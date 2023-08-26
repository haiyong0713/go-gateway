package recommend

import (
	"encoding/json"
	cardproto "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	"go-gateway/app/app-svr/app-show/interface/model/card"
)

// Arc is index show recommend.
type Arc struct {
	Aid         interface{} `json:"aid"`
	Author      string      `json:"author"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Pic         string      `json:"pic"`
	Views       interface{} `json:"play"`
	Comments    int64       `json:"review"`
	Coins       int64       `json:"coins"`
	Danmaku     int         `json:"video_review"`
	Favorites   int64       `json:"favorites"`
	Pts         int64       `json:"pts"`
	Others      []*Arc      `json:"others"`
}

type List struct {
	Aid        int64  `json:"aid"`
	Desc       string `json:"desc"`
	CornerMark int8   `json:"corner_mark"`
}

type CardList struct {
	ID         int64            `json:"id"`
	Goto       string           `json:"goto"`
	FromType   string           `json:"from_type"`
	Desc       string           `json:"desc"`
	CornerMark int8             `json:"corner_mark"`
	CoverGif   string           `json:"cover_gif"`
	Condition  []*CardCondition `json:"condition"`
	HotwordID  int64            `json:"hotword_id"`
	RcmdReason *RcmdReason      `json:"rcmd_reason"`
}

type CardCondition struct {
	Plat      int8   `json:"plat"`
	Condition string `json:"conditions"`
	Build     int    `json:"build"`
}

type RcmdReason struct {
	Content string `json:"content"`
	Style   int8   `json:"style"`
}

// nolint:gomnd
func (c *CardList) CardListChange2() (p *card.PopularCard) {
	p = &card.PopularCard{
		Value:    c.ID,
		Type:     c.Goto,
		FromType: "recommend",
	}
	if c.RcmdReason != nil {
		p.Reason = c.RcmdReason.Content
		p.ReasonType = c.RcmdReason.Style
		if p.CornerMark = c.CornerMark; p.CornerMark == 2 {
			p.CornerMark = 4
		}
	}
	return
}

type HotItem struct {
	ID         int64           `json:"id"`
	Goto       string          `json:"goto"`
	FromType   string          `json:"from_type"`
	Source     string          `json:"source,omitempty"`
	RcmdReason *HotRcmdReason  `json:"rcmd_reason,omitempty"`
	AvFeature  json.RawMessage `json:"av_feature,omitempty"`
	Sticky     int             `json:"sticky,omitempty"`
	IsGif      int             `json:"is_gif,omitempty"`
	HotwordID  int64           `json:"hotword_id,omitempty"`
	TrackID    string          `json:"trackid,omitempty"`
	GifCover   string          `json:"gif_cover,omitempty"`
}

type BizData struct {
	RequestID  string    `json:"request_id"`
	SourceID   int64     `json:"source_id"`
	ResourceID int64     `json:"resource_id"`
	IsAdLoc    bool      `json:"is_ad_loc"`
	ClientIp   string    `json:"client_ip"`
	CardIndex  int64     `json:"card_index"`
	AdContent  AdContent `json:"ad_content"`
}

type AdContent struct {
	CreativeID int64           `json:"creative_id"`
	VideoID    int64           `json:"video_id"`
	AdCb       string          `json:"ad_cb"`
	Extra      json.RawMessage `json:"extra"`
}

type HotRcmdReason struct {
	Content    string `json:"content,omitempty"`
	CornerMark int8   `json:"corner_mark,omitempty"`
}

// nolint:gomnd
func (i *HotItem) HotItemChange() *card.PopularCard {
	p := &card.PopularCard{
		Value:     i.ID,
		Type:      i.Goto,
		FromType:  i.FromType,
		TrackID:   i.TrackID,
		Source:    i.Source,
		AvFeature: i.AvFeature,
		HotwordID: i.HotwordID,
		CoverGif:  i.GifCover,
	}
	if i.RcmdReason != nil {
		p.Reason = i.RcmdReason.Content
		if p.CornerMark = i.RcmdReason.CornerMark; p.CornerMark == 2 {
			p.CornerMark = 4
		}
		if p.Reason != "" {
			p.ReasonType = 3
		}
	}
	return p
}

func (b *BizData) ToCardAdInfo() *cardproto.AdInfo {
	if b == nil {
		return nil
	}
	out := &cardproto.AdInfo{
		Index:      int32(b.CardIndex),
		RequestId:  b.RequestID,
		Source:     int32(b.SourceID),
		IsAdLoc:    b.IsAdLoc,
		ClientIp:   b.ClientIp,
		CreativeId: b.AdContent.CreativeID,
		AdCb:       b.AdContent.AdCb,
		Resource:   b.ResourceID,
		CardIndex:  int32(b.CardIndex),
		CreativeContent: &cardproto.CreativeContent{
			VideoId: b.AdContent.VideoID,
		},
		Extra: b.AdContent.Extra,
	}
	return out
}
