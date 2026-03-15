package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	aidomain "training/backend/internal/domain/ai"
)

type OpenAICompatibleProvider struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
	Logger  *slog.Logger
}

func (p OpenAICompatibleProvider) GeneratePlan(input GenerationInput) (GeneratedPlan, error) {
	if p.BaseURL == "" || p.APIKey == "" || p.Model == "" {
		return GeneratedPlan{}, fmt.Errorf("ai provider is not configured")
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	logger := p.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	requestPayload := map[string]any{
		"model": p.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt(),
			},
			{
				"role":    "user",
				"content": buildUserPrompt(input),
			},
		},
		"temperature": 0.2,
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return GeneratedPlan{}, err
	}
	endpoint := strings.TrimRight(p.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GeneratedPlan{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("ai request failed", "error", err)
		return GeneratedPlan{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(resp.Body)
		logger.Error("ai request rejected", "status", resp.StatusCode)
		return GeneratedPlan{}, fmt.Errorf("ai provider returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return GeneratedPlan{}, err
	}
	if len(envelope.Choices) == 0 || strings.TrimSpace(envelope.Choices[0].Message.Content) == "" {
		return GeneratedPlan{}, fmt.Errorf("ai provider returned empty content")
	}

	return parseStructuredPlan(envelope.Choices[0].Message.Content, input)
}

func parseStructuredPlan(content string, input GenerationInput) (GeneratedPlan, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed aidomain.PlanResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return GeneratedPlan{}, err
	}

	if err := aidomain.ValidatePlanResponse(parsed, aidomain.ValidationInput{
		CandidateExercises: candidateExerciseSet(input.Candidates),
		Availability:       availabilityByWeekday(input.Profile.AvailableTrainingDays),
		Targets: aidomain.Targets{
			DailyCalories: input.Targets.DailyCalories,
			ProteinG:      input.Targets.ProteinG,
			CarbsG:        input.Targets.CarbsG,
			FatG:          input.Targets.FatG,
			DailyWaterML:  input.Targets.DailyWaterML,
		},
	}); err != nil {
		return GeneratedPlan{}, err
	}

	sessions := make([]ScheduleItem, 0)
	for _, week := range parsed.Weeks {
		for _, day := range week.Days {
			sessions = append(sessions, ScheduleItem{
				ID:               token(8),
				Weekday:          strings.ToLower(strings.TrimSpace(day.Weekday)),
				SessionName:      strings.TrimSpace(day.SessionName),
				EstimatedMinutes: day.EstimatedDurationMin,
				Status:           "planned",
			})
		}
	}
	if len(sessions) == 0 {
		return GeneratedPlan{}, fmt.Errorf("ai provider returned no schedule days")
	}

	return GeneratedPlan{
		Title:           parsed.PlanTitle,
		Summary:         parsed.PlanSummary,
		Warnings:        parsed.Warnings,
		Nutrition:       mapNutrition(parsed.Nutrition),
		Sessions:        sessions,
		Weeks:           mapWeeks(input.Locale, parsed.Weeks, input.Candidates),
		AdaptationRules: parsed.Adaptation,
	}, nil
}

func systemPrompt() string {
	return "Return JSON only with no markdown. Use schema_version \"1.0\" and status \"ok\". Use the requested locale for all user-facing text. Never recalculate nutrition targets. Keep nutrition values exactly equal to the provided backend targets. Use only exercise_id values from candidate_exercises. Output keys: schema_version, status, plan_title, plan_summary, warnings, weeks, nutrition, adaptation_rules. Each week day must include weekday, session_name, goal, estimated_duration_min, warmup, exercises, cooldown. Do not exceed available durations or use weekdays outside availability."
}

func buildUserPrompt(input GenerationInput) string {
	payload := map[string]any{
		"schema_version":  "1.0",
		"locale":          input.Locale,
		"generation_type": generationTypeForPrompt(input.History),
		"user_ref":        input.UserRef,
		"profile": map[string]any{
			"age":                      input.Profile.Age,
			"sex":                      input.Profile.BiologicalSex,
			"height_cm":                input.Profile.HeightCM,
			"weight_kg":                input.Profile.CurrentWeightKG,
			"target_weight_kg":         input.Profile.TargetWeightKG,
			"goal":                     input.Profile.PrimaryGoal,
			"experience_level":         input.Profile.ExperienceLevel,
			"activity_level":           input.Profile.ActivityLevel,
			"training_location":        input.Profile.TrainingLocation,
			"injuries":                 input.Profile.Injuries,
			"dietary_preferences":      input.Profile.DietaryPreferences,
			"avoid_foods":              input.Profile.AvoidFoods,
			"timezone":                 input.Profile.Timezone,
			"preferred_training_style": input.Profile.PreferredTrainingStyle,
			"preferred_meal_style":     input.Profile.PreferredMealStyle,
			"hydration_preference":     input.Profile.HydrationPreference,
		},
		"targets": map[string]any{
			"daily_calories": input.Targets.DailyCalories,
			"protein_g":      input.Targets.ProteinG,
			"carbs_g":        input.Targets.CarbsG,
			"fat_g":          input.Targets.FatG,
			"daily_water_ml": input.Targets.DailyWaterML,
		},
		"availability":        input.Profile.AvailableTrainingDays,
		"selected_equipment":  promptEquipment(input.Locale, input.SelectedEquipment),
		"candidate_exercises": promptExercises(input.Locale, input.Candidates),
		"history": map[string]any{
			"previous_plan_summary":            input.History.PreviousPlanSummary,
			"completed_sessions_last_14_days":  input.History.CompletedSessionsLast14Days,
			"missed_sessions_last_14_days":     input.History.MissedSessionsLast14Days,
			"meal_adherence_last_14_days":      input.History.MealAdherenceLast14Days,
			"hydration_adherence_last_14_days": input.History.HydrationAdherenceLast14Days,
			"latest_weight_kg":                 input.History.LatestWeightKG,
			"reason_for_regeneration":          input.History.ReasonForRegeneration,
		},
	}
	raw, _ := json.Marshal(payload)
	return string(raw)
}

func promptEquipment(locale string, items []EquipmentItem) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"equipment_id":  item.ID,
			"name":          translatedValue(locale, item.Names),
			"category":      item.Category,
			"location_type": item.LocationType,
		})
	}
	return result
}

