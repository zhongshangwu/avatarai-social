# 大致流程

- AvatarAI: OAuth Client + MCP Client
  - redirect url


# Neon OAuth
```mermaid
sequenceDiagram
    participant MC as MCP Client<br/>(Cursor/Claude Desktop)
    participant MS as MCP Server<br/>(Neon OAuth Server)
    participant NS as Neon OAuth Server<br/>(Upstream)
    participant U as User Browser

    Note over MC, NS: 1. 客户端注册阶段
    MC->>MS: POST /register<br/>{client_name, redirect_uris, grant_types}
    MS->>MS: 生成 clientId & clientSecret
    MS->>MC: {client_id, client_secret}

    Note over MC, NS: 2. 授权码流程开始
    MC->>MS: GET /authorize?client_id=xxx&<br/>redirect_uri=mcp://callback&<br/>response_type=code&scope=xxx&<br/>code_challenge=xxx&code_challenge_method=S256
    MS->>MS: 验证客户端和参数
    MS->>MS: 检查是否已授权过该客户端

    alt 首次授权
        MS->>U: 显示授权确认页面
        U->>MS: POST /authorize (用户确认)
    end

    MS->>MS: 编码请求参数到state
    MS->>NS: 重定向到上游Neon OAuth<br/>GET /authorize?state=encoded_params

    Note over MC, NS: 3. 上游认证
    NS->>U: 显示Neon登录页面
    U->>NS: 用户登录并授权
    NS->>MS: GET /callback?code=upstream_code&state=xxx

    Note over MC, NS: 4. 回调处理
    MS->>NS: POST /token (交换上游授权码)
    NS->>MS: {access_token, refresh_token, id_token}
    MS->>NS: GET /user (获取用户信息)
    NS->>MS: {user_info}

    MS->>MS: 生成新的授权码 (grantId:nonce)
    MS->>MS: 保存授权码和token映射
    MS->>MC: 重定向 mcp://callback?code=new_auth_code

    Note over MC, NS: 5. Token交换
    MC->>MS: POST /token<br/>{grant_type: authorization_code,<br/>code: auth_code, code_verifier: xxx}
    MS->>MS: 验证授权码和PKCE
    MS->>MS: 保存访问令牌
    MS->>MC: {access_token, refresh_token, expires_in}

    Note over MC, NS: 6. MCP连接建立
    MC->>MS: GET /sse<br/>Authorization: Bearer access_token
    MS->>MS: 验证访问令牌
    MS->>MS: 创建SSE传输和MCP服务器实例
    MS-->>MC: SSE连接建立 (持续连接)

    Note over MC, NS: 7. MCP工具调用
    MC->>MS: POST /messages?sessionId=xxx<br/>{tool_call_request}
    MS->>NS: 使用存储的Neon token调用API
    NS->>MS: API响应
    MS-->>MC: SSE事件推送结果
```

# MCP 完整认证流程图

## 流程时序图

```mermaid
sequenceDiagram
    participant C as MCP Client
    participant M as MCP Server
    participant AS as Authorization Server
    participant U as User

    Note over C,AS: 1. 初始请求与发现阶段
    C->>M: GET /mcp (无 Token)
    M->>C: 401 + WWW-Authenticate<br/>(资源元数据 URL)

    C->>M: GET /.well-known/oauth-protected-resource
    M->>C: 资源元数据<br/>(authorization_servers)

    C->>AS: GET /.well-known/oauth-authorization-server
    AS->>C: 授权服务器元数据<br/>(endpoints, capabilities)

    Note over C,AS: 2. 客户端注册阶段（可选）
    alt 支持动态注册
        C->>AS: POST /register<br/>(client metadata)
        AS->>C: client_id, client_secret
    else 静态配置
        Note over C: 使用预配置凭证
    end

    Note over C,U: 3. 授权流程阶段
    Note over C: 生成 PKCE 参数<br/>(code_verifier, code_challenge)

    C->>AS: GET /authorize<br/>(client_id, redirect_uri, code_challenge, resource, state)
    AS->>U: 重定向到授权页面
    U->>AS: 用户登录并授权
    AS->>C: 重定向回调<br/>(code, state)

    Note over C,AS: 4. Token 获取阶段
    C->>AS: POST /token<br/>(code, code_verifier, client_id, resource)
    AS->>C: access_token, refresh_token, expires_in

    Note over C,M: 5. 资源访问阶段
    C->>M: GET /mcp<br/>Authorization: Bearer <token>
    M->>AS: Token 验证 (可选)
    AS->>M: Token 有效性确认
    M->>C: 返回 MCP 资源

    Note over C,AS: 6. Token 管理阶段
    alt Token 过期时
        C->>AS: POST /token<br/>(refresh_token, grant_type=refresh_token)
        AS->>C: 新的 access_token
    end
```


