package calc

import "math"

type Sex string
type Goal string
type ActivityLevel string

const (
	SexMale   Sex = "male"
	SexFemale Sex = "female"

	GoalLoseFat     Goal = "lose_fat"
	GoalMaintain    Goal = "maintain"
	GoalBuildMuscle Goal = "build_muscle"

	ActivitySedentary ActivityLevel = "sedentary"
	ActivityLight     ActivityLevel = "light"
	ActivityModerate  ActivityLevel = "moderate"
	ActivityHigh      ActivityLevel = "high"
)

type Profile struct {
	Age                  int
	Sex                  Sex
	HeightCM             int
	WeightKG             int
	Goal                 Goal
	ActivityLevel        ActivityLevel
	ProgramDurationWeeks int
}

type Targets struct {
	BMR      int
	TDEE     int
	Calories int
	ProteinG int
	FatG     int
	CarbsG   int
	WaterML  int
}

func CalculateTargets(input Profile) Targets {
	profile := normalizeProfile(input)
	bmr := calculateBMR(profile)
	tdee := int(math.Round(float64(bmr) * activityMultiplier(profile.ActivityLevel)))
	calories := adjustCaloriesForGoal(tdee, profile.Goal)
	protein := int(math.Round(float64(profile.WeightKG) * 2.0))
	fat := int(math.Round(float64(calories) * 0.31 / 9.0))
	carbs := int(math.Round(float64(calories-protein*4-fat*9) / 4.0))
	water := profile.WeightKG*35 + activityWaterBonus(profile.ActivityLevel)

	return Targets{
		BMR:      bmr,
		TDEE:     tdee,
		Calories: calories,
		ProteinG: protein,
		FatG:     fat,
		CarbsG:   carbs,
		WaterML:  water,
	}
}

func normalizeProfile(profile Profile) Profile {
	if profile.ProgramDurationWeeks <= 0 {
		profile.ProgramDurationWeeks = 12
	}
	return profile
}

func calculateBMR(profile Profile) int {
	base := 10*float64(profile.WeightKG) + 6.25*float64(profile.HeightCM) - 5*float64(profile.Age)
	if profile.Sex == SexFemale {
		return int(math.Round(base - 161))
	}
	return int(math.Round(base + 5))
}

func activityMultiplier(level ActivityLevel) float64 {
	switch level {
	case ActivitySedentary:
		return 1.2
	case ActivityModerate:
		return 1.55
	case ActivityHigh:
		return 1.725
	default:
		return 1.375
	}
}

func adjustCaloriesForGoal(tdee int, goal Goal) int {
	switch goal {
	case GoalLoseFat:
		return max(tdee-500, 1400)
	case GoalBuildMuscle:
		return tdee + 250
	default:
		return tdee
	}
}

func activityWaterBonus(level ActivityLevel) int {
	switch level {
	case ActivityModerate:
		return 250
	case ActivityHigh:
		return 500
	default:
		return 0
	}
}
