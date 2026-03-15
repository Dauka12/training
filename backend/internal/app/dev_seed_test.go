package app

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

func TestNewSeedsDevelopmentUsers(t *testing.T) {
	previous := os.Getenv("APP_ENV")
	t.Cleanup(func() {
		if previous == "" {
			_ = os.Unsetenv("APP_ENV")
			return
		}
		_ = os.Setenv("APP_ENV", previous)
	})
	_ = os.Setenv("APP_ENV", "development")

	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))

	admin := app.byMail["admin@local.test"]
	if admin == nil {
		t.Fatal("expected seeded admin user")
	}
	if !admin.Verified || !hasRole(admin, "admin") {
		t.Fatalf("expected seeded admin to be verified admin, got %+v", admin)
	}
	if !verifyPassword(admin.PasswordHash, developmentSeedPassword) {
		t.Fatal("expected seeded admin password to be valid")
	}

	trainer := app.byMail["trainer@local.test"]
	if trainer == nil || !trainer.Verified || !hasRole(trainer, "trainer") {
		t.Fatal("expected seeded trainer user")
	}
}

func TestNewDoesNotSeedDevelopmentUsersOutsideDevelopment(t *testing.T) {
	previous := os.Getenv("APP_ENV")
	t.Cleanup(func() {
		if previous == "" {
			_ = os.Unsetenv("APP_ENV")
			return
		}
		_ = os.Setenv("APP_ENV", previous)
	})
	_ = os.Setenv("APP_ENV", "production")

	app := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	if app.byMail["admin@local.test"] != nil {
		t.Fatal("did not expect development admin seed in production")
	}
}
