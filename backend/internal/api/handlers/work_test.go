package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/internal/api/middleware"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockWorkRepository struct {
	works                map[uint]*models.Work
	worksByUserID        map[uint][]*models.Work
	getByIDErr           error
	getByIDWithAllErr    error
	getByIDWithSelectErr error
	listErr              error
	listPagErr           error
	listByUserIDErr      error
	listByCraftIDErr     error
	listTopErr           error
	listRecommendedErr   error
	createErr            error
	updateErr            error
	updateStatusErr      error
	incrementCountErr    error
	deleteErr            error
	nextID               uint
}

func newMockWorkRepository() *mockWorkRepository {
	return &mockWorkRepository{
		works:         make(map[uint]*models.Work),
		worksByUserID: make(map[uint][]*models.Work),
		nextID:        1,
	}
}

func (m *mockWorkRepository) addTestWork(userID uint, title string, status models.WorkStatus) *models.Work {
	work := &models.Work{
		UserID:   userID,
		Title:    title,
		Content:  "test content",
		Status:   status,
		CraftID:  1,
		RegionID: 1,
	}
	work.ID = m.nextID
	m.works[m.nextID] = work
	m.worksByUserID[userID] = append(m.worksByUserID[userID], work)
	m.nextID++
	return work
}

func (m *mockWorkRepository) Create(ctx context.Context, work *models.Work) error {
	if m.createErr != nil {
		return m.createErr
	}
	work.ID = m.nextID
	m.works[m.nextID] = work
	m.worksByUserID[work.UserID] = append(m.worksByUserID[work.UserID], work)
	m.nextID++
	return nil
}

func (m *mockWorkRepository) GetByID(ctx context.Context, id uint) (*models.Work, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if work, ok := m.works[id]; ok {
		return work, nil
	}
	return nil, repository.ErrWorkNotFound
}

func (m *mockWorkRepository) GetByIDWithAll(ctx context.Context, id uint) (*models.Work, error) {
	if m.getByIDWithAllErr != nil {
		return nil, m.getByIDWithAllErr
	}
	if work, ok := m.works[id]; ok {
		return work, nil
	}
	return nil, repository.ErrWorkNotFound
}

func (m *mockWorkRepository) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.Work, error) {
	if m.getByIDWithSelectErr != nil {
		return nil, m.getByIDWithSelectErr
	}
	if work, ok := m.works[id]; ok {
		return work, nil
	}
	return nil, repository.ErrWorkNotFound
}

func (m *mockWorkRepository) List(ctx context.Context, orderBy string) ([]models.Work, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	works := make([]models.Work, 0, len(m.works))
	for _, w := range m.works {
		works = append(works, *w)
	}
	return works, nil
}

func (m *mockWorkRepository) ListWithAll(ctx context.Context, orderBy string) ([]models.Work, error) {
	return m.List(ctx, orderBy)
}

func (m *mockWorkRepository) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.Work, int64, error) {
	if m.listPagErr != nil {
		return nil, 0, m.listPagErr
	}
	works := make([]models.Work, 0, len(m.works))
	for _, w := range m.works {
		works = append(works, *w)
	}
	return works, int64(len(works)), nil
}

func (m *mockWorkRepository) ListByUserID(ctx context.Context, userID uint) ([]models.Work, error) {
	if m.listByUserIDErr != nil {
		return nil, m.listByUserIDErr
	}
	works := make([]models.Work, 0, len(m.worksByUserID[userID]))
	for _, w := range m.worksByUserID[userID] {
		works = append(works, *w)
	}
	return works, nil
}

func (m *mockWorkRepository) ListByCraftID(ctx context.Context, craftID uint) ([]models.Work, error) {
	if m.listByCraftIDErr != nil {
		return nil, m.listByCraftIDErr
	}
	works := make([]models.Work, 0)
	for _, w := range m.works {
		if w.CraftID == craftID {
			works = append(works, *w)
		}
	}
	return works, nil
}

func (m *mockWorkRepository) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Work, error) {
	return m.List(ctx, "")
}

func (m *mockWorkRepository) ListByStatus(ctx context.Context, status models.WorkStatus) ([]models.Work, error) {
	return m.List(ctx, "")
}

func (m *mockWorkRepository) ListPublished(ctx context.Context, orderBy string) ([]models.Work, error) {
	return m.List(ctx, orderBy)
}

