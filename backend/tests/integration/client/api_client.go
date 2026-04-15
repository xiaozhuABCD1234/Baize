package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"backend/internal/api/middleware"
	"backend/internal/api/routes"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type APIClient struct {
	Echo *echo.Echo
	DB   *gorm.DB
}

func NewAPIClient(db *gorm.DB) *APIClient {
	e := echo.New()
	SetupRouter(e, db)
	return &APIClient{
		Echo: e,
		DB:   db,
	}
}

func SetupRouter(e *echo.Echo, db *gorm.DB) {
	logger := slog.Default()

	workRepo := repository.NewWorkRepository(db)
	workMediaRepo := repository.NewWorkMediaRepository(db)
	craftRepo := repository.NewCraftRepository(db)
	userRepo := repository.NewUserRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	followRepo := repository.NewFollowRepository(db)
	regionRepo := repository.NewRegionRepository(db)
	categoryRepo := repository.NewICHCategoryRepository(db)

	workSvc := svc.NewWorkService(workRepo, workMediaRepo, craftRepo, userRepo, logger)
	commentSvc := svc.NewCommentService(commentRepo, workRepo, userRepo, logger)
	favoriteSvc := svc.NewFavoriteService(favoriteRepo, workRepo, userRepo, logger)
	followSvc := svc.NewFollowService(followRepo, userRepo, logger)
	craftSvc := svc.NewCraftService(craftRepo, categoryRepo, logger)
	regionSvc := svc.NewRegionService(regionRepo, logger)
	categorySvc := svc.NewICHCategoryService(categoryRepo, logger)
	userSvc := svc.NewUserService(userRepo)

	deps := routes.HandlerDeps{
		UserService:     userSvc,
		WorkService:     workSvc,
		CommentService:  commentSvc,
		FavoriteService: favoriteSvc,
		FollowService:   followSvc,
		CraftService:    craftSvc,
		RegionService:   regionSvc,
		CategoryService: categorySvc,
		Logger:          nil,
	}

	routes.SetupRouter(e, deps)
}

func (c *APIClient) DoRequest(method, path string, body interface{}, userID uint) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if userID > 0 {
		req.Header.Set("Authorization", "Bearer test-token")
	}

	rec := httptest.NewRecorder()
	e := c.Echo.NewContext(req, rec)

	if userID > 0 {
		e.Set(middleware.ContextKeyUserID, userID)
		e.Set(middleware.ContextKeyEmail, "test@example.com")
		e.Set(middleware.ContextKeyUserType, "user")
	}

	c.Echo.ServeHTTP(rec, req)
	return rec
}

func (c *APIClient) AuthenticatedRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	c.Echo.ServeHTTP(rec, req)
	return rec
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
	Page    *PageInfo       `json:"page,omitempty"`
}

type APIError struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type PageInfo struct {
	PageNum  int   `json:"page_num"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func ParseResponse(t *testing.T, rec *httptest.ResponseRecorder) *APIResponse {
	var resp APIResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return &resp
}

func CreateTestUser(t *testing.T, db *gorm.DB, username, email string) *models.User {
	user := &models.User{
		Username: username,
		Email:    email,
		Password: "$2a$10$dummy_hash",
		UserType: models.UserTypeUser,
		Status:   models.UserStatusActive,
		Phone:    "138" + fmt.Sprintf("%08d", len(username)*100),
	}
	err := db.Create(user).Error
	assert.NoError(t, err)
	return user
}

func CreateTestWork(t *testing.T, db *gorm.DB, userID uint, title string) *models.Work {
	work := &models.Work{
		UserID:      userID,
		Title:       title,
		Content:     "Test content",
		ContentType: models.ContentTypeImage,
		Status:      models.WorkStatusPublished,
	}
	err := db.Create(work).Error
	assert.NoError(t, err)
	return work
}

func CreateTestComment(t *testing.T, db *gorm.DB, workID, userID uint, content string) *models.Comment {
	comment := &models.Comment{
		WorkID:  workID,
		UserID:  userID,
		Content: content,
		Status:  models.CommentStatusActive,
	}
	err := db.Create(comment).Error
	assert.NoError(t, err)
	return comment
}

func CreateTestCraft(t *testing.T, db *gorm.DB, name string) *models.Craft {
	craft := &models.Craft{
		Name:        name,
		Description: "Test craft description",
		Difficulty:  1,
	}
	err := db.Create(craft).Error
	assert.NoError(t, err)
	return craft
}

func CreateTestRegion(t *testing.T, db *gorm.DB, name string) *models.Region {
	region := &models.Region{
		Name:  name,
		Code:  "test-code",
		Level: 1,
	}
	err := db.Create(region).Error
	assert.NoError(t, err)
	return region
}

func CreateTestCategory(t *testing.T, db *gorm.DB, name string) *models.ICHCategory {
	category := &models.ICHCategory{
		Name:  name,
		Level: 1,
	}
	err := db.Create(category).Error
	assert.NoError(t, err)
	return category
}
