package plans

import "testing"

func TestEvaluateRegenerationTrigger(t *testing.T) {
	reason, triggered := EvaluateRegenerationTrigger(RegenerationInput{
		MissedWorkoutsLast7Days: 2,
		AvailabilityChanged:     false,
		EquipmentChanged:        false,
		InjuryChanged:           false,
		HydrationAdherence:      0.8,
	})
	if !triggered {
		t.Fatal("expected trigger")
	}
	if reason != "missed_workouts" {
		t.Fatalf("expected missed_workouts, got %s", reason)
	}
}
