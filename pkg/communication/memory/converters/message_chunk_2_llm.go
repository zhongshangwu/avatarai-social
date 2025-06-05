package converters

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
)

type MessageChunkConverter struct{}

func (m *MessageChunkConverter) SupportedType() memory.ChunkType {
	return memory.ChunkTypeMessage
}

func (m *MessageChunkConverter) Convert(chunk memory.Chunk) (*llm.PromptMessage, error) {
	messageChunk, ok := chunk.(*memory.MessageChunk)
	if !ok {
		return nil, fmt.Errorf("chunk 类型断言失败，期望 *memory.MessageChunk")
	}

	message := messageChunk.Content
	if message == nil {
		return nil, fmt.Errorf("消息内容为空")
	}

	return m.convertMessage(message)
}

func (m *MessageChunkConverter) convertMessage(message *messages.Message) (*llm.PromptMessage, error) {
	// 根据发送者确定角色
	role := m.determineRole(message)

	// 根据消息类型转换内容
	content, err := m.convertContent(message)
	if err != nil {
		return nil, fmt.Errorf("转换消息内容失败: %w", err)
	}

	return &llm.PromptMessage{
		Role:    role,
		Content: content,
	}, nil
}

func (m *MessageChunkConverter) determineRole(message *messages.Message) llm.PromptMessageRole {
	// 根据消息类型和发送者确定角色
	switch message.MsgType {
	case messages.MessageTypeAgent:
		// AI 代理消息通常是助手角色
		return llm.PromptMessageRoleAssistant
	case messages.MessageTypeSystem:
		return llm.PromptMessageRoleSystem
	default:
		// 其他类型的消息通常是用户角色
		return llm.PromptMessageRoleUser
	}
}

func (m *MessageChunkConverter) convertContent(message *messages.Message) (interface{}, error) {
	switch message.MsgType {
	case messages.MessageTypeText:
		return m.convertTextContent(message)
	case messages.MessageTypeImage:
		return m.convertImageContent(message)
	case messages.MessageTypeVideo:
		return m.convertVideoContent(message)
	case messages.MessageTypeFile:
		return m.convertFileContent(message)
	case messages.MessageTypeAudio:
		return m.convertAudioContent(message)
	case messages.MessageTypeAgent:
		return m.convertAgentContent(message)
	case messages.MessageTypePost:
		return m.convertPostContent(message)
	case messages.MessageTypeSticker:
		return m.convertStickerContent(message)
	default:
		return fmt.Sprintf("[不支持的消息类型: %s]", message.MsgType), nil
	}
}

func (m *MessageChunkConverter) convertTextContent(message *messages.Message) (string, error) {
	textContent, ok := message.Content.(*messages.TextMessageContent)
	if !ok {
		return "", fmt.Errorf("消息内容类型断言失败，期望 *messages.TextMessageContent")
	}

	return textContent.Text, nil
}

func (m *MessageChunkConverter) convertImageContent(message *messages.Message) ([]llm.PromptMessageContent, error) {
	imageContent, ok := message.Content.(*messages.ImageMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息内容类型断言失败，期望 *messages.ImageMessageContent")
	}

	var contents []llm.PromptMessageContent

	// 添加图片内容
	imagePromptContent := &llm.ImagePromptMessageContent{
		MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
			Type: llm.PromptMessageContentTypeImage,
			URL:  imageContent.ImageURL,
		},
		Detail: llm.ImageDetailLevelHigh,
	}
	contents = append(contents, imagePromptContent)

	// 如果有替代文本，添加文本描述
	if imageContent.Alt != "" {
		textContent := &llm.TextPromptMessageContent{
			Type: llm.PromptMessageContentTypeText,
			Data: fmt.Sprintf("图片描述: %s", imageContent.Alt),
		}
		contents = append(contents, textContent)
	}

	return contents, nil
}

