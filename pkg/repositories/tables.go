package repositories

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
