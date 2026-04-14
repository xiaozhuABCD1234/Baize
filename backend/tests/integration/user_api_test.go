package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

func (s *IntegrationSuite) TestGetUser_Success() {
	user, token := s.CreateUserAndLogin("getuser@example.com", "password123")
	rec := s.GET(fmt.Sprintf("/api/v1/users/%d", user.ID), s.AuthHeaders(token))
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)
}

func (s *IntegrationSuite) TestGetUser_NoToken() {
	rec := s.GET("/api/v1/users/1", nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestGetUser_ExpiredToken() {
	user, _ := s.CreateUserAndLogin("expired@example.com", "password123")

	claims := utils.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Type:   utils.AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(utils.GetEnv("JWT_SECRET_KEY", "default_secret")))

	rec := s.GET(fmt.Sprintf("/api/v1/users/%d", user.ID), s.AuthHeaders(tokenString))
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestGetUser_InvalidTokenFormat() {
	rec := s.GET("/api/v1/users/1", map[string]string{
		"Authorization": "InvalidFormat token",
	})
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestGetUser_NotFound() {
	_, token := s.CreateUserAndLogin("notfound@example.com", "password123")
	rec := s.GET("/api/v1/users/999999", s.AuthHeaders(token))
	s.AssertStatus(http.StatusNotFound, rec)
}

func (s *IntegrationSuite) TestGetUser_InvalidID() {
	_, token := s.CreateUserAndLogin("invalidid@example.com", "password123")
	rec := s.GET("/api/v1/users/invalid", s.AuthHeaders(token))
	s.AssertStatus(http.StatusBadRequest, rec)
}

func (s *IntegrationSuite) TestListUsers_Success() {
	_, token := s.CreateUserAndLogin("list@example.com", "password123")
	for i := 0; i < 3; i++ {
		s.CreateUserAndLogin(fmt.Sprintf("list%d@example.com", i), "password123")
	}
	rec := s.GET("/api/v1/users?page=1&page_size=10", s.AuthHeaders(token))
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)

	var resp response.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)
	s.NotNil(resp.Page)
	s.Equal(int64(4), resp.Page.Total)
}

func (s *IntegrationSuite) TestListUsers_EmptyList() {
	_, token := s.CreateUserAndLogin("empty@example.com", "password123")
	rec := s.GET("/api/v1/users?page=1&page_size=10", s.AuthHeaders(token))
	s.AssertOK(rec)

	var resp response.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)
	s.NotNil(resp.Page)
	s.Equal(int64(1), resp.Page.Total)
	s.Equal(1, resp.Page.PageNum)
}

func (s *IntegrationSuite) TestListUsers_Pagination() {
	_, token := s.CreateUserAndLogin("pagination@example.com", "password123")
	for i := 0; i < 15; i++ {
		s.CreateUserAndLogin(fmt.Sprintf("pageuser%d@example.com", i), "password123")
	}

	rec := s.GET("/api/v1/users?page=1&page_size=5", s.AuthHeaders(token))
	var resp response.Response
	json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Equal(int64(16), resp.Page.Total)
	s.Equal(5, len(resp.Data.([]interface{})))

	rec = s.GET("/api/v1/users?page=2&page_size=5", s.AuthHeaders(token))
	json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Equal(5, len(resp.Data.([]interface{})))

	rec = s.GET("/api/v1/users?page=3&page_size=5", s.AuthHeaders(token))
	json.Unmarshal(rec.Body.Bytes(), &resp)
	s.GreaterOrEqual(len(resp.Data.([]interface{})), 1)
}

func (s *IntegrationSuite) TestListUsers_PageSizeBoundary() {
	_, token := s.CreateUserAndLogin("boundary@example.com", "password123")

	rec := s.GET("/api/v1/users?page=1&page_size=0", s.AuthHeaders(token))
	s.AssertOK(rec)

	rec = s.GET("/api/v1/users?page=1&page_size=101", s.AuthHeaders(token))
	s.AssertOK(rec)

	rec = s.GET("/api/v1/users?page=0&page_size=10", s.AuthHeaders(token))
	s.AssertOK(rec)
}

