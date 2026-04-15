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
	ErrICHCategoryNotFound  = errors.New("非遗分类不存在")
	ErrCategoryNameExists   = errors.New("分类名称已存在")
	ErrInvalidCategoryLevel = errors.New("无效的分类级别")
	ErrCannotDeleteCategory = errors.New("无法删除有子节点的分类")
)

type ICHCategoryService interface {
	Create(ctx context.Context, category *model.ICHCategory) (*model.ICHCategoryResponse, error)
	Update(ctx context.Context, id uint, category *model.ICHCategory) (*model.ICHCategoryResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.ICHCategoryResponse, error)
	GetByName(ctx context.Context, name string) (*model.ICHCategoryResponse, error)
	GetByIDWithChildren(ctx context.Context, id uint) (*model.ICHCategoryResponse, error)
	List(ctx context.Context, orderBy string) ([]model.ICHCategoryResponse, error)
	ListRoot(ctx context.Context) ([]model.ICHCategoryResponse, error)
	ListByParentID(ctx context.Context, parentID uint) ([]model.ICHCategoryResponse, error)
	ListByRegionCode(ctx context.Context, regionCode string) ([]model.ICHCategoryResponse, error)
	ListActive(ctx context.Context) ([]model.ICHCategoryResponse, error)
}

type ichCategoryService struct {
	categoryRepo repository.ICHCategoryRepository
	logger       *slog.Logger
}

