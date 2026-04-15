package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/internal/api/middleware"
	apperrs "backend/internal/errors"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockUserRepository struct {
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
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:           make(map[uint]*models.User),
		usersByEmail:    make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
	}
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
	if user.Username != "" {
		m.usersByUsername[user.Username] = user
	}
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

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameErr != nil {
		return nil, m.getByUsernameErr
	}
	if user, ok := m.usersByUsername[username]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *mockUserRepository) List(ctx context.Context, orderBy string) ([]models.User, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	users := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func (m *mockUserRepository) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
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
		return apperrs.ErrUserNotFound
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
		return apperrs.ErrUserNotFound
	}
	m.users[id].Password = hashedPassword
	return nil
}

func (m *mockUserRepository) UpdateEmail(ctx context.Context, id uint, email string) error {
	if m.updateEmailErr != nil {
		return m.updateEmailErr
	}
	if _, ok := m.users[id]; !ok {
		return apperrs.ErrUserNotFound
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
		return apperrs.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) ForceDelete(ctx context.Context, id uint) error {
	if m.forceDeleteErr != nil {
		return m.forceDeleteErr
	}
	if _, ok := m.users[id]; !ok {
		return apperrs.ErrUserNotFound
	}
	email := m.users[id].Email
	username := m.users[id].Username
	delete(m.users, id)
	delete(m.usersByEmail, email)
	delete(m.usersByUsername, username)
	return nil
}

func (m *mockUserRepository) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return m.Create(ctx, user)
}

func (m *mockUserRepository) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepository) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepository) Upsert(ctx context.Context, user *models.User) error {
	return m.Create(ctx, user)
}

func (m *mockUserRepository) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func (m *mockUserRepository) addTestUser(id uint, email, username, password string, userType models.UserType) {
	user := &models.User{
		Email:    email,
		Username: username,
		Password: password,
		UserType: userType,
		Status:   models.UserStatusActive,
	}
	user.ID = id
	m.users[id] = user
	m.usersByEmail[email] = user
	m.usersByUsername[username] = user
}

type testEnv struct {
	restoreJWT func()
}

func setupTestEnv() *testEnv {
	oldAccess := os.Getenv("JWT_ACCESS_EXPIRES")
	oldRefresh := os.Getenv("JWT_REFRESH_EXPIRES")
	os.Setenv("JWT_ACCESS_EXPIRES", "900")
	os.Setenv("JWT_REFRESH_EXPIRES", "604800")
	return &testEnv{
		restoreJWT: func() {
			if oldAccess != "" {
				os.Setenv("JWT_ACCESS_EXPIRES", oldAccess)
			}
			if oldRefresh != "" {
				os.Setenv("JWT_REFRESH_EXPIRES", oldRefresh)
			}
		},
	}
}

func (e *testEnv) cleanup() {
	e.restoreJWT()
}

func createRealHandler(repo repository.UserRepository) *UserHandler {
	service := svc.NewUserService(repo)
	return NewUserHandler(service)
}

func setupEchoContext(method, path string, body string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func setupEchoContextWithParams(method, path string, body string, paramName, paramValue string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if paramName != "" && paramValue != "" {
		c.SetPath("/users/:" + paramName)
		c.SetPathValues([]echo.PathValue{{Name: paramName, Value: paramValue}})
	}

	return c, rec
}

func parseResponse(t *testing.T, rec *httptest.ResponseRecorder) response.Response {
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return resp
}

func assertSuccessResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) response.Response {
	assert.Equal(t, expectedStatus, rec.Code)
	resp := parseResponse(t, rec)
	assert.Nil(t, resp.Error, "expected no error in response")
	assert.NotNil(t, resp.Data, "expected data in response")
	return resp
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assert.Equal(t, expectedStatus, rec.Code)
	resp := parseResponse(t, rec)
	assert.NotNil(t, resp.Error, "expected error in response")
	assert.Equal(t, expectedCode, resp.Error.Code)
	assert.NotEmpty(t, resp.Error.Message)
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"username":"testuser","email":"test@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "testuser", data["username"])
}

