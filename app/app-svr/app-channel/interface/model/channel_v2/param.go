package channel_v2

// Param is
type Param struct {
	AccessKey     string `form:"access_key"`
	MobiApp       string `form:"mobi_app" validate:"required"`
	Platform      string `form:"platform" validate:"required"`
	Device        string `form:"device"`
	Build         int    `form:"build" validate:"required"`
	TeenagersMode int    `form:"teenagers_mode"`
	Lang          string `form:"lang" default:"hans"`
	Channel       string `form:"channel"`
	Offset        string `form:"offset" default:""`
	Spmid         string `form:"spmid"`
	Statistics    string `form:"statistics"`
	TS            int64  `form:"ts"`
	AutoRefresh   int    `form:"auto_refresh" default:"0"`
	PN            int    `json:"pn"`
	OffsetNew     string `form:"offset_new" default:""`
	OffsetRcmd    string `form:"offset_rcmd" default:""`
	Buvid         string `json:"buvid"`
	ReqURL        string `json:"req_url"`
	TimeIso       int64  `json:"timeIso"`
	MID           int64  `json:"mid"`
	NetType       int32
	TfType        int32
	SLocal        string `form:"s_locale"`
	CLocal        string `form:"c_locale"`
}
