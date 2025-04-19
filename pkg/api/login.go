package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

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
	User     *UserData // 指针类型，方便判断是否为 nil
	Messages []string
	URLs     map[string]string
}

type UserData struct {
	Handle string
	// ... 其他用户字段
}

func (a *AvatarAIAPI) HandleOAuthClientMetadata(c echo.Context) error {
	appURL := utils.GetAPPURL(c)
	platform := c.Param("platform")
	clientID := atproto.BuildClientID(appURL, platform)

	metadata := OAuthClientMetadata{
		ClientID:                    clientID,
		DpopBoundAccessTokens:       true,
		ApplicationType:             "web",
		RedirectURIs:                []string{appURL + "api/oauth/callback"},
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

func (a *AvatarAIAPI) HandleOAuthJWKS(c echo.Context) error {
	clientPubJWK := a.Config.ATP.ClientPubJWKMap()
	jwks := JWKS{
		Keys: []interface{}{clientPubJWK},
	}

	return c.JSON(http.StatusOK, jwks)
}

func (a *AvatarAIAPI) HandleOAuthLogin(c echo.Context) error {
	platform := c.QueryParam("platform")
	log.Info("HandleOAuthLogin", "method", c.Request().Method, "platform", platform)
	if c.Request().Method == http.MethodGet {
		data := PageData{
			User: nil,
			URLs: map[string]string{"OAuthLogin": "/api/oauth/signin"},
		}
		return c.Render(http.StatusOK, "layout.html", data)
	}

	username := c.FormValue("username")
	var did, handle, pdsURL, authserverURL, loginHint string

	if atproto.IsValidHandle(username) || atproto.IsValidDID(username) {
		// 如果以帐户标识符开始，解析身份，获取 PDS URL，并解析为授权服务器 URL
		log.Infof("HandleOAuthLogin， username: %s", username)
		loginHint = username
		ident, err := atproto.ResolveIdentity(context.Background(), username)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析身份: " + err.Error(),
			})
		}
		did = string(ident.DID)
		handle = string(ident.Handle)
		pdsURL = atproto.PDSEndpoint(ident)
		authserverURL, err = atproto.ResolvePDSAuthserver(pdsURL)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析授权服务器 URL: " + err.Error(),
			})
		}
	} else if utils.IsSafeURL(username) {
		log.Infof("HandleOAuthLogin else if， username: %s", username)
		// 从授权服务器开始
		did, handle, pdsURL = "", "", ""
		loginHint = ""
		initialURL := username

		// 检查是否为资源服务器(PDS)URL，否则假定为授权服务器
		var err error
		authserverURL, err = atproto.ResolvePDSAuthserver(initialURL)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析授权服务器 URL: " + err.Error(),
			})
		}
	} else {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "不是有效的 handle、DID 或授权服务器 URL",
		})
	}

	// 获取授权服务器元数据
	authserverMeta, err := atproto.FetchAuthserverMeta(authserverURL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("无法获取授权服务器元数据: %s", err),
		})
	}

	// 生成 DPoP 私钥
	// dpopPrivateJWK, err := crypto.GeneratePrivateKeyP256()
	dpopPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	dpopPrivateJWK := jose.JSONWebKey{
		Key:       dpopPrivateKey,
		KeyID:     fmt.Sprintf("demo-%d", time.Now().Unix()),
		Algorithm: "ES256",
		Use:       "sig",
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "无法生成 DPoP 私钥",
		})
	}

	// 请求的 OAuth 范围
	scope := "atproto transition:generic"

	appURL := utils.GetAPPURL(c)
	redirectURI := atproto.BuildRedirectURL(appURL, platform)
	clientID := atproto.BuildClientID(appURL, platform)

	// 提交 OAuth Pushed Authentication Request (PAR)
	pkceVerifier, state, dpopAuthserverNonce, resp, err := atproto.SendPARAuthRequest(
		authserverURL,
		authserverMeta,
		loginHint,
		clientID,
		redirectURI,
		scope,
		a.Config.ATP.ClientSecretJWK(),
		dpopPrivateJWK,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "PAR 请求失败: " + err.Error(),
		})
	}

	// 获取请求 URI
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Error("读取 PAR 响应体失败", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "读取 PAR 响应失败",
		})
	}

	// 检查响应体是否为空
	if len(body) == 0 {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "PAR 响应为空",
		})
	}

	var parResponse map[string]interface{}
	if err := json.Unmarshal(body, &parResponse); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "无法解析 PAR 响应",
		})
	}
	log.Infof("HandleOAuthLogin， parResponse: %+v, redirectURI: %s", parResponse, redirectURI)
	requestURI := parResponse["request_uri"].(string)

	// 保存 OAuth 授权请求到数据库
	dpopPrivateJWKStr, err := dpopPrivateJWK.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", dpopPrivateJWKStr)

	authRequest := database.OAuthAuthRequest{
		State:               state,
		AuthserverIss:       authserverMeta["issuer"].(string),
		Did:                 did,
		Handle:              handle,
		PdsUrl:              pdsURL,
		PkceVerifier:        pkceVerifier,
		Scope:               scope,
		DpopAuthserverNonce: dpopAuthserverNonce,
		DpopPrivateJwk:      string(dpopPrivateJWKStr),
		Platform:            platform,
		ReturnURI:           "",
	}

	log.Infof("HandleOAuthLogin， authRequest: %+v", authRequest)

	if err := database.InsertOAuthAuthRequest(a.metaStore.DB, &authRequest); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存授权请求失败",
		})
	}

	// 重定向用户到授权服务器完成浏览器身份验证流程
	authURL := authserverMeta["authorization_endpoint"].(string)
	if !utils.IsSafeURL(authURL) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "不安全的授权 URL",
		})
	}

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("request_uri", requestURI)

	return c.Redirect(http.StatusFound, authURL+"?"+params.Encode())
}

