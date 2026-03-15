package plans

import (
	"testing"
	"time"
)

func TestFindNextValidReschedule(t *testing.T) {
	now := time.Date(2026, 3, 14, 8, 0, 0, 0, time.UTC)
	slots := []AvailabilitySlot{
		{Weekday: time.Sunday, DurationMin: 45},
		{Weekday: time.Monday, DurationMin: 60},
		{Weekday: time.Tuesday, DurationMin: 60},
	}

	next, ok := FindNextValidReschedule(now, time.Saturday, slots, map[time.Weekday]bool{
		time.Sunday: true,
	})
	if !ok {
		t.Fatal("expected slot found")
	}
	if next.Weekday() != time.Monday {
		t.Fatalf("expected monday, got %s", next.Weekday())
	}
}
