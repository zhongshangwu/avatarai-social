package atproto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	neturl "net/url"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type OAuthClient struct {
	clientSecretJWK jose.JSONWebKey
	appURL          string
}

type AuthRequest struct {
	LoginHint   string
	Platform    string
	Scope       string
	RedirectURI string
}

type AuthResponse struct {
	AuthURL        string
	State          string
	PKCEVerifier   string
	DpopNonce      string
	AuthserverMeta map[string]interface{}
}

type TokenRequest struct {
	Code        string
	State       string
	Platform    string
	RedirectURI string
}

type TokenResponse struct {
	AccessToken         string
	RefreshToken        string
	DpopAuthserverNonce string
	ExpiresIn           int64
	TokenType           string
	Scope               string
}

type RefreshRequest struct {
	SessionDID  string
	Platform    string
	RedirectURI string
}

type PDSRequest struct {
	Method     string
	URL        string
	SessionDID string
	Body       interface{}
}

func NewOAuthClient(appURL string, clientSecretJWK jose.JSONWebKey) *OAuthClient {
	return &OAuthClient{
		clientSecretJWK: clientSecretJWK,
		appURL:          appURL,
	}
}

// StartAuth 开始授权流程
func (c *OAuthClient) StartAuth(req *AuthRequest) (*AuthResponse, error) {
	// 解析 PDS 授权服务器
	authserverURL, err := c.resolvePDSAuthserver(req.LoginHint)
	if err != nil {
		return nil, fmt.Errorf("解析 PDS 授权服务器失败: %w", err)
	}

	// 获取授权服务器元数据
	authserverMeta, err := c.fetchAuthserverMeta(authserverURL)
	if err != nil {
		return nil, fmt.Errorf("获取授权服务器元数据失败: %w", err)
	}

	// 生成 DPoP 密钥对
	dpopPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成 DPoP 密钥失败: %w", err)
	}

	dpopPrivateJWK := jose.JSONWebKey{
		Key:       dpopPrivateKey,
		KeyID:     c.generateToken(16),
		Algorithm: string(jose.ES256),
		Use:       "sig",
	}

	// 构建客户端 ID 和重定向 URI
	clientID := c.buildClientID(req.Platform)
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = c.buildRedirectURL(req.Platform)
	}

	scope := req.Scope
	if scope == "" {
		scope = "atproto transition:generic"
	}

	// 发送 PAR 请求
	pkceVerifier, state, dpopAuthserverNonce, resp, err := c.sendPARAuthRequest(
		authserverURL,
		authserverMeta,
		req.LoginHint,
		clientID,
		redirectURI,
		scope,
		dpopPrivateJWK,
	)
	if err != nil {
		return nil, fmt.Errorf("发送 PAR 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PAR 请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析 PAR 响应
	var parResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&parResponse); err != nil {
		return nil, fmt.Errorf("解析 PAR 响应失败: %w", err)
	}

	requestURI, ok := parResponse["request_uri"].(string)
	if !ok {
		return nil, fmt.Errorf("PAR 响应中缺少 request_uri")
	}

	// 构建授权 URL
	authURL := fmt.Sprintf("%s?client_id=%s&request_uri=%s",
		authserverMeta["authorization_endpoint"].(string),
		clientID,
		requestURI,
	)

	return &AuthResponse{
		AuthURL:        authURL,
		State:          state,
		PKCEVerifier:   pkceVerifier,
		DpopNonce:      dpopAuthserverNonce,
		AuthserverMeta: authserverMeta,
	}, nil
}

// ExchangeToken 交换授权码获取令牌
func (c *OAuthClient) ExchangeToken(req *TokenRequest, authRequest *repositories.OAuthAuthRequest) (*TokenResponse, error) {
	clientID := c.buildClientID(req.Platform)
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = c.buildRedirectURL(req.Platform)
	}

	// 调用内部令牌请求函数
	tokens, dpopAuthserverNonce, err := c.initialTokenRequest(
		authRequest,
		req.Code,
		clientID,
		redirectURI,
	)
	if err != nil {
		return nil, fmt.Errorf("令牌交换失败: %w", err)
	}

	return &TokenResponse{
		AccessToken:         tokens["access_token"].(string),
		RefreshToken:        tokens["refresh_token"].(string),
		DpopAuthserverNonce: dpopAuthserverNonce,
		ExpiresIn:           int64(tokens["expires_in"].(float64)),
		TokenType:           tokens["token_type"].(string),
		Scope:               tokens["scope"].(string),
	}, nil
}

