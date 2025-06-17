package handlers

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
	"strings"
	"time"

	indigo "github.com/bluesky-social/indigo/api/atproto"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"github.com/zhongshangwu/avatarai-social/types"
)

type OAuthHandler struct {
	config        *config.SocialConfig
	metaStore     *repositories.MetaStore
	client        *atproto.OAuthClient
	appReturnHTML string
}

type JWKS struct {
	Keys []interface{} `json:"keys"`
}

type LoginPageData struct {
	User     *LoginPageUserData
	Messages []string
	URLs     map[string]string
}

type LoginPageUserData struct {
	Handle string
}

type ExchangeTokenRequest struct {
	Code string `json:"code"`
}

func NewOAuthHandler(config *config.SocialConfig, metaStore *repositories.MetaStore, appReturnHTML string) *OAuthHandler {
	appURL := fmt.Sprintf("http://%s/", config.Server.HTTP.Address)
	client := atproto.NewOAuthClient(appURL, config.ATP.ClientSecretJWK())
	return &OAuthHandler{
		config:        config,
		metaStore:     metaStore,
		client:        client,
		appReturnHTML: appReturnHTML,
	}
}

func (h *OAuthHandler) OAuthClientMetadata(c echo.Context) error {
	appURL := utils.GetAPPURL(c)
	platform := c.Param("platform")
	clientID := atproto.BuildClientID(appURL, platform)

	// 统一使用 web 客户端模式，因为实际上是服务端作为 OAuth 客户端
	// 不管是 web 端还是移动端，都是通过服务端进行 OAuth 认证
	// 服务端可以安全地存储客户端密钥，因此应该使用 private_key_jwt 认证方法
	options := &atproto.ClientMetadataOptions{
		ClientName:              "AvatarAI Social",
		TokenEndpointAuthMethod: "private_key_jwt",
		JwksURI:                 appURL + "api/oauth/jwks.json",
		UsePrivateKeyJWT:        true,
	}

	metadata := atproto.GetClientMetadataWithOptions(appURL, platform, clientID, options)
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
		// 对于web平台，重定向到前端页面
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/")
		}

		// 对于其他平台，使用模板渲染
		data := LoginPageData{
			User: nil,
			URLs: map[string]string{"OAuthLogin": "/api/oauth/signin?platform=" + platform},
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
			if platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析身份: "+err.Error()))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析身份: " + err.Error(),
			})
		}
		did = string(ident.DID)
		handle = string(ident.Handle)
		pdsURL = atproto.PDSEndpoint(ident)
		authserverURL, err = atproto.ResolvePDSAuthserver(pdsURL)
		if err != nil {
			if platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析授权服务器 URL: "+err.Error()))
			}
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
			if platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析授权服务器 URL: "+err.Error()))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析授权服务器 URL: " + err.Error(),
			})
		}
	} else {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("不是有效的 handle、DID 或授权服务器 URL"))
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "不是有效的 handle、DID 或授权服务器 URL",
		})
	}

	// 获取授权服务器元数据
	authserverMeta, err := atproto.FetchAuthserverMeta(authserverURL)
	if err != nil {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法获取授权服务器元数据: "+err.Error()))
		}
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
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法生成 DPoP 私钥"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "无法生成 DPoP 私钥",
		})
	}

	// 请求的 OAuth 范围
	scope := "atproto transition:generic"

	appURL := utils.GetAPPURL(c)
	redirectURI := atproto.BuildRedirectURL(appURL, platform)
	clientID := atproto.BuildClientID(appURL, platform)

	log.Infof("HandleOAuthLogin， clientID: %s, redirectURI: %s", clientID, redirectURI)

	// 提交 OAuth Pushed Authentication Request (PAR)
	pkceVerifier, state, dpopAuthserverNonce, resp, err := atproto.SendPARAuthRequest(
		authserverURL,
		authserverMeta,
		loginHint,
		clientID,
		redirectURI,
		scope,
		h.config.ATP.ClientSecretJWK(),
		dpopPrivateJWK,
	)
	if err != nil {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("PAR 请求失败: "+err.Error()))
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "PAR 请求失败: " + err.Error(),
		})
	}

	// 获取请求 URI
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Error("读取 PAR 响应体失败", "error", err)
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("读取 PAR 响应失败"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "读取 PAR 响应失败",
		})
	}

	// 检查响应体是否为空
	if len(body) == 0 {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("PAR 响应为空"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "PAR 响应为空",
		})
	}

	var parResponse map[string]interface{}
	if err := json.Unmarshal(body, &parResponse); err != nil {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析 PAR 响应"))
		}
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

	authRequest := repositories.OAuthAuthRequest{
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

	if err := h.metaStore.OAuthRepo.InsertOAuthAuthRequest(&authRequest); err != nil {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("保存授权请求失败"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存授权请求失败",
		})
	}

	// 重定向用户到授权服务器完成浏览器身份验证流程
	authURL := authserverMeta["authorization_endpoint"].(string)
	if !utils.IsSafeURL(authURL) {
		if platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("不安全的授权 URL"))
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "不安全的授权 URL",
		})
	}

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("request_uri", requestURI)

	return c.Redirect(http.StatusFound, authURL+"?"+params.Encode())
}

