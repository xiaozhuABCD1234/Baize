package service

import (
	"context"
	"testing"

	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockUserRepositoryForUser struct {
	createErr        error
	getByIDErr       error
	getByEmailErr    error
	getByUsernameErr error
	listErr          error
	listPagErr       error
	updateErr        error
	updatePassErr    error
	updateEmailErr   error
	deleteErr        error
	forceDeleteErr   error
	users            map[uint]*models.User
	usersByEmail     map[string]*models.User
	usersByUsername  map[string]*models.User
	nextID           uint
}

func newMockUserRepositoryForUser() *mockUserRepositoryForUser {
	return &mockUserRepositoryForUser{
		users:           make(map[uint]*models.User),
		usersByEmail:    make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		nextID:          1,
	}
}

func (m *mockUserRepositoryForUser) Create(ctx context.Context, user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	if user.Username != "" {
		m.usersByUsername[user.Username] = user
	}
	return nil
}

func (m *mockUserRepositoryForUser) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		user.ID = id
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepositoryForUser) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	if user, ok := m.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepositoryForUser) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameErr != nil {
		return nil, m.getByUsernameErr
	}
	if user, ok := m.usersByUsername[username]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepositoryForUser) List(ctx context.Context, orderBy string) ([]models.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.User
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, nil
}

func (m *mockUserRepositoryForUser) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
	if m.listPagErr != nil {
		return nil, 0, m.listPagErr
	}
	var result []models.User
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, int64(len(result)), nil
}

func (m *mockUserRepositoryForUser) Update(ctx context.Context, user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return repository.ErrUserNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepositoryForUser) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	if m.updatePassErr != nil {
		return m.updatePassErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	m.users[id].Password = hashedPassword
	return nil
}

func (m *mockUserRepositoryForUser) UpdateEmail(ctx context.Context, id uint, email string) error {
	if m.updateEmailErr != nil {
		return m.updateEmailErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	oldEmail := m.users[id].Email
	delete(m.usersByEmail, oldEmail)
	m.users[id].Email = email
	m.usersByEmail[email] = m.users[id]
	return nil
}

func (m *mockUserRepositoryForUser) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepositoryForUser) ForceDelete(ctx context.Context, id uint) error {
	if m.forceDeleteErr != nil {
		return m.forceDeleteErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepositoryForUser) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepositoryForUser) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForUser) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForUser) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForUser) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForUser) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepositoryForUser) Upsert(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForUser) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func (m *mockUserRepositoryForUser) addTestUser(id uint, email, username, password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
		UserType: models.UserTypeUser,
		Status:   models.UserStatusActive,
	}
	m.users[id] = user
	m.usersByEmail[email] = user
	m.usersByUsername[username] = user
}

func TestUserService_Register_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	req := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	resp, err := svc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "test@example.com", resp.Email)
}

func TestUserService_Register_DuplicateUsername(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "existing@example.com", "existinguser", "password")
	svc := NewUserService(repo)

	req := RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
	}
	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUsernameExists, err)
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "existing@example.com", "existinguser", "password")
	svc := NewUserService(repo)

	req := RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrEmailExists, err)
}

func TestUserService_Register_InvalidRole(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	req := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		UserType: "invalid_role",
	}
	resp, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidRole, err)
}

func TestUserService_Login_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "password123")
	svc := NewUserService(repo)

	req := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	resp, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "testuser", resp.Username)
	assert.NotEmpty(t, resp.Token)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	req := LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}
	resp, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "correctpassword")
	svc := NewUserService(repo)

	req := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	resp, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidPassword, err)
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "password123")
	svc := NewUserService(repo)

	user, err := svc.GetUserByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Empty(t, user.Password)
}

func TestUserService_GetUserByID_NotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	user, err := svc.GetUserByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestUserService_ListUsers_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "user1@example.com", "user1", "password")
	repo.addTestUser(2, "user2@example.com", "user2", "password")
	svc := NewUserService(repo)

	users, err := svc.ListUsers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserService_ListUsersWithPagination_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "user1@example.com", "user1", "password")
	repo.addTestUser(2, "user2@example.com", "user2", "password")
	svc := NewUserService(repo)

	users, total, err := svc.ListUsersWithPagination(context.Background(), 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "oldpassword")
	svc := NewUserService(repo)

	req := ChangePasswordRequest{
		UserID:      1,
		OldPassword: "oldpassword",
		NewPassword: "newpassword123",
	}
	err := svc.ChangePassword(context.Background(), req)

	assert.NoError(t, err)
}

func TestUserService_ChangePassword_WrongPassword(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "correctpassword")
	svc := NewUserService(repo)

	req := ChangePasswordRequest{
		UserID:      1,
		OldPassword: "wrongpassword",
		NewPassword: "newpassword123",
	}
	err := svc.ChangePassword(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
}

func TestUserService_ChangePassword_SamePassword(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "samepassword")
	svc := NewUserService(repo)

	req := ChangePasswordRequest{
		UserID:      1,
		OldPassword: "samepassword",
		NewPassword: "samepassword",
	}
	err := svc.ChangePassword(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrSamePassword, err)
}

func TestUserService_UpdateEmail_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "old@example.com", "testuser", "password")
	svc := NewUserService(repo)

	err := svc.UpdateEmail(context.Background(), 1, "new@example.com")

	assert.NoError(t, err)
}

func TestUserService_UpdateEmail_SameEmail(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "same@example.com", "testuser", "password")
	svc := NewUserService(repo)

	err := svc.UpdateEmail(context.Background(), 1, "same@example.com")

	assert.Error(t, err)
	assert.Equal(t, ErrEmailNotChanged, err)
}

func TestUserService_UpdateEmail_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "user1@example.com", "user1", "password")
	repo.addTestUser(2, "user2@example.com", "user2", "password")
	svc := NewUserService(repo)

	err := svc.UpdateEmail(context.Background(), 1, "user2@example.com")

	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "password")
	svc := NewUserService(repo)

	err := svc.DeleteUser(context.Background(), 1)

	assert.NoError(t, err)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	err := svc.DeleteUser(context.Background(), 999)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestUserService_ForceDeleteUser_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "password")
	svc := NewUserService(repo)

	err := svc.ForceDeleteUser(context.Background(), 1)

	assert.NoError(t, err)
}

func TestUserService_ForceDeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	err := svc.ForceDeleteUser(context.Background(), 999)

	assert.Error(t, err)
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "old@example.com", "testuser", "password")
	svc := NewUserService(repo)

	req := UpdateUserRequest{
		ID:    1,
		Email: "new@example.com",
	}
	err := svc.UpdateUser(context.Background(), req)

	assert.NoError(t, err)
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	req := UpdateUserRequest{
		ID:    999,
		Email: "new@example.com",
	}
	err := svc.UpdateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestUserService_UpdateUser_InvalidRole(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "password")
	svc := NewUserService(repo)

	req := UpdateUserRequest{
		ID:       1,
		UserType: "invalid",
	}
	err := svc.UpdateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRole, err)
}

func TestUserService_ResetPassword_Success(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	repo.addTestUser(1, "test@example.com", "testuser", "oldpassword")
	svc := NewUserService(repo)

	err := svc.ResetPassword(context.Background(), 1, "newpassword123")

	assert.NoError(t, err)
}

func TestUserService_ResetPassword_NotFound(t *testing.T) {
	repo := newMockUserRepositoryForUser()
	svc := NewUserService(repo)

	err := svc.ResetPassword(context.Background(), 999, "newpassword123")

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}
