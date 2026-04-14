package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const SlowThreshold = 200 * time.Millisecond

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) WithTransaction(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: tx}
}

func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

func (r *BaseRepository[T]) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		fmt.Printf("[SLOW] %s took %v\n", op, elapsed)
	}
}

func (r *BaseRepository[T]) Create(ctx context.Context, data *T) error {
	start := time.Now()
	defer r.logSlow(ctx, "Create", start)
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *BaseRepository[T]) CreateInBatches(ctx context.Context, data []T, batchSize int) error {
	if len(data) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	start := time.Now()
	defer r.logSlow(ctx, "CreateInBatches", start)
	return r.db.WithContext(ctx).CreateInBatches(data, batchSize).Error
}

func (r *BaseRepository[T]) Upsert(ctx context.Context, data *T) error {
	start := time.Now()
	defer r.logSlow(ctx, "Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(data).Error
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	start := time.Now()
	defer r.logSlow(ctx, "GetByID", start)
	var model T
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *BaseRepository[T]) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*T, error) {
	start := time.Now()
	defer r.logSlow(ctx, "GetByIDWithSelect", start)
	var model T
	query := r.db.WithContext(ctx).Where("id = ?", id)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *BaseRepository[T]) List(ctx context.Context, orderBy string) ([]T, error) {
	start := time.Now()
	defer r.logSlow(ctx, "List", start)
	var models []T
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	}
	err := query.Find(&models).Error
	return models, err
}

func (r *BaseRepository[T]) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]T, int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "ListWithPagination", start)
	var total int64
	if err := r.db.WithContext(ctx).Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []T
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error
	return models, total, err
}

func (r *BaseRepository[T]) Update(ctx context.Context, data *T) error {
	start := time.Now()
	defer r.logSlow(ctx, "Update", start)
	result := r.db.WithContext(ctx).Save(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "UpdateFields", start)
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Delete", start)
	result := r.db.WithContext(ctx).Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (r *BaseRepository[T]) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(new(T), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	return nil
}

func (r *BaseRepository[T]) Count(ctx context.Context, cond interface{}, args ...interface{}) (int64, error) {
	var count int64
	start := time.Now()
	defer r.logSlow(ctx, "Count", start)
	err := r.db.WithContext(ctx).Model(new(T)).Where(cond, args...).Count(&count).Error
	return count, err
}

func (r *BaseRepository[T]) Exists(ctx context.Context, cond interface{}, args ...interface{}) (bool, error) {
	var count int64
	start := time.Now()
	defer r.logSlow(ctx, "Exists", start)
	err := r.db.WithContext(ctx).Model(new(T)).Where(cond, args...).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BaseRepository[T]) FindOne(ctx context.Context, cond interface{}, args ...interface{}) (*T, error) {
	start := time.Now()
	defer r.logSlow(ctx, "FindOne", start)
	var model T
	err := r.db.WithContext(ctx).Where(cond, args...).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &model, nil
}

func (r *BaseRepository[T]) FindAll(ctx context.Context, cond interface{}, args ...interface{}) ([]T, error) {
	start := time.Now()
	defer r.logSlow(ctx, "FindAll", start)
	var models []T
	err := r.db.WithContext(ctx).Where(cond, args...).Find(&models).Error
	return models, err
}
