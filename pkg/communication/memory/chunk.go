package memory

import "github.com/zhongshangwu/avatarai-social/pkg/communication/messages"

type Chunk interface {
	GetID() string
	GetType() ChunkType
}

type ChunkType string

const (
	ChunkTypeMessage ChunkType = "message" // 聊天消息
)

type MessageChunk struct {
	ID       string
	Metadata map[string]interface{}
	Content  *messages.Message
}

func (c *MessageChunk) GetID() string {
	return c.ID
}

func (c *MessageChunk) GetType() ChunkType {
	return ChunkTypeMessage
}
