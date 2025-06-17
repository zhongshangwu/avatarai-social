package handlers

import (
	"net/http"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type ActivityHandler struct {
	config       *config.SocialConfig
	metaStore    *repositories.MetaStore
	tagService   *services.TagService
	topicService *services.TopicService
}

func NewActivityHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *ActivityHandler {
	return &ActivityHandler{
		config:       config,
		metaStore:    metaStore,
		tagService:   services.NewTagService(metaStore),
		topicService: services.NewTopicService(metaStore),
	}
}

func (h *ActivityHandler) CreateTag(c *types.APIContext) error {
	ctx := c.Request().Context()

	var req struct {
		Tag string `json:"tag" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "参数错误"})
	}

	tag, err := h.tagService.CreateOrGetTag(ctx, req.Tag, c.User.Did)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	tagView, err := h.tagService.PresentTagView([]*repositories.Tag{tag})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tag": tagView,
	})
}

func (h *ActivityHandler) ListTags(c *types.APIContext) error {
	ctx := c.Request().Context()
	page, pageSize := c.GetPageAndPageSize()
	tags, err := h.tagService.ListTags(ctx, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	tagViews, err := h.tagService.PresentTagView(tags)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	ret := make([]*types.TagView, 0, len(tags))
	for _, tag := range tags {
		ret = append(ret, tagViews[tag.Tag])
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tags": ret,
	})
}

func (h *ActivityHandler) CreateTopic(c *types.APIContext) error {
	ctx := c.Request().Context()

	var req struct {
		Topic string `json:"topic" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "参数错误"})
	}

	topic, err := h.topicService.CreateOrGetTopic(ctx, req.Topic, c.User.Did)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	topicView, err := h.topicService.PresentTopicView([]*repositories.Topic{topic})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"topic": topicView,
	})
}

func (h *ActivityHandler) ListTopics(c *types.APIContext) error {
	ctx := c.Request().Context()
	page, pageSize := c.GetPageAndPageSize()
	topics, err := h.topicService.ListTopics(ctx, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	topicViews, err := h.topicService.PresentTopicView(topics)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	ret := make([]*types.TopicView, 0, len(topics))

	for _, topic := range topics {
		ret = append(ret, topicViews[topic.Topic])
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"topics": ret,
	})
}
