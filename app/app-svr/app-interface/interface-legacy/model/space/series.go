package space

// SeriesParam def.
type SeriesParam struct {
	MobiApp  string `form:"mobi_app"`
	Vmid     int64  `form:"vmid" validate:"required"`
	SeriesId int64  `form:"series_id" validate:"required"`
	Ps       int64  `form:"ps" validate:"required"`
	Next     int64  `form:"next"`
	Clocale  string `form:"clocale"`
	Slocale  string `form:"slocale"`
	Device   string `form:"device"`
	SLocaleP string `form:"s_locale"`
	CLocaleP string `form:"c_locale"`
	Sort     string `form:"sort"`
}
