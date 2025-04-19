package database

import "time"

type OAuthAuthRequest struct {
	ID                  uint   `gorm:"primaryKey;autoIncrement:true"`
	State               string `gorm:"column:state"`
	AuthserverIss       string `gorm:"column:authserver_iss"`
	Did                 string `gorm:"column:did"`
	Handle              string `gorm:"column:handle"`
	PdsUrl              string `gorm:"column:pds_url"`
	PkceVerifier        string `gorm:"column:pkce_verifier"`
	Scope               string `gorm:"column:scope"`
	DpopAuthserverNonce string `gorm:"column:dpop_authserver_nonce"`
	DpopPrivateJwk      string `gorm:"column:dpop_private_jwk"`
	Platform            string `gorm:"column:platform"`
	ReturnURI           string `gorm:"column:return_uri"`
}

type OAuthSession struct {
	ID                  uint      `gorm:"primaryKey;autoIncrement:true"`
	Did                 string    `gorm:"column:did"`
	Handle              string    `gorm:"column:handle"`
	PdsUrl              string    `gorm:"column:pds_url"`
	AuthserverIss       string    `gorm:"column:authserver_iss"`
	AccessToken         string    `gorm:"column:access_token"`
	RefreshToken        string    `gorm:"column:refresh_token"`
	DpopAuthserverNonce string    `gorm:"column:dpop_authserver_nonce"`
	DpopPdsNonce        string    `gorm:"column:dpop_pds_nonce"`
	DpopPrivateJwk      string    `gorm:"column:dpop_private_jwk"`
	ExpiresIn           int64     `gorm:"column:expires_in"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	Platform            string    `gorm:"column:platform"`
	ReturnURI           string    `gorm:"column:return_uri"`
}

type OAuthCode struct {
	ID             uint      `gorm:"primaryKey;autoIncrement:true"`
	Code           string    `gorm:"column:code;uniqueIndex"` // 授权码
	OAuthSessionID uint      `gorm:"column:oauth_session_id"` // 关联的OAuth会话ID
	AvatarDid      string    `gorm:"column:avatar_did"`
	Platform       string    `gorm:"column:platform"`
	ReturnURI      string    `gorm:"column:return_uri"`
	Used           bool      `gorm:"column:used;default:false"` // 是否已使用
	ExpiresAt      time.Time `gorm:"column:expires_at"`         // 过期时间
	CreatedAt      time.Time `gorm:"column:created_at"`
}

type Session struct {
	ID             string    `gorm:"primaryKey"`
	AvatarDid      string    `gorm:"column:avatar_did"`
	AccessToken    string    `gorm:"column:access_token"`
	RefreshToken   string    `gorm:"column:refresh_token"`
	OAuthSessionID uint      `gorm:"column:oauth_session_id"` // 如果是 oauth 登录, 则需要关联 oauth_session_id
	Platform       string    `gorm:"column:platform"`
	ExpiredAt      time.Time `gorm:"column:expired_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

// 具备以下能力:
// 1. bsky chat
// 2. 提供 mcp server
// 3. 提供 response api
type Avatar struct { //  真实的人, 人创建的数字化身, 自注册的 Agent
	ID          uint      `gorm:"primaryKey;autoIncrement:true"`
	Did         string    `gorm:"column:did"`
	Handle      string    `gorm:"column:handle"`
	PdsUrl      string    `gorm:"column:pds_url"`
	IsAster     bool      `gorm:"column:is_aster"`
	CreatorDid  string    `gorm:"column:creator_did"`
	LastLoginAt time.Time `gorm:"column:last_login_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

type AvatarIntegrate struct { // 这个主要是真实用户关联的第三方平台账号, 作为数据来源同步
	ID         uint      `gorm:"primaryKey;autoIncrement:true"`
	AvatarDid  string    `gorm:"column:avatar_did"`
	Provider   string    `gorm:"column:provider"`
	ProviderID string    `gorm:"column:provider_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

type AvatarMCPServer struct {
	ID        uint      `gorm:"primaryKey;autoIncrement:true"`
	AvatarDid string    `gorm:"column:avatar_did"`
	ServerURL string    `gorm:"column:server_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type AvatarBsky struct { // 主要是支持
	ID         uint      `gorm:"primaryKey;autoIncrement:true"`
	AvatarDid  string    `gorm:"column:avatar_did"`
	BskyDid    string    `gorm:"column:bsky_did"`
	BskyHandle string    `gorm:"column:bsky_handle"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

type AvatarResponseAPI struct {
	ID             uint      `gorm:"primaryKey;autoIncrement:true"`
	AvatarDid      string    `gorm:"column:avatar_did"`
	APIEndpointURL string    `gorm:"column:api_endpoint_url"`
	APIKey         string    `gorm:"column:api_key"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}
