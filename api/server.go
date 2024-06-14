package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sp3dr4/bloggogrator/api/middleware"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

type DbApi interface {
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
	GetUserByApiKey(context.Context, string) (database.User, error)
	CreateFeed(context.Context, database.CreateFeedParams) (database.Feed, error)
	ListFeeds(context.Context) ([]database.Feed, error)
	GetFeed(context.Context, uuid.UUID) (database.Feed, error)
	CreateFeedFollow(context.Context, database.CreateFeedFollowParams) (database.FeedFollow, error)
	GetFeedFollow(context.Context, uuid.UUID) (database.FeedFollow, error)
	ListUserFeedFollows(context.Context, uuid.UUID) ([]database.FeedFollow, error)
	DeleteFeedFollow(context.Context, uuid.UUID) error
	GetNextFeedsToFetch(context.Context, int32) ([]database.Feed, error)
	MarkFeedFetched(context.Context, database.MarkFeedFetchedParams) (database.Feed, error)
}

type apiConfig struct {
	DB DbApi
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type respErr struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, respErr{Error: msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, "error encoding response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func (a *apiConfig) handlerHealth(w http.ResponseWriter, r *http.Request) {
	type resp struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, resp{Status: "ok"})
}

func (a *apiConfig) handlerErr(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func Run(db DbApi) {
	cfg := apiConfig{
		DB: db,
	}

	userFetcher := func(ctx context.Context, apiKey string) (interface{}, error) {
		return cfg.DB.GetUserByApiKey(ctx, apiKey)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", cfg.handlerHealth)
	mux.HandleFunc("GET /v1/err", cfg.handlerErr)
	mux.HandleFunc("POST /v1/users", cfg.handlerCreateUser)
	mux.HandleFunc("GET /v1/feeds", cfg.handlerListFeeds)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /users", cfg.handlerGetUser)
	protectedMux.HandleFunc("POST /feeds", cfg.handlerCreateFeed)
	protectedMux.HandleFunc("POST /feed_follows", cfg.handlerCreateFeedFollow)
	protectedMux.HandleFunc("GET /feed_follows", cfg.handlerListUserFeedFollows)
	protectedMux.HandleFunc("DELETE /feed_follows/{feedFollowID}", cfg.handlerDeleteFeedFollow)
	protectedStack := middleware.CreateStack(middleware.AuthFactory(userFetcher))(protectedMux)
	mux.Handle("/v1/", http.StripPrefix("/v1", protectedStack))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: middleware.CreateStack(middleware.Logging)(mux),
	}
	log.Fatal(server.ListenAndServe())
}
