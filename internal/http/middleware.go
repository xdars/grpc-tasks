package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/xdars/grpc-tasks/internal/grpc/interceptors"
	"github.com/xdars/grpc-tasks/internal/jwt"
)

func getToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getToken(r)

		claims, err := jwt.Validate(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), interceptors.UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
