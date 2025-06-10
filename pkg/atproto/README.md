# ATP OAuth 客户端

这是一个干净简洁的 ATP (AT Protocol) OAuth 客户端，将所有 OAuth 逻辑封装在一个统一的客户端中，提供了优雅的接口。

## 特性

- 🔐 完整的 ATP OAuth 2.0 + DPoP 支持
- 🚀 简洁优雅的 API 设计
- 🔄 自动令牌刷新
- 🛡️ 内置安全检查 (SSRF 防护)
- 📦 所有逻辑封装在单一客户端中
- 🎯 零外部依赖，完全自包含

## 快速开始

### 1. 创建客户端

```go
import (
    "github.com/zhongshangwu/avatarai-social/pkg/atproto"
    "github.com/go-jose/go-jose/v4"
    "gorm.io/gorm"
)

// 初始化客户端
client := atproto.NewOAuthClient(
    "https://your-app.com/",  // 应用 URL
    clientSecretJWK,          // 客户端密钥 JWK
    db,                       // GORM 数据库连接
)
```

### 2. 开始授权流程

```go
authResp, err := client.StartAuth(&atproto.AuthRequest{
    LoginHint: "user.bsky.social",           // 用户的 PDS 地址
    Platform:  "web",                       // 平台类型
    Scope:     "atproto transition:generic", // 权限范围
})
if err != nil {
    log.Fatal(err)
}

// 重定向用户到授权 URL
fmt.Printf("请访问: %s\n", authResp.AuthURL)
```

### 3. 交换授权码

```go
// 用户授权后，从回调中获取授权码
tokenResp, err := client.ExchangeToken(&atproto.TokenRequest{
    Code:     authorizationCode,
    State:    authResp.State,
    Platform: "web",
}, authRequest) // 从数据库获取的授权请求
if err != nil {
    log.Fatal(err)
}

fmt.Printf("访问令牌: %s\n", tokenResp.AccessToken)
```

### 4. 发起 PDS 请求