func (m *mockWorkRepository) ListTop(ctx context.Context, limit int) ([]models.Work, error) {
	if m.listTopErr != nil {
		return nil, m.listTopErr
	}
	works := make([]models.Work, 0)
	for _, w := range m.works {
		if w.IsTop && w.Status == models.WorkStatusPublished {
			works = append(works, *w)
			if len(works) >= limit {
				break
			}
		}
	}
	return works, nil
}

func (m *mockWorkRepository) ListRecommended(ctx context.Context, limit int) ([]models.Work, error) {
	if m.listRecommendedErr != nil {
		return nil, m.listRecommendedErr
	}
	works := make([]models.Work, 0)
	for _, w := range m.works {
		if w.IsRecommended && w.Status == models.WorkStatusPublished {
			works = append(works, *w)
			if len(works) >= limit {
				break
			}
		}
	}
	return works, nil
}

func (m *mockWorkRepository) Update(ctx context.Context, work *models.Work) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.works[work.ID]; !ok {
		return repository.ErrWorkNotFound
	}
	m.works[work.ID] = work
	return nil
}

func (m *mockWorkRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.works[id]; !ok {
		return repository.ErrWorkNotFound
	}
	return nil
}

func (m *mockWorkRepository) UpdateStatus(ctx context.Context, id uint, status models.WorkStatus) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if _, ok := m.works[id]; !ok {
		return repository.ErrWorkNotFound
	}
	m.works[id].Status = status
	return nil
}

func (m *mockWorkRepository) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	if m.incrementCountErr != nil {
		return m.incrementCountErr
	}
	if _, ok := m.works[id]; !ok {
		return repository.ErrWorkNotFound
	}
	return nil
}

func (m *mockWorkRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.works[id]; !ok {
		return repository.ErrWorkNotFound
	}
	delete(m.works, id)
	return nil
}

func (m *mockWorkRepository) ForceDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockWorkRepository) Upsert(ctx context.Context, work *models.Work) error {
	return m.Create(ctx, work)
}

func (m *mockWorkRepository) UpsertBatch(ctx context.Context, works []models.Work) error {
	return nil
}

func (m *mockWorkRepository) WithTransaction(tx *gorm.DB) repository.WorkRepository {
	return m
}

type mockWorkMediaRepository struct {
	mediaByWorkID map[uint][]models.WorkMedia
	createErr     error
	deleteErr     error
	nextID        uint
}

func newMockWorkMediaRepository() *mockWorkMediaRepository {
	return &mockWorkMediaRepository{
		mediaByWorkID: make(map[uint][]models.WorkMedia),
		nextID:        1,
	}
}

func (m *mockWorkMediaRepository) Create(ctx context.Context, media *models.WorkMedia) error {
	if m.createErr != nil {
		return m.createErr
	}
	media.ID = m.nextID
	m.mediaByWorkID[media.WorkID] = append(m.mediaByWorkID[media.WorkID], *media)
	m.nextID++
	return nil
}

func (m *mockWorkMediaRepository) CreateBatch(ctx context.Context, mediaList []models.WorkMedia) error {
	if m.createErr != nil {
		return m.createErr
	}
	for i := range mediaList {
		mediaList[i].ID = m.nextID
		m.mediaByWorkID[mediaList[i].WorkID] = append(m.mediaByWorkID[mediaList[i].WorkID], mediaList[i])
		m.nextID++
	}
	return nil
}

func (m *mockWorkMediaRepository) GetByID(ctx context.Context, id uint) (*models.WorkMedia, error) {
	return nil, nil
}

func (m *mockWorkMediaRepository) ListByWorkID(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return m.mediaByWorkID[workID], nil
}

func (m *mockWorkMediaRepository) ListImages(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return m.mediaByWorkID[workID], nil
}

func (m *mockWorkMediaRepository) ListVideos(ctx context.Context, workID uint) ([]models.WorkMedia, error) {
	return m.mediaByWorkID[workID], nil
}

func (m *mockWorkMediaRepository) Update(ctx context.Context, media *models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockWorkMediaRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

func (m *mockWorkMediaRepository) DeleteByWorkID(ctx context.Context, workID uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.mediaByWorkID, workID)
	return nil
}

func (m *mockWorkMediaRepository) DeleteBatch(ctx context.Context, ids []uint) error {
	return nil
}

func (m *mockWorkMediaRepository) Upsert(ctx context.Context, media *models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) UpsertBatch(ctx context.Context, mediaList []models.WorkMedia) error {
	return nil
}

func (m *mockWorkMediaRepository) WithTransaction(tx *gorm.DB) repository.WorkMediaRepository {
	return m
}

