package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	InvalidModPoolKey    = xecode.New(400000) // 无效的mod_pool_key
	ForbiddenOperateMod  = xecode.New(400001) // 无修改资源权限
	DisableMod           = xecode.New(400002) // 限制类型资源禁止修改操作
	ExistMod             = xecode.New(400003) // 已存在相同mod名,不可重复创建
	ProcessingVersion    = xecode.New(400004) // 增量包尚未构建完成,请稍后再试
	DisableVersion       = xecode.New(400005) // 该版本已永久下线
	OfflineVersionNoPush = xecode.New(400006) // 已经下线的资源不能进行推送下线操作
	DisableVersionNoPush = xecode.New(400007) // 资源不存在生效的版本不能进行推送下线操作
	NoExistAppKey        = xecode.New(400008) // appKey不存在
)
