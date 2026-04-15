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

type WorkHandler struct {
	svc svc.WorkService
}

func NewWorkHandler(svc svc.WorkService) *WorkHandler {
	return &WorkHandler{svc: svc}
}

func (h *WorkHandler) RegisterRoutes(g *echo.Group) {
	works := g.Group("/works")
	works.GET("", h.ListWorks)
	works.GET("/top", h.ListTopWorks)
	works.GET("/recommended", h.ListRecommendedWorks)
	works.GET("/:id", h.GetWork)
	works.GET("/:id/detailed", h.GetWorkDetailed)

	worksAuth := works.Group("")
	worksAuth.Use(middleware.JWTAuth())
	worksAuth.POST("", h.CreateWork)
	worksAuth.PUT("/:id", h.UpdateWork)
	worksAuth.DELETE("/:id", h.DeleteWork)
	worksAuth.PUT("/:id/status", h.UpdateStatus, middleware.RequireRole("admin"))
	worksAuth.PUT("/:id/count", h.IncrementCount)
}

// ListWorks godoc
// @Summary     获取作品列表
// @Description 分页获取作品列表，支持筛选和排序
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Param       user_id query int false "用户ID筛选"
// @Param       craft_id query int false "技艺ID筛选"
// @Param       region_id query int false "地区ID筛选"
// @Param       is_master query bool false "是否只看大师作品"
// @Param       order_by query string false "排序方式: newest, hot, weight" default(newest)
// @Param       page query int false "页码" default(1)
// @Param       page_size query int false "每页数量" default(10)
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works [get]
func (h *WorkHandler) ListWorks(c *echo.Context) error {
	var req models.WorkListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	works, total, err := h.svc.List(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.SuccessWithPage(works, req.Page, req.PageSize, total))
}

// ListTopWorks godoc
// @Summary     获取精选作品
// @Description 获取置顶的优秀作品列表
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Param       limit query int false "返回数量" default(10) minimum(1) maximum(50)
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/top [get]
func (h *WorkHandler) ListTopWorks(c *echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	works, err := h.svc.ListTop(c.Request().Context(), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(works))
}

// ListRecommendedWorks godoc
// @Summary     获取推荐作品
// @Description 获取推荐的作品列表
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Param       limit query int false "返回数量" default(10) minimum(1) maximum(50)
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/recommended [get]
func (h *WorkHandler) ListRecommendedWorks(c *echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	works, err := h.svc.ListRecommended(c.Request().Context(), limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(works))
}

// GetWork godoc
// @Summary     获取作品
// @Description 根据ID获取作品信息
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "作品ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id} [get]
func (h *WorkHandler) GetWork(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	work, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(work))
}

// GetWorkDetailed godoc
// @Summary     获取作品详情
// @Description 根据ID获取作品详细信息（包括媒体资源）
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "作品ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id}/detailed [get]
func (h *WorkHandler) GetWorkDetailed(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	work, err := h.svc.GetByIDDetailed(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(work))
}

// CreateWork godoc
// @Summary     创建作品
// @Description 发布新作品或动态
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.CreateWorkRequest true "创建作品请求"
// @Success     201 {object} response.Response "创建成功"
// @Failure     400 {object} response.Response "请求参数错误、技艺不存在、分类不存在或地区不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works [post]
func (h *WorkHandler) CreateWork(c *echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.CreateWorkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	work, err := h.svc.Create(c.Request().Context(), &req, userID)
	if err != nil {
		if errors.Is(err, svc.ErrCraftNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "技艺不存在"))
		}
		if errors.Is(err, svc.ErrCategoryNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "分类不存在"))
		}
		if errors.Is(err, svc.ErrRegionNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "地区不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(work))
}

// UpdateWork godoc
// @Summary     更新作品
// @Description 更新作品信息
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "作品ID"
// @Param       request body models.CreateWorkRequest true "更新作品请求"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id} [put]
func (h *WorkHandler) UpdateWork(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	var req models.CreateWorkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	work, err := h.svc.Update(c.Request().Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(work))
}

// DeleteWork godoc
// @Summary     删除作品
// @Description 删除作品
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "作品ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "请求参数错误或无法删除已发布的作品"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id} [delete]
func (h *WorkHandler) DeleteWork(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		if errors.Is(err, svc.ErrCannotDeleteWork) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无法删除已发布的作品"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}

// UpdateStatus godoc
// @Summary     更新作品状态
// @Description 更新作品审核状态（仅管理员可操作）
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "作品ID"
// @Param       request body map[string]interface{} true "状态: 0草稿 1已发布 2审核中 3未通过 4下架"
// @Success     200 {object} response.Response "状态更新成功"
// @Failure     400 {object} response.Response "请求参数错误或无效的作品状态"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id}/status [put]
func (h *WorkHandler) UpdateStatus(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if err := h.svc.UpdateStatus(c.Request().Context(), uint(id), models.WorkStatus(req.Status)); err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		if errors.Is(err, svc.ErrInvalidWorkStatus) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品状态"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "状态更新成功"}))
}

// IncrementCount godoc
// @Summary     更新作品计数
// @Description 更新作品的浏览、点赞、评论等计数
// @Tags        works
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "作品ID"
// @Param       request body map[string]interface{} true "计数字段和增量"
// @Success     200 {object} response.Response "计数更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "作品不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /works/{id}/count [put]
func (h *WorkHandler) IncrementCount(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	var req struct {
		Field string `json:"field"`
		Delta int    `json:"delta"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if err := h.svc.IncrementCount(c.Request().Context(), uint(id), req.Field, req.Delta); err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "作品不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "计数更新成功"}))
}