type mockCraftRepository struct {
	crafts     map[uint]*models.Craft
	getByIDErr error
	createErr  error
	listErr    error
	nextID     uint
}

func newMockCraftRepository() *mockCraftRepository {
	return &mockCraftRepository{
		crafts: make(map[uint]*models.Craft),
		nextID: 1,
	}
}

func (m *mockCraftRepository) addTestCraft(name string) *models.Craft {
	craft := &models.Craft{
		Name:        name,
		Description: "test craft",
		Difficulty:  1,
	}
	craft.ID = m.nextID
	m.crafts[m.nextID] = craft
	m.nextID++
	return craft
}

func (m *mockCraftRepository) Create(ctx context.Context, craft *models.Craft) error {
	if m.createErr != nil {
		return m.createErr
	}
	craft.ID = m.nextID
	m.crafts[m.nextID] = craft
	m.nextID++
	return nil
}

func (m *mockCraftRepository) GetByID(ctx context.Context, id uint) (*models.Craft, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if craft, ok := m.crafts[id]; ok {
		return craft, nil
	}
	return nil, repository.ErrCraftNotFound
}

func (m *mockCraftRepository) GetByIDWithCategory(ctx context.Context, id uint) (*models.Craft, error) {
	return m.GetByID(ctx, id)
}

func (m *mockCraftRepository) GetByName(ctx context.Context, name string) (*models.Craft, error) {
	for _, craft := range m.crafts {
		if craft.Name == name {
			return craft, nil
		}
	}
	return nil, repository.ErrCraftNotFound
}

func (m *mockCraftRepository) List(ctx context.Context, orderBy string) ([]models.Craft, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	crafts := make([]models.Craft, 0, len(m.crafts))
	for _, c := range m.crafts {
		crafts = append(crafts, *c)
	}
	return crafts, nil
}

func (m *mockCraftRepository) ListWithCategory(ctx context.Context, orderBy string) ([]models.Craft, error) {
	return m.List(ctx, orderBy)
}

func (m *mockCraftRepository) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Craft, error) {
	return m.List(ctx, "")
}

func (m *mockCraftRepository) ListByDifficulty(ctx context.Context, difficulty int8) ([]models.Craft, error) {
	return m.List(ctx, "")
}

func (m *mockCraftRepository) Update(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockCraftRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockCraftRepository) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockCraftRepository) Upsert(ctx context.Context, craft *models.Craft) error {
	return nil
}

func (m *mockCraftRepository) UpsertBatch(ctx context.Context, crafts []models.Craft) error {
	return nil
}

func (m *mockCraftRepository) WithTransaction(tx *gorm.DB) repository.CraftRepository {
	return m
}

func setupWorkTestEnv() func() {
	oldLevel := os.Getenv("LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "error")
	return func() {
		if oldLevel != "" {
			os.Setenv("LOG_LEVEL", oldLevel)
		}
	}
}

func createWorkHandler() (*WorkHandler, *mockWorkRepository, *mockWorkMediaRepository, *mockCraftRepository, *mockUserRepository) {
	mockWorkRepo := newMockWorkRepository()
	mockMediaRepo := newMockWorkMediaRepository()
	mockCraftRepo := newMockCraftRepository()
	mockUserRepo := newMockUserRepository()

	mockCraftRepo.addTestCraft("test craft")

	workSvc := svc.NewWorkService(mockWorkRepo, mockMediaRepo, mockCraftRepo, mockUserRepo, slog.Default())
	h := NewWorkHandler(workSvc)

	return h, mockWorkRepo, mockMediaRepo, mockCraftRepo, mockUserRepo
}

func createEchoContextWithParams(method, path string, paramNames []string, paramValues []string, body ...string) (*echo.Echo, *echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if len(body) > 0 && body[0] != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body[0]))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if len(paramNames) > 0 && len(paramValues) > 0 {
		pv := make(echo.PathValues, len(paramNames))
		for i, name := range paramNames {
			pv[i].Name = name
			if i < len(paramValues) {
				pv[i].Value = paramValues[i]
			}
		}
		c.SetPathValues(pv)
	}
	return e, c, rec
}

// ================================================================================
// ListWorks - 获取作品列表
// ================================================================================

func TestWorkHandler_ListWorks_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)
	mockWorkRepo.addTestWork(1, "work2", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=1&page_size=10", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, int64(2), resp.Page.Total)
}

func TestWorkHandler_ListWorks_EmptyList(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContext("GET", "/works?page=1", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	resp := parseResponse(t, rec)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, int64(0), resp.Page.Total)
}

