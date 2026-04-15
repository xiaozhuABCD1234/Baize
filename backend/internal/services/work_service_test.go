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

type mockWorkRepositoryForWork struct {
	getByIDErr    error
	getByIDAllErr error
	works         map[uint]*models.Work
	nextID        uint
}

func newMockWorkRepositoryForWork() *mockWorkRepositoryForWork {
	return &mockWorkRepositoryForWork{
		works:  make(map[uint]*models.Work),
		nextID: 1,
	}
}

func (m *mockWorkRepositoryForWork) GetByID(ctx context.Context, id uint) (*models.Work, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if work, ok := m.works[id]; ok {
		work.ID = id
		return work, nil
	}
	return nil, apperrs.ErrWorkNotFound
}

func (m *mockWorkRepositoryForWork) GetByIDWithAll(ctx context.Context, id uint) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepositoryForWork) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepositoryForWork) Create(ctx context.Context, work *models.Work) error {
	work.ID = m.nextID
	m.nextID++
	m.works[work.ID] = work
	return nil
}

func (m *mockWorkRepositoryForWork) List(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListWithAll(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.Work, int64, error) {
	return nil, 0, nil
}

func (m *mockWorkRepositoryForWork) ListByUserID(ctx context.Context, userID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListByCraftID(ctx context.Context, craftID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListByStatus(ctx context.Context, status models.WorkStatus) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListPublished(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListTop(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) ListRecommended(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForWork) Update(ctx context.Context, work *models.Work) error {
	if _, ok := m.works[work.ID]; !ok {
		return apperrs.ErrWorkNotFound
	}
	m.works[work.ID] = work
	return nil
}

func (m *mockWorkRepositoryForWork) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockWorkRepositoryForWork) UpdateStatus(ctx context.Context, id uint, status models.WorkStatus) error {
	return nil
}

func (m *mockWorkRepositoryForWork) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	return nil
}

func (m *mockWorkRepositoryForWork) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepositoryForWork) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepositoryForWork) Upsert(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepositoryForWork) UpsertBatch(ctx context.Context, works []models.Work) error {
	return nil
}

func (m *mockWorkRepositoryForWork) WithTransaction(tx *gorm.DB) repository.WorkRepository {
	return m
}

type mockWorkMediaRepository struct {
	createBatchErr error
	media          map[uint][]models.WorkMedia
}

func newMockWorkMediaRepository() *mockWorkMediaRepository {
	return &mockWorkMediaRepository{
		media: make(map[uint][]models.WorkMedia),
	}
}

func (m *mockWorkMediaRepository) Create(ctx context.Context, media *models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) CreateBatch(ctx context.Context, mediaList []models.WorkMedia) error {
	if m.createBatchErr != nil {
		return m.createBatchErr
	}
	if len(mediaList) > 0 {
		m.media[mediaList[0].WorkID] = mediaList
	}
	return nil
}

func (m *mockWorkMediaRepository) GetByID(ctx context.Context, id uint) (*models.WorkMedia, error) {
	return nil, nil
}

func (m *mockWorkMediaRepository) ListByWorkID(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return m.media[workID], nil
}

func (m *mockWorkMediaRepository) ListImages(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return nil, nil
}

func (m *mockWorkMediaRepository) ListVideos(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return nil, nil
}

func (m *mockWorkMediaRepository) Update(ctx context.Context, media *models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockWorkMediaRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkMediaRepository) DeleteByWorkID(ctx context.Context, workID uint) error {
	delete(m.media, workID)
	return nil
}

func (m *mockWorkMediaRepository) DeleteBatch(ctx context.Context, ids []uint) error {
	return nil
}

func (m *mockWorkMediaRepository) Upsert(ctx context.Context, media *models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) UpsertBatch(ctx context.Context, mediaList []models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) WithTransaction(tx *gorm.DB) repository.WorkMediaRepository {
	return m
}

type mockCraftRepositoryForWork struct {
	getByIDErr error
	crafts     map[uint]*models.Craft
}

func newMockCraftRepositoryForWork() *mockCraftRepositoryForWork {
	return &mockCraftRepositoryForWork{
		crafts: make(map[uint]*models.Craft),
	}
}

func (m *mockCraftRepositoryForWork) GetByID(ctx context.Context, id uint) (*models.Craft, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if craft, ok := m.crafts[id]; ok {
		return craft, nil
	}
	return nil, apperrs.ErrCraftNotFound
}

func (m *mockCraftRepositoryForWork) Create(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepositoryForWork) GetByName(ctx context.Context, name string) (*models.Craft, error) {
	return nil, nil
}

