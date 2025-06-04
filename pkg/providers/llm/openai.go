package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/sirupsen/logrus"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type OpenAIClient struct{}

func (o *OpenAIClient) ChatStream(
	ctx context.Context,
	model string,
	credentials map[string]interface{},
	promptMessages []PromptMessage,
	modelParameters map[string]interface{},
	tools []PromptMessageTool,
	stop []string,
) (*streams.Stream[*LLMResultChunk], error) {
	resp, err := o.generate(
		ctx,
		model,
		credentials,
		promptMessages,
		modelParameters,
		tools,
		stop,
		true,
	)
	if err != nil {
		return nil, err
	}
	return resp.(*streams.Stream[*LLMResultChunk]), nil
}

func (o *OpenAIClient) Chat(
	ctx context.Context,
	model string,
	credentials map[string]interface{},
	promptMessages []PromptMessage,
	modelParameters map[string]interface{},
	tools []PromptMessageTool,
	stop []string,
) (*LLMResult, error) {
	resp, err := o.generate(
		ctx,
		model,
		credentials,
		promptMessages,
		modelParameters,
		tools,
		stop,
		true,
	)
	if err != nil {
		return nil, err
	}
	return resp.(*LLMResult), nil
}

func (o *OpenAIClient) generate(
	ctx context.Context,
	model string,
	credentials map[string]interface{},
	promptMessages []PromptMessage,
	modelParameters map[string]interface{},
	tools []PromptMessageTool,
	stop []string,
	stream bool,
) (interface{}, error) {
	options, err := o.toCredentialKwargs(credentials)
	if err != nil {
		return nil, err
	}

	logrus.Infof("openai options: %v", options)
	client := openai.NewClient(options...)

	promptMessages = o.clearIllegalPromptMessages(model, promptMessages)

	messages := o.convertPromptMessagesToOpenAIMessages(promptMessages)

	params := openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
	}

	if err := o.applyModelParameters(&params, modelParameters); err != nil {
		return nil, err
	}

	if len(tools) > 0 {
		toolsParams := make([]openai.ChatCompletionToolParam, 0, len(tools))
		for _, tool := range tools {
			toolsParams = append(toolsParams, openai.ChatCompletionToolParam{
				Type: "function",
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  tool.Parameters,
				},
			})
		}
		params.Tools = toolsParams
	}

	if len(stop) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfChatCompletionNewsStopArray: stop,
		}
	}
	blockAsStream := false
	if isOSeriesModel(model) {
		if maxTokens, ok := modelParameters["max_tokens"].(int); ok {
			delete(modelParameters, "max_tokens")
			modelParameters["max_completion_tokens"] = maxTokens
		}

		if matchesO1Model(model) && stream {
			stream = false
			blockAsStream = true
		}

		params.Stop = openai.ChatCompletionNewParamsStopUnion{}
	}

	if stream {
		stream := client.Chat.Completions.NewStreaming(ctx, params)
		return o.handleChatGenerateStreamResponse(ctx, model, stream, promptMessages, tools)
	} else {
		resp, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			return nil, err
		}

		result := o.handleChatGenerateResponse(ctx, model, resp, promptMessages, tools)

		if blockAsStream {
			return o.handleChatBlockAsStreamResponse(ctx, result, promptMessages, stop)
		}

		return result, nil
	}
}

