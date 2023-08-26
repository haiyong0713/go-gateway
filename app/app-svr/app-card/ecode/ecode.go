package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	// APP
	AppNotData             = xecode.New(78000) // app not data
	AppFlowNotOrdered      = xecode.New(78001) // app该卡号尚未开通哔哩哔哩专属免流服务
	AppFlowExpired         = xecode.New(78002) // app该卡号哔哩哔哩专属免流服务已退订且已过期
	AppVerificationError   = xecode.New(78003) // app短信验证码错误
	AppVerificationExpired = xecode.New(78004) // app短信验证码已过期
	AppSelectedNotValid    = xecode.New(78005) // 热门精选不可通过，前三位未有寄语
	AppQueryExceededLimit  = xecode.New(78006) // app查询数量超过上限
	AIDataExist            = xecode.New(78007) // AI数据已填入，拒绝再次填入
	AppSelectedIDExist     = xecode.New(78020) // 热门精选ID已存在
	AppPreciousNotExist    = xecode.New(78021) // 热门镇站之宝未配置数据
	AppPreciousNotNorm     = xecode.New(78022) // 热门镇站之宝数据被过滤
	AppActNotStart         = xecode.New(78033) // 活动尚未开始，无法投票
	AppActOverLikeLimit    = xecode.New(78034) // 票数已用完，无法投票
	AppNoLikeCondition     = xecode.New(78035) // 无投票资格
	AppPageOffline         = xecode.New(78036) // 活动已经下线
	AppActHasEnd           = xecode.New(78037) // 活动已结束，无法投票
	AppkeyExistErr         = xecode.New(78040) // fawkes appkey已存在
	GarbImageIDIllegal     = xecode.New(78041) // 请求装扮详情的image_id非法
	AppViewForRetry        = xecode.New(78042) // HTTP VIEW 接口，促使 SLB 降级
	// Other
	CoinOverMax = xecode.New(34005) // 超过单个视频投币上限
	// 活动 - 领取挂件
	AppReceiveErr = xecode.New(78044) // 奖励领取失败
	// 大型活动合集
	AppActivitySeasonFallback = xecode.New(78200) // 活动合集页失败可降级普通合集页
	AppTeenagersFilter        = xecode.New(78301) // 播放页青少年模式不展示
	//首映
	PremiereRoomRisk = xecode.New(6006126) //首映房间被风控
	//合集打卡
	ActivityIsExpired      = xecode.New(187006) // 报名时间已过期，无法报名
	ActivityNotExists      = xecode.New(187007) // 活动不存在或已失效
	UserActivityInProgress = xecode.New(187008) // 已成功报名，无需重复操作
	UserActivityNotFound   = xecode.New(187009) // 当前没有进行中的打卡
	UserActivityError      = xecode.New(187010) // 打卡出错了，请刷新查看
	ActivityNetError       = xecode.New(180000) // 网络错误，请刷新查看
)
