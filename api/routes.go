package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

type response struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
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

	resp := response{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 201, resp)
}

func (a *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	apiKey, found := strings.CutPrefix(r.Header.Get("Authorization"), "ApiKey ")
	if !found {
		respondWithError(w, 401, "no authorization header")
	}

	user, err := a.DB.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		log.Printf("user fetching error: %v\n", err)
		msg := "something went wrong"
		code := 500
		if errors.Is(err, sql.ErrNoRows) {
			msg = "user not found"
			code = 404
		}
		respondWithError(w, code, msg)
		return
	}

	resp := response{
		Id:        user.ID.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
	respondWithJSON(w, 200, resp)
}
