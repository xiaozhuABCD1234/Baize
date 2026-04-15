package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockFavoriteRepository struct {
	createErr           error
	getByIDErr          error
	getByUserAndWorkErr error
	listErr             error
	listByUserErr       error
	listByWorkErr       error
	listPagErr          error
	updateErr           error
	updateFolderErr     error
	deleteErr           error
	deleteByUserWorkErr error
	existsErr           error
	countErr            error
	favorites           map[uint]*models.Favorite
	favoritesByUser     map[uint][]*models.Favorite
	favoritesByWork     map[uint][]*models.Favorite
	nextID              uint
}

func newMockFavoriteRepository() *mockFavoriteRepository {
	return &mockFavoriteRepository{
		favorites:       make(map[uint]*models.Favorite),
		favoritesByUser: make(map[uint][]*models.Favorite),
		favoritesByWork: make(map[uint][]*models.Favorite),
		nextID:          1,
	}
}

func (m *mockFavoriteRepository) Create(ctx context.Context, favorite *models.Favorite) error {
	if m.createErr != nil {
		return m.createErr
	}
	favorite.ID = m.nextID
	favorite.CreatedAt = time.Now()
	m.nextID++
	m.favorites[favorite.ID] = favorite
	m.favoritesByUser[favorite.UserID] = append(m.favoritesByUser[favorite.UserID], favorite)
	m.favoritesByWork[favorite.WorkID] = append(m.favoritesByWork[favorite.WorkID], favorite)
	return nil
}

func (m *mockFavoriteRepository) GetByID(ctx context.Context, id uint) (*models.Favorite, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if favorite, ok := m.favorites[id]; ok {
		return favorite, nil
	}
	return nil, repository.ErrFavoriteNotFound
}

func (m *mockFavoriteRepository) GetByUserAndWork(ctx context.Context, userID, workID uint) (*models.Favorite, error) {
	if m.getByUserAndWorkErr != nil {
		return nil, m.getByUserAndWorkErr
	}
	for _, fav := range m.favoritesByUser[userID] {
		if fav.WorkID == workID {
			return fav, nil
		}
	}
	return nil, repository.ErrFavoriteNotFound
}

func (m *mockFavoriteRepository) List(ctx context.Context, orderBy string) ([]models.Favorite, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.Favorite
	for _, fav := range m.favorites {
		result = append(result, *fav)
	}
	return result, nil
}

func (m *mockFavoriteRepository) ListByUserID(ctx context.Context, userID uint) ([]models.Favorite, error) {
	if m.listByUserErr != nil {
		return nil, m.listByUserErr
	}
	var result []models.Favorite
	for _, fav := range m.favoritesByUser[userID] {
		result = append(result, *fav)
	}
	return result, nil
}

func (m *mockFavoriteRepository) ListByWorkID(ctx context.Context, workID uint) ([]models.Favorite, error) {
	if m.listByWorkErr != nil {
		return nil, m.listByWorkErr
	}
	var result []models.Favorite
	for _, fav := range m.favoritesByWork[workID] {
		result = append(result, *fav)
	}
	return result, nil
}

func (m *mockFavoriteRepository) ListByFolderID(ctx context.Context, userID, folderID uint) ([]models.Favorite, error) {
	return nil, nil
}

func (m *mockFavoriteRepository) ListByUserIDWithWorks(ctx context.Context, userID uint, orderBy string) ([]models.Favorite, error) {
	return m.ListByUserID(ctx, userID)
}

func (m *mockFavoriteRepository) ListWithPagination(ctx context.Context, userID uint, page, pageSize int) ([]models.Favorite, int64, error) {
	if m.listPagErr != nil {
		return nil, 0, m.listPagErr
	}
	var result []models.Favorite
	for _, fav := range m.favoritesByUser[userID] {
		result = append(result, *fav)
	}
	return result, int64(len(result)), nil
}

func (m *mockFavoriteRepository) Update(ctx context.Context, favorite *models.Favorite) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.favorites[favorite.ID]; !ok {
		return repository.ErrFavoriteNotFound
	}
	m.favorites[favorite.ID] = favorite
	return nil
}

func (m *mockFavoriteRepository) UpdateFolder(ctx context.Context, id uint, folderID uint) error {
	if m.updateFolderErr != nil {
		return m.updateFolderErr
	}
	if _, ok := m.favorites[id]; !ok {
		return repository.ErrFavoriteNotFound
	}
	m.favorites[id].FolderID = folderID
	return nil
}

func (m *mockFavoriteRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.favorites[id]; !ok {
		return repository.ErrFavoriteNotFound
	}
	delete(m.favorites, id)
	return nil
}

func (m *mockFavoriteRepository) DeleteByUserAndWork(ctx context.Context, userID, workID uint) error {
	if m.deleteByUserWorkErr != nil {
		return m.deleteByUserWorkErr
	}
	for i, fav := range m.favoritesByUser[userID] {
		if fav.WorkID == workID {
			m.favoritesByUser[userID] = append(m.favoritesByUser[userID][:i], m.favoritesByUser[userID][i+1:]...)
			delete(m.favorites, fav.ID)
			return nil
		}
	}
	return repository.ErrFavoriteNotFound
}

func (m *mockFavoriteRepository) DeleteByWorkID(ctx context.Context, workID uint) error {
	return nil
}

