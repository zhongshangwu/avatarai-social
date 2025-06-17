# 活动管理 API 文档

本文档描述了活动管理相关的 API 接口，包括标签和主题的管理功能。

## 基础路径

所有活动管理 API 的基础路径为：`/api/activity`

## 标签管理 API

### 1. 获取指定 URI 的标签

**GET** `/api/activity/tags`

**查询参数：**
- `uri` (必需): 活动的 URI

**响应示例：**
```json
{
  "tags": [
    {
      "id": "tag_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id",
      "tag": "技术",
      "creator": "did:example:creator",
      "createdAt": 1640995200,
      "deleted": false
    }
  ]
}
```

### 2. 根据标签获取 moments

**GET** `/api/activity/tags/moments`

**查询参数：**
- `tag` (必需): 标签名称
- `limit` (可选): 返回数量限制，默认 20
- `cursor` (可选): 分页游标

**响应示例：**
```json
{
  "moments": [
    {
      "id": "moment_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id_1",
      "text": "这是一个关于技术的 moment",
      "creator": "did:example:creator",
      "createdAt": 1640995200
    }
  ],
  "cursor": "1640995200"
}
```

### 3. 创建或获取标签

**POST** `/api/activity/tags`

**请求体：**
```json
{
  "tag": "新标签"
}
```

**响应示例：**
```json
{
  "tag": {
    "id": "tag_id_new",
    "tag": "新标签",
    "creator": "did:example:creator",
    "createdAt": 1640995200,
    "deleted": false
  }
}
```

### 4. 获取标签使用次数

**GET** `/api/activity/tags/usage`

**查询参数：**
- `tag` (必需): 标签名称

**响应示例：**
```json
{
  "tag": "技术",
  "count": 42
}
```

### 5. 获取热门标签

**GET** `/api/activity/tags/popular`

**查询参数：**
- `limit` (可选): 返回数量限制，默认 10

**响应示例：**
```json
{
  "tags": [
    {
      "name": "技术",
      "count": 100,
      "type": "tag"
    },
    {
      "name": "生活",
      "count": 85,
      "type": "tag"
    }
  ]
}
```

## 主题管理 API

### 1. 获取指定 URI 的主题

**GET** `/api/activity/topics`

**查询参数：**
- `uri` (必需): 活动的 URI

**响应示例：**
```json
{
  "topics": [
    {
      "id": "topic_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id",
      "topic": "人工智能",
      "creator": "did:example:creator",
      "createdAt": 1640995200,
      "deleted": false
    }
  ]
}
```

### 2. 根据主题获取 moments

**GET** `/api/activity/topics/moments`

**查询参数：**
- `topic` (必需): 主题名称
- `limit` (可选): 返回数量限制，默认 20
- `cursor` (可选): 分页游标

**响应示例：**
```json
{
  "moments": [
    {
      "id": "moment_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id_1",
      "text": "关于人工智能的讨论",
      "creator": "did:example:creator",
      "createdAt": 1640995200
    }
  ],
  "cursor": "1640995200"
}
```

### 3. 创建或获取主题

**POST** `/api/activity/topics`

**请求体：**
```json
{
  "topic": "新主题"
}
```

**响应示例：**
```json
{
  "topic": {
    "id": "topic_id_new",
    "topic": "新主题",
    "creator": "did:example:creator",
    "createdAt": 1640995200,
    "deleted": false
  }
}
```

### 4. 获取主题使用次数

**GET** `/api/activity/topics/usage`

**查询参数：**
- `topic` (必需): 主题名称

**响应示例：**
```json
{
  "topic": "人工智能",
  "count": 58
}
```

### 5. 获取热门主题

**GET** `/api/activity/topics/popular`

**查询参数：**
- `limit` (可选): 返回数量限制，默认 10

**响应示例：**
```json
{
  "topics": [
    {
      "name": "人工智能",
      "count": 120,
      "type": "topic"
    },
    {
      "name": "区块链",
      "count": 95,
      "type": "topic"
    }
  ]
}
```

## 综合管理 API

### 1. 获取活动的完整元数据

**GET** `/api/activity/metadata`

**查询参数：**
- `uri` (必需): 活动的 URI

**响应示例：**
```json
{
  "tags": [
    {
      "id": "tag_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id",
      "tag": "技术",
      "creator": "did:example:creator",
      "createdAt": 1640995200,
      "deleted": false
    }
  ],
  "topics": [
    {
      "id": "topic_id_1",
      "uri": "at://did:example/app.vtri.activity.moment/moment_id",
      "topic": "人工智能",
      "creator": "did:example:creator",
      "createdAt": 1640995200,
      "deleted": false
    }
  ]
}
```

### 2. 同步活动元数据

**POST** `/api/activity/metadata`

**请求体：**
```json
{
  "uri": "at://did:example/app.vtri.activity.moment/moment_id",
  "tags": ["技术", "编程"],
  "topics": ["人工智能", "机器学习"]
}
```

**响应示例：**
```json
{
  "message": "同步成功"
}
```

### 3. 删除活动元数据

**DELETE** `/api/activity/metadata`

**查询参数：**
- `uri` (必需): 活动的 URI

**响应示例：**
```json
{
  "message": "删除成功"
}
```

## 认证要求

- 所有创建、更新、删除操作需要用户认证 (`mustAuth: true`)
- 查询操作不强制要求认证 (`mustAuth: false`)，但建议提供认证以获得更好的体验

## 错误响应

所有 API 在出错时会返回以下格式的错误响应：

```json
{
  "error": "错误描述信息"
}
```

常见的 HTTP 状态码：
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 需要认证
- `500 Internal Server Error`: 服务器内部错误