package messages

import (
	"encoding/json"
	"fmt"
)

type AIChatMessage struct {
	ID                string                   `json:"id"`                          // 此AI聊天消息的唯一标识符
	MID               string                   `json:"mid"`                         // 消息ID，引用 message.id
	Role              RoleType                 `json:"role"`                        // 消息角色: "user", "assistant", "system"
	Text              string                   `json:"text"`                        // 消息内容(对应的纯文本内容)
	MessageItems      []MessageItem            `json:"messageItems"`                // 模型生成的内容项数组
	InterruptType     int32                    `json:"interruptType"`               // 消息是否有被中断，默认为0
	Status            AiChatMessageStatus      `json:"status"`                      // 响应生成的状态: "completed", "failed", "in_progress", "incomplete"
	Error             *ResponseError           `json:"error,omitempty"`             // 错误信息
	Creator           string                   `json:"creator"`                     // 创建者DID (可能是用户DID，也可以是AssistantDID)
	CreatedAt         int64                    `json:"createdAt"`                   // 创建时间
	UpdatedAt         int64                    `json:"updatedAt"`                   // 更新时间
	IncompleteDetails *IncompleteDetails       `json:"incompleteDetails,omitempty"` // 响应不完整的详细信息
	Usage             *ResponseUsage           `json:"usage,omitempty"`             // 使用情况统计
	Tools             []map[string]interface{} `json:"tools,omitempty"`             // 模型可用的工具
	Metadata          map[string]interface{}   `json:"metadata,omitempty"`          // 响应的其他元数据
}

