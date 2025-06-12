package types

import (
	appbskytypes "github.com/bluesky-social/indigo/api/bsky"
)

type FeedCardType string

const (
	FeedCardTypeMoment FeedCardType = "moment"
)

type Feeds struct {
	Cursor string      `json:"cursor"`
	Feed   []*FeedCard `json:"feed"`
}

type MomentThread struct {
	Moment  *MomentCard     `json:"moment"`
	Replies []*MomentThread `json:"replies"`
}

type FeedCard struct {
	Type FeedCardType `json:"type"`
	Card Card         `json:"card"`
}

type Card interface {
	CardType() FeedCardType
}

type MomentCard struct {
	ID         string                        `json:"id"`
	URI        string                        `json:"uri"`
	CID        string                        `json:"cid"`
	Text       string                        `json:"text"`
	Facets     []*appbskytypes.RichtextFacet `json:"facets,omitempty"`
	Reply      *MomentRelyRef                `json:"reply,omitempty"`
	Embed      *EmbedView                    `json:"embed,omitempty"`
	Langs      []string                      `json:"langs,omitempty"`
	Tags       []string                      `json:"tags,omitempty"`
	ReplyCount int                           `json:"replyCount"`
	LikeCount  int                           `json:"likeCount"`
	CreatedAt  int64                         `json:"createdAt"`
	UpdatedAt  int64                         `json:"updatedAt"`
	Author     *SimpleUserView               `json:"author"`
}

func (c *MomentCard) CardType() FeedCardType {
	return FeedCardTypeMoment
}

type SimpleUserView struct {
	Did         string `json:"did"`
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
}

type EmbedView struct {
	Images   []*ImageView  `json:"images,omitempty"`
	Video    *VideoView    `json:"video,omitempty"`
	External *ExternalView `json:"external,omitempty"`
	Record   *RecordView   `json:"record,omitempty"`
}

type ImageView struct {
	Thumb    string `json:"thumb"`
	Fullsize string `json:"fullsize"`
	Alt      string `json:"alt,omitempty"`
}

type VideoView struct {
	Thumb string `json:"thumb,omitempty"`
	Video string `json:"video"`
	Alt   string `json:"alt,omitempty"`
}

type ExternalView struct {
	URI         string `json:"uri"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Thumb       string `json:"thumb,omitempty"`
}

type RecordView struct {
	URI    string          `json:"uri"`
	CID    string          `json:"cid"`
	Author *SimpleUserView `json:"author"`
	Value  interface{}     `json:"value"`
}
