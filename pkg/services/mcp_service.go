package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	mcptypes "github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/mcp"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type MCPService struct {
	metaStore *repositories.MetaStore
}

func NewMCPService(metaStore *repositories.MetaStore) *MCPService {
	return &MCPService{
		metaStore: metaStore,
	}
}

func (s *MCPService) ListMCPServers(userDid string) ([]*mcp.MCPServerInfo, error) {
	builtinServers := s.getBuiltinMCPServers()

	dbServers, err := s.metaStore.MCPRepo.GetMCPServersByUser(userDid)
	if err != nil {
		return nil, err
	}

	servers := make([]*mcp.MCPServerInfo, 0, len(dbServers))
	for _, dbServer := range dbServers {
		serverInfo, err := s.convertDBServerToAPIServer(dbServer)
		if err != nil {
			logrus.WithError(err).Error("convertDBServerToAPIServer failed")
			return nil, err
		}
		servers = append(servers, serverInfo)
	}

	allServers := s.OverrideMCPServers(builtinServers, servers)
	return allServers, nil
}

func (s *MCPService) OverrideMCPServers(builtinServers []*mcp.MCPServerInfo, dbServers []*mcp.MCPServerInfo) []*mcp.MCPServerInfo {
	allServers := make([]*mcp.MCPServerInfo, 0)
	installed := make([]string, 0)
	for _, dbServer := range dbServers {
		installed = append(installed, dbServer.McpId)
	}
	for _, builtinServer := range builtinServers {
		if !slices.Contains(installed, builtinServer.McpId) {
			allServers = append(allServers, builtinServer)
		}
	}
	return allServers
}

func (s *MCPService) GetMCPServerDetail(mcpID string, userDid string) (*mcp.MCPServerInfo, error) {
	var builtinServer *mcp.MCPServerInfo
	builtinServers := s.getBuiltinMCPServers()
	for _, server := range builtinServers {
		if server.McpId == mcpID {
			builtinServer = server
			break
		}
	}
	dbServer, err := s.metaStore.MCPRepo.GetMCPServerByIDAndUser(mcpID, userDid)
	if err != nil {
		return nil, err
	}
	if dbServer == nil {
		return builtinServer, nil
	}
	dbServerInfo, err := s.convertDBServerToAPIServer(dbServer)
	if err != nil {
		return nil, err
	}
	return s.OverrideMCPServers([]*mcp.MCPServerInfo{builtinServer}, []*mcp.MCPServerInfo{dbServerInfo})[0], nil
}

