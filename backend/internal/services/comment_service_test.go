package service

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockCommentRepository struct {
	createErr          error
	getByIDErr         error
	getByIDWithUserErr error
	listErr            error
	listByWorkErr      error
	listRootByWorkErr  error
	listByUserErr      error
	listByParentErr    error
	listByRootErr      error
	listPagErr         error
	updateErr          error
	updateStatusErr    error
	incrementLikeErr   error
	deleteErr          error
	comments           map[uint]*models.Comment
	commentsByWork     map[uint][]*models.Comment
	nextID             uint
}

func newMockCommentRepository() *mockCommentRepository {
	return &mockCommentRepository{
		comments:       make(map[uint]*models.Comment),
		commentsByWork: make(map[uint][]*models.Comment),
		nextID:         1,
	}
}

func (m *mockCommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	if m.createErr != nil {
		return m.createErr
	}
	comment.ID = m.nextID
	comment.CreatedAt = time.Now()
	m.nextID++
	m.comments[comment.ID] = comment
	m.commentsByWork[comment.WorkID] = append(m.commentsByWork[comment.WorkID], comment)
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
	return m.GetByID(ctx, id)
}

func (m *mockCommentRepository) List(ctx context.Context, orderBy string) ([]models.Comment, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []models.Comment
	for _, c := range m.comments {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockCommentRepository) ListByWorkID(ctx context.Context, workID uint) ([]models.Comment, error) {
	if m.listByWorkErr != nil {
		return nil, m.listByWorkErr
	}
	var result []models.Comment
	for _, c := range m.commentsByWork[workID] {
		result = append(result, *c)
	}
	return result, nil
}

func (m *mockCommentRepository) ListRootByWorkID(ctx context.Context, workID uint) ([]models.Comment, error) {
	if m.listRootByWorkErr != nil {
		return nil, m.listRootByWorkErr
	}
	var result []models.Comment
	for _, c := range m.commentsByWork[workID] {
		if c.ParentID == 0 {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) ListByUserID(ctx context.Context, userID uint) ([]models.Comment, error) {
	if m.listByUserErr != nil {
		return nil, m.listByUserErr
	}
	var result []models.Comment
	for _, c := range m.comments {
		if c.UserID == userID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) ListByParentID(ctx context.Context, parentID uint) ([]models.Comment, error) {
	if m.listByParentErr != nil {
		return nil, m.listByParentErr
	}
	var result []models.Comment
	for _, c := range m.comments {
		if c.ParentID == parentID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) ListByRootID(ctx context.Context, rootID uint) ([]models.Comment, error) {
	if m.listByRootErr != nil {
		return nil, m.listByRootErr
	}
	var result []models.Comment
	for _, c := range m.comments {
		if c.RootID == rootID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) ListByStatus(ctx context.Context, status models.CommentStatus) ([]models.Comment, error) {
	return nil, nil
}

func (m *mockCommentRepository) ListWithPagination(ctx context.Context, workID uint, page, pageSize int) ([]models.Comment, int64, error) {
	if m.listPagErr != nil {
		return nil, 0, m.listPagErr
	}
	var result []models.Comment
	for _, c := range m.commentsByWork[workID] {
		result = append(result, *c)
	}
	return result, int64(len(result)), nil
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
	if m.incrementLikeErr != nil {
		return m.incrementLikeErr
	}
	if _, ok := m.comments[id]; !ok {
		return repository.ErrCommentNotFound
	}
	m.comments[id].LikeCount = uint(int(m.comments[id].LikeCount) + delta)
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
	return nil
}

func (m *mockCommentRepository) Upsert(ctx context.Context, comment *models.Comment) error {
	return nil
}

func (m *mockCommentRepository) WithTransaction(tx *gorm.DB) repository.CommentRepository {
	return m
}

func TestCommentService_Create_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Test comment",
	}
	resp, err := svc.Create(context.Background(), req, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test comment", resp.Content)
}

func TestCommentService_Create_ReplyToRoot(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Root comment",
	}
	rootResp, _ := svc.Create(context.Background(), req, 1)

	replyReq := &models.CreateCommentRequest{
		WorkID:   1,
		ParentID: rootResp.ID,
		Content:  "Reply comment",
	}
	replyResp, err := svc.Create(context.Background(), replyReq, 2)

	assert.NoError(t, err)
	assert.NotNil(t, replyResp)
	assert.Equal(t, rootResp.ID, replyResp.ParentID)
	assert.Equal(t, rootResp.ID, replyResp.RootID)
}

func TestCommentService_Create_CannotReplyToChild(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Root comment",
	}
	rootResp, _ := svc.Create(context.Background(), req, 1)

	replyReq := &models.CreateCommentRequest{
		WorkID:   1,
		ParentID: rootResp.ID,
		Content:  "Reply comment",
	}
	replyResp, _ := svc.Create(context.Background(), replyReq, 2)

	replyToReplyReq := &models.CreateCommentRequest{
		WorkID:   1,
		ParentID: replyResp.ID,
		Content:  "Should fail",
	}
	_, err := svc.Create(context.Background(), replyToReplyReq, 3)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能回复二级及以下的评论")
}

func TestCommentService_GetByID_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Test comment",
	}
	createResp, _ := svc.Create(context.Background(), req, 1)

	getResp, err := svc.GetByID(context.Background(), createResp.ID)

	assert.NoError(t, err)
	assert.NotNil(t, getResp)
	assert.Equal(t, createResp.ID, getResp.ID)
}

func TestCommentService_GetByID_NotFound(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	_, err := svc.GetByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "评论不存在")
}

