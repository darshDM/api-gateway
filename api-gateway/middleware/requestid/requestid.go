package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type RequestIDKey string

const RequestIDKeyName RequestIDKey = "RequestID"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKeyName, id)
		w.Header().Set("x-Request-ID", id)
		r.Header.Set("x-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKeyName).(string); ok {
		return id
	}
	return ""
}
