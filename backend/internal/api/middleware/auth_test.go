package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

func TestJWTAuth_Success(t *testing.T) {
	e := echo.New()

	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		userID := c.Get("user_id")
		if userID != uint(1) {
			t.Errorf("user_id = %v, want %d", userID, 1)
		}
		email := c.Get("email")
		if email != "test@example.com" {
			t.Errorf("email = %v, want %s", email, "test@example.com")
		}
		return nil
	}

	middleware := JWTAuth()
	h := middleware(handler)

	err = h(c)
	if err != nil {
		t.Errorf("JWTAuth() error = %v, want nil", err)
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
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

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
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

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
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

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	e := echo.New()

	refreshToken, err := utils.GenerateRefreshToken(1, "test@example.com", "user")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuth_RefreshTokenInsteadofAccessToken(t *testing.T) {
	e := echo.New()

	refreshToken, err := utils.GenerateRefreshToken(1, "test@example.com", "user")
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	middleware := JWTAuth()
	h := middleware(handler)

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
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

	_ = h(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuth_CaseInsensitiveBearer(t *testing.T) {
	e := echo.New()

	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "BEARER "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c *echo.Context) error {
		handlerCalled = true
		return nil
	}

	middleware := JWTAuth()
	h := middleware(handler)

	_ = h(c)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}
