package services

import (
	"context"
	"fmt"
	"log"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/blobs"
	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/config"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type FeedService struct {
	metaStore         *repositories.MetaStore
	feedGenerator     *FeedGenerator
	mockFeedGenerator *MockFeedGenerator
	momentService     *MomentService
	imageBuilder      *blobs.ImageUriBuilder
}

func NewFeedService(config *config.SocialConfig, metaStore *repositories.MetaStore) *FeedService {
	return &FeedService{
		metaStore:         metaStore,
		feedGenerator:     NewFeedGenerator(metaStore),
		mockFeedGenerator: NewMockFeedGenerator(metaStore),
		momentService:     NewMomentService(metaStore),
		imageBuilder:      blobs.NewImageUriBuilder(config.Server.Domain),
	}
}

func (s *FeedService) Feeds(ctx context.Context, feedName string, limit int, cursor string) (*types.Feeds, error) {
	var uris []string
	var nextCursor string
	var err error

	feeds := &types.Feeds{
		Feed:   nil,
		Cursor: nextCursor,
	}

	switch feedName {
	case "default":
		uris, nextCursor, err = s.feedGenerator.GenerateFeed(ctx, limit, cursor)
	default:
		uris, nextCursor, err = s.mockFeedGenerator.GenerateFeed(ctx, limit, cursor)
	}

	if err != nil {
		return feeds, err
	}

	hydrationState, err := s.hydrate(ctx, uris)
	if err != nil {
		return feeds, err
	}

	cards := s.presentCards(uris, hydrationState)
	feeds.Feed = cards
	return feeds, nil
}

func (s *FeedService) MomentThread(ctx context.Context, uri string, depth int) (*types.MomentThread, error) {
	aturi, err := helper.BuildAtURI(uri)
	if err != nil {
		return nil, err
	}

	if aturi.Collection() != "app.vtri.activity.moment" {
		return nil, fmt.Errorf("只能获取 moment 记录的帖子")
	}

	momentID := string(aturi.RecordKey())

	if depth <= 0 {
		depth = 10 // 默认最大深度
	}

	allMoments, err := s.metaStore.MomentRepo.GetMomentThread(momentID, depth, depth)
	if err != nil {
		return nil, fmt.Errorf("获取 thread moments 失败: %w", err)
	}

	if len(allMoments) == 0 {
		return nil, fmt.Errorf("未找到指定的 moment")
	}

	momentURIs := make([]string, len(allMoments))
	for i, moment := range allMoments {
		momentURIs[i] = moment.URI
	}

	hydrationState, err := s.hydrate(ctx, momentURIs)
	if err != nil {
		return nil, fmt.Errorf("水合数据失败: %w", err)
	}

	thread, err := s.buildMomentThread(momentID, allMoments, hydrationState)
	if err != nil {
		return nil, fmt.Errorf("构建 thread 结构失败: %w", err)
	}

	return thread, nil
}

func (s *FeedService) hydrate(ctx context.Context, uris []string) (map[string]interface{}, error) {
	dids := make([]string, 0, len(uris))
	hydrationState := make(map[string]interface{})

	momentURIs := make([]string, 0, len(uris))
	for _, uri := range uris {
		aturi, err := helper.BuildAtURI(uri)
		if err != nil {
			log.Printf("无法解析记录 URI: %s", err)
			continue // 跳过无法解析的记录
		}

		switch aturi.Collection() {
		case "app.vtri.activity.moment":
			momentURIs = append(momentURIs, uri)
		default:
			log.Printf("记录 URI 不是 moment: %s", uri)
			continue // 跳过非 moment 记录
		}
	}

	moments, momentDids, err := s.hydrateMoments(ctx, momentURIs)
	if err != nil {
		return nil, err
	}
	dids = append(dids, momentDids...)

	profiles, err := s.hydrateProfiles(ctx, dids)
	if err != nil {
		return nil, err
	}

	for key, moment := range moments {
		hydrationState[key] = moment
	}
	for key, profile := range profiles {
		hydrationState[key] = profile
	}
	return hydrationState, nil
}

