package services

import (
	"context"
	"fmt"
	"time"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type TopicService struct {
	metaStore *repositories.MetaStore
}

func NewTopicService(metaStore *repositories.MetaStore) *TopicService {
	return &TopicService{metaStore: metaStore}
}

func (s *TopicService) CreateOrGetTopic(ctx context.Context, topic string, creator string) (*repositories.Topic, error) {
	// 先尝试获取现有主题
	existingTopic, err := s.metaStore.MomentRepo.GetTopicByTopic(topic)
	if err == nil {
		return existingTopic, nil
	}

	// 如果主题不存在，创建新主题
	newTopic := &repositories.Topic{
		ID:        helper.GenerateTID(),
		Topic:     topic,
		CreatedAt: time.Now().Unix(),
		Creator:   creator,
		Deleted:   false,
	}

	if err := s.metaStore.MomentRepo.CreateTopic(newTopic); err != nil {
		return nil, fmt.Errorf("创建主题失败: %w", err)
	}

	return newTopic, nil
}

func (s *TopicService) CreateActivityTopic(ctx context.Context, uri string, topic string, creator string) (*repositories.ActivityTopic, error) {
	activityTopic := &repositories.ActivityTopic{
		ID:         helper.GenerateTID(),
		SubjectURI: uri,
		Topic:      topic,
		CreatedAt:  time.Now().Unix(),
		Creator:    creator,
		Deleted:    false,
	}

	if err := s.metaStore.MomentRepo.CreateActivityTopic(activityTopic); err != nil {
		return nil, fmt.Errorf("创建活动主题关联失败: %w", err)
	}

	return activityTopic, nil
}

func (s *TopicService) SyncActivityTopics(ctx context.Context, uri string, topics []string, creator string) error {
	if len(topics) == 0 {
		return nil
	}

	for _, topic := range topics {
		if topic == "" {
			continue
		}

		// 确保主题存在于主题池中
		_, err := s.CreateOrGetTopic(ctx, topic, creator)
		if err != nil {
			return fmt.Errorf("处理主题 '%s' 失败: %w", topic, err)
		}

		// 创建活动主题关联
		_, err = s.CreateActivityTopic(ctx, uri, topic, creator)
		if err != nil {
			return fmt.Errorf("创建主题关联 '%s' 失败: %w", topic, err)
		}
	}

	return nil
}

func (s *TopicService) UnbindActivityTopics(ctx context.Context, subjectURI string) error {
	// 获取该 subject 的所有主题关联
	activityTopics, err := s.metaStore.MomentRepo.GetActivityTopicsBySubjectURI(subjectURI)
	if err != nil {
		return fmt.Errorf("获取主题关联失败: %w", err)
	}

	// 删除所有主题关联
	for _, activityTopic := range activityTopics {
		if err := s.metaStore.MomentRepo.DeleteActivityTopic(activityTopic.Topic, subjectURI); err != nil {
			return fmt.Errorf("删除主题关联失败: %w", err)
		}
	}

	return nil
}

func (s *TopicService) ListTopics(ctx context.Context, page int, pageSize int) ([]*repositories.Topic, error) {
	return s.metaStore.MomentRepo.ListTopics(page, pageSize)
}

func (s *TopicService) PresentTopicView(topics []*repositories.Topic) (map[string]*types.TopicView, error) {
	ret := make(map[string]*types.TopicView)

	topicList := make([]string, 0, len(topics))
	for _, topic := range topics {
		topicList = append(topicList, topic.Topic)
	}
	creatorList := make([]string, 0, len(topics))
	for _, topic := range topics {
		creatorList = append(creatorList, topic.Creator)
	}

	topicCounts, err := s.metaStore.MomentRepo.GetTopicActivityCounts(topicList)
	if err != nil {
		return nil, fmt.Errorf("获取标签关联失败: %w", err)
	}

	// 2. 批量获取 tag 的创建者
	tagCreators, err := s.metaStore.UserRepo.GetUsersByDIDs(creatorList)
	if err != nil {
		return nil, fmt.Errorf("获取标签创建者失败: %w", err)
	}
	creatorMap := make(map[string]*repositories.Avatar)
	for _, user := range tagCreators {
		creatorMap[user.Did] = user
	}

	for _, topic := range topics {
		creator, ok := creatorMap[topic.Creator]
		if !ok {
			continue
		}

		ret[topic.Topic] = &types.TopicView{
			Topic:   topic.Topic,
			Count:   topicCounts[topic.Topic],
			Creator: topic.Creator,
			IsAster: creator.IsAster,
		}
	}
	return ret, nil
}

func (s *TopicService) PresentActivityTopicView(topics []*repositories.ActivityTopic) (map[string]*types.TopicView, error) {
	ret := make(map[string]*types.TopicView)

	topicList := make([]string, 0, len(topics))
	for _, topic := range topics {
		topicList = append(topicList, topic.Topic)
	}
	creatorList := make([]string, 0, len(topics))
	for _, topic := range topics {
		creatorList = append(creatorList, topic.Creator)
	}

	topicCounts, err := s.metaStore.MomentRepo.GetTopicActivityCounts(topicList)
	if err != nil {
		return nil, fmt.Errorf("获取主题关联失败: %w", err)
	}

	// 2. 批量获取 tag 的创建者
	tagCreators, err := s.metaStore.UserRepo.GetUsersByDIDs(creatorList)
	if err != nil {
		return nil, fmt.Errorf("获取主题创建者失败: %w", err)
	}
	creatorMap := make(map[string]*repositories.Avatar)
	for _, user := range tagCreators {
		creatorMap[user.Did] = user
	}

	for _, topic := range topics {
		creator, ok := creatorMap[topic.Creator]
		if !ok {
			continue
		}

		ret[topic.Topic] = &types.TopicView{
			Topic:   topic.Topic,
			Count:   topicCounts[topic.Topic],
			Creator: topic.Creator,
			IsAster: creator.IsAster,
		}
	}
	return ret, nil
}
