package app

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"log/slog"
)

func TestPGStoreProjectsRuntimeStateIntoRelationalTables(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	store := newTestPGStore(t)
	defer store.Close()

	app := New(
		slog.New(slog.NewTextHandler(ioDiscard{}, nil)),
		WithStateStore(store),
		WithRelationalStore(store),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "admin-repo@example.com", "admin")
	trainerClient, trainerEmail := createUserWithRole(t, app, server.URL, "trainer-repo@example.com", "trainer")
	userClient, userEmail := createUserWithRole(t, app, server.URL, "member-repo@example.com", "user")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")
	trainerCSRF := cookieValue(t, trainerClient, server.URL, "csrf")
	userCSRF := cookieValue(t, userClient, server.URL, "csrf")

	doJSON(t, userClient, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
		"age":                    28,
		"biological_sex":         "male",
		"height_cm":              180,
		"current_weight_kg":      86,
		"target_weight_kg":       80,
		"primary_goal":           "lose_fat",
		"program_duration_weeks": 12,
		"experience_level":       "beginner",
		"activity_level":         "light",
		"training_location":      "mixed",
		"timezone":               "Asia/Qyzylorda",
		"available_training_days": []map[string]any{
			{"weekday": "monday", "duration_min": 60},
			{"weekday": "wednesday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, userCSRF)

	doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/trainers/assign", map[string]any{
		"user_email":    userEmail,
		"trainer_email": trainerEmail,
	}, adminCSRF)

	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/plans/generate", map[string]any{
		"generation_type": "initial",
	}, userCSRF)

	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/support/threads", map[string]any{
		"title": "Need repository help",
		"body":  "Persist this thread",
	}, userCSRF)

	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/discussions/threads", map[string]any{
		"title":    "Repository breakfasts",
		"body":     "Any quick meals?",
		"category": "nutrition",
	}, userCSRF)

	noteResp := doJSON(t, trainerClient, http.MethodPost, server.URL+"/api/v1/trainer/users/"+userEmail+"/notes", map[string]any{
		"body": "Keep sessions short this week",
	}, trainerCSRF)
	if noteResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected trainer note created, got %d", noteResp.StatusCode)
	}

	counts := map[string]int{
		"users":                    0,
		"user_profiles":            0,
		"trainer_assignments":      0,
		"generated_plan_versions":  0,
		"nutrition_plan_versions":  0,
		"workout_schedule_instances": 0,
		"support_threads":          0,
		"public_discussion_threads": 0,
		"trainer_notes":            0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for table := range counts {
		var count int
		if err := store.pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		counts[table] = count
	}

	if counts["users"] < 3 {
		t.Fatalf("expected projected users in relational store, got %+v", counts)
	}
	if counts["trainer_assignments"] != 1 {
		t.Fatalf("expected trainer assignment projection, got %+v", counts)
	}
	if counts["generated_plan_versions"] == 0 || counts["nutrition_plan_versions"] == 0 || counts["workout_schedule_instances"] == 0 {
		t.Fatalf("expected projected plans in relational store, got %+v", counts)
	}
	if counts["support_threads"] == 0 || counts["public_discussion_threads"] == 0 || counts["trainer_notes"] == 0 {
		t.Fatalf("expected projected ops records, got %+v", counts)
	}
}

