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

func getUser(w http.ResponseWriter, r *http.Request, db DbApi, apiKey string) (*database.User, error) {
	user, err := db.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		log.Printf("user fetching error: %v\n", err)
		msg := "something went wrong"
		code := 500
		if errors.Is(err, sql.ErrNoRows) {
			msg = "user not found"
			code = 404
		}
		respondWithError(w, code, msg)
		return nil, err
	}
	return &user, nil
}

func (a *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Context().Value(middleware.AuthApiKey).(string)
	user, err := getUser(w, r, a.DB, apiKey)
	if err != nil {
		return
	}

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
	apiKey := r.Context().Value(middleware.AuthApiKey).(string)
	user, err := getUser(w, r, a.DB, apiKey)
	if err != nil {
		return
	}

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

	resp := feedResponse{
		Id:        feed.ID.String(),
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
		Name:      feed.Name,
		Url:       feed.Url,
		UserId:    feed.UserID.String(),
	}
	respondWithJSON(w, 201, resp)
}
