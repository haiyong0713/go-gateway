package timemachine

import (
	xtime "go-common/library/time"
)

// bcurd -dsn='main_lottery:RWXcNcw1X43S1K1nvB51x7iNjMMjP0ba@tcp(10.221.34.182:4000)/main_lottery?parseTime=true'  -schema=main_lottery -table=user_report_2020_user_info -tmpl=bilibili_log.tmpl > user_info.go

// UserInfo represents a row from 'user_report_2020_user_info'.
type UserInfo struct {
	Mid       int64      `json:"mid"`        // 用户id
	Aid       int64      `json:"aid"`        // 用户发布的稿件id
	IsNew     int8       `json:"is_new"`     // 是否是新up主，0未知，1是，2不是
	LotteryID string     `json:"lottery_id"` // 抽奖id
	Ctime     xtime.Time `json:"ctime"`      // 创建时间
	Mtime     xtime.Time `json:"mtime"`      // 更新时间
}
