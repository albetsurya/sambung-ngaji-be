package service

import (
	"context"
	"strings"
	"time"

	apperrors "pengajian-backend/internal/errors"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

func (s *MemberService) resolveGroup(ctx context.Context, groupID, kelompok string) (id, name string, err error) {
	groupID = strings.TrimSpace(groupID)
	kelompok = strings.TrimSpace(kelompok)
	if groupID != "" {
		if s.groupRepo == nil {
			return groupID, kelompok, nil
		}
		g, err := s.groupRepo.FindByID(ctx, groupID)
		if err != nil {
			return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok tidak dikenal")
		}
		if kelompok == "" {
			kelompok = g.GroupName
		}
		return g.GroupID, kelompok, nil
	}
	if kelompok == "" {
		return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok wajib dipilih")
	}
	if s.groupRepo == nil {
		return "", kelompok, nil
	}
	g, err := s.groupRepo.FindByName(ctx, kelompok)
	if err != nil {
		return "", "", apperrors.Wrap(apperrors.ErrValidation, "kelompok tidak dikenal: "+kelompok)
	}
	return g.GroupID, g.GroupName, nil
}

type CreateMemberInput struct {
	GroupID                string
	NamaLengkap            string
	NamaPanggilan          string
	JenisKelamin           string
	TempatLahir            string
	TanggalLahir           string
	Kelompok               string
	Desa                   string
	Daerah                 string
	AlamatRumah            string
	NoWA                   string
	IsMuballigh            bool
	IsKerja                bool
	IsNikah                bool
	TinggiBadan            string
	BeratBadan             string
	Hobi                   string
	Pekerjaan              string
	FotoURL                string
	StatusPembinaan        string
	TanggalMasuk           string
	JenjangPendidikan      string
	Sekolah                string
	Jurusan                string
	TahunMulaiPendidikan   string
	TahunSelesaiPendidikan string
}

func (s *MemberService) Create(ctx context.Context, in CreateMemberInput) (*model.MemberDetailDTO, error) {

	in.NamaLengkap = util.TitleCaseID(in.NamaLengkap)
	in.NamaPanggilan = util.TitleCaseID(in.NamaPanggilan)
	in.TempatLahir = util.TitleCaseID(in.TempatLahir)
	in.Desa = util.TitleCaseID(in.Desa)
	in.Daerah = util.TitleCaseID(in.Daerah)
	if strings.TrimSpace(in.NamaLengkap) == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "nama_lengkap wajib diisi")
	}

	jk := normalizeGender(in.JenisKelamin)
	var jkPtr *string
	if jk != "" {
		jkPtr = &jk
	}

	var tglLahir interface{}
	if in.TanggalLahir != "" {
		tglLahir = in.TanggalLahir
	}

	status := in.StatusPembinaan
	if status == "" {
		status = "AKTIF"
	}
	tglMasuk := in.TanggalMasuk
	if tglMasuk == "" {
		tglMasuk = time.Now().Format("2006-01-02")
	}

	noWA := ""
	if in.NoWA != "" {
		noWA = util.NormalizePhone(in.NoWA)
	}

	groupID, kelompok, err := s.resolveGroup(ctx, in.GroupID, in.Kelompok)
	if err != nil {
		return nil, err
	}

	memberID := util.NewID("MBR")
	_, err = s.repo.Pool().Exec(ctx, `
		INSERT INTO members (
			member_id, nama_lengkap, nama_panggilan, jenis_kelamin,
			tempat_lahir, tanggal_lahir, foto_url, no_wa,
			alamat_rumah, desa, daerah, kelompok, group_id,
			is_muballigh, is_kerja, is_nikah, tinggi_badan, berat_badan,
			hobi, pekerjaan, status_pembinaan, status_aktif, tanggal_masuk,
			jenjang_pendidikan, sekolah, jurusan,
			tahun_mulai_pendidikan, tahun_selesai_pendidikan,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
			$14,$15,$16,$17,$18,$19,$20,$21,true,$22,
			$23,$24,$25,$26,$27,now(),now()
		)
	`,
		memberID, in.NamaLengkap, in.NamaPanggilan, jkPtr,
		in.TempatLahir, tglLahir, in.FotoURL, noWA,
		in.AlamatRumah, in.Desa, in.Daerah, kelompok, groupID,
		in.IsMuballigh, in.IsKerja, in.IsNikah, in.TinggiBadan, in.BeratBadan,
		in.Hobi, in.Pekerjaan, status, tglMasuk,
		in.JenjangPendidikan, in.Sekolah, in.Jurusan,
		in.TahunMulaiPendidikan, in.TahunSelesaiPendidikan,
	)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	dto := s.toDetailDTO(m)
	return &dto, nil
}

