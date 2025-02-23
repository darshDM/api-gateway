package auth

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/DarshDM/api-gateway/internal/config"
)

func AuthMiddleware(server *config.Server, logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if server.ApiKey == "" || server.ApiKey == r.Header.Get("x-api-key") {
			next.ServeHTTP(w, r)
		} else {
			logger.WithFields(log.Fields{
				"service": "api-gateway",
				"IP":      r.RemoteAddr,
				"Method":  r.Method,
				"URL":     r.URL.Path,
			}).Error("Unauthorized")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}
