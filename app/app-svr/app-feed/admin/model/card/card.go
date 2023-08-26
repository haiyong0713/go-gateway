package card

import (
	"go-common/library/time"
)

type ResourceCard struct {
	Id           int64     `gorm:"column:id" json:"id"`
	CardType     string    `gorm:"column:card_type" json:"card_type"`
	CardSizeType int32     `gorm:"column:card_size_type" json:"card_size_type"`
	Title        string    `gorm:"column:title" json:"title"`
	Desc         string    `gorm:"column:desc" json:"desc"`
	Width        int32     `gorm:"column:width" json:"width"`
	Height       int32     `gorm:"column:height" json:"height"`
	Cover        string    `gorm:"column:cover" json:"cover"`
	Corner       string    `gorm:"column:corner" json:"corner"`
	Button       string    `gorm:"column:button" json:"button"`
	JumpInfo     string    `gorm:"column:jump_info" json:"jump_info"`
	ExtraInfo    string    `gorm:"column:extra_info" json:"extra_info"`
	CUname       string    `gorm:"column:c_uname" json:"c_uname"`
	MUname       string    `gorm:"column:m_uname" json:"m_uname"`
	Ctime        time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime        time.Time `gorm:"column:mtime" json:"mtime"`
	Deleted      int32     `gorm:"column:deleted" json:"deleted"`
}

func (card *ResourceCard) TableName() string {
	return "resource_card"
}
