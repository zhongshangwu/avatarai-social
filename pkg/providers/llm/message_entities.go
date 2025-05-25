package llm

import (
	"encoding/json"
	"fmt"
)

type PromptMessageRole string

const (
	PromptMessageRoleSystem    PromptMessageRole = "system"
	PromptMessageRoleUser      PromptMessageRole = "user"
	PromptMessageRoleAssistant PromptMessageRole = "assistant"
	PromptMessageRoleTool      PromptMessageRole = "tool"
)

func (p *PromptMessageRole) ValueOf(value string) (PromptMessageRole, error) {
	switch value {
	case string(PromptMessageRoleSystem):
		return PromptMessageRoleSystem, nil
	case string(PromptMessageRoleUser):
		return PromptMessageRoleUser, nil
	case string(PromptMessageRoleAssistant):
		return PromptMessageRoleAssistant, nil
	case string(PromptMessageRoleTool):
		return PromptMessageRoleTool, nil
	default:
		return "", fmt.Errorf("invalid prompt message type value %s", value)
	}
}

type PromptMessageTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type PromptMessageFunction struct {
	Type     string            `json:"type"`
	Function PromptMessageTool `json:"function"`
}

type PromptMessageContentType string

const (
	PromptMessageContentTypeText     PromptMessageContentType = "text"
	PromptMessageContentTypeImage    PromptMessageContentType = "image"
	PromptMessageContentTypeAudio    PromptMessageContentType = "audio"
	PromptMessageContentTypeVideo    PromptMessageContentType = "video"
	PromptMessageContentTypeDocument PromptMessageContentType = "document"
)

type PromptMessageContent interface {
	GetType() PromptMessageContentType
	MarshalJSON() ([]byte, error)
}

type TextPromptMessageContent struct {
	Type PromptMessageContentType `json:"type"`
	Data string                   `json:"data"`
}

func (t *TextPromptMessageContent) GetType() PromptMessageContentType {
	return PromptMessageContentTypeText
}

func (t *TextPromptMessageContent) MarshalJSON() ([]byte, error) {
	type Alias TextPromptMessageContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	})
}

type MultiModalPromptMessageContent struct {
	Type       PromptMessageContentType `json:"type"`
	Format     string                   `json:"format"`
	Base64Data string                   `json:"base64_data"`
	URL        string                   `json:"url"`
	MimeType   string                   `json:"mime_type"`
}

func (m *MultiModalPromptMessageContent) GetData() string {
	if m.URL != "" {
		return m.URL
	}
	return fmt.Sprintf("data:%s;base64,%s", m.MimeType, m.Base64Data)
}

type ImageDetailLevel string

const (
	ImageDetailLevelLow  ImageDetailLevel = "low"
	ImageDetailLevelHigh ImageDetailLevel = "high"
)

type ImagePromptMessageContent struct {
	MultiModalPromptMessageContent
	Detail ImageDetailLevel `json:"detail"`
}

func (i *ImagePromptMessageContent) GetType() PromptMessageContentType {
	return PromptMessageContentTypeImage
}

func (i *ImagePromptMessageContent) MarshalJSON() ([]byte, error) {
	type Alias ImagePromptMessageContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(i),
	})
}

type AudioPromptMessageContent struct {
	MultiModalPromptMessageContent
}

func (a *AudioPromptMessageContent) GetType() PromptMessageContentType {
	return PromptMessageContentTypeAudio
}

func (a *AudioPromptMessageContent) MarshalJSON() ([]byte, error) {
	type Alias AudioPromptMessageContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(a),
	})
}

type VideoPromptMessageContent struct {
	MultiModalPromptMessageContent
}

func (v *VideoPromptMessageContent) GetType() PromptMessageContentType {
	return PromptMessageContentTypeVideo
}

func (v *VideoPromptMessageContent) MarshalJSON() ([]byte, error) {
	type Alias VideoPromptMessageContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(v),
	})
}

type DocumentPromptMessageContent struct {
	MultiModalPromptMessageContent
}

func (d *DocumentPromptMessageContent) GetType() PromptMessageContentType {
	return PromptMessageContentTypeDocument
}

