package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/utils"
)

// 定义业务错误
var (
	ErrEmailExists     = errors.New("邮箱已被注册")
	ErrUserNotFound    = errors.New("用户不存在")
	ErrInvalidPassword = errors.New("密码错误")
	ErrEmailNotChanged = errors.New("新邮箱与当前邮箱相同")
	ErrSamePassword    = errors.New("新密码不能与旧密码相同")
)

// UserService 用户业务逻辑层
type UserService struct {
	repo repository.UserRepository
}

// NewUserService 创建 UserService 实例
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ========== 注册/登录 ==========

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// 1. 检查邮箱是否已存在
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// 2. 密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 3. 创建用户
	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return &RegisterResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// 1. 查询用户
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	// 3. 生成 JWT Token
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	return &LoginResponse{
		ID:    user.ID,
		Email: user.Email,
		Token: tokenPair.AccessToken,
	}, nil
}

// ========== 用户查询 ==========

// GetUserByID 根据 ID 获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	// 清除敏感信息
	user.Password = ""
	return user, nil
}

// GetUserByEmail 根据邮箱获取用户（内部使用）
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

// ListUsers 获取所有用户
func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	// 清除所有用户的敏感信息
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

// ListUsersWithPagination 分页获取用户
func (s *UserService) ListUsersWithPagination(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := s.repo.ListWithPagination(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("分页查询用户失败: %w", err)
	}

	// 清除敏感信息
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// ========== 用户更新 ==========

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID    uint   `json:"id" validate:"required"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	// 1. 检查用户是否存在
	existing, err := s.repo.GetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if existing == nil {
		return ErrUserNotFound
	}

	// 2. 如果更新邮箱，检查是否已被使用
	if req.Email != "" && req.Email != existing.Email {
		emailUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err != nil {
			return fmt.Errorf("检查邮箱失败: %w", err)
		}
		if emailUser != nil {
			return ErrEmailExists
		}
		existing.Email = req.Email
	}

	// 3. 更新其他字段
	if req.Name != "" {
		// 如果有 Name 字段才更新
		// existing.Name = req.Name
	}

	// 4. 执行更新
	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	UserID      uint   `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=32"`
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, req ChangePasswordRequest) error {
	// 1. 获取用户
	user, err := s.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 3. 检查新密码是否与旧密码相同
	if req.OldPassword == req.NewPassword {
		return ErrSamePassword
	}

	// 4. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 5. 更新密码
	if err := s.repo.UpdatePassword(ctx, req.UserID, string(hashedPassword)); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	return nil
}

// ResetPassword 重置密码（管理员用）
func (s *UserService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	// 1. 检查用户是否存在
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 3. 更新密码
	if err := s.repo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	return nil
}

// UpdateEmail 更新邮箱
func (s *UserService) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	// 1. 获取当前用户
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. 检查是否相同
	if user.Email == newEmail {
		return ErrEmailNotChanged
	}

	// 3. 检查新邮箱是否已被使用
	existing, err := s.repo.GetByEmail(ctx, newEmail)
	if err != nil {
		return fmt.Errorf("检查邮箱失败: %w", err)
	}
	if existing != nil {
		return ErrEmailExists
	}

	// 4. 更新邮箱
	if err := s.repo.UpdateEmail(ctx, userID, newEmail); err != nil {
		return fmt.Errorf("更新邮箱失败: %w", err)
	}

	return nil
}

// ========== 用户删除 ==========

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	// 1. 检查用户是否存在
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// 2. 执行软删除
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// ForceDeleteUser 强制删除用户（硬删除，管理员用）
func (s *UserService) ForceDeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.ForceDelete(ctx, id); err != nil {
		return fmt.Errorf("强制删除用户失败: %w", err)
	}
	return nil
}
