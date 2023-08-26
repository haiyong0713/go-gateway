package banner

import (
	"encoding/json"

	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	resource "go-gateway/app/app-svr/resource/service/model"
)

type Banner struct {
	ID                  int64                   `json:"id"`
	Title               string                  `json:"title"`
	Image               string                  `json:"image"`
	Hash                string                  `json:"hash"`
	URI                 string                  `json:"uri"`
	RequestID           string                  `json:"request_id,omitempty"`
	CreativeID          int                     `json:"creative_id,omitempty"`
	SrcID               int                     `json:"src_id,omitempty"`
	IsAd                bool                    `json:"is_ad,omitempty"`
	IsAdLoc             bool                    `json:"is_ad_loc,omitempty"`
	IsAdReplace         bool                    `json:"-"`
	AdCb                string                  `json:"ad_cb,omitempty"`
	ShowURL             string                  `json:"show_url,omitempty"`
	ClickURL            string                  `json:"click_url,omitempty"`
	ClientIP            string                  `json:"client_ip,omitempty"`
	ServerType          int                     `json:"server_type"`
	ResourceID          int                     `json:"resource_id,omitempty"`
	Index               int                     `json:"index,omitempty"`
	CmMark              int                     `json:"cm_mark"`
	SplashID            int64                   `json:"splash_id,omitempty"`
	IsTopview           bool                    `json:"is_topview,omitempty"`
	Extra               json.RawMessage         `json:"extra,omitempty"`
	BannerMeta          resourcegrpc.BannerMeta `json:"-"`
	InlineUseSame       int64                   `json:"-"`
	InlineBarrageSwitch int64                   `json:"-"`
}

func (b *Banner) Change(banner *resource.Banner) {
	b.ID = int64(banner.ID)
	b.Title = banner.Title
	b.Image = banner.Image
	b.Hash = banner.Hash
	b.URI = banner.URI
	b.ResourceID = banner.ResourceID
	b.RequestID = banner.RequestId
	b.CreativeID = banner.CreativeId
	b.SrcID = banner.SrcId
	b.IsAd = banner.IsAd
	b.IsAdLoc = banner.IsAdLoc
	b.CmMark = banner.CmMark
	b.AdCb = banner.AdCb
	b.ShowURL = banner.ShowUrl
	b.ClickURL = banner.ClickUrl
	b.ClientIP = banner.ClientIp
	b.Index = banner.Index
	b.ServerType = banner.ServerType
	b.Extra = banner.Extra
	b.SplashID = banner.SplashID
	if b.SplashID > 0 {
		b.IsTopview = true
	}
}

func (b *Banner) FromProto(banner *resourcegrpc.Banner) {
	b.ID = int64(banner.Id)
	b.Title = banner.Title
	b.Image = banner.Image
	b.Hash = banner.Hash
	b.URI = banner.URI
	b.ResourceID = int(banner.ResourceId)
	b.RequestID = banner.RequestId
	b.CreativeID = int(banner.CreativeId)
	b.SrcID = int(banner.SrcId)
	b.IsAd = banner.IsAd
	b.IsAdLoc = banner.IsAdLoc
	b.CmMark = int(banner.CmMark)
	b.AdCb = banner.AdCb
	b.ShowURL = banner.ShowUrl
	b.ClickURL = banner.ClickUrl
	b.ClientIP = banner.ClientIp
	b.Index = int(banner.Index)
	b.ServerType = int(banner.ServerType)
	b.Extra = banner.Extra
	b.SplashID = banner.SplashId
	if b.SplashID > 0 {
		b.IsTopview = true
	}
	b.BannerMeta = banner.BannerMeta
	b.InlineUseSame = banner.InlineUseSame
	b.InlineBarrageSwitch = banner.InlineBarrageSwitch
}
