package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	apperrs "backend/internal/errors"
	model "backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
)

var (
	ErrFollowNotFound   = apperrs.ErrFollowNotFound
	ErrCannotFollowSelf = apperrs.ErrCannotFollowSelf
	ErrAlreadyFollowing = apperrs.ErrAlreadyFollowing
	ErrNotFollowing     = apperrs.ErrNotFollowing
)

type FollowService interface {
	Create(ctx context.Context, followerID, followingID uint) (*model.FollowResponse, error)
	Delete(ctx context.Context, followerID, followingID uint) error
	IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error)
	GetFollowingList(ctx context.Context, userID uint) ([]model.FollowResponse, error)
	GetFollowerList(ctx context.Context, userID uint) ([]model.FollowResponse, error)
	GetFollowingCount(ctx context.Context, userID uint) (int64, error)
	GetFollowerCount(ctx context.Context, userID uint) (int64, error)
}

type followService struct {
	followRepo repository.FollowRepository
	userRepo   repository.UserRepository
	logger     *slog.Logger
}

func NewFollowService(
	followRepo repository.FollowRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) FollowService {
	return &followService{
		followRepo: followRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

func (s *followService) Create(ctx context.Context, followerID, followingID uint) (*model.FollowResponse, error) {
	start := time.Now()
	s.logger.Info("FollowService.Create", "follower_id", followerID, "following_id", followingID)

	if followerID == followingID {
		return nil, fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrCannotFollowSelf)
	}

	if err := s.validateUserExists(ctx, followerID); err != nil {
		return nil, err
	}
	if err := s.validateUserExists(ctx, followingID); err != nil {
		return nil, err
	}

	exists, err := s.followRepo.Exists(ctx, followerID, followingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", response.ResourceConflict, apperrs.ErrAlreadyFollowing)
	}

	follow := &model.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	if err := s.followRepo.Create(ctx, follow); err != nil {
		s.logger.Error("Failed to create follow", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := &model.FollowResponse{
		FollowerID:  follow.FollowerID,
		FollowingID: follow.FollowingID,
		CreatedAt:   follow.CreatedAt,
	}

	s.logger.Info("FollowService.Create success", "duration", time.Since(start))
	return resp, nil
}

func (s *followService) Delete(ctx context.Context, followerID, followingID uint) error {
	start := time.Now()
	s.logger.Info("FollowService.Delete", "follower_id", followerID, "following_id", followingID)

	exists, err := s.followRepo.Exists(ctx, followerID, followingID)
	if err != nil {
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if !exists {
		return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrNotFollowing)
	}

	if err := s.followRepo.Delete(ctx, followerID, followingID); err != nil {
		s.logger.Error("Failed to delete follow", "error", err, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FollowService.Delete success", "duration", time.Since(start))
	return nil
}

func (s *followService) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	start := time.Now()

	isFollowing, err := s.followRepo.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		s.logger.Error("Failed to check follow status", "error", err, "duration", time.Since(start))
		return false, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FollowService.IsFollowing success", "is_following", isFollowing, "duration", time.Since(start))
	return isFollowing, nil
}

func (s *followService) GetFollowingList(ctx context.Context, userID uint) ([]model.FollowResponse, error) {
	start := time.Now()
	s.logger.Info("FollowService.GetFollowingList", "user_id", userID)

	follows, err := s.followRepo.GetFollowingList(ctx, userID, "created_at desc")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.FollowResponse
	for _, follow := range follows {
		responses = append(responses, model.FollowResponse{
			FollowerID:  follow.FollowerID,
			FollowingID: follow.FollowingID,
			CreatedAt:   follow.CreatedAt,
		})
	}

	s.logger.Info("FollowService.GetFollowingList success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *followService) GetFollowerList(ctx context.Context, userID uint) ([]model.FollowResponse, error) {
	start := time.Now()
	s.logger.Info("FollowService.GetFollowerList", "user_id", userID)

	follows, err := s.followRepo.GetFollowerList(ctx, userID, "created_at desc")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.FollowResponse
	for _, follow := range follows {
		responses = append(responses, model.FollowResponse{
			FollowerID:  follow.FollowerID,
			FollowingID: follow.FollowingID,
			CreatedAt:   follow.CreatedAt,
		})
	}

	s.logger.Info("FollowService.GetFollowerList success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *followService) GetFollowingCount(ctx context.Context, userID uint) (int64, error) {
	start := time.Now()

	count, err := s.followRepo.CountFollowing(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count following", "error", err, "duration", time.Since(start))
		return 0, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FollowService.GetFollowingCount success", "count", count, "duration", time.Since(start))
	return count, nil
}

func (s *followService) GetFollowerCount(ctx context.Context, userID uint) (int64, error) {
	start := time.Now()

	count, err := s.followRepo.CountFollowers(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to count followers", "error", err, "duration", time.Since(start))
		return 0, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FollowService.GetFollowerCount success", "count", count, "duration", time.Since(start))
	return count, nil
}

func (s *followService) validateUserExists(ctx context.Context, userID uint) error {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrs.ErrUserNotFound) {
			return fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	return nil
}

var _ FollowService = (*followService)(nil)
