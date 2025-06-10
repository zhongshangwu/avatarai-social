package syncers

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"sync"
// 	"time"

// 	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
// )

// type MomentSyncer struct {
// 	metaStore    *repositories.MetaStore
// 	momentRepo   *repositories.MomentRepository
// 	oauthRepo    *repositories.OAuthRepository
// 	syncInterval time.Duration
// 	batchSize    int
// 	maxRetries   int
// 	retryDelay   time.Duration
// 	running      bool
// 	stopCh       chan struct{}
// 	mu           sync.RWMutex
// 	lastSyncTime time.Time
// 	syncStats    *SyncStats
// }

// // SyncStats 同步统计信息
// type SyncStats struct {
// 	TotalProcessed int64     `json:"total_processed"`
// 	TotalSuccess   int64     `json:"total_success"`
// 	TotalFailed    int64     `json:"total_failed"`
// 	LastSyncTime   time.Time `json:"last_sync_time"`
// 	LastError      string    `json:"last_error,omitempty"`
// }

// // SyncTask 同步任务
// type SyncTask struct {
// 	MomentURI   string                 `json:"moment_uri"`
// 	Did         string                 `json:"did"`
// 	Action      string                 `json:"action"` // create, update, delete
// 	Data        map[string]interface{} `json:"data"`
// 	RetryCount  int                    `json:"retry_count"`
// 	CreatedAt   time.Time              `json:"created_at"`
// 	ScheduledAt time.Time              `json:"scheduled_at"`
// }

// // NewMomentSyncer 创建新的moment同步器
// func NewMomentSyncer(metaStore *repositories.MetaStore) *MomentSyncer {
// 	return &MomentSyncer{
// 		metaStore:    metaStore,
// 		momentRepo:   repositories.NewMomentRepository(metaStore),
// 		oauthRepo:    repositories.NewOAuthRepository(metaStore),
// 		syncInterval: 30 * time.Second, // 默认30秒同步一次
// 		batchSize:    50,               // 每批处理50个
// 		maxRetries:   3,                // 最大重试3次
// 		retryDelay:   5 * time.Second,  // 重试延迟5秒
// 		stopCh:       make(chan struct{}),
// 		syncStats:    &SyncStats{},
// 	}
// }

// // Start 启动同步器
// func (s *MomentSyncer) Start(ctx context.Context) error {
// 	s.mu.Lock()
// 	if s.running {
// 		s.mu.Unlock()
// 		return fmt.Errorf("同步器已经在运行")
// 	}
// 	s.running = true
// 	s.mu.Unlock()

// 	log.Printf("启动Moment同步器，同步间隔: %v", s.syncInterval)

// 	ticker := time.NewTicker(s.syncInterval)
// 	defer ticker.Stop()

// 	// 启动时立即执行一次同步
// 	go s.syncBatch(ctx)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("收到上下文取消信号，停止同步器")
// 			return ctx.Err()
// 		case <-s.stopCh:
// 			log.Println("收到停止信号，停止同步器")
// 			return nil
// 		case <-ticker.C:
// 			go s.syncBatch(ctx)
// 		}
// 	}
// }

// // Stop 停止同步器
// func (s *MomentSyncer) Stop() {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	if !s.running {
// 		return
// 	}

// 	s.running = false
// 	close(s.stopCh)
// 	log.Println("Moment同步器已停止")
// }

// // syncBatch 批量同步处理
// func (s *MomentSyncer) syncBatch(ctx context.Context) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("同步批处理发生panic: %v", r)
// 		}
// 	}()

// 	log.Println("开始执行同步批处理")

// 	// 获取需要同步的moment
// 	pendingMoments, err := s.getPendingMoments()
// 	if err != nil {
// 		log.Printf("获取待同步moment失败: %v", err)
// 		s.updateSyncStats(0, 0, 1, err.Error())
// 		return
// 	}

// 	if len(pendingMoments) == 0 {
// 		log.Println("没有待同步的moment")
// 		return
// 	}

// 	log.Printf("找到 %d 个待同步的moment", len(pendingMoments))

// 	var successCount, failCount int64

// 	// 按用户分组处理，避免并发冲突
// 	userGroups := s.groupMomentsByUser(pendingMoments)

// 	for did, moments := range userGroups {
// 		processed, failed := s.syncUserMoments(ctx, did, moments)
// 		successCount += processed
// 		failCount += failed
// 	}

// 	s.updateSyncStats(successCount, failCount, 0, "")
// 	log.Printf("同步批处理完成，成功: %d, 失败: %d", successCount, failCount)
// }

// // getPendingMoments 获取待同步的moment
// func (s *MomentSyncer) getPendingMoments() ([]*repositories.Moment, error) {
// 	// 查询最近创建但未同步到PDS的moment
// 	// 这里需要添加一个sync_status字段来跟踪同步状态
// 	var moments []*repositories.Moment

// 	// 暂时获取最近的moment，实际应该根据sync_status字段过滤
// 	moments, err := s.momentRepo.GetLatestMoments(s.batchSize, "")
// 	if err != nil {
// 		return nil, fmt.Errorf("查询待同步moment失败: %w", err)
// 	}

// 	return moments, nil
// }

// // groupMomentsByUser 按用户分组moment
// func (s *MomentSyncer) groupMomentsByUser(moments []*repositories.Moment) map[string][]*repositories.Moment {
// 	groups := make(map[string][]*repositories.Moment)

// 	for _, moment := range moments {
// 		groups[moment.Creator] = append(groups[moment.Creator], moment)
// 	}

// 	return groups
// }

