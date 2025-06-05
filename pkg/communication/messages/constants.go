package messages

type MessageType int

const (
	MessageTypeUnspecified MessageType = 0
	MessageTypeText        MessageType = 1
	MessageTypePost        MessageType = 2
	MessageTypeImage       MessageType = 3
	MessageTypeFile        MessageType = 4
	MessageTypeAudio       MessageType = 5
	MessageTypeVideo       MessageType = 6
	MessageTypeSticker     MessageType = 7
	MessageTypeCard        MessageType = 8  // 暂时不实现
	MessageTypeAgent       MessageType = 9  // 客户端 Agent 代理或者作为服务端 Agent 代理发送的消息
	MessageTypeSystem      MessageType = 10 // 暂时不实现
	MessageTypeDelete      MessageType = 11 // 暂时不实现
	MessageTypeRTC         MessageType = 12 // 暂时不实现
)

type ChatEventType string

const (
	EventTypeError                                 ChatEventType = "error"
	EventTypeMessageSend                           ChatEventType = "message.send"
	EventTypeMessageSent                           ChatEventType = "message.sent"
	EventTypeMessageReceived                       ChatEventType = "message.received"
	EventTypeAgentMessageCreated                   ChatEventType = "agent_message.created"
	EventTypeAgentMessageInProgress                ChatEventType = "agent_message.in_progress"
	EventTypeAgentMessageInterrupt                 ChatEventType = "agent_message.interrupt"
	EventTypeAgentMessageCompleted                 ChatEventType = "agent_message.completed"
	EventTypeAgentMessageContentPartAdded          ChatEventType = "agent_message.content_part.added"
	EventTypeAgentMessageContentPartDone           ChatEventType = "agent_message.content_part.done"
	EventTypeAgentMessageFailed                    ChatEventType = "agent_message.failed"
	EventTypeAgentMessageIncomplete                ChatEventType = "agent_message.incomplete"
	EventTypeAgentMessageOutputItemAdded           ChatEventType = "agent_message.output_item.added"
	EventTypeAgentMessageOutputItemDone            ChatEventType = "agent_message.output_item.done"
	EventTypeAgentMessageReasoningSummaryPartAdded ChatEventType = "agent_message.reasoning_summary.part.added"
	EventTypeAgentMessageReasoningSummaryPartDone  ChatEventType = "agent_message.reasoning_summary.part.done"
	EventTypeAgentMessageReasoningSummaryTextDelta ChatEventType = "agent_message.reasoning_summary.text.delta"
	EventTypeAgentMessageReasoningSummaryTextDone  ChatEventType = "agent_message.reasoning_summary.text.done"
	EventTypeAgentMessageRefusalDelta              ChatEventType = "agent_message.refusal.delta"
	EventTypeAgentMessageRefusalDone               ChatEventType = "agent_message.refusal.done"
	EventTypeAgentMessageOutputTextAnnotationAdded ChatEventType = "agent_message.output_text.annotation.added"
	EventTypeAgentMessageOutputTextDelta           ChatEventType = "agent_message.output_text.delta"
	EventTypeAgentMessageOutputTextDone            ChatEventType = "agent_message.output_text.done"

	// 函数调用相关事件
	EventTypeAgentMessageFunctionCallArgumentsDelta ChatEventType = "agent_message.function_call_arguments.delta"
	EventTypeAgentMessageFunctionCallArgumentsDone  ChatEventType = "agent_message.function_call_arguments.done"

	// 文件搜索相关事件
	EventTypeAgentMessageFileSearchCallInProgress ChatEventType = "agent_message.file_search_call.in_progress"
	EventTypeAgentMessageFileSearchCallSearching  ChatEventType = "agent_message.file_search_call.searching"
	EventTypeAgentMessageFileSearchCallCompleted  ChatEventType = "agent_message.file_search_call.completed"

	// Web搜索相关事件
	EventTypeAgentMessageWebSearchCallInProgress ChatEventType = "agent_message.web_search_call.in_progress"
	EventTypeAgentMessageWebSearchCallSearching  ChatEventType = "agent_message.web_search_call.searching"
	EventTypeAgentMessageWebSearchCallCompleted  ChatEventType = "agent_message.web_search_call.completed"

	// 代码解释器相关事件
	EventTypeAgentMessageCodeInterpreterCallInProgress   ChatEventType = "agent_message.code_interpreter_call.in_progress"
	EventTypeAgentMessageCodeInterpreterCallInterpreting ChatEventType = "agent_message.code_interpreter_call.interpreting"
	EventTypeAgentMessageCodeInterpreterCallCompleted    ChatEventType = "agent_message.code_interpreter_call.completed"
	EventTypeAgentMessageCodeInterpreterCallCodeDelta    ChatEventType = "agent_message.code_interpreter_call.code.delta"
	EventTypeAgentMessageCodeInterpreterCallCodeDone     ChatEventType = "agent_message.code_interpreter_call.code.done"

	// 计算机使用相关事件
	EventTypeAgentMessageComputerCallInProgress ChatEventType = "agent_message.computer_call.in_progress"
	EventTypeAgentMessageComputerCallCompleted  ChatEventType = "agent_message.computer_call.completed"

	// 音频相关事件
	EventTypeAgentMessageAudioDelta           ChatEventType = "agent_message.audio.delta"
	EventTypeAgentMessageAudioDone            ChatEventType = "agent_message.audio.done"
	EventTypeAgentMessageAudioTranscriptDelta ChatEventType = "agent_message.audio_transcript.delta"
	EventTypeAgentMessageAudioTranscriptDone  ChatEventType = "agent_message.audio_transcript.done"
)

