package model

import "time"

type MemberMood struct {
	MoodID    string
	MemberID  string
	MoodKey   string
	Date   time.Time
	CreatedAt time.Time
}

type MemberMoodDTO struct {
	MoodID    string `json:"mood_id"`
	MemberID  string `json:"member_id"`
	MoodKey   string `json:"mood_key"`
	Date   string `json:"date"`
	CreatedAt string `json:"created_at"`
}