func (s *FeedService) hydrateMoments(ctx context.Context, uris []string) (map[string]interface{}, []string, error) {
	hydrationState := make(map[string]interface{})
	dids := make([]string, 0, len(uris))

	batchSize := 25
	for i := 0; i < len(uris); i += batchSize {
		end := i + batchSize
		if end > len(uris) {
			end = len(uris)
		}
		batchURIs := uris[i:end]

		moments, err := s.metaStore.MomentRepo.GetMomentsByURIs(batchURIs)
		if err != nil {
			return nil, nil, err
		}

		momentIDs := make([]string, 0, len(moments))
		momentURIs := make([]string, 0, len(moments))
		for _, record := range moments {
			momentIDs = append(momentIDs, record.ID)
			momentURIs = append(momentURIs, record.URI)
		}

		images, err := s.metaStore.MomentRepo.GetMomentImagesByMomentIDs(momentIDs)
		if err != nil {
			return nil, nil, err
		}

		videos, err := s.metaStore.MomentRepo.GetMomentVideoByMomentIDs(momentIDs)
		if err != nil {
			return nil, nil, err
		}

		externals, err := s.metaStore.MomentRepo.GetMomentExternalByMomentIDs(momentIDs)
		if err != nil {
			return nil, nil, err
		}

		activityTags, err := s.metaStore.ActivityRepo.GetActivityTagsBySubjectURIs(momentURIs)
		if err != nil {
			return nil, nil, err
		}

		activityTopics, err := s.metaStore.ActivityRepo.GetActivityTopicsBySubjectURIs(momentURIs)
		if err != nil {
			return nil, nil, err
		}

		for _, record := range moments {
			images := images[record.ID]
			video := videos[record.ID]
			external := externals[record.ID]
			activityTags := activityTags[record.URI]
			activityTopics := activityTopics[record.URI]
			moment := s.momentService.ConvertDBToMoment(record, images, video, external, activityTags, activityTopics)
			hydrationState[record.URI] = moment

			dids = append(dids, record.Creator)
			for _, activityTag := range activityTags {
				dids = append(dids, activityTag.Creator)
			}
			for _, activityTopic := range activityTopics {
				dids = append(dids, activityTopic.Creator)
			}
		}
	}

	return hydrationState, dids, nil
}

func (s *FeedService) hydrateProfiles(ctx context.Context, dids []string) (map[string]interface{}, error) {
	if len(dids) == 0 {
		return nil, nil
	}
	dids = deduplicate(dids)

	hydrationState := make(map[string]interface{})

	batchSize := 25
	for i := 0; i < len(dids); i += batchSize {
		end := i + batchSize
		if end > len(dids) {
			end = len(dids)
		}
		batchDIDs := dids[i:end]

		avatars, err := s.metaStore.UserRepo.GetUsersByDIDs(batchDIDs)
		if err != nil {
			log.Printf("获取用户资料失败: %s", err)
			return nil, err
		}

		for _, avatar := range avatars {
			hydrationState["profile:"+avatar.Did] = avatar
		}
	}
	return hydrationState, nil
}

