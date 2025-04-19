package api

import (
	_ "embed"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

//go:embed templates/app-return.html
var appReturnHTML string

func (a *AvatarAIAPI) HandleAppReturn(c echo.Context) error {
	bundleID := c.Param("bundleID")
	if bundleID == "" {
		return c.JSON(http.StatusNotImplemented, map[string]string{
			"error": "server has no --app-bundle-id set",
		})
	}

	html := strings.Replace(appReturnHTML, "APP_BUNDLE_ID_REPLACE_ME", bundleID, 1)
	return c.HTML(http.StatusOK, html)
}
