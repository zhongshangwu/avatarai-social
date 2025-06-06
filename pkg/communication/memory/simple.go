package memory

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"gorm.io/gorm"
)

type SimpleThreadMemory struct {
	db       *gorm.DB
	roomID   string
	threadID string
}

func NewSimpleThreadMemory(db *gorm.DB, roomID string, threadID string) *SimpleThreadMemory {
	return &SimpleThreadMemory{db: db, roomID: roomID, threadID: threadID}
}

func (m *SimpleThreadMemory) Write(chunk Chunk) error {
	// 对于简单实现，我们不需要额外写入，因为消息已经通过正常流程存储到数据库
	// 这里可以添加一些额外的索引或缓存逻辑
	return nil
}

func (m *SimpleThreadMemory) Retrieve(query Chunk) ([]Chunk, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}

	dbMessages, err := database.ListMessagesHistory(m.db, m.roomID, m.threadID)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	chunks := make([]Chunk, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		message := messages.DBToMessage(dbMsg)
		chunk := &MessageChunk{
			ID: dbMsg.ID,
			Metadata: map[string]interface{}{
				"sender_id":   message.SenderID,
				"receiver_id": message.ReceiverID,
				"sender_at":   message.SenderAt,
				"created_at":  message.CreatedAt,
				"msg_type":    message.MsgType,
			},
			Content: message,
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (m *SimpleThreadMemory) Close() error {
	// 对于简单实现，不需要特殊的关闭操作
	return nil
}
