package main

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"net/http"

	log "github.com/sirupsen/logrus"

	reqLog "github.com/DarshDM/api-gateway/utils/log"

	"github.com/DarshDM/api-gateway/internal/config"
	"github.com/DarshDM/api-gateway/middleware/auth"
	"github.com/DarshDM/api-gateway/middleware/ratelimit"
	"github.com/DarshDM/api-gateway/middleware/requestid"
	gatewayError "github.com/DarshDM/api-gateway/utils/error"
	"github.com/gorilla/mux"
)

type Gateway struct {
	servers   []config.Server
	logger    *log.Logger
	hostIndex sync.Map
}

// Add Logic for basic load balancer
func (g *Gateway) getNextHost(server *config.Server, hostIndex *sync.Map) string {
	key := server.Name
	value, _ := hostIndex.Load(key)
	currentIndex := 0
	if value != nil {
		currentIndex = value.(int)
	}
	nextIndex := (currentIndex + 1) % len(server.Hosts)
	hostIndex.Store(key, nextIndex)
	return server.Hosts[nextIndex]
}

func (g *Gateway) proxyRequest(server *config.Server, req *http.Request) (*http.Response, error) {
	host := g.getNextHost(server, &g.hostIndex)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	proxyReq, err := http.NewRequest(req.Method, host, bytes.NewReader(body))
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}
	proxyReq.Header.Set("x-forwarded-for", req.RemoteAddr)
	client := &http.Client{}
	return client.Do(proxyReq)

}

func Proxyresponse(res http.ResponseWriter, response *http.Response, logger *log.Logger, r *http.Request) {
	requestLogger := reqLog.RequestLogger(logger, r)
	requestLogger.Infof("Service response: %v", response.Status)
	original_header := res.Header()
	for h, val := range response.Header {
		original_header[h] = val
	}
	io.Copy(res, response.Body)
	response.Body.Close()
}
func (g *Gateway) CreateHandler(server *config.Server, logger *log.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		requestLogger := reqLog.RequestLogger(logger, req)
		requestLogger.Infof("Request received")
		response, err := g.proxyRequest(server, req)

		if err != nil {
			requestLogger.Errorf("Service is down: %s", err.Error())
			user_error := fmt.Sprintf("%v service is unavailable", server.Name)
			http.Error(res, user_error, http.StatusBadRequest)
			return
		}
		Proxyresponse(res, response, logger, req)
	}
}

func CreateNoMatchHandler(logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := requestid.GetRequestID(r.Context())
		logger.Printf("[%s] New Request %s %s: %s", requestId, r.RemoteAddr, r.Method, r.URL)
		logger.Printf("[%s] No services registered for: %s %s", requestId, r.Method, r.URL.Path)
		err := &gatewayError.GatewayError{
			Service: r.URL.Path,
			Message: "Service not found",
			Code:    http.StatusBadGateway,
		}
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
func (g *Gateway) AssignHandlers() *mux.Router {
	router := mux.NewRouter()
	rateLimiter := ratelimit.NewRateLimiter(g.servers, g.logger)
	for _, server := range g.servers {
		g.logger.WithFields(log.Fields{
			"service": "api-gateway",
		}).Infof("🚀Registering handler for prefix: %s -> %s", server.Prefix, server.Hosts[0])
		handler := g.CreateHandler(&server, g.logger)
		authHandler := auth.AuthMiddleware(&server, g.logger, handler)
		requestIDHandler := requestid.RequestIDMiddleware(authHandler)

		rateLimitHandler := rateLimiter.RateLimitMiddleware(requestIDHandler)
		router.Path(server.Prefix).Handler(rateLimitHandler).Methods("GET", "POST", "PUT", "DELETE", "PATCH")
		router.PathPrefix(server.Prefix+"/").Handler(requestIDHandler).Methods("GET", "POST", "PUT", "DELETE", "PATCH")
		g.logger.WithFields(log.Fields{
			"service": "api-gateway",
		}).Infof("✅Registed handler for prefix: %s -> %s", server.Prefix, server.Hosts[0])
	}
	handler := CreateNoMatchHandler(g.logger)
	router.PathPrefix("/").Handler(requestid.RequestIDMiddleware(handler))
	return router

}
func NewGateway(servers []config.Server, l *log.Logger) (*Gateway, error) {
	return &Gateway{
			servers:   servers,
			logger:    l,
			hostIndex: sync.Map{},
		},
		nil
}
func main() {

	l := log.New()
	l.SetFormatter(&log.JSONFormatter{})
	cfg, err := config.Load(".", l)
	if err != nil {
		l.Fatal(err)
	}
	gateway, err := NewGateway(cfg.Servers, l)
	if err != nil {
		l.Fatal("Error while creating Gateway instance.")
	}
	router := gateway.AssignHandlers()

	// lmt := tollbooth.NewLimiter(2, nil)
	// lmt.SetIPLookup(limiter.IPLookup{
	// 	Name: "RemoteAddr",
	// })
	http.Handle("/", router)
	l.WithFields(log.Fields{
		"service": "api-gateway",
		"port":    8001,
	}).Info("API-gateway running on port 8001")
	l.Fatal(http.ListenAndServe(":8001", nil))

}
