package service

import (
	"testing"
	"time"
)

func TestNextStreakUsesCalendarDays(t *testing.T) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	if got := nextStreak(4, &yesterday, dateOnly(time.Now().UTC())); got != 5 {
		t.Fatalf("expected consecutive calendar day to extend streak, got %d", got)
	}
	twoDaysAgo := time.Now().UTC().AddDate(0, 0, -2)
	if got := nextStreak(4, &twoDaysAgo, dateOnly(time.Now().UTC())); got != 1 {
		t.Fatalf("expected missed day to reset streak, got %d", got)
	}
}

func TestValidateNgajiEventRejectsUnknownAchievement(t *testing.T) {
	err := validateNgajiEvent("user-1", NgajiProgressEvent{
		EventID: "event-1", GameID: "ayat-hunter", Score: 100, BaseXP: 10,
		AchievementIDs: []string{"not-registered"},
	})
	if err == nil {
		t.Fatal("expected unknown achievement to be rejected")
	}
}
