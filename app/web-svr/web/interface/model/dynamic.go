package model

import (
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"

	cheeseseasongrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

// DynamicBvArcs .
type DynamicBvArcs struct {
	Page     *model.Page `json:"page"`
	Archives []*BvArc    `json:"archives"`
}

type LpDynamicRegionReq struct {
	Rid      int64  `form:"rid" validate:"min=1"`
	Pn       int64  `form:"pn" validate:"min=1"`
	Ps       int64  `form:"ps" validate:"min=1,max=50"`
	Business string `form:"business" validate:"required"`
}

type DyRegionArcs struct {
	Page     *model.Page `json:"page"`
	Archives []*api.Arc  `json:"archives"`
}

type DynamicCard struct {
	Desc struct {
		DynamicID   int64 `json:"dynamic_id"`
		View        int32 `json:"view"`
		Like        int32 `json:"like"`
		UserProfile struct {
			Info struct {
				UID   int64  `json:"uid"`
				UName string `json:"uname"`
				Face  string `json:"face"`
			} `json:"info"`
		} `json:"user_profile"`
	} `json:"desc"`
	Card string `json:"card"`
}

type DrawDetail struct {
	Item DrawItem `json:"item"`
}

type DrawItem struct {
	ID            int64         `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Pictures      []DrawPicture `json:"pictures"`
	PicturesCount int           `json:"pictures_count"`
	Reply         int           `json:"reply"`
	UploadTime    int64         `json:"upload_time"`
	AtControl     string        `json:"at_control"`
}

type DrawPicture struct {
	ImgSrc    string  `json:"img_src"`
	ImgHeight int64   `json:"img_height"`
	ImgWidth  int64   `json:"img_width"`
	ImgSize   float32 `json:"img_size"`
}

type DynamicEntranceParam struct {
	VideoOffset   int64 `form:"video_offset"`
	ArticleOffset int64 `form:"article_offset"`
	AlltypeOffset int64 `form:"alltype_offset"`
}

type DynamicEntrance struct {
	Entrance   *DynamicEntranceItem       `json:"entrance"`
	UpdateInfo *DynamicEntranceUpdateInfo `json:"update_info"`
}

type DynamicCardAdd struct {
	IsAllow       bool                            `json:"is_allow"`
	ErrorMsg      string                          `json:"error_msg"`
	SeasonID      int32                           `json:"season_id"`
	SeasonProfile *cheeseseasongrpc.SeasonProfile `json:"season_profile"`
}

type DynamicCardType struct {
	Items []*DynamicCardTypeItem `json:"items"`
}

type DynamicCardTypeItem struct {
	Title    string `json:"title"`
	CardType int    `json:"card_type"`
}

type DynamicCardCanAddContent struct {
	LinkQuery  bool                              `json:"link_query"`
	UserSeason *cheeseseasongrpc.UserSeasonReply `json:"user_season"`
}

/*
DynamicEntranceItem的type枚举：
none-无红点;
live-直播维度的更新提醒 展示 头像+直播中+红点;
up-up主维度的更新提醒 展示 头像+红点;
dyn-动态维度的图标提醒 展示 头像+红点;
dot-动态维度的红点提醒 展示 红点
*/
type DynamicEntranceItem struct {
	Icon string `json:"icon"`
	Mid  int64  `json:"mid"`
	Type string `json:"type"`
}

/*
DynamicEntranceUpdateInfo的type枚举：
no_point-无红点;
point-红点样式;
count-数字样式
*/
type DynamicEntranceUpdateInfo struct {
	Type string                         `json:"type"`
	Item *DynamicEntranceUpdateInfoItem `json:"item"`
}

type DynamicEntranceUpdateInfoItem struct {
	Count int64 `json:"count"`
}