func (m *mockFavoriteRepository) Exists(ctx context.Context, userID, workID uint) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	for _, fav := range m.favoritesByUser[userID] {
		if fav.WorkID == workID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockFavoriteRepository) CountByWorkID(ctx context.Context, workID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return int64(len(m.favoritesByWork[workID])), nil
}

func (m *mockFavoriteRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return int64(len(m.favoritesByUser[userID])), nil
}

func (m *mockFavoriteRepository) WithTransaction(tx *gorm.DB) repository.FavoriteRepository {
	return m
}

type mockWorkRepository struct {
	getByIDErr    error
	getByIDAllErr error
	works         map[uint]*models.Work
}

func newMockWorkRepository() *mockWorkRepository {
	return &mockWorkRepository{
		works: make(map[uint]*models.Work),
	}
}

func (m *mockWorkRepository) GetByID(ctx context.Context, id uint) (*models.Work, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if work, ok := m.works[id]; ok {
		return work, nil
	}
	return nil, repository.ErrWorkNotFound
}

func (m *mockWorkRepository) GetByIDWithAll(ctx context.Context, id uint) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepository) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepository) Create(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepository) List(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListWithAll(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.Work, int64, error) {
	return nil, 0, nil
}

func (m *mockWorkRepository) ListByUserID(ctx context.Context, userID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListByCraftID(ctx context.Context, craftID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListByStatus(ctx context.Context, status models.WorkStatus) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListPublished(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListTop(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) ListRecommended(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepository) Update(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockWorkRepository) UpdateStatus(ctx context.Context, id uint, status models.WorkStatus) error {
	return nil
}

func (m *mockWorkRepository) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	return nil
}

func (m *mockWorkRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepository) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepository) Upsert(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepository) UpsertBatch(ctx context.Context, works []models.Work) error {
	return nil
}

func (m *mockWorkRepository) WithTransaction(tx *gorm.DB) repository.WorkRepository {
	return m
}

func (m *mockWorkRepository) addTestWork(id uint) {
	m.works[id] = &models.Work{
		Title: "Work " + string(rune('0'+id)),
	}
	m.works[id].ID = id
}

type mockUserRepositoryForFavorite struct {
	getByIDErr error
	users      map[uint]*models.User
}

func newMockUserRepositoryForFavorite() *mockUserRepositoryForFavorite {
	return &mockUserRepositoryForFavorite{
		users: make(map[uint]*models.User),
	}
}

func (m *mockUserRepositoryForFavorite) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepositoryForFavorite) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForFavorite) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForFavorite) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForFavorite) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForFavorite) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForFavorite) List(ctx context.Context, orderBy string) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForFavorite) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepositoryForFavorite) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForFavorite) Update(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) UpdateEmail(ctx context.Context, id uint, email string) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) Upsert(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForFavorite) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func (m *mockUserRepositoryForFavorite) addTestUser(id uint) {
	m.users[id] = &models.User{
		Username: "user" + string(rune('0'+id)),
		Email:    "user" + string(rune('0'+id)) + "@example.com",
	}
}

func TestFavoriteService_Create_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 1, FolderID: 0}
	resp, err := svc.Create(context.Background(), req, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.UserID)
	assert.Equal(t, uint(1), resp.WorkID)
}

func TestFavoriteService_Create_AlreadyFavorited(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 1}
	_, _ = svc.Create(context.Background(), req, 1)
	resp, err := svc.Create(context.Background(), req, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "已经收藏过该作品")
}

func TestFavoriteService_Create_WorkNotFound(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 999}
	resp, err := svc.Create(context.Background(), req, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "作品不存在")
}

func TestFavoriteService_Delete_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 1}
	resp, _ := svc.Create(context.Background(), req, 1)
	err := svc.Delete(context.Background(), resp.ID)

	assert.NoError(t, err)
}

func TestFavoriteService_Delete_NotFound(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	err := svc.Delete(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "收藏不存在")
}

func TestFavoriteService_GetByID_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 1}
	resp, _ := svc.Create(context.Background(), req, 1)
	getResp, err := svc.GetByID(context.Background(), resp.ID)

	assert.NoError(t, err)
	assert.NotNil(t, getResp)
	assert.Equal(t, resp.ID, getResp.ID)
}

func TestFavoriteService_ListByUserID_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	workRepo.addTestWork(2)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 1)
	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 2}, 1)

	list, total, err := svc.ListByUserID(context.Background(), 1, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestFavoriteService_ListByWorkID_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 1)
	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 2)

	list, err := svc.ListByWorkID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFavoriteService_UpdateFolder_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	req := &models.FavoriteRequest{WorkID: 1, FolderID: 0}
	resp, _ := svc.Create(context.Background(), req, 1)

	err := svc.UpdateFolder(context.Background(), resp.ID, 5)

	assert.NoError(t, err)
}

func TestFavoriteService_Exists_True(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 1)

	exists, err := svc.Exists(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestFavoriteService_Exists_False(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	exists, err := svc.Exists(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestFavoriteService_CountByWorkID_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 1)
	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 2)

	count, err := svc.CountByWorkID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestFavoriteService_DeleteByUserAndWork_Success(t *testing.T) {
	favRepo := newMockFavoriteRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)
	userRepo := newMockUserRepositoryForFavorite()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFavoriteService(favRepo, workRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), &models.FavoriteRequest{WorkID: 1}, 1)

	err := svc.DeleteByUserAndWork(context.Background(), 1, 1)

	assert.NoError(t, err)
}
