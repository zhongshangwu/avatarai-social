package messages

type InterruptEvent struct {
	ResponseID string `json:"responseId"` // 响应ID
}

func (i *InterruptEvent) isChatEventBody() {}

type CompletedEvent struct {
	Response *AIChatMessage `json:"response"`
}

func (e *CompletedEvent) isChatEventBody() {}

type ContentPartAddedEvent struct {
	ItemID       string        `json:"itemId"`       // 添加内容部分的输出项的ID
	OutputIndex  int           `json:"outputIndex"`  // 添加内容部分的输出项的索引
	ContentIndex int           `json:"contentIndex"` // 添加的内容部分的索引
	Part         OutputContent `json:"part"`         // 添加的内容部分
}

func (e *ContentPartAddedEvent) isChatEventBody() {}

type ContentPartDoneEvent struct {
	ItemID       string        `json:"itemId"`       // 添加内容部分的输出项的ID
	OutputIndex  int           `json:"outputIndex"`  // 添加内容部分的输出项的索引
	ContentIndex int           `json:"contentIndex"` // 完成的内容部分的索引
	Part         OutputContent `json:"part"`         // 完成的内容部分
}

func (e *ContentPartDoneEvent) isChatEventBody() {}

type CreatedEvent struct {
	Response *AIChatMessage `json:"response"` // 创建的响应
}

func (e *CreatedEvent) isChatEventBody() {}

type ErrorEvent struct {
	Code    *string `json:"code"`    // 错误代码
	Message string  `json:"message"` // 错误消息
	Param   *string `json:"param"`   // 错误参数
}

func (e *ErrorEvent) isChatEventBody() {}

type InProgressEvent struct {
	Response *AIChatMessage `json:"response"` // 正在进行的响应
}

func (e *InProgressEvent) isChatEventBody() {}

type FailedEvent struct {
	Response *AIChatMessage `json:"response"` // 失败的响应
}

func (e *FailedEvent) isChatEventBody() {}

type IncompleteEvent struct {
	Response *AIChatMessage `json:"response"` // 不完整的响应
}

func (e *IncompleteEvent) isChatEventBody() {}

type OutputItemAddedEvent struct {
	OutputIndex int        `json:"outputIndex"` // 添加的输出项的索引
	Item        OutputItem `json:"item"`        // 添加的输出项
}

func (e *OutputItemAddedEvent) isChatEventBody() {}

type OutputItemDoneEvent struct {
	OutputIndex int        `json:"outputIndex"` // 标记为完成的输出项的索引
	Item        OutputItem `json:"item"`        // 标记为完成的输出项
}

func (e *OutputItemDoneEvent) isChatEventBody() {}

type ReasoningSummaryPartAddedEvent struct {
	ItemID       string `json:"itemId"`       // 添加推理摘要部分的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加推理摘要部分的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加推理摘要部分的内容部分的索引
}

func (e *ReasoningSummaryPartAddedEvent) isChatEventBody() {}

type ReasoningSummaryPartDoneEvent struct {
	ItemID       string `json:"itemId"`       // 添加推理摘要部分的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加推理摘要部分的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加推理摘要部分的内容部分的索引
	Text         string `json:"text"`         // 推理摘要部分的文本
}

func (e *ReasoningSummaryPartDoneEvent) isChatEventBody() {}

type ReasoningSummaryTextDeltaEvent struct {
	ItemID       string `json:"itemId"`       // 添加推理摘要文本增量的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加推理摘要文本增量的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加推理摘要文本增量的内容部分的索引
	Delta        string `json:"delta"`        // 添加到推理摘要的增量文本
}

func (e *ReasoningSummaryTextDeltaEvent) isChatEventBody() {}

type ReasoningSummaryTextDoneEvent struct {
	ItemID       string `json:"itemId"`       // 添加推理摘要文本的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加推理摘要文本的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加推理摘要文本的内容部分的索引
	Text         string `json:"text"`         // 完整的推理摘要文本
}

