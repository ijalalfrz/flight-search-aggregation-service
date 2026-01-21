package garuda

import "time"

type SearchFlightResponse struct {
	Status  string   `json:"status"`
	Flights []Flight `json:"flights"`
}

type Flight struct {
	FlightID        string          `json:"flight_id"`
	Airline         string          `json:"airline"`
	AirlineCode     string          `json:"airline_code"`
	Departure       FlightPoint     `json:"departure"`
	Arrival         FlightPoint     `json:"arrival"`
	DurationMinutes int             `json:"duration_minutes"`
	Stops           int             `json:"stops"`
	Aircraft        string          `json:"aircraft"`
	Price           Price           `json:"price"`
	AvailableSeats  int             `json:"available_seats"`
	FareClass       string          `json:"fare_class"`
	Baggage         Baggage         `json:"baggage"`
	Amenities       []string        `json:"amenities"`
	Segments        []FlightSegment `json:"segments"`
}

type FlightSegment struct {
	FlightNumber    string      `json:"flight_number"`
	Departure       FlightPoint `json:"departure"`
	Arrival         FlightPoint `json:"arrival"`
	DurationMinutes int         `json:"duration_minutes"`
	LayoverMinutes  int         `json:"layover_minutes"`
}

type FlightPoint struct {
	Airport  string    `json:"airport"`
	City     string    `json:"city,omitempty"`
	Time     time.Time `json:"time"`
	Terminal string    `json:"terminal,omitempty"`
}

type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type Baggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}
