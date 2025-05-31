package messages

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type SendMsgEvent struct {
	RoomID   string      `json:"roomId"`             // 房间ID
	MsgType  MessageType `json:"msgType"`            // 消息类型
	Body     SendMsgBody `json:"body"`               // 消息体
	SenderID string      `json:"senderId"`           // 发送者ID
	ThreadID string      `json:"threadId,omitempty"` // 话题ID
	QuoteID  string      `json:"quoteId,omitempty"`  // 引用ID
	SenderAt int64       `json:"senderAt,omitempty"` // 发送时间
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

	// 根据 msgType 确定消息体的具体类型
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
	case MessageTypeAIChat:
		body = &AIChatMsgBody{}
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
	ImageCID string `json:"image_cid"`        // 图片内容ID
	Width    int    `json:"width,omitempty"`  // 图片宽度
	Height   int    `json:"height,omitempty"` // 图片高度
	Alt      string `json:"alt,omitempty"`    // 替代文本
}

func (i *ImageMsgBody) isSendMsgBody() {}

type VideoMsgBody struct {
	VideoCID string `json:"video_cid"`        // 视频内容ID
	Duration int    `json:"duration"`         // 视频时长（秒）
	ThumbCID string `json:"thumb_cid"`        // 缩略图内容ID
	Width    int    `json:"width,omitempty"`  // 视频宽度
	Height   int    `json:"height,omitempty"` // 视频高度
}

func (v *VideoMsgBody) isSendMsgBody() {}

type FileMsgBody struct {
	FileCID  string `json:"file_cid"`  // 文件内容ID
	Size     int64  `json:"size"`      // 文件大小（字节）
	FileName string `json:"file_name"` // 文件名
	MimeType string `json:"mime_type"` // MIME类型
	FileType string `json:"file_type"` // 文件类型
}

func (f *FileMsgBody) isSendMsgBody() {}

type AudioMsgBody struct {
	AudioCID   string `json:"audio_cid"`            // 音频内容ID
	Duration   int    `json:"duration"`             // 音频时长（秒）
	Transcript string `json:"transcript,omitempty"` // 转录文本
}

func (a *AudioMsgBody) isSendMsgBody() {}

type AIChatMsgBody struct {
	MessageItems []InputItem `json:"messageItems"` // 消息项列表
}

func (a *AIChatMsgBody) isSendMsgBody() {}

func (a *AIChatMsgBody) UnmarshalJSON(data []byte) error {
	type Alias AIChatMsgBody
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
	StickerCID string `json:"sticker_cid"`      // 表情包内容ID
	Alt        string `json:"alt,omitempty"`    // 替代文本
	Width      int    `json:"width,omitempty"`  // 宽度
	Height     int    `json:"height,omitempty"` // 高度
	IsAnimated bool   `json:"is_animated"`      // 是否为动画表情
}

func (s *StickerMsgBody) isSendMsgBody() {}
