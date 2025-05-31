package messages

type Room struct {
	ID           string   `json:"id"`           // 房间号
	Title        string   `json:"title"`        // 房间标题 (群组可以命名名称)
	Type         string   `json:"type"`         // 房间类型 // 单聊、群聊、ai 对话...
	LastMID      string   `json:"last_mid"`     // 最后一条消息ID
	Participants []string `json:"participants"` // 参与者
	CreatedAt    int64    `json:"created_at"`   // 创建时间
	UpdatedAt    int64    `json:"updated_at"`   // 更新时间
	Deleted      bool     `json:"deleted"`      // 是否被删除
}

type UserRoomStatus struct { // 归属某个具体的用户, 基本是一个 room 的镜像数据
	ID           string   `json:"id"`           // 唯一标识 (业务上使用 room_id + userid 来唯一标识)
	RoomID       string   `json:"room_id"`      // 房间号， 对话等都使用该 room id
	Title        string   `json:"title"`        // 房间标题 (群组可以命名名称)
	Type         string   `json:"type"`         // 房间类型 // 单聊、群聊、ai 对话...
	LastMID      string   `json:"last_mid"`     // 最后一条消息ID
	Participants []string `json:"participants"` // 参与者
	UnreadCount  int32    `json:"unread_count"` // 未读消息数
	Muted        bool     `json:"muted"`        // 是否被静音
	UserID       string   `json:"user_id"`      // 用户ID
	Status       string   `json:"status"`       // 状态 // request, accepted
	CreatedAt    int64    `json:"created_at"`   // 创建时间
	UpdatedAt    int64    `json:"updated_at"`   // 更新时间
	Deleted      bool     `json:"deleted"`      // 是否被删除
}

type ThreadContextMode string

const (
	ThreadContextModeContinuous ThreadContextMode = "continuous" // 连续上下文
	ThreadContextModeIsolated   ThreadContextMode = "isolated"   // 隔离上下文
)

type Thread struct {
	ID             string            `json:"id"`               // 话题ID
	RoomID         string            `json:"room_id"`          // 房间ID
	Title          string            `json:"title"`            // 话题标题
	ContextMode    ThreadContextMode `json:"context_mode"`     // 话题上下文类型: 连续上下文/独立上下文
	RootMID        string            `json:"root_mid"`         // 根消息ID
	ParentThreadID string            `json:"parent_thread_id"` // 父话题ID
	CreatedAt      int64             `json:"created_at"`       // 创建时间
	UpdatedAt      int64             `json:"updated_at"`       // 更新时间
	Deleted        bool              `json:"deleted"`          // 是否被删除
}

type Message struct {
	ID       string `json:"id"`        // 消息ID
	RoomID   string `json:"room_id"`   // 房间ID
	ThreadID string `json:"thread_id"` // 话题ID

	// RootMID   string `json:"root_mid"`   // 消息回复关系：根消息ID为回复树的根节点消息ID
	// ParentMID string `json:"parent_mid"` // 消息回复关系：父消息ID为被回复的信息ID

	MsgType MessageType `json:"msg_type"` // 消息类型
	Content string      `json:"content"`  // 消息内容, JSON 序列化后的内容

	SenderID  string `json:"sender_id"`  // 发送者ID
	QuoteMID  string `json:"quote_mid"`  // 引用消息ID (消息回显用，作为上下文提供, 不作为消息组织结构)
	SenderAt  int64  `json:"sender_at"`  // 发送时间
	CreatedAt int64  `json:"created_at"` // 创建时间
	UpdatedAt int64  `json:"updated_at"` // 更新时间
	Deleted   bool   `json:"deleted"`    // 是否被撤回
}