## 决策流程图

```mermaid
flowchart TD
    A[MCP Client 启动] --> B{选择传输方式}

    B -->|HTTP| C[发起 HTTP 请求]
    B -->|STDIO| D[读取环境变量凭证]
    B -->|其他协议| E[使用协议特定认证]

    C --> F{是否需要认证?}
    F -->|否| G[直接访问资源]
    F -->|是| H[收到 401 响应]

    H --> I[获取资源元数据]
    I --> J[发现授权服务器]
    J --> K{支持动态注册?}

    K -->|是| L[动态注册客户端]
    K -->|否| M[使用静态配置]

    L --> N[生成 PKCE 参数]
    M --> N

    N --> O[构造授权请求]
    O --> P[用户授权]
    P --> Q[获取授权码]
    Q --> R[请求 Access Token]
    R --> S[带 Token 访问资源]

    S --> T{Token 有效?}
    T -->|是| U[返回资源]
    T -->|否| V{有 Refresh Token?}

    V -->|是| W[刷新 Token]
    V -->|否| H
    W --> S

    D --> X[本地认证成功]
    E --> Y[协议认证成功]
    X --> G
    Y --> G
```

## 错误处理流程

```mermaid
flowchart TD
    A[发送请求] --> B{响应状态}

    B -->|200 OK| C[请求成功]
    B -->|401 Unauthorized| D{首次请求?}
    B -->|403 Forbidden| E[权限不足错误]
    B -->|400 Bad Request| F[请求格式错误]

    D -->|是| G[开始认证流程]
    D -->|否| H{Token 过期?}

    H -->|是| I[刷新 Token]
    H -->|否| J[Token 无效错误]

    I --> K{刷新成功?}
    K -->|是| L[重新发送请求]
    K -->|否| M[重新认证]

    G --> N[OAuth 认证流程]
    M --> N
    L --> A
```

## 安全检查清单

```mermaid
flowchart TD
    A[安全检查开始] --> B{使用 HTTPS?}
    B -->|否| C[❌ 必须使用 HTTPS]
    B -->|是| D{实现 PKCE?}

    D -->|否| E[❌ 必须实现 PKCE]
    D -->|是| F{验证 redirect_uri?}

    F -->|否| G[❌ 必须验证重定向 URI]
    F -->|是| H{使用 resource 参数?}

    H -->|否| I[❌ 必须使用 resource 参数]
    H -->|是| J{验证 Token audience?}

    J -->|否| K[❌ 必须验证 Token audience]
    J -->|是| L{Token 短期有效?}

    L -->|否| M[⚠️ 建议缩短 Token 有效期]
    L -->|是| N[✅ 安全检查通过]

    M --> N
```

# Github MCP Server OAuth

```
sequenceDiagram
    participant B as User-Agent (Browser)
    participant C as Client
    participant M as MCP Server (Resource Server)
    participant A as Authorization Server

    C->>M: MCP request without token
    M->>C: HTTP 401 Unauthorized with WWW-Authenticate header
    Note over C: Extract resource_metadata URL from WWW-Authenticate

    C->>M: Request Protected Resource Metadata
    M->>C: Return metadata

    Note over C: Parse metadata and extract authorization server(s)<br/>Client determines AS to use

    C->>A: GET /.well-known/oauth-authorization-server
    A->>C: Authorization server metadata response

    alt Dynamic client registration
        C->>A: POST /register
        A->>C: Client Credentials
    end

    Note over C: Generate PKCE parameters
    C->>B: Open browser with authorization URL + code_challenge
    B->>A: Authorization request
    Note over A: User authorizes
    A->>B: Redirect to callback with authorization code
    B->>C: Authorization code callback
    C->>A: Token request + code_verifier
    A->>C: Access token (+ refresh token)
    C->>M: MCP request with access token
    M-->>C: MCP response
    Note over C,M: MCP communication continues with valid token
```

https://api.notion.com/v1/oauth/authorize?client_id=15ed872b-594c-817b-a08c-0037362900ad&response_type=code&owner=user&redirect_uri=https%3A%2F%2Favatarai.social%2Fapi%2Fmcp%2Foauth%2Fcallback

export MCP_GITHUB_GITHUB_CLIENT_ID="Ov23liXZ68YbB4ILHsyg"
export MCP_GITHUB_GITHUB_CLIENT_SECRET="a7c79cea7177603c833e8b310736a81a8d033f6b"






