package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/exception"
)

// ResponseWithBody is the common method to encode all response types to the client.
func ResponseWithBody(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return fmt.Errorf("encode response body: %w", err)
	}

	return nil
}

func NoContentResponse(_ context.Context, w http.ResponseWriter, _ interface{}) error {
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func CreatedResponse(_ context.Context, w http.ResponseWriter, _ interface{}) error {
	w.WriteHeader(http.StatusCreated)

	return nil
}

// ErrorResponse encodes the error response to the client. it will check if it's a sentinel error or unknown error.
func ErrorResponse(ctx context.Context, err error, respWriter http.ResponseWriter) {
	var (
		appErr  exception.ApplicationError
		message string
	)

	if errors.As(err, &appErr) {
		respWriter.WriteHeader(appErr.StatusCode)

		message = appErr.Message
	} else {
		respWriter.WriteHeader(http.StatusInternalServerError)

		message = err.Error()

		slog.ErrorContext(ctx, message, slog.Any("error", err))
	}

	respWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	//nolint:errcheck,errchkjson
	json.NewEncoder(respWriter).Encode(dto.ErrorResponse{
		Error: message,
	})
}
