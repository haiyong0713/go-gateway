package feed

//go:generate easyjson -all feedmetric.go

//easyjson:json
type Discard struct {
	ID            int64  `json:"id,omitempty"`
	Goto          string `json:"goto,omitempty"`
	DiscardReason int8   `json:"discard_reason,omitempty"`
	Error         string `json:"error,omitempty"`
}

//easyjson:json
type OpenAppURLParam struct {
	Jump  string `json:"jump"`
	Type_ string `json:"type"`
	ID    string `json:"id"`
}

//easyjson:json
type FeedAppListParam struct {
	Mid      int64  `json:"mid"`
	Buvid    string `json:"buvid"`
	MobiApp  string `json:"mobi_app"`
	Device   string `json:"device"`
	Platform string `json:"platform"`
	Build    int    `json:"build"`
	IP       string `json:"ip"`
	Ua       string `json:"ua"`
	Referer  string `json:"referer"`
	Origin   string `json:"origin"`
	CdnIp    string `json:"cdn_ip"`
	Channel  string `json:"channel"`
	Brand    string `json:"brand"`
	Model    string `json:"model"`
	Osver    string `json:"osver"`
	Applist  string `json:"applist"`
}