func (d *DocumentPromptMessageContent) MarshalJSON() ([]byte, error) {
	type Alias DocumentPromptMessageContent
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(d),
	})
}

type PromptMessage struct {
	Role    PromptMessageRole `json:"role"`
	Content interface{}       `json:"content,omitempty"`
	Name    string            `json:"name,omitempty"`
}

func (p *PromptMessage) IsEmpty() bool {
	return p.Content == nil
}

func (p *PromptMessage) UnmarshalJSON(data []byte) error {
	type Alias PromptMessage
	aux := &struct {
		Content json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var contentStr string
	if err := json.Unmarshal(aux.Content, &contentStr); err == nil {
		p.Content = contentStr
		return nil
	}

	var contentArray []json.RawMessage
	if err := json.Unmarshal(aux.Content, &contentArray); err == nil {
		var contents []PromptMessageContent
		for _, item := range contentArray {
			var typeObj struct {
				Type PromptMessageContentType `json:"type"`
			}
			if err := json.Unmarshal(item, &typeObj); err != nil {
				return err
			}

			var content PromptMessageContent
			switch typeObj.Type {
			case PromptMessageContentTypeText:
				var textContent TextPromptMessageContent
				if err := json.Unmarshal(item, &textContent); err != nil {
					return err
				}
				content = &textContent
			case PromptMessageContentTypeImage:
				var imageContent ImagePromptMessageContent
				if err := json.Unmarshal(item, &imageContent); err != nil {
					return err
				}
				content = &imageContent
			case PromptMessageContentTypeAudio:
				var audioContent AudioPromptMessageContent
				if err := json.Unmarshal(item, &audioContent); err != nil {
					return err
				}
				content = &audioContent
			case PromptMessageContentTypeVideo:
				var videoContent VideoPromptMessageContent
				if err := json.Unmarshal(item, &videoContent); err != nil {
					return err
				}
				content = &videoContent
			case PromptMessageContentTypeDocument:
				var docContent DocumentPromptMessageContent
				if err := json.Unmarshal(item, &docContent); err != nil {
					return err
				}
				content = &docContent
			default:
				return fmt.Errorf("unknown content type: %s", typeObj.Type)
			}
			contents = append(contents, content)
		}
		p.Content = contents
	}

	return nil
}

type UserPromptMessage struct {
	PromptMessage
}

func NewUserPromptMessage(content interface{}, name string) *UserPromptMessage {
	return &UserPromptMessage{
		PromptMessage: PromptMessage{
			Role:    PromptMessageRoleUser,
			Content: content,
			Name:    name,
		},
	}
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type AssistantPromptMessage struct {
	PromptMessage
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

func NewAssistantPromptMessage(content interface{}, name string, toolCalls []ToolCall) *AssistantPromptMessage {
	return &AssistantPromptMessage{
		PromptMessage: PromptMessage{
			Role:    PromptMessageRoleAssistant,
			Content: content,
			Name:    name,
		},
		ToolCalls: toolCalls,
	}
}

func (a *AssistantPromptMessage) IsEmpty() bool {
	return a.PromptMessage.IsEmpty() && len(a.ToolCalls) == 0
}

type SystemPromptMessage struct {
	PromptMessage
}

func NewSystemPromptMessage(content interface{}, name string) *SystemPromptMessage {
	return &SystemPromptMessage{
		PromptMessage: PromptMessage{
			Role:    PromptMessageRoleSystem,
			Content: content,
			Name:    name,
		},
	}
}

type ToolPromptMessage struct {
	PromptMessage
	ToolCallID string `json:"tool_call_id"`
}

func NewToolPromptMessage(content interface{}, name string, toolCallID string) *ToolPromptMessage {
	return &ToolPromptMessage{
		PromptMessage: PromptMessage{
			Role:    PromptMessageRoleTool,
			Content: content,
			Name:    name,
		},
		ToolCallID: toolCallID,
	}
}

func (t *ToolPromptMessage) IsEmpty() bool {
	return t.PromptMessage.IsEmpty() && t.ToolCallID == ""
}
