# AvatarAI Social 架构设计

## 概述

AvatarAI Social 是一个基于 ATProto 协议的社交媒体平台，支持用户创建 moment（帖子）、上传文件，并自动同步到 ATProto PDS。

## 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │   Web Frontend  │    │   Mobile Apps   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │      API Gateway         │
                    │   (Echo HTTP Server)     │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    Service Layer         │
                    │  - MomentService         │
                    │  - FileService           │
                    │  - UserService           │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │  Repository Layer        │
                    │  - MomentRepository      │
                    │  - FileRepository        │
                    │  - OAuthRepository       │
                    └─────────────┬─────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │     Database             │
                    │  (PostgreSQL/MySQL)      │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────▼───────┐    ┌─────────▼───────┐    ┌─────────▼───────┐
│ Syncer Manager  │    │  ATProto PDS    │    │  File Storage   │
│ - MomentSyncer  │    │                 │    │                 │
│ - Background    │    │                 │    │                 │
│   Processing    │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 核心组件

### 1. API 层 (pkg/api)

**职责**: 处理 HTTP 请求，路由分发，认证授权

**主要文件**:
- `apiserver.go`: HTTP 服务器配置
- `handlers/moment_handler.go`: Moment 相关 API 处理器
- `middleware/`: 中间件（认证、日志等）

**主要接口**:
- `POST /api/v1/moments`: 创建 moment
- `GET /api/v1/moments/:uri`: 获取 moment 详情
- `GET /api/v1/moments/feed`: 获取 moment feed
- `POST /api/v1/files/upload`: 上传文件
- `GET /api/v1/files/:id`: 获取文件信息

### 2. 服务层 (pkg/services)

**职责**: 业务逻辑处理，事务管理，外部服务调用

**主要服务**:

#### MomentService

- 创建 moment
- 查询 moment
- 处理嵌入内容（图片、视频、外部链接）
- 管理回复关系

#### FileService
- 文件上传处理
- 文件类型验证
- 文件存储管理
- 生成访问 URL

### 3. 数据访问层 (pkg/repositories)

**职责**: 数据库操作，数据模型定义

**主要 Repository**:
- `MomentRepository`: Moment 数据操作
- `FileRepository`: 文件数据操作
- `OAuthRepository`: OAuth 会话管理
- `AtpRepository`: ATProto 记录管理

**数据模型**:
- `Moment`: 核心 moment 数据
- `MomentImage/Video/External`: 嵌入内容
- `UploadFile`: 文件记录
- `OAuthSession`: OAuth 会话
- `AtpRecord`: ATProto 记录

### 4. 同步器 (pkg/pds/syncers)

**职责**: 后台数据同步，ATProto PDS 集成

#### SyncerManager
- 管理所有同步器
- 配置管理
- 健康检查
- 手动触发同步

#### MomentSyncer
- 监控本地 moment 变化
- 批量同步到 ATProto PDS
- 错误重试机制
- 同步状态跟踪

**同步流程**:
1. 定时扫描待同步的 moment
2. 按用户分组处理
3. 构建 ATProto 记录
4. 调用 PDS API 同步
5. 更新本地同步状态

### 5. ATProto 集成 (pkg/atproto)

**职责**: ATProto 协议实现，PDS 通信

**主要组件**:
- `XrpcClient`: XRPC 客户端
- `OAuth`: OAuth 认证流程
- `vtri/`: 自定义 lexicon 定义

## 数据流

### 1. 创建 Moment 流程

```
Client Request
    ↓
API Handler (认证检查)
    ↓
MomentService.CreateMoment()
    ↓
├─ 验证用户 OAuth 会话
├─ 构建 ATProto 记录
├─ 调用 PDS API (暂时模拟)
└─ 保存到本地数据库
    ↓
返回响应给客户端
    ↓
后台 MomentSyncer 检测到新数据
    ↓
同步到 ATProto PDS
```

### 2. 文件上传流程

```
Client Upload
    ↓
API Handler (认证检查)
    ↓
FileService.UploadFile()
    ↓
├─ 文件类型验证
├─ 文件大小检查
├─ 上传到 ATProto Blob Storage (暂时模拟)
└─ 保存文件记录到数据库
    ↓
返回文件信息和 LexBlob
```

### 3. 同步流程

```
定时器触发
    ↓
MomentSyncer.syncBatch()
    ↓
├─ 查询待同步 moment
├─ 按用户分组
└─ 对每个用户:
    ├─ 获取 OAuth 会话
    ├─ 创建 XRPC 客户端
    ├─ 构建 ATProto 记录
    ├─ 调用 PDS API
    └─ 更新同步状态
```

## 配置管理

### 同步器配置
```go
type SyncerConfig struct {
    MomentSyncInterval time.Duration // 同步间隔
    BatchSize          int           // 批处理大小
    MaxRetries         int           // 最大重试次数
    RetryDelay         time.Duration // 重试延迟
    EnableMetrics      bool          // 启用指标
    LogLevel           string        // 日志级别
}
```

### 数据库配置
- 支持 PostgreSQL、MySQL、SQLite
- 连接池配置
- 事务管理

## 错误处理

### API 层
- HTTP 状态码标准化
- 错误消息国际化
- 请求验证

### 服务层
- 业务异常处理
- 事务回滚
- 外部服务调用重试

### 同步器
- 失败重试机制
- 错误统计和报告
- 降级策略

## 监控和日志

### 日志
- 结构化日志 (slog)
- 不同级别日志
- 请求追踪

### 指标
- 同步成功/失败统计
- API 响应时间
- 数据库连接状态

### 健康检查
- 数据库连接检查
- 同步器状态检查
- 外部服务可用性检查

## 扩展性设计

### 水平扩展
- 无状态 API 服务器
- 数据库读写分离
- 缓存层 (Redis)

### 同步器扩展
- 支持多种同步器类型
- 插件化架构
- 动态配置更新

### 存储扩展
- 对象存储集成 (S3, MinIO)
- CDN 集成
- 多地域部署

## 安全考虑

### 认证授权
- OAuth 2.0 + DPoP
- JWT Token 管理
- 权限控制

### 数据安全
- 数据库加密
- 传输加密 (TLS)
- 敏感信息脱敏

### API 安全
- 请求限流
- 输入验证
- CORS 配置

## 部署架构

### 容器化
- Docker 镜像构建
- Kubernetes 部署
- 配置管理 (ConfigMap/Secret)

### 环境分离
- 开发环境
- 测试环境
- 生产环境

### CI/CD
- 自动化测试
- 自动化部署
- 回滚策略

## 未来规划

### 功能扩展
- 实时通知系统
- 搜索功能
- 推荐算法
- 内容审核

### 性能优化
- 缓存策略优化
- 数据库查询优化
- 异步处理优化

### 协议支持
- 完整 ATProto 协议实现
- 联邦化支持
- 跨平台互操作