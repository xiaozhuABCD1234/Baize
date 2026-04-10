package main

import (
	"fmt"
	// "log/slog"
	"os"

	"backend/internal/api/handlers"
	"backend/internal/api/routes"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "backend/docs"
	"github.com/swaggo/echo-swagger/v2"
)

// @title           User Management API
// @version         1.0.0
// @description     用户管理API，提供用户注册、登录、信息管理等功能，基于JWT令牌认证。
// @termsOfService  https://example.com/terms/

// @contact.name   Baize Team
// @contact.url    https://github.com/your-org/baize
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:1323
// @BasePath  /api/v1
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入 "Bearer {token}" 进行JWT认证
func main() {
	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	db.AutoMigrate(&models.User{})

	userRepo := repository.NewUserRepository(db)
	userSvc := svc.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	_ = utils.GetEnv("JWT_SECRET_KEY", "default_secret")

	routes.SetupRouter(e, userHandler)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS("*"))

	f, err := os.OpenFile("log/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}
	defer f.Close()

	// handler := slog.NewJSONHandler(f, nil)
	// e.Logger = slog.New(handler)

	port := utils.GetEnv("PORT", "1323")

	e.Logger.Info("starting server", "port", port)
	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
