package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"training/backend/internal/domain/ai"
	"training/backend/internal/domain/calc"
	planlogic "training/backend/internal/domain/plans"
	trackinglogic "training/backend/internal/domain/tracking"
)

func (a *App) handleOnboarding(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req UserProfile
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Age <= 0 || req.HeightCM <= 0 || req.CurrentWeightKG <= 0 || req.ProgramDurationWeeks <= 0 {
		writeError(w, http.StatusBadRequest, "validation_error", "Incomplete onboarding profile")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user.Profile = req
	user.OnboardingDone = true
	user.WaterTargetML = calculateNutritionTargets(user).DailyWaterML
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]any{"status": "saved", "water_target_ml": user.WaterTargetML})
}

func staticWeeksFromSessions(input GenerationInput, sessions []ScheduleItem, primaryExercise ExerciseItem) []PlanWeek {
	days := make([]PlanDay, 0, len(sessions))
	for _, session := range sessions {
		days = append(days, PlanDay{
			Weekday:          session.Weekday,
			SessionName:      session.SessionName,
			Goal:             localized(input.Locale, "Спокойная базовая силовая тренировка", "Negizgi kush zhattyguy"),
			EstimatedMinutes: session.EstimatedMinutes,
			Warmup: []string{
				localized(input.Locale, "5 минут легкого кардио", "5 minut zhenil kardio"),
				localized(input.Locale, "Мобилизация суставов", "Bukilerdi zhumyldyrynyz"),
			},
			Exercises: []PlanExercise{
				{
					Order:        1,
					ExerciseID:   primaryExercise.ID,
					ExerciseName: translatedValue(input.Locale, primaryExercise.Names),
					Sets:         3,
					Reps:         "8-10",
					RestSec:      90,
					EffortNote:   "RPE 7",
					Notes:        localized(input.Locale, "Контролируйте технику и темп", "Texnikany baqylanyz"),
				},
			},
			Cooldown: []string{
				localized(input.Locale, "Легкая заминка и дыхание", "Zhenil sozylu zhane demalu"),
			},
		})
	}
	return []PlanWeek{{WeekIndex: 1, Days: days}}
}

