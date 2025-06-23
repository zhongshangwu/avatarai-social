package repositories

import (
	"time"
)

// 认证方法常量
const (
	AuthMethodNone      = "none"       // 无认证
	AuthMethodOAuth2    = "oauth2"     // OAuth2认证
	AuthMethodAPIKey    = "api_key"    // API Key认证
	AuthMethodBasicAuth = "basic_auth" // HTTP Basic认证
)

// 认证状态常量
const (
	AuthStatusActive   = "active"   // 激活状态
	AuthStatusExpired  = "expired"  // 已过期
	AuthStatusDisabled = "disabled" // 已禁用
	AuthStatusRevoked  = "revoked"  // 已撤销
)

// Token类型常量
const (
	TokenTypeBearer = "bearer"  // Bearer token
	TokenTypeBasic  = "basic"   // Basic token
	TokenTypeAPIKey = "api_key" // API Key
)

// PKCE Challenge方法常量
const (
	PKCEMethodPlain = "plain" // plain方法
	PKCEMethodS256  = "S256"  // SHA256方法（推荐）
)

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

func (OAuthAuthRequest) TableName() string {
	return "oauth_auth_requests"
}

type OAuthSession struct {
	ID                  uint   `gorm:"primaryKey;autoIncrement:true"`
	Did                 string `gorm:"column:did"`
	Handle              string `gorm:"column:handle"`
	PdsUrl              string `gorm:"column:pds_url"`
	AuthserverIss       string `gorm:"column:authserver_iss"`
	AccessToken         string `gorm:"column:access_token"`
	RefreshToken        string `gorm:"column:refresh_token"`
	DpopAuthserverNonce string `gorm:"column:dpop_authserver_nonce"`
	DpopPdsNonce        string `gorm:"column:dpop_pds_nonce"`
	DpopPrivateJwk      string `gorm:"column:dpop_private_jwk"`
	ExpiresIn           int64  `gorm:"column:expires_in"`
	CreatedAt           int64  `gorm:"column:created_at"`
	Platform            string `gorm:"column:platform"`
	ReturnURI           string `gorm:"column:return_uri"`
}

func (OAuthSession) TableName() string {
	return "oauth_session"
}

type OAuthCode struct {
	ID             uint   `gorm:"primaryKey;autoIncrement:true"`
	Code           string `gorm:"column:code;uniqueIndex"` // 授权码
	OAuthSessionID uint   `gorm:"column:oauth_session_id"` // 关联的OAuth会话ID
	UserDid        string `gorm:"column:user_did"`
	Platform       string `gorm:"column:platform"`
	ReturnURI      string `gorm:"column:return_uri"`
	Used           bool   `gorm:"column:used;default:false"` // 是否已使用
	ExpiresAt      int64  `gorm:"column:expires_at"`         // 过期时间
	CreatedAt      int64  `gorm:"column:created_at"`
}

func (OAuthCode) TableName() string {
	return "oauth_codes"
}

type Session struct {
	ID             string `gorm:"primaryKey"`
	UserDid        string `gorm:"column:user_did"`
	AccessToken    string `gorm:"column:access_token"`
	RefreshToken   string `gorm:"column:refresh_token"`
	OAuthSessionID uint   `gorm:"column:oauth_session_id"` // 如果是 oauth 登录, 则需要关联 oauth_session_id
	Platform       string `gorm:"column:platform"`
	ExpiredAt      int64  `gorm:"column:expired_at"`
	CreatedAt      int64  `gorm:"column:created_at"`
}

func (Session) TableName() string {
	return "session"
}

