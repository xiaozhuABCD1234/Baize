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

type UserProfileRepository interface {
	Create(ctx context.Context, profile *model.UserProfile) error
	GetByID(ctx context.Context, id uint) (*model.UserProfile, error)
	GetByUserID(ctx context.Context, userID uint) (*model.UserProfile, error)
	GetByUserIDWithSelect(ctx context.Context, userID uint, preloads ...string) (*model.UserProfile, error)
	List(ctx context.Context, orderBy string) ([]model.UserProfile, error)
	ListByRegion(ctx context.Context, regionID uint) ([]model.UserProfile, error)
	ListMasters(ctx context.Context) ([]model.UserProfile, error)
	Update(ctx context.Context, profile *model.UserProfile) error
	UpdateByUserID(ctx context.Context, userID uint, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
	Upsert(ctx context.Context, profile *model.UserProfile) error
	WithTransaction(tx *gorm.DB) UserProfileRepository
}

type userProfileRepo struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepo{db: db}
}

func (r *userProfileRepo) WithTransaction(tx *gorm.DB) UserProfileRepository {
	return &userProfileRepo{db: tx}
}

func (r *userProfileRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *userProfileRepo) Create(ctx context.Context, profile *model.UserProfile) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.Create", start)
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *userProfileRepo) GetByID(ctx context.Context, id uint) (*model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.GetByID", start)
	var profile model.UserProfile
	err := r.db.WithContext(ctx).First(&profile, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepo) GetByUserID(ctx context.Context, userID uint) (*model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.GetByUserID", start)
	var profile model.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepo) GetByUserIDWithSelect(ctx context.Context, userID uint, preloads ...string) (*model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.GetByUserIDWithSelect", start)
	var profile model.UserProfile
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrUserProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepo) List(ctx context.Context, orderBy string) ([]model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.List", start)
	var profiles []model.UserProfile
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&profiles).Error
	return profiles, err
}

func (r *userProfileRepo) ListByRegion(ctx context.Context, regionID uint) ([]model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.ListByRegion", start)
	var profiles []model.UserProfile
	err := r.db.WithContext(ctx).
		Where("region_id = ?", regionID).
		Order("created_at desc").
		Find(&profiles).Error
	return profiles, err
}

func (r *userProfileRepo) ListMasters(ctx context.Context) ([]model.UserProfile, error) {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.ListMasters", start)
	var profiles []model.UserProfile
	err := r.db.WithContext(ctx).
		Where("is_master = ?", true).
		Order("created_at desc").
		Find(&profiles).Error
	return profiles, err
}

func (r *userProfileRepo) Update(ctx context.Context, profile *model.UserProfile) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.Update", start)
	result := r.db.WithContext(ctx).Save(profile)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserProfileNotFound
	}
	return nil
}

func (r *userProfileRepo) UpdateByUserID(ctx context.Context, userID uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.UpdateByUserID", start)
	result := r.db.WithContext(ctx).
		Model(&model.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserProfileNotFound
	}
	return nil
}

func (r *userProfileRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.UserProfile{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserProfileNotFound
	}
	return nil
}

func (r *userProfileRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.DeleteByUserID", start)
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.UserProfile{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrUserProfileNotFound
	}
	return nil
}

func (r *userProfileRepo) Upsert(ctx context.Context, profile *model.UserProfile) error {
	start := time.Now()
	defer r.logSlow(ctx, "UserProfile.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(profile).Error
}
