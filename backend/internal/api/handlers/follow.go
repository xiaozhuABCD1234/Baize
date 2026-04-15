package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"backend/internal/api/middleware"
	svc "backend/internal/services"
	"backend/pkg/response"

	"github.com/labstack/echo/v5"
)

type FollowHandler struct {
	svc svc.FollowService
}

func NewFollowHandler(svc svc.FollowService) *FollowHandler {
	return &FollowHandler{svc: svc}
}

func (h *FollowHandler) RegisterRoutes(g *echo.Group) {
	follows := g.Group("/follows")
	follows.GET("/check/:user_id", h.IsFollowing)
	follows.GET("/following/:user_id", h.GetFollowingList)
	follows.GET("/followers/:user_id", h.GetFollowerList)
	follows.GET("/following/:user_id/count", h.GetFollowingCount)
	follows.GET("/followers/:user_id/count", h.GetFollowerCount)

	followsAuth := follows.Group("")
	followsAuth.Use(middleware.JWTAuth())
	followsAuth.POST("", h.CreateFollow)
	followsAuth.DELETE("/:user_id", h.DeleteFollow)
}

// IsFollowing godoc
// @Summary     检查是否关注
// @Description 检查当前用户是否关注了指定用户
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/check/{user_id} [get]
func (h *FollowHandler) IsFollowing(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	currentUserID := middleware.GetUserID(c)

	isFollowing, err := h.svc.IsFollowing(c.Request().Context(), currentUserID, uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"is_following": isFollowing}))
}

// GetFollowingList godoc
// @Summary     获取关注列表
// @Description 获取指定用户的关注列表
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/following/{user_id} [get]
func (h *FollowHandler) GetFollowingList(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	follows, err := h.svc.GetFollowingList(c.Request().Context(), uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(follows))
}

// GetFollowerList godoc
// @Summary     获取粉丝列表
// @Description 获取指定用户的粉丝列表
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/followers/{user_id} [get]
func (h *FollowHandler) GetFollowerList(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	follows, err := h.svc.GetFollowerList(c.Request().Context(), uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(follows))
}

// GetFollowingCount godoc
// @Summary     获取关注数
// @Description 获取指定用户的关注数量
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/following/{user_id}/count [get]
func (h *FollowHandler) GetFollowingCount(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	count, err := h.svc.GetFollowingCount(c.Request().Context(), uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"count": count}))
}

// GetFollowerCount godoc
// @Summary     获取粉丝数
// @Description 获取指定用户的粉丝数量
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Param       user_id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/followers/{user_id}/count [get]
func (h *FollowHandler) GetFollowerCount(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	count, err := h.svc.GetFollowerCount(c.Request().Context(), uint(userID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"count": count}))
}

// CreateFollow godoc
// @Summary     关注用户
// @Description 当前用户关注指定用户
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       request body map[string]interface{} true "关注的用户ID"
// @Success     201 {object} response.Response "关注成功"
// @Failure     400 {object} response.Response "请求参数错误、不能关注自己或用户不存在"
// @Failure     409 {object} response.Response "已经关注过该用户"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows [post]
func (h *FollowHandler) CreateFollow(c *echo.Context) error {
	followerID := middleware.GetUserID(c)

	var req struct {
		FollowingID uint `json:"following_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	follow, err := h.svc.Create(c.Request().Context(), followerID, req.FollowingID)
	if err != nil {
		if errors.Is(err, svc.ErrCannotFollowSelf) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "不能关注自己"))
		}
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "用户不存在"))
		}
		if errors.Is(err, svc.ErrAlreadyFollowing) {
			return c.JSON(http.StatusConflict, response.Fail(response.ResourceConflict, "已经关注过该用户"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusCreated, response.Success(follow))
}

// DeleteFollow godoc
// @Summary     取消关注
// @Description 取消关注指定用户
// @Tags        follows
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       user_id path int true "要取消关注的用户ID"
// @Success     200 {object} response.Response "取消关注成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     404 {object} response.Response "未关注该用户"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /api/v1/follows/{user_id} [delete]
func (h *FollowHandler) DeleteFollow(c *echo.Context) error {
	userIDStr := c.Param("user_id")
	followingID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	followerID := middleware.GetUserID(c)

	if err := h.svc.Delete(c.Request().Context(), followerID, uint(followingID)); err != nil {
		if errors.Is(err, svc.ErrNotFollowing) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "未关注该用户"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "取消关注成功"}))
}
