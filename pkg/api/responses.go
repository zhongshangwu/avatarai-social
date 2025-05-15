package api

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/openai/openai-go/responses"

// 	"github.com/labstack/echo/v4"
// )

// type CreateResponseRequest struct {
// 	responses.ResponseNewParams `json:",inline"`
// 	Stream                      *bool `json:"stream,omitempty"`
// }

// func (a *AvatarAIAPI) Realtime(c echo.Context) error {
// 	var request CreateResponseRequest
// 	if err := c.Bind(&request); err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]interface{}{
// 			"error": map[string]string{
// 				"message": fmt.Sprintf("无法解析请求: %v", err),
// 			},
// 		})
// 	}

// 	stream := request.Stream != nil && *request.Stream

// 	// 3. 创建基础响应对象
// 	resp := responses.NewResponse(
// 		request.Model,
// 		request.MaxOutputTokens,
// 		request.Temperature,
// 		request.TopP,
// 	)

// 	if stream {
// 		return handleStreamResponse(c, resp, request)
// 	}

// 	// 非流式响应处理
// 	return handleNonStreamResponse(c, resp, request)
// }

// // handleStreamResponse 处理流式响应
// func handleStreamResponse(c echo.Context, resp *responses.Response, request responses.CreateRequest) error {
// 	// 设置SSE响应头
// 	c.Response().Header().Set("Content-Type", "text/event-stream")
// 	c.Response().Header().Set("Cache-Control", "no-cache")
// 	c.Response().Header().Set("Connection", "keep-alive")
// 	c.Response().WriteHeader(http.StatusOK)

// 	// 创建一个channel用于处理响应流
// 	streamCh := make(chan interface{})
// 	defer close(streamCh)

// 	// 开始异步处理AI响应
// 	go processAIResponseStream(streamCh, resp, request)

// 	// 创建刷新器，确保数据及时发送到客户端
// 	flusher, ok := c.Response().Writer.(http.Flusher)
// 	if !ok {
// 		return fmt.Errorf("流式响应需要HTTP Flusher")
// 	}

// 	for event := range streamCh {
// 		data, err := json.Marshal(event)
// 		if err != nil {
// 			break
// 		}

// 		// 使用SSE格式写入数据
// 		fmt.Fprintf(c.Response().Writer, "data: %s\n\n", data)
// 		flusher.Flush()
// 	}

// 	// 发送[DONE]标记
// 	fmt.Fprintf(c.Response().Writer, "data: [DONE]\n\n")
// 	flusher.Flush()

// 	return nil
// }

// func handleNonStreamResponse(c echo.Context, resp *responses.Response, request responses.CreateRequest) error {
// 	// 处理AI回答 (实际实现中需要调用您的AI模型服务)
// 	generateAIResponse(resp, request)

// 	resp.Status = responses.ResponseStatusCompleted

// 	return c.JSON(http.StatusOK, resp)
// }

// func processAIResponseStream(streamCh chan<- interface{}, resp *responses.Response, request responses.CreateRequest) {
// 	createdEvent := responses.StreamEvent{
// 		Type:     responses.StreamEventTypeResponseCreated,
// 		Response: resp,
// 	}
// 	streamCh <- createdEvent

// 	// 发送response.in_progress事件
// 	inProgressEvent := responses.StreamEvent{
// 		Type:     responses.StreamEventTypeResponseInProgress,
// 		Response: resp,
// 	}
// 	streamCh <- inProgressEvent

// 	// 模拟生成消息内容 (实际实现应替换为真实的AI调用)
// 	messageID := responses.NewID("msg")
// 	outputText := ""

// 	// 创建一个消息项
// 	message := &responses.MessageItem{
// 		ID:   messageID,
// 		Role: "assistant",
// 		Type: responses.TypeMessage,
// 		Content: []responses.ItemContent{
// 			{
// 				Type: responses.TypeOutputText,
// 				Text: outputText,
// 			},
// 		},
// 	}

// 	// 添加到输出
// 	resp.Output = append(resp.Output, message)

// 	// 模拟生成内容的过程
// 	sentences := []string{
// 		"这是一个基于OpenAI Responses API的实现。",
// 		"它支持流式和非流式响应。",
// 		"您可以根据需要进一步完善大模型交互逻辑。",
// 	}

// 	for _, sentence := range sentences {
// 		// 为每个句子模拟延迟
// 		time.Sleep(500 * time.Millisecond)

// 		// 更新文本内容
// 		outputText += sentence + " "

// 		// 发送增量更新
// 		deltaEvent := responses.ResponseTextDeltaEvent{
// 			ItemID:      messageID,
// 			OutputIndex: 0,
// 			Delta:       sentence + " ",
// 		}
// 		streamCh <- deltaEvent

// 		// 更新响应中的内容
// 		if len(resp.Output) > 0 {
// 			if msgItem, ok := resp.Output[0].(*responses.MessageItem); ok {
// 				if len(msgItem.Content) > 0 {
// 					if textContent, ok := msgItem.Content[0].(responses.ItemContent); ok && textContent.Type == responses.TypeOutputText {
// 						textContent.Text = outputText
// 						msgItem.Content[0] = textContent
// 					}
// 				}
// 			}
// 		}
// 	}

// 	// 标记响应完成
// 	resp.Status = responses.ResponseStatusCompleted
// 	completedEvent := responses.ResponseCompletedEvent{
// 		Response: *resp,
// 	}
// 	streamCh <- completedEvent
// }

// // generateAIResponse 生成AI响应 (非流式)
// func generateAIResponse(resp *responses.Response, request responses.CreateRequest) {
// 	// 创建一个消息项
// 	messageID := responses.NewID("msg")
// 	message := &responses.MessageItem{
// 		ID:   messageID,
// 		Role: "assistant",
// 		Type: responses.TypeMessage,
// 		Content: []responses.ItemContent{
// 			{
// 				Type: responses.TypeOutputText,
// 				Text: "这是一个基于OpenAI Responses API的实现。它支持流式和非流式响应。您可以根据需要进一步完善大模型交互逻辑。",
// 			},
// 		},
// 	}

// 	// 添加到输出
// 	resp.Output = append(resp.Output, message)

// 	// 这里可以添加更复杂的逻辑，如处理工具调用等
// 	// TODO: 实现与大模型的交互逻辑
// }
