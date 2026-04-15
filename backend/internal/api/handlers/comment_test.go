package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"backend/internal/api/middleware"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockCommentRepository struct {
	comments              map[uint]*models.Comment
	getByIDErr            error
	getByIDWithUserErr    error
	listErr               error
	listByWorkIDErr       error
	listRootByWorkIDErr   error
	listByUserIDErr       error
	listByParentIDErr     error
	listByRootIDErr       error
	listWithPagErr        error
	createErr             error
	updateErr             error
	updateStatusErr       error
	incrementLikeCountErr error
	deleteErr             error
	nextID                uint
}

func newMockCommentRepository() *mockCommentRepository {
	return &mockCommentRepository{
		comments: make(map[uint]*models.Comment),
		nextID:   1,
	}
}

func (m *mockCommentRepository) addTestComment(workID, userID uint, status models.CommentStatus) *models.Comment {
	comment := &models.Comment{
		WorkID:   workID,
		UserID:   userID,
		ParentID: 0,
		RootID:   0,
		Content:  "test comment content",
		Status:   status,
	}
	comment.ID = m.nextID
	m.comments[m.nextID] = comment
	m.nextID++
	return comment
}

func (m *mockCommentRepository) addTestReply(parentID, rootID, workID, userID uint) *models.Comment {
	comment := &models.Comment{
		WorkID:   workID,
		UserID:   userID,
		ParentID: parentID,
		RootID:   rootID,
		Content:  "test reply content",
		Status:   models.CommentStatusActive,
	}
	comment.ID = m.nextID
	m.comments[m.nextID] = comment
	m.nextID++
	return comment
}

func (m *mockCommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	if m.createErr != nil {
		return m.createErr
	}
	comment.ID = m.nextID
	m.comments[m.nextID] = comment
	m.nextID++
	return nil
}

func (m *mockCommentRepository) GetByID(ctx context.Context, id uint) (*models.Comment, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if comment, ok := m.comments[id]; ok {
		return comment, nil
	}
	return nil, repository.ErrCommentNotFound
}

func (m *mockCommentRepository) GetByIDWithUser(ctx context.Context, id uint) (*models.Comment, error) {
	if m.getByIDWithUserErr != nil {
		return nil, m.getByIDWithUserErr
	}
	if comment, ok := m.comments[id]; ok {
		return comment, nil
	}
	return nil, repository.ErrCommentNotFound
}

func (m *mockCommentRepository) List(ctx context.Context, orderBy string) ([]models.Comment, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	comments := make([]models.Comment, 0, len(m.comments))
	for _, c := range m.comments {
		comments = append(comments, *c)
	}
	return comments, nil
}

