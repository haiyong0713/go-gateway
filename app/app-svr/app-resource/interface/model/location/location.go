package location

// auth const
const (
	Forbidden = int64(1)
)

// Info ipinfo with the smallest zone_id.
type Info struct {
	Addr        string `json:"addr"`
	Country     string `json:"country"`
	Province    string `json:"province"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	ZoneId      int64  `json:"zone_id"`
	CountryCode int64  `json:"country_code"`
}
