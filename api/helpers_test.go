package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/internal/database"
	"github.com/stretchr/testify/mock"
)

func mustJSON(t *testing.T, v interface{}) string {
	t.Helper()
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

type MockedDbApi struct {
	mock.Mock
}

func (m *MockedDbApi) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockedDbApi) GetUserByApiKey(ctx context.Context, apiKey string) (database.User, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(database.User), args.Error(1)
}

func (m *MockedDbApi) CreateFeed(ctx context.Context, arg database.CreateFeedParams) (database.Feed, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Feed), args.Error(1)
}

func (m *MockedDbApi) ListFeeds(ctx context.Context) ([]database.Feed, error) {
	args := m.Called(ctx)
	return args.Get(0).([]database.Feed), args.Error(1)
}

func (m *MockedDbApi) GetFeed(ctx context.Context, id uuid.UUID) (database.Feed, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Feed), args.Error(1)
}

func (m *MockedDbApi) CreateFeedFollow(ctx context.Context, arg database.CreateFeedFollowParams) (database.FeedFollow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.FeedFollow), args.Error(1)
}

func (m *MockedDbApi) GetFeedFollow(ctx context.Context, id uuid.UUID) (database.FeedFollow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.FeedFollow), args.Error(1)
}

func (m *MockedDbApi) ListUserFeedFollows(ctx context.Context, userID uuid.UUID) ([]database.FeedFollow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]database.FeedFollow), args.Error(1)
}

func (m *MockedDbApi) DeleteFeedFollow(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
