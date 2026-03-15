package ai

import "testing"

func TestValidatePlanResponseAcceptsRichContract(t *testing.T) {
	response := validResponse()

	err := ValidatePlanResponse(response, ValidationInput{
		CandidateExercises: map[string]struct{}{
			"exercise-1": {},
			"exercise-2": {},
		},
		Availability: map[string]int{
			"monday": 45,
		},
		Targets: Targets{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
	})
	if err != nil {
		t.Fatalf("expected valid response, got %v", err)
	}
}

func TestValidatePlanResponseRejectsUnknownExercise(t *testing.T) {
	response := validResponse()
	response.Weeks[0].Days[0].Exercises[0].ExerciseID = "unknown"

	err := ValidatePlanResponse(response, ValidationInput{
		CandidateExercises: map[string]struct{}{"exercise-1": {}},
		Availability:       map[string]int{"monday": 45},
		Targets: Targets{
			DailyCalories: response.Nutrition.DailyCalories,
			ProteinG:      response.Nutrition.ProteinG,
			CarbsG:        response.Nutrition.CarbsG,
			FatG:          response.Nutrition.FatG,
			DailyWaterML:  response.Nutrition.DailyWaterML,
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePlanResponseRejectsDurationOverAvailability(t *testing.T) {
	response := validResponse()
	response.Weeks[0].Days[0].EstimatedDurationMin = 60

	err := ValidatePlanResponse(response, ValidationInput{
		CandidateExercises: map[string]struct{}{
			"exercise-1": {},
			"exercise-2": {},
		},
		Availability: map[string]int{"monday": 45},
		Targets: Targets{
			DailyCalories: response.Nutrition.DailyCalories,
			ProteinG:      response.Nutrition.ProteinG,
			CarbsG:        response.Nutrition.CarbsG,
			FatG:          response.Nutrition.FatG,
			DailyWaterML:  response.Nutrition.DailyWaterML,
		},
	})
	if err == nil {
		t.Fatal("expected duration validation error")
	}
}

func TestValidatePlanResponseRejectsNutritionMismatch(t *testing.T) {
	response := validResponse()
	response.Nutrition.DailyCalories = 2000

	err := ValidatePlanResponse(response, ValidationInput{
		CandidateExercises: map[string]struct{}{
			"exercise-1": {},
			"exercise-2": {},
		},
		Availability: map[string]int{"monday": 45},
		Targets: Targets{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
	})
	if err == nil {
		t.Fatal("expected nutrition validation error")
	}
}

func TestStringListUnmarshalAcceptsSingleString(t *testing.T) {
	var list StringList
	if err := list.UnmarshalJSON([]byte(`"5 минут лёгкого кардио"`)); err != nil {
		t.Fatalf("expected single string to be accepted, got %v", err)
	}
	if len(list) != 1 || list[0] != "5 минут лёгкого кардио" {
		t.Fatalf("unexpected normalized list: %#v", list)
	}
}

func validResponse() PlanResponse {
	return PlanResponse{
		SchemaVersion: "1.0",
		Status:        "ok",
		PlanTitle:     "План на 12 недель",
		PlanSummary:   "Краткое описание плана",
		Warnings:      []string{"Берегите колено"},
		Nutrition: Nutrition{
			DailyCalories:   2200,
			ProteinG:        160,
			CarbsG:          220,
			FatG:            70,
			DailyWaterML:    2600,
			TrainingDayNote: "Больше углеводов рядом с тренировкой",
			RestDayNote:     "Держите простой режим питания",
			HydrationNote:   "Пейте воду равномерно",
			MealExamples: []MealExample{
				{Slot: "breakfast", Examples: []string{"Овсянка + яйца"}},
			},
		},
		Adaptation: []string{"Если тренировка пропущена, перенести на следующий доступный день"},
		Weeks: []PlanWeek{
			{
				WeekIndex: 1,
				Days: []PlanDay{
					{
						Weekday:              "monday",
						SessionName:          "Верх тела A",
						Goal:                 "Силовая тренировка верхней части тела",
						EstimatedDurationMin: 45,
						Warmup:               []string{"5 минут лёгкого кардио"},
						Exercises: []PlanExercise{
							{
								Order:                   1,
								ExerciseID:              "exercise-1",
								Sets:                    3,
								Reps:                    "8-10",
								RestSec:                 90,
								EffortNote:              "RPE 7",
								Notes:                   "Контролируй движение",
								SubstitutionExerciseIDs: []string{"exercise-2"},
							},
						},
						Cooldown: []string{"Лёгкая растяжка"},
					},
				},
			},
		},
	}
}
