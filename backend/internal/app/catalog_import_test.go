package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubCatalogImporter struct {
	equipment []ImportedEquipment
	exercises []ImportedExercise
}

func (s stubCatalogImporter) ImportWger(_ context.Context, limit int) ([]ImportedEquipment, []ImportedExercise, error) {
	if limit <= 0 {
		limit = len(s.exercises)
	}
	equipment := append([]ImportedEquipment(nil), s.equipment...)
	exercises := append([]ImportedExercise(nil), s.exercises...)
	if len(exercises) > limit {
		exercises = exercises[:limit]
	}
	return equipment, exercises, nil
}

func TestAdminCanImportCatalogFromExternalSource(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithCatalogImporter(stubCatalogImporter{
			equipment: []ImportedEquipment{
				{
					SourceID:     "3",
					Name:         "Dumbbell",
					Category:     "weights",
					LocationType: "mixed",
				},
			},
			exercises: []ImportedExercise{
				{
					SourceID:     "57",
					Slug:         "bear-walk",
					NameEN:       "Bear Walk",
					DescriptionEN:"Bodyweight crawl that challenges the whole body.",
					TechniqueEN:  "Move with opposite hand and foot while keeping the trunk braced.",
					Movement:     "conditioning",
					Difficulty:   "beginner",
					LocationType: "home",
					EquipmentIDs: []string{"wger-equipment-3"},
					MediaURL:     "https://wger.de/static/images/example/bear-walk.jpg",
				},
			},
		}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "admin-import@example.com", "admin")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")

	importResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/catalog/import/wger", map[string]any{
		"limit": 1,
	}, adminCSRF)
	if importResp.StatusCode != http.StatusOK {
		t.Fatalf("expected import ok, got %d", importResp.StatusCode)
	}

	var importPayload struct {
		Imported struct {
			Equipment int `json:"equipment"`
			Exercises int `json:"exercises"`
		} `json:"imported"`
	}
	if err := json.NewDecoder(importResp.Body).Decode(&importPayload); err != nil {
		t.Fatal(err)
	}
	if importPayload.Imported.Exercises != 1 {
		t.Fatalf("expected 1 imported exercise, got %+v", importPayload.Imported)
	}

	equipmentResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/catalog/equipment", nil, "")
	if equipmentResp.StatusCode != http.StatusOK {
		t.Fatalf("expected equipment list ok, got %d", equipmentResp.StatusCode)
	}
	var equipmentPayload struct {
		Items []EquipmentItem `json:"items"`
	}
	if err := json.NewDecoder(equipmentResp.Body).Decode(&equipmentPayload); err != nil {
		t.Fatal(err)
	}
	foundEquipment := false
	for _, item := range equipmentPayload.Items {
		if item.Names["ru"] == "Гантели" || item.Names["en"] == "Dumbbell" {
			foundEquipment = true
			break
		}
	}
	if !foundEquipment {
		t.Fatalf("expected imported equipment in catalog, got %+v", equipmentPayload.Items)
	}

	exerciseResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/admin/catalog/exercises", nil, "")
	if exerciseResp.StatusCode != http.StatusOK {
		t.Fatalf("expected exercise list ok, got %d", exerciseResp.StatusCode)
	}
	var exercisePayload struct {
		Items []ExerciseItem `json:"items"`
	}
	if err := json.NewDecoder(exerciseResp.Body).Decode(&exercisePayload); err != nil {
		t.Fatal(err)
	}
	foundExercise := false
	for _, item := range exercisePayload.Items {
		if item.Slug == "bear-walk" {
			foundExercise = true
			if item.MediaURL == "" {
				t.Fatalf("expected media url to be imported for exercise %+v", item)
			}
			break
		}
	}
	if !foundExercise {
		t.Fatalf("expected imported exercise in catalog, got %+v", exercisePayload.Items)
	}
}
