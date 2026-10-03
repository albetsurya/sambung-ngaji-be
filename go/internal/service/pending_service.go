package service

import (
	"context"
	"errors"
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
	groupRepo  *repository.GroupRepo
}

func NewPendingService(
	repo *repository.PendingRepo,
	userRepo *repository.UserRepo,
	memberRepo *repository.MemberRepo,
) *PendingService {
	return &PendingService{
		repo:       repo,
		userRepo:   userRepo,
		memberRepo: memberRepo,
	}
}

func (s *PendingService) SetGroupRepo(gr *repository.GroupRepo) { s.groupRepo = gr }

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

type SubmitRegistrationInput struct {
	GroupID                string
	FullName            string
	Nickname          string
	Gender           string
	BirthPlace            string
	BirthDate           string
	WhatsappNumber                   string
	HomeAddress            string
	Village                   string
	Region                 string
	Occupation              string
	Hobby                   string
	IsMarried                bool
	EducationLevel      string
	School                string
	Major                string
	EducationStartYear   string
	EducationEndYear string
	PhotoURL                string
	Username               string
	Password               string
	ClientIP               string
}

type SubmitRegistrationResult struct {
	SubmissionID string `json:"submission_id"`
	FullName  string `json:"full_name"`
	SubmittedAt  string `json:"submitted_at"`
}

