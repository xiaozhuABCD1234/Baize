package routes

import (
	"log/slog"
	"os"

	"backend/internal/api/handlers"
	"backend/internal/api/middleware"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func SetupRouter(userHandler *handlers.UserHandler) *echo.Echo {
	e := echo.New()
	e.Use(echomiddleware.RequestLogger())
	e.Use(echomiddleware.Recover())

	e.GET("/", func(c *echo.Context) error {
		return c.String(200, "Hello, World!")
	})

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

	f, err := os.OpenFile("log/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		e.Logger.Error("failed to open log file", "error", err)
	} else {
		handler := slog.NewJSONHandler(f, nil)
		e.Logger = slog.New(handler)
	}

	return e
}
