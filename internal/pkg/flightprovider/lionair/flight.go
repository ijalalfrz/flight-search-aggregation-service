package lionair

type SearchFlightResponse struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}

type Data struct {
	AvailableFlights []Flight `json:"available_flights"`
}

type Flight struct {
	ID         string    `json:"id"`
	Carrier    Carrier   `json:"carrier"`
	Route      Route     `json:"route"`
	Schedule   Schedule  `json:"schedule"`
	FlightTime int       `json:"flight_time"`
	IsDirect   bool      `json:"is_direct"`
	StopCount  int       `json:"stop_count,omitempty"`
	Layovers   []Layover `json:"layovers,omitempty"`
	Pricing    Pricing   `json:"pricing"`
	SeatsLeft  int       `json:"seats_left"`
	PlaneType  string    `json:"plane_type"`
	Services   Services  `json:"services"`
}

type Carrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

type Route struct {
	From Airport `json:"from"`
	To   Airport `json:"to"`
}

type Airport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type Schedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

type Layover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

type Pricing struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
	FareType string `json:"fare_type"`
}

type Services struct {
	WifiAvailable    bool             `json:"wifi_available"`
	MealsIncluded    bool             `json:"meals_included"`
	BaggageAllowance BaggageAllowance `json:"baggage_allowance"`
}

type BaggageAllowance struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}