type Avatar struct { //  真实的人, 人创建的数字化身, 自注册的 Agent
	ID          uint   `gorm:"primaryKey;autoIncrement:true"`
	Did         string `gorm:"column:did"`
	Handle      string `gorm:"column:handle"`
	PdsUrl      string `gorm:"column:pds_url"`
	DisplayName string `gorm:"column:display_name"`
	AvatarCID   string `gorm:"column:avatar_cid"`
	Description string `gorm:"column:description"`
	IsAster     bool   `gorm:"column:is_aster"`
	Creator     string `gorm:"column:creator"`
	LastLoginAt int64  `gorm:"column:last_login_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	CreatedAt   int64  `gorm:"column:created_at"`
}

func (Avatar) TableName() string {
	return "avatar"
}

type Moment struct {
	ID            string      `gorm:"primaryKey"`
	URI           string      `gorm:"column:uri"`
	CID           string      `gorm:"column:cid;not null"`
	Text          string      `gorm:"column:text;not null"`
	Facets        string      `gorm:"column:facets"`
	ReplyRootID   string      `gorm:"column:reply_root_id"`
	ReplyParentID string      `gorm:"column:reply_parent_id"`
	Langs         StringArray `gorm:"type:jsonb;column:langs"`
	Tags          StringArray `gorm:"type:jsonb;column:tags"`
	CreatedAt     int64       `gorm:"column:created_at;not null"`
	UpdatedAt     int64       `gorm:"column:updated_at;not null"`
	IndexedAt     int64       `gorm:"column:indexed_at;not null"`
	Creator       string      `gorm:"column:creator;not null"`
	Deleted       bool        `gorm:"column:deleted"`
}

func (Moment) TableName() string {
	return "moments"
}

type MomentVideo struct {
	ID       int64  `gorm:"primaryKey;autoIncrement:true"`
	MomentID string `gorm:"column:moment_id"`
	VideoCID string `gorm:"column:video_cid"`
	Alt      string `gorm:"column:alt"`
}

func (MomentVideo) TableName() string {
	return "moment_videos"
}

type MomentImage struct {
	ID       int64  `gorm:"primaryKey;autoIncrement:true"`
	MomentID string `gorm:"column:moment_id"`
	Position int    `gorm:"column:position"`
	ImageCID string `gorm:"column:image_cid"`
	Alt      string `gorm:"column:alt"`
}

func (MomentImage) TableName() string {
	return "moment_images"
}

type MomentExternal struct {
	ID          int64  `gorm:"primaryKey;autoIncrement:true"`
	MomentID    string `gorm:"column:moment_id"`
	URI         string `gorm:"column:uri"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	ThumbCID    string `gorm:"column:thumb_cid"`
}

func (MomentExternal) TableName() string {
	return "moment_external"
}

type Like struct {
	ID         string `gorm:"primaryKey"`
	URI        string `gorm:"column:uri"`
	CID        string `gorm:"column:cid"`
	Creator    string `gorm:"column:creator"`
	SubjectURI string `gorm:"column:subject_uri"`
	SubjectCid string `gorm:"column:subject_cid"`
	CreatedAt  int64  `gorm:"column:created_at"`
	IndexedAt  int64  `gorm:"column:indexed_at"`
}

func (Like) TableName() string {
	return "likes"
}

type MomentAgg struct {
	URI        string `gorm:"column:uri"`
	LikeCount  int    `gorm:"column:like_count"`
	ReplyCount int    `gorm:"column:reply_count"`
}

func (MomentAgg) TableName() string {
	return "moment_agg"
}

type Tag struct {
	ID        string `gorm:"primaryKey"`
	URI       string `gorm:"column:uri"`
	CID       string `gorm:"column:cid"`
	Tag       string `gorm:"column:tag"` // 索引键 rkey
	CreatedAt int64  `gorm:"column:created_at"`
	Creator   string `gorm:"column:creator"`
	Deleted   bool   `gorm:"column:deleted"`
}

func (Tag) TableName() string {
	return "tags"
}

type ActivityTag struct {
	ID         string `gorm:"primaryKey"`
	SubjectURI string `gorm:"column:subject_uri"` // 关联的实体URI(目前只有moment)
	Tag        string `gorm:"column:tag"`
	CreatedAt  int64  `gorm:"column:created_at"`
	Creator    string `gorm:"column:creator"`
	Deleted    bool   `gorm:"column:deleted"`
}

