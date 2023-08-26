package model

type InfocMsg struct {
	Mid            int64
	Buvid          string
	Host           string
	Path           string
	Method         string
	Header         string
	Query          string
	Body           string
	Referer        string
	IP             string
	Ctime          int64
	ResponseHeader string
	ResponseBody   string
	Sample         int64
}
