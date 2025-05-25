package llm

import (
	"context"
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type ModelManager struct {
	config *config.SocialConfig
}

func NewModelManager(config *config.SocialConfig) *ModelManager {
	return &ModelManager{
		config: config,
	}
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
