package ecode

import xecode "go-common/library/ecode"

var (
	PLayerPolicyNotExist = xecode.New(19001) // 策略不存在
	PLayerPolicyNotStart = xecode.New(19002) // 策略未开始
	PLayerPolicyEnded    = xecode.New(19003) // 策略未开始

	PlayURLNotPay = xecode.New(87005) // 稿件未支付
)
