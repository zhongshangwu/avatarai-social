package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/pkg/services"
	"github.com/zhongshangwu/avatarai-social/types"
)

type MomentHandler struct {
	config        *config.SocialConfig
	metaStore     *repositories.MetaStore
	momentService *services.MomentService
	fileService   *services.FileService
}

func NewMomentHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *MomentHandler {
	return &MomentHandler{
		config:        config,
		metaStore:     metaStore,
		momentService: services.NewMomentService(metaStore),
		fileService:   services.NewFileService(metaStore),
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
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ID参数不能为空")
	}

	moment, err := h.momentService.GetMomentByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "moment不存在: "+err.Error())
	}

	return c.JSON(http.StatusOK, moment)
}
