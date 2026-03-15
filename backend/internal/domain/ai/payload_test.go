package ai

import "testing"

func TestBuildPrivacySafePayloadOmitsPII(t *testing.T) {
	input := PayloadInput{
		UserRef: "internal-123",
		Email:   "user@example.com",
		Locale:  "ru",
		Targets: Targets{
			DailyCalories: 2200,
			ProteinG:      160,
			CarbsG:        220,
			FatG:          70,
			DailyWaterML:  2600,
		},
	}

	payload := BuildPrivacySafePayload(input)
	if payload["email"] != nil {
		t.Fatal("expected email omitted")
	}
	if payload["user_ref"] != "internal-123" {
		t.Fatalf("expected user_ref kept, got %v", payload["user_ref"])
	}
}
