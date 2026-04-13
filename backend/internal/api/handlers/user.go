package handlers

import (
	"errors"
	"net/http"
	"strconv"

	_ "backend/internal/models"
	svc "backend/internal/services"
	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	svc *svc.UserService
}

func NewUserHandler(svc *svc.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func success(c *echo.Context, data any) error {
	return c.JSON(http.StatusOK, response.Success(data))
}

func errorResp(c *echo.Context, status int, msgOrCode string, msg ...string) error {
	if len(msg) > 0 {
		return c.JSON(status, response.Fail(msgOrCode, msg[0]))
	}
	return c.JSON(status, response.Fail(response.InternalError, msgOrCode))
}

// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body svc.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /users/register [post]
func (h *UserHandler) Register(c *echo.Context) error {
	var req svc.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	resp, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrUsernameExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "用户名已被使用"))
		}
		if errors.Is(err, svc.ErrEmailExists) {
			return c.JSON(http.StatusConflict, response.Fail(response.UserEmailExists, "邮箱已被注册"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(resp))
}

// @Summary 用户登录
// @Description 用户登录获取Token
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body svc.LoginRequest true "登录请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/login [post]
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

// @Summary 获取用户
// @Description 根据ID获取用户信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
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

// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response
// @Router /users [get]
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

// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body svc.UpdateUserRequest true "更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	var req svc.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}
	req.ID = uint(id)

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

// @Summary 修改密码
// @Description 修改用户密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body svc.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	var req svc.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}
	req.UserID = uint(id)

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

// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "无效的用户 ID"))
	}

	if err := h.svc.DeleteUser(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, response.Fail(response.UserNotFound, "用户不存在"))
		}
		return c.JSON(http.StatusInternalServerError, response.Fail(response.InternalError, err.Error()))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"message": "删除成功"}))
}

// @Summary 刷新Token
// @Description 使用RefreshToken刷新访问令牌
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body struct{RefreshToken string `json:"refresh_token"`} true "刷新Token请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /users/refresh [post]
func (h *UserHandler) RefreshToken(c *echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	tokenPair, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenInvalid, "无效的 Refresh Token"))
	}

	return c.JSON(http.StatusOK, response.Success(tokenPair))
}
