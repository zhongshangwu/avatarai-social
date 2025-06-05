package messages

import (
	"encoding/json"
	"fmt"
)

type Room struct {
	ID           string   `json:"id"`           // 房间号
	Title        string   `json:"title"`        // 房间标题 (群组可以命名名称)
	Type         string   `json:"type"`         // 房间类型 // 单聊、群聊、ai 对话...
	LastMID      string   `json:"lastMid"`      // 最后一条消息ID
	Participants []string `json:"participants"` // 参与者
	CreatedAt    int64    `json:"createdAt"`    // 创建时间
	UpdatedAt    int64    `json:"updatedAt"`    // 更新时间
	Deleted      bool     `json:"deleted"`      // 是否被删除
}

type UserRoomStatus struct { // 归属某个具体的用户, 基本是一个 room 的镜像数据
	ID           string   `json:"id"`           // 唯一标识 (业务上使用 room_id + userid 来唯一标识)
	RoomID       string   `json:"roomId"`       // 房间号， 对话等都使用该 room id
	Title        string   `json:"title"`        // 房间标题 (群组可以命名名称)
	Type         string   `json:"type"`         // 房间类型 // 单聊、群聊、ai 对话...
	LastMID      string   `json:"lastMid"`      // 最后一条消息ID
	Participants []string `json:"participants"` // 参与者
	UnreadCount  int32    `json:"unreadCount"`  // 未读消息数
	Muted        bool     `json:"muted"`        // 是否被静音
	UserID       string   `json:"userId"`       // 用户ID
	Status       string   `json:"status"`       // 状态 // request, accepted
	CreatedAt    int64    `json:"createdAt"`    // 创建时间
	UpdatedAt    int64    `json:"updatedAt"`    // 更新时间
	Deleted      bool     `json:"deleted"`      // 是否被删除
}

type ThreadContextMode string

const (
	ThreadContextModeContinuous ThreadContextMode = "continuous" // 连续上下文
	ThreadContextModeIsolated   ThreadContextMode = "isolated"   // 隔离上下文
)

type Thread struct {
	ID             string            `json:"id"`             // 话题ID
	RoomID         string            `json:"roomId"`         // 房间ID
	Title          string            `json:"title"`          // 话题标题
	ContextMode    ThreadContextMode `json:"contextMode"`    // 话题上下文类型: 连续上下文/独立上下文
	RootMID        string            `json:"rootMid"`        // 根消息ID
	ParentThreadID string            `json:"parentThreadId"` // 父话题ID
	CreatedAt      int64             `json:"createdAt"`      // 创建时间
	UpdatedAt      int64             `json:"updatedAt"`      // 更新时间
	Deleted        bool              `json:"deleted"`        // 是否被删除
}

type MessageContent interface {
	Type() MessageType
	isMessageContent()
}

type Message struct {
	ID       string `json:"id"`       // 消息ID
	RoomID   string `json:"roomId"`   // 房间ID
	ThreadID string `json:"threadId"` // 话题ID
	// RootMID   string `json:"root_mid"`   // 消息回复关系：根消息ID为回复树的根节点消息ID
	// ParentMID string `json:"parent_mid"` // 消息回复关系：父消息ID为被回复的信息ID

	MsgType MessageType    `json:"msgType"` // 消息类型
	Content MessageContent `json:"content"` // 消息体，可以是不同类型的消息结构

	ReceiverID string `json:"receiverId"` // 接收者ID (有一个 receiver_id 会在很多时候方便一些)
	SenderID   string `json:"senderId"`   // 发送者ID
	QuoteMID   string `json:"quoteMid"`   // 引用消息ID (消息回显用，作为上下文提供, 不作为消息组织结构)
	SenderAt   int64  `json:"senderAt"`   // 发送时间
	CreatedAt  int64  `json:"createdAt"`  // 创建时间
	UpdatedAt  int64  `json:"updatedAt"`  // 更新时间
	Deleted    bool   `json:"deleted"`    // 是否被撤回

	ExternalID string `json:"externalId"` // 外部ID
}

type TextMessageContent struct {
	Text string `json:"text"` // 文本内容
}

func (t *TextMessageContent) Type() MessageType {
	return MessageTypeText
}

func (t *TextMessageContent) isMessageContent() {}

