package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleCheckUsernameAvailability(c *fiber.Ctx, svc *service.PendingService) error {
	username := BodyString(c, "username")
	res, err := svc.CheckUsername(c.Context(), username)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleSubmitPublicRegistration(c *fiber.Ctx, svc *service.PendingService) error {
	clientIP := c.IP()
	if v := c.Get("X-Forwarded-For"); v != "" {
		clientIP = v
	}
	in := service.SubmitRegistrationInput{
		GroupID:                BodyString(c, "group_id"),
		NamaLengkap:            BodyString(c, "nama_lengkap"),
		NamaPanggilan:          BodyString(c, "nama_panggilan"),
		JenisKelamin:           BodyString(c, "jenis_kelamin"),
		TempatLahir:            BodyString(c, "tempat_lahir"),
		TanggalLahir:           BodyString(c, "tanggal_lahir"),
		NoWA:                   BodyString(c, "no_wa"),
		AlamatRumah:            BodyString(c, "alamat_rumah"),
		Desa:                   BodyString(c, "desa"),
		Daerah:                 BodyString(c, "daerah"),
		Pekerjaan:              BodyString(c, "pekerjaan"),
		Hobi:                   BodyString(c, "hobi"),
		IsNikah:                BodyBool(c, "is_nikah"),
		JenjangPendidikan:      BodyString(c, "jenjang_pendidikan"),
		Sekolah:                BodyString(c, "sekolah"),
		Jurusan:                BodyString(c, "jurusan"),
		TahunMulaiPendidikan:   BodyString(c, "tahun_mulai_pendidikan"),
		TahunSelesaiPendidikan: BodyString(c, "tahun_selesai_pendidikan"),
		FotoURL:                BodyString(c, "foto_url"),
		Username:               BodyString(c, "username"),
		Password:               BodyString(c, "password"),
		ClientIP:               clientIP,
	}
	res, err := svc.SubmitRegistration(c.Context(), in)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleGetPendingMembers(c *fiber.Ctx, svc *service.PendingService) error {
	groupID := BodyString(c, "group_id")
	status := BodyString(c, "status")
	items, err := svc.GetPendingMembers(c.Context(), groupID, status)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, items)
}

func handleGetPendingMemberDetail(c *fiber.Ctx, svc *service.PendingService) error {
	id := BodyString(c, "submission_id")
	dto, err := svc.GetPendingMemberDetail(c.Context(), id)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, dto)
}

func handleApprovePendingMember(c *fiber.Ctx, svc *service.PendingService) error {
	u := UserOf(c)
	reviewerID := ""
	if u != nil {
		reviewerID = u.UserID
	}
	res, err := svc.Approve(c.Context(),
		BodyString(c, "submission_id"),
		BodyString(c, "kelompok"),
		reviewerID,
	)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleRejectPendingMember(c *fiber.Ctx, svc *service.PendingService) error {
	u := UserOf(c)
	reviewerID := ""
	if u != nil {
		reviewerID = u.UserID
	}
	if err := svc.Reject(c.Context(),
		BodyString(c, "submission_id"),
		reviewerID,
		BodyString(c, "reason"),
	); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{
		"submission_id": BodyString(c, "submission_id"),
		"status":        "REJECTED",
	})
}
