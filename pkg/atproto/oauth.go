package atproto

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
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
	"github.com/labstack/gommon/log"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

func FetchAuthserverMeta(url string) (map[string]interface{}, error) {
	// 重要：授权服务器 URL 是不受信任的输入，需要 SSRF 缓解措施
	if !utils.IsSafeURL(url) {
		return nil, fmt.Errorf("不安全的 URL: %s", url)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Get(fmt.Sprintf("%s/.well-known/oauth-authorization-server", url))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
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

	fmt.Printf("Auth Server Metadata: %s\n", PrettyJSON(authserverMeta))

	if !IsValidAuthserverMeta(authserverMeta, url) {
		return nil, fmt.Errorf("无效的授权服务器元数据")
	}

	return authserverMeta, nil
}

func IsValidAuthserverMeta(meta map[string]interface{}, url string) bool {
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

	// 检查响应类型支持
	if !containsString(getStringSlice(meta, "response_types_supported"), "code") {
		return false
	}

	// 检查授权类型支持
	grantTypes := getStringSlice(meta, "grant_types_supported")
	if !containsString(grantTypes, "authorization_code") || !containsString(grantTypes, "refresh_token") {
		return false
	}

	// 检查代码挑战方法支持
	if !containsString(getStringSlice(meta, "code_challenge_methods_supported"), "S256") {
		return false
	}

	// 检查令牌端点认证方法支持
	authMethods := getStringSlice(meta, "token_endpoint_auth_methods_supported")
	if !containsString(authMethods, "none") || !containsString(authMethods, "private_key_jwt") {
		return false
	}

	// 检查令牌端点认证签名算法支持
	if !containsString(getStringSlice(meta, "token_endpoint_auth_signing_alg_values_supported"), "ES256") {
		return false
	}

	// 检查作用域支持
	if !containsString(getStringSlice(meta, "scopes_supported"), "atproto") {
		return false
	}

	// 检查其他必需字段
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

	if !containsString(getStringSlice(meta, "dpop_signing_alg_values_supported"), "ES256") {
		return false
	}

	// 检查可选字段（如果存在）
	if requireRequestURIReg, ok := meta["require_request_uri_registration"].(bool); ok && !requireRequestURIReg {
		return false
	}

	clientIDMetadata, ok := meta["client_id_metadata_document_supported"].(bool)
	if !ok || !clientIDMetadata {
		return false
	}

	return true
}

func getStringSlice(meta map[string]interface{}, key string) []string {
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

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func PrettyJSON(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("错误: %v", err)
	}
	return string(bytes)
}

func ResolvePDSAuthserver(url string) (string, error) {
	// 重要：PDS 端点 URL 是不受信任的输入，需要 SSRF 缓解措施
	if !utils.IsSafeURL(url) {
		return "", fmt.Errorf("不安全的 URL: %s", url)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Get(fmt.Sprintf("%s/.well-known/oauth-protected-resource", url))
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

func GenerateToken(length ...int) string {
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

func CreateS256CodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "=") // 移除尾部的等号
}

func ClientAssertionJWT(clientID string, authserverURL string, clientSecretJWK jose.JSONWebKey) (string, error) {
	now := time.Now()

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.ES256,
		Key:       clientSecretJWK.Key,
	}, (&jose.SignerOptions{}).
		WithType("jwt").
		WithHeader("kid", clientSecretJWK.KeyID))
	if err != nil {
		return "", fmt.Errorf("创建签名者失败: %w", err)
	}

	claims := map[string]interface{}{
		"iss": clientID,
		"sub": clientID,
		"aud": []string{authserverURL},
		"jti": GenerateToken(),
		"iat": now.Unix(),
	}

	// 序列化并签名
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

func AuthserverDpopJWT(method string, url string, nonce string, dpopPrivateJWK jose.JSONWebKey) (string, error) {
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
		"jti": GenerateToken(),
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

func PDSDpopJWT(method string, url string, iss string, accessToken string, nonce string, dpopPrivateJWK jose.JSONWebKey) (string, error) {
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
		"jti": GenerateToken(),
		"htm": method,
		"htu": url,
		"ath": CreateS256CodeChallenge(accessToken),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	// 序列化并签名
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
	parURL, ok := authserverMeta["pushed_authorization_request_endpoint"].(string)
	if !ok {
		return "", "", "", nil, fmt.Errorf("无效的 PAR 端点")
	}

	state := GenerateToken()
	pkceVerifier := GenerateToken(48)

	codeChallenge := CreateS256CodeChallenge(pkceVerifier)
	codeChallengeMethod := "S256"

	clientAssertion, err := ClientAssertionJWT(clientID, authserverURL, clientSecretJWK)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("创建客户端断言失败: %w", err)
	}

	dpopAuthserverNonce := ""
	dpopProof, err := AuthserverDpopJWT("POST", parURL, dpopAuthserverNonce, dpopPrivateJWK)
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

	// 重要：推送授权请求 URL 是不受信任的输入，需要 SSRF 缓解措施
	if !utils.IsSafeURL(parURL) {
		return "", "", "", nil, fmt.Errorf("不安全的 URL: %s", parURL)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("POST", parURL, strings.NewReader(parBody.Encode()))
	if err != nil {
		return "", "", "", nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("DPoP", dpopProof)

	log.Infof("dpopProof: %s", dpopProof)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("请求失败: %w", err)
	}

	// 处理 DPoP 缺少/无效 nonce 错误，使用服务器提供的 nonce 重试
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
			fmt.Printf("使用新的授权服务器 DPoP nonce 重试: %s\n", dpopAuthserverNonce)

			dpopProof, err = AuthserverDpopJWT("POST", parURL, dpopAuthserverNonce, dpopPrivateJWK)
			if err != nil {
				return "", "", "", nil, fmt.Errorf("创建 DPoP JWT 失败: %w", err)
			}

			req, err = http.NewRequest("POST", parURL, strings.NewReader(parBody.Encode()))
			if err != nil {
				return "", "", "", nil, fmt.Errorf("创建请求失败: %w", err)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("DPoP", dpopProof)
			resp, err = client.Do(req)
			if err != nil {
				return "", "", "", nil, fmt.Errorf("请求失败: %w", err)
			}
		} else {
			// 重新创建响应体以便后续处理
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}

	return pkceVerifier, state, dpopAuthserverNonce, resp, nil
}

func InitialTokenRequest(
	authRequest *database.OAuthAuthRequest,
	code string,
	clientID string,
	redirectURI string,
	clientSecretJWK jose.JSONWebKey,
) (map[string]interface{}, string, error) {
	authserverURL := authRequest.AuthserverIss
	authserverMeta, err := FetchAuthserverMeta(authserverURL)
	if err != nil {
		return nil, "", fmt.Errorf("获取授权服务器元数据失败: %w", err)
	}

	clientAssertion, err := ClientAssertionJWT(clientID, authserverURL, clientSecretJWK)
	if err != nil {
		return nil, "", fmt.Errorf("创建客户端断言失败: %w", err)
	}

	pkceVerifier := authRequest.PkceVerifier

	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"grant_type":            {"authorization_code"},
		"code":                  {code},
		"code_verifier":         {pkceVerifier},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {clientAssertion},
	}

	tokenURL, ok := authserverMeta["token_endpoint"].(string)
	if !ok {
		return nil, "", fmt.Errorf("无效的令牌端点")
	}

	dpopPrivateJWKStr := authRequest.DpopPrivateJwk

	var dpopPrivateJWK jose.JSONWebKey
	err = dpopPrivateJWK.UnmarshalJSON([]byte(dpopPrivateJWKStr))
	if err != nil {
		return nil, "", fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}
	dpopAuthserverNonce := authRequest.DpopAuthserverNonce
	dpopProof, err := AuthserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
	if err != nil {
		return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
	}

	// 重要：令牌 URL 是不受信任的输入，需要 SSRF 缓解措施
	if !utils.IsSafeURL(tokenURL) {
		return nil, "", fmt.Errorf("不安全的 URL: %s", tokenURL)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("DPoP", dpopProof)
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理 DPoP 缺少/无效 nonce 错误，使用服务器提供的 nonce 重试
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
			fmt.Printf("使用新的授权服务器 DPoP nonce 重试: %s\n", dpopAuthserverNonce)

			dpopProof, err = AuthserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
			if err != nil {
				return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
			}

			req, err = http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
			if err != nil {
				return nil, "", fmt.Errorf("创建请求失败: %w", err)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("DPoP", dpopProof)
			resp, err = client.Do(req)
			if err != nil {
				return nil, "", fmt.Errorf("请求失败: %w", err)
			}
			defer resp.Body.Close()
		} else {
			return nil, "", fmt.Errorf("令牌请求失败: %s", string(body))
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("令牌请求返回非成功状态码: %d, %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("读取响应失败: %w", err)
	}

	var tokenBody map[string]interface{}
	if err := json.Unmarshal(body, &tokenBody); err != nil {
		return nil, "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	// 重要：调用此函数的代码必须验证原始请求中的 'sub' 字段。

	return tokenBody, dpopAuthserverNonce, nil
}

func RefreshTokenRequest(
	session *database.OAuthSession,
	clientID string,
	redirectURI string,
	clientSecretJWK jose.JSONWebKey,
) (map[string]interface{}, string, error) {
	authserverURL := session.AuthserverIss
	authserverMeta, err := FetchAuthserverMeta(authserverURL)
	if err != nil {
		return nil, "", fmt.Errorf("获取授权服务器元数据失败: %w", err)
	}

	clientAssertion, err := ClientAssertionJWT(clientID, authserverURL, clientSecretJWK)
	if err != nil {
		return nil, "", fmt.Errorf("创建客户端断言失败: %w", err)
	}

	refreshToken := session.RefreshToken

	params := url.Values{
		"client_id":             {clientID},
		"grant_type":            {"refresh_token"},
		"refresh_token":         {refreshToken},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {clientAssertion},
	}

	tokenURL, ok := authserverMeta["token_endpoint"].(string)
	if !ok {
		return nil, "", fmt.Errorf("无效的令牌端点")
	}

	dpopPrivateJWKStr := session.DpopPrivateJwk

	var dpopPrivateJWK jose.JSONWebKey
	err = dpopPrivateJWK.UnmarshalJSON([]byte(dpopPrivateJWKStr))
	if err != nil {
		return nil, "", fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}

	dpopAuthserverNonce := session.DpopAuthserverNonce

	dpopProof, err := AuthserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
	if err != nil {
		return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
	}

	// 重要：令牌 URL 是不受信任的输入，需要 SSRF 缓解措施
	if !utils.IsSafeURL(tokenURL) {
		return nil, "", fmt.Errorf("不安全的 URL: %s", tokenURL)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("DPoP", dpopProof)
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 处理 DPoP 缺少/无效 nonce 错误，使用服务器提供的 nonce 重试
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
			fmt.Printf("使用新的授权服务器 DPoP nonce 重试: %s\n", dpopAuthserverNonce)

			dpopProof, err = AuthserverDpopJWT("POST", tokenURL, dpopAuthserverNonce, dpopPrivateJWK)
			if err != nil {
				return nil, "", fmt.Errorf("创建 DPoP JWT 失败: %w", err)
			}

			req, err = http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
			if err != nil {
				return nil, "", fmt.Errorf("创建请求失败: %w", err)
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("DPoP", dpopProof)
			resp, err = client.Do(req)
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
		fmt.Printf("令牌刷新错误: %s\n", string(body))
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

func PDSAuthedReq(method string, url string, session *database.OAuthSession, db interface{}, body interface{}) (*http.Response, error) {
	dpopPrivateJWKStr := session.DpopPrivateJwk
	var dpopPrivateJWK jose.JSONWebKey
	err := dpopPrivateJWK.UnmarshalJSON([]byte(dpopPrivateJWKStr))
	if err != nil {
		return nil, fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}
	dpopPDSNonce := session.DpopPdsNonce
	if dpopPDSNonce == "" {
		dpopPDSNonce = ""
	}

	accessToken := session.AccessToken

	authserverISS := session.AuthserverIss

	did := session.Did

	// 可能需要使用新的 nonce 重试请求
	for i := 0; i < 2; i++ {
		dpopJWT, e := PDSDpopJWT(
			method,
			url,
			authserverISS,
			accessToken,
			dpopPDSNonce,
			dpopPrivateJWK,
		)
		if e != nil {
			return nil, fmt.Errorf("创建 PDS DPoP JWT 失败: %w", e)
		}

		// 重要：PDS URL 是不受信任的输入，需要 SSRF 缓解措施
		if !utils.IsSafeURL(url) {
			return nil, fmt.Errorf("不安全的 URL: %s", url)
		}

		var req *http.Request
		var err error

		if method == "POST" {
			var reqBody []byte
			if body != nil {
				reqBody, err = json.Marshal(body)
				if err != nil {
					return nil, fmt.Errorf("序列化请求体失败: %w", err)
				}
			}
			req, err = http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
			if err != nil {
				return nil, fmt.Errorf("创建请求失败: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest(method, url, nil)
			if err != nil {
				return nil, fmt.Errorf("创建请求失败: %w", err)
			}
		}

		req.Header.Set("Authorization", fmt.Sprintf("DPoP %s", accessToken))
		req.Header.Set("DPoP", dpopJWT)

		client := &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		// 如果收到新的服务器提供的 DPoP nonce，将其存储在数据库中并重试
		// 注意：错误类型也可能在 `WWW-Authenticate` HTTP 响应标头中传达
		if resp.StatusCode == 400 || resp.StatusCode == 401 {
			var respBody []byte
			respBody, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("读取响应失败: %w", err)
			}

			var errorResp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(respBody, &errorResp); err == nil && errorResp.Error == "use_dpop_nonce" {
				dpopPDSNonce = resp.Header.Get("DPoP-Nonce")
				fmt.Printf("使用新的 PDS DPoP nonce 重试: %s\n", dpopPDSNonce)

				// 更新会话数据库中的新 nonce
				dbConn, ok := db.(interface {
					Exec(query string, args ...interface{}) (sql.Result, error)
				})
				if !ok {
					return nil, fmt.Errorf("无效的数据库连接")
				}

				_, err = dbConn.Exec(
					"UPDATE oauth_session SET dpop_pds_nonce = ? WHERE did = ?;",
					dpopPDSNonce,
					did,
				)
				if err != nil {
					return nil, fmt.Errorf("更新数据库失败: %w", err)
				}

				continue
			}

			// 重新创建响应体以便后续处理
			resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
		}

		return resp, nil
	}

	return nil, fmt.Errorf("超过最大重试次数")
}

func BuildRedirectURL(appURL, platform string) string {
	return appURL + "api/oauth/callback"
}

func BuildClientID(appURL, platform string) string {
	if strings.Contains(appURL, "localhost") || strings.Contains(appURL, "127.0.0.1") {
		clientID := fmt.Sprintf("http://localhost/?scope=atproto transition:generic&redirect_uri=%sapi/oauth/callback", appURL)
		return clientID
	}

	return appURL + "api/oauth/" + platform + "/client-metadata.json"
}

func BuildCallbackRedirectURI(appURL, platform string) string {
	if platform == "web" {
		return appURL + "api/oauth/token"
	}
	return appURL + "api/oauth/app-return/oxchat"
}
