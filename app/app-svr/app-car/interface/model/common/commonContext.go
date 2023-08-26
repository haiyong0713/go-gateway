package common

import (
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	bangumimdl "go-gateway/app/app-svr/app-car/interface/model/bangumi"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

type ItemType string

type Otype string

type MaterialType string

const (
	MaterialTypeUGC          = MaterialType("ugc")     // ugc稿件（秒开）
	MaterialTypeUGCPlus      = MaterialType("ugcPlus") // ugc稿件（秒开、区分单p多p）
	MaterialTypeOGVEP        = MaterialType("ogv_ep")
	MaterialTypeOGVSeaon     = MaterialType("ogv_season")
	MaterialTypeUGCView      = MaterialType("ugc_view")
	MaterialTypeOGVView      = MaterialType("ogv_view")
	MaterialTypeVideoSerial  = MaterialType("video_serial")  // 视频合集
	MaterialTypeFmSerial     = MaterialType("fm_serial")     // FM合集
	MaterialTypeVideoChannel = MaterialType("video_channel") // 视频频道
	MaterialTypeFmChannel    = MaterialType("fm_channel")    // FM频道

	ItemTypeUGC          = ItemType("ugc")           // ugc稿件
	ItemTypeOGV          = ItemType("ogv")           // ogv稿件
	ItemTypeUGCSingle    = ItemType("ugc_single")    // ugc单p
	ItemTypeUGCMulti     = ItemType("ugc_multi")     // ugc多p
	ItemTypeVideoSerial  = ItemType("video_serial")  // 视频合集
	ItemTypeVideoChannel = ItemType("video_channel") // 视频频道
	ItemTypeFmSerial     = ItemType("fm_serial")     // FM合集
	ItemTypeFmChannel    = ItemType("fm_channel")    // FM频道

	OtypeUGC  = Otype("ugc")  // ugc
	OtypePGC  = Otype("pgc")  // pgc
	OtypeLive = Otype("live") // live
)

type DeviceInfo struct {
	MobiApp  string `json:"mobi_app"`
	Device   string `json:"device"`
	Platform string `json:"platform"`
	Build    int64  `json:"build"`
}

type Params struct {
	ArchiveReq      *ArchiveReq     // 稿件详情(支持秒开)
	ArchivePlusReq  *ArchivePlusReq // 稿件详情(支持秒开)，和分p信息
	EpisodeReq      *EpisodeReq     // epid获取详情(无秒开)
	SeasonReq       *SeasonReq      // ogv season
	AccountCardReq  *AccountCardReq // 用户信息
	UGCViewReq      *UGCViewReq
	OGVViewReq      *OGVViewReq
	SerialInfosReq  *SerialInfosReq  // 合集基本信息
	SerialArcsReq   *SerialArcsReq   // 合集内部稿件id（分页）
	ChannelInfosReq *ChannelInfosReq // 频道基本信息
	ChannelArcsReq  *ChannelArcsReq  // 频道内部稿件id（分页）
	Mid             int64
	Buvid           string
}

type ArchiveReq struct {
	PlayAvs []*archivegrpc.PlayAv
}

type ArchivePlusReq struct {
	PlayAvs []*archivegrpc.PlayAv
}

type EpisodeReq struct {
	Epids []int32
}

type SeasonReq struct {
	Sids []int32
}

type AccountCardReq struct {
	Mids []int64
}

type UGCViewReq struct {
	Aids []int64
}

type OGVViewReq struct {
	Sid       int64
	AccessKey string
	Cookie    string
	Referer   string
}

type CarContext struct {
	// 列表数据
	OriginData *OriginData
	// 物料详情
	ArchiveResp       map[int64]*archivegrpc.ArcPlayer         // 稿件详情(支持秒开)
	ArchivePlusResp   map[int64]*ArchivePlusResp               // 稿件详情(支持秒开)，和ep信息
	EpisodeResp       map[int32]*episodegrpc.EpisodeCardsProto // epid获取详情(无秒开)
	EpisodeInlineResp map[int32]*pgcinline.EpisodeCard
	SeasonResp        map[int32]*seasongrpc.CardInfoProto
	AccountCardResp   map[int64]*accountgrpc.Card
	UGCViewResp       map[int64]*archivegrpc.ViewReply
	OGVViewResp       *bangumimdl.View
	SerialInfosResp   *SerialInfosResp
	SerialArcsResp    *SerialArcsResp
	ChannelInfosResp  *ChannelInfosResp
	ChannelArcsResp   *ChannelArcsResp
}

type OriginData struct {
	MaterialType MaterialType
	Oid          int64
	Cid          int64
}

type ArchivePlusResp struct {
	Player *archivegrpc.ArcPlayer
	View   *archivegrpc.ViewReply
}
