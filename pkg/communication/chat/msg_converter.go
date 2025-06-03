package chat

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
)

func (actor *ChatActor) convertMsgToInputItems(message *messages.Message) ([]messages.InputItem, error) {
	switch message.MsgType {
	case messages.MessageTypeText:
		return actor.convertTextMsg(message)
	case messages.MessageTypeImage:
		return actor.convertImageMsg(message)
	case messages.MessageTypeVideo:
		return actor.convertVideoMsg(message)
	case messages.MessageTypeFile:
		return actor.convertFileMsg(message)
	case messages.MessageTypeAudio:
		return actor.convertAudioMsg(message)
	case messages.MessageTypeAgent:
		return actor.convertAIChatMsg(message)
	case messages.MessageTypePost:
		return actor.convertPostMsg(message)
	case messages.MessageTypeSticker:
		return actor.convertStickerMsg(message)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", message.MsgType)
	}
}

func (actor *ChatActor) convertTextMsg(message *messages.Message) ([]messages.InputItem, error) {
	textContent, ok := message.Content.(*messages.TextMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 TextMessageContent 类型")
	}

	inputMessage := &messages.InputMessage{
		Type: "message",
		Role: "user",
		Content: []messages.InputContent{
			&messages.InputTextContent{
				Type: "input_text",
				Text: textContent.Text,
			},
		},
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertImageMsg(message *messages.Message) ([]messages.InputItem, error) {
	imageContent, ok := message.Content.(*messages.ImageMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 ImageMessageContent 类型")
	}

	var contents []messages.InputContent

	imageInputContent := &messages.InputImageContent{
		Type:   "input_image",
		FileID: imageContent.ImageCID, // 使用 CID 作为 FileID
		Detail: "auto",
	}
	contents = append(contents, imageInputContent)

	// 如果有替代文本，添加文本内容
	if imageContent.Alt != "" {
		textContent := &messages.InputTextContent{
			Type: "input_text",
			Text: fmt.Sprintf("图片描述: %s", imageContent.Alt),
		}
		contents = append(contents, textContent)
	}

	inputMessage := &messages.InputMessage{
		Type:    "message",
		Role:    "user",
		Content: contents,
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertVideoMsg(message *messages.Message) ([]messages.InputItem, error) {
	videoContent, ok := message.Content.(*messages.VideoMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 VideoMessageContent 类型")
	}

	var contents []messages.InputContent

	// 添加视频文件内容
	videoInputContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: videoContent.VideoCID,
	}
	contents = append(contents, videoInputContent)

	// 添加视频信息文本
	videoInfo := fmt.Sprintf("视频时长: %d秒", videoContent.Duration)
	if videoContent.Width > 0 && videoContent.Height > 0 {
		videoInfo += fmt.Sprintf(", 分辨率: %dx%d", videoContent.Width, videoContent.Height)
	}

	textContent := &messages.InputTextContent{
		Type: "input_text",
		Text: videoInfo,
	}
	contents = append(contents, textContent)

	// 如果有缩略图，也添加进去
	if videoContent.ThumbCID != "" {
		thumbContent := &messages.InputImageContent{
			Type:   "input_image",
			FileID: videoContent.ThumbCID,
			Detail: "low",
		}
		contents = append(contents, thumbContent)
	}

	inputMessage := &messages.InputMessage{
		Type:    "message",
		Role:    "user",
		Content: contents,
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertFileMsg(message *messages.Message) ([]messages.InputItem, error) {
	fileContent, ok := message.Content.(*messages.FileMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 FileMessageContent 类型")
	}

	var contents []messages.InputContent

	// 添加文件内容
	fileInputContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: fileContent.FileCID,
	}
	contents = append(contents, fileInputContent)

	// 添加文件信息文本
	fileInfo := fmt.Sprintf("文件名: %s, 大小: %d 字节, 类型: %s",
		fileContent.FileName, fileContent.Size, fileContent.MimeType)

	textContent := &messages.InputTextContent{
		Type: "input_text",
		Text: fileInfo,
	}
	contents = append(contents, textContent)

	inputMessage := &messages.InputMessage{
		Type:    "message",
		Role:    "user",
		Content: contents,
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertAudioMsg(message *messages.Message) ([]messages.InputItem, error) {
	audioContent, ok := message.Content.(*messages.AudioMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 AudioMessageContent 类型")
	}

	var contents []messages.InputContent

	// 添加音频文件内容
	audioInputContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: audioContent.AudioCID,
	}
	contents = append(contents, audioInputContent)

	// 添加音频信息文本
	audioInfo := fmt.Sprintf("音频时长: %d秒", audioContent.Duration)
	if audioContent.Transcript != "" {
		audioInfo += fmt.Sprintf(", 转录文本: %s", audioContent.Transcript)
	}

	textContent := &messages.InputTextContent{
		Type: "input_text",
		Text: audioInfo,
	}
	contents = append(contents, textContent)

	inputMessage := &messages.InputMessage{
		Type:    "message",
		Role:    "user",
		Content: contents,
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertAIChatMsg(message *messages.Message) ([]messages.InputItem, error) {
	aiChatContent, ok := message.Content.(*messages.AgentMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 AgentMessageContent 类型")
	}

	items := make([]messages.InputItem, 0, len(aiChatContent.AgentMessage.MessageItems))
	for _, item := range aiChatContent.AgentMessage.MessageItems {
		if inputItem, ok := item.(messages.InputItem); ok {
			items = append(items, inputItem)
		}
	}

	return items, nil
}

func (actor *ChatActor) convertPostMsg(message *messages.Message) ([]messages.InputItem, error) {
	postContent, ok := message.Content.(*messages.PostMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 PostMessageContent 类型")
	}

	var textContent string
	if postContent.Title != "" {
		textContent = fmt.Sprintf("标题: %s\n\n", postContent.Title)
	}

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

	inputMessage := &messages.EasyInputMessage{
		Type: "easy_message",
		Role: "user",
		Content: &messages.EasyInputTextContent{
			Text: fmt.Sprintf("分享了一个富文本帖子:\n%s", textContent),
		},
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertStickerMsg(message *messages.Message) ([]messages.InputItem, error) {
	stickerContent, ok := message.Content.(*messages.StickerMessageContent)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 StickerMessageContent 类型")
	}

	var contents []messages.InputContent

	// 添加表情包图片内容
	stickerInputContent := &messages.InputImageContent{
		Type:   "input_image",
		FileID: stickerContent.StickerCID,
		Detail: "low", // 表情包通常不需要高清晰度
	}
	contents = append(contents, stickerInputContent)

	// 添加表情包描述文本
	stickerInfo := "发送了一个表情包"
	if stickerContent.Alt != "" {
		stickerInfo += fmt.Sprintf(": %s", stickerContent.Alt)
	}
	if stickerContent.IsAnimated {
		stickerInfo += " (动画表情)"
	}

	textContent := &messages.InputTextContent{
		Type: "input_text",
		Text: stickerInfo,
	}
	contents = append(contents, textContent)

	inputMessage := &messages.InputMessage{
		Type:    "message",
		Role:    "user",
		Content: contents,
	}

	return []messages.InputItem{inputMessage}, nil
}