func (s *FeedService) presentCards(uris []string, hydrationState map[string]interface{}) []*types.FeedCard {
	var cards []*types.FeedCard

	for _, uri := range uris {
		if _, ok := hydrationState[uri]; !ok {
			continue
		}

		aturi, _ := helper.BuildAtURI(uri)
		switch aturi.Collection() {
		case "app.vtri.activity.moment":
			moment, ok := hydrationState[uri].(*types.Moment)
			if !ok {
				continue
			}

			authorDID := moment.CreatedBy
			authorProfile, hasProfile := hydrationState["profile:"+authorDID].(*repositories.Avatar)

			authorView := &types.SimpleUserView{
				Did: authorDID,
			}

			if hasProfile {
				authorView.Handle = authorProfile.Handle
				authorView.DisplayName = authorProfile.DisplayName
				avatarURL, _ := s.imageBuilder.GetPresetUri(blobs.PresetAvatar, authorDID, authorProfile.AvatarCID)
				authorView.Avatar = avatarURL
				authorView.CreatedAt = authorProfile.CreatedAt
			}

			var embed *types.EmbedView
			if moment.Embed != nil {
				embed = &types.EmbedView{}

				if moment.Embed.External != nil {
					embed.External = &types.ExternalView{
						URI:         moment.Embed.External.URI,
						Title:       moment.Embed.External.Title,
						Description: moment.Embed.External.Description,
					}
				}

				if moment.Embed.Images != nil {
					embed.Images = make([]*types.ImageView, len(moment.Embed.Images))
					for i, image := range moment.Embed.Images {
						thumb, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedThumbnail, authorDID, image.CID)
						fullsize, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedFullsize, authorDID, image.CID)
						embed.Images[i] = &types.ImageView{
							Thumb:    thumb,
							Fullsize: fullsize,
							Alt:      image.Alt,
						}
					}
				}

				if moment.Embed.Video != nil {
					thumb, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedThumbnail, authorDID, moment.Embed.Video.CID)
					embed.Video = &types.VideoView{
						Thumb: thumb,
						Video: moment.Embed.Video.URL,
						Alt:   moment.Embed.Video.Alt,
					}
				}
			}

			tagViews := make([]*types.TagView, 0, len(moment.Tags))
			for _, tag := range moment.Tags {
				var isAster bool
				creatorProfile, hasProfile := hydrationState["profile:"+tag.Creator].(*repositories.Avatar)
				if hasProfile {
					isAster = creatorProfile.IsAster
				}
				tagViews = append(tagViews, &types.TagView{
					Tag:     tag.Tag,
					Creator: tag.Creator,
					IsAster: isAster,
				})
			}

			topicViews := make([]*types.TopicView, 0, len(moment.Topics))
			for _, topic := range moment.Topics {
				var isAster bool
				creatorProfile, hasProfile := hydrationState["profile:"+topic.Creator].(*repositories.Avatar)
				if hasProfile {
					isAster = creatorProfile.IsAster
				}
				topicViews = append(topicViews, &types.TopicView{
					Topic:   topic.Topic,
					Creator: topic.Creator,
					IsAster: isAster,
				})
			}

			momentCard := &types.MomentCard{
				ID:        moment.ID,
				Text:      moment.Text,
				Facets:    moment.Facets,
				Reply:     moment.Reply,
				Embed:     embed,
				Langs:     moment.Langs,
				Tags:      tagViews,
				Topics:    topicViews,
				CreatedAt: moment.CreatedAt,
				UpdatedAt: moment.UpdatedAt,
				Author:    authorView,
			}

			cards = append(cards, &types.FeedCard{
				Type: types.ActivityCardTypeMoment,
				Card: momentCard,
			})
		}
	}
	return cards
}

// buildMomentThread 构建 moment thread 的嵌套结构
func (s *FeedService) buildMomentThread(targetMomentID string, allMoments []*repositories.Moment, hydrationState map[string]interface{}) (*types.MomentThread, error) {
	// 创建 moment 索引
	momentMap := make(map[string]*repositories.Moment)
	for _, moment := range allMoments {
		momentMap[moment.ID] = moment
	}

	// 构建子级映射 (parentID -> children)
	childrenMap := make(map[string][]*repositories.Moment)
	for _, moment := range allMoments {
		if moment.ReplyParentID != "" {
			childrenMap[moment.ReplyParentID] = append(childrenMap[moment.ReplyParentID], moment)
		}
	}

	// 获取目标 moment
	targetMoment, exists := momentMap[targetMomentID]
	if !exists {
		return nil, fmt.Errorf("目标 moment 不存在")
	}

	// 递归构建 thread
	return s.buildMomentThreadRecursive(targetMoment, childrenMap, hydrationState), nil
}

// buildMomentThreadRecursive 递归构建 moment thread
func (s *FeedService) buildMomentThreadRecursive(moment *repositories.Moment, childrenMap map[string][]*repositories.Moment, hydrationState map[string]interface{}) *types.MomentThread {
	// 构建当前 moment 的 card
	momentCard := s.buildMomentCard(moment, hydrationState)

	// 构建回复列表
	var replies []*types.MomentThread
	children := childrenMap[moment.ID]
	for _, child := range children {
		childThread := s.buildMomentThreadRecursive(child, childrenMap, hydrationState)
		replies = append(replies, childThread)
	}

	return &types.MomentThread{
		Moment:  momentCard,
		Replies: replies,
	}
}

