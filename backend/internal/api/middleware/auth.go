package middleware

import (
	"net/http"
	"strings"

	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

const (
	ContextKeyUserID   = "user_id"
	ContextKeyEmail    = "email"
	ContextKeyUserType = "user_type"
)

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenInvalid, "缺少 Authorization header"))
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenInvalid, "Authorization header 格式错误"))
			}

			tokenString := parts[1]
			claims, err := utils.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenExpired, "无效或过期的 Token"))
			}

			if claims.Type != utils.AccessToken {
				return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenTypeInvalid, "无效的 Token 类型"))
			}

			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyEmail, claims.Email)
			c.Set(ContextKeyUserType, claims.UserType)

			return next(c)
		}
	}
}

func RequireAuth() echo.MiddlewareFunc {
	return JWTAuth()
}

func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userType, ok := c.Get(ContextKeyUserType).(string)
			if !ok || userType == "" {
				return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenInvalid, "用户未认证"))
			}

			for _, role := range roles {
				if userType == role {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, response.Fail(response.Forbidden, "无权访问此资源"))
		}
	}
}

func RequireAdmin() echo.MiddlewareFunc {
	return RequireRole("admin")
}

func GetUserID(c *echo.Context) uint {
	if id, ok := c.Get(ContextKeyUserID).(uint); ok {
		return id
	}
	return 0
}

func GetEmail(c *echo.Context) string {
	if email, ok := c.Get(ContextKeyEmail).(string); ok {
		return email
	}
	return ""
}

func GetUserType(c *echo.Context) string {
	if userType, ok := c.Get(ContextKeyUserType).(string); ok {
		return userType
	}
	return ""
}

func IsAdmin(c *echo.Context) bool {
	return GetUserType(c) == "admin"
}
