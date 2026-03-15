package app

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"
)

type StateStore interface {
	Load(context.Context) ([]byte, error)
	Save(context.Context, []byte) error
}

type runtimeState struct {
	Users             map[string]*User             `json:"users"`
	EquipmentCatalog  []EquipmentItem              `json:"equipment_catalog"`
	ExerciseCatalog   []ExerciseItem               `json:"exercise_catalog"`
	TrainerNotes      map[string][]TrainerNote     `json:"trainer_notes"`
	SupportThreads    map[string]*SupportThread    `json:"support_threads"`
	DiscussionThreads map[string]*DiscussionThread `json:"discussion_threads"`
	AuditLogs         []AuditLog                   `json:"audit_logs"`
}

func (a *App) restoreState(ctx context.Context) {
	if a.stateStore == nil {
		return
	}

	payload, err := a.stateStore.Load(ctx)
	if err != nil {
		a.log.Error("state load failed", "error", err)
		return
	}
	if len(payload) == 0 {
		return
	}

	var state runtimeState
	if err := json.Unmarshal(payload, &state); err != nil {
		a.log.Error("state decode failed", "error", err)
		return
	}

	if len(state.Users) > 0 {
		a.users = state.Users
		a.byMail = map[string]*User{}
		for _, user := range a.users {
			a.byMail[strings.ToLower(user.Email)] = user
		}
	}
	if len(state.EquipmentCatalog) > 0 {
		a.equipmentCatalog = state.EquipmentCatalog
	}
	if len(state.ExerciseCatalog) > 0 {
		a.exerciseCatalog = state.ExerciseCatalog
	}
	if len(state.TrainerNotes) > 0 {
		a.trainerNotes = state.TrainerNotes
	}
	if len(state.SupportThreads) > 0 {
		a.supportThreads = state.SupportThreads
	}
	if len(state.DiscussionThreads) > 0 {
		a.discussionThreads = state.DiscussionThreads
	}
	if len(state.AuditLogs) > 0 {
		a.auditLogs = state.AuditLogs
	}
}

func (a *App) persistStateLocked() {
	if a.stateStore == nil && a.relationalStore == nil {
		return
	}
	state := runtimeState{
		Users:             a.users,
		EquipmentCatalog:  a.equipmentCatalog,
		ExerciseCatalog:   a.exerciseCatalog,
		TrainerNotes:      a.trainerNotes,
		SupportThreads:    a.supportThreads,
		DiscussionThreads: a.discussionThreads,
		AuditLogs:         a.auditLogs,
	}
	payload, err := json.Marshal(state)
	if err != nil {
		a.log.Error("state encode failed", "error", err)
		return
	}
	if a.stateStore != nil {
		if err := a.stateStore.Save(context.Background(), payload); err != nil {
			a.log.Error("state save failed", "error", err)
		}
	}
	if a.relationalStore != nil {
		if err := a.relationalStore.Project(context.Background(), state); err != nil {
			a.log.Error("relational projection failed", "error", err)
		}
	}
}

type AuthRateLimiter struct {
	mu        sync.Mutex
	window    time.Duration
	maxFailed int
	attempts  map[string]authAttempt
}

type authAttempt struct {
	Count     int
	StartedAt time.Time
}

func NewAuthRateLimiter() *AuthRateLimiter {
	return &AuthRateLimiter{
		window:    10 * time.Minute,
		maxFailed: 5,
		attempts:  map[string]authAttempt{},
	}
}

func (l *AuthRateLimiter) IsLimited(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.attempts[key]
	if !ok {
		return false
	}
	if now.Sub(item.StartedAt) > l.window {
		delete(l.attempts, key)
		return false
	}
	return item.Count >= l.maxFailed
}

func (l *AuthRateLimiter) RegisterFailure(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.attempts[key]
	if !ok || now.Sub(item.StartedAt) > l.window {
		l.attempts[key] = authAttempt{Count: 1, StartedAt: now}
		return
	}
	item.Count++
	l.attempts[key] = item
}

func (l *AuthRateLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

type LogStateStore struct {
	Logger *slog.Logger
}

func (s LogStateStore) Load(context.Context) ([]byte, error) {
	return nil, nil
}

func (s LogStateStore) Save(_ context.Context, payload []byte) error {
	if s.Logger != nil {
		s.Logger.Debug("state snapshot updated", "bytes", len(payload))
	}
	return nil
}
