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

type CraftRepository interface {
	Create(ctx context.Context, craft *model.Craft) error
	GetByID(ctx context.Context, id uint) (*model.Craft, error)
	GetByIDWithCategory(ctx context.Context, id uint) (*model.Craft, error)
	GetByName(ctx context.Context, name string) (*model.Craft, error)
	List(ctx context.Context, orderBy string) ([]model.Craft, error)
	ListWithCategory(ctx context.Context, orderBy string) ([]model.Craft, error)
	ListByCategoryID(ctx context.Context, categoryID uint) ([]model.Craft, error)
	ListByDifficulty(ctx context.Context, difficulty int8) ([]model.Craft, error)
	Update(ctx context.Context, craft *model.Craft) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	Upsert(ctx context.Context, craft *model.Craft) error
	UpsertBatch(ctx context.Context, crafts []model.Craft) error
	WithTransaction(tx *gorm.DB) CraftRepository
}

type craftRepo struct {
	db *gorm.DB
}

func NewCraftRepository(db *gorm.DB) CraftRepository {
	return &craftRepo{db: db}
}

func (r *craftRepo) WithTransaction(tx *gorm.DB) CraftRepository {
	return &craftRepo{db: tx}
}

func (r *craftRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *craftRepo) Create(ctx context.Context, craft *model.Craft) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.Create", start)
	return r.db.WithContext(ctx).Create(craft).Error
}

func (r *craftRepo) GetByID(ctx context.Context, id uint) (*model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.GetByID", start)
	var craft model.Craft
	err := r.db.WithContext(ctx).First(&craft, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrCraftNotFound
		}
		return nil, err
	}
	return &craft, nil
}

func (r *craftRepo) GetByIDWithCategory(ctx context.Context, id uint) (*model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.GetByIDWithCategory", start)
	var craft model.Craft
	err := r.db.WithContext(ctx).Preload("Category").First(&craft, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrCraftNotFound
		}
		return nil, err
	}
	return &craft, nil
}

func (r *craftRepo) GetByName(ctx context.Context, name string) (*model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.GetByName", start)
	var craft model.Craft
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&craft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrCraftNotFound
		}
		return nil, err
	}
	return &craft, nil
}

func (r *craftRepo) List(ctx context.Context, orderBy string) ([]model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.List", start)
	var crafts []model.Craft
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("name asc")
	}
	err := query.Find(&crafts).Error
	return crafts, err
}

func (r *craftRepo) ListWithCategory(ctx context.Context, orderBy string) ([]model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.ListWithCategory", start)
	var crafts []model.Craft
	query := r.db.WithContext(ctx).Preload("Category")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("name asc")
	}
	err := query.Find(&crafts).Error
	return crafts, err
}

func (r *craftRepo) ListByCategoryID(ctx context.Context, categoryID uint) ([]model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.ListByCategoryID", start)
	var crafts []model.Craft
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("name asc").
		Find(&crafts).Error
	return crafts, err
}

func (r *craftRepo) ListByDifficulty(ctx context.Context, difficulty int8) ([]model.Craft, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.ListByDifficulty", start)
	var crafts []model.Craft
	err := r.db.WithContext(ctx).
		Where("difficulty = ?", difficulty).
		Order("name asc").
		Find(&crafts).Error
	return crafts, err
}

func (r *craftRepo) Update(ctx context.Context, craft *model.Craft) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.Update", start)
	result := r.db.WithContext(ctx).Save(craft)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrCraftNotFound
	}
	return nil
}

func (r *craftRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.Craft{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrCraftNotFound
	}
	return nil
}

func (r *craftRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.Craft{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrCraftNotFound
	}
	return nil
}

func (r *craftRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.Craft{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrCraftNotFound
	}
	return nil
}

func (r *craftRepo) Upsert(ctx context.Context, craft *model.Craft) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(craft).Error
}

func (r *craftRepo) UpsertBatch(ctx context.Context, crafts []model.Craft) error {
	start := time.Now()
	defer r.logSlow(ctx, "Craft.UpsertBatch", start)
	if len(crafts) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(crafts, 100).Error
}