func (m *mockCommentRepository) ListByWorkID(ctx context.Context, workID uint) ([]models.Comment, error) {
	if m.listByWorkIDErr != nil {
		return nil, m.listByWorkIDErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.WorkID == workID {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListRootByWorkID(ctx context.Context, workID uint) ([]models.Comment, error) {
	if m.listRootByWorkIDErr != nil {
		return nil, m.listRootByWorkIDErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.WorkID == workID && c.ParentID == 0 {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListByUserID(ctx context.Context, userID uint) ([]models.Comment, error) {
	if m.listByUserIDErr != nil {
		return nil, m.listByUserIDErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.UserID == userID {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListByParentID(ctx context.Context, parentID uint) ([]models.Comment, error) {
	if m.listByParentIDErr != nil {
		return nil, m.listByParentIDErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.ParentID == parentID {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListByRootID(ctx context.Context, rootID uint) ([]models.Comment, error) {
	if m.listByRootIDErr != nil {
		return nil, m.listByRootIDErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.RootID == rootID {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListByStatus(ctx context.Context, status models.CommentStatus) ([]models.Comment, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var comments []models.Comment
	for _, c := range m.comments {
		if c.Status == status {
			comments = append(comments, *c)
		}
	}
	return comments, nil
}

func (m *mockCommentRepository) ListWithPagination(ctx context.Context, workID uint, page, pageSize int) ([]models.Comment, int64, error) {
	if m.listWithPagErr != nil {
		return nil, 0, m.listWithPagErr
	}
	var total int64
	var comments []models.Comment
	for _, c := range m.comments {
		if c.WorkID == workID {
			total++
			comments = append(comments, *c)
		}
	}
	return comments, total, nil
}

func (m *mockCommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.comments[comment.ID]; !ok {
		return repository.ErrCommentNotFound
	}
	m.comments[comment.ID] = comment
	return nil
}

func (m *mockCommentRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.comments[id]; !ok {
		return repository.ErrCommentNotFound
	}
	return nil
}

func (m *mockCommentRepository) UpdateStatus(ctx context.Context, id uint, status models.CommentStatus) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	if _, ok := m.comments[id]; !ok {
		return repository.ErrCommentNotFound
	}
	m.comments[id].Status = status
	return nil
}

func (m *mockCommentRepository) IncrementLikeCount(ctx context.Context, id uint, delta int) error {
	if m.incrementLikeCountErr != nil {
		return m.incrementLikeCountErr
	}
	if _, ok := m.comments[id]; !ok {
		return repository.ErrCommentNotFound
	}
	return nil
}

func (m *mockCommentRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.comments[id]; !ok {
		return repository.ErrCommentNotFound
	}
	delete(m.comments, id)
	return nil
}

func (m *mockCommentRepository) ForceDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockCommentRepository) DeleteByWorkID(ctx context.Context, workID uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for id, c := range m.comments {
		if c.WorkID == workID {
			delete(m.comments, id)
		}
	}
	return nil
}

func (m *mockCommentRepository) Upsert(ctx context.Context, comment *models.Comment) error {
	return m.Create(ctx, comment)
}

func (m *mockCommentRepository) WithTransaction(tx *gorm.DB) repository.CommentRepository {
	return m
}

type mockWorkRepositoryForComment struct {
	works             map[uint]*models.Work
	getByIDErr        error
	incrementCountErr error
	nextID            uint
}

func newMockWorkRepositoryForComment() *mockWorkRepositoryForComment {
	return &mockWorkRepositoryForComment{
		works:  make(map[uint]*models.Work),
		nextID: 1,
	}
}

func (m *mockWorkRepositoryForComment) addTestWork(status models.WorkStatus) *models.Work {
	work := &models.Work{
		UserID:   1,
		Title:    "test work",
		Content:  "test content",
		Status:   status,
		CraftID:  1,
		RegionID: 1,
	}
	work.ID = m.nextID
	m.works[m.nextID] = work
	m.nextID++
	return work
}

func (m *mockWorkRepositoryForComment) Create(ctx context.Context, work *models.Work) error {
	work.ID = m.nextID
	m.works[m.nextID] = work
	m.nextID++
	return nil
}

func (m *mockWorkRepositoryForComment) GetByID(ctx context.Context, id uint) (*models.Work, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if work, ok := m.works[id]; ok {
		return work, nil
	}
	return nil, repository.ErrWorkNotFound
}

func (m *mockWorkRepositoryForComment) GetByIDWithAll(ctx context.Context, id uint) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepositoryForComment) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.Work, error) {
	return m.GetByID(ctx, id)
}

func (m *mockWorkRepositoryForComment) List(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListWithAll(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.Work, int64, error) {
	return nil, 0, nil
}

func (m *mockWorkRepositoryForComment) ListByUserID(ctx context.Context, userID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListByCraftID(ctx context.Context, craftID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListTop(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListByCategoryID(ctx context.Context, categoryID uint) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListRecommended(ctx context.Context, limit int) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListByStatus(ctx context.Context, status models.WorkStatus) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) ListPublished(ctx context.Context, orderBy string) ([]models.Work, error) {
	return nil, nil
}

func (m *mockWorkRepositoryForComment) Update(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepositoryForComment) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return nil
}

func (m *mockWorkRepositoryForComment) UpdateStatus(ctx context.Context, id uint, status models.WorkStatus) error {
	return nil
}

func (m *mockWorkRepositoryForComment) IncrementCount(ctx context.Context, id uint, field string, delta int) error {
	if m.incrementCountErr != nil {
		return m.incrementCountErr
	}
	return nil
}

func (m *mockWorkRepositoryForComment) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepositoryForComment) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockWorkRepositoryForComment) DeleteByUserID(ctx context.Context, userID uint) error {
	return nil
}

func (m *mockWorkRepositoryForComment) Upsert(ctx context.Context, work *models.Work) error {
	return nil
}

func (m *mockWorkRepositoryForComment) UpsertBatch(ctx context.Context, works []models.Work) error {
	return nil
}

func (m *mockWorkRepositoryForComment) WithTransaction(tx *gorm.DB) repository.WorkRepository {
	return m
}

type mockUserRepositoryForComment struct {
	users      map[uint]*models.User
	getByIDErr error
	nextID     uint
}

func newMockUserRepositoryForComment() *mockUserRepositoryForComment {
	return &mockUserRepositoryForComment{
		users:  make(map[uint]*models.User),
		nextID: 1,
	}
}

func (m *mockUserRepositoryForComment) addTestUser(userType models.UserType) *models.User {
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		UserType: userType,
		Status:   models.UserStatusActive,
	}
	user.ID = m.nextID
	m.users[m.nextID] = user
	m.nextID++
	return user
}

func (m *mockUserRepositoryForComment) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForComment) CreateWithProfile(ctx context.Context, user *models.User, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepositoryForComment) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepositoryForComment) GetByIDWithProfile(ctx context.Context, id uint) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForComment) GetByIDWithSelect(ctx context.Context, id uint, preloads ...string) (*models.User, error) {
	return m.GetByID(ctx, id)
}

func (m *mockUserRepositoryForComment) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepositoryForComment) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepositoryForComment) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepositoryForComment) List(ctx context.Context, orderBy string) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForComment) ListWithPagination(ctx context.Context, page, pageSize int, orderBy string) ([]models.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepositoryForComment) ListByUserType(ctx context.Context, userType models.UserType) ([]models.User, error) {
	return nil, nil
}

func (m *mockUserRepositoryForComment) Update(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForComment) UpdatePassword(ctx context.Context, id uint, password string) error {
	return nil
}

func (m *mockUserRepositoryForComment) UpdateEmail(ctx context.Context, id uint, email string) error {
	return nil
}

func (m *mockUserRepositoryForComment) UpdateStatus(ctx context.Context, id uint, status models.UserStatus) error {
	return nil
}

func (m *mockUserRepositoryForComment) UpdateUserType(ctx context.Context, id uint, userType models.UserType) error {
	return nil
}

func (m *mockUserRepositoryForComment) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForComment) ForceDelete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepositoryForComment) Upsert(ctx context.Context, user *models.User) error {
	return nil
}

func (m *mockUserRepositoryForComment) GetProfile(ctx context.Context, userID uint) (*models.UserProfile, error) {
	return nil, nil
}

func (m *mockUserRepositoryForComment) UpsertProfile(ctx context.Context, userID uint, profile *models.UserProfile) error {
	return nil
}

func (m *mockUserRepositoryForComment) WithTransaction(tx *gorm.DB) repository.UserRepository {
	return m
}

func setupCommentTestEnv() func() {
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	return func() {
		os.Unsetenv("JWT_SECRET")
	}
}

func createCommentHandler() (*CommentHandler, *mockCommentRepository, *mockWorkRepositoryForComment, *mockUserRepositoryForComment) {
	mockCommentRepo := newMockCommentRepository()
	mockWorkRepo := newMockWorkRepositoryForComment()
	mockUserRepo := newMockUserRepositoryForComment()
	commentSvc := svc.NewCommentService(mockCommentRepo, mockWorkRepo, mockUserRepo, slog.Default())
	h := NewCommentHandler(commentSvc)
	return h, mockCommentRepo, mockWorkRepo, mockUserRepo
}

// ================================================================================
// ListByWorkID - 获取作品评论列表
// ================================================================================

func TestCommentHandler_ListByWorkID_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1", []string{"work_id"}, []string{"1"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.NotNil(t, resp.Page)
}