func fallbackExercise(locale string, candidates []ExerciseItem) ExerciseItem {
	for _, candidate := range candidates {
		if !candidate.Active {
			continue
		}
		if translatedValue(locale, candidate.Names) != "" {
			return candidate
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ExerciseItem{
		ID:    "fallback-exercise",
		Names: map[string]string{"ru": "Базовое упражнение", "kk": "Negizgi zhattygu"},
	}
}

func mergeNutritionTargets(base, generated NutritionTarget) NutritionTarget {
	return enrichNutritionTargets(base, generated)
}

func enrichNutritionTargets(base, generated NutritionTarget) NutritionTarget {
	result := base
	if note := strings.TrimSpace(generated.TrainingNote); note != "" {
		result.TrainingNote = note
	}
	if note := strings.TrimSpace(generated.RestNote); note != "" {
		result.RestNote = note
	}
	if note := strings.TrimSpace(generated.HydrationNote); note != "" {
		result.HydrationNote = note
	}
	if len(generated.MealExamples) > 0 {
		result.MealExamples = generated.MealExamples
	}
	return result
}

func (a *App) handleGeneratePlan(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		GenerationType string `json:"generation_type"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !user.OnboardingDone {
		writeError(w, http.StatusBadRequest, "onboarding_required", "Complete onboarding first")
		return
	}
	reason := req.GenerationType
	if reason == "" {
		reason = "manual"
	}
	plan := a.generatePlanLocked(user, reason)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "plan": plan})
}

func (a *App) handleActivePlan(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(user.PlanVersions) == 0 {
		writeError(w, http.StatusNotFound, "plan_not_found", "No active plan")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"plan": user.PlanVersions[len(user.PlanVersions)-1]})
}

func (a *App) handleDashboard(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var currentPlan *PlanVersion
	if len(user.PlanVersions) > 0 {
		currentPlan = &user.PlanVersions[len(user.PlanVersions)-1]
	}
	mealStatus := "not_logged"
	if len(user.MealLogs) > 0 {
		mealStatus = user.MealLogs[len(user.MealLogs)-1].Status
	}
	var todayWorkout any
	var nextSession any
	if currentPlan != nil && len(currentPlan.Schedule) > 0 {
		todayWorkout = currentPlan.Schedule[0]
		if len(currentPlan.Schedule) > 1 {
			nextSession = currentPlan.Schedule[1]
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"today_workout":         todayWorkout,
		"meal_status":           mealStatus,
		"hydration":             map[string]any{"target_ml": user.WaterTargetML, "consumed_ml": user.WaterConsumed},
		"quick_actions":         []string{"log_workout", "log_meal", "log_water"},
		"current_week_progress": map[string]int{"completed_sessions": completedWorkoutCount(user.WorkoutLogs)},
		"latest_weight_trend":   latestWeight(user.WeeklyCheckins, user.Profile.CurrentWeightKG),
		"next_session":          nextSession,
		"plan_health":           a.planHealthForUserLocked(user),
		"notifications_unread":  unreadCount(user.Notifications),
	})
}

func (a *App) handleWorkoutLog(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req WorkoutLog
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user.WorkoutLogs = append(user.WorkoutLogs, req)
	if currentPlan := latestPlan(user); currentPlan != nil {
		for idx := range currentPlan.Schedule {
			if currentPlan.Schedule[idx].ID == req.ScheduleID {
				currentPlan.Schedule[idx].Status = req.Status
				if req.Status == "skipped" {
					a.rescheduleSkippedWorkoutLocked(user, currentPlan, currentPlan.Schedule[idx])
				}
				break
			}
		}
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged"})
}

func (a *App) handleMealLog(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user.MealLogs = append(user.MealLogs, MealLog{Status: req.Status, Note: req.Note, CreatedAt: a.nowRFC3339()})
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged"})
}

func (a *App) handleWaterLog(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		AmountML int `json:"amount_ml"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.AmountML <= 0 || req.AmountML > 5000 {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid water amount")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user.WaterConsumed += req.AmountML
	user.WaterLogs = append(user.WaterLogs, WaterLog{AmountML: req.AmountML, CreatedAt: a.nowRFC3339()})
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]any{"status": "logged", "consumed_ml": user.WaterConsumed, "target_ml": user.WaterTargetML})
}

func (a *App) handleHydrationSummary(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	adherence := 0.0
	if user.WaterTargetML > 0 {
		adherence = float64(user.WaterConsumed) / float64(user.WaterTargetML)
	}
	writeJSON(w, http.StatusOK, map[string]any{"target_ml": user.WaterTargetML, "consumed_ml": user.WaterConsumed, "adherence": adherence})
}

func (a *App) handleWeeklyCheckin(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req WeeklyCheckin
	if !decodeJSON(w, r, &req) {
		return
	}
	req.CreatedAt = a.nowRFC3339()
	a.mu.Lock()
	defer a.mu.Unlock()
	user.WeeklyCheckins = append(user.WeeklyCheckins, req)
	hydrationAdherence := 0.0
	if user.WaterTargetML > 0 {
		hydrationAdherence = float64(user.WaterConsumed) / float64(user.WaterTargetML)
	}
	reason, triggered := planlogic.EvaluateRegenerationTrigger(planlogic.RegenerationInput{
		MissedWorkoutsLast7Days: countWorkoutsByStatus(user.WorkoutLogs, "skipped"),
		AvailabilityChanged:     req.AvailabilityChanged,
		EquipmentChanged:        req.EquipmentChanged,
		InjuryChanged:           req.InjuryChanged,
		HydrationAdherence:      hydrationAdherence,
	})
	var plan *PlanVersion
	if triggered {
		created := a.generatePlanLocked(user, reason)
		plan = &created
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "regenerated": triggered, "reason": reason, "plan": plan})
}

