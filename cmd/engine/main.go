package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/zhongshangwu/avatarai-social/pkg/api"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/pds/syncers"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
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
	metaStore := repositories.NewMetaStore(db)
	if err := metaStore.Init(); err != nil {
		return fmt.Errorf("初始化元数据存储失败: %w", err)
	}

	// 创建同步器管理器
	syncerConfig := syncers.DefaultSyncerConfig()
	syncerManager := syncers.NewSyncerManager(metaStore, syncerConfig)

	// 创建 API 服务器
	apiServer := api.NewAvatarAIAPI(cfg, metaStore)

	// 启动服务
	apiErr := make(chan error, 1)
	syncerErr := make(chan error, 1)

	// 启动同步器管理器
	go func() {
		log.Info("启动同步器管理器")
		if err := syncerManager.Start(); err != nil {
			syncerErr <- fmt.Errorf("同步器管理器启动失败: %w", err)
		}
	}()

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
		shutdownServices(syncerManager)
	case err := <-syncerErr:
		if err != nil {
			log.Error("同步器管理器错误", "err", err)
		}
		shutdownServices(syncerManager)
	case err := <-apiErr:
		if err != nil {
			log.Error("API 服务器错误", "err", err)
		}
		shutdownServices(syncerManager)
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

// 关闭服务
func shutdownServices(syncerManager *syncers.SyncerManager) {
	log.Info("正在关闭服务...")

	// 停止同步器管理器
	if err := syncerManager.Stop(); err != nil {
		log.Error("关闭同步器管理器时出错", "err", err)
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