func (e *ReasoningSummaryTextDoneEvent) isChatEventBody() {}

type RefusalDeltaEvent struct {
	ItemID       string `json:"itemId"`       // 添加拒绝增量的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加拒绝增量的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加拒绝增量的内容部分的索引
	Delta        string `json:"delta"`        // 添加到拒绝的增量文本
}

func (e *RefusalDeltaEvent) isChatEventBody() {}

type RefusalDoneEvent struct {
	ItemID       string `json:"itemId"`       // 添加拒绝文本的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加拒绝文本的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加拒绝文本的内容部分的索引
	Text         string `json:"text"`         // 完整的拒绝文本
}

func (e *RefusalDoneEvent) isChatEventBody() {}

type TextAnnotationDeltaEvent struct {
	ItemID          string     `json:"itemId"`          // 添加文本注释的输出项的ID
	OutputIndex     int        `json:"outputIndex"`     // 添加文本注释的输出项的索引
	ContentIndex    int        `json:"contentIndex"`    // 添加文本注释的内容部分的索引
	AnnotationIndex int        `json:"annotationIndex"` // 添加的注释的索引
	Annotation      Annotation `json:"annotation"`      // 添加的注释
}

func (e *TextAnnotationDeltaEvent) isChatEventBody() {}

type TextDeltaEvent struct {
	ItemID       string `json:"itemId"`       // 添加文本增量的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 添加文本增量的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 添加文本增量的内容部分的索引
	Delta        string `json:"delta"`        // 添加的文本增量
}

func (e *TextDeltaEvent) isChatEventBody() {}

type TextDoneEvent struct {
	ItemID       string `json:"itemId"`       // 文本内容最终确定的输出项的ID
	OutputIndex  int    `json:"outputIndex"`  // 文本内容最终确定的输出项的索引
	ContentIndex int    `json:"contentIndex"` // 文本内容最终确定的内容部分的索引
	Text         string `json:"text"`         // 最终确定的文本内容
}

func (e *TextDoneEvent) isChatEventBody() {}

type FunctionCallArgumentsDeltaEvent struct {
	ItemID      string `json:"itemId"`      // 添加函数调用参数增量的输出项的ID
	OutputIndex int    `json:"outputIndex"` // 添加函数调用参数增量的输出项的索引
	Delta       string `json:"delta"`       // 添加的函数调用参数增量
}

func (e *FunctionCallArgumentsDeltaEvent) isChatEventBody() {}

type FunctionCallArgumentsDoneEvent struct {
	ItemID      string `json:"itemId"`      // 函数调用参数完成的输出项的ID
	OutputIndex int    `json:"outputIndex"` // 函数调用参数完成的输出项的索引
	Arguments   string `json:"arguments"`   // 完成的函数调用参数
}

func (e *FunctionCallArgumentsDoneEvent) isChatEventBody() {}

type FileSearchCallInProgressEvent struct {
	OutputIndex int    `json:"outputIndex"` // 文件搜索调用关联的输出项的索引
	ItemID      string `json:"itemId"`      // 文件搜索调用关联的输出项的ID
}

func (e *FileSearchCallInProgressEvent) isChatEventBody() {}

type FileSearchCallSearchingEvent struct {
	OutputIndex int    `json:"outputIndex"` // 文件搜索调用搜索的输出项的索引
	ItemID      string `json:"itemId"`      // 文件搜索调用搜索的输出项的ID
}

func (e *FileSearchCallSearchingEvent) isChatEventBody() {}

type FileSearchCallCompletedEvent struct {
	OutputIndex int    `json:"outputIndex"` // 文件搜索调用完成的输出项的索引
	ItemID      string `json:"itemId"`      // 文件搜索调用完成的输出项的ID
}

func (e *FileSearchCallCompletedEvent) isChatEventBody() {}

type WebSearchCallInProgressEvent struct {
	OutputIndex int    `json:"outputIndex"` // Web搜索调用关联的输出项的索引
	ItemID      string `json:"itemId"`      // Web搜索调用关联的输出项的ID
}

