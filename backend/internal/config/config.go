package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AppEnv                 string
	ListenAddr             string
	BackendURL             string
	AIAPIBaseURL           string
	AIAPIKey               string
	AIModel                string
	DatabaseURL            string
	SessionSecret          string
	CORSAllowedOrigins     string
	EnableNotificationMail bool
	FrontendURL            string
	GoogleClientID         string
	GoogleClientSecret     string
	GoogleRedirectURL      string
}

func Load() Config {
	loadEnvFile(".env")
	loadEnvFile(filepath.Join("backend", ".env"))

	return Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		ListenAddr:             resolveListenAddr(getEnv("APP_ADDR", ""), getEnv("APP_PORT", "")),
		BackendURL:             getEnv("BACKEND_URL", "http://localhost:8080"),
		AIAPIBaseURL:           getEnv("AI_API_BASE_URL", ""),
		AIAPIKey:               getEnv("AI_API_KEY", ""),
		AIModel:                getEnv("AI_MODEL", ""),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		SessionSecret:          getEnv("SESSION_SECRET", ""),
		CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		EnableNotificationMail: strings.EqualFold(getEnv("ENABLE_NOTIFICATION_EMAILS", "false"), "true"),
		FrontendURL:            getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoogleClientID:         getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:     getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:      getEnv("GOOGLE_REDIRECT_URL", ""),
	}
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveListenAddr(appAddr, appPort string) string {
	if appAddr != "" {
		return appAddr
	}
	if appPort != "" {
		return ":" + appPort
	}
	return ":8080"
}
