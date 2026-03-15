package app

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterVerifyLoginForgotResetFlow(t *testing.T) {
	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newClient(t)

	resp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/register", map[string]any{
		"email":    "user@example.com",
		"password": "supersecret1",
		"locale":   "ru",
	}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	app.mu.Lock()
	verifyToken := app.byMail["user@example.com"].VerifyToken
	app.mu.Unlock()

	verifyResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/verify-email", map[string]any{
		"token": verifyToken,
	}, "")
	if verifyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected verify ok, got %d", verifyResp.StatusCode)
	}

	loginResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "supersecret1",
	}, "")
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected login ok, got %d", loginResp.StatusCode)
	}

	meResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected me ok, got %d", meResp.StatusCode)
	}

	forgotResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/forgot-password", map[string]any{
		"email": "user@example.com",
	}, "")
	if forgotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected forgot ok, got %d", forgotResp.StatusCode)
	}

	app.mu.Lock()
	resetToken := app.byMail["user@example.com"].ResetToken
	app.mu.Unlock()

	resetResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/reset-password", map[string]any{
		"token":        resetToken,
		"new_password": "newsecret123",
	}, "")
	if resetResp.StatusCode != http.StatusOK {
		t.Fatalf("expected reset ok, got %d", resetResp.StatusCode)
	}

	oldSessionResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if oldSessionResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected old session revoked, got %d", oldSessionResp.StatusCode)
	}

	reloginResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "newsecret123",
	}, "")
	if reloginResp.StatusCode != http.StatusOK {
		t.Fatalf("expected relogin ok, got %d", reloginResp.StatusCode)
	}
}

func TestPlanGenerationVersioningAndTracking(t *testing.T) {
	app, server, client, email := createVerifiedSession(t)
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

	onboardingResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
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
			{"weekday": "wednesday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, csrf)
	if onboardingResp.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding ok, got %d", onboardingResp.StatusCode)
	}

	planResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/plans/generate", map[string]any{
		"generation_type": "initial",
	}, csrf)
	if planResp.StatusCode != http.StatusOK {
		t.Fatalf("expected plan generate ok, got %d", planResp.StatusCode)
	}

	activeResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/plans/active", nil, "")
	if activeResp.StatusCode != http.StatusOK {
		t.Fatalf("expected active plan ok, got %d", activeResp.StatusCode)
	}
	var activePayload struct {
		Plan struct {
			ID              string   `json:"id"`
			VersionNo       int      `json:"version_no"`
			AdaptationRules []string `json:"adaptation_rules"`
			Weeks           []struct {
				Days []struct {
					Exercises []struct {
						ExerciseID   string `json:"exercise_id"`
						ExerciseName string `json:"exercise_name"`
					} `json:"exercises"`
				} `json:"days"`
			} `json:"weeks"`
			Schedule []struct {
				ID string `json:"id"`
			} `json:"schedule"`
			NutritionPlan struct {
				MealExamples []struct {
					Slot string `json:"slot"`
				} `json:"meal_examples"`
			} `json:"nutrition"`
		} `json:"plan"`
	}
	_ = json.NewDecoder(activeResp.Body).Decode(&activePayload)
	if activePayload.Plan.VersionNo != 1 {
		t.Fatalf("expected version 1, got %d", activePayload.Plan.VersionNo)
	}
	if len(activePayload.Plan.Weeks) == 0 || len(activePayload.Plan.Weeks[0].Days) == 0 {
		t.Fatalf("expected detailed weeks in active plan, got %+v", activePayload.Plan.Weeks)
	}
	if len(activePayload.Plan.Weeks[0].Days[0].Exercises) == 0 {
		t.Fatalf("expected exercises in detailed weeks, got %+v", activePayload.Plan.Weeks[0].Days[0])
	}
	if activePayload.Plan.Weeks[0].Days[0].Exercises[0].ExerciseID == "" {
		t.Fatalf("expected exercise id in detailed weeks, got %+v", activePayload.Plan.Weeks[0].Days[0].Exercises[0])
	}
	if len(activePayload.Plan.NutritionPlan.MealExamples) == 0 {
		t.Fatalf("expected meal examples in nutrition plan, got %+v", activePayload.Plan.NutritionPlan)
	}
	if len(activePayload.Plan.AdaptationRules) == 0 {
		t.Fatalf("expected adaptation rules in active plan, got %+v", activePayload.Plan.AdaptationRules)
	}

	workoutResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/tracking/workouts/log", map[string]any{
		"schedule_id":     activePayload.Plan.Schedule[0].ID,
		"status":          "skipped",
		"discomfort_flag": true,
		"difficulty":      8,
		"note":            "knee discomfort",
		"completion_time": "2026-03-14T10:00:00Z",
	}, csrf)
	if workoutResp.StatusCode != http.StatusOK {
		t.Fatalf("expected workout log ok, got %d", workoutResp.StatusCode)
	}

	mealResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/tracking/meals", map[string]any{
		"status": "followed",
		"note":   "good day",
	}, csrf)
	if mealResp.StatusCode != http.StatusOK {
		t.Fatalf("expected meal log ok, got %d", mealResp.StatusCode)
	}

	waterResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/tracking/water", map[string]any{
		"amount_ml": 500,
	}, csrf)
	if waterResp.StatusCode != http.StatusOK {
		t.Fatalf("expected water log ok, got %d", waterResp.StatusCode)
	}

	checkinResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/checkins/weekly", map[string]any{
		"weight_kg":            86,
		"energy_level":         2,
		"availability_changed": true,
		"note":                 "schedule changed",
	}, csrf)
	if checkinResp.StatusCode != http.StatusOK {
		t.Fatalf("expected weekly checkin ok, got %d", checkinResp.StatusCode)
	}

	app.mu.Lock()
	user := app.byMail[email]
	versionCount := len(user.PlanVersions)
	latestVersion := user.PlanVersions[len(user.PlanVersions)-1].VersionNo
	app.mu.Unlock()
	if versionCount != 2 {
		t.Fatalf("expected 2 plan versions, got %d", versionCount)
	}
	if latestVersion != 2 {
		t.Fatalf("expected latest version 2, got %d", latestVersion)
	}
}

