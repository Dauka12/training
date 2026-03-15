package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAuthenticatedMutationRequiresCSRFToken(t *testing.T) {
	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newClient(t)

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/register", map[string]any{
		"email":    "csrf@example.com",
		"password": "supersecret1",
		"locale":   "ru",
	}, "")

	app.mu.Lock()
	verifyToken := app.byMail["csrf@example.com"].VerifyToken
	app.mu.Unlock()

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/verify-email", map[string]any{
		"token": verifyToken,
	}, "")

	loginResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"email":    "csrf@example.com",
		"password": "supersecret1",
	}, "")
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login ok, got %d", loginResp.StatusCode)
	}

	csrfToken := cookieValue(t, client, server.URL, "csrf")
	if csrfToken == "" {
		t.Fatal("expected csrf cookie to be set")
	}

	forbiddenResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/tracking/water", map[string]any{
		"amount_ml": 250,
	}, "")
	if forbiddenResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected csrf rejection, got %d", forbiddenResp.StatusCode)
	}

	okResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/tracking/water", map[string]any{
		"amount_ml": 250,
	}, csrfToken)
	if okResp.StatusCode != http.StatusOK {
		t.Fatalf("expected csrf-approved request, got %d", okResp.StatusCode)
	}
}

func TestRepeatedInvalidLoginIsRateLimited(t *testing.T) {
	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newClient(t)

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/register", map[string]any{
		"email":    "limit@example.com",
		"password": "supersecret1",
		"locale":   "ru",
	}, "")

	app.mu.Lock()
	verifyToken := app.byMail["limit@example.com"].VerifyToken
	app.mu.Unlock()

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/verify-email", map[string]any{
		"token": verifyToken,
	}, "")

	var lastStatus int
	for range 6 {
		resp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
			"email":    "limit@example.com",
			"password": "wrong-password",
		}, "")
		lastStatus = resp.StatusCode
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after repeated invalid login attempts, got %d", lastStatus)
	}
}

func TestAppRestoresPersistedStateFromStore(t *testing.T) {
	store := &memoryStateStore{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := New(logger, WithStateStore(store))
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newClient(t)

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/register", map[string]any{
		"email":    "persist@example.com",
		"password": "supersecret1",
		"locale":   "ru",
	}, "")

	app.mu.Lock()
	verifyToken := app.byMail["persist@example.com"].VerifyToken
	app.mu.Unlock()

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/verify-email", map[string]any{
		"token": verifyToken,
	}, "")

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"email":    "persist@example.com",
		"password": "supersecret1",
	}, "")

	doJSON(t, client, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
		"age":                    28,
		"biological_sex":         "male",
		"height_cm":              180,
		"current_weight_kg":      86,
		"target_weight_kg":       78,
		"primary_goal":           "lose_fat",
		"program_duration_weeks": 12,
		"experience_level":       "beginner",
		"activity_level":         "light",
		"training_location":      "mixed",
		"timezone":               "Asia/Qyzylorda",
		"available_training_days": []map[string]any{
			{"weekday": "monday", "duration_min": 60},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, cookieValue(t, client, server.URL, "csrf"))

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/plans/generate", map[string]any{
		"generation_type": "initial",
	}, cookieValue(t, client, server.URL, "csrf"))

	restored := New(logger, WithStateStore(store))
	restored.mu.Lock()
	defer restored.mu.Unlock()

	user := restored.byMail["persist@example.com"]
	if user == nil {
		t.Fatal("expected user restored from persistence store")
	}
	if !user.Verified {
		t.Fatal("expected verified state restored")
	}
	if len(user.PlanVersions) != 1 {
		t.Fatalf("expected plan versions restored, got %d", len(user.PlanVersions))
	}
}

func cookieValue(t *testing.T, client *http.Client, rawURL, name string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range client.Jar.Cookies(parsed) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

type memoryStateStore struct {
	payload []byte
}

func (m *memoryStateStore) Load(context.Context) ([]byte, error) {
	return append([]byte(nil), m.payload...), nil
}

func (m *memoryStateStore) Save(_ context.Context, payload []byte) error {
	m.payload = append([]byte(nil), payload...)
	return nil
}
