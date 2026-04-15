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

type CommentHandler struct {
	svc svc.CommentService
}

func NewCommentHandler(svc svc.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) RegisterRoutes(g *echo.Group) {
	comments := g.Group("/comments")
	comments.GET("/work/:work_id", h.ListByWorkID)
	comments.GET("/work/:work_id/root", h.ListRootByWorkID)
	comments.GET("/user/:user_id", h.ListByUserID)
	comments.GET("/:id", h.GetComment)

	commentsAuth := comments.Group("")
	commentsAuth.Use(middleware.JWTAuth())
	commentsAuth.POST("", h.CreateComment)
	commentsAuth.PUT("/:id", h.UpdateComment)
	commentsAuth.DELETE("/:id", h.DeleteComment)
	commentsAuth.PUT("/:id/status", h.UpdateStatus, middleware.RequireRole("admin"))
	commentsAuth.PUT("/:id/like", h.IncrementLikeCount)
}

// ListByWorkID godoc
// @Summary     获取作品评论列表
// @Description 获取指定作品的所有评论（带分页）
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Param       work_id path int true "作品ID"
// @Param       page query int false "页码" default(1)
// @Param       page_size query int false "每页数量" default(10)
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/work/{work_id} [get]
func (h *CommentHandler) ListByWorkID(c *echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := strconv.ParseUint(workIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	comments, total, err := h.svc.ListByWorkID(c.Request().Context(), uint(workID), page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.SuccessWithPage(comments, page, pageSize, total))
}

// ListRootByWorkID godoc
// @Summary     获取作品根评论
// @Description 获取指定作品的一级评论（不带分页，返回树形结构）
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Param       work_id path int true "作品ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的作品ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/work/{work_id}/root [get]
func (h *CommentHandler) ListRootByWorkID(c *echo.Context) error {
	workIDStr := c.Param("work_id")
	workID, err := strconv.ParseUint(workIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的作品 ID"))
	}

	comments, err := h.svc.ListRootByWorkID(c.Request().Context(), uint(workID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(comments))
}

// ListByUserID godoc
// @Summary     获取用户评论
// @Description 获取指定用户的所有评论
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/user/{user_id} [get]
func (h *CommentHandler) ListByUserID(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	comments, err := h.svc.ListByUserID(c.Request().Context(), uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(comments))
}

// GetComment godoc
// @Summary     获取评论
// @Description 根据ID获取评论详情
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "评论ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的评论ID"
// @Failure     404 {object} response.Response "评论不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/{id} [get]
func (h *CommentHandler) GetComment(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论 ID"))
	}

	comment, err := h.svc.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "评论不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(comment))
}

// CreateComment godoc
// @Summary     发表评论
// @Description 对作品发表评论或回复
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body models.CreateCommentRequest true "评论请求"
// @Success     201 {object} response.Response "评论成功"
// @Failure     400 {object} response.Response "请求参数错误、作品不存在、父评论不存在或不能回复二级及以下的评论"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments [post]
func (h *CommentHandler) CreateComment(c *echo.Context) error {
	userID := middleware.GetUserID(c)

	var req models.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	comment, err := h.svc.Create(c.Request().Context(), &req, userID)
	if err != nil {
		if errors.Is(err, svc.ErrWorkNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "作品不存在"))
		}
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "父评论不存在"))
		}
		if errors.Is(err, svc.ErrCannotReplyChild) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "不能回复二级及以下的评论"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(comment))
}

// UpdateComment godoc
// @Summary     更新评论
// @Description 更新评论内容
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "评论ID"
// @Param       request body map[string]string true "评论内容"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "评论不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论 ID"))
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	comment, err := h.svc.Update(c.Request().Context(), uint(id), req.Content)
	if err != nil {
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "评论不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(comment))
}

// DeleteComment godoc
// @Summary     删除评论
// @Description 删除评论
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "评论ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "无效的评论ID"
// @Failure     404 {object} response.Response "评论不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论 ID"))
	}

	if err := h.svc.Delete(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "评论不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}

// UpdateStatus godoc
// @Summary     更新评论状态
// @Description 更新评论审核状态（仅管理员可操作）
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "评论ID"
// @Param       request body map[string]interface{} true "状态: 0删除 1正常 2审核中"
// @Success     200 {object} response.Response "状态更新成功"
// @Failure     400 {object} response.Response "请求参数错误或无效的评论状态"
// @Failure     404 {object} response.Response "评论不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/{id}/status [put]
func (h *CommentHandler) UpdateStatus(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论 ID"))
	}

	var req struct {
		Status int8 `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if err := h.svc.UpdateStatus(c.Request().Context(), uint(id), models.CommentStatus(req.Status)); err != nil {
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "评论不存在"))
		}
		if errors.Is(err, svc.ErrInvalidStatus) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论状态"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "状态更新成功"}))
}

// IncrementLikeCount godoc
// @Summary     更新评论点赞数
// @Description 更新评论的点赞数量
// @Tags        comments
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "评论ID"
// @Param       request body map[string]interface{} true "增量"
// @Success     200 {object} response.Response "点赞更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     404 {object} response.Response "评论不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/comments/{id}/like [put]
func (h *CommentHandler) IncrementLikeCount(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的评论 ID"))
	}

	var req struct {
		Delta int `json:"delta"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if err := h.svc.IncrementLikeCount(c.Request().Context(), uint(id), req.Delta); err != nil {
		if errors.Is(err, svc.ErrCommentNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "评论不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "点赞更新成功"}))
}
