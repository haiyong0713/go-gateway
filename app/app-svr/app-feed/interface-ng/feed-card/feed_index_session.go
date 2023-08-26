package feedcard

import (
	cdm "go-gateway/app/app-svr/app-card/interface/model"
)

type IndexParam struct {
	Idx               int64            `form:"idx" default:"0"`
	Pull              bool             `form:"pull" default:"true"`
	Column            cdm.ColumnStatus `form:"column"`
	LoginEvent        int              `form:"login_event" default:"0"`
	OpenEvent         string           `form:"open_event"`
	BannerHash        string           `form:"banner_hash"`
	AdExtra           string           `form:"ad_extra"`
	Interest          string           `form:"interest"`
	Flush             int              `form:"flush"`
	AutoPlayCard      int              `form:"autoplay_card"`
	DeviceType        int              `form:"device_type"`
	ParentMode        int              `form:"parent_mode"`
	ForceHost         int              `form:"force_host"`
	RecsysMode        int              `form:"recsys_mode"`
	TeenagersMode     int              `form:"teenagers_mode"`
	LessonsMode       int              `form:"lessons_mode"`
	DeviceName        string           `form:"device_name"`
	AccessKey         string           `form:"access_key"`
	ActionKey         string           `form:"actionKey"`
	Statistics        string           `form:"statistics"`
	Appver            int              `form:"appver"`
	Filtered          int              `form:"filtered"`
	AppKey            string           `form:"appkey"`
	HttpsUrlReq       int              `form:"https_url_req"`
	InterestV2        string           `form:"interest_v2"`
	SplashID          int64            `form:"splash_id"`
	Guidance          int              `form:"guidance"`
	AppList           string
	DeviceInfo        string
	ColumnTimestamp   int64 `form:"column_timestamp"`
	AutoplayTimestamp int64 `form:"autoplay_timestamp"`
	DisableRcmd       int   `form:"disable_rcmd"`
}
