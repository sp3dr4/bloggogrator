package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
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

func dbUserToUser(o database.User) userResponse {
	return userResponse{
		Id:        o.ID.String(),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Name:      o.Name,
		ApiKey:    o.ApiKey,
	}
}

type feedResponse struct {
	Id            string     `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `json:"name"`
	Url           string     `json:"url"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	UserId        string     `json:"user_id"`
}

func dbFeedToFeed(o database.Feed) feedResponse {
	var fetchedAt *time.Time
	if o.LastFetchedAt.Valid {
		fetchedAt = &o.LastFetchedAt.Time
	}
	return feedResponse{
		Id:            o.ID.String(),
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
		Name:          o.Name,
		Url:           o.Url,
		LastFetchedAt: fetchedAt,
		UserId:        o.UserID.String(),
	}
}

type followResponse struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	FeedId    string    `json:"feed_id"`
	UserId    string    `json:"user_id"`
}

func dbFollowToFollow(o database.FeedFollow) followResponse {
	return followResponse{
		Id:        o.ID.String(),
		CreatedAt: o.CreatedAt,
		UserId:    o.UserID.String(),
		FeedId:    o.FeedID.String(),
	}
}

type postResponse struct {
	Id          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Url         string    `json:"url"`
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	PublishedAt time.Time `json:"published_at"`
	FeedId      string    `json:"feed_id"`
}

func dbPostToPost(o database.Post) postResponse {
	var title *string
	if o.Title.Valid {
		title = &o.Title.String
	}
	var descr *string
	if o.Description.Valid {
		descr = &o.Description.String
	}
	return postResponse{
		Id:          o.ID.String(),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
		Url:         o.Url,
		Title:       title,
		Description: descr,
		PublishedAt: o.PublishedAt,
		FeedId:      o.FeedID.String(),
	}
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

	respondWithJSON(w, 201, dbUserToUser(user))
}

func (a *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	respondWithJSON(w, 200, dbUserToUser(user))
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
		respFeeds = append(respFeeds, dbFeedToFeed(o))
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

	resp := struct {
		Feed       feedResponse   `json:"feed"`
		FeedFollow followResponse `json:"feed_follow"`
	}{
		Feed:       dbFeedToFeed(feed),
		FeedFollow: dbFollowToFollow(follow),
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

	respondWithJSON(w, 201, dbFollowToFollow(follow))
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
		respFollows = append(respFollows, dbFollowToFollow(o))
	}
	respondWithJSON(w, 200, respFollows)
}

func (a *apiConfig) handlerListPosts(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.AuthUser).(database.User)

	var limit int32 = 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limitInt, err := strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			respondWithError(w, 400, "invalid limit query parameter")
			return
		}
		limit = int32(limitInt)
	}

	params := database.GetUserPostsParams{UserID: user.ID, Limit: limit}
	posts, err := a.DB.GetUserPosts(r.Context(), params)
	if err != nil {
		respondWithError(w, 500, "error retrieving user posts")
		return
	}

	respPosts := make([]postResponse, 0, len(posts))
	for _, o := range posts {
		respPosts = append(respPosts, dbPostToPost(o))
	}
	respondWithJSON(w, 200, respPosts)
}
