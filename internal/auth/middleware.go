package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware проверяет JWT и добавляет userID в контекст запроса
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Auth: no Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Auth: invalid header format: %s", authHeader)
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		userID, err := VerifyToken(tokenString)
		if err != nil {
			log.Printf("Auth: token verification failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		log.Printf("Auth: user %s authenticated", userID)
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}