func (o *OpenAIClient) handleChatGenerateResponse(
	ctx context.Context,
	model string,
	response *openai.ChatCompletion,
	promptMessages []PromptMessage,
	tools []PromptMessageTool,
) *LLMResult {
	if len(response.Choices) == 0 {
		return &LLMResult{
			Model:          model,
			PromptMessages: promptMessages,
			Message:        NewAssistantPromptMessage("", "", nil),
		}
	}

	assistantMessage := response.Choices[0].Message

	// Extract tool calls from response
	var toolCalls []ToolCall
	if len(assistantMessage.ToolCalls) > 0 {
		for _, toolCall := range assistantMessage.ToolCalls {
			if toolCall.Type == "function" {
				functionCall := ToolCall{
					ID:   toolCall.ID,
					Type: string(toolCall.Type),
					Function: ToolCallFunction{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				}
				toolCalls = append(toolCalls, functionCall)
			}
		}
	}

	assistantPromptMessage := NewAssistantPromptMessage(
		assistantMessage.Content,
		"",
		toolCalls,
	)

	promptTokens := o.numTokensFromMessages(model, promptMessages, tools)
	completionTokens := o.numTokensFromMessages(model, []PromptMessage{assistantPromptMessage.PromptMessage}, nil)

	usage := &Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}

	return &LLMResult{
		Model:             response.Model,
		PromptMessages:    promptMessages,
		Message:           assistantPromptMessage,
		Usage:             usage,
		SystemFingerprint: response.SystemFingerprint,
	}
}

func (o *OpenAIClient) handleChatGenerateStreamResponse(
	ctx context.Context,
	model string,
	openaiStream *ssestream.Stream[openai.ChatCompletionChunk],
	promptMessages []PromptMessage,
	tools []PromptMessageTool,
) (*streams.Stream[*LLMResultChunk], error) {
	respStream := streams.NewStream[*LLMResultChunk](ctx, 5)

	go func() {
		defer respStream.CloseSend()

		acc := openai.ChatCompletionAccumulator{}
		finalToolCalls := []ToolCall{}
		var usage *Usage

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// 使用非阻塞的方式检查 openaiStream
			if !openaiStream.Next() {
				break
			}

			chunk := openaiStream.Current()
			acc.AddChunk(chunk)

			// 检查并累积工具调用
			if toolCall, ok := acc.JustFinishedToolCall(); ok {
				functionCall := ToolCall{
					ID:   toolCall.Id,
					Type: "function",
					Function: ToolCallFunction{
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
					},
				}
				finalToolCalls = append(finalToolCalls, functionCall)
			}

			// 处理 usage 信息
			if chunk.Usage.TotalTokens > 0 {
				usage = &Usage{
					PromptTokens:     chunk.Usage.PromptTokens,
					CompletionTokens: chunk.Usage.CompletionTokens,
					TotalTokens:      chunk.Usage.TotalTokens,
				}
			}

			// 如果没有选择，继续处理下一个块
			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0]
			hasFinishReason := delta.FinishReason != ""

			// 提取增量内容
			content := delta.Delta.Content

			// 提取工具调用
			var toolCalls []ToolCall
			for _, toolCall := range delta.Delta.ToolCalls {
				if toolCall.Type == "function" {
					functionCall := ToolCall{
						ID:   toolCall.ID,
						Type: toolCall.Type,
						Function: ToolCallFunction{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					}
					toolCalls = append(toolCalls, functionCall)
				}
			}

			// 创建助手消息
			assistantPromptMessage := NewAssistantPromptMessage(content, "", toolCalls)

			// 发送结果块
			resultChunk := &LLMResultChunk{
				Model:             chunk.Model,
				PromptMessages:    promptMessages,
				SystemFingerprint: chunk.SystemFingerprint,
				Delta: LLMResultChunkDelta{
					Index:   delta.Index,
					Message: assistantPromptMessage,
				},
			}

			// 如果是最终块，添加完成原因和使用情况
			if hasFinishReason {
				resultChunk.Delta.FinishReason = delta.FinishReason
				resultChunk.Delta.Usage = usage
			}

			respStream.Send(resultChunk)
			logrus.Infof("send result chunk sleep 3 seconds....: %v", resultChunk.Delta.Message.Content)
			time.Sleep(3 * time.Second)
		}

		// 处理流错误
		if err := openaiStream.Err(); err != nil {
			respStream.SendError(err)
			fmt.Printf("Stream error: %v\n", err)
		}
	}()

	return respStream, nil
}

