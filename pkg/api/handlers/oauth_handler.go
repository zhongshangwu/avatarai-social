package handlers

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type OAuthHandler struct {
	config    *config.SocialConfig
	metaStore *repositories.MetaStore
	client    *atproto.OAuthClient
}

type OAuthClientMetadata struct {
	ClientID                    string   `json:"client_id"`
	DpopBoundAccessTokens       bool     `json:"dpop_bound_access_tokens"`
	ApplicationType             string   `json:"application_type"`
	RedirectURIs                []string `json:"redirect_uris"`
	GrantTypes                  []string `json:"grant_types"`
	ResponseTypes               []string `json:"response_types"`
	Scope                       string   `json:"scope"`
	TokenEndpointAuthMethod     string   `json:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg string   `json:"token_endpoint_auth_signing_alg"`
	JwksURI                     string   `json:"jwks_uri"`
	ClientName                  string   `json:"client_name"`
	ClientURI                   string   `json:"client_uri"`
}

type JWKS struct {
	Keys []interface{} `json:"keys"`
}

type OAuthSession struct {
	DID                 string    `json:"did"`
	Handle              string    `json:"handle"`
	PDSURL              string    `json:"pds_url"`
	AuthserverISS       string    `json:"authserver_iss"`
	AccessToken         string    `json:"access_token"`
	RefreshToken        string    `json:"refresh_token"`
	DpopAuthserverNonce string    `json:"dpop_authserver_nonce"`
	DpopPrivateJWK      string    `json:"dpop_private_jwk"`
	CreatedAt           time.Time `json:"created_at"`
}

type OAuthAuthRequest struct {
	State               string    `json:"state"`
	AuthserverISS       string    `json:"authserver_iss"`
	DID                 string    `json:"did"`
	Handle              string    `json:"handle"`
	PDSURL              string    `json:"pds_url"`
	PKCEVerifier        string    `json:"pkce_verifier"`
	Scope               string    `json:"scope"`
	DpopAuthserverNonce string    `json:"dpop_authserver_nonce"`
	DpopPrivateJWK      string    `json:"dpop_private_jwk"`
	CreatedAt           time.Time `json:"created_at"`
}

type PageData struct {
	User     *UserData
	Messages []string
	URLs     map[string]string
}

type UserData struct {
	Handle string
}

type ExchangeTokenRequest struct {
	Code string `json:"code"`
}

func NewOAuthHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *OAuthHandler {
	appURL := fmt.Sprintf("http://%s", config.Server.HTTP.Address)
	if !utils.IsSafeURL(appURL) {
		appURL = "http://localhost:8080/"
	}

	client := atproto.NewOAuthClient(appURL, config.ATP.ClientSecretJWK())

	return &OAuthHandler{
		config:    config,
		metaStore: metaStore,
		client:    client,
	}
}

func (h *OAuthHandler) errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]string{"error": message})
}

func (h *OAuthHandler) successResponse(c echo.Context, data map[string]string) error {
	return c.JSON(http.StatusOK, data)
}

func (h *OAuthHandler) resolveIdentity(username string) (did, handle, pdsURL, authserverURL, loginHint string, err error) {
	if atproto.IsValidHandle(username) || atproto.IsValidDID(username) {
		loginHint = username
		ident, identErr := atproto.ResolveIdentity(context.Background(), username)
		if identErr != nil {
			err = fmt.Errorf("无法解析身份: %w", identErr)
			return
		}
		did = string(ident.DID)
		handle = string(ident.Handle)
		pdsURL = atproto.PDSEndpoint(ident)

		authserverURL, err = atproto.ResolvePDSAuthserver(pdsURL)
		if err != nil {
			err = fmt.Errorf("无法解析授权服务器 URL: %w", err)
			return
		}
	} else if utils.IsSafeURL(username) {
		did, handle, pdsURL = "", "", ""
		loginHint = ""
		initialURL := username

		authserverURL, err = atproto.ResolvePDSAuthserver(initialURL)
		if err != nil {
			err = fmt.Errorf("无法解析授权服务器 URL: %w", err)
			return
		}
	} else {
		err = fmt.Errorf("不是有效的 handle、DID 或授权服务器 URL")
		return
	}
	return
}

func (h *OAuthHandler) generateDpopKeyPair() (jose.JSONWebKey, string, error) {
	dpopPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return jose.JSONWebKey{}, "", fmt.Errorf("生成 DPoP 密钥失败: %w", err)
	}

	dpopPrivateJWK := jose.JSONWebKey{
		Key:       dpopPrivateKey,
		KeyID:     fmt.Sprintf("dpop-%d", time.Now().Unix()),
		Algorithm: string(jose.ES256),
		Use:       "sig",
	}

	dpopPrivateJWKStr, err := dpopPrivateJWK.MarshalJSON()
	if err != nil {
		return jose.JSONWebKey{}, "", fmt.Errorf("序列化 DPoP 密钥失败: %w", err)
	}

	return dpopPrivateJWK, string(dpopPrivateJWKStr), nil
}

func (h *OAuthHandler) buildRedirectURL(appURL string) string {
	return appURL + "api/oauth/callback"
}

func (h *OAuthHandler) HandleOAuthClientMetadata(c echo.Context) error {
	appURL := utils.GetAPPURL(c)
	platform := c.Param("platform")
	clientID := atproto.BuildClientID(appURL, platform)

	metadata := OAuthClientMetadata{
		ClientID:                    clientID,
		DpopBoundAccessTokens:       true,
		ApplicationType:             "web",
		RedirectURIs:                []string{h.buildRedirectURL(appURL)},
		GrantTypes:                  []string{"authorization_code", "refresh_token"},
		ResponseTypes:               []string{"code"},
		Scope:                       "atproto transition:generic",
		TokenEndpointAuthMethod:     "private_key_jwt",
		TokenEndpointAuthSigningAlg: "ES256",
		JwksURI:                     appURL + "api/oauth/jwks.json",
		ClientName:                  "ATProto OAuth Go Backend",
		ClientURI:                   appURL,
	}
	return c.JSON(http.StatusOK, metadata)
}

func (h *OAuthHandler) HandleOAuthJWKS(c echo.Context) error {
	clientPubJWK := h.config.ATP.ClientPubJWKMap()
	jwks := JWKS{
		Keys: []interface{}{clientPubJWK},
	}

	return c.JSON(http.StatusOK, jwks)
}

func (h *OAuthHandler) OAuthLogin(c echo.Context) error {
	platform := c.QueryParam("platform")
	if platform == "" {
		platform = "web"
	}

	log.Info("HandleOAuthLogin", "method", c.Request().Method, "platform", platform)

	if c.Request().Method == http.MethodGet {
		data := PageData{
			User: nil,
			URLs: map[string]string{"OAuthLogin": "/api/oauth/signin?platform=" + platform},
		}
		return c.Render(http.StatusOK, "layout.html", data)
	}

	username := c.FormValue("username")

	did, handle, pdsURL, authserverURL, loginHint, err := h.resolveIdentity(username)
	if err != nil {
		return h.errorResponse(c, http.StatusBadRequest, err.Error())
	}

	appURL := utils.GetAPPURL(c)
	authReq := &atproto.AuthRequest{
		LoginHint:   loginHint,
		Platform:    platform,
		Scope:       "atproto transition:generic",
		RedirectURI: h.buildRedirectURL(appURL),
	}

	authResp, err := h.client.StartAuth(authReq)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "启动授权流程失败: "+err.Error())
	}

	_, dpopPrivateJWKStr, err := h.generateDpopKeyPair()
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, err.Error())
	}

	authRequest := &repositories.OAuthAuthRequest{
		State:               authResp.State,
		AuthserverIss:       authserverURL,
		Did:                 did,
		Handle:              handle,
		PdsUrl:              pdsURL,
		PkceVerifier:        authResp.PKCEVerifier,
		Scope:               authReq.Scope,
		DpopAuthserverNonce: authResp.DpopNonce,
		DpopPrivateJwk:      dpopPrivateJWKStr,
		Platform:            platform,
		ReturnURI:           "",
	}

	if err := h.metaStore.OAuthRepo.InsertOAuthAuthRequest(authRequest); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "保存授权请求失败: "+err.Error())
	}

	return c.Redirect(http.StatusFound, authResp.AuthURL)
}

func (h *OAuthHandler) HandleOAuthCallback(c echo.Context) error {
	code := c.QueryParam("code")
	state := c.QueryParam("state")
	errorParam := c.QueryParam("error")
	errorDescription := c.QueryParam("error_description")

	if errorParam != "" {
		return h.errorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("OAuth 错误: %s - %s", errorParam, errorDescription))
	}

	if code == "" || state == "" {
		return h.errorResponse(c, http.StatusBadRequest, "缺少必需的参数")
	}

	authReq, err := h.metaStore.OAuthRepo.GetOAuthAuthRequest(state)
	if err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "无效的状态参数")
	}

	appURL := utils.GetAPPURL(c)
	tokenReq := &atproto.TokenRequest{
		Code:        code,
		State:       state,
		Platform:    authReq.Platform,
		RedirectURI: h.buildRedirectURL(appURL),
	}

	tokenResp, err := h.client.ExchangeToken(tokenReq, authReq)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "令牌交换失败: "+err.Error())
	}

	oauthSession := &repositories.OAuthSession{
		Did:                 authReq.Did,
		Handle:              authReq.Handle,
		PdsUrl:              authReq.PdsUrl,
		AuthserverIss:       authReq.AuthserverIss,
		AccessToken:         tokenResp.AccessToken,
		RefreshToken:        tokenResp.RefreshToken,
		DpopAuthserverNonce: tokenResp.DpopAuthserverNonce,
		DpopPrivateJwk:      authReq.DpopPrivateJwk,
		ExpiresIn:           tokenResp.ExpiresIn,
		CreatedAt:           utils.Timestamp(),
		Platform:            authReq.Platform,
		ReturnURI:           authReq.ReturnURI,
	}

	if err := h.metaStore.OAuthRepo.SaveOAuthSession(oauthSession); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "保存会话失败")
	}

	if err := h.metaStore.OAuthRepo.DeleteOAuthAuthRequest(state); err != nil {
		log.Warnf("删除授权请求失败: %v", err)
	}

	return h.successResponse(c, map[string]string{
		"message": "授权成功",
		"did":     oauthSession.Did,
		"handle":  oauthSession.Handle,
	})
}

func (h *OAuthHandler) HandleOAuthToken(c echo.Context) error {
	var req ExchangeTokenRequest
	if err := c.Bind(&req); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "无效的请求格式")
	}

	return h.successResponse(c, map[string]string{
		"message": "令牌处理成功",
	})
}

func (h *OAuthHandler) HandleOAuthRefresh(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

	if oauthSession == nil {
		return h.errorResponse(c, http.StatusUnauthorized, "未找到有效会话")
	}

	appURL := utils.GetAPPURL(c)
	refreshReq := &atproto.RefreshRequest{
		SessionDID:  oauthSession.Did,
		Platform:    oauthSession.Platform,
		RedirectURI: h.buildRedirectURL(appURL),
	}

	tokenResp, err := h.client.RefreshToken(refreshReq, oauthSession)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "刷新令牌失败: "+err.Error())
	}

	updateSession := &repositories.OAuthSession{
		Did:                 oauthSession.Did,
		AccessToken:         tokenResp.AccessToken,
		RefreshToken:        tokenResp.RefreshToken,
		DpopAuthserverNonce: tokenResp.DpopAuthserverNonce,
	}

	if err := h.metaStore.OAuthRepo.UpdateOAuthSession(updateSession); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "更新会话失败")
	}

	return h.successResponse(c, map[string]string{
		"message": "令牌刷新成功",
	})
}

func (h *OAuthHandler) HandleOAuthLogout(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	oauthSession := ac.OauthSession

	if oauthSession != nil {
		if err := h.metaStore.OAuthRepo.DeleteOAuthSessionByDID(oauthSession.Did); err != nil {
			log.Warnf("删除会话失败: %v", err)
		}
	}

	return h.successResponse(c, map[string]string{
		"message": "登出成功",
	})
}

func (h *OAuthHandler) HandleAppReturn(c echo.Context) error {
	bundleID := c.Param("bundleID")
	return h.successResponse(c, map[string]string{
		"bundle_id": bundleID,
		"message":   "应用返回处理成功",
	})
}

func (h *OAuthHandler) HandleBskyPost(c echo.Context) error {
	return h.successResponse(c, map[string]string{
		"message": "Bsky 发布处理成功",
	})
}
