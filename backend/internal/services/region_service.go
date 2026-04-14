package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	model "backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
)

var (
	ErrRegionCodeExists   = errors.New("地区代码已存在")
	ErrInvalidRegionLevel = errors.New("无效的地区级别")
	ErrCannotDeleteRegion = errors.New("无法删除有子节点的地区")
)

type RegionService interface {
	Create(ctx context.Context, region *model.Region) (*model.RegionResponse, error)
	Update(ctx context.Context, id uint, region *model.Region) (*model.RegionResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.RegionResponse, error)
	GetByCode(ctx context.Context, code string) (*model.RegionResponse, error)
	GetByIDWithChildren(ctx context.Context, id uint) (*model.RegionResponse, error)
	List(ctx context.Context, orderBy string) ([]model.RegionResponse, error)
	ListRoot(ctx context.Context) ([]model.RegionResponse, error)
	ListByParentID(ctx context.Context, parentID uint) ([]model.RegionResponse, error)
	ListByLevel(ctx context.Context, level int8) ([]model.RegionResponse, error)
	ListHeritageCenters(ctx context.Context) ([]model.RegionResponse, error)
}

type regionService struct {
	regionRepo repository.RegionRepository
	logger     *slog.Logger
}

func NewRegionService(
	regionRepo repository.RegionRepository,
	logger *slog.Logger,
) RegionService {
	return &regionService{
		regionRepo: regionRepo,
		logger:     logger,
	}
}

func (s *regionService) Create(ctx context.Context, region *model.Region) (*model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.Create", "name", region.Name, "code", region.Code)

	if err := s.validateLevel(region.Level); err != nil {
		return nil, err
	}

	existing, err := s.regionRepo.GetByCode(ctx, region.Code)
	if err != nil && !errors.Is(err, repository.ErrRegionNotFound) {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: %w", response.ResourceConflict, ErrRegionCodeExists)
	}

	if region.ParentID != 0 {
		parent, err := s.regionRepo.GetByID(ctx, region.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrRegionNotFound) {
				return nil, fmt.Errorf("%w: 上级地区不存在", response.BadRequest)
			}
			return nil, fmt.Errorf("%w: %w", response.InternalError, err)
		}
		if parent.Level >= region.Level {
			return nil, fmt.Errorf("%w: 下级地区的级别必须大于上级地区", response.BadRequest)
		}
	}

	if err := s.regionRepo.Create(ctx, region); err != nil {
		s.logger.Error("Failed to create region", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	resp, err := s.GetByID(ctx, region.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("RegionService.Create success", "region_id", region.ID, "duration", time.Since(start))
	return resp, nil
}

func (s *regionService) Update(ctx context.Context, id uint, region *model.Region) (*model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.Update", "region_id", id)

	existing, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRegionNotFound) {
			return nil, fmt.Errorf("%w: %w", response.UserNotFound, ErrRegionNotFound)
		}
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	if region.Code != "" && region.Code != existing.Code {
		codeExists, err := s.regionRepo.GetByCode(ctx, region.Code)
		if err != nil && !errors.Is(err, repository.ErrRegionNotFound) {
			return nil, fmt.Errorf("%w: %w", response.InternalError, err)
		}
		if codeExists != nil && codeExists.ID != id {
			return nil, fmt.Errorf("%w: %w", response.ResourceConflict, ErrRegionCodeExists)
		}
	}

	if region.Level != 0 && region.Level != existing.Level {
		if err := s.validateLevel(region.Level); err != nil {
			return nil, err
		}
	}

	region.ID = id
	if err := s.regionRepo.Update(ctx, region); err != nil {
		s.logger.Error("Failed to update region", "error", err, "region_id", id, "duration", time.Since(start))
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	resp, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("RegionService.Update success", "region_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *regionService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("RegionService.Delete", "region_id", id)

	children, err := s.regionRepo.ListByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %w", response.InternalError, err)
	}
	if len(children) > 0 {
		return fmt.Errorf("%w: %w", response.BadRequest, ErrCannotDeleteRegion)
	}

	if err := s.regionRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete region", "error", err, "region_id", id, "duration", time.Since(start))
		return fmt.Errorf("%w: %w", response.InternalError, err)
	}

	s.logger.Info("RegionService.Delete success", "region_id", id, "duration", time.Since(start))
	return nil
}

func (s *regionService) GetByID(ctx context.Context, id uint) (*model.RegionResponse, error) {
	start := time.Now()

	region, err := s.regionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRegionNotFound) {
			return nil, fmt.Errorf("%w: %w", response.UserNotFound, ErrRegionNotFound)
		}
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	resp := s.toRegionResponse(region)

	s.logger.Info("RegionService.GetByID success", "region_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *regionService) GetByCode(ctx context.Context, code string) (*model.RegionResponse, error) {
	start := time.Now()

	region, err := s.regionRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrRegionNotFound) {
			return nil, fmt.Errorf("%w: %w", response.UserNotFound, ErrRegionNotFound)
		}
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	resp := s.toRegionResponse(region)

	s.logger.Info("RegionService.GetByCode success", "code", code, "duration", time.Since(start))
	return resp, nil
}

