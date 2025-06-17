package services

import (
	"context"
	"fmt"
	"time"

	"github.com/zhongshangwu/avatarai-social/pkg/atproto/helper"
	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
	"github.com/zhongshangwu/avatarai-social/types"
)

type TagService struct {
	metaStore *repositories.MetaStore
}

func NewTagService(metaStore *repositories.MetaStore) *TagService {
	return &TagService{metaStore: metaStore}
}

func (s *TagService) CreateOrGetTag(ctx context.Context, tag string, creator string) (*repositories.Tag, error) {
	// 先尝试获取现有标签
	existingTag, err := s.metaStore.MomentRepo.GetTagByTag(tag)
	if err == nil {
		return existingTag, nil
	}

	// 如果标签不存在，创建新标签
	newTag := &repositories.Tag{
		ID:        helper.GenerateTID(),
		Tag:       tag,
		CreatedAt: time.Now().Unix(),
		Creator:   creator,
		Deleted:   false,
	}

	if err := s.metaStore.MomentRepo.CreateTag(newTag); err != nil {
		return nil, fmt.Errorf("创建标签失败: %w", err)
	}

	return newTag, nil
}

func (s *TagService) CreateActivityTag(ctx context.Context, uri string, tag string, creator string) (*repositories.ActivityTag, error) {
	activityTag := &repositories.ActivityTag{
		ID:         helper.GenerateTID(),
		SubjectURI: uri,
		Tag:        tag,
		CreatedAt:  time.Now().Unix(),
		Creator:    creator,
		Deleted:    false,
	}

	if err := s.metaStore.MomentRepo.CreateActivityTag(activityTag); err != nil {
		return nil, fmt.Errorf("创建活动标签关联失败: %w", err)
	}

	return activityTag, nil
}

func (s *TagService) ListTags(ctx context.Context, page int, pageSize int) ([]*repositories.Tag, error) {
	return s.metaStore.MomentRepo.ListTags(page, pageSize)
}

func (s *TagService) SyncActivityTags(ctx context.Context, subjectURI string, tags []string, creator string) ([]*repositories.ActivityTag, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	ret := make([]*repositories.ActivityTag, 0, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		// 确保标签存在于标签池中
		_, err := s.CreateOrGetTag(ctx, tag, creator)
		if err != nil {
			return nil, fmt.Errorf("处理标签 '%s' 失败: %w", tag, err)
		}

		// 创建活动标签关联
		activityTag, err := s.CreateActivityTag(ctx, subjectURI, tag, creator)
		if err != nil {
			return nil, fmt.Errorf("创建标签关联 '%s' 失败: %w", tag, err)
		}
		ret = append(ret, activityTag)
	}

	return ret, nil
}

func (s *TagService) UnbindActivityTags(ctx context.Context, subjectURI string) error {
	// 获取该 subject 的所有标签关联
	activityTags, err := s.metaStore.MomentRepo.GetActivityTagsBySubjectURI(subjectURI)
	if err != nil {
		return fmt.Errorf("获取标签关联失败: %w", err)
	}

	// 删除所有标签关联
	for _, activityTag := range activityTags {
		if err := s.metaStore.MomentRepo.DeleteActivityTag(activityTag.Tag, subjectURI); err != nil {
			return fmt.Errorf("删除标签关联失败: %w", err)
		}
	}

	return nil
}

func (s *TagService) GetTagUsageCount(ctx context.Context, tag string) (int, error) {
	activityTags, err := s.metaStore.MomentRepo.GetActivityTagsByTag(tag)
	if err != nil {
		return 0, fmt.Errorf("获取标签使用情况失败: %w", err)
	}
	return len(activityTags), nil
}

func (s *TagService) PresentTagView(tags []*repositories.Tag) (map[string]*types.TagView, error) {
	ret := make(map[string]*types.TagView)

	tagList := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, tag.Tag)
	}
	creatorList := make([]string, 0, len(tags))
	for _, tag := range tags {
		creatorList = append(creatorList, tag.Creator)
	}

	tagCounts, err := s.metaStore.MomentRepo.GetTagActivityCounts(tagList)
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

	for _, tag := range tags {
		creator, ok := creatorMap[tag.Creator]
		if !ok {
			continue
		}

		ret[tag.Tag] = &types.TagView{
			Tag:     tag.Tag,
			Count:   tagCounts[tag.Tag],
			Creator: tag.Creator,
			IsAster: creator.IsAster,
		}
	}
	return ret, nil
}

func (s *TagService) PresentActivityTagView(tags []*repositories.ActivityTag) (map[string]*types.TagView, error) {
	ret := make(map[string]*types.TagView)

	tagList := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, tag.Tag)
	}
	creatorList := make([]string, 0, len(tags))
	for _, tag := range tags {
		creatorList = append(creatorList, tag.Creator)
	}

	tagCounts, err := s.metaStore.MomentRepo.GetTagActivityCounts(tagList)
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

	for _, tag := range tags {
		creator, ok := creatorMap[tag.Creator]
		if !ok {
			continue
		}

		ret[tag.Tag] = &types.TagView{
			Tag:     tag.Tag,
			Count:   tagCounts[tag.Tag],
			Creator: tag.Creator,
			IsAster: creator.IsAster,
		}
	}
	return ret, nil
}
