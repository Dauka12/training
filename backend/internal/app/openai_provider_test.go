package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleProviderGeneratePlanUsesStrictContract(t *testing.T) {
	var authHeader string
	var requestPath string
	var requestBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		requestPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"schema_version\":\"1.0\",\"status\":\"ok\",\"plan_title\":\"План на 12 недель\",\"plan_summary\":\"Краткий план\",\"warnings\":[\"Берегите колено\"],\"weeks\":[{\"week_index\":1,\"days\":[{\"weekday\":\"monday\",\"session_name\":\"Верх тела A\",\"goal\":\"Силовая тренировка\",\"estimated_duration_min\":45,\"warmup\":[\"5 минут лёгкого кардио\"],\"exercises\":[{\"order\":1,\"exercise_id\":\"20000000-0000-0000-0000-000000000001\",\"sets\":3,\"reps\":\"8-10\",\"rest_sec\":90,\"effort_note\":\"RPE 7\",\"notes\":\"Контролируй движение\",\"substitution_exercise_ids\":[\"20000000-0000-0000-0000-000000000002\"]}],\"cooldown\":[\"Лёгкая растяжка\"]}]}],\"nutrition\":{\"daily_calories\":2200,\"protein_g\":160,\"carbs_g\":220,\"fat_g\":70,\"daily_water_ml\":2600},\"adaptation_rules\":[\"Если тренировка пропущена, перенести на следующий доступный день\"]}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gemini-test",
		Client:  server.Client(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	plan, err := provider.GeneratePlan(GenerationInput{
		Locale:  "ru",
		UserRef: "internal-1",
		Profile: UserProfile{
			Age:                   28,
			BiologicalSex:         "male",
			HeightCM:              180,
			CurrentWeightKG:       86,
			PrimaryGoal:           "lose_fat",
			ExperienceLevel:       "beginner",
			ActivityLevel:         "light",
			TrainingLocation:      "mixed",
			Injuries:              []string{"knee_discomfort"},
			DietaryPreferences:    []string{"high_protein"},
			AvoidFoods:            []string{"lactose"},
			PreferredTrainingStyle: "low_impact_strength",
			PreferredMealStyle:     "simple_prep",
			HydrationPreference:    "small_frequent_sips",
			AvailableTrainingDays: []AvailabilityDay{{Weekday: "monday", DurationMin: 45}},
			EquipmentIDs:          []string{"10000000-0000-0000-0000-000000000001"},
		},
		Targets: NutritionTarget{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
		Candidates:        seedExerciseCatalog(),
		SelectedEquipment: seedEquipmentCatalog(),
		History: GenerationHistory{
			CompletedSessionsLast14Days:  1,
			MissedSessionsLast14Days:     0,
			MealAdherenceLast14Days:      70,
			HydrationAdherenceLast14Days: 80,
			LatestWeightKG:               86,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if authHeader != "Bearer test-key" {
		t.Fatalf("expected bearer auth, got %q", authHeader)
	}
	if requestPath != "/chat/completions" {
		t.Fatalf("expected /chat/completions, got %q", requestPath)
	}
	if !strings.Contains(requestBody, "\"model\":\"gemini-test\"") {
		t.Fatalf("expected model in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "candidate_exercises") {
		t.Fatalf("expected candidate exercises in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "selected_equipment") {
		t.Fatalf("expected selected equipment in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "history") {
		t.Fatalf("expected history in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "preferred_training_style") {
		t.Fatalf("expected preferred training style in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "preferred_meal_style") {
		t.Fatalf("expected preferred meal style in request body, got %s", requestBody)
	}
	if !strings.Contains(requestBody, "hydration_preference") {
		t.Fatalf("expected hydration preference in request body, got %s", requestBody)
	}
	if strings.Contains(requestBody, "user@example.com") {
		t.Fatalf("expected request body to omit email, got %s", requestBody)
	}
	if plan.Title != "План на 12 недель" {
		t.Fatalf("unexpected title: %s", plan.Title)
	}
	if len(plan.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(plan.Sessions))
	}
	if plan.Sessions[0].Weekday != "monday" || plan.Sessions[0].EstimatedMinutes != 45 {
		t.Fatalf("unexpected session mapping: %+v", plan.Sessions[0])
	}
	if len(plan.Weeks) != 1 || len(plan.Weeks[0].Days) != 1 {
		t.Fatalf("expected weeks to be preserved, got %+v", plan.Weeks)
	}
	if len(plan.Weeks[0].Days[0].Exercises) != 1 {
		t.Fatalf("expected exercise details to be preserved, got %+v", plan.Weeks[0].Days[0].Exercises)
	}
	if plan.Weeks[0].Days[0].Exercises[0].ExerciseName == "" {
		t.Fatalf("expected exercise name to be resolved, got %+v", plan.Weeks[0].Days[0].Exercises[0])
	}
	if len(plan.AdaptationRules) != 1 {
		t.Fatalf("expected adaptation rules to be preserved, got %+v", plan.AdaptationRules)
	}
	if got := len(plan.Weeks[0].Days[0].Warmup); got != 1 {
		t.Fatalf("expected warmup string to normalize into one item, got %d", got)
	}
	if got := len(plan.Weeks[0].Days[0].Cooldown); got != 1 {
		t.Fatalf("expected cooldown string to normalize into one item, got %d", got)
	}
}

func TestOpenAICompatibleProviderRejectsUnknownExerciseID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"schema_version\":\"1.0\",\"status\":\"ok\",\"plan_title\":\"План на 12 недель\",\"plan_summary\":\"Краткий план\",\"warnings\":[],\"weeks\":[{\"week_index\":1,\"days\":[{\"weekday\":\"monday\",\"session_name\":\"Верх тела A\",\"goal\":\"Силовая тренировка\",\"estimated_duration_min\":45,\"warmup\":[],\"exercises\":[{\"order\":1,\"exercise_id\":\"unknown-exercise\",\"sets\":3,\"reps\":\"8-10\",\"rest_sec\":90,\"effort_note\":\"RPE 7\",\"notes\":\"Контролируй движение\",\"substitution_exercise_ids\":[]}],\"cooldown\":[]}]}],\"nutrition\":{\"daily_calories\":2200,\"protein_g\":160,\"carbs_g\":220,\"fat_g\":70,\"daily_water_ml\":2600},\"adaptation_rules\":[]}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gemini-test",
		Client:  server.Client(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	_, err := provider.GeneratePlan(GenerationInput{
		Locale: "ru",
		Profile: UserProfile{
			AvailableTrainingDays: []AvailabilityDay{{Weekday: "monday", DurationMin: 45}},
		},
		Targets: NutritionTarget{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
		Candidates: seedExerciseCatalog(),
	})
	if err == nil {
		t.Fatal("expected validation error for unknown exercise id")
	}
}

func TestOpenAICompatibleProviderAcceptsStringWarmupAndCooldown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "{\"schema_version\":\"1.0\",\"status\":\"ok\",\"plan_title\":\"План на 12 недель\",\"plan_summary\":\"Краткий план\",\"warnings\":[],\"weeks\":[{\"week_index\":1,\"days\":[{\"weekday\":\"monday\",\"session_name\":\"Верх тела A\",\"goal\":\"Силовая тренировка\",\"estimated_duration_min\":45,\"warmup\":\"5 минут лёгкого кардио\",\"exercises\":[{\"order\":1,\"exercise_id\":\"20000000-0000-0000-0000-000000000001\",\"sets\":3,\"reps\":\"8-10\",\"rest_sec\":90,\"effort_note\":\"RPE 7\",\"notes\":\"Контролируй движение\",\"substitution_exercise_ids\":[]}],\"cooldown\":\"Лёгкая растяжка\"}]}],\"nutrition\":{\"daily_calories\":2200,\"protein_g\":160,\"carbs_g\":220,\"fat_g\":70,\"daily_water_ml\":2600},\"adaptation_rules\":[]}"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gemini-test",
		Client:  server.Client(),
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	plan, err := provider.GeneratePlan(GenerationInput{
		Locale: "ru",
		Profile: UserProfile{
			AvailableTrainingDays: []AvailabilityDay{{Weekday: "monday", DurationMin: 45}},
		},
		Targets: NutritionTarget{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
		Candidates: seedExerciseCatalog(),
	})
	if err != nil {
		t.Fatalf("expected tolerant parsing to succeed, got %v", err)
	}
	if len(plan.Sessions) != 1 || plan.Sessions[0].SessionName != "Верх тела A" {
		t.Fatalf("unexpected plan parsed from string warmup/cooldown: %+v", plan)
	}
}