func TestWorkHandler_ListWorks_PageZero(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=0", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 1, resp.Page.PageNum)
}

func TestWorkHandler_ListWorks_PageNegative(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=-5", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 1, resp.Page.PageNum)
}

func TestWorkHandler_ListWorks_PageSizeZero(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=1&page_size=0", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 10, resp.Page.PageSize)
}

func TestWorkHandler_ListWorks_PageSizeExceedsMax(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=1&page_size=101", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 10, resp.Page.PageSize)
}

func TestWorkHandler_ListWorks_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.listPagErr = errors.New("database error")

	c, rec := setupEchoContext("GET", "/works?page=1", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// ListTopWorks - 获取精选作品
// ================================================================================

func TestWorkHandler_ListTopWorks_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "top work", models.WorkStatusPublished)
	work.IsTop = true

	c, rec := setupEchoContext("GET", "/works/top?limit=10", "")

	err := h.ListTopWorks(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_ListTopWorks_DefaultLimit(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "top work", models.WorkStatusPublished)
	work.IsTop = true

	c, rec := setupEchoContext("GET", "/works/top", "")

	err := h.ListTopWorks(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_ListTopWorks_LimitZero(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "top work", models.WorkStatusPublished)
	work.IsTop = true

	c, rec := setupEchoContext("GET", "/works/top?limit=0", "")

	err := h.ListTopWorks(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_ListTopWorks_LimitExceedsMax(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "top work", models.WorkStatusPublished)
	work.IsTop = true

	c, rec := setupEchoContext("GET", "/works/top?limit=51", "")

	err := h.ListTopWorks(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_ListTopWorks_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.listTopErr = errors.New("database error")

	c, rec := setupEchoContext("GET", "/works/top?limit=10", "")

	err := h.ListTopWorks(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// ListRecommendedWorks - 获取推荐作品
// ================================================================================

func TestWorkHandler_ListRecommendedWorks_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "recommended work", models.WorkStatusPublished)
	work.IsRecommended = true

	c, rec := setupEchoContext("GET", "/works/recommended?limit=10", "")

	err := h.ListRecommendedWorks(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_ListRecommendedWorks_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.listRecommendedErr = errors.New("database error")

	c, rec := setupEchoContext("GET", "/works/recommended?limit=10", "")

	err := h.ListRecommendedWorks(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// GetWork - 获取单个作品
// ================================================================================

func TestWorkHandler_GetWork_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "test work", models.WorkStatusPublished)

	_, c, rec := createEchoContextWithParams("GET", "/works/1", []string{"id"}, []string{"1"})

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_GetWork_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContextWithParams("GET", "/works/abc", "", "id", "abc")

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_GetWork_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	_, c, rec := createEchoContextWithParams("GET", "/works/999", []string{"id"}, []string{"999"})

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_GetWork_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.getByIDWithSelectErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("GET", "/works/1", []string{"id"}, []string{"1"})

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// GetWorkDetailed - 获取作品详情
// ================================================================================

func TestWorkHandler_GetWorkDetailed_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "test work", models.WorkStatusPublished)

	_, c, rec := createEchoContextWithParams("GET", "/works/1/detailed", []string{"id"}, []string{"1"})

	err := h.GetWorkDetailed(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_GetWorkDetailed_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContext("GET", "/works/abc/detailed", "")

	err := h.GetWorkDetailed(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_GetWorkDetailed_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	_, c, rec := createEchoContextWithParams("GET", "/works/999/detailed", []string{"id"}, []string{"999"})

	err := h.GetWorkDetailed(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

// ================================================================================
// CreateWork - 创建作品
// ================================================================================

func TestWorkHandler_CreateWork_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, mockCraftRepo, _ := createWorkHandler()
	mockCraftRepo.addTestCraft("test craft")

	body := `{"title":"new work","content":"content","content_type":1,"craft_id":1}`
	c, rec := setupEchoContext("POST", "/works", body)
	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateWork(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusCreated)
}

func TestWorkHandler_CreateWork_InvalidJSON(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{invalid json}`
	c, rec := setupEchoContext("POST", "/works", body)
	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_CreateWork_CraftNotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"title":"new work","content":"content","content_type":1,"craft_id":999}`
	c, rec := setupEchoContext("POST", "/works", body)
	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_CreateWork_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.createErr = errors.New("database error")

	body := `{"title":"new work","content":"content","content_type":1,"craft_id":1}`
	c, rec := setupEchoContext("POST", "/works", body)
	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// UpdateWork - 更新作品
// ================================================================================

func TestWorkHandler_UpdateWork_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "original title", models.WorkStatusDraft)

	body := `{"title":"updated title","content":"updated content","content_type":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1", []string{"id"}, []string{"1"}, body)

	err := h.UpdateWork(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_UpdateWork_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"title":"updated title"}`
	c, rec := setupEchoContextWithParams("PUT", "/works/abc", body, "id", "abc")

	err := h.UpdateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_UpdateWork_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"title":"updated title"}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/999", []string{"id"}, []string{"999"}, body)

	err := h.UpdateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_UpdateWork_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "original", models.WorkStatusDraft)
	mockWorkRepo.updateErr = errors.New("database error")

	body := `{"title":"updated"}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1", []string{"id"}, []string{"1"}, body)

	err := h.UpdateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// DeleteWork - 删除作品
// ================================================================================

func TestWorkHandler_DeleteWork_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "draft work", models.WorkStatusDraft)

	_, c, rec := createEchoContextWithParams("DELETE", "/works/1", []string{"id"}, []string{"1"})

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_DeleteWork_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContextWithParams("DELETE", "/works/abc", "", "id", "abc")

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_DeleteWork_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	_, c, rec := createEchoContextWithParams("DELETE", "/works/999", []string{"id"}, []string{"999"})

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_DeleteWork_CannotDeletePublished(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "published work", models.WorkStatusPublished)

	_, c, rec := createEchoContextWithParams("DELETE", "/works/1", []string{"id"}, []string{"1"})

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_DeleteWork_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "draft", models.WorkStatusDraft)
	mockWorkRepo.deleteErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("DELETE", "/works/1", []string{"id"}, []string{"1"})

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// UpdateStatus - 更新作品状态 (Admin)
// ================================================================================

func TestWorkHandler_UpdateStatus_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)

	body := `{"status":3}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_UpdateStatus_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"status":3}`
	c, rec := setupEchoContext("PUT", "/works/abc/status", body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_UpdateStatus_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"status":3}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/999/status", []string{"id"}, []string{"999"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)

	body := `{"status":99}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

// ================================================================================
// IncrementCount - 递增计数
// ================================================================================

func TestWorkHandler_IncrementCount_Success(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)

	body := `{"field":"view_count","delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1/count", []string{"id"}, []string{"1"}, body)

	err := h.IncrementCount(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestWorkHandler_IncrementCount_InvalidID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"field":"view_count","delta":1}`
	c, rec := setupEchoContext("PUT", "/works/abc/count", body)

	err := h.IncrementCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_IncrementCount_NotFound(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"field":"view_count","delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/999/count", []string{"id"}, []string{"999"}, body)

	err := h.IncrementCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_IncrementCount_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)
	mockWorkRepo.incrementCountErr = errors.New("database error")

	body := `{"field":"view_count","delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/works/1/count", []string{"id"}, []string{"1"}, body)

	err := h.IncrementCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// Pagination Boundary Tests - 分页边界测试
// ================================================================================

func TestWorkHandler_ListWorks_PaginationBoundaries(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	for i := 0; i < 15; i++ {
		mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)
	}

	tests := []struct {
		name             string
		page             string
		pageSize         string
		expectedPageNum  int
		expectedPageSize int
	}{
		{"page=1_size=5", "1", "5", 1, 5},
		{"page=2_size=5", "2", "5", 2, 5},
		{"page=0_defaults_to_1", "0", "10", 1, 10},
		{"page=-1_defaults_to_1", "-1", "10", 1, 10},
		{"page_size=0_defaults_to_10", "1", "0", 1, 10},
		{"page_size=-5_defaults_to_10", "1", "-5", 1, 10},
		{"page_size=100_is_valid", "1", "100", 1, 100},
		{"page_size=10000_defaults_to_10", "1", "10000", 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := setupEchoContext("GET", "/works?page="+tt.page+"&page_size="+tt.pageSize, "")

			err := h.ListWorks(c)
			assert.NoError(t, err)
			resp := assertSuccessResponse(t, rec, http.StatusOK)
			assert.Equal(t, tt.expectedPageNum, resp.Page.PageNum)
			assert.Equal(t, tt.expectedPageSize, resp.Page.PageSize)
		})
	}
}

// ================================================================================
// ID Boundary Tests - ID边界测试
// ================================================================================

func TestWorkHandler_GetWork_ZeroID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContextWithParams("GET", "/works/0", "", "id", "0")

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_GetWork_MaxUint32(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContextWithParams("GET", "/works/4294967295", "", "id", "4294967295")

	err := h.GetWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.WorkNotFound)
}

func TestWorkHandler_UpdateWork_ZeroID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"title":"test"}`
	c, rec := setupEchoContext("PUT", "/works/0", body)

	err := h.UpdateWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_DeleteWork_ZeroID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContext("DELETE", "/works/0", "")

	err := h.DeleteWork(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_UpdateStatus_ZeroID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"status":1}`
	c, rec := setupEchoContext("PUT", "/works/0/status", body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestWorkHandler_IncrementCount_ZeroID(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	body := `{"field":"view_count","delta":1}`
	c, rec := setupEchoContext("PUT", "/works/0/count", body)

	err := h.IncrementCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

// ================================================================================
// ListTopWorks/Recommended - Limit边界测试
// ================================================================================

func TestWorkHandler_ListTopWorks_LimitBoundaries(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "top", models.WorkStatusPublished)
	work.IsTop = true

	tests := []struct {
		name  string
		limit string
	}{
		{"limit=1", "1"},
		{"limit=50", "50"},
		{"limit=51_defaults_to_10", "51"},
		{"limit=0_defaults_to_10", "0"},
		{"limit=-1", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := setupEchoContext("GET", "/works/top?limit="+tt.limit, "")

			err := h.ListTopWorks(c)
			assert.NoError(t, err)
			assertSuccessResponse(t, rec, http.StatusOK)
		})
	}
}

func TestWorkHandler_ListRecommendedWorks_LimitBoundaries(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	work := mockWorkRepo.addTestWork(1, "recommended", models.WorkStatusPublished)
	work.IsRecommended = true

	tests := []struct {
		name  string
		limit string
	}{
		{"limit=1", "1"},
		{"limit=50", "50"},
		{"limit=51_defaults_to_10", "51"},
		{"limit=0_defaults_to_10", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := setupEchoContext("GET", "/works/recommended?limit="+tt.limit, "")

			err := h.ListRecommendedWorks(c)
			assert.NoError(t, err)
			assertSuccessResponse(t, rec, http.StatusOK)
		})
	}
}

// ================================================================================
// Response Format Tests - 响应格式测试
// ================================================================================

func TestWorkHandler_ListWorks_ResponseFormat(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work1", models.WorkStatusPublished)

	c, rec := setupEchoContext("GET", "/works?page=1&page_size=10", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)

	var resp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Data)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, 1, resp.Page.PageNum)
	assert.Equal(t, 10, resp.Page.PageSize)
	assert.Equal(t, int64(1), resp.Page.Total)
}

func TestWorkHandler_GetWork_ResponseFormat(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "test work", models.WorkStatusPublished)

	c, rec := setupEchoContextWithParams("GET", "/works/1", "", "id", "1")

	err := h.GetWork(c)
	assert.NoError(t, err)

	var resp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Data)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test work", data["title"])
}

func TestWorkHandler_ErrorResponse_Format(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, _, _, _, _ := createWorkHandler()

	c, rec := setupEchoContextWithParams("GET", "/works/999", "", "id", "999")

	err := h.GetWork(c)
	assert.NoError(t, err)

	var resp response.Response
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, response.WorkNotFound, resp.Error.Code)
}

// ================================================================================
// Service Error Propagation Tests - 服务错误传播测试
// ================================================================================

func TestWorkHandler_GetWorkDetailed_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)
	mockWorkRepo.getByIDWithAllErr = errors.New("database error")

	c, rec := setupEchoContextWithParams("GET", "/works/1/detailed", "", "id", "1")

	err := h.GetWorkDetailed(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestWorkHandler_ListByCraftID_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.listByCraftIDErr = errors.New("database error")

	c, rec := setupEchoContext("GET", "/works?craft_id=1", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestWorkHandler_ListByUserID_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.listByUserIDErr = errors.New("database error")

	c, rec := setupEchoContext("GET", "/works?user_id=1", "")

	err := h.ListWorks(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestWorkHandler_UpdateStatus_ServiceError(t *testing.T) {
	restore := setupWorkTestEnv()
	defer restore()

	h, mockWorkRepo, _, _, _ := createWorkHandler()
	mockWorkRepo.addTestWork(1, "work", models.WorkStatusPublished)
	mockWorkRepo.updateStatusErr = errors.New("database error")

	body := `{"status":3}`
	c, rec := setupEchoContextWithParams("PUT", "/works/1/status", body, "id", "1")

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}
