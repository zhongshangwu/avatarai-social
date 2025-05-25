package chat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestClientEvent_ProtobufSerialization(t *testing.T) {
	tests := []struct {
		name  string
		event *ClientEvent
	}{
		{
			name: "SendMsg事件_文本消息",
			event: &ClientEvent{
				EventId:   "test-event-1",
				EventType: "send_msg",
				Event: &ClientEvent_SendMsg{
					SendMsg: &SendMsgEvent{
						RoomId:  "room-123",
						MsgType: MessageType_MESSAGE_TYPE_TEXT,
						Body: &SendMsgEvent_Payload{
							Payload: "Hello, World!",
						},
						SenderId: "user-123",
						ThreadId: "thread-123",
						QuoteId:  "quote-123",
						SenderAt: "2024-03-15T10:00:00Z",
					},
				},
			},
		},
		{
			name: "SendMsg事件_AI聊天消息",
			event: &ClientEvent{
				EventId:   "test-event-2",
				EventType: "send_msg",
				Event: &ClientEvent_SendMsg{
					SendMsg: &SendMsgEvent{
						RoomId:  "room-456",
						MsgType: MessageType_MESSAGE_TYPE_AI_CHAT,
						Body: &SendMsgEvent_AiChatBody{
							AiChatBody: &AIChatMessageBody{
								Id:      "msg-123",
								Role:    "user",
								Content: "What's the weather like?",
								Status:  1,
								MessageItems: []*AIChatMessageBody_AIChatMessageItem{
									{
										Message: &AIChatMessageItemMessage{
											Id:   "item-1",
											Type: "message",
											Role: "user",
											Content: []*AIChatMessageItemMessage_Content{
												{
													Content: "What's the weather like?",
												},
											},
										},
									},
								},
							},
						},
						SenderId: "user-456",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 Protobuf 序列化
			data, err := proto.Marshal(tt.event)
			assert.NoError(t, err, "Protobuf序列化失败")
			assert.NotEmpty(t, data, "序列化后的数据不应为空")

			// 测试 Protobuf 反序列化
			decoded := &ClientEvent{}
			err = proto.Unmarshal(data, decoded)
			assert.NoError(t, err, "Protobuf反序列化失败")
			assert.True(t, proto.Equal(tt.event, decoded), "反序列化后的数据与原数据不匹配")
		})
	}
}

func TestClientEvent_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		event    *ClientEvent
		wantJSON string
	}{
		{
			name: "文本消息JSON序列化",
			event: &ClientEvent{
				EventId:   "test-event-1",
				EventType: "send_msg",
				Event: &ClientEvent_SendMsg{
					SendMsg: &SendMsgEvent{
						RoomId:  "room-123",
						MsgType: MessageType_MESSAGE_TYPE_TEXT,
						Body: &SendMsgEvent_Payload{
							Payload: "Hello, World!",
						},
						SenderId: "user-123",
					},
				},
			},
		},
		{
			name: "AI聊天消息JSON序列化",
			event: &ClientEvent{
				EventId:   "test-event-2",
				EventType: "send_msg",
				Event: &ClientEvent_SendMsg{
					SendMsg: &SendMsgEvent{
						RoomId:  "room-456",
						MsgType: MessageType_MESSAGE_TYPE_AI_CHAT,
						Body: &SendMsgEvent_AiChatBody{
							AiChatBody: &AIChatMessageBody{
								Id:      "msg-123",
								Role:    "user",
								Content: "Tell me a story",
								MessageItems: []*AIChatMessageBody_AIChatMessageItem{
									{
										Message: &AIChatMessageItemMessage{
											Id:   "item-1",
											Type: "message",
											Role: "user",
											Content: []*AIChatMessageItemMessage_Content{
												{
													Content: "Tell me a story",
												},
											},
										},
									},
									{
										FunctionCall: &AIChatMessageItemFunctionCall{
											Id:   "item-2",
											Type: "function_call",
											Function: &AIChatMessageItemFunctionCall_Function{
												Name:      "tell_story",
												Arguments: "{\"story_length\": 5}",
											},
										},
									},
								},
								Status: 1,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 protojson 进行序列化
			marshaler := protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			}

			// 序列化为JSON
			jsonData, err := marshaler.Marshal(tt.event)
			assert.NoError(t, err, "JSON序列化失败")
			assert.NotEmpty(t, jsonData, "JSON数据不应为空")

			// 反序列化JSON
			unmarshaler := protojson.UnmarshalOptions{
				DiscardUnknown: true,
			}
			decoded := &ClientEvent{}
			err = unmarshaler.Unmarshal(jsonData, decoded)
			assert.NoError(t, err, "JSON反序列化失败")

			// 验证反序列化后的数据
			assert.Equal(t, tt.event.EventId, decoded.EventId)
			assert.Equal(t, tt.event.EventType, decoded.EventType)

			// 使用标准json包进行美化输出（用于调试）
			var prettyJSON map[string]interface{}
			err = json.Unmarshal(jsonData, &prettyJSON)
			assert.NoError(t, err, "JSON美化失败")

			prettyOutput, err := json.MarshalIndent(prettyJSON, "", "  ")
			assert.NoError(t, err, "JSON美化输出失败")
			t.Logf("美化后的JSON输出:\n%s", string(prettyOutput))
		})
	}
}

func TestClientEvent_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		event       *ClientEvent
		shouldError bool
	}{
		{
			name:        "空事件",
			event:       &ClientEvent{},
			shouldError: false,
		},
		{
			name: "只有ID的事件",
			event: &ClientEvent{
				EventId: "test-event",
			},
			shouldError: false,
		},
		{
			name: "无效的消息类型",
			event: &ClientEvent{
				EventId:   "test-event",
				EventType: "send_msg",
				Event: &ClientEvent_SendMsg{
					SendMsg: &SendMsgEvent{
						MsgType: MessageType_MESSAGE_TYPE_UNSPECIFIED,
					},
				},
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Protobuf序列化测试
			data, err := proto.Marshal(tt.event)
			if tt.shouldError {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")

				// 反序列化测试
				decoded := &ClientEvent{}
				err = proto.Unmarshal(data, decoded)
				assert.NoError(t, err, "反序列化不应该返回错误")
			}

			// JSON序列化测试
			marshaler := protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			}

			jsonData, err := marshaler.Marshal(tt.event)
			if tt.shouldError {
				assert.Error(t, err, "JSON序列化应该返回错误")
			} else {
				assert.NoError(t, err, "JSON序列化不应该返回错误")
				t.Logf("边界情况JSON输出:\n%s", string(jsonData))
			}
		})
	}
}

func TestClientEvent_MessageTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{
			name:     "文本消息类型",
			msgType:  MessageType_MESSAGE_TYPE_TEXT,
			expected: "MESSAGE_TYPE_TEXT",
		},
		{
			name:     "AI聊天消息类型",
			msgType:  MessageType_MESSAGE_TYPE_AI_CHAT,
			expected: "MESSAGE_TYPE_AI_CHAT",
		},
		{
			name:     "未指定消息类型",
			msgType:  MessageType_MESSAGE_TYPE_UNSPECIFIED,
			expected: "MESSAGE_TYPE_UNSPECIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &SendMsgEvent{
				MsgType: tt.msgType,
			}

			// 测试消息类型的字符串表示
			assert.Equal(t, tt.expected, tt.msgType.String(), "消息类型字符串表示不匹配")

			// 测试消息类型的序列化
			data, err := proto.Marshal(event)
			assert.NoError(t, err, "消息类型序列化失败")

			// 测试消息类型的反序列化
			decoded := &SendMsgEvent{}
			err = proto.Unmarshal(data, decoded)
			assert.NoError(t, err, "消息类型反序列化失败")
			assert.Equal(t, tt.msgType, decoded.MsgType, "消息类型不匹配")
		})
	}
}
