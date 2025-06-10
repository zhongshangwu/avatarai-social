package atproto

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/carlmjohnson/versioninfo"
	"github.com/go-jose/go-jose/v4"
	"github.com/zhongshangwu/avatarai-social/types"
)

type NonceUpdateCallback func(did, newNonce string) error

type XrpcClient struct {
	httpClient *http.Client
	userAgent  string
	headers    map[string]string

	session        *types.OAuthSession
	dpopPrivateJwk jose.JSONWebKey
	onNonceUpdate  NonceUpdateCallback
}

type ClientOption func(*XrpcClient)

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *XrpcClient) {
		c.httpClient = client
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *XrpcClient) {
		c.userAgent = userAgent
	}
}

func WithHeaders(headers map[string]string) ClientOption {
	return func(c *XrpcClient) {
		c.headers = headers
	}
}

func WithNonceUpdateCallback(callback NonceUpdateCallback) ClientOption {
	return func(c *XrpcClient) {
		c.onNonceUpdate = callback
	}
}

func NewXrpcClient(session *types.OAuthSession, options ...ClientOption) (*XrpcClient, error) {
	var dpopPrivateJWK jose.JSONWebKey
	if err := dpopPrivateJWK.UnmarshalJSON([]byte(session.DpopPrivateJwk)); err != nil {
		return nil, fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}

	client := &XrpcClient{
		session:        session,
		dpopPrivateJwk: dpopPrivateJWK,
		userAgent:      "atproto-oauth/" + versioninfo.Short(),
	}

	for _, option := range options {
		option(client)
	}

	if client.httpClient == nil {
		client.httpClient = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}

	return client, nil
}

func (c *XrpcClient) Query(ctx context.Context, method string, params map[string]any, out any) error {
	return c.do(ctx, xrpc.Query, "", method, params, nil, out)
}

func (c *XrpcClient) Procedure(ctx context.Context, method string, params map[string]any, body any, out any) error {
	return c.do(ctx, xrpc.Procedure, "application/json", method, params, body, out)
}

func (c *XrpcClient) ProcedureWithEncoding(ctx context.Context, method string, encoding string, params map[string]any, body any, out any) error {
	return c.do(ctx, xrpc.Procedure, encoding, method, params, body, out)
}

func (c *XrpcClient) UpdateSession(session *types.OAuthSession) error {
	var dpopPrivateJWK jose.JSONWebKey
	if err := dpopPrivateJWK.UnmarshalJSON([]byte(session.DpopPrivateJwk)); err != nil {
		return fmt.Errorf("解析 DPoP 私钥失败: %w", err)
	}

	c.session = session
	c.dpopPrivateJwk = dpopPrivateJWK
	return nil
}

func (c *XrpcClient) GetSession() *types.OAuthSession {
	return c.session
}

func (c *XrpcClient) do(ctx context.Context, kind xrpc.XRPCRequestType, encoding, method string, params map[string]any, bodyobj any, out any) error {
	// 最多重试 2 次（处理 nonce 更新）
	for attempt := 0; attempt < 2; attempt++ {
		if err := c.makeRequest(ctx, kind, encoding, method, params, bodyobj, out); err != nil {
			// 如果是 nonce 相关错误且还有重试机会，继续重试
			if xrpcErr, ok := err.(*xrpc.XRPCError); ok &&
				(xrpcErr.ErrStr == "use_dpop_nonce") &&
				attempt < 1 {
				continue
			}
			return err
		}
		return nil
	}
	return fmt.Errorf("请求失败，已达到最大重试次数")
}

func (c *XrpcClient) makeRequest(ctx context.Context, kind xrpc.XRPCRequestType, encoding, method string, params map[string]any, bodyobj any, out any) error {
	var body io.Reader
	if bodyobj != nil {
		if rr, ok := bodyobj.(io.Reader); ok {
			body = rr
		} else {
			b, err := json.Marshal(bodyobj)
			if err != nil {
				return fmt.Errorf("序列化请求体失败: %w", err)
			}
			body = bytes.NewReader(b)
		}
	}

	var httpMethod string
	switch kind {
	case xrpc.Query:
		httpMethod = "GET"
	case xrpc.Procedure:
		httpMethod = "POST"
	default:
		return fmt.Errorf("不支持的请求类型: %d", kind)
	}

	var paramStr string
	if len(params) > 0 {
		paramStr = "?" + makeParams(params)
	}
	requestURL := c.session.PdsUrl + "/xrpc/" + method + paramStr

	req, err := http.NewRequestWithContext(ctx, httpMethod, requestURL, body)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if bodyobj != nil && encoding != "" {
		req.Header.Set("Content-Type", encoding)
	}
	req.Header.Set("User-Agent", c.userAgent)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	if err := c.setAuthHeaders(req, httpMethod, requestURL); err != nil {
		return fmt.Errorf("设置认证头失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求执行失败: %w", err)
	}
	defer resp.Body.Close()

	return c.handleResponse(resp, out)
}

