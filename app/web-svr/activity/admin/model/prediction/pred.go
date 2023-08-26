package prediction

import (
	"time"
)

// BatchAdd .
type BatchAdd struct {
	Sid   int64  `json:"sid"`
	Min   int    `json:"min"`
	Max   int    `json:"max"`
	Name  string `json:"name" `
	Pid   int64  `json:"pid"`
	Type  int    `json:"type"`
	State int    `json:"state"`
}

// PredSearch .
type PredSearch struct {
	ID    int64 `form:"id"`
	Sid   int64 `form:"sid" validate:"min=1"`
	Pid   int64 `form:"pid" default:"-1"`
	Type  int   `form:"type" default:"-1"`
	State int   `form:"state" default:"-1"`
	Pn    int   `form:"pn" default:"0" validate:"min=0"`
	Ps    int   `form:"ps" default:"15" validate:"min=1,max=100"`
}

// PresUp .
type PresUp struct {
	ID    int64  `form:"id" validate:"min=1"`
	Min   int    `form:"min" validate:"min=0"`
	Max   int    `form:"max" validate:"min=0"`
	Name  string `form:"name" validate:"required"`
	Type  int    `form:"type" validate:"min=0"`
	State int    `form:"state" validate:"min=0"`
}

// ItemAdd .
type ItemAdd struct {
	Sid   int64  `form:"sid" validate:"min=1"`
	Pid   int64  `form:"pid" validate:"min=1"`
	Desc  string `form:"desc"`
	Image string `form:"image"`
	State int    `form:"state" default:"1"`
}

// ItemUp .
type ItemUp struct {
	ID    int64  `form:"id" validate:"min=1"`
	Desc  string `form:"desc"`
	Image string `form:"image"`
	State int    `form:"state" validate:"min=0"`
}

// Prediction .
type Prediction struct {
	ID    int64     `json:"id"  gorm:"column:id"`
	Sid   int64     `json:"sid"  gorm:"column:sid"`
	Pid   int64     `json:"pid"  gorm:"column:pid"`
	Min   int       `json:"min"  gorm:"column:min"`
	Max   int       `json:"max"  gorm:"column:max"`
	Type  int       `json:"type"  gorm:"column:type"`
	Name  string    `json:"name"  gorm:"column:name"`
	State int       `json:"state"  gorm:"column:state"`
	Ctime time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// SearchRes .
type SearchRes struct {
	List []*Prediction `json:"list"`
	Pages
}

// Pages .
type Pages struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// TableName Prediction def
func (Prediction) TableName() string {
	return "prediction"
}

// PredItem .
type PredItem struct {
	ID    int64     `json:"id"  gorm:"column:id"`
	Sid   int64     `json:"sid"  gorm:"column:sid"`
	Pid   int64     `json:"pid"  gorm:"column:pid"`
	Desc  string    `json:"desc" gorm:"column:desc"`
	Image string    `json:"image" gorm:"column:image"`
	State int       `json:"state"  gorm:"column:state"`
	Ctime time.Time `json:"ctime"  gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime"  gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// TableName PredItem def
func (PredItem) TableName() string {
	return "prediction_item"
}

// ItemSearch .
type ItemSearch struct {
	ID    int64 `form:"id"`
	Pid   int64 `form:"pid" validate:"required,min=1"`
	State int   `form:"state" default:"-1"`
	Pn    int   `form:"pn" default:"0" validate:"min=0"`
	Ps    int   `form:"ps" default:"15" validate:"min=1,max=1000"`
}

// ItemSearchRes .
type ItemSearchRes struct {
	List []*PredItem `json:"list"`
	Pages
}
