package services

import (
	"encoding/json"
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
)

type MessageConverter struct {
	messageRepo *repositories.MessageRepository
}

func NewMessageConverter(messageRepo *repositories.MessageRepository) *MessageConverter {
	return &MessageConverter{
		messageRepo: messageRepo,
	}
}

func (c *MessageConverter) DBToMessage(dbMsg *repositories.Message) *messages.Message {
	if dbMsg == nil {
		return nil
	}

	message := &messages.Message{
		ID:         dbMsg.ID,
		RoomID:     dbMsg.RoomID,
		ThreadID:   dbMsg.ThreadID,
		MsgType:    messages.MessageType(dbMsg.MsgType),
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
		content, err := c.parseMessageContent(messages.MessageType(dbMsg.MsgType), dbMsg.Content)
		if err != nil {
			// 记录错误但不返回nil，保持消息的其他信息
			// TODO: 添加日志记录
			return message
		}
		message.Content = content
	}

	return message
}

func (c *MessageConverter) MessageToDB(message *messages.Message) *repositories.Message {
	if message == nil {
		return nil
	}

	var contentStr string
	if message.Content != nil {
		content := c.serializeMessageContent(message.Content)
		if content != nil {
			contentBytes, _ := json.Marshal(content)
			contentStr = string(contentBytes)
		}
	}

	return &repositories.Message{
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

func (c *MessageConverter) DBToAgentMessage(dbAgentMessage *repositories.AgentMessage) *messages.AgentMessage {
	if dbAgentMessage == nil {
		return nil
	}

	agentMessage := &messages.AgentMessage{
		ID:            dbAgentMessage.ID,
		MessageID:     dbAgentMessage.MessageID,
		Role:          messages.RoleType(dbAgentMessage.Role),
		AltText:       dbAgentMessage.AltText,
		InterruptType: dbAgentMessage.InterruptType,
		Status:        messages.AgentMessageStatus(dbAgentMessage.Status),
		Creator:       dbAgentMessage.Creator,
		CreatedAt:     dbAgentMessage.CreatedAt,
		UpdatedAt:     dbAgentMessage.UpdatedAt,
		MessageItems:  make([]messages.MessageItem, 0),
		Metadata:      make(map[string]interface{}),
	}

	// 解析JSON字段
	c.parseJSONField(dbAgentMessage.Error, &agentMessage.Error)
	c.parseJSONField(dbAgentMessage.Usage, &agentMessage.Usage)
	c.parseJSONField(dbAgentMessage.Metadata, &agentMessage.Metadata)
	c.parseJSONField(dbAgentMessage.IncompleteDetails, &agentMessage.IncompleteDetails)

	return agentMessage
}

func (c *MessageConverter) AgentMessageToDB(agentMessage *messages.AgentMessage) *repositories.AgentMessage {
	if agentMessage == nil {
		return nil
	}

	return &repositories.AgentMessage{
		ID:                agentMessage.ID,
		MessageID:         agentMessage.MessageID,
		Role:              string(agentMessage.Role),
		AltText:           agentMessage.AltText,
		Status:            string(agentMessage.Status),
		InterruptType:     agentMessage.InterruptType,
		Error:             c.serializeJSONField(agentMessage.Error),
		Usage:             c.serializeJSONField(agentMessage.Usage),
		Metadata:          c.serializeJSONField(agentMessage.Metadata),
		IncompleteDetails: c.serializeJSONField(agentMessage.IncompleteDetails),
		Creator:           agentMessage.Creator,
		CreatedAt:         agentMessage.CreatedAt,
		UpdatedAt:         agentMessage.UpdatedAt,
	}
}

func (c *MessageConverter) parseJSONField(jsonStr string, target interface{}) {
	if jsonStr != "" && jsonStr != "null" {
		json.Unmarshal([]byte(jsonStr), target)
	}
}

func (c *MessageConverter) serializeJSONField(data interface{}) string {
	if data == nil {
		return ""
	}
	if bytes, err := json.Marshal(data); err == nil {
		return string(bytes)
	}
	return ""
}

func (c *MessageConverter) parseMessageContent(msgType messages.MessageType, contentStr string) (messages.MessageContent, error) {
	var contentMap map[string]interface{}
	if err := json.Unmarshal([]byte(contentStr), &contentMap); err != nil {
		return nil, fmt.Errorf("解析JSON内容失败: %w", err)
	}

	switch msgType {
	case messages.MessageTypeText:
		return c.parseTextContent(contentMap)
	case messages.MessageTypeImage:
		return c.parseImageContent(contentMap)
	case messages.MessageTypeVideo:
		return c.parseVideoContent(contentMap)
	case messages.MessageTypeFile:
		return c.parseFileContent(contentMap)
	case messages.MessageTypeAudio:
		return c.parseAudioContent(contentMap)
	case messages.MessageTypeAgent:
		return c.parseAgentContent(contentMap)
	case messages.MessageTypePost:
		return c.parsePostContent(contentMap)
	case messages.MessageTypeSticker:
		return c.parseStickerContent(contentMap)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %d", msgType)
	}
}

func (c *MessageConverter) serializeMessageContent(content messages.MessageContent) map[string]interface{} {
	switch c := content.(type) {
	case *messages.TextMessageContent:
		return map[string]interface{}{
			"text": c.Text,
		}
	case *messages.ImageMessageContent:
		return map[string]interface{}{
			"imageCid": c.ImageCID,
			"width":    c.Width,
			"height":   c.Height,
			"alt":      c.Alt,
		}
	case *messages.VideoMessageContent:
		return map[string]interface{}{
			"videoCid": c.VideoCID,
			"duration": c.Duration,
			"thumbCid": c.ThumbCID,
			"width":    c.Width,
			"height":   c.Height,
		}
	case *messages.FileMessageContent:
		return map[string]interface{}{
			"fileCid":  c.FileCID,
			"size":     c.Size,
			"fileName": c.FileName,
			"mimeType": c.MimeType,
			"fileType": c.FileType,
		}
	case *messages.AudioMessageContent:
		return map[string]interface{}{
			"audioCid":   c.AudioCID,
			"duration":   c.Duration,
			"transcript": c.Transcript,
		}
	case *messages.AgentMessageContent:
		return map[string]interface{}{
			"agentMessageId": c.AgentMessage.ID,
		}
	case *messages.PostMessageContent:
		return map[string]interface{}{
			"title":   c.Title,
			"content": c.Content,
		}
	case *messages.StickerMessageContent:
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
func (c *MessageConverter) parseTextContent(contentMap map[string]interface{}) (*messages.TextMessageContent, error) {
	text, ok := contentMap["text"].(string)
	if !ok {
		return nil, fmt.Errorf("文本消息缺少text字段")
	}
	return &messages.TextMessageContent{Text: text}, nil
}

func (c *MessageConverter) parseImageContent(contentMap map[string]interface{}) (*messages.ImageMessageContent, error) {
	content := &messages.ImageMessageContent{
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

func (c *MessageConverter) parseVideoContent(contentMap map[string]interface{}) (*messages.VideoMessageContent, error) {
	content := &messages.VideoMessageContent{
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

func (c *MessageConverter) parseFileContent(contentMap map[string]interface{}) (*messages.FileMessageContent, error) {
	content := &messages.FileMessageContent{
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

func (c *MessageConverter) parseAudioContent(contentMap map[string]interface{}) (*messages.AudioMessageContent, error) {
	content := &messages.AudioMessageContent{
		AudioCID:   utils.GetStringFromMap(contentMap, "audioCid"),
		Duration:   utils.GetIntFromMap(contentMap, "duration"),
		Transcript: utils.GetStringFromMap(contentMap, "transcript"),
	}

	if content.AudioCID == "" {
		return nil, fmt.Errorf("音频消息缺少audioCid字段")
	}

	return content, nil
}

func (c *MessageConverter) parseAgentContent(contentMap map[string]interface{}) (*messages.AgentMessageContent, error) {
	agentMessageID := utils.GetStringFromMap(contentMap, "agentMessageId")
	if agentMessageID == "" {
		return nil, fmt.Errorf("Agent消息缺少agentMessageId字段")
	}

	// 从数据库查询完整的 AgentMessage 对象
	dbAgentMessage, err := c.messageRepo.GetAgentMessageByID(agentMessageID)
	if err != nil {
		return nil, fmt.Errorf("查询AgentMessage失败: %w", err)
	}

	// 转换为业务对象
	agentMessage := c.DBToAgentMessage(dbAgentMessage)
	if agentMessage == nil {
		return nil, fmt.Errorf("转换AgentMessage失败")
	}

	// 查询 AgentMessage 的 MessageItems
	messageItems, err := c.loadAgentMessageItems(agentMessageID)
	if err != nil {
		// 记录错误但不中断，MessageItems 可以为空
		// TODO: 添加日志记录
		messageItems = make([]messages.MessageItem, 0)
	}
	agentMessage.MessageItems = messageItems

	return &messages.AgentMessageContent{
		AgentMessage: *agentMessage,
	}, nil
}

func (c *MessageConverter) loadAgentMessageItems(agentMessageID string) ([]messages.MessageItem, error) {
	// 查询数据库中的 AgentMessageItem 记录
	dbItems, err := c.messageRepo.GetAgentMessageItems(agentMessageID)
	if err != nil {
		return nil, fmt.Errorf("查询AgentMessageItems失败: %w", err)
	}

	// 转换为业务对象
	messageItems := make([]messages.MessageItem, 0, len(dbItems))
	for _, dbItem := range dbItems {
		messageItem, err := c.parseMessageItem(dbItem)
		if err != nil {
			// 记录错误但继续处理其他项
			// TODO: 添加日志记录
			continue
		}
		if messageItem != nil {
			messageItems = append(messageItems, messageItem)
		}
	}

	return messageItems, nil
}

func (c *MessageConverter) parseMessageItem(dbItem *repositories.AgentMessageItem) (messages.MessageItem, error) {
	if dbItem == nil || dbItem.Item == "" {
		return nil, fmt.Errorf("AgentMessageItem为空")
	}

	// 根据 ItemType 进行不同的解析
	switch dbItem.ItemType {
	case "message":
		var outputMessage messages.OutputMessage
		if err := json.Unmarshal([]byte(dbItem.Item), &outputMessage); err != nil {
			return nil, fmt.Errorf("解析OutputMessage失败: %w", err)
		}
		return &outputMessage, nil

	case "function_call":
		var functionCall messages.FunctionToolCall
		if err := json.Unmarshal([]byte(dbItem.Item), &functionCall); err != nil {
			return nil, fmt.Errorf("解析FunctionToolCall失败: %w", err)
		}
		return &functionCall, nil

	case "function_call_output":
		var functionOutput messages.FunctionToolCallOutput
		if err := json.Unmarshal([]byte(dbItem.Item), &functionOutput); err != nil {
			return nil, fmt.Errorf("解析FunctionToolCallOutput失败: %w", err)
		}
		return &functionOutput, nil

	case "reasoning":
		var reasoning messages.ReasoningItem
		if err := json.Unmarshal([]byte(dbItem.Item), &reasoning); err != nil {
			return nil, fmt.Errorf("解析ReasoningItem失败: %w", err)
		}
		return &reasoning, nil

	case "file_search_call":
		var fileSearch messages.FileSearchToolCall
		if err := json.Unmarshal([]byte(dbItem.Item), &fileSearch); err != nil {
			return nil, fmt.Errorf("解析FileSearchToolCall失败: %w", err)
		}
		return &fileSearch, nil

	case "computer_call":
		var computerCall messages.ComputerToolCall
		if err := json.Unmarshal([]byte(dbItem.Item), &computerCall); err != nil {
			return nil, fmt.Errorf("解析ComputerToolCall失败: %w", err)
		}
		return &computerCall, nil

	case "computer_call_output":
		var computerOutput messages.ComputerToolCallOutput
		if err := json.Unmarshal([]byte(dbItem.Item), &computerOutput); err != nil {
			return nil, fmt.Errorf("解析ComputerToolCallOutput失败: %w", err)
		}
		return &computerOutput, nil

	default:
		return nil, fmt.Errorf("不支持的MessageItem类型: %s", dbItem.ItemType)
	}
}

func (c *MessageConverter) parsePostContent(contentMap map[string]interface{}) (*messages.PostMessageContent, error) {
	title := utils.GetStringFromMap(contentMap, "title")

	// 解析富文本内容
	var content [][]messages.RichTextNode
	if contentData, ok := contentMap["content"]; ok {
		// 这里需要根据实际的富文本结构进行解析
		// 由于富文本结构比较复杂，这里提供一个基础实现
		contentBytes, err := json.Marshal(contentData)
		if err != nil {
			return nil, fmt.Errorf("解析富文本内容失败: %w", err)
		}

		if err := json.Unmarshal(contentBytes, &content); err != nil {
			// 如果解析失败，创建一个空的内容数组
			content = make([][]messages.RichTextNode, 0)
		}
	}

	return &messages.PostMessageContent{
		Title:   title,
		Content: content,
	}, nil
}

func (c *MessageConverter) parseStickerContent(contentMap map[string]interface{}) (*messages.StickerMessageContent, error) {
	content := &messages.StickerMessageContent{
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
