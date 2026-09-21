package util

import (
	"strings"

	"pengajian-backend/internal/model"
)

// MeetingStatusLibur adalah nilai kolom status meeting untuk jadwal libur.
const MeetingStatusLibur = "LIBUR"

// IsLiburMeeting mengembalikan true jika meeting berstatus LIBUR.
// Case-insensitive — toleran terhadap variasi data.
func IsLiburMeeting(m model.Meeting) bool {
	return strings.EqualFold(m.Status, MeetingStatusLibur)
}

// LiburMeetingIDs mengembalikan set meeting_id yang berstatus LIBUR.
// Berguna untuk filtering cepat di loop attendance.
func LiburMeetingIDs(meetings []model.Meeting) map[string]bool {
	out := make(map[string]bool, len(meetings))
	for _, m := range meetings {
		if IsLiburMeeting(m) {
			out[m.MeetingID] = true
		}
	}
	return out
}
