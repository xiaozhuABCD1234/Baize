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

type UserHandler struct {
	svc *svc.UserService
}

func NewUserHandler(svc *svc.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoutes(g *echo.Group) {
	users := g.Group("/users")
	users.POST("/register", h.Register)
	users.POST("/login", h.Login)
	users.GET("", h.ListUsers)
	users.GET("/:id", h.GetUser)

	usersAuth := users.Group("")
	usersAuth.Use(middleware.JWTAuth())
	usersAuth.PUT("/:id", h.UpdateUser)
	usersAuth.PUT("/:id/password", h.ChangePassword)
	usersAuth.DELETE("/:id", h.DeleteUser)
	usersAuth.DELETE("/:id/force", h.ForceDeleteUser, middleware.RequireRole("admin"))
}

// Register godoc
// @Summary     用户注册
// @Description 创建新用户账号
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Param       request body map[string]interface{} true "注册请求"
// @Success     200 {object} response.Response "注册成功"
// @Failure     400 {object} response.Response "请求参数错误或无效的角色类型"
// @Failure     409 {object} response.Response "用户名或邮箱已被使用"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/register [post]
func (h *UserHandler) Register(c *echo.Context) error {
	var req svc.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	resp, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrUsernameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserNameExists, "用户名已被使用"))
		}
		if errors.Is(err, svc.ErrEmailExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "邮箱已被注册"))
		}
		if errors.Is(err, svc.ErrInvalidRole) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的角色类型"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// Login godoc
// @Summary     用户登录
// @Description 用户登录获取Access Token和Refresh Token
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Param       request body map[string]interface{} true "登录请求"
// @Success     200 {object} response.Response "登录成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     401 {object} response.Response "用户不存在或密码错误"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/login [post]
func (h *UserHandler) Login(c *echo.Context) error {
	var req svc.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	resp, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, response.Fail(response.UserNotFound, "用户不存在"))
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, response.Fail(response.InvalidPassword, "密码错误"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// GetUser godoc
// @Summary     获取用户
// @Description 根据ID获取用户信息
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Param       id path int true "用户ID"
// @Success     200 {object} response.Response "获取成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     404 {object} response.Response "用户不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/{id} [get]
func (h *UserHandler) GetUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	user, err := h.svc.GetUserByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(user))
}

// ListUsers godoc
// @Summary     获取用户列表
// @Description 分页获取用户列表
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Param       page query int false "页码" default(1)
// @Param       page_size query int false "每页数量" default(10)
// @Success     200 {object} response.Response "获取成功"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users [get]
func (h *UserHandler) ListUsers(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.svc.ListUsersWithPagination(c.Request().Context(), page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.SuccessWithPage(users, page, pageSize, total))
}

// UpdateUser godoc
// @Summary     更新用户
// @Description 更新用户信息（仅本人或管理员可操作）
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Param       request body map[string]interface{} true "更新请求"
// @Success     200 {object} response.Response "更新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     403 {object} response.Response "无权修改此用户"
// @Failure     404 {object} response.Response "用户不存在"
// @Failure     409 {object} response.Response "邮箱已被使用"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/{id} [put]
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	currentUserID := middleware.GetUserID(c)
	currentUserType := middleware.GetUserType(c)
	targetID := uint(id)

	if currentUserType != "admin" && currentUserID != targetID {
		return c.JSON(http.StatusForbidden, response.Fail(response.Forbidden, "无权修改此用户"))
	}

	var req svc.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}
	req.ID = targetID

	if err := h.svc.UpdateUser(c.Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		if errors.Is(err, svc.ErrEmailExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "邮箱已被使用"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "更新成功"}))
}

// ChangePassword godoc
// @Summary     修改密码
// @Description 修改用户密码（仅本人或管理员可操作）
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Param       request body map[string]string true "密码修改请求"
// @Success     200 {object} response.Response "密码修改成功"
// @Failure     400 {object} response.Response "请求参数错误或新密码不能与旧密码相同"
// @Failure     401 {object} response.Response "旧密码错误"
// @Failure     403 {object} response.Response "无权修改此用户密码"
// @Failure     404 {object} response.Response "用户不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	currentUserID := middleware.GetUserID(c)
	currentUserType := middleware.GetUserType(c)
	targetID := uint(id)

	if currentUserType != "admin" && currentUserID != targetID {
		return c.JSON(http.StatusForbidden, response.Fail(response.Forbidden, "无权修改此用户密码"))
	}

	var req svc.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}
	req.UserID = targetID

	if err := h.svc.ChangePassword(c.Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, response.Fail(response.InvalidPassword, "旧密码错误"))
		}
		if errors.Is(err, svc.ErrSamePassword) {
			return c.JSON(http.StatusBadRequest, response.Fail(response.SamePassword, "新密码不能与旧密码相同"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "密码修改成功"}))
}

// DeleteUser godoc
// @Summary     删除用户
// @Description 删除用户账号（仅本人或管理员可操作）
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Success     200 {object} response.Response "删除成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     403 {object} response.Response "无权删除此用户"
// @Failure     404 {object} response.Response "用户不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	currentUserID := middleware.GetUserID(c)
	currentUserType := middleware.GetUserType(c)
	targetID := uint(id)

	if currentUserType != "admin" && currentUserID != targetID {
		return c.JSON(http.StatusForbidden, response.Fail(response.Forbidden, "无权删除此用户"))
	}

	if err := h.svc.DeleteUser(c.Request().Context(), targetID); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}

// ForceDeleteUser godoc
// @Summary     永久删除用户
// @Description 永久删除用户账号（仅管理员可操作）
// @Tags        users
// @Accept      application/json
// @Produce     application/json
// @Security    BearerAuth
// @Param       id path int true "用户ID"
// @Success     200 {object} response.Response "永久删除成功"
// @Failure     400 {object} response.Response "无效的用户ID"
// @Failure     404 {object} response.Response "用户不存在"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /users/{id}/force [delete]
func (h *UserHandler) ForceDeleteUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	targetID := uint(id)

	if err := h.svc.ForceDeleteUser(c.Request().Context(), targetID); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "永久删除成功"}))
}
