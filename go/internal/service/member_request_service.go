package service

import (
	"context"
	"errors"
	"strings"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type MemberRequestService struct {
	requestRepo *repository.MemberRequestRepo
	memberRepo  *repository.MemberRepo
	userRepo    *repository.UserAdminRepo
}

func NewMemberRequestService(
	requestRepo *repository.MemberRequestRepo,
	memberRepo *repository.MemberRepo,
	userRepo *repository.UserAdminRepo,
) *MemberRequestService {
	return &MemberRequestService{requestRepo: requestRepo, memberRepo: memberRepo, userRepo: userRepo}
}

type RequestBecomeMemberResult struct {
	AutoCreated *string `json:"auto_created,omitempty"`
	RequestSent *string `json:"request_sent,omitempty"`
	Message     string  `json:"message"`
}

func (s *MemberRequestService) Request(ctx context.Context, user *model.User) (*RequestBecomeMemberResult, error) {
	if user == nil {
		return nil, errors.New("Unauthorized")
	}
	if user.MemberID != nil && *user.MemberID != "" {
		return nil, errors.New("Akun Anda sudah menjadi member")
	}

	if user.Role == "ADMIN" || user.Role == "SUPER_ADMIN" {
		memberID, err := s.autoCreateMember(ctx, user.UserID, user.Name)
		if err != nil {
			return nil, err
		}
		return &RequestBecomeMemberResult{
			AutoCreated: &memberID,
			Message:     "Data member berhasil dibuat untuk akun Anda",
		}, nil
	}

	existing, err := s.requestRepo.FindByUserID(ctx, user.UserID)
	if err == nil && existing != nil && strings.ToUpper(existing.Status) == "PENDING" {
		return &RequestBecomeMemberResult{
			RequestSent: &existing.RequestID,
			Message:     "Permintaan menjadi member sudah dikirim dan menunggu persetujuan admin",
		}, nil
	}

	req := &model.MemberRequest{
		RequestID: util.NewID("REQ"),
		UserID:    user.UserID,
		Name:      user.Name,
		Status:    "PENDING",
	}
	if err := s.requestRepo.Insert(ctx, req); err != nil {
		return nil, err
	}
	return &RequestBecomeMemberResult{
		RequestSent: &req.RequestID,
		Message:     "Permintaan menjadi member terkirim. Silakan tunggu persetujuan admin.",
	}, nil
}

func (s *MemberRequestService) autoCreateMember(ctx context.Context, userID, name string) (string, error) {
	memberID := util.NewID("MBR")
	in := repository.NewMemberInput{
		MemberID:        memberID,
		FullName:     util.TitleCaseID(strings.TrimSpace(name)),
		Nickname:   util.TitleCaseID(strings.TrimSpace(name)),
		MentoringStatus: "AKTIF",
	}
	if err := s.memberRepo.Insert(ctx, in); err != nil {
		return "", err
	}
	if err := s.userRepo.Update(ctx, userID, map[string]interface{}{"member_id": memberID}); err != nil {
		return "", err
	}
	return memberID, nil
}

func (s *MemberRequestService) List(ctx context.Context, status, groupID string) ([]model.MemberRequestDTO, error) {
	rows, err := s.requestRepo.FindByStatus(ctx, strings.ToUpper(strings.TrimSpace(status)), groupID)
	if err != nil {
		return nil, err
	}
	out := make([]model.MemberRequestDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMemberRequestDTO(r))
	}
	return out, nil
}

type ApproveMemberRequestInput struct {
	RequestID  string
	ReviewerID string
}

func (s *MemberRequestService) Approve(ctx context.Context, in ApproveMemberRequestInput) (*model.MemberRequestDTO, error) {
	if in.RequestID == "" {
		return nil, errors.New("request_id wajib diisi")
	}
	req, err := s.requestRepo.FindByID(ctx, in.RequestID)
	if err != nil || req == nil {
		return nil, errors.New("Permintaan tidak ditemukan")
	}
	if strings.ToUpper(req.Status) != "PENDING" {
		return nil, errors.New("Permintaan sudah diproses")
	}
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user == nil {
		return nil, errors.New("User tidak ditemukan")
	}
	if user.MemberID != nil && *user.MemberID != "" {
		return nil, errors.New("User ini sudah memiliki data member")
	}

	memberID, err := s.autoCreateMember(ctx, user.UserID, req.Name)
	if err != nil {
		return nil, err
	}
	if err := s.requestRepo.UpdateApproved(ctx, req.RequestID, memberID, in.ReviewerID); err != nil {
		return nil, err
	}
	fresh, _ := s.requestRepo.FindByID(ctx, req.RequestID)
	if fresh == nil {
		return nil, errors.New("gagal ambil data permintaan")
	}
	dto := toMemberRequestDTO(*fresh)
	return &dto, nil
}

func (s *MemberRequestService) Reject(ctx context.Context, requestID, reviewerID, reason string) error {
	if requestID == "" {
		return errors.New("request_id wajib diisi")
	}
	req, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil || req == nil {
		return errors.New("Permintaan tidak ditemukan")
	}
	if strings.ToUpper(req.Status) != "PENDING" {
		return errors.New("Permintaan sudah diproses")
	}
	if strings.TrimSpace(reason) == "" {
		reason = "Tidak memenuhi syarat"
	}
	return s.requestRepo.UpdateRejected(ctx, requestID, reviewerID, reason)
}

func toMemberRequestDTO(m model.MemberRequest) model.MemberRequestDTO {
	dto := model.MemberRequestDTO{
		RequestID: m.RequestID,
		UserID:    m.UserID,
		Name:      m.Name,
		Status:    m.Status,
		Reason:    m.Reason,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	if m.MemberID != nil {
		dto.MemberID = *m.MemberID
	}
	if m.ReviewedBy != nil {
		dto.ReviewedBy = *m.ReviewedBy
	}
	if m.ReviewedAt != nil {
		dto.ReviewedAt = m.ReviewedAt.Format("2006-01-02T15:04:05.000Z07:00")
	}
	return dto
}
