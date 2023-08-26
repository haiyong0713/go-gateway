package ecode

import (
	"go-common/library/ecode"
)

var (
	// 亲子平台：78060-78070
	FamilyNotRealnamed   = ecode.New(78060) //未实名认证
	FamilyInvalidQrcode  = ecode.New(78061) //二维码超时或者已被使用
	FamilyExceedLimit    = ecode.New(78062) //账号绑定超过上限
	FamilyNotSupportBind = ecode.New(78063) //当前账号不支持绑定此关系
	FamilyLockExceed     = ecode.New(78064) //当前操作人数过多
)
