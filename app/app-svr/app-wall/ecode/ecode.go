package ecode

import (
	xecode "go-common/library/ecode"
)

const OrderFailedMessage = "因运营商同步信息延迟，当前无法激活，请您稍后重试，带来不便请谅解"

var (
	// APP
	AppNotData             = xecode.New(78000) // app not data
	AppFlowNotOrdered      = xecode.New(78001) // app该卡号尚未开通哔哩哔哩专属免流服务
	AppFlowExpired         = xecode.New(78002) // app该卡号哔哩哔哩专属免流服务已退订且已过期
	AppVerificationError   = xecode.New(78003) // app短信验证码错误
	AppVerificationExpired = xecode.New(78004) // app短信验证码已过期
	AppQueryExceededLimit  = xecode.New(78006) // app查询数量超过上限
	// 福利社
	AppComicUserNotExist                    = xecode.New(78031) // 该用户未登录过哔哩哔哩漫画
	AppWelfareClubNoBinding                 = xecode.New(78032) // 用户未绑定福利社
	AppWelfareClubOnlySupportCard           = xecode.New(78100) // 该业务只支持哔哩哔哩免流卡
	AppWelfareClubNotFree                   = xecode.New(78101) // 该卡号尚未开通哔哩哔哩专属免流服务
	AppWelfareClubCancelOrExpire            = xecode.New(78102) // 该卡号哔哩哔哩专属免流服务已退订且已过期
	AppWelfareClubPackNotExist              = xecode.New(78103) // 该礼包不存在
	AppWelfareClubLackIntegral              = xecode.New(78104) // 福利点不足
	AppWelfareClubWaitOneMinute             = xecode.New(78105) // 请间隔一分钟之后再领取流量包
	AppWelfareClubLackFlow                  = xecode.New(78106) // 可用流量不足
	AppWelfareClubWaitResult                = xecode.New(78107) // 请稍后查看结果
	AppWelfareClubOnlyOnce                  = xecode.New(78108) // 每张卡只能领取一次特权礼包哦
	AppWelfareClubPackLack                  = xecode.New(78109) // 该礼包库存不足，请兑换其他礼包
	AppWelfareClubFlowOrderForbidden        = xecode.New(78110) // 您当前是生效的免流卡故暂无法订购免流包
	AppWelfareClubActiveFailed              = xecode.New(78111) // 激活失败，请重新输入验证码激活
	AppWelfareClubBinded                    = xecode.New(78112) // 该账户已绑定过手机号
	AppWelfareClubRegistered                = xecode.New(78113) // 该手机号已被注册
	AppWelfareClubOrderCancelFailed         = xecode.New(78114) // 免流产品退订失败，请联系运营商进行退订
	AppWelfareClubOrderFailed               = xecode.New(78115) // 激活失败,该卡号尚未开通哔哩哔哩专属免流服务
	AppWelfareClubPackCountLimit            = xecode.New(78116) // 该优惠券领取次数超过上限，请兑换其他福利
	AppWelfareClubNoAllowPackExchange       = xecode.New(78117) // 本月5000点兑换额度已用完，请下月再兑换
	AppWelfareClubRejectPackExchange        = xecode.New(78118) // 兑换过于频繁，请明日再来兑换
	AppWelfareClubOnlySupportCardFromUnicom = xecode.New(78119) // 该业务只支持哔哩哔哩免流卡
	AppWelfareClubRequestUnicom             = xecode.New(78120) // 请求联通服务错误
	AppWelfareClubUnicomServiceUpgrade      = xecode.New(78121) // 联通系统维护升级
)
