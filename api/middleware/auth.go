package middleware

import (
	"context"
	"net/http"
	"strings"
)

type MiddlewareContextKey string

const AuthApiKey MiddlewareContextKey = "middleware.auth.apiKey"

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")

		if !strings.HasPrefix(authorization, "ApiKey ") {
			writeUnauthed(w)
			return
		}

		key := strings.TrimPrefix(authorization, "ApiKey ")
		if len(strings.TrimSpace(key)) < 1 {
			writeUnauthed(w)
			return
		}

		ctx := context.WithValue(r.Context(), AuthApiKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
