package http

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/cors"
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

					logger.Error("panic occurred", slog.Any("message", rvr), slog.String("stack_trace", string(debug.Stack())))
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