func (a *App) generatePlanLocked(user *User, reason string) PlanVersion {
	targets := calculateNutritionTargets(user)
	_ = ai.BuildPrivacySafePayload(ai.PayloadInput{
		UserRef: user.ID,
		Locale:  user.Locale,
		Targets: ai.Targets{
			DailyCalories: targets.DailyCalories,
			ProteinG:      targets.ProteinG,
			CarbsG:        targets.CarbsG,
			FatG:          targets.FatG,
			DailyWaterML:  targets.DailyWaterML,
		},
	})

	input := GenerationInput{
		Locale:            user.Locale,
		UserRef:           user.ID,
		Profile:           user.Profile,
		Targets:           targets,
		Candidates:        a.exerciseCatalog,
		SelectedEquipment: selectedEquipment(a.equipmentCatalog, user.Profile.EquipmentIDs),
		History:           generationHistoryForUser(user, targets, reason, a.now()),
	}

	var (
		generated GeneratedPlan
		err       error
	)
	for attempt := 1; attempt <= 2; attempt++ {
		generated, err = a.aiProvider.GeneratePlan(input)
		if err != nil {
			a.appendAIGenerationLogLocked(user, "provider_error", providerName(a.aiProvider), generated.Title, err.Error(), attempt)
			continue
		}
		if validateErr := validateGeneratedPlan(generated); validateErr != nil {
			err = validateErr
			a.appendAIGenerationLogLocked(user, "invalid_response", providerName(a.aiProvider), generated.Title, validateErr.Error(), attempt)
			continue
		}
		a.appendAIGenerationLogLocked(user, "ok", providerName(a.aiProvider), generated.Title, "", attempt)
		err = nil
		break
	}
	if err != nil {
		a.log.Error("ai generation failed, using static fallback", "error", err)
		generated, _ = StaticAIProvider{}.GeneratePlan(input)
		generated.Warnings = append(generated.Warnings, localized(user.Locale, "Использован резервный шаблон плана", "Qosalqy jospar ulgisi qoldanyldy"))
		a.appendAIGenerationLogLocked(user, "fallback", providerName(StaticAIProvider{}), generated.Title, err.Error(), 0)
	}

	if latest := latestPlan(user); latest != nil {
		latest.SupersededAt = a.nowRFC3339()
	}
	parentID := ""
	if len(user.PlanVersions) > 0 {
		parentID = user.PlanVersions[len(user.PlanVersions)-1].ID
	}
	plan := PlanVersion{
		ID:                 token(10),
		VersionNo:          len(user.PlanVersions) + 1,
		ParentPlanID:       parentID,
		RegenerationReason: reason,
		CreatedAt:          a.nowRFC3339(),
		Title:              generated.Title,
		Summary:            generated.Summary,
		Nutrition:          mergeNutritionTargets(targets, generated.Nutrition),
		Schedule:           generated.Sessions,
		Weeks:              generated.Weeks,
		Warnings:           generated.Warnings,
		AdaptationRules:    generated.AdaptationRules,
	}
	user.PlanVersions = append(user.PlanVersions, plan)
	a.persistStateLocked()
	a.pushNotificationLocked(user, "plan_regenerated", localized(user.Locale, "План обновлен", "Jospar zhanartyldy"), "/plan")
	return plan
}

func (a *App) rescheduleSkippedWorkoutLocked(user *User, plan *PlanVersion, skipped ScheduleItem) {
	var slots []planlogic.AvailabilitySlot
	for _, day := range user.Profile.AvailableTrainingDays {
		slots = append(slots, planlogic.AvailabilitySlot{
			Weekday:     weekdayToTime(day.Weekday),
			DurationMin: day.DurationMin,
		})
	}
	occupied := map[time.Weekday]bool{}
	for _, item := range plan.Schedule {
		occupied[weekdayToTime(item.Weekday)] = true
	}
	next, ok := planlogic.FindNextValidReschedule(a.now(), weekdayToTime(skipped.Weekday), slots, occupied)
	if !ok {
		return
	}
	plan.Schedule = append(plan.Schedule, ScheduleItem{
		ID:                token(8),
		Weekday:           strings.ToLower(next.Weekday().String()),
		SessionName:       skipped.SessionName + " (rescheduled)",
		EstimatedMinutes:  skipped.EstimatedMinutes,
		Status:            "rescheduled",
		RescheduledFromID: skipped.ID,
	})
}

func validateGeneratedPlan(plan GeneratedPlan) error {
	if strings.TrimSpace(plan.Title) == "" {
		return fmt.Errorf("plan title is required")
	}
	if len(plan.Sessions) == 0 {
		return fmt.Errorf("at least one session is required")
	}
	if len(plan.Weeks) == 0 {
		return fmt.Errorf("at least one week is required")
	}
	for _, session := range plan.Sessions {
		if strings.TrimSpace(session.Weekday) == "" {
			return fmt.Errorf("session weekday is required")
		}
		if strings.TrimSpace(session.SessionName) == "" {
			return fmt.Errorf("session name is required")
		}
		if session.EstimatedMinutes <= 0 {
			return fmt.Errorf("session duration must be positive")
		}
	}
	return nil
}

func providerName(provider AIProvider) string {
	return fmt.Sprintf("%T", provider)
}

func (a *App) appendAIGenerationLogLocked(user *User, status, provider, title, errMessage string, attempt int) {
	user.AIGenerationLogs = append(user.AIGenerationLogs, AIGenerationLog{
		ID:          token(6),
		RequestedAt: a.nowRFC3339(),
		Status:      status,
		Provider:    provider,
		PlanTitle:   title,
		Error:       errMessage,
		Attempt:     attempt,
	})
}

