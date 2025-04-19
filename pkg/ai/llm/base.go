package llm

import "context"

type LLM interface {
	Generate(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
}
