package log

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func RequestLogger(logger *log.Logger, r *http.Request) *log.Entry {
	return logger.WithFields(log.Fields{
		"service":   "api-gateway",
		"requestId": r.Header.Get("X-Request-Id"),
		"IP":        r.RemoteAddr,
		"Method":    r.Method,
		"URL":       r.URL.Path,
	})
}
