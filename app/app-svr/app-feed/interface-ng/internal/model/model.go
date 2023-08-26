package model

import (
	"time"

	"go-gateway/app/app-svr/app-card/interface/model/card"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"

	locationgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

type CompareSessionReply struct {
	CardType string `json:"card_type"`
	CardGoto string `json:"card_goto"`
	Goto     string `json:"goto"`
	Param    string `json:"param"`
	MobiApp  string `json:"mobi_app"`
	Result   string `json:"result"`
}

type IndexNgReq struct {
	Mid        int64               `json:"mid"`
	FeedParam  feed.IndexParam     `json:"feed_param"`
	Style      int                 `json:"style"`
	AppList    string              `json:"app_list"`
	DeviceInfo string              `json:"device_info"`
	Device     *feedcard.CtxDevice `json:"device"`
}

type IndexNgReply struct {
	Items  []card.Handler `json:"items"`
	Config *feed.Config   `json:"config"`
}

type AiReq struct {
	IndexNgReq   IndexNgReq              `json:"index_ng_req"`
	Group        int                     `json:"group"`
	AvAdResource int64                   `json:"av_ad_resource"`
	AutoPlay     string                  `json:"auto_play"`
	NoCache      bool                    `json:"no_cache"`
	ResourceID   int                     `json:"resource_id"`
	BannerExp    int                     `json:"banner_exp"`
	AdExp        int64                   `json:"ad_exp"`
	Zone         *locationgrpc.InfoReply `json:"zone"`
	Now          time.Time               `json:"now"`
}

type ConstructCardParam struct {
	IndexNgReq *IndexNgReq             `json:"index_ng_req"`
	AIResponse *feed.AIResponse        `json:"ai_response"`
	Zone       *locationgrpc.InfoReply `json:"zone"`
}