func calculateNutritionTargets(user *User) NutritionTarget {
	targets := calc.CalculateTargets(calc.Profile{
		Age:                  user.Profile.Age,
		Sex:                  calc.Sex(user.Profile.BiologicalSex),
		HeightCM:             user.Profile.HeightCM,
		WeightKG:             user.Profile.CurrentWeightKG,
		Goal:                 calc.Goal(user.Profile.PrimaryGoal),
		ActivityLevel:        calc.ActivityLevel(user.Profile.ActivityLevel),
		ProgramDurationWeeks: user.Profile.ProgramDurationWeeks,
	})
	water := targets.WaterML
	if user.WaterOverrideML > 0 {
		water = user.WaterOverrideML
	}
	return NutritionTarget{
		DailyCalories: targets.Calories,
		ProteinG:      targets.ProteinG,
		CarbsG:        targets.CarbsG,
		FatG:          targets.FatG,
		DailyWaterML:  water,
		TrainingNote:  localized(user.Locale, "В тренировочные дни держите углеводы вокруг тренировки", "Zhattygu kuni komirsulardy baqylanyz"),
		RestNote:      localized(user.Locale, "В дни отдыха сохраняйте простой режим", "Demalys kuni qarapaiym rejimdi saqtanyz"),
		HydrationNote: localized(user.Locale, "Пейте воду равномерно в течение дня", "Sudy kun boiy birqalypty ishiniz"),
	}
}

func latestPlan(user *User) *PlanVersion {
	if len(user.PlanVersions) == 0 {
		return nil
	}
	return &user.PlanVersions[len(user.PlanVersions)-1]
}

func generationHistoryForUser(user *User, targets NutritionTarget, reason string, now time.Time) GenerationHistory {
	history := GenerationHistory{
		CompletedSessionsLast14Days:  countRecentWorkoutsByStatus(user.WorkoutLogs, "done", now),
		MissedSessionsLast14Days:     countRecentWorkoutsByStatus(user.WorkoutLogs, "skipped", now),
		MealAdherenceLast14Days:      recentMealAdherence(user.MealLogs, now),
		HydrationAdherenceLast14Days: recentHydrationAdherence(user.WaterLogs, targets.DailyWaterML, now),
		LatestWeightKG:               latestWeight(user.WeeklyCheckins, user.Profile.CurrentWeightKG),
		ReasonForRegeneration:        reason,
	}
	if latest := latestPlan(user); latest != nil {
		history.PreviousPlanSummary = latest.Summary
	}
	return history
}

func selectedEquipment(catalog []EquipmentItem, ids []string) []EquipmentItem {
	allowed := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	items := make([]EquipmentItem, 0, len(ids))
	for _, item := range catalog {
		if _, ok := allowed[item.ID]; ok {
			items = append(items, item)
		}
	}
	return items
}

func latestWeight(checkins []WeeklyCheckin, fallback int) int {
	if len(checkins) == 0 {
		return fallback
	}
	return checkins[len(checkins)-1].WeightKG
}

func countWorkoutsByStatus(logs []WorkoutLog, status string) int {
	total := 0
	for _, item := range logs {
		if item.Status == status {
			total++
		}
	}
	return total
}

func countRecentWorkoutsByStatus(logs []WorkoutLog, status string, now time.Time) int {
	cutoff := now.Add(-14 * 24 * time.Hour)
	total := 0
	for _, item := range logs {
		if item.Status != status {
			continue
		}
		if item.CompletionTime == "" {
			continue
		}
		completedAt, err := time.Parse(time.RFC3339, item.CompletionTime)
		if err != nil || completedAt.Before(cutoff) {
			continue
		}
		total++
	}
	return total
}

func completedWorkoutCount(logs []WorkoutLog) int {
	return countWorkoutsByStatus(logs, "done")
}

func countMealsByStatus(logs []MealLog, status string) int {
	total := 0
	for _, item := range logs {
		if item.Status == status {
			total++
		}
	}
	return total
}

func recentMealAdherence(logs []MealLog, now time.Time) int {
	cutoff := now.Add(-14 * 24 * time.Hour)
	total := 0
	followed := 0
	for _, item := range logs {
		createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil || createdAt.Before(cutoff) {
			continue
		}
		total++
		if item.Status == "followed" {
			followed++
		}
	}
	if total == 0 {
		return 0
	}
	return followed * 100 / total
}

func recentHydrationAdherence(logs []WaterLog, dailyTarget int, now time.Time) int {
	if dailyTarget <= 0 {
		return 0
	}
	cutoff := now.Add(-14 * 24 * time.Hour)
	totalML := 0
	for _, item := range logs {
		createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil || createdAt.Before(cutoff) {
			continue
		}
		totalML += item.AmountML
	}
	if totalML == 0 {
		return 0
	}
	targetWindow := dailyTarget * 14
	adherence := totalML * 100 / targetWindow
	if adherence > 100 {
		return 100
	}
	return adherence
}

