package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleGetUsers(c *fiber.Ctx, svc *service.UserService) error {
	items, err := svc.GetUsers(c.Context())
	if err != nil {
		return Fail(c, "Gagal ambil users: "+err.Error())
	}
	return Ok(c, items)
}

func handleGetUserDetail(c *fiber.Ctx, svc *service.UserService) error {
	id := BodyString(c, "user_id")
	dto, err := svc.GetUserDetail(c.Context(), id)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleCreateUser(c *fiber.Ctx, svc *service.UserService) error {
	u := UserOf(c)
	adminID := ""
	if u != nil {
		adminID = u.UserID
	}
	in := service.CreateUserInput{
		Username: BodyString(c, "username"),
		Password: BodyString(c, "password"),
		Role:     BodyString(c, "role"),
		Nama:     BodyString(c, "nama"),
		MemberID: BodyString(c, "member_id"),
		AdminID:  adminID,
	}
	dto, err := svc.CreateUser(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateUser(c *fiber.Ctx, svc *service.UserService) error {
	body := BodyOf(c)
	in := service.UpdateUserInput{
		UserID: BodyString(c, "user_id"),
	}
	if v, ok := body["nama"].(string); ok {
		in.Nama = &v
	}
	if v, ok := body["role"].(string); ok {
		in.Role = &v
	}
	if v, ok := body["status_aktif"].(bool); ok {
		in.StatusAktif = &v
	}
	if v, ok := body["password"].(string); ok && v != "" {
		in.Password = &v
	}
	dto, err := svc.UpdateUser(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleUpdateUserRole(c *fiber.Ctx, svc *service.UserService) error {
	in := service.UpdateRoleInput{
		UserID: BodyString(c, "user_id"),
		Role:   BodyString(c, "role"),
	}
	dto, err := svc.UpdateUserRole(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleGetMemberUserStatus(c *fiber.Ctx, svc *service.UserService) error {
	res, err := svc.GetMemberUserStatus(c.Context(), BodyString(c, "member_id"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleChangeMyPassword(c *fiber.Ctx, svc *service.UserService) error {
	u := UserOf(c)
	if u == nil {
		return Fail(c, "Unauthorized")
	}
	in := service.ChangePasswordInput{
		UserID:      u.UserID,
		OldPassword: BodyString(c, "old_password"),
		NewPassword: BodyString(c, "new_password"),
	}
	if err := svc.ChangeMyPassword(c.Context(), in); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"changed": true})
}

func handleResetUserPassword(c *fiber.Ctx, svc *service.UserService) error {
	if err := svc.ResetPassword(c.Context(),
		BodyString(c, "user_id"),
		BodyString(c, "new_password"),
	); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"reset": true})
}
