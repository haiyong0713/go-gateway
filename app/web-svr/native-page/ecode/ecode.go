package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	// activity ecode
	ActivityLikeIPFrequence    = xecode.New(75034) //点赞活动ip访问过于频繁-访问过快
	ActivityLikeScoreLower     = xecode.New(75035) //账户score过低不支持点赞-异常账号!
	ActivityLikeLevelLimit     = xecode.New(75036) //-用户等级不够!
	ActivityLikeNotStart       = xecode.New(75038) //-评分未开始!
	ActivityLikeHasEnd         = xecode.New(75039) //-评分已结束!
	ActivityLikeRegisterLimit  = xecode.New(75042) //-晚于活动限制注册时间!
	ActivityLikeBeforeRegister = xecode.New(75046) //-早于活动限制注册时间!
	ActivityTelValid           = xecode.New(75055) //-未绑定有效手机号码
	ActivityOverLikeLimit      = xecode.New(75089) //-票数已用完，无法投票

	NativePageOffline   = xecode.New(75094) //-活动已经下线
	ActivityFrequence   = xecode.New(75114) // 操作过快
	ActivityNtUserLimit = xecode.New(75611) // up发起活动：用户不合法
	ActivityNtNoBind    = xecode.New(75612) //暂不支持编辑
	UpBindOtherPage     = xecode.New(75613) //已绑定其他活动
	NotOnline           = xecode.New(75614) //非上线状态
	SpaceUnbindFail     = xecode.New(75615) //空间tab解绑失败
	UpActIllegal        = xecode.New(75616) //暂不满足参与条件
	UpActBusIllegal     = xecode.New(75617) //企业号专属参与通道，请私信@哔哩哔哩广告娘，解锁独家权益
	NaAuditFail         = xecode.New(75618) //当前活动审核不通过
	NaAuditChanged      = xecode.New(75619) //内容已更新，请刷新页面
	// 176000-176999
	NotUpgradeTopic = xecode.New(176000) //非自动升级话题
)
