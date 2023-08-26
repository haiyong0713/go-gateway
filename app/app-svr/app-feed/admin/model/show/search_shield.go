package show

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// SearchShield search web
type SearchShield struct {
	ID        int64                `json:"id" form:"id"`
	CardType  int                  `json:"card_type" form:"card_type"`
	CardValue string               `json:"card_value" form:"card_value"`
	Check     int                  `json:"check" form:"check"`
	Person    string               `json:"person" form:"person"`
	Reason    string               `json:"reason" form:"reason"`
	Query     []*SearchShieldQuery `json:"query" form:"query" gorm:"-"`
	Title     string               `json:"title"`
	Season    interface{}          `json:"season"`
	Mtime     xtime.Time           `json:"mtime"`
	BvID      string               `json:"bvid" gorm:"-"`
}

// SearchShieldPager .
type SearchShieldPager struct {
	Item []*SearchShield `json:"item"`
	Page common.Page     `json:"page"`
}

// SearchShieldValid
type SearchShieldValid struct {
	ID        int64
	Query     string
	CardValue string
	CardType  int
}

// TableName .
func (a SearchShield) TableName() string {
	return "search_shield"
}

/*
---------------------------
 struct param
---------------------------
*/

// SearchShieldAP add param
type SearchShieldAP struct {
	ID        int64  `json:"id" form:"id"`
	CardType  int    `json:"card_type" form:"card_type" validate:"required"`
	CardValue string `json:"card_value" form:"card_value" validate:"required"`
	Person    string `json:"person" form:"person"`
	Reason    string `json:"reason" form:"reason"`
	Query     string `json:"query" form:"query" gorm:"-" validate:"required"`
}

// SearchShieldUP update param
type SearchShieldUP struct {
	ID        int64  `form:"id" validate:"required"`
	CardType  int    `json:"card_type" form:"card_type" validate:"required"`
	CardValue string `json:"card_value" form:"card_value" validate:"required"`
	Person    string `json:"person" form:"person"`
	Reason    string `json:"reason" form:"reason"`
	Query     string `json:"query" form:"query" gorm:"-" validate:"required"`
}

// SearchShieldLP list param
type SearchShieldLP struct {
	ID       int64  `form:"-"`
	IDNew    string `form:"id"`
	Check    int    `form:"check"`
	Query    string `form:"query"`
	Ps       int    `form:"ps" default:"20"`
	Pn       int    `form:"pn" default:"1"`
	CardType int    `form:"card_type"`
}

// SearchShieldOption option web card
type SearchShieldOption struct {
	ID    int64  `form:"id" validate:"required"`
	Check int    `form:"check" validate:"required"`
	Name  string `gorm:"-"`
	UID   int64  `gorm:"-"`
}

// TableName .
func (a SearchShieldOption) TableName() string {
	return "search_shield"
}

// TableName .
func (a SearchShieldAP) TableName() string {
	return "search_shield"
}

// TableName .
func (a SearchShieldUP) TableName() string {
	return "search_shield"
}
