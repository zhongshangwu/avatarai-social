package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcpclienttransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type DBTokenStore struct {
	mcpId      string
	userDid    string
	serverInfo *MCPServerInfo

	metaStore *repositories.MetaStore

	token *mcpclient.Token
	mu    sync.RWMutex
}

func NewDBTokenStore(metaStore *repositories.MetaStore, serverInfo *MCPServerInfo) *DBTokenStore {
	if serverInfo.Authorization.Method != MCPServerAuthorizationMethodOAuth2 {
		return nil
	}

	var token *mcpclient.Token
	if serverInfo.Authorization.Status == repositories.AuthStatusActive {
		token = &mcpclient.Token{
			AccessToken:  GetString(serverInfo.Authorization.Credentials, "access_token"),
			RefreshToken: GetString(serverInfo.Authorization.Credentials, "refresh_token"),
			Scope:        serverInfo.Authorization.Scopes,
			TokenType:    "Bearer",
			ExpiresAt:    time.Unix(serverInfo.Authorization.ExpireAt, 0),
		}
	}

	return &DBTokenStore{
		metaStore:  metaStore,
		token:      token,
		mu:         sync.RWMutex{},
		mcpId:      serverInfo.McpId,
		userDid:    serverInfo.UserID,
		serverInfo: serverInfo,
	}
}

func (s *DBTokenStore) GetToken() (*mcpclient.Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.token == nil {
		return nil, errors.New("no token available")
	}
	return s.token, nil
}

func (s *DBTokenStore) SaveToken(token *mcpclient.Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.token = token

	credentials, _ := json.Marshal(token)
	config, _ := json.Marshal(s.serverInfo.Authorization.Config)

	auth := &repositories.MCPServerAuth{
		McpId:       s.mcpId,
		UserDid:     s.userDid,
		Scope:       token.Scope,
		Status:      repositories.AuthStatusActive,
		AuthConfig:  string(config),
		AuthMethod:  string(MCPServerAuthorizationMethodOAuth2),
		Credentials: string(credentials),
		ExpiresAt:   token.ExpiresAt.Unix(),
		UpdatedAt:   time.Now().Unix(),
		CreatedAt:   time.Now().Unix(),
	}
	return s.metaStore.MCPRepo.CreateOrUpdateMCPServerAuth(auth)
}

type MCPClient struct {
	ServerInfo *MCPServerInfo

	client       *mcpclient.Client
	oauthHandler *mcpclienttransport.OAuthHandler
}

// 使用本地修改版本的库后，这些函数就不再需要了

func NewMCPClient(metaStore *repositories.MetaStore, serverInfo *MCPServerInfo) (*MCPClient, error) {
	var client *mcpclient.Client
	var oauthHandler *mcpclienttransport.OAuthHandler
	var err error

	if serverInfo.Authorization.Method != MCPServerAuthorizationMethodOAuth2 {
		return nil, fmt.Errorf("invalid authorization method: %s", serverInfo.Authorization.Method)
	} else {
		tokenStore := NewDBTokenStore(metaStore, serverInfo)
		oAuthConfig := mcpclient.OAuthConfig{
			ClientID:     GetString(serverInfo.Authorization.Config, "client_id"),
			ClientSecret: GetString(serverInfo.Authorization.Config, "client_secret"),
			RedirectURI:  GetString(serverInfo.Authorization.Config, "redirect_uri"),
			Scopes:       strings.Split(serverInfo.Authorization.Scopes, " "),
			PKCEEnabled:  true,
			TokenStore:   tokenStore,
		}
		oauthHandler = mcpclienttransport.NewOAuthHandler(oAuthConfig)
		baseURL, err := extractBaseURL(serverInfo.Endpoint.Url)
		if err != nil {
			logrus.WithError(err).Errorf("Failed to extract base URL: %v", err)
			return nil, err
		}
		oauthHandler.SetBaseURL(baseURL)
		oauthHandler.SetHTTPClient(mcpclienttransport.CreateDebugHTTPClientWithProxy("", ""))
		// oauthHandler.SetHTTPClient(mcpclienttransport.CreateDebugHTTPClientWithProxy("172.26.128.1", "7890"))
		switch serverInfo.Endpoint.Type {
		default:
			return nil, fmt.Errorf("invalid endpoint type: %s", serverInfo.Endpoint.Type)
		case MCPServerEndpointTypeStdio:
			client = nil
		case MCPServerEndpointTypeSSE:
			client, err = mcpclient.NewOAuthSSEClient(serverInfo.Endpoint.Url, oAuthConfig)
		case MCPServerEndpointTypeStreamableHttp:
			client, err = mcpclient.NewOAuthStreamableHttpClient(serverInfo.Endpoint.Url, oAuthConfig)
		}
	}

	if err != nil {
		logrus.WithError(err).Errorf("Failed to create OAuth Streamable Http Client: %v", err)
		return nil, err
	}
	return &MCPClient{ServerInfo: serverInfo, client: client, oauthHandler: oauthHandler}, nil
}

func (c *MCPClient) GenerateCodeChallenge() (string, string, error) {
	codeVerifier, err := mcpclient.GenerateCodeVerifier()
	if err != nil {
		logrus.WithError(err).Errorf("Failed to generate code verifier: %v", err)
		return "", "", err
	}
	codeChallenge := mcpclient.GenerateCodeChallenge(codeVerifier)
	return codeVerifier, codeChallenge, nil
}

func (c *MCPClient) GenerateState() (string, error) {
	state, err := mcpclient.GenerateState()
	if err != nil {
		logrus.WithError(err).Errorf("Failed to generate state: %v", err)
		return "", err
	}
	return state, nil
}

func (c *MCPClient) GetAuthorizationURL(ctx context.Context, state string, codeChallenge string) (authURL string, err error) {
	authURL, err = c.oauthHandler.GetAuthorizationURL(ctx, state, codeChallenge)
	logrus.Infof("Authorization URL: %s", authURL)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get authorization URL: %v", err)
		return "", err
	}
	return authURL, nil
}

func (c *MCPClient) ExchangeCode(ctx context.Context, expectedState string, code string, state string, codeVerifier string) error {
	c.oauthHandler.SetExpectedState(expectedState)
	err := c.oauthHandler.ProcessAuthorizationResponse(ctx, code, state, codeVerifier)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to exchange code: %v", err)
		return err
	}
	return nil
}

func extractBaseURL(mcpServerURL string) (string, error) {
	parsedURL, err := url.Parse(mcpServerURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), nil
}

func GetString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return v.(string)
	}
	return ""
}
