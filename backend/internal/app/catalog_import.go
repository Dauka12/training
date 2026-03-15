package app

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type CatalogImporter interface {
	ImportWger(ctx context.Context, limit int) ([]ImportedEquipment, []ImportedExercise, error)
}

type ImportedEquipment struct {
	SourceID     string
	Name         string
	Description  string
	Category     string
	LocationType string
}

type ImportedExercise struct {
	SourceID      string
	Slug          string
	NameEN        string
	DescriptionEN string
	TechniqueEN   string
	Movement      string
	Difficulty    string
	LocationType  string
	EquipmentIDs  []string
	MediaURL      string
}

type WgerCatalogImporter struct {
	Client  *http.Client
	BaseURL string
}

func NewWgerCatalogImporter(client *http.Client, baseURL string) WgerCatalogImporter {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://wger.de/api/v2"
	}
	return WgerCatalogImporter{
		Client:  client,
		BaseURL: baseURL,
	}
}

func (w WgerCatalogImporter) ImportWger(ctx context.Context, limit int) ([]ImportedEquipment, []ImportedExercise, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 24 {
		limit = 24
	}

	equipment, err := w.fetchEquipment(ctx)
	if err != nil {
		return nil, nil, err
	}

	exercises, err := w.fetchExercises(ctx, limit)
	if err != nil {
		return nil, nil, err
	}

	return equipment, exercises, nil
}

type wgerEquipmentResponse struct {
	Results []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"results"`
}

type wgerExerciseResponse struct {
	Results []struct {
		ID        int `json:"id"`
		Category  struct {
			Name string `json:"name"`
		} `json:"category"`
		Equipment []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"equipment"`
		Images []struct {
			Image string `json:"image"`
		} `json:"images"`
		Translations []struct {
			Language    int    `json:"language"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"translations"`
	} `json:"results"`
}

func (w WgerCatalogImporter) fetchEquipment(ctx context.Context) ([]ImportedEquipment, error) {
	var payload wgerEquipmentResponse
	if err := w.getJSON(ctx, w.BaseURL+"/equipment/?limit=50", &payload); err != nil {
		return nil, fmt.Errorf("fetch wger equipment: %w", err)
	}

	items := make([]ImportedEquipment, 0, len(payload.Results))
	for _, item := range payload.Results {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		items = append(items, ImportedEquipment{
			SourceID:     fmt.Sprintf("wger-equipment-%d", item.ID),
			Name:         name,
			Description:  name,
			Category:     "imported",
			LocationType: locationTypeForEquipment(name),
		})
	}
	return items, nil
}

func (w WgerCatalogImporter) fetchExercises(ctx context.Context, limit int) ([]ImportedExercise, error) {
	var payload wgerExerciseResponse
	if err := w.getJSON(ctx, fmt.Sprintf("%s/exerciseinfo/?limit=%d", w.BaseURL, limit), &payload); err != nil {
		return nil, fmt.Errorf("fetch wger exercises: %w", err)
	}

	items := make([]ImportedExercise, 0, len(payload.Results))
	for _, item := range payload.Results {
		translation := preferredTranslation(item.Translations)
		name := strings.TrimSpace(translation.Name)
		if name == "" {
			continue
		}
		description := cleanHTML(translation.Description)
		mediaURL := ""
		if len(item.Images) > 0 {
			mediaURL = strings.TrimSpace(item.Images[0].Image)
		}

		equipmentIDs := make([]string, 0, len(item.Equipment))
		locationType := "mixed"
		if len(item.Equipment) == 0 {
			locationType = "home"
		}
		for _, equipment := range item.Equipment {
			equipmentIDs = append(equipmentIDs, fmt.Sprintf("wger-equipment-%d", equipment.ID))
			if strings.Contains(strings.ToLower(equipment.Name), "none") {
				locationType = "home"
			}
		}

		items = append(items, ImportedExercise{
			SourceID:      fmt.Sprintf("wger-exercise-%d", item.ID),
			Slug:          slugify(name),
			NameEN:        name,
			DescriptionEN: trimSentence(description, 220),
			TechniqueEN:   trimSentence(description, 320),
			Movement:      fallbackString(slugify(item.Category.Name), "general"),
			Difficulty:    "beginner",
			LocationType:  locationType,
			EquipmentIDs:  equipmentIDs,
			MediaURL:      mediaURL,
		})
	}
	return items, nil
}

func (w WgerCatalogImporter) getJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := w.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func preferredTranslation(items []struct {
	Language    int    `json:"language"`
	Name        string `json:"name"`
	Description string `json:"description"`
}) struct {
	Language    int    `json:"language"`
	Name        string `json:"name"`
	Description string `json:"description"`
} {
	for _, item := range items {
		if item.Language == 2 && strings.TrimSpace(item.Name) != "" {
			return item
		}
	}
	for _, item := range items {
		if strings.TrimSpace(item.Name) != "" {
			return item
		}
	}
	return struct {
		Language    int    `json:"language"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}{}
}

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)
var slugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func cleanHTML(value string) string {
	value = html.UnescapeString(value)
	value = htmlTagPattern.ReplaceAllString(value, " ")
	value = strings.Join(strings.Fields(value), " ")
	return strings.TrimSpace(value)
}

func trimSentence(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) <= limit {
		return value
	}
	return strings.TrimSpace(value[:limit]) + "..."
}

func slugify(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = slugPattern.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "imported-item"
	}
	return normalized
}

func locationTypeForEquipment(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(lower, "bench"), strings.Contains(lower, "barbell"), strings.Contains(lower, "machine"):
		return "gym"
	case strings.Contains(lower, "none"), strings.Contains(lower, "bodyweight"), strings.Contains(lower, "mat"):
		return "home"
	default:
		return "mixed"
	}
}
