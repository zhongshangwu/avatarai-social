package avatarai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type AvatarEngine struct {
	db     *gorm.DB
	log    *slog.Logger
	config AvatarConfig

	modelClient ModelClient

	// 同步操作的锁
	opLock sync.Mutex

	// 管理活跃的会话
	sessionsLk sync.RWMutex
	sessions   map[string]*AvatarSession
}

type ModelClient interface {
	Generate(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	StreamGenerate(ctx context.Context, prompt string, options map[string]interface{}) (<-chan string, error)
}

type AvatarSession struct {
	ID         string
	UserID     string
	StartedAt  time.Time
	LastActive time.Time
	Completed  bool
	Result     string
}

type AvatarConfig struct {
	APIKey                string
	EndpointURL           string
	Timeout               time.Duration
	MaxConcurrentRequests int
	Logger                *slog.Logger
	AdminToken            string
}

func DefaultAvatarConfig() *AvatarConfig {
	return &AvatarConfig{
		Timeout:               time.Second * 30,
		MaxConcurrentRequests: 10,
	}
}

func NewAvatarEngine(db *gorm.DB, modelClient ModelClient, config *AvatarConfig) (*AvatarEngine, error) {
	if config == nil {
		config = DefaultAvatarConfig()
	}

	engine := &AvatarEngine{
		db:          db,
		modelClient: modelClient,
		log:         config.Logger,
		config:      *config,
		sessions:    make(map[string]*AvatarSession),
	}

	if engine.log == nil {
		engine.log = slog.Default().With("system", "avatar-engine")
	}

	return engine, nil
}

func (ae *AvatarEngine) Start(addr string, logWriter io.Writer) error {
	var lc net.ListenConfig
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	li, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	return ae.StartWithListener(li, logWriter)
}

func (ae *AvatarEngine) StartWithListener(listen net.Listener, logWriter io.Writer) error {
	e := echo.New()
	e.Logger.SetOutput(logWriter)
	e.HideBanner = true

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} latency=${latency_human}\n",
	}))

	api := e.Group("/api")

	api.GET("/avatar/profile", ae.handleGetAvatarProfile)
	api.POST("/avatar/init", ae.handleInitAvatar)
	api.GET("/avatar/status", ae.handleGetAvatarStatus)

	e.GET("/_health", ae.handleHealthCheck)

	e.Listener = listen
	srv := &http.Server{}
	return e.StartServer(srv)
}

func (ae *AvatarEngine) Shutdown(ctx context.Context) error {
	ae.sessionsLk.Lock()
	defer ae.sessionsLk.Unlock()

	ae.log.Info("shutting down avatar engine")
	ae.sessions = make(map[string]*AvatarSession)

	return nil
}

// func (ae *AvatarEngine) GenerateAvatar(ctx context.Context, userID string, prompt string, options map[string]interface{}) (string, error) {
// 	sessionID := generateUniqueID()

// 	session := &AvatarSession{
// 		ID:         sessionID,
// 		UserID:     userID,
// 		StartedAt:  time.Now(),
// 		LastActive: time.Now(),
// 		Completed:  false,
// 	}

// 	ae.sessionsLk.Lock()
// 	ae.sessions[sessionID] = session
// 	ae.sessionsLk.Unlock()

// 	go func() {
// 		result, err := ae.modelClient.Generate(context.Background(), prompt, options)
// 		ae.sessionsLk.Lock()
// 		if s, exists := ae.sessions[sessionID]; exists {
// 			s.LastActive = time.Now()
// 			s.Completed = true
// 			if err == nil {
// 				s.Result = result
// 				// 保存到数据库
// 				avatar := database.Avatar{
// 					DID:        sessionID,
// 					CreatorDID: userID,
// 					CreatedAt:  session.StartedAt,
// 				}
// 				ae.db.Create(&avatar)
// 			} else {
// 				ae.log.Error("failed to generate avatar", "session_id", sessionID, "error", err)
// 			}
// 		}
// 		ae.sessionsLk.Unlock()
// 	}()

// 	return sessionID, nil
// }

// func (ae *AvatarEngine) GetAvatarStatus(ctx context.Context, sessionID string) (*AvatarSession, error) {
// 	ae.sessionsLk.RLock()
// 	defer ae.sessionsLk.RUnlock()

// 	session, exists := ae.sessions[sessionID]
// 	if !exists {
// 		// 如果会话不在内存中，尝试从数据库查找
// 		var avatar database.Avatar
// 		if err := ae.db.Where("did = ?", sessionID).First(&avatar).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				return nil, fmt.Errorf("avatar session not found: %s", sessionID)
// 			}
// 			return nil, err
// 		}

// 		// 从数据库中恢复会话
// 		session = &AvatarSession{
// 			ID:         sessionID,
// 			StartedAt:  avatar.CreatedAt,
// 			LastActive: avatar.CreatedAt,
// 			Completed:  true,
// 		}
// 	}

// 	return session, nil
// }

func (ae *AvatarEngine) handleGetAvatarStatus(c echo.Context) error {
	// 获取生成状态
	return nil // 待实现
}

func (ae *AvatarEngine) handleGetAvatarProfile(c echo.Context) error {
	// 获取数字身份
	return nil // 待实现
}

func (ae *AvatarEngine) handleInitAvatar(c echo.Context) error {
	// 初始化数字身份
	return nil // 待实现
}

func (ae *AvatarEngine) handleHealthCheck(c echo.Context) error {
	if err := ae.db.Exec("SELECT 1").Error; err != nil {
		ae.log.Error("healthcheck失败", "err", err)
		return c.JSON(500, map[string]interface{}{
			"status":  "error",
			"message": "无法连接到数据库",
		})
	}
	return c.JSON(200, map[string]interface{}{
		"status": "ok",
	})
}

func (ae *AvatarEngine) checkAdminAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth != "Bearer "+ae.config.AdminToken {
			return echo.ErrForbidden
		}
		return next(c)
	}
}

// 辅助函数
func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
