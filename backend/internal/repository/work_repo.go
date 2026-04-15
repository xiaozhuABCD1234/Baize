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

type WorkRepository interface {
	Create(ctx context.Context, work *model.Work) error
	GetByID(ctx context.Context, id uint) (*model.Work, error)
	GetByIDWithAll(ctx context.Context, id uint) (*model.Work, error)
	GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*model.Work, error)
	List(ctx context.Context, orderBy string) ([]model.Work, error)
	ListWithAll(ctx context.Context, orderBy string) ([]model.Work, error)
	ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]model.Work, int64, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.Work, error)
	ListByCraftID(ctx context.Context, craftID uint) ([]model.Work, error)
	ListByCategoryID(ctx context.Context, categoryID uint) ([]model.Work, error)
	ListByStatus(ctx context.Context, status model.WorkStatus) ([]model.Work, error)
	ListPublished(ctx context.Context, orderBy string) ([]model.Work, error)
	ListTop(ctx context.Context, limit int) ([]model.Work, error)
	ListRecommended(ctx context.Context, limit int) ([]model.Work, error)
	Update(ctx context.Context, work *model.Work) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	UpdateStatus(ctx context.Context, id uint, status model.WorkStatus) error
	IncrementCount(ctx context.Context, id uint, field string, delta int) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	Upsert(ctx context.Context, work *model.Work) error
	UpsertBatch(ctx context.Context, works []model.Work) error
	WithTransaction(tx *gorm.DB) WorkRepository
}

type workRepo struct {
	db *gorm.DB
}

func NewWorkRepository(db *gorm.DB) WorkRepository {
	return &workRepo{db: db}
}

func (r *workRepo) WithTransaction(tx *gorm.DB) WorkRepository {
	return &workRepo{db: tx}
}

func (r *workRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *workRepo) Create(ctx context.Context, work *model.Work) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.Create", start)
	return r.db.WithContext(ctx).Create(work).Error
}

func (r *workRepo) GetByID(ctx context.Context, id uint) (*model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.GetByID", start)
	var work model.Work
	err := r.db.WithContext(ctx).First(&work, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrWorkNotFound
		}
		return nil, err
	}
	return &work, nil
}

func (r *workRepo) GetByIDWithAll(ctx context.Context, id uint) (*model.Work, error) {
	return r.GetByIDWithSelect(ctx, id, "User", "Craft", "Media")
}

func (r *workRepo) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.GetByIDWithSelect", start)
	var work model.Work
	query := r.db.WithContext(ctx).Where("id = ?", id)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	err := query.First(&work).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrs.ErrWorkNotFound
		}
		return nil, err
	}
	return &work, nil
}

func (r *workRepo) List(ctx context.Context, orderBy string) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.List", start)
	var works []model.Work
	query := r.db.WithContext(ctx)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&works).Error
	return works, err
}

func (r *workRepo) ListWithAll(ctx context.Context, orderBy string) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListWithAll", start)
	var works []model.Work
	query := r.db.WithContext(ctx).Preload("User").Preload("Craft").Preload("Media")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&works).Error
	return works, err
}

func (r *workRepo) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]model.Work, int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListWithPagination", start)
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Work{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var works []model.Work
	query := r.db.WithContext(ctx).Preload("User").Preload("Craft").Preload("Media")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&works).Error
	return works, total, err
}

func (r *workRepo) ListByUserID(ctx context.Context, userID uint) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListByUserID", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&works).Error
	return works, err
}

func (r *workRepo) ListByCraftID(ctx context.Context, craftID uint) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListByCraftID", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Where("craft_id = ?", craftID).
		Order("created_at desc").
		Find(&works).Error
	return works, err
}

func (r *workRepo) ListByCategoryID(ctx context.Context, categoryID uint) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListByCategoryID", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Where("category_id = ?", categoryID).
		Order("created_at desc").
		Find(&works).Error
	return works, err
}

func (r *workRepo) ListByStatus(ctx context.Context, status model.WorkStatus) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListByStatus", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at desc").
		Find(&works).Error
	return works, err
}

func (r *workRepo) ListPublished(ctx context.Context, orderBy string) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListPublished", start)
	var works []model.Work
	query := r.db.WithContext(ctx).
		Where("status = ?", model.WorkStatusPublished)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("published_at desc, created_at desc")
	}
	err := query.Find(&works).Error
	return works, err
}

func (r *workRepo) ListTop(ctx context.Context, limit int) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListTop", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Craft").
		Where("is_top = ? AND status = ?", true, model.WorkStatusPublished).
		Order("weight desc, created_at desc").
		Limit(limit).
		Find(&works).Error
	return works, err
}

func (r *workRepo) ListRecommended(ctx context.Context, limit int) ([]model.Work, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ListRecommended", start)
	var works []model.Work
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Craft").
		Where("is_recommended = ? AND status = ?", true, model.WorkStatusPublished).
		Order("weight desc, created_at desc").
		Limit(limit).
		Find(&works).Error
	return works, err
}

func (r *workRepo) Update(ctx context.Context, work *model.Work) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.Update", start)
	result := r.db.WithContext(ctx).Save(work)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.Work{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) UpdateStatus(ctx context.Context, id uint, status model.WorkStatus) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.UpdateStatus", start)
	result := r.db.WithContext(ctx).
		Model(&model.Work{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.IncrementCount", start)
	validFields := map[string]bool{
		"view_count": true, "like_count": true, "comment_count": true,
		"favorite_count": true, "share_count": true,
	}
	if !validFields[field] {
		return errors.New("invalid field name")
	}
	result := r.db.WithContext(ctx).
		Model(&model.Work{}).
		Where("id = ?", id).
		Update(field, gorm.Expr(field+"+?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.Work{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.Work{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrs.ErrWorkNotFound
	}
	return nil
}

func (r *workRepo) Upsert(ctx context.Context, work *model.Work) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(work).Error
}

func (r *workRepo) UpsertBatch(ctx context.Context, works []model.Work) error {
	start := time.Now()
	defer r.logSlow(ctx, "Work.UpsertBatch", start)
	if len(works) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(works, 100).Error
}
