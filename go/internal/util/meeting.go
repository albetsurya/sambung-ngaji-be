package util

import (
	"strings"

	"pengajian-backend/internal/model"
)

const MeetingStatusLibur = "LIBUR"

func IsLiburMeeting(m model.Meeting) bool {
	return strings.EqualFold(m.Status, MeetingStatusLibur)
}

func LiburMeetingIDs(meetings []model.Meeting) map[string]bool {
	out := make(map[string]bool, len(meetings))
	for _, m := range meetings {
		if IsLiburMeeting(m) {
			out[m.MeetingID] = true
		}
	}
	return out
}
