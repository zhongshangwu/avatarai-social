package api

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	mw "github.com/zhongshangwu/avatarai-social/pkg/api/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(pattern string) *TemplateRenderer {
	return &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}

type AvatarAIAPI struct {
	Config    *config.SocialConfig
	echo      *echo.Echo
	metaStore *database.MetaStore
}

func NewAvatarAIAPI(config *config.SocialConfig, metaStore *database.MetaStore) *AvatarAIAPI {
	e := echo.New()

	renderer := NewTemplateRenderer("pkg/api/templates/*.html")
	e.Renderer = renderer
	return &AvatarAIAPI{
		Config:    config,
		echo:      e,
		metaStore: metaStore,
	}
}

func (a *AvatarAIAPI) InstallRoutes() {
	api := a.echo.Group("/api")

	api.GET("/healthz", a.HandleHealthz)

	oauth := api.Group("/oauth")
	oauth.GET("/login", a.HandleOAuthLogin)
	oauth.POST("/signin", a.HandleOAuthLogin)
	oauth.GET("/jwks.json", a.HandleOAuthJWKS)
	oauth.GET("/callback", a.HandleOAuthCallback)
	oauth.GET("/token", a.HandleOAuthToken)
	oauth.GET("/app-return/:bundleID", a.HandleAppReturn)
	oauth.GET("/:platform/client-metadata.json", a.HandleOAuthClientMetadata)
	oauth.GET("/refresh", mw.NewRequireAuth(a.metaStore, a.Config)(a.HandleOAuthRefresh))
	oauth.POST("/refresh", mw.NewRequireAuth(a.metaStore, a.Config)(a.HandleOAuthRefresh))
	oauth.GET("/logout", mw.NewRequireAuth(a.metaStore, a.Config)(a.HandleOAuthLogout))

	avatar := api.Group("/avatar")
	avatar.Use(mw.NewRequireAuth(a.metaStore, a.Config))
	avatar.GET("/profile", echo.HandlerFunc(a.HandleAvatarProfile))
	avatar.POST("/profile", echo.HandlerFunc(a.HandleUpdateAvatarProfile))

	aster := api.Group("/aster")
	aster.Use(mw.NewRequireAuth(a.metaStore, a.Config))
	aster.GET("/profile", a.HandleAsterProfile)
	aster.POST("/mint", a.HandleAsterMint)

	moment := api.Group("/moments")
	moment.Use(mw.NewRequireAuth(a.metaStore, a.Config))
	moment.POST("", a.HandleMomentCreate)
	moment.GET("/detail", a.HandleMomentDetail)

	feed := api.Group("/feed")
	feed.Use(mw.NewRequireAuth(a.metaStore, a.Config))
	feed.GET("", a.HandleMomentFeed)

	blob := api.Group("/blobs")
	blob.Use(mw.NewRequireAuth(a.metaStore, a.Config))
	blob.POST("", a.UploadBlobHandler)
}

func (a *AvatarAIAPI) InstallMiddleware() {
	a.echo.Use(middleware.Logger())
	a.echo.Use(middleware.Recover())
	a.echo.Use(middleware.CORS())
	a.echo.Use(middleware.RequestID())
}

func (a *AvatarAIAPI) Start() error {
	a.InstallMiddleware()
	a.InstallRoutes()
	return a.echo.Start(a.Config.Server.HTTP.Address)
}

func (a *AvatarAIAPI) HandleHealthz(c echo.Context) error {
	return c.JSON(200, map[string]string{
		"status": "ok",
	})
}
