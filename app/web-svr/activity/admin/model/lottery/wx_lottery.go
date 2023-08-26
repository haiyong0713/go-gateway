package lottery

import (
	"fmt"

	"go-common/library/time"
)

type WxLotteryLog struct {
	Mid         int64     `json:"mid" gorm:"column:mid"`
	Uname       string    `json:"uname" gorm:"-"`
	ActName     string    `json:"act_name" gorm:"-"`
	GiftID      int64     `json:"gift_id" gorm:"gift_id"`
	GiftType    int64     `json:"gift_type" gorm:"gift_type"`
	GiftName    string    `json:"gift_name" gorm:"gift_name"`
	GiftCount   int64     `json:"gift_count" gorm:"-"`
	GiftStatus  int64     `json:"gift_status" gorm:"-"`
	OrderStatus int64     `json:"-" gorm:"order_status"`
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
}

func (t WxLotteryLog) TableName() string {
	return fmt.Sprintf("wx_lottery_log_%02d", t.Mid%100)
}
