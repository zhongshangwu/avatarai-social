package messages

type RichTextNode interface {
	GetTag() RichTextNodeType
	isRichTextNode()
}

type RichTextNodeType string

const (
	PostNodeText      RichTextNodeType = "text"       // 文本
	PostNodeLink      RichTextNodeType = "a"          // 超链接
	PostNodeAt        RichTextNodeType = "at"         // @用户
	PostNodeImage     RichTextNodeType = "img"        // 图片
	PostNodeMedia     RichTextNodeType = "media"      // 视频
	PostNodeEmotion   RichTextNodeType = "emotion"    // 表情
	PostNodeHr        RichTextNodeType = "hr"         // 分割线
	PostNodeCodeBlock RichTextNodeType = "code_block" // 代码块
	PostNodeMarkdown  RichTextNodeType = "md"         // Markdown
)

type RichTextNodeText struct {
	Tag      string                      `json:"tag"`                // 标签类型，固定为"text"
	Text     string                      `json:"text"`               // 文本内容
	UnEscape bool                        `json:"unEscape,omitempty"` // 是否unescape解码
	Style    []RichTextNodeTextStypeType `json:"style,omitempty"`    // 文本样式
}

func (r *RichTextNodeText) GetTag() RichTextNodeType { return PostNodeText }
func (r *RichTextNodeText) isRichTextNode()          {}

type RichTextNodeLink struct {
	Tag   string                      `json:"tag"`             // 标签类型，固定为"a"
	Text  string                      `json:"text"`            // 超链接文本
	Href  string                      `json:"href"`            // 超链接地址
	Style []RichTextNodeTextStypeType `json:"style,omitempty"` // 文本样式
}

func (r *RichTextNodeLink) GetTag() RichTextNodeType { return PostNodeLink }
func (r *RichTextNodeLink) isRichTextNode()          {}

type RichTextNodeAt struct {
	Tag    string                      `json:"tag"`             // 标签类型，固定为"at"
	UserID string                      `json:"userId"`          // 用户ID
	Style  []RichTextNodeTextStypeType `json:"style,omitempty"` // 文本样式
}

func (r *RichTextNodeAt) GetTag() RichTextNodeType { return PostNodeAt }
func (r *RichTextNodeAt) isRichTextNode()          {}

type RichTextNodeImage struct {
	Tag      string `json:"tag"`      // 标签类型，固定为"img"
	ImageKey string `json:"imageKey"` // 图片Key
}

func (r *RichTextNodeImage) GetTag() RichTextNodeType { return PostNodeImage }
func (r *RichTextNodeImage) isRichTextNode()          {}

type RichTextNodeVideo struct {
	Tag      string `json:"tag"`                // 标签类型，固定为"media"
	FileKey  string `json:"fileKey"`            // 视频文件Key
	ImageKey string `json:"imageKey,omitempty"` // 视频封面图片Key
}

func (r *RichTextNodeVideo) GetTag() RichTextNodeType { return PostNodeMedia }
func (r *RichTextNodeVideo) isRichTextNode()          {}

type RichTextNodeEmotion struct {
	Tag       string `json:"tag"`       // 标签类型，固定为"emotion"
	EmojiType string `json:"emojiType"` // 表情类型
}

func (r *RichTextNodeEmotion) GetTag() RichTextNodeType { return PostNodeEmotion }
func (r *RichTextNodeEmotion) isRichTextNode()          {}

type RichTextNodeHr struct {
	Tag string `json:"tag"` // 标签类型，固定为"hr"
}

func (r *RichTextNodeHr) GetTag() RichTextNodeType { return PostNodeHr }
func (r *RichTextNodeHr) isRichTextNode()          {}

type RichTextNodeCodeBlock struct {
	Tag      string `json:"tag"`                // 标签类型，固定为"code_block"
	Language string `json:"language,omitempty"` // 代码语言
	Text     string `json:"text"`               // 代码内容
}

func (r *RichTextNodeCodeBlock) GetTag() RichTextNodeType { return PostNodeCodeBlock }
func (r *RichTextNodeCodeBlock) isRichTextNode()          {}

type RichTextNodeMarkdown struct {
	Tag  string `json:"tag"`  // 标签类型，固定为"md"
	Text string `json:"text"` // Markdown内容
}

func (r *RichTextNodeMarkdown) GetTag() RichTextNodeType { return PostNodeMarkdown }
func (r *RichTextNodeMarkdown) isRichTextNode()          {}
