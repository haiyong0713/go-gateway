package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	PopularGRPCRecordNotFound = xecode.New(79000)  // 热门grpc返回为空
	DynViewNotFound           = xecode.New(165000) // 动态详情页，当前动态不存在
)
