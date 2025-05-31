package prompt

import (
	"fmt"
	"strings"

	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
	"github.com/zhongshangwu/avatarai-social/pkg/types/messages"
)

type LLMEntitiesTransform struct {
}

func (t *LLMEntitiesTransform) TransformInputItems(items []messages.InputItem) ([]llm.PromptMessage, error) {
	messageItems := make([]messages.MessageItem, len(items))
	for i, item := range items {
		messageItems[i] = item
	}
	return t.Transform(messageItems)
}

func (t *LLMEntitiesTransform) Transform(entities []messages.MessageItem) ([]llm.PromptMessage, error) {
	var promptMessages []llm.PromptMessage

	for _, entity := range entities {
		switch entity.GetType() {
		case "message":
			// 处理输入消息
			if inputMsg, ok := entity.(*messages.InputMessage); ok {
				promptMsg, err := t.convertInputMessage(inputMsg)
				if err != nil {
					return nil, fmt.Errorf("转换输入消息失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
			// 处理输出消息
			if outputMsg, ok := entity.(*messages.OutputMessage); ok {
				promptMsg, err := t.convertOutputMessage(outputMsg)
				if err != nil {
					return nil, fmt.Errorf("转换输出消息失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		case "tool_call", "function_call":
			// 处理函数工具调用
			if toolCall, ok := entity.(*messages.FunctionToolCall); ok {
				promptMsg, err := t.convertFunctionToolCall(toolCall)
				if err != nil {
					return nil, fmt.Errorf("转换函数工具调用失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		case "function_call_output":
			// 处理函数工具调用输出
			if toolOutput, ok := entity.(*messages.FunctionToolCallOutput); ok {
				promptMsg, err := t.convertFunctionToolCallOutput(toolOutput)
				if err != nil {
					return nil, fmt.Errorf("转换函数工具调用输出失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		case "item_reference":
			// 处理项目引用参数
			if itemRef, ok := entity.(*messages.ItemReferenceParam); ok {
				promptMsg, err := t.convertItemReference(itemRef)
				if err != nil {
					return nil, fmt.Errorf("转换项目引用失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		// 注意：以下类型在 llm.PromptMessage 中没有直接对应的结构，将被忽略或转换为文本消息
		case "reasoning":
			// 推理项目 - 转换为系统消息
			if reasoning, ok := entity.(*messages.ReasoningItem); ok {
				promptMsg := t.convertReasoningItem(reasoning)
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		case "file_search_call", "web_search_call", "code_interpreter_call", "computer_call":
			// 各种工具调用 - 转换为助手消息的工具调用
			promptMsg := t.convertToolCallToAssistantMessage(entity)
			if promptMsg != nil {
				promptMessages = append(promptMessages, *promptMsg)
			}
		case "computer_call_output":
			// 计算机工具调用输出
			if computerOutput, ok := entity.(*messages.ComputerToolCallOutput); ok {
				promptMsg, err := t.convertComputerToolCallOutput(computerOutput)
				if err != nil {
					return nil, fmt.Errorf("转换计算机工具调用输出失败: %w", err)
				}
				if promptMsg != nil {
					promptMessages = append(promptMessages, *promptMsg)
				}
			}
		default:
			// 未知类型，记录警告但继续处理
			fmt.Printf("警告: 未知的消息项类型: %s\n", entity.GetType())
		}
	}

	return promptMessages, nil
}

func (t *LLMEntitiesTransform) TransformTools(tools []map[string]interface{}) []llm.PromptMessageTool {
	var llmTools []llm.PromptMessageTool

	for _, tool := range tools {
		// 处理嵌套的 function 结构
		var functionDef map[string]interface{}
		if toolType, ok := tool["type"].(string); ok && toolType == "function" {
			if fn, ok := tool["function"].(map[string]interface{}); ok {
				functionDef = fn
			}
		} else {
			// 直接使用工具定义
			functionDef = tool
		}

		if name, ok := functionDef["name"].(string); ok {
			llmTool := llm.PromptMessageTool{
				Name: name,
			}

			if desc, ok := functionDef["description"].(string); ok {
				llmTool.Description = desc
			}

			if params, ok := functionDef["parameters"].(map[string]interface{}); ok {
				llmTool.Parameters = params
			}

			llmTools = append(llmTools, llmTool)
		}
	}

	return llmTools
}

func (t *LLMEntitiesTransform) convertInputMessage(inputMsg *messages.InputMessage) (*llm.PromptMessage, error) {
	// 转换角色
	var role llm.PromptMessageRole
	switch inputMsg.Role {
	case "user":
		role = llm.PromptMessageRoleUser
	case "assistant":
		role = llm.PromptMessageRoleAssistant
	case "system":
		role = llm.PromptMessageRoleSystem
	case "developer":
		role = llm.PromptMessageRoleSystem // developer 映射为 system
	default:
		role = llm.PromptMessageRoleUser
	}

	// 转换内容
	var content interface{}
	if len(inputMsg.Content) == 1 {
		// 单个内容项，如果是文本则直接使用字符串
		if textContent, ok := inputMsg.Content[0].(*messages.InputTextContent); ok {
			content = textContent.Text
		} else {
			content = t.convertInputContents(inputMsg.Content)
		}
	} else if len(inputMsg.Content) > 1 {
		// 多个内容项，使用数组
		content = t.convertInputContents(inputMsg.Content)
	} else {
		// 空内容
		content = ""
	}

	return &llm.PromptMessage{
		Role:    role,
		Content: content,
	}, nil
}

func (t *LLMEntitiesTransform) convertOutputMessage(outputMsg *messages.OutputMessage) (*llm.PromptMessage, error) {
	// 输出消息通常是助手角色
	role := llm.PromptMessageRoleAssistant

	// 转换内容
	var content interface{}
	if len(outputMsg.Content) == 1 {
		// 单个内容项
		if textContent, ok := outputMsg.Content[0].(*messages.OutputTextContent); ok {
			content = textContent.Text
		} else {
			content = t.convertOutputContents(outputMsg.Content)
		}
	} else if len(outputMsg.Content) > 1 {
		// 多个内容项
		content = t.convertOutputContents(outputMsg.Content)
	} else {
		// 空内容
		content = ""
	}

	return &llm.PromptMessage{
		Role:    role,
		Content: content,
	}, nil
}

func (t *LLMEntitiesTransform) convertFunctionToolCall(toolCall *messages.FunctionToolCall) (*llm.PromptMessage, error) {
	// 创建工具调用
	llmToolCall := llm.ToolCall{
		ID:   toolCall.ID,
		Type: "function",
		Function: llm.ToolCallFunction{
			Name:      toolCall.Name,
			Arguments: toolCall.Arguments,
		},
	}

	assistantMsg := &llm.AssistantPromptMessage{
		PromptMessage: llm.PromptMessage{
			Role:    llm.PromptMessageRoleAssistant,
			Content: nil, // 工具调用通常没有文本内容
		},
		ToolCalls: []llm.ToolCall{llmToolCall},
	}

	// 返回嵌入的 PromptMessage
	return &assistantMsg.PromptMessage, nil
}

func (t *LLMEntitiesTransform) convertFunctionToolCallOutput(toolOutput *messages.FunctionToolCallOutput) (*llm.PromptMessage, error) {
	toolMsg := &llm.ToolPromptMessage{
		PromptMessage: llm.PromptMessage{
			Role:    llm.PromptMessageRoleTool,
			Content: toolOutput.Output,
		},
		ToolCallID: toolOutput.CallID,
	}

	// 返回嵌入的 PromptMessage
	return &toolMsg.PromptMessage, nil
}

func (t *LLMEntitiesTransform) convertItemReference(itemRef *messages.ItemReferenceParam) (*llm.PromptMessage, error) {
	// 项目引用转换为用户消息，内容为引用信息
	content := fmt.Sprintf("[引用项目: %s]", itemRef.ID)

	return &llm.PromptMessage{
		Role:    llm.PromptMessageRoleUser,
		Content: content,
	}, nil
}

func (t *LLMEntitiesTransform) convertReasoningItem(reasoning *messages.ReasoningItem) *llm.PromptMessage {
	// 将推理内容转换为系统消息
	var summaryTexts []string
	for _, summary := range reasoning.Summary {
		summaryTexts = append(summaryTexts, summary.Text)
	}

	content := fmt.Sprintf("[推理过程] %s", strings.Join(summaryTexts, " "))

	return &llm.PromptMessage{
		Role:    llm.PromptMessageRoleSystem,
		Content: content,
	}
}

func (t *LLMEntitiesTransform) convertToolCallToAssistantMessage(entity messages.MessageItem) *llm.PromptMessage {
	var content string

	switch entity.GetType() {
	case "file_search_call":
		if fileSearch, ok := entity.(*messages.FileSearchToolCall); ok {
			content = fmt.Sprintf("[文件搜索] 查询: %s", strings.Join(fileSearch.Queries, ", "))
		}
	case "web_search_call":
		if webSearch, ok := entity.(*messages.WebSearchToolCall); ok {
			content = fmt.Sprintf("[Web搜索] ID: %s", webSearch.ID)
		}
	case "code_interpreter_call":
		if codeInterpreter, ok := entity.(*messages.CodeInterpreterToolCall); ok {
			content = fmt.Sprintf("[代码解释器] 代码: %s", codeInterpreter.Code)
		}
	case "computer_call":
		if computerCall, ok := entity.(*messages.ComputerToolCall); ok {
			content = fmt.Sprintf("[计算机操作] 类型: %s", computerCall.Action.GetType())
		}
	default:
		content = fmt.Sprintf("[工具调用] 类型: %s", entity.GetType())
	}

	return &llm.PromptMessage{
		Role:    llm.PromptMessageRoleAssistant,
		Content: content,
	}
}

func (t *LLMEntitiesTransform) convertComputerToolCallOutput(computerOutput *messages.ComputerToolCallOutput) (*llm.PromptMessage, error) {
	var content string

	switch computerOutput.Output.GetType() {
	case "screenshot":
		if screenshot, ok := computerOutput.Output.(*messages.ComputerScreenshotResult); ok {
			content = fmt.Sprintf("[计算机截图] URL: %s", screenshot.ImageURL)
		}
	case "action":
		if action, ok := computerOutput.Output.(*messages.ComputerActionResult); ok {
			content = fmt.Sprintf("[计算机操作结果] 成功: %t, 消息: %s", action.Success, action.Message)
		}
	default:
		content = fmt.Sprintf("[计算机输出] 类型: %s", computerOutput.Output.GetType())
	}

	toolMsg := &llm.ToolPromptMessage{
		PromptMessage: llm.PromptMessage{
			Role:    llm.PromptMessageRoleTool,
			Content: content,
		},
		ToolCallID: computerOutput.CallID,
	}

	// 返回嵌入的 PromptMessage
	return &toolMsg.PromptMessage, nil
}

func (t *LLMEntitiesTransform) convertInputContents(contents []messages.InputContent) []llm.PromptMessageContent {
	var promptContents []llm.PromptMessageContent

	for _, content := range contents {
		switch content.GetType() {
		case "input_text":
			textContent := content.(*messages.InputTextContent)
			promptContents = append(promptContents, &llm.TextPromptMessageContent{
				Type: llm.PromptMessageContentTypeText,
				Data: textContent.Text,
			})
		case "input_image":
			imageContent := content.(*messages.InputImageContent)
			var detail llm.ImageDetailLevel
			switch imageContent.Detail {
			case "high":
				detail = llm.ImageDetailLevelHigh
			case "low":
				detail = llm.ImageDetailLevelLow
			default:
				detail = llm.ImageDetailLevelHigh
			}

			promptContents = append(promptContents, &llm.ImagePromptMessageContent{
				MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
					Type: llm.PromptMessageContentTypeImage,
					URL:  imageContent.ImageURL,
				},
				Detail: detail,
			})
		case "input_file":
			fileContent := content.(*messages.InputFileContent)
			// 文件内容转换为文档类型
			promptContents = append(promptContents, &llm.DocumentPromptMessageContent{
				MultiModalPromptMessageContent: llm.MultiModalPromptMessageContent{
					Type: llm.PromptMessageContentTypeDocument,
					URL:  fmt.Sprintf("file://%s", fileContent.FileID), // 使用文件ID作为URL
				},
			})
		}
	}

	return promptContents
}

func (t *LLMEntitiesTransform) convertOutputContents(contents []messages.OutputContent) []llm.PromptMessageContent {
	var promptContents []llm.PromptMessageContent

	for _, content := range contents {
		switch content.GetType() {
		case "output_text":
			textContent := content.(*messages.OutputTextContent)
			promptContents = append(promptContents, &llm.TextPromptMessageContent{
				Type: llm.PromptMessageContentTypeText,
				Data: textContent.Text,
			})
		case "refusal":
			refusalContent := content.(*messages.RefusalContent)
			promptContents = append(promptContents, &llm.TextPromptMessageContent{
				Type: llm.PromptMessageContentTypeText,
				Data: fmt.Sprintf("[拒绝] %s", refusalContent.Refusal),
			})
		}
	}

	return promptContents
}