```go
resp, err := client.MakePDSRequest(&atproto.PDSRequest{
    Method: "GET",
    URL:    "https://bsky.social/xrpc/com.atproto.repo.getRecord",
    Body:   nil,
}, session) // OAuth 会话
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

### 5. 刷新令牌

```go
// 检查会话是否过期
if client.IsSessionExpired(session) {
    newTokenResp, err := client.RefreshToken(&atproto.RefreshRequest{
        SessionDID: session.Did,
        Platform:   "web",
    }, session)
    if err != nil {
        log.Fatal(err)
    }

    // 更新会话
    session.AccessToken = newTokenResp.AccessToken
    session.RefreshToken = newTokenResp.RefreshToken
}
```

## API 参考

### 类型定义

#### AuthRequest
```go
type AuthRequest struct {
    LoginHint   string  // PDS 地址或用户标识
    Platform    string  // 平台类型 ("web", "ios", "android")
    Scope       string  // 权限范围 (可选，默认 "atproto transition:generic")
    RedirectURI string  // 重定向 URI (可选，使用默认值)
}
```

#### AuthResponse
```go
type AuthResponse struct {
    AuthURL        string                 // 授权 URL
    State          string                 // 状态参数
    PKCEVerifier   string                 // PKCE 验证码
    DpopNonce      string                 // DPoP nonce
    AuthserverMeta map[string]interface{} // 授权服务器元数据
}
```

#### TokenRequest
```go
type TokenRequest struct {
    Code        string  // 授权码
    State       string  // 状态参数
    Platform    string  // 平台类型
    RedirectURI string  // 重定向 URI (可选)
}
```

#### TokenResponse
```go
type TokenResponse struct {
    AccessToken         string  // 访问令牌
    RefreshToken        string  // 刷新令牌
    DpopAuthserverNonce string  // DPoP nonce
    ExpiresIn           int64   // 过期时间 (秒)
    TokenType           string  // 令牌类型
    Scope               string  // 权限范围
}
```

### 主要方法

#### StartAuth
开始 OAuth 授权流程，返回授权 URL 和相关参数。

#### ExchangeToken
使用授权码交换访问令牌。

#### RefreshToken
刷新过期的访问令牌。

#### MakePDSRequest
使用访问令牌发起 PDS API 请求。

#### IsSessionExpired
检查会话是否已过期。

#### GenerateClientMetadata
生成 OAuth 客户端元数据 (用于发现端点)。

#### GenerateJWKS
生成 JWKS (JSON Web Key Set)。

## 架构设计

### 统一封装
所有 OAuth 逻辑都封装在 `OAuthClient` 结构体中，包括：

- **授权服务器发现**: 自动解析 PDS 授权服务器
- **元数据验证**: 验证授权服务器元数据的完整性和安全性
- **DPoP 处理**: 自动生成和管理 DPoP 密钥和 JWT
- **PAR 请求**: 推送授权请求处理
- **令牌管理**: 令牌交换、刷新和验证
- **PDS 请求**: 认证的 PDS API 请求
- **错误重试**: 自动处理 DPoP nonce 重试

### 内部方法
客户端包含以下内部方法（不对外暴露）：

- `fetchAuthserverMeta()`: 获取授权服务器元数据
- `isValidAuthserverMeta()`: 验证授权服务器元数据
- `resolvePDSAuthserver()`: 解析 PDS 授权服务器
- `sendPARAuthRequest()`: 发送推送授权请求
- `initialTokenRequest()`: 初始令牌请求
- `refreshTokenRequest()`: 刷新令牌请求
- `pdsAuthedReq()`: PDS 认证请求
- `clientAssertionJWT()`: 创建客户端断言 JWT
- `authserverDpopJWT()`: 创建授权服务器 DPoP JWT
- `pdsDpopJWT()`: 创建 PDS DPoP JWT

## 注意事项

1. **安全性**: 客户端内置了 SSRF 防护，会验证所有外部 URL
2. **数据库**: 需要使用 GORM 数据库连接
3. **DPoP**: 自动处理 DPoP (Demonstration of Proof-of-Possession) 流程
4. **错误处理**: 所有方法都返回详细的错误信息
5. **重试机制**: 内置 DPoP nonce 重试机制
6. **自包含**: 不依赖外部 OAuth 函数，所有逻辑都在客户端内部

## 迁移指南

如果你正在从原有的分散 OAuth 实现迁移：

1. 创建 `OAuthClient` 实例替代分散的函数调用
2. 使用新的结构化请求/响应类型
3. 利用客户端的自动状态管理功能
4. 保持现有的数据库结构不变
5. 所有原有的安全检查和 DPoP 逻辑都已保留

## 依赖

- `github.com/go-jose/go-jose/v4` - JWT/JWK 处理
- `gorm.io/gorm` - 数据库 ORM
- `github.com/zhongshangwu/avatarai-social/pkg/repositories` - 数据模型
- `github.com/zhongshangwu/avatarai-social/pkg/utils` - 工具函数

## 完整示例

```go
package main

import (
    "fmt"
    "log"

    "github.com/zhongshangwu/avatarai-social/pkg/atproto"
    "github.com/go-jose/go-jose/v4"
    "gorm.io/gorm"
)

func main() {
    // 创建客户端
    client := atproto.NewOAuthClient(appURL, clientSecretJWK, db)

    // 开始授权
    authResp, err := client.StartAuth(&atproto.AuthRequest{
        LoginHint: "user.bsky.social",
        Platform:  "web",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("访问授权 URL: %s\n", authResp.AuthURL)

    // 用户完成授权后...
    tokenResp, err := client.ExchangeToken(&atproto.TokenRequest{
        Code:     authorizationCode,
        State:    authResp.State,
        Platform: "web",
    }, authRequest)
    if err != nil {
        log.Fatal(err)
    }

    // 使用令牌发起 API 请求
    resp, err := client.MakePDSRequest(&atproto.PDSRequest{
        Method: "GET",
        URL:    "https://bsky.social/xrpc/com.atproto.repo.listRecords",
    }, session)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    fmt.Printf("API 请求成功: %d\n", resp.StatusCode)
}
```