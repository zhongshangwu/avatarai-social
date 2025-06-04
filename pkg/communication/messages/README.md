# Avatar AI Social Chat Protocol Specification

这是 Avatar AI Social 聊天协议的完整规格文档。本协议定义了一个功能丰富的聊天系统，支持多种消息类型、AI 聊天交互、富文本内容和实时事件流。

## 概述

本协议基于 JSON Schema 定义，提供了类型安全的消息传递机制，支持：

- **12种消息类型**：文本、图片、视频、文件、音频、AI聊天、富文本、表情包、卡片、系统消息、删除消息和RTC通话
- **AI聊天集成**：完整的AI助手交互，包括工具调用、推理过程和多模态内容
- **富文本支持**：支持格式化文本、链接、@用户、图片、视频、表情、代码块等
- **实时事件流**：基于事件的实时通信机制
- **房间和线程管理**：支持群聊、私聊和话题组织

## 核心数据结构

### 1. 房间 (Room)

房间是消息交换的基本容器，支持单聊、群聊。

```json
{
  "id": "room_123",
  "title": "项目讨论组",
  "type": "group",
  "last_mid": "msg_456",
  "participants": ["user_1", "user_2", "ai_assistant"],
  "created_at": 1703123456,
  "updated_at": 1703123456,
  "deleted": false
}
```

### 2. 消息 (Message)

消息是协议的核心实体，包含元数据和具体内容。

```json
{
  "id": "msg_123",
  "room_id": "room_123",
  "thread_id": "thread_456",
  "msg_type": 1,
  "content": {
    "text": "Hello, World!"
  },
  "receiver_id": "user_2",
  "sender_id": "user_1",
  "quote_mid": "",
  "sender_at": 1703123456,
  "created_at": 1703123456,
  "updated_at": 1703123456,
  "deleted": false,
  "external_id": "ext_789"
}
```

### 3. 线程 (Thread)

线程用于组织相关消息，支持连续上下文和隔离上下文两种模式。

```json
{
  "id": "thread_123",
  "room_id": "room_123",
  "title": "AI助手讨论",
  "context_mode": "continuous",
  "root_mid": "msg_100",
  "parent_thread_id": "",
  "created_at": 1703123456,
  "updated_at": 1703123456,
  "deleted": false
}
```

## 消息类型详解

### 消息类型枚举

| 类型值 | 名称        | 描述             | 内容结构                |
| ------ | ----------- | ---------------- | ----------------------- |
| 0      | Unspecified | 未指定           | -                       |
| 1      | Text        | 文本消息         | `TextMessageContent`    |
| 2      | Post        | 富文本消息       | `PostMessageContent`    |
| 3      | Image       | 图片消息         | `ImageMessageContent`   |
| 4      | File        | 文件消息         | `FileMessageContent`    |
| 5      | Audio       | 音频消息         | `AudioMessageContent`   |
| 6      | Video       | 视频消息         | `VideoMessageContent`   |
| 7      | Sticker     | 表情包消息       | `StickerMessageContent` |
| 8      | Card        | 卡片消息(未实现) | `CardMessageContent`    |
| 9      | AIChat      | AI聊天消息       | `AIChatMessageContent`  |
| 10     | System      | 系统消息         | `SystemMessageContent`  |
| 11     | Delete      | 删除消息         | `DeleteMessageContent`  |
| 12     | RTC         | RTC通话消息      | `RTCMessageContent`     |

### 1. 文本消息 (Type: 1)

最基础的消息类型，包含纯文本内容。

```json
{
  "msg_type": 1,
  "content": {
    "text": "这是一条文本消息"
  }
}
```

### 2. 图片消息 (Type: 3)

支持图片分享，包含尺寸和替代文本信息。

```json
{
  "msg_type": 3,
  "content": {
    "image_url": "https://example.com/image.jpg",
    "image_cid": "img_123",
    "width": 800,
    "height": 600,
    "alt": "美丽的风景照片"
  }
}
```

### 3. AI聊天消息 (Type: 9)

最复杂的消息类型，支持AI助手的完整交互流程。

```json
{
  "msg_type": 9,
  "content": {
    "message": {
      "id": "ai_msg_123",
      "message_id": "external_456",
      "role": "assistant",
      "altText": "AI助手回复了一条消息",
      "messageItems": [
        {
          "type": "message",
          "role": "assistant",
          "content": [
            {
              "type": "output_text",
              "text": "我可以帮助您解决这个问题。"
            }
          ],
          "status": "completed"
        }
      ],
      "interruptType": 0,
      "status": "completed",
      "creator": "ai_assistant",
      "createdAt": 1703123456,
      "updatedAt": 1703123456
    }
  }
}
```

### 4. 富文本消息 (Type: 2)

