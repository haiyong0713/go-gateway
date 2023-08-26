package history

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"

	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
)

var (
	//历史记录tab
	historyBusTabV2 = []*history.BusTab{
		{
			Business: "all",
			Name:     "全部",
			Router:   "bilibili://main/history/all",
		},
		{
			Business: "archive",
			Name:     "视频",
			Router:   "bilibili://main/history/video",
		},
		{
			Business: "live",
			Name:     "直播",
			Router:   "bilibili://main/history/live",
		},
		{
			Business: "article",
			Name:     "专栏",
			Router:   "bilibili://main/history/article",
		},
	}
	//搜索记录tab
	searchBusTabV2 = []*history.BusTab{
		{
			Business: "all",
			Name:     "全部",
			Router:   "bilibili://main/history/search/all",
		},
		{
			Business: "archive",
			Name:     "视频",
			Router:   "bilibili://main/history/search/video",
		},
		{
			Business: "live",
			Name:     "直播",
			Router:   "bilibili://main/history/search/live",
		},
		{
			Business: "article",
			Name:     "专栏",
			Router:   "bilibili://main/history/search/article",
		},
	}
)

func (s *Service) HistoryTabGRPCV2(c context.Context, mid int64, buvid string, arg *api.HistoryTabReq, isLessonMode bool) (*api.HistoryTabReply, error) {
	var (
		businesses []string
		tempBusTab []*history.BusTab
	)
	switch arg.Source {
	case api.HistorySource_history:
		tempBusTab = historyBusTab
	case api.HistorySource_shopping:
		tempBusTab = shoppingBusTab
	}
	for _, businessTab := range tempBusTab {
		businesses = append(businesses, busTabMap[businessTab.Business]...)
	}
	func() {
		dev, ok := device.FromContext(c)
		if !ok {
			return
		}
		plat := model.Plat(dev.RawMobiApp, dev.Device)
		if model.IsPinkAndBlue(plat) {
			businesses = append(businesses, _cheeseStr)
		}
		if plat == model.PlatIPad || plat == model.PlatIpadHD {
			businesses = append(businesses, _cheeseIPadStr)
		}
	}()

	if len(businesses) == 0 {
		return nil, ecode.RequestErr
	}
	query := &history.SearchQuery{
		Businesses: businesses,
		Mid:        mid,
		Buvid:      buvid,
		Keyword:    arg.Keyword,
	}
	var (
		rawHisTabMap map[string]bool
		err          error
		isSearch     = arg.Keyword != ""
	)
	if !isSearch { //历史记录list页
		rawHisTabMap, err = s.historyDao.HasHistory(c, query)
	} else { //历史记录搜索页
		rawHisTabMap, err = s.historyDao.SearchHasHistory(c, query)
	}
	if err != nil {
		log.Error("HistoryTabGRPCV2(%+v) mid(%v) buvid(%v)", err, mid, buvid)
		return nil, err
	}
	if len(rawHisTabMap) == 0 {
		return &api.HistoryTabReply{}, nil
	}
	s.gameTabFilter(c, rawHisTabMap)
	listHasValue, vipHasValue, gameHasValue, focusBusiness := findTabsAndFocus(rawHisTabMap, isLessonMode)
	return buildListResTabs(isSearch, listHasValue, vipHasValue, gameHasValue, focusBusiness)
}

func (s *Service) gameTabFilter(ctx context.Context, in map[string]bool) {
	if !in[_game] { //没有游戏tab则不需要判断是否关闭游戏tab
		return
	}
	tabSwitch, err := s.gameDao.GameTabSwitch(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	const gameTabClose = 0
	if tabSwitch == gameTabClose {
		in[_game] = false
	}
	return
}

func findTabsAndFocus(rawHisTabMap map[string]bool, isLessonMode bool) (bool, bool, bool, string) {
	var (
		vipHasValue, listHasValue, gameHasValue bool
		vipFocusMap                             = make(map[string]struct{}, len(shoppingBusTab))
		focusBusiness                           string
	)
	//将历史记录细分的8种tab类型聚合成6种tab类型
	for tab, hasContent := range rawHisTabMap {
		if !hasContent {
			continue
		}
		value, ok := historyBusinessToBusTabMap[tab]
		if !ok {
			continue
		}
		switch value {
		case _game:
			if !isLessonMode {
				gameHasValue = true
			}
		case _goods, _show:
			if !isLessonMode {
				vipHasValue = true
				vipFocusMap[value] = struct{}{}
			}
		case _arcStr, _artStr, _liveStr, _cheeseStr, _cheeseIPadStr:
			listHasValue = true
			focusBusiness = _allStr
		default:
			continue
		}
	}
	if focusBusiness == _allStr {
		return listHasValue, vipHasValue, gameHasValue, focusBusiness
	}
	//没有定位到全部tab，先尝试在商业的tab下定位
	for _, v := range shoppingBusTab {
		if _, ok := vipFocusMap[v.Business]; ok {
			focusBusiness = v.Business
			return listHasValue, vipHasValue, gameHasValue, focusBusiness
		}
	}
	//没有定位到全部和商业tab，在游戏tab下定位
	for _, v := range gameBusTab {
		if gameHasValue {
			focusBusiness = v.Business
			break
		}
	}
	return listHasValue, vipHasValue, gameHasValue, focusBusiness
}

func buildListResTabs(isSearch, listHasValue, vipHasValue, gameHasValue bool, focusBusiness string) (*api.HistoryTabReply, error) {
	var res = new(api.HistoryTabReply)
	tempHisBusTabs := historyBusTabV2
	tempShopBusTabs := shoppingBusTab
	temGameBusTabs := gameBusTab
	if isSearch {
		tempHisBusTabs = searchBusTabV2
		tempShopBusTabs = shoppingBusSearchTab
	}
	//如果视频，专栏，直播任一tab有内容则出三个tab，定位在全部tab上
	//如果商业任一tab下有内容则出商业所有tab
	if listHasValue {
		for _, v := range tempHisBusTabs {
			res.Tab = append(res.Tab, &api.CursorTab{
				Business: v.Business,
				Name:     v.Name,
				Router:   v.Router,
				Focus:    v.Business == focusBusiness,
			})
		}
	}
	if gameHasValue {
		for _, v := range temGameBusTabs {
			res.Tab = append(res.Tab, &api.CursorTab{
				Business: v.Business,
				Name:     v.Name,
				Router:   v.Router,
				Focus:    v.Business == focusBusiness,
			})
		}
	}
	if vipHasValue {
		for _, v := range tempShopBusTabs {
			res.Tab = append(res.Tab, &api.CursorTab{
				Business: v.Business,
				Name:     v.Name,
				Router:   v.Router,
				Focus:    v.Business == focusBusiness,
			})
		}
	}
	return res, nil
}
