package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/DarshDM/api-gateway/internal/config"
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
		logger.Printf("[%s] %s %s: %s", requestId, req.RemoteAddr, req.Method, req.URL)

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
		logger.Printf("[%s] Proxy call to %v", requestId, server.Host)
		proxyReq.Header.Set("x-forwarded-for", req.RemoteAddr)
		client := &http.Client{}
		response, err := client.Do(proxyReq)

		if err != nil {
			logger.Printf("%v is down,%s", server.Name, err.Error())
			user_error := fmt.Sprintf("%v service is unavailable", server.Name)
			http.Error(res, user_error, http.StatusBadRequest)
			return
		}
		logger.Printf("[%s] Service response: %v", requestId, response.Status)
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
			Service: "gateway",
			Message: "Service not found",
			Code:    http.StatusBadGateway,
		}
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}
func (g *Gateway) AssignHandlers() *mux.Router {
	router := mux.NewRouter()

	for _, server := range g.servers {
		g.logger.Printf("🚀Registering handler for prefix: %s -> %s", server.Prefix, server.Host)
		handler := CreateHandler(&server, g.logger)
		router.PathPrefix(server.Prefix).Handler(requestid.RequestIDMiddleware(handler)).Methods("GET", "POST", "PUT", "DELETE", "PATCH")
		g.logger.Printf("✅Registed handler for prefix: %s -> %s", server.Prefix, server.Host)
	}
	handler := CreateNoMatchHandler(g.logger)
	router.PathPrefix("/").Handler(requestid.RequestIDMiddleware(handler))
	return router

}
func NewGateway(servers []config.Server, l *log.Logger) (*Gateway, error) {
	return &Gateway{servers: servers, logger: l}, nil
}
func main() {

	// Add logger
	l := log.New(os.Stdout, "api-gateway:", log.LstdFlags)

	cfg, err := config.Load(".", l)
	if err != nil {
		l.Fatal(err)
	}
	gateway, err := NewGateway(cfg.Servers, l)
	if err != nil {
		l.Fatal("Error while creating Gateway instance.")
	}
	router := gateway.AssignHandlers()
	http.Handle("/", router)
	l.Fatal(http.ListenAndServe(":8001", nil))

}