func TestRoleAccessAndSupportDiscussionNotifications(t *testing.T) {
	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, adminEmail := createUserWithRole(t, app, server.URL, "admin@example.com", "admin")
	trainerClient, trainerEmail := createUserWithRole(t, app, server.URL, "trainer@example.com", "trainer")
	userClient, userEmail := createUserWithRole(t, app, server.URL, "member@example.com", "user")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")
	trainerCSRF := cookieValue(t, trainerClient, server.URL, "csrf")
	userCSRF := cookieValue(t, userClient, server.URL, "csrf")

	forbiddenResp := doJSON(t, userClient, http.MethodGet, server.URL+"/api/v1/admin/catalog/equipment", nil, "")
	if forbiddenResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected admin forbidden, got %d", forbiddenResp.StatusCode)
	}

	assignResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/trainers/assign", map[string]any{
		"user_email":    userEmail,
		"trainer_email": trainerEmail,
	}, adminCSRF)
	if assignResp.StatusCode != http.StatusOK {
		t.Fatalf("expected trainer assign ok, got %d", assignResp.StatusCode)
	}

	trainerUsersResp := doJSON(t, trainerClient, http.MethodGet, server.URL+"/api/v1/trainer/users", nil, "")
	if trainerUsersResp.StatusCode != http.StatusOK {
		t.Fatalf("expected trainer users ok, got %d", trainerUsersResp.StatusCode)
	}

	threadResp := doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/support/threads", map[string]any{
		"title": "Need help",
		"body":  "My knee hurts",
	}, userCSRF)
	if threadResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected support thread created, got %d", threadResp.StatusCode)
	}
	var threadPayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	_ = json.NewDecoder(threadResp.Body).Decode(&threadPayload)

	replyResp := doJSON(t, trainerClient, http.MethodPost, server.URL+"/api/v1/support/threads/"+threadPayload.Thread.ID+"/messages", map[string]any{
		"body": "Please reduce intensity this week",
	}, trainerCSRF)
	if replyResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected support reply created, got %d", replyResp.StatusCode)
	}

	discussionResp := doJSON(t, userClient, http.MethodPost, server.URL+"/api/v1/discussions/threads", map[string]any{
		"title":    "Meal prep ideas",
		"body":     "Any simple breakfasts?",
		"category": "nutrition",
	}, userCSRF)
	if discussionResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected discussion created, got %d", discussionResp.StatusCode)
	}
	var discussionPayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	_ = json.NewDecoder(discussionResp.Body).Decode(&discussionPayload)

	replyDiscussionResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/discussions/threads/"+discussionPayload.Thread.ID+"/replies", map[string]any{
		"body": "Try oats and eggs",
	}, adminCSRF)
	if replyDiscussionResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected discussion reply created, got %d", replyDiscussionResp.StatusCode)
	}

	app.mu.Lock()
	user := app.byMail[userEmail]
	notificationCount := len(user.Notifications)
	app.mu.Unlock()
	if notificationCount < 2 {
		t.Fatalf("expected support and discussion notifications, got %d", notificationCount)
	}

	_ = adminEmail
}

