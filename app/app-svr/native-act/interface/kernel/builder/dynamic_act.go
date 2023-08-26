package builder

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
)

type DynamicAct struct{}

func (bu DynamicAct) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	dynCfg, ok := cfg.(*config.DynamicAct)
	if !ok {
		logCfgAssertionError(config.DynamicAct{})
		return nil
	}
	likesRly, ok := material.ActLikesRlys[dynCfg.ActLikesReqID]
	if !ok || likesRly.Subject == nil || len(likesRly.List) == 0 {
		return nil
	}
	mlFactory := kernel.NewMatLoaderFactory(c, dep, ss)
	rid2DynIds := bu.getRid2DynIds(mlFactory, likesRly)
	if len(rid2DynIds) == 0 {
		return nil
	}
	dynDetails := bu.getDynDetails(mlFactory, rid2DynIds)
	if len(dynDetails) == 0 {
		return nil
	}
	if dynCfg.IsFeed {
		if model.IsFromIndex(ss.ReqFrom) {
			return bu.buildFromFeedIndex(dynCfg, ss)
		}
	}
	module := bu.buildModuleBase(dynCfg, ss.ReqFrom)
	for _, v := range likesRly.List {
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		dynDetail, ok := dynDetails[v.Item.Wid]
		if !ok {
			continue
		}
		module.ModuleItems = append(module.ModuleItems, &api.ModuleItem{
			CardType: model.CardTypeDynamic.String(),
			CardId:   strconv.FormatInt(rid2DynIds[v.Item.Wid], 10),
			CardDetail: &api.ModuleItem_DynamicCard{
				DynamicCard: &api.DynamicCard{Dynamic: dynDetail},
			},
		})
	}
	if likesRly.HasMore == 1 {
		module.HasMore = true
		if !dynCfg.IsFeed && model.IsFromIndex(ss.ReqFrom) {
			// 只有非无限feed模式的首页才返回更多卡
			module.ModuleItems = append(module.ModuleItems, bu.buildMoreCard(dynCfg, likesRly.Offset))
		} else {
			module.SubpageParams = bu.subpageParams(module.ModuleId, dynCfg.SortType, likesRly.Offset)
		}
	}
	return module
}

func (bu DynamicAct) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu DynamicAct) buildDynType(subject *activitygrpc.Subject) int64 {
	dynType := model.DynTypeVideo
	if subject.Type == model.ActSubTypeArticle {
		dynType = model.DynTypeArticle
	}
	return int64(dynType)
}

// rid->DynamicItem
func (bu DynamicAct) getDynDetails(mlFactory *kernel.MatLoaderFactory, rid2DynIds map[int64]int64) map[int64]*appdyngrpc.DynamicItem {
	var dynIds []int64
	for _, dynId := range rid2DynIds {
		dynIds = append(dynIds, dynId)
	}
	ml := mlFactory.NewMaterialLoader()
	reqID, err := ml.AddItem(model.MaterialDynDetail, &appdyngrpc.DynServerDetailsReq{DynamicIds: dynIds})
	if err != nil {
		return nil
	}
	rawDynDetails := ml.Load(nil).DynDetails[reqID]
	dynDetails := make(map[int64]*appdyngrpc.DynamicItem, len(rid2DynIds))
	for rid, dynId := range rid2DynIds {
		if dynDetail, ok := rawDynDetails[dynId]; ok {
			dynDetails[rid] = dynDetail
		}
	}
	return dynDetails
}

func (bu DynamicAct) getRid2DynIds(mlFactory *kernel.MatLoaderFactory, likesRly *activitygrpc.LikesReply) map[int64]int64 {
	dynType := bu.buildDynType(likesRly.Subject)
	dynRevsIds := make([]*dyncommongrpc.DynRevsId, 0, len(likesRly.List))
	for _, v := range likesRly.List {
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		dynRevsIds = append(dynRevsIds, &dyncommongrpc.DynRevsId{Rid: v.Item.Wid, DynType: dynType})
	}
	if len(dynRevsIds) == 0 {
		return nil
	}
	ml := mlFactory.NewMaterialLoader()
	reqID, err := ml.AddItem(model.MaterialDynRevsRly, &dynfeedgrpc.FetchDynIdByRevsReq{DynRevsIds: dynRevsIds})
	if err != nil {
		return nil
	}
	dynRevsRly, ok := ml.Load(nil).DynRevsRlys[reqID]
	if !ok {
		return nil
	}
	rid2DynIds := make(map[int64]int64, len(dynRevsRly.Items))
	for _, v := range dynRevsRly.Items {
		if v.DynRevsId == nil {
			continue
		}
		rid2DynIds[v.DynRevsId.Rid] = v.DynId
	}
	return rid2DynIds
}

func (bu DynamicAct) buildModuleBase(cfg *config.DynamicAct, reqFrom string) *api.Module {
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
		}
		if cfg.TextTitle != "" {
			module.ModuleItems = append(module.ModuleItems, card.NewTextTitle(cfg.TextTitle).Build())
		}
	}
	return module
}

func (bu DynamicAct) buildFromFeedIndex(cfg *config.DynamicAct, ss *kernel.Session) *api.Module {
	module := bu.buildModuleBase(cfg, ss.ReqFrom)
	module.SubpageParams = bu.subpageParams(cfg.ModuleBase().ModuleID, cfg.SortType, 0)
	return module
}

func (bu DynamicAct) buildMoreCard(cfg *config.DynamicAct, offset int64) *api.ModuleItem {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("page_id", strconv.FormatInt(cfg.PageID, 10))
	// TODO 待新版覆盖率高时去掉
	params.Set(model.SubParamsField, bu.subpageParams(cfg.ModuleBase().ModuleID, cfg.SortType, offset))
	uri := fmt.Sprintf("bilibili://following/activity_detail/%d?%s", cfg.ModuleBase().ModuleID, params.Encode())
	subpageData := buildSubpageData(cfg.SubpageTitle, cfg.SortList, func(sort int64) string {
		if sort == SubpageCurrSortKey {
			sort = cfg.SortType
		}
		var realOffset int64
		if sort == cfg.SortType {
			realOffset = offset
		}
		return bu.subpageParams(cfg.ModuleBase().ModuleID, sort, realOffset)
	})
	return card.NewDynamicActMore("查看更多", uri, subpageData).Build()
}

func (bu DynamicAct) subpageParams(moduleID, sort, offset int64) string {
	return passthrough.Marshal(&api.DynamicParams{Offset: offset, ModuleId: moduleID, SortType: sort})
}
