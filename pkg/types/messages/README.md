# Messages 模块

这个模块定义了聊天系统中所有消息类型的结构和处理逻辑。

## 主要特性

- **基于 msgType 的消息体解析**：根据消息类型自动选择正确的消息体结构进行反序列化
- **完整的消息类型支持**：支持文本、图片、视频、文件、音频、AI聊天、富文本、表情包、卡片、系统消息、删除消息和RTC通话等12种消息类型
- **类型安全**：使用接口和类型断言确保消息体类型的正确性
- **扩展性强**：易于添加新的消息类型

## 消息类型

| 类型 | 值 | 描述 | 消息体类型 |
|------|----|----- |-----------|
| MessageTypeText | 1 | 文本消息 | TextMsgBody |
| MessageTypePost | 2 | 富文本消息 | PostMsgBody |
| MessageTypeImage | 3 | 图片消息 | ImageMsgBody |
| MessageTypeFile | 4 | 文件消息 | FileMsgBody |
| MessageTypeAudio | 5 | 音频消息 | AudioMsgBody |
| MessageTypeVideo | 6 | 视频消息 | VideoMsgBody |
| MessageTypeSticker | 7 | 表情包消息 | StickerMsgBody |
| MessageTypeCard | 8 | 卡片消息 | CardMsgBody |
| MessageTypeAIChat | 9 | AI聊天消息 | AIChatMsgBody |
| MessageTypeSystem | 10 | 系统消息 | SystemMsgBody |
| MessageTypeDelete | 11 | 删除消息 | DeleteMsgBody |
| MessageTypeRTC | 12 | RTC通话消息 | RTCMsgBody |

## 使用示例

### 1. 文本消息

```go
import "encoding/json"

// JSON 数据
jsonData := `{
    "roomId": "room123",
    "msgType": 1,
    "body": {
        "text": "Hello, World!"
    },
    "senderId": "user123"
}`

// 反序列化
var event SendMsgEvent
err := json.Unmarshal([]byte(jsonData), &event)
if err != nil {
    log.Fatal(err)
}

// 类型断言获取具体的消息体
if textBody, ok := event.Body.(*TextMsgBody); ok {
    fmt.Println("文本内容:", textBody.Text)
}
```

### 2. 图片消息

```go
jsonData := `{
    "roomId": "room123",
    "msgType": 3,
    "body": {
        "image_cid": "img123",
        "width": 800,
        "height": 600,
        "alt": "测试图片"
    },
    "senderId": "user123"
}`

var event SendMsgEvent
err := json.Unmarshal([]byte(jsonData), &event)
if err != nil {
    log.Fatal(err)
}

if imageBody, ok := event.Body.(*ImageMsgBody); ok {
    fmt.Printf("图片: %s (%dx%d)\n", imageBody.ImageCID, imageBody.Width, imageBody.Height)
}
```

### 3. AI聊天消息

```go
jsonData := `{
    "roomId": "room123",
    "msgType": 9,
    "body": {
        "messageItems": [
            {
                "type": "message",
                "role": "user",
                "content": [
                    {
                        "type": "input_text",
                        "text": "Hello AI!"
                    }
                ]
            }
        ]
    },
    "senderId": "user123"
}`

var event SendMsgEvent
err := json.Unmarshal([]byte(jsonData), &event)
if err != nil {
    log.Fatal(err)
}

if aiChatBody, ok := event.Body.(*AIChatMsgBody); ok {
    fmt.Printf("AI聊天消息包含 %d 个消息项\n", len(aiChatBody.MessageItems))
}
```

### 4. 卡片消息

```go
jsonData := `{
    "roomId": "room123",
    "msgType": 8,
    "body": {
        "card_type": "link",
        "title": "测试链接",
        "description": "这是一个测试链接卡片",
        "url": "https://example.com",
        "actions": [
            {
                "type": "button",
                "label": "点击访问",
                "url": "https://example.com"
            }
        ]
    },
    "senderId": "user123"
}`

var event SendMsgEvent
err := json.Unmarshal([]byte(jsonData), &event)
if err != nil {
    log.Fatal(err)
}

if cardBody, ok := event.Body.(*CardMsgBody); ok {
    fmt.Printf("卡片: %s - %s\n", cardBody.Title, cardBody.Description)
}
```

## 消息体接口

所有消息体类型都实现了 `SendMsgBody` 接口：

```go
type SendMsgBody interface {
    isSendMsgBody()
}
```

## 错误处理

当遇到不支持的消息类型或无效的消息体时，`UnmarshalJSON` 方法会返回相应的错误：

- `不支持的消息类型: <type>` - 当 msgType 不在支持的范围内时
- `解析消息体失败: <error>` - 当消息体格式不正确时

## 扩展新的消息类型

要添加新的消息类型，需要：

1. 在 `constants.go` 中添加新的 `MessageType` 常量
2. 创建新的消息体结构体并实现 `SendMsgBody` 接口
3. 在 `SendMsgEvent.UnmarshalJSON` 方法中添加对应的 case
4. 添加相应的测试用例

示例：

```go
// 1. 添加新的消息类型
const MessageTypeLocation MessageType = 13

// 2. 创建消息体结构体
type LocationMsgBody struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Address   string  `json:"address,omitempty"`
}

func (l *LocationMsgBody) isSendMsgBody() {}

// 3. 在 UnmarshalJSON 中添加 case
case MessageTypeLocation:
    body = &LocationMsgBody{}
```

## 测试

运行测试：

```bash
go test ./pkg/types/messages -v
```

测试覆盖了所有消息类型的序列化/反序列化，以及错误情况的处理。

## AI Chat 主要结构类型

#### 输入项 (InputItem)
- ✅ EasyInputMessage (已重构为类型安全)
- ✅ InputMessage (完整的消息结构)
- ✅ ItemReferenceParam
- ✅ FunctionToolCall

#### 输出项 (OutputItem)
- ✅ OutputMessage
- ✅ FunctionToolCall
- ✅ ReasoningItem
- ✅ FileSearchToolCall (已重构Attributes为类型安全)
- ✅ WebSearchToolCall
- ✅ CodeInterpreterToolCall
- ✅ ComputerToolCall
- ✅ FunctionToolCallOutput
- ✅ ComputerToolCallOutput

#### 内容类型 (Content)
- ✅ InputTextContent
- ✅ InputImageContent
- ✅ InputFileContent
- ✅ OutputTextContent
- ✅ RefusalContent

#### 注释类型 (Annotation)
- ✅ FileCitationBody
- ✅ UrlCitationBody
- ✅ FilePathBody

#### 计算机操作 (ComputerAction)
- ✅ ClickAction (含Button字段)
- ✅ DoubleClickAction
- ✅ DragAction
- ✅ KeyPressAction
- ✅ MoveAction
- ✅ ScreenshotAction
- ✅ ScrollAction
- ✅ TypeAction
- ✅ WaitAction

#### 工具调用结果
- ✅ ComputerScreenshotResult
- ✅ ComputerActionResult
- ✅ CodeInterpreterTextOutput
- ✅ CodeInterpreterFileOutput (含文件详情)

### 事件类型
所有主要的响应事件类型都已在 consts.go 和相应的事件结构中定义，包括:
- 函数调用事件
- 文件搜索事件
- Web搜索事件
- 代码解释器事件
- 计算机使用事件
- 音频相关事件
