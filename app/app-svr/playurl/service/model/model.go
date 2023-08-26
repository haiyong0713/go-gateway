package model

import (
	"fmt"

	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
)

const (
	Total = "total"
	Web   = "web"
	App   = "app"
	//broadcast新接口web端的前缀为video
	Video = "video"
	_2min = 2 * 60
)

type LitePlayerInfoc struct {
	Buvid    string
	GroupID  int64
	JoinTime int64
}

type CopyRightRestriction struct {
	Aid         int64 `json:"aid"`
	BanBackend  bool  `json:"ban_backend"`
	BanPip      bool  `json:"ban_pip"`
	BanMiracast bool  `json:"ban_miracast"`
}

func OnlineKey(aid int64) string {
	return fmt.Sprintf("oc_%d", aid)
}

type OnlineInfo struct {
	WebCidCount map[int64]int64 `json:"web_cid_count"`
	AppCidCount map[int64]int64 `json:"app_cid_count"`
	Time        int64           `json:"time"`
	AidTotal    int64           `json:"aid_total"`
}

type OnlineCount struct {
	Total int64
	Web   int64
	App   int64
}

type GlanceMsg struct {
	Mid       int64
	Duration  int64
	IsSp      bool
	CanGlance bool
	Group     v2.Group
}

func (msg *GlanceMsg) SupportGlance() bool {
	return msg.Mid > 0 && !msg.IsSp && msg.Duration >= _2min
}

func (msg *GlanceMsg) FetchGlanceTime(defTime int64, ratio int64) int64 {
	//取整后加1，给用户多送10s试看
	glanceTime := (msg.Duration/100 + 1) * ratio
	if glanceTime >= defTime {
		return defTime
	}
	return glanceTime
}

type ConfValueEdit struct {
	ConfValue *v2.ConfValue
	ConfType  v2.ConfType
}
