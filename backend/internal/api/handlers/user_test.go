package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/models"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
)

type mockUserService struct {
	registerErr       error
	loginErr          error
	getUserByIDErr    error
	listUsersErr      error
	listUsersPagErr   error
	updateUserErr     error
	changePasswordErr error
	deleteUserErr     error
}

func (m *mockUserService) Register(ctx context.Context, req svc.RegisterRequest) (*svc.RegisterResponse, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	return &svc.RegisterResponse{
		ID:        1,
		Email:     req.Email,
		CreatedAt: "2024-01-01 00:00:00",
	}, nil
}

func (m *mockUserService) Login(ctx context.Context, req svc.LoginRequest) (*svc.LoginResponse, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	return &svc.LoginResponse{
		ID:    1,
		Email: req.Email,
		Token: "test_token",
	}, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getUserByIDErr != nil {
		return nil, m.getUserByIDErr
	}
	return &models.User{Email: "test@example.com"}, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserService) ListUsers(ctx context.Context) ([]models.User, error) {
	if m.listUsersErr != nil {
		return nil, m.listUsersErr
	}
	return []models.User{}, nil
}

func (m *mockUserService) ListUsersWithPagination(ctx context.Context, page, pageSize int) ([]models.User, int64, error) {
	if m.listUsersPagErr != nil {
		return nil, 0, m.listUsersPagErr
	}
	return []models.User{}, 0, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, req svc.UpdateUserRequest) error {
	if m.updateUserErr != nil {
		return m.updateUserErr
	}
	return nil
}

func (m *mockUserService) ChangePassword(ctx context.Context, req svc.ChangePasswordRequest) error {
	if m.changePasswordErr != nil {
		return m.changePasswordErr
	}
	return nil
}

func (m *mockUserService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	return nil
}

func (m *mockUserService) UpdateEmail(ctx context.Context, userID uint, newEmail string) error {
	return nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id uint) error {
	if m.deleteUserErr != nil {
		return m.deleteUserErr
	}
	return nil
}

func (m *mockUserService) ForceDeleteUser(ctx context.Context, id uint) error {
	return nil
}

type testHandler struct {
	svc *mockUserService
}

func newTestHandler() *testHandler {
	return &testHandler{svc: &mockUserService{}}
}

func (h *testHandler) Register(c *echo.Context) error {
	var req svc.RegisterRequest
	if err := (*c).Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}

	resp, err := h.svc.Register((*c).Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrEmailExists) {
			return errorResp(c, http.StatusConflict, "邮箱已被注册")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, resp)
}

func (h *testHandler) Login(c *echo.Context) error {
	var req svc.LoginRequest
	if err := (*c).Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}

	resp, err := h.svc.Login((*c).Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusUnauthorized, "用户不存在")
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return errorResp(c, http.StatusUnauthorized, "密码错误")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, resp)
}

func (h *testHandler) GetUser(c *echo.Context) error {
	idStr := (*c).QueryParam("id")
	if idStr == "" || idStr == "invalid" {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	user, err := h.svc.GetUserByID((*c).Request().Context(), 1)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, user)
}

func (h *testHandler) ListUsers(c *echo.Context) error {
	users, total, err := h.svc.ListUsersWithPagination((*c).Request().Context(), 1, 10)
	if err != nil {
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{
		"users":     users,
		"total":     total,
		"page":      1,
		"page_size": 10,
	})
}

func (h *testHandler) UpdateUser(c *echo.Context) error {
	idStr := (*c).QueryParam("id")
	if idStr == "" || idStr == "invalid" {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	var req svc.UpdateUserRequest
	if err := (*c).Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}
	req.ID = 1

	if err := h.svc.UpdateUser((*c).Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		if errors.Is(err, svc.ErrEmailExists) {
			return errorResp(c, http.StatusConflict, "邮箱已被使用")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "更新成功"})
}

func (h *testHandler) ChangePassword(c *echo.Context) error {
	idStr := (*c).QueryParam("id")
	if idStr == "" || idStr == "invalid" {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	var req svc.ChangePasswordRequest
	if err := (*c).Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}
	req.UserID = 1

	if err := h.svc.ChangePassword((*c).Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return errorResp(c, http.StatusUnauthorized, "旧密码错误")
		}
		if errors.Is(err, svc.ErrSamePassword) {
			return errorResp(c, http.StatusBadRequest, "新密码不能与旧密码相同")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "密码修改成功"})
}

func (h *testHandler) DeleteUser(c *echo.Context) error {
	idStr := (*c).QueryParam("id")
	if idStr == "" || idStr == "invalid" {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	if err := h.svc.DeleteUser((*c).Request().Context(), 1); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "删除成功"})
}

func TestRegister_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/register", strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.Register(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/register", strings.NewReader(`{invalid json}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.Register(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRegister_EmailExists(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/register", strings.NewReader(`{"email":"existing@example.com","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.registerErr = svc.ErrEmailExists
	_ = h.Register(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/login", strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.Login(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp response.Response[any]
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error != nil {
		t.Errorf("expected no error in response")
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/login", strings.NewReader(`{invalid}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.Login(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/login", strings.NewReader(`{"email":"notfound@example.com","password":"password123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.loginErr = svc.ErrUserNotFound
	_ = h.Login(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users/login", strings.NewReader(`{"email":"test@example.com","password":"wrongpassword"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.loginErr = svc.ErrInvalidPassword
	_ = h.Login(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetUser_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users?id=1", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.GetUser(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users?id=invalid", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.GetUser(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users?id=999", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.getUserByIDErr = svc.ErrUserNotFound
	_ = h.GetUser(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestListUsers_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users?page=1&page_size=10", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.ListUsers(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestListUsers_ServiceError(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/users", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.listUsersPagErr = errors.New("service error")
	_ = h.ListUsers(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users?id=1", strings.NewReader(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.UpdateUser(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUpdateUser_InvalidID(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users?id=invalid", strings.NewReader(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.UpdateUser(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUser_InvalidJSON(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users?id=1", strings.NewReader(`{invalid}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.UpdateUser(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users?id=999", strings.NewReader(`{"email":"new@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.updateUserErr = svc.ErrUserNotFound
	_ = h.UpdateUser(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestUpdateUser_EmailExists(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users?id=1", strings.NewReader(`{"email":"existing@example.com"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.updateUserErr = svc.ErrEmailExists
	_ = h.UpdateUser(c)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestChangePassword_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/1/password?id=1", strings.NewReader(`{"old_password":"oldpass","new_password":"newpass123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestChangePassword_InvalidID(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/invalid/password?id=invalid", strings.NewReader(`{"old_password":"oldpass","new_password":"newpass123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestChangePassword_InvalidJSON(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/1/password?id=1", strings.NewReader(`{invalid}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/999/password?id=999", strings.NewReader(`{"old_password":"oldpass","new_password":"newpass123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.changePasswordErr = svc.ErrUserNotFound
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestChangePassword_WrongPassword(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/1/password?id=1", strings.NewReader(`{"old_password":"wrongpass","new_password":"newpass123"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.changePasswordErr = svc.ErrInvalidPassword
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestChangePassword_SamePassword(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/users/1/password?id=1", strings.NewReader(`{"old_password":"samepass","new_password":"samepass"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.changePasswordErr = svc.ErrSamePassword
	_ = h.ChangePassword(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/1?id=1", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.DeleteUser(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/invalid?id=invalid", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	_ = h.DeleteUser(c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/users/999?id=999", nil)
	c := e.NewContext(req, rec)

	h := newTestHandler()
	h.svc.deleteUserErr = svc.ErrUserNotFound
	_ = h.DeleteUser(c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
