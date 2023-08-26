package search

import (
	"encoding/json"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"

	account "git.bilibili.co/bapis/bapis-go/account/service"
)

type ADContent struct {
	CreativeID int64           `json:"creative_id"`
	CardType   int64           `json:"card_type"`
	UPMid      int64           `json:"up_mid"`
	Aids       []int64         `json:"aids"`
	ADCB       string          `json:"ad_cb"`
	Extra      json.RawMessage `json:"extra"`
	GameID     int64           `json:"game_id"`
}

type ADResource struct {
	RequestID  string     `json:"request_id"`
	SourceID   int64      `json:"source_id"`
	ResourceID int64      `json:"resource_id"`
	IsADLoc    bool       `json:"is_ad_loc"`
	ClientIP   string     `json:"client_ip"`
	ADContent  *ADContent `json:"ad_content"`
}

type ADInfo struct {
	Resource  int64  `json:"resource,omitempty"`
	Source    int64  `json:"source,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	IsAdLoc   bool   `json:"is_ad_loc,omitempty"`
	ClientIP  string `json:"client_ip,omitempty"`

	CreativeID int64           `json:"creative_id,omitempty"`
	CardType   int64           `json:"card_type,omitempty"`
	ADCB       string          `json:"ad_cb,omitempty"`
	Extra      json.RawMessage `json:"extra,omitempty"`
	UPMid      int64           `json:"up_mid,omitempty"`
	Aids       []int64         `json:"aids,omitempty"`
	GameID     int64           `json:"game_id,omitempty"`
}

type BrandAD struct {
	BIZData *ADResource `json:"biz_data"`
}

type GameAD struct {
	BIZData *ADResource `json:"biz_data"`
}

func (b *BrandAD) GetADContent() *ADContent {
	if b == nil {
		return nil
	}
	if b.BIZData == nil {
		return nil
	}
	return b.BIZData.ADContent
}

type BrandADInline struct {
	BIZData *ADResource `json:"biz_data"`
}

func (b *BrandADInline) GetADContent() *ADContent {
	if b == nil {
		return nil
	}
	if b.BIZData == nil {
		return nil
	}
	return b.BIZData.ADContent
}

type BrandADArc struct {
	Param string `json:"param,omitempty"`
	Goto  string `json:"goto,omitempty"`

	Aid           int64  `json:"aid,omitempty"`
	Play          int64  `json:"play,omitempty"`
	Reply         int64  `json:"reply,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Author        string `json:"author,omitempty"`
	Title         string `json:"title,omitempty"`
	URI           string `json:"uri,omitempty"`
	Cover         string `json:"cover,omitempty"`
	ShowCardDesc2 string `json:"show_card_desc_2,omitempty"`
}

type BrandADAccount struct {
	Param string `json:"param,omitempty"`
	Goto  string `json:"goto,omitempty"`

	Mid            int64             `json:"mid,omitempty"`
	Name           string            `json:"name,omitempty"`
	Face           string            `json:"face,omitempty"`
	Sign           string            `json:"sign,omitempty"`
	Relation       *cardmdl.Relation `json:"relation,omitempty"`
	RoomID         int64             `json:"roomid,omitempty"`
	LiveStatus     int64             `json:"live_status,omitempty"`
	LiveLink       string            `json:"live_link,omitempty"`
	OfficialVerify *OfficialVerify   `json:"official_verify,omitempty"`
	Vip            *account.VipInfo  `json:"vip,omitempty"`
	URI            string            `json:"uri,omitempty"`
	FaceNftNew     int32             `json:"face_nft_new,omitempty"`
}

func AsADInfo(arg *ADResource) (*ADInfo, bool) {
	if arg == nil {
		return nil, false
	}
	if arg.ADContent == nil {
		return nil, false
	}
	adInfo := &ADInfo{
		Resource:  arg.ResourceID,
		Source:    arg.SourceID,
		RequestID: arg.RequestID,
		IsAdLoc:   true,
		ClientIP:  arg.ClientIP,
	}
	adInfo.CardType = arg.ADContent.CardType
	adInfo.CreativeID = arg.ADContent.CreativeID
	adInfo.ADCB = arg.ADContent.ADCB
	adInfo.Extra = arg.ADContent.Extra
	adInfo.UPMid = arg.ADContent.UPMid
	adInfo.Aids = arg.ADContent.Aids
	adInfo.GameID = arg.ADContent.GameID
	return adInfo, true
}
