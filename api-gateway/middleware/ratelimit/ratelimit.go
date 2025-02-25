package ratelimit

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/DarshDM/api-gateway/internal/config"
)

type RateLimiter struct {
	servers         []config.Server
	logger          *log.Logger
	clientLimiters  sync.Map
	serviceLimiters sync.Map
}

func newRateLimiter(servers []config.Server, logger *log.Logger) *RateLimiter {
	// TODO: remaining to add per service client rate limiter
	r := &RateLimiter{
		servers: servers,
		logger:  logger,
	}
	fmt.Println(r)
	return r
}
