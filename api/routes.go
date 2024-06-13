package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/api/middleware"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

type userResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

type feedResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	UserId    string    `json:"user_id"`
}

type followResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	FeedId    string    `json:"feed_id"`
	UserId    string    `json:"user_id"`
}

func (a *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}

	createParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      request.Name,
	}
	user, err := a.DB.CreateUser(r.Context(), createParams)
	if err != nil {
		log.Printf("user creation error: %v\n", err)
		respondWithError(w, 500, "error creating user")
		return
	}

	resp := userResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 201, resp)
}

func (a *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	resp := userResponse{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 200, resp)
}

func (a *apiConfig) handlerListFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := a.DB.ListFeeds(r.Context())
	if err != nil {
		log.Printf("feeds listing error: %v\n", err)
		respondWithError(w, 500, "error listing feeds")
		return
	}
	respFeeds := make([]feedResponse, 0, len(feeds))
	for _, o := range feeds {
		respFeeds = append(respFeeds, feedResponse{
			Id:        o.ID.String(),
			CreatedAt: o.CreatedAt,
			UpdatedAt: o.UpdatedAt,
			Name:      o.Name,
			Url:       o.Url,
			UserId:    o.UserID.String(),
		})
	}
	respondWithJSON(w, 200, respFeeds)
}

func (a *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	var request struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}

	createParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      request.Name,
		Url:       request.Url,
		UserID:    user.ID,
	}
	feed, err := a.DB.CreateFeed(r.Context(), createParams)
	if err != nil {
		log.Printf("feed creation error: %v\n", err)
		respondWithError(w, 500, "error creating feed")
		return
	}

	createFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	follow, err := a.DB.CreateFeedFollow(r.Context(), createFollowParams)
	if err != nil {
		log.Printf("follow creation error: %v\n", err)
		respondWithError(w, 500, "error creating feed follow")
		return
	}

	feedResp := feedResponse{
		Id:        feed.ID.String(),
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
		Name:      feed.Name,
		Url:       feed.Url,
		UserId:    feed.UserID.String(),
	}
	followResp := followResponse{
		Id:        follow.ID.String(),
		CreatedAt: follow.CreatedAt,
		UserId:    follow.UserID.String(),
		FeedId:    follow.FeedID.String(),
	}
	resp := struct {
		Feed       feedResponse   `json:"feed"`
		FeedFollow followResponse `json:"feed_follow"`
	}{
		Feed:       feedResp,
		FeedFollow: followResp,
	}
	respondWithJSON(w, 201, resp)
}

func (a *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	var request struct {
		FeedId string `json:"feed_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, 400, "error decoding request body")
		return
	}

	feed_uuid, err := uuid.Parse(request.FeedId)
	if err != nil {
		respondWithError(w, 400, "invalid feed id")
		return
	}
	feed, err := a.DB.GetFeed(r.Context(), feed_uuid)
	if err != nil {
		respondWithError(w, 404, "feed not found")
		return
	}

	createParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	follow, err := a.DB.CreateFeedFollow(r.Context(), createParams)
	if err != nil {
		log.Printf("follow creation error: %v\n", err)
		respondWithError(w, 500, "error creating feed follow")
		return
	}

	resp := followResponse{
		Id:        follow.ID.String(),
		CreatedAt: follow.CreatedAt,
		FeedId:    follow.FeedID.String(),
		UserId:    follow.UserID.String(),
	}
	respondWithJSON(w, 201, resp)
}

func (a *apiConfig) handlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)
	followId, err := uuid.Parse(r.PathValue("feedFollowID"))
	if err != nil {
		respondWithError(w, 400, "invalid feed id")
		return
	}

	follow, err := a.DB.GetFeedFollow(r.Context(), followId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, 500, "error retrieving feed follow")
		return
	}
	if follow.UserID != user.ID {
		respondWithError(w, 403, "operation not allowed")
		return
	}

	if err := a.DB.DeleteFeedFollow(r.Context(), followId); err != nil {
		respondWithError(w, 500, "error deleting feed follow")
		return
	}

	respondWithJSON(w, 204, struct{}{})
}

func (a *apiConfig) handlerListUserFeedFollows(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	follows, err := a.DB.ListUserFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 500, "error retrieving feed follows")
		return
	}

	respFollows := make([]followResponse, 0, len(follows))
	for _, o := range follows {
		respFollows = append(respFollows, followResponse{
			Id:        o.ID.String(),
			CreatedAt: o.CreatedAt,
			FeedId:    o.FeedID.String(),
			UserId:    o.UserID.String(),
		})
	}
	respondWithJSON(w, 200, respFollows)
}
