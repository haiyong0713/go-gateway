package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Header struct {
	BaseCfgManager

	BgColor             string
	SponsorContent      string
	DisplayUser         bool
	DisplayH5Header     bool
	SponsorMid          int64
	TopicID             int64
	DisplayViewNum      bool //是否展示浏览、讨论数
	DisplaySubscribeBtn bool //是否展示订阅按钮
	ActiveUsersReqID    kernel.RequestID
}
