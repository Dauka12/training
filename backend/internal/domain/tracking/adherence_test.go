package tracking

import "testing"

func TestPlanHealth(t *testing.T) {
	health := ComputePlanHealth(Summary{
		WorkoutAdherence:   0.4,
		MealAdherence:      0.5,
		HydrationAdherence: 0.3,
		MissedLast7Days:    2,
	})
	if health != PlanHealthAdaptationRecommended {
		t.Fatalf("expected adaptation recommended, got %s", health)
	}
}
