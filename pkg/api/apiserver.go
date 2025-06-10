package api

import (
	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/api/handlers"
	mw "github.com/zhongshangwu/avatarai-social/pkg/api/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

//go:embed templates/app-return.html
var appReturnHTML string

type AvatarAIAPI struct {
	Config          *config.SocialConfig
	echo            *echo.Echo
	MetaStore       *repositories.MetaStore
	HealthHandler   *handlers.HealthHandler
	AuthHandler     *handlers.OAuthHandler
	UserHandler     *handlers.UserHandler
	AsterHandler    *handlers.AsterHandler
	MomentsHandler  *handlers.MomentHandler
	FeedHandler     *handlers.FeedHandler
	BlobsHandler    *handlers.BlobHandler
	MessagesHandler *handlers.MessageHandler
	ChatHandler     *handlers.ChatHandler
}

func NewAvatarAIAPI(config *config.SocialConfig, metaStore *repositories.MetaStore) *AvatarAIAPI {
	e := echo.New()

	renderer := NewTemplateRenderer("pkg/api/templates/*.html")
	e.Renderer = renderer

	healthHandler := handlers.NewHealthHandler(config, metaStore)
	oauthHandler := handlers.NewOAuthHandler(config, metaStore, appReturnHTML)
	userHandler := handlers.NewUserHandler(config, metaStore)
	asterHandler := handlers.NewAsterHandler(config, metaStore)
	momentHandler := handlers.NewMomentHandler(config, metaStore)
	blobHandler := handlers.NewBlobHandler(config, metaStore)
	messageHandler := handlers.NewMessageHandler(config, metaStore)
	chatHandler := handlers.NewChatHandler(config, metaStore)

	return &AvatarAIAPI{
		Config:          config,
		echo:            e,
		MetaStore:       metaStore,
		HealthHandler:   healthHandler,
		AuthHandler:     oauthHandler,
		UserHandler:     userHandler,
		AsterHandler:    asterHandler,
		MomentsHandler:  momentHandler,
		BlobsHandler:    blobHandler,
		MessagesHandler: messageHandler,
		ChatHandler:     chatHandler,
	}
}

func (a *AvatarAIAPI) InstallRoutes() {
	a.echo.GET("/healthz", a.HealthHandler.Healthz)

	api := a.echo.Group("/api")
	withAuth := mw.NewContextWrapper(a.MetaStore, a.Config)

	oauth := api.Group("/oauth")
	oauth.GET("/login", a.AuthHandler.OAuthLogin)
	oauth.POST("/signin", a.AuthHandler.OAuthLogin)
	oauth.GET("/jwks.json", a.AuthHandler.HandleOAuthJWKS)
	oauth.GET("/callback", a.AuthHandler.HandleOAuthCallback)
	oauth.GET("/token", a.AuthHandler.HandleOAuthToken)
	oauth.GET("/app-return/:bundleID", a.AuthHandler.HandleAppReturn)
	oauth.GET("/:platform/client-metadata.json", a.AuthHandler.OAuthClientMetadata)
	oauth.POST("/refresh", withAuth(a.AuthHandler.HandleOAuthRefresh, false))
	oauth.GET("/logout", withAuth(a.AuthHandler.HandleOAuthLogout, true))

	avatar := api.Group("/avatar")
	avatar.GET("/profile", withAuth(a.UserHandler.CurrentUserProfile, true))

	aster := api.Group("/aster")
	aster.POST("/mint", withAuth(a.AsterHandler.HandleAsterMint, true))
	aster.GET("/profile", withAuth(a.AsterHandler.GetAsterProfile, true))

	// 聊天相关路由
	chat := api.Group("/chat")
	chat.GET("/stream", withAuth(a.ChatHandler.ChatStream, true))

	moment := api.Group("/moments")
	moment.POST("", withAuth(a.MomentsHandler.CreateMoment, true))
	moment.GET("/detail", withAuth(a.MomentsHandler.GetMoment, true))

	feed := api.Group("/feeds")
	feed.GET("", withAuth(a.FeedHandler.Feeds, true))

	blob := api.Group("/blobs")
	blob.POST("", withAuth(a.BlobsHandler.UploadFile, true))
	blob.GET("", withAuth(a.BlobsHandler.GetFile, true))

	messages := api.Group("/messages")
	messages.GET("/history", withAuth(a.MessagesHandler.HistoryMessages, true))
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
