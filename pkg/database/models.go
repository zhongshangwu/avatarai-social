package database

import (
	"time"
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
	return "oauth_auth_request"
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

func (OAuthSession) TableName() string {
	return "oauth_session"
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

func (OAuthCode) TableName() string {
	return "oauth_code"
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

func (Session) TableName() string {
	return "session"
}

type Avatar struct { //  真实的人, 人创建的数字化身, 自注册的 Agent
	ID          uint      `gorm:"primaryKey;autoIncrement:true"`
	Did         string    `gorm:"column:did"`
	Handle      string    `gorm:"column:handle"`
	PdsUrl      string    `gorm:"column:pds_url"`
	DisplayName string    `gorm:"column:display_name"`
	AvatarCID   string    `gorm:"column:avatar_cid"`
	Description string    `gorm:"column:description"`
	IsAster     bool      `gorm:"column:is_aster"`
	CreatorDid  string    `gorm:"column:creator_did"`
	LastLoginAt time.Time `gorm:"column:last_login_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (Avatar) TableName() string {
	return "avatar"
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

type Moment struct {
	URI            string      `gorm:"primaryKey;column:uri"`
	CID            string      `gorm:"column:cid;not null"`
	Creator        string      `gorm:"column:creator;not null"`
	Text           string      `gorm:"column:text;not null"`
	ReplyRoot      string      `gorm:"column:reply_root"`
	ReplyRootCID   string      `gorm:"column:reply_root_cid"`
	ReplyParent    string      `gorm:"column:reply_parent"`
	ReplyParentCID string      `gorm:"column:reply_parent_cid"`
	CreatedAt      string      `gorm:"column:created_at;not null"`
	IndexedAt      string      `gorm:"column:indexed_at;not null"`
	SortAt         string      `gorm:"column:sort_at;not null"`
	Langs          StringArray `gorm:"type:jsonb;column:langs"`
	Tags           StringArray `gorm:"type:jsonb;column:tags"`
}

func (Moment) TableName() string {
	return "moment"
}

type MomentVideo struct {
	MomentURI string `gorm:"column:moment_uri"`
	VideoCID  string `gorm:"column:video_cid"`
	Alt       string `gorm:"column:alt"`
}

func (MomentVideo) TableName() string {
	return "moment_video"
}

type MomentImage struct {
	MomentURI string `gorm:"column:moment_uri"`
	Position  int    `gorm:"column:position"`
	ImageCID  string `gorm:"column:image_cid"`
	Alt       string `gorm:"column:alt"`
}

func (MomentImage) TableName() string {
	return "moment_image"
}

type MomentExternal struct {
	MomentURI   string `gorm:"column:moment_uri"`
	URI         string `gorm:"column:uri"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	ThumbCID    string `gorm:"column:thumb_cid"`
}

func (MomentExternal) TableName() string {
	return "moment_external"
}

type AtpRecord struct {
	URI       string `gorm:"column:uri"`
	CID       string `gorm:"column:cid"`
	Did       string `gorm:"column:did"`
	JSON      string `gorm:"column:json"`
	IndexedAt string `gorm:"column:indexed_at"`
}

func (AtpRecord) TableName() string {
	return "atp_record"
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

type AIChatMessage struct {
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

func (AIChatMessage) TableName() string {
	return "ai_chat_messages"
}

type AIChatMessageItem struct {
	ID        string `gorm:"primaryKey"`
	MessageID string `gorm:"column:message_id"`
	ItemType  string `gorm:"column:item_type"`
	Item      string `gorm:"column:item"`
	Position  int    `gorm:"column:position"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
	Deleted   bool   `gorm:"column:deleted"`
}

func (AIChatMessageItem) TableName() string {
	return "ai_chat_message_items"
}
