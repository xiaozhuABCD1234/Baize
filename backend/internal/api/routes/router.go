package routes

import (
	"backend/internal/api/handlers"
	"backend/internal/api/middleware"

	"github.com/labstack/echo/v5"
)

func SetupRouter(e *echo.Echo, userHandler *handlers.UserHandler) {
	api := e.Group("/api/v1")

	users := api.Group("/users")
	users.POST("/register", userHandler.Register)
	users.POST("/login", userHandler.Login)
	users.POST("/refresh", userHandler.RefreshToken)

	protected := users.Group("")
	protected.Use(middleware.JWTAuth())
	protected.GET("", userHandler.ListUsers)
	protected.GET("/:id", userHandler.GetUser)
	protected.PUT("/:id", userHandler.UpdateUser)
	protected.PUT("/:id/password", userHandler.ChangePassword)
	protected.DELETE("/:id", userHandler.DeleteUser)
}
