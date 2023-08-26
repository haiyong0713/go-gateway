package like

import (
	"time"
)

type Reserve struct {
	ID       int64  `json:"id"`
	Mid      int64  `json:"mid"`
	State    int64  `json:"state"`
	Num      int64  `json:"num"`
	Sid      int64  `json:"sid"`
	Platform string `json:"platform"`
	From     string `json:"from"`
	Name     string `json:"name"`
}

type ProgressNotifyMessage struct {
	Sid     int64       `json:"sid"`
	Mid     int64       `json:"mid"`
	RuleID  int64       `json:"rule_id"`
	Num     int64       `json:"num"`
	MTime   time.Time   `json:"mtime"`
	Subject *ActSubject `json:"subject"`
}

type ReserveTunnel struct {
	ID       int64     `json:"id"`
	Mid      int64     `json:"mid"`
	State    int64     `json:"state"`
	Num      int64     `json:"num"`
	Sid      int64     `json:"sid"`
	Platform string    `json:"platform"`
	From     string    `json:"from"`
	Ctime    time.Time `json:"ctime"`
}

type UpActReserveRelationMonitor struct {
	Action      string                `json:"action"`
	Old         *UpActReserveRelation `json:"old"`
	New         *UpActReserveRelation `json:"new"`
	TimeVersion int64                 `json:"time_version"`
}

const (
	NotifyMessageTypeReserve        = 3 // 私信通知卡
	NotifyMessageTypeResetReserve   = 4 // 私信撤销通知卡
	NotifyMessageTypeLotteryReserve = 5 // 私信抽奖预约
	NotifyMessageTypePushUpVerify14 = 6 // 私信提醒up主核销预约-14天时
	NotifyMessageTypePushUpVerify30 = 7 // 私信提醒up主核销预约-30天时
)

type LotteryReserveNotify struct {
	BizID        int64   `json:"biz_id"`
	UniqueID     int64   `json:"unique_id"`
	CardUniqueID int64   `json:"card_unique_id"`
	Mids         []int64 `json:"mids"`
	State        int64   `json:"state"`
	Timestamp    int64   `json:"timestamp"`
}