func (m *AIChatMessage) UnmarshalJSON(data []byte) error {
	type Alias AIChatMessage
	aux := &struct {
		MessageItems []json.RawMessage `json:"messageItems"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.MessageItems = make([]MessageItem, len(aux.MessageItems))
	for i, raw := range aux.MessageItems {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var item MessageItem
		switch typeName {
		case "message": // FIXME: 这里是否需要区分 input message 和 output message
			item = &OutputMessage{}
		case "tool_call":
			item = &FunctionToolCall{}
		case "function_call":
			item = &FunctionToolCall{}
		case "reasoning":
			item = &ReasoningItem{}
		case "file_search_call":
			item = &FileSearchToolCall{}
		case "web_search_call":
			item = &WebSearchToolCall{}
		case "code_interpreter_call":
			item = &CodeInterpreterToolCall{}
		case "computer_call":
			item = &ComputerToolCall{}
		case "function_call_output":
			item = &FunctionToolCallOutput{}
		case "computer_call_output":
			item = &ComputerToolCallOutput{}
		default:
			return fmt.Errorf("unknown output item type: %s", typeName)
		}

		if err := json.Unmarshal(raw, item); err != nil {
			return err
		}

		m.MessageItems[i] = item
	}

	return nil
}

type IncompleteDetails struct {
	Reason IncompleteReason `json:"reason"` // 响应不完整的原因: "max_output_tokens", "content_filter"
}

type InputContent interface {
	GetType() string
	isInputContent()
}

type InputMessage struct {
	Type    string         `json:"type,omitempty"`   // 消息输入的类型，始终设置为"message"
	Role    string         `json:"role"`             // 消息输入的角色: "user", "system", "developer"
	Status  string         `json:"status,omitempty"` // 项目的状态
	Content []InputContent `json:"content"`          // 内容可以是InputTextContent、InputImageContent或InputFileContent
}

func (i *InputMessage) GetType() string { return i.Type }
func (i *InputMessage) isInputItem()    {}
func (i *InputMessage) isMessageItem()  {}

func (m *InputMessage) UnmarshalJSON(data []byte) error {
	type Alias InputMessage
	aux := &struct {
		Content []json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Content = make([]InputContent, len(aux.Content))
	for i, raw := range aux.Content {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var content InputContent
		switch typeName {
		case "input_text":
			content = &InputTextContent{}
		case "input_image":
			content = &InputImageContent{}
		case "input_file":
			content = &InputFileContent{}
		default:
			return fmt.Errorf("未知的输入内容类型: %s", typeName)
		}

		if err := json.Unmarshal(raw, content); err != nil {
			return err
		}

		m.Content[i] = content
	}

	return nil
}

type InputTextContent struct {
	Type string `json:"type"` // 输入项的类型，始终为"input_text"
	Text string `json:"text"` // 模型的文本输入
}

func (i *InputTextContent) GetType() string { return i.Type }
func (i *InputTextContent) isInputContent() {}

type InputImageContent struct {
	Type     string `json:"type"`               // 输入项的类型，始终为"input_image"
	ImageURL string `json:"imageUrl,omitempty"` // 图像URL
	FileID   string `json:"fileId,omitempty"`   // 要发送到模型的文件的ID
	Detail   string `json:"detail"`             // 要发送到模型的图像的细节级别: "high", "low", "auto"
}

func (i *InputImageContent) GetType() string { return i.Type }
func (i *InputImageContent) isInputContent() {}

type InputFileContent struct {
	Type   string `json:"type"`   // 输入项的类型，始终为"input_file"
	FileID string `json:"fileId"` // 要发送到模型的文件的ID
}

func (i *InputFileContent) GetType() string { return i.Type }
func (i *InputFileContent) isInputContent() {}

type EasyInputMessage struct {
	Type    string                  `json:"type"`    // 消息输入的类型，始终设置为"message"
	Role    string                  `json:"role"`    // 消息输入的角色: "user", "assistant", "system", "developer"
	Content EasyInputMessageContent `json:"content"` // 内容可以是字符串或InputContent数组
}

func (e *EasyInputMessage) GetType() string { return e.Type }
func (e *EasyInputMessage) isInputItem()    {}
func (e *EasyInputMessage) isMessageItem()  {}

type EasyInputMessageContent interface {
	isEasyInputMessageContent()
}

type EasyInputTextContent struct {
	Text string
}

func (e *EasyInputTextContent) isEasyInputMessageContent() {}

type EasyInputContentList struct {
	Content []InputContent
}

func (e *EasyInputContentList) isEasyInputMessageContent() {}

func (e *EasyInputMessage) UnmarshalJSON(data []byte) error {
	type Alias EasyInputMessage
	aux := &struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 尝试解析为字符串
	var textContent string
	if err := json.Unmarshal(aux.Content, &textContent); err == nil {
		e.Content = &EasyInputTextContent{Text: textContent}
		return nil
	}

	// 尝试解析为InputContent数组
	var contentList []json.RawMessage
	if err := json.Unmarshal(aux.Content, &contentList); err == nil {
		contents := make([]InputContent, len(contentList))
		for i, raw := range contentList {
			typeName, err := ExtractType(raw)
			if err != nil {
				return err
			}

			var content InputContent
			switch typeName {
			case "input_text":
				content = &InputTextContent{}
			case "input_image":
				content = &InputImageContent{}
			case "input_file":
				content = &InputFileContent{}
			default:
				return fmt.Errorf("未知的输入内容类型: %s", typeName)
			}

			if err := json.Unmarshal(raw, content); err != nil {
				return err
			}

			contents[i] = content
		}

		e.Content = &EasyInputContentList{Content: contents}
		return nil
	}

	return fmt.Errorf("无法解析EasyInputMessage的content字段")
}

func (e *EasyInputMessage) MarshalJSON() ([]byte, error) {
	type Alias EasyInputMessage

	switch content := e.Content.(type) {
	case *EasyInputTextContent:
		return json.Marshal(&struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			Type:    e.Type,
			Role:    e.Role,
			Content: content.Text,
		})
	case *EasyInputContentList:
		return json.Marshal(&struct {
			Type    string         `json:"type"`
			Role    string         `json:"role"`
			Content []InputContent `json:"content"`
		}{
			Type:    e.Type,
			Role:    e.Role,
			Content: content.Content,
		})
	default:
		return nil, fmt.Errorf("未知的EasyInputMessage内容类型")
	}
}

type ItemReferenceParam struct {
	Type string `json:"type"` // 引用类型，始终为"item_reference"
	ID   string `json:"id"`   // 要引用的项目ID
}

func (i *ItemReferenceParam) GetType() string { return i.Type }
func (i *ItemReferenceParam) isInputItem()    {}
func (i *ItemReferenceParam) isMessageItem()  {}

type ResponseError struct {
	Code    ResponseErrorCode `json:"code"`    // 错误代码
	Message string            `json:"message"` // 错误的人类可读描述
}

type ResponseUsage struct {
	InputTokens         int64               `json:"inputTokens"`         // 输入令牌数
	InputTokensDetails  InputTokensDetails  `json:"inputTokensDetails"`  // 输入令牌的详细分类
	OutputTokens        int64               `json:"outputTokens"`        // 输出令牌数
	OutputTokensDetails OutputTokensDetails `json:"outputTokensDetails"` // 输出令牌的详细分类
	TotalTokens         int64               `json:"totalTokens"`         // 使用的总令牌数
}

type InputTokensDetails struct {
	CachedTokens int `json:"cachedTokens"` // 从缓存中检索的令牌数
}

type OutputTokensDetails struct {
	ReasoningTokens int `json:"reasoningTokens"` // 推理令牌数
}

type InputItem interface {
	GetType() string
	isInputItem()
	isMessageItem()
}

type OutputItem interface {
	GetID() string
	GetType() string
	GetStatus() string
	isOutputItem()
	isMessageItem()
}

type MessageItem interface {
	GetType() string
	isMessageItem()
}

type OutputMessage struct {
	ID      string          `json:"id"`      // 输出消息的唯一ID
	Type    string          `json:"type"`    // 输出消息的类型，始终为"message"
	Role    string          `json:"role"`    // 输出消息的角色，始终为"assistant"
	Content []OutputContent `json:"content"` // 输出消息的内容
	Status  string          `json:"status"`  // 消息输入的状态
}

func (o *OutputMessage) GetID() string     { return o.ID }
func (o *OutputMessage) GetType() string   { return o.Type }
func (o *OutputMessage) GetStatus() string { return o.Status }
func (o *OutputMessage) isOutputItem()     {}
func (o *OutputMessage) isMessageItem()    {}

type FunctionToolCall struct {
	ID        string `json:"id"`                  // 函数工具调用的唯一ID
	Type      string `json:"type"`                // 输出项的类型，始终为"tool_call"
	Name      string `json:"name"`                // 被调用的函数名称
	Status    string `json:"status"`              // 函数工具调用的状态
	Arguments string `json:"arguments,omitempty"` // 函数的参数，作为JSON字符串
}

func (f *FunctionToolCall) GetID() string     { return f.ID }
func (f *FunctionToolCall) GetType() string   { return f.Type }
func (f *FunctionToolCall) GetStatus() string { return f.Status }
func (f *FunctionToolCall) isOutputItem()     {}
func (f *FunctionToolCall) isInputItem()      {}
func (f *FunctionToolCall) isMessageItem()    {}

type ReasoningItem struct {
	Type    string        `json:"type"`             // 对象的类型，始终为"reasoning"
	ID      string        `json:"id"`               // 推理内容的唯一标识符
	Summary []SummaryText `json:"summary"`          // 推理文本内容
	Status  string        `json:"status,omitempty"` // 项目的状态
}

func (r *ReasoningItem) GetID() string     { return r.ID }
func (r *ReasoningItem) GetType() string   { return r.Type }
func (r *ReasoningItem) GetStatus() string { return r.Status }
func (r *ReasoningItem) isOutputItem()     {}
func (r *ReasoningItem) isMessageItem()    {}

type SummaryText struct {
	Type string `json:"type"` // 对象的类型，始终为"summary_text"
	Text string `json:"text"` // 模型生成响应时使用的推理的简短摘要
}

type OutputContent interface {
	GetType() string
	isOutputContent()
}

type OutputTextContent struct {
	Type        string       `json:"type"`                  // 输出文本的类型，始终为"output_text"
	Text        string       `json:"text"`                  // 模型的文本输出
	Annotations []Annotation `json:"annotations,omitempty"` // 文本输出的注释
}

func (o *OutputTextContent) GetType() string  { return o.Type }
func (o *OutputTextContent) isOutputContent() {}

type RefusalContent struct {
	Type    string `json:"type"`    // 拒绝的类型，始终为"refusal"
	Refusal string `json:"refusal"` // 模型的拒绝解释
}

func (r *RefusalContent) GetType() string  { return r.Type }
func (r *RefusalContent) isOutputContent() {}

type Annotation interface {
	GetType() string
	isAnnotation()
}

type FileCitationBody struct {
	Type   string `json:"type"`   // 文件引用的类型，始终为"file_citation"
	FileID string `json:"fileId"` // 文件的ID
	Index  int    `json:"index"`  // 文件在文件列表中的索引
}

func (f *FileCitationBody) GetType() string { return f.Type }
func (f *FileCitationBody) isAnnotation()   {}

type UrlCitationBody struct {
	Type       string `json:"type"`       // URL引用的类型，始终为"url_citation"
	URL        string `json:"url"`        // Web资源的URL
	StartIndex int    `json:"startIndex"` // URL引用在消息中的第一个字符的索引
	EndIndex   int    `json:"endIndex"`   // URL引用在消息中的最后一个字符的索引
	Title      string `json:"title"`      // Web资源的标题
}

func (u *UrlCitationBody) GetType() string { return u.Type }
func (u *UrlCitationBody) isAnnotation()   {}

type FilePathBody struct {
	Type   string `json:"type"`   // 文件路径的类型，始终为"file_path"
	FileID string `json:"fileId"` // 文件的ID
	Index  int    `json:"index"`  // 文件在文件列表中的索引
}

func (f *FilePathBody) GetType() string { return f.Type }
func (f *FilePathBody) isAnnotation()   {}

type FileSearchToolCall struct {
	ID      string             `json:"id"`                // 文件搜索工具调用的唯一ID
	Type    string             `json:"type"`              // 工具调用类型，始终为"file_search_call"
	Status  ToolCallStatus     `json:"status"`            // 工具调用状态
	Queries []string           `json:"queries"`           // 用于搜索文件的查询
	Results []FileSearchResult `json:"results,omitempty"` // 文件搜索结果
}

func (f *FileSearchToolCall) GetID() string     { return f.ID }
func (f *FileSearchToolCall) GetType() string   { return f.Type }
func (f *FileSearchToolCall) GetStatus() string { return string(f.Status) }
func (f *FileSearchToolCall) isOutputItem()     {}
func (f *FileSearchToolCall) isInputItem()      {}
func (f *FileSearchToolCall) isMessageItem()    {}

type FileSearchResult struct {
	FileID     string                    `json:"fileId"`               // 文件的唯一ID
	Text       string                    `json:"text"`                 // 从文件中检索的文本
	Filename   string                    `json:"filename"`             // 文件名
	Attributes VectorStoreFileAttributes `json:"attributes,omitempty"` // 文件属性
	Score      float64                   `json:"score"`                // 相关性分数，0-1之间的值
}

type VectorStoreFileAttributes map[string]VectorStoreFileAttributeValue

type VectorStoreFileAttributeValue interface {
	isVectorStoreFileAttributeValue()
}

type StringAttributeValue struct {
	Value string
}

func (s *StringAttributeValue) isVectorStoreFileAttributeValue() {}

type NumberAttributeValue struct {
	Value float64
}

func (n *NumberAttributeValue) isVectorStoreFileAttributeValue() {}

type BooleanAttributeValue struct {
	Value bool
}

func (b *BooleanAttributeValue) isVectorStoreFileAttributeValue() {}

func (v VectorStoreFileAttributes) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	for key, value := range v {
		switch val := value.(type) {
		case *StringAttributeValue:
			result[key] = val.Value
		case *NumberAttributeValue:
			result[key] = val.Value
		case *BooleanAttributeValue:
			result[key] = val.Value
		default:
			return nil, fmt.Errorf("未知的属性值类型: %T", value)
		}
	}
	return json.Marshal(result)
}

func (v *VectorStoreFileAttributes) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*v = make(VectorStoreFileAttributes)
	for key, value := range raw {
		switch val := value.(type) {
		case string:
			(*v)[key] = &StringAttributeValue{Value: val}
		case float64:
			(*v)[key] = &NumberAttributeValue{Value: val}
		case bool:
			(*v)[key] = &BooleanAttributeValue{Value: val}
		default:
			return fmt.Errorf("不支持的属性值类型: %T", value)
		}
	}
	return nil
}

type WebSearchToolCall struct {
	ID     string         `json:"id"`     // Web搜索工具调用的唯一ID
	Type   string         `json:"type"`   // 工具调用类型，始终为"web_search_call"
	Status ToolCallStatus `json:"status"` // 工具调用状态
}

func (w *WebSearchToolCall) GetID() string     { return w.ID }
func (w *WebSearchToolCall) GetType() string   { return w.Type }
func (w *WebSearchToolCall) GetStatus() string { return string(w.Status) }
func (w *WebSearchToolCall) isOutputItem()     {}
func (w *WebSearchToolCall) isInputItem()      {}
func (w *WebSearchToolCall) isMessageItem()    {}

type CodeInterpreterToolCall struct {
	ID      string                      `json:"id"`      // 代码解释器工具调用的唯一ID
	Type    string                      `json:"type"`    // 工具调用类型，始终为"code_interpreter_call"
	Code    string                      `json:"code"`    // 要运行的代码
	Status  ToolCallStatus              `json:"status"`  // 工具调用状态
	Results []CodeInterpreterToolOutput `json:"results"` // 代码解释器工具调用的结果
}

func (c *CodeInterpreterToolCall) GetID() string     { return c.ID }
func (c *CodeInterpreterToolCall) GetType() string   { return c.Type }
func (c *CodeInterpreterToolCall) GetStatus() string { return string(c.Status) }
func (c *CodeInterpreterToolCall) isOutputItem()     {}
func (c *CodeInterpreterToolCall) isInputItem()      {}
func (c *CodeInterpreterToolCall) isMessageItem()    {}

type CodeInterpreterToolOutput interface {
	GetType() string
	isCodeInterpreterOutput()
}

type CodeInterpreterTextOutput struct {
	Type string `json:"type"` // 输出类型，始终为"text"
	Text string `json:"text"` // 文本输出
}

func (c *CodeInterpreterTextOutput) GetType() string          { return c.Type }
func (c *CodeInterpreterTextOutput) isCodeInterpreterOutput() {}

type CodeInterpreterFileOutput struct {
	Type  string                      `json:"type"`  // 输出类型，始终为"files"
	Files []CodeInterpreterFileDetail `json:"files"` // 生成的文件列表
}

type CodeInterpreterFileDetail struct {
	MimeType string `json:"mimeType"` // 文件的MIME类型
	FileID   string `json:"fileId"`   // 文件的ID
}

func (c *CodeInterpreterFileOutput) GetType() string          { return c.Type }
func (c *CodeInterpreterFileOutput) isCodeInterpreterOutput() {}

type ComputerToolCall struct {
	ID                  string                `json:"id"`                  // 计算机工具调用的唯一ID
	Type                string                `json:"type"`                // 工具调用类型，始终为"computer_call"
	CallID              string                `json:"callId"`              // 调用ID
	Action              ComputerAction        `json:"action"`              // 计算机操作
	PendingSafetyChecks []ComputerSafetyCheck `json:"pendingSafetyChecks"` // 待处理的安全检查
	Status              ToolCallStatus        `json:"status"`              // 工具调用状态
}

func (c *ComputerToolCall) GetID() string     { return c.ID }
func (c *ComputerToolCall) GetType() string   { return c.Type }
func (c *ComputerToolCall) GetStatus() string { return string(c.Status) }
func (c *ComputerToolCall) isOutputItem()     {}
func (c *ComputerToolCall) isInputItem()      {}
func (c *ComputerToolCall) isMessageItem()    {}

type ComputerAction interface {
	GetType() string
	isComputerAction()
}

type ComputerSafetyCheck struct {
	ID      string `json:"id"`                // 安全检查的ID
	Code    string `json:"code,omitempty"`    // 安全检查的类型
	Message string `json:"message,omitempty"` // 安全检查的详细信息
}

type ClickAction struct {
	Type   string `json:"type"`   // 操作类型，始终为"click"
	Button string `json:"button"` // 鼠标按钮: "left", "right", "wheel", "back", "forward"
	X      int    `json:"x"`      // 点击的x坐标
	Y      int    `json:"y"`      // 点击的y坐标
}

func (c *ClickAction) GetType() string   { return c.Type }
func (c *ClickAction) isComputerAction() {}

type TypeAction struct {
	Type string `json:"type"` // 操作类型，始终为"type"
	Text string `json:"text"` // 要输入的文本
}

func (t *TypeAction) GetType() string   { return t.Type }
func (t *TypeAction) isComputerAction() {}

type ScreenshotAction struct {
	Type string `json:"type"` // 操作类型，始终为"screenshot"
}

func (s *ScreenshotAction) GetType() string   { return s.Type }
func (s *ScreenshotAction) isComputerAction() {}

type DoubleClickAction struct {
	Type string `json:"type"` // 操作类型，始终为"double_click"
	X    int    `json:"x"`    // 双击的x坐标
	Y    int    `json:"y"`    // 双击的y坐标
}

func (d *DoubleClickAction) GetType() string   { return d.Type }
func (d *DoubleClickAction) isComputerAction() {}

type DragAction struct {
	Type string `json:"type"` // 操作类型，始终为"drag"
	X    int    `json:"x"`    // 拖拽起始x坐标
	Y    int    `json:"y"`    // 拖拽起始y坐标
	ToX  int    `json:"toX"`  // 拖拽结束x坐标
	ToY  int    `json:"toY"`  // 拖拽结束y坐标
}

func (d *DragAction) GetType() string   { return d.Type }
func (d *DragAction) isComputerAction() {}

type KeyPressAction struct {
	Type string   `json:"type"` // 操作类型，始终为"key_press"
	Keys []string `json:"keys"` // 要按下的键
}

func (k *KeyPressAction) GetType() string   { return k.Type }
func (k *KeyPressAction) isComputerAction() {}

type MoveAction struct {
	Type string `json:"type"` // 操作类型，始终为"move"
	X    int    `json:"x"`    // 移动到的x坐标
	Y    int    `json:"y"`    // 移动到的y坐标
}

func (m *MoveAction) GetType() string   { return m.Type }
func (m *MoveAction) isComputerAction() {}

type ScrollAction struct {
	Type    string `json:"type"`    // 操作类型，始终为"scroll"
	X       int    `json:"x"`       // 滚动位置的x坐标
	Y       int    `json:"y"`       // 滚动位置的y坐标
	ScrollX int    `json:"scrollX"` // 水平滚动距离
	ScrollY int    `json:"scrollY"` // 垂直滚动距离
}

func (s *ScrollAction) GetType() string   { return s.Type }
func (s *ScrollAction) isComputerAction() {}

type WaitAction struct {
	Type     string `json:"type"`     // 操作类型，始终为"wait"
	Duration int    `json:"duration"` // 等待时间（毫秒）
}

func (w *WaitAction) GetType() string   { return w.Type }
func (w *WaitAction) isComputerAction() {}

type FunctionToolCallOutput struct {
	ID     string `json:"id"`     // 函数工具调用输出的唯一ID
	Type   string `json:"type"`   // 输出类型，始终为"function_call_output"
	CallID string `json:"callId"` // 函数工具调用的ID
	Output string `json:"output"` // 函数调用的输出结果
	Status string `json:"status"` // 输出状态
}

func (f *FunctionToolCallOutput) GetID() string     { return f.ID }
func (f *FunctionToolCallOutput) GetType() string   { return f.Type }
func (f *FunctionToolCallOutput) GetStatus() string { return f.Status }
func (f *FunctionToolCallOutput) isOutputItem()     {}
func (f *FunctionToolCallOutput) isInputItem()      {}
func (f *FunctionToolCallOutput) isMessageItem()    {}

type ComputerToolCallOutput struct {
	ID     string                 `json:"id"`     // 计算机工具调用输出的唯一ID
	Type   string                 `json:"type"`   // 输出类型，始终为"computer_call_output"
	CallID string                 `json:"callId"` // 计算机工具调用的ID
	Output ComputerToolCallResult `json:"output"` // 计算机调用的输出结果
	Status string                 `json:"status"` // 输出状态
}

func (c *ComputerToolCallOutput) GetID() string     { return c.ID }
func (c *ComputerToolCallOutput) GetType() string   { return c.Type }
func (c *ComputerToolCallOutput) GetStatus() string { return c.Status }
func (c *ComputerToolCallOutput) isOutputItem()     {}
func (c *ComputerToolCallOutput) isInputItem()      {}
func (c *ComputerToolCallOutput) isMessageItem()    {}

type ComputerToolCallResult interface {
	GetType() string
	isComputerToolCallResult()
}

type ComputerScreenshotResult struct {
	Type     string `json:"type"`     // 结果类型，始终为"screenshot"
	ImageURL string `json:"imageUrl"` // 截图的图片URL
}

func (c *ComputerScreenshotResult) GetType() string           { return c.Type }
func (c *ComputerScreenshotResult) isComputerToolCallResult() {}

type ComputerActionResult struct {
	Type    string `json:"type"`    // 结果类型，始终为"action"
	Success bool   `json:"success"` // 操作是否成功
	Message string `json:"message"` // 操作结果消息
}

func (c *ComputerActionResult) GetType() string           { return c.Type }
func (c *ComputerActionResult) isComputerToolCallResult() {}

func (o *OutputMessage) UnmarshalJSON(data []byte) error {
	type Alias OutputMessage
	aux := &struct {
		Content []json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	o.Content = make([]OutputContent, len(aux.Content))
	for i, raw := range aux.Content {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var content OutputContent
		switch typeName {
		case "output_text":
			content = &OutputTextContent{}
		case "refusal":
			content = &RefusalContent{}
		default:
			return fmt.Errorf("unknown output content type: %s", typeName)
		}

		if err := json.Unmarshal(raw, content); err != nil {
			return err
		}

		o.Content[i] = content
	}

	return nil
}

func (o *OutputTextContent) UnmarshalJSON(data []byte) error {
	type Alias OutputTextContent
	aux := &struct {
		Annotations []json.RawMessage `json:"annotations,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Annotations == nil {
		return nil
	}

	o.Annotations = make([]Annotation, len(aux.Annotations))
	for i, raw := range aux.Annotations {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var annotation Annotation
		switch typeName {
		case "file_citation":
			annotation = &FileCitationBody{}
		case "url_citation":
			annotation = &UrlCitationBody{}
		case "file_path":
			annotation = &FilePathBody{}
		default:
			return fmt.Errorf("unknown annotation type: %s", typeName)
		}

		if err := json.Unmarshal(raw, annotation); err != nil {
			return err
		}

		o.Annotations[i] = annotation
	}

	return nil
}

