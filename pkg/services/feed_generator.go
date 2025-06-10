package services

import (
	"context"

	"github.com/zhongshangwu/avatarai-social/pkg/repositories"
)

type FeedGenerator struct {
	metaStore *repositories.MetaStore
}

func NewFeedGenerator(metaStore *repositories.MetaStore) *FeedGenerator {
	return &FeedGenerator{metaStore: metaStore}
}

func (f *FeedGenerator) GenerateFeed(ctx context.Context, limit int, cursor string) ([]string, string, error) {
	var uris []string
	var nextCursor string

	uris, err := f.metaStore.MomentRepo.GetLatestMomentURIs(limit, cursor)
	if err != nil {
		return nil, "", err
	}

	if len(uris) >= limit {
		nextCursor = uris[len(uris)-1]
	} else {
		nextCursor = ""
	}

	return uris, nextCursor, nil
}

type MockFeedGenerator struct {
	metaStore *repositories.MetaStore
}

func NewMockFeedGenerator(metaStore *repositories.MetaStore) *MockFeedGenerator {
	return &MockFeedGenerator{metaStore: metaStore}
}

func (f *MockFeedGenerator) GenerateFeed(ctx context.Context, limit int, cursor string) ([]string, string, error) {
	uris := []string{
		"at://did:plc:example1/app.vtri.activity.moment/1",
		"at://did:plc:example2/app.vtri.activity.moment/2",
		"at://did:plc:example3/app.vtri.activity.moment/3",
	}
	nextCursor := "mock_cursor_value"
	return uris, nextCursor, nil
}
