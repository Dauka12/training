package app

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	emailintegration "training/backend/internal/integrations/email"
)

type App struct {
	log               *slog.Logger
	mu                sync.Mutex
	users             map[string]*User
	byMail            map[string]*User
	equipmentCatalog  []EquipmentItem
	exerciseCatalog   []ExerciseItem
	trainerNotes      map[string][]TrainerNote
	supportThreads    map[string]*SupportThread
	discussionThreads map[string]*DiscussionThread
	aiProvider        AIProvider
	stateStore        StateStore
	relationalStore   RelationalStore
	authLimiter       *AuthRateLimiter
	allowedOrigins    []string
	emailSender       emailintegration.Sender
	clock             Clock
	auditLogs         []AuditLog
}

type Option func(*App)

func WithAIProvider(provider AIProvider) Option {
	return func(app *App) {
		if provider != nil {
			app.aiProvider = provider
		}
	}
}

func WithStateStore(store StateStore) Option {
	return func(app *App) {
		app.stateStore = store
	}
}

func WithRelationalStore(store RelationalStore) Option {
	return func(app *App) {
		app.relationalStore = store
	}
}

func WithAllowedOrigins(origins ...string) Option {
	return func(app *App) {
		app.allowedOrigins = append([]string(nil), origins...)
	}
}

func WithEmailSender(sender emailintegration.Sender) Option {
	return func(app *App) {
		app.emailSender = sender
	}
}

func WithClock(clock Clock) Option {
	return func(app *App) {
		if clock != nil {
			app.clock = clock
		}
	}
}

func New(logger *slog.Logger, options ...Option) *App {
	app := &App{
		log:               logger,
		users:             map[string]*User{},
		byMail:            map[string]*User{},
		equipmentCatalog:  seedEquipmentCatalog(),
		exerciseCatalog:   seedExerciseCatalog(),
		trainerNotes:      map[string][]TrainerNote{},
		supportThreads:    map[string]*SupportThread{},
		discussionThreads: map[string]*DiscussionThread{},
		aiProvider:        StaticAIProvider{},
		authLimiter:       NewAuthRateLimiter(),
		allowedOrigins:    []string{"http://localhost:5173"},
		clock:             RealClock{},
	}
	for _, option := range options {
		option(app)
	}
	app.restoreState(context.Background())
	app.seedDevelopmentUsers()
	return app
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/readyz", a.handleReady)
	mux.HandleFunc("/api/v1/auth/register", a.handleRegister)
	mux.HandleFunc("/api/v1/auth/verify-email", a.handleVerifyEmail)
	mux.HandleFunc("/api/v1/auth/login", a.handleLogin)
	mux.HandleFunc("/api/v1/auth/logout", a.withAuth(a.handleLogout))
	mux.HandleFunc("/api/v1/auth/forgot-password", a.handleForgotPassword)
	mux.HandleFunc("/api/v1/auth/reset-password", a.handleResetPassword)
	mux.HandleFunc("/api/v1/me", a.withAuth(a.handleMe))
	mux.HandleFunc("/api/v1/me/preferences", a.withAuth(a.handlePreferences))
	mux.HandleFunc("/api/v1/onboarding", a.withAuth(a.handleOnboarding))
	mux.HandleFunc("/api/v1/plans/generate", a.withAuth(a.handleGeneratePlan))
	mux.HandleFunc("/api/v1/plans/active", a.withAuth(a.handleActivePlan))
	mux.HandleFunc("/api/v1/dashboard/today", a.withAuth(a.handleDashboard))
	mux.HandleFunc("/api/v1/tracking/workouts/log", a.withAuth(a.handleWorkoutLog))
	mux.HandleFunc("/api/v1/tracking/meals", a.withAuth(a.handleMealLog))
	mux.HandleFunc("/api/v1/tracking/water", a.withAuth(a.handleWaterLog))
	mux.HandleFunc("/api/v1/tracking/hydration/summary", a.withAuth(a.handleHydrationSummary))
	mux.HandleFunc("/api/v1/checkins/weekly", a.withAuth(a.handleWeeklyCheckin))
	mux.HandleFunc("/api/v1/notifications", a.withAuth(a.handleNotifications))
	mux.HandleFunc("/api/v1/notifications/read", a.withAuth(a.handleReadNotification))
	mux.HandleFunc("/api/v1/notifications/preferences", a.withAuth(a.handleNotificationPreferences))
	mux.HandleFunc("/api/v1/trainer/users", a.withAuth(a.withRoles(a.handleTrainerUsers, "trainer", "admin")))
	mux.HandleFunc("/api/v1/trainer/users/", a.withAuth(a.withRoles(a.handleTrainerUserSubroutes, "trainer", "admin")))
	mux.HandleFunc("/api/v1/admin/users", a.withAuth(a.withRoles(a.handleAdminUsers, "admin")))
	mux.HandleFunc("/api/v1/admin/trainers", a.withAuth(a.withRoles(a.handleAdminTrainers, "admin")))
	mux.HandleFunc("/api/v1/admin/trainers/assign", a.withAuth(a.withRoles(a.handleAssignTrainer, "admin")))
	mux.HandleFunc("/api/v1/admin/catalog/equipment", a.withAuth(a.withRoles(a.handleAdminEquipment, "admin")))
	mux.HandleFunc("/api/v1/admin/catalog/exercises", a.withAuth(a.withRoles(a.handleAdminExercises, "admin")))
	mux.HandleFunc("/api/v1/admin/logs/ai", a.withAuth(a.withRoles(a.handleAdminAILogs, "admin")))
	mux.HandleFunc("/api/v1/admin/logs/email", a.withAuth(a.withRoles(a.handleAdminEmailLogs, "admin")))
	mux.HandleFunc("/api/v1/admin/logs/audit", a.withAuth(a.withRoles(a.handleAdminAuditLogs, "admin")))
	mux.HandleFunc("/api/v1/admin/logs/notifications", a.withAuth(a.withRoles(a.handleAdminNotificationLogs, "admin")))
	mux.HandleFunc("/api/v1/admin/support/threads", a.withAuth(a.withRoles(a.handleAdminSupportThreads, "admin")))
	mux.HandleFunc("/api/v1/admin/support/threads/", a.withAuth(a.withRoles(a.handleAdminSupportThreadSubroutes, "admin")))
	mux.HandleFunc("/api/v1/admin/discussions/threads", a.withAuth(a.withRoles(a.handleAdminDiscussionThreads, "admin")))
	mux.HandleFunc("/api/v1/admin/discussions/threads/", a.withAuth(a.withRoles(a.handleAdminDiscussionThreadSubroutes, "admin")))
	mux.HandleFunc("/api/v1/support/threads", a.withAuth(a.handleSupportThreads))
	mux.HandleFunc("/api/v1/support/threads/", a.withAuth(a.handleSupportThreadSubroutes))
	mux.HandleFunc("/api/v1/discussions/threads", a.withAuth(a.handleDiscussionThreads))
	mux.HandleFunc("/api/v1/discussions/threads/", a.withAuth(a.handleDiscussionThreadSubroutes))
	return withSecurityHeaders(withRequestID(withCORS(mux, a.allowedOrigins...)))
}
