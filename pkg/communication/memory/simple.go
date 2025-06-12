package memory

import (
	"fmt"

	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"gorm.io/gorm"
)

type SimpleThreadMemory struct {
	db       *gorm.DB
	roomID   string
	threadID string

	messageService *services.MessageService
}

func NewSimpleThreadMemory(db *gorm.DB, roomID string, threadID string) *SimpleThreadMemory {
	metaStore := &repositories.MetaStore{DB: db}
	messageService := services.NewMessageService(metaStore)
	return &SimpleThreadMemory{db: db, roomID: roomID, threadID: threadID, messageService: messageService}
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

	// 创建临时的 MessageRepository 来查询消息
	metaStore := &repositories.MetaStore{DB: m.db}
	messageRepo := repositories.NewMessageRepository(metaStore)

	dbMessages, err := messageRepo.ListMessagesHistory(m.roomID, m.threadID)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	chunks := make([]Chunk, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		message := m.messageService.Converter.DBToMessage(dbMsg)
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
