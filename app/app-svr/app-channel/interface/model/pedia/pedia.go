package pedia

import (
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
)

type NavReq struct {
	Bid int64 `form:"bid" validate:"required"`
}

type NavList struct {
	Nid   int64  `json:"nid"`
	Title string `json:"title"`
}

type Navigation struct {
	List []*NavList `json:"list"`
}

type NavPart struct {
	Nid      int64      `json:"nid"`
	Position int        `json:"position,omitempty"`
	Title    string     `json:"title"`
	Part     []*NavPart `json:"part"`
}

type BaikeTree struct {
	ContentTitle string     `json:"content_title"`
	Part         []*NavPart `json:"part"`
}

type BaikeInfo struct {
	BaikeName string `json:"baike_name"`
	Desc      string `json:"desc"`
}

type NavResponse struct {
	Version    string      `json:"version"`
	Navigation *Navigation `json:"navigation"`
	BaikeTree  *BaikeTree  `json:"baike_tree"`
	BaikeInfo  *BaikeInfo  `json:"baike_info"`
}

type FeedReq struct {
	Bid      int64  `form:"bid" validate:"required"`
	Nid      int64  `form:"nid"`
	Vertical int32  `form:"vertical"`
	Offset   string `form:"offset"`
	Version  string `form:"version" validate:"required"`
	Ps       int32  `form:"ps"`
}

type FeedItem struct {
	CardType   string `json:"card_type"`
	NavNid     int64  `json:"nav_nid"`
	ContentNid int64  `json:"content_nid"`
	FirstNid   int64  `json:"first_nid"`
	SecondNid  int64  `json:"second_nid"`
	BaikeTitle string `json:"baike_title"`
	*jsoncard.LargeCoverInline
	Desc  string `json:"desc"`
	Image string `json:"image"`
}

type FeedResponse struct {
	Items      []*FeedItem `json:"items"`
	UpOffset   string      `json:"up_offset,omitempty"`
	DownOffset string      `json:"down_offset,omitempty"`
	UpMore     bool        `json:"up_more,omitempty"`
	DownMore   bool        `json:"down_more,omitempty"`
}
