// package types

// type IMessage interface {
// 	GetMessageID() string
// 	GetMessageType() MessageType
// }

// type MessageType int

// const (
// 	TEXT    MessageType = 1  // 文本
// 	POST    MessageType = 2  // 富文本
// 	IMAGE   MessageType = 3  // 图片
// 	FILE    MessageType = 4  // 文件
// 	AUDIO   MessageType = 5  // 语音
// 	VIDEO   MessageType = 6  // 视频
// 	STICKER MessageType = 7  // 表情
// 	CARD    MessageType = 8  // 卡片 ( content 内部可以再细分不同的卡片类型 )
// 	AI_CHAT MessageType = 9  // AI聊天 - 发送消息 ( OpenAI Response API 调用 )
// 	SYSTEM  MessageType = 10 // 系统消息
// 	DELETE  MessageType = 11 // 撤回消息
// 	RTC     MessageType = 12 // RTC 通话 ( 这个过程只能接收 音频和视频??? )
// )

// type Message struct {
// 	ID         string      `json:"id"`          // 消息ID
// 	RoomID     string      `json:"room_id"`     // 房间ID
// 	RootID     string      `json:"root_id"`     // 根消息ID ( 回复消息时，根消息ID为被回复的消息ID )
// 	ParentID   string      `json:"parent_id"`   // 父消息ID ( 回复消息时，父消息ID为被回复的消息ID )
// 	MsgType    MessageType `json:"msg_type"`    // 消息类型
// 	Content    string      `json:"content"`     // 消息内容, JSON 序列化后的内容
// 	SenderID   string      `json:"sender_id"`   // 发送者ID
// 	ThreadID   string      `json:"thread_id"`   // 话题ID
// 	QuoteID    string      `json:"quote_id"`    // 引用消息ID
// 	SenderAt   int64       `json:"sender_at"`   // 发送时间
// 	CreateTime int64       `json:"create_time"` // 创建时间
// 	UpdateTime int64       `json:"update_time"` // 更新时间
// 	Deleted    bool        `json:"deleted"`     // 是否被撤回
// }

// type ThreadType string

// const (
// 	THREAD_CONTINUATION ThreadType = "continuous" // 连续上下文
// 	THREAD_ISOLATED     ThreadType = "isolated"   // 隔离上下文
// )

// type Thread struct {
// 	ID       string     `json:"id"`        // 话题ID
// 	Title    string     `json:"title"`     // 话题标题
// 	Type     ThreadType `json:"type"`      // 话题类型: 连续上下文/独立上下文
// 	CreateAt int64      `json:"create_at"` // 创建时间
// 	UpdateAt int64      `json:"update_at"` // 更新时间
// 	Deleted  bool       `json:"deleted"`   // 是否被删除
// }

// type AIChatMessage struct {
// 	ID              string `json:"id"`               // 消息ID
// 	MessageID       string `json:"message_id"`       // 消息ID, 引用 message.id
// 	ConversationID  string `json:"conversation_id"`  // 会话ID
// 	Role            string `json:"role"`             // 消息角色: user, assistant, system
// 	Content         string `json:"content"`          // 消息内容(纯文本)
// 	InterruptType   int32  `json:"interrupt_type"`   // 消息是否有被中断,0:默认值,未被中断 1:用户手动停止生成
// 	Status          int32  `json:"status"`           // 消息状态: 当前消息是否已结束,0:消息正在生成中, 1:消息被中断结束 2:err导致消息结束 3:触发安全审核导致消息结束 4:消息正常结束
// 	Error           string `json:"error"`            // 错误信息
// 	MessageMetadata string `json:"message_metadata"` // 消息元数据, 包含 Usage 等
// 	UserID          string `json:"user_id"`          // 用户ID (可能是用户ID， 也可以是 AssistantID)
// 	CreatedTime     int64  `json:"created_time"`     // 创建时间
// 	UpdatedTime     int64  `json:"updated_time"`     // 更新时间
// }