func (ActivityTag) TableName() string {
	return "activity_tags"
}

type Topic struct {
	ID        string `gorm:"primaryKey"`
	URI       string `gorm:"column:uri"`
	CID       string `gorm:"column:cid"`
	Topic     string `gorm:"column:topic"` // 索引键 rkey
	CreatedAt int64  `gorm:"column:created_at"`
	Creator   string `gorm:"column:creator"`
	Deleted   bool   `gorm:"column:deleted"`
}

func (Topic) TableName() string {
	return "topics"
}

type ActivityTopic struct {
	ID         string `gorm:"primaryKey"`
	SubjectURI string `gorm:"column:subject_uri"` // 关联的实体URI(目前只有moment)
	Topic      string `gorm:"column:topic"`
	CreatedAt  int64  `gorm:"column:created_at"`
	Creator    string `gorm:"column:creator"`
	Deleted    bool   `gorm:"column:deleted"`
}

func (ActivityTopic) TableName() string {
	return "activity_topics"
}

type Message struct {
	ID         string `gorm:"primaryKey"`
	ExternalID string `gorm:"column:external_id"`
	RoomID     string `gorm:"column:room_id"`
	ThreadID   string `gorm:"column:thread_id"`
	MsgType    int    `gorm:"column:msg_type"`
	Content    string `gorm:"column:content"`
	QuoteMID   string `gorm:"column:quote_mid"`
	ReceiverID string `gorm:"column:receiver_id"`
	SenderID   string `gorm:"column:sender_id"`
	SenderAt   int64  `gorm:"column:sender_at"`
	CreatedAt  int64  `gorm:"column:created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at"`
	Deleted    bool   `gorm:"column:deleted"`
}

func (Message) TableName() string {
	return "messages"
}

