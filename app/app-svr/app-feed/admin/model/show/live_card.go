package show

import "time"

type PopLiveCardRes struct {
	Items []*PopLiveCard `json:"items"`
	Pager PagerCfg       `json:"pager"`
}

type PopLiveCard struct {
	ID       int64     `json:"id" form:"id"`
	CardType string    `json:"card_type" form:"-"`
	RID      int64     `json:"rid" form:"rid" gorm:"column:rid" validate:"required"`
	CreateBy string    `json:"create_by" form:"-" gorm:"column:create_by"`
	Cover    string    `json:"cover" form:"cover" gorm:"column:cover"`
	State    int       `json:"state" form:"state"`
	Mtime    time.Time `json:"-" form:"-"`
	MtimeStr string    `json:"mtime" form:"-"`
}

type PopLiveCardAD struct {
	ID       int64  `json:"id" form:"id"`
	CardType string `json:"card_type" form:"-"`
	RID      int64  `json:"rid" form:"rid" gorm:"column:rid" validate:"required"`
	CreateBy string `json:"create_by" form:"-" gorm:"column:create_by"`
	Cover    string `json:"cover" form:"cover" gorm:"column:cover"`
	State    int    `json:"state" form:"state"`
}

type PopLiveCardUP struct {
	ID       int64  `form:"id" validate:"required"`
	CardType string `json:"card_type" form:"-"`
	RID      int64  `json:"rid" form:"rid" gorm:"column:rid"`
	CreateBy string `json:"create_by" form:"-" gorm:"column:create_by"`
	Cover    string `json:"cover" form:"cover" gorm:"column:cover"`
	State    int    `json:"state" form:"state"`
}

// TableName .
func (a PopLiveCard) TableName() string {
	return "popular_live_card"
}

// TableName .
func (a PopLiveCardAD) TableName() string {
	return "popular_live_card"
}

// TableName .
func (a PopLiveCardUP) TableName() string {
	return "popular_live_card"
}
