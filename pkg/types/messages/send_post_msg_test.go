package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostMsgBody_UnmarshalJSON(t *testing.T) {
	// 测试数据
	jsonData := `{
		"title": "测试富文本消息",
		"content": [
			[
				{
					"tag": "text",
					"text": "这是一段文本",
					"style": ["bold"]
				},
				{
					"tag": "a",
					"text": "链接文本",
					"href": "https://example.com"
				}
			],
			[
				{
					"tag": "at",
					"user_id": "user123"
				},
				{
					"tag": "img",
					"image_key": "img_key_123"
				}
			],
			[
				{
					"tag": "code_block",
					"language": "go",
					"text": "fmt.Println(\"Hello World\")"
				}
			]
		]
	}`

	var postMsg PostMsgBody
	err := json.Unmarshal([]byte(jsonData), &postMsg)
	require.NoError(t, err)

	// 验证标题
	assert.Equal(t, "测试富文本消息", postMsg.Title)

	// 验证内容结构
	assert.Len(t, postMsg.Content, 3)
	assert.Len(t, postMsg.Content[0], 2)
	assert.Len(t, postMsg.Content[1], 2)
	assert.Len(t, postMsg.Content[2], 1)

	// 验证第一行第一个节点（文本节点）
	textNode, ok := postMsg.Content[0][0].(*RichTextNodeText)
	require.True(t, ok)
	assert.Equal(t, PostNodeText, textNode.GetTag())
	assert.Equal(t, "这是一段文本", textNode.Text)
	assert.Equal(t, []string{"bold"}, textNode.Style)

	// 验证第一行第二个节点（链接节点）
	linkNode, ok := postMsg.Content[0][1].(*RichTextNodeLink)
	require.True(t, ok)
	assert.Equal(t, PostNodeLink, linkNode.GetTag())
	assert.Equal(t, "链接文本", linkNode.Text)
	assert.Equal(t, "https://example.com", linkNode.Href)

	// 验证第二行第一个节点（@节点）
	atNode, ok := postMsg.Content[1][0].(*RichTextNodeAt)
	require.True(t, ok)
	assert.Equal(t, PostNodeAt, atNode.GetTag())
	assert.Equal(t, "user123", atNode.UserID)

	// 验证第二行第二个节点（图片节点）
	imgNode, ok := postMsg.Content[1][1].(*RichTextNodeImage)
	require.True(t, ok)
	assert.Equal(t, PostNodeImage, imgNode.GetTag())
	assert.Equal(t, "img_key_123", imgNode.ImageKey)

	// 验证第三行第一个节点（代码块节点）
	codeNode, ok := postMsg.Content[2][0].(*RichTextNodeCodeBlock)
	require.True(t, ok)
	assert.Equal(t, PostNodeCodeBlock, codeNode.GetTag())
	assert.Equal(t, "go", codeNode.Language)
	assert.Equal(t, "fmt.Println(\"Hello World\")", codeNode.Text)
}

func TestPostMsgBody_MarshalJSON(t *testing.T) {
	// 创建测试数据
	postMsg := PostMsgBody{
		Title: "测试消息",
		Content: [][]RichTextNode{
			{
				&RichTextNodeText{
					Tag:  "text",
					Text: "Hello",
				},
				&RichTextNodeLink{
					Tag:  "a",
					Text: "链接",
					Href: "https://example.com",
				},
			},
		},
	}

	// 序列化
	data, err := json.Marshal(postMsg)
	require.NoError(t, err)

	// 反序列化验证
	var result PostMsgBody
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, postMsg.Title, result.Title)
	assert.Len(t, result.Content, 1)
	assert.Len(t, result.Content[0], 2)

	textNode := result.Content[0][0].(*RichTextNodeText)
	assert.Equal(t, "Hello", textNode.Text)

	linkNode := result.Content[0][1].(*RichTextNodeLink)
	assert.Equal(t, "链接", linkNode.Text)
	assert.Equal(t, "https://example.com", linkNode.Href)
}

func TestRichTextNode_Interface(t *testing.T) {
	// 测试所有富文本节点都实现了接口
	var nodes []RichTextNode = []RichTextNode{
		&RichTextNodeText{},
		&RichTextNodeLink{},
		&RichTextNodeAt{},
		&RichTextNodeImage{},
		&RichTextNodeVideo{},
		&RichTextNodeEmotion{},
		&RichTextNodeHr{},
		&RichTextNodeCodeBlock{},
		&RichTextNodeMarkdown{},
	}

	expectedTags := []RichTextNodeType{
		PostNodeText,
		PostNodeLink,
		PostNodeAt,
		PostNodeImage,
		PostNodeMedia,
		PostNodeEmotion,
		PostNodeHr,
		PostNodeCodeBlock,
		PostNodeMarkdown,
	}

	for i, node := range nodes {
		assert.Equal(t, expectedTags[i], node.GetTag())
		// 确保 isRichTextNode 方法存在（编译时检查）
		node.isRichTextNode()
	}
}

func TestPostMsgBody_InvalidTag(t *testing.T) {
	// 测试无效标签
	jsonData := `{
		"title": "测试",
		"content": [
			[
				{
					"tag": "invalid_tag",
					"text": "无效节点"
				}
			]
		]
	}`

	var postMsg PostMsgBody
	err := json.Unmarshal([]byte(jsonData), &postMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的富文本节点类型")
}
