package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestJWTEnv() func() {
	oldAccess := os.Getenv("JWT_ACCESS_EXPIRES")
	oldRefresh := os.Getenv("JWT_REFRESH_EXPIRES")
	os.Setenv("JWT_ACCESS_EXPIRES", "900")
	os.Setenv("JWT_REFRESH_EXPIRES", "604800")
	return func() {
		if oldAccess != "" {
			os.Setenv("JWT_ACCESS_EXPIRES", oldAccess)
		}
		if oldRefresh != "" {
			os.Setenv("JWT_REFRESH_EXPIRES", oldRefresh)
		}
	}
}

func TestJWTAuth_Success(t *testing.T) {
	restore := setupTestJWTEnv()
	defer restore()

	e := echo.New()
	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var capturedUserID uint
	var capturedEmail string
	var capturedUserType string
	handler := func(c *echo.Context) error {
		capturedUserID = GetUserID(c)
		capturedEmail = GetEmail(c)
		capturedUserType = GetUserType(c)
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uint(1), capturedUserID)
	assert.Equal(t, "test@example.com", capturedEmail)
	assert.Equal(t, "user", capturedUserType)
}

func TestJWTAuth_MissingAuthHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_InvalidFormat_NotBearer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	restore := setupTestJWTEnv()
	os.Setenv("JWT_ACCESS_EXPIRES", "1")
	defer restore()

	e := echo.New()
	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_RefreshTokenInsteadOfAccessToken(t *testing.T) {
	restore := setupTestJWTEnv()
	defer restore()

	e := echo.New()
	refreshToken, err := utils.GenerateRefreshToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_EmptyToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_CaseInsensitiveBearer(t *testing.T) {
	restore := setupTestJWTEnv()
	defer restore()

	e := echo.New()
	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "BEARER "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestJWTAuth_EmptyAuthorizationHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_Success(t *testing.T) {
	restore := setupTestJWTEnv()
	defer restore()

	e := echo.New()
	token, err := utils.GenerateAccessToken(1, "test@example.com", "admin")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "admin")

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireRole("admin")
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "master")

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireRole("admin", "master", "institution")
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestRequireRole_NoRoleMatch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "user")

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireRole("admin", "master")
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRequireRole_EmptyUserType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireRole("admin")
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireRole_NoUserTypeSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "")

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireRole("admin")
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAdmin_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "admin")

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireAdmin()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}

func TestRequireAdmin_Failure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(ContextKeyUserType, "user")

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireAdmin()
	h := middleware(handler)

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestGetUserID_WithValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyUserID, uint(42))

	id := GetUserID(c)
	assert.Equal(t, uint(42), id)
}

func TestGetUserID_NoValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)

	id := GetUserID(c)
	assert.Equal(t, uint(0), id)
}

func TestGetUserID_WrongType(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyUserID, "not a uint")

	id := GetUserID(c)
	assert.Equal(t, uint(0), id)
}

func TestGetEmail_WithValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyEmail, "test@example.com")

	email := GetEmail(c)
	assert.Equal(t, "test@example.com", email)
}

func TestGetEmail_NoValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)

	email := GetEmail(c)
	assert.Equal(t, "", email)
}

func TestGetUserType_WithValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyUserType, "admin")

	userType := GetUserType(c)
	assert.Equal(t, "admin", userType)
}

func TestGetUserType_NoValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)

	userType := GetUserType(c)
	assert.Equal(t, "", userType)
}

func TestIsAdmin_True(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyUserType, "admin")

	isAdmin := IsAdmin(c)
	assert.True(t, isAdmin)
}

func TestIsAdmin_False(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)
	c.Set(ContextKeyUserType, "user")

	isAdmin := IsAdmin(c)
	assert.False(t, isAdmin)
}

func TestIsAdmin_NoValue(t *testing.T) {
	e := echo.New()
	c := e.NewContext(nil, nil)

	isAdmin := IsAdmin(c)
	assert.False(t, isAdmin)
}

func TestRequireAuth_AliasForJWTAuth(t *testing.T) {
	restore := setupTestJWTEnv()
	defer restore()

	e := echo.New()
	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "OK")
	}

	middleware := RequireAuth()
	h := middleware(handler)

	err = h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled)
}
