package types

type IMessage interface {
	GetMessageID() string
	GetMessageType() MessageType
}

type MessageType int

const (
	TEXT    MessageType = 1  // 文本
	POST    MessageType = 2  // 富文本
	IMAGE   MessageType = 3  // 图片
	FILE    MessageType = 4  // 文件
	AUDIO   MessageType = 5  // 语音
	VIDEO   MessageType = 6  // 视频
	STICKER MessageType = 7  // 表情
	CARD    MessageType = 8  // 卡片 ( content 内部可以再细分不同的卡片类型 )
	AI_CHAT MessageType = 9  // AI聊天 - 发送消息 ( OpenAI Response API 调用 )
	SYSTEM  MessageType = 10 // 系统消息
	DELETE  MessageType = 11 // 撤回消息
	RTC     MessageType = 12 // RTC 通话 ( 这个过程只能接收 音频和视频??? )
)

type Message struct {
	ID         string      `json:"id"`          // 消息ID
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
	CreatedAt       int64  `json:"created_at"`       // 创建时间
	UpdatedAt       int64  `json:"updated_at"`       // 更新时间
}

type IAIChatMessageItem interface {
	GetID() string
	GetType() string
	GetPosition() int
	GetMessageID() string
}

type AIChatMessageItem struct {
	ID        string `json:"id"`         // 消息ID
	MessageID string `json:"message_id"` // 消息ID
	Type      string `json:"type"`       // 消息类型: message, function_call, function_call_output
	Position  int32  `json:"position"`   // 消息位置
	Content   string `json:"content"`    // 消息内容, JSON 序列化后的内容
}

// ```
// message:

// {
// 	"type": "message",
// 	"id": "msg_67ccd3acc8d48190a77525dc6de64b4104becb25c45c1d41",
// 	"status": "completed",
// 	"role": "assistant",
// 	"content": [
// 	  {
// 		"type": "output_text",
// 		"text": "The image depicts a scenic landscape with a wooden boardwalk or pathway leading through lush, green grass under a blue sky with some clouds. The setting suggests a peaceful natural area, possibly a park or nature reserve. There are trees and shrubs in the background.",
// 		"annotations": []
// 	  }
// 	]
// }

// function_call:

// {
//     "type": "function_call",
//     "id": "fc_12345xyz",
//     "call_id": "call_12345xyz",
//     "name": "get_weather",
//     "arguments": "{\"location\":\"Paris, France\"}"
// }

// function_call_output:

// {                               # append result message
//     "type": "function_call_output",
//     "call_id": tool_call.call_id,
//     "output": str(result)
// }

// ```
