package middleware

import (
	"net/http"

	"github.com/pedy4000/noker/internal/api/response"
	"github.com/pedy4000/noker/pkg/config"
)

func Auth(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if key != cfg.Server.APIKey {
				response.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