func TestCommentHandler_ListByWorkID_InvalidWorkID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/abc", []string{"work_id"}, []string{"abc"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_ListByWorkID_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.listWithPagErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1", []string{"work_id"}, []string{"1"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestCommentHandler_ListByWorkID_PaginationBoundaries(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1?page=0", []string{"work_id"}, []string{"1"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 1, resp.Page.PageNum)
}

func TestCommentHandler_ListByWorkID_PageSizeBoundaries(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1?page_size=0", []string{"work_id"}, []string{"1"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 10, resp.Page.PageSize)
}

func TestCommentHandler_ListByWorkID_PageSizeExceedsMax(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1?page_size=200", []string{"work_id"}, []string{"1"}, "")

	err := h.ListByWorkID(c)
	assert.NoError(t, err)
	resp := assertSuccessResponse(t, rec, http.StatusOK)
	assert.Equal(t, 10, resp.Page.PageSize)
}

// ================================================================================
// ListRootByWorkID - 获取作品根评论
// ================================================================================

func TestCommentHandler_ListRootByWorkID_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1/root", []string{"work_id"}, []string{"1"}, "")

	err := h.ListRootByWorkID(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_ListRootByWorkID_InvalidWorkID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/xyz/root", []string{"work_id"}, []string{"xyz"}, "")

	err := h.ListRootByWorkID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_ListRootByWorkID_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.listRootByWorkIDErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("GET", "/comments/work/1/root", []string{"work_id"}, []string{"1"}, "")

	err := h.ListRootByWorkID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// ListByUserID - 获取用户评论
// ================================================================================

func TestCommentHandler_ListByUserID_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/user/1", []string{"user_id"}, []string{"1"}, "")

	err := h.ListByUserID(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_ListByUserID_InvalidUserID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("GET", "/comments/user/abc", []string{"user_id"}, []string{"abc"}, "")

	err := h.ListByUserID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_ListByUserID_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.listByUserIDErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("GET", "/comments/user/1", []string{"user_id"}, []string{"1"}, "")

	err := h.ListByUserID(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// GetComment - 获取评论详情
// ================================================================================

func TestCommentHandler_GetComment_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("GET", "/comments/1", []string{"id"}, []string{"1"}, "")

	err := h.GetComment(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_GetComment_InvalidID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("GET", "/comments/abc", []string{"id"}, []string{"abc"}, "")

	err := h.GetComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_GetComment_NotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("GET", "/comments/999", []string{"id"}, []string{"999"}, "")

	err := h.GetComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestCommentHandler_GetComment_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.getByIDErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("GET", "/comments/1", []string{"id"}, []string{"1"}, "")

	err := h.GetComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// CreateComment - 发表评论
// ================================================================================

func TestCommentHandler_CreateComment_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)

	body := `{"work_id":1,"content":"test comment"}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusCreated)
}

func TestCommentHandler_CreateComment_InvalidJSON(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{invalid json}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_CreateComment_WorkNotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"work_id":999,"content":"test comment"}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_CreateComment_ParentNotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)

	body := `{"work_id":1,"parent_id":999,"content":"reply to non-existent"}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_CreateComment_CannotReplyChild(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	rootComment := mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)
	replyComment := mockCommentRepo.addTestReply(rootComment.ID, rootComment.ID, 1, 1)

	body := `{"work_id":1,"parent_id":` + string(rune(replyComment.ID+'0')) + `,"content":"nested reply"}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_CreateComment_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, mockWorkRepo, _ := createCommentHandler()
	mockWorkRepo.addTestWork(models.WorkStatusPublished)
	mockCommentRepo.createErr = errors.New("database error")

	body := `{"work_id":1,"content":"test comment"}`
	_, c, rec := createEchoContextWithParams("POST", "/comments", nil, nil, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.CreateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// UpdateComment - 更新评论
// ================================================================================

func TestCommentHandler_UpdateComment_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{"content":"updated content"}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1", []string{"id"}, []string{"1"}, body)

	c.Set(middleware.ContextKeyUserID, uint(1))

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_UpdateComment_InvalidID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"content":"updated content"}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/abc", []string{"id"}, []string{"abc"}, body)

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_UpdateComment_InvalidJSON(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{invalid json}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1", []string{"id"}, []string{"1"}, body)

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_UpdateComment_NotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"content":"updated content"}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/999", []string{"id"}, []string{"999"}, body)

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestCommentHandler_UpdateComment_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)
	mockCommentRepo.updateErr = errors.New("database error")

	body := `{"content":"updated content"}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1", []string{"id"}, []string{"1"}, body)

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestCommentHandler_UpdateComment_EmptyBody(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1", []string{"id"}, []string{"1"}, body)

	err := h.UpdateComment(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

// ================================================================================
// DeleteComment - 删除评论
// ================================================================================

func TestCommentHandler_DeleteComment_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	_, c, rec := createEchoContextWithParams("DELETE", "/comments/1", []string{"id"}, []string{"1"}, "")

	err := h.DeleteComment(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_DeleteComment_InvalidID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("DELETE", "/comments/abc", []string{"id"}, []string{"abc"}, "")

	err := h.DeleteComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_DeleteComment_NotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	_, c, rec := createEchoContextWithParams("DELETE", "/comments/999", []string{"id"}, []string{"999"}, "")

	err := h.DeleteComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestCommentHandler_DeleteComment_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)
	mockCommentRepo.deleteErr = errors.New("database error")

	_, c, rec := createEchoContextWithParams("DELETE", "/comments/1", []string{"id"}, []string{"1"}, "")

	err := h.DeleteComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// UpdateStatus - 更新评论状态（管理员）
// ================================================================================

func TestCommentHandler_UpdateStatus_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{"status":0}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_UpdateStatus_InvalidID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"status":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/abc/status", []string{"id"}, []string{"abc"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_UpdateStatus_InvalidJSON(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{invalid json}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{"status":99}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_UpdateStatus_NotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"status":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/999/status", []string{"id"}, []string{"999"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestCommentHandler_UpdateStatus_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)
	mockCommentRepo.updateStatusErr = errors.New("database error")

	body := `{"status":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/status", []string{"id"}, []string{"1"}, body)

	err := h.UpdateStatus(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

// ================================================================================
// IncrementLikeCount - 更新评论点赞数
// ================================================================================

func TestCommentHandler_IncrementLikeCount_Success(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{"delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/like", []string{"id"}, []string{"1"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

func TestCommentHandler_IncrementLikeCount_InvalidID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/abc/like", []string{"id"}, []string{"abc"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_IncrementLikeCount_InvalidJSON(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{invalid json}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/like", []string{"id"}, []string{"1"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}

func TestCommentHandler_IncrementLikeCount_NotFound(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	body := `{"delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/999/like", []string{"id"}, []string{"999"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusNotFound, response.UserNotFound)
}

func TestCommentHandler_IncrementLikeCount_ServiceError(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)
	mockCommentRepo.incrementLikeCountErr = errors.New("database error")

	body := `{"delta":1}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/like", []string{"id"}, []string{"1"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusInternalServerError, response.InternalError)
}

func TestCommentHandler_IncrementLikeCount_NegativeDelta(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, mockCommentRepo, _, _ := createCommentHandler()
	mockCommentRepo.addTestComment(1, 1, models.CommentStatusActive)

	body := `{"delta":-5}`
	_, c, rec := createEchoContextWithParams("PUT", "/comments/1/like", []string{"id"}, []string{"1"}, body)

	err := h.IncrementLikeCount(c)
	assert.NoError(t, err)
	assertSuccessResponse(t, rec, http.StatusOK)
}

// ================================================================================
// ID Boundary Tests
// ================================================================================

func TestCommentHandler_GetComment_MaxUintID(t *testing.T) {
	restore := setupCommentTestEnv()
	defer restore()

	h, _, _, _ := createCommentHandler()

	maxUintStr := "18446744073709551615"
	_, c, rec := createEchoContextWithParams("GET", "/comments/"+maxUintStr, []string{"id"}, []string{maxUintStr}, "")

	err := h.GetComment(c)
	assert.NoError(t, err)
	assertErrorResponse(t, rec, http.StatusBadRequest, response.BadRequest)
}
