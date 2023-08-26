package ecode

import xecode "go-common/library/ecode"

var (
	PlNameTooLong      = xecode.New(63001) // 播单标题超出最大限制
	PlDescTooLong      = xecode.New(63002) // 播单简介超出最大限制
	PlMaxCount         = xecode.New(63003) // 播单个数超出最大限制
	PlCanNotDelDefault = xecode.New(63004) // 不能删除默认播单
	PlExist            = xecode.New(63005) // 已经存在相同标题的播单
	PlNotExist         = xecode.New(63006) // 播单无法访问
	PlAlreadyDel       = xecode.New(63007) // 播单已经删除
	PlDenied           = xecode.New(63008) // 播单暂未开放
	PlVideoOverflow    = xecode.New(63009) // 播单内视频个数超出最大限制
	PlVideoExist       = xecode.New(63010) // 视频已经添加进此播单
	PlVideoAlreadyDel  = xecode.New(63011) // 视频已经不属于此播单
	PlSortOverflow     = xecode.New(63012) // 播单内视频排序不生效
	PlFavExist         = xecode.New(63013) // 播单已经收藏
	PlFavAlreadyDel    = xecode.New(63014) // 播单未收藏
	PlNotUser          = xecode.New(63015) // 非创建者不能修改此播单
)
