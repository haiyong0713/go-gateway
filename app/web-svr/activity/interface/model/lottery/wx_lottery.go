package lottery

import (
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
	lotteryV2 "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
)

const (
	GiftTypeReal      = 1
	GiftTypeVipCoupon = 2
	GiftTypePendant   = 3
	GiftTypeMoney     = 4
	_lotteryFromWx    = 1 //小程序来源，1微信
	_lotteryFromQQ    = 2 //小程序来源，2qq
)

type WxLotteryLog struct {
	ID        int64      `json:"id"`
	Mid       int64      `json:"mid"`
	Buvid     string     `json:"buvid"`
	LotteryID string     `json:"lottery_id"`
	GiftType  int64      `json:"gift_type"`
	GiftID    int64      `json:"gift_id"`
	GiftName  string     `json:"gift_name"`
	GiftMoney int64      `json:"gift_money"`
	Ctime     xtime.Time `json:"ctime"`
	Mtime     xtime.Time `json:"mtime"`
}

type WxLotteryHis struct {
	ID    int64      `json:"id"`
	Mid   int64      `json:"mid"`
	Buvid string     `json:"buvid"`
	Ctime xtime.Time `json:"ctime"`
	Mtime xtime.Time `json:"mtime"`
}

type WxAwardRes struct {
	NotLottery int    `json:"not_lottery"`
	PlayWindow int    `json:"play_window"`
	Mid        int64  `json:"mid"`
	GiftID     int64  `json:"gift_id"`
	GiftName   string `json:"gift_name"`
	ImgURL     string `json:"img_url"`
	Type       int64  `json:"type"`
	JumpURL    string `json:"jump_url"`
}

type WxPlayWindowRes struct {
}

type WxLotteryRes struct {
	Mid       int64  `json:"mid"`
	GiftID    int64  `json:"gift_id"`
	GiftName  string `json:"gift_name"`
	ImgURL    string `json:"img_url"`
	Type      int64  `json:"type"`
	JumpURL   string `json:"jump_url"`
	IsNew     int    `json:"is_new"`
	LotteryID string `json:"-"`
	Money     int64  `json:"-"`
	UserType  int64  `json:"-"`
}

type WxLotteryGiftRes struct {
	List []*WxLotteryGift `json:"list"`
}

type WxLotteryGift struct {
	ID   string `json:"id"`
	Vid  string `json:"vid"`
	Name string `json:"name"`
	Data *struct {
		Name  string `json:"name"`
		Image string `json:"image"`
		ID    string `json:"id"`
	} `json:"data"`
}

func (out *WxLotteryRes) FromRecordDetail(in *LotteryRecordDetail, sid string, userType int64, moneyMap map[string]int64) {
	out.Mid = in.Mid
	out.LotteryID = sid
	out.UserType = userType
	if in.GiftID > 0 {
		out.GiftID = in.GiftID
		out.GiftName = strings.TrimSpace(in.GiftName)
		out.ImgURL = in.ImgURL
		var outType int64
		switch in.Type {
		case 1: // 实物奖品
			outType = GiftTypeReal
		case 6: // 大会员券
			outType = GiftTypeVipCoupon
		case 3: // 头像挂件
			outType = GiftTypePendant
		case 7: // 虚拟奖励(现金)
			if money, ok := moneyMap[out.GiftName]; ok {
				out.Money = money
				outType = GiftTypeMoney
			}
		default:
			outType = 0
		}
		if outType == 0 {
			log.Warn("FromRecordDetail sid:%s giftID:%d type(%d) not support", sid, in.GiftID, in.Type)
		}
		out.Type = outType
	}
}

func HandleRecordDetail(in *lotteryV2.RecordDetail, sid string, userType int64, moneyMap map[string]int64) *WxLotteryRes {
	out := new(WxLotteryRes)
	out.Mid = in.Mid
	out.LotteryID = sid
	out.UserType = userType
	if in.GiftID > 0 {
		out.GiftID = in.GiftID
		out.GiftName = strings.TrimSpace(in.GiftName)
		out.ImgURL = in.ImgURL
		var outType int64
		switch in.Type {
		case 1: // 实物奖品
			outType = GiftTypeReal
		case 6: // 大会员券
			outType = GiftTypeVipCoupon
		case 3: // 头像挂件
			outType = GiftTypePendant
		case 10: // 虚拟奖励(现金)
			if money, ok := moneyMap[out.GiftName]; ok {
				out.Money = money
				outType = GiftTypeMoney
			}
		default:
			outType = 0
		}
		if outType == 0 {
			log.Warn("HandleRecordDetail sid:%s giftID:%d type(%d) not support", sid, in.GiftID, in.Type)
		}
		out.Type = outType
	}
	return out
}

func HandleLotteryFrom(referer string) int64 {
	if referer == "" {
		return 0
	}

	if ok := strings.Contains(referer, "servicewechat.com"); ok {
		return _lotteryFromWx
	}

	if ok := strings.Contains(referer, "appservice.qq.com"); ok {
		return _lotteryFromQQ
	}
	return 0
}