// // syncUserMoments 同步特定用户的moment
// func (s *MomentSyncer) syncUserMoments(ctx context.Context, did string, moments []*repositories.Moment) (int64, int64) {
// 	// 获取用户的OAuth会话
// 	oauthSession, err := s.oauthRepo.GetOAuthSessionByDID(did)
// 	if err != nil {
// 		log.Printf("获取用户 %s 的OAuth会话失败: %v", did, err)
// 		return 0, int64(len(moments))
// 	}

// 	if oauthSession == nil {
// 		log.Printf("用户 %s 没有有效的OAuth会话", did)
// 		return 0, int64(len(moments))
// 	}

// 	// 创建XRPC客户端 - 暂时跳过实际的客户端创建
// 	// 在实际实现中需要正确处理XRPC客户端

// 	var successCount, failCount int64

// 	for _, moment := range moments {
// 		if err := s.syncSingleMoment(ctx, moment, nil, oauthSession); err != nil {
// 			log.Printf("同步moment %s 失败: %v", moment.URI, err)
// 			failCount++
// 		} else {
// 			successCount++
// 		}
// 	}

// 	return successCount, failCount
// }

// // syncSingleMoment 同步单个moment
// func (s *MomentSyncer) syncSingleMoment(ctx context.Context, moment *repositories.Moment, client interface{}, session *repositories.OAuthSession) error {
// 	// 检查moment是否已经同步
// 	if moment.CID != "" {
// 		log.Printf("Moment %s 已经同步，跳过", moment.URI)
// 		return nil
// 	}

// 	// 构建ATProto记录
// 	record, err := s.buildAtprotoRecord(moment)
// 	if err != nil {
// 		return fmt.Errorf("构建ATProto记录失败: %w", err)
// 	}

// 	// 同步到PDS - 暂时模拟
// 	uri, cid, err := s.putRecordToPDS(ctx, client, session, record)
// 	if err != nil {
// 		return fmt.Errorf("同步到PDS失败: %w", err)
// 	}

// 	// 更新本地记录
// 	if err := s.updateMomentSyncStatus(moment.URI, uri, cid); err != nil {
// 		log.Printf("更新moment同步状态失败: %v", err)
// 		// 不返回错误，因为PDS同步已经成功
// 	}

// 	log.Printf("成功同步moment %s 到 %s", moment.URI, uri)
// 	return nil
// }

// // buildAtprotoRecord 构建ATProto记录
// func (s *MomentSyncer) buildAtprotoRecord(moment *repositories.Moment) (map[string]interface{}, error) {
// 	record := map[string]interface{}{
// 		"$type":     "app.vtri.activity.moment",
// 		"text":      moment.Text,
// 		"createdAt": moment.CreatedAt,
// 	}

// 	if len(moment.Langs) > 0 {
// 		record["langs"] = moment.Langs
// 	}

// 	if len(moment.Tags) > 0 {
// 		record["tags"] = moment.Tags
// 	}

// 	// 处理回复关系
// 	if moment.ReplyRoot != "" {
// 		record["reply"] = map[string]interface{}{
// 			"root": map[string]interface{}{
// 				"uri": moment.ReplyRoot,
// 				"cid": moment.ReplyRootCID,
// 			},
// 		}

// 		if moment.ReplyParent != "" {
// 			record["reply"].(map[string]interface{})["parent"] = map[string]interface{}{
// 				"uri": moment.ReplyParent,
// 				"cid": moment.ReplyParentCID,
// 			}
// 		}
// 	}

// 	// TODO: 处理嵌入内容（图片、视频等）
// 	// 这里需要查询相关的MomentImage, MomentVideo等表

// 	return record, nil
// }

// // putRecordToPDS 将记录提交到PDS
// func (s *MomentSyncer) putRecordToPDS(ctx context.Context, client interface{}, session *repositories.OAuthSession, record map[string]interface{}) (string, string, error) {
// 	// 生成记录键
// 	rkey := s.generateRecordKey()

// 	// 暂时模拟PDS调用，实际实现中需要使用真正的XRPC客户端
// 	uri := fmt.Sprintf("at://%s/app.vtri.activity.moment/%s", session.Did, rkey)
// 	cid := "bafyrei" + rkey // 简化的CID生成

// 	return uri, cid, nil
// }

// // generateRecordKey 生成记录键
// func (s *MomentSyncer) generateRecordKey() string {
// 	// 使用时间戳生成唯一键
// 	return fmt.Sprintf("%d", time.Now().UnixNano())
// }

// // updateMomentSyncStatus 更新moment同步状态
// func (s *MomentSyncer) updateMomentSyncStatus(momentURI, atprotoURI, cid string) error {
// 	updates := map[string]interface{}{
// 		"uri": atprotoURI,
// 		"cid": cid,
// 	}

// 	return s.momentRepo.UpdateMoment(momentURI, updates)
// }

// // updateSyncStats 更新同步统计
// func (s *MomentSyncer) updateSyncStats(success, failed, errors int64, lastError string) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	s.syncStats.TotalProcessed += success + failed
// 	s.syncStats.TotalSuccess += success
// 	s.syncStats.TotalFailed += failed
// 	s.syncStats.LastSyncTime = time.Now()

// 	if lastError != "" {
// 		s.syncStats.LastError = lastError
// 	}
// }

// // GetSyncStats 获取同步统计信息
// func (s *MomentSyncer) GetSyncStats() *SyncStats {
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()

// 	// 返回副本避免并发问题
// 	stats := *s.syncStats
// 	return &stats
// }

// // SetSyncInterval 设置同步间隔
// func (s *MomentSyncer) SetSyncInterval(interval time.Duration) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	s.syncInterval = interval
// }

// // SetBatchSize 设置批处理大小
// func (s *MomentSyncer) SetBatchSize(size int) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	s.batchSize = size
// }
