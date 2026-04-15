package e2e

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/internal/api/routes"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testServer struct {
	Echo     *echo.Echo
	DB       *gorm.DB
	Handlers *routes.HandlerDeps
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.Region{},
		&models.ICHCategory{},
		&models.Craft{},
		&models.Work{},
		&models.WorkMedia{},
		&models.Comment{},
		&models.Favorite{},
		&models.Follow{},
		&models.Like{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func setupTestServer(t *testing.T) *testServer {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	userSvc := svc.NewUserService(userRepo)

	commentRepo := repository.NewCommentRepository(db)
	workRepo := repository.NewWorkRepository(db)
	workMediaRepo := repository.NewWorkMediaRepository(db)
	craftRepo := repository.NewCraftRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	followRepo := repository.NewFollowRepository(db)
	regionRepo := repository.NewRegionRepository(db)
	categoryRepo := repository.NewICHCategoryRepository(db)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	workSvc := svc.NewWorkService(workRepo, workMediaRepo, craftRepo, userRepo, logger)
	commentSvc := svc.NewCommentService(commentRepo, workRepo, userRepo, logger)
	favoriteSvc := svc.NewFavoriteService(favoriteRepo, workRepo, userRepo, logger)
	followSvc := svc.NewFollowService(followRepo, userRepo, logger)
	craftSvc := svc.NewCraftService(craftRepo, categoryRepo, logger)
	regionSvc := svc.NewRegionService(regionRepo, logger)
	categorySvc := svc.NewICHCategoryService(categoryRepo, logger)

	deps := &routes.HandlerDeps{
		UserService:     userSvc,
		WorkService:     workSvc,
		CommentService:  commentSvc,
		FavoriteService: favoriteSvc,
		FollowService:   followSvc,
		CraftService:    craftSvc,
		RegionService:   regionSvc,
		CategoryService: categorySvc,
		Logger:          logger,
	}

	e := echo.New()
	routes.SetupRouter(e, *deps)

	return &testServer{
		Echo:     e,
		DB:       db,
		Handlers: deps,
	}
}

func (ts *testServer) close() {
	sqlDB, _ := ts.DB.DB()
	sqlDB.Close()
}

type requestHelper struct {
	server    *testServer
	authToken string
}

func newRequestHelper(server *testServer) *requestHelper {
	return &requestHelper{server: server}
}

func (rh *requestHelper) setAuthToken(token string) {
	rh.authToken = token
}

func (rh *requestHelper) post(path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if rh.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+rh.authToken)
	}
	rec := httptest.NewRecorder()
	rh.server.Echo.ServeHTTP(rec, req)
	return rec
}

func (rh *requestHelper) get(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	if rh.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+rh.authToken)
	}
	rec := httptest.NewRecorder()
	rh.server.Echo.ServeHTTP(rec, req)
	return rec
}

func (rh *requestHelper) put(path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PUT", path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if rh.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+rh.authToken)
	}
	rec := httptest.NewRecorder()
	rh.server.Echo.ServeHTTP(rec, req)
	return rec
}

func (rh *requestHelper) delete(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)
	if rh.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+rh.authToken)
	}
	rec := httptest.NewRecorder()
	rh.server.Echo.ServeHTTP(rec, req)
	return rec
}

func parseResponse(rec *httptest.ResponseRecorder) response.Response {
	var resp response.Response
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	if err != nil {
		return response.Response{}
	}
	return resp
}

func assertSuccessResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) response.Response {
	assert.Equal(t, expectedStatus, rec.Code)
	resp := parseResponse(rec)
	assert.Nil(t, resp.Error, "expected no error in response")
	assert.NotNil(t, resp.Data, "expected data in response")
	return resp
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assert.Equal(t, expectedStatus, rec.Code)
	resp := parseResponse(rec)
	assert.NotNil(t, resp.Error, "expected error in response")
	if expectedCode != "" {
		assert.Equal(t, expectedCode, resp.Error.Code)
	}
}

type authResult struct {
	UserID    uint
	Email     string
	Username  string
	Token     string
	ExpiresIn int
}

func registerUser(t *testing.T, rh *requestHelper, email, username, password string) *authResult {
	body := `{"email":"` + email + `","username":"` + username + `","password":"` + password + `"}`
	rec := rh.post("/api/v1/users/register", body)

	if rec.Code != http.StatusOK {
		return nil
	}

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data := resp.Data.(map[string]interface{})

	token, _ := data["token"].(string)
	expiresIn, _ := data["expires_in"].(float64)

	return &authResult{
		Email:     email,
		Username:  username,
		Token:     token,
		ExpiresIn: int(expiresIn),
	}
}

func loginUser(t *testing.T, rh *requestHelper, email, password string) *authResult {
	body := `{"email":"` + email + `","password":"` + password + `"}`
	rec := rh.post("/api/v1/users/login", body)

	if rec.Code != http.StatusOK {
		return nil
	}

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data := resp.Data.(map[string]interface{})

	token, _ := data["token"].(string)
	expiresIn, _ := data["expires_in"].(float64)

	return &authResult{
		Email:     email,
		Token:     token,
		ExpiresIn: int(expiresIn),
	}
}

func generateTestToken(userID uint, email, userType string) string {
	token, _ := utils.GenerateAccessToken(userID, email, userType)
	return token
}

func setupJWTTestEnv() func() {
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

func createTestCraft(t *testing.T, db *gorm.DB, name string) *models.Craft {
	craft := &models.Craft{
		Name:        name,
		Description: "test description",
		Difficulty:  3,
	}
	db.Create(craft)
	return craft
}

func createTestRegion(t *testing.T, db *gorm.DB, name string) *models.Region {
	region := &models.Region{
		Name:  name,
		Code:  "TEST-" + name,
		Level: 1,
	}
	db.Create(region)
	return region
}