func (a *AvatarAIAPI) HandleOAuthCallback(c echo.Context) error {
	state := c.QueryParam("state")
	authserverISS := c.QueryParam("iss")
	authorizationCode := c.QueryParam("code")

	// 通过 state 令牌查找授权请求
	authRequest, err := database.GetOAuthAuthRequest(a.metaStore.DB, state)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "未找到 OAuth 请求",
		})
	}

	// 删除行以防止响应重放
	if err := database.DeleteOAuthAuthRequest(a.metaStore.DB, state); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "无法删除授权请求",
		})
	}

	// 验证查询参数 "iss" 与早期 oauth 请求 "iss" 相符
	if authRequest.AuthserverIss != authserverISS {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "授权服务器不匹配",
		})
	}

	// 通过从授权服务器请求身份验证令牌来完成身份验证流程
	appURL := utils.GetAPPURL(c)
	tokens, dpopAuthserverNonce, err := atproto.InitialTokenRequest(
		authRequest,
		authorizationCode,
		atproto.BuildClientID(appURL, authRequest.Platform),
		atproto.BuildRedirectURL(appURL, authRequest.Platform),
		a.Config.ATP.ClientSecretJWK(),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取令牌失败: " + err.Error(),
		})
	}

	log.Infof("HandleOAuthCallback， tokens: %+v", tokens)

	// 验证账号身份与原始请求
	var did, handle, pdsURL string
	if authRequest.Did != "" {
		// 如果我们以账号标识符开始，这很简单
		did, handle, pdsURL = authRequest.Did, authRequest.Handle, authRequest.PdsUrl
		if tokens["sub"].(string) != did {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "令牌主题与请求的 DID 不匹配",
			})
		}
	} else {
		// 如果我们以授权服务器 URL 开始，现在需要解析身份
		did = tokens["sub"].(string)
		if !atproto.IsValidDID(did) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无效的 DID",
			})
		}

		ident, err := atproto.ResolveIdentity(context.Background(), did)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析身份: " + err.Error(),
			})
		}
		pdsURL = atproto.PDSEndpoint(ident)
		authserverURL, err := atproto.ResolvePDSAuthserver(pdsURL)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析授权服务器 URL: " + err.Error(),
			})
		}

		// 验证授权服务器匹配
		if authserverURL != authserverISS {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "授权服务器不匹配",
			})
		}
	}

	// 验证返回的作用域与请求匹配
	if authRequest.Scope != tokens["scope"].(string) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "作用域不匹配",
		})
	}

	// 保存会话(包括身份验证令牌)到数据库
	oauthSession := database.OAuthSession{
		Did:                 did,
		Handle:              handle,
		PdsUrl:              pdsURL,
		AuthserverIss:       authserverISS,
		AccessToken:         tokens["access_token"].(string),
		RefreshToken:        tokens["refresh_token"].(string),
		DpopAuthserverNonce: dpopAuthserverNonce,
		DpopPrivateJwk:      authRequest.DpopPrivateJwk,
		ExpiresIn:           utils.ConvertInt64(tokens["expires_in"]),
		CreatedAt:           time.Now(),
		Platform:            authRequest.Platform,
		ReturnURI:           "",
	}

	if err := database.SaveOAuthSession(a.metaStore.DB, &oauthSession); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存会话失败",
		})
	}

	avatar, err := database.GetOrCreateAvatar(a.metaStore.DB, did, handle, pdsURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "创建或获取 Avatar 失败",
		})
	}

	code, err := utils.GenerateCode()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成授权码失败",
		})
	}

	oauthCode := database.OAuthCode{
		Code:           code,
		OAuthSessionID: oauthSession.ID,
		AvatarDid:      avatar.Did,
		Used:           false,
		Platform:       authRequest.Platform,
		ReturnURI:      "",
		ExpiresAt:      time.Now().Add(time.Hour * 24 * 30),
	}
	if err := database.SaveOAuthCode(a.metaStore.DB, &oauthCode); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存授权码失败",
		})
	}

	redirectURI := atproto.BuildCallbackRedirectURI(appURL, authRequest.Platform)
	return c.Redirect(http.StatusFound, redirectURI+fmt.Sprintf("?code=%s", code))
}

