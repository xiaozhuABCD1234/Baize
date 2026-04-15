package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRegister_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	body := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rec := rh.post("/api/v1/users/register", body)

	assert.Equal(t, http.StatusOK, rec.Code)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data := resp.Data.(map[string]interface{})

	assert.NotEmpty(t, data["id"])
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "testuser", data["username"])
}

func TestUserRegister_DuplicateEmail(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	body := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", body)

	body = `{"email":"test@example.com","username":"anotheruser","password":"password456"}`
	rec := rh.post("/api/v1/users/register", body)

	assertErrorResponse(t, rec, http.StatusConflict, "USER_002")
}

func TestUserRegister_DuplicateUsername(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	body := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", body)

	body = `{"email":"another@example.com","username":"testuser","password":"password456"}`
	rec := rh.post("/api/v1/users/register", body)

	assertErrorResponse(t, rec, http.StatusConflict, "USER_005")
}

func TestUserRegister_InvalidJSON(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	body := `{invalid json}`
	rec := rh.post("/api/v1/users/register", body)

	assertErrorResponse(t, rec, http.StatusBadRequest, "COMMON_001")
}

func TestUserLogin_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"password123"}`
	rec := rh.post("/api/v1/users/login", loginBody)

	assert.Equal(t, http.StatusOK, rec.Code)

	resp := assertSuccessResponse(t, rec, http.StatusOK)
	data := resp.Data.(map[string]interface{})

	assert.NotEmpty(t, data["token"])
}

func TestUserLogin_InvalidEmail(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	body := `{"email":"nonexistent@example.com","password":"password123"}`
	rec := rh.post("/api/v1/users/login", body)

	assertErrorResponse(t, rec, http.StatusUnauthorized, "USER_001")
}

func TestUserLogin_InvalidPassword(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"wrongpassword"}`
	rec := rh.post("/api/v1/users/login", loginBody)

	assertErrorResponse(t, rec, http.StatusUnauthorized, "USER_003")
}

func TestCreateWork_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	craft := createTestCraft(t, server.DB, "苏绣")
	region := createTestRegion(t, server.DB, "北京市")

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"password123"}`
	loginRec := rh.post("/api/v1/users/login", loginBody)
	loginResp := assertSuccessResponse(t, loginRec, http.StatusOK)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["token"].(string)
	rh.setAuthToken(token)

	workBody := `{"title":"测试作品","content":"这是测试内容","content_type":1,"craft_id":` + itoa(int(craft.ID)) + `,"region_id":` + itoa(int(region.ID)) + `,"media":[{"url":"https://example.com/image.jpg","media_type":1}]}`
	rec := rh.post("/api/v1/works", workBody)

	assert.Equal(t, http.StatusCreated, rec.Code)

	resp := assertSuccessResponse(t, rec, http.StatusCreated)
	workData := resp.Data.(map[string]interface{})
	assert.NotEmpty(t, workData["id"])
	assert.Equal(t, "测试作品", workData["title"])
}

func TestCreateWork_Unauthorized(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	craft := createTestCraft(t, server.DB, "苏绣")
	region := createTestRegion(t, server.DB, "北京市")

	workBody := `{"title":"测试作品","content":"这是测试内容","content_type":1,"craft_id":` + itoa(int(craft.ID)) + `,"region_id":` + itoa(int(region.ID)) + `,"media":[{"url":"https://example.com/image.jpg","media_type":1}]}`
	rec := rh.post("/api/v1/works", workBody)

	assertErrorResponse(t, rec, http.StatusUnauthorized, "AUTH_001")
}

func TestDeleteWork_NotFound(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"password123"}`
	loginRec := rh.post("/api/v1/users/login", loginBody)
	loginResp := assertSuccessResponse(t, loginRec, http.StatusOK)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["token"].(string)
	rh.setAuthToken(token)

	rec := rh.delete("/api/v1/works/99999")
	assertErrorResponse(t, rec, http.StatusNotFound, "WORK_001")
}

func TestGetWork_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	craft := createTestCraft(t, server.DB, "苏绣")
	region := createTestRegion(t, server.DB, "北京市")

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"password123"}`
	loginRec := rh.post("/api/v1/users/login", loginBody)
	loginResp := assertSuccessResponse(t, loginRec, http.StatusOK)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["token"].(string)
	rh.setAuthToken(token)

	workBody := `{"title":"测试作品","content":"这是测试内容","content_type":1,"craft_id":` + itoa(int(craft.ID)) + `,"region_id":` + itoa(int(region.ID)) + `,"media":[{"url":"https://example.com/image.jpg","media_type":1}]}`
	createRec := rh.post("/api/v1/works", workBody)
	createResp := assertSuccessResponse(t, createRec, http.StatusCreated)
	workData := createResp.Data.(map[string]interface{})
	workID := itoa(int(workData["id"].(float64)))

	rh.setAuthToken("")
	getRec := rh.get("/api/v1/works/" + workID)
	assert.Equal(t, http.StatusOK, getRec.Code)

	getResp := assertSuccessResponse(t, getRec, http.StatusOK)
	getData := getResp.Data.(map[string]interface{})
	assert.Equal(t, "测试作品", getData["title"])
}

func TestListWorks_Success(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	craft := createTestCraft(t, server.DB, "苏绣")
	region := createTestRegion(t, server.DB, "北京市")

	registerBody := `{"email":"test@example.com","username":"testuser","password":"password123"}`
	rh.post("/api/v1/users/register", registerBody)

	loginBody := `{"email":"test@example.com","password":"password123"}`
	loginRec := rh.post("/api/v1/users/login", loginBody)
	loginResp := assertSuccessResponse(t, loginRec, http.StatusOK)
	loginData := loginResp.Data.(map[string]interface{})
	token := loginData["token"].(string)
	rh.setAuthToken(token)

	workBody := `{"title":"测试作品1","content":"内容1","content_type":1,"craft_id":` + itoa(int(craft.ID)) + `,"region_id":` + itoa(int(region.ID)) + `,"media":[{"url":"https://example.com/image1.jpg","media_type":1}]}`
	rh.post("/api/v1/works", workBody)

	workBody = `{"title":"测试作品2","content":"内容2","content_type":1,"craft_id":` + itoa(int(craft.ID)) + `,"region_id":` + itoa(int(region.ID)) + `,"media":[{"url":"https://example.com/image2.jpg","media_type":1}]}`
	rh.post("/api/v1/works", workBody)

	rh.setAuthToken("")
	rec := rh.get("/api/v1/works")
	assert.Equal(t, http.StatusOK, rec.Code)

	listResp := assertSuccessResponse(t, rec, http.StatusOK)
	listData := listResp.Data.([]interface{})
	assert.GreaterOrEqual(t, len(listData), 2)
}

func TestHealthCheck(t *testing.T) {
	server := setupTestServer(t)
	defer server.close()

	rh := newRequestHelper(server)

	rec := rh.get("/health")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	digits := ""
	for i > 0 {
		digits = string(byte('0'+i%10)) + digits
		i /= 10
	}
	return digits
}