// RefreshToken 刷新访问令牌
func (c *OAuthClient) RefreshToken(req *RefreshRequest, session *repositories.OAuthSession) (*TokenResponse, error) {
	clientID := c.buildClientID(req.Platform)
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = c.buildRedirectURL(req.Platform)
	}

	// 调用内部刷新令牌函数
	tokens, dpopAuthserverNonce, err := c.refreshTokenRequest(
		session,
		clientID,
		redirectURI,
	)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌失败: %w", err)
	}

	return &TokenResponse{
		AccessToken:         tokens["access_token"].(string),
		RefreshToken:        tokens["refresh_token"].(string),
		DpopAuthserverNonce: dpopAuthserverNonce,
		ExpiresIn:           int64(tokens["expires_in"].(float64)),
		TokenType:           tokens["token_type"].(string),
		Scope:               tokens["scope"].(string),
	}, nil
}

// IsSessionExpired 检查会话是否过期
func (c *OAuthClient) IsSessionExpired(session *repositories.OAuthSession) bool {
	return time.Now().Unix() > session.CreatedAt.Unix()+session.ExpiresIn
}

// GenerateClientMetadata 生成客户端元数据
func (c *OAuthClient) GenerateClientMetadata(platform string) map[string]interface{} {
	clientID := c.buildClientID(platform)

	return map[string]interface{}{
		"client_id":                       clientID,
		"dpop_bound_access_tokens":        true,
		"application_type":                "web",
		"redirect_uris":                   []string{c.appURL + "api/oauth/callback"},
		"grant_types":                     []string{"authorization_code", "refresh_token"},
		"response_types":                  []string{"code"},
		"scope":                           "atproto transition:generic",
		"token_endpoint_auth_method":      "private_key_jwt",
		"token_endpoint_auth_signing_alg": "ES256",
		"jwks_uri":                        c.appURL + "api/oauth/jwks.json",
		"client_name":                     "ATProto OAuth Go Backend",
		"client_uri":                      c.appURL,
	}
}

// GenerateJWKS 生成 JWKS
func (c *OAuthClient) GenerateJWKS() map[string]interface{} {
	return map[string]interface{}{
		"keys": []interface{}{c.clientSecretJWK.Public()},
	}
}

// 通用 HTTP 客户端创建方法
func (c *OAuthClient) createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

// 通用 HTTP 请求方法
func (c *OAuthClient) makeHTTPRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	if !utils.IsSafeURL(url) {
		return nil, fmt.Errorf("不安全的 URL: %s", url)
	}

	client := c.createHTTPClient()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认 Content-Type
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// 添加自定义头部
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return client.Do(req)
}

