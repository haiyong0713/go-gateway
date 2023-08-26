package search

import (
	"encoding/json"

	arcmdl "go-gateway/app/app-svr/archive/service/api"
)

// SearchVideoRes .
type SearchVideoRes struct {
	Code           int             `json:"code,omitempty"`
	SeID           string          `json:"seid,omitempty"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"pagesize,omitempty"`
	Total          int             `json:"total,omitempty"`
	NumResults     int             `json:"numResults"`
	NumPages       int             `json:"numPages"`
	SuggestKeyword string          `json:"suggest_keyword"`
	RqtType        string          `json:"rqt_type,omitempty"`
	CostTime       json.RawMessage `json:"cost_time,omitempty"`
	ExpList        json.RawMessage `json:"exp_list,omitempty"`
	EggHit         int             `json:"egg_hit"`
	PageInfo       json.RawMessage `json:"pageinfo,omitempty"`
	Result         []*SearchVideo  `json:"result,omitempty"`
	ShowColumn     int             `json:"show_column"`
	InBlackKey     int8            `json:"in_black_key"`
	InWhiteKey     int8            `json:"in_white_key"`
}

// SearchVideo search video .
type SearchVideo struct {
	Type         string             `json:"type"`
	ID           int64              `json:"id"`
	Author       string             `json:"author"`
	Mid          int64              `json:"mid"`
	Typeid       string             `json:"typeid"`
	Typename     string             `json:"typename"`
	Arcurl       string             `json:"arcurl"`
	Aid          int64              `json:"aid"`
	Bvid         string             `json:"bvid"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Arcrank      string             `json:"arcrank"`
	Pic          string             `json:"pic"`
	Play         interface{}        `json:"play"`
	VideoReview  int64              `json:"video_review"`
	Favorites    int64              `json:"favorites"`
	Tag          string             `json:"tag"`
	Review       int                `json:"review"`
	Pubdate      int                `json:"pubdate"`
	Senddate     int                `json:"senddate"`
	Duration     string             `json:"duration"`
	Badgepay     bool               `json:"badgepay"`
	HitColumns   []string           `json:"hit_columns"`
	ViewType     string             `json:"view_type"`
	IsPay        int                `json:"is_pay"`
	IsUnionVideo int                `json:"is_union_video"`
	RecTags      interface{}        `json:"rec_tags"`
	NewRecTags   []*SearchNewRecTag `json:"new_rec_tags"`
	RankScore    int64              `json:"rank_score"`
	Like         int64              `json:"like"`
	Upic         string             `json:"upic"`
	// special_card
	Corner    string `json:"corner"`
	Cover     string `json:"cover"`
	Desc      string `json:"desc"`
	URL       string `json:"url"`
	RecReason string `json:"rec_reason"`
	Danmaku   int32  `json:"danmaku"`
}

// Fill fill search video data.
func (v *SearchVideo) Fill(arc *arcmdl.Arc) {
	if arc == nil {
		return
	}
	v.Upic = arc.Author.Face
	v.Play = arc.Stat.View
	v.Pubdate = int(arc.PubDate)
	v.Danmaku = arc.Stat.Danmaku
}
