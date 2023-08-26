package model

import (
	"encoding/json"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"

	"go-common/library/log"
)

// Suggest3 struct
type Suggest3 struct {
	Code    int    `json:"code"`
	TrackID string `json:"trackid"`
	ExpStr  string `json:"exp_str"`
	Result  []*Sug `json:"result"`
}

// Sug struct
type Sug struct {
	ShowName  string          `json:"show_name,omitempty"`
	Term      string          `json:"term,omitempty"`
	Ref       int64           `json:"ref,omitempty"`
	TermType  int             `json:"term_type,omitempty"`
	SubType   string          `json:"sub_type,omitempty"`
	Pos       int             `json:"pos,omitempty"`
	Cover     string          `json:"cover,omitempty"`
	CoverSize float64         `json:"cover_size,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	PGC       *SugPGC         `json:"-"`
	User      *SugUser        `json:"-"`
}

// SugChange chagne sug value
func (s *Sug) SugChange() {
	var err error
	switch s.TermType {
	case SuggestionJumpPGC:
		err = json.Unmarshal(s.Value, &s.PGC)
	case SuggestionJumpUser:
		err = json.Unmarshal(s.Value, &s.User)
	}
	if err != nil {
		log.Error("SugChange json.Unmarshal(%s) error(%+v)", s.Value, err)
	}
}

// SugPGC fro sug
type SugPGC struct {
	MediaID        int64                `json:"media_id,omitempty"`
	SeasonID       int64                `json:"season_id,omitempty"`
	Title          string               `json:"title,omitempty"`
	MediaType      int                  `json:"media_type,omitempty"`
	GotoURL        string               `json:"goto_url,omitempty"`
	Areas          string               `json:"areas,omitempty"`
	Pubtime        xtime.Time           `json:"pubtime,omitempty"`
	FixPubTime     string               `json:"fix_pubtime_str,omitempty"`
	Styles         string               `json:"styles,omitempty"`
	CV             string               `json:"cv,omitempty"`
	Staff          string               `json:"staff,omitempty"`
	MediaScore     float64              `json:"media_score,omitempty"`
	MediaUserCount int                  `json:"media_user_cnt,omitempty"`
	Cover          string               `json:"cover,omitempty"`
	Badges         []*model.ReasonStyle `json:"badges,omitempty"`
}

// SugUser fro sug
type SugUser struct {
	Mid                int64  `json:"uid,omitempty"`
	Face               string `json:"face,omitempty"`
	Name               string `json:"uname,omitempty"`
	Fans               int    `json:"fans,omitempty"`
	Videos             int    `json:"videos,omitempty"`
	Level              int    `json:"level,omitempty"`
	OfficialVerifyType int    `json:"verify_type,omitempty"`
}

// SearchSuggestReq is
type SearchSuggestReq struct {
	Mid       int64     `form:"-"`
	Platform  string    `form:"platform"`
	Buvid     string    `form:"-"`
	Term      string    `form:"-"`
	Device    string    `form:"device"`
	Build     int64     `form:"build"`
	Highlight int64     `form:"-"`
	MobiApp   string    `form:"mobi_app"`
	Now       time.Time `form:"-"`
}

const (
	SuggestionAccount = 4

	SuggestionJumpUser = 81
	SuggestionJumpPGC  = 82
)
