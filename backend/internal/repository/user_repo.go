package repository

import (
	"context"
	"errors"
	"time"

	apperrs "backend/internal/errors"
	model "backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	CreateWithProfile(ctx context.Context, user *model.User, profile *model.UserProfile) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByIDWithProfile(ctx context.Context, id uint) (*model.User, error)
	GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	List(ctx context.Context, orderBy string) ([]model.User, error)
	ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]model.User, int64, error)
	ListByUserType(ctx context.Context, userType model.UserType) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error
	UpdateEmail(ctx context.Context, id uint, email string) error
	UpdateStatus(ctx context.Context, id uint, status model.UserStatus) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	Upsert(ctx context.Context, user *model.User) error
	WithTransaction(tx *gorm.DB) UserRepository
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) WithTransaction(tx *gorm.DB) UserRepository {
	return &userRepo{db: tx}
}

func (r *userRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.Create", start)
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) CreateWithProfile(ctx context.Context, user *model.User, profile *model.UserProfile) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.CreateWithProfile", start)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.ID
		if err := tx.Create(profile).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.GetByID", start)
	var user model.User
	err := r.db.WithContext(ctx).Preload("Profile").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByIDWithProfile(ctx context.Context, id uint) (*model.User, error) {
	return r.GetByIDWithSelect(ctx, id, "Profile")
}

func (r *userRepo) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.GetByIDWithSelect", start)
	var user model.User
	query := r.db.WithContext(ctx).Where("id = ?", id)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.GetByEmail", start)
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.GetByUsername", start)
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.GetByPhone", start)
	var user model.User
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) List(ctx context.Context, orderBy string) ([]model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.List", start)
	var users []model.User
	query := r.db.WithContext(ctx).Preload("Profile")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&users).Error
	return users, err
}

func (r *userRepo) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]model.User, int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.ListWithPagination", start)
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	query := r.db.WithContext(ctx).Preload("Profile")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *userRepo) ListByUserType(ctx context.Context, userType model.UserType) ([]model.User, error) {
	start := time.Now()
	defer r.logSlow(ctx, "User.ListByUserType", start)
	var users []model.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Where("user_type = ?", userType).
		Order("created_at desc").
		Find(&users).Error
	return users, err
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.Update", start)
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.UpdatePassword", start)
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdateEmail(ctx context.Context, id uint, email string) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.UpdateEmail", start)
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("email", email)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) UpdateStatus(ctx context.Context, id uint, status model.UserStatus) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.UpdateStatus", start)
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) Upsert(ctx context.Context, user *model.User) error {
	start := time.Now()
	defer r.logSlow(ctx, "User.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(user).Error
}
