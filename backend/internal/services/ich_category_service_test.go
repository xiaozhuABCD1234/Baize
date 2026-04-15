package service

import (
	"context"
	"log/slog"
	"testing"

	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockICHCategoryRepository struct {
	createErr        error
	getByIDErr       error
	getByNameErr     error
	listErr          error
	listRootErr      error
	listByParentErr  error
	listByRegionErr  error
	listActiveErr    error
	updateErr        error
	deleteErr        error
	categories       map[uint]*models.ICHCategory
	categoriesByName map[string]*models.ICHCategory
	nextID           uint
}

func newMockICHCategoryRepository() *mockICHCategoryRepository {
	return &mockICHCategoryRepository{
		categories:       make(map[uint]*models.ICHCategory),
		categoriesByName: make(map[string]*models.ICHCategory),
		nextID:           1,
	}
}

func (m *mockICHCategoryRepository) Create(ctx context.Context, category *models.ICHCategory) error {
	if m.createErr != nil {
		return m.createErr
	}
	category.ID = m.nextID
	m.nextID++
	m.categories[category.ID] = category
	m.categoriesByName[category.Name] = category
	return nil
}

func (m *mockICHCategoryRepository) GetByID(ctx context.Context, id uint) (*models.ICHCategory, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if category, ok := m.categories[id]; ok {
		return category, nil
	}
	return nil, repository.ErrICHCategoryNotFound
}

func (m *mockICHCategoryRepository) GetByName(ctx context.Context, name string) (*models.ICHCategory, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if category, ok := m.categoriesByName[name]; ok {
		return category, nil
	}
	return nil, repository.ErrICHCategoryNotFound
}

func (m *mockICHCategoryRepository) GetByIDWithChildren(ctx context.Context, id uint) (*models.ICHCategory, error) {
	return m.GetByID(ctx, id)
}

func (m *mockICHCategoryRepository) List(ctx context.Context, orderBy string) ([]models.ICHCategory, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.ICHCategory
	for _, c := range m.categories {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockICHCategoryRepository) ListRoot(ctx context.Context) ([]models.ICHCategory, error) {
	if m.listRootErr != nil {
		return nil, m.listRootErr
	}
	var result []models.ICHCategory
	for _, c := range m.categories {
		if c.ParentID == 0 {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockICHCategoryRepository) ListByParentID(ctx context.Context, parentID uint) ([]models.ICHCategory, error) {
	if m.listByParentErr != nil {
		return nil, m.listByParentErr
	}
	var result []models.ICHCategory
	for _, c := range m.categories {
		if c.ParentID == parentID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockICHCategoryRepository) ListByRegionCode(ctx context.Context, regionCode string) ([]models.ICHCategory, error) {
	if m.listByRegionErr != nil {
		return nil, m.listByRegionErr
	}
	var result []models.ICHCategory
	for _, c := range m.categories {
		if c.RegionCode == regionCode {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockICHCategoryRepository) ListActive(ctx context.Context) ([]models.ICHCategory, error) {
	if m.listActiveErr != nil {
		return nil, m.listActiveErr
	}
	var result []models.ICHCategory
	for _, c := range m.categories {
		if c.Status == 1 {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockICHCategoryRepository) Update(ctx context.Context, category *models.ICHCategory) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.categories[category.ID]; !ok {
		return repository.ErrICHCategoryNotFound
	}
	m.categories[category.ID] = category
	return nil
}

func (m *mockICHCategoryRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockICHCategoryRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.categories[id]; !ok {
		return repository.ErrICHCategoryNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockICHCategoryRepository) ForceDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockICHCategoryRepository) Upsert(ctx context.Context, category *models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepository) UpsertBatch(ctx context.Context, categories []models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepository) WithTransaction(tx *gorm.DB) repository.ICHCategoryRepository {
	return m
}

func (m *mockICHCategoryRepository) addTestCategory(id uint, name string, level int8, parentID uint) {
	m.categories[id] = &models.ICHCategory{
		Name:     name,
		Level:    level,
		ParentID: parentID,
		Status:   1,
	}
	m.categoriesByName[name] = m.categories[id]
	m.categoriesByName[name].ID = id
}

func TestICHCategoryService_Create_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	category := &models.ICHCategory{
		Name:  "Traditional Crafts",
		Level: 1,
	}
	resp, err := svc.Create(context.Background(), category)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Traditional Crafts", resp.Name)
}

func TestICHCategoryService_Create_DuplicateName(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	category := &models.ICHCategory{
		Name:  "Traditional Crafts",
		Level: 1,
	}
	resp, err := svc.Create(context.Background(), category)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "分类名称已存在")
}

func TestICHCategoryService_Create_InvalidLevel(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	category := &models.ICHCategory{
		Name:  "Test",
		Level: 10,
	}
	resp, err := svc.Create(context.Background(), category)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "无效的分类级别")
}

func TestICHCategoryService_GetByID_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	resp, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Traditional Crafts", resp.Name)
}

func TestICHCategoryService_GetByID_NotFound(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	_, err := svc.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "非遗分类不存在")
}

func TestICHCategoryService_GetByName_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	resp, err := svc.GetByName(context.Background(), "Traditional Crafts")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
}

func TestICHCategoryService_Update_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	category := &models.ICHCategory{
		Name: "Updated Crafts",
	}
	resp, err := svc.Update(context.Background(), 1, category)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Crafts", resp.Name)
}

func TestICHCategoryService_Delete_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestICHCategoryService_Delete_WithChildren(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	categoryRepo.addTestCategory(2, "Woodworking", 2, 1)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法删除有子节点的分类")
}

func TestICHCategoryService_List_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Crafts 1", 1, 0)
	categoryRepo.addTestCategory(2, "Crafts 2", 1, 0)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	list, err := svc.List(context.Background(), "name asc")

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestICHCategoryService_ListRoot_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	categoryRepo.addTestCategory(2, "Woodworking", 2, 1)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	list, err := svc.ListRoot(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Traditional Crafts", list[0].Name)
}

func TestICHCategoryService_ListByParentID_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Traditional Crafts", 1, 0)
	categoryRepo.addTestCategory(2, "Woodworking", 2, 1)
	categoryRepo.addTestCategory(3, "Metalwork", 2, 1)
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	list, err := svc.ListByParentID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestICHCategoryService_ListActive_Success(t *testing.T) {
	categoryRepo := newMockICHCategoryRepository()
	categoryRepo.addTestCategory(1, "Active Craft", 1, 0)
	categoryRepo.categories[2] = &models.ICHCategory{
		Name:   "Inactive Craft",
		Level:  1,
		Status: 0,
	}
	logger := slog.Default()
	svc := NewICHCategoryService(categoryRepo, logger)

	list, err := svc.ListActive(context.Background())

	assert.NoError(t, err)
	assert.Len(t, list, 1)
}
