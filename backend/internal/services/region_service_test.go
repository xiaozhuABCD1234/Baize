package service

import (
	"context"
	"log/slog"
	"testing"

	apperrs "backend/internal/errors"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockRegionRepository struct {
	createErr       error
	getByIDErr      error
	getByCodeErr    error
	listErr         error
	listRootErr     error
	listByParentErr error
	listByLevelErr  error
	listHeritageErr error
	updateErr       error
	deleteErr       error
	regions         map[uint]*models.Region
	regionsByCode   map[string]*models.Region
	nextID          uint
}

func newMockRegionRepository() *mockRegionRepository {
	return &mockRegionRepository{
		regions:       make(map[uint]*models.Region),
		regionsByCode: make(map[string]*models.Region),
		nextID:        1,
	}
}

func (m *mockRegionRepository) Create(ctx context.Context, region *models.Region) error {
	if m.createErr != nil {
		return m.createErr
	}
	region.ID = m.nextID
	m.nextID++
	m.regions[region.ID] = region
	m.regionsByCode[region.Code] = region
	return nil
}

func (m *mockRegionRepository) GetByID(ctx context.Context, id uint) (*models.Region, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if region, ok := m.regions[id]; ok {
		return region, nil
	}
	return nil, apperrs.ErrRegionNotFound
}

func (m *mockRegionRepository) GetByCode(ctx context.Context, code string) (*models.Region, error) {
	if m.getByCodeErr != nil {
		return nil, m.getByCodeErr
	}
	if region, ok := m.regionsByCode[code]; ok {
		return region, nil
	}
	return nil, apperrs.ErrRegionNotFound
}

func (m *mockRegionRepository) GetByIDWithChildren(ctx context.Context, id uint) (*models.Region, error) {
	return m.GetByID(ctx, id)
}

func (m *mockRegionRepository) List(ctx context.Context, orderBy string) ([]models.Region, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.Region
	for _, r := range m.regions {
		result = append(result, *r)
	}
	return result, nil
}

func (m *mockRegionRepository) ListRoot(ctx context.Context) ([]models.Region, error) {
	if m.listRootErr != nil {
		return nil, m.listRootErr
	}
	var result []models.Region
	for _, r := range m.regions {
		if r.ParentID == 0 {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *mockRegionRepository) ListByParentID(ctx context.Context, parentID uint) ([]models.Region, error) {
	if m.listByParentErr != nil {
		return nil, m.listByParentErr
	}
	var result []models.Region
	for _, r := range m.regions {
		if r.ParentID == parentID {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *mockRegionRepository) ListByLevel(ctx context.Context, level int8) ([]models.Region, error) {
	if m.listByLevelErr != nil {
		return nil, m.listByLevelErr
	}
	var result []models.Region
	for _, r := range m.regions {
		if r.Level == level {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *mockRegionRepository) ListHeritageCenters(ctx context.Context) ([]models.Region, error) {
	if m.listHeritageErr != nil {
		return nil, m.listHeritageErr
	}
	var result []models.Region
	for _, r := range m.regions {
		if r.IsHeritageCenter {
			result = append(result, *r)
		}
	}
	return result, nil
}

func (m *mockRegionRepository) Update(ctx context.Context, region *models.Region) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.regions[region.ID]; !ok {
		return apperrs.ErrRegionNotFound
	}
	m.regions[region.ID] = region
	return nil
}

func (m *mockRegionRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockRegionRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.regions[id]; !ok {
		return apperrs.ErrRegionNotFound
	}
	delete(m.regions, id)
	return nil
}

func (m *mockRegionRepository) ForceDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockRegionRepository) Upsert(ctx context.Context, region *models.Region) error {
	return nil
}

func (m *mockRegionRepository) UpsertBatch(ctx context.Context, regions []models.Region) error {
	return nil
}

func (m *mockRegionRepository) WithTransaction(tx *gorm.DB) repository.RegionRepository {
	return m
}

func (m *mockRegionRepository) addTestRegion(id uint, code string, level int8, parentID uint) {
	m.regions[id] = &models.Region{
		Code:     code,
		Name:     "Region " + code,
		Level:    level,
		ParentID: parentID,
	}
	m.regions[id].ID = id
	m.regionsByCode[code] = m.regions[id]
}

func TestRegionService_Create_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	region := &models.Region{
		Name:  "Beijing",
		Code:  "BJ",
		Level: 1,
	}
	resp, err := svc.Create(context.Background(), region)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Beijing", resp.Name)
	assert.Equal(t, "BJ", resp.Code)
}

func TestRegionService_Create_DuplicateCode(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	region := &models.Region{
		Name:  "Beijing 2",
		Code:  "BJ",
		Level: 1,
	}
	resp, err := svc.Create(context.Background(), region)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "地区代码已存在")
}

func TestRegionService_Create_InvalidLevel(t *testing.T) {
	regionRepo := newMockRegionRepository()
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	region := &models.Region{
		Name:  "Test",
		Code:  "TEST",
		Level: 10,
	}
	resp, err := svc.Create(context.Background(), region)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "无效的地区级别")
}

func TestRegionService_GetByID_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	resp, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "BJ", resp.Code)
}

func TestRegionService_GetByID_NotFound(t *testing.T) {
	regionRepo := newMockRegionRepository()
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	_, err := svc.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "地区不存在")
}

func TestRegionService_GetByCode_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	resp, err := svc.GetByCode(context.Background(), "BJ")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
}

func TestRegionService_Update_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	region := &models.Region{
		Name: "Beijing Updated",
	}
	resp, err := svc.Update(context.Background(), 1, region)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Beijing Updated", resp.Name)
}

func TestRegionService_Delete_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestRegionService_Delete_WithChildren(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	regionRepo.addTestRegion(2, "CY", 2, 1)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法删除有子节点的地区")
}

func TestRegionService_List_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	regionRepo.addTestRegion(2, "SH", 1, 0)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	list, err := svc.List(context.Background(), "code asc")

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRegionService_ListRoot_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	regionRepo.addTestRegion(2, "CY", 2, 1)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	list, err := svc.ListRoot(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "BJ", list[0].Code)
}

func TestRegionService_ListByParentID_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	regionRepo.addTestRegion(2, "CY", 2, 1)
	regionRepo.addTestRegion(3, "HD", 2, 1)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	list, err := svc.ListByParentID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestRegionService_ListByLevel_Success(t *testing.T) {
	regionRepo := newMockRegionRepository()
	regionRepo.addTestRegion(1, "BJ", 1, 0)
	regionRepo.addTestRegion(2, "SH", 1, 0)
	regionRepo.addTestRegion(3, "CY", 2, 1)
	logger := slog.Default()
	svc := NewRegionService(regionRepo, logger)

	list, err := svc.ListByLevel(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}