func NewICHCategoryService(
	categoryRepo repository.ICHCategoryRepository,
	logger *slog.Logger,
) ICHCategoryService {
	return &ichCategoryService{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *ichCategoryService) Create(ctx context.Context, category *model.ICHCategory) (*model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.Create", "name", category.Name)

	if err := s.validateLevel(category.Level); err != nil {
		return nil, err
	}

	existing, err := s.categoryRepo.GetByName(ctx, category.Name)
	if err != nil && !errors.Is(err, repository.ErrICHCategoryNotFound) {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%s: %w", response.ResourceConflict, ErrCategoryNameExists)
	}

	if category.ParentID != 0 {
		parent, err := s.categoryRepo.GetByID(ctx, category.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrICHCategoryNotFound) {
				return nil, fmt.Errorf("%s: 上级分类不存在", response.BadRequest)
			}
			return nil, fmt.Errorf("%s: %w", response.InternalError, err)
		}
		if parent.Level >= category.Level {
			return nil, fmt.Errorf("%s: 下级分类的级别必须大于上级分类", response.BadRequest)
		}
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		s.logger.Error("Failed to create ICH category", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp, err := s.GetByID(ctx, category.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("ICHCategoryService.Create success", "category_id", category.ID, "duration", time.Since(start))
	return resp, nil
}

func (s *ichCategoryService) Update(ctx context.Context, id uint, category *model.ICHCategory) (*model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.Update", "category_id", id)

	existing, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrICHCategoryNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrICHCategoryNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if category.Name != "" && category.Name != existing.Name {
		nameExists, err := s.categoryRepo.GetByName(ctx, category.Name)
		if err != nil && !errors.Is(err, repository.ErrICHCategoryNotFound) {
			return nil, fmt.Errorf("%s: %w", response.InternalError, err)
		}
		if nameExists != nil && nameExists.ID != id {
			return nil, fmt.Errorf("%s: %w", response.ResourceConflict, ErrCategoryNameExists)
		}
	}

	if category.Level != 0 && category.Level != existing.Level {
		if err := s.validateLevel(category.Level); err != nil {
			return nil, err
		}
	}

	category.ID = id
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		s.logger.Error("Failed to update ICH category", "error", err, "category_id", id, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("ICHCategoryService.Update success", "category_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *ichCategoryService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("ICHCategoryService.Delete", "category_id", id)

	children, err := s.categoryRepo.ListByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if len(children) > 0 {
		return fmt.Errorf("%s: %w", response.BadRequest, ErrCannotDeleteCategory)
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete ICH category", "error", err, "category_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("ICHCategoryService.Delete success", "category_id", id, "duration", time.Since(start))
	return nil
}

func (s *ichCategoryService) GetByID(ctx context.Context, id uint) (*model.ICHCategoryResponse, error) {
	start := time.Now()

	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrICHCategoryNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrICHCategoryNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toICHCategoryResponse(category)

	s.logger.Info("ICHCategoryService.GetByID success", "category_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *ichCategoryService) GetByName(ctx context.Context, name string) (*model.ICHCategoryResponse, error) {
	start := time.Now()

	category, err := s.categoryRepo.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrICHCategoryNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrICHCategoryNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toICHCategoryResponse(category)

	s.logger.Info("ICHCategoryService.GetByName success", "name", name, "duration", time.Since(start))
	return resp, nil
}

func (s *ichCategoryService) GetByIDWithChildren(ctx context.Context, id uint) (*model.ICHCategoryResponse, error) {
	start := time.Now()

	category, err := s.categoryRepo.GetByIDWithChildren(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrICHCategoryNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, ErrICHCategoryNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toICHCategoryResponse(category)

	s.logger.Info("ICHCategoryService.GetByIDWithChildren success", "category_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *ichCategoryService) List(ctx context.Context, orderBy string) ([]model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.List", "orderBy", orderBy)

	categories, err := s.categoryRepo.List(ctx, orderBy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.ICHCategoryResponse
	for _, category := range categories {
		responses = append(responses, *s.toICHCategoryResponse(&category))
	}

	s.logger.Info("ICHCategoryService.List success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *ichCategoryService) ListRoot(ctx context.Context) ([]model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.ListRoot")

	categories, err := s.categoryRepo.ListRoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.ICHCategoryResponse
	for _, category := range categories {
		responses = append(responses, *s.toICHCategoryResponse(&category))
	}

	s.logger.Info("ICHCategoryService.ListRoot success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *ichCategoryService) ListByParentID(ctx context.Context, parentID uint) ([]model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.ListByParentID", "parent_id", parentID)

	categories, err := s.categoryRepo.ListByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.ICHCategoryResponse
	for _, category := range categories {
		responses = append(responses, *s.toICHCategoryResponse(&category))
	}

	s.logger.Info("ICHCategoryService.ListByParentID success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *ichCategoryService) ListByRegionCode(ctx context.Context, regionCode string) ([]model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.ListByRegionCode", "region_code", regionCode)

	categories, err := s.categoryRepo.ListByRegionCode(ctx, regionCode)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.ICHCategoryResponse
	for _, category := range categories {
		responses = append(responses, *s.toICHCategoryResponse(&category))
	}

	s.logger.Info("ICHCategoryService.ListByRegionCode success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *ichCategoryService) ListActive(ctx context.Context) ([]model.ICHCategoryResponse, error) {
	start := time.Now()
	s.logger.Info("ICHCategoryService.ListActive")

	categories, err := s.categoryRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.ICHCategoryResponse
	for _, category := range categories {
		responses = append(responses, *s.toICHCategoryResponse(&category))
	}

	s.logger.Info("ICHCategoryService.ListActive success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *ichCategoryService) validateLevel(level int8) error {
	if level < 1 || level > 5 {
		return fmt.Errorf("%s: %w", response.BadRequest, ErrInvalidCategoryLevel)
	}
	return nil
}

func (s *ichCategoryService) toICHCategoryResponse(category *model.ICHCategory) *model.ICHCategoryResponse {
	if category == nil {
		return nil
	}

	return &model.ICHCategoryResponse{
		ID:          category.ID,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Level:       category.Level,
		RegionCode:  category.RegionCode,
		Description: category.Description,
		IconURL:     category.IconURL,
		SortOrder:   category.SortOrder,
		Status:      category.Status,
	}
}

var _ ICHCategoryService = (*ichCategoryService)(nil)
