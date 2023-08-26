package lottery

// AddressInfo ...
type AddressInfo struct {
	ID      int64  `json:"id"`
	Type    int64  `json:"type"`
	Def     int64  `json:"def"`
	ProvID  int64  `json:"prov_id"`
	CityID  int64  `json:"city_id"`
	AreaID  int64  `json:"area_id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Addr    string `json:"addr"`
	ZipCode string `json:"zip_code"`
	Prov    string `json:"prov"`
	City    string `json:"city"`
	Area    string `json:"area"`
}
