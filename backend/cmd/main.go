package main

import (
	"fmt"
	"os"

	"backend/internal/api/handlers"
	"backend/internal/api/routes"
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"
	svc "backend/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "backend/docs"
	"github.com/swaggo/echo-swagger/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	db.AutoMigrate(&models.User{})

	userRepo := repository.NewUserRepository(db)
	userSvc := svc.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	routes.SetupRouter(e, userHandler)

	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS("*"))

	f, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}
	defer f.Close()

	port := cfg.App.Port

	e.Logger.Info("starting server", "port", port, "env", cfg.App.Env)
	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
