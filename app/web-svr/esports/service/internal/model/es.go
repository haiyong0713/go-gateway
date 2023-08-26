package model

import (
	"go-common/library/database/elastic"
	"time"
)

const (
	_pageFirstNumber   = 1
	SortSTimeDesc      = 1
	SortSTimeAsc       = 0
	AllPlatform        = 3
	_defaultTimeString = "2006-01-02 15:04:05"
)

type ContestQueryResponse struct {
	Page   *Page                      `json:"page"`
	Result []*ContestQueryResultModel `json:"result"`
}

type ContestsQueryParamsModel struct {
	MatchId       int64   `json:"match_id"`
	Gid           int64   `json:"gid"`
	HomeId        int64   `json:"home_id"`
	AwayId        int64   `json:"away_id"`
	Tid           int64   `json:"tid"`
	Stime         string  `json:"stime"`
	Etime         string  `json:"etime"`
	GState        string  `json:"g_state"`
	Sids          []int64 `json:"sids"`
	Forbid        int     `json:"forbid"`
	RoomIds       []int64 `json:"roomIds"`
	Sort          int     `json:"sort"`
	Pn            int     `json:"pn"  validate:"gt=0"`
	Ps            int     `json:"ps"  validate:"gt=0,lte=100"`
	ContestIds    []int64 `json:"contest_ids"`
	GuessType     int     `json:"gs_type"`
	GuessRecT     int64   `json:"guess_rec_t"`
	CursorPage    bool    `json:"cursor_page"`
	Cursor        int64   `json:"cursor"`
	CursorSize    int     `json:"cursor_size"`
	Channel       []int64 `json:"channel"`
	Debug         bool    `json:"debug"`
	ContestStatus int64   `json:"contest_status"`
	NeedInvalid   bool    `json:"need_invalid"`
}

type ContestQueryResultModel struct {
	ID int64 `json:"id"`
}

// Page es page
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

func (model *ContestsQueryParamsModel) ContestQueryPageParams(esReq *elastic.Request) {
	if model.CursorPage {
		// es暂时无法通过游标方式，后续优化
		esReq.Pn(int(model.Cursor + 1)).Ps(model.CursorSize)
	} else {
		esReq.Pn(model.Pn).Ps(model.Ps)
	}
	if model.Sort == SortSTimeDesc {
		esReq.Order("stime", elastic.OrderDesc).Order("id", elastic.OrderAsc)
	} else {
		esReq.Order("stime", elastic.OrderAsc).Order("id", elastic.OrderAsc)
	}
}

func (model *ContestsQueryParamsModel) ContestQueryTeamsParams(esReq *elastic.Request) {
	if model.Tid > 0 {
		esReq.WhereOr("home_id", model.Tid).WhereOr("away_id", model.Tid)
	}
	if model.HomeId > 0 {
		esReq.WhereEq("home_id", model.Tid)
	}
	if model.AwayId > 0 {
		esReq.WhereEq("away_id", model.Tid)
	}
}
func (model *ContestsQueryParamsModel) ContestQueryTimeParams(esReq *elastic.Request) {
	if model.Stime != "" && model.Etime != "" {
		start, _ := time.ParseInLocation(_defaultTimeString, model.Stime, time.Local)
		end, _ := time.ParseInLocation(_defaultTimeString, model.Etime, time.Local)
		esReq.WhereRange("stime", start.Unix(), end.Unix(), elastic.RangeScopeLcRc)
	} else if model.Stime != "" && model.Etime == "" {
		start, _ := time.ParseInLocation(_defaultTimeString, model.Stime, time.Local)
		esReq.WhereRange("stime", start.Unix(), "", elastic.RangeScopeLcRo)
	} else if model.Stime == "" && model.Etime != "" {
		end, _ := time.ParseInLocation(_defaultTimeString, model.Etime, time.Local)
		esReq.WhereRange("stime", "", end.Unix(), elastic.RangeScopeLoRc)
	}
}
func (model *ContestsQueryParamsModel) ContestQueryGuessParams(esReq *elastic.Request) {
	if model.GuessType != 0 {
		esReq.WhereEq("guess_type", model.GuessType)
	}
	if model.GuessRecT > 0 {
		esReq.WhereRange("etime", "", model.GuessRecT, elastic.RangeScopeLcRo)
	}
}

func (model *ContestsQueryParamsModel) ContestQuerySyncPlatformParams(esReq *elastic.Request) {
	if len(model.Channel) > 0 {
		channelList := make([]int64, 0)
		channelList = append(channelList, model.Channel...)
		channelList = append(channelList, AllPlatform)
		esReq.WhereIn("sync_platform", channelList)
	}
}

func (model *ContestsQueryParamsModel) ContestQueryStatusParams(esReq *elastic.Request) {
	if model.NeedInvalid {
		return
	}
	esReq.WhereEq("status", FreezeFalse)
}
