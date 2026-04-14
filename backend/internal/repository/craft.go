package repository

import (
	"context"
	"errors"

	model "backend/internal/models"

	"gorm.io/gorm"
)

type CraftRepositoryInterface interface {
	Create(ctx context.Context, craft *model.Craft) error
	GetByID(ctx context.Context, id uint) (*model.Craft, error)
	GetByName(ctx context.Context, name string) (*model.Craft, error)
	List(ctx context.Context) ([]model.Craft, error)
	ListByCategory(ctx context.Context, categoryID uint) ([]model.Craft, error)
	ListByDifficulty(ctx context.Context, difficulty int8) ([]model.Craft, error)
	ListWithPagination(ctx context.Context, page, pageSize int) ([]model.Craft, int64, error)
	Update(ctx context.Context, craft *model.Craft) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
}

type CraftRepository struct {
	db *gorm.DB
}

func NewCraftRepository(db *gorm.DB) *CraftRepository {
	return &CraftRepository{db: db}
}

var ErrCraftNotFound = errors.New("工艺不存在或已被删除")

func (r *CraftRepository) Create(ctx context.Context, craft *model.Craft) error {
	return r.db.WithContext(ctx).Create(craft).Error
}

func (r *CraftRepository) GetByID(ctx context.Context, id uint) (*model.Craft, error) {
	var craft model.Craft
	err := r.db.WithContext(ctx).First(&craft, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &craft, nil
}

func (r *CraftRepository) GetByName(ctx context.Context, name string) (*model.Craft, error) {
	var craft model.Craft
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&craft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &craft, nil
}

func (r *CraftRepository) List(ctx context.Context) ([]model.Craft, error) {
	var crafts []model.Craft
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&crafts).Error
	return crafts, err
}

func (r *CraftRepository) ListByCategory(ctx context.Context, categoryID uint) ([]model.Craft, error) {
	var crafts []model.Craft
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("created_at desc").
		Find(&crafts).Error
	return crafts, err
}

func (r *CraftRepository) ListByDifficulty(ctx context.Context, difficulty int8) ([]model.Craft, error) {
	var crafts []model.Craft
	err := r.db.WithContext(ctx).
		Where("difficulty = ?", difficulty).
		Order("created_at desc").
		Find(&crafts).Error
	return crafts, err
}

func (r *CraftRepository) ListWithPagination(ctx context.Context, page, pageSize int) ([]model.Craft, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Craft{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var crafts []model.Craft
	err := r.db.WithContext(ctx).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at desc").
		Find(&crafts).Error

	return crafts, total, err
}

func (r *CraftRepository) Update(ctx context.Context, craft *model.Craft) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", craft.ID).
		Updates(craft)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCraftNotFound
	}
	return nil
}

func (r *CraftRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.Craft{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCraftNotFound
	}
	return nil
}

func (r *CraftRepository) ForceDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().
		Where("id = ?", id).
		Delete(&model.Craft{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCraftNotFound
	}
	return nil
}
