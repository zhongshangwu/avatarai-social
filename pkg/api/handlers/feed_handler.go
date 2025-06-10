package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type FeedHandler struct {
	config      *config.SocialConfig
	metaStore   *repositories.MetaStore
	feedService *services.FeedService
}

func NewFeedHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *FeedHandler {
	return &FeedHandler{
		config:      config,
		metaStore:   metaStore,
		feedService: services.NewFeedService(metaStore),
	}
}

func (h *FeedHandler) Feeds(c *types.APIContext) error {
	limitStr := c.QueryParam("limit")
	limit := 20 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cursor := c.QueryParam("cursor")
	feedName := c.QueryParam("feed")
	if feedName == "" {
		feedName = "default"
	}

	feeds, err := h.feedService.Feeds(c.Request().Context(), feedName, limit, cursor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取feed失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, feeds)
}
