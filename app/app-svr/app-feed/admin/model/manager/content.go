package manager

import "encoding/json"

type ContentCard struct {
	ID         int64           `gorm:"column:id" json:"id"`
	Title      string          `gorm:"column:title" json:"title"`
	Cover      string          `gorm:"column:cover" json:"cover"`
	ContentStr json.RawMessage `gorm:"column:content" json:"-"`
	Weight     int64           `gorm:"column:weight" json:"weight"`
	State      int64           `gorm:"column:state" json:"state"`
	ReType     int64           `gorm:"column:re_type" json:"re_type"`
	ReValue    string          `gorm:"column:re_value" json:"re_value"`
	BtnReType  int64           `gorm:"-" json:"btn_re_type"`
	BtnReValue string          `gorm:"-" json:"btn_re_value"`
	Contents   []*Contents     `gorm:"-" json:"contents"`
}

type Contents struct {
	Ctype  string `json:"ctype,omitempty"`
	Cvalue string `json:"cvalue,omitempty"`
}

// TableName .
func (a ContentCard) TableName() string {
	return "content_card"
}
