package chat

type ChatEventType string

const (
	EventTypeSendMsg                         ChatEventType = "send_msg"
	EventTypeError                           ChatEventType = "error"
	EventTypeAIChatInterrupt                 ChatEventType = "ai_chat.interrupt"
	EventTypeAIChatCompleted                 ChatEventType = "ai_chat.completed"
	EventTypeAIChatContentPartAdded          ChatEventType = "ai_chat.content_part.added"
	EventTypeAIChatContentPartDone           ChatEventType = "ai_chat.content_part.done"
	EventTypeAIChatCreated                   ChatEventType = "ai_chat.created"
	EventTypeAIChatInProgress                ChatEventType = "ai_chat.in_progress"
	EventTypeAIChatFailed                    ChatEventType = "ai_chat.failed"
	EventTypeAIChatIncomplete                ChatEventType = "ai_chat.incomplete"
	EventTypeAIChatOutputItemAdded           ChatEventType = "ai_chat.output_item.added"
	EventTypeAIChatOutputItemDone            ChatEventType = "ai_chat.output_item.done"
	EventTypeAIChatReasoningSummaryPartAdded ChatEventType = "ai_chat.reasoning_summary.part.added"
	EventTypeAIChatReasoningSummaryPartDone  ChatEventType = "ai_chat.reasoning_summary.part.done"
	EventTypeAIChatReasoningSummaryTextDelta ChatEventType = "ai_chat.reasoning_summary.text.delta"
	EventTypeAIChatReasoningSummaryTextDone  ChatEventType = "ai_chat.reasoning_summary.text.done"
	EventTypeAIChatRefusalDelta              ChatEventType = "ai_chat.refusal.delta"
	EventTypeAIChatRefusalDone               ChatEventType = "ai_chat.refusal.done"
	EventTypeAIChatOutputTextAnnotationAdded ChatEventType = "ai_chat.output_text.annotation.added"
	EventTypeAIChatOutputTextDelta           ChatEventType = "ai_chat.output_text.delta"
	EventTypeAIChatOutputTextDone            ChatEventType = "ai_chat.output_text.done"
)

type RoleType string

const (
	RoleTypeUser      RoleType = "user"
	RoleTypeAssistant RoleType = "assistant"
	RoleTypeSystem    RoleType = "system"
)

type AiChatMessageStatus string

const (
	AiChatMessageStatusCompleted  AiChatMessageStatus = "completed"
	AiChatMessageStatusFailed     AiChatMessageStatus = "failed"
	AiChatMessageStatusInProgress AiChatMessageStatus = "in_progress"
	AiChatMessageStatusIncomplete AiChatMessageStatus = "incomplete"
)

type InterruptType int

const (
	InterruptTypeDefault InterruptType = iota
	InterruptTypeUser
	InterruptTypeSystem
)

type IncompleteReason string

const (
	IncompleteReasonMaxOutputTokens IncompleteReason = "max_output_tokens"
	IncompleteReasonContentFilter   IncompleteReason = "content_filter"
)
