package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/datatypes"

	model "backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
)

var (
	ErrWorkNotFound      = errors.New("作品不存在")
	ErrWorkMediaNotFound = errors.New("作品媒体不存在")
	ErrInvalidWorkStatus = errors.New("无效的作品状态")
	ErrCannotDeleteWork  = errors.New("无法删除已发布的作品")
	ErrCraftNotFound     = errors.New("技艺不存在")
	ErrCategoryNotFound  = errors.New("分类不存在")
	ErrRegionNotFound    = errors.New("地区不存在")
)

type WorkService interface {
	Create(ctx context.Context, req *model.CreateWorkRequest, userID uint) (*model.WorkResponse, error)
	Update(ctx context.Context, id uint, req *model.CreateWorkRequest) (*model.WorkResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.WorkResponse, error)
	GetByIDDetailed(ctx context.Context, id uint) (*model.WorkResponse, error)
	List(ctx context.Context, req *model.WorkListRequest) ([]model.WorkResponse, int64, error)
	ListTop(ctx context.Context, limit int) ([]model.WorkResponse, error)
	ListRecommended(ctx context.Context, limit int) ([]model.WorkResponse, error)
	UpdateStatus(ctx context.Context, id uint, status model.WorkStatus) error
	IncrementCount(ctx context.Context, id uint, field string, delta int) error
}

type workService struct {
	workRepo  repository.WorkRepository
	mediaRepo repository.WorkMediaRepository
	craftRepo repository.CraftRepository
	userRepo  repository.UserRepository
	logger    *slog.Logger
}

