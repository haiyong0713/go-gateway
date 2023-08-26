package common

import "go-gateway/app/app-svr/app-car/interface/model"

type FavoriteReq struct {
	model.DeviceInfo
}

type Favorite struct {
	Fid   int64  `json:"fid"`
	Mid   int64  `json:"mid"`
	State int32  `json:"state"`
	Count int    `json:"count"`
	Name  string `json:"name"`
	Cover string `json:"cover"`
}

type FavoriteVideoReq struct {
	model.DeviceInfo
	Fid      int64  `json:"fid" form:"fid"`
	PageNext string `json:"page_next" form:"page_next"`
	Ps       int    `json:"ps" form:"ps"`
}

type FavoriteVideoResp struct {
	Items    []*Item                `json:"items"`
	PageNext *FavoriteVideoPageNext `json:"page_next"`
}

type FavoriteVideoPageNext struct {
	Pn int `json:"pn"`
	Ps int `json:"ps"`
}

type FavoriteBangumiReq struct {
	model.DeviceInfo
	PageNext string `json:"page_next" form:"page_next"`
	Ps       int    `json:"ps" form:"ps"`
}

type FavoriteBangumiResp struct {
	Items    []*Item              `json:"items"`
	PageNext *FavoriteOGVPageNext `json:"page_next"`
}

type FavoriteCinemaReq struct {
	model.DeviceInfo
	PageNext string `json:"page_next" form:"page_next"`
	Ps       int    `json:"ps" form:"ps"`
}

type FavoriteCinemaResp struct {
	Items    []*Item              `json:"items"`
	PageNext *FavoriteOGVPageNext `json:"page_next"`
}

type FavoriteOGVPageNext struct {
	Pn int `json:"pn"`
	Ps int `json:"ps"`
}

type FavoriteToViewReq struct {
	model.DeviceInfo
	Ps       int    `json:"ps" form:"ps"`
	PageNext string `json:"page_next" form:"page_next"`
}

type FavoriteToViewResp struct {
	Items    []*Item                 `json:"items"`
	PageNext *FavoriteToViewPageNext `json:"page_next"`
}

type FavoriteToViewPageNext struct {
	Pn int `json:"pn"`
	Ps int `json:"ps"`
}
