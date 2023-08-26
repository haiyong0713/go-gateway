package show

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// SearchWeb search web
type SearchWeb struct {
	ID          int64             `json:"id" form:"id"`
	CardType    int               `json:"card_type" form:"card_type"`
	CardValue   string            `json:"card_value" form:"card_value"`
	Stime       xtime.Time        `json:"stime" form:"stime"`
	Etime       xtime.Time        `json:"etime" form:"etime"`
	Check       int               `json:"check" form:"check"`
	Status      int               `json:"status" form:"status"`
	Priority    int               `json:"priority" form:"priority"`
	Person      string            `json:"person" form:"person"`
	ApplyReason string            `json:"apply_reason" form:"apply_reason"`
	RecReason   string            `json:"rec_reason" form:"rec_reason"`
	Deleted     int               `json:"deleted" form:"deleted"`
	Query       []*SearchWebQuery `json:"query" form:"query" gorm:"-"`
	PlatVer     []*SearchWebPlat  `json:"plat_ver" form:"plat_ver" gorm:"-"`
	Card        interface{}       `json:"card" gorm:"-"`
}

type SearchWebPlat struct {
	ID         int64  `gorm:"column:id" json:"-" form:"-"`
	SId        int64  `gorm:"column:sid" json:"-" form:"-"`
	Plat       int    `gorm:"column:plat" json:"platforms" form:"platforms"`
	Conditions string `gorm:"column:conditions" json:"conditions" form:"conditions"`
	Build      string `gorm:"column:build" json:"values" form:"values"`
	Deleted    int    `gorm:"column:deleted" json:"-" form:"-"`
}

type OpenSearchWebPlat struct {
	Plat       int    `json:"platforms"`
	Conditions string `json:"conditions"`
	Build      string `json:"build"`
}

// SearchWebPager .
type SearchWebPager struct {
	Item []*SearchWeb `json:"item"`
	Page common.Page  `json:"page"`
}

// OpenSearchWeb search web
type OpenSearchWeb struct {
	ID          int64                `json:"id" form:"id"`
	CardType    int                  `json:"card_type" form:"card_type"`
	CardValue   string               `json:"card_value" form:"card_value"`
	Stime       xtime.Time           `json:"stime" form:"stime"`
	Etime       xtime.Time           `json:"etime" form:"etime"`
	Check       int                  `json:"check" form:"check"`
	Status      int                  `json:"status" form:"status"`
	Priority    int                  `json:"priority" form:"priority"`
	Person      string               `json:"person" form:"person"`
	ApplyReason string               `json:"apply_reason" form:"apply_reason"`
	RecReason   string               `json:"rec_reason" form:"rec_reason"`
	Deleted     int                  `json:"deleted" form:"deleted"`
	Query       []*SearchWebQuery    `json:"query" form:"query" gorm:"-"`
	PlatVer     []*OpenSearchWebPlat `json:"plat_ver" form:"plat_ver" gorm:"-"`
	Card        interface{}          `json:"card" gorm:"-"`
}

// TableName .
func (a SearchWeb) TableName() string {
	return "search_web"
}

func (p SearchWebPlat) TableName() string {
	return "search_web_plat"
}

func (s SearchWeb) Convert() *OpenSearchWeb {
	ret := &OpenSearchWeb{
		ID:          s.ID,
		CardType:    s.CardType,
		CardValue:   s.CardValue,
		Stime:       s.Stime,
		Etime:       s.Etime,
		Check:       s.Check,
		Status:      s.Status,
		Priority:    s.Priority,
		Person:      s.Person,
		ApplyReason: s.ApplyReason,
		RecReason:   s.RecReason,
		Deleted:     s.Deleted,
		Query:       s.Query,
		Card:        s.Card,
		PlatVer:     make([]*OpenSearchWebPlat, len(s.PlatVer)),
	}
	for idx, v := range s.PlatVer {
		ret.PlatVer[idx] = &OpenSearchWebPlat{
			Plat:       v.Plat,
			Conditions: v.Conditions,
			Build:      v.Build,
		}
	}
	return ret
}

/*
---------------------------
 struct param
---------------------------
*/

// SearchWebAP add param
type SearchWebAP struct {
	ID          int64            `json:"id" form:"id"`
	CardType    int              `json:"card_type" form:"card_type" validate:"required"`
	CardValue   string           `json:"card_value" form:"card_value" validate:"required"`
	Stime       xtime.Time       `json:"stime" form:"stime" validate:"required"`
	Etime       xtime.Time       `json:"etime" form:"etime" validate:"required"`
	Priority    int              `json:"priority" form:"priority" validate:"required"`
	Check       int              `form:"check" default:"1"`
	Person      string           `json:"person" form:"person"`
	ApplyReason string           `json:"apply_reason" form:"apply_reason"`
	RecReason   string           `json:"rec_reason" form:"rec_reason" gorm:"rec_reason"`
	PlatVerStr  string           `json:"plat_ver" form:"plat_ver" gorm:"-"`
	Query       string           `json:"query" form:"query" gorm:"-" validate:"required"`
	PlatVer     []*SearchWebPlat `json:"-" form:"-" gorm:"-"`
}

// SearchWebUP update param
type SearchWebUP struct {
	ID          int64            `form:"id" validate:"required"`
	CardType    int              `json:"card_type" form:"card_type"`
	CardValue   string           `json:"card_value" form:"card_value"`
	Stime       xtime.Time       `json:"stime" form:"stime"`
	Etime       xtime.Time       `json:"etime" form:"etime"`
	Check       int              `json:"check" form:"check"`
	Status      int              `json:"status" form:"status"`
	Priority    int              `json:"priority" form:"priority"`
	Person      string           `json:"person" form:"person"`
	RecReason   string           `json:"rec_reason" form:"rec_reason" gorm:"rec_reason"`
	PlatVerStr  string           `json:"plat_ver" form:"plat_ver" gorm:"-"`
	ApplyReason string           `json:"apply_reason" form:"apply_reason"`
	Query       string           `json:"query" form:"query" gorm:"-" validate:"required"`
	PlatVer     []*SearchWebPlat `json:"-" form:"-" gorm:"-"`
}

// SearchWebLP list param
type SearchWebLP struct {
	ID       int    `form:"id"`
	Check    int    `form:"check"`
	Person   string `form:"person"`
	STime    string `form:"stime"`
	ETime    string `form:"etime"`
	Ps       int    `form:"ps" default:"20"`
	Pn       int    `form:"pn" default:"1"`
	CardType int    `form:"card_type"`
	Keyword  string `form:"keyword"`
}

// SearchWebOption option web card (online,hidden,pass,reject)
type SearchWebOption struct {
	ID     int64 `form:"id" validate:"required"`
	Check  int   `json:"check" form:"check"`
	Status int   `json:"status" form:"status"`
}

// SWTimeValid option web card (online,hidden,pass,reject)
type SWTimeValid struct {
	ID        int64
	Query     string
	Priority  int
	STime     xtime.Time
	ETime     xtime.Time
	CardValue string
	CardType  int
	PlatVer   []*SearchWebPlat
	Plat      int
}

// TableName .
func (a SearchWebOption) TableName() string {
	return "search_web"
}

// TableName .
func (a SearchWebAP) TableName() string {
	return "search_web"
}

// TableName .
func (a SearchWebUP) TableName() string {
	return "search_web"
}
