package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	apperrs "backend/internal/errors"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockFollowRepository struct {
	createErr       error
	deleteErr       error
	existsErr       error
	getFollowingErr error
	getFollowerErr  error
	countErr        error
	isFollowingErr  error
	follows         map[uint]map[uint]*models.Follow
	countFollowing  map[uint]int64
	countFollowers  map[uint]int64
}

func newMockFollowRepository() *mockFollowRepository {
	return &mockFollowRepository{
		follows:        make(map[uint]map[uint]*models.Follow),
		countFollowing: make(map[uint]int64),
		countFollowers: make(map[uint]int64),
	}
}

func (m *mockFollowRepository) Create(ctx context.Context, follow *models.Follow) error {
	if m.createErr != nil {
		return m.createErr
	}
	if len(m.follows[follow.FollowerID]) == 0 {
		m.follows[follow.FollowerID] = make(map[uint]*models.Follow)
	}
	follow.CreatedAt = time.Now()
	m.follows[follow.FollowerID][follow.FollowingID] = follow
	m.countFollowing[follow.FollowerID]++
	m.countFollowers[follow.FollowingID]++
	return nil
}

func (m *mockFollowRepository) Delete(ctx context.Context, followerID, followingID uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.follows[followerID]; ok {
		if _, ok := m.follows[followerID][followingID]; ok {
			delete(m.follows[followerID], followingID)
			m.countFollowing[followerID]--
			m.countFollowers[followingID]--
			return nil
		}
	}
	return apperrs.ErrFollowNotFound
}

func (m *mockFollowRepository) Exists(ctx context.Context, followerID, followingID uint) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	if _, ok := m.follows[followerID]; ok {
		if _, ok := m.follows[followerID][followingID]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockFollowRepository) GetFollowingList(ctx context.Context, userID uint, orderBy string) ([]models.Follow, error) {
	if m.getFollowingErr != nil {
		return nil, m.getFollowingErr
	}
	var result []models.Follow
	if follows, ok := m.follows[userID]; ok {
		for _, f := range follows {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (m *mockFollowRepository) GetFollowerList(ctx context.Context, userID uint, orderBy string) ([]models.Follow, error) {
	if m.getFollowerErr != nil {
		return nil, m.getFollowerErr
	}
	var result []models.Follow
	for _, followers := range m.follows {
		for _, f := range followers {
			if f.FollowingID == userID {
				result = append(result, *f)
			}
		}
	}
	return result, nil
}

func (m *mockFollowRepository) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.countFollowing[userID], nil
}

func (m *mockFollowRepository) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return m.countFollowers[userID], nil
}

func (m *mockFollowRepository) IsFollowing(ctx context.Context, followerID, followingID uint) (bool, error) {
	if m.isFollowingErr != nil {
		return false, m.isFollowingErr
	}
	return m.Exists(ctx, followerID, followingID)
}

func (m *mockFollowRepository) WithTransaction(tx *gorm.DB) repository.FollowRepository {
	return m
}

type mockUserRepository struct {
	getByIDErr error
	users      map[uint]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[uint]*models.User),
	}
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, apperrs.ErrUserNotFound
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepository) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepository) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepository) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) List(ctx context.Context, orderBy string) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return nil
}

func (m *mockUserRepository) UpdateEmail(ctx context.Context, id uint, email string) error {
	return nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepository) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepository) Upsert(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepository) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func (m *mockUserRepository) addTestUser(id uint) {
	m.users[id] = &models.User{
		Username: "user" + string(rune('0'+id)),
		Email:    "user" + string(rune('0'+id)) + "@example.com",
	}
}

func TestFollowService_Create_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	resp, err := svc.Create(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.FollowerID)
	assert.Equal(t, uint(2), resp.FollowingID)
}

func TestFollowService_Create_SelfFollow(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	resp, err := svc.Create(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "不能关注自己")
}

func TestFollowService_Create_UserNotFound(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	resp, err := svc.Create(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "用户不存在")
}

func TestFollowService_Create_AlreadyFollowing(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 1, 2)
	resp, err := svc.Create(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "已经关注过该用户")
}

func TestFollowService_Create_RepositoryError(t *testing.T) {
	followRepo := newMockFollowRepository()
	followRepo.createErr = errors.New("database error")
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	resp, err := svc.Create(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestFollowService_Delete_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 1, 2)
	err := svc.Delete(context.Background(), 1, 2)

	assert.NoError(t, err)
}

func TestFollowService_Delete_NotFollowing(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	err := svc.Delete(context.Background(), 1, 2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未关注该用户")
}

func TestFollowService_IsFollowing_True(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 1, 2)
	isFollowing, err := svc.IsFollowing(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.True(t, isFollowing)
}

func TestFollowService_IsFollowing_False(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	isFollowing, err := svc.IsFollowing(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.False(t, isFollowing)
}

func TestFollowService_GetFollowingList_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)
	userRepo.addTestUser(3)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 1, 2)
	_, _ = svc.Create(context.Background(), 1, 3)

	list, err := svc.GetFollowingList(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFollowService_GetFollowerList_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)
	userRepo.addTestUser(3)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 2, 1)
	_, _ = svc.Create(context.Background(), 3, 1)

	list, err := svc.GetFollowerList(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestFollowService_GetFollowingCount_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)
	userRepo.addTestUser(3)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 1, 2)
	_, _ = svc.Create(context.Background(), 1, 3)

	count, err := svc.GetFollowingCount(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestFollowService_GetFollowerCount_Success(t *testing.T) {
	followRepo := newMockFollowRepository()
	userRepo := newMockUserRepository()
	userRepo.addTestUser(1)
	userRepo.addTestUser(2)
	userRepo.addTestUser(3)

	logger := slog.Default()
	svc := NewFollowService(followRepo, userRepo, logger)

	_, _ = svc.Create(context.Background(), 2, 1)
	_, _ = svc.Create(context.Background(), 3, 1)

	count, err := svc.GetFollowerCount(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