支持复杂的格式化内容，包括文本样式、链接、@用户、图片、视频等。

```json
{
  "msg_type": 2,
  "content": {
    "title": "项目更新",
    "content": [
      [
        {
          "tag": "text",
          "text": "项目进展顺利，",
          "style": ["bold"]
        },
        {
          "tag": "at",
          "user_id": "user_123"
        },
        {
          "tag": "text",
          "text": " 请查看最新代码。"
        }
      ],
      [
        {
          "tag": "img",
          "image_key": "img_456"
        }
      ]
    ]
  }
}
```

## AI聊天系统详解

AI聊天系统是协议的核心特性，支持复杂的AI助手交互。

### 消息项类型 (MessageItem)

AI聊天消息包含多种消息项：

#### 1. 输入消息 (InputMessage)

用户向AI发送的消息。

```json
{
  "type": "message",
  "role": "user",
  "content": [
    {
      "type": "input_text",
      "text": "请帮我分析这张图片"
    },
    {
      "type": "input_image",
      "fileId": "file_123",
      "detail": "high"
    }
  ]
}
```

#### 2. 输出消息 (OutputMessage)

AI助手的回复消息。

```json
{
  "id": "output_123",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "output_text",
      "text": "这张图片显示了一个美丽的日落场景。",
      "annotations": [
        {
          "type": "file_citation",
          "fileId": "file_123",
          "index": 0
        }
      ]
    }
  ],
  "status": "completed"
}
```

#### 3. 工具调用 (Tool Calls)

AI助手可以调用各种工具来完成任务。

##### 函数工具调用

```json
{
  "id": "tool_123",
  "type": "tool_call",
  "name": "get_weather",
  "status": "completed",
  "arguments": "{\"location\": \"北京\"}"
}
```

##### 文件搜索工具调用

```json
{
  "id": "search_123",
  "type": "file_search_call",
  "status": "completed",
  "queries": ["机器学习", "深度学习"],
  "results": [
    {
      "fileId": "file_456",
      "text": "机器学习是人工智能的一个分支...",
      "filename": "ml_guide.pdf",
      "score": 0.95
    }
  ]
}
```

##### 代码解释器工具调用

```json
{
  "id": "code_123",
  "type": "code_interpreter_call",
  "code": "import numpy as np\nprint(np.array([1, 2, 3]))",
  "status": "completed",
  "results": [
    {
      "type": "text",
      "text": "[1 2 3]"
    }
  ]
}
```

##### 计算机工具调用

```json
{
  "id": "computer_123",
  "type": "computer_call",
  "callId": "call_456",
  "action": {
    "type": "click",
    "button": "left",
    "x": 100,
    "y": 200
  },
  "pendingSafetyChecks": [],
  "status": "completed"
}
```

### 推理过程 (Reasoning)

AI助手的思考过程可以被记录和展示。

```json
{
  "type": "reasoning",
  "id": "reasoning_123",
  "summary": [
    {
      "type": "summary_text",
      "text": "用户询问了关于机器学习的问题，我需要提供准确和有用的信息。"
    }
  ],
  "status": "completed"
}
```

## 事件系统

协议支持实时事件流，用于传递消息状态变化和AI处理进度。

### 事件结构

```json
{
  "eventId": "event_123",
  "eventType": "ai_chat.output_text.delta",
  "event": {
    "itemId": "output_456",
    "outputIndex": 0,
    "contentIndex": 0,
    "delta": "这是增量文本"
  }
}
```

### 主要事件类型

#### 1. 消息发送事件

```json
{
  "eventId": "event_123",
  "eventType": "send_msg",
  "event": {
    "roomId": "room_123",
    "msgType": 1,
    "body": {
      "text": "Hello"
    },
    "senderId": "user_123",
    "receiverId": "user_456"
  }
}
```

#### 2. AI聊天事件

- `ai_chat.created`: AI聊天会话创建
- `ai_chat.in_progress`: AI正在处理
- `ai_chat.output_text.delta`: 文本增量更新
- `ai_chat.output_text.done`: 文本输出完成
- `ai_chat.completed`: AI聊天完成
- `ai_chat.failed`: AI聊天失败

#### 3. 工具调用事件

- `ai_chat.function_call_arguments.delta`: 函数参数增量
- `ai_chat.file_search_call.searching`: 文件搜索中
- `ai_chat.code_interpreter_call.interpreting`: 代码执行中
- `ai_chat.computer_call.in_progress`: 计算机操作进行中

## 富文本节点详解

富文本系统支持复杂的内容格式化。

### 节点类型

#### 1. 文本节点

```json
{
  "tag": "text",
  "text": "这是加粗的文本",
  "style": ["bold", "italic"],
  "un_escape": false
}
```

