package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"training/backend/internal/app"
	"training/backend/internal/config"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var options []app.Option
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		migrationsDir := resolveMigrationsDir()
		if err := retry(ctx, 8, time.Second, func() error {
			return app.RunMigrations(ctx, cfg.DatabaseURL, migrationsDir)
		}); err != nil {
			logger.Error("migration failed", "error", err)
			os.Exit(1)
		}
		var store *app.PGStateStore
		err := retry(ctx, 8, time.Second, func() error {
			var openErr error
			store, openErr = app.NewPGStateStore(ctx, cfg.DatabaseURL)
			return openErr
		})
		if err != nil {
			logger.Error("postgres store init failed", "error", err)
			os.Exit(1)
		}
		defer store.Close()
		options = append(options, app.WithStateStore(store), app.WithRelationalStore(store))
	}
	if cfg.CORSAllowedOrigins != "" {
		options = append(options, app.WithAllowedOrigins(splitCSV(cfg.CORSAllowedOrigins)...))
	}
	options = append(options, app.WithFrontendURL(cfg.FrontendURL))
	if cfg.AIAPIKey != "" && cfg.AIAPIBaseURL != "" && cfg.AIModel != "" {
		options = append(options, app.WithAIProvider(app.OpenAICompatibleProvider{
			BaseURL: cfg.AIAPIBaseURL,
			APIKey:  cfg.AIAPIKey,
			Model:   cfg.AIModel,
			Logger:  logger,
		}))
	}
	if provider := app.NewGoogleOAuthProvider(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL); provider != nil {
		options = append(options, app.WithGoogleAuthProvider(provider))
	}
	application := app.New(logger, options...)
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           application.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := app.ShutdownContext()
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt == attempts {
				break
			}
			select {
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) && lastErr != nil {
					return lastErr
				}
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		return nil
	}
	return lastErr
}

func splitCSV(value string) []string {
	var items []string
	for _, raw := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func resolveMigrationsDir() string {
	candidates := []string{
		filepath.Join("backend", "migrations"),
		"migrations",
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return filepath.Join("backend", "migrations")
}
