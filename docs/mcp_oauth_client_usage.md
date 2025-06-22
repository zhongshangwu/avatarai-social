# MCP OAuth Client 使用指南

本文档介绍如何使用后端服务作为 MCP Client 进行 OAuth2 认证的接口。

## 接口概览

基于流程图，我们实现了以下 OAuth2 认证流程的接口：

### 1. 资源发现阶段

#### 获取资源元数据
```http
GET /api/mcp/oauth/discover-resource?resource_url=https://mcp-server.example.com
```

响应示例：
```json
{
  "authorization_servers": ["https://auth.example.com"],
  "resource": "https://mcp-server.example.com"
}
```

#### 获取授权服务器元数据
```http
GET /api/mcp/oauth/discover-auth-server?auth_server_url=https://auth.example.com
```

响应示例：
```json
{
  "issuer": "https://auth.example.com",
  "authorization_endpoint": "https://auth.example.com/authorize",
  "token_endpoint": "https://auth.example.com/token",
  "registration_endpoint": "https://auth.example.com/register",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code", "refresh_token"],
  "code_challenge_methods_supported": ["S256"]
}
```

### 2. 客户端注册阶段（可选）

```http
POST /api/mcp/oauth/register-client?registration_url=https://auth.example.com/register
Content-Type: application/json

{
  "client_name": "AvatarAI Social MCP Client",
  "redirect_uris": ["https://avatarai.social/api/mcp/oauth/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "scope": "openid profile mcp:access"
}
```

响应示例：
```json
{
  "client_id": "generated_client_id",
  "client_secret": "generated_client_secret"
}
```

### 3. 授权流程阶段

#### 开始授权流程
```http
GET /api/mcp/oauth/start-authorization?auth_server_url=https://auth.example.com&client_id=your_client_id&resource=https://mcp-server.example.com
```

响应示例：
```json
{
  "authorization_url": "https://auth.example.com/authorize?response_type=code&client_id=your_client_id&redirect_uri=https://avatarai.social/api/mcp/oauth/callback&resource=https://mcp-server.example.com&state=generated_state&code_challenge=generated_challenge&code_challenge_method=S256&scope=openid+profile",
  "state": "generated_state"
}
```

用户需要访问 `authorization_url` 进行授权。

#### 处理授权回调
```http
GET /api/mcp/oauth/callback?code=authorization_code&state=generated_state
```

响应示例：
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "refresh_token_here",
  "scope": "openid profile mcp:access"
}
```

### 4. Token 管理阶段

#### 刷新 Token
```http
POST /api/mcp/oauth/refresh-token?refresh_token=your_refresh_token&auth_server_url=https://auth.example.com&client_id=your_client_id
```

响应示例：
```json
{
  "access_token": "new_access_token",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "new_refresh_token"
}
```

### 5. 资源访问阶段

#### 使用 Token 访问资源
```http
GET /api/mcp/oauth/access-resource?resource_url=https://mcp-server.example.com/mcp&access_token=your_access_token
```

响应示例：
```json
{
  "data": {
    "version": "2024-11-05",
    "capabilities": {
      "tools": {},
      "resources": {}
    }
  }
}
```

## 完整的认证流程示例

### 步骤 1: 发现资源和授权服务器
```bash
# 1. 发现资源元数据
curl "https://avatarai.social/api/mcp/oauth/discover-resource?resource_url=https://mcp-server.example.com"

# 2. 发现授权服务器元数据
curl "https://avatarai.social/api/mcp/oauth/discover-auth-server?auth_server_url=https://auth.example.com"
```

### 步骤 2: 客户端注册（如果需要）
```bash
curl -X POST "https://avatarai.social/api/mcp/oauth/register-client?registration_url=https://auth.example.com/register" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "AvatarAI Social MCP Client",
    "redirect_uris": ["https://avatarai.social/api/mcp/oauth/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"]
  }'
```

### 步骤 3: 开始授权流程
```bash
curl "https://avatarai.social/api/mcp/oauth/start-authorization?auth_server_url=https://auth.example.com&client_id=your_client_id&resource=https://mcp-server.example.com"
```

### 步骤 4: 用户授权
用户需要访问返回的 `authorization_url` 进行授权。

### 步骤 5: 获取 Token
授权成功后，授权服务器会重定向到回调 URL，系统自动处理并返回 Token。

### 步骤 6: 使用 Token 访问资源
```bash
curl "https://avatarai.social/api/mcp/oauth/access-resource?resource_url=https://mcp-server.example.com/mcp&access_token=your_access_token"
```

### 步骤 7: 刷新 Token（当需要时）
```bash
curl -X POST "https://avatarai.social/api/mcp/oauth/refresh-token?refresh_token=your_refresh_token&auth_server_url=https://auth.example.com&client_id=your_client_id"
```

## 数据存储

系统会自动存储以下信息：

1. **OAuth 会话** (`mcp_oauth_sessions`): 存储授权流程中的临时会话信息
2. **OAuth Token** (`mcp_oauth_tokens`): 存储获取到的访问令牌和刷新令牌
3. **客户端注册** (`mcp_client_registrations`): 存储动态注册的客户端信息

## 安全考虑

1. **PKCE**: 所有授权流程都使用 PKCE (Proof Key for Code Exchange) 来增强安全性
2. **State 参数**: 使用随机生成的 state 参数防止 CSRF 攻击
3. **Token 过期**: 系统会自动处理 Token 过期和刷新
4. **会话清理**: 使用后的会话信息会被自动清理

## 错误处理

所有接口都会返回适当的 HTTP 状态码和错误信息：

- `400 Bad Request`: 参数错误或无效请求
- `401 Unauthorized`: 认证失败
- `500 Internal Server Error`: 服务器内部错误

错误响应格式：
```json
{
  "error": "错误描述",
  "error_description": "详细错误信息（可选）"
}
```

## 注意事项

1. 这是一个基础框架实现，实际使用时需要根据具体的 MCP 服务器和授权服务器进行调整
2. 用户认证集成：在实际应用中，需要集成用户认证系统来关联 OAuth Token 和用户
3. 错误处理：可以根据需要增加更详细的错误处理和重试逻辑
4. 监控和日志：建议添加适当的监控和日志记录