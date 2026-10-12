package handlers

import (
	"testing"
	"time"
)

func TestCalculateStreakUsesConsecutiveLocalCompletionDates(t *testing.T) {
	location := time.FixedZone("test", 5*60*60+30*60)
	today := time.Date(2026, 10, 12, 0, 0, 0, 0, location)
	completions := []time.Time{
		time.Date(2026, 10, 8, 10, 0, 0, 0, location),
		time.Date(2026, 10, 9, 10, 0, 0, 0, location),
		time.Date(2026, 10, 9, 16, 0, 0, 0, location),
		time.Date(2026, 10, 11, 10, 0, 0, 0, location),
		time.Date(2026, 10, 12, 10, 0, 0, 0, location),
		time.Date(2026, 10, 13, 10, 0, 0, 0, location),
	}

	current, longest, last := calculateStreak(completions, today)
	if current != 2 {
		t.Errorf("current streak = %d, want 2", current)
	}
	if longest != 2 {
		t.Errorf("longest streak = %d, want 2", longest)
	}
	if last != "2026-10-12" {
		t.Errorf("last study date = %q, want 2026-10-12", last)
	}
}

func TestCalculateStreakExpiresAfterMissedDay(t *testing.T) {
	location := time.UTC
	today := time.Date(2026, 10, 12, 0, 0, 0, 0, location)
	completions := []time.Time{time.Date(2026, 10, 9, 8, 0, 0, 0, location)}

	current, longest, _ := calculateStreak(completions, today)
	if current != 0 {
		t.Errorf("current streak = %d, want 0 after missed days", current)
	}
	if longest != 1 {
		t.Errorf("longest streak = %d, want 1", longest)
	}
}

func TestCalculateStreakCanCarryFromYesterday(t *testing.T) {
	today := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	completions := []time.Time{
		time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 11, 8, 0, 0, 0, time.UTC),
	}
	current, _, _ := calculateStreak(completions, today)
	if current != 2 {
		t.Errorf("current streak = %d, want 2", current)
	}
}

func TestCalendarDateOffsetCountsCalendarDays(t *testing.T) {
	location := time.FixedZone("test", 5*60*60+30*60)
	start := time.Date(2026, 10, 12, 0, 0, 0, 0, location)
	day := time.Date(2026, 10, 18, 23, 30, 0, 0, location)
	if offset := calendarDateOffset(day, start); offset != 6 {
		t.Errorf("calendar day offset = %d, want 6", offset)
	}
}
