package display

// Zone ip struct info.
type Zone struct {
	ID          int64   `json:"id"`
	Addr        string  `json:"addr"`
	ISP         string  `json:"isp"`
	Country     string  `json:"country"`
	Province    string  `json:"province"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CountryCode int     `json:"country_code,omitempty"`
}