func (c *ComputerToolCall) UnmarshalJSON(data []byte) error {
	type Alias ComputerToolCall
	aux := &struct {
		Action json.RawMessage `json:"action"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	typeName, err := ExtractType(aux.Action)
	if err != nil {
		return err
	}

	var action ComputerAction
	switch typeName {
	case "click":
		action = &ClickAction{}
	case "double_click":
		action = &DoubleClickAction{}
	case "drag":
		action = &DragAction{}
	case "key_press":
		action = &KeyPressAction{}
	case "move":
		action = &MoveAction{}
	case "screenshot":
		action = &ScreenshotAction{}
	case "scroll":
		action = &ScrollAction{}
	case "type":
		action = &TypeAction{}
	case "wait":
		action = &WaitAction{}
	default:
		return fmt.Errorf("unknown computer action type: %s", typeName)
	}

	if err := json.Unmarshal(aux.Action, action); err != nil {
		return err
	}

	c.Action = action
	return nil
}

func (c *ComputerToolCallOutput) UnmarshalJSON(data []byte) error {
	type Alias ComputerToolCallOutput
	aux := &struct {
		Output json.RawMessage `json:"output"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	typeName, err := ExtractType(aux.Output)
	if err != nil {
		return err
	}

	var output ComputerToolCallResult
	switch typeName {
	case "screenshot":
		output = &ComputerScreenshotResult{}
	case "action":
		output = &ComputerActionResult{}
	default:
		return fmt.Errorf("unknown computer tool call result type: %s", typeName)
	}

	if err := json.Unmarshal(aux.Output, output); err != nil {
		return err
	}

	c.Output = output
	return nil
}

func (c *CodeInterpreterToolCall) UnmarshalJSON(data []byte) error {
	type Alias CodeInterpreterToolCall
	aux := &struct {
		Results []json.RawMessage `json:"results"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Results = make([]CodeInterpreterToolOutput, len(aux.Results))
	for i, raw := range aux.Results {
		typeName, err := ExtractType(raw)
		if err != nil {
			return err
		}

		var result CodeInterpreterToolOutput
		switch typeName {
		case "text":
			result = &CodeInterpreterTextOutput{}
		case "files":
			result = &CodeInterpreterFileOutput{}
		default:
			return fmt.Errorf("unknown code interpreter output type: %s", typeName)
		}

		if err := json.Unmarshal(raw, result); err != nil {
			return err
		}

		c.Results[i] = result
	}

	return nil
}
