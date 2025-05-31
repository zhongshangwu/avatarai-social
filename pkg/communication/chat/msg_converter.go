package chat

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

func (actor *ChatActor) convertMsgToInputItems(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	switch sendMsgEvent.MsgType {
	case messages.MessageTypeText:
		return actor.convertTextMsg(sendMsgEvent)
	case messages.MessageTypeImage:
		return actor.convertImageMsg(sendMsgEvent)
	case messages.MessageTypeVideo:
		return actor.convertVideoMsg(sendMsgEvent)
	case messages.MessageTypeFile:
		return actor.convertFileMsg(sendMsgEvent)
	case messages.MessageTypeAudio:
		return actor.convertAudioMsg(sendMsgEvent)
	case messages.MessageTypeAIChat:
		return actor.convertAIChatMsg(sendMsgEvent)
	case messages.MessageTypePost:
		return actor.convertPostMsg(sendMsgEvent)
	case messages.MessageTypeSticker:
		return actor.convertStickerMsg(sendMsgEvent)
	default:
		return nil, fmt.Errorf("不支持的消息类型: %s", sendMsgEvent.MsgType)
	}
}

func (actor *ChatActor) convertTextMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	textBody, ok := sendMsgEvent.Body.(*messages.TextMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 TextMsgBody 类型")
	}

	inputMessage := &messages.EasyInputMessage{
		Type: "easy_message",
		Role: "user",
		Content: &messages.EasyInputTextContent{
			Text: textBody.Text,
		},
	}

	return []messages.InputItem{inputMessage}, nil
}

func (actor *ChatActor) convertImageMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	imageBody, ok := sendMsgEvent.Body.(*messages.ImageMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 ImageMsgBody 类型")
	}

	// 构建包含图片和可选文本的内容
	var contents []messages.InputContent

	// 添加图片内容
	imageContent := &messages.InputImageContent{
		Type:   "input_image",
		FileID: imageBody.ImageCID, // 使用 CID 作为 FileID
		Detail: "auto",
	}
	contents = append(contents, imageContent)

	// 如果有替代文本，添加文本内容
	if imageBody.Alt != "" {
		textContent := &messages.InputTextContent{
			Type: "input_text",
			Text: fmt.Sprintf("图片描述: %s", imageBody.Alt),
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

func (actor *ChatActor) convertVideoMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	videoBody, ok := sendMsgEvent.Body.(*messages.VideoMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 VideoMsgBody 类型")
	}

	var contents []messages.InputContent

	// 添加视频文件内容
	videoContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: videoBody.VideoCID,
	}
	contents = append(contents, videoContent)

	// 添加视频信息文本
	videoInfo := fmt.Sprintf("视频时长: %d秒", videoBody.Duration)
	if videoBody.Width > 0 && videoBody.Height > 0 {
		videoInfo += fmt.Sprintf(", 分辨率: %dx%d", videoBody.Width, videoBody.Height)
	}

	textContent := &messages.InputTextContent{
		Type: "input_text",
		Text: videoInfo,
	}
	contents = append(contents, textContent)

	// 如果有缩略图，也添加进去
	if videoBody.ThumbCID != "" {
		thumbContent := &messages.InputImageContent{
			Type:   "input_image",
			FileID: videoBody.ThumbCID,
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

func (actor *ChatActor) convertFileMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	fileBody, ok := sendMsgEvent.Body.(*messages.FileMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 FileMsgBody 类型")
	}

	var contents []messages.InputContent

	// 添加文件内容
	fileContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: fileBody.FileCID,
	}
	contents = append(contents, fileContent)

	// 添加文件信息文本
	fileInfo := fmt.Sprintf("文件名: %s, 大小: %d 字节, 类型: %s",
		fileBody.FileName, fileBody.Size, fileBody.MimeType)

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

func (actor *ChatActor) convertAudioMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	audioBody, ok := sendMsgEvent.Body.(*messages.AudioMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 AudioMsgBody 类型")
	}

	var contents []messages.InputContent

	// 添加音频文件内容
	audioContent := &messages.InputFileContent{
		Type:   "input_file",
		FileID: audioBody.AudioCID,
	}
	contents = append(contents, audioContent)

	// 添加音频信息文本
	audioInfo := fmt.Sprintf("音频时长: %d秒", audioBody.Duration)
	if audioBody.Transcript != "" {
		audioInfo += fmt.Sprintf(", 转录文本: %s", audioBody.Transcript)
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

func (actor *ChatActor) convertAIChatMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	aiChatBody, ok := sendMsgEvent.Body.(*messages.AIChatMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 AIChatMsgBody 类型")
	}

	return aiChatBody.MessageItems, nil
}

func (actor *ChatActor) convertPostMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	postBody, ok := sendMsgEvent.Body.(*messages.PostMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 PostMsgBody 类型")
	}

	// 将富文本内容转换为纯文本
	var textContent string
	if postBody.Title != "" {
		textContent = fmt.Sprintf("标题: %s\n\n", postBody.Title)
	}

	// 处理富文本内容
	for _, row := range postBody.Content {
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

func (actor *ChatActor) convertStickerMsg(sendMsgEvent *messages.SendMsgEvent) ([]messages.InputItem, error) {
	stickerBody, ok := sendMsgEvent.Body.(*messages.StickerMsgBody)
	if !ok {
		return nil, fmt.Errorf("消息体类型转换失败，非 StickerMsgBody 类型")
	}

	var contents []messages.InputContent

	// 添加表情包图片内容
	stickerContent := &messages.InputImageContent{
		Type:   "input_image",
		FileID: stickerBody.StickerCID,
		Detail: "low", // 表情包通常不需要高清晰度
	}
	contents = append(contents, stickerContent)

	// 添加表情包描述文本
	stickerInfo := "发送了一个表情包"
	if stickerBody.Alt != "" {
		stickerInfo += fmt.Sprintf(": %s", stickerBody.Alt)
	}
	if stickerBody.IsAnimated {
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
