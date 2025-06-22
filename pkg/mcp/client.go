package mcp

import (
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/types"
)

type MCPClient struct {
	ServerInfo types.MCPServerInfo
	client     *mcpclient.Client
}

func NewMCPClient(serverInfo types.MCPServerInfo) *MCPClient {
	var client *mcpclient.Client
	var err error
	switch serverInfo.Endpoint.Type {
	case types.MCPServerEndpointTypeStdio:
		client = nil
	case types.MCPServerEndpointTypeSSE:
		client = nil
	case types.MCPServerEndpointTypeStreamableHttp:
		oAuthConfig := mcpclient.OAuthConfig{
			// ClientID:     serverInfo.ClientID,
			// ClientSecret: serverInfo.ClientSecret,
			// RedirectURI:  serverInfo.RedirectURI,
			// Scopes:       serverInfo.Scopes,
		}
		client, err = mcpclient.NewOAuthStreamableHttpClient(serverInfo.Endpoint.Url, oAuthConfig)
	default:
		return nil
	}
	if err != nil {
		logrus.WithError(err).Errorf("Failed to create OAuth Streamable Http Client: %v", err)
	}
	return &MCPClient{ServerInfo: serverInfo, client: client}
}

type TwitterMCPClient struct {
	ServerInfo types.MCPServerInfo
	client     *mcpclient.Client
}

func NewTwitterMCPClient(serverInfo types.MCPServerInfo) *TwitterMCPClient {
	var client *mcpclient.Client
	var err error
	switch serverInfo.Endpoint.Type {
	case types.MCPServerEndpointTypeStdio:
		client = nil
	case types.MCPServerEndpointTypeSSE:
		client = nil
	case types.MCPServerEndpointTypeStreamableHttp:
		oAuthConfig := mcpclient.OAuthConfig{
			ClientID:     "VC1yaFhoWktuVzhEdGxTUjF6VEI6MTpjaQ",
			ClientSecret: "XfwfAPzjsgPiGQ_ZYneJwADcOyXAIXBxZlO6rt0pD8Duih9MBN",
			RedirectURI:  "https://avatarai.social/api/mcp/oauth-callback",
			Scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access", "follows.read", "follows.write"},
		}
		client, err = mcpclient.NewOAuthStreamableHttpClient(serverInfo.Endpoint.Url, oAuthConfig)
	default:
		return nil
	}
	if err != nil {
		logrus.WithError(err).Errorf("Failed to create OAuth Streamable Http Client: %v", err)
	}
	return &TwitterMCPClient{ServerInfo: serverInfo, client: client}
}
