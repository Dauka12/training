package app

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	emailintegration "training/backend/internal/integrations/email"
)

func TestPlanGenerationRetriesInvalidAIAndStoresLogs(t *testing.T) {
	provider := &scriptedAIProvider{
		results: []scriptedAIResult{
			{plan: GeneratedPlan{Title: "", Sessions: nil}},
			{plan: validGeneratedPlan("Reliable plan", "Retry succeeded")},
		},
	}

	app, server, client, _ := createVerifiedSessionWithOptions(t, WithAIProvider(provider))
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

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
			{"weekday": "monday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, csrf)

	resp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/plans/generate", map[string]any{
		"generation_type": "initial",
	}, csrf)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected plan generation ok, got %d", resp.StatusCode)
	}

	app.mu.Lock()
	defer app.mu.Unlock()
	user := app.byMail["member@example.com"]
	if user == nil {
		t.Fatal("expected user to exist")
	}
	if len(user.AIGenerationLogs) != 2 {
		t.Fatalf("expected 2 AI logs for retry flow, got %d", len(user.AIGenerationLogs))
	}
	if user.AIGenerationLogs[0].Status != "invalid_response" {
		t.Fatalf("expected first ai log to be invalid_response, got %s", user.AIGenerationLogs[0].Status)
	}
	if user.AIGenerationLogs[1].Status != "ok" {
		t.Fatalf("expected second ai log to be ok, got %s", user.AIGenerationLogs[1].Status)
	}
	if len(user.PlanVersions) != 1 || user.PlanVersions[0].Title != "Reliable plan" {
		t.Fatal("expected final plan saved from successful retry")
	}
}

func TestNotificationPreferencesControlHydrationReminderAndEmail(t *testing.T) {
	sender := &recordingSender{}
	now := time.Date(2026, time.March, 14, 10, 0, 0, 0, time.UTC)

	app, server, client, _ := createVerifiedSessionWithOptions(
		t,
		WithEmailSender(sender),
		WithClock(fixedClock{now: now}),
	)
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

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
			{"weekday": "monday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, csrf)

	prefsOff := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/notifications/preferences", map[string]any{
		"hydration_reminder": false,
		"email_enabled":      false,
	}, csrf)
	if prefsOff.StatusCode != http.StatusOK {
		t.Fatalf("expected prefs save ok, got %d", prefsOff.StatusCode)
	}

	app.mu.Lock()
	user := app.byMail["member@example.com"]
	app.runReminderSweepLocked(now)
	offCount := len(user.Notifications)
	app.mu.Unlock()
	if offCount != 0 {
		t.Fatalf("expected no reminders with hydration preference disabled, got %d", offCount)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("expected no email messages when email disabled, got %d", len(sender.messages))
	}

	prefsOn := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/notifications/preferences", map[string]any{
		"hydration_reminder": true,
		"email_enabled":      true,
	}, csrf)
	if prefsOn.StatusCode != http.StatusOK {
		t.Fatalf("expected prefs enable ok, got %d", prefsOn.StatusCode)
	}

	app.mu.Lock()
	app.runReminderSweepLocked(now.Add(4 * time.Hour))
	onCount := len(user.Notifications)
	prefs := user.NotificationPreferences
	emailLogs := len(user.EmailLogs)
	app.mu.Unlock()

	if !prefs.HydrationReminder || !prefs.EmailEnabled {
		t.Fatal("expected preferences persisted")
	}
	if onCount == 0 {
		t.Fatal("expected hydration reminder notification to be created")
	}
	if len(sender.messages) == 0 {
		t.Fatal("expected hydration reminder email to be sent")
	}
	if emailLogs == 0 {
		t.Fatal("expected email log to be stored")
	}
}

func TestAdminCanViewAIEmailAndAuditLogs(t *testing.T) {
	sender := &recordingSender{}
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithEmailSender(sender),
		WithClock(fixedClock{now: time.Date(2026, time.March, 14, 9, 0, 0, 0, time.UTC)}),
		WithAIProvider(&scriptedAIProvider{results: []scriptedAIResult{{
			plan: validGeneratedPlan("Admin visible plan", "Visible in logs"),
		}}}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "admin@example.com", "admin")
	userClient, userEmail := createUserWithRole(t, app, server.URL, "member@example.com", "user")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")
	userCSRF := cookieValue(t, userClient, server.URL, "csrf")

	doJSON(t, userClient, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
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
			{"weekday": "monday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, userCSRF)
	doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/plans/generate", map[string]any{"generation_type": "initial"}, userCSRF)
	doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/trainers/assign", map[string]any{
		"user_email":    userEmail,
		"trainer_email": "admin@example.com",
	}, adminCSRF)

	aiResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/logs/ai", nil, "")
	if aiResp.StatusCode != http.StatusOK {
		t.Fatalf("expected ai logs ok, got %d", aiResp.StatusCode)
	}
	emailResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/logs/email", nil, "")
	if emailResp.StatusCode != http.StatusOK {
		t.Fatalf("expected email logs ok, got %d", emailResp.StatusCode)
	}
	auditResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/logs/audit", nil, "")
	if auditResp.StatusCode != http.StatusOK {
		t.Fatalf("expected audit logs ok, got %d", auditResp.StatusCode)
	}
}

type scriptedAIProvider struct {
	results []scriptedAIResult
	calls   int
}

type scriptedAIResult struct {
	plan GeneratedPlan
	err  error
}

func (s *scriptedAIProvider) GeneratePlan(GenerationInput) (GeneratedPlan, error) {
	index := s.calls
	s.calls++
	if index >= len(s.results) {
		return GeneratedPlan{}, errors.New("no scripted ai result")
	}
	return s.results[index].plan, s.results[index].err
}

func validGeneratedPlan(title, summary string) GeneratedPlan {
	return GeneratedPlan{
		Title:   title,
		Summary: summary,
		Sessions: []ScheduleItem{{
			ID:               "session-1",
			Weekday:          "monday",
			SessionName:      "Strength A",
			EstimatedMinutes: 45,
			Status:           "planned",
		}},
		Weeks: []PlanWeek{{
			WeekIndex: 1,
			Days: []PlanDay{{
				Weekday:          "monday",
				SessionName:      "Strength A",
				Goal:             "Strength",
				EstimatedMinutes: 45,
				Warmup:           []string{"Cardio"},
				Exercises: []PlanExercise{{
					Order:        1,
					ExerciseID:   "exercise-1",
					ExerciseName: "Press",
					Sets:         3,
					Reps:         "8-10",
					RestSec:      90,
				}},
				Cooldown: []string{"Stretch"},
			}},
		}},
		AdaptationRules: []string{"Move missed sessions to the next valid slot"},
	}
}

type recordingSender struct {
	messages []emailintegration.Message
}

func (r *recordingSender) Send(message emailintegration.Message) error {
	r.messages = append(r.messages, message)
	return nil
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}