// buildMomentCard 构建 moment card
func (s *FeedService) buildMomentCard(moment *repositories.Moment, hydrationState map[string]interface{}) *types.MomentCard {
	// 获取 moment 数据
	momentData, ok := hydrationState[moment.URI].(*types.Moment)
	if !ok {
		// 如果水合数据中没有，创建基本的 moment 数据
		momentData = &types.Moment{
			ID:        moment.ID,
			Text:      moment.Text,
			CreatedAt: moment.CreatedAt,
			UpdatedAt: moment.UpdatedAt,
			CreatedBy: moment.Creator,
		}
	}

	// 获取作者信息
	authorProfile, hasProfile := hydrationState["profile:"+moment.Creator].(*repositories.Avatar)

	authorView := &types.SimpleUserView{
		Did: moment.Creator,
	}

	if hasProfile {
		authorView.Handle = authorProfile.Handle
		authorView.DisplayName = authorProfile.DisplayName
		avatarURL := fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg", authorProfile.Did, authorProfile.AvatarCID)
		authorView.Avatar = avatarURL
		authorView.CreatedAt = authorProfile.CreatedAt
	}

	// 构建嵌入内容视图
	var embed *types.EmbedView
	if momentData.Embed != nil {
		embed = &types.EmbedView{}

		if momentData.Embed.External != nil {
			embed.External = &types.ExternalView{
				URI:         momentData.Embed.External.URI,
				Title:       momentData.Embed.External.Title,
				Description: momentData.Embed.External.Description,
			}
		}

		if momentData.Embed.Images != nil {
			embed.Images = make([]*types.ImageView, len(momentData.Embed.Images))
			for i, image := range momentData.Embed.Images {
				thumb, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedThumbnail, moment.Creator, image.CID)
				fullsize, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedFullsize, moment.Creator, image.CID)
				embed.Images[i] = &types.ImageView{
					Thumb:    thumb,
					Fullsize: fullsize,
					Alt:      image.Alt,
				}
			}
		}

		if momentData.Embed.Video != nil {
			thumb, _ := s.imageBuilder.GetPresetUri(blobs.PresetFeedThumbnail, moment.Creator, momentData.Embed.Video.CID)
			embed.Video = &types.VideoView{
				Thumb: thumb,
				Video: momentData.Embed.Video.URL,
				Alt:   momentData.Embed.Video.Alt,
			}
		}
	}

	tagViews := make([]*types.TagView, 0, len(momentData.Tags))

	for _, tag := range momentData.Tags {
		var isAster bool
		creatorProfile, hasProfile := hydrationState["profile:"+tag.Creator].(*repositories.Avatar)
		if hasProfile {
			isAster = creatorProfile.IsAster
		}
		tagViews = append(tagViews, &types.TagView{
			Tag:     tag.Tag,
			Creator: tag.Creator,
			IsAster: isAster,
		})
	}

	topicViews := make([]*types.TopicView, 0, len(momentData.Topics))
	for _, topic := range momentData.Topics {
		var isAster bool
		creatorProfile, hasProfile := hydrationState["profile:"+topic.Creator].(*repositories.Avatar)
		if hasProfile {
			isAster = creatorProfile.IsAster
		}
		topicViews = append(topicViews, &types.TopicView{
			Topic:   topic.Topic,
			Creator: topic.Creator,
			IsAster: isAster,
		})
	}

	return &types.MomentCard{
		ID:        momentData.ID,
		URI:       moment.URI,
		CID:       moment.CID,
		Text:      momentData.Text,
		Facets:    momentData.Facets,
		Reply:     momentData.Reply,
		Embed:     embed,
		Langs:     momentData.Langs,
		Tags:      tagViews,
		Topics:    topicViews,
		CreatedAt: momentData.CreatedAt,
		UpdatedAt: momentData.UpdatedAt,
		Author:    authorView,
	}
}

func deduplicate(dids []string) []string {
	seen := make(map[string]bool)
	var ret []string
	for _, did := range dids {
		if !seen[did] {
			seen[did] = true
			ret = append(ret, did)
		}
	}
	return ret
}
