package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (s *Service) BottomTab(ctx context.Context, req *api.BottomTabReq) (rly *api.BottomTabRly, err error) {
	defer func() {
		// 兜底逻辑
		if err == nil || req.PageId <= 0 {
			return
		}
		rly = bottomTabErrLimitRly(req.PageId)
		err = nil
	}()
	// 获取物料
	modulesRly, err := s.dao.Natpage().NatTabModules(ctx, req.TabId)
	if err != nil {
		return nil, err
	}
	modules := filterTabModules(modulesRly)
	if modulesRly == nil || modulesRly.Tab == nil || len(modules) == 0 {
		return nil, ecode.NothingFound
	}
	pages := func() map[int64]*natpagegrpc.NativePage {
		pageIds := pageIdsOfTabModules(modules)
		if len(pageIds) == 0 {
			return nil
		}
		if rly, err := s.dao.Natpage().NativePages(ctx, &natpagegrpc.NativePagesReq{Pids: pageIds}); err == nil && rly != nil {
			return rly.List
		}
		return nil
	}()
	// 组装响应
	var selected *api.BottomTabItem
	items := make([]*api.BottomTabItem, 0, len(modules))
	for _, module := range modules {
		item, ok := buildBottomTabItem(module)
		if !ok {
			continue
		}
		if item.TabModuleId == req.TabModuleId {
			item.Selected = true
			selected = item
		}
		if module.IsTabPage() {
			if page, ok := pages[module.Pid]; ok && page != nil && page.IsOnline() {
				item.PageTitle = page.Title
				item.PageFid = page.ForeignID
			} else if !item.Selected {
				continue
			}
		}
		items = append(items, item)
	}
	if selected == nil {
		return nil, ecode.NothingFound
	}
	return &api.BottomTabRly{
		Tab: buildBottomTab(modulesRly.Tab, selected, items),
	}, nil
}

func bottomTabErrLimitRly(pageId int64) *api.BottomTabRly {
	return &api.BottomTabRly{
		ErrLimit: &api.BottomTabErrLimit{
			Code:    int64(ecode.NothingFound.Code()),
			Message: "当前页面状态发生变化",
			Button: &api.BottomTabErrLimit_Button{
				Content: "前往活动页面",
				Url:     fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d", pageId),
			},
		},
	}
}

func filterTabModules(rly *natpagegrpc.NatTabModulesReply) []*natpagegrpc.NativeTabModule {
	if rly == nil {
		return nil
	}
	modules := make([]*natpagegrpc.NativeTabModule, 0, len(rly.List))
	for _, module := range rly.List {
		if module == nil || !module.IsOnline() {
			continue
		}
		modules = append(modules, module)
	}
	return modules
}

func pageIdsOfTabModules(modules []*natpagegrpc.NativeTabModule) []int64 {
	pageIds := make([]int64, 0, len(modules))
	for _, module := range modules {
		if module.IsTabPage() && module.Pid > 0 {
			pageIds = append(pageIds, module.Pid)
		}
	}
	return pageIds
}

func buildBottomTab(in *natpagegrpc.NativeActTab, selected *api.BottomTabItem, items []*api.BottomTabItem) *api.BottomTab {
	tab := &api.BottomTab{
		BgType:              api.BottomTabBgType(in.BgType),
		BgImage:             in.BgImg,
		BgColor:             in.BgColor,
		IconType:            api.BottomTabIconType(in.IconType),
		SelectedFontColor:   in.ActiveColor,
		UnselectedFontColor: in.InactiveColor,
		Items:               items,
	}
	if !isBottomTabOnline(in) {
		tab.Items = []*api.BottomTabItem{selected}
	}
	return tab
}

func isBottomTabOnline(tab *natpagegrpc.NativeActTab) bool {
	if !tab.IsOnline() {
		return false
	}
	now := time.Now().Unix()
	return int64(tab.Stime) > 0 && int64(tab.Stime) <= now && (int64(tab.Etime) >= now || int64(tab.Etime) <= 0)
}

func buildBottomTabItem(in *natpagegrpc.NativeTabModule) (*api.BottomTabItem, bool) {
	item := &api.BottomTabItem{
		TabId:           in.TabID,
		TabModuleId:     in.ID,
		Title:           in.Title,
		Selected:        false,
		SelectedImage:   in.ActiveImg,
		UnselectedImage: in.InactiveImg,
		ShareOrigin:     model.ShareOriginTab,
	}
	switch {
	case in.IsTabPage():
		if in.Pid <= 0 {
			return nil, false
		}
		item.Goto = api.BottomTabGoto_BTGNaPage
		item.PageId = in.Pid
	case in.IsTabUrl():
		if in.URL == "" {
			return nil, false
		}
		item.Goto = api.BottomTabGoto_BTGRedirect
		item.Url = in.URL
	default:
		return nil, false
	}
	return item, true
}
