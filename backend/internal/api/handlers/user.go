package handlers

import (
	"errors"
	"net/http"
	"strconv"

	svc "backend/internal/services"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

type UserHandler struct {
	svc *svc.UserService
}

func NewUserHandler(svc *svc.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func success(c *echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Code: http.StatusOK,
		Data: data,
	})
}

func errorResp(c *echo.Context, status int, msg string) error {
	return c.JSON(status, ErrorResponse{
		Code:    status,
		Message: msg,
	})
}

// @Summary 用户注册
// @Description 创建新用户账号
// @Tags users
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "注册请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /users/register [post]
func (h *UserHandler) Register(c *echo.Context) error {
	var req svc.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}

	resp, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrEmailExists) {
			return errorResp(c, http.StatusConflict, "邮箱已被注册")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, resp)
}

// @Summary 用户登录
// @Description 用户登录获取Token
// @Tags users
// @Accept json
// @Produce json
// @Param request body object{email=string,password=string} true "登录请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) Login(c *echo.Context) error {
	var req svc.LoginRequest
	if err := c.Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}

	resp, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusUnauthorized, "用户不存在")
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return errorResp(c, http.StatusUnauthorized, "密码错误")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, resp)
}

// @Summary 获取用户
// @Description 根据ID获取用户信息
// @Tags users
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	user, err := h.svc.GetUserByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, user)
}

// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags users
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} SuccessResponse
// @Security BearerAuth
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
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// @Summary 更新用户
// @Description 更新用户信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body object{name=string,email=string} true "更新请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	var req svc.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}
	req.ID = uint(id)

	if err := h.svc.UpdateUser(c.Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		if errors.Is(err, svc.ErrEmailExists) {
			return errorResp(c, http.StatusConflict, "邮箱已被使用")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "更新成功"})
}

// @Summary 修改密码
// @Description 修改用户密码
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body object{old_password=string,new_password=string} true "修改密码请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/password [put]
func (h *UserHandler) ChangePassword(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	var req svc.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}
	req.UserID = uint(id)

	if err := h.svc.ChangePassword(c.Request().Context(), req); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		if errors.Is(err, svc.ErrInvalidPassword) {
			return errorResp(c, http.StatusUnauthorized, "旧密码错误")
		}
		if errors.Is(err, svc.ErrSamePassword) {
			return errorResp(c, http.StatusBadRequest, "新密码不能与旧密码相同")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "密码修改成功"})
}

// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags users
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return errorResp(c, http.StatusBadRequest, "无效的用户 ID")
	}

	if err := h.svc.DeleteUser(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, svc.ErrUserNotFound) {
			return errorResp(c, http.StatusNotFound, "用户不存在")
		}
		return errorResp(c, http.StatusInternalServerError, err.Error())
	}

	return success(c, map[string]interface{}{"message": "删除成功"})
}

// @Summary 刷新Token
// @Description 使用RefreshToken获取新的AccessToken
// @Tags users
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "刷新Token请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/refresh [post]
func (h *UserHandler) RefreshToken(c *echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return errorResp(c, http.StatusBadRequest, "请求参数错误")
	}

	tokenPair, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return errorResp(c, http.StatusUnauthorized, "无效的 Refresh Token")
	}

	return success(c, tokenPair)
}
