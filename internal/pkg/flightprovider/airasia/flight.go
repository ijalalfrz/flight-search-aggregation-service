package airasia

import "time"

type SearchFlightResponse struct {
	Status  string   `json:"status"`
	Flights []Flight `json:"flights"`
}

type Flight struct {
	FlightCode    string    `json:"flight_code"`
	Airline       string    `json:"airline"`
	FromAirport   string    `json:"from_airport"`
	ToAirport     string    `json:"to_airport"`
	DepartTime    time.Time `json:"depart_time"`
	ArriveTime    time.Time `json:"arrive_time"`
	DurationHours float64   `json:"duration_hours"`
	DirectFlight  bool      `json:"direct_flight"`
	PriceIDR      int       `json:"price_idr"`
	Seats         int       `json:"seats"`
	CabinClass    string    `json:"cabin_class"`
	BaggageNote   string    `json:"baggage_note"`
	Stops         []Stop    `json:"stops"`
}

type Stop struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int    `json:"wait_time_minutes"`
}
