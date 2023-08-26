package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	//playurl
	PlayURLNotLogin      = xecode.New(87000) //用户未登录
	PlayURLNotPay        = xecode.New(87001) //稿件未付费
	PlayURLSteinsUpgrade = xecode.New(87002) // 互动视频提示升级
	ProjectInvalidOtt    = xecode.New(87003) // 未过审OTT不支持投屏
	PlayURLArcPayUpgrade = xecode.New(87004) // 付费稿件低版本提示升级/引导
	PlayURLArcPayNotice  = xecode.New(87005) // 付费稿件未付费不可观看
)
