package mcp

import (
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerInfo struct {
	ID                  string
	Name                string
	Description         string
	About               string
	Icon                string
	Schema              string
	SchemaKind          string
	Endpoint            *MCPServerEndpoint
	Version             string
	ProtocolVersion     string
	Capabilities        mcp.ServerCapabilities // json string
	Instructions        *string
	Author              string
	AuthorzationMethod  MCPServerAuthorizationMethod
	Disabled            bool
	Status              MCPServerStatus
	Error               *string
	UserID              string
	CreatedAt           int64
	UpdatedAt           int64
	LastSyncResourcesAt int64
}

type MCPServerEndpoint struct {
	Type    MCPServerEndpointType
	Command string
	Args    []string
	Env     map[string]string
	Url     string
	Headers map[string]string
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

func NewTwitterServerInfo() *MCPServerInfo {
	return &MCPServerInfo{
		ID:          "twitter",
		Name:        "Twitter",
		Description: "Twitter MCP Server",
		About:       "Twitter MCP Server",
		Icon:        "https://twitter.com/favicon.ico",
		Schema:      "https://twitter.com/schema.json",
		Endpoint: &MCPServerEndpoint{
			Type:    MCPServerEndpointTypeStreamableHttp,
			Url:     "https://api.twitter.com/2/mcp",
			Headers: map[string]string{},
		},
		AuthorzationMethod:  MCPServerAuthorizationMethodOAuth2,
		Disabled:            false,
		Status:              MCPServerStatusConnected,
		Error:               nil,
		Capabilities:        mcp.ServerCapabilities{},
		Instructions:        nil,
		Author:              "AvatarAI",
		UserID:              "1234567890",
		CreatedAt:           time.Now().Unix(),
		UpdatedAt:           time.Now().Unix(),
		LastSyncResourcesAt: time.Now().Unix(),
	}
}
