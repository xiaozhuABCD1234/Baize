package repository

import (
	"context"
	"errors"
	"time"

	model "backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id uint) (*model.Comment, error)
	GetByIDWithUser(ctx context.Context, id uint) (*model.Comment, error)
	List(ctx context.Context, orderBy string) ([]model.Comment, error)
	ListByWorkID(ctx context.Context, workID uint) ([]model.Comment, error)
	ListRootByWorkID(ctx context.Context, workID uint) ([]model.Comment, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.Comment, error)
	ListByParentID(ctx context.Context, parentID uint) ([]model.Comment, error)
	ListByRootID(ctx context.Context, rootID uint) ([]model.Comment, error)
	ListByStatus(ctx context.Context, status model.CommentStatus) ([]model.Comment, error)
	ListWithPagination(ctx context.Context, workID uint, page, pageSize int) ([]model.Comment, int64, error)
	Update(ctx context.Context, comment *model.Comment) error
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
	UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error
	IncrementLikeCount(ctx context.Context, id uint, delta int) error
	Delete(ctx context.Context, id uint) error
	ForceDelete(ctx context.Context, id uint) error
	DeleteByWorkID(ctx context.Context, workID uint) error
	Upsert(ctx context.Context, comment *model.Comment) error
	WithTransaction(tx *gorm.DB) CommentRepository
}

type commentRepo struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepo{db: db}
}

func (r *commentRepo) WithTransaction(tx *gorm.DB) CommentRepository {
	return &commentRepo{db: tx}
}

var ErrCommentNotFound = errors.New("comment not found")

func (r *commentRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *commentRepo) Create(ctx context.Context, comment *model.Comment) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.Create", start)
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepo) GetByID(ctx context.Context, id uint) (*model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.GetByID", start)
	var comment model.Comment
	err := r.db.WithContext(ctx).First(&comment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepo) GetByIDWithUser(ctx context.Context, id uint) (*model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.GetByIDWithUser", start)
	var comment model.Comment
	err := r.db.WithContext(ctx).Preload("User").First(&comment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepo) List(ctx context.Context, orderBy string) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.List", start)
	var comments []model.Comment
	query := r.db.WithContext(ctx).Preload("User")
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListByWorkID(ctx context.Context, workID uint) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListByWorkID", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("work_id = ?", workID).
		Order("created_at desc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListRootByWorkID(ctx context.Context, workID uint) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListRootByWorkID", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("work_id = ? AND parent_id = 0", workID).
		Order("created_at desc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListByUserID(ctx context.Context, userID uint) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListByUserID", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListByParentID(ctx context.Context, parentID uint) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListByParentID", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("parent_id = ?", parentID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListByRootID(ctx context.Context, rootID uint) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListByRootID", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("root_id = ?", rootID).
		Order("created_at asc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListByStatus(ctx context.Context, status model.CommentStatus) ([]model.Comment, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListByStatus", start)
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("status = ?", status).
		Order("created_at desc").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepo) ListWithPagination(ctx context.Context, workID uint, page, pageSize int) ([]model.Comment, int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ListWithPagination", start)
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("work_id = ?", workID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("work_id = ?", workID).
		Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error
	return comments, total, err
}

func (r *commentRepo) Update(ctx context.Context, comment *model.Comment) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.Update", start)
	result := r.db.WithContext(ctx).Save(comment)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.UpdateFields", start)
	result := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.UpdateStatus", start)
	result := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) IncrementLikeCount(ctx context.Context, id uint, delta int) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.IncrementLikeCount", start)
	result := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ?", id).
		Update("like_count", gorm.Expr("like_count+?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.Delete", start)
	result := r.db.WithContext(ctx).Delete(&model.Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) ForceDelete(ctx context.Context, id uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.ForceDelete", start)
	result := r.db.WithContext(ctx).Unscoped().Delete(&model.Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func (r *commentRepo) DeleteByWorkID(ctx context.Context, workID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.DeleteByWorkID", start)
	result := r.db.WithContext(ctx).
		Where("work_id = ?", workID).
		Delete(&model.Comment{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *commentRepo) Upsert(ctx context.Context, comment *model.Comment) error {
	start := time.Now()
	defer r.logSlow(ctx, "Comment.Upsert", start)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(comment).Error
}