type RoleType string

const (
	RoleTypeUser      RoleType = "user"
	RoleTypeAssistant RoleType = "assistant"
	RoleTypeSystem    RoleType = "system" // 没有这种实现, 后续删除
)

type AgentMessageStatus string

const (
	AgentMessageStatusCompleted  AgentMessageStatus = "completed"
	AgentMessageStatusFailed     AgentMessageStatus = "failed"
	AgentMessageStatusInProgress AgentMessageStatus = "in_progress"
	AgentMessageStatusIncomplete AgentMessageStatus = "incomplete"
)

type InterruptType int32

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

type ToolCallStatus string

const (
	ToolCallStatusInProgress   ToolCallStatus = "in_progress"
	ToolCallStatusSearching    ToolCallStatus = "searching"
	ToolCallStatusInterpreting ToolCallStatus = "interpreting"
	ToolCallStatusCompleted    ToolCallStatus = "completed"
	ToolCallStatusIncomplete   ToolCallStatus = "incomplete"
	ToolCallStatusFailed       ToolCallStatus = "failed"
)

type ToolType string

const (
	ToolTypeFileSearchCall      ToolType = "file_search_call"
	ToolTypeWebSearchCall       ToolType = "web_search_call"
	ToolTypeCodeInterpreterCall ToolType = "code_interpreter_call"
	ToolTypeComputerCall        ToolType = "computer_call"
	ToolTypeFunctionCall        ToolType = "tool_call"
)

type ContentType string

const (
	ContentTypeInputText   ContentType = "input_text"
	ContentTypeInputImage  ContentType = "input_image"
	ContentTypeInputFile   ContentType = "input_file"
	ContentTypeOutputText  ContentType = "output_text"
	ContentTypeRefusal     ContentType = "refusal"
	ContentTypeMessage     ContentType = "message"
	ContentTypeReasoning   ContentType = "reasoning"
	ContentTypeSummaryText ContentType = "summary_text"
)

type AnnotationType string

const (
	AnnotationTypeFileCitation AnnotationType = "file_citation"
	AnnotationTypeUrlCitation  AnnotationType = "url_citation"
	AnnotationTypeFilePath     AnnotationType = "file_path"
)

