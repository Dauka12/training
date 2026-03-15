package app

import (
	"net/http"
	"strings"
	"time"

	emailintegration "training/backend/internal/integrations/email"
)

func (a *App) handleNotifications(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": user.Notifications})
}

func (a *App) handleReadNotification(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for idx := range user.Notifications {
		if user.Notifications[idx].ID == req.ID {
			user.Notifications[idx].Read = true
		}
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (a *App) handleAdminAILogs(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListAILogs(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin ai logs repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []AIGenerationLog
	for _, user := range a.users {
		items = append(items, user.AIGenerationLogs...)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminEmailLogs(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListEmailLogs(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin email logs repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []EmailLog
	for _, user := range a.users {
		items = append(items, user.EmailLogs...)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminAuditLogs(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListAuditLogs(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin audit logs repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"items": a.auditLogs})
}

func (a *App) handleTrainerUsers(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListTrainerUsers(r.Context(), user.Email, hasRole(user, "admin"))
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("trainer users repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []map[string]any
	for _, candidate := range a.users {
		if hasRole(user, "admin") || candidate.AssignedTrainerEmail == user.Email {
			items = append(items, map[string]any{
				"email":       candidate.Email,
				"plan_health": a.planHealthForUserLocked(candidate),
				"workouts":    len(candidate.WorkoutLogs),
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleTrainerUserSubroutes(w http.ResponseWriter, r *http.Request, user *User) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/trainer/users/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}
	target := a.byMail[normalizeEmail(parts[0])]
	if target == nil {
		writeError(w, http.StatusNotFound, "user_not_found", "User not found")
		return
	}
	if !hasRole(user, "admin") && target.AssignedTrainerEmail != user.Email {
		writeError(w, http.StatusForbidden, "forbidden", "Forbidden")
		return
	}
	if len(parts) == 2 && parts[1] == "notes" {
		switch r.Method {
		case http.MethodGet:
			if a.relationalStore != nil {
				items, err := a.relationalStore.ListTrainerNotes(r.Context(), target.Email)
				if err == nil {
					writeJSON(w, http.StatusOK, map[string]any{"items": items})
					return
				}
				a.log.Error("trainer notes repository read failed", "error", err)
			}
			a.mu.Lock()
			items := append([]TrainerNote(nil), a.trainerNotes[target.Email]...)
			a.mu.Unlock()
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		case http.MethodPost:
			var req struct {
				Body string `json:"body"`
			}
			if !decodeJSON(w, r, &req) {
				return
			}
			note := TrainerNote{
				ID:           token(8),
				TrainerEmail: user.Email,
				UserEmail:    target.Email,
				Body:         req.Body,
				CreatedAt:    a.nowRFC3339(),
			}
			a.mu.Lock()
			a.trainerNotes[target.Email] = append(a.trainerNotes[target.Email], note)
			a.pushNotificationLocked(target, "trainer_message", localized(target.Locale, "Новое сообщение от тренера", "Zhattyqtyrushydan zhana habar"), "/trainer")
			a.persistStateLocked()
			a.mu.Unlock()
			writeJSON(w, http.StatusCreated, map[string]any{"note": note})
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		if a.relationalStore != nil {
			detail, err := a.relationalStore.GetTrainerUserDetail(r.Context(), target.Email)
			if err == nil {
				writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
					"email":            detail.Email,
					"plan_versions":    detail.PlanVersions,
					"meal_logs":        detail.MealLogs,
					"water_ml":         detail.WaterML,
					"weekly_checkins":  detail.WeeklyCheckins,
					"assigned_trainer": detail.AssignedTrainer,
				}})
				return
			}
			a.log.Error("trainer user detail repository read failed", "error", err)
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{
			"email":         target.Email,
			"plan_versions": len(target.PlanVersions),
			"meal_logs":     len(target.MealLogs),
			"water_ml":      target.WaterConsumed,
		}})
		return
	}
	if len(parts) == 2 && parts[1] == "regenerate-plan" && r.Method == http.MethodPost {
		a.mu.Lock()
		plan := a.generatePlanLocked(target, "trainer_triggered")
		a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "Not found")
}

func (a *App) handleAdminUsers(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListAdminUsers(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin users repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []map[string]any
	for _, candidate := range a.users {
		activePlans := 0
		if latest := latestPlan(candidate); latest != nil && latest.SupersededAt == "" {
			activePlans = 1
		}
		items = append(items, map[string]any{
			"email":                  candidate.Email,
			"roles":                  candidate.Roles,
			"assigned_trainer_email": candidate.AssignedTrainerEmail,
			"onboarding_done":        candidate.OnboardingDone,
			"active_plan_versions":   activePlans,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminTrainers(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListAdminTrainers(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin trainers repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []map[string]any
	for _, candidate := range a.users {
		if !hasRole(candidate, "trainer") {
			continue
		}
		assignedUsers := 0
		trainerNoteCount := 0
		for _, member := range a.users {
			if member.AssignedTrainerEmail == candidate.Email {
				assignedUsers++
			}
		}
		for _, notes := range a.trainerNotes {
			for _, note := range notes {
				if note.TrainerEmail == candidate.Email {
					trainerNoteCount++
				}
			}
		}
		items = append(items, map[string]any{
			"email":              candidate.Email,
			"assigned_users":     assignedUsers,
			"trainer_note_count": trainerNoteCount,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAssignTrainer(w http.ResponseWriter, r *http.Request, actingUser *User) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		UserEmail    string `json:"user_email"`
		TrainerEmail string `json:"trainer_email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	member := a.byMail[normalizeEmail(req.UserEmail)]
	trainer := a.byMail[normalizeEmail(req.TrainerEmail)]
	if member == nil || trainer == nil || !hasRole(trainer, "trainer", "admin") {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid assignment")
		return
	}
	member.AssignedTrainerEmail = trainer.Email
	a.recordAuditLocked(actingUser.Email, "assign_trainer", "user", member.Email)
	a.persistStateLocked()
	a.pushNotificationLocked(member, "trainer_message", localized(member.Locale, "Назначен тренер", "Zhattyqtyrushy tagaiyndaldy"), "/trainer")
	writeJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (a *App) handleAdminEquipment(w http.ResponseWriter, r *http.Request, user *User) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		defer a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"items": a.equipmentCatalog})
	case http.MethodPost:
		var req EquipmentItem
		if !decodeJSON(w, r, &req) {
			return
		}
		req.ID = token(8)
		req.Active = true
		a.mu.Lock()
		a.equipmentCatalog = append(a.equipmentCatalog, req)
		a.recordAuditLocked(user.Email, "create_equipment", "equipment", req.ID)
		a.persistStateLocked()
		a.mu.Unlock()
		writeJSON(w, http.StatusCreated, map[string]any{"item": req})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminExercises(w http.ResponseWriter, r *http.Request, user *User) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		defer a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"items": a.exerciseCatalog})
	case http.MethodPost:
		var req ExerciseItem
		if !decodeJSON(w, r, &req) {
			return
		}
		req.ID = token(8)
		req.Active = true
		a.mu.Lock()
		a.exerciseCatalog = append(a.exerciseCatalog, req)
		a.recordAuditLocked(user.Email, "create_exercise", "exercise", req.ID)
		a.persistStateLocked()
		a.mu.Unlock()
		writeJSON(w, http.StatusCreated, map[string]any{"item": req})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminNotificationLogs(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListNotificationLogs(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin notification logs repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []map[string]any
	for _, candidate := range a.users {
		for _, notification := range candidate.Notifications {
			items = append(items, map[string]any{
				"id":         notification.ID,
				"user_email": candidate.Email,
				"type":       notification.Type,
				"title":      notification.Title,
				"target_url": notification.TargetURL,
				"read":       notification.Read,
				"created_at": notification.CreatedAt,
			})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminSupportThreads(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListSupportThreads(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin support repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []*SupportThread
	for _, thread := range a.supportThreads {
		items = append(items, thread)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminSupportThreadSubroutes(w http.ResponseWriter, r *http.Request, user *User) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/support/threads/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "status" {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Status          string `json:"status"`
		AssignedToEmail string `json:"assigned_to_email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if !isSupportStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid support status")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	thread := a.supportThreads[parts[0]]
	if thread == nil {
		writeError(w, http.StatusNotFound, "thread_not_found", "Thread not found")
		return
	}
	if req.AssignedToEmail != "" {
		assignee := a.byMail[normalizeEmail(req.AssignedToEmail)]
		if assignee == nil || !hasRole(assignee, "trainer", "admin") {
			writeError(w, http.StatusBadRequest, "validation_error", "Invalid assignee")
			return
		}
		thread.AssignedToEmail = assignee.Email
	}
	thread.Status = req.Status
	a.recordAuditLocked(user.Email, "moderate_support_thread", "support_thread", thread.ID)
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]any{"status": "saved", "thread": thread})
}

func (a *App) handleAdminDiscussionThreads(w http.ResponseWriter, r *http.Request, _ *User) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.relationalStore != nil {
		items, err := a.relationalStore.ListDiscussionThreads(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}
		a.log.Error("admin discussion repository read failed", "error", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var items []*DiscussionThread
	for _, thread := range a.discussionThreads {
		items = append(items, thread)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleAdminDiscussionThreadSubroutes(w http.ResponseWriter, r *http.Request, user *User) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/discussions/threads/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "moderation" {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if !isDiscussionStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid discussion status")
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	thread := a.discussionThreads[parts[0]]
	if thread == nil {
		writeError(w, http.StatusNotFound, "thread_not_found", "Thread not found")
		return
	}
	thread.Status = req.Status
	a.recordAuditLocked(user.Email, "moderate_discussion_thread", "discussion_thread", thread.ID)
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]any{"status": "saved", "thread": thread})
}

func (a *App) handleSupportThreads(w http.ResponseWriter, r *http.Request, user *User) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		defer a.mu.Unlock()
		var items []*SupportThread
		for _, thread := range a.supportThreads {
			if hasRole(user, "admin") || thread.UserEmail == user.Email || thread.AssignedToEmail == user.Email {
				items = append(items, thread)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if !decodeJSON(w, r, &req) {
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		thread := &SupportThread{
			ID:              token(8),
			UserEmail:       user.Email,
			Title:           req.Title,
			Status:          "open",
			AssignedToEmail: user.AssignedTrainerEmail,
			CreatedAt:       a.nowRFC3339(),
			Messages:        []SupportMessage{{AuthorEmail: user.Email, Body: req.Body, CreatedAt: a.nowRFC3339()}},
		}
		a.supportThreads[thread.ID] = thread
		a.persistStateLocked()
		writeJSON(w, http.StatusCreated, map[string]any{"thread": thread})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleSupportThreadSubroutes(w http.ResponseWriter, r *http.Request, user *User) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/support/threads/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "messages" {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	thread := a.supportThreads[parts[0]]
	if thread == nil {
		writeError(w, http.StatusNotFound, "thread_not_found", "Thread not found")
		return
	}
	if !hasRole(user, "admin") && thread.UserEmail != user.Email && thread.AssignedToEmail != user.Email {
		writeError(w, http.StatusForbidden, "forbidden", "Forbidden")
		return
	}
	thread.Messages = append(thread.Messages, SupportMessage{AuthorEmail: user.Email, Body: req.Body, CreatedAt: a.nowRFC3339()})
	if user.Email != thread.UserEmail {
		if member := a.byMail[thread.UserEmail]; member != nil {
			a.pushNotificationLocked(member, "support_reply", localized(member.Locale, "Есть ответ в поддержке", "Qoldauda zhauap bar"), "/support")
		}
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusCreated, map[string]any{"thread": thread})
}

func (a *App) handleDiscussionThreads(w http.ResponseWriter, r *http.Request, user *User) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		defer a.mu.Unlock()
		var items []*DiscussionThread
		for _, thread := range a.discussionThreads {
			if !hasRole(user, "admin") && thread.Status != "visible" {
				continue
			}
			items = append(items, thread)
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	case http.MethodPost:
		var req struct {
			Title    string `json:"title"`
			Body     string `json:"body"`
			Category string `json:"category"`
		}
		if !decodeJSON(w, r, &req) {
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		thread := &DiscussionThread{
			ID:          token(8),
			AuthorEmail: user.Email,
			Title:       req.Title,
			Body:        req.Body,
			Category:    req.Category,
			Status:      "visible",
			CreatedAt:   a.nowRFC3339(),
		}
		a.discussionThreads[thread.ID] = thread
		a.persistStateLocked()
		writeJSON(w, http.StatusCreated, map[string]any{"thread": thread})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleDiscussionThreadSubroutes(w http.ResponseWriter, r *http.Request, user *User) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/discussions/threads/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "replies" {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	thread := a.discussionThreads[parts[0]]
	if thread == nil {
		writeError(w, http.StatusNotFound, "thread_not_found", "Thread not found")
		return
	}
	if !hasRole(user, "admin") && thread.Status != "visible" {
		writeError(w, http.StatusForbidden, "forbidden", "Forbidden")
		return
	}
	thread.Replies = append(thread.Replies, DiscussionReply{AuthorEmail: user.Email, Body: req.Body, CreatedAt: a.nowRFC3339()})
	if user.Email != thread.AuthorEmail {
		if member := a.byMail[thread.AuthorEmail]; member != nil {
			a.pushNotificationLocked(member, "discussion_reply", localized(member.Locale, "Новый ответ в обсуждении", "Talqylauda zhana zhauap"), "/discussions")
		}
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusCreated, map[string]any{"thread": thread})
}

func (a *App) pushNotificationLocked(user *User, kind, title, targetURL string) {
	user.Notifications = append(user.Notifications, Notification{
		ID:        token(6),
		Type:      kind,
		Title:     title,
		CreatedAt: a.nowRFC3339(),
		TargetURL: targetURL,
	})
}

func (a *App) runReminderSweepLocked(now time.Time) {
	for _, user := range a.users {
		if !user.OnboardingDone || !user.NotificationPreferences.HydrationReminder {
			continue
		}
		if user.WaterTargetML <= 0 || user.WaterConsumed >= user.WaterTargetML {
			continue
		}
		if sentRecently(user.LastHydrationReminderAt, now, 3*time.Hour) {
			continue
		}
		a.pushNotificationLocked(user, "hydration_reminder", localized(user.Locale, "Пора выпить воды", "Su ishu uaqyty"), "/track")
		user.LastHydrationReminderAt = now.UTC().Format(time.RFC3339)
		if user.NotificationPreferences.EmailEnabled {
			a.sendEmailLocked(user, "hydration_reminder", localized(user.Locale, "Напоминание о воде", "Su turaly eske salu"), localized(user.Locale, "Добавьте немного воды в трекер.", "Trekerge biraz su engiziniz."))
		}
	}
	a.persistStateLocked()
}

func sentRecently(raw string, now time.Time, window time.Duration) bool {
	if raw == "" {
		return false
	}
	last, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return false
	}
	return now.Sub(last) < window
}

func (a *App) sendEmailLocked(user *User, kind, subject, text string) {
	logEntry := EmailLog{
		ID:        token(6),
		Type:      kind,
		To:        user.Email,
		Subject:   subject,
		CreatedAt: a.nowRFC3339(),
	}
	if a.emailSender == nil {
		logEntry.Status = "skipped"
		user.EmailLogs = append(user.EmailLogs, logEntry)
		return
	}
	if err := a.emailSender.Send(emailintegration.Message{
		To:      user.Email,
		Subject: subject,
		Text:    text,
	}); err != nil {
		logEntry.Status = "failed"
		logEntry.Error = err.Error()
	} else {
		logEntry.Status = "sent"
	}
	user.EmailLogs = append(user.EmailLogs, logEntry)
}

func (a *App) recordAuditLocked(actorEmail, action, resourceType, resourceID string) {
	a.auditLogs = append(a.auditLogs, AuditLog{
		ID:           token(6),
		ActorEmail:   actorEmail,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    a.nowRFC3339(),
	})
}

func isSupportStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "open", "in_progress", "resolved", "closed":
		return true
	default:
		return false
	}
}

func isDiscussionStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "visible", "hidden", "flagged", "archived":
		return true
	default:
		return false
	}
}

func seedEquipmentCatalog() []EquipmentItem {
	return []EquipmentItem{
		{
			ID: "10000000-0000-0000-0000-000000000001",
			Names: map[string]string{
				"ru": "Гантели",
				"kk": "Gantel",
			},
			Descriptions: map[string]string{
				"ru": "Пара свободных весов",
				"kk": "Erkin salmaq jup",
			},
			Category:     "weights",
			LocationType: "mixed",
			MediaURL:     "https://example.com/media/dumbbells.jpg",
			Active:       true,
		},
	}
}

func seedExerciseCatalog() []ExerciseItem {
	return []ExerciseItem{
		{
			ID:   "20000000-0000-0000-0000-000000000001",
			Slug: "goblet-squat",
			Names: map[string]string{
				"ru": "Присед с гантелью",
				"kk": "Gantelmen otyru",
			},
			Descriptions: map[string]string{
				"ru": "Базовое упражнение на ноги",
				"kk": "Aiaq ushin negizgi zhattygu",
			},
			Technique: map[string]string{
				"ru": "Держите спину ровно",
				"kk": "Arqany tike ustanyz",
			},
			Movement:     "squat",
			Difficulty:   "beginner",
			LocationType: "mixed",
			EquipmentIDs: []string{"10000000-0000-0000-0000-000000000001"},
			Active:       true,
		},
		{
			ID:   "20000000-0000-0000-0000-000000000002",
			Slug: "push-up",
			Names: map[string]string{
				"ru": "Отжимания",
				"kk": "Iterilu",
			},
			Descriptions: map[string]string{
				"ru": "Базовое упражнение на верх тела",
				"kk": "Jogargy bolik ushin negizgi zhattygu",
			},
			Technique: map[string]string{
				"ru": "Корпус прямой",
				"kk": "Deneni tike ustanyz",
			},
			Movement:     "push",
			Difficulty:   "beginner",
			LocationType: "home",
			Active:       true,
		},
	}
}