func (a *App) planHealthForUserLocked(user *User) trackinglogic.PlanHealth {
	workoutAdherence := 1.0
	if len(user.WorkoutLogs) > 0 {
		workoutAdherence = float64(completedWorkoutCount(user.WorkoutLogs)) / float64(len(user.WorkoutLogs))
	}
	mealAdherence := 1.0
	if len(user.MealLogs) > 0 {
		mealAdherence = float64(countMealsByStatus(user.MealLogs, "followed")) / float64(len(user.MealLogs))
	}
	hydrationAdherence := 0.0
	if user.WaterTargetML > 0 {
		hydrationAdherence = float64(user.WaterConsumed) / float64(user.WaterTargetML)
	}
	return trackinglogic.ComputePlanHealth(trackinglogic.Summary{
		WorkoutAdherence:   workoutAdherence,
		MealAdherence:      mealAdherence,
		HydrationAdherence: hydrationAdherence,
		MissedLast7Days:    countWorkoutsByStatus(user.WorkoutLogs, "skipped"),
	})
}

func weekdayToTime(weekday string) time.Weekday {
	switch strings.ToLower(weekday) {
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Sunday
	}
}

type StaticAIProvider struct{}

func (StaticAIProvider) GeneratePlan(input GenerationInput) (GeneratedPlan, error) {
	return buildStaticPlan(input), nil
}

func buildStaticPlan(input GenerationInput) GeneratedPlan {
	sessions := make([]ScheduleItem, 0, len(input.Profile.AvailableTrainingDays))
	weeks := []PlanWeek{{WeekIndex: 1}}
	primaryExercise := fallbackExercise(input.Locale, input.Candidates)
	for index, day := range input.Profile.AvailableTrainingDays {
		sessions = append(sessions, ScheduleItem{
			ID:               token(8),
			Weekday:          day.Weekday,
			SessionName:      localized(input.Locale, "Силовая сессия", "Kush zhattyguy"),
			EstimatedMinutes: day.DurationMin,
			Status:           "planned",
		})
		if index >= 2 {
			break
		}
	}
	if len(sessions) == 0 {
		sessions = append(sessions, ScheduleItem{
			ID:               token(8),
			Weekday:          "monday",
			SessionName:      localized(input.Locale, "Силовая сессия", "Kush zhattyguy"),
			EstimatedMinutes: 45,
			Status:           "planned",
		})
	}
	weeks = staticWeeksFromSessions(input, sessions, primaryExercise)
	nutrition := enrichNutritionTargets(input.Targets, NutritionTarget{
		TrainingNote:  localized(input.Locale, "В тренировочные дни держите углеводы рядом с тренировкой", "Zhattygu kuni komirsulardy zhattygu ainalasynda ustanyz"),
		RestNote:      localized(input.Locale, "В дни отдыха сохраняйте простой режим питания", "Demalys kuni qarapaiym tamaq rejimin saqtanyz"),
		HydrationNote: localized(input.Locale, "Пейте воду небольшими порциями в течение дня", "Sudy kun boiy azdap ishiniz"),
		MealExamples: []MealExample{
			{Slot: "breakfast", Examples: []string{localized(input.Locale, "Овсянка + яйца", "Sulii botqa + zhumyrtqa")}},
			{Slot: "lunch", Examples: []string{localized(input.Locale, "Рис + курица + овощи", "Kurish + tauyq + konis")}},
			{Slot: "dinner", Examples: []string{localized(input.Locale, "Рыба + картофель + салат", "Balyq + kartop + salat")}},
		},
	})
	return GeneratedPlan{
		Title:     localized(input.Locale, "План на 12 недель", "12 aptalyq jospar"),
		Summary:   localized(input.Locale, "Базовый адаптивный план", "Negizgi beyimdelgen jospar"),
		Warnings:  []string{localized(input.Locale, "При боли снизьте нагрузку", "Auruy bolsa zhuktamany azaytynyz")},
		Nutrition: nutrition,
		Sessions:  sessions,
		Weeks:     weeks,
		AdaptationRules: []string{
			localized(input.Locale, "Если тренировка пропущена, перенесите ее на следующий доступный день", "Zhattygu otkizilip alinsa, ony kelesi qolai kunge auystyrynyz"),
			localized(input.Locale, "Если пропущены две тренировки за неделю, запросите адаптацию плана", "Bir aptada eki zhattygu otip ketse, jospar beyimdeuin suranyz"),
		},
	}
}
