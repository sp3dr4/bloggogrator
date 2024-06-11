package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/api/middleware"
	"github.com/sp3dr4/bloggogrator/internal/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var now = time.Now()

func setupUser() database.User {
	return database.User{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "FooBar",
		ApiKey:    "qwerty-12345-asdf",
	}
}

func compareUser(t *testing.T, expected database.User, actual userResponse) {
	t.Helper()
	require.Equal(t, expected.ID.String(), actual.Id)
	require.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Duration(1))
	require.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Duration(1))
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.ApiKey, actual.ApiKey)
}

func compareError(t *testing.T, rw *httptest.ResponseRecorder, expectedCode int, expectedMsg string) {
	require.Equal(t, expectedCode, rw.Code)
	var resp map[string]string
	err := json.NewDecoder(rw.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, expectedMsg, resp["error"])
}

func setupCreateUserTest(t *testing.T, mockDbApi *MockedDbApi, user database.User, err error) (*httptest.ResponseRecorder, *http.Request, apiConfig) {
	t.Helper()
	testApi := apiConfig{DB: mockDbApi}
	mockDbApi.On("CreateUser", mock.Anything, mock.Anything).Return(user, err)

	body, err := json.Marshal(map[string]string{"name": "FooBar"})
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
	require.NoError(t, err)
	rw := httptest.NewRecorder()

	return rw, req, testApi
}

func TestCreateUserHandler(t *testing.T) {
	t.Run("return 201", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		user := setupUser()
		rw, req, testApi := setupCreateUserTest(t, mockDbApi, user, nil)

		testApi.handlerCreateUser(rw, req)

		require.Equal(t, http.StatusCreated, rw.Code)
		var resp userResponse
		err := json.NewDecoder(rw.Body).Decode(&resp)
		require.NoError(t, err)
		compareUser(t, user, resp)

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 500", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		rw, req, testApi := setupCreateUserTest(t, mockDbApi, database.User{}, context.DeadlineExceeded)

		testApi.handlerCreateUser(rw, req)

		compareError(t, rw, http.StatusInternalServerError, "error creating user")

		mockDbApi.AssertExpectations(t)
	})
}

func setupGetUserTest(t *testing.T, mockDbApi *MockedDbApi, user database.User, err error) (*httptest.ResponseRecorder, *http.Request, apiConfig) {
	t.Helper()
	testApi := apiConfig{DB: mockDbApi}
	mockDbApi.On("GetUserByApiKey", mock.Anything, mock.Anything).Return(user, err)

	req, err := http.NewRequest(http.MethodGet, "/v1/users", nil)
	require.NoError(t, err)
	key := "anything"
	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", key))
	ctx := context.WithValue(req.Context(), middleware.AuthApiKey, key)
	rw := httptest.NewRecorder()

	return rw, req.WithContext(ctx), testApi
}

func TestGetUserHandler(t *testing.T) {
	t.Run("return 200", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		user := setupUser()
		rw, req, testApi := setupGetUserTest(t, mockDbApi, user, nil)

		testApi.handlerGetUser(rw, req)

		require.Equal(t, http.StatusOK, rw.Code)
		var resp userResponse
		err := json.NewDecoder(rw.Body).Decode(&resp)
		require.NoError(t, err)
		compareUser(t, user, resp)

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 404", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		rw, req, testApi := setupGetUserTest(t, mockDbApi, database.User{}, sql.ErrNoRows)

		testApi.handlerGetUser(rw, req)

		compareError(t, rw, http.StatusNotFound, "user not found")

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 500", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		rw, req, testApi := setupGetUserTest(t, mockDbApi, database.User{}, context.DeadlineExceeded)

		testApi.handlerGetUser(rw, req)

		compareError(t, rw, http.StatusInternalServerError, "something went wrong")

		mockDbApi.AssertExpectations(t)
	})
}

func setupFeed() database.Feed {
	return database.Feed{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "TestFeed",
		Url:       "http://example.com",
		UserID:    uuid.New(),
	}
}

func compareFeed(t *testing.T, expected database.Feed, actual feedResponse) {
	t.Helper()
	require.Equal(t, expected.ID.String(), actual.Id)
	require.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Duration(1))
	require.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Duration(1))
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Url, actual.Url)
	require.Equal(t, expected.UserID.String(), actual.UserId)
}

func setupCreateFeedTest(t *testing.T, mockDbApi *MockedDbApi, user database.User, feed database.Feed, err error) (*httptest.ResponseRecorder, *http.Request, apiConfig) {
	t.Helper()
	testApi := apiConfig{DB: mockDbApi}
	mockDbApi.On("GetUserByApiKey", mock.Anything, mock.Anything).Return(user, nil)
	mockDbApi.On("CreateFeed", mock.Anything, mock.Anything).Return(feed, err)

	body, err := json.Marshal(map[string]string{"name": "TestFeed", "url": "http://example.com"})
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/feed", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", user.ApiKey))
	ctx := context.WithValue(req.Context(), middleware.AuthApiKey, user.ApiKey)
	rw := httptest.NewRecorder()

	return rw, req.WithContext(ctx), testApi
}

func TestCreateFeedHandler(t *testing.T) {
	t.Run("return 201", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		user := setupUser()
		feed := setupFeed()
		rw, req, testApi := setupCreateFeedTest(t, mockDbApi, user, feed, nil)

		testApi.handlerCreateFeed(rw, req)

		require.Equal(t, http.StatusCreated, rw.Code)
		var resp feedResponse
		err := json.NewDecoder(rw.Body).Decode(&resp)
		require.NoError(t, err)
		compareFeed(t, feed, resp)

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 400", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		user := setupUser()
		mockDbApi.On("GetUserByApiKey", mock.Anything, mock.Anything).Return(user, nil)
		testApi := apiConfig{DB: mockDbApi}
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/feed", bytes.NewBufferString(`{invalid json}`))
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("ApiKey %s", user.ApiKey))
		ctx := context.WithValue(req.Context(), middleware.AuthApiKey, user.ApiKey)

		testApi.handlerCreateFeed(rw, req.WithContext(ctx))

		compareError(t, rw, http.StatusBadRequest, "error decoding request body")

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 500", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		user := setupUser()
		rw, req, testApi := setupCreateFeedTest(t, mockDbApi, user, database.Feed{}, context.DeadlineExceeded)

		testApi.handlerCreateFeed(rw, req)

		compareError(t, rw, http.StatusInternalServerError, "error creating feed")

		mockDbApi.AssertExpectations(t)
	})
}
