package main

// key info ...
const (
	_reqUrl      = "http://api.bilibili.co/x/internal/filter"
	_replaceWord = "<ep>"
	_filterArea  = "ep_ci"
	_filterLevel = 20
)

// CommonResp ...
type CommonResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
}

// filterResp ...
type filterResp struct {
	CommonResp
	Data *FilterData `json:"data"`
}

// FilterData ...
type FilterData struct {
	Level  int64    `json:"level"`
	Limit  int64    `json:"limit"`
	Msg    string   `json:"msg"`
	TypeID []int64  `json:"typeid"`
	Hit    []string `json:"hit"`
}
