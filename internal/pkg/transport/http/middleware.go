package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/logger"
)

type MiddlewareFunc func(http.Handler) http.Handler

func Recoverer(logger *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if err, _ := rvr.(error); errors.Is(err, http.ErrAbortHandler) {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					logger.ErrorContext(req.Context(), "panic occurred", slog.Any("message", rvr), slog.String("stack_trace", string(debug.Stack())))
					respWriter.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(respWriter, req)
		})
	}
}

// CORSMiddleware set CORS related headers.
func CORSMiddleware() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:8444"}, // allow swagger
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Origin", "Content-Type"},
	})
}

// RequestID add request id to context and response header.
func RequestID() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