func (m *mockCraftRepositoryForWork) GetByIDWithCategory(ctx context.Context, id uint) (*models.Craft, error) {
	return m.GetByID(ctx, id)
}

func (m *mockCraftRepositoryForWork) List(ctx context.Context, orderBy string) ([]models.Craft, error) {
	return nil, nil
}

func (m *mockCraftRepositoryForWork) ListWithCategory(ctx context.Context, orderBy string) ([]models.Craft, error) {
	return nil, nil
}

func (m *mockCraftRepositoryForWork) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Craft, error) {
	return nil, nil
}

func (m *mockCraftRepositoryForWork) ListByDifficulty(ctx context.Context, difficulty int8) ([]models.Craft, error) {
	return nil, nil
}

func (m *mockCraftRepositoryForWork) Update(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepositoryForWork) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockCraftRepositoryForWork) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockCraftRepositoryForWork) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockCraftRepositoryForWork) Upsert(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepositoryForWork) UpsertBatch(ctx context.Context, crafts []models.Craft) error {
	return nil
}

func (m *mockCraftRepositoryForWork) WithTransaction(tx *gorm.DB) repository.CraftRepository {
	return m
}

func (m *mockCraftRepositoryForWork) addTestCraft(id uint) {
	m.crafts[id] = &models.Craft{
		Name: "Craft " + string(rune('0'+id)),
	}
}

type mockUserRepositoryForWork struct {
	getByIDErr error
	users      map[uint]*models.User
}

func newMockUserRepositoryForWork() *mockUserRepositoryForWork {
	return &mockUserRepositoryForWork{
		users: make(map[uint]*models.User),
	}
}

func (m *mockUserRepositoryForWork) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, apperrs.ErrUserNotFound
}

func (m *mockUserRepositoryForWork) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForWork) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepositoryForWork) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForWork) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForWork) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForWork) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForWork) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForWork) List(ctx context.Context, orderBy string) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForWork) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepositoryForWork) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForWork) Update(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForWork) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return nil
}

func (m *mockUserRepositoryForWork) UpdateEmail(ctx context.Context, id uint, email string) error {
	return nil
}

func (m *mockUserRepositoryForWork) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepositoryForWork) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForWork) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForWork) Upsert(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForWork) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func (m *mockUserRepositoryForWork) addTestUser(id uint) {
	m.users[id] = &models.User{
		Username: "user" + string(rune('0'+id)),
		Email:    "user" + string(rune('0'+id)) + "@example.com",
	}
}

func TestWorkService_Create_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	craftRepo.addTestCraft(1)
	userRepo := newMockUserRepositoryForWork()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	req := &models.CreateWorkRequest{
		Title:      "My First Work",
		Content:    "This is the content",
		CraftID:    1,
		CategoryID: 0,
		RegionID:   0,
	}
	resp, err := svc.Create(context.Background(), req, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "My First Work", resp.Title)
}

func TestWorkService_GetByID_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Test Work"}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	resp, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Work", resp.Title)
}

func TestWorkService_GetByID_NotFound(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	_, err := svc.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "作品不存在")
}

func TestWorkService_Update_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Original Title"}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	craftRepo.addTestCraft(1)
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	req := &models.CreateWorkRequest{
		Title:   "Updated Title",
		CraftID: 1,
	}
	resp, err := svc.Update(context.Background(), 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Updated Title", resp.Title)
}

func TestWorkService_Delete_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Draft Work", Status: models.WorkStatusDraft}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.NoError(t, err)
}

func TestWorkService_Delete_PublishedWork(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Published Work", Status: models.WorkStatusPublished}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	err := svc.Delete(context.Background(), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法删除已发布的作品")
}

func TestWorkService_UpdateStatus_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Test Work", Status: models.WorkStatusDraft}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	err := svc.UpdateStatus(context.Background(), 1, models.WorkStatusPublished)

	assert.NoError(t, err)
}

func TestWorkService_UpdateStatus_Invalid(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Test Work", Status: models.WorkStatusDraft}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	err := svc.UpdateStatus(context.Background(), 1, 99)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的作品状态")
}

func TestWorkService_IncrementCount_Success(t *testing.T) {
	workRepo := newMockWorkRepositoryForWork()
	workRepo.works[1] = &models.Work{Title: "Test Work", ViewCount: 10}
	mediaRepo := newMockWorkMediaRepository()
	craftRepo := newMockCraftRepositoryForWork()
	userRepo := newMockUserRepositoryForWork()

	logger := slog.Default()
	svc := NewWorkService(workRepo, mediaRepo, craftRepo, userRepo, logger)

	err := svc.IncrementCount(context.Background(), 1, "view_count", 1)

	assert.NoError(t, err)
}
