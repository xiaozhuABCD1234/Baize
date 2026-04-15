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
	ErrFavoriteNotFound = apperrs.ErrFavoriteNotFound
	ErrAlreadyFavorited = apperrs.ErrAlreadyFavorited
	ErrNotFavorited     = apperrs.ErrNotFavorited
)

type FavoriteService interface {
	Create(ctx context.Context, req *model.FavoriteRequest, userID uint) (*model.FavoriteResponse, error)
	Delete(ctx context.Context, id uint) error
	DeleteByUserAndWork(ctx context.Context, userID, workID uint) error
	GetByID(ctx context.Context, id uint) (*model.FavoriteResponse, error)
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.FavoriteResponse, int64, error)
	ListByWorkID(ctx context.Context, workID uint) ([]model.FavoriteResponse, error)
	UpdateFolder(ctx context.Context, id uint, folderID uint) error
	Exists(ctx context.Context, userID, workID uint) (bool, error)
	CountByWorkID(ctx context.Context, workID uint) (int64, error)
}

type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
	workRepo     repository.WorkRepository
	userRepo     repository.UserRepository
	logger       *slog.Logger
}

func NewFavoriteService(
	favoriteRepo repository.FavoriteRepository,
	workRepo repository.WorkRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) FavoriteService {
	return &favoriteService{
		favoriteRepo: favoriteRepo,
		workRepo:     workRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

func (s *favoriteService) Create(ctx context.Context, req *model.FavoriteRequest, userID uint) (*model.FavoriteResponse, error) {
	start := time.Now()
	s.logger.Info("FavoriteService.Create", "user_id", userID, "work_id", req.WorkID)

	exists, err := s.favoriteRepo.Exists(ctx, userID, req.WorkID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if exists {
		return nil, fmt.Errorf("%s: %w", response.ResourceConflict, apperrs.ErrAlreadyFavorited)
	}

	_, err = s.workRepo.GetByID(ctx, req.WorkID)
	if err != nil {
		if errors.Is(err, apperrs.ErrWorkNotFound) {
			return nil, fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrWorkNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	favorite := &model.Favorite{
		UserID:   userID,
		WorkID:   req.WorkID,
		FolderID: req.FolderID,
		Remark:   req.Remark,
	}

	if err := s.favoriteRepo.Create(ctx, favorite); err != nil {
		s.logger.Error("Failed to create favorite", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, req.WorkID, "favorite_count", 1); err != nil {
		s.logger.Warn("Failed to increment work favorite count", "error", err, "work_id", req.WorkID)
	}

	resp, err := s.GetByID(ctx, favorite.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("FavoriteService.Create success", "favorite_id", favorite.ID, "duration", time.Since(start))
	return resp, nil
}

func (s *favoriteService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("FavoriteService.Delete", "favorite_id", id)

	favorite, err := s.favoriteRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrFavoriteNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrFavoriteNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.favoriteRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete favorite", "error", err, "favorite_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, favorite.WorkID, "favorite_count", -1); err != nil {
		s.logger.Warn("Failed to decrement work favorite count", "error", err, "work_id", favorite.WorkID)
	}

	s.logger.Info("FavoriteService.Delete success", "favorite_id", id, "duration", time.Since(start))
	return nil
}

func (s *favoriteService) DeleteByUserAndWork(ctx context.Context, userID, workID uint) error {
	start := time.Now()
	s.logger.Info("FavoriteService.DeleteByUserAndWork", "user_id", userID, "work_id", workID)

	favorite, err := s.favoriteRepo.GetByUserAndWork(ctx, userID, workID)
	if err != nil {
		if errors.Is(err, apperrs.ErrFavoriteNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrFavoriteNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.favoriteRepo.DeleteByUserAndWork(ctx, userID, workID); err != nil {
		s.logger.Error("Failed to delete favorite", "error", err, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, favorite.WorkID, "favorite_count", -1); err != nil {
		s.logger.Warn("Failed to decrement work favorite count", "error", err, "work_id", favorite.WorkID)
	}

	s.logger.Info("FavoriteService.DeleteByUserAndWork success", "duration", time.Since(start))
	return nil
}

func (s *favoriteService) GetByID(ctx context.Context, id uint) (*model.FavoriteResponse, error) {
	start := time.Now()

	favorite, err := s.favoriteRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrFavoriteNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrFavoriteNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toFavoriteResponse(favorite)

	work, err := s.workRepo.GetByIDWithAll(ctx, favorite.WorkID)
	if err == nil && work != nil {
		workResp := s.toWorkResponse(work)
		resp.Work = workResp
	}

	s.logger.Info("FavoriteService.GetByID success", "favorite_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *favoriteService) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.FavoriteResponse, int64, error) {
	start := time.Now()
	s.logger.Info("FavoriteService.ListByUserID", "user_id", userID, "page", page, "pageSize", pageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	favorites, total, err := s.favoriteRepo.ListWithPagination(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.FavoriteResponse
	for _, fav := range favorites {
		resp := s.toFavoriteResponse(&fav)
		if fav.Work.ID != 0 {
			resp.Work = s.toWorkResponse(&fav.Work)
		}
		responses = append(responses, *resp)
	}

	s.logger.Info("FavoriteService.ListByUserID success", "total", total, "count", len(responses), "duration", time.Since(start))
	return responses, total, nil
}

func (s *favoriteService) ListByWorkID(ctx context.Context, workID uint) ([]model.FavoriteResponse, error) {
	start := time.Now()
	s.logger.Info("FavoriteService.ListByWorkID", "work_id", workID)

	favorites, err := s.favoriteRepo.ListByWorkID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.FavoriteResponse
	for _, fav := range favorites {
		responses = append(responses, *s.toFavoriteResponse(&fav))
	}

	s.logger.Info("FavoriteService.ListByWorkID success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *favoriteService) UpdateFolder(ctx context.Context, id uint, folderID uint) error {
	start := time.Now()
	s.logger.Info("FavoriteService.UpdateFolder", "favorite_id", id, "folder_id", folderID)

	_, err := s.favoriteRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrFavoriteNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrFavoriteNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.favoriteRepo.UpdateFolder(ctx, id, folderID); err != nil {
		s.logger.Error("Failed to update favorite folder", "error", err, "favorite_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FavoriteService.UpdateFolder success", "favorite_id", id, "duration", time.Since(start))
	return nil
}

func (s *favoriteService) Exists(ctx context.Context, userID, workID uint) (bool, error) {
	start := time.Now()

	exists, err := s.favoriteRepo.Exists(ctx, userID, workID)
	if err != nil {
		s.logger.Error("Failed to check favorite exists", "error", err, "duration", time.Since(start))
		return false, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FavoriteService.Exists success", "exists", exists, "duration", time.Since(start))
	return exists, nil
}

func (s *favoriteService) CountByWorkID(ctx context.Context, workID uint) (int64, error) {
	start := time.Now()

	count, err := s.favoriteRepo.CountByWorkID(ctx, workID)
	if err != nil {
		s.logger.Error("Failed to count favorites", "error", err, "duration", time.Since(start))
		return 0, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("FavoriteService.CountByWorkID success", "count", count, "duration", time.Since(start))
	return count, nil
}

func (s *favoriteService) toFavoriteResponse(favorite *model.Favorite) *model.FavoriteResponse {
	if favorite == nil {
		return nil
	}

	return &model.FavoriteResponse{
		ID:        favorite.ID,
		UserID:    favorite.UserID,
		WorkID:    favorite.WorkID,
		FolderID:  favorite.FolderID,
		Remark:    favorite.Remark,
		CreatedAt: favorite.CreatedAt,
	}
}

func (s *favoriteService) toWorkResponse(work *model.Work) *model.WorkResponse {
	if work == nil {
		return nil
	}

	return &model.WorkResponse{
		ID:            work.ID,
		CreatedAt:     work.CreatedAt,
		UpdatedAt:     work.UpdatedAt,
		UserID:        work.UserID,
		Title:         work.Title,
		Content:       work.Content,
		ContentType:   int8(work.ContentType),
		CraftID:       work.CraftID,
		CategoryID:    work.CategoryID,
		RegionID:      work.RegionID,
		ViewCount:     work.ViewCount,
		LikeCount:     work.LikeCount,
		CommentCount:  work.CommentCount,
		FavoriteCount: work.FavoriteCount,
		ShareCount:    work.ShareCount,
		Status:        int8(work.Status),
		IsTop:         work.IsTop,
		IsRecommended: work.IsRecommended,
		Weight:        work.Weight,
		PublishedAt:   work.PublishedAt,
	}
}

var _ FavoriteService = (*favoriteService)(nil)
