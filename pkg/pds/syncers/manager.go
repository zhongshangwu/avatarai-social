package syncers

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

// SyncerManager 同步器管理器
type SyncerManager struct {
	metaStore    *repositories.MetaStore
	momentSyncer *MomentSyncer
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	startTime    time.Time
	config       *SyncerConfig
}

// SyncerConfig 同步器配置
type SyncerConfig struct {
	MomentSyncInterval time.Duration `json:"moment_sync_interval"`
	BatchSize          int           `json:"batch_size"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	EnableMetrics      bool          `json:"enable_metrics"`
	LogLevel           string        `json:"log_level"`
}

// DefaultSyncerConfig 默认同步器配置
func DefaultSyncerConfig() *SyncerConfig {
	return &SyncerConfig{
		MomentSyncInterval: 30 * time.Second,
		BatchSize:          50,
		MaxRetries:         3,
		RetryDelay:         5 * time.Second,
		EnableMetrics:      true,
		LogLevel:           "info",
	}
}

// NewSyncerManager 创建新的同步器管理器
func NewSyncerManager(metaStore *repositories.MetaStore, config *SyncerConfig) *SyncerManager {
	if config == nil {
		config = DefaultSyncerConfig()
	}

	return &SyncerManager{
		metaStore: metaStore,
		config:    config,
	}
}

// Start 启动同步器管理器
func (sm *SyncerManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("同步器管理器已经在运行")
	}

	sm.ctx, sm.cancel = context.WithCancel(context.Background())
	sm.startTime = time.Now()
	sm.running = true

	log.Println("启动同步器管理器")

	// 初始化各个同步器
	if err := sm.initializeSyncers(); err != nil {
		sm.running = false
		sm.cancel()
		return fmt.Errorf("初始化同步器失败: %w", err)
	}

	// 启动各个同步器
	if err := sm.startSyncers(); err != nil {
		sm.running = false
		sm.cancel()
		return fmt.Errorf("启动同步器失败: %w", err)
	}

	log.Println("同步器管理器启动完成")
	return nil
}

// Stop 停止同步器管理器
func (sm *SyncerManager) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return nil
	}

	log.Println("停止同步器管理器")

	// 停止各个同步器
	sm.stopSyncers()

	// 取消上下文
	if sm.cancel != nil {
		sm.cancel()
	}

	sm.running = false
	log.Println("同步器管理器已停止")
	return nil
}

// initializeSyncers 初始化同步器
func (sm *SyncerManager) initializeSyncers() error {
	// 初始化Moment同步器
	sm.momentSyncer = NewMomentSyncer(sm.metaStore)
	sm.momentSyncer.SetSyncInterval(sm.config.MomentSyncInterval)
	sm.momentSyncer.SetBatchSize(sm.config.BatchSize)

	log.Println("同步器初始化完成")
	return nil
}

// startSyncers 启动同步器
func (sm *SyncerManager) startSyncers() error {
	// 启动Moment同步器
	go func() {
		if err := sm.momentSyncer.Start(sm.ctx); err != nil {
			log.Printf("Moment同步器启动失败: %v", err)
		}
	}()

	// 等待一小段时间确保同步器启动
	time.Sleep(100 * time.Millisecond)

	log.Println("所有同步器启动完成")
	return nil
}

// stopSyncers 停止同步器
func (sm *SyncerManager) stopSyncers() {
	if sm.momentSyncer != nil {
		sm.momentSyncer.Stop()
	}

	log.Println("所有同步器已停止")
}

// IsRunning 检查管理器是否在运行
func (sm *SyncerManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// GetStatus 获取同步器状态
func (sm *SyncerManager) GetStatus() *SyncerStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status := &SyncerStatus{
		Running:   sm.running,
		StartTime: sm.startTime,
		Uptime:    time.Since(sm.startTime),
	}

	if sm.momentSyncer != nil {
		status.MomentSyncStats = sm.momentSyncer.GetSyncStats()
	}

	return status
}

// SyncerStatus 同步器状态
type SyncerStatus struct {
	Running         bool          `json:"running"`
	StartTime       time.Time     `json:"start_time"`
	Uptime          time.Duration `json:"uptime"`
	MomentSyncStats *SyncStats    `json:"moment_sync_stats,omitempty"`
}

// UpdateConfig 更新配置
func (sm *SyncerManager) UpdateConfig(config *SyncerConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	sm.config = config

	// 如果同步器正在运行，更新其配置
	if sm.running && sm.momentSyncer != nil {
		sm.momentSyncer.SetSyncInterval(config.MomentSyncInterval)
		sm.momentSyncer.SetBatchSize(config.BatchSize)
	}

	log.Println("同步器配置已更新")
	return nil
}

// GetConfig 获取当前配置
func (sm *SyncerManager) GetConfig() *SyncerConfig {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 返回配置副本
	config := *sm.config
	return &config
}

// TriggerSync 手动触发同步
func (sm *SyncerManager) TriggerSync(syncType string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.running {
		return fmt.Errorf("同步器管理器未运行")
	}

	switch syncType {
	case "moment":
		if sm.momentSyncer != nil {
			go sm.momentSyncer.syncBatch(sm.ctx)
			log.Println("手动触发Moment同步")
			return nil
		}
		return fmt.Errorf("Moment同步器未初始化")
	case "all":
		if sm.momentSyncer != nil {
			go sm.momentSyncer.syncBatch(sm.ctx)
		}
		log.Println("手动触发所有同步器")
		return nil
	default:
		return fmt.Errorf("未知的同步类型: %s", syncType)
	}
}

// GetMomentSyncer 获取Moment同步器实例
func (sm *SyncerManager) GetMomentSyncer() *MomentSyncer {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.momentSyncer
}

// HealthCheck 健康检查
func (sm *SyncerManager) HealthCheck() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.running {
		return fmt.Errorf("同步器管理器未运行")
	}

	// 检查各个同步器的健康状态
	// 这里可以添加更详细的健康检查逻辑

	return nil
}