func TestRegister_InvalidJSON(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{invalid json}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestRegister_EmailExists(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "existing@example.com", "existing", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"username":"newuser","email":"existing@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusConflict, response.UserEmailExists)
}

func TestRegister_UsernameExists(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "a@example.com", "existing", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"username":"existing","email":"new@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusConflict, response.UserNameExists)
}

func TestRegister_InvalidRole(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"username":"testuser","email":"test@example.com","password":"password123","user_type":"invalid_role"}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestRegister_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	repo.createErr = errors.New("database connection failed")
	h := createRealHandler(repo)

	body := `{"username":"testuser","email":"test@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/register", body)

	err := h.Register(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepository()
	hashedPassword, _ := bcryptHash("password123")
	repo.addTestUser(1, "test@example.com", "testuser", hashedPassword, models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"email":"test@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/login", body)

	err := h.Login(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["token"])
}

func TestLogin_InvalidJSON(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{invalid}`
	c, rec := setupEchoContext("POST", "/users/login", body)

	err := h.Login(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"email":"notfound@example.com","password":"password123"}`
	c, rec := setupEchoContext("POST", "/users/login", body)

	err := h.Login(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusUnauthorized, response.UserNotFound)
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := newMockUserRepository()
	hashedPassword, _ := bcryptHash("correctpassword")
	repo.addTestUser(1, "test@example.com", "testuser", hashedPassword, models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	c, rec := setupEchoContext("POST", "/users/login", body)

	err := h.Login(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusUnauthorized, response.InvalidPassword)
}

func TestGetUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "test@example.com", "testuser", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("GET", "/users/1", "", "id", "1")

	err := h.GetUser(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.NotNil(t, resp.Data)
}

func TestGetUser_InvalidID(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("GET", "/users/abc", "", "id", "abc")

	err := h.GetUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestGetUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("GET", "/users/999", "", "id", "999")

	err := h.GetUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestGetUser_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	repo.getByIDErr = errors.New("database error")
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("GET", "/users/1", "", "id", "1")

	err := h.GetUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestListUsers_Success(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "a@example.com", "user1", "hash", models.UserTypeUser)
	repo.addTestUser(2, "b@example.com", "user2", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	c, rec := setupEchoContext("GET", "/users?page=1&page_size=10", "")

	err := h.ListUsers(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, 1, resp.Page.PageNum)
	assert.Equal(t, 10, resp.Page.PageSize)
	assert.Equal(t, int64(2), resp.Page.Total)
}

func TestListUsers_EmptyList(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContext("GET", "/users?page=1", "")

	err := h.ListUsers(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, int64(0), resp.Page.Total)
}

func TestListUsers_CustomPagination(t *testing.T) {
	repo := newMockUserRepository()
	for i := 0; i < 10; i++ {
		repo.addTestUser(uint(i+1), "test@example.com", "testuser", "hash", models.UserTypeUser)
	}
	h := createRealHandler(repo)

	c, rec := setupEchoContext("GET", "/users?page=2&page_size=5", "")

	err := h.ListUsers(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 2, resp.Page.PageNum)
	assert.Equal(t, 5, resp.Page.PageSize)
}

func TestListUsers_DefaultPagination(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContext("GET", "/users?page=-1&page_size=0", "")

	err := h.ListUsers(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 1, resp.Page.PageNum)
	assert.Equal(t, 10, resp.Page.PageSize)
}

func TestListUsers_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	repo.listPagErr = errors.New("database error")
	h := createRealHandler(repo)

	c, rec := setupEchoContext("GET", "/users?page=1", "")

	err := h.ListUsers(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestUpdateUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "old@example.com", "testuser", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"email":"new@example.com"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "更新成功", data["message"])
}

func TestUpdateUser_InvalidID(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"email":"new@example.com"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/abc", body, "id", "abc")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestUpdateUser_InvalidJSON(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "old@example.com", "testuser", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{invalid}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestUpdateUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"email":"new@example.com"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/999", body, "id", "999")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "admin")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestUpdateUser_EmailExists(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "old@example.com", "user1", "hash", models.UserTypeUser)
	repo.addTestUser(2, "existing@example.com", "user2", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"email":"existing@example.com"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusConflict, response.UserEmailExists)
}

func TestUpdateUser_InvalidRole(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "test@example.com", "testuser", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"user_type":"invalid_role"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestUpdateUser_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "old@example.com", "testuser", "hash", models.UserTypeUser)
	repo.updateErr = errors.New("database error")
	h := createRealHandler(repo)

	body := `{"email":"new@example.com"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.UpdateUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestChangePassword_Success(t *testing.T) {
	repo := newMockUserRepository()
	hashedOld, _ := bcryptHash("oldpassword")
	repo.addTestUser(1, "test@example.com", "testuser", hashedOld, models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"old_password":"oldpassword","new_password":"newpassword123"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1/password", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "密码修改成功", data["message"])
}

func TestChangePassword_InvalidID(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"old_password":"oldpass","new_password":"newpass123"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/abc/password", body, "id", "abc")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{invalid}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1/password", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestChangePassword_UserNotFound(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	body := `{"old_password":"oldpass","new_password":"newpass123"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/999/password", body, "id", "999")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "admin")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestChangePassword_WrongPassword(t *testing.T) {
	repo := newMockUserRepository()
	hashedOld, _ := bcryptHash("correctpassword")
	repo.addTestUser(1, "test@example.com", "testuser", hashedOld, models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"old_password":"wrongpassword","new_password":"newpass123"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1/password", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusUnauthorized, response.InvalidPassword)
}

func TestChangePassword_SamePassword(t *testing.T) {
	repo := newMockUserRepository()
	hashedOld, _ := bcryptHash("samepassword")
	repo.addTestUser(1, "test@example.com", "testuser", hashedOld, models.UserTypeUser)
	h := createRealHandler(repo)

	body := `{"old_password":"samepassword","new_password":"samepassword"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1/password", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.SamePassword)
}

func TestChangePassword_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	hashedOld, _ := bcryptHash("oldpassword")
	repo.addTestUser(1, "test@example.com", "testuser", hashedOld, models.UserTypeUser)
	repo.updatePassErr = errors.New("database error")
	h := createRealHandler(repo)

	body := `{"old_password":"oldpassword","new_password":"newpass123"}`
	c, rec := setupEchoContextWithParams("PUT", "/users/1/password", body, "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.ChangePassword(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestDeleteUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "test@example.com", "testuser", "hash", models.UserTypeUser)
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("DELETE", "/users/1", "", "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.DeleteUser(c)
	assert.NoError(t, err)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "删除成功", data["message"])
}

func TestDeleteUser_InvalidID(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("DELETE", "/users/abc", "", "id", "abc")

	err := h.DeleteUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestDeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("DELETE", "/users/999", "", "id", "999")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "admin")

	err := h.DeleteUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestDeleteUser_ServiceError(t *testing.T) {
	repo := newMockUserRepository()
	repo.addTestUser(1, "test@example.com", "testuser", "hash", models.UserTypeUser)
	repo.deleteErr = errors.New("database error")
	h := createRealHandler(repo)

	c, rec := setupEchoContextWithParams("DELETE", "/users/1", "", "id", "1")
	c.Set(middleware.ContextKeyUserID, uint(1))
	c.Set(middleware.ContextKeyUserType, "user")

	err := h.DeleteUser(c)
	assert.NoError(t, err)

	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
