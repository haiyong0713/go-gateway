package cards

import (
	"go-common/library/time"
)

// Cards ...
type Cards struct {
	ID        int64     `json:"id" gorm:"column:id"`
	Name      string    `json:"name" gorm:"column:name"`
	SID       string    `json:"sid" gorm:"column:sid"`
	LotteryID int64     `json:"lottery_id" gorm:"column:lottery_id"`
	ReserveID int64     `json:"reserve_id" gorm:"column:reserve_id"`
	CardsNum  int64     `json:"cards_num" gorm:"column:cards_num"`
	Cards     string    `json:"cards" gorm:"column:cards"`
	Ctime     time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}