{
    "servers": [
        {
            "mcpId": "notion-mcp",
            "userId": "did:plc:mop7aiqx3dgxciovzmx7o6xe",
            "isBuiltin": false,
            "name": "Notion MCP Server",
            "description": "Notion MCP 允许您使用Notion API 和第三方客户端（如Cursor）进行交互。要使用Notion MCP，您需要在Notion中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的Notion页面和数据库。",
            "about": "",
            "icon": "",
            "schema": "",
            "schemaKind": "",
            "endpoint": {
                "type": "streamableHttp",
                "command": "",
                "args": null,
                "env": null,
                "url": "http://localhost:8091/mcp",
                "headers": {}
            },
            "version": "1.0.0",
            "protocolVersion": "1.0.0",
            "capabilities": {},
            "instructions": null,
            "author": "AvatarAI",
            "authorization": {
                "method": "none",
                "status": "inactive",
                "scopes": "",
                "config": null,
                "credentials": null,
                "expireAt": 0
            },
            "status": "disconnected",
            "error": null,
            "enabled": false,
            "syncResources": false,
            "createdAt": 1750774000,
            "updatedAt": 1750774000,
            "lastSyncResourcesAt": 0
        },
        {
            "mcpId": "github-mcp",
            "userId": "did:plc:mop7aiqx3dgxciovzmx7o6xe",
            "isBuiltin": false,
            "name": "GitHub MCP Server",
            "description": "GitHub MCP Server 允许您使用GitHub API 和第三方客户端（如Cursor）进行交互。要使用GitHub MCP，您需要在GitHub中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的GitHub仓库。",
            "about": "",
            "icon": "",
            "schema": "",
            "schemaKind": "",
            "endpoint": {
                "type": "streamableHttp",
                "command": "",
                "args": null,
                "env": null,
                "url": "http://localhost:8089/mcp",
                "headers": {}
            },
            "version": "1.2.0",
            "protocolVersion": "1.0.0",
            "capabilities": {},
            "instructions": null,
            "author": "AvatarAI",
            "authorization": {
                "method": "oauth2",
                "status": "active",
                "scopes": "repo",
                "config": {
                    "client_id": "Ov23liXZ68YbB4ILHsyg",
                    "client_secret": "a7c79cea7177603c833e8b310736a81a8d033f6b",
                    "redirect_uri": "https://avatarai.social/api/mcp/oauth/callback"
                },
                "credentials": {
                    "access_token": "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",
                    "expires_at": "0001-01-01T00:00:00Z",
                    "scope": "repo",
                    "token_type": "bearer"
                },
                "expireAt": -62135596800
            },
            "status": "disconnected",
            "error": null,
            "enabled": false,
            "syncResources": false,
            "createdAt": 1750772013,
            "updatedAt": 1750772013,
            "lastSyncResourcesAt": 0
        },
        {
            "mcpId": "twitter-mcp",
            "userId": "did:plc:mop7aiqx3dgxciovzmx7o6xe",
            "isBuiltin": false,
            "name": "Twitter MCP Server",
            "description": "Twitter MCP Server 允许您使用Twitter API 和第三方客户端（如Cursor）进行交互。要使用Twitter MCP，您需要在Twitter中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的Twitter账号。",
            "about": "",
            "icon": "",
            "schema": "",
            "schemaKind": "",
            "endpoint": {
                "type": "streamableHttp",
                "command": "",
                "args": null,
                "env": null,
                "url": "http://localhost:8090/mcp",
                "headers": {}
            },
            "version": "1.0.0",
            "protocolVersion": "1.0.0",
            "capabilities": {},
            "instructions": null,
            "author": "AvatarAI",
            "authorization": {
                "method": "oauth2",
                "status": "active",
                "scopes": "follows.read offline.access tweet.write media.write like.write like.read users.read tweet.read follows.write",
                "config": {
                    "client_id": "VC1yaFhoWktuVzhEdGxTUjF6VEI6MTpjaQ",
                    "client_secret": "EjdsctDBgUAaKOYmtTrKtlawGxtBYYQA5qk29XCnwSJfFhHFJH",
                    "redirect_uri": "https://avatarai.social/api/mcp/oauth/callback"
                },
                "credentials": {
                    "access_token": "S3AtQWZiOWt5OFVON2xmWUN5ajVsc1JXdDNJbzlOQk04Q3FnUVFORVgxWDNyOjE3NTA3NzI1NTk2OTM6MTowOmF0OjE",
                    "expires_at": "2025-06-24T23:42:39.898818343+08:00",
                    "expires_in": 7200,
                    "refresh_token": "S3pNUnBneElVeW1KSThLREtBeVFXTjVjOE5hV3BhdGMyYmhHcDlaaU9zRkJKOjE3NTA3NzI1NTk2OTM6MTowOnJ0OjE",
                    "scope": "follows.read offline.access tweet.write media.write like.write like.read users.read tweet.read follows.write",
                    "token_type": "bearer"
                },
                "expireAt": 1750779759
            },
            "status": "connected",
            "error": null,
            "enabled": true,
            "syncResources": false,
            "createdAt": 1750771409,
            "updatedAt": 1750772959,
            "lastSyncResourcesAt": 0
        }
    ]
}