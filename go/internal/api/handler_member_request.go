package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleRequestBecomeMember(c *fiber.Ctx, svc *service.MemberRequestService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	res, err := svc.Request(c.Context(), u)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleListMemberRequests(c *fiber.Ctx, svc *service.MemberRequestService) error {
	status := BodyString(c, "status")
	/* Akun ber-group_label hanya melihat request user kelompoknya. */
	groupID := ""
	if g, isSuper := ActorOf(c); !isSuper && g != "" && g != UnassignedGroup {
		groupID = g
	}
	res, err := svc.List(c.Context(), status, groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleApproveMemberRequest(c *fiber.Ctx, svc *service.MemberRequestService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	res, err := svc.Approve(c.Context(), service.ApproveMemberRequestInput{
		RequestID:  BodyString(c, "request_id"),
		ReviewerID: u.UserID,
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleRejectMemberRequest(c *fiber.Ctx, svc *service.MemberRequestService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	err := svc.Reject(c.Context(), BodyString(c, "request_id"), u.UserID, BodyString(c, "reason"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, map[string]bool{"rejected": true})
}
