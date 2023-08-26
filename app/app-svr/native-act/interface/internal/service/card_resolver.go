package service

import (
	"context"
	"strconv"
	"sync/atomic"

	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/resolver"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	GlobalCardResolver = &cardResolver{resolvers: map[int64]resolver.Resolver{}}
)

func init() {
	GlobalCardResolver.Register(natpagegrpc.ModuleEditor, resolver.Editor{})
	GlobalCardResolver.Register(natpagegrpc.ModuleEditorOrigin, resolver.EditorOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleParticipation, resolver.Participation{})
	GlobalCardResolver.Register(natpagegrpc.ModuleBaseHead, resolver.Header{})
	GlobalCardResolver.Register(natpagegrpc.ModuleDynamic, resolver.Dynamic{})
	GlobalCardResolver.Register(natpagegrpc.ModuleVideo, resolver.DynamicAct{})
	GlobalCardResolver.Register(natpagegrpc.ModuleLive, resolver.LiveID{})
	GlobalCardResolver.Register(natpagegrpc.ModuleCarouselImg, resolver.CarouselImg{})
	GlobalCardResolver.Register(natpagegrpc.ModuleCarouselWord, resolver.CarouselWord{})
	GlobalCardResolver.Register(natpagegrpc.ModuleCarouselSource, resolver.CarouselOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleGame, resolver.Game{})
	GlobalCardResolver.Register(natpagegrpc.ModuleResourceID, resolver.ResourceID{})
	GlobalCardResolver.Register(natpagegrpc.ModuleResourceAct, resolver.ResourceAct{})
	GlobalCardResolver.Register(natpagegrpc.ModuleResourceDynamic, resolver.ResourceTopic{})
	GlobalCardResolver.Register(natpagegrpc.ModuleResourceRole, resolver.ResourceRole{})
	GlobalCardResolver.Register(natpagegrpc.ModuleResourceOrigin, resolver.ResourceOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewVideoAvid, resolver.VideoID{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewVideoAct, resolver.VideoAct{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewVideoDyn, resolver.VideoTopic{})
	GlobalCardResolver.Register(natpagegrpc.ModuleRecommend, resolver.Rcmd{})
	GlobalCardResolver.Register(natpagegrpc.ModuleRcmdSource, resolver.RcmdOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleRcmdVertical, resolver.RcmdVertical{})
	GlobalCardResolver.Register(natpagegrpc.ModuleRcmdVerticalSource, resolver.RcmdVerticalOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleAct, resolver.Relativeact{})
	GlobalCardResolver.Register(natpagegrpc.ModuleActCapsule, resolver.RelativeactCapsule{})
	GlobalCardResolver.Register(natpagegrpc.ModuleStatement, resolver.Statement{})
	GlobalCardResolver.Register(natpagegrpc.ModuleIcon, resolver.Icon{})
	GlobalCardResolver.Register(natpagegrpc.VoteModule, resolver.Vote{})
	GlobalCardResolver.Register(natpagegrpc.ModuleReserve, resolver.Reserve{})
	GlobalCardResolver.Register(natpagegrpc.ModuleTimelineIDs, resolver.Timeline{})
	GlobalCardResolver.Register(natpagegrpc.ModuleTimelineSource, resolver.TimelineOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleOgvSeasonID, resolver.Ogv{})
	GlobalCardResolver.Register(natpagegrpc.ModuleOgvSeasonSource, resolver.OgvOrigin{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNavigation, resolver.Navigation{})
	GlobalCardResolver.Register(natpagegrpc.ModuleReply, resolver.Reply{})
	GlobalCardResolver.Register(natpagegrpc.ModuleInlineTab, resolver.Tab{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewactHeader, resolver.NewactHeader{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewactAward, resolver.NewactAward{})
	GlobalCardResolver.Register(natpagegrpc.ModuleNewactStatement, resolver.NewactStatement{})
	GlobalCardResolver.Register(natpagegrpc.ModuleProgress, resolver.Progress{})
	GlobalCardResolver.Register(natpagegrpc.ModuleSelect, resolver.Select{})
	GlobalCardResolver.Register(natpagegrpc.ModuleClick, resolver.Click{})
	GlobalCardResolver.Register(natpagegrpc.ModuleBaseHoverButton, resolver.HoverButton{})
	GlobalCardResolver.Register(natpagegrpc.BottomButtonModule, resolver.BottomButton{})
}

type cardResolver struct {
	moduleWhitelist atomic.Value
	pageBlacklist   atomic.Value
	resolvers       map[int64]resolver.Resolver //module.category=>resolver
}

func InitGlobalCardResolver(whitelist *ModuleWhitelist, pageBl PageModuleBlacklist) {
	GlobalCardResolver.moduleWhitelist.Store(whitelist)
	GlobalCardResolver.pageBlacklist.Store(pageBl)
}

func (cr *cardResolver) Register(category int64, pr resolver.Resolver) {
	if _, ok := cr.resolvers[category]; ok {
		log.Warn("resolver=%+v has already registered", category)
	}
	cr.resolvers[category] = pr
}

func (cr *cardResolver) getResolver(category int64) (resolver.Resolver, bool) {
	r, ok := cr.resolvers[category]
	return r, ok
}

func (cr *cardResolver) Resolve(c context.Context, ss *kernel.Session, page *natpagegrpc.NativePage, modules []*natpagegrpc.Module) []config.BaseCfgManager {
	cfgs := make([]config.BaseCfgManager, 0, len(modules))
	for _, v := range modules {
		if v == nil || v.NativeModule == nil {
			log.Warnc(c, "natpagegrpc.Module is empty")
			continue
		}
		if !cr.moduleAllowed(v.NativeModule.Category) {
			continue
		}
		if !cr.pageModuleAllowed(page.Type, v.NativeModule.Category) {
			continue
		}
		r, ok := cr.getResolver(v.NativeModule.Category)
		if !ok {
			log.Warnc(c, "resolver not found, category=%+v", v.NativeModule.Category)
			continue
		}
		if cfg := r.Resolve(c, ss, page, v); cfg != nil {
			cfgs = append(cfgs, cfg)
		}
	}
	return cfgs
}

func (cr *cardResolver) moduleAllowed(category int64) bool {
	whitelist := cr.moduleWhitelist.Load().(*ModuleWhitelist)
	if whitelist == nil {
		return false
	}
	if whitelist.AllowAll {
		return true
	}
	cat := strconv.FormatInt(category, 10)
	if allowed, ok := whitelist.Category[cat]; ok && allowed {
		return true
	}
	return false
}

func (cr *cardResolver) pageModuleAllowed(pageType, category int64) bool {
	blacklist := cr.pageBlacklist.Load().(PageModuleBlacklist)
	list, ok := blacklist[strconv.FormatInt(pageType, 10)]
	if !ok {
		return true
	}
	if disabled, ok := list[strconv.FormatInt(category, 10)]; ok && disabled {
		return false
	}
	return true
}
