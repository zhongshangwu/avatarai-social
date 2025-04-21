package helper

import (
	"fmt"
	"regexp"
)

type ImagePreset string

const (
	PresetAvatar        ImagePreset = "avatar"
	PresetBanner        ImagePreset = "banner"
	PresetFeedThumbnail ImagePreset = "feed_thumbnail"
	PresetFeedFullsize  ImagePreset = "feed_fullsize"
)

// Options 定义图片处理选项
type Options struct {
	Format string
	Fit    string
	Height int
	Width  int
	Min    bool
}

// BlobLocation 定义图片位置信息
type BlobLocation struct {
	CID string
	DID string
}

// ImageUriBuilder 图片URI构建器
type ImageUriBuilder struct {
	Endpoint string
}

var (
	pathRegex = regexp.MustCompile(`^/(.+?)/plain/(.+?)/(.+?)@(.+?)$`)
	presets   = map[ImagePreset]Options{
		PresetAvatar: {
			Format: "jpeg",
			Fit:    "cover",
			Height: 1000,
			Width:  1000,
			Min:    true,
		},
		PresetBanner: {
			Format: "jpeg",
			Fit:    "cover",
			Height: 1000,
			Width:  3000,
			Min:    true,
		},
		PresetFeedThumbnail: {
			Format: "jpeg",
			Fit:    "inside",
			Height: 2000,
			Width:  2000,
			Min:    true,
		},
		PresetFeedFullsize: {
			Format: "jpeg",
			Fit:    "inside",
			Height: 1000,
			Width:  1000,
			Min:    true,
		},
	}
)

// NewImageUriBuilder 创建新的URI构建器
func NewImageUriBuilder(endpoint string) *ImageUriBuilder {
	return &ImageUriBuilder{
		Endpoint: endpoint,
	}
}

// GetPresetUri 获取预设URI
func (b *ImageUriBuilder) GetPresetUri(preset ImagePreset, did, cid string) (string, error) {
	if _, ok := presets[preset]; !ok {
		return "", fmt.Errorf("未识别的预设类型: %s", preset)
	}

	path := GetPath(preset, did, cid)
	return b.Endpoint + path, nil
}

// GetPath 获取路径
func GetPath(preset ImagePreset, did, cid string) string {
	format := presets[preset].Format
	return fmt.Sprintf("/%s/plain/%s/%s@%s", preset, did, cid, format)
}

// GetOptions 从路径解析选项
func GetOptions(path string) (*Options, *BlobLocation, ImagePreset, error) {
	matches := pathRegex.FindStringSubmatch(path)
	if matches == nil {
		return nil, nil, "", fmt.Errorf("无效的路径")
	}

	presetStr := ImagePreset(matches[1])
	did := matches[2]
	cid := matches[3]
	format := matches[4]

	if _, ok := presets[presetStr]; !ok {
		return nil, nil, "", fmt.Errorf("无效的预设类型")
	}

	if format != "jpeg" && format != "png" {
		return nil, nil, "", fmt.Errorf("无效的格式")
	}

	opts := presets[presetStr]
	loc := &BlobLocation{
		CID: cid,
		DID: did,
	}

	return &opts, loc, presetStr, nil
}

// builder := NewImageUriBuilder("https://bsky.avatar.ai")
// uri, err := builder.GetPresetUri(PresetAvatar, "did:example:123", "cid123")
// if err != nil {
//     // 处理错误
// }
// fmt.Println(uri)