func TestAdminAndTrainerOpsEndpointsReadFromRepository(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	store := newTestPGStore(t)
	defer store.Close()

	app := New(
		slog.New(slog.NewTextHandler(ioDiscard{}, nil)),
		WithStateStore(store),
		WithRelationalStore(store),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "admin-ops@example.com", "admin")
	trainerClient, trainerEmail := createUserWithRole(t, app, server.URL, "trainer-ops@example.com", "trainer")
	userClient, userEmail := createUserWithRole(t, app, server.URL, "member-ops@example.com", "user")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")
	trainerCSRF := cookieValue(t, trainerClient, server.URL, "csrf")
	userCSRF := cookieValue(t, userClient, server.URL, "csrf")

	doJSON(t, userClient, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
		"age":                    28,
		"biological_sex":         "male",
		"height_cm":              180,
		"current_weight_kg":      86,
		"target_weight_kg":       80,
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
	}, userCSRF)

	doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/trainers/assign", map[string]any{
		"user_email":    userEmail,
		"trainer_email": trainerEmail,
	}, adminCSRF)
	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/support/threads", map[string]any{
		"title": "Need ops help",
		"body":  "Support projection check",
	}, userCSRF)
	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/discussions/threads", map[string]any{
		"title":    "Ops meals",
		"body":     "Any practical meals?",
		"category": "nutrition",
	}, userCSRF)
	doJSON(t, trainerClient, http.MethodPost, server.URL+"/api/v1/trainer/users/"+userEmail+"/notes", map[string]any{
		"body": "Trainer note for repository view",
	}, trainerCSRF)

	adminUsersResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/users", nil, "")
	if adminUsersResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin users ok, got %d", adminUsersResp.StatusCode)
	}
	var adminUsersPayload struct {
		Items []struct {
			Email              string   `json:"email"`
			Roles              []string `json:"roles"`
			AssignedTrainer    string   `json:"assigned_trainer_email"`
			OnboardingDone     bool     `json:"onboarding_done"`
			ActivePlanVersions int      `json:"active_plan_versions"`
		} `json:"items"`
	}
	_ = json.NewDecoder(adminUsersResp.Body).Decode(&adminUsersPayload)
	if len(adminUsersPayload.Items) == 0 {
		t.Fatal("expected admin users list from repository")
	}

	adminTrainersResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/trainers", nil, "")
	if adminTrainersResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin trainers ok, got %d", adminTrainersResp.StatusCode)
	}

	notesResp := doJSON(t, trainerClient, http.MethodGet, server.URL+"/api/v1/trainer/users/"+userEmail+"/notes", nil, "")
	if notesResp.StatusCode != http.StatusOK {
		t.Fatalf("expected trainer notes ok, got %d", notesResp.StatusCode)
	}

	supportResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/support/threads", nil, "")
	if supportResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin support threads ok, got %d", supportResp.StatusCode)
	}

	discussionResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/discussions/threads", nil, "")
	if discussionResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin discussion threads ok, got %d", discussionResp.StatusCode)
	}

	notificationLogsResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/logs/notifications", nil, "")
	if notificationLogsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected notification log view ok, got %d", notificationLogsResp.StatusCode)
	}
}

