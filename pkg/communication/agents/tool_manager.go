package agents

// type ToolEngine struct {
// 	Sender     *ChatEventSender
// 	LLMManager *llm.ModelManager
// }

// func NewToolEngine(sender *ChatEventSender, llmManager *llm.ModelManager) *ToolEngine {
// 	return &ToolEngine{
// 		Sender:     sender,
// 		LLMManager: llmManager,
// 	}
// }

// func (a *ToolEngine) HandleToolCalls(ctx context.Context, toolCalls []llm.ToolCall) error {
// 	for _, toolCall := range toolCalls {
// 		if toolCall.Type == "function" {
// 			if err := a.handleFunctionCall(ctx, toolCall); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (a *ToolEngine) handleFunctionCall(ctx context.Context, toolCall llm.ToolCall) error {
// 	// 创建函数工具调用
// 	functionCall := &messages.FunctionToolCall{
// 		ID:        toolCall.ID,
// 		Type:      "tool_call",
// 		Name:      toolCall.Function.Name,
// 		Status:    "in_progress",
// 		Arguments: toolCall.Function.Arguments,
// 	}

// 	// 添加到消息项
// 	ctx.AIChatMessage.MessageItems = append(ctx.AIChatMessage.MessageItems, functionCall)
// 	currentIndex := len(ctx.AIChatMessage.MessageItems) - 1

// 	// 发送输出项添加事件
// 	if err := a.Sender.sendOutputItemAdded(ctx, currentIndex, functionCall); err != nil {
// 		return err
// 	}

// 	// 发送函数调用参数增量事件（如果有参数）
// 	if functionCall.Arguments != "" {
// 		if err := a.Sender.sendFunctionCallArgumentsDelta(ctx, functionCall.ID, currentIndex, functionCall.Arguments); err != nil {
// 			return err
// 		}

// 		// 发送函数调用参数完成事件
// 		if err := a.Sender.sendFunctionCallArgumentsDone(ctx, functionCall.ID, currentIndex, functionCall.Arguments); err != nil {
// 			return err
// 		}
// 	}

// 	// 异步执行工具调用
// 	go a.executeFunctionCall(ctx, functionCall, currentIndex)

// 	return nil
// }

// func (a *ToolEngine) executeFunctionCall(ctx context.Context, functionCall *messages.FunctionToolCall, outputIndex int) {
// 	logrus.Infof("开始执行工具: %s", functionCall.Name)

// 	// 执行工具
// 	result, err := a.LLMManager.ExecuteTool(ctx.Context, functionCall.Name, functionCall.Arguments)

// 	if err != nil {
// 		logrus.Errorf("执行工具 %s 失败: %v", functionCall.Name, err)
// 		functionCall.Status = "failed"

// 		// 可以在这里发送工具调用失败事件
// 		// 目前先记录错误
// 	} else {
// 		logrus.Infof("工具 %s 执行成功，结果: %s", functionCall.Name, result)
// 		functionCall.Status = "completed"

// 		// 创建函数调用输出
// 		functionOutput := &messages.FunctionToolCallOutput{
// 			ID:     uuid.New().String(),
// 			Type:   "function_call_output",
// 			CallID: functionCall.ID,
// 			Output: result,
// 			Status: "completed",
// 		}

// 		// 添加到消息项
// 		ctx.AIChatMessage.MessageItems = append(ctx.AIChatMessage.MessageItems, functionOutput)
// 		outputOutputIndex := len(ctx.AIChatMessage.MessageItems) - 1

// 		// 发送函数调用输出添加事件
// 		if err := a.Sender.sendOutputItemAdded(ctx, outputOutputIndex, functionOutput); err != nil {
// 			logrus.Errorf("发送函数调用输出添加事件失败: %v", err)
// 		}

// 		// 发送函数调用输出完成事件
// 		if err := a.Sender.sendOutputItemDone(ctx, outputOutputIndex, functionOutput); err != nil {
// 			logrus.Errorf("发送函数调用输出完成事件失败: %v", err)
// 		}
// 	}

// 	// 发送函数调用完成事件
// 	if err := a.Sender.sendOutputItemDone(ctx, outputIndex, functionCall); err != nil {
// 		logrus.Errorf("发送函数调用完成事件失败: %v", err)
// 	}
// }
