package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func setupAuthTestEnv() func() {
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

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	restore := setupAuthTestEnv()
	defer restore()

	h := NewAuthHandler()

	token, err := utils.GenerateRefreshToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	body := `{"refresh_token":"` + token + `"}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
	assert.NotEmpty(t, data["expires_in"])
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	h := NewAuthHandler()

	body := `{invalid}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_RefreshToken_EmptyToken(t *testing.T) {
	h := NewAuthHandler()

	body := `{"refresh_token":""}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	h := NewAuthHandler()

	body := `{"refresh_token":"invalid.token.here"}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_RefreshToken_ExpiredToken(t *testing.T) {
	os.Setenv("JWT_REFRESH_EXPIRES", "1")
	defer func() {
		os.Setenv("JWT_REFRESH_EXPIRES", "604800")
	}()

	h := NewAuthHandler()

	token, err := utils.GenerateRefreshToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)

	body := `{"refresh_token":"` + token + `"}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_RefreshToken_WrongTokenType(t *testing.T) {
	restore := setupAuthTestEnv()
	defer restore()

	h := NewAuthHandler()

	token, err := utils.GenerateAccessToken(1, "test@example.com", "user")
	assert.NoError(t, err)

	body := `{"refresh_token":"` + token + `"}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_RefreshToken_MissingRefreshTokenField(t *testing.T) {
	h := NewAuthHandler()

	body := `{}`
	e := echo.New()
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.RefreshToken(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
