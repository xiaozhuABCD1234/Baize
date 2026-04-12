package integration

import (
	"net/http"
)

func (s *IntegrationSuite) TestRegister_Success() {
	body := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	rec := s.POST("/api/v1/users/register", body, nil)
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)
}

func (s *IntegrationSuite) TestRegister_DuplicateEmail() {
	body := map[string]string{
		"email":    "duplicate@example.com",
		"password": "password123",
	}
	s.POST("/api/v1/users/register", body, nil)
	rec := s.POST("/api/v1/users/register", body, nil)
	s.AssertStatus(http.StatusConflict, rec)
	s.AssertFailResponse(rec)
}

func (s *IntegrationSuite) TestRegister_InvalidJSON() {
	rec := s.POST("/api/v1/users/register", "{invalid}", nil)
	s.AssertStatus(http.StatusBadRequest, rec)
}

func (s *IntegrationSuite) TestLogin_Success() {
	s.CreateUserAndLogin("login@example.com", "password123")
	body := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}
	rec := s.POST("/api/v1/users/login", body, nil)
	s.AssertOK(rec)
	s.AssertSuccessResponse(rec)
}

func (s *IntegrationSuite) TestLogin_WrongPassword() {
	s.CreateUserAndLogin("user@example.com", "correctpassword")
	body := map[string]string{
		"email":    "user@example.com",
		"password": "wrongpassword",
	}
	rec := s.POST("/api/v1/users/login", body, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
	s.AssertFailResponse(rec)
}

func (s *IntegrationSuite) TestLogin_UserNotFound() {
	body := map[string]string{
		"email":    "notfound@example.com",
		"password": "password123",
	}
	rec := s.POST("/api/v1/users/login", body, nil)
	s.AssertStatus(http.StatusUnauthorized, rec)
	s.AssertFailResponse(rec)
}

func (s *IntegrationSuite) TestLogin_InvalidJSON() {
	rec := s.POST("/api/v1/users/login", "{invalid}", nil)
	s.AssertStatus(http.StatusBadRequest, rec)
}