// fetchAuthserverMeta 获取授权服务器元数据
func (c *OAuthClient) fetchAuthserverMeta(url string) (map[string]interface{}, error) {
	resp, err := c.makeHTTPRequest("GET", fmt.Sprintf("%s/.well-known/oauth-authorization-server", url), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var authserverMeta map[string]interface{}
	if err := json.Unmarshal(body, &authserverMeta); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if !c.isValidAuthserverMeta(authserverMeta, url) {
		return nil, fmt.Errorf("无效的授权服务器元数据")
	}

	return authserverMeta, nil
}

// isValidAuthserverMeta 验证授权服务器元数据
func (c *OAuthClient) isValidAuthserverMeta(meta map[string]interface{}, url string) bool {
	fetchURL, err := neturl.Parse(url)
	if err != nil {
		return false
	}

	issuer, ok := meta["issuer"].(string)
	if !ok {
		return false
	}

	issuerURL, err := neturl.Parse(issuer)
	if err != nil {
		return false
	}

	// 验证 issuer URL 的各个部分
	if issuerURL.Hostname() != fetchURL.Hostname() {
		return false
	}
	if issuerURL.Scheme != "https" {
		return false
	}
	if issuerURL.Port() != "" {
		return false
	}
	if issuerURL.Path != "" && issuerURL.Path != "/" {
		return false
	}
	if issuerURL.RawQuery != "" {
		return false
	}
	if issuerURL.Fragment != "" {
		return false
	}

	// 检查必需的支持
	if !c.containsString(c.getStringSlice(meta, "response_types_supported"), "code") {
		return false
	}

	grantTypes := c.getStringSlice(meta, "grant_types_supported")
	if !c.containsString(grantTypes, "authorization_code") || !c.containsString(grantTypes, "refresh_token") {
		return false
	}

	if !c.containsString(c.getStringSlice(meta, "code_challenge_methods_supported"), "S256") {
		return false
	}

	authMethods := c.getStringSlice(meta, "token_endpoint_auth_methods_supported")
	if !c.containsString(authMethods, "none") || !c.containsString(authMethods, "private_key_jwt") {
		return false
	}

	if !c.containsString(c.getStringSlice(meta, "token_endpoint_auth_signing_alg_values_supported"), "ES256") {
		return false
	}

	if !c.containsString(c.getStringSlice(meta, "scopes_supported"), "atproto") {
		return false
	}

	authRespIssParam, ok := meta["authorization_response_iss_parameter_supported"].(bool)
	if !ok || !authRespIssParam {
		return false
	}

	parEndpoint, ok := meta["pushed_authorization_request_endpoint"]
	if !ok || parEndpoint == nil {
		return false
	}

	requirePAR, ok := meta["require_pushed_authorization_requests"].(bool)
	if !ok || !requirePAR {
		return false
	}

	if !c.containsString(c.getStringSlice(meta, "dpop_signing_alg_values_supported"), "ES256") {
		return false
	}

	if requireRequestURIReg, ok := meta["require_request_uri_registration"].(bool); ok && !requireRequestURIReg {
		return false
	}

	clientIDMetadata, ok := meta["client_id_metadata_document_supported"].(bool)
	if !ok || !clientIDMetadata {
		return false
	}

	return true
}

// resolvePDSAuthserver 解析 PDS 授权服务器
func (c *OAuthClient) resolvePDSAuthserver(url string) (string, error) {
	resp, err := c.makeHTTPRequest("GET", fmt.Sprintf("%s/.well-known/oauth-protected-resource", url), nil, nil)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务器返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var respData struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if len(respData.AuthorizationServers) == 0 {
		return "", fmt.Errorf("未找到授权服务器")
	}

	return respData.AuthorizationServers[0], nil
}

// generateToken 生成随机令牌
func (c *OAuthClient) generateToken(length ...int) string {
	tokenLength := 32
	if len(length) > 0 && length[0] > 0 {
		tokenLength = length[0]
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, tokenLength)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// createS256CodeChallenge 创建 S256 代码挑战
func (c *OAuthClient) createS256CodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "=")
}

// clientAssertionJWT 创建客户端断言 JWT
func (c *OAuthClient) clientAssertionJWT(clientID string, authserverURL string) (string, error) {
	now := time.Now()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.ES256,
		Key:       c.clientSecretJWK.Key,
	}, (&jose.SignerOptions{}).
		WithType("jwt").
		WithHeader("kid", c.clientSecretJWK.KeyID))
	if err != nil {
		return "", fmt.Errorf("创建签名者失败: %w", err)
	}

	claims := map[string]interface{}{
		"iss": clientID,
		"sub": clientID,
		"aud": []string{authserverURL},
		"jti": c.generateToken(),
		"iat": now.Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("序列化 claims 失败: %w", err)
	}

	signature, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	token, err := signature.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("序列化 token 失败: %w", err)
	}

	return token, nil
}

