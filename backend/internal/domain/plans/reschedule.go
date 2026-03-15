package plans

import "time"

type AvailabilitySlot struct {
	Weekday     time.Weekday
	DurationMin int
}

func FindNextValidReschedule(now time.Time, _ time.Weekday, slots []AvailabilitySlot, occupied map[time.Weekday]bool) (time.Time, bool) {
	for offset := 1; offset <= 14; offset++ {
		candidate := now.AddDate(0, 0, offset)
		for _, slot := range slots {
			if candidate.Weekday() == slot.Weekday && !occupied[candidate.Weekday()] {
				return candidate, true
			}
		}
	}
	return time.Time{}, false
}

type RegenerationInput struct {
	MissedWorkoutsLast7Days int
	AvailabilityChanged     bool
	EquipmentChanged        bool
	InjuryChanged           bool
	HydrationAdherence      float64
}

func EvaluateRegenerationTrigger(input RegenerationInput) (string, bool) {
	switch {
	case input.MissedWorkoutsLast7Days >= 2:
		return "missed_workouts", true
	case input.AvailabilityChanged:
		return "availability_changed", true
	case input.EquipmentChanged:
		return "equipment_changed", true
	case input.InjuryChanged:
		return "injury_changed", true
	case input.HydrationAdherence > 0 && input.HydrationAdherence < 0.3:
		return "hydration_decline", true
	default:
		return "", false
	}
}
