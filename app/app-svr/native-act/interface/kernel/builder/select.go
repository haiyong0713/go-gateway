package builder

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Select struct{}

func (se Select) buildColor(cfg *config.Select) *api.Color {
	return &api.Color{
		BgColor:                cfg.BgColor,
		TopFontColor:           cfg.TopFontColor,
		PanelBgColor:           cfg.PanelBgColor,
		PanelSelectColor:       cfg.PanelSelectColor,
		PanelSelectFontColor:   cfg.PanelSelectFontColor,
		PanelNtSelectFontColor: cfg.PanelNtSelectFontColor,
	}
}

func (se Select) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	pgCfg, ok := cfg.(*config.Select)
	if !ok {
		logCfgAssertionError(config.Select{})
		return nil
	}
	items := se.buildModuleItems(pgCfg, ss, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeSelect.String(),
		ModuleId:    pgCfg.ModuleBase().ModuleID,
		ModuleColor: se.buildColor(pgCfg),
		ModuleItems: items,
		ModuleUkey:  pgCfg.ModuleBase().Ukey,
	}
	return module
}

func (se Select) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (se Select) defaultTab(tabs []*config.SelectItem) int64 {
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

func (se Select) currentTab(cfg *config.SelectItem) string {
	if cfg.Type == "" || cfg.LocationKey == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", cfg.Type, cfg.LocationKey)
}

// nolint:gocognit
func (se Select) buildModuleItems(cfg *config.Select, ss *kernel.Session, material *kernel.Material) []*api.ModuleItem {
	if len(cfg.Items) == 0 {
		return nil
	}
	cd := &api.SelectCard{}
	var (
		tabIndex   int64
		findCurTab bool
	)
	defaultTab := se.defaultTab(cfg.Items)
	for _, v := range cfg.Items {
		if v == nil {
			continue
		}
		pageInfo, ok := material.NativePages[v.PageID]
		if !ok || pageInfo == nil || !pageInfo.IsOnline() || pageInfo.Title == "" {
			continue
		}
		tabItem := &api.SelectItem{
			PageId: v.PageID,
			Title:  pageInfo.Title,
		}
		func() {
			if v.SelectExt.LocationKey == "" {
				return
			}
			weekid, e := strconv.ParseInt(v.SelectExt.LocationKey, 10, 64)
			if e != nil {
				return
			}
			var title, desc, img string
			switch v.SelectExt.Type {
			case natpagegrpc.SelectWeek:
				// 获取每周必看的share数据
				weekInfo, ok := material.WeekCard[weekid]
				if !ok || weekInfo == nil {
					return
				}
				title = weekInfo.ShareTitle
				desc = weekInfo.ShareSubtitle
				img = "https://i0.hdslb.com/bfs/activity-plat/static/20200121/df3e2ff90b315fca2f8d24a29cb68a47/mvxKMWL-V.png"
			case natpagegrpc.SelectMiao:
				if cfg.ShareInfo == nil {
					return
				}
				title = cfg.ShareInfo.Title
				desc = cfg.ShareInfo.Caption
				img = cfg.ShareInfo.Image
			default:
				return
			}
			tabItem.PageShare = &api.PageShare{
				Image: img,                     //分享图
				Type:  model.ShareTypeActivity, //分享类型
				Title: title,                   //分享内容
				Desc:  desc,                    //分享文案
			}
			if ss.ShareReq != nil && ss.ShareReq.ShareOrigin == model.ShareOriginTab && ss.ShareReq.TabID > 0 && ss.ShareReq.TabModuleID > 0 {
				//"pageid,tabid,tabModuleId,type,id,current_tab"
				tabItem.PageShare.Sid = fmt.Sprintf("%d,%d,%d,%s,%d,%s-%s", cfg.PrimaryPageID, ss.ShareReq.TabID, ss.ShareReq.TabModuleID, v.SelectExt.Type, weekid, v.SelectExt.Type, v.SelectExt.LocationKey)
				tabItem.PageShare.Origin = model.ShareOriginInlineTab
				tabItem.PageShare.InsideUri = fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d&current_tab=%s-%s", cfg.PrimaryPageID, ss.ShareReq.TabID, ss.ShareReq.TabModuleID, time.Now().Unix(), v.SelectExt.Type, v.SelectExt.LocationKey)
			} else {
				//"pageid,type,id,current_tab"
				tabItem.PageShare.Sid = fmt.Sprintf("%d,%s,%d,%s-%s", cfg.PrimaryPageID, v.SelectExt.Type, weekid, v.SelectExt.Type, v.SelectExt.LocationKey)
				tabItem.PageShare.Origin = model.ShareOriginSimpleInlineTab
				tabItem.PageShare.InsideUri = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d&current_tab=%s-%s", cfg.PrimaryPageID, time.Now().Unix(), v.SelectExt.Type, v.SelectExt.LocationKey) //分享动态增加时间戳参数
			}
			if tabItem.PageShare.OutsideUri == "" {
				tabItem.PageShare.OutsideUri = tabItem.PageShare.InsideUri
			}
		}()
		// 当【页面URL含有定位参数】与【页面设置默认tab】同时存在时，则优先以页面URL的定位参数为准
		if ss.CurrentTab != "" && se.currentTab(v) == ss.CurrentTab {
			findCurTab = true
			cd.CurrentTab = tabIndex
		}
		// 没有指定定位  && 有默认tab
		if !findCurTab && v.PageID == defaultTab {
			cd.CurrentTab = tabIndex
		}
		tabIndex++
		cd.Items = append(cd.Items, tabItem)
	}
	if len(cd.Items) == 0 {
		return nil
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeSelect.String(),
		CardDetail: &api.ModuleItem_SelectCard{SelectCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}
