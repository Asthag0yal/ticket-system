package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"ticket-system/internal/auth"
	"ticket-system/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserContext struct {
	UserID string
	Email  string
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(r *http.Request) (UserContext, bool) {
	user, ok := r.Context().Value(UserContextKey).(UserContext)
	return user, ok
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse{Error: message})
}
