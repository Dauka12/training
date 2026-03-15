package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleStateCookieName = "google_oauth_state"

type GoogleIdentity struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type GoogleAuthProvider interface {
	Enabled() bool
	AuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (GoogleIdentity, error)
}

type GoogleOAuthProvider struct {
	config     *oauth2.Config
	httpClient *http.Client
}

func NewGoogleOAuthProvider(clientID, clientSecret, redirectURL string) *GoogleOAuthProvider {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" || strings.TrimSpace(redirectURL) == "" {
		return nil
	}
	return &GoogleOAuthProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"openid", "email", "profile"},
		},
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *GoogleOAuthProvider) Enabled() bool {
	return p != nil && p.config != nil
}

func (p *GoogleOAuthProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (GoogleIdentity, error) {
	if p == nil || p.config == nil {
		return GoogleIdentity{}, errors.New("google auth disabled")
	}
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return GoogleIdentity{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return GoogleIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return GoogleIdentity{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoogleIdentity{}, errors.New("google userinfo request failed")
	}

	var identity GoogleIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return GoogleIdentity{}, err
	}
	return identity, nil
}

func (a *App) handleGoogleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	state := token(16)
	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if a.googleAuth != nil && a.googleAuth.Enabled() {
		http.Redirect(w, r, a.googleAuth.AuthURL(state), http.StatusFound)
		return
	}

	if isDevelopmentMode() {
		http.Redirect(w, r, "/api/v1/auth/google/dev-callback?state="+url.QueryEscape(state), http.StatusFound)
		return
	}

	writeError(w, http.StatusServiceUnavailable, "google_auth_unavailable", "Google sign-in is unavailable")
}

func (a *App) handleGoogleDevCallback(w http.ResponseWriter, r *http.Request) {
	if !isDevelopmentMode() {
		writeError(w, http.StatusNotFound, "not_found", "Not found")
		return
	}

	if err := validateGoogleState(r); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_state", "Invalid Google state")
		return
	}

	identity := GoogleIdentity{
		Email:         "google-dev@example.com",
		EmailVerified: true,
		Name:          "Google Dev",
	}
	a.completeGoogleLogin(w, r, identity)
}

func (a *App) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if a.googleAuth == nil || !a.googleAuth.Enabled() {
		writeError(w, http.StatusServiceUnavailable, "google_auth_unavailable", "Google sign-in is unavailable")
		return
	}
	if err := validateGoogleState(r); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_state", "Invalid Google state")
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing_code", "Missing Google code")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	identity, err := a.googleAuth.ExchangeCode(ctx, code)
	if err != nil {
		writeError(w, http.StatusBadGateway, "google_exchange_failed", "Could not sign in with Google")
		return
	}

	a.completeGoogleLogin(w, r, identity)
}

func validateGoogleState(r *http.Request) error {
	stateCookie, err := r.Cookie(googleStateCookieName)
	if err != nil || stateCookie.Value == "" {
		return errors.New("missing state cookie")
	}
	if subtleHash(stateCookie.Value) != subtleHash(r.URL.Query().Get("state")) {
		return errors.New("state mismatch")
	}
	return nil
}

func (a *App) completeGoogleLogin(w http.ResponseWriter, r *http.Request, identity GoogleIdentity) {
	email := normalizeEmail(identity.Email)
	if email == "" || !identity.EmailVerified {
		writeError(w, http.StatusBadRequest, "invalid_google_identity", "Verified Google email required")
		return
	}

	a.mu.Lock()
	user := a.byMail[email]
	if user == nil {
		user = &User{
			ID:           token(8),
			Email:        email,
			Verified:     true,
			Locale:       normalizeLocale(r.URL.Query().Get("locale")),
			Theme:        "light",
			Roles:        []string{"user"},
			PasswordHash: "",
			NotificationPreferences: NotificationPreferences{
				HydrationReminder: true,
			},
			AuthProvider: "google",
		}
		a.users[user.ID] = user
		a.byMail[user.Email] = user
	} else {
		user.Verified = true
		if user.AuthProvider == "" {
			user.AuthProvider = "google"
		}
	}

	rawSession := token(16)
	rawCSRF := token(16)
	user.SessionTokenHash = subtleHash(rawSession)
	user.CSRFTokenHash = subtleHash(rawCSRF)
	a.persistStateLocked()
	a.mu.Unlock()

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
	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	target := "/profile"
	if user.OnboardingDone {
		target = "/today"
	}
	http.Redirect(w, r, strings.TrimRight(a.frontendURL, "/")+target, http.StatusFound)
}
