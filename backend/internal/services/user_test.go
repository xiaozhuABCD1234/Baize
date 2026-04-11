package service

import (
	"context"
	"errors"
	"testing"

	"backend/internal/models"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	users          map[uint]*models.User
	usersByEmail   map[string]*models.User
	createErr      error
	getByIDErr     error
	getByEmailErr  error
	listErr        error
	listPagErr     error
	updateErr      error
	updatePassErr  error
	updateEmailErr error
	deleteErr      error
	forceDeleteErr error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:        make(map[uint]*models.User),
		usersByEmail: make(map[string]*models.User),
	}
}

func createTestUser(id uint, email, password string) *models.User {
	user := &models.User{Email: email, Password: password}
	user.ID = id
	return user
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if len(m.users) == 0 {
		user.ID = 1
	} else {
		maxID := uint(0)
		for id := range m.users {
			if id > maxID {
				maxID = id
			}
		}
		user.ID = maxID + 1
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	if user, ok := m.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepository) List(ctx context.Context) ([]models.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *mockUserRepository) ListWithPagination(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	if m.listPagErr != nil {
		return nil, 0, m.listPagErr
	}
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	total := int64(len(users))
	return users, total, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return repository.ErrUserNotFound
	}
	oldEmail := m.users[user.ID].Email
	delete(m.usersByEmail, oldEmail)
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	if m.updatePassErr != nil {
		return m.updatePassErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	m.users[id].Password = hashedPassword
	return nil
}

func (m *mockUserRepository) UpdateEmail(ctx context.Context, id uint, email string) error {
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

func (m *mockUserRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) ForceDelete(ctx context.Context, id uint) error {
	if m.forceDeleteErr != nil {
		return m.forceDeleteErr
	}
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	email := m.users[id].Email
	delete(m.users, id)
	delete(m.usersByEmail, email)
	return nil
}

func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name      string
		req       RegisterRequest
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name: "success",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserRepository) {},
			wantErr:   nil,
		},
		{
			name: "email exists",
			req: RegisterRequest{
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserRepository) {
				user := &models.User{Email: "existing@example.com", Password: "hash"}
				user.ID = 1
				m.users[1] = user
				m.usersByEmail["existing@example.com"] = user
			},
			wantErr: ErrEmailExists,
		},
		{
			name: "create error",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserRepository) {
				m.createErr = errors.New("database error")
			},
			wantErr: errors.New("创建用户失败: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			resp, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Register() expected error %v, got nil", tt.wantErr)
					return
				}
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("Register() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Register() unexpected error: %v", err)
				return
			}

			if resp.Email != tt.req.Email {
				t.Errorf("Register() email = %v, want %v", resp.Email, tt.req.Email)
			}
			if resp.ID == 0 {
				t.Error("Register() returned zero ID")
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMye.IzQGPQvqF8gPZJ6VZrN6dVJ6Z6V6Z6"

	tests := []struct {
		name      string
		req       LoginRequest
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name: "user not found",
			req: LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
		{
			name: "invalid password",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", hashedPassword)
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			_, err := svc.Login(context.Background(), tt.req)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Login() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", "hashed")
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name:      "user not found",
			userID:    999,
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			user, err := svc.GetUserByID(context.Background(), tt.userID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetUserByID() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetUserByID() unexpected error: %v", err)
				return
			}

			if user.Password != "" {
				t.Error("GetUserByID() should clear password")
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mockUserRepository)
		wantErr   error
		wantCount int
	}{
		{
			name: "success",
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "a@example.com", "hash1")
				m.users[2] = createTestUser(2, "b@example.com", "hash2")
				m.usersByEmail["a@example.com"] = m.users[1]
				m.usersByEmail["b@example.com"] = m.users[2]
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:      "empty list",
			setupMock: func(m *mockUserRepository) {},
			wantErr:   nil,
			wantCount: 0,
		},
		{
			name: "list error",
			setupMock: func(m *mockUserRepository) {
				m.listErr = errors.New("database error")
			},
			wantErr:   errors.New("查询用户列表失败: database error"),
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			users, err := svc.ListUsers(context.Background())

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("ListUsers() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ListUsers() unexpected error: %v", err)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("ListUsers() count = %d, want %d", len(users), tt.wantCount)
			}

			for _, user := range users {
				if user.Password != "" {
					t.Error("ListUsers() should clear all passwords")
				}
			}
		})
	}
}