func (m *MessageChunkConverter) convertVideoContent(message *messages.Message) ([]llm.PromptMessageContent, error) {
	videoContent, ok := message.Content.(*messages.VideoMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息内容类型断言失败，期望 *messages.VideoMessageContent")
	}

	var contents []llm.PromptMessageContent

	// 添加视频文件内容
	videoPromptContent := &llm.VideoPromptMessageContent{
		MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
			Type: llm.PromptMessageContentTypeVideo,
			URL:  videoContent.VideoURL,
		},
	}
	contents = append(contents, videoPromptContent)

	// 添加视频信息文本
	videoInfo := fmt.Sprintf("视频时长: %d秒", videoContent.Duration)
	if videoContent.Width > 0 && videoContent.Height > 0 {
		videoInfo += fmt.Sprintf(", 分辨率: %dx%d", videoContent.Width, videoContent.Height)
	}

	textContent := &llm.TextPromptMessageContent{
		Type: llm.PromptMessageContentTypeText,
		Data: videoInfo,
	}
	contents = append(contents, textContent)

	// 如果有缩略图，也添加进去
	if videoContent.ThumbURL != "" {
		thumbContent := &llm.ImagePromptMessageContent{
			MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
				Type: llm.PromptMessageContentTypeImage,
				URL:  videoContent.ThumbURL,
			},
			Detail: llm.ImageDetailLevelLow,
		}
		contents = append(contents, thumbContent)
	}

	return contents, nil
}

func (m *MessageChunkConverter) convertFileContent(message *messages.Message) ([]llm.PromptMessageContent, error) {
	fileContent, ok := message.Content.(*messages.FileMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息内容类型断言失败，期望 *messages.FileMessageContent")
	}

	var contents []llm.PromptMessageContent

	// 添加文档内容
	docContent := &llm.DocumentPromptMessageContent{
		MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
			Type: llm.PromptMessageContentTypeDocument,
			URL:  fileContent.FileURL,
		},
	}
	contents = append(contents, docContent)

	// 添加文件信息文本
	fileInfo := fmt.Sprintf("文件名: %s, 大小: %d 字节, 类型: %s",
		fileContent.FileName, fileContent.Size, fileContent.MimeType)

	textContent := &llm.TextPromptMessageContent{
		Type: llm.PromptMessageContentTypeText,
		Data: fileInfo,
	}
	contents = append(contents, textContent)

	return contents, nil
}

func (m *MessageChunkConverter) convertAudioContent(message *messages.Message) ([]llm.PromptMessageContent, error) {
	audioContent, ok := message.Content.(*messages.AudioMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息内容类型断言失败，期望 *messages.AudioMessageContent")
	}

	var contents []llm.PromptMessageContent

	// 添加音频内容
	audioPromptContent := &llm.AudioPromptMessageContent{
		MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
			Type: llm.PromptMessageContentTypeAudio,
			URL:  audioContent.AudioURL,
		},
	}
	contents = append(contents, audioPromptContent)

	// 添加音频信息文本
	audioInfo := fmt.Sprintf("音频时长: %d秒", audioContent.Duration)
	if audioContent.Transcript != "" {
		audioInfo += fmt.Sprintf(", 转录文本: %s", audioContent.Transcript)
	}

	textContent := &llm.TextPromptMessageContent{
		Type: llm.PromptMessageContentTypeText,
		Data: audioInfo,
	}
	contents = append(contents, textContent)

	return contents, nil
}

func (m *MessageChunkConverter) convertAgentContent(message *messages.Message) (string, error) {
	agentContent, ok := message.Content.(*messages.AgentMessageContent)
	if !ok {
		return "", fmt.Errorf("消息内容类型断言失败，期望 *messages.AgentMessageContent")
	}

	// 提取 AI 代理消息的文本内容
	return agentContent.AgentMessage.AltText, nil
}

