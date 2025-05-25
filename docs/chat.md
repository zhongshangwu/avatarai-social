# Realtime Chat API

基于 Websocket 的全双工实时聊天协议, 支持普通的聊天消息类型和 AIChat 消息类型。

## 获取历史消息

### 接口地址

> GET https://avatar.ai/api/chat/history

### 请求参数

- **before** (string) (Optional):
    - An message ID to list messages after, used in pagination.
- **after** (string) (Optional):
    - An message ID to list messages after, used in pagination.
- **limit** (integet) (Optional) Defaults to 20:
    - A limit on the number of messages to be returned. Limit can range between 1 and 100, and the default is 20.

### 响应内容

- **data**: (array[`Message`])
    - 返回的历史消息列表
- **first_id**: (string)
    - 第一条消息的 ID
- **last_id**: (string)
    - 最后一条消息的 ID
- **has_more**: (boolean)
    - 是否有更多的消息

### 消息结构定义

#### 通用消息结构

```json
{
    "id": "msg_1xxxxx",
    "msg_type": "text",
    "payload": {
        "text": "这是一个简单的纯文本"
    },
    "room_id": "room_123",
    "thread_id": "thread_xxxxx",
    "root_id": "msg_xxxx",
    "parent_id": "msg_xxxxx",
    "quote_id": "msg_xxxx",
    "sender_at": 1333,
    "create_time": 11111
}
```

#### 文本

```json
{
    "id": "msg_1xxxxx",
    "msg_type": "text",
    "payload": {
        "text": "这是一个简单的纯文本"
    }
}
```

#### AI 聊天

```json
{
    "id": "msg_1xxxxx",
    "msg_type": "ai_chat",
    "payload": {
        "chat_id": "chat_id",
        "message_id": "xxxx",
        "role": "user",
        "content": "xxxxxxxxxxxxxx",
        "extra": {
            "webpages": [

            ],
            "context": [

            ]
        },
        "metadata": {
            "usage": {

            }
        },
    }
}
```

#### AI 实时音视频聊天

```json
{
    "id": "msg_1xxxxx",
    "msg_type": "ai_rtc",
    "payload": {
        "session_id": "chat_id",
    }
}
```

## 实时聊天

> wss://backend.avatar.ai/api/chat/realtime

### 事件结构定义

#### 通用事件

```json
{
    "event_id": "event_abc",
    "type": "message.create",
    "message": {
        "id": "msg_xxxxx",
        "type": "message",
        "role": "user",
        "content": [
            {
                "type": "input_text",
                "text": "你好"
            }
        ]
    }
}
```

#### Client Event

##### send_msg.ai_chat (发送消息)

##### send_msg.text (发送消息)

##### ai_chat_response.cancel (取消正在进行的回复)

#### Server Event

##### ai_chat_response.created

##### ai_chat_response.completed

##### ai_chat_response.failed

##### ai_chat_response.in_complete

## 实时音视频聊天
