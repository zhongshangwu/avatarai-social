package converters

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/communication/memory"
	"github.com/zhongshangwu/avatarai-social/pkg/providers/llm"
)

type ChunkConverter interface {
	Convert(chunk memory.Chunk) (*llm.PromptMessage, error)
	SupportedType() memory.ChunkType
}

type ChunkToLLMConverter struct {
	converters map[memory.ChunkType]ChunkConverter
}

func NewChunkToLLMConverter() *ChunkToLLMConverter {
	converter := &ChunkToLLMConverter{
		converters: make(map[memory.ChunkType]ChunkConverter),
	}

	converter.RegisterConverter(&MessageChunkConverter{})

	return converter
}

func (c *ChunkToLLMConverter) RegisterConverter(converter ChunkConverter) {
	c.converters[converter.SupportedType()] = converter
}

func (c *ChunkToLLMConverter) Convert(chunk memory.Chunk) (*llm.PromptMessage, error) {
	converter, exists := c.converters[chunk.GetType()]
	if !exists {
		return nil, fmt.Errorf("不支持的 chunk 类型: %s", chunk.GetType())
	}

	return converter.Convert(chunk)
}

func (c *ChunkToLLMConverter) ConvertBatch(chunks []memory.Chunk) ([]*llm.PromptMessage, error) {
	var promptMessages []*llm.PromptMessage

	for _, chunk := range chunks {
		promptMsg, err := c.Convert(chunk)
		if err != nil {
			return nil, fmt.Errorf("转换 chunk [%s] 失败: %w", chunk.GetID(), err)
		}

		if promptMsg != nil {
			promptMessages = append(promptMessages, promptMsg)
		}
	}

	return promptMessages, nil
}
