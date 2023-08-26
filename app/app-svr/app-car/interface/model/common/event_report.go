package common

import "go-gateway/app/app-svr/app-car/interface/model"

const (
	// 杜比相关
	ET_Dobly = 1
)

type EventReportReq struct {
	EventType int `json:"event_type" form:"event_type"`
	model.DeviceInfo
	Ctime    string `json:"ctime" form:"ctime"`
	Mid      int64  `json:"mid" form:"mid"`
	Buvid    string `json:"buvid" form:"buvid"`
	MobiApp  string `form:"mobi_app"`
	Platform string `form:"platform"`
	Build    int    `form:"build"`

	//dobly
	Avid  int64  `json:"avid" form:"avid"`
	Cid   int64  `json:"cid" form:"cid"`
	Scene string `form:"scene"`
}
