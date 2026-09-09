package auth

import (
	"log"
	"net/http"

	"zorion/internal/config"
)

// AdminAuth проверяет пароль в заголовке X-Admin-Password
func AdminAuth(next http.HandlerFunc) http.HandlerFunc {
	cfg := config.Load()
	return func(w http.ResponseWriter, r *http.Request) {
		password := r.Header.Get("X-Admin-Password")
		if password == "" {
			// Если пароль не передан, но запрос из браузера, можно показать страницу с формой
			// Но для простоты просто возвращаем 401
			http.Error(w, "Admin password required", http.StatusUnauthorized)
			return
		}
		if password != cfg.AdminPassword {
			log.Printf("Admin auth failed: wrong password")
			http.Error(w, "Invalid admin password", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}