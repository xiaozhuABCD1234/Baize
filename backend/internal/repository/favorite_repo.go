package repository

import (
	"context"
	"errors"
	"time"

	apperrs "backend/internal/errors"
	model "backend/internal/models"

	"gorm.io/gorm"
)

type FavoriteRepository interface {
	Create(ctx context.Context, favorite *model.Favorite) error
	GetByID(ctx context.Context, id uint) (*model.Favorite, error)
	GetByUserAndWork(ctx context.Context, userID, workID uint) (*model.Favorite, error)
	List(ctx context.Context, orderBy string) ([]model.Favorite, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.Favorite, error)
	ListByWorkID(ctx context.Context, workID uint) ([]model.Favorite, error)
	ListByFolderID(ctx context.Context, userID, folderID uint) ([]model.Favorite, error)
	ListByUserIDWithWorks(ctx context.Context, userID uint, orderBy string) ([]model.Favorite, error)
	ListWithPagination(ctx context.Context, userID uint, page, pageSize int) ([]model.Favorite, int64, error)
	Update(ctx context.Context, favorite *model.Favorite) error
	UpdateFolder(ctx context.Context, id uint, folderID uint) error
	Delete(ctx context.Context, id uint) error
	DeleteByUserAndWork(ctx context.Context, userID, workID uint) error
	DeleteByWorkID(ctx context.Context, workID uint) error
	Exists(ctx context.Context, userID, workID uint) (bool, error)
	CountByWorkID(ctx context.Context, workID uint) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	WithTransaction(tx *gorm.DB) FavoriteRepository
}

type favoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepo{db: db}
}

func (r *favoriteRepo) WithTransaction(tx *gorm.DB) FavoriteRepository {
	return &favoriteRepo{db: tx}
}

func (r *favoriteRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *favoriteRepo) Create(ctx context.Context, favorite *model.Favorite) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.Create", start)
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *favoriteRepo) GetByID(ctx context.Context, id uint) (*model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.GetByID", start)
	var favorite model.Favorite
	err := r.db.WithContext(ctx).First(&favorite, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrFavoriteNotFound
		}
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepo) GetByUserAndWork(ctx context.Context, userID, workID uint) (*model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.GetByUserAndWork", start)
	var favorite model.Favorite
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND work_id = ?", userID, workID).
		First(&favorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrFavoriteNotFound
		}
		return nil, err
	}
	return &favorite, nil
}

func (r *favoriteRepo) List(ctx context.Context, orderBy string) ([]model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.List", start)
	var favorites []model.Favorite
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepo) ListByUserID(ctx context.Context, userID uint) ([]model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.ListByUserID", start)
	var favorites []model.Favorite
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepo) ListByWorkID(ctx context.Context, workID uint) ([]model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.ListByWorkID", start)
	var favorites []model.Favorite
	err := r.db.WithContext(ctx).
		Where("work_id = ?", workID).
		Order("created_at desc").
		Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepo) ListByFolderID(ctx context.Context, userID, folderID uint) ([]model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.ListByFolderID", start)
	var favorites []model.Favorite
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND folder_id = ?", userID, folderID).
		Order("created_at desc").
		Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepo) ListByUserIDWithWorks(ctx context.Context, userID uint, orderBy string) ([]model.Favorite, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.ListByUserIDWithWorks", start)
	var favorites []model.Favorite
	query := r.db.WithContext(ctx).Preload("Work")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Where("user_id = ?", userID).Find(&favorites).Error
	return favorites, err
}

func (r *favoriteRepo) ListWithPagination(ctx context.Context, userID uint, page, pageSize int) ([]model.Favorite, int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.ListWithPagination", start)
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Favorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var favorites []model.Favorite
	err := r.db.WithContext(ctx).
		Preload("Work").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&favorites).Error
	return favorites, total, err
}

func (r *favoriteRepo) Update(ctx context.Context, favorite *model.Favorite) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.Update", start)
	result := r.db.WithContext(ctx).Save(favorite)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrFavoriteNotFound
	}
	return nil
}

func (r *favoriteRepo) UpdateFolder(ctx context.Context, id uint, folderID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.UpdateFolder", start)
	result := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("id = ?", id).
		Update("folder_id", folderID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrFavoriteNotFound
	}
	return nil
}

func (r *favoriteRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.Favorite{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrFavoriteNotFound
	}
	return nil
}

func (r *favoriteRepo) DeleteByUserAndWork(ctx context.Context, userID, workID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.DeleteByUserAndWork", start)
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Delete(&model.Favorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrFavoriteNotFound
	}
	return nil
}

func (r *favoriteRepo) DeleteByWorkID(ctx context.Context, workID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.DeleteByWorkID", start)
	result := r.db.WithContext(ctx).
		Where("work_id = ?", workID).
		Delete(&model.Favorite{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *favoriteRepo) Exists(ctx context.Context, userID, workID uint) (bool, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.Exists", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *favoriteRepo) CountByWorkID(ctx context.Context, workID uint) (int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.CountByWorkID", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("work_id = ?", workID).
		Count(&count).Error
	return count, err
}

func (r *favoriteRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Favorite.CountByUserID", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
