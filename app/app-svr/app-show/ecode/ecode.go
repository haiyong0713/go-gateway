package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	HotWordAIErr         = xecode.New(78038) // 热门热词AI返回为空
	HotWordNoAuditingErr = xecode.New(78039) // 热门热词审核未通过
	DynamicBuildLimit    = xecode.New(78042) // 话题活动-当前版本较低，无法显示完全，请更新至最新版本后查看
	ActivityNothingMore  = xecode.New(78045) //话题活动-没有更多内容了
	ActivityHasLock      = xecode.New(78046) //还未解锁，敬请期待
	HasVoted             = xecode.New(78047) //您已经投过票了
)
