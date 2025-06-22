package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
)

// MCPOAuthHandler 作为 MCP Client 实现 OAuth2 认证
type MCPOAuthHandler struct {
	config       *config.SocialConfig
	metaStore    *repositories.MetaStore
	oauthService *services.MCPOAuthService
}

// OAuth2 相关结构体定义

// ResourceMetadata 资源元数据
type ResourceMetadata struct {
	AuthorizationServers []string `json:"authorization_servers"`
	Resource             string   `json:"resource"`
}

// AuthServerMetadata 授权服务器元数据
type AuthServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
	JwksURI                           string   `json:"jwks_uri,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// ClientRegistrationRequest 客户端注册请求
type ClientRegistrationRequest struct {
	ClientName    string   `json:"client_name"`
	RedirectURIs  []string `json:"redirect_uris"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types"`
	Scope         string   `json:"scope,omitempty"`
}

// ClientRegistrationResponse 客户端注册响应
type ClientRegistrationResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// PKCEParams PKCE 参数
type PKCEParams struct {
	CodeVerifier  string
	CodeChallenge string
}

// OAuthSession OAuth 会话信息
type OAuthSession struct {
	State         string
	CodeVerifier  string
	AuthServerURL string
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	Resource      string
	CreatedAt     time.Time
}

func NewMCPOAuthHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MCPOAuthHandler {
	return &MCPOAuthHandler{
		config:       config,
		metaStore:    metaStore,
		oauthService: services.NewMCPOAuthService(metaStore),
	}
}

// 1. 资源发现阶段 - 获取资源元数据
func (h *MCPOAuthHandler) DiscoverResource(c echo.Context) error {
	resourceURL := c.QueryParam("resource_url")
	if resourceURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "resource_url 参数是必需的",
		})
	}

	// 构造资源元数据 URL
	metadataURL := strings.TrimSuffix(resourceURL, "/") + "/.well-known/oauth-protected-resource"

	// 获取资源元数据
	metadata, err := h.fetchResourceMetadata(metadataURL)
	if err != nil {
		log.Errorf("获取资源元数据失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取资源元数据失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, metadata)
}

// 2. 授权服务器发现 - 获取授权服务器元数据
func (h *MCPOAuthHandler) DiscoverAuthServer(c echo.Context) error {
	authServerURL := c.QueryParam("auth_server_url")
	if authServerURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "auth_server_url 参数是必需的",
		})
	}

	// 构造授权服务器元数据 URL
	metadataURL := strings.TrimSuffix(authServerURL, "/") + "/.well-known/oauth-authorization-server"

	// 获取授权服务器元数据
	metadata, err := h.fetchAuthServerMetadata(metadataURL)
	if err != nil {
		log.Errorf("获取授权服务器元数据失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取授权服务器元数据失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, metadata)
}

// 3. 客户端注册（可选）
func (h *MCPOAuthHandler) RegisterClient(c echo.Context) error {
	var req ClientRegistrationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无效的请求参数",
		})
	}

	registrationURL := c.QueryParam("registration_url")
	if registrationURL == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "registration_url 参数是必需的",
		})
	}

	// 发送客户端注册请求
	response, err := h.registerClient(registrationURL, &req)
	if err != nil {
		log.Errorf("客户端注册失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "客户端注册失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// 4. 开始授权流程
func (h *MCPOAuthHandler) StartAuthorization(c echo.Context) error {
	authServerURL := c.QueryParam("auth_server_url")
	clientID := c.QueryParam("client_id")
	resource := c.QueryParam("resource")

	if authServerURL == "" || clientID == "" || resource == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "auth_server_url, client_id, resource 参数都是必需的",
		})
	}

	// 生成 PKCE 参数
	pkce, err := h.generatePKCE()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成 PKCE 参数失败",
		})
	}

	// 生成 state 参数
	state := h.generateState()

	// 构造重定向 URI
	redirectURI := fmt.Sprintf("%s/api/mcp/oauth/callback", h.getBaseURL(c))

	// 保存会话信息到数据库
	session := h.oauthService.CreateOAuthSessionFromParams(
		"", // userDID - 在实际应用中应该从认证上下文获取
		state,
		pkce.CodeVerifier,
		authServerURL,
		clientID,
		"", // clientSecret - 可选
		redirectURI,
		resource,
	)

	if err := h.oauthService.SaveOAuthSession(session); err != nil {
		log.Errorf("保存 OAuth 会话失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存会话信息失败",
		})
	}

	// 构造授权 URL
	authURL, err := h.buildAuthorizationURL(authServerURL, clientID, redirectURI, resource, state, pkce.CodeChallenge)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "构造授权 URL 失败",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"authorization_url": authURL,
		"state":             state,
	})
}