func TestUserService_ListUsersWithPagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		pageSize  int
		setupMock func(*mockUserRepository)
		wantTotal int64
		wantPage  int
		wantSize  int
	}{
		{
			name:     "default values",
			page:     0,
			pageSize: 0,
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "a@example.com", "hash")
				m.users[2] = createTestUser(2, "b@example.com", "hash")
				m.usersByEmail["a@example.com"] = m.users[1]
				m.usersByEmail["b@example.com"] = m.users[2]
			},
			wantTotal: 2,
			wantPage:  1,
			wantSize:  10,
		},
		{
			name:     "custom values",
			page:     2,
			pageSize: 5,
			setupMock: func(m *mockUserRepository) {
				for i := 1; i <= 10; i++ {
					m.users[uint(i)] = createTestUser(uint(i), "a@example.com", "hash")
				}
			},
			wantTotal: 10,
			wantPage:  2,
			wantSize:  5,
		},
		{
			name:     "pageSize too large",
			page:     1,
			pageSize: 200,
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "a@example.com", "hash")
			},
			wantTotal: 1,
			wantPage:  1,
			wantSize:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			_, total, err := svc.ListUsersWithPagination(context.Background(), tt.page, tt.pageSize)

			if err != nil {
				t.Errorf("ListUsersWithPagination() unexpected error: %v", err)
				return
			}

			if total != tt.wantTotal {
				t.Errorf("ListUsersWithPagination() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		req       UpdateUserRequest
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name: "success",
			req: UpdateUserRequest{
				ID:    1,
				Email: "new@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "old@example.com", "hash")
				m.usersByEmail["old@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			req: UpdateUserRequest{
				ID:    999,
				Email: "new@example.com",
			},
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
		{
			name: "email already exists",
			req: UpdateUserRequest{
				ID:    1,
				Email: "existing@example.com",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "old@example.com", "hash")
				m.users[2] = createTestUser(2, "existing@example.com", "hash")
				m.usersByEmail["old@example.com"] = m.users[1]
				m.usersByEmail["existing@example.com"] = m.users[2]
			},
			wantErr: ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.UpdateUser(context.Background(), tt.req)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	hashedSamePassword, _ := bcrypt.GenerateFromPassword([]byte("samepassword"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		req       ChangePasswordRequest
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name: "success",
			req: ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", string(hashedOldPassword))
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			req: ChangePasswordRequest{
				UserID:      999,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
		{
			name: "wrong old password",
			req: ChangePasswordRequest{
				UserID:      1,
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", string(hashedOldPassword))
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: ErrInvalidPassword,
		},
		{
			name: "same password",
			req: ChangePasswordRequest{
				UserID:      1,
				OldPassword: "samepassword",
				NewPassword: "samepassword",
			},
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", string(hashedSamePassword))
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: ErrSamePassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.ChangePassword(context.Background(), tt.req)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ChangePassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		newPassword string
		setupMock   func(*mockUserRepository)
		wantErr     error
	}{
		{
			name:        "success",
			userID:      1,
			newPassword: "newpassword123",
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", "oldhash")
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name:        "user not found",
			userID:      999,
			newPassword: "newpassword123",
			setupMock:   func(m *mockUserRepository) {},
			wantErr:     ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.ResetPassword(context.Background(), tt.userID, tt.newPassword)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ResetPassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_UpdateEmail(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		newEmail  string
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name:     "success",
			userID:   1,
			newEmail: "new@example.com",
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "old@example.com", "hash")
				m.usersByEmail["old@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name:      "user not found",
			userID:    999,
			newEmail:  "new@example.com",
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
		{
			name:     "same email",
			userID:   1,
			newEmail: "old@example.com",
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "old@example.com", "hash")
				m.usersByEmail["old@example.com"] = m.users[1]
			},
			wantErr: ErrEmailNotChanged,
		},
		{
			name:     "email already exists",
			userID:   1,
			newEmail: "existing@example.com",
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "old@example.com", "hash")
				m.users[2] = createTestUser(2, "existing@example.com", "hash")
				m.usersByEmail["old@example.com"] = m.users[1]
				m.usersByEmail["existing@example.com"] = m.users[2]
			},
			wantErr: ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.UpdateEmail(context.Background(), tt.userID, tt.newEmail)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateEmail() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", "hash")
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name:      "user not found",
			userID:    999,
			setupMock: func(m *mockUserRepository) {},
			wantErr:   ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.DeleteUser(context.Background(), tt.userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("DeleteUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_ForceDeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		setupMock func(*mockUserRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(m *mockUserRepository) {
				m.users[1] = createTestUser(1, "test@example.com", "hash")
				m.usersByEmail["test@example.com"] = m.users[1]
			},
			wantErr: nil,
		},
		{
			name:      "user not found",
			userID:    999,
			setupMock: func(m *mockUserRepository) {},
			wantErr:   repository.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepository()
			tt.setupMock(repo)
			svc := NewUserService(repo)

			err := svc.ForceDeleteUser(context.Background(), tt.userID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ForceDeleteUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
