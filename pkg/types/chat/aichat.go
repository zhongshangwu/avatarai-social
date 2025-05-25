package chat

import (
	"encoding/json"
	"fmt"
)

type AIChatMessage struct {
	ID                string                   `json:"id"`                          // 此AI聊天消息的唯一标识符
	MessageID         string                   `json:"messageId"`                   // 消息ID，引用 message.id
	Role              RoleType                 `json:"role"`                        // 消息角色: "user", "assistant", "system"
	Text              string                   `json:"text"`                        // 消息内容(纯文本)
	MessageItems      []OutputItem             `json:"messageItems"`                // 模型生成的内容项数组
	InterruptType     int                      `json:"interruptType"`               // 消息是否有被中断，默认为0
	Status            AiChatMessageStatus      `json:"status"`                      // 响应生成的状态: "completed", "failed", "in_progress", "incomplete"
	Error             *ResponseError           `json:"error,omitempty"`             // 错误信息
	UserID            string                   `json:"userId"`                      // 用户ID (可能是用户ID，也可以是AssistantID)
	CreatedAt         int64                    `json:"createdAt"`                   // 创建时间
	UpdatedAt         int64                    `json:"updatedAt"`                   // 更新时间
	IncompleteDetails *IncompleteDetails       `json:"incompleteDetails,omitempty"` // 响应不完整的详细信息
	Usage             *ResponseUsage           `json:"usage,omitempty"`             // 使用情况统计
	Tools             []map[string]interface{} `json:"tools,omitempty"`             // 模型可用的工具
	Metadata          map[string]interface{}   `json:"metadata,omitempty"`          // 响应的其他元数据
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
		var typeInfo struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(raw, &typeInfo); err != nil {
			return err
		}

		var content InputContent
		switch typeInfo.Type {
		case "input_text":
			content = &InputTextContent{}
		case "input_image":
			content = &InputImageContent{}
		case "input_file":
			content = &InputFileContent{}
		default:
			return fmt.Errorf("unknown input content type: %s", typeInfo.Type)
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

type ResponseError struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误的人类可读描述
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
}

type OutputItem interface {
	GetID() string
	GetType() string
	GetStatus() string
	isOutputItem()
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

	m.MessageItems = make([]OutputItem, len(aux.MessageItems))
	for i, raw := range aux.MessageItems {
		var typeInfo struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(raw, &typeInfo); err != nil {
			return err
		}

		var item OutputItem
		switch typeInfo.Type {
		case "message":
			item = &OutputMessage{}
		case "tool_call":
			item = &FunctionToolCall{}
		case "reasoning":
			item = &ReasoningItem{}
		default:
			return fmt.Errorf("unknown output item type: %s", typeInfo.Type)
		}

		if err := json.Unmarshal(raw, item); err != nil {
			return err
		}

		m.MessageItems[i] = item
	}

	return nil
}

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
		var typeInfo struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(raw, &typeInfo); err != nil {
			return err
		}

		var content OutputContent
		switch typeInfo.Type {
		case "output_text":
			content = &OutputTextContent{}
		case "refusal":
			content = &RefusalContent{}
		default:
			return fmt.Errorf("unknown output content type: %s", typeInfo.Type)
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
		var typeInfo struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(raw, &typeInfo); err != nil {
			return err
		}

		var annotation Annotation
		switch typeInfo.Type {
		case "file_citation":
			annotation = &FileCitationBody{}
		case "url_citation":
			annotation = &UrlCitationBody{}
		default:
			return fmt.Errorf("unknown annotation type: %s", typeInfo.Type)
		}

		if err := json.Unmarshal(raw, annotation); err != nil {
			return err
		}

		o.Annotations[i] = annotation
	}

	return nil
}
