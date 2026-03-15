package app

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type RelationalStore interface {
	Project(context.Context, runtimeState) error
	ListAdminUsers(context.Context) ([]AdminUserRecord, error)
	ListAdminTrainers(context.Context) ([]AdminTrainerRecord, error)
	ListTrainerUsers(context.Context, string, bool) ([]TrainerUserRecord, error)
	GetTrainerUserDetail(context.Context, string) (TrainerUserDetailRecord, error)
	ListTrainerNotes(context.Context, string) ([]TrainerNote, error)
	ListSupportThreads(context.Context) ([]SupportThreadRecord, error)
	ListDiscussionThreads(context.Context) ([]DiscussionThreadRecord, error)
	ListNotificationLogs(context.Context) ([]NotificationLogRecord, error)
	ListAILogs(context.Context) ([]AIGenerationLog, error)
	ListEmailLogs(context.Context) ([]EmailLog, error)
	ListAuditLogs(context.Context) ([]AuditLog, error)
}

type AdminUserRecord struct {
	Email              string
	Roles              []string
	AssignedTrainer    string
	OnboardingDone     bool
	ActivePlanVersions int
}

type AdminTrainerRecord struct {
	Email            string
	AssignedUsers    int
	TrainerNoteCount int
}

type TrainerUserRecord struct {
	Email      string
	PlanHealth string
	Workouts   int
}

type TrainerUserDetailRecord struct {
	Email           string
	PlanVersions    int
	MealLogs        int
	WaterML         int
	WeeklyCheckins  int
	AssignedTrainer string
}

type SupportThreadRecord struct {
	ID              string
	UserEmail       string
	Status          string
	AssignedToEmail string
	Title           string
	MessageCount    int
	CreatedAt       string
}

type DiscussionThreadRecord struct {
	ID          string
	AuthorEmail string
	Category    string
	Title       string
	Status      string
	ReplyCount  int
	CreatedAt   string
}

type NotificationLogRecord struct {
	ID        string
	UserEmail string
	Type      string
	Title     string
	TargetURL string
	Read      bool
	CreatedAt string
}

