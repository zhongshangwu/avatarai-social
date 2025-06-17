package blobs

import (
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
)

type ImagePreset string

const (
	PresetAvatar        ImagePreset = "avatar"
	PresetBanner        ImagePreset = "banner"
	PresetFeedThumbnail ImagePreset = "feed_thumbnail"
	PresetFeedFullsize  ImagePreset = "feed_fullsize"
)

type Options struct {
	Format string
	Fit    string
	Height int
	Width  int
	Min    bool
}

type BlobLocation struct {
	CID string
	DID string
}

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

func NewImageUriBuilder(endpoint string) *ImageUriBuilder {
	return &ImageUriBuilder{
		Endpoint: endpoint + "/img",
	}
}

func (b *ImageUriBuilder) GetPresetUri(preset ImagePreset, did, cid string) (string, error) {
	if cid == "" {
		return "", nil
	}

	if _, ok := presets[preset]; !ok {
		return "", fmt.Errorf("未识别的预设类型: %s", preset)
	}

	logrus.Infof("GetPresetUri: %s, %s, %s, %s", b.Endpoint, preset, did, cid)

	path := GetPath(preset, did, cid)
	return b.Endpoint + path, nil
}

func GetPath(preset ImagePreset, did, cid string) string {
	format := presets[preset].Format
	return fmt.Sprintf("/%s/plain/%s/%s@%s", preset, did, cid, format)
}

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
