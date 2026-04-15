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

type mockCraftRepository struct {
	createErr           error
	getByIDErr          error
	getByNameErr        error
	listErr             error
	listByCategoryErr   error
	listByDifficultyErr error
	updateErr           error
	deleteErr           error
	crafts              map[uint]*models.Craft
	craftsByName        map[string]*models.Craft
	nextID              uint
}

func newMockCraftRepository() *mockCraftRepository {
	return &mockCraftRepository{
		crafts:       make(map[uint]*models.Craft),
		craftsByName: make(map[string]*models.Craft),
		nextID:       1,
	}
}

func (m *mockCraftRepository) Create(ctx context.Context, craft *models.Craft) error {
	if m.createErr != nil {
		return m.createErr
	}
	craft.ID = m.nextID
	m.nextID++
	m.crafts[craft.ID] = craft
	m.craftsByName[craft.Name] = craft
	return nil
}

func (m *mockCraftRepository) GetByID(ctx context.Context, id uint) (*models.Craft, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if craft, ok := m.crafts[id]; ok {
		return craft, nil
	}
	return nil, repository.ErrCraftNotFound
}

func (m *mockCraftRepository) GetByIDWithCategory(ctx context.Context, id uint) (*models.Craft, error) {
	return m.GetByID(ctx, id)
}

func (m *mockCraftRepository) GetByName(ctx context.Context, name string) (*models.Craft, error) {
	if m.getByNameErr != nil {
		return nil, m.getByNameErr
	}
	if craft, ok := m.craftsByName[name]; ok {
		return craft, nil
	}
	return nil, repository.ErrCraftNotFound
}

func (m *mockCraftRepository) List(ctx context.Context, orderBy string) ([]models.Craft, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.Craft
	for _, c := range m.crafts {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockCraftRepository) ListWithCategory(ctx context.Context, orderBy string) ([]models.Craft, error) {
	return m.List(ctx, orderBy)
}

func (m *mockCraftRepository) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Craft, error) {
	if m.listByCategoryErr != nil {
		return nil, m.listByCategoryErr
	}
	var result []models.Craft
	for _, c := range m.crafts {
		if c.CategoryID == categoryID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCraftRepository) ListByDifficulty(ctx context.Context, difficulty int8) ([]models.Craft, error) {
	if m.listByDifficultyErr != nil {
		return nil, m.listByDifficultyErr
	}
	var result []models.Craft
	for _, c := range m.crafts {
		if c.Difficulty == difficulty {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCraftRepository) Update(ctx context.Context, craft *models.Craft) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.crafts[craft.ID]; !ok {
		return repository.ErrCraftNotFound
	}
	m.crafts[craft.ID] = craft
	return nil
}

func (m *mockCraftRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockCraftRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.crafts[id]; !ok {
		return repository.ErrCraftNotFound
	}
	delete(m.crafts, id)
	return nil
}

func (m *mockCraftRepository) ForceDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockCraftRepository) Upsert(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepository) UpsertBatch(ctx context.Context, crafts []models.Craft) error {
	return nil
}

func (m *mockCraftRepository) WithTransaction(tx *gorm.DB) repository.CraftRepository {
	return m
}

func (m *mockCraftRepository) addTestCraft(id uint, name string, categoryID uint, difficulty int8) {
	m.crafts[id] = &models.Craft{
		Name:       name,
		CategoryID: categoryID,
		Difficulty: difficulty,
	}
	m.craftsByName[name] = m.crafts[id]
}

type mockICHCategoryRepositoryForCraft struct {
	getByIDErr error
	categories map[uint]*models.ICHCategory
}

func newMockICHCategoryRepositoryForCraft() *mockICHCategoryRepositoryForCraft {
	return &mockICHCategoryRepositoryForCraft{
		categories: make(map[uint]*models.ICHCategory),
	}
}

func (m *mockICHCategoryRepositoryForCraft) GetByID(ctx context.Context, id uint) (*models.ICHCategory, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if category, ok := m.categories[id]; ok {
		return category, nil
	}
	return nil, repository.ErrICHCategoryNotFound
}

func (m *mockICHCategoryRepositoryForCraft) Create(ctx context.Context, category *models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) GetByName(ctx context.Context, name string) (*models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) GetByIDWithChildren(ctx context.Context, id uint) (*models.ICHCategory, error) {
	return m.GetByID(ctx, id)
}

func (m *mockICHCategoryRepositoryForCraft) List(ctx context.Context, orderBy string) ([]models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) ListRoot(ctx context.Context) ([]models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) ListByParentID(ctx context.Context, parentID uint) ([]models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) ListByRegionCode(ctx context.Context, regionCode string) ([]models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) ListActive(ctx context.Context) ([]models.ICHCategory, error) {
	return nil, nil
}

func (m *mockICHCategoryRepositoryForCraft) Update(ctx context.Context, category *models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) Upsert(ctx context.Context, category *models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) UpsertBatch(ctx context.Context, categories []models.ICHCategory) error {
	return nil
}

func (m *mockICHCategoryRepositoryForCraft) WithTransaction(tx *gorm.DB) repository.ICHCategoryRepository {
	return m
}

func (m *mockICHCategoryRepositoryForCraft) addTestCategory(id uint) {
	m.categories[id] = &models.ICHCategory{
		Name:   "Category " + string(rune('0'+id)),
		Level:  1,
		Status: 1,
	}
}

func TestCraftService_Create_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	categoryRepo.addTestCategory(1)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	craft := &models.Craft{
		Name:       "Woodcarving",
		CategoryID: 1,
		Difficulty: 3,
	}
	resp, err := svc.Create(context.Background(), craft)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Woodcarving", resp.Name)
	assert.Equal(t, int8(3), resp.Difficulty)
}