func (c *XrpcClient) setAuthHeaders(req *http.Request, method, url string) error {
	dpopJwt, err := c.pdsDpopJWT(
		method,
		url,
		c.session.AuthserverIss,
		c.session.AccessToken,
		c.session.DpopPdsNonce,
		c.dpopPrivateJwk,
	)
	if err != nil {
		return fmt.Errorf("生成 DPoP JWT 失败: %w", err)
	}

	req.Header.Set("DPoP", dpopJwt)
	req.Header.Set("Authorization", "DPoP "+c.session.AccessToken)
	return nil
}

func (c *XrpcClient) handleResponse(resp *http.Response, out any) error {
	if resp.StatusCode != 200 {
		var xe xrpc.XRPCError
		if err := json.NewDecoder(resp.Body).Decode(&xe); err != nil {
			return errorFromHTTPResponse(resp, fmt.Errorf("解码错误响应失败: %w", err))
		}

		// 处理 nonce 更新
		if (resp.StatusCode == 400 || resp.StatusCode == 401) && xe.ErrStr == "use_dpop_nonce" {
			newNonce := resp.Header.Get("DPoP-Nonce")
			if newNonce != "" {
				c.session.DpopPdsNonce = newNonce
				if c.onNonceUpdate != nil {
					if err := c.onNonceUpdate(c.session.Did, newNonce); err != nil {
						log.Printf("nonce 更新回调失败: %v", err)
					}
				}
			}
		}

		return errorFromHTTPResponse(resp, &xe)
	}

	// 解析响应体
	if out != nil {
		if buf, ok := out.(*bytes.Buffer); ok {
			if resp.ContentLength < 0 {
				_, err := io.Copy(buf, resp.Body)
				if err != nil {
					return fmt.Errorf("读取响应体失败: %w", err)
				}
			} else {
				n, err := io.CopyN(buf, resp.Body, resp.ContentLength)
				if err != nil {
					return fmt.Errorf("读取限长响应体失败 (%d < %d): %w", n, resp.ContentLength, err)
				}
			}
		} else {
			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				return fmt.Errorf("解码响应失败: %w", err)
			}
		}
	}

	return nil
}

func (c *XrpcClient) pdsDpopJWT(method string, url string, iss string, accessToken string, nonce string, dpopPrivateJWK jose.JSONWebKey) (string, error) {
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

func (c *XrpcClient) generateToken(length ...int) string {
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

func (c *XrpcClient) createS256CodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "=")
}

// GenerateCodeChallenge 生成 PKCE 代码挑战
func GenerateCodeChallenge(pkceVerifier string) string {
	h := sha256.New()
	h.Write([]byte(pkceVerifier))
	hash := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(hash)
}

func errorFromHTTPResponse(resp *http.Response, err error) error {
	r := &xrpc.Error{
		StatusCode: resp.StatusCode,
		Wrapped:    err,
	}
	if resp.Header.Get("ratelimit-limit") != "" {
		r.Ratelimit = &xrpc.RatelimitInfo{
			Policy: resp.Header.Get("ratelimit-policy"),
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-reset"), 10, 64); err == nil {
			r.Ratelimit.Reset = time.Unix(n, 0)
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-limit"), 10, 64); err == nil {
			r.Ratelimit.Limit = int(n)
		}
		if n, err := strconv.ParseInt(resp.Header.Get("ratelimit-remaining"), 10, 64); err == nil {
			r.Ratelimit.Remaining = int(n)
		}
	}
	return r
}

func makeParams(p map[string]any) string {
	params := url.Values{}
	for k, v := range p {
		if s, ok := v.([]string); ok {
			for _, v := range s {
				params.Add(k, v)
			}
		} else {
			params.Add(k, fmt.Sprint(v))
		}
	}
	return params.Encode()
}
