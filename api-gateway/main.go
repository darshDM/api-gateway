package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/DarshDM/api-gateway/internal/config"
	"github.com/gorilla/mux"
)

type Gateway struct {
	servers []config.Server
	logger  *log.Logger
}

func CreateHandler(server *config.Server, logger *log.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		proxyReq, err := http.NewRequest(req.Method, server.Host, bytes.NewReader(body))
		if err != nil {
			http.Error(res, "There are issues with request url", http.StatusBadRequest)
		}
		proxyReq.Header = make(http.Header)
		for h, val := range req.Header {
			proxyReq.Header[h] = val
		}

		client := &http.Client{}
		response, err := client.Do(proxyReq)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		original_header := res.Header()
		for h, val := range response.Header {
			original_header[h] = val
		}
		io.Copy(res, response.Body)
		response.Body.Close()
	}
}
func (g *Gateway) AssignHandlers() *mux.Router {
	router := mux.NewRouter()
	for _, server := range g.servers {
		g.logger.Printf("Registering handler for prefix: %s -> %s", server.Prefix, server.Host)
		router.PathPrefix(server.Prefix).Handler(CreateHandler(&server, g.logger)).Methods("GET", "POST", "PUT", "DELETE", "PATCH")
	}
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.logger.Printf("Catch-all handler hit: %s %s", r.Method, r.URL.Path)
		http.Error(w, "No matching route", http.StatusNotFound)
	})
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
	l.Fatal(http.ListenAndServe(":8000", nil))

}
