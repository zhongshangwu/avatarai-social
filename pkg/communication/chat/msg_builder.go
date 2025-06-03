package chat

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
)

func BuildMessageFromSendMsgEvent(sendMsgEvent *messages.SendMsgEvent) (*messages.Message, error) {
	message := &messages.Message{
		ID:         GenerateMessageID(),
		RoomID:     sendMsgEvent.RoomID,
		ThreadID:   sendMsgEvent.ThreadID,
		QuoteMID:   sendMsgEvent.QuoteMID,
		MsgType:    sendMsgEvent.MsgType,
		ReceiverID: sendMsgEvent.ReceiverID,
		SenderID:   sendMsgEvent.SenderID,
		SenderAt:   sendMsgEvent.SenderAt, // 定时发送，暂时没有实现
		ExternalID: sendMsgEvent.ExternalID,
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
		Deleted:    false,
	}

	switch sendMsgEvent.MsgType {
	case messages.MessageTypeText:
		body := sendMsgEvent.Body.(*messages.TextMsgBody)
		message.Content = &messages.TextMessageContent{
			Text: body.Text,
		}
	case messages.MessageTypeImage:
		body := sendMsgEvent.Body.(*messages.ImageMsgBody)
		message.Content = &messages.ImageMessageContent{
			ImageCID: body.ImageCID,
			Width:    body.Width,
			Height:   body.Height,
			Alt:      body.Alt,
		}
	case messages.MessageTypeVideo:
		body := sendMsgEvent.Body.(*messages.VideoMsgBody)
		message.Content = &messages.VideoMessageContent{
			VideoCID: body.VideoCID,
			Duration: body.Duration,
			ThumbCID: body.ThumbCID,
		}
	case messages.MessageTypeAudio:
		body := sendMsgEvent.Body.(*messages.AudioMsgBody)
		message.Content = &messages.AudioMessageContent{
			AudioCID: body.AudioCID,
			Duration: body.Duration,
		}
	case messages.MessageTypeAgent:
		body := sendMsgEvent.Body.(*messages.AgentMsgBody)
		messageItems := make([]messages.MessageItem, len(body.MessageItems))
		for i, item := range body.MessageItems {
			messageItems[i] = messages.MessageItem(item)
		}
		message.Content = &messages.AgentMessageContent{
			AgentMessage: messages.AgentMessage{
				ID:           GenerateMessageID(),
				MessageID:    body.MessageID,
				Role:         messages.RoleType(body.Role),
				MessageItems: messageItems,
				Metadata:     body.Metadata,
			},
		}
	case messages.MessageTypeSticker:
		body := sendMsgEvent.Body.(*messages.StickerMsgBody)
		message.Content = &messages.StickerMessageContent{
			StickerCID: body.StickerCID,
			Alt:        body.Alt,
			Width:      body.Width,
			Height:     body.Height,
		}
	case messages.MessageTypePost:
		body := sendMsgEvent.Body.(*messages.PostMsgBody)
		message.Content = &messages.PostMessageContent{
			Title:   body.Title,
			Content: body.Content,
		}
	case messages.MessageTypeFile:
		body := sendMsgEvent.Body.(*messages.FileMsgBody)
		message.Content = &messages.FileMessageContent{
			FileCID:  body.FileCID,
			Size:     body.Size,
			FileName: body.FileName,
			MimeType: body.MimeType,
		}
	default:
		return nil, fmt.Errorf("unsupported message type: %s", sendMsgEvent.MsgType)
	}
	return message, nil
}

func (actor *ChatActor) InitRespondMessage(input *messages.Message) (*messages.Message, error) {
	message := &messages.Message{
		ID:         GenerateMessageID(),
		RoomID:     input.RoomID,
		ThreadID:   input.ThreadID,
		QuoteMID:   "",
		MsgType:    messages.MessageTypeAgent,
		ReceiverID: input.SenderID,
		SenderID:   input.ReceiverID,
		SenderAt:   time.Now().UnixMilli(),
		ExternalID: "",
		CreatedAt:  time.Now().UnixMilli(),
		UpdatedAt:  time.Now().UnixMilli(),
		Deleted:    false,
	}
	aiChatMessage := &messages.AgentMessage{
		ID:            GenerateAgentMessageID(),
		MessageID:     "",
		Role:          messages.RoleTypeAssistant,
		AltText:       "",
		MessageItems:  make([]messages.MessageItem, 0),
		InterruptType: 0,
		Status:        messages.AgentMessageStatusInProgress,
		Creator:       input.ReceiverID,
		CreatedAt:     time.Now().UnixMilli(),
		UpdatedAt:     time.Now().UnixMilli(),
		Metadata:      make(map[string]interface{}),
	}
	dbAgentMessage := aiChatMessage.ToDB()
	if err := database.InsertAgentMessage(actor.DB, dbAgentMessage); err != nil {
		logrus.Errorf("insert ai chat message failed: %v", err)
		return nil, err
	}

	message.Content = &messages.AgentMessageContent{
		AgentMessage: *aiChatMessage,
	}
	dbMessage := message.ToDB()
	if err := database.InsertMessage(actor.DB, dbMessage); err != nil {
		logrus.Errorf("insert message failed: %v", err)
		return nil, err
	}
	return message, nil
}

func GenerateMessageID() string {
	return uuid.New().String()
}

func GenerateAgentMessageID() string {
	return uuid.New().String()
}
