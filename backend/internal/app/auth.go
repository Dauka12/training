package app

import (
	"net/http"
	"os"
	"strings"
)

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleReady(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.authLimiter.IsLimited(authKey(r, "register"), a.now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many attempts")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Locale   string `json:"locale"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if !isValidEmail(req.Email) || len(req.Password) < 10 {
		a.authLimiter.RegisterFailure(authKey(r, "register"), a.now())
		writeError(w, http.StatusBadRequest, "validation_error", "Invalid email or password")
		return
	}
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not register")
		return
	}
	rawToken := token(16)

	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.byMail[req.Email]; exists {
		a.authLimiter.RegisterFailure(authKey(r, "register"), a.now())
		writeError(w, http.StatusConflict, "email_in_use", "Email already in use")
		return
	}
	user := &User{
		ID:           token(8),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Locale:       normalizeLocale(req.Locale),
		Theme:        "light",
		Roles:        []string{"user"},
		NotificationPreferences: NotificationPreferences{
			HydrationReminder: true,
		},
		VerifyTokenHash: subtleHash(rawToken),
		VerifyToken:     rawToken,
	}
	a.users[user.ID] = user
	a.byMail[user.Email] = user
	a.persistStateLocked()
	a.authLimiter.Reset(authKey(r, "register"))
	response := map[string]any{"status": "registered"}
	if isDevelopmentMode() {
		response["dev_verification_token"] = user.VerifyToken
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, user := range a.users {
		if user.VerifyTokenHash != "" && subtleHash(req.Token) == user.VerifyTokenHash {
			user.Verified = true
			user.VerifyTokenHash = ""
			user.VerifyToken = ""
			a.persistStateLocked()
			writeJSON(w, http.StatusOK, map[string]string{"status": "verified"})
			return
		}
	}
	writeError(w, http.StatusBadRequest, "invalid_token", "Invalid token")
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	key := authKey(r, "login")
	if a.authLimiter.IsLimited(key, a.now()) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many attempts")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user := a.byMail[normalizeEmail(req.Email)]
	if user == nil || !verifyPassword(user.PasswordHash, req.Password) {
		a.authLimiter.RegisterFailure(key, a.now())
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
		return
	}
	if !user.Verified {
		a.authLimiter.RegisterFailure(key, a.now())
		writeError(w, http.StatusForbidden, "email_not_verified", "Verify email first")
		return
	}
	rawSession := token(16)
	rawCSRF := token(16)
	user.SessionTokenHash = subtleHash(rawSession)
	user.CSRFTokenHash = subtleHash(rawCSRF)
	a.persistStateLocked()
	a.authLimiter.Reset(key)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    rawSession,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    rawCSRF,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "csrf_token": rawCSRF})
}

func (a *App) handleLogout(w http.ResponseWriter, _ *http.Request, user *User) {
	a.mu.Lock()
	defer a.mu.Unlock()
	user.SessionTokenHash = ""
	user.CSRFTokenHash = ""
	a.persistStateLocked()
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	http.SetCookie(w, &http.Cookie{Name: "csrf", Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteStrictMode})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (a *App) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if user := a.byMail[normalizeEmail(req.Email)]; user != nil {
		rawToken := token(16)
		user.ResetTokenHash = subtleHash(rawToken)
		user.ResetToken = rawToken
		a.persistStateLocked()
	}
	response := map[string]any{"status": "sent_if_exists"}
	if isDevelopmentMode() {
		if user := a.byMail[normalizeEmail(req.Email)]; user != nil && user.ResetToken != "" {
			response["dev_reset_token"] = user.ResetToken
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if len(req.NewPassword) < 10 {
		writeError(w, http.StatusBadRequest, "validation_error", "Password too short")
		return
	}
	passwordHash, err := hashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Could not reset password")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, user := range a.users {
		if user.ResetTokenHash != "" && subtleHash(req.Token) == user.ResetTokenHash {
			user.PasswordHash = passwordHash
			user.ResetTokenHash = ""
			user.ResetToken = ""
			user.SessionTokenHash = ""
			user.CSRFTokenHash = ""
			a.persistStateLocked()
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
			http.SetCookie(w, &http.Cookie{Name: "csrf", Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteStrictMode})
			writeJSON(w, http.StatusOK, map[string]string{"status": "password_reset"})
			return
		}
	}
	writeError(w, http.StatusBadRequest, "invalid_token", "Invalid token")
}

func (a *App) withAuth(next func(http.ResponseWriter, *http.Request, *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
			return
		}
		a.mu.Lock()
		var matched *User
		for _, user := range a.users {
			if user.SessionTokenHash != "" && subtleHash(cookie.Value) == user.SessionTokenHash {
				matched = user
				break
			}
		}
		a.mu.Unlock()
		if matched == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Unauthorized")
			return
		}
		if requiresCSRF(r.Method) {
			csrfCookie, cookieErr := r.Cookie("csrf")
			csrfHeader := r.Header.Get("X-CSRF-Token")
			if cookieErr != nil || csrfCookie.Value == "" || csrfHeader == "" || csrfHeader != csrfCookie.Value || subtleHash(csrfHeader) != matched.CSRFTokenHash {
				writeError(w, http.StatusForbidden, "csrf_invalid", "CSRF token mismatch")
				return
			}
		}
		next(w, r, matched)
	}
}

func (a *App) withRoles(next func(http.ResponseWriter, *http.Request, *User), roles ...string) func(http.ResponseWriter, *http.Request, *User) {
	return func(w http.ResponseWriter, r *http.Request, user *User) {
		if !hasRole(user, roles...) {
			writeError(w, http.StatusForbidden, "forbidden", "Forbidden")
			return
		}
		next(w, r, user)
	}
}

func (a *App) handleMe(w http.ResponseWriter, _ *http.Request, user *User) {
	writeJSON(w, http.StatusOK, map[string]any{
		"id":              user.ID,
		"email":           user.Email,
		"locale":          user.Locale,
		"theme":           user.Theme,
		"roles":           user.Roles,
		"verified":        user.Verified,
		"onboarding_done": user.OnboardingDone,
		"profile":         user.Profile,
	})
}

func (a *App) handlePreferences(w http.ResponseWriter, r *http.Request, user *User) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Locale          string `json:"locale"`
		Theme           string `json:"theme"`
		WaterOverrideML int    `json:"water_override_ml"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user.Locale = normalizeLocale(req.Locale)
	user.Theme = normalizeTheme(req.Theme)
	if req.WaterOverrideML > 0 {
		user.WaterOverrideML = req.WaterOverrideML
		user.WaterTargetML = req.WaterOverrideML
	}
	a.persistStateLocked()
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (a *App) handleNotificationPreferences(w http.ResponseWriter, r *http.Request, user *User) {
	switch r.Method {
	case http.MethodGet:
		a.mu.Lock()
		defer a.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"preferences": user.NotificationPreferences})
	case http.MethodPut:
		var req NotificationPreferences
		if !decodeJSON(w, r, &req) {
			return
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		user.NotificationPreferences = req
		a.persistStateLocked()
		writeJSON(w, http.StatusOK, map[string]any{"status": "saved", "preferences": user.NotificationPreferences})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isDevelopmentMode() bool {
	return strings.EqualFold(os.Getenv("APP_ENV"), "development")
}

func requiresCSRF(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func authKey(r *http.Request, action string) string {
	return action + ":" + strings.Split(r.RemoteAddr, ":")[0]
}
