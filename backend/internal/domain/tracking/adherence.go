package tracking

type Summary struct {
	WorkoutAdherence   float64
	MealAdherence      float64
	HydrationAdherence float64
	MissedLast7Days    int
}

type PlanHealth string

const (
	PlanHealthHealthy               PlanHealth = "healthy"
	PlanHealthAttentionNeeded       PlanHealth = "attention_needed"
	PlanHealthAdaptationRecommended PlanHealth = "adaptation_recommended"
)

func ComputePlanHealth(summary Summary) PlanHealth {
	if summary.MissedLast7Days >= 2 || summary.WorkoutAdherence < 0.5 || summary.HydrationAdherence < 0.4 {
		return PlanHealthAdaptationRecommended
	}
	if summary.WorkoutAdherence < 0.75 || summary.MealAdherence < 0.75 || summary.HydrationAdherence < 0.6 {
		return PlanHealthAttentionNeeded
	}
	return PlanHealthHealthy
}
