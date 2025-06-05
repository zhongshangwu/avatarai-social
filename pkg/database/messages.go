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
