package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sp3dr4/bloggogrator/api/middleware"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

type DbApi interface {
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
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

func Run() {
	godotenv.Load()
	dbURL := os.Getenv("CONN")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("connection open error: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("connection ping error: %v", err)
	}

	dbQueries := database.New(db)

	cfg := apiConfig{
		DB: dbQueries,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", cfg.handlerHealth)
	mux.HandleFunc("GET /v1/err", cfg.handlerErr)

	mux.HandleFunc("POST /v1/users", cfg.handlerCreateUser)

	stack := middleware.CreateStack(
		middleware.Logging,
	)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: stack(mux),
	}
	log.Fatal(server.ListenAndServe())
}
