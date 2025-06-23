package repositories

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type MCPRepository struct {
	metaStore *MetaStore
}

func NewMCPRepository(metaStore *MetaStore) *MCPRepository {
	return &MCPRepository{metaStore: metaStore}
}

func (r *MCPRepository) GetMCPServersByUser(userDid string) ([]*MCPServer, error) {
	var servers []*MCPServer
	if err := r.metaStore.DB.Where("user_did = ?", userDid).Find(&servers).Error; err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *MCPRepository) GetMCPServerByIDAndUser(mcpID string, userDid string) (*MCPServer, error) {
	var server MCPServer
	if err := r.metaStore.DB.Where("mcp_id = ? AND user_did = ?", mcpID, userDid).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

func (r *MCPRepository) CreateMCPServer(server *MCPServer) error {
	now := time.Now().Unix()
	server.CreatedAt = now
	server.UpdatedAt = now
	return r.metaStore.DB.Create(server).Error
}

func (r *MCPRepository) DeleteMCPServer(mcpID string, userDid string) error {
	return r.metaStore.DB.Where("mcp_id = ? AND user_did = ?", mcpID, userDid).Delete(&MCPServer{}).Error
}

func (r *MCPRepository) GetMCPServerEndpoint(mcpID uint) (*MCPServerEndpoint, error) {
	var endpoint MCPServerEndpoint
	if err := r.metaStore.DB.Where("mcp_id = ?", mcpID).First(&endpoint).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &endpoint, nil
}

func (r *MCPRepository) CreateOrUpdateMCPServerEndpoint(endpoint *MCPServerEndpoint) error {
	now := time.Now().Unix()
	endpoint.UpdatedAt = now

	// 检查是否存在
	var existing MCPServerEndpoint
	if err := r.metaStore.DB.Where("mcp_id = ?", endpoint.McpID).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新的
			endpoint.CreatedAt = now
			return r.metaStore.DB.Create(endpoint).Error
		}
		return err
	}

	// 存在，更新
	endpoint.ID = existing.ID
	endpoint.CreatedAt = existing.CreatedAt
	return r.metaStore.DB.Save(endpoint).Error
}

func (r *MCPRepository) GetMCPServerAuth(mcpID uint) (*MCPServerAuth, error) {
	var auth MCPServerAuth
	if err := r.metaStore.DB.Where("mcp_id = ?", mcpID).First(&auth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &auth, nil
}

func (r *MCPRepository) CreateOrUpdateMCPServerAuth(auth *MCPServerAuth) error {
	now := time.Now().Unix()
	auth.UpdatedAt = now

	var existing MCPServerAuth
	if err := r.metaStore.DB.Where("mcp_id = ?", auth.McpId).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新的
			auth.CreatedAt = now
			return r.metaStore.DB.Create(auth).Error
		}
		return err
	}

	// 存在，更新
	auth.ID = existing.ID
	auth.CreatedAt = existing.CreatedAt
	return r.metaStore.DB.Save(auth).Error
}

func (r *MCPRepository) UpdateSyncResourcesStatus(mcpID string, userDid string, syncResources bool) error {
	updates := map[string]interface{}{
		"sync_resources": syncResources,
		"updated_at":     time.Now().Unix(),
	}
	return r.metaStore.DB.Model(&MCPServer{}).
		Where("mcp_id = ? AND user_did = ?", mcpID, userDid).
		Updates(updates).Error
}

func (r *MCPRepository) UpdateEnabledStatus(mcpID string, userDid string, enabled bool) error {
	updates := map[string]interface{}{
		"enabled":    enabled,
		"updated_at": time.Now().Unix(),
	}
	return r.metaStore.DB.Model(&MCPServer{}).
		Where("mcp_id = ? AND user_did = ?", mcpID, userDid).
		Updates(updates).Error
}

func (r *MCPRepository) CreateMCPServerOAuthCode(code *MCPServerOAuthCode) error {
	now := time.Now().Unix()
	code.CreatedAt = now
	code.UpdatedAt = now
	return r.metaStore.DB.Create(code).Error
}

func (r *MCPRepository) GetMCPServerOAuthCode(issuer string, state string) (*MCPServerOAuthCode, error) {
	var code MCPServerOAuthCode
	if err := r.metaStore.DB.Where("issuer = ? AND state = ?", issuer, state).First(&code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &code, nil
}