func (s *MCPService) GetMCPServerAuth(mcpID string, userDid string) (*repositories.MCPServerAuth, error) {
	var auth repositories.MCPServerAuth
	err := s.metaStore.DB.Where("mcp_id = ? AND user_did = ?", mcpID, userDid).First(&auth).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func (s *MCPService) AddMCPServer(name string, endpoint *mcp.MCPServerEndpoint, userDid string) (string, error) {
	dbServer := &repositories.MCPServer{
		McpID:       name,
		UserDid:     userDid,
		Name:        name,
		Description: "",
		About:       "",
		Icon:        "",
	}

	if err := s.metaStore.MCPRepo.CreateMCPServer(dbServer); err != nil {
		return "", err
	}

	dbEndpoint, err := s.convertAPIEndpointToDBEndpoint(endpoint, dbServer.McpID)
	if err != nil {
		return "", err
	}
	if err := s.metaStore.MCPRepo.CreateOrUpdateMCPServerEndpoint(dbEndpoint); err != nil {
		return "", err
	}
	return dbServer.McpID, nil
}

func (s *MCPService) DeleteMCPServer(mcpID string, userDid string) error {
	builtinServers := s.getBuiltinMCPServers()
	for _, server := range builtinServers {
		if server.McpId == mcpID {
			return fmt.Errorf("cannot_delete_builtin")
		}
	}
	return s.metaStore.MCPRepo.DeleteMCPServer(mcpID, userDid)
}

func (s *MCPService) UpdateSyncResourcesStatus(mcpID string, userDid string, syncResources bool) error {
	return s.metaStore.MCPRepo.UpdateSyncResourcesStatus(mcpID, userDid, syncResources)
}

func (s *MCPService) EnforceToDB(serverInfo *mcp.MCPServerInfo, userDid string) error {
	dbServer, err := s.metaStore.MCPRepo.GetMCPServerByIDAndUser(serverInfo.McpId, userDid)
	if err != nil {
		return err
	}
	if dbServer == nil {
		dbServer, err = s.convertAPIServerToDBServer(serverInfo, userDid)
		if err != nil {
			return err
		}
		if err := s.metaStore.MCPRepo.CreateMCPServer(dbServer); err != nil {
			return err
		}

		dbEndpoint, err := s.convertAPIEndpointToDBEndpoint(serverInfo.Endpoint, serverInfo.McpId)
		if err != nil {
			return err
		}
		if err := s.metaStore.MCPRepo.CreateOrUpdateMCPServerEndpoint(dbEndpoint); err != nil {
			return err
		}

		dbAuth := s.convertAPIServerToDBAuth(&serverInfo.Authorization, serverInfo.McpId, serverInfo.UserID)
		if err := s.metaStore.MCPRepo.CreateOrUpdateMCPServerAuth(dbAuth); err != nil {
			return err
		}
	}

	return nil
}

func (s *MCPService) UpdateEnabledStatus(mcpID string, userDid string, enabled bool) error {
	return s.metaStore.MCPRepo.UpdateEnabledStatus(mcpID, userDid, enabled)
}

func (s *MCPService) CreateOAuthCode(mcpID string, userDid string, state string, codeVerifier string, codeChallenge string, challengeMethod string, redirectURI string, scope string, expiresAt int64) (*repositories.MCPServerOAuthCode, error) {
	dbCode := &repositories.MCPServerOAuthCode{
		McpID:           mcpID,
		UserDid:         userDid,
		State:           state,
		CodeVerifier:    codeVerifier,
		CodeChallenge:   codeChallenge,
		ChallengeMethod: challengeMethod,
		RedirectURI:     redirectURI,
		Scope:           scope,
		ExpiresAt:       expiresAt,
	}

	if err := s.metaStore.MCPRepo.CreateMCPServerOAuthCode(dbCode); err != nil {
		return nil, err
	}
	return dbCode, nil
}

func (s *MCPService) GetMCPServerOAuthCode(issuer string, state string) (*repositories.MCPServerOAuthCode, error) {
	dbCode, err := s.metaStore.MCPRepo.GetMCPServerOAuthCode(issuer, state)
	if err != nil {
		return nil, err
	}
	return dbCode, nil
}

func (s *MCPService) convertDBServerToAPIServer(dbServer *repositories.MCPServer) (*mcp.MCPServerInfo, error) {
	capabilities := mcptypes.ServerCapabilities{}
	if err := json.Unmarshal([]byte(dbServer.Capabilities), &capabilities); err != nil {
		return nil, err
	}
	serverInfo := &mcp.MCPServerInfo{
		McpId:           dbServer.McpID,
		IsBuiltin:       false,
		Name:            dbServer.Name,
		Description:     dbServer.Description,
		About:           dbServer.About,
		Icon:            dbServer.Icon,
		Author:          dbServer.Author,
		Version:         dbServer.Version,
		ProtocolVersion: dbServer.ProtocolVersion,
		Capabilities:    capabilities,
		Authorization: mcp.MCPServerAuthorization{
			Method: mcp.MCPServerAuthorizationMethodNone,
			Status: mcp.MCPServerAuthorizationStatusActive,
		},
		Enabled:             dbServer.Enabled,
		SyncResources:       dbServer.SyncResources,
		Status:              mcp.MCPServerStatusDisconnected, // 默认状态
		UserID:              dbServer.UserDid,
		CreatedAt:           dbServer.CreatedAt,
		UpdatedAt:           dbServer.UpdatedAt,
		LastSyncResourcesAt: dbServer.LastSyncResourcesAt,
	}

	if dbServer.Instructions != "" {
		serverInfo.Instructions = &dbServer.Instructions
	}
	if dbServer.Enabled {
		serverInfo.Status = mcp.MCPServerStatusConnected
	}

	endpoint, err := s.metaStore.MCPRepo.GetMCPServerEndpoint(dbServer.ID)
	if err != nil {
		return nil, err
	}
	if endpoint != nil {
		serverInfo.Endpoint, err = s.convertDBEndpointToAPIEndpoint(endpoint)
		if err != nil {
			return nil, err
		}
	}

	auth, err := s.metaStore.MCPRepo.GetMCPServerAuth(dbServer.ID)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		config := map[string]string{}
		if err := json.Unmarshal([]byte(auth.AuthConfig), &config); err != nil {
			return nil, err
		}
		credentials := map[string]string{}
		if err := json.Unmarshal([]byte(auth.Credentials), &credentials); err != nil {
			return nil, err
		}
		serverInfo.Authorization = mcp.MCPServerAuthorization{
			Method:      mcp.MCPServerAuthorizationMethod(auth.AuthMethod),
			Status:      mcp.MCPServerAuthorizationStatus(auth.Status),
			Scopes:      auth.Scope,
			Config:      config,
			Credentials: credentials,
			ExpireAt:    auth.ExpiresAt,
		}
	}

	return serverInfo, nil
}

