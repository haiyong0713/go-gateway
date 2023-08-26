package model

import (
	"encoding/json"
	"time"
)

type Scope string

const (
	ScopeRange        = Scope("range")
	ScopeRangeContain = Scope("range_contain")
	ScopeEqual        = Scope("equal")
)

type ArcsPassedResult struct {
	Aid       int64               `json:"aid,omitempty"`
	Title     json.RawMessage     `json:"title,omitempty"`
	Content   json.RawMessage     `json:"content,omitempty"`
	TitleItem []string            `json:"title.item,omitempty"`
	Pubtime   string              `json:"pubtime,omitempty"`
	Click     int64               `json:"click,omitempty"`
	Fav       int64               `json:"fav,omitempty"`
	Share     int64               `json:"share,omitempty"`
	Reply     int64               `json:"reply,omitempty"`
	Coin      int64               `json:"coin,omitempty"`
	Dm        int64               `json:"dm,omitempty"`
	Likes     int64               `json:"likes,omitempty"`
	Highlight *ArcPassedHighlight `json:"highlight,omitempty"`
}

type Page struct {
	Num   int64 `json:"num,omitempty"`
	Size  int64 `json:"size,omitempty"`
	Total int64 `json:"total,omitempty"`
}

type ArcPassedSearchReply struct {
	Result []*ArcsPassedResult `json:"result,omitempty"`
	Page   *Page               `json:"page,omitempty"`
}

type ArcPassedHighlight struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type ArcSortBuckets struct {
	Key string `json:"key"`
}

type ArcSortGroupByAid struct {
	Buckets []*ArcSortBuckets `json:"buckets"`
}

type ArcSortGroupByMid struct {
	GroupByAid *ArcSortGroupByAid `json:"group_by_aid"`
	Key        string             `json:"key"`
}

type ArcSortResult struct {
	GroupByMid []*ArcSortGroupByMid `json:"group_by_mid"`
}

type ArcSortReply struct {
	Result *ArcSortResult `json:"result"`
}

type ArcScoreResult struct {
	Score int64 `json:"score,omitempty"`
}

type ArcSearchReply struct {
	Result []*ArcsResult `json:"result"`
	Page   *Page         `json:"page"`
}

type ArcsResult struct {
	Aid       int64               `json:"aid"`
	Highlight *ArcPassedHighlight `json:"highlight"`
}

type ArcCursorAidSearchReply struct {
	Result []*ArcsCursorAidResult `json:"result,omitempty"`
	Page   *Page                  `json:"page,omitempty"`
}

type ArcsCursorAidResult struct {
	Aid   int64 `json:"aid,omitempty"`
	Score int64 `json:"score,omitempty"`
}

type ResultCache struct {
	Reply json.RawMessage `json:"reply"`
	Ctime time.Time       `json:"ctime"`
}