type Room struct {
	ID        string `gorm:"primaryKey"`
	Title     string `gorm:"column:title"`
	Type      string `gorm:"column:type"`
	LastMID   string `gorm:"column:last_mid"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	Deleted   bool   `gorm:"column:deleted"`
}

func (Room) TableName() string {
	return "rooms"
}

type UserRoomStatus struct {
	ID          string `gorm:"primaryKey"`
	RoomID      string `gorm:"column:room_id"`
	UnreadCount int32  `gorm:"column:unread_count"`
	Muted       bool   `gorm:"column:muted"`
	UserID      string `gorm:"column:user_id"`
	Status      string `gorm:"column:status"` // request, accepted
	CreatedAt   int64  `gorm:"column:created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at"`
	Deleted     bool   `gorm:"column:deleted"`
}

func (UserRoomStatus) TableName() string {
	return "user_room_status"
}

type Thread struct {
	ID             string `gorm:"primaryKey"`
	RoomID         string `gorm:"column:room_id"`
	Title          string `gorm:"column:title"`
	ContextMode    string `gorm:"column:context_mode"` // continuous, isolated
	RootMID        string `gorm:"column:root_mid"`
	ParentThreadID string `gorm:"column:parent_thread_id"`
	CreatedAt      int64  `gorm:"column:created_at"`
	UpdatedAt      int64  `gorm:"column:updated_at"`
	Deleted        bool   `gorm:"column:deleted"`
}

func (Thread) TableName() string {
	return "threads"
}

type AgentMessage struct {
	ID                string `gorm:"primaryKey"`
	MessageID         string `gorm:"column:message_id"`
	Role              string `gorm:"column:role"`
	AltText           string `gorm:"column:alt_text"`
	InterruptType     int32  `gorm:"column:interrupt_type"`
	Status            string `gorm:"column:status"`
	Error             string `gorm:"column:error"`
	Usage             string `gorm:"column:usage"`
	Metadata          string `gorm:"column:metadata"`
	IncompleteDetails string `gorm:"column:incomplete_details"`
	Creator           string `gorm:"column:creator"`
	CreatedAt         int64  `gorm:"column:created_at"`
	UpdatedAt         int64  `gorm:"column:updated_at"`
	Deleted           bool   `gorm:"column:deleted"`
}

func (AgentMessage) TableName() string {
	return "agent_messages"
}

type AgentMessageItem struct {
	ID             string `gorm:"primaryKey"`
	AgentMessageID string `gorm:"column:agent_message_id"`
	ItemType       string `gorm:"column:item_type"`
	Item           string `gorm:"column:item"`
	Position       int    `gorm:"column:position"`
	CreatedAt      int64  `gorm:"column:created_at"`
	UpdatedAt      int64  `gorm:"column:updated_at"`
	Deleted        bool   `gorm:"column:deleted"`
}

func (AgentMessageItem) TableName() string {
	return "agent_message_items"
}

type UploadFile struct {
	ID        string `gorm:"primaryKey"`
	CID       string `gorm:"column:cid"`
	URI       string `gorm:"column:uri"`
	BlobCID   string `gorm:"column:blob_cid"`
	Size      int64  `gorm:"column:size"`
	Filename  string `gorm:"column:filename"`
	Extension string `gorm:"column:extension"`
	MimeType  string `gorm:"column:mime_type"`
	CreatedBy string `gorm:"column:created_by"`
	CreatedAt int64  `gorm:"column:created_at"`
}

func (UploadFile) TableName() string {
	return "upload_files"
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

type AvatarIntegrate struct { // 这个主要是真实用户关联的第三方平台账号, 作为数据来源同步
	ID         uint      `gorm:"primaryKey;autoIncrement:true"`
	AvatarDid  string    `gorm:"column:avatar_did"`
	Provider   string    `gorm:"column:provider"`
	ProviderID string    `gorm:"column:provider_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

type AtpRecord struct {
	URI       string `gorm:"column:uri"`
	CID       string `gorm:"column:cid"`
	Did       string `gorm:"column:did"`
	JSON      string `gorm:"column:json"`
	IndexedAt string `gorm:"column:indexed_at"`
}

func (AtpRecord) TableName() string {
	return "atp_records"
}

type MCPServer struct {
	ID                  uint   `gorm:"primaryKey;autoIncrement:true"`
	McpID               string `gorm:"column:mcp_id;uniqueIndex:idx_user_mcp"`   // MCP服务器唯一标识
	UserDid             string `gorm:"column:user_did;uniqueIndex:idx_user_mcp"` // 用户DID，组合唯一索引
	Name                string `gorm:"column:name;not null"`                     // 服务器名称
	Description         string `gorm:"column:description"`                       // 服务器描述
	About               string `gorm:"column:about"`                             // 关于信息
	Icon                string `gorm:"column:icon"`                              // 图标URL
	Author              string `gorm:"column:author"`                            // 作者
	Version             string `gorm:"column:version"`                           // 服务器版本
	ProtocolVersion     string `gorm:"column:protocol_version"`                  // MCP协议版本
	Instructions        string `gorm:"column:instructions"`                      // 使用说明
	Capabilities        string `gorm:"column:capabilities"`                      // 服务器能力
	Enabled             bool   `gorm:"column:enabled;default:false"`             // 是否开启
	SyncResources       bool   `gorm:"column:sync_resources;default:false"`      // 是否同步资源到PDS
	UpdatedAt           int64  `gorm:"column:updated_at;not null"`
	CreatedAt           int64  `gorm:"column:created_at;not null"`
	LastSyncResourcesAt int64  `gorm:"column:last_sync_resources_at"`
}

func (MCPServer) TableName() string {
	return "mcp_servers"
}

type MCPServerEndpoint struct {
	ID        uint   `gorm:"primaryKey;autoIncrement:true"`
	McpID     string `gorm:"column:mcp_id;not null"`   // 关联MCPServer
	Type      string `gorm:"column:type;not null"`     // stdio, sse, streamableHttp
	Command   string `gorm:"column:command"`           // 命令（stdio类型）
	Args      string `gorm:"type:text;column:args"`    // 参数列表JSON字符串
	Env       string `gorm:"type:text;column:env"`     // 环境变量JSON字符串
	URL       string `gorm:"column:url"`               // URL（HTTP类型）
	Headers   string `gorm:"type:text;column:headers"` // HTTP头JSON字符串
	CreatedAt int64  `gorm:"column:created_at;not null"`
	UpdatedAt int64  `gorm:"column:updated_at;not null"`
}

func (MCPServerEndpoint) TableName() string {
	return "mcp_server_endpoints"
}

type MCPServerOAuthCode struct {
	ID              uint   `gorm:"primaryKey;autoIncrement:true"`
	McpID           string `gorm:"column:mcp_id;not null;index"`                  // 关联MCPServer，统一使用mcp_server_id
	UserDid         string `gorm:"column:user_did;not null;index"`                // 关联用户DID
	Issuer          string `gorm:"column:issuer;not null;index"`                  // OAuth2 issuer
	State           string `gorm:"column:state;not null;uniqueIndex"`             // OAuth2 state参数
	CodeVerifier    string `gorm:"column:code_verifier;not null"`                 // PKCE code verifier
	CodeChallenge   string `gorm:"column:code_challenge;not null"`                // PKCE code challenge
	ChallengeMethod string `gorm:"column:challenge_method;not null;default:S256"` // code challenge方法
	RedirectURI     string `gorm:"column:redirect_uri;not null"`                  // OAuth2重定向URI
	Scope           string `gorm:"column:scope"`                                  // OAuth2 scope
	ExpiresAt       int64  `gorm:"column:expires_at;not null"`                    // 过期时间（统一使用expires_at）
	CreatedAt       int64  `gorm:"column:created_at;not null"`
	UpdatedAt       int64  `gorm:"column:updated_at;not null"`
}

func (MCPServerOAuthCode) TableName() string {
	return "mcp_server_oauth_codes"
}

type MCPServerAuth struct {
	ID          uint   `gorm:"primaryKey;autoIncrement:true"`
	McpId       string `gorm:"column:mcp_id;not null;index:idx_mcp_server_user"`   // 关联MCPServer，统一命名
	UserDid     string `gorm:"column:user_did;not null;index:idx_mcp_server_user"` // 关联用户DID
	AuthMethod  string `gorm:"column:auth_method;not null;index"`                  // 认证方式: none, oauth2, api_key, basic_auth
	AuthConfig  string `gorm:"type:text;column:auth_config"`                       // 认证配置JSON（非敏感信息）
	Credentials string `gorm:"type:text;column:credentials"`                       // 敏感凭据JSON（加密存储）
	Scope       string `gorm:"column:scope"`                                       // 授权范围
	Status      string `gorm:"column:status;default:active;index"`                 // 状态: active, expired, disabled
	ExpiresAt   int64  `gorm:"column:expires_at;index"`                            // 凭据过期时间
	LastUsedAt  int64  `gorm:"column:last_used_at"`                                // 最后使用时间
	CreatedAt   int64  `gorm:"column:created_at;not null"`
	UpdatedAt   int64  `gorm:"column:updated_at;not null"`
}

func (MCPServerAuth) TableName() string {
	return "mcp_server_auth"
}

func (auth *MCPServerAuth) IsExpired() bool {
	if auth.ExpiresAt == 0 {
		return false // 永不过期
	}
	return time.Now().Unix() > auth.ExpiresAt
}

func (auth *MCPServerAuth) IsActive() bool {
	return auth.Status == AuthStatusActive && !auth.IsExpired()
}

func (auth *MCPServerAuth) UpdateLastUsed() {
	auth.LastUsedAt = time.Now().Unix()
}

func (code *MCPServerOAuthCode) IsExpired() bool {
	return time.Now().Unix() > code.ExpiresAt
}

func (code *MCPServerOAuthCode) VerifyPKCE(verifier string) bool {
	if code.ChallengeMethod == PKCEMethodPlain {
		return code.CodeChallenge == verifier
	}
	// TODO: 实现SHA256验证逻辑
	return false
}

// Resource ,  Tool, Prompt 暂时不需要考虑
// // MCP服务器资源缓存
// type MCPServerResource struct {
// 	ID          uint   `gorm:"primaryKey;autoIncrement:true"`
// 	McpID       uint   `gorm:"column:mcp_id;not null"`       // 关联MCPServer
// 	URI         string `gorm:"column:uri;not null"`          // 资源URI
// 	Name        string `gorm:"column:name;not null"`         // 资源名称
// 	Description string `gorm:"column:description"`           // 资源描述
// 	MimeType    string `gorm:"column:mime_type"`             // MIME类型
// 	Annotations string `gorm:"type:text;column:annotations"` // 注解信息JSON字符串
// 	CachedAt    int64  `gorm:"column:cached_at;not null"`    // 缓存时间
// 	ExpireAt    int64  `gorm:"column:expire_at"`             // 缓存过期时间
// }

// func (MCPServerResource) TableName() string {
// 	return "mcp_server_resources"
// }

// // MCP服务器工具缓存
// type MCPServerTool struct {
// 	ID          uint   `gorm:"primaryKey;autoIncrement:true"`
// 	McpID       uint   `gorm:"column:mcp_id;not null"`        // 关联MCPServer
// 	Name        string `gorm:"column:name;not null"`          // 工具名称
// 	Description string `gorm:"column:description"`            // 工具描述
// 	InputSchema string `gorm:"type:text;column:input_schema"` // 输入参数schema JSON字符串
// 	CachedAt    int64  `gorm:"column:cached_at;not null"`     // 缓存时间
// 	ExpireAt    int64  `gorm:"column:expire_at"`              // 缓存过期时间
// }

// func (MCPServerTool) TableName() string {
// 	return "mcp_server_tools"
// }

// // MCP服务器状态日志（用于监控和调试）
// type MCPServerStatusLog struct {
// 	ID        uint   `gorm:"primaryKey;autoIncrement:true"`
// 	McpID     uint   `gorm:"column:mcp_id;not null"`    // 关联MCPServer
// 	Status    string `gorm:"column:status;not null"`    // connected, disconnected, connecting, error
// 	Error     string `gorm:"column:error"`              // 错误信息
// 	Metadata  string `gorm:"type:text;column:metadata"` // 额外元数据JSON字符串
// 	CreatedAt int64  `gorm:"column:created_at;not null"`
// }

// func (MCPServerStatusLog) TableName() string {
// 	return "mcp_server_status_logs"
// }

// // MCP服务器使用统计
// type MCPServerUsage struct {
// 	ID               uint  `gorm:"primaryKey;autoIncrement:true"`
// 	McpID            uint  `gorm:"column:mcp_id;not null"`              // 关联MCPServer
// 	Date             int64 `gorm:"column:date;not null"`                // 统计日期（YYYYMMDD格式的时间戳）
// 	ToolCallCount    int64 `gorm:"column:tool_call_count;default:0"`    // 工具调用次数
// 	ResourceGetCount int64 `gorm:"column:resource_get_count;default:0"` // 资源获取次数
// 	ErrorCount       int64 `gorm:"column:error_count;default:0"`        // 错误次数
// 	CreatedAt        int64 `gorm:"column:created_at;not null"`
// 	UpdatedAt        int64 `gorm:"column:updated_at;not null"`
// }

// func (MCPServerUsage) TableName() string {
// 	return "mcp_server_usage"
// }
