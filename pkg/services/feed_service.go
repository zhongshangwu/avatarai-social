package services

import (
	"context"
	"fmt"
	"log"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type FeedService struct {
	metaStore         *repositories.MetaStore
	feedGenerator     *FeedGenerator
	mockFeedGenerator *MockFeedGenerator
	momentService     *MomentService
}

func NewFeedService(metaStore *repositories.MetaStore) *FeedService {
	return &FeedService{
		metaStore:         metaStore,
		feedGenerator:     NewFeedGenerator(metaStore),
		mockFeedGenerator: NewMockFeedGenerator(metaStore),
		momentService:     NewMomentService(metaStore),
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

		records, err := s.metaStore.MomentRepo.GetMomentsByURIs(batchURIs)
		if err != nil {
			return nil, nil, err
		}

		images, err := s.metaStore.MomentRepo.GetMomentImagesByMomentIDs(batchURIs)
		if err != nil {
			return nil, nil, err
		}

		videos, err := s.metaStore.MomentRepo.GetMomentVideoByMomentIDs(batchURIs)
		if err != nil {
			return nil, nil, err
		}

		externals, err := s.metaStore.MomentRepo.GetMomentExternalByMomentIDs(batchURIs)
		if err != nil {
			return nil, nil, err
		}

		for _, record := range records {
			images := images[record.ID]
			video := videos[record.ID]
			external := externals[record.ID]

			moment := s.momentService.ConvertDBToMoment(record, images, video, external)
			hydrationState[record.URI] = moment
			dids = append(dids, record.Creator)
		}
	}

	return hydrationState, dids, nil
}

func (s *FeedService) hydrateProfiles(ctx context.Context, dids []string) (map[string]interface{}, error) {
	if len(dids) == 0 {
		return nil, nil
	}

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
				avatarURL := fmt.Sprintf("https://bsky.avatar.ai/img/avatar/plain/%s/%s@jpeg", authorProfile.Did, authorProfile.AvatarCID)
				authorView.Avatar = avatarURL
				authorView.CreatedAt = authorProfile.CreatedAt
			}

			momentCard := &types.MomentCard{
				ID:        moment.ID,
				Text:      moment.Text,
				Facets:    moment.Facets,
				Reply:     moment.Reply,
				Langs:     moment.Langs,
				Tags:      moment.Tags,
				CreatedAt: moment.CreatedAt,
				UpdatedAt: moment.UpdatedAt,
				Author:    authorView,
			}

			cards = append(cards, &types.FeedCard{
				Type: types.FeedCardTypeMoment,
				Card: momentCard,
			})
		}
	}
	return cards
}
