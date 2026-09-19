package service

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

var (
	usernameRegex     = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
	minPasswordLength = 6
	maxSubmPerIPDay   = 3
)

type PendingService struct {
	repo       *repository.PendingRepo
	userRepo   *repository.UserRepo
	memberRepo *repository.MemberRepo
	wa         *WASender
	waEnabled  bool
}

func NewPendingService(
	repo *repository.PendingRepo,
	userRepo *repository.UserRepo,
	memberRepo *repository.MemberRepo,
	wa *WASender,
) *PendingService {
	return &PendingService{
		repo:       repo,
		userRepo:   userRepo,
		memberRepo: memberRepo,
		wa:         wa,
		waEnabled:  os.Getenv("WA_NOTIF_ENABLED") == "true",
	}
}

/* sendWA — soft-fail, hanya kirim kalau WA_NOTIF_ENABLED=true */
func (s *PendingService) sendWA(ctx context.Context, to, msg string) {
	if !s.waEnabled || s.wa == nil {
		return
	}
	_ = s.wa.Send(ctx, to, msg)
}

/* ===== Public: check username ===== */

type CheckUsernameResult struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

func (s *PendingService) CheckUsername(ctx context.Context, username string) (*CheckUsernameResult, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if !usernameRegex.MatchString(username) {
		return &CheckUsernameResult{Available: false, Reason: "invalid"}, nil
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return &CheckUsernameResult{Available: false, Reason: "taken"}, nil
	}

	n, err := s.repo.CountByUsernamePending(ctx, username)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		return &CheckUsernameResult{Available: false, Reason: "pending"}, nil
	}

	return &CheckUsernameResult{Available: true}, nil
}

/* ===== Public: submit registration ===== */

type SubmitRegistrationInput struct {
	NamaLengkap            string
	NamaPanggilan          string
	JenisKelamin           string
	TempatLahir            string
	TanggalLahir           string
	NoWA                   string
	AlamatRumah            string
	Desa                   string
	Daerah                 string
	Pekerjaan              string
	Hobi                   string
	IsNikah                bool
	JenjangPendidikan      string
	Sekolah                string
	Jurusan                string
	TahunMulaiPendidikan   string
	TahunSelesaiPendidikan string
	FotoURL                string
	Username               string
	Password               string
	ClientIP               string
}

type SubmitRegistrationResult struct {
	SubmissionID string `json:"submission_id"`
	NamaLengkap  string `json:"nama_lengkap"`
	SubmittedAt  string `json:"submitted_at"`
}

