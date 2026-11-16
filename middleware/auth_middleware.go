package middleware

import (
	"context"
	"net/http"
	"strings"

	"study-planner/utils"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserID(r *http.Request) uint {
	userID, _ := r.Context().Value(userIDKey).(uint)
	return userID
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeUnauthorized(w)
			return
		}

		userID, err := utils.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeUnauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	utils.JSONError(w, http.StatusUnauthorized, "Missing or invalid token")
}
