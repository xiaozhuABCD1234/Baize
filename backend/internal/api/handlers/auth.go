package handlers

import (
	"net/http"

	"backend/pkg/response"
	"backend/pkg/utils"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	auth := g.Group("/auth")
	auth.POST("/refresh", h.RefreshToken)
}

// RefreshToken godoc
// @Summary     刷新Access Token
// @Description 使用Refresh Token获取新的Access Token
// @Tags        auth
// @Accept      application/json
// @Produce     application/json
// @Param       request body map[string]string true "Refresh Token"
// @Success     200 {object} response.Response "刷新成功"
// @Failure     400 {object} response.Response "请求参数错误"
// @Failure     401 {object} response.Response "无效的Refresh Token"
// @Failure     500 {object} response.Response "服务器内部错误"
// @Router      /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "请求参数错误"))
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, response.Fail(response.BadRequest, "refresh_token 不能为空"))
	}

	tokenPair, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, response.Fail(response.TokenInvalid, "无效的 Refresh Token"))
	}

	return c.JSON(http.StatusOK, response.Success(tokenPair))
}
