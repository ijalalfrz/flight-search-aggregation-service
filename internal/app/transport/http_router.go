package transport

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/config"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/endpoints"
	httptransport "github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/transport/http"
)

// MakeHTTPRouter builds the HTTP router with all the service endpoints.
func MakeHTTPRouter(
	cfg *config.Config,
	endpts endpoints.Endpoints,
) *chi.Mux {
	// Initialize Router
	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router.Route("/api/v1/flights", func(router chi.Router) {
		router.Use(
			httptransport.RequestID(),
			httptransport.CORSMiddleware(),
			httptransport.Recoverer(slog.Default()),
			render.SetContentType(render.ContentTypeJSON),
		)

		router.Post("/search", httptransport.MakeHandlerFunc(
			endpts.AggregatorEndpoint.SearchFlights,
			httptransport.DecodeRequest[dto.SearchCriteria],
			httptransport.ResponseWithBody,
		))
	})

	return router
}