type ExchangeTokenRequest struct {
	Code string `json:"code"`
}

func (a *AvatarAIAPI) HandleOAuthToken(c echo.Context) error {
	request := ExchangeTokenRequest{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法解析请求",
		})
	}

	oauthCode, err := database.GetOAuthCode(a.metaStore.DB, request.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法找到授权码",
		})
	}

	// check code is used or expired
	if oauthCode.Used || oauthCode.ExpiresAt.Before(time.Now()) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "授权码已使用或已过期",
		})
	}

	oauthSession, err := database.GetOauthSessionByID(a.metaStore.DB, oauthCode.OAuthSessionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法找到授权会话",
		})
	}
	avatar, err := database.GetAvatarByDID(a.metaStore.DB, oauthCode.AvatarDid)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到用户信息",
		})
	}

	sessionID := uuid.New().String()
	// 生成自己服务的访问令牌和刷新令牌
	accessToken, err := utils.GenerateAccessToken(a.Config, sessionID, avatar)
	if err != nil {
		log.Errorf("HandleOAuthCallback，生成访问令牌失败: %+v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成访问令牌失败",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(a.Config, sessionID)
	if err != nil {
		log.Errorf("HandleOAuthCallback，生成刷新令牌失败: %+v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成刷新令牌失败",
		})
	}

	session := database.Session{
		ID:             sessionID,
		AvatarDid:      oauthSession.Did,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		OAuthSessionID: oauthSession.ID,
		ExpiredAt:      time.Now().Add(time.Hour * 24 * 30),
		Platform:       oauthSession.Platform,
	}
	if err := database.SaveSession(a.metaStore.DB, &session); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存会话失败",
		})
	}

	// 同时设置JWT令牌到cookie中
	SetTokenCookie(c.Response().Writer, "avatarai_token", accessToken, "", true, true)

	// 返回访问令牌和刷新令牌
	return c.JSON(http.StatusOK, map[string]string{
		"did":           oauthSession.Did,
		"handle":        oauthSession.Handle,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    "86400", // 24小时，您可以根据需要调整
	})
}

