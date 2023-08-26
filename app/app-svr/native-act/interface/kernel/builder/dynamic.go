package builder

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	"github.com/gogo/protobuf/proto"
	ptypes "github.com/gogo/protobuf/types"
	"go-common/library/log"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
)

func init() {
	// bapis注册在"github.com/golang/protobuf/proto"，为了使用MarshalAny，需在"github.com/gogo/protobuf/proto"重新注册
	proto.RegisterType((*dyntopicgrpc.FeedOffset)(nil), "dynamic.service.topic.v1.FeedOffset")
}

type Dynamic struct{}

func (bu Dynamic) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	dynCfg, ok := cfg.(*config.Dynamic)
	if !ok {
		logCfgAssertionError(config.Dynamic{})
		return nil
	}
	if dynCfg.IsFeed {
		if model.IsFromIndex(ss.ReqFrom) {
			return bu.buildFromFeedIndex(dynCfg, material, ss)
		}
	}
	mlFactory := kernel.NewMatLoaderFactory(c, dep, ss)
	return bu.buildFromListIndex(mlFactory, dynCfg, material, ss)
}

func (bu Dynamic) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Dynamic) buildModuleBase(cfg *config.Dynamic, reqFrom string) *api.Module {
	module := &api.Module{
		ModuleType:  model.ModuleTypeDynamic.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: cfg.BgColor, FontColor: cfg.FontColor},
		ModuleUkey:  cfg.ModuleBase().Ukey,
		IsFeed:      cfg.IsFeed,
	}
	if model.IsFromIndex(reqFrom) {
		if cfg.ImageTitle != "" {
			module.ModuleItems = append(module.ModuleItems, card.NewImageTitle(cfg.ImageTitle).Build())
		} else if cfg.TextTitle != "" {
			module.ModuleItems = append(module.ModuleItems, card.NewTextTitle(cfg.TextTitle).Build())
		}
	}
	return module
}

func (bu Dynamic) hasDyns(cfg *config.Dynamic, material *kernel.Material) bool {
	hasDyns, ok := material.HasDynsRlys[cfg.HasDynsReqID]
	if !ok || !hasDyns.HasDyns {
		return false
	}
	return true
}

// 无限feed首刷
func (bu Dynamic) buildFromFeedIndex(cfg *config.Dynamic, material *kernel.Material, ss *kernel.Session) *api.Module {
	if !bu.hasDyns(cfg, material) {
		return nil
	}
	module := bu.buildModuleBase(cfg, ss.ReqFrom)
	module.SubpageParams = bu.subpageParams(cfg, nil, model.DynGroupEmpty)
	return module
}

// 正常列表首刷
func (bu Dynamic) buildFromListIndex(mlFactory *kernel.MatLoaderFactory, cfg *config.Dynamic, material *kernel.Material, ss *kernel.Session) *api.Module {
	listRly, ok := material.ListDynsRlys[cfg.ListDynsReqID]
	if !ok {
		return nil
	}
	dynIDs := bu.mergeListDynIDs(listRly)
	if len(dynIDs) == 0 {
		return nil
	}
	ml := mlFactory.NewMaterialLoader()
	reqID, err := ml.AddItem(model.MaterialDynDetail, &appdyngrpc.DynServerDetailsReq{DynamicIds: dynIDs, IsMaster: cfg.IsMaster, TopDynamicIds: bu.extractDynIDs(listRly.TopList)})
	if err != nil {
		return nil
	}
	dynDetails := ml.Load(nil).DynDetails[reqID]
	if len(dynDetails) == 0 {
		return nil
	}
	module := bu.buildModuleBase(cfg, ss.ReqFrom)
	var lastGroup int64
	var moduleItems []*api.ModuleItem
	if topItems := bu.buildDynModuleItems(model.DynGroupTop, listRly.TopList, dynDetails, ss.LastGroup); len(topItems) > 0 {
		moduleItems = append(moduleItems, topItems...)
		lastGroup = model.DynGroupTop
	}
	if hotItems := bu.buildDynModuleItems(model.DynGroupHot, listRly.HotList, dynDetails, ss.LastGroup); len(hotItems) > 0 {
		moduleItems = append(moduleItems, hotItems...)
		lastGroup = model.DynGroupHot
	}
	if feedItems := bu.buildDynModuleItems(model.DynGroupFeed, listRly.FeedList, dynDetails, ss.LastGroup); len(feedItems) > 0 {
		moduleItems = append(moduleItems, feedItems...)
		lastGroup = model.DynGroupFeed
	}
	if len(moduleItems) == 0 {
		return nil
	}
	if listRly.HasMore {
		module.HasMore = true
		if !cfg.IsFeed && model.IsFromIndex(ss.ReqFrom) {
			// 只有非无限feed模式的首页才返回更多卡
			moduleItems = append(moduleItems, bu.buildMoreCard(cfg, material.Tags, listRly, lastGroup))
		} else {
			module.SubpageParams = bu.subpageParams(cfg, listRly.Offset, lastGroup)
		}
	}
	module.ModuleItems = append(module.ModuleItems, moduleItems...)
	return module
}

