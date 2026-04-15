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

type CategoryHandler struct {
	svc svc.ICHCategoryService
}

func NewCategoryHandler(svc svc.ICHCategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) RegisterRoutes(g *echo.Group) {
	categories := g.Group("/categories")
	categories.GET("", h.ListCategories)
	categories.GET("/root", h.ListRoot)
	categories.GET("/parent/:parent_id", h.ListByParentID)
	categories.GET("/region/:region_code", h.ListByRegionCode)
	categories.GET("/active", h.ListActive)
	categories.GET("/:id", h.GetCategory)
	categories.GET("/:id/with-children", h.GetCategoryWithChildren)
	categories.GET("/name/:name", h.GetCategoryByName)

	categoriesAdmin := categories.Group("")
	categoriesAdmin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
	categoriesAdmin.POST("", h.CreateCategory)
	categoriesAdmin.PUT("/:id", h.UpdateCategory)
	categoriesAdmin.DELETE("/:id", h.DeleteCategory)
}

// ListCategories godoc
// @Summary     获取分类列表
// @Description 获取所有非遗分类列表
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       order_by query string false "排序字段"
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories [get]
func (h *CategoryHandler) ListCategories(c *echo.Context) error {
	orderBy := c.QueryParam("order_by")

	categories, err := h.svc.List(c.Request().Context(), orderBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(categories))
}

// ListRoot godoc
// @Summary     获取根分类
// @Description 获取一级分类（无父级的分类）
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/root [get]
func (h *CategoryHandler) ListRoot(c *echo.Context) error {
	categories, err := h.svc.ListRoot(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(categories))
}

// ListByParentID godoc
// @Summary     获取子分类
// @Description 根据父分类ID获取子分类列表
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       parent_id path int true "父分类ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的分类ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/parent/{parent_id} [get]
func (h *CategoryHandler) ListByParentID(c *echo.Context) error {
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	categories, err := h.svc.ListByParentID(c.Request().Context(), uint(parentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(categories))
}

// ListByRegionCode godoc
// @Summary     获取地区分类
// @Description 根据地区编码获取该地区的分类列表
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       region_code path string true "地区编码"
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/region/{region_code} [get]
func (h *CategoryHandler) ListByRegionCode(c *echo.Context) error {
	regionCode := c.Param("region_code")

	categories, err := h.svc.ListByRegionCode(c.Request().Context(), regionCode)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(categories))
}

// ListActive godoc
// @Summary     获取启用分类
// @Description 获取所有状态为启用的分类
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/active [get]
func (h *CategoryHandler) ListActive(c *echo.Context) error {
	categories, err := h.svc.ListActive(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(categories))
}

// GetCategory godoc
// @Summary     获取分类
// @Description 根据ID获取分类信息
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "分类ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的分类ID"
// @Failure     404 {object} response.Response "分类不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	category, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrICHCategoryNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "分类不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(category))
}

// GetCategoryWithChildren godoc
// @Summary     获取分类及子分类
// @Description 根据ID获取分类信息及其所有子分类
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "分类ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的分类ID"
// @Failure     404 {object} response.Response "分类不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/{id}/with-children [get]
func (h *CategoryHandler) GetCategoryWithChildren(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	category, err := h.svc.GetByIDWithChildren(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrICHCategoryNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "分类不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(category))
}

// GetCategoryByName godoc
// @Summary     根据名称获取分类
// @Description 根据分类名称获取分类信息
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Param       name path string true "分类名称"
// @Success     200 {object} response.Response "获取成功"
// @Failure     404 {object} response.Response "分类不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/name/{name} [get]
func (h *CategoryHandler) GetCategoryByName(c *echo.Context) error {
	name := c.Param("name")

	category, err := h.svc.GetByName(c.Request().Context(), name)
	if err != nil {
		if errors.Is(err, svc.ErrICHCategoryNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "分类不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(category))
}

// CreateCategory godoc
// @Summary     创建分类
// @Description 创建新的非遗分类（仅管理员可操作）
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.ICHCategory true "分类信息"
// @Success     201 {object} response.Response "创建成功"
// @Failure     400 {object} response.Response "请求参数错误或无效的分类级别"
// @Failure     409 {object} response.Response "分类名称已存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories [post]
func (h *CategoryHandler) CreateCategory(c *echo.Context) error {
	var category models.ICHCategory
	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Create(c.Request().Context(), &category)
	if err != nil {
		if errors.Is(err, svc.ErrCategoryNameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "分类名称已存在"))
		}
		if errors.Is(err, svc.ErrInvalidCategoryLevel) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类级别"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(result))
}

// UpdateCategory godoc
// @Summary     更新分类
// @Description 更新分类信息（仅管理员可操作）
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "分类ID"
// @Param       request body models.ICHCategory true "分类信息"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "分类不存在"
// @Failure     409 {object} response.Response "分类名称已存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	var category models.ICHCategory
	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Update(c.Request().Context(), uint(id), &category)
	if err != nil {
		if errors.Is(err, svc.ErrICHCategoryNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "分类不存在"))
		}
		if errors.Is(err, svc.ErrCategoryNameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "分类名称已存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(result))
}

// DeleteCategory godoc
// @Summary     删除分类
// @Description 删除分类（仅管理员可操作）
// @Tags        categories
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "分类ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "请求参数错误或无法删除有子节点的分类"
// @Failure     404 {object} response.Response "分类不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的分类 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrICHCategoryNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "分类不存在"))
		}
		if errors.Is(err, svc.ErrCannotDeleteCategory) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无法删除有子节点的分类"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}
