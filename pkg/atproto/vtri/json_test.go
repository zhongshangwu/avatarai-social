package vtri

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAIChatMessage_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		message  *ChatAiChat_Message
		wantJSON string
	}{
		{
			name: "基础AI聊天消息",
			message: &ChatAiChat_Message{
				LexiconTypeID: "app.vtri.chat.aiChat",
				Id:            "msg-123",
				Role:          "assistant",
				Status:        "completed",
				Text:          "你好，我是AI助手",
				MessageItems: []*ChatAiChat_OutputItem{
					{
						ChatAiChat_OutputMessage: &ChatAiChat_OutputMessage{
							LexiconTypeID: "app.vtri.chat.aiChat#OutputMessage",
							Content: []*ChatAiChat_OutputContent{
								{
									ChatAiChat_OutputTextContent: &ChatAiChat_OutputTextContent{
										Text: "你好，我是AI助手",
										Type: "output_text",
									},
								},
							},
							Id:     "item-1",
							Role:   "assistant",
							Status: "completed",
							Type:   "message",
						},
					},
				},
			},
			wantJSON: `{
				"$type": "app.vtri.chat.aiChat",
				"id": "msg-123",
				"role": "assistant",
				"status": "completed",
				"text": "你好，我是AI助手",
				"content": [
					{
						"$type": "app.vtri.chat.aiChat#OutputMessage",
						"content": [
							{
								"$type": "app.vtri.chat.aiChat#OutputTextContent",
								"text": "你好，我是AI助手",
								"type": "output_text"
							}
						],
						"id": "item-1",
						"role": "assistant",
						"status": "completed",
						"type": "message"
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			got, err := json.MarshalIndent(tt.message, "", "  ")
			assert.NoError(t, err)
			fmt.Println("--------------------------------")
			fmt.Println(string(got))
			fmt.Println("--------------------------------")

			var expectedJSON interface{}
			err = json.Unmarshal([]byte(tt.wantJSON), &expectedJSON)
			assert.NoError(t, err)

			var gotJSON interface{}
			err = json.Unmarshal(got, &gotJSON)
			assert.NoError(t, err)

			assert.Equal(t, expectedJSON, gotJSON)

			// 测试反序列化
			var decoded ChatAiChat_Message
			err = json.Unmarshal([]byte(tt.wantJSON), &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.message.Id, decoded.Id)
			assert.Equal(t, tt.message.Role, decoded.Role)
			assert.Equal(t, tt.message.Status, decoded.Status)
			assert.Equal(t, tt.message.Text, decoded.Text)
		})
	}
}

func TestChatEvent_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		event    *ChatEvent
		wantJSON string
	}{
		{
			name: "发送消息事件",
			event: &ChatEvent{
				EventId:   "evt-123",
				EventType: "send_msg",
				Event: &ChatEvent_Event{
					ChatEvent_SendMsgEvent: &ChatEvent_SendMsgEvent{
						LexiconTypeID: "app.vtri.chat.event#sendMsgEvent",
						RoomId:        "room-123",
						MsgType:       1,
						Body: &ChatEvent_SendMsgEvent_Body{
							ChatEvent_AiChatMsg: &ChatEvent_AiChatMsg{
								LexiconTypeID: "app.vtri.chat.event#aiChatMsg",
								Role:          "user",
								Content: []*ChatEvent_AiChatMsg_Content_Elem{
									{
										ChatAiChat_InputMessage: &ChatAiChat_InputMessage{
											LexiconTypeID: "app.vtri.chat.aiChat#InputMessage",
											Content: []*ChatAiChat_InputMessage_Content_Elem{
												{
													ChatAiChat_InputTextContent: &ChatAiChat_InputTextContent{
														Text: "你能帮我写一个测试用例吗？",
														Type: "input_text",
													},
												},
											},
										},
									},
								},
							},
						},
						SenderId: "user-123",
					},
				},
			},
			wantJSON: `{
				"event_id": "evt-123",
				"event_type": "send_msg",
				"event": {
					"$type": "app.vtri.chat.event#sendMsgEvent",
					"room_id": "room-123",
					"msg_type": 1,
					"body": {
						"$type": "app.vtri.chat.event#aiChatMsg",
						"role": "user",
						"content": [
							{
								"$type": "app.vtri.chat.aiChat#InputMessage",
								"content": [
									{
										"$type": "app.vtri.chat.aiChat#InputTextContent",
										"text": "你能帮我写一个测试用例吗？",
										"type": "input_text"
									}
								]
							}
						]
					},
					"sender_id": "user-123"
				}
			}`,
		},
		{
			name: "AI聊天流创建事件",
			event: &ChatEvent{
				EventId:   "evt-456",
				EventType: "created",
				Event: &ChatEvent_Event{
					ChatAiChatStream_CreatedEvent: &ChatAiChatStream_CreatedEvent{
						LexiconTypeID: "app.vtri.chat.aiChatStream#CreatedEvent",
						Response: &ChatAiChat_Message{
							LexiconTypeID: "app.vtri.chat.aiChat",
							Id:            "msg-456",
							Role:          "assistant",
							Status:        "created",
						},
					},
				},
			},
			wantJSON: `{
				"event_id": "evt-456",
				"event_type": "created",
				"event": {
					"$type": "app.vtri.chat.aiChatStream#CreatedEvent",
					"response": {
						"$type": "app.vtri.chat.aiChat",
						"id": "msg-456",
						"role": "assistant",
						"status": "created"
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试序列化
			got, err := json.MarshalIndent(tt.event, "", "  ")
			assert.NoError(t, err)
			fmt.Println("--------------------------------")
			fmt.Println(string(got))
			fmt.Println("--------------------------------")

			var expectedJSON interface{}
			err = json.Unmarshal([]byte(tt.wantJSON), &expectedJSON)
			assert.NoError(t, err)

			var gotJSON interface{}
			err = json.Unmarshal(got, &gotJSON)
			assert.NoError(t, err)

			assert.Equal(t, expectedJSON, gotJSON)

			// 测试反序列化
			var decoded ChatEvent
			err = json.Unmarshal([]byte(tt.wantJSON), &decoded)
			assert.NoError(t, err)
			assert.Equal(t, tt.event.EventId, decoded.EventId)
			assert.Equal(t, tt.event.EventType, decoded.EventType)
		})
	}
}
