package repositories

import (
	"encoding/json"
	"time"
)

type MessageRepository struct {
	metaStore *MetaStore
}

func NewMessageRepository(metastore *MetaStore) *MessageRepository {
	return &MessageRepository{
		metaStore: metastore,
	}
}

// Message 相关操作
func (r *MessageRepository) InsertMessage(message *Message) error {
	return r.metaStore.DB.Create(message).Error
}

func (r *MessageRepository) GetMessageByID(id string) (*Message, error) {
	var message Message
	if err := r.metaStore.DB.Where("id = ?", id).First(&message).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) ListMessagesHistory(roomID string, threadID string) ([]*Message, error) {
	var messages []*Message
	if err := r.metaStore.DB.Where("room_id = ? AND thread_id = ?", roomID, threadID).Order("created_at DESC").Limit(3).Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// MessagePaginationResult 分页查询结果
type MessagePaginationResult struct {
	Messages []*Message `json:"messages"`
	HasMore  bool       `json:"hasMore"`
	FirstID  string     `json:"firstId,omitempty"`
	LastID   string     `json:"lastId,omitempty"`
}

func (r *MessageRepository) ListMessagesHistoryWithPagination(roomID string, threadID string, beforeMsgID string, beforeCount int, afterMsgID string, afterCount int) (*MessagePaginationResult, error) {
	var messages []*Message
	var hasMore bool

	// 构建基础查询
	baseQuery := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false)

	// 如果指定了 beforeMsgID，获取该消息之前的消息
	if beforeMsgID != "" && beforeCount > 0 {
		// 先获取指定消息的创建时间
		var refMessage Message
		if err := r.metaStore.DB.Where("id = ?", beforeMsgID).First(&refMessage).Error; err != nil {
			return nil, err
		}

		// 多查询一条来判断是否还有更多
		query := baseQuery.Where("created_at < ?", refMessage.CreatedAt).
			Order("created_at DESC").
			Limit(beforeCount + 1)

		var beforeMessages []*Message
		if err := query.Find(&beforeMessages).Error; err != nil {
			return nil, err
		}

		// 判断是否有更多数据
		if len(beforeMessages) > beforeCount {
			hasMore = true
			beforeMessages = beforeMessages[:beforeCount] // 移除多查询的那一条
		}

		messages = append(messages, beforeMessages...)
	} else if afterMsgID != "" && afterCount > 0 {
		// 如果指定了 afterMsgID，获取该消息之后的消息
		var refMessage Message
		if err := r.metaStore.DB.Where("id = ?", afterMsgID).First(&refMessage).Error; err != nil {
			return nil, err
		}

		// 多查询一条来判断是否还有更多
		query := baseQuery.Where("created_at > ?", refMessage.CreatedAt).
			Order("created_at ASC").
			Limit(afterCount + 1)

		var afterMessages []*Message
		if err := query.Find(&afterMessages).Error; err != nil {
			return nil, err
		}

		// 判断是否有更多数据
		if len(afterMessages) > afterCount {
			hasMore = true
			afterMessages = afterMessages[:afterCount] // 移除多查询的那一条
		}

		// 将 after 消息按时间倒序插入到结果中
		for i := len(afterMessages) - 1; i >= 0; i-- {
			messages = append([]*Message{afterMessages[i]}, messages...)
		}
	} else {
		// 如果没有指定任何参数，返回最新的消息
		limit := 20 // 默认限制
		if beforeCount > 0 {
			limit = beforeCount
		} else if afterCount > 0 {
			limit = afterCount
		}

		// 多查询一条来判断是否还有更多
		query := baseQuery.Order("created_at DESC").Limit(limit + 1)

		if err := query.Find(&messages).Error; err != nil {
			return nil, err
		}

		// 判断是否有更多数据
		if len(messages) > limit {
			hasMore = true
			messages = messages[:limit] // 移除多查询的那一条
		}
	}

	// 构建结果
	result := &MessagePaginationResult{
		Messages: messages,
		HasMore:  hasMore,
	}

	// 设置首尾ID
	if len(messages) > 0 {
		result.FirstID = messages[0].ID
		result.LastID = messages[len(messages)-1].ID
	}

	return result, nil
}

