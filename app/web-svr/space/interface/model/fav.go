package model

import (
	xtime "go-common/library/time"

	arc "git.bilibili.co/bapis/bapis-go/archive/service"
)

// FavNav fav nav struct.
type FavNav struct {
	Archive   []*VideoFolder `json:"archive"`
	Arc       int64          `json:"arc"`
	Playlist  int64          `json:"playlist"`
	Topic     int64          `json:"topic"`
	Article   int64          `json:"article"`
	Album     int            `json:"album"`
	Movie     int            `json:"movie"`
	Pugv      int64          `json:"pugv"`
	Note      int64          `json:"note"`
	TopicList int64          `json:"topic_list"` // 新话题
}

// FavArcArg .
type FavArcArg struct {
	Vmid    int64  `form:"vmid" validate:"min=1"`
	Fid     int64  `form:"fid" validate:"min=-1"`
	Tid     int64  `form:"tid"`
	Keyword string `form:"keyword"`
	Order   string `form:"order"`
	Pn      int    `form:"pn" default:"1" validate:"min=1"`
	Ps      int    `form:"ps" default:"30" validate:"min=1"`
}

type FavUpper struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
}

type FavCntInfo struct {
	Collect int64 `json:"collect"`
	Play    int64 `json:"play"`
}

type FavInfo struct {
	ID         int64       `json:"id"`
	SeasonType int64       `json:"season_type"`
	Title      string      `json:"title"`
	Cover      string      `json:"cover"`
	Upper      *FavUpper   `json:"upper"`
	CntInfo    *FavCntInfo `json:"cnt_info"`
	MediaCount int64       `json:"media_count"`
}

type FavMedia struct {
	ID       int64       `json:"id"`
	Title    string      `json:"title"`
	Cover    string      `json:"cover"`
	Duration int64       `json:"duration"`
	Pubtime  int64       `json:"pubtime"`
	Bvid     string      `json:"bvid"`
	Upper    *FavUpper   `json:"upper"`
	CntInfo  *FavCntInfo `json:"cnt_info"`
}

type FavSeasonList struct {
	Info   *FavInfo    `json:"info"`
	Medias []*FavMedia `json:"medias"`
}

type SearchArchive struct {
	Code           int    `json:"code,omitempty"`
	Seid           string `json:"seid"`
	Page           int    `json:"page"`
	PageSize       int    `json:"pagesize"`
	NumPages       int    `json:"numPages,omitempty"`
	PageCount      int    `json:"pagecount"`
	NumResults     int    `json:"numResults,omitempty"`
	Total          int    `json:"total"`
	SuggestKeyword string `json:"suggest_keyword"`
	Mid            int64  `json:"mid"`
	Fid            int64  `json:"fid"`
	Tid            int    `json:"tid"`
	Order          string `json:"order"`
	Keyword        string `json:"keyword"`
	TList          []struct {
		Tid   int    `json:"tid"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	} `json:"tlist,omitempty"`
	Result   []*SearchArchiveResult `json:"result,omitempty"`
	Archives []*FavArchive          `json:"archives"`
}

type SearchArchiveResult struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Play    string `json:"play"`
	FavTime int64  `json:"fav_time"`
}

type FavArchive struct {
	*arc.Arc
	FavAt          int64  `json:"fav_at"`
	PlayNum        string `json:"play_num"`
	HighlightTitle string `json:"highlight_title"`
}

type VideoFolder struct {
	MediaId    int64      `json:"media_id"`
	Fid        int64      `json:"fid"`
	Mid        int64      `json:"mid"`
	Name       string     `json:"name"`
	MaxCount   int        `json:"max_count"`
	CurCount   int        `json:"cur_count"`
	AttenCount int        `json:"atten_count"`
	Favoured   int8       `json:"favoured"`
	State      int8       `json:"state"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
	Cover      []*Cover   `json:"cover,omitempty"`
}

type Cover struct {
	Aid  int64  `json:"aid"`
	Pic  string `json:"pic"`
	Type int32  `json:"type"`
}
