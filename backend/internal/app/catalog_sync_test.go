package app

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func TestRunCatalogSyncOnceImportsAndAudits(t *testing.T) {
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithCatalogImporter(stubCatalogImporter{
			equipment: []ImportedEquipment{
				{SourceID: "50", Name: "Pull-up Bar", Category: "bodyweight", LocationType: "home"},
			},
			exercises: []ImportedExercise{
				{
					SourceID:      "501",
					Slug:          "pull-up",
					NameEN:        "Pull Up",
					DescriptionEN: "Upper body pulling exercise.",
					TechniqueEN:   "Start from a dead hang and pull to the bar.",
					Movement:      "pull",
					Difficulty:    "intermediate",
					LocationType:  "home",
					EquipmentIDs:  []string{"wger-equipment-50"},
					MediaURL:      "https://wger.de/static/images/example/pull-up.jpg",
				},
			},
		}),
	)

	if err := app.RunCatalogSyncOnce(context.Background(), 3); err != nil {
		t.Fatalf("expected sync ok, got %v", err)
	}

	app.mu.Lock()
	defer app.mu.Unlock()

	foundExercise := false
	for _, item := range app.exerciseCatalog {
		if item.ID == "wger-exercise-501" {
			foundExercise = true
			break
		}
	}
	if !foundExercise {
		t.Fatalf("expected synced exercise in catalog, got %+v", app.exerciseCatalog)
	}
	if len(app.auditLogs) == 0 || app.auditLogs[len(app.auditLogs)-1].Action != "sync_catalog_wger" {
		t.Fatalf("expected sync audit log, got %+v", app.auditLogs)
	}
}
