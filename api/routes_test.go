package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/internal/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUserHandler(t *testing.T) {
	t.Run("return 201", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		testApi := apiConfig{DB: mockDbApi}
		userID := uuid.New()
		nowTime := time.Now()
		mockDbApi.On("CreateUser", mock.Anything, mock.Anything).Return(database.User{
			ID:        userID,
			CreatedAt: nowTime,
			UpdatedAt: nowTime,
			Name:      "FooBar",
		}, nil)

		body, _ := json.Marshal(map[string]string{"name": "FooBar"})
		req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
		require.NoError(t, err)
		rw := httptest.NewRecorder()

		testApi.handlerCreateUser(rw, req)

		require.Equal(t, http.StatusCreated, rw.Code)
		var resp response
		err = json.NewDecoder(rw.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, userID.String(), resp.Id)
		require.WithinDuration(t, nowTime, resp.CreatedAt, time.Duration(1))
		require.WithinDuration(t, nowTime, resp.UpdatedAt, time.Duration(1))
		require.Equal(t, "FooBar", resp.Name)

		mockDbApi.AssertExpectations(t)
	})

	t.Run("return 500", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		testApi := apiConfig{DB: mockDbApi}
		mockDbApi.On("CreateUser", mock.Anything, mock.Anything).Return(database.User{}, context.DeadlineExceeded)

		body, _ := json.Marshal(map[string]string{"name": "FooBar"})
		req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(body))
		require.NoError(t, err)
		rw := httptest.NewRecorder()

		testApi.handlerCreateUser(rw, req)

		require.Equal(t, http.StatusInternalServerError, rw.Code)
		var resp map[string]string
		err = json.NewDecoder(rw.Body).Decode(&resp)
		require.NoError(t, err)
		require.Equal(t, "error creating user", resp["error"])

		mockDbApi.AssertExpectations(t)
	})
}