// type IAIChatMessageItem interface {
// 	GetID() string
// 	GetType() string
// 	GetPosition() int
// 	GetMessageID() string
// }

// type AIChatMessageItem struct {
// 	ID        string `json:"id"`         // 消息ID
// 	MessageID string `json:"message_id"` // 消息ID
// 	Type      string `json:"type"`       // 消息类型: message, function_call, function_call_output
// 	Position  int32  `json:"position"`   // 消息位置
// 	Content   string `json:"content"`    // 消息内容, JSON 序列化后的内容
// }

// type Room struct {
// 	RoomID        string `json:"room_id"`         // 房间号
// 	Title         string `json:"title"`           // 房间标题 (群组可以命名名称)
// 	Type          string `json:"type"`            // 房间类型 // chat, group
// 	LastMessageID string `json:"last_message_id"` // 最后一条消息ID
// 	CreatedTime   int64  `json:"create_time"`     // 创建时间
// 	UpdateTime    int64  `json:"update_time"`     // 更新时间
// 	Deleted       bool   `json:"deleted"`         // 是否被删除
// }

// type UserRoomStatus struct { // 归属某个具体的用户
// 	ID          string `json:"id"`           // 唯一标识 (业务上使用 room_id + userid 来唯一标识)
// 	RoomID      string `json:"room_id"`      // 房间号， 对话等都使用该 room id
// 	UserID      string `json:"user_id"`      // 用户ID
// 	Status      string `json:"status"`       // 状态 // request, accepted
// 	UnreadCount int32  `json:"unread_count"` // 未读消息数
// 	Muted       bool   `json:"muted"`        // 是否被静音
// 	CreatedTime int64  `json:"create_time"`  // 创建时间
// 	UpdateTime  int64  `json:"update_time"`  // 更新时间
// 	Deleted     bool   `json:"deleted"`      // 是否被删除
// }

// type AIChatMessageBody struct {
// 	Query string `json:"query"`
// }

// // ```
// // message:

// // {
// // 	"type": "message",
// // 	"id": "msg_67ccd3acc8d48190a77525dc6de64b4104becb25c45c1d41",
// // 	"status": "completed",
// // 	"role": "assistant",
// // 	"content": [
// // 	  {
// // 		"type": "output_text",
// // 		"text": "The image depicts a scenic landscape with a wooden boardwalk or pathway leading through lush, green grass under a blue sky with some clouds. The setting suggests a peaceful natural area, possibly a park or nature reserve. There are trees and shrubs in the background.",
// // 		"annotations": []
// // 	  }
// // 	]
// // }

// // function_call:

// // {
// //     "type": "function_call",
// //     "id": "fc_12345xyz",
// //     "call_id": "call_12345xyz",
// //     "name": "get_weather",
// //     "arguments": "{\"location\":\"Paris, France\"}"
// // }

// // function_call_output:

// // {                               # append result message
// //     "type": "function_call_output",
// //     "call_id": tool_call.call_id,
// //     "output": str(result)
// // }

// // ```

// package chat

// import (
// 	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
// )

// type Event interface {
// 	GetEventID() string
// 	GetEventType() string
// }

// type SendMsgEvent struct {
// 	RoomID   string               `json:"room_id"`
// 	MsgType  messages.MessageType `json:"msg_type"`
// 	Payload  *string              `json:"payload"` // 普通消息内容，与AIChatBody互斥
// 	SenderID string               `json:"sender_id"`
// 	ThreadID string               `json:"thread_id"`
// 	QuoteID  string               `json:"quote_id"`
// 	SenderAt string               `json:"sender_at"`
// }

// type AIChatMessageItem interface {
// 	GetID() string
// 	GetType() string
// 	GetContent() string
// }

// type AIChatMessageItemMessage struct {
// 	ID      string `json:"id"`
// 	Type    string `json:"type"` // 常量: "message"
// 	Content string `json:"content"`
// }

// func (m *AIChatMessageItemMessage) GetID() string {
// 	return m.ID
// }

// func (m *AIChatMessageItemMessage) GetType() string {
// 	return m.Type
// }

