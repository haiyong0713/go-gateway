package model

type AdReq struct {
	Mid          int64
	Buvid        string
	Build        int
	Resource     []int64
	Country      string
	Province     string
	City         string
	Network      string
	MobiApp      string
	Device       string
	OpenEvent    string
	AdExtra      string
	Style        int
	MayResistGif int
}
