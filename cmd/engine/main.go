package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/zhongshangwu/avatarai-social/pkg/api"
	"github.com/zhongshangwu/avatarai-social/pkg/avatarai"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/database"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var log = slog.Default().With("system", "avatarai-engine")

func main() {
	if err := run(os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	app := cli.App{
		Name:    "avatarai-engine",
		Usage:   "AvatarAI 引擎服务",
		Version: "0.1.0",
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "conf",
			Usage:   "配置文件路径",
			EnvVars: []string{"AVATARAI_CONFIG"},
		},
	}

	app.Action = runAvatarEngine
	return app.Run(args)
}

func runAvatarEngine(cctx *cli.Context) error {
	// 加载配置
	cfg, err := config.LoadConfig(cctx.String("conf"))
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 捕获 SIGINT 信号以触发关闭
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// 确保数据目录存在
	if err := os.MkdirAll(cfg.Storage.DataDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 设置日志
	logger := slog.Default().With("system", "avatar-engine")
	slog.SetDefault(logger)

	// 设置数据库连接
	db, err := setupDatabase(cfg.Database)
	if err != nil {
		return fmt.Errorf("设置数据库失败: %w", err)
	}

	// 初始化元数据存储
	metaStore := database.NewMetaStore(db)
	if err := metaStore.Init(); err != nil {
		return fmt.Errorf("初始化元数据存储失败: %w", err)
	}

	// 设置模型客户端
	modelClient := createModelClient()

	// 创建引擎配置
	engineConfig := &avatarai.AvatarConfig{
		APIKey:                cfg.Avatar.LLM.APIKey,
		EndpointURL:           cfg.Avatar.LLM.APIURL,
		Timeout:               cfg.Server.HTTP.ReadTimeout,
		MaxConcurrentRequests: cfg.Database.MaxConnections, // 使用数据库最大连接数作为并发请求限制
		Logger:                logger,
		AdminToken:            cfg.Server.AdminKey,
	}

	// 创建头像引擎
	avatarEngine, err := avatarai.NewAvatarEngine(db, modelClient, engineConfig)
	if err != nil {
		return fmt.Errorf("创建头像引擎失败: %w", err)
	}

	// 创建 API 服务器
	apiServer := api.NewAvatarAIAPI(cfg, metaStore)

	// 启动服务
	engineErr := make(chan error, 1)
	apiErr := make(chan error, 1)

	// 启动引擎服务
	// go func() {
	// 	log.Info("启动头像引擎服务", "地址", cfg.Server.HTTP.Address)
	// 	err := avatarEngine.Start(cfg.Server.HTTP.Address, os.Stdout)
	// 	engineErr <- err
	// }()

	// 启动 API 服务器
	go func() {
		log.Info("启动 API 服务器", "地址", cfg.Server.HTTP.Address)
		err := apiServer.Start()
		apiErr <- err
	}()

	log.Info("服务启动完成")

	// 等待信号或错误
	select {
	case <-signals:
		log.Info("收到关闭信号")
		shutdownServices(avatarEngine)
	case err := <-engineErr:
		if err != nil {
			log.Error("引擎服务错误", "err", err)
		}
		shutdownServices(avatarEngine)
	case err := <-apiErr:
		if err != nil {
			log.Error("API 服务器错误", "err", err)
		}
		shutdownServices(avatarEngine)
	}

	log.Info("关闭完成")
	return nil
}

func setupDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "sqlite":
		dialector = sqlite.Open(cfg.DSN)
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.ConnectionLifetime)

	return db, nil
}

func createModelClient() avatarai.ModelClient {
	// 创建一个简单的模型客户端实现
	// 实际实现应当根据配置创建合适的客户端
	return &mockModelClient{}
}

// 关闭服务
func shutdownServices(engine *avatarai.AvatarEngine) {
	log.Info("正在关闭服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := engine.Shutdown(ctx); err != nil {
		log.Error("关闭 AvatarAI 引擎时出错", "err", err)
	}
}

// mock 实现模型客户端接口
type mockModelClient struct{}

func (m *mockModelClient) Generate(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	return "mock generated avatar", nil
}

func (m *mockModelClient) StreamGenerate(ctx context.Context, prompt string, options map[string]interface{}) (<-chan string, error) {
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
		ch <- "mock generated avatar"
	}()
	return ch, nil
}
