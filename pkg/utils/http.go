package utils

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type AvatarAIContext struct {
	echo.Context
	Avatar       *repositories.Avatar
	Session      *repositories.Session
	OauthSession *repositories.OAuthSession
}

// IsSafeURL 检查URL是否安全，防止SSRF攻击
// 这只是一个部分缓解措施，实际的HTTP客户端还需要防止其他攻击和行为
func IsSafeURL(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// 检查URL是否使用HTTPS协议，并且有效的主机名
	if parsedURL.Scheme != "https" ||
		parsedURL.Hostname() == "" ||
		parsedURL.Hostname() != parsedURL.Host ||
		parsedURL.User != nil {
		return false
	}

	// 检查主机名格式
	segments := strings.Split(parsedURL.Hostname(), ".")
	if len(segments) < 2 {
		return false
	}

	// 检查顶级域名是否在禁用列表中
	lastSegment := segments[len(segments)-1]
	if lastSegment == "local" ||
		lastSegment == "arpa" ||
		lastSegment == "internal" ||
		lastSegment == "localhost" {
		return false
	}

	// 检查顶级域名是否为纯数字
	if isDigit(lastSegment) {
		return false
	}

	return true
}

// isDigit 检查字符串是否只包含数字
func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// 创建一个加固的HTTP客户端配置
type HardenedHTTPClientConfig struct {
	Timeout           time.Duration
	DisableRedirects  bool
	BlockLoopbackIPs  bool
	UserAgentOverride string
}

// NewDefaultHardenedConfig 返回默认的加固HTTP客户端配置
func NewDefaultHardenedConfig() *HardenedHTTPClientConfig {
	return &HardenedHTTPClientConfig{
		Timeout:           10 * time.Second,
		DisableRedirects:  true,
		BlockLoopbackIPs:  true,
		UserAgentOverride: "AtprotoGoClient",
	}
}

// NewHardenedHTTPClient 创建一个加固的HTTP客户端
func NewHardenedHTTPClient(config *HardenedHTTPClientConfig) *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// 实现IP过滤
		// DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 	// 如果需要阻止本地回环IP请求
		// 	if config.BlockLoopbackIPs {
		// 		host, _, err := net.SplitHostPort(addr)
		// 		if err != nil {
		// 			return nil, err
		// 		}

		// 		ip := net.ParseIP(host)
		// 		if ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		// 			return nil, fmt.Errorf("连接到内部IP地址被阻止: %s", host)
		// 		}
		// 	}

		// 	dialer := &net.Dialer{
		// 		Timeout:   30 * time.Second,
		// 		KeepAlive: 30 * time.Second,
		// 	}
		// 	return dialer.DialContext(ctx, network, addr)
		// },
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// 禁用重定向
	if config.DisableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

// Do 执行HTTP请求，添加自定义User-Agent并验证URL安全性
func Do(client *http.Client, req *http.Request, config *HardenedHTTPClientConfig) (*http.Response, error) {
	// 验证URL安全性
	if !IsSafeURL(req.URL.String()) {
		return nil, fmt.Errorf("不安全的URL: %s", req.URL.String())
	}

	// 设置User-Agent
	if config.UserAgentOverride != "" {
		req.Header.Set("User-Agent", config.UserAgentOverride)
	}

	return client.Do(req)
}

func GetAPPURL(c echo.Context) string {
	scheme := "https"
	if c.Request().TLS == nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request().Host + "/"
}

type Claims struct {
	DID                  string    `json:"did"`
	Handle               string    `json:"handle"`
	ExpiredAt            time.Time `json:"expired_at"`
	jwt.RegisteredClaims           // 嵌入标准 JWT 字段
}

func GenerateAvataraiToken(config *config.SocialConfig, sessionID string, avatar *repositories.Avatar) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Hour * 24 * 30) // 30天过期

	avataraiClaims := Claims{
		DID:    avatar.Did,
		Handle: avatar.Handle,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "avatarai-social",             // 令牌签发者
			Subject:   avatar.Did,                    // 令牌主题（通常是用户标识）
			Audience:  []string{"avatarai-app"},      // 令牌接收者
			ExpiresAt: jwt.NewNumericDate(expiresAt), // 过期时间
			NotBefore: jwt.NewNumericDate(now),       // 生效时间
			IssuedAt:  jwt.NewNumericDate(now),       // 签发时间
			ID:        sessionID,                     // JWT ID
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, avataraiClaims)
	aisAccessToken, err := jwtToken.SignedString(config.Security.GetRSAPrivateKey())
	if err != nil {
		return "", err
	}

	return aisAccessToken, nil
}

func ValidateAccessToken(config *config.SocialConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保使用了正确的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return config.Security.GetRSAPublicKey(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析token失败: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 验证是否过期
		if time.Now().After(claims.ExpiresAt.Time) {
			return nil, fmt.Errorf("token已过期")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("无效的token")
}

// GenerateAccessToken 生成短期访问令牌
func GenerateAccessToken(config *config.SocialConfig, sessionID string, avatar *repositories.Avatar) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Hour * 24) // 访问令牌有效期24小时

	accessClaims := Claims{
		DID:    avatar.Did,
		Handle: avatar.Handle,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "avatarai-social",             // 令牌签发者
			Subject:   avatar.Did,                    // 令牌主题（通常是用户标识）
			Audience:  []string{"avatarai-app"},      // 令牌接收者
			ExpiresAt: jwt.NewNumericDate(expiresAt), // 过期时间
			NotBefore: jwt.NewNumericDate(now),       // 生效时间
			IssuedAt:  jwt.NewNumericDate(now),       // 签发时间
			ID:        sessionID,                     // JWT ID
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken, err := jwtToken.SignedString(config.Security.GetRSAPrivateKey())
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GenerateRefreshToken 生成长期刷新令牌
func GenerateRefreshToken(config *config.SocialConfig, sessionID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Hour * 24 * 30) // 刷新令牌有效期30天

	refreshClaims := jwt.RegisteredClaims{
		Issuer:    "avatarai-social",                    // 令牌签发者
		Subject:   sessionID,                            // 以会话ID作为主题
		Audience:  []string{"avatarai-app"},             // 令牌接收者
		ExpiresAt: jwt.NewNumericDate(expiresAt),        // 过期时间
		NotBefore: jwt.NewNumericDate(now),              // 生效时间
		IssuedAt:  jwt.NewNumericDate(now),              // 签发时间
		ID:        fmt.Sprintf("refresh-%s", sessionID), // 刷新令牌ID
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshToken, err := jwtToken.SignedString(config.Security.GetRSAPrivateKey())
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(config *config.SocialConfig, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 确保使用了正确的签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return config.Security.GetRSAPublicKey(), nil
	})

	if err != nil {
		return "", fmt.Errorf("解析刷新令牌失败: %v", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// 验证是否过期
		if time.Now().After(claims.ExpiresAt.Time) {
			return "", fmt.Errorf("刷新令牌已过期")
		}

		// 返回会话ID (Subject字段)
		return claims.Subject, nil
	}

	return "", fmt.Errorf("无效的刷新令牌")
}
