package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	xtime "go-common/library/time"
	api "go-gateway/app/app-svr/topic/interface/api"
)

type TopicReportReq struct {
	TopicId int64  `form:"topic_id"`
	Reason  string `form:"reason"`
}

type TopicResReportReq struct {
	TopicId  int64  `form:"topic_id"`
	ResId    int64  `form:"res_id"`
	ResIdStr string `form:"res_id_str"` // 举报资源id的string形式，优先取res_id，如果res_id为0，则转res_id_str为数字使用
	ResType  int64  `form:"res_type"`
	Reason   string `form:"reason"`
}

type TopicLikeReq struct {
	Business string `form:"business" default:"topic"`
	TopicId  int64  `form:"topic_id" validate:"required"`
	UpMid    int64  `form:"up_mid"`
	Action   string `form:"action"`
}

type TopicDislikeReq struct {
	TopicId int64 `form:"topic_id" validate:"required"`
}

type StoryItemFromTopic struct {
	Vmid       int64
	TopicId    int64
	SortBy     int64
	ServerInfo string
	CornerMark string
}

type TopicDetailsAll struct {
	ResAll      *api.TopicDetailsAllReply
	ReserveInfo *api.ReserveRelationInfo
	EsportInfo  *api.EsportInfo
}

func ConstructReserveDescText1(typ activitygrpc.UpActReserveRelationType, time xtime.Time, desc string) string {
	if time <= 0 {
		return ""
	}
	switch typ {
	case activitygrpc.UpActReserveRelationType_Live:
		return fmt.Sprintf("%s直播", pubTimeToString(time.Time()))
	case activitygrpc.UpActReserveRelationType_ESports:
		return desc
	default:
	}
	return ""
}

func pubTimeToString(t time.Time) string {
	now := time.Now()
	if now.Year() == t.Year() {
		if now.Month() == t.Month() && now.Day() == t.Day() {
			return "今天 " + t.Format("15:04")
		}
		if now.Month() == t.Month() && now.Day() < t.Day() && (t.Day()-now.Day() == 1) {
			return "明天 " + t.Format("15:04")
		}
		return t.Format("01-02 15:04")
	}
	return t.Format("2006-01-02 15:04")
}

func ConstructReserveDescText2(total int64) string {
	return statNumberToString(total, "人预约")
}

func statNumberToString(number int64, suffix string) string {
	if number < 50 {
		return ""
	}
	if number < 10000 {
		return strconv.FormatInt(number, 10) + suffix
	}
	var rawFormat string
	if number < 100000000 {
		rawFormat = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(rawFormat, ".0") + "万" + suffix
	}
	rawFormat = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(rawFormat, ".0") + "亿" + suffix
}

func FormMatchTime(stime, localTime int64) (label string) {
	// 计算时区差值(默认服务端固定东八区)
	// 与客户端约定：东一至东十二区分别1到12; 0时区0; 西一至西十一分别-1到-11
	dd, _ := time.ParseDuration(fmt.Sprintf("%dh", localTime-8))
	// 用户所在地的相对开赛时间
	ls := time.Unix(stime, 0).Add(dd)
	// 用户所在地的标准时间
	lt := time.Now().Add(dd)
	if lt.Year() == ls.Year() {
		if lt.YearDay()-ls.YearDay() == 1 {
			label = fmt.Sprintf("昨天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == 0 {
			label = fmt.Sprintf("今天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == -1 {
			label = fmt.Sprintf("明天 %v", ls.Format("15:04"))
			return
		} else {
			label = ls.Format("01-02 15:04")
		}
	} else {
		label = ls.Format("2006-01-02 15:04")
	}
	return
}

func FormMatchState(matchState int32, match *esportsgrpc.ContestDetail, liveEntry *livexroomgate.EntryRoomInfoResp_EntryList) (label, uri, liveLink string, state int32) {
	//nolint:gomnd
	var LiveEntryHandler = func(l *livexroomgate.EntryRoomInfoResp_EntryList, entryFrom string) func(uri string) string {
		return func(uri string) string {
			if l == nil {
				return uri
			}
			if entryFrom != "" {
				entryURI, ok := l.JumpUrl[entryFrom]
				if ok {
					return entryURI
				}
			}
			if l.LiveScreenType == 0 || l.LiveScreenType == 1 {
				return fmt.Sprintf("%s?broadcast_type=%d", uri, l.LiveScreenType)
			}
			return uri
		}
	}
	switch matchState {
	case 1: // 赛前
		if match.IsSubscribed == esportsgrpc.SubscribedStatusEnum_CanSubSubed {
			state = 1
			label = "已订阅"
		} else if match.LiveRoom != 0 {
			state = 2
			label = "订阅"
		} else {
			state = 3
			label = "敬请期待"
		}
	case 2:
		if match.LiveRoom != 0 {
			state = 4
			label = "观看直播"
			uri = fillURI(strconv.FormatInt(match.LiveRoom, 10), nil)
			liveLink = fillURI(strconv.FormatInt(match.LiveRoom, 10), LiveEntryHandler(liveEntry, "NONE"))
		} else {
			state = 5
			label = "敬请期待"
		}
	case 3:
		if match.Playback != "" {
			state = 6
			label = "观看回放"
			uri = match.Playback
		} else if match.CollectionURL != "" {
			state = 7
			label = "观看集锦"
			uri = match.CollectionURL
		} else if match.LiveRoom != 0 {
			state = 8
			label = "直播间"
			uri = fillURI(strconv.FormatInt(match.LiveRoom, 10), nil)
		} else {
			state = 9
			label = "敬请期待"
		}
	}
	return
}

func fillURI(param string, f func(uri string) string) (uri string) {
	uri = "https://live.bilibili.com/" + param
	if f != nil {
		uri = f(uri)
	}
	return
}
