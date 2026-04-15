package routes

import (
	"log/slog"
	"net/http"

	"backend/internal/api/handlers"
	svc "backend/internal/services"

	"github.com/labstack/echo/v5"
)

type HandlerDeps struct {
	UserService     *svc.UserService
	WorkService     svc.WorkService
	CommentService  svc.CommentService
	FavoriteService svc.FavoriteService
	FollowService   svc.FollowService
	CraftService    svc.CraftService
	RegionService   svc.RegionService
	CategoryService svc.ICHCategoryService
	Logger          *slog.Logger
}

func SetupRouter(e *echo.Echo, deps HandlerDeps) {
	api := e.Group("/api/v1")

	userHandler := handlers.NewUserHandler(deps.UserService)
	authHandler := handlers.NewAuthHandler()
	workHandler := handlers.NewWorkHandler(deps.WorkService)
	commentHandler := handlers.NewCommentHandler(deps.CommentService)
	favoriteHandler := handlers.NewFavoriteHandler(deps.FavoriteService)
	followHandler := handlers.NewFollowHandler(deps.FollowService)
	craftHandler := handlers.NewCraftHandler(deps.CraftService)
	regionHandler := handlers.NewRegionHandler(deps.RegionService)
	categoryHandler := handlers.NewCategoryHandler(deps.CategoryService)

	userHandler.RegisterRoutes(api)
	authHandler.RegisterRoutes(api)
	workHandler.RegisterRoutes(api)
	commentHandler.RegisterRoutes(api)
	favoriteHandler.RegisterRoutes(api)
	followHandler.RegisterRoutes(api)
	craftHandler.RegisterRoutes(api)
	regionHandler.RegisterRoutes(api)
	categoryHandler.RegisterRoutes(api)

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
}