// func (m *AIChatMessageItemMessage) GetContent() string {
// 	return m.Content
// }

// type AIChatMessageItemFunctionCall struct {
// 	ID      string `json:"id"`
// 	Type    string `json:"type"` // 常量: "function_call"
// 	Content string `json:"content"`
// }

// func (f *AIChatMessageItemFunctionCall) GetID() string {
// 	return f.ID
// }

// func (f *AIChatMessageItemFunctionCall) GetType() string {
// 	return f.Type
// }

// func (f *AIChatMessageItemFunctionCall) GetContent() string {
// 	return f.Content
// }

// type AIChatMessageBody struct {
// 	Role            string                 `json:"role"`
// 	Content         string                 `json:"content"`
// 	Status          int32                  `json:"status"`
// 	InterruptType   int32                  `json:"interrupt_type"`
// 	Error           string                 `json:"error"`
// 	MessageItems    []AIChatMessageItem    `json:"message_items"`    // 使用接口类型实现union
// 	MessageMetadata map[string]interface{} `json:"message_metadata"` // 消息元数据，包含Usage等
// }

// // 枚举类型的实现
// type CloseReason int32

// const (
// 	CloseReasonUnspecified       CloseReason = 0
// 	CloseReasonUserExit          CloseReason = 1
// 	CloseReasonSessionTimeout    CloseReason = 2
// 	CloseReasonServerClose       CloseReason = 3
// 	CloseReasonNetworkDisconnect CloseReason = 4
// 	CloseReasonUserRisk          CloseReason = 5
// 	CloseReasonBotRisk           CloseReason = 6
// 	CloseReasonIdleTimeout       CloseReason = 7
// 	CloseReasonSystemError       CloseReason = 99
// )

// // 获取枚举值的字符串表示
// func (c CloseReason) String() string {
// 	switch c {
// 	case CloseReasonUnspecified:
// 		return "CLOSE_REASON_UNSPECIFIED"
// 	case CloseReasonUserExit:
// 		return "CLOSE_REASON_USER_EXIT"
// 	case CloseReasonSessionTimeout:
// 		return "CLOSE_REASON_SESSION_TIMEOUT"
// 	case CloseReasonServerClose:
// 		return "CLOSE_REASON_SERVER_CLOSE"
// 	case CloseReasonNetworkDisconnect:
// 		return "CLOSE_REASON_NETWORK_DISCONNECT"
// 	case CloseReasonUserRisk:
// 		return "CLOSE_REASON_USER_RISK"
// 	case CloseReasonBotRisk:
// 		return "CLOSE_REASON_BOT_RISK"
// 	case CloseReasonIdleTimeout:
// 		return "CLOSE_REASON_IDLE_TIMEOUT"
// 	case CloseReasonSystemError:
// 		return "CLOSE_REASON_SYSTEM_ERROR"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// type BotStatus int32

// const (
// 	BotStatusUnspecified BotStatus = 0
// 	BotStatusSpeaking    BotStatus = 1
// 	BotStatusThinking    BotStatus = 2
// 	BotStatusListening   BotStatus = 3
// )

// func (b BotStatus) String() string {
// 	switch b {
// 	case BotStatusUnspecified:
// 		return "BOT_STATUS_UNSPECIFIED"
// 	case BotStatusSpeaking:
// 		return "BOT_STATUS_SPEAKING"
// 	case BotStatusThinking:
// 		return "BOT_STATUS_THINKING"
// 	case BotStatusListening:
// 		return "BOT_STATUS_LISTENING"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// type Role int32

// const (
// 	RoleUnspecified Role = 0
// 	RoleUser        Role = 1
// 	RoleBot         Role = 2
// )

// func (r Role) String() string {
// 	switch r {
// 	case RoleUnspecified:
// 		return "ROLE_UNSPECIFIED"
// 	case RoleUser:
// 		return "ROLE_USER"
// 	case RoleBot:
// 		return "ROLE_BOT"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// type MessageStatus int32

