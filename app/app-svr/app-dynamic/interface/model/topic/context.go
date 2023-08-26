package topic

import (
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

type Context struct {
	Activitys map[int64]*natpagegrpc.NativePage
	Accounts  map[int64]*accountgrpc.Card
	Archives  map[int64]*archivegrpc.ArcPlayer
	Draws     map[int64]*dynmdlV2.DrawDetailRes
}