func TestAdminCanModerateThreadsAndReadEmailLogsFromRepository(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	store := newTestPGStore(t)
	defer store.Close()

	app := New(
		slog.New(slog.NewTextHandler(ioDiscard{}, nil)),
		WithStateStore(store),
		WithRelationalStore(store),
		WithEmailSender(&recordingSender{}),
		WithClock(fixedClock{now: time.Date(2026, time.March, 15, 9, 0, 0, 0, time.UTC)}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "admin-moderation@example.com", "admin")
	userClient, userEmail := createUserWithRole(t, app, server.URL, "member-moderation@example.com", "user")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")
	userCSRF := cookieValue(t, userClient, server.URL, "csrf")

	doJSON(t, userClient, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
		"age":                      30,
		"biological_sex":           "female",
		"height_cm":                168,
		"current_weight_kg":        74,
		"target_weight_kg":         68,
		"primary_goal":             "lose_fat",
		"program_duration_weeks":   12,
		"experience_level":         "beginner",
		"activity_level":           "light",
		"training_location":        "home",
		"timezone":                 "Asia/Qyzylorda",
		"preferred_training_style": "low_impact_strength",
		"preferred_meal_style":     "simple_prep",
		"hydration_preference":     "small_frequent_sips",
		"available_training_days": []map[string]any{
			{"weekday": "monday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, userCSRF)

	doJSON(t, userClient, http.MethodPut, server.URL+"/api/v1/notifications/preferences", map[string]any{
		"hydration_reminder": true,
		"email_enabled":      true,
	}, userCSRF)

	supportResp := doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/support/threads", map[string]any{
		"title": "Moderate my support thread",
		"body":  "Please review this issue",
	}, userCSRF)
	if supportResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected support thread created, got %d", supportResp.StatusCode)
	}
	var supportPayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	_ = json.NewDecoder(supportResp.Body).Decode(&supportPayload)

	discussionResp := doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/discussions/threads", map[string]any{
		"title":    "Moderate my discussion",
		"body":     "Please review this post",
		"category": "general",
	}, userCSRF)
	if discussionResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected discussion thread created, got %d", discussionResp.StatusCode)
	}
	var discussionPayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	_ = json.NewDecoder(discussionResp.Body).Decode(&discussionPayload)

	supportModerationResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/support/threads/"+supportPayload.Thread.ID+"/status", map[string]any{
		"status": "resolved",
	}, adminCSRF)
	if supportModerationResp.StatusCode != http.StatusOK {
		t.Fatalf("expected support moderation ok, got %d", supportModerationResp.StatusCode)
	}

	discussionModerationResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/discussions/threads/"+discussionPayload.Thread.ID+"/moderation", map[string]any{
		"status": "hidden",
	}, adminCSRF)
	if discussionModerationResp.StatusCode != http.StatusOK {
		t.Fatalf("expected discussion moderation ok, got %d", discussionModerationResp.StatusCode)
	}

	app.mu.Lock()
	app.runReminderSweepLocked(time.Date(2026, time.March, 15, 13, 0, 0, 0, time.UTC))
	app.mu.Unlock()

	adminSupportResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/support/threads", nil, "")
	if adminSupportResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin support threads ok, got %d", adminSupportResp.StatusCode)
	}
	var adminSupportPayload struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	_ = json.NewDecoder(adminSupportResp.Body).Decode(&adminSupportPayload)
	foundResolved := false
	for _, item := range adminSupportPayload.Items {
		if item.Status == "resolved" {
			foundResolved = true
		}
	}
	if !foundResolved {
		t.Fatalf("expected moderated support status in repository response, got %+v", adminSupportPayload.Items)
	}

	adminDiscussionResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/discussions/threads", nil, "")
	if adminDiscussionResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin discussion threads ok, got %d", adminDiscussionResp.StatusCode)
	}
	var adminDiscussionPayload struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	_ = json.NewDecoder(adminDiscussionResp.Body).Decode(&adminDiscussionPayload)
	foundHidden := false
	for _, item := range adminDiscussionPayload.Items {
		if item.Status == "hidden" {
			foundHidden = true
		}
	}
	if !foundHidden {
		t.Fatalf("expected moderated discussion status in repository response, got %+v", adminDiscussionPayload.Items)
	}

	emailLogsResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/logs/email", nil, "")
	if emailLogsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin email logs ok, got %d", emailLogsResp.StatusCode)
	}
	var emailLogsPayload struct {
		Items []struct {
			Type   string `json:"type"`
			Status string `json:"status"`
			To     string `json:"to"`
		} `json:"items"`
	}
	_ = json.NewDecoder(emailLogsResp.Body).Decode(&emailLogsPayload)
	if len(emailLogsPayload.Items) == 0 {
		t.Fatal("expected repository-backed email logs to be returned")
	}
	foundHydrationEmail := false
	for _, item := range emailLogsPayload.Items {
		if item.Type == "hydration_reminder" && item.Status == "sent" && item.To == userEmail {
			foundHydrationEmail = true
		}
	}
	if !foundHydrationEmail {
		t.Fatalf("expected hydration reminder email log in repository response, got %+v", emailLogsPayload.Items)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func newTestPGStore(t *testing.T) *PGStateStore {
	t.Helper()

	databaseURL := discoverTestDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := RunMigrations(ctx, databaseURL, discoverMigrationsDir(t)); err != nil {
		t.Skipf("postgres migrations unavailable for repository integration test: %v", err)
	}

	store, err := NewPGStateStore(ctx, databaseURL)
	if err != nil {
		t.Skipf("postgres not available for repository integration test: %v", err)
	}
	truncateProjectionTables(t, store)
	return store
}

func truncateProjectionTables(t *testing.T, store *PGStateStore) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := store.pool.Exec(ctx, `
		TRUNCATE TABLE
			app_runtime_state,
			admin_audit_logs,
			trainer_notes,
			public_discussion_replies,
			public_discussion_threads,
			support_messages,
			support_threads,
			email_logs,
			ai_generation_logs,
			notification_preferences,
			notifications,
			weekly_checkins,
			hydration_targets,
			water_logs,
			meal_logs,
			workout_logs,
			workout_schedule_instances,
			nutrition_plan_versions,
			workout_day_exercises,
			workout_day_templates,
			generated_plan_versions,
			exercise_media,
			availability_rules,
			user_equipment,
			dietary_preferences,
			user_injuries,
			user_goals,
			trainer_assignments,
			user_profiles,
			password_reset_tokens,
			email_verification_tokens,
			sessions,
			user_roles,
			roles,
			users,
			exercise_catalog_translations,
			exercise_catalog,
			equipment_catalog_translations,
			equipment_catalog
		`)
	if err != nil {
		t.Fatalf("truncate projection tables: %v", err)
	}
}

func discoverTestDatabaseURL(t *testing.T) string {
	t.Helper()

	if raw := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); raw != "" {
		return raw
	}
	if raw := strings.TrimSpace(os.Getenv("DATABASE_URL")); raw != "" {
		return normalizeTestDatabaseURL(raw)
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to locate current file for database discovery")
	}
	rootDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	envPath := filepath.Join(rootDir, ".env")
	file, err := os.Open(envPath)
	if err != nil {
		t.Skipf("unable to open root env for database discovery: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "DATABASE_URL=") {
			continue
		}
		return normalizeTestDatabaseURL(strings.TrimPrefix(line, "DATABASE_URL="))
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan env: %v", err)
	}
	t.Skip("DATABASE_URL not found for repository integration tests")
	return ""
}

func discoverMigrationsDir(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to locate current file for migrations discovery")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "migrations"))
}

func normalizeTestDatabaseURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	if parsed.Hostname() == "postgres" {
		parsed.Host = "localhost:5433"
	}
	return parsed.String()
}