// authserverDpopJWT 创建授权服务器 DPoP JWT
func (c *OAuthClient) authserverDpopJWT(method string, url string, nonce string, dpopPrivateJWK jose.JSONWebKey) (string, error) {
	now := time.Now()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.ES256,
		Key:       dpopPrivateJWK.Key,
	}, (&jose.SignerOptions{}).
		WithType("dpop+jwt").
		WithHeader("jwk", dpopPrivateJWK.Public()))
	if err != nil {
		return "", fmt.Errorf("创建签名者失败: %w", err)
	}

	claims := map[string]interface{}{
		"typ": "dpop+jwt",
		"jti": c.generateToken(),
		"htm": method,
		"htu": url,
		"iat": now.Unix(),
		"exp": now.Add(30 * time.Second).Unix(),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("序列化 claims 失败: %w", err)
	}

	signature, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	token, err := signature.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("序列化 token 失败: %w", err)
	}

	return token, nil
}

// pdsDpopJWT 创建 PDS DPoP JWT
func (c *OAuthClient) pdsDpopJWT(method string, url string, iss string, accessToken string, nonce string, dpopPrivateJWK jose.JSONWebKey) (string, error) {
	now := time.Now()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.ES256,
		Key:       dpopPrivateJWK.Key,
	}, (&jose.SignerOptions{}).
		WithType("dpop+jwt").
		WithHeader("jwk", dpopPrivateJWK.Public()))
	if err != nil {
		return "", fmt.Errorf("创建签名者失败: %w", err)
	}

	claims := map[string]interface{}{
		"typ": "dpop+jwt",
		"iss": iss,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Second).Unix(),
		"jti": c.generateToken(),
		"htm": method,
		"htu": url,
		"ath": c.createS256CodeChallenge(accessToken),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("序列化 claims 失败: %w", err)
	}

	signature, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	token, err := signature.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("序列化 token 失败: %w", err)
	}

	return token, nil
}

// 工具方法
func (c *OAuthClient) getStringSlice(meta map[string]interface{}, key string) []string {
	interfaceSlice, ok := meta[key].([]interface{})
	if !ok {
		return []string{}
	}

	result := make([]string, 0, len(interfaceSlice))
	for _, v := range interfaceSlice {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func (c *OAuthClient) containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (c *OAuthClient) buildRedirectURL(platform string) string {
	return c.appURL + "api/oauth/callback"
}

func (c *OAuthClient) buildClientID(platform string) string {
	if strings.Contains(c.appURL, "localhost") || strings.Contains(c.appURL, "127.0.0.1") {
		return fmt.Sprintf("http://localhost/?scope=atproto transition:generic&redirect_uri=%sapi/oauth/callback", c.appURL)
	}
	return c.appURL + "api/oauth/" + platform + "/client-metadata.json"
}

// sendPARAuthRequest 发送推送授权请求
func (c *OAuthClient) sendPARAuthRequest(
	authserverURL string,
	authserverMeta map[string]interface{},
	loginHint string,
	clientID string,
	redirectURI string,
	scope string,
	dpopPrivateJWK jose.JSONWebKey,
) (string, string, string, *http.Response, error) {
	parURL, ok := authserverMeta["pushed_authorization_request_endpoint"].(string)
	if !ok {
		return "", "", "", nil, fmt.Errorf("无效的 PAR 端点")
	}

	state := c.generateToken()
	pkceVerifier := c.generateToken(48)
	codeChallenge := c.createS256CodeChallenge(pkceVerifier)
	codeChallengeMethod := "S256"

	clientAssertion, err := c.clientAssertionJWT(clientID, authserverURL)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("创建客户端断言失败: %w", err)
	}

	dpopAuthserverNonce := ""
	dpopProof, err := c.authserverDpopJWT("POST", parURL, dpopAuthserverNonce, dpopPrivateJWK)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("创建 DPoP JWT 失败: %w", err)
	}

	parBody := url.Values{
		"response_type":         {"code"},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {codeChallengeMethod},
		"client_id":             {clientID},
		"state":                 {state},
		"redirect_uri":          {redirectURI},
		"scope":                 {scope},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {clientAssertion},
	}

	if loginHint != "" {
		parBody.Set("login_hint", loginHint)
	}

	headers := map[string]string{
		"DPoP": dpopProof,
	}

	resp, err := c.makeHTTPRequest("POST", parURL, strings.NewReader(parBody.Encode()), headers)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("请求失败: %w", err)
	}

	// 处理 DPoP 缺少/无效 nonce 错误
	if resp.StatusCode == 400 {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", "", "", nil, fmt.Errorf("读取响应失败: %w", err)
		}

		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error == "use_dpop_nonce" {
			dpopAuthserverNonce = resp.Header.Get("DPoP-Nonce")

			dpopProof, err = c.authserverDpopJWT("POST", parURL, dpopAuthserverNonce, dpopPrivateJWK)
			if err != nil {
				return "", "", "", nil, fmt.Errorf("创建 DPoP JWT 失败: %w", err)
			}

			headers["DPoP"] = dpopProof
			resp, err = c.makeHTTPRequest("POST", parURL, strings.NewReader(parBody.Encode()), headers)
			if err != nil {
				return "", "", "", nil, fmt.Errorf("请求失败: %w", err)
			}
		} else {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	return pkceVerifier, state, dpopAuthserverNonce, resp, nil
}

