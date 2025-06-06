package database

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

func InsertMessage(db *gorm.DB, message *Message) error {
	return db.Create(message).Error
}

func ListMessagesHistory(db *gorm.DB, roomID string, threadID string) ([]*Message, error) {
	var messages []*Message
	if err := db.Where("room_id = ? AND thread_id = ?", roomID, threadID).Order("created_at DESC").Limit(3).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// ListMessagesHistoryWithPagination 根据 before/after 参数获取历史消息
func ListMessagesHistoryWithPagination(db *gorm.DB, roomID string, threadID string, beforeMsgID string, beforeCount int, afterMsgID string, afterCount int) ([]*Message, error) {
	var messages []*Message

	// 构建基础查询
	query := db.Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false)

	// 如果指定了 beforeMsgID，获取该消息之前的消息
	if beforeMsgID != "" && beforeCount > 0 {
		// 先获取指定消息的创建时间
		var refMessage Message
		if err := db.Where("id = ?", beforeMsgID).First(&refMessage).Error; err != nil {
			return nil, err
		}

		// 获取该时间之前的消息
		beforeQuery := query.Where("created_at < ?", refMessage.CreatedAt).
			Order("created_at DESC").
			Limit(beforeCount)

		var beforeMessages []*Message
		if err := beforeQuery.Find(&beforeMessages).Error; err != nil {
			return nil, err
		}
		messages = append(messages, beforeMessages...)
	}

	// 如果指定了 afterMsgID，获取该消息之后的消息
	if afterMsgID != "" && afterCount > 0 {
		// 先获取指定消息的创建时间
		var refMessage Message
		if err := db.Where("id = ?", afterMsgID).First(&refMessage).Error; err != nil {
			return nil, err
		}

		// 获取该时间之后的消息
		afterQuery := query.Where("created_at > ?", refMessage.CreatedAt).
			Order("created_at ASC").
			Limit(afterCount)

		var afterMessages []*Message
		if err := afterQuery.Find(&afterMessages).Error; err != nil {
			return nil, err
		}

		// 将 after 消息按时间倒序插入到结果中
		for i := len(afterMessages) - 1; i >= 0; i-- {
			messages = append([]*Message{afterMessages[i]}, messages...)
		}
	}

	// 如果没有指定任何参数，返回最新的消息
	if beforeMsgID == "" && afterMsgID == "" {
		limit := 20 // 默认限制
		if beforeCount > 0 {
			limit = beforeCount
		} else if afterCount > 0 {
			limit = afterCount
		}

		if err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error; err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func InsertAgentMessage(db *gorm.DB, message *AgentMessage) error {
	return db.Create(message).Error
}

func GetAgentMessageByID(db *gorm.DB, agentMessageID string) (*AgentMessage, error) {
	var agentMessage AgentMessage
	if err := db.Where("id = ?", agentMessageID).First(&agentMessage).Error; err != nil {
		return nil, err
	}
	return &agentMessage, nil
}

func UpdateAgentMessage(db *gorm.DB, agentMessageID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return db.Model(&AgentMessage{}).Where("id = ?", agentMessageID).Updates(updates).Error
}

func UpdateAgentMessageStatus(db *gorm.DB, agentMessageID string, status string) error {
	return UpdateAgentMessage(db, agentMessageID, map[string]interface{}{
		"status": status,
	})
}

func UpdateAgentMessageWithError(db *gorm.DB, agentMessageID string, status string, errorInfo interface{}) error {
	errorStr, _ := json.Marshal(errorInfo)
	return UpdateAgentMessage(db, agentMessageID, map[string]interface{}{
		"status": status,
		"error":  string(errorStr),
	})
}

func UpdateAgentMessageWithUsage(db *gorm.DB, agentMessageID string, status string, usage interface{}, altText string) error {
	usageStr, _ := json.Marshal(usage)
	updates := map[string]interface{}{
		"status": status,
		"usage":  string(usageStr),
	}
	if altText != "" {
		updates["alt_text"] = altText
	}
	return UpdateAgentMessage(db, agentMessageID, updates)
}

func UpdateAgentMessageIncomplete(db *gorm.DB, agentMessageID string, interruptType int32, errorInfo interface{}, incompleteDetails interface{}) error {
	updates := map[string]interface{}{
		"status":         "incomplete",
		"interrupt_type": interruptType,
	}

	if errorInfo != nil {
		errorStr, _ := json.Marshal(errorInfo)
		updates["error"] = string(errorStr)
	}

	if incompleteDetails != nil {
		incompleteStr, _ := json.Marshal(incompleteDetails)
		updates["incomplete_details"] = string(incompleteStr)
	}

	return UpdateAgentMessage(db, agentMessageID, updates)
}

func InsertAgentMessageItem(db *gorm.DB, item *AgentMessageItem) error {
	return db.Create(item).Error
}

func UpdateAgentMessageItem(db *gorm.DB, itemID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return db.Model(&AgentMessageItem{}).Where("id = ?", itemID).Updates(updates).Error
}

func UpdateAgentMessageItemByPosition(db *gorm.DB, agentMessageID string, position int, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return db.Model(&AgentMessageItem{}).
		Where("agent_message_id = ? AND position = ?", agentMessageID, position).
		Updates(updates).Error
}
