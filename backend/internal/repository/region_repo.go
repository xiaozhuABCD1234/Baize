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

type RegionRepository interface {
	Create(ctx context.Context, region *model.Region) error
	GetByID(ctx context.Context, id uint) (*model.Region, error)
	GetByCode(ctx context.Context, code string) (*model.Region, error)
	GetByIDWithChildren(ctx context.Context, id uint) (*model.Region, error)
	List(ctx context.Context, orderBy string) ([]model.Region, error)
	ListRoot(ctx context.Context) ([]model.Region, error)
	ListByParentID(ctx context.Context, parentID uint) ([]model.Region, error)
	ListByLevel(ctx context.Context, level int8) ([]model.Region, error)
	ListHeritageCenters(ctx context.Context) ([]model.Region, error)
	Update(ctx context.Context, region *model.Region) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	Upsert(ctx context.Context, region *model.Region) error
	UpsertBatch(ctx context.Context, regions []model.Region) error
	WithTransaction(tx *gorm.DB) RegionRepository
}

type regionRepo struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) RegionRepository {
	return &regionRepo{db: db}
}

func (r *regionRepo) WithTransaction(tx *gorm.DB) RegionRepository {
	return &regionRepo{db: tx}
}

func (r *regionRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *regionRepo) Create(ctx context.Context, region *model.Region) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.Create", start)
	return r.db.WithContext(ctx).Create(region).Error
}

func (r *regionRepo) GetByID(ctx context.Context, id uint) (*model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.GetByID", start)
	var region model.Region
	err := r.db.WithContext(ctx).First(&region, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrRegionNotFound
		}
		return nil, err
	}
	return &region, nil
}

func (r *regionRepo) GetByCode(ctx context.Context, code string) (*model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.GetByCode", start)
	var region model.Region
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&region).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrRegionNotFound
		}
		return nil, err
	}
	return &region, nil
}

func (r *regionRepo) GetByIDWithChildren(ctx context.Context, id uint) (*model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.GetByIDWithChildren", start)
	var region model.Region
	err := r.db.WithContext(ctx).Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Order("name asc")
	}).First(&region, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrRegionNotFound
		}
		return nil, err
	}
	return &region, nil
}

func (r *regionRepo) List(ctx context.Context, orderBy string) ([]model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.List", start)
	var regions []model.Region
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("code asc")
	}
	err := query.Find(&regions).Error
	return regions, err
}

func (r *regionRepo) ListRoot(ctx context.Context) ([]model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.ListRoot", start)
	var regions []model.Region
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", 0).
		Order("code asc").
		Find(&regions).Error
	return regions, err
}

func (r *regionRepo) ListByParentID(ctx context.Context, parentID uint) ([]model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.ListByParentID", start)
	var regions []model.Region
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("name asc").
		Find(&regions).Error
	return regions, err
}

func (r *regionRepo) ListByLevel(ctx context.Context, level int8) ([]model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.ListByLevel", start)
	var regions []model.Region
	err := r.db.WithContext(ctx).
		Where("level = ?", level).
		Order("code asc").
		Find(&regions).Error
	return regions, err
}

func (r *regionRepo) ListHeritageCenters(ctx context.Context) ([]model.Region, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Region.ListHeritageCenters", start)
	var regions []model.Region
	err := r.db.WithContext(ctx).
		Where("is_heritage_center = ?", true).
		Order("name asc").
		Find(&regions).Error
	return regions, err
}

func (r *regionRepo) Update(ctx context.Context, region *model.Region) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.Update", start)
	result := r.db.WithContext(ctx).Save(region)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrRegionNotFound
	}
	return nil
}

func (r *regionRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.Region{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrRegionNotFound
	}
	return nil
}

func (r *regionRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.Region{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrRegionNotFound
	}
	return nil
}

func (r *regionRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.Region{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrRegionNotFound
	}
	return nil
}

func (r *regionRepo) Upsert(ctx context.Context, region *model.Region) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(region).Error
}

func (r *regionRepo) UpsertBatch(ctx context.Context, regions []model.Region) error {
	start := time.Now()
	defer r.logSlow(ctx, "Region.UpsertBatch", start)
	if len(regions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(regions, 100).Error
}
