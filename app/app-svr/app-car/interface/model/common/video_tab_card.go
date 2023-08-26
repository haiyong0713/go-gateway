package common

import "go-gateway/app/app-svr/app-car/interface/model"

type VideoTabCardResp struct {
	Items    *VideoTabCardItems `json:"items"`
	PageNext *PageNext          `json:"page_next"`
}

type VideoTabCardItems struct {
	Banners []*Item `json:"banners"`
	Cards   []*Item `json:"cards"`
}

type PageNext struct {
	Ps int `json:"ps"`
	Pn int `json:"pn"`
}

type VideoTabCardReq struct {
	model.DeviceInfo
	TabId       int64     `form:"tab_id"`
	PageNextStr string    `form:"page_next"`
	LoginEvent  int       `form:"login_event"`
	Buvid       string    `form:"-"`
	Mid         int64     `form:"-"`
	PageNext    *PageNext `form:"-"`
	Mode        int       `form:"mode"` // 0默认，开启个性化   1关闭个性化推荐
}

type CardPlaylistReq struct {
	model.DeviceInfo
	Id    int64  `form:"id"`
	Type  string `form:"type"`
	Mid   int64  `form:"-"`
	Buvid string `form:"-"`
}

type CardPlaylistResp struct {
	Cards []*Item `json:"cards"`
}

type PageInfo struct {
	Ps          int   `json:"ps,omitempty"`
	Oid         int64 `json:"oid,omitempty"`
	Pn          int   `json:"pn,omitempty"`
	WithCurrent bool  `json:"with_current,omitempty"` // 是否包含当前游标的稿件，默认不包含
}

type VideoTabsReq struct {
	model.DeviceInfo
	Mid   int64  `form:"-"`
	Buvid string `form:"-"`
}
