package model

import (
	"time"

	"go-gateway/app/app-svr/app-card/interface/model"
)

type RecommendReq struct {
	Mid            int64              `json:"mid"`
	Plat           int8               `json:"plat"`
	Buvid          string             `json:"buvid"`
	Build          int                `json:"build"`
	LoginEvent     int                `json:"login_event"`
	ParentMode     int                `json:"parent_mode"`
	RecsysMode     int                `json:"recsys_mode"`
	TeenagersMode  int                `json:"teenagers_mode"`
	LessonsMode    int                `json:"lessons_mode"`
	ZoneID         int64              `json:"zone_id"`
	Group          int                `json:"group"`
	Interest       string             `json:"interest"`
	Network        string             `json:"network"`
	Style          int                `json:"style"`
	Column         model.ColumnStatus `json:"column"`
	Flush          int                `json:"flush"`
	IndexCount     int                `json:"index_count"`
	DeviceType     int                `json:"device_type"`
	AvAdResource   int64              `json:"av_ad_resource"`
	AdResource     int64              `json:"resource"`
	AutoPlay       string             `json:"auto_play"`
	DeviceName     string             `json:"device_name"`
	OpenEvent      string             `json:"open_event"`
	BannerHash     string             `json:"banner_hash"`
	AppList        string             `json:"app_list"`
	DeviceInfo     string             `json:"device_info"`
	InterestSelect string             `json:"interest_select"`
	ResourceID     int                `json:"resource_id"`
	BannerExp      int                `json:"banner_exp"`
	AdExp          int                `json:"ad_exp"`
	MobiApp        string             `json:"mobi_app"`
	AdExtra        string             `json:"ad_extra"`
	Pull           bool               `json:"pull"`
	RedPoint       int64              `json:"red_point"`
	InlineSound    int64              `json:"inline_sound"`
	InlineDanmu    int64              `json:"inline_danmu"`
	Now            time.Time          `json:"now"`
}
