package builder

import (
	"context"
	"fmt"
	"time"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Tab struct{}

func (bu Tab) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	tabCfg, ok := cfg.(*config.Tab)
	if !ok {
		logCfgAssertionError(config.Tab{})
		return nil
	}
	items := bu.buildModuleItems(tabCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeTab.String(),
		ModuleId:      tabCfg.ModuleBase().ModuleID,
		ModuleColor:   bu.buildColor(tabCfg),
		ModuleSetting: &api.Setting{DisplayUnfoldButton: tabCfg.DisplayUnfoldButton},
		ModuleItems:   items,
		ModuleUkey:    tabCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Tab) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Tab) buildColor(cfg *config.Tab) *api.Color {
	if bu.style(cfg.Style) != api.TabStyle_TabStyleColor {
		return nil
	}
	return &api.Color{
		BgColor:             cfg.BgColor,
		SelectedFontColor:   cfg.SelectedFontColor,
		UnselectedFontColor: cfg.UnselectedFontColor,
	}
}

func (bu Tab) buildModuleItems(cfg *config.Tab, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	if len(cfg.Items) == 0 {
		return nil
	}
	cd := &api.TabCard{
		Style: bu.style(cfg.Style),
		Items: make([]*api.TabItem, 0, len(cfg.Items)),
	}
	if cd.Style == api.TabStyle_TabStyleImage {
		cd.BgImage = bu.bgImage(&cfg.BgImage)
	}
	var (
		tabIndex   int64
		findCurTab bool
	)
	defaultTab := bu.defaultTab(cfg.Items)
	for _, item := range cfg.Items {
		page, ok := material.NativePages[item.PageID]
		if !ok || !page.IsOnline() || page.Title == "" {
			continue
		}
		tabItem := &api.TabItem{
			PageId: item.PageID,
			Title:  page.Title,
		}
		isLocked, ok := bu.handleLock(page, tabItem)
		if !ok {
			continue
		}
		if cd.Style == api.TabStyle_TabStyleImage {
			if isLocked {
				tabItem.UnselectedImage = item.UnI.ToSizeImage()
			} else {
				tabItem.SelectedImage = item.SI.ToSizeImage()
				tabItem.UnselectedImage = item.UnSI.ToSizeImage()
			}
		}
		if !isLocked {
			// 当【页面URL含有定位参数】与【页面设置默认tab】同时存在时，则优先以页面URL的定位参数为准
			if ss.CurrentTab != "" && bu.currentTab(item) == ss.CurrentTab {
				findCurTab = true
				cd.CurrentTab = tabIndex
			}
			// 没有指定定位 && 没有锁定 && 有默认tab
			if !findCurTab && item.PageID == defaultTab {
				cd.CurrentTab = tabIndex
			}
		}
		cd.Items = append(cd.Items, tabItem)
		tabIndex++
	}
	if len(cd.Items) == 0 {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeTab.String(),
		CardDetail: &api.ModuleItem_TabCard{TabCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}

func (bu Tab) style(cfgStyle int64) api.TabStyle {
	style := api.TabStyle_TabStyleColor
	if cfgStyle == 1 {
		style = api.TabStyle_TabStyleImage
	}
	return style
}

func (bu Tab) bgImage(cfg *config.SizeImage) *api.SizeImage {
	if cfg == nil {
		return nil
	}
	bgImage := cfg.ToSizeImage()
	if bgImage.Image == "" || bgImage.Width <= 0 || bgImage.Height <= 0 {
		bgImage.Width = 1125
		bgImage.Height = 120
	}
	return bgImage
}

func (bu Tab) defaultTab(tabs []*config.TabItem) int64 {
	var defTab, timingTab int64
	now := time.Now().Unix()
	for _, tab := range tabs {
		switch tab.DefType {
		case model.EftTypeImmediately:
			defTab = tab.PageID
		case model.EftTypeTiming:
			if tab.DStime <= now && tab.DEtime > now {
				timingTab = tab.PageID
			}
		}
	}
	// 默认tab优先级，若立即生效的时间，与定时生效的时间一致，则优先以定时生效的为准
	if timingTab == 0 {
		timingTab = defTab
	}
	return timingTab
}

func (bu Tab) handleLock(page *natpagegrpc.NativePage, tabItem *api.TabItem) (isLocked bool, ok bool) {
	lock := page.ConfSetUnmarshal()
	if lock.DT != model.TabLockNeed {
		return false, true
	}
	if lock.DC == model.TabLockTypeTime && lock.Stime <= time.Now().Unix() {
		return false, true
	}
	switch lock.UnLock {
	case model.TabNotUnlockDisableClick:
		tabItem.PageId = 0
		tabItem.DisableClick = true
		tabItem.DisableClickToast = "还未解锁，敬请期待"
		if lock.Tip != "" {
			tabItem.DisableClickToast = lock.Tip
		}
		return true, true
	default:
		return false, false
	}
}

func (bu Tab) currentTab(cfg *config.TabItem) string {
	if cfg.Type == "" || cfg.LocationKey == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", cfg.Type, cfg.LocationKey)
}
