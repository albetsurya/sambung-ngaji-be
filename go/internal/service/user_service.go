package service

import (
	"context"
	"errors"
	"strings"

	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type UserService struct {
	repo       *repository.UserAdminRepo
	authRepo   *repository.UserRepo
	memberRepo *repository.MemberRepo
	audit      *AuditService
}

func NewUserService(
	repo *repository.UserAdminRepo,
	authRepo *repository.UserRepo,
	memberRepo *repository.MemberRepo,
	audit *AuditService,
) *UserService {
	return &UserService{
		repo:       repo,
		authRepo:   authRepo,
		memberRepo: memberRepo,
		audit:      audit,
	}
}

func (s *UserService) GetUsers(ctx context.Context) ([]model.UserDTO, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, toUserDTO(u))
	}
	return out, nil
}

func (s *UserService) GetUserDetail(ctx context.Context, id string) (*model.UserDTO, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	dto := toUserDTO(*u)
	return &dto, nil
}

type CreateUserInput struct {
	Username string
	Password string
	Role     string
	Nama     string
	MemberID string
	AdminID  string
}

func (s *UserService) CreateUser(ctx context.Context, in CreateUserInput) (*model.UserDTO, error) {
	if in.Username == "" || in.Password == "" || in.Role == "" {
		return nil, errors.New("username, password, role wajib diisi")
	}
	if !validRole(in.Role) {
		return nil, errors.New("role tidak valid")
	}
	if in.MemberID == "" {
		return nil, errors.New("member_id wajib diisi. Setiap user harus terhubung ke jamaah.")
	}

	if _, err := s.memberRepo.FindByID(ctx, in.MemberID); err != nil {
		return nil, errors.New("jamaah dengan member_id tersebut tidak ditemukan")
	}

	hasActive, err := s.authRepo.MemberHasActiveUser(ctx, in.MemberID)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, errors.New("jamaah ini sudah punya akun aktif")
	}

	exists, err := s.authRepo.UsernameExists(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username sudah digunakan")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	userID := util.NewID("USR")
	nama := in.Nama
	if nama == "" {
		nama = in.Username
	}

	if err := s.authRepo.Insert(ctx, repository.NewUserInput{
		UserID:       userID,
		Username:     in.Username,
		PasswordHash: hash,
		Nama:         nama,
		Role:         in.Role,
		MemberID:     in.MemberID,
	}); err != nil {
		return nil, err
	}

	u, _ := s.repo.FindByID(ctx, userID)
	dto := toUserDTO(*u)
	return &dto, nil
}

type UpdateUserInput struct {
	UserID      string
	Nama        *string
	Role        *string
	StatusAktif *bool
	Password    *string
	AdminID     string
}

func (s *UserService) UpdateUser(ctx context.Context, in UpdateUserInput) (*model.UserDTO, error) {
	if in.UserID == "" {
		return nil, errors.New("user_id wajib diisi")
	}
	existing, err := s.repo.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	if in.Role != nil && !validRole(*in.Role) {
		return nil, errors.New("role tidak valid")
	}

	// Proteksi super admin terakhir
	if in.Role != nil && existing.Role == "SUPER_ADMIN" && *in.Role != "SUPER_ADMIN" {
		n, _ := s.repo.CountActiveSuperAdminsExcept(ctx, in.UserID)
		if n == 0 {
			return nil, errors.New("tidak bisa mengubah role SUPER_ADMIN terakhir")
		}
	}
	if in.StatusAktif != nil && !*in.StatusAktif && existing.Role == "SUPER_ADMIN" {
		n, _ := s.repo.CountActiveSuperAdminsExcept(ctx, in.UserID)
		if n == 0 {
			return nil, errors.New("tidak bisa menonaktifkan SUPER_ADMIN terakhir")
		}
	}

	patch := map[string]interface{}{}
	shouldInvalidate := false

	if in.Nama != nil {
		patch["nama"] = *in.Nama
	}
	if in.Role != nil && *in.Role != existing.Role {
		patch["role"] = *in.Role
		shouldInvalidate = true
	}
	if in.StatusAktif != nil && *in.StatusAktif != existing.StatusAktif {
		patch["status_aktif"] = *in.StatusAktif
		shouldInvalidate = true
	}
	if in.Password != nil && *in.Password != "" {
		hash, err := auth.HashPassword(*in.Password)
		if err != nil {
			return nil, err
		}
		patch["password_hash"] = hash
		shouldInvalidate = true
	}

	if len(patch) > 0 {
		if err := s.repo.Update(ctx, in.UserID, patch); err != nil {
			return nil, err
		}
	}

	if shouldInvalidate {
		_ = s.repo.DeleteAllSessions(ctx, in.UserID)
	}

	u, _ := s.repo.FindByID(ctx, in.UserID)
	dto := toUserDTO(*u)
	return &dto, nil
}

