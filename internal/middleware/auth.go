package middleware

import (
	"context"
	"net/http"
	"strings"

	"tracker/internal/service"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			userID, err := authSvc.ParseAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDKey, userID)))
		})
	}
}