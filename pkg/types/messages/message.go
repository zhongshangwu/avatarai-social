package messages

type MessageType int

const (
	MessageTypeUnspecified MessageType = 0
	MessageTypeText        MessageType = 1
	MessageTypePost        MessageType = 2
	MessageTypeImage       MessageType = 3
	MessageTypeFile        MessageType = 4
	MessageTypeAudio       MessageType = 5
	MessageTypeVideo       MessageType = 6
	MessageTypeSticker     MessageType = 7
	MessageTypeCard        MessageType = 8
	MessageTypeAIChat      MessageType = 9
	MessageTypeSystem      MessageType = 10
	MessageTypeDelete      MessageType = 11
	MessageTypeRTC         MessageType = 12
)

type Message struct {
	ID         string      `json:"id"`          // 消息ID
	RoomID     string      `json:"room_id"`     // 房间ID
	RootID     string      `json:"root_id"`     // 根消息ID ( 回复消息时，根消息ID为被回复的消息ID )
	ParentID   string      `json:"parent_id"`   // 父消息ID ( 回复消息时，父消息ID为被回复的消息ID )
	MsgType    MessageType `json:"msg_type"`    // 消息类型
	Content    string      `json:"content"`     // 消息内容, JSON 序列化后的内容
	SenderID   string      `json:"sender_id"`   // 发送者ID
	ThreadID   string      `json:"thread_id"`   // 话题ID
	QuoteID    string      `json:"quote_id"`    // 引用消息ID
	SenderAt   int64       `json:"sender_at"`   // 发送时间
	CreateTime int64       `json:"create_time"` // 创建时间
	UpdateTime int64       `json:"update_time"` // 更新时间
	Deleted    bool        `json:"deleted"`     // 是否被撤回
}

type ThreadType string

const (
	THREAD_CONTINUATION ThreadType = "continuous" // 连续上下文
	THREAD_ISOLATED     ThreadType = "isolated"   // 隔离上下文
)

type Thread struct {
	ID       string     `json:"id"`        // 话题ID
	Title    string     `json:"title"`     // 话题标题
	Type     ThreadType `json:"type"`      // 话题类型: 连续上下文/独立上下文
	CreateAt int64      `json:"create_at"` // 创建时间
	UpdateAt int64      `json:"update_at"` // 更新时间
	Deleted  bool       `json:"deleted"`   // 是否被删除
}

type Room struct {
	RoomID        string `json:"room_id"`         // 房间号
	Title         string `json:"title"`           // 房间标题 (群组可以命名名称)
	Type          string `json:"type"`            // 房间类型 // chat, group
	LastMessageID string `json:"last_message_id"` // 最后一条消息ID
	CreatedTime   int64  `json:"create_time"`     // 创建时间
	UpdateTime    int64  `json:"update_time"`     // 更新时间
	Deleted       bool   `json:"deleted"`         // 是否被删除
}

type UserRoomStatus struct { // 归属某个具体的用户
	ID          string `json:"id"`           // 唯一标识 (业务上使用 room_id + userid 来唯一标识)
	RoomID      string `json:"room_id"`      // 房间号， 对话等都使用该 room id
	UserID      string `json:"user_id"`      // 用户ID
	Status      string `json:"status"`       // 状态 // request, accepted
	UnreadCount int32  `json:"unread_count"` // 未读消息数
	Muted       bool   `json:"muted"`        // 是否被静音
	CreatedTime int64  `json:"create_time"`  // 创建时间
	UpdateTime  int64  `json:"update_time"`  // 更新时间
	Deleted     bool   `json:"deleted"`      // 是否被删除
}

type AIChatMessage struct {
	ID              string `json:"id"`               // 消息ID
	MessageID       string `json:"message_id"`       // 消息ID, 引用 message.id
	ConversationID  string `json:"conversation_id"`  // 会话ID
	Role            string `json:"role"`             // 消息角色: user, assistant, system
	Content         string `json:"content"`          // 消息内容(纯文本)
	InterruptType   int32  `json:"interrupt_type"`   // 消息是否有被中断,0:默认值,未被中断 1:用户手动停止生成
	Status          int32  `json:"status"`           // 消息状态: 当前消息是否已结束,0:消息正在生成中, 1:消息被中断结束 2:err导致消息结束 3:触发安全审核导致消息结束 4:消息正常结束
	Error           string `json:"error"`            // 错误信息
	MessageMetadata string `json:"message_metadata"` // 消息元数据, 包含 Usage 等
	UserID          string `json:"user_id"`          // 用户ID (可能是用户ID， 也可以是 AssistantID)
	CreatedTime     int64  `json:"created_time"`     // 创建时间
	UpdatedTime     int64  `json:"updated_time"`     // 更新时间
}
