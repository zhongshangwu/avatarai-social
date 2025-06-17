package main

import (
	"reflect"

	"github.com/bluesky-social/indigo/mst"

	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/vtri"
)

func main() {
	var typVals []any
	for _, typ := range mst.CBORTypes() {
		typVals = append(typVals, reflect.New(typ).Elem().Interface())
	}

	genCfg := cbg.Gen{
		MaxStringLength: 1_000_000,
	}

	if err := genCfg.WriteMapEncodersToFile("pkg/atproto/vtri/cbor_gen.go", "vtri",
		vtri.AvatarProfile{},
		vtri.AsterProfile{},
		vtri.EntityFile{},
		vtri.EntityExternal{},
		vtri.EntityExternal_External{},
		vtri.EntityImages{},
		vtri.EntityImages_Image{},
		vtri.EntityVideo{},
		vtri.EntityVideo_Caption{},
		vtri.ActivityMoment_ReplyRef{},
		vtri.EntityRecord{},
		vtri.EntityDefs_AspectRatio{},
		vtri.ActivityMoment_Embed{},
		vtri.ActivityMoment{},
		vtri.ActivityLike{},
		vtri.ActivityRelationship{},
		vtri.ActivityTopic{},
		vtri.ActivityTag{},

		vtri.ChatEvent{},
		vtri.ChatEvent_TextMsg{},
		vtri.ChatEvent_AiChatMsg{},
		vtri.ChatEvent_Event{},
		vtri.ChatEvent_SendMsgEvent{},
		vtri.ChatEvent_SendMsgEvent_Body{},
		vtri.ChatEvent_AiChatMsg_Content_Elem{},

		vtri.ChatMessage{},

		vtri.ChatAiChat_FunctionToolCall{},
		vtri.ChatAiChat_InputMessage{},
		vtri.ChatAiChat_InputMessage_Content_Elem{},
		vtri.ChatAiChat_InputTextContent{},
		vtri.ChatAiChat_InputImageContent{},
		vtri.ChatAiChat_InputFileContent{},
		vtri.ChatAiChat_ResponseError{},
		vtri.ChatAiChat_ResponseUsage{},
		vtri.ChatAiChat_ResponseUsage_InputTokensDetails{},
		vtri.ChatAiChat_ResponseUsage_OutputTokensDetails{},
		vtri.ChatAiChat_OutputItem{},
		vtri.ChatAiChat_OutputMessage{},
		vtri.ChatAiChat_ReasoningItem{},
		vtri.ChatAiChat_ReasoningItem_Summary_Elem{},
		vtri.ChatAiChat_OutputContent{},
		vtri.ChatAiChat_OutputTextContent{},
		vtri.ChatAiChat_RefusalContent{},
		vtri.ChatAiChat_Annotation{},
		vtri.ChatAiChat_FileCitationBody{},
		vtri.ChatAiChat_UrlCitationBody{},
		vtri.ChatAiChat_Message_IncompleteDetails{},
		vtri.ChatAiChat_Message_Metadata{},
		vtri.ChatAiChat_Message_Tools_Elem{},
		vtri.ChatAiChat_Message{},

		vtri.ChatAiChatStream_InterruptEvent{},
		vtri.ChatAiChatStream_CompletedEvent{},
		vtri.ChatAiChatStream_ContentPartAddedEvent{},
		vtri.ChatAiChatStream_ContentPartDoneEvent{},
		vtri.ChatAiChatStream_CreatedEvent{},
		vtri.ChatAiChatStream_ErrorEvent{},
		vtri.ChatAiChatStream_InProgressEvent{},
		vtri.ChatAiChatStream_FailedEvent{},
		vtri.ChatAiChatStream_IncompleteEvent{},
		vtri.ChatAiChatStream_OutputItemAddedEvent{},
		vtri.ChatAiChatStream_OutputItemDoneEvent{},
		vtri.ChatAiChatStream_ReasoningSummaryPartAddedEvent{},
		vtri.ChatAiChatStream_ReasoningSummaryPartDoneEvent{},
		vtri.ChatAiChatStream_ReasoningSummaryTextDeltaEvent{},
		vtri.ChatAiChatStream_ReasoningSummaryTextDoneEvent{},
		vtri.ChatAiChatStream_RefusalDeltaEvent{},
		vtri.ChatAiChatStream_RefusalDoneEvent{},
		vtri.ChatAiChatStream_TextAnnotationDeltaEvent{},
		vtri.ChatAiChatStream_TextDeltaEvent{},
		vtri.ChatAiChatStream_TextDoneEvent{},

		vtri.ChatThread{},
		vtri.ChatRoom{},
	); err != nil {
		panic(err)
	}
}
