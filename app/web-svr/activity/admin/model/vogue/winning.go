package vogue

import "go-common/library/time"

type WinningItem struct {
	Id           int       `json:"-" gorm:"column:id"`
	Uid          int       `json:"uid" gorm:"column:uid"`
	UName        string    `json:"uname" gorm:"column:uname"`
	GoodsName    string    `json:"goods_name" gorm:"column:goods_name"`
	WinningTime  time.Time `json:"winning_time" gorm:"column:winning_time"`
	GoodsAddress string    `json:"goods_address" gorm:"column:goods_address"`
	GoodsType    int       `json:"goods_type" gorm:"-"`
	HasError     string    `json:"has_error" gorm:"column:has_error"`
}

type WinningList struct {
	List []*WinningItem `json:"list"`
	Page Page           `json:"page"`
}
