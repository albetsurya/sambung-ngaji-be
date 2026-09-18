package api

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

// AuditableActions — whitelist action yang perlu di-audit.
var AuditableActions = map[string]bool{
	"createMember":                   true,
	"updateMember":                   true,
	"deactivateMember":               true,
	"saveGroup":                      true,
	"createMeeting":                  true,
	"updateMeeting":                  true,
	"deleteMeeting":                  true,
	"deleteMeetingsBulk":             true,
	"bulkCreateMeetings":             true,
	"saveAttendance":                 true,
	"bulkSaveAttendance":             true,
	"deleteAttendance":               true,
	"deleteAttendanceByMeeting":      true,
	"deleteAttendanceByMember":       true,
	"createMonitoring":               true,
	"updateMonitoring":               true,
	"updateSettings":                 true,
	"createAnnouncementTemplate":     true,
	"updateAnnouncementTemplate":     true,
	"deleteAnnouncementTemplate":     true,
	"createTemplateFromAnnouncement": true,
	"createAnnouncement":             true,
	"updateAnnouncement":             true,
	"generateWeeklyAnnouncements":    true,
	"approvePendingMember":           true,
	"rejectPendingMember":            true,
	"createUser":                     true,
	"updateUser":                     true,
	"updateUserRole":                 true,
	"changeMyPassword":               true,
	"changeMyUsername":               true,
	"resetUserPassword":              true,
	"updateMyProfile":                true,
	"uploadPhoto":                    true,
	"deletePhoto":                    true,
	"setAIProvider":                  true,
}

var actionToTargetType = map[string]string{
	"createMember":         "member",
	"updateMember":         "member",
	"deactivateMember":     "member",
	"saveGroup":            "group",
	"createMeeting":        "meeting",
	"updateMeeting":        "meeting",
	"deleteMeeting":        "meeting",
	"deleteMeetingsBulk":   "meeting",
	"bulkCreateMeetings":   "meeting",
	"saveAttendance":       "attendance",
	"bulkSaveAttendance":   "attendance",
	"deleteAttendance":     "attendance",
	"createMonitoring":     "monitoring",
	"updateMonitoring":     "monitoring",
	"updateSettings":       "settings",
	"createAnnouncement":   "announcement",
	"updateAnnouncement":   "announcement",
	"approvePendingMember": "pending",
	"rejectPendingMember":  "pending",
	"createUser":           "user",
	"updateUser":           "user",
	"updateUserRole":       "user",
	"resetUserPassword":    "user",
	"changeMyUsername":     "user",
	"updateMyProfile":      "member",
	"uploadPhoto":          "member",
	"deletePhoto":          "member",
	"setAIProvider":        "settings",
}

// AuditMiddleware — auto-log action mutation yang sukses.
func AuditMiddleware(auditSvc *service.AuditService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			return err
		}

		success, _ := c.Locals(LocalsSuccess).(bool)
		if !success {
			return nil
		}

		action, _ := BodyOf(c)["action"].(string)
		if action == "" || !AuditableActions[action] {
			return nil
		}

		u := UserOf(c)
		if u == nil {
			return nil
		}

		targetID := extractTargetID(c, action)
		targetType := actionToTargetType[action]
		if targetType == "" {
			targetType = "unknown"
		}

		go func(userID, act, tt, tid string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			auditSvc.Log(ctx, userID, act, tt, tid)
		}(u.UserID, action, targetType, targetID)

		return nil
	}
}

func extractTargetID(c *fiber.Ctx, action string) string {
	body := BodyOf(c)
	candidates := []string{
		"member_id", "group_id", "meeting_id", "attendance_id",
		"monitoring_id", "user_id", "submission_id", "template_id",
		"announcement_id", "target_id",
	}
	for _, k := range candidates {
		if v, ok := body[k].(string); ok && v != "" {
			return v
		}
	}
	if strings.HasPrefix(action, "create") {
		return "(new)"
	}
	return ""
}
