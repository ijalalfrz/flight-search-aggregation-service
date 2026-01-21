package providerutils

import (
	"net/http"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/exception"
)

var ErrProviderInternalError = exception.ApplicationError{
	StatusCode: http.StatusInternalServerError,
	Message:    "provider internal error or temporary unavailable",
}

var ErrRetryExceeded = exception.ApplicationError{
	StatusCode: http.StatusInternalServerError,
	Message:    "retry exceeded",
}

var ErrProviderRateLimitExceeded = exception.ApplicationError{
	StatusCode: http.StatusTooManyRequests,
	Message:    "provider rate limit exceeded",
}