func (e *WebSearchCallInProgressEvent) isChatEventBody() {}

type WebSearchCallSearchingEvent struct {
	OutputIndex int    `json:"outputIndex"` // Web搜索调用搜索的输出项的索引
	ItemID      string `json:"itemId"`      // Web搜索调用搜索的输出项的ID
}

func (e *WebSearchCallSearchingEvent) isChatEventBody() {}

type WebSearchCallCompletedEvent struct {
	OutputIndex int    `json:"outputIndex"` // Web搜索调用完成的输出项的索引
	ItemID      string `json:"itemId"`      // Web搜索调用完成的输出项的ID
}

func (e *WebSearchCallCompletedEvent) isChatEventBody() {}

type CodeInterpreterCallInProgressEvent struct {
	OutputIndex int    `json:"outputIndex"` // 代码解释器调用关联的输出项的索引
	ItemID      string `json:"itemId"`      // 代码解释器调用关联的输出项的ID
}

func (e *CodeInterpreterCallInProgressEvent) isChatEventBody() {}

type CodeInterpreterCallInterpretingEvent struct {
	OutputIndex int    `json:"outputIndex"` // 代码解释器调用解释的输出项的索引
	ItemID      string `json:"itemId"`      // 代码解释器调用解释的输出项的ID
}

func (e *CodeInterpreterCallInterpretingEvent) isChatEventBody() {}

type CodeInterpreterCallCompletedEvent struct {
	OutputIndex int    `json:"outputIndex"` // 代码解释器调用完成的输出项的索引
	ItemID      string `json:"itemId"`      // 代码解释器调用完成的输出项的ID
}

func (e *CodeInterpreterCallCompletedEvent) isChatEventBody() {}

type CodeInterpreterCallCodeDeltaEvent struct {
	OutputIndex int    `json:"outputIndex"` // 代码解释器调用代码增量的输出项的索引
	ItemID      string `json:"itemId"`      // 代码解释器调用代码增量的输出项的ID
	Delta       string `json:"delta"`       // 添加的代码增量
}

func (e *CodeInterpreterCallCodeDeltaEvent) isChatEventBody() {}

type CodeInterpreterCallCodeDoneEvent struct {
	OutputIndex int    `json:"outputIndex"` // 代码解释器调用代码完成的输出项的索引
	ItemID      string `json:"itemId"`      // 代码解释器调用代码完成的输出项的ID
	Code        string `json:"code"`        // 完成的代码
}

func (e *CodeInterpreterCallCodeDoneEvent) isChatEventBody() {}

type ComputerCallInProgressEvent struct {
	OutputIndex int    `json:"outputIndex"` // 计算机调用关联的输出项的索引
	ItemID      string `json:"itemId"`      // 计算机调用关联的输出项的ID
}

func (e *ComputerCallInProgressEvent) isChatEventBody() {}

type ComputerCallCompletedEvent struct {
	OutputIndex int    `json:"outputIndex"` // 计算机调用完成的输出项的索引
	ItemID      string `json:"itemId"`      // 计算机调用完成的输出项的ID
}

func (e *ComputerCallCompletedEvent) isChatEventBody() {}

type AudioDeltaEvent struct {
	Delta string `json:"delta"` // 音频增量数据
}

func (e *AudioDeltaEvent) isChatEventBody() {}

type AudioDoneEvent struct {
	ResponseID string `json:"responseId"` // 响应ID
}

func (e *AudioDoneEvent) isChatEventBody() {}

type AudioTranscriptDeltaEvent struct {
	Delta string `json:"delta"` // 音频转录增量
}

func (e *AudioTranscriptDeltaEvent) isChatEventBody() {}

type AudioTranscriptDoneEvent struct {
	Transcript string `json:"transcript"` // 完整的音频转录
}

func (e *AudioTranscriptDoneEvent) isChatEventBody() {}
