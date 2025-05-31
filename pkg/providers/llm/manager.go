package llm

import (
	"context"
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type ModelManager struct {
	config        *config.SocialConfig
	toolExecutors map[string]ToolExecutor
}

// ToolExecutor 定义工具执行器接口
type ToolExecutor interface {
	Execute(ctx context.Context, arguments string) (string, error)
	GetName() string
	GetDescription() string
	GetParameters() map[string]interface{}
}

func NewModelManager(config *config.SocialConfig) *ModelManager {
	return &ModelManager{
		config:        config,
		toolExecutors: make(map[string]ToolExecutor),
	}
}

// RegisterTool 注册工具执行器
func (m *ModelManager) RegisterTool(executor ToolExecutor) {
	m.toolExecutors[executor.GetName()] = executor
}

// GetAvailableTools 获取可用的工具定义
func (m *ModelManager) GetAvailableTools() []PromptMessageTool {
	var tools []PromptMessageTool
	for _, executor := range m.toolExecutors {
		tools = append(tools, PromptMessageTool{
			Name:        executor.GetName(),
			Description: executor.GetDescription(),
			Parameters:  executor.GetParameters(),
		})
	}
	return tools
}

// ExecuteTool 执行工具
func (m *ModelManager) ExecuteTool(ctx context.Context, toolName, arguments string) (string, error) {
	executor, exists := m.toolExecutors[toolName]
	if !exists {
		return "", fmt.Errorf("tool not found: %s", toolName)
	}

	return executor.Execute(ctx, arguments)
}

func (m *ModelManager) ChatStream(
	ctx context.Context,
	promptMessages []PromptMessage,
	modelParameters map[string]interface{},
	tools []PromptMessageTool,
	stop []string,
) (*streams.Stream[*LLMResultChunk], error) {
	model := m.config.Avatar.LLM.Model
	provider := m.config.Avatar.LLM.Provider

	switch provider {
	case "openai":
		client := OpenAIClient{}
		return client.ChatStream(ctx, model, map[string]interface{}{
			"api_key":  m.config.Avatar.LLM.APIKey,
			"base_url": m.config.Avatar.LLM.APIURL,
		}, promptMessages, modelParameters, tools, stop)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// Chat 非流式聊天
func (m *ModelManager) Chat(
	ctx context.Context,
	promptMessages []PromptMessage,
	modelParameters map[string]interface{},
	tools []PromptMessageTool,
	stop []string,
) (*LLMResult, error) {
	model := m.config.Avatar.LLM.Model
	provider := m.config.Avatar.LLM.Provider

	switch provider {
	case "openai":
		client := OpenAIClient{}
		return client.Chat(ctx, model, map[string]interface{}{
			"api_key":  m.config.Avatar.LLM.APIKey,
			"base_url": m.config.Avatar.LLM.APIURL,
		}, promptMessages, modelParameters, tools, stop)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
