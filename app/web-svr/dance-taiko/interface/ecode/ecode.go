package ecode

import xecode "go-common/library/ecode"

var (
	JsonFormatErr = xecode.New(14801) //
	GameIDErr     = xecode.New(14802) //
	GameStatusErr = xecode.New(14803)
	PlayerErr     = xecode.New(14804)
	RestartErr    = xecode.New(14805) // 重新开始失败
	PlayerBeyound = xecode.New(14806) // 人数超过上限
)
