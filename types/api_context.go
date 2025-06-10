package types

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

type APIContext struct {
	echo.Context

	// 用户信息
	IsAuthenticated bool
	User            *User

	// 会话信息
	Session *Session

	// 第三方认证信息
	OauthProvider OAuthProviderType
	OauthSession  *OAuthSession
}

func (c *APIContext) IsAster() bool {
	if c.User == nil {
		return false
	}
	return c.User.IsAster
}

func (c *APIContext) IsUser() bool {
	if c.User == nil {
		return false
	}
	return !c.User.IsAster
}

func (c *APIContext) RedirectToLogin(redirectURI string) error {
	if redirectURI == "" {
		return c.Redirect(http.StatusFound, "/api/oauth/login")
	}
	return c.Redirect(http.StatusFound, "/api/oauth/login?redirectURI="+url.QueryEscape(redirectURI))
}

func (c *APIContext) InvalidRequest(code, message string) error {
	return c.JSON(http.StatusBadRequest, &APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func (c *APIContext) InternalServerError(message string) error {
	return c.JSON(http.StatusInternalServerError, &APIResponse{
		Code:    string(ErrorCodeInternalServerError),
		Message: message,
		Data:    nil,
	})
}

func (c *APIContext) AuthToken() string {
	var token string
	jwtCookie, err := c.Cookie("avatarai_token")
	if err != nil || jwtCookie.Value == "" {
		// 尝试从 Authorization header 中获取
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			}
		}
	} else {
		token = jwtCookie.Value
	}

	return token
}

type APIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorCode string

const (
	ErrorCodeInternalServerError  ErrorCode = "internal_server_error"
	ErrorCodeInvalidRequestParams ErrorCode = "invalid_request_params"
)
