package llm

type LLMResultChunkDelta struct {
	Index        int64                   `json:"index"`
	Message      *AssistantPromptMessage `json:"message"`
	FinishReason string                  `json:"finish_reason,omitempty"`
	Usage        *Usage                  `json:"usage,omitempty"`
}

type LLMResultChunk struct {
	Model             string              `json:"model"`
	PromptMessages    []*PromptMessage    `json:"prompt_messages"`
	SystemFingerprint string              `json:"system_fingerprint,omitempty"`
	Delta             LLMResultChunkDelta `json:"delta"`
}

type LLMResult struct {
	Model             string                  `json:"model"`
	PromptMessages    []*PromptMessage        `json:"prompt_messages"`
	Message           *AssistantPromptMessage `json:"message"`
	Usage             *Usage                  `json:"usage,omitempty"`
	SystemFingerprint string                  `json:"system_fingerprint,omitempty"`
}

type Usage struct {
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	TotalPrice       float64 `json:"total_price,omitempty"`
	Currency         string  `json:"currency,omitempty"`
}