// const (
// 	MessageStatusUnspecified MessageStatus = 0
// 	MessageStatusCreated     MessageStatus = 1
// 	MessageStatusCompleted   MessageStatus = 2
// )

// func (m MessageStatus) String() string {
// 	switch m {
// 	case MessageStatusUnspecified:
// 		return "MESSAGE_STATUS_UNSPECIFIED"
// 	case MessageStatusCreated:
// 		return "MESSAGE_STATUS_CREATED"
// 	case MessageStatusCompleted:
// 		return "MESSAGE_STATUS_COMPLETED"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// type HistoryStrategy int32

// const (
// 	HistoryStrategyAuto     HistoryStrategy = 0
// 	HistoryStrategyDiscard  HistoryStrategy = 1
// 	HistoryStrategyPreserve HistoryStrategy = 2
// )

// func (h HistoryStrategy) String() string {
// 	switch h {
// 	case HistoryStrategyAuto:
// 		return "Auto"
// 	case HistoryStrategyDiscard:
// 		return "Discard"
// 	case HistoryStrategyPreserve:
// 		return "Preserve"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// type RTCNodeInfo struct {
// 	Host string `json:"host"`
// 	Port int32  `json:"port"`
// }

// type AVParams struct {
// 	UpstreamAudioSampleRate   int32 `json:"upstream_audio_sample_rate"`
// 	UpstreamAudioChannels     int32 `json:"upstream_audio_channels"`
// 	UpstreamImageInterval     int32 `json:"upstream_image_interval"`
// 	DownstreamAudioBitRate    int32 `json:"downstream_audio_bit_rate"`
// 	DownstreamAudioSampleRate int32 `json:"downstream_audio_sample_rate"`
// 	DownstreamAudioChannels   int32 `json:"downstream_audio_channels"`
// }

// type TextContent struct {
// 	Text string `json:"text"`
// }

// type MessageContent struct {
// 	Text *TextContent `json:"text,omitempty"`
// }

// type Message struct {
// 	Round   int32             `json:"round"`
// 	Content []*MessageContent `json:"content"`
// 	Role    Role              `json:"role"`
// }

// type OasisHeader struct {
// 	AppId          int32  `json:"app_id"`
// 	Did            string `json:"did"`
// 	WebId          string `json:"web_id"`
// 	Platform       string `json:"platform"`
// 	Buvid          int64  `json:"buvid"`
// 	ExtraDid       string `json:"extra_did"`
// 	OsName         string `json:"os_name"`
// 	OsVersion      string `json:"os_version"`
// 	Channel        string `json:"channel"`
// 	DeviceBrand    string `json:"device_brand"`
// 	DeviceModel    string `json:"device_model"`
// 	AppVersion     string `json:"app_version"`
// 	VersionCode    string `json:"version_code"`
// 	NetworkCarrier string `json:"network_carrier"`
// 	NetworkType    string `json:"network_type"`
// 	TimeZone       string `json:"time_zone"`
// }

// type TTSParams struct {
// 	VoiceId string `json:"voice_id"`
// }

// type ToolCallParams struct {
// 	Enabled bool `json:"enabled"`
// }

// type ReplayTurnParams struct {
// 	Enabled  bool   `json:"enabled"`
// 	BotText  string `json:"bot_text"`
// 	UserText string `json:"user_text"`
// }

// type ProfileParams struct {
// 	Mode     string `json:"mode"`
// 	Language string `json:"language"`
// }

// type BackdoorConfig struct {
// 	Agent           string `json:"agent"`
// 	AqtaEndpoint    string `json:"aqta_endpoint"`
// 	AqtaModel       string `json:"aqta_model"`
// 	AsrEndpoint     string `json:"asr_endpoint"`
// 	AsrModel        string `json:"asr_model"`
// 	AsrMiniEndpoint string `json:"asr_mini_endpoint"`
// 	AsrMiniModel    string `json:"asr_mini_model"`
// }

