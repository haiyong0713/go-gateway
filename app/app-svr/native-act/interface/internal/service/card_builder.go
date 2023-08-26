package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	GlobalCardBuilder = &cardBuilder{builders: map[int64]builder.Builder{}}
)

func init() {
	GlobalCardBuilder.Register(natpagegrpc.ModuleEditor, builder.Editor{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleEditorOrigin, builder.EditorOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleParticipation, builder.Participation{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleBaseHead, builder.Header{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleDynamic, builder.Dynamic{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleVideo, builder.DynamicAct{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleLive, builder.LiveID{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleCarouselImg, builder.CarouselImg{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleCarouselWord, builder.CarouselWord{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleCarouselSource, builder.CarouselOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleGame, builder.Game{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleResourceID, builder.ResourceID{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleResourceAct, builder.ResourceAct{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleResourceDynamic, builder.ResourceTopic{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleResourceRole, builder.ResourceRole{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleResourceOrigin, builder.ResourceOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewVideoAvid, builder.VideoID{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewVideoAct, builder.VideoAct{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewVideoDyn, builder.VideoTopic{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleRecommend, builder.Rcmd{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleRcmdSource, builder.RcmdOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleRcmdVertical, builder.RcmdVertical{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleRcmdVerticalSource, builder.RcmdVerticalOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleAct, builder.Relativeact{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleActCapsule, builder.RelativeactCapsule{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleStatement, builder.Statement{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleIcon, builder.Icon{})
	GlobalCardBuilder.Register(natpagegrpc.VoteModule, builder.Vote{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleReserve, builder.Reserve{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleTimelineIDs, builder.Timeline{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleTimelineSource, builder.TimelineOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleOgvSeasonID, builder.Ogv{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleOgvSeasonSource, builder.OgvOrigin{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNavigation, builder.Navigation{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleReply, builder.Reply{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleInlineTab, builder.Tab{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewactHeader, builder.NewactHeader{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewactAward, builder.NewactAward{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleNewactStatement, builder.NewactStatement{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleProgress, builder.Progress{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleSelect, builder.Select{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleClick, builder.Click{})
	GlobalCardBuilder.Register(natpagegrpc.ModuleBaseHoverButton, builder.HoverButton{})
	GlobalCardBuilder.Register(natpagegrpc.BottomButtonModule, builder.BottomButton{})
}

type cardBuilder struct {
	builders map[int64]builder.Builder //module.category=>builder
}

func (cb *cardBuilder) Register(category int64, builder builder.Builder) {
	if _, ok := cb.builders[category]; ok {
		log.Warn("builder=%+v has already registered", category)
	}
	cb.builders[category] = builder
}

func (cb *cardBuilder) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfgs []config.BaseCfgManager, material *kernel.Material) []*api.Module {
	modules := make(map[int64]*api.Module, len(cfgs))
	eg := errgroup.WithContext(c)
	mu := sync.Mutex{}
	for _, v := range cfgs {
		cfg := v
		bu, ok := GlobalCardBuilder.builders[cfg.ModuleBase().Category]
		if !ok {
			log.Warnc(c, "builder not found, category=%+v", cfg.ModuleBase().Category)
			continue
		}
		eg.Go(func(ctx context.Context) error {
			module := bu.Build(ctx, ss, dep, cfg, material)
			if module == nil {
				return nil
			}
			mu.Lock()
			defer mu.Unlock()
			modules[cfg.ModuleBase().ModuleID] = module
			return nil
		})
	}
	if err := eg.Wait(); err != nil { //错误降级，不处理
		log.Errorc(c, "Fail to build modules, error=%+v", err)
	}
	return cb.after(c, cfgs, modules, ss)
}

func (cb *cardBuilder) after(c context.Context, cfgs []config.BaseCfgManager, modules map[int64]*api.Module, ss *kernel.Session) []*api.Module {
	data := &builder.AfterContextData{}
	if ss.PageRlyContext.HasNavigation {
		data.NaviItems = builder.Navigation{}.BuildNavigationItems(cfgs)
	}
	list := make([]*api.Module, 0, len(modules))
	for _, cfg := range cfgs {
		bu, ok := GlobalCardBuilder.builders[cfg.ModuleBase().Category]
		if !ok {
			log.Warnc(c, "builder not found, category=%+v", cfg.ModuleBase().Category)
			continue
		}
		module, ok := modules[cfg.ModuleBase().ModuleID]
		if !ok {
			continue
		}
		if ok := bu.After(data, module); ok {
			list = append(list, module)
		}
	}
	return list
}

func buildPageResp(ss *kernel.Session, natPage *natpagegrpc.NativePage, modules []*api.Module, baseModules []*api.Module) *api.PageResp {
	resp := &api.PageResp{
		IsOnline:           natPage.IsOnline(),
		IgnoreAppDarkTheme: natPage.IsAttrNotNightModule() == natpagegrpc.AttrModuleYes,
		PageColor:          &api.Color{BgColor: natPage.BgColor},
		ModuleList:         modules,
		SponsorType:        int64(natPage.FromType),
		HoverButton:        fetchModule(baseModules, model.ModuleTypeHoverButton),
		BottomButton:       fetchModule(baseModules, model.ModuleTypeBottomButton),
	}
	if natPage.IsTopicAct() {
		resp.TopicInfo = &api.TopicInfo{
			TopicId: natPage.ForeignID,
			Title:   natPage.Title,
		}
		resp.PageShare = buildPageShare(ss, natPage)
		resp.PageHeader = fetchModule(baseModules, model.ModuleTypeHeader)
		resp.Participation = fetchModule(baseModules, model.ModuleTypeParticipation)
	}
	if natPage.IsMenuAct() {
		resp.TopTab = buildTopTabFromConfSet(natPage.ConfSetUnmarshal())
	}
	if natPage.IsNewact() {
		resp.PageShare = buildPageShare(ss, natPage)
	}
	//组建互斥逻辑
	//吸底按钮 >自定义悬浮组件 > 悬浮投稿按钮
	if resp.BottomButton != nil {
		resp.HoverButton = nil
		resp.Participation = nil
	}
	//组建互斥逻辑
	return resp
}

func buildPageShare(ss *kernel.Session, page *natpagegrpc.NativePage) *api.PageShare {
	share := &api.PageShare{
		Type:                  model.ShareTypeActivity,
		Title:                 page.Title,
		Desc:                  page.ShareTitle,
		Image:                 page.ShareImage,
		OutsideUri:            page.ShareURL,
		SpacePageUrl:          fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=topic_sync_space&act_id=%d", page.ID),
		SpaceExclusivePageUrl: fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=topic_set_space&act_id=%d", page.ID),
	}
	if page.ShareCaption != "" {
		share.Title = page.ShareCaption
	}
	if ss.ShareReq != nil && ss.ShareReq.ShareOrigin == model.ShareOriginTab && ss.ShareReq.TabID > 0 && ss.ShareReq.TabModuleID > 0 {
		share.InsideUri = fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d",
			page.ID, ss.ShareReq.TabID, ss.ShareReq.TabModuleID, time.Now().Unix())
	} else {
		share.InsideUri = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", page.ID, time.Now().Unix())
	}
	if share.OutsideUri == "" {
		share.OutsideUri = share.InsideUri
	}
	return share
}

func setExtraIndexResp(req *api.IndexReq, resp *api.PageResp, materials *kernel.Material, reqID *IndexReqID) {
	if req.DynamicId > 0 && model.NeedLayerDynamic(req.ActivityFrom) {
		func() {
			dynDetail, ok := materials.DynDetails[reqID.DynDetail][req.DynamicId]
			if !ok {
				return
			}
			title := "来自首页"
			if req.ActivityFrom == model.ActFromDt {
				title = "活动推荐"
			}
			resp.LayerDynamic = &api.LayerDynamic{Title: title, Dynamic: dynDetail}
		}()
	}
	resp.PageId = req.PageId
}

func buildTopTabFromConfSet(confSet *natpagegrpc.ConfSet) *api.TopTab {
	if confSet == nil {
		return nil
	}
	switch confSet.BgType {
	case model.TopTabBgImg:
		return &api.TopTab{
			BgImage1:  confSet.BgImage1,
			BgImage2:  confSet.BgImage2,
			FontColor: confSet.FontColor,
			BarType:   int64(confSet.BarType),
		}
	case model.TopTabBgColor:
		return &api.TopTab{
			TabTopColor:    confSet.TabTopColor,
			TabMiddleColor: confSet.TabMiddleColor,
			TabBottomColor: confSet.TabBottomColor,
			FontColor:      confSet.FontColor,
			BarType:        int64(confSet.BarType),
		}
	default:
		return nil
	}
}

func fetchModule(modules []*api.Module, mt model.ModuleType) *api.Module {
	for _, m := range modules {
		if m.ModuleType == mt.String() {
			return m
		}
	}
	return nil
}
