package messages

import (
	"encoding/json"
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

func DBToMessage(dbMsg *database.Message) *Message {
	if dbMsg == nil {
		return nil
	}

	message := &Message{
		ID:         dbMsg.ID,
		RoomID:     dbMsg.RoomID,
		ThreadID:   dbMsg.ThreadID,
		MsgType:    MessageType(dbMsg.MsgType),
		ReceiverID: dbMsg.ReceiverID,
		SenderID:   dbMsg.SenderID,
		QuoteMID:   dbMsg.QuoteMID,
		SenderAt:   dbMsg.SenderAt,
		CreatedAt:  dbMsg.CreatedAt,
		UpdatedAt:  dbMsg.UpdatedAt,
		Deleted:    dbMsg.Deleted,
		ExternalID: dbMsg.ExternalID,
	}

	if dbMsg.Content != "" {
		content, err := parseMessageContent(MessageType(dbMsg.MsgType), dbMsg.Content)
		if err != nil {
			// 记录错误但不返回nil，保持消息的其他信息
			return message
		}
		message.Content = content
	}

	return message
}

func MessageToDB(message *Message) *database.Message {
	if message == nil {
		return nil
	}

	var contentStr string
	if message.Content != nil {
		content := serializeMessageContent(message.Content)
		if content != nil {
			contentBytes, _ := json.Marshal(content)
			contentStr = string(contentBytes)
		}
	}

	return &database.Message{
		ID:         message.ID,
		RoomID:     message.RoomID,
		ThreadID:   message.ThreadID,
		MsgType:    int(message.MsgType),
		Content:    contentStr,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		QuoteMID:   message.QuoteMID,
		SenderAt:   message.SenderAt,
		ExternalID: message.ExternalID,
		CreatedAt:  message.CreatedAt,
		UpdatedAt:  message.UpdatedAt,
		Deleted:    message.Deleted,
	}
}

func DBToAgentMessage(dbAgentMessage *database.AgentMessage) *AgentMessage {
	if dbAgentMessage == nil {
		return nil
	}

	agentMessage := &AgentMessage{
		ID:            dbAgentMessage.ID,
		MessageID:     dbAgentMessage.MessageID,
		Role:          RoleType(dbAgentMessage.Role),
		AltText:       dbAgentMessage.AltText,
		InterruptType: dbAgentMessage.InterruptType,
		Status:        AgentMessageStatus(dbAgentMessage.Status),
		Creator:       dbAgentMessage.Creator,
		CreatedAt:     dbAgentMessage.CreatedAt,
		UpdatedAt:     dbAgentMessage.UpdatedAt,
		MessageItems:  make([]MessageItem, 0),
		Metadata:      make(map[string]interface{}),
	}

	// 解析JSON字段
	if dbAgentMessage.Error != "" && dbAgentMessage.Error != "null" {
		var errorInfo ResponseError
		if err := json.Unmarshal([]byte(dbAgentMessage.Error), &errorInfo); err == nil {
			agentMessage.Error = &errorInfo
		}
	}

	if dbAgentMessage.Usage != "" && dbAgentMessage.Usage != "null" {
		var usage ResponseUsage
		if err := json.Unmarshal([]byte(dbAgentMessage.Usage), &usage); err == nil {
			agentMessage.Usage = &usage
		}
	}

	if dbAgentMessage.Metadata != "" && dbAgentMessage.Metadata != "null" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(dbAgentMessage.Metadata), &metadata); err == nil {
			agentMessage.Metadata = metadata
		}
	}

	if dbAgentMessage.IncompleteDetails != "" && dbAgentMessage.IncompleteDetails != "null" {
		var incompleteDetails IncompleteDetails
		if err := json.Unmarshal([]byte(dbAgentMessage.IncompleteDetails), &incompleteDetails); err == nil {
			agentMessage.IncompleteDetails = &incompleteDetails
		}
	}

	return agentMessage
}