// 保留原有方法以保持兼容性，但标记为废弃
// Deprecated: 使用 ListMessagesHistoryWithPagination 替代
func (r *MessageRepository) ListMessagesHistoryWithPaginationOld(roomID string, threadID string, beforeMsgID string, beforeCount int, afterMsgID string, afterCount int) ([]*Message, error) {
	result, err := r.ListMessagesHistoryWithPagination(roomID, threadID, beforeMsgID, beforeCount, afterMsgID, afterCount)
	if err != nil {
		return nil, err
	}
	return result.Messages, nil
}

func (r *MessageRepository) UpdateMessage(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&Message{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MessageRepository) DeleteMessage(id string) error {
	return r.metaStore.DB.Model(&Message{}).Where("id = ?", id).Update("deleted", true).Error
}

// Room 相关操作
func (r *MessageRepository) CreateRoom(room *Room) error {
	return r.metaStore.DB.Create(room).Error
}

func (r *MessageRepository) GetRoomByID(id string) (*Room, error) {
	var room Room
	if err := r.metaStore.DB.Where("id = ?", id).First(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *MessageRepository) UpdateRoom(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&Room{}).Where("id = ?", id).Updates(updates).Error
}

// Thread 相关操作
func (r *MessageRepository) CreateThread(thread *Thread) error {
	return r.metaStore.DB.Create(thread).Error
}

func (r *MessageRepository) GetThreadByID(id string) (*Thread, error) {
	var thread Thread
	if err := r.metaStore.DB.Where("id = ?", id).First(&thread).Error; err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *MessageRepository) UpdateThread(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&Thread{}).Where("id = ?", id).Updates(updates).Error
}

// UserRoomStatus 相关操作
func (r *MessageRepository) CreateUserRoomStatus(status *UserRoomStatus) error {
	return r.metaStore.DB.Create(status).Error
}

func (r *MessageRepository) GetUserRoomStatus(userID, roomID string) (*UserRoomStatus, error) {
	var status UserRoomStatus
	if err := r.metaStore.DB.Where("user_id = ? AND room_id = ?", userID, roomID).First(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *MessageRepository) UpdateUserRoomStatus(userID, roomID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&UserRoomStatus{}).
		Where("user_id = ? AND room_id = ?", userID, roomID).
		Updates(updates).Error
}

// AgentMessage 相关操作
func (r *MessageRepository) InsertAgentMessage(message *AgentMessage) error {
	return r.metaStore.DB.Create(message).Error
}

func (r *MessageRepository) GetAgentMessageByID(agentMessageID string) (*AgentMessage, error) {
	var agentMessage AgentMessage
	if err := r.metaStore.DB.Where("id = ?", agentMessageID).First(&agentMessage).Error; err != nil {
		return nil, err
	}
	return &agentMessage, nil
}

func (r *MessageRepository) UpdateAgentMessage(agentMessageID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&AgentMessage{}).Where("id = ?", agentMessageID).Updates(updates).Error
}

func (r *MessageRepository) UpdateAgentMessageStatus(agentMessageID string, status string) error {
	return r.UpdateAgentMessage(agentMessageID, map[string]interface{}{
		"status": status,
	})
}

func (r *MessageRepository) UpdateAgentMessageWithError(agentMessageID string, status string, errorInfo interface{}) error {
	errorStr, _ := json.Marshal(errorInfo)
	return r.UpdateAgentMessage(agentMessageID, map[string]interface{}{
		"status": status,
		"error":  string(errorStr),
	})
}

func (r *MessageRepository) UpdateAgentMessageWithUsage(agentMessageID string, status string, usage interface{}, altText string) error {
	usageStr, _ := json.Marshal(usage)
	updates := map[string]interface{}{
		"status": status,
		"usage":  string(usageStr),
	}
	if altText != "" {
		updates["alt_text"] = altText
	}
	return r.UpdateAgentMessage(agentMessageID, updates)
}

func (r *MessageRepository) UpdateAgentMessageIncomplete(agentMessageID string, interruptType int32, errorInfo interface{}, incompleteDetails interface{}) error {
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

	return r.UpdateAgentMessage(agentMessageID, updates)
}