type ImageMessageContent struct {
	ImageURL string `json:"imageUrl"`         // 图片URL
	ImageCID string `json:"imageCid"`         // 图片内容ID
	Width    int    `json:"width,omitempty"`  // 图片宽度
	Height   int    `json:"height,omitempty"` // 图片高度
	Alt      string `json:"alt,omitempty"`    // 替代文本
}

func (i *ImageMessageContent) Type() MessageType {
	return MessageTypeImage
}

func (i *ImageMessageContent) isMessageContent() {}

type VideoMessageContent struct {
	VideoURL string `json:"videoUrl"`         // 视频URL
	VideoCID string `json:"videoCid"`         // 视频内容ID
	Duration int    `json:"duration"`         // 视频时长（秒）
	ThumbURL string `json:"thumbUrl"`         // 缩略图URL
	ThumbCID string `json:"thumbCid"`         // 缩略图内容ID
	Width    int    `json:"width,omitempty"`  // 视频宽度
	Height   int    `json:"height,omitempty"` // 视频高度
}

func (v *VideoMessageContent) Type() MessageType {
	return MessageTypeVideo
}

func (v *VideoMessageContent) isMessageContent() {}

type FileMessageContent struct {
	FileURL  string `json:"fileUrl"`  // 文件URL
	FileCID  string `json:"fileCid"`  // 文件内容ID
	Size     int64  `json:"size"`     // 文件大小（字节）
	FileName string `json:"fileName"` // 文件名
	MimeType string `json:"mimeType"` // MIME类型
	FileType string `json:"fileType"` // 文件类型
}

func (f *FileMessageContent) Type() MessageType {
	return MessageTypeFile
}

func (f *FileMessageContent) isMessageContent() {}

type AudioMessageContent struct {
	AudioURL   string `json:"audioUrl"`             // 音频URL
	AudioCID   string `json:"audioCid"`             // 音频内容ID
	Duration   int    `json:"duration"`             // 音频时长（秒）
	Transcript string `json:"transcript,omitempty"` // 转录文本
}

func (a *AudioMessageContent) Type() MessageType {
	return MessageTypeAudio
}

func (a *AudioMessageContent) isMessageContent() {}

type AgentMessageContent struct {
	AgentMessage AgentMessage `json:"message"`
}

func (a *AgentMessageContent) Type() MessageType {
	return MessageTypeAgent
}

func (a *AgentMessageContent) isMessageContent() {}

type StickerMessageContent struct {
	StickerURL string `json:"stickerUrl"`       // 表情包URL
	StickerCID string `json:"stickerCid"`       // 表情包内容ID
	Alt        string `json:"alt,omitempty"`    // 替代文本
	Width      int    `json:"width,omitempty"`  // 宽度
	Height     int    `json:"height,omitempty"` // 高度
	IsAnimated bool   `json:"isAnimated"`       // 是否为动画表情
}

func (s *StickerMessageContent) Type() MessageType {
	return MessageTypeSticker
}

func (s *StickerMessageContent) isMessageContent() {}

type PostMessageContent struct {
	Title   string           `json:"title"`   // 富文本标题
	Content [][]RichTextNode `json:"content"` // 富文本内容
}

func (p *PostMessageContent) Type() MessageType {
	return MessageTypePost
}

func (p *PostMessageContent) isMessageContent() {}

func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
		Content json.RawMessage `json:"content"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Content) > 0 {
		var body MessageContent
		switch m.MsgType {
		case MessageTypeText:
			body = &TextMessageContent{}
		case MessageTypeImage:
			body = &ImageMessageContent{}
		case MessageTypeVideo:
			body = &VideoMessageContent{}
		case MessageTypeFile:
			body = &FileMessageContent{}
		case MessageTypeAudio:
			body = &AudioMessageContent{}
		case MessageTypeAgent:
			body = &AgentMessageContent{}
		case MessageTypePost:
			body = &PostMessageContent{}
		case MessageTypeSticker:
			body = &StickerMessageContent{}
		default:
			return fmt.Errorf("不支持的消息类型: %d", m.MsgType)
		}

		if err := json.Unmarshal(aux.Content, body); err != nil {
			return fmt.Errorf("解析消息体失败: %w", err)
		}

		m.Content = body
	}
	return nil
}
