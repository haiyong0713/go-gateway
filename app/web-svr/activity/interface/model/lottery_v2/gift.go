package lottery

import (
	"context"
	"go-common/library/log"

	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
)

const (
	// Efficient 加入奖池
	Efficient = 1
	// Remark ...
	Remark = "活动抽奖所得"
	// GiftLeastMark 保底奖品
	GiftLeastMark = 1
	// GiftMidNumKey 用户维度上限
	GiftMidNumKey = "mid"
	// IsShow 1
	IsShow = 1
)

// GiftMemberInfo 用户信息
type GiftMemberInfo struct {
	Mid int64
	IP  string
}

// DayNum 日库存
type DayNum map[string]int64

// Gift ...
type Gift struct {
	ID             int64             `json:"id"`
	Sid            string            `json:"sid"`
	Name           string            `json:"name"`
	Num            int64             `json:"num"`
	Type           int               `json:"type"`
	Source         string            `json:"source"`
	ImgURL         string            `json:"img_url"`
	IsShow         int               `json:"is_show"`
	LeastMark      int               `json:"least_mark"`
	MessageTitle   string            `json:"message_title"`
	MessageContent string            `json:"message_content"`
	SendNum        int64             `json:"send_num"`
	DaySendNum     DayNum            `json:"day_send_num"`
	Efficient      int               `json:"efficient"`
	State          int               `json:"state"`
	MemberGroup    []int64           `json:"member_group"`
	TimeLimit      xtime.Time        `json:"time_limit"`
	DayNum         DayNum            `json:"day_num"`
	OtherSendNum   DayNum            `json:"other_send_num"`
	Probability    int64             `json:"probability"`
	Params         string            `json:"params"`
	Extra          map[string]string `json:"extra"`
}

type GiftRes struct {
	ID     int64  `json:"id"`
	Sid    string `json:"sid"`
	Name   string `json:"name"`
	Type   int    `json:"type"`
	ImgURL string `json:"img_url"`
}

// GiftMid ...
type GiftMid struct {
	GiftID   int64      `json:"gift_id"`
	GiftName string     `json:"gift_name"`
	ImgURL   string     `json:"gift_img_url"`
	Mid      int64      `json:"mid"`
	Ctime    xtime.Time `json:"ctime"`
}

// MidWinList ...
type MidWinList struct {
	ID     int64      `json:"id"`
	Mid    int64      `json:"mid"`
	GiftID int64      `json:"gift_id"`
	Cdkey  string     `json:"cdkey"`
	Mtime  xtime.Time `json:"mtime"`
}

// GiftDB db struct
type GiftDB struct {
	ID             int64      `json:"id"`
	Sid            string     `json:"sid"`
	Name           string     `json:"name"`
	Num            int64      `json:"num"`
	Type           int        `json:"type"`
	Source         string     `json:"source"`
	ImgURL         string     `json:"img_url"`
	IsShow         int        `json:"is_show"`
	LeastMark      int        `json:"least_mark"`
	MessageTitle   string     `json:"message_title"`
	MessageContent string     `json:"message_content"`
	SendNum        int64      `json:"send_num"`
	DaySendNum     string     `json:"day_send_num"`
	Efficient      int        `json:"efficient"`
	State          int        `json:"state"`
	MemberGroup    string     `json:"member_group"`
	DayNum         string     `json:"day_num"`
	Params         string     `json:"params"`
	Probability    int64      `json:"probability"`
	Extra          string     `json:"extra"`
	TimeLimit      xtime.Time `json:"time_limit"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
}

// CheckStore check store
func (g *Gift) CheckStore(c context.Context) error {
	if g.SendNum >= g.Num {
		log.Errorc(c, "giftID (%d) send num >= gift num ,sendNum(%d) giftNum (%d)", g.ID, g.SendNum, g.Num)
		return ecode.ActivityLotteryGiftNostoreErr
	}
	if g.DayNum == nil {
		return nil
	}
	for key, v := range g.DayNum {
		if g.DaySendNum != nil {
			if daySendNum, ok := g.DaySendNum[key]; ok {
				if v != 0 && v <= daySendNum {
					log.Errorc(c, "giftID (%d) daySendNum >= giftNum ,daynum(%d) daySendNum (%d)", g.ID, v, daySendNum)
					return ecode.ActivityLotteryGiftNostoreErr
				}
			}
		}

	}

	return nil
}

// CheckSendStore check store
func (g *Gift) CheckSendStore(c context.Context) error {
	if g.SendNum > g.Num {
		return ecode.ActivityLotteryGiftNostoreErr
	}
	if g.DayNum == nil {
		return nil
	}
	for key, v := range g.DayNum {
		if g.DaySendNum != nil {
			if daySendNum, ok := g.DaySendNum[key]; ok {

				if v != 0 && daySendNum > v {
					log.Errorc(c, "giftID (%d) key(%v) sendDayNum> giftDayNum ,sendDayNum(%d) giftDayNum (%d)", g.ID, key, daySendNum, v)
					return ecode.ActivityLotteryGiftNostoreErr
				}
			}
		}
		if g.OtherSendNum != nil {
			if dayOtherNum, ok := g.OtherSendNum[key]; ok {
				if v != 0 && dayOtherNum > v {
					log.Errorc(c, "giftID (%d) key(%v) dayOtherNum > giftDayNum , midSendNum(%d) giftmidNum (%d)", g.ID, key, dayOtherNum, v)

					return ecode.ActivityLotteryGiftNostoreErr
				}
			}
		}

	}
	return nil
}