// AgentMessageItem 相关操作
func (r *MessageRepository) InsertAgentMessageItem(item *AgentMessageItem) error {
	return r.metaStore.DB.Create(item).Error
}

func (r *MessageRepository) UpdateAgentMessageItem(itemID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&AgentMessageItem{}).Where("id = ?", itemID).Updates(updates).Error
}

func (r *MessageRepository) UpdateAgentMessageItemByPosition(agentMessageID string, position int, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now().UnixMilli()
	return r.metaStore.DB.Model(&AgentMessageItem{}).
		Where("agent_message_id = ? AND position = ?", agentMessageID, position).
		Updates(updates).Error
}

// GetMessageCountByRoom 获取房间内的消息总数
func (r *MessageRepository) GetMessageCountByRoom(roomID string, threadID string) (int64, error) {
	var count int64
	err := r.metaStore.DB.Model(&Message{}).
		Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Count(&count).Error
	return count, err
}

// GetUnreadMessageCount 获取用户在房间内的未读消息数
func (r *MessageRepository) GetUnreadMessageCount(roomID string, threadID string, lastReadMessageID string) (int64, error) {
	var count int64
	query := r.metaStore.DB.Model(&Message{}).
		Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false)

	if lastReadMessageID != "" {
		// 获取最后已读消息的时间
		var lastReadMsg Message
		if err := r.metaStore.DB.Where("id = ?", lastReadMessageID).First(&lastReadMsg).Error; err != nil {
			return 0, err
		}
		query = query.Where("created_at > ?", lastReadMsg.CreatedAt)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetMessagesByTimeRange 按时间范围查询消息
func (r *MessageRepository) GetMessagesByTimeRange(roomID string, threadID string, startTime, endTime int64, limit int) ([]*Message, error) {
	var messages []*Message
	query := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ? AND created_at BETWEEN ? AND ?",
		roomID, threadID, false, startTime, endTime).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// GetLatestMessageInRoom 获取房间内最新的一条消息
func (r *MessageRepository) GetLatestMessageInRoom(roomID string, threadID string) (*Message, error) {
	var message Message
	err := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Order("created_at DESC").
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetMessagesByIDs 批量获取消息
func (r *MessageRepository) GetMessagesByIDs(messageIDs []string) ([]*Message, error) {
	var messages []*Message
	err := r.metaStore.DB.Where("id IN ? AND deleted = ?", messageIDs, false).
		Order("created_at DESC").
		Find(&messages).Error
	return messages, err
}

// SearchMessages 搜索消息内容
func (r *MessageRepository) SearchMessages(roomID string, threadID string, keyword string, limit int, offset int) ([]*Message, error) {
	var messages []*Message
	query := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ? AND content LIKE ?",
		roomID, threadID, false, "%"+keyword+"%").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&messages).Error
	return messages, err
}

// GetMessageStatsByRoom 获取房间消息统计信息
func (r *MessageRepository) GetMessageStatsByRoom(roomID string, threadID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总消息数
	var totalCount int64
	if err := r.metaStore.DB.Model(&Message{}).
		Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// 最早消息时间
	var earliestMsg Message
	if err := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Order("created_at ASC").
		First(&earliestMsg).Error; err == nil {
		stats["earliest_message_time"] = earliestMsg.CreatedAt
	}

	// 最新消息时间
	var latestMsg Message
	if err := r.metaStore.DB.Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Order("created_at DESC").
		First(&latestMsg).Error; err == nil {
		stats["latest_message_time"] = latestMsg.CreatedAt
	}

	// 按消息类型统计
	var typeStats []struct {
		MsgType int   `json:"msg_type"`
		Count   int64 `json:"count"`
	}
	if err := r.metaStore.DB.Model(&Message{}).
		Select("msg_type, COUNT(*) as count").
		Where("room_id = ? AND thread_id = ? AND deleted = ?", roomID, threadID, false).
		Group("msg_type").
		Scan(&typeStats).Error; err == nil {
		stats["message_types"] = typeStats
	}

	return stats, nil
}