func TestCraftService_Create_DuplicateName(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	categoryRepo.addTestCategory(1)
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	craft := &models.Craft{
		Name:       "Woodcarving",
		CategoryID: 1,
		Difficulty: 3,
	}
	resp, err := svc.Create(context.Background(), craft)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "技艺名称已存在")
}

func TestCraftService_Create_InvalidDifficulty(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	categoryRepo.addTestCategory(1)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	craft := &models.Craft{
		Name:       "Woodcarving",
		CategoryID: 1,
		Difficulty: 10,
	}
	resp, err := svc.Create(context.Background(), craft)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "无效的难度等级")
}

func TestCraftService_GetByID_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	resp, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Woodcarving", resp.Name)
}

func TestCraftService_GetByID_NotFound(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	_, err := svc.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "技艺不存在")
}

func TestCraftService_Update_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	categoryRepo.addTestCategory(1)
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	craft := &models.Craft{
		Name: "Updated Woodcarving",
	}
	resp, err := svc.Update(context.Background(), 1, craft)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Woodcarving", resp.Name)
}

func TestCraftService_Delete_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestCraftService_List_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)
	craftRepo.addTestCraft(2, "Metalwork", 1, 4)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	list, err := svc.List(context.Background(), "name asc")

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCraftService_ListByCategory_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	categoryRepo.addTestCategory(1)
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)
	craftRepo.addTestCraft(2, "Metalwork", 2, 4)
	craftRepo.addTestCraft(3, "Stonework", 1, 5)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	list, err := svc.ListByCategory(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCraftService_ListByDifficulty_Success(t *testing.T) {
	craftRepo := newMockCraftRepository()
	categoryRepo := newMockICHCategoryRepositoryForCraft()
	craftRepo.addTestCraft(1, "Woodcarving", 1, 3)
	craftRepo.addTestCraft(2, "Metalwork", 1, 4)
	craftRepo.addTestCraft(3, "Stonework", 1, 3)

	logger := slog.Default()
	svc := NewCraftService(craftRepo, categoryRepo, logger)

	list, err := svc.ListByDifficulty(context.Background(), 3)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}
