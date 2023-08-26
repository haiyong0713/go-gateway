package model

import (
	"context"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

type DynSchemaCtx struct {
	Ctx                      context.Context
	DynCtx                   *dynmdlV2.DynamicContext
	DynCmtMode               map[int64]*DynCmtMeta // 动态评论模式
	TopicId                  int64
	SortBy                   int64
	Offset                   string
	ItemFrom                 map[int64]string                             // 话题来源
	HiddenAttached           map[int64]bool                               // 隐式关联
	ServerInfo               map[int64]string                             // 服务端透传数据
	IsDisableInt64MidVersion bool                                         // 禁止int64 mid版本
	TopicCreatorMid          int64                                        // 话题创建者mid
	OwnerAppear              int32                                        // 是否点赞外露
	TopicCreatorLike         map[string]*thumgrpc.MultiStatsReply_Records //话题创建者点赞信息
	MergedResource           map[int64]MergedResource                     // 收拢信息
}

type MergedResource struct {
	MergeType    int32 // 收拢类型
	MergedResCnt int32 // 收拢资源组中的数量
}

func (dyn *DynSchemaCtx) CanBeForward() bool {
	return dyn.DynCtx.Dyn.IsForward()
}

func (dyn *DynSchemaCtx) CanBeAv() bool {
	return dyn.DynCtx.Dyn.IsAv()
}

func (dyn *DynSchemaCtx) CanBeDraw() bool {
	return dyn.DynCtx.Dyn.IsDraw()
}

func (dyn *DynSchemaCtx) CanBeWord() bool {
	return dyn.DynCtx.Dyn.IsWord()
}

func (dyn *DynSchemaCtx) CanBeArticle() bool {
	if pd.WithContext(dyn.Ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build("<", int64(65400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33200000))
	}).MustFinish() {
		return false
	}
	return dyn.DynCtx.Dyn.IsArticle()
}

func (dyn *DynSchemaCtx) CanBePGC() bool {
	if pd.WithContext(dyn.Ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build("<", int64(65400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33200000))
	}).MustFinish() {
		return false
	}
	return dyn.DynCtx.Dyn.IsPGC()
}

func (dyn *DynSchemaCtx) CanBeCommon() bool {
	if pd.WithContext(dyn.Ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build("<", int64(65400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33200000))
	}).MustFinish() {
		return false
	}
	return dyn.DynCtx.Dyn.IsCommon()
}

//nolint:gosimple
func (dyn *DynSchemaCtx) CanBeAdditionalReserve() bool {
	if pd.WithContext(dyn.Ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build("<", int64(65400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33200000))
	}).MustFinish() {
		return false
	}
	return true
}

type DynCmtMeta struct {
	CmtShowStat int32 // 评论外露是否展示
	CmtMode     int64 // 评论外露模式
}

type DynMetaCardListParam struct {
	DynId          int64
	DynCmtMeta     *DynCmtMeta
	ItemFrom       string
	HiddenAttached bool   // 隐式关联
	ServerInfo     string //透传字段
	MergedResource MergedResource
}