func TestOnboardingPersistsExtendedPlanningPreferencesAndReturnsProfile(t *testing.T) {
	_, server, client, _ := createVerifiedSession(t)
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

	onboardingResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
		"age":                      31,
		"biological_sex":           "female",
		"height_cm":                167,
		"current_weight_kg":        72,
		"target_weight_kg":         65,
		"primary_goal":             "lose_fat",
		"program_duration_weeks":   10,
		"experience_level":         "intermediate",
		"activity_level":           "moderate",
		"training_location":        "home",
		"timezone":                 "Asia/Qyzylorda",
		"preferred_training_style": "low_impact_strength",
		"preferred_meal_style":     "simple_prep",
		"hydration_preference":     "small_frequent_sips",
		"injuries":                 []string{"knee_discomfort"},
		"dietary_preferences":      []string{"high_protein"},
		"avoid_foods":              []string{"lactose"},
		"available_training_days": []map[string]any{
			{"weekday": "tuesday", "duration_min": 40},
			{"weekday": "thursday", "duration_min": 50},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, csrf)
	if onboardingResp.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding ok, got %d", onboardingResp.StatusCode)
	}

	meResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected me ok, got %d", meResp.StatusCode)
	}

	var payload struct {
		OnboardingDone bool `json:"onboarding_done"`
		Profile struct {
			PreferredTrainingStyle string `json:"preferred_training_style"`
			PreferredMealStyle     string `json:"preferred_meal_style"`
			HydrationPreference    string `json:"hydration_preference"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(meResp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload.OnboardingDone {
		t.Fatal("expected onboarding_done to be true")
	}
	if payload.Profile.PreferredTrainingStyle != "low_impact_strength" {
		t.Fatalf("expected preferred training style saved, got %q", payload.Profile.PreferredTrainingStyle)
	}
	if payload.Profile.PreferredMealStyle != "simple_prep" {
		t.Fatalf("expected preferred meal style saved, got %q", payload.Profile.PreferredMealStyle)
	}
	if payload.Profile.HydrationPreference != "small_frequent_sips" {
		t.Fatalf("expected hydration preference saved, got %q", payload.Profile.HydrationPreference)
	}
}

func TestPreferencesExposeWaterOverrideAndAllowReset(t *testing.T) {
	now := time.Date(2026, time.March, 15, 9, 0, 0, 0, time.UTC)
	_, server, client, _ := createVerifiedSessionWithOptions(t, WithClock(fixedClock{now: now}))
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

	onboardingResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
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
	}, csrf)
	if onboardingResp.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding ok, got %d", onboardingResp.StatusCode)
	}

	meResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected me ok, got %d", meResp.StatusCode)
	}
	var initial struct {
		WaterTargetML   int `json:"water_target_ml"`
		WaterOverrideML int `json:"water_override_ml"`
	}
	if err := json.NewDecoder(meResp.Body).Decode(&initial); err != nil {
		t.Fatal(err)
	}
	if initial.WaterTargetML <= 0 {
		t.Fatalf("expected deterministic water target, got %d", initial.WaterTargetML)
	}
	if initial.WaterOverrideML != 0 {
		t.Fatalf("expected no override by default, got %d", initial.WaterOverrideML)
	}

	saveOverrideResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/me/preferences", map[string]any{
		"locale":            "ru",
		"theme":             "light",
		"water_override_ml": 3100,
	}, csrf)
	if saveOverrideResp.StatusCode != http.StatusOK {
		t.Fatalf("expected save override ok, got %d", saveOverrideResp.StatusCode)
	}

	meWithOverrideResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if meWithOverrideResp.StatusCode != http.StatusOK {
		t.Fatalf("expected me with override ok, got %d", meWithOverrideResp.StatusCode)
	}
	var withOverride struct {
		WaterTargetML   int `json:"water_target_ml"`
		WaterOverrideML int `json:"water_override_ml"`
	}
	if err := json.NewDecoder(meWithOverrideResp.Body).Decode(&withOverride); err != nil {
		t.Fatal(err)
	}
	if withOverride.WaterOverrideML != 3100 {
		t.Fatalf("expected override 3100, got %d", withOverride.WaterOverrideML)
	}
	if withOverride.WaterTargetML != 3100 {
		t.Fatalf("expected target to follow override, got %d", withOverride.WaterTargetML)
	}

	resetOverrideResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/me/preferences", map[string]any{
		"locale":            "ru",
		"theme":             "light",
		"water_override_ml": 0,
	}, csrf)
	if resetOverrideResp.StatusCode != http.StatusOK {
		t.Fatalf("expected reset override ok, got %d", resetOverrideResp.StatusCode)
	}

	meAfterResetResp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/me", nil, "")
	if meAfterResetResp.StatusCode != http.StatusOK {
		t.Fatalf("expected me after reset ok, got %d", meAfterResetResp.StatusCode)
	}
	var afterReset struct {
		WaterTargetML   int `json:"water_target_ml"`
		WaterOverrideML int `json:"water_override_ml"`
	}
	if err := json.NewDecoder(meAfterResetResp.Body).Decode(&afterReset); err != nil {
		t.Fatal(err)
	}
	if afterReset.WaterOverrideML != 0 {
		t.Fatalf("expected override reset to 0, got %d", afterReset.WaterOverrideML)
	}
	if afterReset.WaterTargetML != initial.WaterTargetML {
		t.Fatalf("expected deterministic target restored to %d, got %d", initial.WaterTargetML, afterReset.WaterTargetML)
	}
}

func TestWeeklyCheckinIgnoresOldSkippedWorkoutsOutsideSevenDayWindow(t *testing.T) {
	now := time.Date(2026, time.March, 15, 9, 0, 0, 0, time.UTC)
	app, server, client, email := createVerifiedSessionWithOptions(t, WithClock(fixedClock{now: now}))
	defer server.Close()
	csrf := cookieValue(t, client, server.URL, "csrf")

	onboardingResp := doJSON(t, client, http.MethodPut, server.URL+"/api/v1/onboarding", map[string]any{
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
			{"weekday": "wednesday", "duration_min": 45},
		},
		"equipment_ids": []string{"10000000-0000-0000-0000-000000000001"},
	}, csrf)
	if onboardingResp.StatusCode != http.StatusOK {
		t.Fatalf("expected onboarding ok, got %d", onboardingResp.StatusCode)
	}

	app.mu.Lock()
	user := app.byMail[email]
	user.WorkoutLogs = []WorkoutLog{
		{Status: "skipped", CompletionTime: now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)},
		{Status: "skipped", CompletionTime: now.Add(-20 * 24 * time.Hour).Format(time.RFC3339)},
	}
	app.mu.Unlock()

	checkinResp := doJSON(t, client, http.MethodPost, server.URL+"/api/v1/checkins/weekly", map[string]any{
		"weight_kg":            85,
		"energy_level":         3,
		"availability_changed": false,
		"equipment_changed":    false,
		"injury_changed":       false,
		"note":                 "steady week",
	}, csrf)
	if checkinResp.StatusCode != http.StatusOK {
		t.Fatalf("expected weekly checkin ok, got %d", checkinResp.StatusCode)
	}

	var payload struct {
		Regenerated bool   `json:"regenerated"`
		Reason      string `json:"reason"`
	}
	if err := json.NewDecoder(checkinResp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.Regenerated {
		t.Fatalf("expected old skipped workouts to be ignored, got regeneration reason %q", payload.Reason)
	}
}

func createVerifiedSession(t *testing.T) (*App, *httptest.Server, *http.Client, string) {
	return createVerifiedSessionWithOptions(t)
}

func createVerifiedSessionWithOptions(t *testing.T, options ...Option) (*App, *httptest.Server, *http.Client, string) {
	t.Helper()
	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)), options...)
	server := httptest.NewServer(app.Routes())
	client := newClient(t)
	email := "member@example.com"

	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret1",
		"locale":   "ru",
	}, "")
	app.mu.Lock()
	token := app.byMail[email].VerifyToken
	app.mu.Unlock()
	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/verify-email", map[string]any{"token": token}, "")
	doJSON(t, client, http.MethodPost, server.URL+"/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "supersecret1",
	}, "")
	return app, server, client, email
}

func createUserWithRole(t *testing.T, app *App, baseURL, email, role string) (*http.Client, string) {
	t.Helper()
	client := newClient(t)
	doJSON(t, client, http.MethodPost, baseURL+"/api/v1/auth/register", map[string]any{
		"email":    email,
		"password": "supersecret1",
		"locale":   "ru",
	}, "")
	app.mu.Lock()
	token := app.byMail[email].VerifyToken
	app.byMail[email].Roles = []string{role}
	app.mu.Unlock()
	doJSON(t, client, http.MethodPost, baseURL+"/api/v1/auth/verify-email", map[string]any{"token": token}, "")
	doJSON(t, client, http.MethodPost, baseURL+"/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": "supersecret1",
	}, "")
	return client, email
}

func newClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Jar: jar}
}

func doJSON(t *testing.T, client *http.Client, method, url string, payload any, csrf string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, url, &body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
