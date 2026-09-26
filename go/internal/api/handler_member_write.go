package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func handleCreateMember(c *fiber.Ctx, svc *service.MemberService) error {
	claims := ClaimsOf(c)
	groupName := ""
	if claims != nil && claims.Role != "SUPER_ADMIN" && claims.GroupID != nil && *claims.GroupID != "" {
		if grp, err := svc.GroupRepo().FindByID(c.Context(), *claims.GroupID); err == nil && grp != nil {
			groupName = grp.GroupName
		}
	}

	in := service.CreateMemberInput{
		NamaLengkap:            BodyString(c, "nama_lengkap"),
		NamaPanggilan:          BodyString(c, "nama_panggilan"),
		JenisKelamin:           BodyString(c, "jenis_kelamin"),
		TempatLahir:            BodyString(c, "tempat_lahir"),
		TanggalLahir:           BodyString(c, "tanggal_lahir"),
		Kelompok:               func() string { if groupName != "" { return groupName }; return BodyString(c, "kelompok") }(),
		Desa:                   BodyString(c, "desa"),
		Daerah:                 BodyString(c, "daerah"),
		AlamatRumah:            BodyString(c, "alamat_rumah"),
		NoWA:                   BodyString(c, "no_wa"),
		IsMuballigh:            BodyBool(c, "is_muballigh"),
		IsKerja:                BodyBool(c, "is_kerja"),
		IsNikah:                BodyBool(c, "is_nikah"),
		TinggiBadan:            BodyString(c, "tinggi_badan"),
		BeratBadan:             BodyString(c, "berat_badan"),
		Hobi:                   BodyString(c, "hobi"),
		Pekerjaan:              BodyString(c, "pekerjaan"),
		FotoURL:                BodyString(c, "foto_url"),
		StatusPembinaan:        BodyString(c, "status_pembinaan"),
		TanggalMasuk:           BodyString(c, "tanggal_masuk"),
		JenjangPendidikan:      BodyString(c, "jenjang_pendidikan"),
		Sekolah:                BodyString(c, "sekolah"),
		Jurusan:                BodyString(c, "jurusan"),
		TahunMulaiPendidikan:   BodyString(c, "tahun_mulai_pendidikan"),
		TahunSelesaiPendidikan: BodyString(c, "tahun_selesai_pendidikan"),
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
			body["kelompok"] = grp.GroupName
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

	in.NamaLengkap = pickStr("nama_lengkap")
	in.NamaPanggilan = pickStr("nama_panggilan")
	in.JenisKelamin = pickStr("jenis_kelamin")
	in.TempatLahir = pickStr("tempat_lahir")
	in.TanggalLahir = pickStr("tanggal_lahir")
	in.Kelompok = pickStr("kelompok")
	in.Desa = pickStr("desa")
	in.Daerah = pickStr("daerah")
	in.AlamatRumah = pickStr("alamat_rumah")
	in.NoWA = pickStr("no_wa")
	in.IsMuballigh = pickBool("is_muballigh")
	in.IsKerja = pickBool("is_kerja")
	in.IsNikah = pickBool("is_nikah")
	in.TinggiBadan = pickStr("tinggi_badan")
	in.BeratBadan = pickStr("berat_badan")
	in.Hobi = pickStr("hobi")
	in.Pekerjaan = pickStr("pekerjaan")
	in.FotoURL = pickStr("foto_url")
	in.StatusPembinaan = pickStr("status_pembinaan")
	in.TanggalMasuk = pickStr("tanggal_masuk")
	in.TanggalKeluar = pickStr("tanggal_keluar")
	in.JenjangPendidikan = pickStr("jenjang_pendidikan")
	in.Sekolah = pickStr("sekolah")
	in.Jurusan = pickStr("jurusan")
	in.TahunMulaiPendidikan = pickStr("tahun_mulai_pendidikan")
	in.TahunSelesaiPendidikan = pickStr("tahun_selesai_pendidikan")

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
