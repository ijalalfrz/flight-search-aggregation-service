package batikair

type SearchFlightResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Results []Flight `json:"results"`
}

type Flight struct {
	FlightNumber      string       `json:"flightNumber"`
	AirlineName       string       `json:"airlineName"`
	AirlineIATA       string       `json:"airlineIATA"`
	Origin            string       `json:"origin"`
	Destination       string       `json:"destination"`
	DepartureDateTime string       `json:"departureDateTime"`
	ArrivalDateTime   string       `json:"arrivalDateTime"`
	TravelTime        string       `json:"travelTime"`
	NumberOfStops     int          `json:"numberOfStops"`
	Connections       []Connection `json:"connections"`
	Fare              Fare         `json:"fare"`
	SeatsAvailable    int          `json:"seatsAvailable"`
	AircraftModel     string       `json:"aircraftModel"`
	BaggageInfo       string       `json:"baggageInfo"`
	OnboardServices   []string     `json:"onboardServices"`
}

type Connection struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

type Fare struct {
	BasePrice    int    `json:"basePrice"`
	Taxes        int    `json:"taxes"`
	TotalPrice   int    `json:"totalPrice"`
	CurrencyCode string `json:"currencyCode"`
	Class        string `json:"class"`
}