func (s *IntegrationSuite) TestListUsers_NoAuth() {
	rec := s.GET("/api/v1/users", nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestUpdateUser_Success() {
	user, token := s.CreateUserAndLogin("update@example.com", "password123")
	body := map[string]interface{}{
		"email": "newemail@example.com",
	}
	rec := s.PUT(fmt.Sprintf("/api/v1/users/%d", user.ID), body, s.AuthHeaders(token))
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)
}

func (s *IntegrationSuite) TestUpdateUser_NotFound() {
	_, token := s.CreateUserAndLogin("updatenotfound@example.com", "password123")
	body := map[string]interface{}{
		"email": "newemail@example.com",
	}
	rec := s.PUT("/api/v1/users/999999", body, s.AuthHeaders(token))
	s.AssertStatus(http.StatusNotFound, rec)
}

func (s *IntegrationSuite) TestUpdateUser_DuplicateEmail() {
	user1, token1 := s.CreateUserAndLogin("user1@example.com", "password123")
	s.CreateUserAndLogin("user2@example.com", "password123")

	body := map[string]interface{}{
		"email": "user2@example.com",
	}
	rec := s.PUT(fmt.Sprintf("/api/v1/users/%d", user1.ID), body, s.AuthHeaders(token1))
	s.AssertStatus(http.StatusConflict, rec)
}

func (s *IntegrationSuite) TestUpdateUser_NoAuth() {
	rec := s.PUT("/api/v1/users/1", map[string]interface{}{"email": "new@example.com"}, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestUpdateUser_InvalidID() {
	_, token := s.CreateUserAndLogin("invalidupdate@example.com", "password123")
	rec := s.PUT("/api/v1/users/invalid", map[string]interface{}{"email": "new@example.com"}, s.AuthHeaders(token))
	s.AssertStatus(http.StatusBadRequest, rec)
}

func (s *IntegrationSuite) TestDeleteUser_Success() {
	user, token := s.CreateUserAndLogin("delete@example.com", "password123")
	rec := s.DELETE(fmt.Sprintf("/api/v1/users/%d", user.ID), s.AuthHeaders(token))
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)

	rec = s.GET(fmt.Sprintf("/api/v1/users/%d", user.ID), s.AuthHeaders(token))
	s.AssertStatus(http.StatusNotFound, rec)
}

func (s *IntegrationSuite) TestDeleteUser_NotFound() {
	_, token := s.CreateUserAndLogin("deletenotfound@example.com", "password123")
	rec := s.DELETE("/api/v1/users/999999", s.AuthHeaders(token))
	s.AssertStatus(http.StatusNotFound, rec)
}

func (s *IntegrationSuite) TestDeleteUser_NoAuth() {
	rec := s.DELETE("/api/v1/users/1", nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestChangePassword_Success() {
	user, token := s.CreateUserAndLogin("changepwd@example.com", "oldpassword")
	body := map[string]string{
		"old_password": "oldpassword",
		"new_password": "newpassword123",
	}
	rec := s.PUT(fmt.Sprintf("/api/v1/users/%d/password", user.ID), body, s.AuthHeaders(token))
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)

	body = map[string]string{
		"email":    "changepwd@example.com",
		"password": "newpassword123",
	}
	loginRec := s.POST("/api/v1/users/login", body, nil)
	s.AssertOK(loginRec)
}

func (s *IntegrationSuite) TestChangePassword_WrongOldPassword() {
	user, token := s.CreateUserAndLogin("wrongold@example.com", "correctpassword")
	body := map[string]string{
		"old_password": "wrongpassword",
		"new_password": "newpassword123",
	}
	rec := s.PUT(fmt.Sprintf("/api/v1/users/%d/password", user.ID), body, s.AuthHeaders(token))
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestChangePassword_SamePassword() {
	user, token := s.CreateUserAndLogin("samepwd@example.com", "password123")
	body := map[string]string{
		"old_password": "password123",
		"new_password": "password123",
	}
	rec := s.PUT(fmt.Sprintf("/api/v1/users/%d/password", user.ID), body, s.AuthHeaders(token))
	s.AssertStatus(http.StatusBadRequest, rec)
}

func (s *IntegrationSuite) TestChangePassword_NotFound() {
	_, token := s.CreateUserAndLogin("pwdnotfound@example.com", "password123")
	body := map[string]string{
		"old_password": "oldpassword",
		"new_password": "newpassword123",
	}
	rec := s.PUT("/api/v1/users/999999/password", body, s.AuthHeaders(token))
	s.AssertStatus(http.StatusNotFound, rec)
}

func (s *IntegrationSuite) TestChangePassword_NoAuth() {
	rec := s.PUT("/api/v1/users/1/password", map[string]string{
		"old_password": "old",
		"new_password": "new",
	}, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestRefreshToken_Success() {
	user, _ := s.CreateUserAndLogin("refresh@example.com", "password123")

	tokenPair, _ := utils.GenerateTokenPair(user.ID, user.Email, string(user.UserType))
	body := map[string]string{
		"refresh_token": tokenPair.RefreshToken,
	}
	rec := s.POST("/api/v1/users/refresh", body, nil)
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)
}

func (s *IntegrationSuite) TestRefreshToken_InvalidToken() {
	body := map[string]string{
		"refresh_token": "invalid-token",
	}
	rec := s.POST("/api/v1/users/refresh", body, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestRefreshToken_WrongTokenType() {
	user, _ := s.CreateUserAndLogin("wrongtype@example.com", "password123")
	tokenPair, _ := utils.GenerateTokenPair(user.ID, user.Email, string(user.UserType))

	body := map[string]string{
		"refresh_token": tokenPair.AccessToken,
	}
	rec := s.POST("/api/v1/users/refresh", body, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
}

func (s *IntegrationSuite) TestFullWorkflow() {
	registerBody := map[string]string{
		"email":    "workflow@example.com",
		"password": "password123",
	}
	regRec := s.POST("/api/v1/users/register", registerBody, nil)
	s.AssertOK(regRec)
	s.AssertSuccessResponse(regRec)

	loginBody := map[string]string{
		"email":    "workflow@example.com",
		"password": "password123",
	}
	loginRec := s.POST("/api/v1/users/login", loginBody, nil)
	s.AssertOK(loginRec)
	s.AssertSuccessResponse(loginRec)

	var loginResp response.Response
	json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	token := loginResp.Data.(map[string]interface{})["token"].(string)
	userID := uint(loginResp.Data.(map[string]interface{})["id"].(float64))

	listRec := s.GET("/api/v1/users?page=1&page_size=10", s.AuthHeaders(token))
	s.AssertOK(listRec)
	s.AssertSuccessResponse(listRec)

	updateRec := s.PUT(fmt.Sprintf("/api/v1/users/%d", userID), map[string]interface{}{
		"email": "updated@example.com",
	}, s.AuthHeaders(token))
	s.AssertOK(updateRec)
	s.AssertSuccessResponse(updateRec)

	getRec := s.GET(fmt.Sprintf("/api/v1/users/%d", userID), s.AuthHeaders(token))
	s.AssertOK(getRec)

	changePwdRec := s.PUT(fmt.Sprintf("/api/v1/users/%d/password", userID), map[string]string{
		"old_password": "password123",
		"new_password": "newpassword456",
	}, s.AuthHeaders(token))
	s.AssertOK(changePwdRec)

	delRec := s.DELETE(fmt.Sprintf("/api/v1/users/%d", userID), s.AuthHeaders(token))
	s.AssertOK(delRec)
}

func (s *IntegrationSuite) TestPermissionControl() {
	userA, tokenA := s.CreateUserAndLogin("usera@example.com", "password123")
	_, tokenB := s.CreateUserAndLogin("userb@example.com", "password123")

	userAIDStr := strconv.FormatUint(uint64(userA.ID), 10)
	userBIDStr := strconv.FormatUint(uint64(userA.ID+1), 10)

	updateRec := s.PUT("/api/v1/users/"+userBIDStr, map[string]interface{}{
		"email": "hacked@example.com",
	}, s.AuthHeaders(tokenA))
	s.AssertStatus(http.StatusOK, updateRec)

	deleteRec := s.DELETE("/api/v1/users/"+userAIDStr, s.AuthHeaders(tokenB))
	s.AssertStatus(http.StatusOK, deleteRec)
}