func promptExercises(locale string, items []ExerciseItem) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"exercise_id":               item.ID,
			"name":                      translatedValue(locale, item.Names),
			"movement_pattern":          item.Movement,
			"difficulty":                item.Difficulty,
			"location_type":             item.LocationType,
			"required_equipment_ids":    item.EquipmentIDs,
			"contraindication_tags":     []string{},
			"substitution_exercise_ids": substitutionIDs(items, item.ID),
		})
	}
	return result
}

func substitutionIDs(items []ExerciseItem, currentID string) []string {
	result := make([]string, 0, 2)
	for _, item := range items {
		if item.ID == currentID {
			continue
		}
		result = append(result, item.ID)
		if len(result) == 2 {
			break
		}
	}
	return result
}

func translatedValue(locale string, values map[string]string) string {
	normalizedLocale := normalizeLocale(locale)
	if value := strings.TrimSpace(values[normalizedLocale]); value != "" {
		return value
	}
	if value := strings.TrimSpace(values["ru"]); value != "" {
		return value
	}
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func generationTypeForPrompt(history GenerationHistory) string {
	if strings.TrimSpace(history.ReasonForRegeneration) == "" || history.ReasonForRegeneration == "initial" {
		return "initial"
	}
	return "adaptation"
}

func availabilityByWeekday(days []AvailabilityDay) map[string]int {
	result := make(map[string]int, len(days))
	for _, day := range days {
		result[strings.ToLower(strings.TrimSpace(day.Weekday))] = day.DurationMin
	}
	return result
}

func candidateExerciseSet(items []ExerciseItem) map[string]struct{} {
	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		result[item.ID] = struct{}{}
	}
	return result
}

func mapNutrition(nutrition aidomain.Nutrition) NutritionTarget {
	mealExamples := make([]MealExample, 0, len(nutrition.MealExamples))
	for _, item := range nutrition.MealExamples {
		mealExamples = append(mealExamples, MealExample{
			Slot:     strings.TrimSpace(item.Slot),
			Examples: trimStrings(item.Examples),
		})
	}
	return NutritionTarget{
		DailyCalories: nutrition.DailyCalories,
		ProteinG:      nutrition.ProteinG,
		CarbsG:        nutrition.CarbsG,
		FatG:          nutrition.FatG,
		DailyWaterML:  nutrition.DailyWaterML,
		TrainingNote:  strings.TrimSpace(nutrition.TrainingDayNote),
		RestNote:      strings.TrimSpace(nutrition.RestDayNote),
		HydrationNote: strings.TrimSpace(nutrition.HydrationNote),
		MealExamples:  mealExamples,
	}
}

func mapWeeks(locale string, weeks []aidomain.PlanWeek, candidates []ExerciseItem) []PlanWeek {
	exerciseNames := make(map[string]string, len(candidates))
	for _, item := range candidates {
		exerciseNames[item.ID] = translatedValue(locale, item.Names)
	}

	result := make([]PlanWeek, 0, len(weeks))
	for _, week := range weeks {
		mappedDays := make([]PlanDay, 0, len(week.Days))
		for _, day := range week.Days {
			mappedExercises := make([]PlanExercise, 0, len(day.Exercises))
			for _, exercise := range day.Exercises {
				mappedExercises = append(mappedExercises, PlanExercise{
					Order:                   exercise.Order,
					ExerciseID:              strings.TrimSpace(exercise.ExerciseID),
					ExerciseName:            strings.TrimSpace(exerciseNames[exercise.ExerciseID]),
					Sets:                    exercise.Sets,
					Reps:                    strings.TrimSpace(exercise.Reps),
					RestSec:                 exercise.RestSec,
					EffortNote:              strings.TrimSpace(exercise.EffortNote),
					Notes:                   strings.TrimSpace(exercise.Notes),
					SubstitutionExerciseIDs: trimStrings(exercise.SubstitutionExerciseIDs),
					SubstitutionNames:       lookupExerciseNames(exerciseNames, exercise.SubstitutionExerciseIDs),
				})
			}

			mappedDays = append(mappedDays, PlanDay{
				Weekday:          strings.ToLower(strings.TrimSpace(day.Weekday)),
				SessionName:      strings.TrimSpace(day.SessionName),
				Goal:             strings.TrimSpace(day.Goal),
				EstimatedMinutes: day.EstimatedDurationMin,
				Warmup:           trimStrings(day.Warmup),
				Exercises:        mappedExercises,
				Cooldown:         trimStrings(day.Cooldown),
			})
		}

		result = append(result, PlanWeek{
			WeekIndex: week.WeekIndex,
			Days:      mappedDays,
		})
	}
	return result
}

func lookupExerciseNames(names map[string]string, ids []string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		name := strings.TrimSpace(names[id])
		if name == "" {
			continue
		}
		result = append(result, name)
	}
	return result
}

func trimStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
