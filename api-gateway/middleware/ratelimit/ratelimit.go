package ratelimit

import (
	"net/http"
	"time"

	"github.com/DarshDM/api-gateway/internal/config"
	gatewayError "github.com/DarshDM/api-gateway/utils/error"
	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	log "github.com/sirupsen/logrus"
)

type RateLimiter struct {
	ServiceLimiter map[string]*limiter.Limiter
	Servers        []config.Server
	Logger         *log.Logger
}

func NewRateLimiter(servers []config.Server, logger *log.Logger) *RateLimiter {
	rl := &RateLimiter{
		Servers:        servers,
		Logger:         logger,
		ServiceLimiter: make(map[string]*limiter.Limiter),
	}

	for _, server := range servers {
		if server.RateLimit > 0 {
			lmt := tollbooth.NewLimiter(float64(server.RateLimit), &limiter.ExpirableOptions{
				DefaultExpirationTTL: time.Second,
			})
			lmt.SetBurst(5)
			rl.ServiceLimiter[server.Name] = lmt
		}
	}
	return rl
}
func (rl *RateLimiter) getServer(r *http.Request) *config.Server {
	for _, server := range rl.Servers {

		if server.Name == r.URL.Path[1:len(server.Name)+1] {
			return &server
		}
	}
	return nil
}
func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := rl.getServer(r)

		httpError := tollbooth.LimitByKeys(rl.ServiceLimiter[server.Name], []string{server.Name})
		if httpError != nil {
			rl.Logger.Warnf("Rate limit exceeded for %s", r.RemoteAddr)
			err := &gatewayError.GatewayError{
				Service: r.URL.Path,
				Message: "Rate limit exceeded",
				Code:    http.StatusTooManyRequests,
			}
			http.Error(w, err.Error(), err.Code)
			return
		}

		next.ServeHTTP(w, r)
	})
}
