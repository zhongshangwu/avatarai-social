package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerInfo struct {
	McpId               string                 `json:"mcp_id"`
	UserID              string                 `json:"user_id"`
	IsBuiltin           bool                   `json:"is_builtin"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	About               string                 `json:"about"`
	Icon                string                 `json:"icon"`
	Schema              string                 `json:"schema"`
	SchemaKind          string                 `json:"schema_kind"`
	Endpoint            *MCPServerEndpoint     `json:"endpoint"`
	Version             string                 `json:"version"`
	ProtocolVersion     string                 `json:"protocol_version"`
	Capabilities        mcp.ServerCapabilities `json:"capabilities"`
	Instructions        *string                `json:"instructions"`
	Author              string                 `json:"author"`
	Authorization       MCPServerAuthorization `json:"authorization"` // 授权信息
	Status              MCPServerStatus        `json:"status"`        // 连接状态, 现在没有 local mcp server， 所以这里连接状态只要启用都是 connected
	Error               *string                `json:"error"`
	Enabled             bool                   `json:"enabled"`        // 是否开启
	SyncResources       bool                   `json:"sync_resources"` // 是否开启同步资源到 PDS
	CreatedAt           int64                  `json:"created_at"`
	UpdatedAt           int64                  `json:"updated_at"`
	LastSyncResourcesAt int64                  `json:"last_sync_resources_at"`
}

type MCPServerAuthorization struct {
	Method      MCPServerAuthorizationMethod `json:"method"`
	Status      MCPServerAuthorizationStatus `json:"status"`
	Scopes      string                       `json:"scopes"`
	Config      map[string]string            `json:"config"`      // 配置: client_id, client_secret, scopes, etc.
	Credentials map[string]string            `json:"credentials"` // 凭证: access_token, refresh_token, expires_in, etc.
	ExpireAt    int64                        `json:"expire_at"`
}

type MCPServerEndpoint struct {
	Type    MCPServerEndpointType `json:"type"`
	Command string                `json:"command"`
	Args    []string              `json:"args"`
	Env     map[string]string     `json:"env"`
	Url     string                `json:"url"`
	Headers map[string]string     `json:"headers"`
}

type MCPServerEndpointType string

const (
	MCPServerEndpointTypeStdio          MCPServerEndpointType = "stdio"
	MCPServerEndpointTypeSSE            MCPServerEndpointType = "sse"
	MCPServerEndpointTypeStreamableHttp MCPServerEndpointType = "streamableHttp"
)

type MCPServerAuthorizationMethod string

const (
	MCPServerAuthorizationMethodNone   MCPServerAuthorizationMethod = "none"
	MCPServerAuthorizationMethodOAuth2 MCPServerAuthorizationMethod = "oauth2"
)

type MCPServerStatus string

const (
	MCPServerStatusConnected    MCPServerStatus = "connected"
	MCPServerStatusDisconnected MCPServerStatus = "disconnected"
	MCPServerStatusConnecting   MCPServerStatus = "connecting"
)

type MCPServerAuthorizationStatus string

const (
	MCPServerAuthorizationStatusActive   MCPServerAuthorizationStatus = "active"
	MCPServerAuthorizationStatusExpired  MCPServerAuthorizationStatus = "expired"
	MCPServerAuthorizationStatusDisabled MCPServerAuthorizationStatus = "disabled"
)
