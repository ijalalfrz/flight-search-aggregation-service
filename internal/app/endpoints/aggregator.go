package endpoints

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

type AggregatorService interface {
	SearchFlights(ctx context.Context, req dto.SearchCriteria) (dto.SearchFlightResponse, error)
}

type AggregatorEndpoint struct {
	SearchFlights endpoint.Endpoint
}

func MakeAggregatorEndpoint(service AggregatorService) AggregatorEndpoint {
	return AggregatorEndpoint{
		SearchFlights: makeSearchFlightsEndpoint(service),
	}
}

func makeSearchFlightsEndpoint(service AggregatorService) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		request, ok := req.(*dto.SearchCriteria)
		if !ok || request == nil {
			return nil, errors.New("invalid type")
		}

		flight, err := service.SearchFlights(ctx, *request)
		if err != nil {
			return nil, fmt.Errorf("aggregator service: %w", err)
		}

		return flight, nil
	}
}
