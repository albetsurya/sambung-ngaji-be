package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleCreateMember(c *fiber.Ctx, svc *service.MemberService) error {
	claims := ClaimsOf(c)
	groupID := BodyString(c, "group_id")
	groupName := ""
	if claims != nil && claims.Role != "SUPER_ADMIN" && claims.GroupID != nil && *claims.GroupID != "" {
		groupID = *claims.GroupID
		if grp, err := svc.GroupRepo().FindByID(c.Context(), *claims.GroupID); err == nil && grp != nil {
			groupName = grp.GroupName
		}
	}

	in := service.CreateMemberInput{
		GroupID:       groupID,
		FullName:   BodyString(c, "full_name"),
		Nickname: BodyString(c, "nickname"),
		Gender:  BodyString(c, "gender"),
		BirthPlace:   BodyString(c, "birth_place"),
		BirthDate:  BodyString(c, "birth_date"),
		GroupLabel: func() string {
			if groupName != "" {
				return groupName
			}
			return BodyString(c, "group_label")
		}(),
		Village:                   BodyString(c, "village"),
		Region:                 BodyString(c, "region"),
		HomeAddress:            BodyString(c, "home_address"),
		WhatsappNumber:                   BodyString(c, "whatsapp_number"),
		IsPreacher:            BodyBool(c, "is_preacher"),
		IsEmployed:                BodyBool(c, "is_employed"),
		IsMarried:                BodyBool(c, "is_married"),
		Height:            BodyString(c, "height"),
		Weight:             BodyString(c, "weight"),
		Hobby:                   BodyString(c, "hobby"),
		Occupation:              BodyString(c, "occupation"),
		PhotoURL:                BodyString(c, "photo_url"),
		MentoringStatus:        BodyString(c, "mentoring_status"),
		JoinedDate:           BodyString(c, "joined_date"),
		EducationLevel:      BodyString(c, "education_level"),
		School:                BodyString(c, "school"),
		Major:                BodyString(c, "major"),
		EducationStartYear:   BodyString(c, "education_start_year"),
		EducationEndYear: BodyString(c, "education_end_year"),
	}
	m, err := svc.Create(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, m)
}

func handleUpdateMember(c *fiber.Ctx, svc *service.MemberService) error {
	body := BodyOf(c)
	in := service.UpdateMemberInput{MemberID: BodyString(c, "member_id")}

	claims := ClaimsOf(c)
	if claims != nil && claims.Role != "SUPER_ADMIN" && claims.GroupID != nil && *claims.GroupID != "" {
		if grp, err := svc.GroupRepo().FindByID(c.Context(), *claims.GroupID); err == nil && grp != nil {
			body["group_label"] = grp.GroupName
			body["group_id"] = grp.GroupID
		}
	}

	pickStr := func(k string) *string {
		if v, ok := body[k].(string); ok {
			return &v
		}
		return nil
	}
	pickBool := func(k string) *bool {
		if v, ok := body[k].(bool); ok {
			return &v
		}
		return nil
	}

	in.FullName = pickStr("full_name")
	in.Nickname = pickStr("nickname")
	in.Gender = pickStr("gender")
	in.BirthPlace = pickStr("birth_place")
	in.BirthDate = pickStr("birth_date")
	in.GroupLabel = pickStr("group_label")
	in.GroupID = pickStr("group_id")
	in.Village = pickStr("village")
	in.Region = pickStr("region")
	in.HomeAddress = pickStr("home_address")
	in.WhatsappNumber = pickStr("whatsapp_number")
	in.IsPreacher = pickBool("is_preacher")
	in.IsEmployed = pickBool("is_employed")
	in.IsMarried = pickBool("is_married")
	in.Height = pickStr("height")
	in.Weight = pickStr("weight")
	in.Hobby = pickStr("hobby")
	in.Occupation = pickStr("occupation")
	in.PhotoURL = pickStr("photo_url")
	in.MentoringStatus = pickStr("mentoring_status")
	in.JoinedDate = pickStr("joined_date")
	in.LeftDate = pickStr("left_date")
	in.EducationLevel = pickStr("education_level")
	in.School = pickStr("school")
	in.Major = pickStr("major")
	in.EducationStartYear = pickStr("education_start_year")
	in.EducationEndYear = pickStr("education_end_year")

	m, err := svc.UpdateFull(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, m)
}

func handleDeleteMember(c *fiber.Ctx, svc *service.MemberService) error {
	if err := svc.DeleteMember(c.Context(), BodyString(c, "member_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func handleDeactivateMember(c *fiber.Ctx, svc *service.MemberService) error {
	if err := svc.Deactivate(c.Context(), BodyString(c, "member_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deactivated": true})
}

func handleGetMembersForExport(c *fiber.Ctx, svc *service.MemberService) error {
	items, err := svc.FindForExport(c.Context())
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}