// 5. 处理授权回调
func (h *MCPOAuthHandler) HandleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")

	if errorParam != "" {
		errorDescription := c.QueryParam("error_description")
		log.Errorf("授权失败: %s - %s", errorParam, errorDescription)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error":             errorParam,
			"error_description": errorDescription,
		})
	}

	if code == "" || state == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "缺少 code 或 state 参数",
		})
	}

	// 从存储中获取会话信息
	sessionData, err := h.oauthService.GetOAuthSession(state)
	if err != nil {
		log.Errorf("获取 OAuth 会话失败: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无效的 state 参数或会话已过期",
		})
	}

	// 转换为内部使用的会话结构
	session := &OAuthSession{
		State:         sessionData.State,
		CodeVerifier:  sessionData.CodeVerifier,
		AuthServerURL: sessionData.AuthServerURL,
		ClientID:      sessionData.ClientID,
		ClientSecret:  sessionData.ClientSecret,
		RedirectURI:   sessionData.RedirectURI,
		Resource:      sessionData.Resource,
	}

	// 交换授权码获取 token
	tokenResponse, err := h.exchangeCodeForToken(session, code)
	if err != nil {
		log.Errorf("交换 token 失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "交换 token 失败: " + err.Error(),
		})
	}

	// 存储 token 信息到数据库
	token := h.oauthService.CreateOAuthTokenFromResponse(
		sessionData.UserDID,
		session.ClientID,
		session.Resource,
		tokenResponse.AccessToken,
		tokenResponse.TokenType,
		tokenResponse.RefreshToken,
		tokenResponse.Scope,
		tokenResponse.ExpiresIn,
	)

	if err := h.oauthService.SaveOAuthToken(token); err != nil {
		log.Errorf("保存 OAuth Token 失败: %v", err)
		// 不返回错误，因为 Token 交换已经成功
	}

	// 清理已使用的会话
	if err := h.oauthService.DeleteOAuthSession(state); err != nil {
		log.Errorf("删除 OAuth 会话失败: %v", err)
	}

	return c.JSON(http.StatusOK, tokenResponse)
}

// 6. 刷新 Token
func (h *MCPOAuthHandler) RefreshToken(c echo.Context) error {
	refreshToken := c.QueryParam("refresh_token")
	authServerURL := c.QueryParam("auth_server_url")
	clientID := c.QueryParam("client_id")

	if refreshToken == "" || authServerURL == "" || clientID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "refresh_token, auth_server_url, client_id 参数都是必需的",
		})
	}

	// 刷新 token
	tokenResponse, err := h.refreshAccessToken(authServerURL, clientID, refreshToken)
	if err != nil {
		log.Errorf("刷新 token 失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "刷新 token 失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, tokenResponse)
}

// 7. 使用 Token 访问资源
func (h *MCPOAuthHandler) AccessResource(c echo.Context) error {
	resourceURL := c.QueryParam("resource_url")
	accessToken := c.QueryParam("access_token")

	if resourceURL == "" || accessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "resource_url 和 access_token 参数都是必需的",
		})
	}

	// 使用 access token 访问资源
	data, err := h.accessProtectedResource(resourceURL, accessToken)
	if err != nil {
		log.Errorf("访问资源失败: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "访问资源失败: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": data,
	})
}

// 辅助方法实现

func (h *MCPOAuthHandler) fetchResourceMetadata(metadataURL string) (*ResourceMetadata, error) {
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var metadata ResourceMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (h *MCPOAuthHandler) fetchAuthServerMetadata(metadataURL string) (*AuthServerMetadata, error) {
	resp, err := http.Get(metadataURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var metadata AuthServerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (h *MCPOAuthHandler) registerClient(registrationURL string, req *ClientRegistrationRequest) (*ClientRegistrationResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(registrationURL, "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var response ClientRegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (h *MCPOAuthHandler) generatePKCE() (*PKCEParams, error) {
	// 生成 code_verifier (43-128 字符的随机字符串)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// 生成 code_challenge (SHA256 hash of code_verifier)
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEParams{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
	}, nil
}

func (h *MCPOAuthHandler) generateState() string {
	return uuid.New().String()
}

func (h *MCPOAuthHandler) buildAuthorizationURL(authServerURL, clientID, redirectURI, resource, state, codeChallenge string) (string, error) {
	baseURL := strings.TrimSuffix(authServerURL, "/") + "/authorize"

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("resource", resource)
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("scope", "openid profile")

	return baseURL + "?" + params.Encode(), nil
}

func (h *MCPOAuthHandler) exchangeCodeForToken(session *OAuthSession, code string) (*TokenResponse, error) {
	tokenURL := strings.TrimSuffix(session.AuthServerURL, "/") + "/token"

	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("redirect_uri", session.RedirectURI)
	params.Set("client_id", session.ClientID)
	params.Set("code_verifier", session.CodeVerifier)
	params.Set("resource", session.Resource)

	resp, err := http.PostForm(tokenURL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func (h *MCPOAuthHandler) refreshAccessToken(authServerURL, clientID, refreshToken string) (*TokenResponse, error) {
	tokenURL := strings.TrimSuffix(authServerURL, "/") + "/token"

	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", refreshToken)
	params.Set("client_id", clientID)

	resp, err := http.PostForm(tokenURL, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func (h *MCPOAuthHandler) accessProtectedResource(resourceURL, accessToken string) (interface{}, error) {
	req, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func (h *MCPOAuthHandler) getBaseURL(c echo.Context) string {
	scheme := "http"
	if c.Request().TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request().Host)
}
