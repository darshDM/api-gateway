package main

import (
	"bytes"
	"fmt"
	"io"

	"net/http"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	log "github.com/sirupsen/logrus"

	"github.com/DarshDM/api-gateway/internal/config"
	"github.com/DarshDM/api-gateway/middleware/auth"
	"github.com/DarshDM/api-gateway/middleware/requestid"
	gatewayError "github.com/DarshDM/api-gateway/utils/error"
	"github.com/gorilla/mux"
)

type Gateway struct {
	servers []config.Server
	logger  *log.Logger
}

func CreateHandler(server *config.Server, logger *log.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestId := requestid.GetRequestID(req.Context())
		// logger.Printf("[%s] %s %s: %s", requestId, req.RemoteAddr, req.Method, req.URL)
		logger.WithFields(log.Fields{
			"service":   "api-gateway",
			"requestId": requestId,
			"IP":        req.RemoteAddr,
			"Method":    req.Method,
			"URL":       req.URL.Path,
		}).Info("New Request")

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		proxyReq, err := http.NewRequest(req.Method, server.Host, bytes.NewReader(body))
		if err != nil {
			http.Error(res, "invalid request Check request parameters and body.", http.StatusBadRequest)
			return
		}
		proxyReq.Header = make(http.Header)
		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}
		logger.WithFields(log.Fields{
			"service":   "api-gateway",
			"requestId": requestId,
			"IP":        req.RemoteAddr,
			"Method":    req.Method,
			"URL":       req.URL.Path,
		}).Info("Calling service")
		logger.WithFields(log.Fields{
			"service":   "api-gateway",
			"requestId": requestId,
			"IP":        req.RemoteAddr,
			"Method":    req.Method,
			"URL":       req.URL.Path,
		}).Infof("Proxy call to %v", server.Host)
		proxyReq.Header.Set("x-forwarded-for", req.RemoteAddr)
		client := &http.Client{}
		response, err := client.Do(proxyReq)

		if err != nil {
			logger.WithFields(log.Fields{
				"service":   server.Name,
				"requestId": requestId,
				"IP":        req.RemoteAddr,
				"Method":    req.Method,
				"URL":       req.URL.Path,
			}).Errorf("Service is down: %s", err.Error())
			user_error := fmt.Sprintf("%v service is unavailable", server.Name)
			http.Error(res, user_error, http.StatusBadRequest)
			return
		}
		logger.WithFields(log.Fields{
			"service":   server.Name,
			"requestId": requestId,
			"IP":        req.RemoteAddr,
			"Method":    req.Method,
			"URL":       req.URL.Path,
		}).Infof("Service response: %v", response.Status)
		original_header := res.Header()
		for h, val := range response.Header {
			original_header[h] = val
		}
		io.Copy(res, response.Body)
		response.Body.Close()
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

	for _, server := range g.servers {
		g.logger.WithFields(log.Fields{
			"service": "api-gateway",
		}).Infof("🚀Registering handler for prefix: %s -> %s", server.Prefix, server.Host)
		handler := CreateHandler(&server, g.logger)
		authHandler := auth.AuthMiddleware(&server, g.logger, handler)
		requestIDHandler := requestid.RequestIDMiddleware(authHandler)
		router.PathPrefix(server.Prefix+"/").Handler(requestIDHandler).Methods("GET", "POST", "PUT", "DELETE", "PATCH")
		g.logger.WithFields(log.Fields{
			"service": "api-gateway",
		}).Infof("✅Registed handler for prefix: %s -> %s", server.Prefix, server.Host)
	}
	handler := CreateNoMatchHandler(g.logger)
	router.PathPrefix("/").Handler(requestid.RequestIDMiddleware(handler))
	return router

}
func NewGateway(servers []config.Server, l *log.Logger) (*Gateway, error) {
	return &Gateway{servers: servers, logger: l}, nil
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

	lmt := tollbooth.NewLimiter(2, nil)
	lmt.SetIPLookup(limiter.IPLookup{
		Name: "RemoteAddr",
	})
	http.Handle("/", tollbooth.HTTPMiddleware(lmt)(router))
	l.WithFields(log.Fields{
		"service": "api-gateway",
		"port":    8001,
	}).Info("API-gateway running on port 8001")
	l.Fatal(http.ListenAndServe(":8001", nil))

}
