package types

import (
	appbskytypes "github.com/bluesky-social/indigo/api/bsky"
)

type Moment struct {
	ID        string                        `json:"id"`
	Text      string                        `json:"text"`
	Facets    []*appbskytypes.RichtextFacet `json:"facets,omitempty"`
	Langs     []string                      `json:"langs"`
	Tags      []string                      `json:"tags"`
	Reply     *MomentRelyRef                `json:"reply,omitempty"`
	Embed     *EmbedContent                 `json:"embed,omitempty"`
	CreatedAt int64                         `json:"createdAt"`
	UpdatedAt int64                         `json:"updatedAt"`
	IndexedAt int64                         `json:"indexedAt"`
	CreatedBy string                        `json:"createdBy"`
	Deleted   bool                          `json:"deleted"`
}

type MomentRelyRef struct {
	Parent *RefLink `json:"parent,omitempty"`
	Root   *RefLink `json:"root,omitempty"`
}

type RefLink struct {
	ID string `json:"id"`
}

type EmbedContent struct {
	Images   []*ImageEmbed  `json:"images,omitempty"`
	Video    *VideoEmbed    `json:"video,omitempty"`
	External *ExternalEmbed `json:"external,omitempty"`
	Record   *RecordEmbed   `json:"record,omitempty"`
}

type ImageEmbed struct {
	CID string `json:"cid"`
	Alt string `json:"alt,omitempty"`
	URL string `json:"url,omitempty"`
}

type VideoEmbed struct {
	CID string `json:"cid"`
	Alt string `json:"alt,omitempty"`
	URL string `json:"url,omitempty"`
}

type ExternalEmbed struct {
	URI         string `json:"uri"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ThumbCID    string `json:"thumbCid,omitempty"`
	ThumbURL    string `json:"thumbURL,omitempty"`
}

type RecordEmbed struct {
	Type     string `json:"type"`
	RecordID string `json:"recordId"`
}