func (a *AvatarAIAPI) HandleOAuthRefresh(c echo.Context) error {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法解析请求",
		})
	}

	// 验证刷新令牌
	sessionID, err := utils.ValidateRefreshToken(a.Config, request.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无效的刷新令牌: " + err.Error(),
		})
	}

	// 获取会话信息
	session, err := database.GetSessionByID(a.metaStore.DB, sessionID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到会话",
		})
	}

	// 获取 OAuth 会话信息
	oauthSession, err := database.GetOauthSessionByID(a.metaStore.DB, session.OAuthSessionID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到 OAuth 会话",
		})
	}

	// 获取用户信息
	avatar, err := database.GetAvatarByDID(a.metaStore.DB, session.AvatarDid)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到用户信息",
		})
	}

	// 刷新 ATP OAuth 令牌
	appURL := utils.GetAPPURL(c)
	atpTokens, dpopAuthserverNonce, err := atproto.RefreshTokenRequest(
		oauthSession,
		atproto.BuildClientID(appURL, oauthSession.Platform),
		atproto.BuildRedirectURL(appURL, oauthSession.Platform),
		a.Config.ATP.ClientSecretJWK(),
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "刷新 ATP 令牌失败: " + err.Error(),
		})
	}

	// 将更新的 ATP 令牌保存到数据库
	oauthSession.AccessToken = atpTokens["access_token"].(string)
	oauthSession.RefreshToken = atpTokens["refresh_token"].(string)
	oauthSession.DpopAuthserverNonce = dpopAuthserverNonce

	if err := database.UpdateOAuthSession(a.metaStore.DB, oauthSession); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "更新 OAuth 会话失败",
		})
	}

	// 生成新的本地访问令牌和刷新令牌
	accessToken, err := utils.GenerateAccessToken(a.Config, session.ID, avatar)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成访问令牌失败: " + err.Error(),
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(a.Config, session.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成刷新令牌失败: " + err.Error(),
		})
	}

	SetTokenCookie(c.Response().Writer, "avatarai_token", accessToken, "", true, true)

	// 返回新的令牌
	return c.JSON(http.StatusOK, map[string]string{
		"did":           avatar.Did,
		"handle":        avatar.Handle,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    "86400", // 24小时
	})
}

func (a *AvatarAIAPI) HandleOAuthLogout(c echo.Context) error {
	userDID, err := c.Cookie("user_did")
	if err == nil {
		database.DeleteOAuthSessionByDID(a.metaStore.DB, userDID.Value)
	}

	c.SetCookie(&http.Cookie{
		Name:   "user_did",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	c.SetCookie(&http.Cookie{
		Name:   "user_handle",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	return c.Redirect(http.StatusFound, "/")
}

func (a *AvatarAIAPI) HandleUserProfile(c echo.Context) error {
	ac := c.(*utils.AvatarAIContext)
	avatar := ac.Avatar
	oauthSession := ac.OauthSession
	if avatar == nil {
		return ac.Redirect(http.StatusFound, "/api/oauth/login")
	}

	xrpcCli := atproto.NewXrpcClient(oauthSession, a.metaStore.DB)
	authArgs := atproto.GetOauthSessionAuthArgs(oauthSession)
	var atpSession comatproto.ServerGetSession_Output
	if err := xrpcCli.Do(c.Request().Context(), authArgs, xrpc.Query, "", "com.atproto.server.getSession", nil, nil, &atpSession); err != nil {
		return err
	}

	return ac.JSON(http.StatusOK, map[string]any{
		"user_did": avatar.Did,
		"handle":   atpSession.Did,
		"atp":      atpSession,
	})
}

func (a *AvatarAIAPI) HandleBskyPost(c echo.Context) error {
	userDID, err := c.Cookie("user_did")
	if err != nil {
		return c.Redirect(http.StatusFound, "/api/oauth/login")
	}

	session, err := database.GetOAuthSessionByDID(a.metaStore.DB, userDID.Value)
	if err != nil {
		return c.Redirect(http.StatusFound, "/api/oauth/login")
	}

	if c.Request().Method == http.MethodGet {
		return c.Render(http.StatusOK, "bsky_post.html", nil)
	}

	postText := c.FormValue("post_text")

	pdsURL := session.PdsUrl
	reqURL := pdsURL + "/xrpc/com.atproto.repo.createRecord"

	now := time.Now().UTC().Format(time.RFC3339)
	body := map[string]interface{}{
		"repo":       session.Did,
		"collection": "app.bsky.feed.post",
		"record": map[string]interface{}{
			"$type":     "app.bsky.feed.post",
			"text":      postText,
			"createdAt": now,
		},
	}

	_, err = pdsAuthedReq("POST", reqURL, body, session)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "发布失败: " + err.Error(),
		})
	}

	return c.Render(http.StatusOK, "bsky_post.html", map[string]interface{}{
		"message": "帖子已在 PDS 中创建！",
	})
}

func pdsAuthedReq(method string, url string, body interface{}, session *database.OAuthSession) (string, error) {
	// 实现对 PDS 进行身份验证请求的逻辑
	return "", nil
}

func SetTokenCookie(w http.ResponseWriter, cookieName, jwtToken, domain string, httpOnly, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    jwtToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   8 * 60 * 60,
		HttpOnly: httpOnly,
		Secure:   secure,
	})
}

func UnsetTokenCookie(w http.ResponseWriter, cookieName, domain string, httpOnly, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: httpOnly,
		Secure:   secure,
	})
}