func NewWorkService(
	workRepo repository.WorkRepository,
	mediaRepo repository.WorkMediaRepository,
	craftRepo repository.CraftRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) WorkService {
	return &workService{
		workRepo:  workRepo,
		mediaRepo: mediaRepo,
		craftRepo: craftRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

func (s *workService) Create(ctx context.Context, req *model.CreateWorkRequest, userID uint) (*model.WorkResponse, error) {
	start := time.Now()
	s.logger.Info("WorkService.Create", "user_id", userID, "title", req.Title)

	if err := s.validateCraftExists(ctx, req.CraftID); err != nil {
		return nil, err
	}
	if err := s.validateCategoryExists(ctx, req.CategoryID); err != nil {
		return nil, err
	}
	if err := s.validateRegionExists(ctx, req.RegionID); err != nil {
		return nil, err
	}

	techniqueTags, err := s.convertToJSON(req.TechniqueTags)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.BadRequest, err)
	}

	materials, err := s.convertToJSON(req.Materials)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.BadRequest, err)
	}

	work := &model.Work{
		UserID:        userID,
		Title:         req.Title,
		Content:       req.Content,
		ContentType:   model.ContentType(req.ContentType),
		CraftID:       req.CraftID,
		CategoryID:    req.CategoryID,
		RegionID:      req.RegionID,
		TechniqueTags: techniqueTags,
		Materials:     materials,
		CreationTime:  req.CreationTime,
		Status:        model.WorkStatusPublished,
	}
	now := time.Now()
	work.PublishedAt = &now

	if err := s.workRepo.Create(ctx, work); err != nil {
		s.logger.Error("Failed to create work", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.createMedia(ctx, work.ID, req.Media); err != nil {
		s.logger.Error("Failed to create work media", "error", err, "work_id", work.ID, "duration", time.Since(start))
		return nil, err
	}

	workResp, err := s.GetByIDDetailed(ctx, work.ID)
	if err != nil {
		s.logger.Error("Failed to get created work", "error", err, "work_id", work.ID, "duration", time.Since(start))
		return nil, err
	}

	s.logger.Info("WorkService.Create success", "work_id", work.ID, "duration", time.Since(start))
	return workResp, nil
}

func (s *workService) Update(ctx context.Context, id uint, req *model.CreateWorkRequest) (*model.WorkResponse, error) {
	start := time.Now()
	s.logger.Info("WorkService.Update", "work_id", id)

	work, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if req.CraftID != 0 {
		if err := s.validateCraftExists(ctx, req.CraftID); err != nil {
			return nil, err
		}
		work.CraftID = req.CraftID
	}
	if req.CategoryID != 0 {
		if err := s.validateCategoryExists(ctx, req.CategoryID); err != nil {
			return nil, err
		}
		work.CategoryID = req.CategoryID
	}
	if req.RegionID != 0 {
		if err := s.validateRegionExists(ctx, req.RegionID); err != nil {
			return nil, err
		}
		work.RegionID = req.RegionID
	}

	work.Title = req.Title
	work.Content = req.Content
	work.ContentType = model.ContentType(req.ContentType)

	if req.TechniqueTags != nil {
		techniqueTags, err := s.convertToJSON(req.TechniqueTags)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", response.BadRequest, err)
		}
		work.TechniqueTags = techniqueTags
	}

	if req.Materials != nil {
		materials, err := s.convertToJSON(req.Materials)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", response.BadRequest, err)
		}
		work.Materials = materials
	}

	if req.CreationTime != nil {
		work.CreationTime = req.CreationTime
	}

	if err := s.workRepo.Update(ctx, work); err != nil {
		s.logger.Error("Failed to update work", "error", err, "work_id", id, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.updateMedia(ctx, id, req.Media); err != nil {
		s.logger.Error("Failed to update work media", "error", err, "work_id", id, "duration", time.Since(start))
		return nil, err
	}

	workResp, err := s.GetByIDDetailed(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("WorkService.Update success", "work_id", id, "duration", time.Since(start))
	return workResp, nil
}

func (s *workService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("WorkService.Delete", "work_id", id)

	work, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if work.Status == model.WorkStatusPublished {
		return fmt.Errorf("%s: %w", response.BadRequest, ErrCannotDeleteWork)
	}

	if err := s.mediaRepo.DeleteByWorkID(ctx, id); err != nil {
		s.logger.Error("Failed to delete work media", "error", err, "work_id", id)
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete work", "error", err, "work_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("WorkService.Delete success", "work_id", id, "duration", time.Since(start))
	return nil
}

func (s *workService) GetByID(ctx context.Context, id uint) (*model.WorkResponse, error) {
	start := time.Now()

	work, err := s.workRepo.GetByIDWithSelect(ctx, id, "User", "Craft")
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toWorkResponse(work)

	s.logger.Info("WorkService.GetByID success", "work_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *workService) GetByIDDetailed(ctx context.Context, id uint) (*model.WorkResponse, error) {
	start := time.Now()

	work, err := s.workRepo.GetByIDWithAll(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toWorkResponse(work)

	s.logger.Info("WorkService.GetByIDDetailed success", "work_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *workService) List(ctx context.Context, req *model.WorkListRequest) ([]model.WorkResponse, int64, error) {
	start := time.Now()
	s.logger.Info("WorkService.List", "req", req)

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	var works []model.Work
	var total int64
	var err error

	orderBy := s.buildOrderBy(req.OrderBy)

	switch {
	case req.UserID != 0:
		works, err = s.workRepo.ListByUserID(ctx, req.UserID)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", response.InternalError, err)
		}
		total = int64(len(works))
	case req.CraftID != 0:
		works, err = s.workRepo.ListByCraftID(ctx, req.CraftID)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", response.InternalError, err)
		}
		total = int64(len(works))
	default:
		if req.IsMaster {
			works, total, err = s.workRepo.ListWithPagination(ctx, req.Page, req.PageSize, orderBy)
			works = s.filterMasterWorks(works)
		} else {
			works, total, err = s.workRepo.ListWithPagination(ctx, req.Page, req.PageSize, orderBy)
		}
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", response.InternalError, err)
		}
	}

	var responses []model.WorkResponse
	for _, work := range works {
		responses = append(responses, *s.toWorkResponse(&work))
	}

	s.logger.Info("WorkService.List success", "total", total, "duration", time.Since(start))
	return responses, total, nil
}

func (s *workService) ListTop(ctx context.Context, limit int) ([]model.WorkResponse, error) {
	start := time.Now()
	s.logger.Info("WorkService.ListTop", "limit", limit)

	if limit < 1 || limit > 50 {
		limit = 10
	}

	works, err := s.workRepo.ListTop(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.WorkResponse
	for _, work := range works {
		responses = append(responses, *s.toWorkResponse(&work))
	}

	s.logger.Info("WorkService.ListTop success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *workService) ListRecommended(ctx context.Context, limit int) ([]model.WorkResponse, error) {
	start := time.Now()
	s.logger.Info("WorkService.ListRecommended", "limit", limit)

	if limit < 1 || limit > 50 {
		limit = 10
	}

	works, err := s.workRepo.ListRecommended(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.WorkResponse
	for _, work := range works {
		responses = append(responses, *s.toWorkResponse(&work))
	}

	s.logger.Info("WorkService.ListRecommended success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *workService) UpdateStatus(ctx context.Context, id uint, status model.WorkStatus) error {
	start := time.Now()
	s.logger.Info("WorkService.UpdateStatus", "work_id", id, "status", status)

	if !s.isValidWorkStatus(status) {
		return fmt.Errorf("%s: %w", response.BadRequest, ErrInvalidWorkStatus)
	}

	_, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("Failed to update work status", "error", err, "work_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("WorkService.UpdateStatus success", "work_id", id, "duration", time.Since(start))
	return nil
}

func (s *workService) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	start := time.Now()
	s.logger.Info("WorkService.IncrementCount", "work_id", id, "field", field, "delta", delta)

	_, err := s.workRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrWorkNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, ErrWorkNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, id, field, delta); err != nil {
		s.logger.Error("Failed to increment count", "error", err, "work_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("WorkService.IncrementCount success", "work_id", id, "duration", time.Since(start))
	return nil
}

func (s *workService) createMedia(ctx context.Context, workID uint, mediaItems []model.MediaItem) error {
	if len(mediaItems) == 0 {
		return nil
	}

	var mediaList []model.WorkMedia
	for _, item := range mediaItems {
		mediaList = append(mediaList, model.WorkMedia{
			WorkID:       workID,
			MediaType:    model.MediaType(item.MediaType),
			URL:          item.URL,
			ThumbnailURL: item.ThumbnailURL,
			Description:  item.Description,
			SortOrder:    item.SortOrder,
		})
	}

	if err := s.mediaRepo.CreateBatch(ctx, mediaList); err != nil {
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	return nil
}

func (s *workService) updateMedia(ctx context.Context, workID uint, mediaItems []model.MediaItem) error {
	if err := s.mediaRepo.DeleteByWorkID(ctx, workID); err != nil {
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if len(mediaItems) == 0 {
		return nil
	}

	return s.createMedia(ctx, workID, mediaItems)
}

func (s *workService) validateCraftExists(ctx context.Context, craftID uint) error {
	if craftID == 0 {
		return nil
	}
	_, err := s.craftRepo.GetByID(ctx, craftID)
	if err != nil {
		if errors.Is(err, repository.ErrCraftNotFound) {
			return fmt.Errorf("%s: %w", response.BadRequest, ErrCraftNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	return nil
}

func (s *workService) validateCategoryExists(ctx context.Context, categoryID uint) error {
	if categoryID == 0 {
		return nil
	}
	return nil
}

func (s *workService) validateRegionExists(ctx context.Context, regionID uint) error {
	if regionID == 0 {
		return nil
	}
	return nil
}

func (s *workService) isValidWorkStatus(status model.WorkStatus) bool {
	switch status {
	case model.WorkStatusDraft, model.WorkStatusPublished, model.WorkStatusReviewing,
		model.WorkStatusRejected, model.WorkStatusOffline:
		return true
	default:
		return false
	}
}

func (s *workService) buildOrderBy(orderBy string) string {
	switch orderBy {
	case "newest":
		return "created_at desc"
	case "hot":
		return "view_count desc, like_count desc"
	case "weight":
		return "weight desc, created_at desc"
	default:
		return "created_at desc"
	}
}

func (s *workService) filterMasterWorks(works []model.Work) []model.Work {
	var filtered []model.Work
	for _, work := range works {
		if work.User.UserType == model.UserTypeMaster {
			filtered = append(filtered, work)
		}
	}
	return filtered
}

func (s *workService) convertToJSON(data []string) (datatypes.JSON, error) {
	if data == nil {
		return nil, nil
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *workService) parseJSONArray(data datatypes.JSON) map[string]interface{} {
	if data == nil {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

func (s *workService) toWorkResponse(work *model.Work) *model.WorkResponse {
	if work == nil {
		return nil
	}

	resp := &model.WorkResponse{
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
		CreationTime:  work.CreationTime,
	}

	if work.TechniqueTags != nil {
		resp.TechniqueTags = s.parseJSONArray(work.TechniqueTags)
	}
	if work.Materials != nil {
		resp.Materials = s.parseJSONArray(work.Materials)
	}

	if work.User.ID != 0 {
		resp.User = &model.UserResponse{
			ID:        work.User.ID,
			CreatedAt: work.User.CreatedAt,
			UpdatedAt: work.User.UpdatedAt,
			Username:  work.User.Username,
			Email:     work.User.Email,
			Phone:     work.User.Phone,
			UserType:  string(work.User.UserType),
			Status:    int8(work.User.Status),
		}
	}

	if work.Craft.ID != 0 {
		resp.Craft = &model.CraftResponse{
			ID:          work.Craft.ID,
			CreatedAt:   work.Craft.CreatedAt,
			UpdatedAt:   work.Craft.UpdatedAt,
			CategoryID:  work.Craft.CategoryID,
			Name:        work.Craft.Name,
			Description: work.Craft.Description,
			History:     work.Craft.History,
			Difficulty:  work.Craft.Difficulty,
		}
		if work.Craft.Tools != nil {
			resp.Craft.Tools = s.parseJSONArray(work.Craft.Tools)
		}
		if work.Craft.RegionFeatures != nil {
			resp.Craft.RegionFeatures = s.parseJSONArray(work.Craft.RegionFeatures)
		}
	}

	if len(work.Media) > 0 {
		for _, media := range work.Media {
			resp.Media = append(resp.Media, model.WorkMediaResponse{
				ID:           media.ID,
				MediaType:    int8(media.MediaType),
				URL:          media.URL,
				ThumbnailURL: media.ThumbnailURL,
				Width:        media.Width,
				Height:       media.Height,
				Duration:     media.Duration,
				FileSize:     media.FileSize,
				SortOrder:    media.SortOrder,
				Description:  media.Description,
			})
		}
	}

	return resp
}

var _ WorkService = (*workService)(nil)
