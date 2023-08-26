package search

import (
	"encoding/json"

	seasonmdl "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
)

// SearchPGCRes .
type SearchPGCRes struct {
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
	Result         []*SearchSeason `json:"result,omitempty"`
	ShowColumn     int             `json:"show_column"`
	InBlackKey     int8            `json:"in_black_key"`
	InWhiteKey     int8            `json:"in_white_key"`
}

// SearchSeason search all result season.
type SearchSeason struct {
	Type       string   `json:"type"`
	MediaID    int64    `json:"media_id"`
	Title      string   `json:"title"`
	OrgTitle   string   `json:"org_title"`
	MediaType  int      `json:"media_type"`
	Cv         string   `json:"cv"`
	Staff      string   `json:"staff"`
	SeasonID   int32    `json:"season_id"`
	IsAvid     bool     `json:"is_avid"`
	HitColumns []string `json:"hit_columns"`
	HitEpids   string   `json:"hit_epids"`
	// 由于历史原因，只能沿用内嵌，如果要新加字段，请不要往这里面添加
	*SeasonInfo
}

type SeasonInfo struct {
	SeasonType     int32                         `json:"season_type"`
	SeasonTypeName string                        `json:"season_type_name"`
	SelectionStyle string                        `json:"selection_style"`
	EpSize         int32                         `json:"ep_size"`
	URL            string                        `json:"url"`
	ButtonText     string                        `json:"button_text"`
	IsFollow       int32                         `json:"is_follow"`
	IsSelection    int32                         `json:"is_selection"`
	Eps            []*searchEp                   `json:"eps"`
	Badges         []*seasonmdl.SearchBadgeProto `json:"badges"`
	Cover          string                        `json:"cover"`
	Areas          string                        `json:"areas"`
	Styles         string                        `json:"styles"`
	GotoURL        string                        `json:"goto_url"`
	Desc           string                        `json:"desc"`
	Pubtime        int64                         `json:"pubtime"`
	MediaMode      int32                         `json:"media_mode"`
	FixPubtimeStr  string                        `json:"fix_pubtime_str"`
	MediaScore     *MediaScore                   `json:"media_score"`
	DisplayInfo    []*seasonmdl.SearchBadgeProto `json:"display_info"`
	PGCSeasonID    int32                         `json:"pgc_season_id"`
	Corner         int32                         `json:"corner"`
	IndexShow      string                        `json:"index_show"`
}

type MediaScore struct {
	Score     float32 `json:"score"`
	UserCount int32   `json:"user_count"`
}

type searchEp struct {
	ID          int32                         `json:"id"`
	Cover       string                        `json:"cover"`
	Title       string                        `json:"title"`
	URL         string                        `json:"url"`
	ReleaseDate string                        `json:"release_date"`
	Badges      []*seasonmdl.SearchBadgeProto `json:"badges"`
	IndexTitle  string                        `json:"index_title"`
	LongTitle   string                        `json:"long_title"`
}

// Fill fill search season data.
func (season *SearchSeason) Fill(card *seasonmdl.SearchCardProto) {
	var eps []*searchEp
	for _, v := range card.Eps {
		ep := &searchEp{
			ID:          v.Id,
			Cover:       v.Cover,
			Title:       v.Title,
			URL:         v.Url,
			ReleaseDate: v.ReleaseDate,
			IndexTitle:  v.IndexTitle,
			LongTitle:   v.LongTitle,
			Badges:      v.Badges,
		}
		eps = append(eps, ep)
	}
	season.SeasonInfo = &SeasonInfo{
		Cover:  card.SeasonCover,
		Areas:  card.Areas,
		Styles: card.Style,
		MediaScore: &MediaScore{
			Score:     card.GetRating().GetScore(),
			UserCount: card.GetRating().GetCount(),
		},
		GotoURL:        card.Url,
		Desc:           card.Evaluate,
		Pubtime:        card.GetReleaseDate().GetSeconds(),
		MediaMode:      card.Mode,
		FixPubtimeStr:  card.ReleaseDateShow,
		DisplayInfo:    card.Badges,
		Badges:         card.Badges,
		PGCSeasonID:    card.SeasonId,
		SeasonType:     card.SeasonType,
		SeasonTypeName: card.SeasonTypeName,
		SelectionStyle: card.SelectionStyle,
		EpSize:         card.EpSize,
		URL:            card.Url,
		ButtonText:     card.ButtonText,
		IsFollow:       card.IsFollow,
		IsSelection:    card.IsSelection,
		Corner:         card.Status,
		Eps:            eps,
		IndexShow:      card.IndexShow,
	}
}
