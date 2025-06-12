package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	appbskytypes "github.com/bluesky-social/indigo/api/bsky"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type CreateMomentRequest struct {
	Text     string                        `json:"text"`               // 文本内容
	Facets   []*appbskytypes.RichtextFacet `json:"facets,omitempty"`   // 富文本注解
	RootID   string                        `json:"rootId,omitempty"`   // 根帖子ID
	ParentID string                        `json:"parentId,omitempty"` // 父帖子ID
	Images   []*BlobData                   `json:"images,omitempty"`   // 图片引用
	Video    *BlobData                     `json:"video,omitempty"`    // 视频引用
	External *ExternalData                 `json:"external,omitempty"` // 外部链接
	Langs    []string                      `json:"langs,omitempty"`    // 语言标签
	Tags     []string                      `json:"tags,omitempty"`     // 标签
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
	momentId := s.GenerateMomentID()

	facets, err := json.Marshal(req.Facets)
	if err != nil {
		return nil, fmt.Errorf("序列化富文本注解失败: %w", err)
	}

	dbMoment := &repositories.Moment{
		ID:        momentId,
		URI:       s.BuildAtURI(creatorDid, momentId),
		CID:       "",
		Creator:   creatorDid,
		Text:      req.Text,
		Facets:    string(facets),
		Langs:     req.Langs,
		Tags:      req.Tags,
		CreatedAt: now,
		IndexedAt: 0,
	}

	// 处理回复逻辑
	if req.ParentID != "" {
		dbMoment.ReplyParentID = req.ParentID

		// 如果没有指定 RootID，需要自动查找或设置
		if req.RootID != "" {
			dbMoment.ReplyRootID = req.RootID
		} else {
			// 获取父 moment，确定根 moment
			parentMoment, err := s.metaStore.MomentRepo.GetMomentByID(req.ParentID)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("获取父 moment 失败: %w", err)
			}

			if parentMoment.ReplyRootID != "" {
				// 父 moment 是回复，使用相同的根
				dbMoment.ReplyRootID = parentMoment.ReplyRootID
			} else {
				// 父 moment 是顶级帖子，它就是根
				dbMoment.ReplyRootID = parentMoment.ID
			}
		}
	} else if req.RootID != "" {
		// 如果只指定了 RootID 而没有 ParentID，则是直接回复根帖子
		dbMoment.ReplyRootID = req.RootID
		dbMoment.ReplyParentID = req.RootID
	}

	if err := s.metaStore.MomentRepo.CreateMoment(dbMoment); err != nil {
		return nil, fmt.Errorf("保存moment记录失败: %w", err)
	}

	var (
		images   []*repositories.MomentImage
		video    *repositories.MomentVideo
		external *repositories.MomentExternal
	)

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

func (s *MomentService) LikeMoment(ctx context.Context, uri string, did string) (*repositories.Like, error) {
	aturi, err := helper.BuildAtURI(uri)
	if err != nil {
		return nil, err
	}

	if aturi.Collection() != "app.vtri.activity.moment" {
		return nil, fmt.Errorf("只能点赞 moment 记录的帖子")
	}

	momentID := string(aturi.RecordKey())

	moment, err := s.metaStore.MomentRepo.GetMomentByID(momentID)
	if err != nil {
		return nil, fmt.Errorf("获取 moment 失败: %w", err)
	}

	id := helper.GenerateTID()
	like := &repositories.Like{
		ID:         id,
		URI:        s.BuildLikeURI(did, id),
		CID:        "",
		Creator:    did,
		SubjectURI: moment.URI,
		SubjectCid: moment.CID,
		CreatedAt:  time.Now().Unix(),
		IndexedAt:  0,
	}
	if err := s.metaStore.MomentRepo.CreateLike(like); err != nil {
		return nil, fmt.Errorf("点赞失败: %w", err)
	}
	return like, nil
}

func (s *MomentService) UndoLikeMoment(ctx context.Context, uri string, did string, likeURI string) error {
	aturi, err := helper.BuildAtURI(uri)
	if err != nil {
		return err
	}

	if aturi.Collection() != "app.vtri.activity.moment" {
		return fmt.Errorf("只能点赞 moment 记录的帖子")
	}

	likeAturi, err := helper.BuildAtURI(likeURI)
	if err != nil {
		return err
	}
	if likeAturi.Collection() != "app.vtri.activity.like" {
		return fmt.Errorf("只能取消点赞 moment 记录的帖子")
	}

	if err := s.metaStore.MomentRepo.DeleteLike(likeURI); err != nil {
		return fmt.Errorf("取消点赞失败: %w", err)
	}
	return nil
}

func (s *MomentService) loadEmbedContent(momentID string) (
	[]*repositories.MomentImage,
	*repositories.MomentVideo,
	*repositories.MomentExternal,
	error,
) {
	images, err := s.metaStore.MomentRepo.GetMomentImages(momentID)
	// 需要先获取 moment 来得到 URI（或者修改数据库方法使用 moment_id）
	moment, errMoment := s.metaStore.MomentRepo.GetMomentByID(momentID)
	if errMoment != nil {
		return nil, nil, nil, errMoment
	}

	video, err2 := s.metaStore.MomentRepo.GetMomentVideo(moment.URI)
	external, err3 := s.metaStore.MomentRepo.GetMomentExternal(moment.URI)
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
	var reply *types.MomentRelyRef
	if moment.ReplyRootID != "" || moment.ReplyParentID != "" {
		reply = &types.MomentRelyRef{
			Root: &types.RefLink{
				ID: moment.ReplyRootID,
			},
			Parent: &types.RefLink{
				ID: moment.ReplyParentID,
			},
		}
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

func (s *MomentService) GenerateMomentID() string {
	return helper.GenerateTID()
}

func (s *MomentService) BuildAtURI(did string, rkey string) string {
	return fmt.Sprintf("at://%s/app.vtri.activity.moment/%s", did, rkey)
}

func (s *MomentService) BuildLikeURI(did string, rkey string) string {
	return fmt.Sprintf("at://%s/app.vtri.activity.like/%s", did, rkey)
}
