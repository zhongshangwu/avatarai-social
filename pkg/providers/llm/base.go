package llm

import (
	"context"

	"github.com/zhongshangwu/avatarai-social/pkg/streams"
)

type LLM interface {
	ChatStream(ctx context.Context,
		model string,
		credentials map[string]interface{},
		promptMessages []PromptMessage,
		modelParameters map[string]interface{},
		tools []PromptMessageTool,
		stop []string) (streams.Stream[*LLMResultChunk], error)
	Chat(ctx context.Context,
		model string,
		credentials map[string]interface{},
		promptMessages []PromptMessage,
		modelParameters map[string]interface{},
		tools []PromptMessageTool,
		stop []string) (*LLMResult, error)
}
