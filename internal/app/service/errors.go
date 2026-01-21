package service

import (
	"net/http"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/exception"
)

var ErrNoFlightsFound = exception.ApplicationError{
	Message:    "no flights found",
	StatusCode: http.StatusNotFound,
}