func (o *OpenAIClient) handleChatBlockAsStreamResponse(
	ctx context.Context,
	blockResult *LLMResult,
	promptMessages []PromptMessage,
	stop []string,
) (*streams.Stream[*LLMResultChunk], error) {
	respStream := streams.NewStream[*LLMResultChunk](ctx, 5)

	go func() {
		defer respStream.CloseSend()

		text := ""
		if content, ok := blockResult.Message.Content.(string); ok {
			text = content

			// Apply stop tokens if needed
			if len(stop) > 0 {
				text = o.enforceStopTokens(text, stop)
			}
		}

		respStream.Send(&LLMResultChunk{
			Model:             blockResult.Model,
			PromptMessages:    promptMessages,
			SystemFingerprint: blockResult.SystemFingerprint,
			Delta: LLMResultChunkDelta{
				Index:        0,
				Message:      blockResult.Message,
				FinishReason: "stop",
				Usage:        blockResult.Usage,
			},
		})
	}()

	return respStream, nil
}

func (o *OpenAIClient) convertPromptMessagesToOpenAIMessages(messages []PromptMessage) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, message := range messages {
		switch message.Role {
		case PromptMessageRoleUser:
			if content, ok := message.Content.(string); ok {
				msg := openai.UserMessage(content)
				if message.Name != "" {
					msg = openai.UserMessage(content)
				}
				result = append(result, msg)
			} else if contentList, ok := message.Content.([]PromptMessageContent); ok {
				var contents []openai.ChatCompletionContentPartUnionParam

				for _, item := range contentList {
					switch item.GetType() {
					case PromptMessageContentTypeText:
						textContent, _ := item.(*TextPromptMessageContent)
						contents = append(contents, openai.TextContentPart(textContent.Data))
					case PromptMessageContentTypeImage:
						imageContent, _ := item.(*ImagePromptMessageContent)
						contents = append(contents, openai.ImageContentPart(
							openai.ChatCompletionContentPartImageImageURLParam{
								URL:    imageContent.GetData(),
								Detail: string(imageContent.Detail),
							},
						))
					}
				}

				msg := openai.UserMessage(contents)
				if message.Name != "" {
					msg = openai.UserMessage(contents)
				}
				result = append(result, msg)
			}
		case PromptMessageRoleAssistant:
			if content, ok := message.Content.(string); ok {
				msg := openai.AssistantMessage(content)

				// Add tool calls if present
				// if assistantMessage, ok := interface{}(&message).(*AssistantPromptMessage); ok && len(assistantMessage.ToolCalls) > 0 {
				// 	toolCall := assistantMessage.ToolCalls[0]
				// 	msg = openai.AssistantMessage(
				// 		content,
				// 		openai.WithFunctionCall(toolCall.Function.Name, toolCall.Function.Arguments),
				// 	)
				// }

				result = append(result, msg)
			}
		case PromptMessageRoleSystem:
			if content, ok := message.Content.(string); ok {
				msg := openai.SystemMessage(content)
				result = append(result, msg)
			}
		case PromptMessageRoleTool:
			if content, ok := message.Content.(string); ok {
				if toolMessage, ok := interface{}(&message).(*ToolPromptMessage); ok {
					msg := openai.ToolMessage(content, toolMessage.ToolCallID)
					result = append(result, msg)
				}
			}
		}
	}

	return result
}

func (o *OpenAIClient) clearIllegalPromptMessages(model string, promptMessages []PromptMessage) []PromptMessage {
	if strings.Contains(model, "gpt-4-turbo") {
		userMessageCount := 0
		for _, msg := range promptMessages {
			if msg.Role == PromptMessageRoleUser {
				userMessageCount++
			}
		}

		if userMessageCount > 1 {
			// Convert multi-modal content to plain text
			for i := range promptMessages {
				if promptMessages[i].Role == PromptMessageRoleUser {
					if contentList, ok := promptMessages[i].Content.([]PromptMessageContent); ok {
						var textParts []string
						for _, item := range contentList {
							if textContent, ok := item.(*TextPromptMessageContent); ok {
								textParts = append(textParts, textContent.Data)
							} else if item.GetType() == PromptMessageContentTypeImage {
								textParts = append(textParts, "[IMAGE]")
							}
						}
						promptMessages[i].Content = strings.Join(textParts, "\n")
					}
				}
			}
		}
	}

	// O-series models compatibility
	if isOSeriesModel(model) {
		// Convert system messages to user messages for O-series models
		systemMessageCount := 0
		for _, msg := range promptMessages {
			if msg.Role == PromptMessageRoleSystem {
				systemMessageCount++
			}
		}

		if systemMessageCount > 0 {
			newPromptMessages := make([]PromptMessage, 0, len(promptMessages))
			for _, msg := range promptMessages {
				if msg.Role == PromptMessageRoleSystem {
					newMsg := PromptMessage{
						Role:    PromptMessageRoleUser,
						Content: msg.Content,
						Name:    msg.Name,
					}
					newPromptMessages = append(newPromptMessages, newMsg)
				} else {
					newPromptMessages = append(newPromptMessages, msg)
				}
			}
			promptMessages = newPromptMessages
		}
	}

	return promptMessages
}