func TestCommentService_Update_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Original content",
	}
	createResp, _ := svc.Create(context.Background(), req, 1)

	updateResp, err := svc.Update(context.Background(), createResp.ID, "Updated content")

	assert.NoError(t, err)
	assert.NotNil(t, updateResp)
	assert.Equal(t, "Updated content", updateResp.Content)
}

func TestCommentService_Delete_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Test comment",
	}
	createResp, _ := svc.Create(context.Background(), req, 1)

	err := svc.Delete(context.Background(), createResp.ID)

	assert.NoError(t, err)
}

func TestCommentService_ListByWorkID_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	_, _ = svc.Create(context.Background(), &models.CreateCommentRequest{WorkID: 1, Content: "Comment 1"}, 1)
	_, _ = svc.Create(context.Background(), &models.CreateCommentRequest{WorkID: 1, Content: "Comment 2"}, 1)

	list, total, err := svc.ListByWorkID(context.Background(), 1, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)
}

func TestCommentService_ListRootByWorkID_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	root1, _ := svc.Create(context.Background(), &models.CreateCommentRequest{WorkID: 1, Content: "Root 1"}, 1)
	_, _ = svc.Create(context.Background(), &models.CreateCommentRequest{WorkID: 1, ParentID: root1.ID, Content: "Reply 1"}, 2)
	_, _ = svc.Create(context.Background(), &models.CreateCommentRequest{WorkID: 1, Content: "Root 2"}, 1)

	list, err := svc.ListRootByWorkID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCommentService_UpdateStatus_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Test comment",
	}
	createResp, _ := svc.Create(context.Background(), req, 1)

	err := svc.UpdateStatus(context.Background(), createResp.ID, models.CommentStatusDeleted)

	assert.NoError(t, err)
}

func TestCommentService_IncrementLikeCount_Success(t *testing.T) {
	commentRepo := newMockCommentRepository()
	workRepo := newMockWorkRepository()
	workRepo.addTestWork(1)

	logger := slog.Default()
	svc := NewCommentService(commentRepo, workRepo, nil, logger)

	req := &models.CreateCommentRequest{
		WorkID:  1,
		Content: "Test comment",
	}
	createResp, _ := svc.Create(context.Background(), req, 1)

	err := svc.IncrementLikeCount(context.Background(), createResp.ID, 1)

	assert.NoError(t, err)
}
