package integration

import (
	"fmt"
	"net/http"
	"testing"

	"backend/internal/models"
	"backend/tests/integration/client"
	"backend/tests/integration/db"
	"backend/tests/integration/helpers"

	"github.com/stretchr/testify/assert"
)

func setupCommentTest(t *testing.T) (*client.APIClient, *models.User, *models.Work) {
	testDB := db.SetupTestDB(t)
	apiClient := client.NewAPIClient(testDB)
	user := client.CreateTestUser(t, testDB, "testuser", "test@example.com")
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Test Work")
	return apiClient, user, work
}

func TestComment_Create_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"work_id": work.ID,
		"content": "This is a test comment",
	}

	rec := apiClient.AuthenticatedRequest("POST", "/api/v1/comments", body, token)

	assert.Equal(t, http.StatusCreated, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_Create_Without_Auth(t *testing.T) {
	apiClient, _, work := setupCommentTest(t)

	body := map[string]interface{}{
		"work_id": work.ID,
		"content": "This is a test comment",
	}

	rec := apiClient.DoRequest("POST", "/api/v1/comments", body, 0)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestComment_ListByWorkID_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "First comment")
	client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "Second comment")

	rec := apiClient.DoRequest("GET", "/api/v1/comments/work/"+fmt.Sprintf("%d", work.ID), nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_ListByWorkID_Empty(t *testing.T) {
	apiClient, _, work := setupCommentTest(t)

	rec := apiClient.DoRequest("GET", "/api/v1/comments/work/"+fmt.Sprintf("%d", work.ID), nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_Update_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	comment := client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "Original comment")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"content": "Updated comment content",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/comments/"+fmt.Sprintf("%d", comment.ID), body, token)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_Update_NotOwner(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	comment := client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "Original comment")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"content": "Updated comment",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/comments/"+fmt.Sprintf("%d", comment.ID), body, token)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestComment_Update_NotFound(t *testing.T) {
	apiClient, user, _ := setupCommentTest(t)
	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"content": "Updated comment",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/comments/99999", body, token)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestComment_Delete_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	comment := client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "To be deleted")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	rec := apiClient.AuthenticatedRequest("DELETE", "/api/v1/comments/"+fmt.Sprintf("%d", comment.ID), nil, token)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestComment_NestedReply_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	parentComment := client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "Parent comment")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"work_id":   work.ID,
		"parent_id": parentComment.ID,
		"root_id":   parentComment.ID,
		"content":   "This is a reply",
	}

	rec := apiClient.AuthenticatedRequest("POST", "/api/v1/comments", body, token)

	assert.Equal(t, http.StatusCreated, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_GetByID_Success(t *testing.T) {
	apiClient, user, work := setupCommentTest(t)
	comment := client.CreateTestComment(t, apiClient.DB, work.ID, user.ID, "Test comment")

	rec := apiClient.DoRequest("GET", "/api/v1/comments/"+fmt.Sprintf("%d", comment.ID), nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestComment_GetByID_NotFound(t *testing.T) {
	apiClient, _, _ := setupCommentTest(t)

	rec := apiClient.DoRequest("GET", "/api/v1/comments/99999", nil, 0)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