type projectionRefs struct {
	userIDs      map[string]string
	userByEmail  map[string]string
	roleIDs      map[string]string
	planIDs      map[string]string
	scheduleIDs  map[string]string
	schedulePlan map[string]string
}

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func (s *PGStateStore) Project(ctx context.Context, state runtimeState) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin projection tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		TRUNCATE TABLE
			admin_audit_logs,
			trainer_notes,
			public_discussion_replies,
			public_discussion_threads,
			support_messages,
			support_threads,
			email_logs,
			ai_generation_logs,
			notification_preferences,
			notifications,
			weekly_checkins,
			hydration_targets,
			water_logs,
			meal_logs,
			workout_logs,
			workout_schedule_instances,
			nutrition_plan_versions,
			workout_day_exercises,
			workout_day_templates,
			generated_plan_versions,
			exercise_media,
			availability_rules,
			user_equipment,
			dietary_preferences,
			user_injuries,
			user_goals,
			trainer_assignments,
			user_profiles,
			password_reset_tokens,
			email_verification_tokens,
			sessions,
			user_roles,
			roles,
			users,
			exercise_catalog_translations,
			exercise_catalog,
			equipment_catalog_translations,
			equipment_catalog
	`); err != nil {
		return fmt.Errorf("truncate relational tables: %w", err)
	}

	refs := newProjectionRefs(state)
	if err := projectCatalogs(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := projectUsers(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := projectPlans(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := projectTracking(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := projectNotifications(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := projectOps(ctx, tx, state, refs); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit projection tx: %w", err)
	}
	return nil
}

func (s *PGStateStore) ListAdminUsers(ctx context.Context) ([]AdminUserRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			u.email,
			COALESCE(array_agg(r.code ORDER BY r.code) FILTER (WHERE r.code IS NOT NULL), ARRAY[]::text[]),
			COALESCE(trainer.email, ''),
			CASE WHEN up.user_id IS NULL THEN FALSE ELSE TRUE END,
			COALESCE(active_plans.active_count, 0)
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id
		LEFT JOIN user_profiles up ON up.user_id = u.id
		LEFT JOIN trainer_assignments ta ON ta.user_id = u.id
		LEFT JOIN users trainer ON trainer.id = ta.trainer_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS active_count
			FROM generated_plan_versions
			WHERE status = 'active'
			GROUP BY user_id
		) active_plans ON active_plans.user_id = u.id
		GROUP BY u.id, u.email, trainer.email, up.user_id, active_plans.active_count
		ORDER BY u.email
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminUserRecord
	for rows.Next() {
		var item AdminUserRecord
		if err := rows.Scan(&item.Email, &item.Roles, &item.AssignedTrainer, &item.OnboardingDone, &item.ActivePlanVersions); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListAdminTrainers(ctx context.Context) ([]AdminTrainerRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			u.email,
			COUNT(DISTINCT ta.user_id) AS assigned_users,
			COUNT(DISTINCT tn.id) AS trainer_note_count
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id AND r.code = 'trainer'
		LEFT JOIN trainer_assignments ta ON ta.trainer_id = u.id
		LEFT JOIN trainer_notes tn ON tn.trainer_id = u.id
		GROUP BY u.id, u.email
		ORDER BY u.email
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminTrainerRecord
	for rows.Next() {
		var item AdminTrainerRecord
		if err := rows.Scan(&item.Email, &item.AssignedUsers, &item.TrainerNoteCount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListTrainerUsers(ctx context.Context, actorEmail string, isAdmin bool) ([]TrainerUserRecord, error) {
	query := `
		SELECT
			u.email,
			COALESCE(COUNT(wl.id), 0) AS workouts,
			COALESCE(SUM(CASE WHEN wl.status = 'skipped' THEN 1 ELSE 0 END), 0) AS skipped_workouts
		FROM users u
		LEFT JOIN generated_plan_versions gpv ON gpv.user_id = u.id
		LEFT JOIN workout_logs wl ON wl.plan_version_id = gpv.id
	`
	args := []any{}
	if !isAdmin {
		query += `
			JOIN trainer_assignments ta ON ta.user_id = u.id
			JOIN users trainer ON trainer.id = ta.trainer_id AND trainer.email = $1
		`
		args = append(args, actorEmail)
	}
	query += `
		GROUP BY u.id, u.email
		ORDER BY u.email
	`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TrainerUserRecord
	for rows.Next() {
		var item TrainerUserRecord
		var skipped int
		if err := rows.Scan(&item.Email, &item.Workouts, &skipped); err != nil {
			return nil, err
		}
		item.PlanHealth = projectionPlanHealth(item.Workouts, skipped)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) GetTrainerUserDetail(ctx context.Context, targetEmail string) (TrainerUserDetailRecord, error) {
	var item TrainerUserDetailRecord
	err := s.pool.QueryRow(ctx, `
		SELECT
			u.email,
			COALESCE((SELECT COUNT(*) FROM generated_plan_versions gpv WHERE gpv.user_id = u.id), 0),
			COALESCE((SELECT COUNT(*) FROM meal_logs ml WHERE ml.user_id = u.id), 0),
			COALESCE((SELECT SUM(wl.amount_ml) FROM water_logs wl WHERE wl.user_id = u.id), 0),
			COALESCE((SELECT COUNT(*) FROM weekly_checkins wc WHERE wc.user_id = u.id), 0),
			COALESCE(trainer.email, '')
		FROM users u
		LEFT JOIN trainer_assignments ta ON ta.user_id = u.id
		LEFT JOIN users trainer ON trainer.id = ta.trainer_id
		WHERE u.email = $1
	`, targetEmail).Scan(
		&item.Email,
		&item.PlanVersions,
		&item.MealLogs,
		&item.WaterML,
		&item.WeeklyCheckins,
		&item.AssignedTrainer,
	)
	return item, err
}

func (s *PGStateStore) ListTrainerNotes(ctx context.Context, targetEmail string) ([]TrainerNote, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			tn.id::text,
			trainer.email,
			member.email,
			tn.body,
			tn.created_at::text
		FROM trainer_notes tn
		JOIN users trainer ON trainer.id = tn.trainer_id
		JOIN users member ON member.id = tn.user_id
		WHERE member.email = $1
		ORDER BY tn.created_at DESC
	`, targetEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TrainerNote
	for rows.Next() {
		var item TrainerNote
		if err := rows.Scan(&item.ID, &item.TrainerEmail, &item.UserEmail, &item.Body, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListSupportThreads(ctx context.Context) ([]SupportThreadRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			st.id::text,
			member.email,
			st.status,
			COALESCE(assignee.email, ''),
			st.title,
			COUNT(sm.id) AS message_count,
			st.created_at::text
		FROM support_threads st
		JOIN users member ON member.id = st.user_id
		LEFT JOIN users assignee ON assignee.id = st.assigned_to
		LEFT JOIN support_messages sm ON sm.thread_id = st.id
		GROUP BY st.id, member.email, st.status, assignee.email, st.title, st.created_at
		ORDER BY st.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SupportThreadRecord
	for rows.Next() {
		var item SupportThreadRecord
		if err := rows.Scan(&item.ID, &item.UserEmail, &item.Status, &item.AssignedToEmail, &item.Title, &item.MessageCount, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListDiscussionThreads(ctx context.Context) ([]DiscussionThreadRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			dt.id::text,
			author.email,
			dt.category,
			dt.title,
			dt.moderation_status,
			COUNT(dr.id) AS reply_count,
			dt.created_at::text
		FROM public_discussion_threads dt
		JOIN users author ON author.id = dt.author_user_id
		LEFT JOIN public_discussion_replies dr ON dr.thread_id = dt.id
		GROUP BY dt.id, author.email, dt.category, dt.title, dt.moderation_status, dt.created_at
		ORDER BY dt.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DiscussionThreadRecord
	for rows.Next() {
		var item DiscussionThreadRecord
		if err := rows.Scan(&item.ID, &item.AuthorEmail, &item.Category, &item.Title, &item.Status, &item.ReplyCount, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListNotificationLogs(ctx context.Context) ([]NotificationLogRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			n.id::text,
			u.email,
			n.type,
			n.title,
			COALESCE(n.target_url, ''),
			n.is_read,
			n.created_at::text
		FROM notifications n
		JOIN users u ON u.id = n.user_id
		ORDER BY n.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []NotificationLogRecord
	for rows.Next() {
		var item NotificationLogRecord
		if err := rows.Scan(&item.ID, &item.UserEmail, &item.Type, &item.Title, &item.TargetURL, &item.Read, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListAILogs(ctx context.Context) ([]AIGenerationLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id::text,
			created_at::text,
			status,
			COALESCE(response_payload->>'provider', ''),
			COALESCE(response_payload->>'plan_title', ''),
			COALESCE(response_payload->>'error', ''),
			COALESCE(NULLIF(response_payload->>'attempt', ''), '0')::int
		FROM ai_generation_logs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AIGenerationLog
	for rows.Next() {
		var item AIGenerationLog
		if err := rows.Scan(&item.ID, &item.RequestedAt, &item.Status, &item.Provider, &item.PlanTitle, &item.Error, &item.Attempt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListEmailLogs(ctx context.Context) ([]EmailLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			el.id::text,
			el.type,
			el.recipient_email,
			el.subject,
			el.status,
			COALESCE(el.error, ''),
			el.created_at::text
		FROM email_logs el
		ORDER BY el.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []EmailLog
	for rows.Next() {
		var item EmailLog
		if err := rows.Scan(&item.ID, &item.Type, &item.To, &item.Subject, &item.Status, &item.Error, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PGStateStore) ListAuditLogs(ctx context.Context) ([]AuditLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			al.id::text,
			COALESCE(u.email, ''),
			al.action_type,
			al.resource_type,
			al.resource_id,
			al.created_at::text
		FROM admin_audit_logs al
		LEFT JOIN users u ON u.id = al.admin_user_id
		ORDER BY al.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AuditLog
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(&item.ID, &item.ActorEmail, &item.Action, &item.ResourceType, &item.ResourceID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func newProjectionRefs(state runtimeState) projectionRefs {
	refs := projectionRefs{
		userIDs:      map[string]string{},
		userByEmail:  map[string]string{},
		roleIDs:      map[string]string{},
		planIDs:      map[string]string{},
		scheduleIDs:  map[string]string{},
		schedulePlan: map[string]string{},
	}
	for _, user := range state.Users {
		userUUID := runtimeUUID("user", user.ID, user.Email)
		refs.userIDs[user.ID] = userUUID
		refs.userByEmail[user.Email] = userUUID
		for _, role := range user.Roles {
			refs.roleIDs[role] = runtimeUUID("role", role)
		}
		for _, plan := range user.PlanVersions {
			planUUID := runtimeUUID("plan", user.ID, plan.ID)
			refs.planIDs[plan.ID] = planUUID
			for _, item := range plan.Schedule {
				scheduleUUID := runtimeUUID("schedule", plan.ID, item.ID)
				refs.scheduleIDs[item.ID] = scheduleUUID
				refs.schedulePlan[item.ID] = planUUID
			}
		}
	}
	return refs
}

func projectCatalogs(ctx context.Context, tx pgx.Tx, state runtimeState, _ projectionRefs) error {
	equipment := append([]EquipmentItem(nil), state.EquipmentCatalog...)
	sort.Slice(equipment, func(i, j int) bool { return equipment[i].ID < equipment[j].ID })
	for _, item := range equipment {
		equipmentID := runtimeUUID("equipment", item.ID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO equipment_catalog (id, slug, category, location_type, media_url, active)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, equipmentID, projectSlug(translatedValue("ru", item.Names), item.ID), item.Category, item.LocationType, item.MediaURL, item.Active); err != nil {
			return fmt.Errorf("insert equipment catalog: %w", err)
		}
		for _, locale := range []string{"ru", "kk"} {
			name := translatedValue(locale, item.Names)
			description := translatedValue(locale, item.Descriptions)
			if description == "" {
				description = name
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO equipment_catalog_translations (equipment_id, locale, name, description)
				VALUES ($1, $2, $3, $4)
			`, equipmentID, locale, fallbackString(name, item.ID), fallbackString(description, fallbackString(name, item.ID))); err != nil {
				return fmt.Errorf("insert equipment translation: %w", err)
			}
		}
	}

	exercises := append([]ExerciseItem(nil), state.ExerciseCatalog...)
	sort.Slice(exercises, func(i, j int) bool { return exercises[i].ID < exercises[j].ID })
	for _, item := range exercises {
		exerciseID := runtimeUUID("exercise", item.ID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO exercise_catalog (id, slug, movement_pattern, difficulty, location_type, media_url, active)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, exerciseID, fallbackString(item.Slug, projectSlug(translatedValue("ru", item.Names), item.ID)), item.Movement, item.Difficulty, item.LocationType, "", item.Active); err != nil {
			return fmt.Errorf("insert exercise catalog: %w", err)
		}
		for _, locale := range []string{"ru", "kk"} {
			name := translatedValue(locale, item.Names)
			description := translatedValue(locale, item.Descriptions)
			technique := translatedValue(locale, item.Technique)
			if description == "" {
				description = name
			}
			if technique == "" {
				technique = description
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO exercise_catalog_translations (exercise_id, locale, name, description, technique_steps)
				VALUES ($1, $2, $3, $4, $5)
			`, exerciseID, locale, fallbackString(name, item.ID), fallbackString(description, fallbackString(name, item.ID)), fallbackString(technique, fallbackString(description, fallbackString(name, item.ID)))); err != nil {
				return fmt.Errorf("insert exercise translation: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO exercise_media (id, exercise_id, media_url, media_type)
			VALUES ($1, $2, $3, $4)
		`, runtimeUUID("exercise-media", item.ID), exerciseID, "https://example.com/media/"+fallbackString(item.Slug, item.ID), "image"); err != nil {
			return fmt.Errorf("insert exercise media: %w", err)
		}
	}
	return nil
}

func projectUsers(ctx context.Context, tx pgx.Tx, state runtimeState, refs projectionRefs) error {
	users := orderedUsers(state.Users)
	now := time.Now().UTC()
	for _, user := range users {
		userID := refs.userIDs[user.ID]
		var verifiedAt any
		if user.Verified {
			verifiedAt = now
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (id, email, password_hash, email_verified_at, locale, theme, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, userID, user.Email, user.PasswordHash, verifiedAt, user.Locale, user.Theme, now); err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		for _, role := range sortedStrings(user.Roles) {
			roleID := refs.roleIDs[role]
			if _, err := tx.Exec(ctx, `INSERT INTO roles (id, code) VALUES ($1, $2)`, roleID, role); err != nil {
				return fmt.Errorf("insert role: %w", err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID); err != nil {
				return fmt.Errorf("insert user role: %w", err)
			}
		}
		if user.SessionTokenHash != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("session", user.ID), userID, user.SessionTokenHash, now.Add(30*24*time.Hour), now); err != nil {
				return fmt.Errorf("insert session: %w", err)
			}
		}
		if user.VerifyTokenHash != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO email_verification_tokens (id, user_id, token_hash, expires_at, used_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("verify-token", user.ID), userID, user.VerifyTokenHash, now.Add(24*time.Hour), nil); err != nil {
				return fmt.Errorf("insert email verification token: %w", err)
			}
		}
		if user.ResetTokenHash != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, used_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("reset-token", user.ID), userID, user.ResetTokenHash, now.Add(2*time.Hour), nil); err != nil {
				return fmt.Errorf("insert password reset token: %w", err)
			}
		}
		if user.OnboardingDone {
			if _, err := tx.Exec(ctx, `
				INSERT INTO user_profiles (
					user_id, age, biological_sex, height_cm, current_weight_kg, target_weight_kg,
					primary_goal, program_duration_weeks, experience_level, activity_level,
					training_location, timezone, preferred_training_style, preferred_meal_style, hydration_preference,
					created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			`, userID, user.Profile.Age, user.Profile.BiologicalSex, user.Profile.HeightCM, user.Profile.CurrentWeightKG, user.Profile.TargetWeightKG, user.Profile.PrimaryGoal, user.Profile.ProgramDurationWeeks, user.Profile.ExperienceLevel, user.Profile.ActivityLevel, user.Profile.TrainingLocation, user.Profile.Timezone, nullIfEmpty(user.Profile.PreferredTrainingStyle), nullIfEmpty(user.Profile.PreferredMealStyle), nullIfEmpty(user.Profile.HydrationPreference), now, now); err != nil {
				return fmt.Errorf("insert user profile: %w", err)
			}
		}
		for _, goal := range sortedStrings(nonEmptyStrings([]string{user.Profile.PrimaryGoal})) {
			if _, err := tx.Exec(ctx, `INSERT INTO user_goals (user_id, goal_code) VALUES ($1, $2)`, userID, goal); err != nil {
				return fmt.Errorf("insert user goal: %w", err)
			}
		}
		for _, injury := range sortedStrings(user.Profile.Injuries) {
			if _, err := tx.Exec(ctx, `INSERT INTO user_injuries (user_id, injury_code) VALUES ($1, $2)`, userID, injury); err != nil {
				return fmt.Errorf("insert user injury: %w", err)
			}
		}
		for _, preference := range sortedStrings(user.Profile.DietaryPreferences) {
			if _, err := tx.Exec(ctx, `INSERT INTO dietary_preferences (user_id, preference_code) VALUES ($1, $2)`, userID, preference); err != nil {
				return fmt.Errorf("insert dietary preference: %w", err)
			}
		}
		for _, equipmentID := range sortedStrings(user.Profile.EquipmentIDs) {
			if _, err := tx.Exec(ctx, `INSERT INTO user_equipment (user_id, equipment_id) VALUES ($1, $2)`, userID, runtimeUUID("equipment", equipmentID)); err != nil {
				return fmt.Errorf("insert user equipment: %w", err)
			}
		}
		for idx, day := range user.Profile.AvailableTrainingDays {
			if _, err := tx.Exec(ctx, `
				INSERT INTO availability_rules (id, user_id, weekday, duration_min)
				VALUES ($1, $2, $3, $4)
			`, runtimeUUID("availability", user.ID, strconv.Itoa(idx), day.Weekday), userID, day.Weekday, day.DurationMin); err != nil {
				return fmt.Errorf("insert availability rule: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO hydration_targets (id, user_id, target_ml, override_ml, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, runtimeUUID("hydration-target", user.ID), userID, user.WaterTargetML, nullablePositive(user.WaterOverrideML), now); err != nil {
			return fmt.Errorf("insert hydration target: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO notification_preferences (user_id, workout_reminder, weekly_checkin_reminder, hydration_reminder, email_enabled)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, true, true, user.NotificationPreferences.HydrationReminder, user.NotificationPreferences.EmailEnabled); err != nil {
			return fmt.Errorf("insert notification preferences: %w", err)
		}
		if user.AssignedTrainerEmail != "" && refs.userByEmail[user.AssignedTrainerEmail] != "" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO trainer_assignments (user_id, trainer_id, assigned_at)
				VALUES ($1, $2, $3)
			`, userID, refs.userByEmail[user.AssignedTrainerEmail], now); err != nil {
				return fmt.Errorf("insert trainer assignment: %w", err)
			}
		}
	}
	return nil
}

func projectPlans(ctx context.Context, tx pgx.Tx, state runtimeState, refs projectionRefs) error {
	users := orderedUsers(state.Users)
	for _, user := range users {
		userID := refs.userIDs[user.ID]
		for planIndex, plan := range user.PlanVersions {
			planID := refs.planIDs[plan.ID]
			status := "superseded"
			if planIndex == len(user.PlanVersions)-1 && plan.SupersededAt == "" {
				status = "active"
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO generated_plan_versions (
					id, user_id, parent_plan_version_id, version_no, status, regeneration_reason, superseded_at, created_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, planID, userID, nullableUUID(plan.ParentPlanID, "plan", user.ID), plan.VersionNo, status, nullIfEmpty(plan.RegenerationReason), nullableTime(plan.SupersededAt), parseTimestamp(plan.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert plan version: %w", err)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO nutrition_plan_versions (
					id, plan_version_id, daily_calories, protein_g, carbs_g, fat_g, daily_water_ml
				) VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, runtimeUUID("nutrition", plan.ID), planID, plan.Nutrition.DailyCalories, plan.Nutrition.ProteinG, plan.Nutrition.CarbsG, plan.Nutrition.FatG, plan.Nutrition.DailyWaterML); err != nil {
				return fmt.Errorf("insert nutrition plan version: %w", err)
			}

			createdAt := parseTimestamp(plan.CreatedAt, time.Now().UTC())
			for scheduleIndex, item := range plan.Schedule {
				if _, err := tx.Exec(ctx, `
					INSERT INTO workout_schedule_instances (id, plan_version_id, scheduled_for, status, rescheduled_from_id)
					VALUES ($1, $2, $3, $4, $5)
				`, refs.scheduleIDs[item.ID], planID, projectedScheduleTime(createdAt, item.Weekday, scheduleIndex), item.Status, nullableUUID(item.RescheduledFromID, "schedule", plan.ID)); err != nil {
					return fmt.Errorf("insert workout schedule instance: %w", err)
				}
			}

			for _, week := range plan.Weeks {
				for dayIndex, day := range week.Days {
					templateID := runtimeUUID("workout-template", plan.ID, strconv.Itoa(week.WeekIndex), strconv.Itoa(dayIndex), day.Weekday, day.SessionName)
					if _, err := tx.Exec(ctx, `
						INSERT INTO workout_day_templates (id, plan_version_id, week_index, weekday, session_name)
						VALUES ($1, $2, $3, $4, $5)
					`, templateID, planID, week.WeekIndex, day.Weekday, day.SessionName); err != nil {
						return fmt.Errorf("insert workout day template: %w", err)
					}
					for _, exercise := range day.Exercises {
						if _, err := tx.Exec(ctx, `
							INSERT INTO workout_day_exercises (
								id, template_id, exercise_id, exercise_order, sets_count, reps_text, rest_sec
							) VALUES ($1, $2, $3, $4, $5, $6, $7)
						`, runtimeUUID("workout-day-exercise", templateID, strconv.Itoa(exercise.Order), exercise.ExerciseID), templateID, runtimeUUID("exercise", exercise.ExerciseID), exercise.Order, exercise.Sets, exercise.Reps, exercise.RestSec); err != nil {
							return fmt.Errorf("insert workout day exercise: %w", err)
						}
					}
				}
			}
		}
	}
	return nil
}

func projectTracking(ctx context.Context, tx pgx.Tx, state runtimeState, refs projectionRefs) error {
	users := orderedUsers(state.Users)
	for _, user := range users {
		userID := refs.userIDs[user.ID]
		for idx, log := range user.WorkoutLogs {
			planID := refs.schedulePlan[log.ScheduleID]
			if planID == "" && len(user.PlanVersions) > 0 {
				planID = refs.planIDs[user.PlanVersions[len(user.PlanVersions)-1].ID]
			}
			scheduleID := refs.scheduleIDs[log.ScheduleID]
			if scheduleID == "" {
				scheduleID = runtimeUUID("schedule", user.ID, log.ScheduleID)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO workout_logs (
					id, plan_version_id, schedule_instance_id, status, discomfort_flag, difficulty, note, completed_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, runtimeUUID("workout-log", user.ID, strconv.Itoa(idx), log.ScheduleID), planID, scheduleID, log.Status, log.DiscomfortFlag, nullablePositive(log.Difficulty), nullIfEmpty(log.Note), nullableTime(log.CompletionTime)); err != nil {
				return fmt.Errorf("insert workout log: %w", err)
			}
		}
		for idx, log := range user.MealLogs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO meal_logs (id, user_id, status, note, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("meal-log", user.ID, strconv.Itoa(idx)), userID, log.Status, nullIfEmpty(log.Note), parseTimestamp(log.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert meal log: %w", err)
			}
		}
		for idx, log := range user.WaterLogs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO water_logs (id, user_id, amount_ml, created_at)
				VALUES ($1, $2, $3, $4)
			`, runtimeUUID("water-log", user.ID, strconv.Itoa(idx)), userID, log.AmountML, parseTimestamp(log.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert water log: %w", err)
			}
		}
		for idx, checkin := range user.WeeklyCheckins {
			if _, err := tx.Exec(ctx, `
				INSERT INTO weekly_checkins (id, user_id, weight_kg, energy_level, note, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, runtimeUUID("weekly-checkin", user.ID, strconv.Itoa(idx)), userID, checkin.WeightKG, nullablePositive(checkin.EnergyLevel), nullIfEmpty(checkin.Note), parseTimestamp(checkin.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert weekly checkin: %w", err)
			}
		}
	}
	return nil
}

func projectNotifications(ctx context.Context, tx pgx.Tx, state runtimeState, refs projectionRefs) error {
	users := orderedUsers(state.Users)
	for _, user := range users {
		userID := refs.userIDs[user.ID]
		for idx, notification := range user.Notifications {
			if _, err := tx.Exec(ctx, `
				INSERT INTO notifications (id, user_id, type, title, body, target_url, is_read, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, runtimeUUID("notification", user.ID, strconv.Itoa(idx), notification.ID), userID, notification.Type, notification.Title, "", nullIfEmpty(notification.TargetURL), notification.Read, parseTimestamp(notification.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert notification: %w", err)
			}
		}
		for idx, log := range user.AIGenerationLogs {
			requestPayload := mustJSON(map[string]any{
				"user_ref": user.ID,
				"status":   log.Status,
			})
			responsePayload := mustJSON(map[string]any{
				"provider":   log.Provider,
				"plan_title": log.PlanTitle,
				"error":      log.Error,
				"attempt":    log.Attempt,
			})
			if _, err := tx.Exec(ctx, `
				INSERT INTO ai_generation_logs (id, user_id, generation_type, request_payload, response_payload, status, created_at)
				VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)
			`, runtimeUUID("ai-log", user.ID, strconv.Itoa(idx), log.ID), userID, "plan_generation", requestPayload, responsePayload, log.Status, parseTimestamp(log.RequestedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert ai generation log: %w", err)
			}
		}
	}
	return nil
}

func projectOps(ctx context.Context, tx pgx.Tx, state runtimeState, refs projectionRefs) error {
	for _, user := range orderedUsers(state.Users) {
		userID := refs.userIDs[user.ID]
		for idx, logEntry := range user.EmailLogs {
			if _, err := tx.Exec(ctx, `
				INSERT INTO email_logs (id, user_id, type, recipient_email, subject, status, error, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, runtimeUUID("email-log", user.ID, strconv.Itoa(idx), logEntry.ID), userID, logEntry.Type, user.Email, logEntry.Subject, logEntry.Status, nullIfEmpty(logEntry.Error), parseTimestamp(logEntry.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert email log: %w", err)
			}
		}
	}
	for _, note := range orderedTrainerNotes(state.TrainerNotes) {
		trainerID := refs.userByEmail[note.TrainerEmail]
		userID := refs.userByEmail[note.UserEmail]
		if trainerID == "" || userID == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO trainer_notes (id, trainer_id, user_id, body, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, runtimeUUID("trainer-note", note.ID, note.TrainerEmail, note.UserEmail), trainerID, userID, note.Body, parseTimestamp(note.CreatedAt, time.Now().UTC())); err != nil {
			return fmt.Errorf("insert trainer note: %w", err)
		}
	}
	for _, thread := range orderedSupportThreads(state.SupportThreads) {
		userID := refs.userByEmail[thread.UserEmail]
		if userID == "" {
			continue
		}
		var assignedTo any
		if refs.userByEmail[thread.AssignedToEmail] != "" {
			assignedTo = refs.userByEmail[thread.AssignedToEmail]
		}
		threadID := runtimeUUID("support-thread", thread.ID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO support_threads (id, user_id, status, assigned_to, title, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, threadID, userID, thread.Status, assignedTo, thread.Title, parseTimestamp(thread.CreatedAt, time.Now().UTC())); err != nil {
			return fmt.Errorf("insert support thread: %w", err)
		}
		for idx, message := range thread.Messages {
			authorID := refs.userByEmail[message.AuthorEmail]
			if authorID == "" {
				continue
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO support_messages (id, thread_id, author_user_id, body, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("support-message", thread.ID, strconv.Itoa(idx)), threadID, authorID, message.Body, parseTimestamp(message.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert support message: %w", err)
			}
		}
	}
	for _, thread := range orderedDiscussionThreads(state.DiscussionThreads) {
		authorID := refs.userByEmail[thread.AuthorEmail]
		if authorID == "" {
			continue
		}
		threadID := runtimeUUID("discussion-thread", thread.ID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO public_discussion_threads (id, author_user_id, category, title, body, moderation_status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, threadID, authorID, thread.Category, thread.Title, thread.Body, thread.Status, parseTimestamp(thread.CreatedAt, time.Now().UTC())); err != nil {
			return fmt.Errorf("insert discussion thread: %w", err)
		}
		for idx, reply := range thread.Replies {
			replyAuthorID := refs.userByEmail[reply.AuthorEmail]
			if replyAuthorID == "" {
				continue
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO public_discussion_replies (id, thread_id, author_user_id, body, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, runtimeUUID("discussion-reply", thread.ID, strconv.Itoa(idx)), threadID, replyAuthorID, reply.Body, parseTimestamp(reply.CreatedAt, time.Now().UTC())); err != nil {
				return fmt.Errorf("insert discussion reply: %w", err)
			}
		}
	}
	for idx, item := range state.AuditLogs {
		adminID := refs.userByEmail[item.ActorEmail]
		if adminID == "" {
			adminID = runtimeUUID("admin-actor", item.ActorEmail)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO admin_audit_logs (id, admin_user_id, action_type, resource_type, resource_id, details, created_at)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
		`, runtimeUUID("audit-log", strconv.Itoa(idx), item.ID), adminID, item.Action, item.ResourceType, item.ResourceID, mustJSON(map[string]any{"actor_email": item.ActorEmail}), parseTimestamp(item.CreatedAt, time.Now().UTC())); err != nil {
			return fmt.Errorf("insert admin audit log: %w", err)
		}
	}
	return nil
}

func orderedUsers(items map[string]*User) []*User {
	result := make([]*User, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Email < result[j].Email })
	return result
}

func orderedSupportThreads(items map[string]*SupportThread) []*SupportThread {
	result := make([]*SupportThread, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func orderedDiscussionThreads(items map[string]*DiscussionThread) []*DiscussionThread {
	result := make([]*DiscussionThread, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func orderedTrainerNotes(items map[string][]TrainerNote) []TrainerNote {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var result []TrainerNote
	for _, key := range keys {
		notes := append([]TrainerNote(nil), items[key]...)
		sort.Slice(notes, func(i, j int) bool { return notes[i].CreatedAt < notes[j].CreatedAt })
		result = append(result, notes...)
	}
	return result
}

func runtimeUUID(namespace string, parts ...string) string {
	values := append([]string{namespace}, parts...)
	sum := sha1.Sum([]byte(strings.Join(values, "\x1f")))
	buf := sum[:16]
	buf[6] = (buf[6] & 0x0f) | 0x50
	buf[8] = (buf[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

func projectedScheduleTime(base time.Time, weekday string, offset int) time.Time {
	start := time.Date(base.Year(), base.Month(), base.Day(), 9, 0, 0, 0, time.UTC)
	target := weekdayToTime(weekday)
	shift := (int(target) - int(start.Weekday()) + 7) % 7
	return start.AddDate(0, 0, shift+offset*7)
}

func parseTimestamp(raw string, fallback time.Time) time.Time {
	if strings.TrimSpace(raw) == "" {
		return fallback.UTC()
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return fallback.UTC()
	}
	return parsed.UTC()
}

func nullableTime(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return parsed.UTC()
}

func nullableUUID(raw string, namespace string, parts ...string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	values := append(parts, raw)
	return runtimeUUID(namespace, values...)
}

func nullablePositive(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func mustJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func projectSlug(value, fallback string) string {
	base := strings.ToLower(strings.TrimSpace(value))
	base = strings.ReplaceAll(base, "_", "-")
	base = nonSlugChars.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = strings.ToLower(strings.TrimSpace(fallback))
		base = strings.ReplaceAll(base, "_", "-")
		base = nonSlugChars.ReplaceAllString(base, "-")
		base = strings.Trim(base, "-")
	}
	if base == "" {
		return "item"
	}
	return base
}

func sortedStrings(values []string) []string {
	items := append([]string(nil), values...)
	sort.Strings(items)
	return items
}

func nonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func projectionPlanHealth(workouts, skipped int) string {
	switch {
	case skipped >= 2:
		return "adaptation_recommended"
	case skipped > 0 || workouts == 0:
		return "attention_needed"
	default:
		return "healthy"
	}
}