// 通用令牌请求方法
func (c *OAuthClient) makeTokenRequest(
	authserverURL string,
	params url.Values,
	dpopPrivateJWK jose.JSONWebKey,
	dpopAuthserverNonce string,
) (map[string]interface{}, string, error) {
	authserverMeta, err := c.fetchAuthserverMeta(authserverURL)
	if err != nil {
		return nil, "", fmt.Errorf("获取授权服务器元数据失败: %w", err)
	}

	tokenURL, ok := authserverMeta["token_endpoint"].(string)
	if !ok {
		return nil, "", fmt.Errorf("无效的令牌端点")
	}

	dpopProof, err := c.authserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
	if err != nil {
		return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
	}

	headers := map[string]string{
		"DPoP": dpopProof,
	}

	resp, err := c.makeHTTPRequest("POST", tokenURL, strings.NewReader(params.Encode()), headers)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理 DPoP nonce 重试
	if resp.StatusCode == 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("读取响应失败: %w", err)
		}

		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error == "use_dpop_nonce" {
			dpopAuthserverNonce = resp.Header.Get("DPoP-Nonce")

			dpopProof, err = c.authserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
			if err != nil {
				return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
			}

			headers["DPoP"] = dpopProof
			resp, err = c.makeHTTPRequest("POST", tokenURL, strings.NewReader(params.Encode()), headers)
			if err != nil {
				return nil, "", fmt.Errorf("请求失败: %w", err)
			}
			defer resp.Body.Close()
		} else {
			return nil, "", fmt.Errorf("令牌请求失败: %s", string(body))
		}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("令牌请求错误: %s\n", string(body))
		return nil, "", fmt.Errorf("令牌请求返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应失败: %w", err)
	}

	var tokenBody map[string]interface{}
	if err := json.Unmarshal(body, &tokenBody); err != nil {
		return nil, "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return tokenBody, dpopAuthserverNonce, nil
}

// initialTokenRequest 初始令牌请求
func (c *OAuthClient) initialTokenRequest(
	authRequest *repositories.OAuthAuthRequest,
	code string,
	clientID string,
	redirectURI string,
) (map[string]interface{}, string, error) {
	authserverURL := authRequest.AuthserverIss

	clientAssertion, err := c.clientAssertionJWT(clientID, authserverURL)
	if err != nil {
		return nil, "", fmt.Errorf("创建客户端断言失败: %w", err)
	}

	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"grant_type":            {"authorization_code"},
		"code":                  {code},
		"code_verifier":         {authRequest.PkceVerifier},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {clientAssertion},
	}

	var dpopPrivateJWK jose.JSONWebKey
	err = dpopPrivateJWK.UnmarshalJSON([]byte(authRequest.DpopPrivateJwk))
	if err != nil {
		return nil, "", fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}

	return c.makeTokenRequest(authserverURL, params, dpopPrivateJWK, authRequest.DpopAuthserverNonce)
}

