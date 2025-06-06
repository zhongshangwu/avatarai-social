package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
)

// GetMessagesHistoryResponse 历史消息响应结构
type GetMessagesHistoryResponse struct {
	Messages   []*messages.Message `json:"messages"`
	Pagination PaginationInfo      `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Before  string `json:"before,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	After   string `json:"after,omitempty"`
	Count   int    `json:"count"`
	HasMore bool   `json:"hasMore"`
}

// GetMessagesHistoryHandler 获取历史消息接口
func (a *AvatarAIAPI) GetMessagesHistoryHandler(c echo.Context) error {
	// ac := c.(*utils.AvatarAIContext)
	// avatar := ac.Avatar

	// if avatar == nil {
	// 	return echo.NewHTTPError(http.StatusUnauthorized, "未授权访问")
	// }

	// 获取查询参数
	roomID := c.QueryParam("roomId")
	threadID := c.QueryParam("threadId")
	beforeMsgID := c.QueryParam("before")
	limitStr := c.QueryParam("limit")
	afterMsgID := c.QueryParam("after")

	// 验证必需参数
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "roomId 参数是必需的")
	}

	if threadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "threadId 参数是必需的")
	}

	// 解析计数参数
	limit := 0

	if limitStr != "" {
		if count, err := strconv.Atoi(limitStr); err == nil && count > 0 && count <= 100 {
			limit = count
		}
	}

	// 如果没有指定任何计数，默认获取最新的20条消息
	if limit == 0 {
		limit = 20
	}

	// 从数据库获取消息
	dbMessages, err := database.ListMessagesHistoryWithPagination(
		a.metaStore.DB,
		roomID,
		threadID,
		beforeMsgID,
		limit,
		afterMsgID,
		limit,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取历史消息失败: "+err.Error())
	}

	// 转换数据库消息为API消息格式
	apiMessages := make([]*messages.Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		apiMsg := messages.DBToMessage(dbMsg)
		if apiMsg != nil {
			apiMessages = append(apiMessages, apiMsg)
		}
	}

	// 检查是否还有更多消息
	hasMore := false
	if len(apiMessages) > 0 {
		// 检查是否还有更早的消息
		if beforeMsgID != "" || (beforeMsgID == "" && afterMsgID == "") {
			oldestMsg := apiMessages[len(apiMessages)-1]
			var count int64
			a.metaStore.DB.Model(&database.Message{}).
				Where("room_id = ? AND thread_id = ? AND deleted = ? AND created_at < ?",
					roomID, threadID, false, oldestMsg.CreatedAt).
				Count(&count)
			hasMore = count > 0
		}

		// 检查是否还有更新的消息
		if afterMsgID != "" {
			newestMsg := apiMessages[0]
			var count int64
			a.metaStore.DB.Model(&database.Message{}).
				Where("room_id = ? AND thread_id = ? AND deleted = ? AND created_at > ?",
					roomID, threadID, false, newestMsg.CreatedAt).
				Count(&count)
			if count > 0 {
				hasMore = true
			}
		}
	}

	// 构建响应
	response := GetMessagesHistoryResponse{
		Messages: apiMessages,
		Pagination: PaginationInfo{
			Before:  beforeMsgID,
			Limit:   limit,
			After:   afterMsgID,
			Count:   len(apiMessages),
			HasMore: hasMore,
		},
	}

	return c.JSON(http.StatusOK, response)
}
