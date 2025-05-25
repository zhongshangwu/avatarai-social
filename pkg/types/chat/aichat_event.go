package chat

type InterruptEvent struct {
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
