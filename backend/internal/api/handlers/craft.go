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

type CraftHandler struct {
	svc svc.CraftService
}

func NewCraftHandler(svc svc.CraftService) *CraftHandler {
	return &CraftHandler{svc: svc}
}

func (h *CraftHandler) RegisterRoutes(g *echo.Group) {
	crafts := g.Group("/crafts")
	crafts.GET("", h.ListCrafts)
	crafts.GET("/category/:category_id", h.ListByCategory)
	crafts.GET("/difficulty/:level", h.ListByDifficulty)
	crafts.GET("/:id", h.GetCraft)
	crafts.GET("/:id/with-category", h.GetCraftWithCategory)

	craftsAdmin := crafts.Group("")
	craftsAdmin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
	craftsAdmin.POST("", h.CreateCraft)
	craftsAdmin.PUT("/:id", h.UpdateCraft)
	craftsAdmin.DELETE("/:id", h.DeleteCraft)
}

// ListCrafts godoc
// @Summary     获取技艺列表
// @Description 获取所有技艺列表
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Param       order_by query string false "排序字段"
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts [get]
func (h *CraftHandler) ListCrafts(c *echo.Context) error {
	orderBy := c.QueryParam("order_by")

	crafts, err := h.svc.List(c.Request().Context(), orderBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(crafts))
}

// ListByCategory godoc
// @Summary     获取分类下的技艺
// @Description 根据分类ID获取该分类下的所有技艺
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Param       category_id path int true "分类ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的分类ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/category/{category_id} [get]
func (h *CraftHandler) ListByCategory(c *echo.Context) error {
	categoryIDStr := c.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	crafts, err := h.svc.ListByCategory(c.Request().Context(), uint(categoryID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(crafts))
}

// ListByDifficulty godoc
// @Summary     获取指定难度技艺
// @Description 根据难度等级获取技艺列表
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Param       level path int true "难度等级"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的难度等级"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/difficulty/{level} [get]
func (h *CraftHandler) ListByDifficulty(c *echo.Context) error {
	levelStr := c.Param("level")
	level, err := strconv.ParseInt(levelStr, 10, 8)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的难度等级"))
	}

	crafts, err := h.svc.ListByDifficulty(c.Request().Context(), int8(level))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidDifficulty) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的难度等级"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(crafts))
}

// GetCraft godoc
// @Summary     获取技艺
// @Description 根据ID获取技艺信息
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "技艺ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的技艺ID"
// @Failure     404 {object} response.Response "技艺不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/{id} [get]
func (h *CraftHandler) GetCraft(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的技艺 ID"))
	}

	craft, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrCraftNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "技艺不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(craft))
}

// GetCraftWithCategory godoc
// @Summary     获取技艺及分类
// @Description 根据ID获取技艺信息及其所属分类
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "技艺ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的技艺ID"
// @Failure     404 {object} response.Response "技艺不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/{id}/with-category [get]
func (h *CraftHandler) GetCraftWithCategory(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的技艺 ID"))
	}

	craft, err := h.svc.GetByIDWithCategory(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrCraftNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "技艺不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(craft))
}

// CreateCraft godoc
// @Summary     创建技艺
// @Description 创建新的技艺（仅管理员可操作）
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.Craft true "技艺信息"
// @Success     201 {object} response.Response "创建成功"
// @Failure     400 {object} response.Response "请求参数错误、技艺名称已存在、分类不存在或无效的难度等级"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts [post]
func (h *CraftHandler) CreateCraft(c *echo.Context) error {
	var craft models.Craft
	if err := c.Bind(&craft); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Create(c.Request().Context(), &craft)
	if err != nil {
		if errors.Is(err, svc.ErrCraftNameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "技艺名称已存在"))
		}
		if errors.Is(err, svc.ErrCategoryNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "分类不存在"))
		}
		if errors.Is(err, svc.ErrInvalidDifficulty) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的难度等级"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(result))
}

// UpdateCraft godoc
// @Summary     更新技艺
// @Description 更新技艺信息（仅管理员可操作）
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "技艺ID"
// @Param       request body models.Craft true "技艺信息"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "技艺不存在"
// @Failure     409 {object} response.Response "技艺名称已存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/{id} [put]
func (h *CraftHandler) UpdateCraft(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的技艺 ID"))
	}

	var craft models.Craft
	if err := c.Bind(&craft); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Update(c.Request().Context(), uint(id), &craft)
	if err != nil {
		if errors.Is(err, svc.ErrCraftNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "技艺不存在"))
		}
		if errors.Is(err, svc.ErrCraftNameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "技艺名称已存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(result))
}

// DeleteCraft godoc
// @Summary     删除技艺
// @Description 删除技艺（仅管理员可操作）
// @Tags        crafts
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "技艺ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "无效的技艺ID"
// @Failure     404 {object} response.Response "技艺不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/crafts/{id} [delete]
func (h *CraftHandler) DeleteCraft(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的技艺 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrCraftNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "技艺不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}
