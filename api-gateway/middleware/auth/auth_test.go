package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DarshDM/api-gateway/internal/config"
	log "github.com/sirupsen/logrus"
)

func TestAuthMiddleware(t *testing.T) {
	logger := log.New()
	server := &config.Server{
		ApiKey: "secret_key",
	}

	//create new request
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("x-api-key", "secret_key")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

	})
	// Test valid request
	middleware := AuthMiddleware(server, logger, handler)
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("x-api-key", "wrong_key")
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

}
