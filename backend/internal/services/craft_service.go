package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/datatypes"

	apperrs "backend/internal/errors"
	model "backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
)

var (
	ErrCraftNameExists   = apperrs.ErrCraftNameExists
	ErrInvalidDifficulty = apperrs.ErrInvalidDifficulty
)

type CraftService interface {
	Create(ctx context.Context, craft *model.Craft) (*model.CraftResponse, error)
	Update(ctx context.Context, id uint, craft *model.Craft) (*model.CraftResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.CraftResponse, error)
	GetByIDWithCategory(ctx context.Context, id uint) (*model.CraftResponse, error)
	List(ctx context.Context, orderBy string) ([]model.CraftResponse, error)
	ListByCategory(ctx context.Context, categoryID uint) ([]model.CraftResponse, error)
	ListByDifficulty(ctx context.Context, difficulty int8) ([]model.CraftResponse, error)
}

type craftService struct {
	craftRepo    repository.CraftRepository
	categoryRepo repository.ICHCategoryRepository
	logger       *slog.Logger
}

func NewCraftService(
	craftRepo repository.CraftRepository,
	categoryRepo repository.ICHCategoryRepository,
	logger *slog.Logger,
) CraftService {
	return &craftService{
		craftRepo:    craftRepo,
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

func (s *craftService) Create(ctx context.Context, craft *model.Craft) (*model.CraftResponse, error) {
	start := time.Now()
	s.logger.Info("CraftService.Create", "name", craft.Name)

	if err := s.validateCategoryExists(ctx, craft.CategoryID); err != nil {
		return nil, err
	}

	if err := s.validateDifficulty(craft.Difficulty); err != nil {
		return nil, err
	}

	existing, err := s.craftRepo.GetByName(ctx, craft.Name)
	if err != nil && !errors.Is(err, apperrs.ErrCraftNotFound) {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%s: %w", response.ResourceConflict, apperrs.ErrCraftNameExists)
	}

	if err := s.craftRepo.Create(ctx, craft); err != nil {
		s.logger.Error("Failed to create craft", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp, err := s.GetByIDWithCategory(ctx, craft.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("CraftService.Create success", "craft_id", craft.ID, "duration", time.Since(start))
	return resp, nil
}

func (s *craftService) Update(ctx context.Context, id uint, craft *model.Craft) (*model.CraftResponse, error) {
	start := time.Now()
	s.logger.Info("CraftService.Update", "craft_id", id)

	existing, err := s.craftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCraftNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCraftNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if craft.CategoryID != 0 && craft.CategoryID != existing.CategoryID {
		if err := s.validateCategoryExists(ctx, craft.CategoryID); err != nil {
			return nil, err
		}
	}

	if err := s.validateDifficulty(craft.Difficulty); err != nil {
		return nil, err
	}

	if craft.Name != "" && craft.Name != existing.Name {
		duplicate, err := s.craftRepo.GetByName(ctx, craft.Name)
		if err != nil && !errors.Is(err, apperrs.ErrCraftNotFound) {
			return nil, fmt.Errorf("%s: %w", response.InternalError, err)
		}
		if duplicate != nil && duplicate.ID != id {
			return nil, fmt.Errorf("%s: %w", response.ResourceConflict, apperrs.ErrCraftNameExists)
		}
	}

	craft.ID = id
	if err := s.craftRepo.Update(ctx, craft); err != nil {
		s.logger.Error("Failed to update craft", "error", err, "craft_id", id, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp, err := s.GetByIDWithCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("CraftService.Update success", "craft_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *craftService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("CraftService.Delete", "craft_id", id)

	_, err := s.craftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCraftNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCraftNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.craftRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete craft", "error", err, "craft_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("CraftService.Delete success", "craft_id", id, "duration", time.Since(start))
	return nil
}

func (s *craftService) GetByID(ctx context.Context, id uint) (*model.CraftResponse, error) {
	start := time.Now()

	craft, err := s.craftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCraftNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCraftNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toCraftResponse(craft)

	s.logger.Info("CraftService.GetByID success", "craft_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *craftService) GetByIDWithCategory(ctx context.Context, id uint) (*model.CraftResponse, error) {
	start := time.Now()

	craft, err := s.craftRepo.GetByIDWithCategory(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCraftNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCraftNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toCraftResponse(craft)
	if craft.Category.ID != 0 {
		resp.Category = s.toICHCategoryResponse(&craft.Category)
	}

	s.logger.Info("CraftService.GetByIDWithCategory success", "craft_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *craftService) List(ctx context.Context, orderBy string) ([]model.CraftResponse, error) {
	start := time.Now()
	s.logger.Info("CraftService.List", "orderBy", orderBy)

	crafts, err := s.craftRepo.ListWithCategory(ctx, orderBy)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CraftResponse
	for _, craft := range crafts {
		resp := s.toCraftResponse(&craft)
		if craft.Category.ID != 0 {
			resp.Category = s.toICHCategoryResponse(&craft.Category)
		}
		responses = append(responses, *resp)
	}

	s.logger.Info("CraftService.List success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *craftService) ListByCategory(ctx context.Context, categoryID uint) ([]model.CraftResponse, error) {
	start := time.Now()
	s.logger.Info("CraftService.ListByCategory", "category_id", categoryID)

	crafts, err := s.craftRepo.ListByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CraftResponse
	for _, craft := range crafts {
		responses = append(responses, *s.toCraftResponse(&craft))
	}

	s.logger.Info("CraftService.ListByCategory success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *craftService) ListByDifficulty(ctx context.Context, difficulty int8) ([]model.CraftResponse, error) {
	start := time.Now()
	s.logger.Info("CraftService.ListByDifficulty", "difficulty", difficulty)

	if err := s.validateDifficulty(difficulty); err != nil {
		return nil, err
	}

	crafts, err := s.craftRepo.ListByDifficulty(ctx, difficulty)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CraftResponse
	for _, craft := range crafts {
		responses = append(responses, *s.toCraftResponse(&craft))
	}

	s.logger.Info("CraftService.ListByDifficulty success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *craftService) validateCategoryExists(ctx context.Context, categoryID uint) error {
	if categoryID == 0 {
		return nil
	}
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, apperrs.ErrICHCategoryNotFound) {
			return fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrCategoryNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	return nil
}

func (s *craftService) validateDifficulty(difficulty int8) error {
	if difficulty < 0 || difficulty > 5 {
		return fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrInvalidDifficulty)
	}
	return nil
}

func (s *craftService) toCraftResponse(craft *model.Craft) *model.CraftResponse {
	if craft == nil {
		return nil
	}

	resp := &model.CraftResponse{
		ID:          craft.ID,
		CreatedAt:   craft.CreatedAt,
		UpdatedAt:   craft.UpdatedAt,
		CategoryID:  craft.CategoryID,
		Name:        craft.Name,
		Description: craft.Description,
		History:     craft.History,
		Difficulty:  craft.Difficulty,
	}

	if craft.Tools != nil {
		resp.Tools = s.parseJSON(craft.Tools)
	}
	if craft.RegionFeatures != nil {
		resp.RegionFeatures = s.parseJSON(craft.RegionFeatures)
	}

	return resp
}

func (s *craftService) toICHCategoryResponse(category *model.ICHCategory) *model.ICHCategoryResponse {
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

func (s *craftService) parseJSON(data datatypes.JSON) map[string]interface{} {
	if data == nil {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

var _ CraftService = (*craftService)(nil)