func (s *MCPService) convertAPIServerToDBServer(serverInfo *mcp.MCPServerInfo, userDid string) (*repositories.MCPServer, error) {
	instructions := ""
	if serverInfo.Instructions != nil {
		instructions = *serverInfo.Instructions
	}
	dbServer := &repositories.MCPServer{
		McpID:           serverInfo.McpId,
		UserDid:         userDid,
		Name:            serverInfo.Name,
		Description:     serverInfo.Description,
		About:           serverInfo.About,
		Icon:            serverInfo.Icon,
		Author:          serverInfo.Author,
		Version:         serverInfo.Version,
		ProtocolVersion: serverInfo.ProtocolVersion,
		Instructions:    instructions,
		Enabled:         serverInfo.Enabled,
		SyncResources:   serverInfo.SyncResources,
	}
	return dbServer, nil
}

func (s *MCPService) convertDBEndpointToAPIEndpoint(dbEndpoint *repositories.MCPServerEndpoint) (*mcp.MCPServerEndpoint, error) {
	endpoint := &mcp.MCPServerEndpoint{
		Type:    mcp.MCPServerEndpointType(dbEndpoint.Type),
		Command: dbEndpoint.Command,
		Url:     dbEndpoint.URL,
	}

	if dbEndpoint.Args != "" {
		var args []string
		if err := json.Unmarshal([]byte(dbEndpoint.Args), &args); err != nil {
			return nil, err
		}
		endpoint.Args = args
	}

	if dbEndpoint.Env != "" {
		var env map[string]string
		if err := json.Unmarshal([]byte(dbEndpoint.Env), &env); err != nil {
			return nil, err
		}
		endpoint.Env = env
	}

	if dbEndpoint.Headers != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(dbEndpoint.Headers), &headers); err != nil {
			return nil, err
		}
		endpoint.Headers = headers
	}

	return endpoint, nil
}

func (s *MCPService) convertAPIEndpointToDBEndpoint(apiEndpoint *mcp.MCPServerEndpoint, mcpID string) (*repositories.MCPServerEndpoint, error) {
	endpoint := &repositories.MCPServerEndpoint{
		McpID:   mcpID,
		Type:    string(apiEndpoint.Type),
		Command: apiEndpoint.Command,
		Args:    marshalSlice(apiEndpoint.Args),
		Env:     marshalMap(apiEndpoint.Env),
		URL:     apiEndpoint.Url,
		Headers: marshalMap(apiEndpoint.Headers),
	}
	return endpoint, nil
}

func (s *MCPService) convertAPIServerToDBAuth(apiAuth *mcp.MCPServerAuthorization, mcpID string, userDid string) *repositories.MCPServerAuth {
	auth := &repositories.MCPServerAuth{
		McpId:       mcpID,
		UserDid:     userDid,
		AuthMethod:  string(apiAuth.Method),
		AuthConfig:  marshalMap(apiAuth.Config),
		Credentials: marshalMap(apiAuth.Credentials),
		Scope:       apiAuth.Scopes,
		Status:      string(apiAuth.Status),
		ExpiresAt:   apiAuth.ExpireAt,
		UpdatedAt:   time.Now().Unix(),
		CreatedAt:   time.Now().Unix(),
	}
	return auth
}

