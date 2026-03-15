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

func TestAdminCanPreviewImportAndUpdateExerciseMediaAndTechnique(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithCatalogImporter(stubCatalogImporter{
			equipment: []ImportedEquipment{
				{
					SourceID:     "9",
					Name:         "Kettlebell",
					Category:     "weights",
					LocationType: "mixed",
				},
			},
			exercises: []ImportedExercise{
				{
					SourceID:      "90",
					Slug:          "kettlebell-swing",
					NameEN:        "Kettlebell Swing",
					DescriptionEN: "Hip hinge based power exercise.",
					TechniqueEN:   "Drive the kettlebell with the hips.",
					Movement:      "hinge",
					Difficulty:    "intermediate",
					LocationType:  "mixed",
					EquipmentIDs:  []string{"wger-equipment-9"},
					MediaURL:      "https://wger.de/static/images/example/kettlebell-swing.jpg",
				},
			},
		}),
	)
	server := httptest.NewServer(app.Routes())
	defer server.Close()

	adminClient, _ := createUserWithRole(t, app, server.URL, "ops-admin@example.com", "admin")
	adminCSRF := cookieValue(t, adminClient, server.URL, "csrf")

	previewResp := doJSON(t, adminClient, http.MethodPost, server.URL+"/api/v1/admin/catalog/import/wger/preview", map[string]any{
		"limit": 1,
	}, adminCSRF)
	if previewResp.StatusCode != http.StatusOK {
		t.Fatalf("expected preview ok, got %d", previewResp.StatusCode)
	}

	var previewPayload struct {
		Preview struct {
			Equipment []ImportedEquipment `json:"equipment"`
			Exercises []ImportedExercise  `json:"exercises"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(previewResp.Body).Decode(&previewPayload); err != nil {
		t.Fatal(err)
	}
	if len(previewPayload.Preview.Exercises) != 1 || previewPayload.Preview.Exercises[0].MediaURL == "" {
		t.Fatalf("expected preview exercise with media, got %+v", previewPayload.Preview)
	}

	updateResp := doJSON(t, adminClient, http.MethodPut, server.URL+"/api/v1/admin/catalog/exercises/20000000-0000-0000-0000-000000000001", map[string]any{
		"slug":                  "goblet-squat",
		"names":                 map[string]string{"ru": "Присед с гантелью", "kk": "Gantelmen otyru"},
		"descriptions":          map[string]string{"ru": "Обновленное описание", "kk": "Jangartylgan sipattama"},
		"technique":             map[string]string{"ru": "Локти смотрят вниз, корпус стабилен", "kk": "Shyntaq tomenge, dene turaqty"},
		"movement_pattern":      "squat",
		"difficulty":            "beginner",
		"location_type":         "mixed",
		"equipment_ids":         []string{"10000000-0000-0000-0000-000000000001"},
		"media_url":             "https://example.com/updated-goblet-squat.jpg",
		"contraindication_tags": []string{"knee_pain", "low_back_sensitivity"},
		"substitution_ids":      []string{"20000000-0000-0000-0000-000000000002"},
		"active":                true,
	}, adminCSRF)
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("expected exercise update ok, got %d", updateResp.StatusCode)
	}

	detailResp := doJSON(t, adminClient, http.MethodGet, server.URL+"/api/v1/catalog/exercises/20000000-0000-0000-0000-000000000001", nil, "")
	if detailResp.StatusCode != http.StatusOK {
		t.Fatalf("expected exercise detail ok, got %d", detailResp.StatusCode)
	}
	var detailPayload struct {
		Exercise struct {
			MediaURL             string   `json:"media_url"`
			Technique            string   `json:"technique"`
			ContraindicationTags []string `json:"contraindication_tags"`
			Substitutions        []struct {
				ID string `json:"id"`
			} `json:"substitutions"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(detailResp.Body).Decode(&detailPayload); err != nil {
		t.Fatal(err)
	}
	if detailPayload.Exercise.MediaURL != "https://example.com/updated-goblet-squat.jpg" {
		t.Fatalf("expected updated media url, got %+v", detailPayload.Exercise)
	}
	if len(detailPayload.Exercise.ContraindicationTags) != 2 {
		t.Fatalf("expected updated contraindications, got %+v", detailPayload.Exercise)
	}
	if len(detailPayload.Exercise.Substitutions) != 1 {
		t.Fatalf("expected updated substitutions, got %+v", detailPayload.Exercise)
	}
}
