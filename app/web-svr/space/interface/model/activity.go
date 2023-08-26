package model

import (
	"encoding/json"
	xtime "go-common/library/time"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

type AddReserveReq struct {
	Sid      int64  `form:"sid" validate:"min=1"`
	Mid      int64  `form:"mid"`
	From     string `form:"from"`
	Type     string `form:"type"`
	Oid      string `form:"oid"`
	Platform string `form:"platform"`
	Mobiapp  string `form:"mobi_app"`
	Buvid    string `form:"buvid"`
	Spmid    string `form:"spmid"`
}

type UpActReserveRelationInfo struct {
	Sid                int64                                   `json:"sid"`
	Name               string                                  `json:"name"`
	Total              int64                                   `json:"total"`
	Stime              xtime.Time                              `json:"stime"`
	Etime              xtime.Time                              `json:"etime"`
	IsFollow           int64                                   `json:"is_follow"`
	State              actgrpc.UpActReserveRelationState       `json:"state"`
	Oid                string                                  `json:"oid"`
	Type               actgrpc.UpActReserveRelationType        `json:"type"`
	Upmid              int64                                   `json:"up_mid"`
	ReserveRecordCtime xtime.Time                              `json:"reserve_record_ctime"`
	LivePlanStartTime  xtime.Time                              `json:"live_plan_start_time"`
	LotteryType        actgrpc.UpActReserveRelationLotteryType `json:"lottery_type,omitempty"`
	LotteryPrizeInfo   *LotteryPrizeInfo                       `json:"lottery_prize_info,omitempty"`
	ShowTotal          bool                                    `json:"show_total"`
	Subtitle           string                                  `json:"subtitle"`
	AttachedBadgeText  string                                  `json:"attached_badge_text"`
}

func (i *UpActReserveRelationInfo) FromUpActReserveRelationInfo(s *actgrpc.UpActReserveRelationInfo, isOwner bool) {
	i.Sid = s.Sid
	i.Name = s.Title
	i.Stime = s.Stime
	i.Etime = s.Etime
	i.IsFollow = s.IsFollow
	i.State = s.State
	i.Oid = s.Oid
	i.Type = s.Type
	i.Upmid = s.Upmid
	i.ReserveRecordCtime = s.ReserveRecordCtime
	i.LivePlanStartTime = s.LivePlanStartTime
	if isOwner || s.Type != actgrpc.UpActReserveRelationType_Live || s.Total >= s.ReserveTotalShowLimit {
		i.Total = s.Total
		i.ShowTotal = true
	}
	if s.Type == actgrpc.UpActReserveRelationType_ESports {
		i.Subtitle = s.Desc
	}
	i.AttachedBadgeText = constructReserveAttatchedBadgeText(s.Type, s.Ext)
}

func constructReserveAttatchedBadgeText(typ actgrpc.UpActReserveRelationType, ext string) string {
	switch typ {
	case actgrpc.UpActReserveRelationType_Live:
		if ext == "" {
			return ""
		}
		tmp := &actgrpc.UpActReserveRelationInfoExtend{}
		// 大航海
		if err := json.Unmarshal([]byte(ext), &tmp); err == nil && tmp.SubType == 1 {
			return "大航海专属"
		}
	default:
	}
	return ""
}

func (i *UpActReserveRelationInfo) FromUpActReserveLotteryInfo(s *actgrpc.UpActReserveRelationInfo) {
	const (
		_upActReserveRelationLotteryTypeCron = 1 // 预约抽奖类型定时抽奖
		// 预约抽奖预制icon
		_upActReserveRelationLotteryIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/rgHplMQyiX.png"
	)
	switch s.LotteryType {
	case _upActReserveRelationLotteryTypeCron:
		i.LotteryType = s.LotteryType
		if s.PrizeInfo != nil {
			i.LotteryPrizeInfo = &LotteryPrizeInfo{
				Text:        s.PrizeInfo.Text,
				LotteryIcon: _upActReserveRelationLotteryIcon,
				JumpUrl:     s.PrizeInfo.JumpUrl,
			}
		}
	default:
	}
}

type LotteryPrizeInfo struct {
	Text        string `json:"text,omitempty"`
	LotteryIcon string `json:"lottery_icon,omitempty"`
	JumpUrl     string `json:"jump_url,omitempty"`
}
