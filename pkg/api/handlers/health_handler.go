package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type HealthHandler struct {
	config    *config.SocialConfig
	metaStore *repositories.MetaStore
}

func NewHealthHandler(config *config.SocialConfig, metaStore *repositories.MetaStore) *HealthHandler {
	return &HealthHandler{
		config:    config,
		metaStore: metaStore,
	}
}

func (h *HealthHandler) Healthz(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "ok",
	})
}
