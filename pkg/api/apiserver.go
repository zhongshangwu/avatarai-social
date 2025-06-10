package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/api/handlers"
	mw "github.com/zhongshangwu/avatarai-social/pkg/api/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type AvatarAIAPI struct {
	Config          *config.SocialConfig
	echo            *echo.Echo
	MetaStore       *repositories.MetaStore
	HealthHandler   *handlers.HealthHandler
	AuthHandler     *handlers.OAuthHandler
	UserHandler     *handlers.UserHandler
	AsterHandler    *handlers.AsterHandler
	MomentsHandler  *handlers.MomentHandler
	BlobsHandler    *handlers.BlobHandler
	MessagesHandler *handlers.MessageHandler
	ChatHandler     *handlers.ChatHandler
}

func NewAvatarAIAPI(config *config.SocialConfig, metaStore *repositories.MetaStore) *AvatarAIAPI {
	e := echo.New()

	renderer := NewTemplateRenderer("pkg/api/templates/*.html")
	e.Renderer = renderer

	healthHandler := handlers.NewHealthHandler(config, metaStore)
	oauthHandler := handlers.NewOAuthHandler(config, metaStore)
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

	oauth := api.Group("/oauth")
	oauth.GET("/login", a.AuthHandler.OAuthLogin)
	oauth.POST("/signin", a.AuthHandler.OAuthLogin)
	oauth.GET("/jwks.json", a.AuthHandler.HandleOAuthJWKS)
	oauth.GET("/callback", a.AuthHandler.HandleOAuthCallback)
	oauth.GET("/token", a.AuthHandler.HandleOAuthToken)
	oauth.GET("/app-return/:bundleID", a.AuthHandler.HandleAppReturn)
	oauth.GET("/:platform/client-metadata.json", a.AuthHandler.HandleOAuthClientMetadata)
	oauth.GET("/refresh", mw.NewRequireAuth(a.MetaStore, a.Config)(a.AuthHandler.HandleOAuthRefresh))
	oauth.POST("/refresh", mw.NewRequireAuth(a.MetaStore, a.Config)(a.AuthHandler.HandleOAuthRefresh))
	oauth.GET("/logout", mw.NewRequireAuth(a.MetaStore, a.Config)(a.AuthHandler.HandleOAuthLogout))

	avatar := api.Group("/avatar")
	avatar.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	avatar.GET("/profile", mw.WrapHandler(a.UserHandler.CurrentUserProfile))
	avatar.POST("/profile", mw.WrapHandler(a.UserHandler.UpdateUserProfile))

	aster := api.Group("/aster")
	aster.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	aster.POST("/mint", mw.WrapHandler(a.AsterHandler.HandleAsterMint))
	aster.GET("/profile", mw.WrapHandler(a.AsterHandler.GetAsterProfile))

	// 聊天相关路由
	api.GET("/demo/chat-stream", mw.WrapHandler(a.ChatHandler.ChatStream))

	chat := api.Group("/chat")
	// chat.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	chat.GET("/history", mw.WrapHandler(a.ChatHandler.ChatHistoryMessages))

	moment := api.Group("/moments")
	moment.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	moment.POST("", a.MomentsHandler.HandleMomentCreate)
	moment.GET("/detail", a.MomentsHandler.HandleMomentDetail)

	feed := api.Group("/feed")
	feed.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	feed.GET("", a.MomentsHandler.HandleMomentFeed)

	blob := api.Group("/blobs")
	blob.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	blob.POST("", mw.WrapHandler(a.BlobsHandler.UploadBlobHandler))
	blob.GET("", mw.WrapHandler(a.BlobsHandler.GetUploadFilesHandler))

	messages := api.Group("/messages")
	// messages.Use(mw.NewRequireAuth(a.MetaStore, a.Config))
	messages.GET("/history", a.MessagesHandler.GetMessagesHistoryHandler)
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
