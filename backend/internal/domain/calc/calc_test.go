package calc

import "testing"

func TestCalculateTargets(t *testing.T) {
	profile := Profile{
		Age:                  28,
		Sex:                  SexMale,
		HeightCM:             180,
		WeightKG:             86,
		Goal:                 GoalLoseFat,
		ActivityLevel:        ActivityLight,
		ProgramDurationWeeks: 12,
	}

	targets := CalculateTargets(profile)

	if targets.BMR != 1850 {
		t.Fatalf("expected BMR 1850, got %d", targets.BMR)
	}
	if targets.TDEE != 2544 {
		t.Fatalf("expected TDEE 2544, got %d", targets.TDEE)
	}
	if targets.Calories != 2044 {
		t.Fatalf("expected calories 2044, got %d", targets.Calories)
	}
	if targets.ProteinG != 172 {
		t.Fatalf("expected protein 172, got %d", targets.ProteinG)
	}
	if targets.FatG != 70 {
		t.Fatalf("expected fat 70, got %d", targets.FatG)
	}
	if targets.CarbsG != 182 {
		t.Fatalf("expected carbs 182, got %d", targets.CarbsG)
	}
	if targets.WaterML != 3010 {
		t.Fatalf("expected water 3010, got %d", targets.WaterML)
	}
}
