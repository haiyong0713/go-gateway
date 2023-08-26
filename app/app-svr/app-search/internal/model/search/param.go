package search

type Param struct {
	AccessKey     string `form:"access_key"`
	MobiApp       string `form:"mobi_app" validate:"required"`
	Platform      string `form:"platform" validate:"required"`
	Device        string `form:"device"`
	Build         int64  `form:"build" validate:"required"`
	Keyword       string `form:"keyword"`
	TeenagersMode int    `form:"teenagers_mode"`
	LessonsMode   int    `form:"lessons_mode"`
	PS            int    `form:"ps"`
	PN            int    `form:"pn"`
	Buvid         string `form:"buvid"`
	MID           int64  `form:"mid"`
	Plat          int8   `form:"plat"`
	Spmid         string `form:"spmid"`
}

type SiriCommandReq struct {
	Command  string `form:"command" validate:"required"`
	Debug    bool   `form:"__debug"`
	MobiApp  string `form:"mobi_app" validate:"required"`
	Platform string `form:"platform" validate:"required"`
	Device   string `form:"device" validate:"required"`
	Build    int64  `form:"build" validate:"required"`
	Channel  string `form:"channel"`
	Buvid    string
	Mid      int64
}

type DefaultWordsExtParam struct {
	Tab         string
	EventId     string
	Avid        string
	Query       string
	An          int64
	IsFresh     int64
	DisableRcmd int64
}