func (o *OpenAIClient) applyModelParameters(params *openai.ChatCompletionNewParams, modelParams map[string]interface{}) error {
	for key, value := range modelParams {
		switch key {
		case "temperature":
			if temp, ok := value.(float64); ok {
				params.Temperature = openai.Float(temp)
			}
		case "top_p":
			if topP, ok := value.(float64); ok {
				params.TopP = openai.Float(topP)
			}
		case "max_tokens":
			if maxTokens, ok := value.(int64); ok {
				params.MaxTokens = openai.Int(maxTokens)
			}
		case "presence_penalty":
			if penalty, ok := value.(float64); ok {
				params.PresencePenalty = openai.Float(penalty)
			}
		case "frequency_penalty":
			if penalty, ok := value.(float64); ok {
				params.FrequencyPenalty = openai.Float(penalty)
			}
		case "response_format":
			// if format, ok := value.(string); ok {
			// 	if format == "json" {
			// 		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			// 			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			// 		}
			// 	}
			// } else if formatMap, ok := value.(map[string]interface{}); ok {
			// 	if formatType, ok := formatMap["type"].(string); ok {
			// 		if formatType == "json_object" {
			// 			params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			// 				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			// 			}
			// 		}
			// 	}
			// }
		case "seed":
			if seed, ok := value.(int64); ok {
				params.Seed = openai.Int(seed)
			}
		}
	}
	return nil
}

// numTokensFromString calculates the number of tokens in a string
func (o *OpenAIClient) numTokensFromString(model string, text string) int64 {
	// This is a simplified implementation - in a real implementation,
	// you'd use a proper tokenizer like tiktoken
	// For now, we'll use a rough approximation
	words := int64(len(strings.Fields(text)))
	return words * 4 / 3 // Rough approximation: 4 tokens per 3 words
}

// numTokensFromMessages calculates the number of tokens in messages
func (o *OpenAIClient) numTokensFromMessages(model string, messages []PromptMessage, tools []PromptMessageTool) int64 {
	// This is a simplified implementation
	var (
		tokensPerMessage int64 = 3
		tokensPerName    int64 = 1
		numTokens        int64 = 0
	)

	// Process messages
	for _, message := range messages {
		numTokens += tokensPerMessage

		// Add tokens for content
		if content, ok := message.Content.(string); ok {
			numTokens += o.numTokensFromString(model, content)
		} else if contentList, ok := message.Content.([]PromptMessageContent); ok {
			for _, item := range contentList {
				if textContent, ok := item.(*TextPromptMessageContent); ok {
					numTokens += o.numTokensFromString(model, textContent.Data)
				}
				// Note: Image tokens would require more complex calculation
			}
		}

		// Add tokens for name if present
		if message.Name != "" {
			numTokens += tokensPerName
		}

		// Add tokens for tool calls if present
		if assistantMessage, ok := interface{}(&message).(*AssistantPromptMessage); ok {
			for _, toolCall := range assistantMessage.ToolCalls {
				numTokens += o.numTokensFromString(model, toolCall.ID)
				numTokens += o.numTokensFromString(model, toolCall.Type)
				numTokens += o.numTokensFromString(model, toolCall.Function.Name)
				numTokens += o.numTokensFromString(model, toolCall.Function.Arguments)
			}
		}
	}

	// Add tokens for tools
	if len(tools) > 0 {
		numTokens += o.numTokensForTools(model, tools)
	}

	// Every reply is primed with <im_start>assistant
	numTokens += 3

	return numTokens
}