func AgentMessageToDB(agentMessage *AgentMessage) *database.AgentMessage {
	if agentMessage == nil {
		return nil
	}

	var errStr, usageStr, metadataStr, incompleteDetailsStr string

	if agentMessage.Error != nil {
		if errBytes, err := json.Marshal(agentMessage.Error); err == nil {
			errStr = string(errBytes)
		}
	}

	if agentMessage.Usage != nil {
		if usageBytes, err := json.Marshal(agentMessage.Usage); err == nil {
			usageStr = string(usageBytes)
		}
	}

	if agentMessage.Metadata != nil {
		if metadataBytes, err := json.Marshal(agentMessage.Metadata); err == nil {
			metadataStr = string(metadataBytes)
		}
	}

	if agentMessage.IncompleteDetails != nil {
		if incompleteBytes, err := json.Marshal(agentMessage.IncompleteDetails); err == nil {
			incompleteDetailsStr = string(incompleteBytes)
		}
	}

	return &database.AgentMessage{
		ID:                agentMessage.ID,
		MessageID:         agentMessage.MessageID,
		Role:              string(agentMessage.Role),
		AltText:           agentMessage.AltText,
		Status:            string(agentMessage.Status),
		InterruptType:     agentMessage.InterruptType,
		Error:             errStr,
		Usage:             usageStr,
		Metadata:          metadataStr,
		IncompleteDetails: incompleteDetailsStr,
		Creator:           agentMessage.Creator,
		CreatedAt:         agentMessage.CreatedAt,
		UpdatedAt:         agentMessage.UpdatedAt,
	}
}

func parseMessageContent(msgType MessageType, contentStr string) (MessageContent, error) {
	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(contentStr), &contentMap); err != nil {
		return nil, fmt.Errorf("解析JSON内容失败: %w", err)
	}

	switch msgType {
	case MessageTypeText:
		return parseTextContent(contentMap)
	case MessageTypeImage:
		return parseImageContent(contentMap)
	case MessageTypeVideo:
		return parseVideoContent(contentMap)
	case MessageTypeFile:
		return parseFileContent(contentMap)
	case MessageTypeAudio:
		return parseAudioContent(contentMap)
	case MessageTypeAgent:
		return parseAgentContent(contentMap)
	case MessageTypePost:
		return parsePostContent(contentMap)
	case MessageTypeSticker:
		return parseStickerContent(contentMap)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %d", msgType)
	}
}

func serializeMessageContent(content MessageContent) map[string]interface{} {
	switch c := content.(type) {
	case *TextMessageContent:
		return map[string]interface{}{
			"text": c.Text,
		}
	case *ImageMessageContent:
		return map[string]interface{}{
			"imageCid": c.ImageCID,
			"width":    c.Width,
			"height":   c.Height,
			"alt":      c.Alt,
		}
	case *VideoMessageContent:
		return map[string]interface{}{
			"videoCid": c.VideoCID,
			"duration": c.Duration,
			"thumbCid": c.ThumbCID,
			"width":    c.Width,
			"height":   c.Height,
		}
	case *FileMessageContent:
		return map[string]interface{}{
			"fileCid":  c.FileCID,
			"size":     c.Size,
			"fileName": c.FileName,
			"mimeType": c.MimeType,
			"fileType": c.FileType,
		}
	case *AudioMessageContent:
		return map[string]interface{}{
			"audioCid":   c.AudioCID,
			"duration":   c.Duration,
			"transcript": c.Transcript,
		}
	case *AgentMessageContent:
		return map[string]interface{}{
			"agentMessageId": c.AgentMessage.ID,
		}
	case *PostMessageContent:
		return map[string]interface{}{
			"title":   c.Title,
			"content": c.Content,
		}
	case *StickerMessageContent:
		return map[string]interface{}{
			"stickerCid": c.StickerCID,
			"alt":        c.Alt,
			"width":      c.Width,
			"height":     c.Height,
			"isAnimated": c.IsAnimated,
		}
	default:
		return nil
	}
}

func parseTextContent(contentMap map[string]interface{}) (*TextMessageContent, error) {
	text, ok := contentMap["text"].(string)
	if !ok {
		return nil, fmt.Errorf("文本消息缺少text字段")
	}
	return &TextMessageContent{Text: text}, nil
}