type UpdateMemberInput struct {
	MemberID               string
	GroupID                *string
	NamaLengkap            *string
	NamaPanggilan          *string
	JenisKelamin           *string
	TempatLahir            *string
	TanggalLahir           *string
	Kelompok               *string
	Desa                   *string
	Daerah                 *string
	AlamatRumah            *string
	NoWA                   *string
	IsMuballigh            *bool
	IsKerja                *bool
	IsNikah                *bool
	TinggiBadan            *string
	BeratBadan             *string
	Hobi                   *string
	Pekerjaan              *string
	FotoURL                *string
	StatusPembinaan        *string
	TanggalMasuk           *string
	TanggalKeluar          *string
	JenjangPendidikan      *string
	Sekolah                *string
	Jurusan                *string
	TahunMulaiPendidikan   *string
	TahunSelesaiPendidikan *string
}

func (s *MemberService) UpdateFull(ctx context.Context, in UpdateMemberInput) (*model.MemberDetailDTO, error) {

	if in.NamaLengkap != nil {
		v := util.TitleCaseID(*in.NamaLengkap)
		in.NamaLengkap = &v
	}
	if in.NamaPanggilan != nil {
		v := util.TitleCaseID(*in.NamaPanggilan)
		in.NamaPanggilan = &v
	}
	if in.TempatLahir != nil {
		v := util.TitleCaseID(*in.TempatLahir)
		in.TempatLahir = &v
	}
	if in.Desa != nil {
		v := util.TitleCaseID(*in.Desa)
		in.Desa = &v
	}
	if in.Daerah != nil {
		v := util.TitleCaseID(*in.Daerah)
		in.Daerah = &v
	}
	if in.MemberID == "" {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, in.MemberID); err != nil {
		return nil, apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}

	patch := map[string]interface{}{}
	addStr := func(col string, v *string) {
		if v != nil {
			patch[col] = *v
		}
	}
	addBool := func(col string, v *bool) {
		if v != nil {
			patch[col] = *v
		}
	}

	addStr("nama_lengkap", in.NamaLengkap)
	addStr("nama_panggilan", in.NamaPanggilan)
	if in.JenisKelamin != nil {
		if jk := normalizeGender(*in.JenisKelamin); jk != "" {
			patch["jenis_kelamin"] = jk
		}
	}
	addStr("tempat_lahir", in.TempatLahir)
	if in.TanggalLahir != nil {
		if *in.TanggalLahir == "" {
			patch["tanggal_lahir"] = nil
		} else {
			patch["tanggal_lahir"] = *in.TanggalLahir
		}
	}
	if in.Kelompok != nil || in.GroupID != nil {
		gid := ""
		if in.GroupID != nil {
			gid = *in.GroupID
		}
		knama := ""
		if in.Kelompok != nil {
			knama = *in.Kelompok
		}
		resolvedID, resolvedName, err := s.resolveGroup(ctx, gid, knama)
		if err != nil {
			return nil, err
		}
		patch["group_id"] = resolvedID
		patch["kelompok"] = resolvedName
	}
	addStr("desa", in.Desa)
	addStr("daerah", in.Daerah)
	addStr("alamat_rumah", in.AlamatRumah)
	if in.NoWA != nil && *in.NoWA != "" {
		patch["no_wa"] = util.NormalizePhone(*in.NoWA)
	}
	addBool("is_muballigh", in.IsMuballigh)
	addBool("is_kerja", in.IsKerja)
	addBool("is_nikah", in.IsNikah)
	addStr("tinggi_badan", in.TinggiBadan)
	addStr("berat_badan", in.BeratBadan)
	addStr("hobi", in.Hobi)
	addStr("pekerjaan", in.Pekerjaan)
	addStr("foto_url", in.FotoURL)
	addStr("status_pembinaan", in.StatusPembinaan)
	if in.TanggalMasuk != nil {
		if *in.TanggalMasuk == "" {
			patch["tanggal_masuk"] = nil
		} else {
			patch["tanggal_masuk"] = *in.TanggalMasuk
		}
	}
	if in.TanggalKeluar != nil {
		if *in.TanggalKeluar == "" {
			patch["tanggal_keluar"] = nil
		} else {
			patch["tanggal_keluar"] = *in.TanggalKeluar
		}
	}
	addStr("jenjang_pendidikan", in.JenjangPendidikan)
	addStr("sekolah", in.Sekolah)
	addStr("jurusan", in.Jurusan)
	addStr("tahun_mulai_pendidikan", in.TahunMulaiPendidikan)
	addStr("tahun_selesai_pendidikan", in.TahunSelesaiPendidikan)

	if len(patch) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrValidation, "tidak ada perubahan")
	}
	if err := s.repo.Update(ctx, in.MemberID, patch); err != nil {
		return nil, err
	}
	if _, ok := patch["group_id"]; ok {
		_ = s.repo.SyncUsersGroupByMember(ctx, in.MemberID)
	}
	m, err := s.repo.FindByID(ctx, in.MemberID)
	if err != nil {
		return nil, err
	}
	dto := s.toDetailDTO(m)
	return &dto, nil
}

