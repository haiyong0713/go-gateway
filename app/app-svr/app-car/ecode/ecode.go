package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	AppNotVedio          = xecode.New(142000) // 当前视频不存在
	AppAreaLimit         = xecode.New(142001) // 当前海外限制
	AppAttrBitSteinsGate = xecode.New(142002) // 当前互动视频不能播放
	AppCannotPlay        = xecode.New(142003) // 不支持该视频格式
	AppMediaNotData      = xecode.New(142010) // 当前媒资接口没有更多数据
	AppCarVipOnlyOnce    = xecode.New(142011) // 你已经领取过了，快去追喜欢的影片吧
	AppCarVipActivityEnd = xecode.New(142012) // 活动已结束
	AppCarVipRiskUser    = xecode.New(142013) // 你暂时无法活动
	AppReportPlayError   = xecode.New(142014) // 当前音频历史记录上报失败
	AppCarVipError       = xecode.New(142015) // 大会员兑换失败
	AppVideoInsecurity   = xecode.New(142016) // 该视频已失效（内容安全过滤专用）
	AppVideoReachBottom  = xecode.New(142017) // 已加载到底啦（内容安全过滤专用）
)