func (bu Dynamic) mergeListDynIDs(listRly *dyntopicgrpc.ListDynsRsp) []int64 {
	var dynIDs []int64
	for _, v := range listRly.TopList {
		dynIDs = append(dynIDs, v.GetDynId())
	}
	for _, v := range listRly.HotList {
		dynIDs = append(dynIDs, v.GetDynId())
	}
	for _, v := range listRly.FeedList {
		dynIDs = append(dynIDs, v.GetDynId())
	}
	return dynIDs
}

func (bu Dynamic) buildDynModuleItems(group int64, infos []*dyntopicgrpc.DynInfo, dynDetails map[int64]*appdyngrpc.DynamicItem, lastGroup int64) []*api.ModuleItem {
	var items []*api.ModuleItem
	for _, v := range infos {
		dynDetail, ok := dynDetails[v.DynId]
		if !ok {
			continue
		}
		items = append(items, &api.ModuleItem{
			CardType: model.CardTypeDynamic.String(),
			CardId:   strconv.FormatInt(v.DynId, 10),
			CardDetail: &api.ModuleItem_DynamicCard{
				DynamicCard: &api.DynamicCard{Dynamic: dynDetail},
			},
		})
	}
	if len(items) == 0 {
		return nil
	}
	// lastGroup之前的动态组标题已经返回
	if group <= lastGroup {
		return items
	}
	return append([]*api.ModuleItem{card.NewText(model.DynGroupNames[group]).Build()}, items...)
}

func (bu Dynamic) buildMoreCard(cfg *config.Dynamic, tags map[int64]*taggrpc.Tag, listRly *dyntopicgrpc.ListDynsRsp, lastGroup int64) *api.ModuleItem {
	params := url.Values{}
	params.Set("title", cfg.ModuleTitle)
	params.Set("sort", cfg.Sort)
	name := cfg.PageTitle
	if tag, ok := tags[cfg.TopicID]; ok {
		name = tag.Name
	}
	params.Set("name", name)
	params.Set("module_id", strconv.FormatInt(cfg.ModuleBase().ModuleID, 10))
	params.Set("page_id", strconv.FormatInt(cfg.PageID, 10))
	params.Set("sortby", strconv.FormatInt(int64(cfg.SortBy), 10))
	params.Set("offset", bu.oldOffset(cfg.SortBy, listRly.Offset))
	params.Set(model.SubParamsField, bu.subpageParams(cfg, listRly.Offset, lastGroup))
	uri := fmt.Sprintf("bilibili://following/topic_content_list/%d?%s", cfg.TopicID, params.Encode())
	subpageData := buildSubpageData(cfg.ModuleTitle, nil, func(sort int64) string {
		return bu.subpageParams(cfg, listRly.Offset, lastGroup)
	})
	return card.NewDynamicMore("查看更多", uri, subpageData).Build()
}

func (bu Dynamic) oldOffset(sortBy int32, feedOffset *dyntopicgrpc.FeedOffset) string {
	if feedOffset == nil {
		return ""
	}
	if sortBy == model.DynSortTime || sortBy == model.DynSortCompre {
		return strconv.FormatInt(feedOffset.DynamicId, 10)
	}
	switch sortBy {
	case model.DynSortTime, model.DynSortCompre:
		return strconv.FormatInt(feedOffset.DynamicId, 10)
	case model.DynSortHot:
		return fmt.Sprintf("%d_%d", feedOffset.TimeSlot, feedOffset.Index)
	}
	return ""
}

func (bu Dynamic) subpageParams(cfg *config.Dynamic, offset *dyntopicgrpc.FeedOffset, lastGroup int64) string {
	params := &api.DynamicParams{ModuleId: cfg.ModuleBase().ModuleID, LastGroup: lastGroup}
	if offset != nil {
		offsetPB, err := ptypes.MarshalAny(offset)
		if err != nil {
			log.Error("Fail to MarshalAny feedOffset, offset=%+v error=%+v", offset, err)
			return ""
		}
		params.FeedOffset = offsetPB
	}
	return passthrough.Marshal(params)
}

func (bu Dynamic) extractDynIDs(list []*dyntopicgrpc.DynInfo) []int64 {
	dynIDs := make([]int64, 0, len(list))
	for _, v := range list {
		dynIDs = append(dynIDs, v.DynId)
	}
	return dynIDs
}
