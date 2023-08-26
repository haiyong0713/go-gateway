package ecode

import (
	xecode "go-common/library/ecode"
)

var (

	// ActivityInviterTelIsOld 手机号是旧的
	ActivityInviterTelIsOld = xecode.New(75802)
	// ActivityMobileNotAllow 手机号是旧的
	ActivityMobileNotAllow = xecode.New(75803)
	// ActivityInviterJoinNotAllow 不允许加入
	ActivityInviterJoinNotAllow = xecode.New(75804)
	// ActivityTokenNotFind token未找到
	ActivityTokenNotFind = xecode.New(75805)
	// ActivityInviteMidNotFind 邀请人未找到
	ActivityInviteMidNotFind = xecode.New(75806)
	// ActivityInviterNoBindTelGetErr 邀请未绑定手机
	ActivityInviterNoBindTelGetErr = xecode.New(75807)
	// ActivityInviterShareNotAllow 不允许分享
	ActivityInviterShareNotAllow = xecode.New(75808)
)
