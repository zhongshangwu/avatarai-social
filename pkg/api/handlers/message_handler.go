package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/communication/messages"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type GetMessagesHistoryResponse struct {
	Messages   []*messages.Message `json:"messages"`
	Pagination PaginationInfo      `json:"pagination"`
}

type PaginationInfo struct {
	Before  string `json:"before,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	After   string `json:"after,omitempty"`
	Count   int    `json:"count"`
	HasMore bool   `json:"hasMore"`
	FirstID string `json:"firstId,omitempty"`
	LastID  string `json:"lastId,omitempty"`
}

type MessageHandler struct {
	config    *config.SocialConfig
	metaStore *repositories.MetaStore
}

func NewMessageHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MessageHandler {
	return &MessageHandler{
		config:    config,
		metaStore: metaStore,
	}
}

func (h *MessageHandler) GetMessagesHistoryHandler(c echo.Context) error {
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

	// 从数据库获取消息 - 使用新的分页方法
	result, err := h.metaStore.MessageRepo.ListMessagesHistoryWithPagination(
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
	apiMessages := make([]*messages.Message, 0, len(result.Messages))
	for _, dbMsg := range result.Messages {
		apiMsg := messages.DBToMessage(dbMsg)
		if apiMsg != nil {
			apiMessages = append(apiMessages, apiMsg)
		}
	}

	// 构建响应 - 使用 repository 返回的分页信息
	response := GetMessagesHistoryResponse{
		Messages: apiMessages,
		Pagination: PaginationInfo{
			Before:  beforeMsgID,
			Limit:   limit,
			After:   afterMsgID,
			Count:   len(apiMessages),
			HasMore: result.HasMore,
			FirstID: result.FirstID,
			LastID:  result.LastID,
		},
	}

	return c.JSON(http.StatusOK, response)
}

func (h *MessageHandler) GetMessageStatsHandler(c echo.Context) error {
	roomID := c.QueryParam("roomId")
	threadID := c.QueryParam("threadId")

	// 验证必需参数
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "roomId 参数是必需的")
	}

	if threadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "threadId 参数是必需的")
	}

	// 获取统计信息
	stats, err := h.metaStore.MessageRepo.GetMessageStatsByRoom(roomID, threadID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取统计信息失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}

func (h *MessageHandler) SearchMessagesHandler(c echo.Context) error {
	roomID := c.QueryParam("roomId")
	threadID := c.QueryParam("threadId")
	keyword := c.QueryParam("keyword")
	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	// 验证必需参数
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "roomId 参数是必需的")
	}

	if threadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "threadId 参数是必需的")
	}

	if keyword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "keyword 参数是必需的")
	}

	// 解析分页参数
	limit := 20
	if limitStr != "" {
		if count, err := strconv.Atoi(limitStr); err == nil && count > 0 && count <= 100 {
			limit = count
		}
	}

	offset := 0
	if offsetStr != "" {
		if count, err := strconv.Atoi(offsetStr); err == nil && count >= 0 {
			offset = count
		}
	}

	// 搜索消息
	dbMessages, err := h.metaStore.MessageRepo.SearchMessages(roomID, threadID, keyword, limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "搜索消息失败: "+err.Error())
	}

	// 转换为API格式
	apiMessages := make([]*messages.Message, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		apiMsg := messages.DBToMessage(dbMsg)
		if apiMsg != nil {
			apiMessages = append(apiMessages, apiMsg)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"messages": apiMessages,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(apiMessages),
		},
	})
}

func (h *MessageHandler) GetUnreadCountHandler(c echo.Context) error {
	roomID := c.QueryParam("roomId")
	threadID := c.QueryParam("threadId")
	lastReadMessageID := c.QueryParam("lastReadMessageId")

	// 验证必需参数
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "roomId 参数是必需的")
	}

	if threadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "threadId 参数是必需的")
	}

	// 获取未读消息数
	unreadCount, err := h.metaStore.MessageRepo.GetUnreadMessageCount(roomID, threadID, lastReadMessageID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取未读消息数失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"unreadCount": unreadCount,
	})
}
