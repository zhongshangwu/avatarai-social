package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	appbskytypes "github.com/bluesky-social/indigo/api/bsky"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type CreateMomentRequest struct {
	Text           string                        `json:"text"`                     // 文本内容
	Facets         []*appbskytypes.RichtextFacet `json:"facets,omitempty"`         // 富文本注解
	RootMomentID   string                        `json:"rootMomentId,omitempty"`   // 根帖子ID
	ParentMomentID string                        `json:"parentMomentId,omitempty"` // 父帖子ID
	Images         []*BlobData                   `json:"images,omitempty"`         // 图片引用
	Video          *BlobData                     `json:"video,omitempty"`          // 视频引用
	External       *ExternalData                 `json:"external,omitempty"`       // 外部链接
	Langs          []string                      `json:"langs,omitempty"`          // 语言标签
	Tags           []string                      `json:"tags,omitempty"`           // 标签
}

type BlobData struct {
	CID string `json:"cid"`
}

type ExternalData struct {
	URI         string `json:"uri"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	ThumbCID    string `json:"thumbCid,omitempty"`
}

type MomentService struct {
	metaStore *repositories.MetaStore
}

func NewMomentService(metaStore *repositories.MetaStore) *MomentService {
	return &MomentService{
		metaStore: metaStore,
	}
}

func (s *MomentService) CreateMoment(ctx context.Context, creatorDid string, req *CreateMomentRequest) (*types.Moment, error) {
	tx := s.metaStore.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("开始数据库事务失败: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now().Unix()
	momentId := s.metaStore.MomentRepo.GenerateMomentID()

	facets, err := json.Marshal(req.Facets)
	if err != nil {
		return nil, fmt.Errorf("序列化富文本注解失败: %w", err)
	}

	dbMoment := &repositories.Moment{
		ID:        momentId,
		URI:       "",
		CID:       "",
		Creator:   creatorDid,
		Text:      req.Text,
		Facets:    string(facets),
		Langs:     req.Langs,
		Tags:      req.Tags,
		CreatedAt: now,
		IndexedAt: 0,
	}

	if req.RootMomentID != "" {
		dbMoment.ReplyRootID = req.RootMomentID
	}

	if req.ParentMomentID != "" {
		dbMoment.ReplyParentID = req.ParentMomentID
	}

	if err := s.metaStore.MomentRepo.CreateMoment(dbMoment); err != nil {
		return nil, fmt.Errorf("保存moment记录失败: %w", err)
	}

	var (
		images   []*repositories.MomentImage
		video    *repositories.MomentVideo
		external *repositories.MomentExternal
	)

	// 保存图片
	if len(req.Images) > 0 {
		for i, img := range req.Images {
			dbImage := &repositories.MomentImage{
				MomentID: momentId,
				Position: i,
				ImageCID: img.CID,
				Alt:      "",
			}
			images = append(images, dbImage)
			if err := s.metaStore.MomentRepo.CreateMomentImage(dbImage); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("保存图片记录失败: %w", err)
			}
		}
	}

	// 保存视频
	if req.Video != nil {
		video = &repositories.MomentVideo{
			MomentID: momentId,
			VideoCID: req.Video.CID,
			Alt:      "",
		}
		if err := s.metaStore.MomentRepo.CreateMomentVideo(video); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("保存视频记录失败: %w", err)
		}
	}

	// 保存外部链接
	if req.External != nil {
		external = &repositories.MomentExternal{
			MomentID:    momentId,
			URI:         req.External.URI,
			Title:       req.External.Title,
			Description: req.External.Description,
			ThumbCID:    req.External.ThumbCID,
		}
		if err := s.metaStore.MomentRepo.CreateMomentExternal(external); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("保存外部链接记录失败: %w", err)
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return s.ConvertDBToMoment(dbMoment, images, video, external), nil
}

func (s *MomentService) GetMomentByID(ctx context.Context, id string) (*types.Moment, error) {
	moment, err := s.metaStore.MomentRepo.GetMomentByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取moment失败: %w", err)
	}
	images, video, external, err := s.loadEmbedContent(id)
	if err != nil {
		log.Printf("加载嵌入内容失败: %v", err)
		return nil, err
	}

	return s.ConvertDBToMoment(moment, images, video, external), nil
}

func (s *MomentService) loadEmbedContent(momentID string) (
	[]*repositories.MomentImage,
	*repositories.MomentVideo,
	*repositories.MomentExternal,
	error,
) {
	images, err := s.metaStore.MomentRepo.GetMomentImages(momentID)
	video, err2 := s.metaStore.MomentRepo.GetMomentVideo(momentID)
	external, err3 := s.metaStore.MomentRepo.GetMomentExternal(momentID)
	unionError := errors.Join(err, err2, err3)
	return images, video, external, unionError
}

func (s *MomentService) ConvertDBToMoment(
	moment *repositories.Moment,
	images []*repositories.MomentImage,
	video *repositories.MomentVideo,
	external *repositories.MomentExternal,
) *types.Moment {

	facets := make([]*appbskytypes.RichtextFacet, 0)

	if moment.Facets != "" {
		if err := json.Unmarshal([]byte(moment.Facets), &facets); err != nil {
			log.Printf("反序列化富文本注解失败: %v", err)
		}
	}

	reply := &types.MomentRelyRef{
		Root: &types.RefLink{
			ID: moment.ReplyRootID,
		},
		Parent: &types.RefLink{
			ID: moment.ReplyParentID,
		},
	}

	embed := &types.EmbedContent{
		Images: make([]*types.ImageEmbed, 0, len(images)),
	}
	if len(images) > 0 {
		for _, img := range images {
			embed.Images = append(embed.Images, &types.ImageEmbed{
				CID: img.ImageCID,
				Alt: img.Alt,
			})
		}
	}
	if video != nil {
		embed.Video = &types.VideoEmbed{
			CID: video.VideoCID,
			Alt: video.Alt,
		}
	}
	if external != nil {
		embed.External = &types.ExternalEmbed{
			URI:         external.URI,
			Title:       external.Title,
			Description: external.Description,
			ThumbCID:    external.ThumbCID,
		}
	}

	return &types.Moment{
		ID:        moment.ID,
		Text:      moment.Text,
		Facets:    facets,
		Langs:     moment.Langs,
		Tags:      moment.Tags,
		Reply:     reply,
		Embed:     embed,
		CreatedAt: moment.CreatedAt,
		UpdatedAt: moment.UpdatedAt,
		IndexedAt: moment.IndexedAt,
		CreatedBy: moment.Creator,
		Deleted:   moment.Deleted,
	}
}

func parseTime(timeStr string) time.Time {
	if timeStr == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Printf("解析时间失败: %v", err)
		return time.Time{}
	}

	return t
}
