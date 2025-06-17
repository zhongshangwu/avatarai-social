package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/pkg/utils"
	"github.com/zhongshangwu/avatarai-social/types"
)

type MomentHandler struct {
	config        *config.SocialConfig
	metaStore     *repositories.MetaStore
	momentService *services.MomentService
	feedService   *services.FeedService
	fileService   *services.FileService
}

func NewMomentHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MomentHandler {
	return &MomentHandler{
		config:        config,
		metaStore:     metaStore,
		momentService: services.NewMomentService(metaStore),
		feedService:   services.NewFeedService(config, metaStore),
		fileService:   services.NewFileService(config, metaStore),
	}
}

func (h *MomentHandler) CreateMoment(c *types.APIContext) error {
	if !c.IsAuthenticated {
		return c.RedirectToLogin("")
	}

	var req services.CreateMomentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "请求格式错误: "+err.Error())
	}

	if req.Text == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "文本内容不能为空")
	}

	response, err := h.momentService.CreateMoment(c.Request().Context(), c.User.Did, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "创建moment失败: "+err.Error())
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *MomentHandler) GetMoment(c *types.APIContext) error {
	uri := c.Param("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID参数不能为空")
	}

	moment, err := h.momentService.GetMomentByID(c.Request().Context(), uri)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "moment不存在: "+err.Error())
	}

	return c.JSON(http.StatusOK, moment)
}

func (h *MomentHandler) GetMomentThread(c *types.APIContext) error {
	uri := c.QueryParam("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "uri参数不能为空")
	}

	// 获取深度参数，默认为10
	depth := 10
	if depthStr := c.QueryParam("depth"); depthStr != "" {
		if d, err := utils.ParseInt(depthStr); err == nil && d > 0 && d <= 50 {
			depth = d
		}
	}

	thread, err := h.feedService.MomentThread(c.Request().Context(), uri, depth)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "获取 moment thread 失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, thread)
}

func (h *MomentHandler) LikeMoment(c *types.APIContext) error {
	uri := c.QueryParam("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "uri参数不能为空")
	}

	like, err := h.momentService.LikeMoment(c.Request().Context(), uri, c.User.Did)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "点赞失败: "+err.Error())
	}

	return c.JSON(http.StatusOK, like)
}

func (h *MomentHandler) RemoveLikeMoment(c *types.APIContext) error {
	uri := c.QueryParam("uri")
	if uri == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "uri参数不能为空")
	}

	likeURI := c.QueryParam("likeURI")
	if likeURI == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "likeURI参数不能为空")
	}

	err := h.momentService.RemoveLikeMoment(c.Request().Context(), uri, c.User.Did, likeURI)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "取消点赞失败: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}
