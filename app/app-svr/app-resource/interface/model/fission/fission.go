package fission

// ParamCheck  fission check new params.
type ParamCheck struct {
	Mid      int64  `form:"-"`
	Buvid    string `form:"-"`
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Platform string `form:"platform"`
	Build    int64  `form:"build"`
}
