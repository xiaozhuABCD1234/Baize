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

type RegionHandler struct {
	svc svc.RegionService
}

func NewRegionHandler(svc svc.RegionService) *RegionHandler {
	return &RegionHandler{svc: svc}
}

func (h *RegionHandler) RegisterRoutes(g *echo.Group) {
	regions := g.Group("/regions")
	regions.GET("", h.ListRegions)
	regions.GET("/root", h.ListRoot)
	regions.GET("/parent/:parent_id", h.ListByParentID)
	regions.GET("/level/:level", h.ListByLevel)
	regions.GET("/heritage-centers", h.ListHeritageCenters)
	regions.GET("/:id", h.GetRegion)
	regions.GET("/:id/with-children", h.GetRegionWithChildren)
	regions.GET("/code/:code", h.GetRegionByCode)

	regionsAdmin := regions.Group("")
	regionsAdmin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
	regionsAdmin.POST("", h.CreateRegion)
	regionsAdmin.PUT("/:id", h.UpdateRegion)
	regionsAdmin.DELETE("/:id", h.DeleteRegion)
}

// ListRegions godoc
// @Summary     获取地区列表
// @Description 获取所有地区列表
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       order_by query string false "排序字段"
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions [get]
func (h *RegionHandler) ListRegions(c *echo.Context) error {
	orderBy := c.QueryParam("order_by")

	regions, err := h.svc.List(c.Request().Context(), orderBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(regions))
}

// ListRoot godoc
// @Summary     获取根地区
// @Description 获取一级地区（无父级的地区）
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/root [get]
func (h *RegionHandler) ListRoot(c *echo.Context) error {
	regions, err := h.svc.ListRoot(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(regions))
}

// ListByParentID godoc
// @Summary     获取子地区
// @Description 根据父地区ID获取子地区列表
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       parent_id path int true "父地区ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的地区ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/parent/{parent_id} [get]
func (h *RegionHandler) ListByParentID(c *echo.Context) error {
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区 ID"))
	}

	regions, err := h.svc.ListByParentID(c.Request().Context(), uint(parentID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(regions))
}

// ListByLevel godoc
// @Summary     获取指定级别地区
// @Description 根据级别获取地区列表
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       level path int true "地区级别"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的地区级别"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/level/{level} [get]
func (h *RegionHandler) ListByLevel(c *echo.Context) error {
	levelStr := c.Param("level")
	level, err := strconv.ParseInt(levelStr, 10, 8)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区级别"))
	}

	regions, err := h.svc.ListByLevel(c.Request().Context(), int8(level))
	if err != nil {
		if errors.Is(err, svc.ErrInvalidRegionLevel) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区级别"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(regions))
}

// ListHeritageCenters godoc
// @Summary     获取非遗保护中心
// @Description 获取所有非遗保护中心地区
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/heritage-centers [get]
func (h *RegionHandler) ListHeritageCenters(c *echo.Context) error {
	regions, err := h.svc.ListHeritageCenters(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(regions))
}

// GetRegion godoc
// @Summary     获取地区
// @Description 根据ID获取地区信息
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "地区ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的地区ID"
// @Failure     404 {object} response.Response "地区不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/{id} [get]
func (h *RegionHandler) GetRegion(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区 ID"))
	}

	region, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "地区不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(region))
}

// GetRegionWithChildren godoc
// @Summary     获取地区及子地区
// @Description 根据ID获取地区信息及其所有子地区
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "地区ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的地区ID"
// @Failure     404 {object} response.Response "地区不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/{id}/with-children [get]
func (h *RegionHandler) GetRegionWithChildren(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区 ID"))
	}

	region, err := h.svc.GetByIDWithChildren(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "地区不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(region))
}

// GetRegionByCode godoc
// @Summary     根据编码获取地区
// @Description 根据地区编码获取地区信息
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Param       code path string true "地区编码"
// @Success     200 {object} response.Response "获取成功"
// @Failure     404 {object} response.Response "地区不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/code/{code} [get]
func (h *RegionHandler) GetRegionByCode(c *echo.Context) error {
	code := c.Param("code")

	region, err := h.svc.GetByCode(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "地区不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(region))
}

// CreateRegion godoc
// @Summary     创建地区
// @Description 创建新的地区（仅管理员可操作）
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.Region true "地区信息"
// @Success     201 {object} response.Response "创建成功"
// @Failure     400 {object} response.Response "请求参数错误或无效的地区级别"
// @Failure     409 {object} response.Response "地区代码已存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions [post]
func (h *RegionHandler) CreateRegion(c *echo.Context) error {
	var region models.Region
	if err := c.Bind(&region); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Create(c.Request().Context(), &region)
	if err != nil {
		if errors.Is(err, svc.ErrRegionCodeExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "地区代码已存在"))
		}
		if errors.Is(err, svc.ErrInvalidRegionLevel) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区级别"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(result))
}

// UpdateRegion godoc
// @Summary     更新地区
// @Description 更新地区信息（仅管理员可操作）
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "地区ID"
// @Param       request body models.Region true "地区信息"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "地区不存在"
// @Failure     409 {object} response.Response "地区代码已存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/{id} [put]
func (h *RegionHandler) UpdateRegion(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区 ID"))
	}

	var region models.Region
	if err := c.Bind(&region); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	result, err := h.svc.Update(c.Request().Context(), uint(id), &region)
	if err != nil {
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "地区不存在"))
		}
		if errors.Is(err, svc.ErrRegionCodeExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "地区代码已存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(result))
}

// DeleteRegion godoc
// @Summary     删除地区
// @Description 删除地区（仅管理员可操作）
// @Tags        regions
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "地区ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "请求参数错误或无法删除有子节点的地区"
// @Failure     404 {object} response.Response "地区不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/regions/{id} [delete]
func (h *RegionHandler) DeleteRegion(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的地区 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "地区不存在"))
		}
		if errors.Is(err, svc.ErrCannotDeleteRegion) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无法删除有子节点的地区"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}
