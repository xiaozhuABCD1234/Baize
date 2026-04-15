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

type WorkMediaRepository interface {
	Create(ctx context.Context, media *model.WorkMedia) error
	CreateBatch(ctx context.Context, mediaList []model.WorkMedia) error
	GetByID(ctx context.Context, id uint) (*model.WorkMedia, error)
	ListByWorkID(ctx context.Context, workID uint) ([]model.WorkMedia, error)
	ListImages(ctx context.Context, workID uint) ([]model.WorkMedia, error)
	ListVideos(ctx context.Context, workID uint) ([]model.WorkMedia, error)
	Update(ctx context.Context, media *model.WorkMedia) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	DeleteByWorkID(ctx context.Context, workID uint) error
	DeleteBatch(ctx context.Context, ids []uint) error
	Upsert(ctx context.Context, media *model.WorkMedia) error
	UpsertBatch(ctx context.Context, mediaList []model.WorkMedia) error
	WithTransaction(tx *gorm.DB) WorkMediaRepository
}

type workMediaRepo struct {
	db *gorm.DB
}

func NewWorkMediaRepository(db *gorm.DB) WorkMediaRepository {
	return &workMediaRepo{db: db}
}

func (r *workMediaRepo) WithTransaction(tx *gorm.DB) WorkMediaRepository {
	return &workMediaRepo{db: tx}
}

func (r *workMediaRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *workMediaRepo) Create(ctx context.Context, media *model.WorkMedia) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.Create", start)
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *workMediaRepo) CreateBatch(ctx context.Context, mediaList []model.WorkMedia) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.CreateBatch", start)
	if len(mediaList) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(mediaList, 100).Error
}

func (r *workMediaRepo) GetByID(ctx context.Context, id uint) (*model.WorkMedia, error) {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.GetByID", start)
	var media model.WorkMedia
	err := r.db.WithContext(ctx).First(&media, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrWorkMediaNotFound
		}
		return nil, err
	}
	return &media, nil
}

func (r *workMediaRepo) ListByWorkID(ctx context.Context, workID uint) ([]model.WorkMedia, error) {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.ListByWorkID", start)
	var mediaList []model.WorkMedia
	err := r.db.WithContext(ctx).
		Where("work_id = ?", workID).
		Order("sort_order asc, id asc").
		Find(&mediaList).Error
	return mediaList, err
}

func (r *workMediaRepo) ListImages(ctx context.Context, workID uint) ([]model.WorkMedia, error) {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.ListImages", start)
	var mediaList []model.WorkMedia
	err := r.db.WithContext(ctx).
		Where("work_id = ? AND media_type = ?", workID, model.MediaTypeImage).
		Order("sort_order asc, id asc").
		Find(&mediaList).Error
	return mediaList, err
}

func (r *workMediaRepo) ListVideos(ctx context.Context, workID uint) ([]model.WorkMedia, error) {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.ListVideos", start)
	var mediaList []model.WorkMedia
	err := r.db.WithContext(ctx).
		Where("work_id = ? AND media_type = ?", workID, model.MediaTypeVideo).
		Order("sort_order asc, id asc").
		Find(&mediaList).Error
	return mediaList, err
}

func (r *workMediaRepo) Update(ctx context.Context, media *model.WorkMedia) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.Update", start)
	result := r.db.WithContext(ctx).Save(media)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkMediaNotFound
	}
	return nil
}

func (r *workMediaRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.WorkMedia{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkMediaNotFound
	}
	return nil
}

func (r *workMediaRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.WorkMedia{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkMediaNotFound
	}
	return nil
}

func (r *workMediaRepo) DeleteByWorkID(ctx context.Context, workID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.DeleteByWorkID", start)
	result := r.db.WithContext(ctx).
		Where("work_id = ?", workID).
		Delete(&model.WorkMedia{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *workMediaRepo) DeleteBatch(ctx context.Context, ids []uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.DeleteBatch", start)
	if len(ids) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.WorkMedia{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *workMediaRepo) Upsert(ctx context.Context, media *model.WorkMedia) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(media).Error
}

func (r *workMediaRepo) UpsertBatch(ctx context.Context, mediaList []model.WorkMedia) error {
	start := time.Now()
	defer r.logSlow(ctx, "WorkMedia.UpsertBatch", start)
	if len(mediaList) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(mediaList, 100).Error
}