func (m *MessageChunkConverter) convertPostContent(message *messages.Message) (string, error) {
	postContent, ok := message.Content.(*messages.PostMessageContent)
	if !ok {
		return "", fmt.Errorf("消息内容类型断言失败，期望 *messages.PostMessageContent")
	}

	var textContent string
	if postContent.Title != "" {
		textContent = fmt.Sprintf("标题: %s\n\n", postContent.Title)
	}

	// 简化的富文本内容提取
	for _, row := range postContent.Content {
		for _, node := range row {
			switch node.GetTag() {
			case messages.PostNodeText:
				if textNode, ok := node.(*messages.RichTextNodeText); ok {
					textContent += textNode.Text
				}
			case messages.PostNodeLink:
				if linkNode, ok := node.(*messages.RichTextNodeLink); ok {
					textContent += fmt.Sprintf("[%s](%s)", linkNode.Text, linkNode.Href)
				}
			case messages.PostNodeAt:
				if atNode, ok := node.(*messages.RichTextNodeAt); ok {
					textContent += fmt.Sprintf("@%s", atNode.UserID)
				}
			case messages.PostNodeImage:
				if imgNode, ok := node.(*messages.RichTextNodeImage); ok {
					textContent += fmt.Sprintf("[图片: %s]", imgNode.ImageKey)
				}
			case messages.PostNodeMedia:
				if videoNode, ok := node.(*messages.RichTextNodeVideo); ok {
					textContent += fmt.Sprintf("[视频: %s]", videoNode.FileKey)
				}
			case messages.PostNodeEmotion:
				if emotionNode, ok := node.(*messages.RichTextNodeEmotion); ok {
					textContent += fmt.Sprintf("[表情: %s]", emotionNode.EmojiType)
				}
			case messages.PostNodeHr:
				textContent += "\n---\n"
			case messages.PostNodeCodeBlock:
				if codeNode, ok := node.(*messages.RichTextNodeCodeBlock); ok {
					if codeNode.Language != "" {
						textContent += fmt.Sprintf("\n```%s\n%s\n```\n", codeNode.Language, codeNode.Text)
					} else {
						textContent += fmt.Sprintf("\n```\n%s\n```\n", codeNode.Text)
					}
				}
			case messages.PostNodeMarkdown:
				if mdNode, ok := node.(*messages.RichTextNodeMarkdown); ok {
					textContent += mdNode.Text
				}
			}
		}
		textContent += "\n" // 每行结束后添加换行
	}

	return fmt.Sprintf("分享了一个富文本帖子:\n%s", textContent), nil
}

func (m *MessageChunkConverter) convertStickerContent(message *messages.Message) ([]llm.PromptMessageContent, error) {
	stickerContent, ok := message.Content.(*messages.StickerMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息内容类型断言失败，期望 *messages.StickerMessageContent")
	}

	var contents []llm.PromptMessageContent

	// 添加表情包图片内容
	stickerPromptContent := &llm.ImagePromptMessageContent{
		MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
			Type: llm.PromptMessageContentTypeImage,
			URL:  stickerContent.StickerURL,
		},
		Detail: llm.ImageDetailLevelLow, // 表情包通常不需要高清晰度
	}
	contents = append(contents, stickerPromptContent)

	// 添加表情包描述文本
	stickerInfo := "发送了一个表情包"
	if stickerContent.Alt != "" {
		stickerInfo += fmt.Sprintf(": %s", stickerContent.Alt)
	}
	if stickerContent.IsAnimated {
		stickerInfo += " (动画表情)"
	}

	textContent := &llm.TextPromptMessageContent{
		Type: llm.PromptMessageContentTypeText,
		Data: stickerInfo,
	}
	contents = append(contents, textContent)

	return contents, nil
}
func ChunkToLLM(chunk memory.Chunk) *llm.PromptMessage {
	converter := NewChunkToLLMConverter()
	promptMsg, err := converter.Convert(chunk)
	if err != nil {
		// 返回错误信息作为文本消息
		return &llm.PromptMessage{
			Role:    llm.PromptMessageRoleUser,
			Content: fmt.Sprintf("[转换错误: %s]", err.Error()),
		}
	}
	return promptMsg
}
