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

github oauth apps
Ov23liXZ68YbB4ILHsyg
a7c79cea7177603c833e8b310736a81a8d033f6b


twitter oauth


export MCP_GITHUB_GITHUB_CLIENT_ID="Ov23liXZ68YbB4ILHsyg"
export MCP_GITHUB_GITHUB_CLIENT_SECRET="a7c79cea7177603c833e8b310736a81a8d033f6b"