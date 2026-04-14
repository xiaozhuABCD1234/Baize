package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"backend/internal/api/handlers"
	"backend/internal/api/routes"
	"backend/internal/models"
	repo "backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type IntegrationSuite struct {
	suite.Suite
	DB          *gorm.DB
	Echo        *echo.Echo
	Repo        *repo.UserRepository
	Svc         *svc.UserService
	tmpDB       *os.File
	userCounter int64
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) SetupSuite() {
	tmpFile, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		s.T().Fatalf("Failed to create temp file: %v", err)
	}
	s.tmpDB = tmpFile
	tmpFile.Close()

	db, err := gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
	if err != nil {
		s.T().Fatalf("Failed to connect to SQLite: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		s.T().Fatalf("Failed to migrate schema: %v", err)
	}

	s.DB = db
	s.Repo = repo.NewUserRepository(db)
	s.Svc = svc.NewUserService(s.Repo)

	s.Echo = echo.New()
	userHandler := handlers.NewUserHandler(s.Svc)
	routes.SetupRouter(s.Echo, userHandler)
}

func (s *IntegrationSuite) SetupTest() {
	s.DB.Exec("DELETE FROM users")
}

func (s *IntegrationSuite) TearDownSuite() {
	sqlDB, err := s.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
	os.Remove(s.tmpDB.Name())
}

func (s *IntegrationSuite) CreateUserAndLogin(email, password string) (*models.User, string) {
	phoneNum := atomic.AddInt64(&s.userCounter, 1)
	registerReq := svc.RegisterRequest{
		Username: email,
		Email:    email,
		Password: password,
		Phone:    fmt.Sprintf("138%011d", phoneNum),
	}
	resp, err := s.Svc.Register(context.Background(), registerReq)
	if err != nil {
		s.T().Fatalf("Failed to create user: %v", err)
	}

	createdUser, err := s.Repo.GetByID(context.Background(), resp.ID)
	if err != nil {
		s.T().Fatalf("Failed to get created user: %v", err)
	}

	loginReq := svc.LoginRequest{
		Email:    email,
		Password: password,
	}
	loginResp, err := s.Svc.Login(context.Background(), loginReq)
	if err != nil {
		s.T().Fatalf("Failed to login: %v", err)
	}

	return createdUser, loginResp.Token
}

func (s *IntegrationSuite) AuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

func (s *IntegrationSuite) POST(path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody string
	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = str
		} else {
			jsonBytes, _ := json.Marshal(body)
			reqBody = string(jsonBytes)
		}
	}

	req := httptest.NewRequest("POST", path, strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) GET(path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) PUT(path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody string
	if body != nil {
		if str, ok := body.(string); ok {
			reqBody = str
		} else {
			jsonBytes, _ := json.Marshal(body)
			reqBody = string(jsonBytes)
		}
	}

	req := httptest.NewRequest("PUT", path, strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) DELETE(path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	s.Echo.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) AssertOK(rec *httptest.ResponseRecorder) {
	s.Equal(http.StatusOK, rec.Code)
}

func (s *IntegrationSuite) AssertStatus(code int, rec *httptest.ResponseRecorder) {
	s.Equal(code, rec.Code)
}

func (s *IntegrationSuite) AssertSuccessResponse(rec *httptest.ResponseRecorder) {
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Nil(err)
	s.Nil(resp.Error)
}

func (s *IntegrationSuite) AssertFailResponse(rec *httptest.ResponseRecorder) {
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Nil(err)
	s.NotNil(resp.Error)
}

func (s *IntegrationSuite) GetResponseData(rec *httptest.ResponseRecorder, v interface{}) {
	json.Unmarshal(rec.Body.Bytes(), &response.Response{
		Data: v,
	})
}
