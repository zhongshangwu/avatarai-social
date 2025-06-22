package services

import (
	"encoding/json"
	"time"

	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
	"gorm.io/gorm"
)

// MCPOAuthService 处理 MCP OAuth 相关的业务逻辑
type MCPOAuthService struct {
	metaStore *repositories.MetaStore
}

// NewMCPOAuthService 创建新的 MCP OAuth 服务
func NewMCPOAuthService(metaStore *repositories.MetaStore) *MCPOAuthService {
	return &MCPOAuthService{
		metaStore: metaStore,
	}
}

// SaveOAuthSession 保存 OAuth 会话信息
func (s *MCPOAuthService) SaveOAuthSession(session *types.MCPOAuthSession) error {
	// 设置会话过期时间（默认 10 分钟）
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(10 * time.Minute)
	}

	return s.metaStore.DB.Create(session).Error
}

// GetOAuthSession 根据 state 获取 OAuth 会话信息
func (s *MCPOAuthService) GetOAuthSession(state string) (*types.MCPOAuthSession, error) {
	var session types.MCPOAuthSession
	err := s.metaStore.DB.Where("state = ? AND expires_at > ?", state, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteOAuthSession 删除 OAuth 会话信息
func (s *MCPOAuthService) DeleteOAuthSession(state string) error {
	return s.metaStore.DB.Where("state = ?", state).Delete(&types.MCPOAuthSession{}).Error
}

// SaveOAuthToken 保存 OAuth Token 信息
func (s *MCPOAuthService) SaveOAuthToken(token *types.MCPOAuthToken) error {
	// 如果已存在相同的 UserDID + ClientID + Resource 组合，则更新
	var existingToken types.MCPOAuthToken
	err := s.metaStore.DB.Where("user_did = ? AND client_id = ? AND resource = ?",
		token.UserDID, token.ClientID, token.Resource).First(&existingToken).Error

	if err == nil {
		// 更新现有 token
		existingToken.AccessToken = token.AccessToken
		existingToken.TokenType = token.TokenType
		existingToken.RefreshToken = token.RefreshToken
		existingToken.ExpiresAt = token.ExpiresAt
		existingToken.Scope = token.Scope
		return s.metaStore.DB.Save(&existingToken).Error
	} else if err == gorm.ErrRecordNotFound {
		// 创建新 token
		return s.metaStore.DB.Create(token).Error
	} else {
		return err
	}
}

// GetOAuthToken 获取 OAuth Token 信息
func (s *MCPOAuthService) GetOAuthToken(userDID, clientID, resource string) (*types.MCPOAuthToken, error) {
	var token types.MCPOAuthToken
	err := s.metaStore.DB.Where("user_did = ? AND client_id = ? AND resource = ?",
		userDID, clientID, resource).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetValidOAuthToken 获取有效的 OAuth Token
func (s *MCPOAuthService) GetValidOAuthToken(userDID, clientID, resource string) (*types.MCPOAuthToken, error) {
	var token types.MCPOAuthToken
	err := s.metaStore.DB.Where("user_did = ? AND client_id = ? AND resource = ? AND expires_at > ?",
		userDID, clientID, resource, time.Now()).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateOAuthToken 更新 OAuth Token
func (s *MCPOAuthService) UpdateOAuthToken(token *types.MCPOAuthToken) error {
	return s.metaStore.DB.Save(token).Error
}

// DeleteOAuthToken 删除 OAuth Token
func (s *MCPOAuthService) DeleteOAuthToken(userDID, clientID, resource string) error {
	return s.metaStore.DB.Where("user_did = ? AND client_id = ? AND resource = ?",
		userDID, clientID, resource).Delete(&types.MCPOAuthToken{}).Error
}

// ListUserOAuthTokens 列出用户的所有 OAuth Token
func (s *MCPOAuthService) ListUserOAuthTokens(userDID string) ([]types.MCPOAuthToken, error) {
	var tokens []types.MCPOAuthToken
	err := s.metaStore.DB.Where("user_did = ?", userDID).Find(&tokens).Error
	return tokens, err
}

// SaveClientRegistration 保存客户端注册信息
func (s *MCPOAuthService) SaveClientRegistration(registration *types.MCPClientRegistration) error {
	// 如果已存在相同的 UserDID + AuthServerURL 组合，则更新
	var existingReg types.MCPClientRegistration
	err := s.metaStore.DB.Where("user_did = ? AND auth_server_url = ?",
		registration.UserDID, registration.AuthServerURL).First(&existingReg).Error

	if err == nil {
		// 更新现有注册信息
		existingReg.ClientID = registration.ClientID
		existingReg.ClientSecret = registration.ClientSecret
		existingReg.ClientName = registration.ClientName
		existingReg.RedirectURIs = registration.RedirectURIs
		existingReg.GrantTypes = registration.GrantTypes
		existingReg.ResponseTypes = registration.ResponseTypes
		existingReg.Scope = registration.Scope
		existingReg.RegistrationMetadata = registration.RegistrationMetadata
		return s.metaStore.DB.Save(&existingReg).Error
	} else if err == gorm.ErrRecordNotFound {
		// 创建新注册信息
		return s.metaStore.DB.Create(registration).Error
	} else {
		return err
	}
}

// GetClientRegistration 获取客户端注册信息
func (s *MCPOAuthService) GetClientRegistration(userDID, authServerURL string) (*types.MCPClientRegistration, error) {
	var registration types.MCPClientRegistration
	err := s.metaStore.DB.Where("user_did = ? AND auth_server_url = ?",
		userDID, authServerURL).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// GetClientRegistrationByClientID 根据 ClientID 获取客户端注册信息
func (s *MCPOAuthService) GetClientRegistrationByClientID(clientID string) (*types.MCPClientRegistration, error) {
	var registration types.MCPClientRegistration
	err := s.metaStore.DB.Where("client_id = ?", clientID).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// ListUserClientRegistrations 列出用户的所有客户端注册信息
func (s *MCPOAuthService) ListUserClientRegistrations(userDID string) ([]types.MCPClientRegistration, error) {
	var registrations []types.MCPClientRegistration
	err := s.metaStore.DB.Where("user_did = ?", userDID).Find(&registrations).Error
	return registrations, err
}

// DeleteClientRegistration 删除客户端注册信息
func (s *MCPOAuthService) DeleteClientRegistration(userDID, authServerURL string) error {
	return s.metaStore.DB.Where("user_did = ? AND auth_server_url = ?",
		userDID, authServerURL).Delete(&types.MCPClientRegistration{}).Error
}

// CleanupExpiredSessions 清理过期的会话
func (s *MCPOAuthService) CleanupExpiredSessions() error {
	return s.metaStore.DB.Where("expires_at < ?", time.Now()).Delete(&types.MCPOAuthSession{}).Error
}

// CleanupExpiredTokens 清理过期的 Token
func (s *MCPOAuthService) CleanupExpiredTokens() error {
	return s.metaStore.DB.Where("expires_at < ?", time.Now()).Delete(&types.MCPOAuthToken{}).Error
}

// 辅助方法：将字符串数组转换为 JSON
func (s *MCPOAuthService) StringArrayToJSON(arr []string) (string, error) {
	if arr == nil {
		return "", nil
	}
	data, err := json.Marshal(arr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 辅助方法：将 JSON 转换为字符串数组
func (s *MCPOAuthService) JSONToStringArray(jsonStr string) ([]string, error) {
	if jsonStr == "" {
		return nil, nil
	}
	var arr []string
	err := json.Unmarshal([]byte(jsonStr), &arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// CreateOAuthSessionFromParams 从参数创建 OAuth 会话
func (s *MCPOAuthService) CreateOAuthSessionFromParams(userDID, state, codeVerifier, authServerURL, clientID, clientSecret, redirectURI, resource string) *types.MCPOAuthSession {
	return &types.MCPOAuthSession{
		State:         state,
		CodeVerifier:  codeVerifier,
		AuthServerURL: authServerURL,
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		Resource:      resource,
		UserDID:       userDID,
		ExpiresAt:     time.Now().Add(10 * time.Minute),
	}
}

// CreateOAuthTokenFromResponse 从 Token 响应创建 OAuth Token
func (s *MCPOAuthService) CreateOAuthTokenFromResponse(userDID, clientID, resource string, accessToken, tokenType, refreshToken, scope string, expiresIn int) *types.MCPOAuthToken {
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	if expiresIn == 0 {
		// 如果没有指定过期时间，默认设置为 1 小时
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	return &types.MCPOAuthToken{
		UserDID:      userDID,
		ClientID:     clientID,
		Resource:     resource,
		AccessToken:  accessToken,
		TokenType:    tokenType,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Scope:        scope,
	}
}

// 验证 Token 是否需要刷新（在过期前 5 分钟刷新）
func (s *MCPOAuthService) ShouldRefreshToken(token *types.MCPOAuthToken) bool {
	if token.RefreshToken == "" {
		return false
	}
	// 在 Token 过期前 5 分钟进行刷新
	refreshThreshold := token.ExpiresAt.Add(-5 * time.Minute)
	return time.Now().After(refreshThreshold)
}

// GetTokenStats 获取 Token 统计信息
func (s *MCPOAuthService) GetTokenStats(userDID string) (map[string]interface{}, error) {
	var total int64
	var expired int64
	var valid int64

	// 总数
	if err := s.metaStore.DB.Model(&types.MCPOAuthToken{}).Where("user_did = ?", userDID).Count(&total).Error; err != nil {
		return nil, err
	}

	// 过期数
	if err := s.metaStore.DB.Model(&types.MCPOAuthToken{}).Where("user_did = ? AND expires_at < ?", userDID, time.Now()).Count(&expired).Error; err != nil {
		return nil, err
	}

	// 有效数
	if err := s.metaStore.DB.Model(&types.MCPOAuthToken{}).Where("user_did = ? AND expires_at > ?", userDID, time.Now()).Count(&valid).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":   total,
		"expired": expired,
		"valid":   valid,
	}, nil
}
