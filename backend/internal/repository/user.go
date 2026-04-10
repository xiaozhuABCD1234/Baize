package repository

import (
	"context"
	"errors"

	model "backend/internal/models"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	ListWithPagination(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error
	UpdateEmail(ctx context.Context, id uint, email string) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ========== CREATE ==========

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return gorm.G[model.User](r.db).Create(ctx, user)
}

// ========== READ ==========

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := gorm.G[model.User](r.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱查询用户（登录用）
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := gorm.G[model.User](r.db).
		Where("email = ?", email).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// List 查询所有用户
func (r *UserRepository) List(ctx context.Context) ([]model.User, error) {
	return gorm.G[model.User](r.db).
		Order("created_at desc").
		Find(ctx)
}

// ListWithPagination 分页查询
func (r *UserRepository) ListWithPagination(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	var total int64

	// 统计总数
	err := gorm.G[model.User](r.db).
		Raw("SELECT COUNT(*) FROM users").
		Scan(ctx, &total)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	users, err := gorm.G[model.User](r.db).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at desc").
		Find(ctx)

	return users, total, err
}

// ========== UPDATE ==========

// Update 更新用户（全字段）
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	rowsAffected, err := gorm.G[model.User](r.db).
		Where("id = ?", user.ID).
		Updates(ctx, *user)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdatePassword 只更新密码
func (r *UserRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	rowsAffected, err := gorm.G[model.User](r.db).
		Where("id = ?", id).
		Update(ctx, "password", hashedPassword)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateEmail 只更新邮箱
func (r *UserRepository) UpdateEmail(ctx context.Context, id uint, email string) error {
	rowsAffected, err := gorm.G[model.User](r.db).
		Where("id = ?", id).
		Update(ctx, "email", email)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ========== DELETE ==========

var ErrUserNotFound = errors.New("用户不存在或已被删除")

// Delete 软删除（使用 gorm.Model 的 DeletedAt）
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[model.User](r.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ForceDelete 硬删除
func (r *UserRepository) ForceDelete(ctx context.Context, id uint) error {
	rowsAffected, err := gorm.G[model.User](r.db.Unscoped()).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
