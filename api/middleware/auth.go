package middleware

import (
	"context"
	"net/http"
	"strings"
)

type MiddlewareContextKey string

const AuthUser MiddlewareContextKey = "middleware.auth.user"

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}

type UserFetcher func(ctx context.Context, apikey string) (interface{}, error)

func AuthFactory(fetchUser UserFetcher) Middleware {
	return func(next http.Handler) http.Handler {
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

			user, err := fetchUser(r.Context(), key)
			if err != nil {
				writeUnauthed(w)
				return
			}

			ctx := context.WithValue(r.Context(), AuthUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
