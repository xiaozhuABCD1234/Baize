package main

import (
	// "log/slog"
	// "net/http"
	"fmt"
	"os"

	// "github.com/labstack/echo/v5"
	// "github.com/labstack/echo/v5/middleware"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// e := echo.New()
	// e.Use(middleware.RequestLogger())

	// e.GET("/", func(c *echo.Context) error {
	// 	return c.String(http.StatusOK, "Hello, World!")
	// })

	// // 打开日志文件
	// f, err := os.OpenFile("log/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// // 创建 slog.Handler，直接指定输出到文件
	// handler := slog.NewJSONHandler(f, nil)
	// e.Logger = slog.New(handler)

	// if err := e.Start(":1323"); err != nil {
	// 	e.Logger.Error("failed to start server", "error", err)
	// }

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	_, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
