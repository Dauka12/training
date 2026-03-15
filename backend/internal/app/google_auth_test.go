package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type fakeGoogleProvider struct {
	authURL     string
	identity    GoogleIdentity
	exchangeErr error
}

func (f fakeGoogleProvider) Enabled() bool {
	return true
}

func (f fakeGoogleProvider) AuthURL(state string) string {
	return f.authURL + url.QueryEscape(state)
}

func (f fakeGoogleProvider) ExchangeCode(_ context.Context, _ string) (GoogleIdentity, error) {
	if f.exchangeErr != nil {
		return GoogleIdentity{}, f.exchangeErr
	}
	return f.identity, nil
}

func TestGoogleAuthCreatesVerifiedSessionAndRedirectsToProfile(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithGoogleAuthProvider(fakeGoogleProvider{
			authURL: "https://accounts.google.test/auth?state=",
			identity: GoogleIdentity{
				Email:         "google-user@example.com",
				EmailVerified: true,
				Name:          "Google User",
			},
		}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newNoRedirectClient(t)

	startResp, err := client.Get(server.URL + "/api/v1/auth/google/start")
	if err != nil {
		t.Fatal(err)
	}
	if startResp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 start response, got %d", startResp.StatusCode)
	}

	stateCookie := cookieValueFromJar(t, client, server.URL, "google_oauth_state")
	if stateCookie == "" {
		t.Fatal("expected google_oauth_state cookie")
	}

	callbackResp, err := client.Get(server.URL + "/api/v1/auth/google/callback?code=fake-code&state=" + url.QueryEscape(stateCookie))
	if err != nil {
		t.Fatal(err)
	}
	if callbackResp.StatusCode != http.StatusFound {
		t.Fatalf("expected 302 callback response, got %d", callbackResp.StatusCode)
	}
	if location := callbackResp.Header.Get("Location"); !strings.HasSuffix(location, "/profile") {
		t.Fatalf("expected redirect to /profile, got %q", location)
	}

	meResp, err := client.Get(server.URL + "/api/v1/me")
	if err != nil {
		t.Fatal(err)
	}
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected authenticated me response, got %d", meResp.StatusCode)
	}

	app.mu.Lock()
	defer app.mu.Unlock()
	user := app.byMail["google-user@example.com"]
	if user == nil {
		t.Fatal("expected google user to be created")
	}
	if !user.Verified {
		t.Fatal("expected google user to be verified")
	}
}

func TestGoogleAuthRejectsUnverifiedIdentity(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithGoogleAuthProvider(fakeGoogleProvider{
			authURL: "https://accounts.google.test/auth?state=",
			identity: GoogleIdentity{
				Email:         "unverified@example.com",
				EmailVerified: false,
			},
		}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newNoRedirectClient(t)
	if _, err := client.Get(server.URL + "/api/v1/auth/google/start"); err != nil {
		t.Fatal(err)
	}
	stateCookie := cookieValueFromJar(t, client, server.URL, "google_oauth_state")

	callbackResp, err := client.Get(server.URL + "/api/v1/auth/google/callback?code=fake-code&state=" + url.QueryEscape(stateCookie))
	if err != nil {
		t.Fatal(err)
	}
	if callbackResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 callback response, got %d", callbackResp.StatusCode)
	}
}

func TestGoogleAuthHandlesExchangeFailure(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithGoogleAuthProvider(fakeGoogleProvider{
			authURL:     "https://accounts.google.test/auth?state=",
			exchangeErr: errors.New("exchange failed"),
		}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	client := newNoRedirectClient(t)
	if _, err := client.Get(server.URL + "/api/v1/auth/google/start"); err != nil {
		t.Fatal(err)
	}
	stateCookie := cookieValueFromJar(t, client, server.URL, "google_oauth_state")

	callbackResp, err := client.Get(server.URL + "/api/v1/auth/google/callback?code=fake-code&state=" + url.QueryEscape(stateCookie))
	if err != nil {
		t.Fatal(err)
	}
	if callbackResp.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected 502 callback response, got %d", callbackResp.StatusCode)
	}
}

func newNoRedirectClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func cookieValueFromJar(t *testing.T, client *http.Client, rawURL, name string) string {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range client.Jar.Cookies(parsed) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}
