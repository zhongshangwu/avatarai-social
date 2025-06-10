# API Handlers 重构

## 概述

这个模块已经被重构为使用独立的handler结构，每个handler负责特定的功能域。

## Handler 结构

### 1. HealthHandler
- `HandleHealthz()` - 健康检查

### 2. OAuthHandler
- `HandleOAuthLogin()` - OAuth登录
- `HandleOAuthCallback()` - OAuth回调
- `HandleOAuthToken()` - 令牌交换
- `HandleOAuthRefresh()` - 令牌刷新
- `HandleOAuthLogout()` - 登出
- `HandleOAuthJWKS()` - JWKS端点
- `HandleOAuthClientMetadata()` - 客户端元数据
- `HandleAppReturn()` - 应用返回处理
- `HandleBskyPost()` - Bluesky发布

### 3. AvatarHandler
- `HandleAvatarProfile()` - 获取用户资料
- `HandleUpdateAvatarProfile()` - 更新用户资料

### 4. AsterHandler
- `HandleAsterProfile()` - Aster个人资料
- `HandleAsterUpdateProfile()` - 更新Aster资料
- `HandleAsterMint()` - Aster铸造

### 5. MomentHandler
- `HandleMomentCreate()` - 创建动态
- `HandleMomentDetail()` - 获取动态详情
- `HandleMomentFeed()` - 获取动态流

### 6. BlobHandler
- `UploadBlobHandler()` - 上传文件
- `GetUploadFilesHandler()` - 获取上传文件列表

### 7. MessageHandler
- `GetMessagesHistoryHandler()` - 获取消息历史

### 8. ChatHandler
- `ChatStream()` - 聊天流

## 使用方式

```go
// 创建handler实例
userHandler := handlers.NewAvatarHandler(config, metaStore)
oauthHandler := handlers.NewOAuthHandler(config, metaStore)

// 在路由中使用
router.GET("/avatar/profile", userHandler.HandleAvatarProfile)
router.POST("/oauth/login", oauthHandler.HandleOAuthLogin)
```

## 优势

1. **模块化**: 每个handler专注于特定功能
2. **可测试性**: 更容易进行单元测试
3. **可维护性**: 代码结构更清晰
4. **可扩展性**: 容易添加新的handler
5. **依赖注入**: 通过构造函数注入依赖

## 迁移指南

原来的方法调用：
```go
api.GET("/healthz", a.HandleHealthz)
```

现在的方法调用：
```go
api.GET("/healthz", a.healthHandler.HandleHealthz)
```