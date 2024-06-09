package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type apiConfig struct {
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
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/healthz", cfg.handlerHealth)
	mux.HandleFunc("GET /v1/err", cfg.handlerErr)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
