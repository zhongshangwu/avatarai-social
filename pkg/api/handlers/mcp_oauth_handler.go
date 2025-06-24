package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/mcp"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type MCPOAuthHandler struct {
	config     *config.SocialConfig
	metaStore  *repositories.MetaStore
	mcpService *services.MCPService
}

func NewMCPOAuthHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MCPOAuthHandler {
	return &MCPOAuthHandler{
		config:     config,
		metaStore:  metaStore,
		mcpService: services.NewMCPService(metaStore),
	}
}

func (h *MCPOAuthHandler) Authorize(c *types.APIContext) error {
	userDid := c.User.Did
	mcpId := c.QueryParam("mcpId")
	returnURI := c.QueryParam("returnUri") // 前端希望授权完成后跳转的地址

	if mcpId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "mcpId 参数是必需的",
		})
	}

	serverInfo, err := h.mcpService.GetMCPServerDetail(mcpId, userDid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取 MCP 服务器信息失败",
		})
	}

	if serverInfo.IsBuiltin {
		if err := h.mcpService.InstallBuiltinIfNotExists(serverInfo, userDid); err != nil {
			return c.InternalServerError(err.Error())
		}
	}

	if serverInfo.Authorization.Method != mcp.MCPServerAuthorizationMethodOAuth2 {
		return c.InvalidRequest("invalid_request", "MCP 服务器不支持 OAuth2 认证")
	}

	existingAuth, err := h.mcpService.GetMCPServerAuth(mcpId, userDid)
	if err == nil && existingAuth != nil && existingAuth.Status == repositories.AuthStatusActive {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"already_authorized": true,
			"status":             "active",
			"expires_at":         existingAuth.ExpiresAt,
			"scopes":             existingAuth.Scope,
		})
	}

	client, err := mcp.NewMCPClient(h.metaStore, serverInfo)
	if err != nil {
		logrus.WithError(err).Error("初始化 MCP 客户端失败")
		return c.InternalServerError(err.Error())
	}

	state, err := client.GenerateState()
	if err != nil {
		logrus.WithError(err).Error("生成 state 失败")
		return c.InternalServerError(err.Error())
	}

	codeVerifier, codeChallenge, err := client.GenerateCodeChallenge()
	if err != nil {
		logrus.WithError(err).Error("生成 code verifier 和 code challenge 失败")
		return c.InternalServerError(err.Error())
	}

	authURL, err := client.GetAuthorizationURL(c.Request().Context(), state, codeChallenge)
	if err != nil {
		logrus.WithError(err).Error("获取授权 URL 失败")
		return c.InternalServerError(err.Error())
	}

	_, err = h.mcpService.CreateOAuthCode(
		mcpId,
		userDid,
		state,
		codeVerifier,
		codeChallenge,
		"S256",
		returnURI,
		serverInfo.Authorization.Scopes,
		time.Now().Add(time.Hour).Unix(), // 1小时过期，更短的过期时间
	)

	if err != nil {
		log.Errorf("保存 OAuth 会话失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存会话信息失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"authorization_url": authURL,
		"state":             state,
		"expires_in":        3600, // state有效期（秒）
		"scopes":            serverInfo.Authorization.Scopes,
	})
}

func (h *MCPOAuthHandler) OAuthCallback(c *types.APIContext) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")
	issuer := c.QueryParam("iss") // 部分授权服务会返回 iss, 防止多授权服务冲突

	if errorParam != "" {
		errorDescription := c.QueryParam("error_description")
		log.Errorf("授权失败: %s - %s", errorParam, errorDescription)
		return c.InvalidRequest("invalid_request", errorParam)
	}

	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "缺少 code 或 state 参数",
		})
	}

	oauthCode, err := h.mcpService.GetMCPServerOAuthCode(issuer, state)
	if err != nil {
		log.Errorf("获取 OAuth 会话失败: %v", err)
		return c.InternalServerError(err.Error())
	}

	serverInfo, err := h.mcpService.GetMCPServerDetail(oauthCode.McpID, oauthCode.UserDid)
	if err != nil {
		return c.InternalServerError(err.Error())
	}

	if serverInfo.Authorization.Method != mcp.MCPServerAuthorizationMethodOAuth2 {
		return c.InvalidRequest("invalid_request", "MCP 服务器不支持 OAuth2 认证")
	}

	client, err := mcp.NewMCPClient(h.metaStore, serverInfo)
	if err != nil {
		logrus.WithError(err).Error("初始化 MCP 客户端失败")
		return c.InternalServerError(err.Error())
	}

	if err = client.ExchangeCode(c.Request().Context(), oauthCode.State, code, state, oauthCode.CodeVerifier); err != nil {
		logrus.WithError(err).Error("交换授权码失败")
		return c.InternalServerError(err.Error())
	}

	redirectUri := oauthCode.RedirectURI
	if redirectUri == "" {
		redirectUri = h.config.Server.Domain
	}
	return c.Redirect(http.StatusFound, redirectUri)
}

func (h *MCPOAuthHandler) getBaseURL(c echo.Context) string {
	return h.config.Server.Domain
}