// type CreateSessionEvent struct {
// 	UserId           int64             `json:"user_id"`
// 	ChatId           int64             `json:"chat_id"`
// 	UserIdent        string            `json:"user_ident"`
// 	RoomId           string            `json:"room_id"`
// 	BotId            string            `json:"bot_id"`
// 	SessionId        string            `json:"session_id"`
// 	NodeInfo         *RTCNodeInfo      `json:"node_info,omitempty"`
// 	AvParams         *AVParams         `json:"av_params,omitempty"`
// 	OasisHeader      *OasisHeader      `json:"oasis_header,omitempty"`
// 	TtsParams        *TTSParams        `json:"tts_params,omitempty"`
// 	ToolCallParams   *ToolCallParams   `json:"tool_call_params,omitempty"`
// 	ReplayTurnParams *ReplayTurnParams `json:"replay_turn_params,omitempty"`
// 	ProfileParams    *ProfileParams    `json:"profile_params,omitempty"`
// 	BackdoorConfig   *BackdoorConfig   `json:"backdoor_config,omitempty"`
// }

// type CloseSessionEvent struct {
// 	Reason CloseReason `json:"reason"`
// }

// type InterruptEvent struct {
// 	// 无内容字段
// }

// type ChangeVoiceEvent struct {
// 	VoiceId     string `json:"voice_id"`
// 	ProfileMode string `json:"profile_mode"`
// }

// type ResumeEvent struct {
// 	// 无内容字段
// }

// type SwitchVideoStreamEvent struct {
// 	WithVideo bool            `json:"with_video"`
// 	Strategy  HistoryStrategy `json:"strategy"`
// }

// // 服务端事件子类型
// type SessionCreatedEvent struct {
// 	SessionId string `json:"session_id"`
// }

// type SessionClosedEvent struct {
// 	Reason CloseReason `json:"reason"`
// }

// type BotStatusEvent struct {
// 	Status BotStatus `json:"status"`
// 	Turn   int32     `json:"turn"`
// }

// type HistoryAppendEvent struct {
// 	Message *Message `json:"message"`
// }

// type MessageDeltaEvent struct {
// 	Turn   int32         `json:"turn"`
// 	Role   Role          `json:"role"`
// 	Delta  string        `json:"delta"`
// 	Status MessageStatus `json:"status"`
// }

// type VoiceChangedEvent struct {
// 	VoiceId string `json:"voice_id"`
// }

// type NotificationEvent struct {
// 	Content string `json:"content"`
// }

// type ResetTurnOutputEvent struct {
// 	Turn int32 `json:"turn"`
// }

// type ResetTurnEvent struct {
// 	Turn      int32  `json:"turn"`
// 	RiskLevel string `json:"risk_level"`
// }

// type HistoryResetEvent struct {
// 	// 无内容字段
// }

// type ClientEvent struct {
// 	EventID   string `json:"event_id"`
// 	EventType string `json:"event_type"`

// 	CreateSession     *CreateSessionEvent     `json:"create_session,omitempty"`
// 	CloseSession      *CloseSessionEvent      `json:"close_session,omitempty"`
// 	Interrupt         *InterruptEvent         `json:"interrupt,omitempty"`
// 	ChangeVoice       *ChangeVoiceEvent       `json:"change_voice,omitempty"`
// 	Resume            *ResumeEvent            `json:"resume,omitempty"`
// 	SwitchVideoStream *SwitchVideoStreamEvent `json:"switch_video_stream,omitempty"`
// }

// func (c *ClientEvent) GetEventID() string {
// 	return c.EventID
// }

// func (c *ClientEvent) GetEventType() string {
// 	return c.EventType
// }

// // 创建不同类型的客户端事件的工厂方法
// func NewCreateSessionClientEvent(eventID string, event *CreateSessionEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:       eventID,
// 		EventType:     "create_session",
// 		CreateSession: event,
// 	}
// }

// func NewCloseSessionClientEvent(eventID string, event *CloseSessionEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:      eventID,
// 		EventType:    "close_session",
// 		CloseSession: event,
// 	}
// }

// func NewInterruptClientEvent(eventID string, event *InterruptEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:   eventID,
// 		EventType: "interrupt",
// 		Interrupt: event,
// 	}
// }

