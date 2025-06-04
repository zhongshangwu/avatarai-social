package messages

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type SendMsgEvent struct {
	RoomID     string      `json:"roomId"`               // 房间ID
	MsgType    MessageType `json:"msgType"`              // 消息类型
	Body       SendMsgBody `json:"body"`                 // 消息体
	ReceiverID string      `json:"receiverId"`           // 接收者ID
	SenderID   string      `json:"senderId"`             // 发送者ID
	ThreadID   string      `json:"threadId,omitempty"`   // 话题ID
	QuoteMID   string      `json:"quoteMid,omitempty"`   // 引用ID
	SenderAt   int64       `json:"senderAt,omitempty"`   // 发送时间
	ExternalID string      `json:"externalId,omitempty"` // 外部ID
}

func (s *SendMsgEvent) isChatEventBody() {}

func (s *SendMsgEvent) UnmarshalJSON(data []byte) error {
	type Alias SendMsgEvent
	aux := &struct {
		Body json.RawMessage `json:"body"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var body SendMsgBody
	switch s.MsgType {
	case MessageTypeText:
		body = &TextMsgBody{}
	case MessageTypeImage:
		body = &ImageMsgBody{}
	case MessageTypeVideo:
		body = &VideoMsgBody{}
	case MessageTypeFile:
		body = &FileMsgBody{}
	case MessageTypeAudio:
		body = &AudioMsgBody{}
	case MessageTypeAgent:
		body = &AgentMsgBody{}
	case MessageTypePost:
		body = &PostMsgBody{}
	case MessageTypeSticker:
		body = &StickerMsgBody{}
	default:
		return fmt.Errorf("不支持的消息类型: %d", s.MsgType)
	}

	if err := json.Unmarshal(aux.Body, body); err != nil {
		logrus.WithError(err).Errorf("解析消息体失败，消息类型: %d", s.MsgType)
		return fmt.Errorf("解析消息体失败: %w", err)
	}

	s.Body = body
	return nil
}

type SendMsgBody interface {
	isSendMsgBody()
}

type TextMsgBody struct {
	Text string `json:"text"` // 文本内容
}

func (t *TextMsgBody) isSendMsgBody() {}

type ImageMsgBody struct {
	ImageCID string `json:"imageCid"`         // 图片内容ID
	Width    int    `json:"width,omitempty"`  // 图片宽度
	Height   int    `json:"height,omitempty"` // 图片高度
	Alt      string `json:"alt,omitempty"`    // 替代文本
}

func (i *ImageMsgBody) isSendMsgBody() {}

type VideoMsgBody struct {
	VideoCID string `json:"videoCid"`         // 视频内容ID
	Duration int    `json:"duration"`         // 视频时长（秒）
	ThumbCID string `json:"thumbCid"`         // 缩略图内容ID
	Width    int    `json:"width,omitempty"`  // 视频宽度
	Height   int    `json:"height,omitempty"` // 视频高度
}

func (v *VideoMsgBody) isSendMsgBody() {}

type FileMsgBody struct {
	FileCID  string `json:"fileCid"`  // 文件内容ID
	Size     int64  `json:"size"`     // 文件大小（字节）
	FileName string `json:"fileName"` // 文件名
	MimeType string `json:"mimeType"` // MIME类型
	FileType string `json:"fileType"` // 文件类型
}

func (f *FileMsgBody) isSendMsgBody() {}

type AudioMsgBody struct {
	AudioCID   string `json:"audioCid"`             // 音频内容ID
	Duration   int    `json:"duration"`             // 音频时长（秒）
	Transcript string `json:"transcript,omitempty"` // 转录文本
}

func (a *AudioMsgBody) isSendMsgBody() {}

type AgentMsgBody struct {
	Role         string         `json:"role"`         // 角色
	MessageItems []InputItem    `json:"messageItems"` // 消息项列表
	Metadata     map[string]any `json:"metadata"`     // 元数据
	// MessageID    string         `json:"messageId"`    // 消息ID （由创建方提供的 messageid )
}

func (a *AgentMsgBody) isSendMsgBody() {}

func (a *AgentMsgBody) UnmarshalJSON(data []byte) error {
	type Alias AgentMsgBody
	aux := &struct {
		MessageItems []json.RawMessage `json:"messageItems"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.MessageItems = make([]InputItem, len(aux.MessageItems))
	for i, raw := range aux.MessageItems {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var item InputItem
		switch typeName {
		case "message":
			item = &InputMessage{}
		case "tool_call", "function_call":
			item = &FunctionToolCall{}
		case "item_reference":
			item = &ItemReferenceParam{}
		case "easy_message":
			item = &EasyInputMessage{}
		default:
			return fmt.Errorf("未知的输入项类型: %s", typeName)
		}

		if err := json.Unmarshal(raw, item); err != nil {
			return err
		}

		a.MessageItems[i] = item
	}

	return nil
}

type StickerMsgBody struct {
	StickerCID string `json:"stickerCid"`       // 表情包内容ID
	Alt        string `json:"alt,omitempty"`    // 替代文本
	Width      int    `json:"width,omitempty"`  // 宽度
	Height     int    `json:"height,omitempty"` // 高度
	IsAnimated bool   `json:"isAnimated"`       // 是否为动画表情
}

func (s *StickerMsgBody) isSendMsgBody() {}

type PostMsgBody struct {
	Title   string           `json:"title,omitempty"` // 富文本标题
	Content [][]RichTextNode `json:"content"`         // 富文本内容
}

func (p *PostMsgBody) isSendMsgBody() {}

func (p *PostMsgBody) UnmarshalJSON(data []byte) error {
	type Alias PostMsgBody
	aux := &struct {
		Content [][]json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 解析富文本内容
	p.Content = make([][]RichTextNode, len(aux.Content))
	for i, row := range aux.Content {
		p.Content[i] = make([]RichTextNode, len(row))
		for j, nodeData := range row {
			tag, err := ExtractTag(nodeData)
			if err != nil {
				logrus.WithError(err).Errorf("提取富文本节点标签失败")
				return fmt.Errorf("提取富文本节点标签失败: %w", err)
			}

			var node RichTextNode
			switch RichTextNodeType(tag) {
			case PostNodeText:
				node = &RichTextNodeText{}
			case PostNodeLink:
				node = &RichTextNodeLink{}
			case PostNodeAt:
				node = &RichTextNodeAt{}
			case PostNodeImage:
				node = &RichTextNodeImage{}
			case PostNodeMedia:
				node = &RichTextNodeVideo{}
			case PostNodeEmotion:
				node = &RichTextNodeEmotion{}
			case PostNodeHr:
				node = &RichTextNodeHr{}
			case PostNodeCodeBlock:
				node = &RichTextNodeCodeBlock{}
			case PostNodeMarkdown:
				node = &RichTextNodeMarkdown{}
			default:
				return fmt.Errorf("不支持的富文本节点类型: %s", tag)
			}

			if err := json.Unmarshal(nodeData, node); err != nil {
				logrus.WithError(err).Errorf("解析富文本节点失败，节点类型: %s", tag)
				return fmt.Errorf("解析富文本节点失败: %w", err)
			}

			p.Content[i][j] = node
		}
	}

	return nil
}

type MessageSentEvent struct {
	MessageID string `json:"messageId"` // 消息ID
	EventID   string `json:"eventId"`   // 原始发送消息的事件ID
}

func (s *MessageSentEvent) isChatEventBody() {}

type MessageReceivedEvent struct {
	Message *Message `json:"message"` // 消息
}

func (o *MessageReceivedEvent) isChatEventBody() {}
