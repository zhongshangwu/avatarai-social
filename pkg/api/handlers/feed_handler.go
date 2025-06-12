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
		feedService: services.NewFeedService(config, metaStore),
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
	_ = c.Request().Context()

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

func (h *FeedHandler) MomentThread(c *types.APIContext) error {
	uri := c.QueryParam("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "uri参数不能为空")
	}

	depth := 10
	if depthStr := c.QueryParam("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 50 {
			depth = d
		}
	}

	thread, err := h.feedService.MomentThread(c.Request().Context(), uri, depth)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "获取帖子失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, thread)
}
