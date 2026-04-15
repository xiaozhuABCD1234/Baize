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

func setupWorkTest(t *testing.T) (*client.APIClient, *models.User) {
	testDB := db.SetupTestDB(t)
	apiClient := client.NewAPIClient(testDB)
	user := client.CreateTestUser(t, testDB, "testuser", "test@example.com")
	return apiClient, user
}

func TestWork_Create_Success(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	craft := client.CreateTestCraft(t, apiClient.DB, "Test Craft")
	region := client.CreateTestRegion(t, apiClient.DB, "Test Region")
	category := client.CreateTestCategory(t, apiClient.DB, "Test Category")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"title":          "My Test Work",
		"content":        "This is test content",
		"content_type":   1,
		"craft_id":       craft.ID,
		"region_id":      region.ID,
		"category_id":    category.ID,
		"technique_tags": []string{"tag1", "tag2"},
		"materials":      []string{"material1"},
		"media": []map[string]interface{}{
			{"url": "https://example.com/image.jpg", "media_type": 1},
		},
	}

	rec := apiClient.AuthenticatedRequest("POST", "/api/v1/works", body, token)

	assert.Equal(t, http.StatusCreated, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Data)
}

func TestWork_Create_Without_Auth(t *testing.T) {
	apiClient, _ := setupWorkTest(t)
	craft := client.CreateTestCraft(t, apiClient.DB, "Test Craft")

	body := map[string]interface{}{
		"title":        "My Test Work",
		"content":      "This is test content",
		"content_type": 1,
		"craft_id":     craft.ID,
	}

	rec := apiClient.DoRequest("POST", "/api/v1/works", body, 0)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWork_GetByID_Success(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Test Work")

	rec := apiClient.DoRequest("GET", "/api/v1/works/"+fmt.Sprintf("%d", work.ID), nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestWork_GetByID_NotFound(t *testing.T) {
	apiClient, _ := setupWorkTest(t)

	rec := apiClient.DoRequest("GET", "/api/v1/works/99999", nil, 0)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWork_List_Success(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	for i := 0; i < 5; i++ {
		client.CreateTestWork(t, apiClient.DB, user.ID, "Test Work "+string(rune('A'+i)))
	}

	rec := apiClient.DoRequest("GET", "/api/v1/works?page=1&page_size=10", nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
	assert.NotNil(t, resp.Page)
	assert.Equal(t, int64(5), resp.Page.Total)
}

func TestWork_List_Pagination(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	for i := 0; i < 15; i++ {
		client.CreateTestWork(t, apiClient.DB, user.ID, "Test Work "+string(rune('A'+i)))
	}

	rec := apiClient.DoRequest("GET", "/api/v1/works?page=1&page_size=5", nil, 0)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Equal(t, int64(15), resp.Page.Total)
	assert.Equal(t, 5, resp.Page.PageSize)
}

func TestWork_Update_Success(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Original Title")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated content",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/works/"+fmt.Sprintf("%d", work.ID), body, token)

	assert.Equal(t, http.StatusOK, rec.Code)
	resp := client.ParseResponse(t, rec)
	assert.Nil(t, resp.Error)
}

func TestWork_Update_NotOwner(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Test Work")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"title": "Updated Title",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/works/"+fmt.Sprintf("%d", work.ID), body, token)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWork_Update_NotFound(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	body := map[string]interface{}{
		"title": "Updated Title",
	}

	rec := apiClient.AuthenticatedRequest("PUT", "/api/v1/works/99999", body, token)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestWork_Delete_Success(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Draft Work")
	work.Status = models.WorkStatusDraft
	apiClient.DB.Save(work)

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	rec := apiClient.AuthenticatedRequest("DELETE", "/api/v1/works/"+fmt.Sprintf("%d", work.ID), nil, token)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWork_Delete_PublishedWork(t *testing.T) {
	apiClient, user := setupWorkTest(t)
	work := client.CreateTestWork(t, apiClient.DB, user.ID, "Published Work")

	token := helpers.GenerateTestToken(t, user.ID, user.Email, string(user.UserType))

	rec := apiClient.AuthenticatedRequest("DELETE", "/api/v1/works/"+fmt.Sprintf("%d", work.ID), nil, token)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
