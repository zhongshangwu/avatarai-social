package types

import (
	"time"

	"gorm.io/gorm"
)

// MCPOAuthSession 存储 OAuth 会话信息
type MCPOAuthSession struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	State         string         `json:"state" gorm:"uniqueIndex;not null"`
	CodeVerifier  string         `json:"code_verifier" gorm:"not null"`
	AuthServerURL string         `json:"auth_server_url" gorm:"not null"`
	ClientID      string         `json:"client_id" gorm:"not null"`
	ClientSecret  string         `json:"client_secret"`
	RedirectURI   string         `json:"redirect_uri" gorm:"not null"`
	Resource      string         `json:"resource" gorm:"not null"`
	UserDID       string         `json:"user_did" gorm:"index"`
	ExpiresAt     time.Time      `json:"expires_at" gorm:"index"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// MCPOAuthToken 存储 OAuth Token 信息
type MCPOAuthToken struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserDID      string         `json:"user_did" gorm:"index;not null"`
	ClientID     string         `json:"client_id" gorm:"index;not null"`
	Resource     string         `json:"resource" gorm:"not null"`
	AccessToken  string         `json:"access_token" gorm:"not null"`
	TokenType    string         `json:"token_type" gorm:"default:Bearer"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresAt    time.Time      `json:"expires_at" gorm:"index"`
	Scope        string         `json:"scope"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// MCPClientRegistration 存储客户端注册信息
type MCPClientRegistration struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	UserDID              string         `json:"user_did" gorm:"index;not null"`
	AuthServerURL        string         `json:"auth_server_url" gorm:"not null"`
	ClientID             string         `json:"client_id" gorm:"uniqueIndex;not null"`
	ClientSecret         string         `json:"client_secret"`
	ClientName           string         `json:"client_name" gorm:"not null"`
	RedirectURIs         string         `json:"redirect_uris" gorm:"type:text"`  // JSON 数组存储
	GrantTypes           string         `json:"grant_types" gorm:"type:text"`    // JSON 数组存储
	ResponseTypes        string         `json:"response_types" gorm:"type:text"` // JSON 数组存储
	Scope                string         `json:"scope"`
	RegistrationMetadata string         `json:"registration_metadata" gorm:"type:text"` // JSON 存储完整的注册响应
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName 指定表名
func (MCPOAuthSession) TableName() string {
	return "mcp_oauth_sessions"
}

func (MCPOAuthToken) TableName() string {
	return "mcp_oauth_tokens"
}

func (MCPClientRegistration) TableName() string {
	return "mcp_client_registrations"
}

// IsExpired 检查会话是否过期
func (s *MCPOAuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsExpired 检查 Token 是否过期
func (t *MCPOAuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid 检查 Token 是否有效（未过期且未删除）
func (t *MCPOAuthToken) IsValid() bool {
	return !t.IsExpired() && t.DeletedAt.Time.IsZero()
}