type ComputerActionType string

const (
	ComputerActionTypeClick       ComputerActionType = "click"
	ComputerActionTypeDoubleClick ComputerActionType = "double_click"
	ComputerActionTypeDrag        ComputerActionType = "drag"
	ComputerActionTypeKeyPress    ComputerActionType = "key_press"
	ComputerActionTypeMove        ComputerActionType = "move"
	ComputerActionTypeScreenshot  ComputerActionType = "screenshot"
	ComputerActionTypeScroll      ComputerActionType = "scroll"
	ComputerActionTypeType        ComputerActionType = "type"
	ComputerActionTypeWait        ComputerActionType = "wait"
)

type ToolCallOutputType string

const (
	ToolCallOutputTypeFunctionCallOutput ToolCallOutputType = "function_call_output"
	ToolCallOutputTypeComputerCallOutput ToolCallOutputType = "computer_call_output"
)

type CodeInterpreterOutputType string

const (
	CodeInterpreterOutputTypeText  CodeInterpreterOutputType = "text"
	CodeInterpreterOutputTypeFiles CodeInterpreterOutputType = "files"
)

type ComputerToolCallResultType string

const (
	ComputerToolCallResultTypeScreenshot ComputerToolCallResultType = "screenshot"
	ComputerToolCallResultTypeAction     ComputerToolCallResultType = "action"
)

type ResponseErrorCode string

const (
	ResponseErrorCodeServerError                 ResponseErrorCode = "server_error"
	ResponseErrorCodeRateLimitExceeded           ResponseErrorCode = "rate_limit_exceeded"
	ResponseErrorCodeInvalidPrompt               ResponseErrorCode = "invalid_prompt"
	ResponseErrorCodeVectorStoreTimeout          ResponseErrorCode = "vector_store_timeout"
	ResponseErrorCodeInvalidImage                ResponseErrorCode = "invalid_image"
	ResponseErrorCodeInvalidImageFormat          ResponseErrorCode = "invalid_image_format"
	ResponseErrorCodeInvalidBase64Image          ResponseErrorCode = "invalid_base64_image"
	ResponseErrorCodeInvalidImageURL             ResponseErrorCode = "invalid_image_url"
	ResponseErrorCodeImageTooLarge               ResponseErrorCode = "image_too_large"
	ResponseErrorCodeImageTooSmall               ResponseErrorCode = "image_too_small"
	ResponseErrorCodeImageParseError             ResponseErrorCode = "image_parse_error"
	ResponseErrorCodeImageContentPolicyViolation ResponseErrorCode = "image_content_policy_violation"
	ResponseErrorCodeInvalidImageMode            ResponseErrorCode = "invalid_image_mode"
	ResponseErrorCodeImageFileTooLarge           ResponseErrorCode = "image_file_too_large"
	ResponseErrorCodeUnsupportedImageMediaType   ResponseErrorCode = "unsupported_image_media_type"
	ResponseErrorCodeEmptyImageFile              ResponseErrorCode = "empty_image_file"
	ResponseErrorCodeFailedToDownloadImage       ResponseErrorCode = "failed_to_download_image"
	ResponseErrorCodeImageFileNotFound           ResponseErrorCode = "image_file_not_found"
	ResponseErrorCodeLLMRequestLength            ResponseErrorCode = "llm_request_length"
)

type RichTextNodeTextStypeType string

const (
	RichTextNodeTextStypeTypeBold        RichTextNodeTextStypeType = "bold"        // 加粗
	RichTextNodeTextStypeTypeUnderline   RichTextNodeTextStypeType = "underline"   // 下划线
	RichTextNodeTextStypeTypeLineThrough RichTextNodeTextStypeType = "lineThrough" // 删除线
	RichTextNodeTextStypeTypeItalic      RichTextNodeTextStypeType = "italic"      // 斜体
)