func (s *MCPService) GenerateMcpId() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *MCPService) getBuiltinMCPServers() []*mcp.MCPServerInfo {
	// 内置 mcp servers 的配置信息
	// 开关顺序: 授权 -> 是否启用 -> 是否同步资源
	// 将内置 mcp server 的配置信息存储到个人数据库的时机:
	// 1. 授权
	// 2. 启用
	return []*mcp.MCPServerInfo{
		{
			McpId:       "notion-mcp",
			IsBuiltin:   true,
			Name:        "Notion MCP Server",
			Description: "Notion MCP 允许您使用Notion API 和第三方客户端（如Cursor）进行交互。要使用Notion MCP，您需要在Notion中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的Notion页面和数据库。",
			Version:     "1.0.0",
			Author:      "AvatarAI",
			Status:      mcp.MCPServerStatusDisconnected,
			Endpoint: &mcp.MCPServerEndpoint{
				Type:    mcp.MCPServerEndpointTypeStreamableHttp,
				Url:     "https://api.notion.com/mcp",
				Headers: map[string]string{},
			},
			ProtocolVersion: "1.0.0",
			Capabilities:    mcptypes.ServerCapabilities{},
			Instructions:    nil,
			Authorization: mcp.MCPServerAuthorization{
				Method: mcp.MCPServerAuthorizationMethodOAuth2,
				Status: mcp.MCPServerAuthorizationStatusDisabled,
				Scopes: "notion.api",
				Config: map[string]string{
					"client_id":     "notion-mcp",
					"client_secret": "notion-mcp",
					"redirect_uri":  "https://avatarai.social/api/mcp/oauth-callback",
				},
				Credentials: map[string]string{},
			},
			Enabled:             false,
			SyncResources:       false,
			UpdatedAt:           time.Now().Unix(),
			CreatedAt:           time.Now().Unix(),
			LastSyncResourcesAt: time.Now().Unix(),
		},
		{
			McpId:       "github-mcp",
			IsBuiltin:   true,
			Name:        "GitHub MCP Server",
			Description: "GitHub MCP Server 允许您使用GitHub API 和第三方客户端（如Cursor）进行交互。要使用GitHub MCP，您需要在GitHub中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的GitHub仓库。",
			Version:     "1.2.0",
			Author:      "AvatarAI",
			Status:      mcp.MCPServerStatusDisconnected,
			Endpoint: &mcp.MCPServerEndpoint{
				Type:    mcp.MCPServerEndpointTypeStreamableHttp,
				Url:     "https://api.github.com/mcp",
				Headers: map[string]string{},
			},
			ProtocolVersion: "1.0.0",
			Capabilities:    mcptypes.ServerCapabilities{},
			Instructions:    nil,
			Authorization: mcp.MCPServerAuthorization{
				Method: mcp.MCPServerAuthorizationMethodOAuth2,
				Status: mcp.MCPServerAuthorizationStatusDisabled,
				Scopes: "repo",
				Config: map[string]string{
					"client_id":     "github-mcp",
					"client_secret": "github-mcp",
					"redirect_uri":  "https://avatarai.social/api/mcp/oauth-callback",
				},
				Credentials: map[string]string{},
			},
			Enabled:             false,
			SyncResources:       false,
			UpdatedAt:           time.Now().Unix(),
			CreatedAt:           time.Now().Unix(),
			LastSyncResourcesAt: time.Now().Unix(),
		},
		{
			McpId:       "twitter-mcp",
			IsBuiltin:   true,
			Name:        "Twitter MCP Server",
			Description: "Twitter MCP Server 允许您使用Twitter API 和第三方客户端（如Cursor）进行交互。要使用Twitter MCP，您需要在Twitter中创建集成，获取内部集成令牌，并在MCP客户端中配置这些信息，以便客户端可以访问和操作您的Twitter账号。",
			Version:     "1.0.0",
			Author:      "AvatarAI",
			Status:      mcp.MCPServerStatusConnected,
			Endpoint: &mcp.MCPServerEndpoint{
				Type:    mcp.MCPServerEndpointTypeStreamableHttp,
				Url:     "https://api.twitter.com/2/mcp",
				Headers: map[string]string{},
			},
			ProtocolVersion: "1.0.0",
			Capabilities:    mcptypes.ServerCapabilities{},
			Instructions:    nil,
			Authorization: mcp.MCPServerAuthorization{
				Method: mcp.MCPServerAuthorizationMethodOAuth2,
				Status: mcp.MCPServerAuthorizationStatusDisabled,
				Scopes: "tweet.read tweet.write users.read offline.access follows.read follows.write",
				Config: map[string]string{
					"client_id":     "twitter-mcp",
					"client_secret": "twitter-mcp",
					"redirect_uri":  "https://avatarai.social/api/mcp/oauth-callback",
				},
				Credentials: map[string]string{},
			},
			Enabled:             false,
			SyncResources:       false,
			UpdatedAt:           time.Now().Unix(),
			CreatedAt:           time.Now().Unix(),
			LastSyncResourcesAt: time.Now().Unix(),
		},
	}
}

func marshalMap(m map[string]string) string {
	bytes, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func marshalSlice(s []string) string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(bytes)
}