func (s *PendingService) SubmitRegistration(ctx context.Context, in SubmitRegistrationInput) (*SubmitRegistrationResult, error) {

	in.FullName = util.TitleCaseID(in.FullName)
	in.Nickname = util.TitleCaseID(in.Nickname)
	in.BirthPlace = util.TitleCaseID(in.BirthPlace)
	in.Village = util.TitleCaseID(in.Village)
	in.Region = util.TitleCaseID(in.Region)
	name := strings.TrimSpace(in.FullName)
	jk := strings.ToUpper(strings.TrimSpace(in.Gender))
	noWA := strings.TrimSpace(in.WhatsappNumber)
	username := strings.ToLower(strings.TrimSpace(in.Username))
	password := in.Password

	if len(name) < 3 {
		return nil, errors.New("Name lengkap minimal 3 karakter")
	}
	if jk != "L" && jk != "P" {
		return nil, errors.New("Type kelamin harus L atau P")
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

	allMembers, err := s.memberRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range allMembers {
		if util.NormalizePhone(m.WhatsappNumber) == normalizedWA && m.WhatsappNumber != "" {
			return nil, errors.New("Nomor WhatsApp sudah terdaftar sebagai jamaah")
		}
	}

	waPending, err := s.repo.CountByWAPending(ctx, normalizedWA)
	if err != nil {
		return nil, err
	}
	if waPending > 0 {
		return nil, errors.New("Pendaftaran dengan nomor ini sedang menunggu verifikasi")
	}

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
	if in.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", in.BirthDate); err == nil {
			tglLahir = &t
		}
	}

	var jkPtr *string
	if jk != "" {
		jkPtr = &jk
	}

	var groupIDPtr *string
	if in.GroupID != "" {
		groupIDPtr = &in.GroupID
	}

	p := &model.PendingMember{
		SubmissionID:           util.NewID("SUB"),
		GroupID:                groupIDPtr,
		FullName:            name,
		Nickname:          strings.TrimSpace(in.Nickname),
		Gender:           jkPtr,
		BirthPlace:            strings.TrimSpace(in.BirthPlace),
		BirthDate:           tglLahir,
		WhatsappNumber:                   normalizedWA,
		HomeAddress:            strings.TrimSpace(in.HomeAddress),
		Village:                   strings.TrimSpace(in.Village),
		Region:                 strings.TrimSpace(in.Region),
		Occupation:              strings.TrimSpace(in.Occupation),
		Hobby:                   strings.TrimSpace(in.Hobby),
		IsMarried:                in.IsMarried,
		EducationLevel:      strings.TrimSpace(in.EducationLevel),
		School:                strings.TrimSpace(in.School),
		Major:                strings.TrimSpace(in.Major),
		EducationStartYear:   strings.TrimSpace(in.EducationStartYear),
		EducationEndYear: strings.TrimSpace(in.EducationEndYear),
		PhotoURL:                strings.TrimSpace(in.PhotoURL),
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
		FullName:  p.FullName,
		SubmittedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

func (s *PendingService) GetPendingMembers(ctx context.Context, groupID, status string) ([]model.PendingMemberDTO, error) {
	rows, err := s.repo.FindAll(ctx, groupID, status)
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

type ApproveResult struct {
	MemberID     string `json:"member_id"`
	SubmissionID string `json:"submission_id"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
}

func (s *PendingService) Approve(ctx context.Context, submissionID, group_label, reviewerID string) (*ApproveResult, error) {
	return s.ApproveWithGroup(ctx, submissionID, "", group_label, reviewerID)
}

func (s *PendingService) ApproveWithGroup(ctx context.Context, submissionID, groupID, group_label, reviewerID string) (*ApproveResult, error) {
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

	p.FullName = util.TitleCaseID(p.FullName)
	p.Nickname = util.TitleCaseID(p.Nickname)
	p.BirthPlace = util.TitleCaseID(p.BirthPlace)
	p.Village = util.TitleCaseID(p.Village)
	p.Region = util.TitleCaseID(p.Region)

	memberID := util.NewID("MBR")
	// Standard: group_id adalah FK tunggal. Resolve dari param group_id,
	// fallback ke param group_label (name), fallback terakhir ke p.GroupID saat daftar.
	resolvedGroupID := strings.TrimSpace(groupID)
	resolvedKelompok := strings.TrimSpace(group_label)
	if resolvedGroupID == "" && p.GroupID != nil {
		resolvedGroupID = strings.TrimSpace(*p.GroupID)
	}
	var groupIDPtr *string
	if resolvedGroupID != "" && s.groupRepo != nil {
		if g, err := s.groupRepo.FindByID(ctx, resolvedGroupID); err == nil && g != nil {
			groupIDPtr = &g.GroupID
			resolvedGroupID = g.GroupID
			resolvedKelompok = g.GroupName
		} else {
			return nil, errors.New("kelompok tidak dikenal")
		}
	} else if resolvedKelompok != "" && s.groupRepo != nil {
		if g, err := s.groupRepo.FindByName(ctx, resolvedKelompok); err == nil && g != nil {
			groupIDPtr = &g.GroupID
			resolvedGroupID = g.GroupID
			resolvedKelompok = g.GroupName
		} else {
			return nil, errors.New("kelompok tidak dikenal: " + resolvedKelompok)
		}
	} else if resolvedGroupID != "" {
		groupIDPtr = &resolvedGroupID
	} else if resolvedKelompok == "" {
		return nil, errors.New("kelompok wajib dipilih saat approve")
	}
	memberIn := repository.NewMemberInput{
		MemberID:               memberID,
		GroupID:                groupIDPtr,
		FullName:            p.FullName,
		Nickname:          p.Nickname,
		Gender:           p.Gender,
		BirthPlace:            p.BirthPlace,
		BirthDate:           p.BirthDate,
		PhotoURL:                p.PhotoURL,
		WhatsappNumber:                   p.WhatsappNumber,
		HomeAddress:            p.HomeAddress,
		Village:                   p.Village,
		Region:                 p.Region,
		GroupLabel:               resolvedKelompok,
		IsPreacher:            false,
		IsEmployed:                false,
		IsMarried:                p.IsMarried,
		Hobby:                   p.Hobby,
		Occupation:              p.Occupation,
		MentoringStatus:        "AKTIF",
		EducationLevel:      p.EducationLevel,
		School:                p.School,
		Major:                p.Major,
		EducationStartYear:   p.EducationStartYear,
		EducationEndYear: p.EducationEndYear,
	}
	if err := s.memberRepo.Insert(ctx, memberIn); err != nil {
		return nil, err
	}

	userID := util.NewID("USR")
	if err := s.userRepo.Insert(ctx, repository.NewUserInput{
		UserID:       userID,
		Username:     username,
		PasswordHash: p.PasswordHash,
		Name:         p.FullName,
		Role:         "MEMBER",
		MemberID:     memberID,
		GroupID:      groupIDPtr,
	}); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateApproved(ctx, submissionID, reviewerID, memberID); err != nil {
		return nil, err
	}

	return &ApproveResult{
		MemberID:     memberID,
		SubmissionID: submissionID,
		UserID:       userID,
		Username:     username,
	}, nil
}

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

	return nil
}

func toPendingDTO(p model.PendingMember) model.PendingMemberDTO {
	jk := ""
	if p.Gender != nil {
		jk = *p.Gender
	}
	tgl := ""
	if p.BirthDate != nil {
		tgl = p.BirthDate.Format("2006-01-02")
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
	grpID := ""
	if p.GroupID != nil {
		grpID = *p.GroupID
	}
	return model.PendingMemberDTO{
		SubmissionID:           p.SubmissionID,
		GroupID:                grpID,
		FullName:            p.FullName,
		Nickname:          p.Nickname,
		Gender:           jk,
		BirthPlace:            p.BirthPlace,
		BirthDate:           tgl,
		WhatsappNumber:                   p.WhatsappNumber,
		HomeAddress:            p.HomeAddress,
		Village:                   p.Village,
		Region:                 p.Region,
		Occupation:              p.Occupation,
		Hobby:                   p.Hobby,
		IsMarried:                p.IsMarried,
		EducationLevel:      p.EducationLevel,
		School:                p.School,
		Major:                p.Major,
		EducationStartYear:   p.EducationStartYear,
		EducationEndYear: p.EducationEndYear,
		PhotoURL:                p.PhotoURL,
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