func (c *OAuthClient) refreshTokenRequest(
	session *repositories.OAuthSession,
	clientID string,
	redirectURI string,
) (map[string]interface{}, string, error) {
	authserverURL := session.AuthserverIss

	clientAssertion, err := c.clientAssertionJWT(clientID, authserverURL)
	if err != nil {
		return nil, "", fmt.Errorf("创建客户端断言失败: %w", err)
	}

	params := url.Values{
		"client_id":             {clientID},
		"grant_type":            {"refresh_token"},
		"refresh_token":         {session.RefreshToken},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {clientAssertion},
	}

	var dpopPrivateJWK jose.JSONWebKey
	err = dpopPrivateJWK.UnmarshalJSON([]byte(session.DpopPrivateJwk))
	if err != nil {
		return nil, "", fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}

	return c.makeTokenRequest(authserverURL, params, dpopPrivateJWK, session.DpopAuthserverNonce)
}

// 公共函数，供其他包使用

// ResolvePDSAuthserver 解析 PDS 授权服务器（公共函数）
func ResolvePDSAuthserver(pdsURL string) (string, error) {
	client := &OAuthClient{}
	resp, err := client.makeHTTPRequest("GET", fmt.Sprintf("%s/.well-known/oauth-protected-resource", pdsURL), nil, nil)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务器返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var respData struct {
		AuthorizationServers []string `json:"authorization_servers"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if len(respData.AuthorizationServers) == 0 {
		return "", fmt.Errorf("未找到授权服务器")
	}

	return respData.AuthorizationServers[0], nil
}

// FetchAuthserverMeta 获取授权服务器元数据（公共函数）
func FetchAuthserverMeta(authserverURL string) (map[string]interface{}, error) {
	client := &OAuthClient{}
	resp, err := client.makeHTTPRequest("GET", fmt.Sprintf("%s/.well-known/oauth-authorization-server", authserverURL), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回非成功状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var authserverMeta map[string]interface{}
	if err := json.Unmarshal(body, &authserverMeta); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return authserverMeta, nil
}

// GeneratePKCEVerifier 生成 PKCE 验证器（公共函数）
func GeneratePKCEVerifier() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 48) // 48 字节，符合 PKCE 规范
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// GeneratePKCEChallenge 生成 PKCE 挑战（公共函数）
func GeneratePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "=")
}

// SendPARAuthRequest 发送推送授权请求（公共函数）
func SendPARAuthRequest(
	authserverURL string,
	authserverMeta map[string]interface{},
	loginHint string,
	clientID string,
	redirectURI string,
	scope string,
	clientSecretJWK jose.JSONWebKey,
	dpopPrivateJWK jose.JSONWebKey,
) (string, string, string, *http.Response, error) {
	client := NewOAuthClient(authserverURL, clientSecretJWK)
	return client.sendPARAuthRequest(
		authserverURL,
		authserverMeta,
		loginHint,
		clientID,
		redirectURI,
		scope,
		dpopPrivateJWK,
	)
}

// InitialTokenRequest 初始令牌请求（公共函数）
func InitialTokenRequest(
	authRequest *repositories.OAuthAuthRequest,
	code string,
	clientID string,
	redirectURI string,
	clientSecretJWK jose.JSONWebKey,
) (map[string]interface{}, string, error) {
	client := NewOAuthClient("", clientSecretJWK)
	return client.initialTokenRequest(authRequest, code, clientID, redirectURI)
}

// RefreshTokenRequest 刷新令牌请求（公共函数）
func RefreshTokenRequest(
	session *repositories.OAuthSession,
	clientID string,
	redirectURI string,
	clientSecretJWK jose.JSONWebKey,
) (map[string]interface{}, string, error) {
	client := NewOAuthClient("", clientSecretJWK)
	return client.refreshTokenRequest(session, clientID, redirectURI)
}
