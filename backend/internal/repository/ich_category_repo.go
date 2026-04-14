package repository

import (
	"context"
	"errors"
	"time"

	model "backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ICHCategoryRepository interface {
	Create(ctx context.Context, category *model.ICHCategory) error
	GetByID(ctx context.Context, id uint) (*model.ICHCategory, error)
	GetByIDWithChildren(ctx context.Context, id uint) (*model.ICHCategory, error)
	GetByName(ctx context.Context, name string) (*model.ICHCategory, error)
	List(ctx context.Context, orderBy string) ([]model.ICHCategory, error)
	ListRoot(ctx context.Context) ([]model.ICHCategory, error)
	ListByParentID(ctx context.Context, parentID uint) ([]model.ICHCategory, error)
	ListByRegionCode(ctx context.Context, regionCode string) ([]model.ICHCategory, error)
	ListActive(ctx context.Context) ([]model.ICHCategory, error)
	Update(ctx context.Context, category *model.ICHCategory) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	Upsert(ctx context.Context, category *model.ICHCategory) error
	UpsertBatch(ctx context.Context, categories []model.ICHCategory) error
	WithTransaction(tx *gorm.DB) ICHCategoryRepository
}

type ichCategoryRepo struct {
	db *gorm.DB
}

func NewICHCategoryRepository(db *gorm.DB) ICHCategoryRepository {
	return &ichCategoryRepo{db: db}
}

func (r *ichCategoryRepo) WithTransaction(tx *gorm.DB) ICHCategoryRepository {
	return &ichCategoryRepo{db: tx}
}

var ErrICHCategoryNotFound = errors.New("ICH category not found")

func (r *ichCategoryRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *ichCategoryRepo) Create(ctx context.Context, category *model.ICHCategory) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.Create", start)
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *ichCategoryRepo) GetByID(ctx context.Context, id uint) (*model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.GetByID", start)
	var category model.ICHCategory
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrICHCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *ichCategoryRepo) GetByIDWithChildren(ctx context.Context, id uint) (*model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.GetByIDWithChildren", start)
	var category model.ICHCategory
	err := r.db.WithContext(ctx).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order asc, name asc")
		}).
		First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrICHCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *ichCategoryRepo) GetByName(ctx context.Context, name string) (*model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.GetByName", start)
	var category model.ICHCategory
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrICHCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}

func (r *ichCategoryRepo) List(ctx context.Context, orderBy string) ([]model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.List", start)
	var categories []model.ICHCategory
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("sort_order asc, name asc")
	}
	err := query.Find(&categories).Error
	return categories, err
}

func (r *ichCategoryRepo) ListRoot(ctx context.Context) ([]model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.ListRoot", start)
	var categories []model.ICHCategory
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", 0).
		Order("sort_order asc, name asc").
		Find(&categories).Error
	return categories, err
}

func (r *ichCategoryRepo) ListByParentID(ctx context.Context, parentID uint) ([]model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.ListByParentID", start)
	var categories []model.ICHCategory
	err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Order("sort_order asc, name asc").
		Find(&categories).Error
	return categories, err
}

func (r *ichCategoryRepo) ListByRegionCode(ctx context.Context, regionCode string) ([]model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.ListByRegionCode", start)
	var categories []model.ICHCategory
	err := r.db.WithContext(ctx).
		Where("region_code = ?", regionCode).
		Order("sort_order asc, name asc").
		Find(&categories).Error
	return categories, err
}

func (r *ichCategoryRepo) ListActive(ctx context.Context) ([]model.ICHCategory, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.ListActive", start)
	var categories []model.ICHCategory
	err := r.db.WithContext(ctx).
		Where("status = ?", 1).
		Order("sort_order asc, name asc").
		Find(&categories).Error
	return categories, err
}

func (r *ichCategoryRepo) Update(ctx context.Context, category *model.ICHCategory) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.Update", start)
	result := r.db.WithContext(ctx).Save(category)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrICHCategoryNotFound
	}
	return nil
}

func (r *ichCategoryRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.ICHCategory{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrICHCategoryNotFound
	}
	return nil
}

func (r *ichCategoryRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.ICHCategory{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrICHCategoryNotFound
	}
	return nil
}

func (r *ichCategoryRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.ICHCategory{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrICHCategoryNotFound
	}
	return nil
}

func (r *ichCategoryRepo) Upsert(ctx context.Context, category *model.ICHCategory) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(category).Error
}

func (r *ichCategoryRepo) UpsertBatch(ctx context.Context, categories []model.ICHCategory) error {
	start := time.Now()
	defer r.logSlow(ctx, "ICHCategory.UpsertBatch", start)
	if len(categories) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(categories, 100).Error
}
