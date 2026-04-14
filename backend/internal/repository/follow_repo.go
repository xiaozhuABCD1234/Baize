package repository

import (
	"context"
	"errors"
	"time"

	model "backend/internal/models"

	"gorm.io/gorm"
)

type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, followerID, followingID uint) error
	Exists(ctx context.Context, followerID, followingID uint) (bool, error)
	GetFollowingList(ctx context.Context, userID uint, orderBy string) ([]model.Follow, error)
	GetFollowerList(ctx context.Context, userID uint, orderBy string) ([]model.Follow, error)
	CountFollowing(ctx context.Context, userID uint) (int64, error)
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error)
	WithTransaction(tx *gorm.DB) FollowRepository
}

type followRepo struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepo{db: db}
}

func (r *followRepo) WithTransaction(tx *gorm.DB) FollowRepository {
	return &followRepo{db: tx}
}

var ErrFollowNotFound = errors.New("follow relationship not found")

func (r *followRepo) logSlow(ctx context.Context, op string, start time.Time) {
	elapsed := time.Since(start)
	if elapsed > SlowThreshold {
		// Log slow query
	}
}

func (r *followRepo) Create(ctx context.Context, follow *model.Follow) error {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.Create", start)
	return r.db.WithContext(ctx).Create(follow).Error
}

func (r *followRepo) Delete(ctx context.Context, followerID, followingID uint) error {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.Delete", start)
	result := r.db.WithContext(ctx).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&model.Follow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFollowNotFound
	}
	return nil
}

func (r *followRepo) Exists(ctx context.Context, followerID, followingID uint) (bool, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.Exists", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *followRepo) GetFollowingList(ctx context.Context, userID uint, orderBy string) ([]model.Follow, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.GetFollowingList", start)
	var follows []model.Follow
	query := r.db.WithContext(ctx).
		Where("follower_id = ?", userID)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&follows).Error
	return follows, err
}

func (r *followRepo) GetFollowerList(ctx context.Context, userID uint, orderBy string) ([]model.Follow, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.GetFollowerList", start)
	var follows []model.Follow
	query := r.db.WithContext(ctx).
		Where("following_id = ?", userID)
	if orderBy != "" {
		query = query.Order(orderBy)
	} else {
		query = query.Order("created_at desc")
	}
	err := query.Find(&follows).Error
	return follows, err
}

func (r *followRepo) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.CountFollowing", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepo) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	start := time.Now()
	defer r.logSlow(ctx, "Follow.CountFollowers", start)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Follow{}).
		Where("following_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepo) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	return r.Exists(ctx, followerID, followingID)
}
