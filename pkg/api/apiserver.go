package api

import (
	_ "embed"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/api/handlers"
	mw "github.com/zhongshangwu/avatarai-social/pkg/api/middleware"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/blobs"
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
	ImageViewer     *blobs.ImageViewer
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
	feedHandler := handlers.NewFeedHandler(config, metaStore)

	viewerConfig := blobs.DefaultImageViewerConfig()
	viewerConfig.CacheLocation = "/tmp/avatar-ai-image-cache"
	viewerConfig.MaxResponseSize = 50 << 20 // 50MB
	viewerConfig.UserAgent = "AvatarAI-Social/1.0"

	viewer, err := blobs.NewImageViewer(viewerConfig)
	if err != nil {
		e.Logger.Errorf("创建图片查看器失败: %v", err)
		panic(err)
	}

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
		FeedHandler:     feedHandler,
		ImageViewer:     viewer,
	}
}

func (a *AvatarAIAPI) InstallRoutes() {
	a.echo.GET("/healthz", a.HealthHandler.Healthz)

	// 静态文件服务 - 提供前端页面
	a.echo.Static("/", "web")

	api := a.echo.Group("/api")
	withAuth := mw.NewContextWrapper(a.MetaStore, a.Config)

	oauth := api.Group("/oauth")
	oauth.GET("/login", a.AuthHandler.OAuthLogin)
	oauth.POST("/signin", a.AuthHandler.OAuthLogin)
	oauth.GET("/jwks.json", a.AuthHandler.HandleOAuthJWKS)
	oauth.GET("/callback", a.AuthHandler.HandleOAuthCallback)
	oauth.GET("/token", a.AuthHandler.HandleOAuthToken)
	oauth.POST("/token", a.AuthHandler.HandleOAuthToken)
	oauth.GET("/app-return/:bundleID", a.AuthHandler.HandleAppReturn)
	oauth.GET("/:platform/client-metadata.json", a.AuthHandler.OAuthClientMetadata)
	oauth.POST("/refresh", withAuth(a.AuthHandler.HandleOAuthRefresh, false))
	oauth.GET("/logout", withAuth(a.AuthHandler.HandleOAuthLogout, true))
	oauth.POST("/bsky-post", withAuth(a.AuthHandler.HandleBskyPost, true))

	avatar := api.Group("/avatar")
	avatar.GET("/profile", withAuth(a.UserHandler.CurrentUserProfile, true))
	avatar.POST("/profile", withAuth(a.UserHandler.UpdateUserProfile, true))

	aster := api.Group("/aster")
	aster.POST("/mint", withAuth(a.AsterHandler.HandleAsterMint, true))
	aster.GET("/profile", withAuth(a.AsterHandler.GetAsterProfile, true))

	chat := api.Group("/chat")
	chat.GET("/stream", withAuth(a.ChatHandler.ChatStream, true))

	moment := api.Group("/moments")
	moment.POST("", withAuth(a.MomentsHandler.CreateMoment, true))
	moment.GET("/detail", withAuth(a.MomentsHandler.GetMoment, true))
	moment.GET("/thread", withAuth(a.MomentsHandler.GetMomentThread, false))
	moment.POST("/like", withAuth(a.MomentsHandler.LikeMoment, true))
	moment.DELETE("/like", withAuth(a.MomentsHandler.UndoLikeMoment, true))

	feed := api.Group("/feeds")
	feed.GET("", withAuth(a.FeedHandler.Feeds, true))

	blob := api.Group("/blobs")
	blob.POST("", withAuth(a.BlobsHandler.UploadFile, true))
	blob.GET("", withAuth(a.BlobsHandler.GetFile, true))

	messages := api.Group("/messages")
	messages.GET("/history", withAuth(a.MessagesHandler.HistoryMessages, true))

	img := a.echo.Group("/img")
	img.Use(echo.WrapMiddleware(a.ImageViewer.CreateMiddleware("/img/")))
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

// func (a *AvatarAIAPI) handleATProtoBlob(c *types.APIContext) error {
// 	var req struct {
// 		DID string `json:"did" validate:"required"`
// 		CID string `json:"cid" validate:"required"`
// 	}

// 	if err := c.Bind(&req); err != nil {
// 		return c.JSON(400, map[string]string{"error": "参数错误"})
// 	}

// 	options := helper.BlobOptions{
// 		DID:         req.DID,
// 		CID:         req.CID,
// 		Identifier:  req.CID,
// 		StorageType: helper.StorageATProto,
// 	}

// 	data, err := a.BlobReader.GetBlob(c.Request().Context(), options)
// 	if err != nil {
// 		return c.JSON(500, map[string]string{"error": err.Error()})
// 	}

// 	return c.JSON(200, map[string]interface{}{
// 		"size":         data.Size,
// 		"content_type": data.ContentType,
// 		"source":       data.Source,
// 		"metadata":     data.Metadata,
// 		"data":         data.Data, // 注意：实际应用中可能需要base64编码
// 	})
// }