func (s *regionService) GetByIDWithChildren(ctx context.Context, id uint) (*model.RegionResponse, error) {
	start := time.Now()

	region, err := s.regionRepo.GetByIDWithChildren(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRegionNotFound) {
			return nil, fmt.Errorf("%w: %w", response.UserNotFound, ErrRegionNotFound)
		}
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	resp := s.toRegionResponse(region)

	s.logger.Info("RegionService.GetByIDWithChildren success", "region_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *regionService) List(ctx context.Context, orderBy string) ([]model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.List", "orderBy", orderBy)

	regions, err := s.regionRepo.List(ctx, orderBy)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	var responses []model.RegionResponse
	for _, region := range regions {
		responses = append(responses, *s.toRegionResponse(&region))
	}

	s.logger.Info("RegionService.List success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *regionService) ListRoot(ctx context.Context) ([]model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.ListRoot")

	regions, err := s.regionRepo.ListRoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	var responses []model.RegionResponse
	for _, region := range regions {
		responses = append(responses, *s.toRegionResponse(&region))
	}

	s.logger.Info("RegionService.ListRoot success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *regionService) ListByParentID(ctx context.Context, parentID uint) ([]model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.ListByParentID", "parent_id", parentID)

	regions, err := s.regionRepo.ListByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	var responses []model.RegionResponse
	for _, region := range regions {
		responses = append(responses, *s.toRegionResponse(&region))
	}

	s.logger.Info("RegionService.ListByParentID success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *regionService) ListByLevel(ctx context.Context, level int8) ([]model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.ListByLevel", "level", level)

	if err := s.validateLevel(level); err != nil {
		return nil, err
	}

	regions, err := s.regionRepo.ListByLevel(ctx, level)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	var responses []model.RegionResponse
	for _, region := range regions {
		responses = append(responses, *s.toRegionResponse(&region))
	}

	s.logger.Info("RegionService.ListByLevel success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *regionService) ListHeritageCenters(ctx context.Context) ([]model.RegionResponse, error) {
	start := time.Now()
	s.logger.Info("RegionService.ListHeritageCenters")

	regions, err := s.regionRepo.ListHeritageCenters(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", response.InternalError, err)
	}

	var responses []model.RegionResponse
	for _, region := range regions {
		responses = append(responses, *s.toRegionResponse(&region))
	}

	s.logger.Info("RegionService.ListHeritageCenters success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *regionService) validateLevel(level int8) error {
	if level < 1 || level > 5 {
		return fmt.Errorf("%w: %w", response.BadRequest, ErrInvalidRegionLevel)
	}
	return nil
}

func (s *regionService) toRegionResponse(region *model.Region) *model.RegionResponse {
	if region == nil {
		return nil
	}

	return &model.RegionResponse{
		ID:               region.ID,
		CreatedAt:        region.CreatedAt,
		UpdatedAt:        region.UpdatedAt,
		ParentID:         region.ParentID,
		Name:             region.Name,
		Code:             region.Code,
		Level:            region.Level,
		IsHeritageCenter: region.IsHeritageCenter,
		CultureDesc:      region.CultureDesc,
	}
}

var _ RegionService = (*regionService)(nil)