#### 2. 链接节点

```json
{
  "tag": "a",
  "text": "点击这里",
  "href": "https://example.com",
  "style": ["underline"]
}
```

#### 3. @用户节点

```json
{
  "tag": "at",
  "user_id": "user_123",
  "style": ["bold"]
}
```

#### 4. 图片节点

```json
{
  "tag": "img",
  "image_key": "img_123"
}
```

#### 5. 视频节点

```json
{
  "tag": "media",
  "file_key": "video_123",
  "image_key": "thumb_456"
}
```

#### 6. 代码块节点

```json
{
  "tag": "code_block",
  "language": "python",
  "text": "print('Hello, World!')"
}
```

#### 7. Markdown节点

```json
{
  "tag": "md",
  "text": "# 标题\n\n这是**加粗**文本。"
}
```

## 错误处理

协议定义了完整的错误处理机制。

### 错误代码

```json
{
  "error": {
    "code": "invalid_image_format",
    "message": "不支持的图片格式，请使用 JPEG 或 PNG"
  }
}
```

### 常见错误代码

- `server_error`: 服务器内部错误
- `rate_limit_exceeded`: 请求频率超限
- `invalid_prompt`: 无效的提示词
- `invalid_image_format`: 无效的图片格式
- `image_too_large`: 图片文件过大
- `vector_store_timeout`: 向量存储超时

## 使用示例

### 1. 发送文本消息

```json
{
  "eventType": "send_msg",
  "event": {
    "roomId": "room_123",
    "msgType": 1,
    "body": {
      "text": "大家好！"
    },
    "senderId": "user_123",
    "receiverId": "room_123"
  }
}
```

### 2. 发送AI聊天消息

```json
{
  "eventType": "send_msg",
  "event": {
    "roomId": "room_123",
    "msgType": 9,
    "body": {
      "role": "user",
      "message_id": "msg_456",
      "messageItems": [
        {
          "type": "message",
          "role": "user",
          "content": [
            {
              "type": "input_text",
              "text": "请帮我写一个Python函数来计算斐波那契数列"
            }
          ]
        }
      ],
      "metadata": {}
    },
    "senderId": "user_123",
    "receiverId": "ai_assistant"
  }
}
```

### 3. 发送富文本消息

```json
{
  "eventType": "send_msg",
  "event": {
    "roomId": "room_123",
    "msgType": 2,
    "body": {
      "title": "会议纪要",
      "content": [
        [
          {
            "tag": "text",
            "text": "今日会议要点：",
            "style": ["bold"]
          }
        ],
        [
          {
            "tag": "text",
            "text": "1. 项目进度更新 - "
          },
          {
            "tag": "at",
            "user_id": "user_456"
          },
          {
            "tag": "text",
            "text": " 负责"
          }
        ],
        [
          {
            "tag": "code_block",
            "language": "javascript",
            "text": "const progress = calculateProgress();"
          }
        ]
      ]
    },
    "senderId": "user_123",
    "receiverId": "room_123"
  }
}
```

## 最佳实践

### 1. 消息ID管理

- 使用UUID或雪花算法生成唯一ID
- 保持ID的可读性和可追踪性
- 在分布式环境中确保ID的全局唯一性

### 2. 错误处理

- 始终检查消息格式的有效性
- 提供有意义的错误信息
- 实现重试机制处理临时性错误

### 3. 性能优化

- 对大型富文本内容进行分页
- 使用增量更新减少数据传输
- 实现消息缓存机制

### 4. 安全考虑

- 验证用户权限
- 过滤恶意内容
- 实现消息加密（如需要）

## 版本兼容性

本协议遵循语义化版本控制：

- **主版本号**：不兼容的API变更
- **次版本号**：向后兼容的功能性新增
- **修订号**：向后兼容的问题修正

当前版本：`1.0.0`

## 扩展性

协议设计考虑了未来的扩展需求：

1. **新消息类型**：可以通过增加新的 `MessageType` 枚举值来支持
2. **新工具类型**：AI聊天系统支持插件式的工具扩展
3. **新事件类型**：事件系统支持自定义事件类型
4. **新富文本节点**：富文本系统支持新的节点类型

## 总结

Avatar AI Social Chat Protocol 提供了一个功能完整、类型安全、可扩展的聊天协议规格。通过 JSON Schema 的定义，确保了数据的一致性和可验证性，同时支持复杂的AI交互和富媒体内容。

协议的设计考虑了现代聊天应用的各种需求，从简单的文本消息到复杂的AI助手交互，都能得到很好的支持。通过事件驱动的架构，实现了实时性和响应性的平衡。

如需更多技术细节，请参考 `spec.json` 文件中的完整 JSON Schema 定义。
