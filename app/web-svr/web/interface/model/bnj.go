package model

import (
	arcmdl "go-gateway/app/app-svr/archive/service/api"

	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
)

// Bnj2019 .
type Bnj2019 struct {
	*Bnj2019View
	Elec    *ElecShow         `json:"elec"`
	Related []*Bnj2019Related `json:"related"`
	ReqUser *ReqUser          `json:"req_user"`
}

// Bnj2019View .
type Bnj2019View struct {
	*arcmdl.Arc
	Pages []*arcmdl.Page `json:"pages"`
}

// Bnj2019Related .
type Bnj2019Related struct {
	*arcmdl.Arc
	Pages []*arcmdl.Page `json:"pages"`
}

// ReqUser req user.
type ReqUser struct {
	Attention bool  `json:"attention"`
	Favorite  bool  `json:"favorite"`
	SeasonFav bool  `json:"season_fav"`
	Like      bool  `json:"like"`
	Dislike   bool  `json:"dislike"`
	Coin      int64 `json:"coin"`
}

// Timeline bnj timeline.
type Timeline struct {
	Name     string   `json:"name"`
	Start    int64    `json:"start"`
	End      int64    `json:"end"`
	Cover    string   `json:"cover"`
	H5Cover  string   `json:"h5_cover"`
	Subtitle string   `json:"subtitle"`
	Type     int      `json:"type"`
	Tag      []string `json:"tag"`
}

type Bnj20Cache struct {
	MainView    *arcmdl.ViewReply
	SpView      *arcmdl.SteinsGateViewReply
	ElecInfo    *payrank.BNJRankWithPanelReply
	RelatedList []*arcmdl.ViewReply
	LiveArc     *arcmdl.ArcReply
	GrayUids    map[int64]struct{}
	LiveGiftCnt int64
}

type BnjElec struct {
	TotalCount int64 `json:"total_count"`
}

type Bnj2020 struct {
	*arcmdl.ViewReply
	SpView  *arcmdl.SteinsGateViewReply `json:"sp_view"`
	Elec    *BnjElec                    `json:"elec"`
	Related []*arcmdl.ViewReply         `json:"related"`
	ReqUser *ReqUser                    `json:"req_user"`
}

type Bnj20Item struct {
	Staff   []*BnjStaff `json:"staff"`
	ReqUser *ReqUser    `json:"req_user"`
	Stat    arcmdl.Stat `json:"stat"`
	Banner  []*Banner   `json:"banner"`
}

type BnjStaff struct {
	*Staff
	Attention bool `json:"attention"`
}

type Banner struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Image      string `json:"image"`
	Hash       string `json:"hash"`
	URI        string `json:"uri"`
	RequestID  string `json:"request_id,omitempty"`
	CreativeID int64  `json:"creative_id,omitempty"`
	SrcID      int64  `json:"src_id,omitempty"`
	IsAd       bool   `json:"is_ad,omitempty"`
	IsAdLoc    bool   `json:"is_ad_loc,omitempty"`
	AdCb       string `json:"ad_cb,omitempty"`
	ShowURL    string `json:"show_url,omitempty"`
	ClickURL   string `json:"click_url,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	ServerType int64  `json:"server_type"`
	ResourceID int64  `json:"resource_id,omitempty"`
	Index      int64  `json:"index,omitempty"`
	CmMark     int64  `json:"cm_mark"`
}