type UpdateRoleInput struct {
	UserID  string
	Role    string
	AdminID string
}

func (s *UserService) UpdateUserRole(ctx context.Context, in UpdateRoleInput) (*model.UserDTO, error) {
	if in.UserID == "" || in.Role == "" {
		return nil, errors.New("user_id dan role wajib diisi")
	}
	if !validRole(in.Role) {
		return nil, errors.New("role tidak valid")
	}
	existing, err := s.repo.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	if existing.Role == "SUPER_ADMIN" && in.Role != "SUPER_ADMIN" {
		n, _ := s.repo.CountActiveSuperAdminsExcept(ctx, in.UserID)
		if n == 0 {
			return nil, errors.New("tidak bisa mengubah role SUPER_ADMIN terakhir")
		}
	}
	if err := s.repo.Update(ctx, in.UserID, map[string]interface{}{"role": in.Role}); err != nil {
		return nil, err
	}
	if in.Role != existing.Role {
		_ = s.repo.DeleteAllSessions(ctx, in.UserID)
	}
	u, _ := s.repo.FindByID(ctx, in.UserID)
	dto := toUserDTO(*u)
	return &dto, nil
}

func (s *UserService) GetMemberUserStatus(ctx context.Context, memberID string) (map[string]interface{}, error) {
	if memberID == "" {
		return nil, errors.New("member_id wajib diisi")
	}
	u, err := s.repo.FindActiveUserByMemberID(ctx, memberID)
	if err != nil {
		return map[string]interface{}{
			"has_user": false,
			"user":     nil,
		}, nil
	}
	return map[string]interface{}{
		"has_user": true,
		"user":     toUserDTO(*u),
	}, nil
}

/* ===== Password ===== */

type ChangePasswordInput struct {
	UserID       string
	OldPassword  string
	NewPassword  string
	CurrentToken string
}

func (s *UserService) ChangeMyPassword(ctx context.Context, in ChangePasswordInput) error {
	if in.OldPassword == "" {
		return errors.New("password lama wajib diisi")
	}
	if in.NewPassword == "" {
		return errors.New("password baru wajib diisi")
	}
	if len(in.NewPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	if in.NewPassword == in.OldPassword {
		return errors.New("password baru harus berbeda dari password lama")
	}
	u, err := s.repo.FindByID(ctx, in.UserID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}
	match, _ := auth.VerifyPassword(in.OldPassword, u.PasswordHash)
	if !match {
		return errors.New("password lama salah")
	}
	hash, err := auth.HashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordHash(ctx, in.UserID, hash); err != nil {
		return err
	}
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	if userID == "" {
		return errors.New("user_id wajib diisi")
	}
	if newPassword == "" {
		return errors.New("password baru wajib diisi")
	}
	if len(newPassword) < 6 {
		return errors.New("password baru minimal 6 karakter")
	}
	if _, err := s.repo.FindByID(ctx, userID); err != nil {
		return errors.New("user tidak ditemukan")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePasswordHash(ctx, userID, hash); err != nil {
		return err
	}
	_ = s.repo.DeleteAllSessions(ctx, userID)
	return nil
}

/* ===== Helpers ===== */

var validRoles = map[string]bool{
	"SUPER_ADMIN": true,
	"ADMIN":       true,
	"TIM_PNKB":    true,
	"TIM_ABSENSI": true,
	"PENGAWAS":    true,
	"MEMBER":      true,
}

func validRole(r string) bool { return validRoles[r] }

func toUserDTO(u model.User) model.UserDTO {
	mid := ""
	if u.MemberID != nil {
		mid = *u.MemberID
	}
	lla := ""
	if u.LastLoginAt != nil {
		lla = u.LastLoginAt.Format("2006-01-02T15:04:05.000Z07:00")
	}
	return model.UserDTO{
		UserID:      u.UserID,
		Username:    u.Username,
		Nama:        u.Nama,
		Role:        u.Role,
		MemberID:    mid,
		StatusAktif: u.StatusAktif,
		CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		UpdatedAt:   u.UpdatedAt.Format("2006-01-02T15:04:05.000Z07:00"),
		LastLoginAt: lla,
	}
}

var _ = strings.TrimSpace