// numTokensForTools calculates the number of tokens used by tools
func (o *OpenAIClient) numTokensForTools(model string, tools []PromptMessageTool) int64 {
	var numTokens int64 = 0

	for _, tool := range tools {
		// Type and function tokens
		numTokens += o.numTokensFromString(model, "type")
		numTokens += o.numTokensFromString(model, "function")

		// Function object tokens
		numTokens += o.numTokensFromString(model, "name")
		numTokens += o.numTokensFromString(model, tool.Name)
		numTokens += o.numTokensFromString(model, "description")
		numTokens += o.numTokensFromString(model, tool.Description)

		// Parameters tokens
		numTokens += o.numTokensFromString(model, "parameters")

		if title, ok := tool.Parameters["title"].(string); ok {
			numTokens += o.numTokensFromString(model, "title")
			numTokens += o.numTokensFromString(model, title)
		}

		if typeStr, ok := tool.Parameters["type"].(string); ok {
			numTokens += o.numTokensFromString(model, "type")
			numTokens += o.numTokensFromString(model, typeStr)
		}

		if properties, ok := tool.Parameters["properties"].(map[string]interface{}); ok {
			numTokens += o.numTokensFromString(model, "properties")

			for key, value := range properties {
				numTokens += o.numTokensFromString(model, key)

				if propObj, ok := value.(map[string]interface{}); ok {
					for fieldKey, fieldValue := range propObj {
						numTokens += o.numTokensFromString(model, fieldKey)

						if fieldKey == "enum" {
							if enumValues, ok := fieldValue.([]interface{}); ok {
								for _, enumValue := range enumValues {
									numTokens += 3 // Approximate tokens for enum value
									if enumStr, ok := enumValue.(string); ok {
										numTokens += o.numTokensFromString(model, enumStr)
									}
								}
							}
						} else {
							if fieldStr, ok := fieldValue.(string); ok {
								numTokens += o.numTokensFromString(model, fieldStr)
							} else {
								// For non-string values, add approximate tokens
								numTokens += 3
							}
						}
					}
				}
			}
		}

		if required, ok := tool.Parameters["required"].([]interface{}); ok {
			numTokens += o.numTokensFromString(model, "required")

			for _, req := range required {
				numTokens += 3 // Approximate tokens for required field
				if reqStr, ok := req.(string); ok {
					numTokens += o.numTokensFromString(model, reqStr)
				}
			}
		}
	}

	return numTokens
}

// enforceStopTokens ensures that the text doesn't contain stop sequences
func (o *OpenAIClient) enforceStopTokens(text string, stop []string) string {
	result := text
	for _, stopSeq := range stop {
		if idx := strings.Index(result, stopSeq); idx >= 0 {
			result = result[:idx]
		}
	}
	return result
}

func isOSeriesModel(model string) bool {
	return strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}

func matchesO1Model(model string) bool {
	return strings.HasPrefix(model, "o1") ||
		strings.Contains(model, "-o1-")
}

func (o *OpenAIClient) toCredentialKwargs(credentials map[string]interface{}) ([]option.RequestOption, error) {
	options := make([]option.RequestOption, 0)

	if apiKey, ok := credentials["api_key"].(string); ok {
		logrus.Infof("openai api key: %s", apiKey)
		options = append(options, option.WithAPIKey(apiKey))
		options = append(options, option.WithHeader("X-Stepcast-Auth-Token", apiKey))
	}

	if baseURL, ok := credentials["base_url"].(string); ok {
		options = append(options, option.WithBaseURL(baseURL))
	}

	if organization, ok := credentials["organization"].(string); ok {
		options = append(options, option.WithOrganization(organization))
	}

	return options, nil
}
