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
	MessageTypeCard        MessageType = 8
	MessageTypeAIChat      MessageType = 9 // 客户端 Agent 代理或者作为服务端 Agent 代理发送的消息
	MessageTypeSystem      MessageType = 10
	MessageTypeDelete      MessageType = 11
	MessageTypeRTC         MessageType = 12
)

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

	// 函数调用相关事件
	EventTypeAIChatFunctionCallArgumentsDelta ChatEventType = "ai_chat.function_call_arguments.delta"
	EventTypeAIChatFunctionCallArgumentsDone  ChatEventType = "ai_chat.function_call_arguments.done"

	// 文件搜索相关事件
	EventTypeAIChatFileSearchCallInProgress ChatEventType = "ai_chat.file_search_call.in_progress"
	EventTypeAIChatFileSearchCallSearching  ChatEventType = "ai_chat.file_search_call.searching"
	EventTypeAIChatFileSearchCallCompleted  ChatEventType = "ai_chat.file_search_call.completed"

	// Web搜索相关事件
	EventTypeAIChatWebSearchCallInProgress ChatEventType = "ai_chat.web_search_call.in_progress"
	EventTypeAIChatWebSearchCallSearching  ChatEventType = "ai_chat.web_search_call.searching"
	EventTypeAIChatWebSearchCallCompleted  ChatEventType = "ai_chat.web_search_call.completed"

	// 代码解释器相关事件
	EventTypeAIChatCodeInterpreterCallInProgress   ChatEventType = "ai_chat.code_interpreter_call.in_progress"
	EventTypeAIChatCodeInterpreterCallInterpreting ChatEventType = "ai_chat.code_interpreter_call.interpreting"
	EventTypeAIChatCodeInterpreterCallCompleted    ChatEventType = "ai_chat.code_interpreter_call.completed"
	EventTypeAIChatCodeInterpreterCallCodeDelta    ChatEventType = "ai_chat.code_interpreter_call.code.delta"
	EventTypeAIChatCodeInterpreterCallCodeDone     ChatEventType = "ai_chat.code_interpreter_call.code.done"

	// 计算机使用相关事件
	EventTypeAIChatComputerCallInProgress ChatEventType = "ai_chat.computer_call.in_progress"
	EventTypeAIChatComputerCallCompleted  ChatEventType = "ai_chat.computer_call.completed"

	// 音频相关事件
	EventTypeAIChatAudioDelta           ChatEventType = "ai_chat.audio.delta"
	EventTypeAIChatAudioDone            ChatEventType = "ai_chat.audio.done"
	EventTypeAIChatAudioTranscriptDelta ChatEventType = "ai_chat.audio_transcript.delta"
	EventTypeAIChatAudioTranscriptDone  ChatEventType = "ai_chat.audio_transcript.done"
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
)

type RichTextNodeTextStypeType string

const (
	RichTextNodeTextStypeTypeBold        RichTextNodeTextStypeType = "bold"        // 加粗
	RichTextNodeTextStypeTypeUnderline   RichTextNodeTextStypeType = "underline"   // 下划线
	RichTextNodeTextStypeTypeLineThrough RichTextNodeTextStypeType = "lineThrough" // 删除线
	RichTextNodeTextStypeTypeItalic      RichTextNodeTextStypeType = "italic"      // 斜体
)
