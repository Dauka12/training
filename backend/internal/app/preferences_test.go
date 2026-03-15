package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNotificationPreferencesCanBeRead(t *testing.T) {
	_, server, client, _ := createVerifiedSession(t)
	defer server.Close()

	resp := doJSON(t, client, http.MethodGet, server.URL+"/api/v1/notifications/preferences", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected notification preferences ok, got %d", resp.StatusCode)
	}

	var payload struct {
		Preferences NotificationPreferences `json:"preferences"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Preferences.HydrationReminder {
		t.Fatal("expected hydration reminder preference enabled by default")
	}
}