func (s *PendingService) SubmitRegistration(ctx context.Context, in SubmitRegistrationInput) (*SubmitRegistrationResult, error) {
	nama := strings.TrimSpace(in.NamaLengkap)
	jk := strings.ToUpper(strings.TrimSpace(in.JenisKelamin))
	noWA := strings.TrimSpace(in.NoWA)
	username := strings.ToLower(strings.TrimSpace(in.Username))
	password := in.Password

	if len(nama) < 3 {
		return nil, errors.New("Nama lengkap minimal 3 karakter")
	}
	if jk != "L" && jk != "P" {
		return nil, errors.New("Jenis kelamin harus L atau P")
	}
	if noWA == "" {
		return nil, errors.New("Nomor WhatsApp wajib diisi")
	}
	normalizedWA := util.NormalizePhone(noWA)
	if len(normalizedWA) < 10 || len(normalizedWA) > 15 {
		return nil, errors.New("Nomor WhatsApp tidak valid")
	}
	if !usernameRegex.MatchString(username) {
		return nil, errors.New("Username tidak valid. Gunakan huruf kecil, angka, atau underscore (3-20 karakter).")
	}
	if len(password) < minPasswordLength {
		return nil, errors.New("Password minimal 6 karakter")
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Username sudah dipakai. Coba yang lain.")
	}

	clientIP := strings.TrimSpace(in.ClientIP)
	if clientIP == "" {
		clientIP = "unknown"
	}
	if len(clientIP) > 50 {
		clientIP = clientIP[:50]
	}

	today := time.Now().Format("2006-01-02")
	n, err := s.repo.CountByIPToday(ctx, clientIP, today)
	if err != nil {
		return nil, err
	}
	if n >= maxSubmPerIPDay {
		return nil, errors.New("Terlalu banyak pendaftaran dari perangkat ini. Coba lagi besok.")
	}

	// Cek WA existing di members
	allMembers, err := s.memberRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range allMembers {
		if util.NormalizePhone(m.NoWA) == normalizedWA && m.NoWA != "" {
			return nil, errors.New("Nomor WhatsApp sudah terdaftar sebagai jamaah")
		}
	}

	// Cek WA pending
	waPending, err := s.repo.CountByWAPending(ctx, normalizedWA)
	if err != nil {
		return nil, err
	}
	if waPending > 0 {
		return nil, errors.New("Pendaftaran dengan nomor ini sedang menunggu verifikasi")
	}

	// Cek username pending
	userPending, err := s.repo.CountByUsernamePending(ctx, username)
	if err != nil {
		return nil, err
	}
	if userPending > 0 {
		return nil, errors.New("Username sedang menunggu verifikasi. Coba yang lain.")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var tglLahir *time.Time
	if in.TanggalLahir != "" {
		if t, err := time.Parse("2006-01-02", in.TanggalLahir); err == nil {
			tglLahir = &t
		}
	}

	var jkPtr *string
	if jk != "" {
		jkPtr = &jk
	}

	p := &model.PendingMember{
		SubmissionID:           util.NewID("SUB"),
		NamaLengkap:            nama,
		NamaPanggilan:          strings.TrimSpace(in.NamaPanggilan),
		JenisKelamin:           jkPtr,
		TempatLahir:            strings.TrimSpace(in.TempatLahir),
		TanggalLahir:           tglLahir,
		NoWA:                   normalizedWA,
		AlamatRumah:            strings.TrimSpace(in.AlamatRumah),
		Desa:                   strings.TrimSpace(in.Desa),
		Daerah:                 strings.TrimSpace(in.Daerah),
		Pekerjaan:              strings.TrimSpace(in.Pekerjaan),
		Hobi:                   strings.TrimSpace(in.Hobi),
		IsNikah:                in.IsNikah,
		JenjangPendidikan:      strings.TrimSpace(in.JenjangPendidikan),
		Sekolah:                strings.TrimSpace(in.Sekolah),
		Jurusan:                strings.TrimSpace(in.Jurusan),
		TahunMulaiPendidikan:   strings.TrimSpace(in.TahunMulaiPendidikan),
		TahunSelesaiPendidikan: strings.TrimSpace(in.TahunSelesaiPendidikan),
		FotoURL:                strings.TrimSpace(in.FotoURL),
		Username:               username,
		PasswordHash:           hash,
		Status:                 "PENDING",
		SubmittedIP:            clientIP,
	}

	if err := s.repo.Insert(ctx, p); err != nil {
		return nil, err
	}

	return &SubmitRegistrationResult{
		SubmissionID: p.SubmissionID,
		NamaLengkap:  p.NamaLengkap,
		SubmittedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

/* ===== Admin: list & detail ===== */

func (s *PendingService) GetPendingMembers(ctx context.Context, status string) ([]model.PendingMemberDTO, error) {
	rows, err := s.repo.FindAll(ctx, status)
	if err != nil {
		return nil, err
	}
	out := make([]model.PendingMemberDTO, 0, len(rows))
	for _, p := range rows {
		out = append(out, toPendingDTO(p))
	}
	return out, nil
}

func (s *PendingService) GetPendingMemberDetail(ctx context.Context, id string) (*model.PendingMemberDTO, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("Pendaftaran tidak ditemukan")
	}
	dto := toPendingDTO(*p)
	return &dto, nil
}

/* ===== Admin: approve ===== */

type ApproveResult struct {
	MemberID     string `json:"member_id"`
	SubmissionID string `json:"submission_id"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
}

func (s *PendingService) Approve(ctx context.Context, submissionID, kelompok, reviewerID string) (*ApproveResult, error) {
	if submissionID == "" {
		return nil, errors.New("submission_id wajib diisi")
	}
	p, err := s.repo.FindByID(ctx, submissionID)
	if err != nil {
		return nil, errors.New("Pendaftaran tidak ditemukan")
	}
	if strings.ToUpper(p.Status) != "PENDING" {
		return nil, errors.New("Pendaftaran sudah diproses")
	}
	username := strings.ToLower(strings.TrimSpace(p.Username))
	if !usernameRegex.MatchString(username) {
		return nil, errors.New("Data pendaftar tidak memiliki username yang valid")
	}
	if p.PasswordHash == "" {
		return nil, errors.New("Data pendaftar tidak memiliki password")
	}

	exists, err := s.userRepo.UsernameExists(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Username sudah dipakai. Tolak pendaftar dan minta daftar ulang dengan username lain.")
	}

	memberID := util.NewID("MBR")
	memberIn := repository.NewMemberInput{
		MemberID:               memberID,
		NamaLengkap:            p.NamaLengkap,
		NamaPanggilan:          p.NamaPanggilan,
		JenisKelamin:           p.JenisKelamin,
		TempatLahir:            p.TempatLahir,
		TanggalLahir:           p.TanggalLahir,
		FotoURL:                p.FotoURL,
		NoWA:                   p.NoWA,
		AlamatRumah:            p.AlamatRumah,
		Desa:                   p.Desa,
		Daerah:                 p.Daerah,
		Kelompok:               kelompok,
		IsMuballigh:            false,
		IsKerja:                false,
		IsNikah:                p.IsNikah,
		Hobi:                   p.Hobi,
		Pekerjaan:              p.Pekerjaan,
		StatusPembinaan:        "AKTIF",
		JenjangPendidikan:      p.JenjangPendidikan,
		Sekolah:                p.Sekolah,
		Jurusan:                p.Jurusan,
		TahunMulaiPendidikan:   p.TahunMulaiPendidikan,
		TahunSelesaiPendidikan: p.TahunSelesaiPendidikan,
	}
	if err := s.memberRepo.Insert(ctx, memberIn); err != nil {
		return nil, err
	}

	userID := util.NewID("USR")
	if err := s.userRepo.Insert(ctx, repository.NewUserInput{
		UserID:       userID,
		Username:     username,
		PasswordHash: p.PasswordHash,
		Nama:         p.NamaLengkap,
		Role:         "MEMBER",
		MemberID:     memberID,
	}); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateApproved(ctx, submissionID, reviewerID, memberID); err != nil {
		return nil, err
	}

	// Kirim WA (soft-fail)
	if p.NoWA != "" {
		sapaan := util.BuildSapaan(derefStr(p.JenisKelamin), p.TanggalLahir)
		doa := util.BuildDoa(derefStr(p.JenisKelamin))
		msg := "Assalamu'alaikum " + sapaan + ",\n\n" +
			"Alhamdulillah, pendaftaran " + sapaan + " di *Sambung Ngaji* sudah disetujui.\n\n" +
			"Silakan masuk menggunakan:\nUsername: *" + username + "*\nPassword: sesuai yang " + sapaan + " daftarkan\n\n" +
			"Barakallahu fiik.\n" + doa + " 🤍"
		s.sendWA(ctx, p.NoWA, msg)
	}

	return &ApproveResult{
		MemberID:     memberID,
		SubmissionID: submissionID,
		UserID:       userID,
		Username:     username,
	}, nil
}

/* ===== Admin: reject ===== */

func (s *PendingService) Reject(ctx context.Context, submissionID, reviewerID, reason string) error {
	if submissionID == "" {
		return errors.New("submission_id wajib diisi")
	}
	p, err := s.repo.FindByID(ctx, submissionID)
	if err != nil {
		return errors.New("Pendaftaran tidak ditemukan")
	}
	if strings.ToUpper(p.Status) != "PENDING" {
		return errors.New("Pendaftaran sudah diproses")
	}
	if strings.TrimSpace(reason) == "" {
		reason = "Tidak memenuhi syarat"
	}

	if err := s.repo.UpdateRejected(ctx, submissionID, reviewerID, reason); err != nil {
		return err
	}

	if p.NoWA != "" {
		sapaan := util.BuildSapaan(derefStr(p.JenisKelamin), p.TanggalLahir)
		doa := util.BuildDoa(derefStr(p.JenisKelamin))
		msg := "Assalamu'alaikum " + sapaan + ",\n\n" +
			"Mohon maaf, pendaftaran " + sapaan + " di *Sambung Ngaji* belum bisa kami setujui.\n\n" +
			"Alasan: " + reason + "\n\n" +
			"Silakan hubungi admin untuk informasi lebih lanjut.\n\n" + doa + " 🙏"
		s.sendWA(ctx, p.NoWA, msg)
	}

	return nil
}

/* ===== Helpers ===== */

func toPendingDTO(p model.PendingMember) model.PendingMemberDTO {
	jk := ""
	if p.JenisKelamin != nil {
		jk = *p.JenisKelamin
	}
	tgl := ""
	if p.TanggalLahir != nil {
		tgl = p.TanggalLahir.Format("2006-01-02")
	}
	rvBy := ""
	if p.ReviewedBy != nil {
		rvBy = *p.ReviewedBy
	}
	rvAt := ""
	if p.ReviewedAt != nil {
		rvAt = p.ReviewedAt.Format("2006-01-02T15:04:05.000Z07:00")
	}
	cmid := ""
	if p.CreatedMemberID != nil {
		cmid = *p.CreatedMemberID
	}
	return model.PendingMemberDTO{
		SubmissionID:           p.SubmissionID,
		NamaLengkap:            p.NamaLengkap,
		NamaPanggilan:          p.NamaPanggilan,
		JenisKelamin:           jk,
		TempatLahir:            p.TempatLahir,
		TanggalLahir:           tgl,
		NoWA:                   p.NoWA,
		AlamatRumah:            p.AlamatRumah,
		Desa:                   p.Desa,
		Daerah:                 p.Daerah,
		Pekerjaan:              p.Pekerjaan,
		Hobi:                   p.Hobi,
		IsNikah:                p.IsNikah,
		JenjangPendidikan:      p.JenjangPendidikan,
		Sekolah:                p.Sekolah,
		Jurusan:                p.Jurusan,
		TahunMulaiPendidikan:   p.TahunMulaiPendidikan,
		TahunSelesaiPendidikan: p.TahunSelesaiPendidikan,
		FotoURL:                p.FotoURL,
		Username:               p.Username,
		Status:                 p.Status,
		SubmittedAt:            p.SubmittedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		SubmittedIP:            p.SubmittedIP,
		ReviewedBy:             rvBy,
		ReviewedAt:             rvAt,
		RejectionReason:        p.RejectionReason,
		CreatedMemberID:        cmid,
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