func (h *OAuthHandler) HandleOAuthCallback(c echo.Context) error {
	state := c.QueryParam("state")
	authserverISS := c.QueryParam("iss")
	authorizationCode := c.QueryParam("code")

	// 通过 state 令牌查找授权请求
	authRequest, err := h.metaStore.OAuthRepo.GetOAuthAuthRequest(state)
	if err != nil {
		// 对于web平台，重定向到前端页面显示错误
		return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("未找到 OAuth 请求"))
	}

	// 删除行以防止响应重放
	if err := h.metaStore.OAuthRepo.DeleteOAuthAuthRequest(state); err != nil {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法删除授权请求"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "无法删除授权请求",
		})
	}

	// 验证查询参数 "iss" 与早期 oauth 请求 "iss" 相符
	if authRequest.AuthserverIss != authserverISS {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("授权服务器不匹配"))
		}
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
		h.config.ATP.ClientSecretJWK(),
	)
	if err != nil {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("获取令牌失败: "+err.Error()))
		}
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
			if authRequest.Platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("令牌主题与请求的 DID 不匹配"))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "令牌主题与请求的 DID 不匹配",
			})
		}
	} else {
		// 如果我们以授权服务器 URL 开始，现在需要解析身份
		did = tokens["sub"].(string)
		if !atproto.IsValidDID(did) {
			if authRequest.Platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无效的 DID"))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无效的 DID",
			})
		}

		ident, err := atproto.ResolveIdentity(context.Background(), did)
		if err != nil {
			if authRequest.Platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析身份: "+err.Error()))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析身份: " + err.Error(),
			})
		}
		pdsURL = atproto.PDSEndpoint(ident)
		authserverURL, err := atproto.ResolvePDSAuthserver(pdsURL)
		if err != nil {
			if authRequest.Platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("无法解析授权服务器 URL: "+err.Error()))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "无法解析授权服务器 URL: " + err.Error(),
			})
		}

		// 验证授权服务器匹配
		if authserverURL != authserverISS {
			if authRequest.Platform == "web" {
				return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("授权服务器不匹配"))
			}
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "授权服务器不匹配",
			})
		}
	}

	// 验证返回的作用域与请求匹配
	if authRequest.Scope != tokens["scope"].(string) {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("作用域不匹配"))
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "作用域不匹配",
		})
	}

	// 保存会话(包括身份验证令牌)到数据库
	oauthSession := repositories.OAuthSession{
		Did:                 did,
		Handle:              handle,
		PdsUrl:              pdsURL,
		AuthserverIss:       authserverISS,
		AccessToken:         tokens["access_token"].(string),
		RefreshToken:        tokens["refresh_token"].(string),
		DpopAuthserverNonce: dpopAuthserverNonce,
		DpopPrivateJwk:      authRequest.DpopPrivateJwk,
		ExpiresIn:           utils.ConvertInt64(tokens["expires_in"]),
		CreatedAt:           time.Now().Unix(),
		Platform:            authRequest.Platform,
		ReturnURI:           "",
	}

	if err := h.metaStore.OAuthRepo.SaveOAuthSession(&oauthSession); err != nil {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("保存会话失败"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存会话失败",
		})
	}

	avatar, err := h.metaStore.UserRepo.GetOrCreateAvatar(did, handle, pdsURL)
	if err != nil {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("创建或获取 Avatar 失败"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "创建或获取 Avatar 失败",
		})
	}

	code, err := utils.GenerateCode()
	if err != nil {
		if authRequest.Platform == "web" {
			return c.Redirect(http.StatusFound, "/?error="+url.QueryEscape("生成授权码失败"))
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成授权码失败",
		})
	}

	oauthCode := repositories.OAuthCode{
		Code:           code,
		OAuthSessionID: oauthSession.ID,
		UserDid:        avatar.Did,
		Used:           false,
		Platform:       authRequest.Platform,
		ReturnURI:      "",
		ExpiresAt:      time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	if err := h.metaStore.OAuthRepo.SaveOAuthCode(&oauthCode); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "保存授权码失败",
		})
	}

	// 对于web平台，重定向到前端页面
	if authRequest.Platform == "web" {
		return c.Redirect(http.StatusFound, fmt.Sprintf("/?code=%s", code))
	}

	// 对于其他平台，使用原来的重定向逻辑
	redirectURI := atproto.BuildCallbackRedirectURI(appURL, authRequest.Platform)
	return c.Redirect(http.StatusFound, redirectURI+fmt.Sprintf("?code=%s", code))
}

