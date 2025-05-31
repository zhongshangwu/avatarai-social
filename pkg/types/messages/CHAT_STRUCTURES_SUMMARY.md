### 主要结构类型:

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
