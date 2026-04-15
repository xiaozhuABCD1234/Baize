package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"backend/internal/api/middleware"
	"backend/internal/models"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
)

type FavoriteHandler struct {
	svc svc.FavoriteService
}

func NewFavoriteHandler(svc svc.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{svc: svc}
}

func (h *FavoriteHandler) RegisterRoutes(g *echo.Group) {
	favorites := g.Group("/favorites")
	favorites.GET("/work/:work_id", h.ListByWorkID)
	favorites.GET("/check/:work_id", h.CheckExists)
	favorites.GET("/:id", h.GetFavorite)
	favorites.GET("/user/:user_id", h.ListByUserID)

	favoritesAuth := favorites.Group("")
	favoritesAuth.Use(middleware.JWTAuth())
	favoritesAuth.POST("", h.CreateFavorite)
	favoritesAuth.DELETE("/:id", h.DeleteFavorite)
	favoritesAuth.DELETE("/work/:work_id", h.DeleteByWork)
	favoritesAuth.PUT("/:id/folder", h.UpdateFolder)
}

// ListByWorkID godoc
// @Summary     获取作品收藏列表
// @Description 获取收藏了指定作品的用户列表
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Param       work_id path int true "作品ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/work/{work_id} [get]
func (h *FavoriteHandler) ListByWorkID(c *echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := strconv.ParseUint(workIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	favorites, err := h.svc.ListByWorkID(c.Request().Context(), uint(workID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(favorites))
}

// CheckExists godoc
// @Summary     检查收藏状态
// @Description 检查当前用户是否收藏了指定作品
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       work_id path int true "作品ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/check/{work_id} [get]
func (h *FavoriteHandler) CheckExists(c *echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := strconv.ParseUint(workIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	userID := middleware.GetUserID(c)

	exists, err := h.svc.Exists(c.Request().Context(), userID, uint(workID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"favorited": exists}))
}

// GetFavorite godoc
// @Summary     获取收藏
// @Description 根据ID获取收藏详情
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "收藏ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的收藏ID"
// @Failure     404 {object} response.Response "收藏不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/{id} [get]
func (h *FavoriteHandler) GetFavorite(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的收藏 ID"))
	}

	favorite, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrFavoriteNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "收藏不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(favorite))
}

// ListByUserID godoc
// @Summary     获取用户收藏列表
// @Description 获取指定用户的收藏列表（带分页）
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Param       page query int false "页码" default(1)
// @Param       page_size query int false "每页数量" default(10)
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/user/{user_id} [get]
func (h *FavoriteHandler) ListByUserID(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	favorites, total, err := h.svc.ListByUserID(c.Request().Context(), uint(userID), page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.SuccessWithPage(favorites, page, pageSize, total))
}

// CreateFavorite godoc
// @Summary     收藏作品
// @Description 收藏指定作品
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.FavoriteRequest true "收藏请求"
// @Success     201 {object} response.Response "收藏成功"
// @Failure     400 {object} response.Response "请求参数错误或作品不存在"
// @Failure     409 {object} response.Response "已经收藏过该作品"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites [post]
func (h *FavoriteHandler) CreateFavorite(c *echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.FavoriteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	favorite, err := h.svc.Create(c.Request().Context(), &req, userID)
	if err != nil {
		if errors.Is(err, svc.ErrAlreadyFavorited) {
			return c.JSON(http.StatusConflict, response.Fail(response.ResourceConflict, "已经收藏过该作品"))
		}
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "作品不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(favorite))
}

// DeleteFavorite godoc
// @Summary     取消收藏
// @Description 取消收藏指定作品
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "收藏ID"
// @Success     200 {object} response.Response "取消收藏成功"
// @Failure     400 {object} response.Response "无效的收藏ID"
// @Failure     404 {object} response.Response "收藏不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/{id} [delete]
func (h *FavoriteHandler) DeleteFavorite(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的收藏 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrFavoriteNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "收藏不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "取消收藏成功"}))
}

// DeleteByWork godoc
// @Summary     取消收藏作品
// @Description 根据作品ID取消收藏
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       work_id path int true "作品ID"
// @Success     200 {object} response.Response "取消收藏成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     404 {object} response.Response "收藏不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/work/{work_id} [delete]
func (h *FavoriteHandler) DeleteByWork(c *echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := strconv.ParseUint(workIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	userID := middleware.GetUserID(c)

	if err := h.svc.DeleteByUserAndWork(c.Request().Context(), userID, uint(workID)); err != nil {
		if errors.Is(err, svc.ErrFavoriteNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "收藏不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "取消收藏成功"}))
}

// UpdateFolder godoc
// @Summary     移动收藏
// @Description 将收藏移动到指定文件夹
// @Tags        favorites
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "收藏ID"
// @Param       request body map[string]interface{} true "文件夹ID"
// @Success     200 {object} response.Response "移动成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "收藏不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/favorites/{id}/folder [put]
func (h *FavoriteHandler) UpdateFolder(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的收藏 ID"))
	}

	var req struct {
		FolderID uint `json:"folder_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if err := h.svc.UpdateFolder(c.Request().Context(), uint(id), req.FolderID); err != nil {
		if errors.Is(err, svc.ErrFavoriteNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "收藏不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "移动成功"}))
}