func (h *OAuthHandler) HandleOAuthToken(c echo.Context) error {
	code := c.QueryParam("code")

	// 支持POST请求的JSON格式
	if c.Request().Method == http.MethodPost {
		request := ExchangeTokenRequest{}
		if err := c.Bind(&request); err != nil {
			log.Errorf("HandleOAuthToken，从请求中获取code失败: %+v", err)
		} else if request.Code != "" {
			code = request.Code
		}
	}

	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法解析请求",
		})
	}

	oauthCode, err := h.metaStore.OAuthRepo.GetOAuthCode(code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法找到授权码",
		})
	}
	currentTime := time.Now().Unix()

	if oauthCode.Used || oauthCode.ExpiresAt <= currentTime {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "授权码已使用或已过期",
		})
	}

	oauthSession, err := h.metaStore.OAuthRepo.GetOAuthSessionByID(oauthCode.OAuthSessionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "无法找到授权会话",
		})
	}
	avatar, err := h.metaStore.UserRepo.GetAvatarByDID(oauthCode.UserDid)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到用户信息",
		})
	}

	sessionID := uuid.New().String()
	// 生成自己服务的访问令牌和刷新令牌
	accessToken, err := utils.GenerateAccessToken(h.config, sessionID, avatar)
	if err != nil {
		log.Errorf("HandleOAuthCallback，生成访问令牌失败: %+v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成访问令牌失败",
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(h.config, sessionID)
	if err != nil {
		log.Errorf("HandleOAuthCallback，生成刷新令牌失败: %+v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成刷新令牌失败",
		})
	}

	session := repositories.Session{
		ID:             sessionID,
		UserDid:        oauthCode.UserDid,
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		OAuthSessionID: oauthSession.ID,
		ExpiredAt:      time.Now().Add(time.Hour * 24 * 30).Unix(),
		Platform:       oauthSession.Platform,
	}
	if err := h.metaStore.OAuthRepo.SaveSession(&session); err != nil {
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

func (h *OAuthHandler) HandleOAuthRefresh(c *types.APIContext) error {
	refreshToken := c.RefreshToken()
	// 验证刷新令牌
	log.Infof("HandleOAuthRefresh， refreshToken: %+v", refreshToken)
	sessionID, err := utils.ValidateRefreshToken(h.config, refreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无效的刷新令牌: " + err.Error(),
		})
	}

	session, err := h.metaStore.OAuthRepo.GetSessionByID(sessionID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到会话",
		})
	}

	oauthSession, err := h.metaStore.OAuthRepo.GetOAuthSessionByID(session.OAuthSessionID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到 OAuth 会话",
		})
	}

	avatar, err := h.metaStore.UserRepo.GetAvatarByDID(oauthSession.Did)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "无法找到用户信息",
		})
	}

	appURL := utils.GetAPPURL(c)
	refreshReq := &atproto.RefreshRequest{
		SessionDID:  oauthSession.Did,
		Platform:    oauthSession.Platform,
		RedirectURI: atproto.BuildRedirectURL(appURL, oauthSession.Platform),
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

	// 生成新的本地访问令牌和刷新令牌
	accessToken, err := utils.GenerateAccessToken(h.config, session.ID, avatar)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成访问令牌失败: " + err.Error(),
		})
	}

	refreshToken, err = utils.GenerateRefreshToken(h.config, session.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "生成刷新令牌失败: " + err.Error(),
		})
	}

	SetTokenCookie(c.Response().Writer, "avatarai_token", accessToken, "", true, true)

	return c.JSON(http.StatusOK, map[string]string{
		"did":           avatar.Did,
		"handle":        avatar.Handle,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    "86400", // 24小时
	})
}

func (h *OAuthHandler) HandleOAuthLogout(c *types.APIContext) error {
	oauthSession := c.OauthSession
	if oauthSession != nil {
		if err := h.metaStore.OAuthRepo.DeleteOAuthSessionByDID(oauthSession.Did); err != nil {
			log.Warnf("删除会话失败: %v", err)
		}
	}

	return h.successResponse(c, map[string]string{
		"message": "登出成功",
	})
}

func (a *OAuthHandler) HandleAppReturn(c echo.Context) error {
	bundleID := c.Param("bundleID")
	if bundleID == "" {
		return c.JSON(http.StatusNotImplemented, map[string]string{
			"error": "server has no --app-bundle-id set",
		})
	}

	html := strings.Replace(a.appReturnHTML, "APP_BUNDLE_ID_REPLACE_ME", bundleID, 1)
	return c.HTML(http.StatusOK, html)
}

func (a *OAuthHandler) HandleBskyPost(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}

	userDID := c.User.Did
	session, err := a.metaStore.OAuthRepo.GetOAuthSessionByDID(userDID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "获取会话失败: " + err.Error(),
		})
	}

	if c.Request().Method == http.MethodGet {
		return c.Render(http.StatusOK, "bsky_post.html", nil)
	}

	postText := c.FormValue("post_text")

	xrpcCli, err := atproto.NewXrpcClient(c.OauthSession)
	if err != nil {
		return c.InternalServerError("创建 XRPC 客户端失败: " + err.Error())
	}

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

	putOutput := indigo.RepoCreateRecord_Output{}

	err = xrpcCli.Procedure(c.Request().Context(), "com.atproto.repo.createRecord", nil, body, &putOutput)
	if err != nil {
		return c.InternalServerError("创建Aster PDS记录失败: " + err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "帖子已在 PDS 中创建！",
		"record":  putOutput,
	})
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