func parseImageContent(contentMap map[string]interface{}) (*ImageMessageContent, error) {
	content := &ImageMessageContent{
		ImageCID: utils.GetStringFromMap(contentMap, "imageCid"),
		Width:    utils.GetIntFromMap(contentMap, "width"),
		Height:   utils.GetIntFromMap(contentMap, "height"),
		Alt:      utils.GetStringFromMap(contentMap, "alt"),
	}

	if content.ImageCID == "" {
		return nil, fmt.Errorf("图片消息缺少imageCid字段")
	}

	return content, nil
}

func parseVideoContent(contentMap map[string]interface{}) (*VideoMessageContent, error) {
	content := &VideoMessageContent{
		VideoCID: utils.GetStringFromMap(contentMap, "videoCid"),
		Duration: utils.GetIntFromMap(contentMap, "duration"),
		ThumbCID: utils.GetStringFromMap(contentMap, "thumbCid"),
		Width:    utils.GetIntFromMap(contentMap, "width"),
		Height:   utils.GetIntFromMap(contentMap, "height"),
	}

	if content.VideoCID == "" {
		return nil, fmt.Errorf("视频消息缺少videoCid字段")
	}

	return content, nil
}

func parseFileContent(contentMap map[string]interface{}) (*FileMessageContent, error) {
	content := &FileMessageContent{
		FileCID:  utils.GetStringFromMap(contentMap, "fileCid"),
		Size:     utils.GetInt64FromMap(contentMap, "size"),
		FileName: utils.GetStringFromMap(contentMap, "fileName"),
		MimeType: utils.GetStringFromMap(contentMap, "mimeType"),
		FileType: utils.GetStringFromMap(contentMap, "fileType"),
	}

	if content.FileCID == "" {
		return nil, fmt.Errorf("文件消息缺少fileCid字段")
	}

	return content, nil
}

func parseAudioContent(contentMap map[string]interface{}) (*AudioMessageContent, error) {
	content := &AudioMessageContent{
		AudioCID:   utils.GetStringFromMap(contentMap, "audioCid"),
		Duration:   utils.GetIntFromMap(contentMap, "duration"),
		Transcript: utils.GetStringFromMap(contentMap, "transcript"),
	}

	if content.AudioCID == "" {
		return nil, fmt.Errorf("音频消息缺少audioCid字段")
	}

	return content, nil
}

func parseAgentContent(contentMap map[string]interface{}) (*AgentMessageContent, error) {
	agentMessageID := utils.GetStringFromMap(contentMap, "agentMessageId")
	if agentMessageID == "" {
		return nil, fmt.Errorf("Agent消息缺少agentMessageId字段")
	}

	// 注意：这里只存储了AgentMessage的ID，实际的AgentMessage对象需要单独查询
	// 在实际使用中，可能需要通过数据库查询来获取完整的AgentMessage对象
	return &AgentMessageContent{
		AgentMessage: AgentMessage{ID: agentMessageID},
	}, nil
}

func parsePostContent(contentMap map[string]interface{}) (*PostMessageContent, error) {
	title := utils.GetStringFromMap(contentMap, "title")

	// 解析富文本内容
	var content [][]RichTextNode
	if contentData, ok := contentMap["content"]; ok {
		// 这里需要根据实际的富文本结构进行解析
		// 由于富文本结构比较复杂，这里提供一个基础实现
		contentBytes, err := json.Marshal(contentData)
		if err != nil {
			return nil, fmt.Errorf("解析富文本内容失败: %w", err)
		}

		if err := json.Unmarshal(contentBytes, &content); err != nil {
			// 如果解析失败，创建一个空的内容数组
			content = make([][]RichTextNode, 0)
		}
	}

	return &PostMessageContent{
		Title:   title,
		Content: content,
	}, nil
}

func parseStickerContent(contentMap map[string]interface{}) (*StickerMessageContent, error) {
	content := &StickerMessageContent{
		StickerCID: utils.GetStringFromMap(contentMap, "stickerCid"),
		Alt:        utils.GetStringFromMap(contentMap, "alt"),
		Width:      utils.GetIntFromMap(contentMap, "width"),
		Height:     utils.GetIntFromMap(contentMap, "height"),
		IsAnimated: utils.GetBoolFromMap(contentMap, "isAnimated"),
	}

	if content.StickerCID == "" {
		return nil, fmt.Errorf("表情包消息缺少stickerCid字段")
	}

	return content, nil
}