func (s *MemberService) Deactivate(ctx context.Context, memberID string) error {
	if memberID == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, memberID); err != nil {
		return apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}
	return s.repo.Deactivate(ctx, memberID)
}

func (s *MemberService) FindForExport(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.repo.FindAllIncludingInactive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		jk := ""
		if r.JenisKelamin != nil {
			jk = *r.JenisKelamin
		}
		tgl := ""
		if r.TanggalLahir != nil {
			tgl = r.TanggalLahir.Format("2006-01-02")
		}
		out = append(out, map[string]interface{}{
			"nama_lengkap":     r.NamaLengkap,
			"nama_panggilan":   r.NamaPanggilan,
			"jenis_kelamin":    jk,
			"tempat_lahir":     r.TempatLahir,
			"tanggal_lahir":    tgl,
			"usia":             util.GetAge(r.TanggalLahir),
			"kategori":         util.GetMemberCategory(r.TanggalLahir, r.JenjangPendidikan, r.IsNikah),
			"kelompok":         r.Kelompok,
			"desa":             r.Desa,
			"daerah":           r.Daerah,
			"alamat_rumah":     r.AlamatRumah,
			"no_wa":            r.NoWA,
			"pekerjaan":        r.Pekerjaan,
			"hobi":             r.Hobi,
			"status_pembinaan": r.StatusPembinaan,
		})
	}
	return out, nil
}

func normalizeGender(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "l", "laki-laki", "laki laki", "pria", "male":
		return "L"
	case "p", "perempuan", "wanita", "female":
		return "P"
	}
	return ""
}

var _ = repository.NewMemberRepo

func (s *MemberService) UpdateFotoURL(ctx context.Context, memberID, url string) error {
	return s.repo.UpdateFotoURL(ctx, memberID, url)
}

func (s *MemberService) GetFotoURL(ctx context.Context, memberID string) (string, error) {
	return s.repo.GetFotoURL(ctx, memberID)
}

func (s *MemberService) DeleteMember(ctx context.Context, memberID string) error {
	if memberID == "" {
		return apperrors.Wrap(apperrors.ErrValidation, "member_id wajib diisi")
	}
	if _, err := s.repo.FindByID(ctx, memberID); err != nil {
		return apperrors.Wrap(apperrors.ErrNotFound, "jamaah tidak ditemukan")
	}
	n, err := s.repo.CountUsersByMemberID(ctx, memberID)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperrors.Wrap(apperrors.ErrValidation, "member sudah punya akun user, hapus lewat Kelola Akun")
	}
	return s.repo.Delete(ctx, memberID)
}
