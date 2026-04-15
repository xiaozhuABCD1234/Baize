package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	apperrs "backend/internal/errors"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/utils"
)

var (
	ErrEmailExists     = apperrs.ErrEmailExists
	ErrUsernameExists  = apperrs.ErrUsernameExists
	ErrUserNotFound    = apperrs.ErrUserNotFound
	ErrInvalidPassword = apperrs.ErrInvalidPassword
	ErrEmailNotChanged = apperrs.ErrEmailNotChanged
	ErrSamePassword    = apperrs.ErrSamePassword
	ErrInvalidRole     = apperrs.ErrInvalidRole
)

type UserServiceInterface interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
	ListUsersWithPagination(ctx context.Context, page, pageSize int) ([]models.User, int64, error)
	UpdateUser(ctx context.Context, req UpdateUserRequest) error
	ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	ResetPassword(ctx context.Context, userID uint, newPassword string) error
	UpdateEmail(ctx context.Context, userID uint, newEmail string) error
	DeleteUser(ctx context.Context, id uint) error
	ForceDeleteUser(ctx context.Context, id uint) error
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

type RegisterRequest struct {
	Username string            `json:"username" validate:"required,min=3,max=50"`
	Email    string            `json:"email" validate:"required,email"`
	Password string            `json:"password" validate:"required,min=6,max=32"`
	Phone    string            `json:"phone,omitempty" validate:"omitempty,e164"`
	UserType string            `json:"user_type,omitempty"`
	Status   models.UserStatus `json:"status,omitempty"`
}

type RegisterResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	UserType  string `json:"user_type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, apperrs.ErrUserNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if existingUser != nil {
		return nil, apperrs.ErrUsernameExists
	}

	existingUser, err = s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, apperrs.ErrUserNotFound) {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if existingUser != nil {
		return nil, apperrs.ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	userType := req.UserType
	if userType == "" {
		userType = string(models.UserTypeUser)
	}
	if !isValidUserType(userType) {
		return nil, apperrs.ErrInvalidRole
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Phone:    req.Phone,
		UserType: models.UserType(userType),
		Status:   models.UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &RegisterResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		UserType:  string(user.UserType),
		Status:    strconv.FormatInt(int64(user.Status), 10),
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func isValidUserType(t string) bool {
	switch models.UserType(t) {
	case models.UserTypeUser, models.UserTypeMaster, models.UserTypeInstitution, models.UserTypeAdmin:
		return true
	default:
		return false
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Status   string `json:"status"`
	Token    string `json:"token"`
}

func (s *UserService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrs.ErrUserNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, apperrs.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrs.ErrInvalidPassword
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email, string(user.UserType))
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	return &LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		UserType: string(user.UserType),
		Status:   strconv.FormatInt(int64(user.Status), 10),
		Token:    tokenPair.AccessToken,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, apperrs.ErrUserNotFound
	}
	user.Password = ""
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.List(ctx, "created_at desc")
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

func (s *UserService) ListUsersWithPagination(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.repo.ListWithPagination(ctx, page, pageSize, "created_at desc")
	if err != nil {
		return nil, 0, fmt.Errorf("分页查询用户失败: %w", err)
	}

	for i := range users {
		users[i].Password = ""
	}

	return users, total, err
}

type UpdateUserRequest struct {
	ID       uint   `json:"id" validate:"required"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	UserType string `json:"user_type,omitempty" validate:"omitempty,oneof=user master institution admin"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=0 1 2"`
}

func (s *UserService) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	existing, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if existing == nil {
		return apperrs.ErrUserNotFound
	}

	if req.Email != "" && req.Email != existing.Email {
		emailUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err != nil {
			return fmt.Errorf("检查邮箱失败: %w", err)
		}
		if emailUser != nil {
			return apperrs.ErrEmailExists
		}
		existing.Email = req.Email
	}

	if req.UserType != "" {
		if !isValidUserType(req.UserType) {
			return apperrs.ErrInvalidRole
		}
		existing.UserType = models.UserType(req.UserType)
	}

	if req.Status != "" {
		status := models.UserStatus(0)
		if _, err := fmt.Sscanf(req.Status, "%d", &status); err != nil {
			return fmt.Errorf("无效的 Status 值: %w", err)
		}
		existing.Status = status
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

type ChangePasswordRequest struct {
	UserID      uint   `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=32"`
}

func (s *UserService) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return apperrs.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return apperrs.ErrInvalidPassword
	}

	if req.OldPassword == req.NewPassword {
		return apperrs.ErrSamePassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, req.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return apperrs.ErrUserNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	return nil
}

func (s *UserService) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return apperrs.ErrUserNotFound
	}

	if user.Email == newEmail {
		return apperrs.ErrEmailNotChanged
	}

	existing, err := s.repo.GetByEmail(ctx, newEmail)
	if err != nil {
		return fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existing != nil {
		return apperrs.ErrEmailExists
	}

	if err := s.repo.UpdateEmail(ctx, userID, newEmail); err != nil {
		return fmt.Errorf("更新邮箱失败: %w", err)
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return apperrs.ErrUserNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

func (s *UserService) ForceDeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.ForceDelete(ctx, id); err != nil {
		return fmt.Errorf("强制删除用户失败: %w", err)
	}
	return nil
}