// func NewChangeVoiceClientEvent(eventID string, event *ChangeVoiceEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:     eventID,
// 		EventType:   "change_voice",
// 		ChangeVoice: event,
// 	}
// }

// func NewResumeClientEvent(eventID string, event *ResumeEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:   eventID,
// 		EventType: "resume",
// 		Resume:    event,
// 	}
// }

// func NewSwitchVideoStreamClientEvent(eventID string, event *SwitchVideoStreamEvent) *ClientEvent {
// 	return &ClientEvent{
// 		EventID:           eventID,
// 		EventType:         "switch_video_stream",
// 		SwitchVideoStream: event,
// 	}
// }

// type ServerEvent struct {
// 	EventID   string `json:"event_id"`
// 	EventType string `json:"event_type"`

// 	// 各种事件类型，只有一个会被设置
// 	SessionCreated  *SessionCreatedEvent  `json:"session_created,omitempty"`
// 	SessionClosed   *SessionClosedEvent   `json:"session_closed,omitempty"`
// 	BotStatus       *BotStatusEvent       `json:"bot_status,omitempty"`
// 	HistoryAppend   *HistoryAppendEvent   `json:"history_append,omitempty"`
// 	MessageDelta    *MessageDeltaEvent    `json:"message_delta,omitempty"`
// 	VoiceChanged    *VoiceChangedEvent    `json:"voice_changed,omitempty"`
// 	Notification    *NotificationEvent    `json:"notification,omitempty"`
// 	ResetTurnOutput *ResetTurnOutputEvent `json:"reset_turn_output,omitempty"`
// 	ResetTurn       *ResetTurnEvent       `json:"reset_turn,omitempty"`
// 	HistoryReset    *HistoryResetEvent    `json:"history_reset,omitempty"`
// }

// func (s *ServerEvent) GetEventID() string {
// 	return s.EventID
// }

// func (s *ServerEvent) GetEventType() string {
// 	return s.EventType
// }

// // 创建不同类型的服务端事件的工厂方法
// func NewSessionCreatedServerEvent(eventID string, event *SessionCreatedEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:        eventID,
// 		EventType:      "session_created",
// 		SessionCreated: event,
// 	}
// }

// func NewSessionClosedServerEvent(eventID string, event *SessionClosedEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:       eventID,
// 		EventType:     "session_closed",
// 		SessionClosed: event,
// 	}
// }

// func NewBotStatusServerEvent(eventID string, event *BotStatusEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:   eventID,
// 		EventType: "bot_status",
// 		BotStatus: event,
// 	}
// }

// func NewHistoryAppendServerEvent(eventID string, event *HistoryAppendEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:       eventID,
// 		EventType:     "history_append",
// 		HistoryAppend: event,
// 	}
// }

// func NewMessageDeltaServerEvent(eventID string, event *MessageDeltaEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:      eventID,
// 		EventType:    "message_delta",
// 		MessageDelta: event,
// 	}
// }

// func NewVoiceChangedServerEvent(eventID string, event *VoiceChangedEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:      eventID,
// 		EventType:    "voice_changed",
// 		VoiceChanged: event,
// 	}
// }

// func NewNotificationServerEvent(eventID string, event *NotificationEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:      eventID,
// 		EventType:    "notification",
// 		Notification: event,
// 	}
// }

// func NewResetTurnOutputServerEvent(eventID string, event *ResetTurnOutputEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:         eventID,
// 		EventType:       "reset_turn_output",
// 		ResetTurnOutput: event,
// 	}
// }

// func NewResetTurnServerEvent(eventID string, event *ResetTurnEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:   eventID,
// 		EventType: "reset_turn",
// 		ResetTurn: event,
// 	}
// }

// func NewHistoryResetServerEvent(eventID string, event *HistoryResetEvent) *ServerEvent {
// 	return &ServerEvent{
// 		EventID:      eventID,
// 		EventType:    "history_reset",
// 		HistoryReset: event,
// 	}
// }
