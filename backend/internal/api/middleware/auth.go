package middleware

import (
	"net/http"
	"strings"

	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

type AuthResponse struct {
	Message string `json:"message"`
}

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, AuthResponse{
					Message: "缺少 Authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return c.JSON(http.StatusUnauthorized, AuthResponse{
					Message: "Authorization header 格式错误",
				})
			}

			tokenString := parts[1]
			claims, err := utils.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, AuthResponse{
					Message: "无效或过期的 Token",
				})
			}

			if claims.Type != utils.AccessToken {
				return c.JSON(http.StatusUnauthorized, AuthResponse{
					Message: "无效的 Token 类型",
				})
			}

			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}
