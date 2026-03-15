package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type PlanResponse struct {
	SchemaVersion string     `json:"schema_version"`
	Status        string     `json:"status"`
	PlanTitle     string     `json:"plan_title"`
	PlanSummary   string     `json:"plan_summary"`
	Warnings      []string   `json:"warnings"`
	Weeks         []PlanWeek `json:"weeks"`
	Nutrition     Nutrition  `json:"nutrition"`
	Adaptation    []string   `json:"adaptation_rules"`
}

type PlanWeek struct {
	WeekIndex int       `json:"week_index"`
	Days      []PlanDay `json:"days"`
}

type PlanDay struct {
	Weekday              string         `json:"weekday"`
	SessionName          string         `json:"session_name"`
	Goal                 string         `json:"goal"`
	EstimatedDurationMin int            `json:"estimated_duration_min"`
	Warmup               StringList     `json:"warmup"`
	Exercises            []PlanExercise `json:"exercises"`
	Cooldown             StringList     `json:"cooldown"`
}

type PlanExercise struct {
	Order                   int      `json:"order"`
	ExerciseID              string   `json:"exercise_id"`
	Sets                    int      `json:"sets"`
	Reps                    string   `json:"reps"`
	RestSec                 int      `json:"rest_sec"`
	EffortNote              string   `json:"effort_note"`
	Notes                   string   `json:"notes"`
	SubstitutionExerciseIDs []string `json:"substitution_exercise_ids"`
}

type PayloadInput struct {
	UserRef string
	Email   string
	Locale  string
	Targets Targets
}

type Nutrition struct {
	DailyCalories   int           `json:"daily_calories"`
	ProteinG        int           `json:"protein_g"`
	CarbsG          int           `json:"carbs_g"`
	FatG            int           `json:"fat_g"`
	DailyWaterML    int           `json:"daily_water_ml"`
	MealExamples    []MealExample `json:"meal_examples"`
	TrainingDayNote string        `json:"training_day_note"`
	RestDayNote     string        `json:"rest_day_note"`
	HydrationNote   string        `json:"hydration_note"`
}

type MealExample struct {
	Slot     string   `json:"slot"`
	Examples []string `json:"examples"`
}

type Targets struct {
	DailyCalories int `json:"daily_calories"`
	ProteinG      int `json:"protein_g"`
	CarbsG        int `json:"carbs_g"`
	FatG          int `json:"fat_g"`
	DailyWaterML  int `json:"daily_water_ml"`
}

type ValidationInput struct {
	CandidateExercises map[string]struct{}
	Availability       map[string]int
	Targets            Targets
}

type StringList []string

func ValidatePlanResponse(response PlanResponse, input ValidationInput) error {
	if response.SchemaVersion != "1.0" {
		return fmt.Errorf("invalid schema version")
	}
	if response.Status != "ok" {
		return fmt.Errorf("invalid status")
	}
	if strings.TrimSpace(response.PlanTitle) == "" {
		return fmt.Errorf("plan title is required")
	}
	if strings.TrimSpace(response.PlanSummary) == "" {
		return fmt.Errorf("plan summary is required")
	}
	if len(response.Weeks) == 0 {
		return fmt.Errorf("at least one week is required")
	}
	if hasExpectedTargets(input.Targets) && !matchesTargets(response.Nutrition, input.Targets) {
		return fmt.Errorf("nutrition targets mismatch")
	}
	for _, week := range response.Weeks {
		if week.WeekIndex <= 0 {
			return fmt.Errorf("week index must be positive")
		}
		if len(week.Days) == 0 {
			return fmt.Errorf("week must contain at least one day")
		}
		for _, day := range week.Days {
			weekday := strings.ToLower(strings.TrimSpace(day.Weekday))
			if weekday == "" {
				return fmt.Errorf("weekday is required")
			}
			if strings.TrimSpace(day.SessionName) == "" {
				return fmt.Errorf("session name is required")
			}
			if day.EstimatedDurationMin <= 0 {
				return fmt.Errorf("estimated duration must be positive")
			}
			if len(input.Availability) > 0 {
				allowedDuration, ok := input.Availability[weekday]
				if !ok {
					return fmt.Errorf("weekday is outside availability: %s", weekday)
				}
				if allowedDuration > 0 && day.EstimatedDurationMin > allowedDuration {
					return fmt.Errorf("estimated duration exceeds availability for %s", weekday)
				}
			}
			if len(day.Exercises) == 0 {
				return fmt.Errorf("session must contain at least one exercise")
			}
			for _, exercise := range day.Exercises {
				if strings.TrimSpace(exercise.ExerciseID) == "" {
					return fmt.Errorf("exercise_id is required")
				}
				if _, ok := input.CandidateExercises[exercise.ExerciseID]; !ok {
					return fmt.Errorf("unknown exercise_id: %s", exercise.ExerciseID)
				}
				for _, substitutionID := range exercise.SubstitutionExerciseIDs {
					if substitutionID == "" {
						continue
					}
					if _, ok := input.CandidateExercises[substitutionID]; !ok {
						return fmt.Errorf("unknown substitution_exercise_id: %s", substitutionID)
					}
				}
			}
		}
	}
	return nil
}

func hasExpectedTargets(targets Targets) bool {
	return targets.DailyCalories > 0 ||
		targets.ProteinG > 0 ||
		targets.CarbsG > 0 ||
		targets.FatG > 0 ||
		targets.DailyWaterML > 0
}

func matchesTargets(nutrition Nutrition, targets Targets) bool {
	return nutrition.DailyCalories == targets.DailyCalories &&
		nutrition.ProteinG == targets.ProteinG &&
		nutrition.CarbsG == targets.CarbsG &&
		nutrition.FatG == targets.FatG &&
		nutrition.DailyWaterML == targets.DailyWaterML
}

func (s *StringList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}

	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*s = list
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		trimmed := strings.TrimSpace(single)
		if trimmed == "" {
			*s = nil
			return nil
		}
		*s = []string{trimmed}
		return nil
	}

	return fmt.Errorf("expected string or string array")
}

func BuildPrivacySafePayload(input PayloadInput) map[string]any {
	return map[string]any{
		"schema_version": "1.0",
		"user_ref":       input.UserRef,
		"locale":         input.Locale,
		"targets": map[string]any{
			"daily_calories": input.Targets.DailyCalories,
			"protein_g":      input.Targets.ProteinG,
			"carbs_g":        input.Targets.CarbsG,
			"fat_g":          input.Targets.FatG,
			"daily_water_ml": input.Targets.DailyWaterML,
		},
	}
}
