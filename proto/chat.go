package types

import (
	"encoding/json"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/zhongshangwu/avatarai-social/proto/chat"
)

type ServerEvent struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}

type ClientEvent struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
	responses.ResponseAudioDoneEvent
	openai.ChatCompletionMessageParamUnion
}

type SendMsgEvent struct {
	chat.SendMsgEvent
}
