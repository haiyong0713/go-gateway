package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	// kvo
	KvoTimestampErr   = xecode.New(23001) // 时间戳不合法
	KvoCheckSumErr    = xecode.New(23002) // checksum不合法
	KvoDataOverLimit  = xecode.New(23003) // 数据超过限制
	KvoNotModified    = xecode.New(23004) // 数据没有修改
	KvoHashConflict   = xecode.New(23005) // hash key冲突
	KvoModuleNotExist = xecode.New(23006) // module 不存在
)
