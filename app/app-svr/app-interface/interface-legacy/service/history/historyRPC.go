package history

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"

	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
	historyQuery "go-gateway/app/app-svr/app-interface/interface-legacy/model/history"

	v1 "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

var (
	//历史记录tab
	historyBusTab = []*history.BusTab{
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
		{
			Business: "goods",
			Name:     "商品",
			Router:   "bilibili://mall/history/goods",
		},
		{
			Business: "show",
			Name:     "展演",
			Router:   "bilibili://mall/history/ticket",
		},
		{
			Business: "game",
			Name:     "游戏",
			Router:   "bilibili://game_center/history?sourcefrom=1000240011",
		},
	}
	//搜索记录tab
	searchBusTab = []*history.BusTab{
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
		{
			Business: "goods",
			Name:     "商品",
			Router:   "bilibili://mall/history/search/goods",
		},
		{
			Business: "show",
			Name:     "展演",
			Router:   "bilibili://mall/history/search/ticket",
		},
		{
			Business: "game",
			Name:     "游戏",
			Router:   "bilibili://game_center/history?sourcefrom=1000240011",
		},
	}
	//会员购tab
	shoppingBusTab = []*history.BusTab{
		{
			Business: "goods",
			Name:     "商品",
			Router:   "bilibili://mall/history/goods",
		},
		{
			Business: "show",
			Name:     "展演",
			Router:   "bilibili://mall/history/ticket",
		},
	}
	//会员购搜索的tab
	shoppingBusSearchTab = []*history.BusTab{
		{
			Business: "goods",
			Name:     "商品",
			Router:   "bilibili://mall/history/search/goods",
		},
		{
			Business: "show",
			Name:     "展演",
			Router:   "bilibili://mall/history/search/ticket",
		},
	}
	//游戏相关tab
	gameBusTab = []*history.BusTab{
		{
			Business: "game",
			Name:     "游戏",
			Router:   "bilibili://game_center/history?sourcefrom=1000240011",
		},
	}
	// 历史记录business对应tab
	historyBusinessToBusTabMap = map[string]string{
		_artStr:        "article",
		_corpusStr:     "article",
		_arcStr:        "archive",
		_pgcStr:        "archive",
		_liveStr:       "live",
		_mallGoods:     "goods",
		_mallShow:      "show",
		_game:          "game",
		_cheeseStr:     "cheese",
		_cheeseIPadStr: "cheese-ipad",
	}
)

// HistoryTabRPC is 查询需要展示的tab
func (s *Service) HistoryTabGRPC(c context.Context, mid int64, dev device.Device, arg *api.HistoryTabReq, isLessonMode bool) (*api.HistoryTabReply, error) {
	// 确定搜索businesses
	businesses, targetBusTab := buildSearchBusinesses(arg.Source, arg.Keyword)
	// 接口查询
	query := &historyQuery.SearchQuery{
		Mid:        mid,
		Buvid:      dev.Buvid,
		Businesses: businesses,
		Keyword:    arg.Keyword,
	}
	var (
		//nolint:ineffassign
		hasMap = make(map[string]bool)
		err    error
	)
	if arg.Business == "" {
		hasMap, err = s.historyDao.HasHistory(c, query)
	} else {
		hasMap, err = s.historyDao.SearchHasHistory(c, query)
	}
	if err != nil {
		log.Error("HasHistory error(%+v) mid(%v) buvid(%v)", err, mid, dev.Buvid)
		return nil, err
	}
	var (
		resultBusTab  = make(map[string]bool)
		hasContent    bool
		vipHasContent bool
	)
	for k, v := range hasMap {
		if !v {
			continue
		}
		tab, ok := historyBusinessToBusTabMap[k]
		if ok {
			//商业有任一tag有内容则全部tag都要下发
			if tab == _goods || tab == _show {
				vipHasContent = true
			}
			resultBusTab[tab] = true
			hasContent = true
		}
	}
	// 定位到的focus的business
	focusBusiness := fetchFocusBusiness(resultBusTab, targetBusTab, arg.Business)
	//只要tag不全是空，视频，直播，专栏tag都下发,仅历史记录页
	//会员购的两个tag绑定要么都下发，要么都不下发,仅历史记录页
	if arg.Business == "" {
		resultBusTab[_arcStr] = hasContent
		resultBusTab[_liveStr] = hasContent
		resultBusTab[_artStr] = hasContent
		resultBusTab[_goods] = vipHasContent
		resultBusTab[_show] = vipHasContent
	}
	//课堂模式下搜索和历史记录页需要屏蔽商品tag
	if isLessonMode {
		resultBusTab[_goods] = false
		resultBusTab[_show] = false
	}
	res := &api.HistoryTabReply{}
	for _, v := range targetBusTab {
		exist, ok := resultBusTab[v.Business]
		if !ok || !exist {
			continue
		}
		res.Tab = append(res.Tab, &api.CursorTab{
			Business: v.Business,
			Name:     v.Name,
			Router:   v.Router,
			Focus:    v.Business == focusBusiness,
		})
	}
	return res, nil
}

// CursorV2GRPC
func (s *Service) CursorV2GRPC(ctx context.Context, mid int64, dev device.Device, plat int8, arg *api.CursorV2Req, net network.Network) (*api.CursorV2Reply, error) {
	res := &api.CursorV2Reply{
		EmptyLink: s.c.HisEmptyLink[arg.Business],
	}
	//构建历史记录查询参数
	param := buildHistoryParam(mid, arg, dev)
	liveParam := &history.LiveParam{
		Uid:        mid,
		Platform:   param.Platform,
		DeviceName: dev.Model,
		Build:      param.Build,
		NetWork:    fetchLiveNetType(net),
		ReqBiz:     _hisCursorV2,
	}
	hisCursor, hasMore, err := s.CursorV2(ctx, param, plat, liveParam)
	if err != nil {
		return nil, err
	}
	if hisCursor == nil {
		return res, nil
	}
	if hisCursor.Cursor != nil {
		res.Cursor = &api.Cursor{
			MaxTp: hisCursor.Cursor.MaxTP,
			Max:   hisCursor.Cursor.Max,
		}
	}
	if len(hisCursor.List) == 0 {
		return res, nil
	}
	//历史服务端未到底+网关测当页有数据 认为还可翻页
	res.HasMore = hasMore
	res.Items = s.buildGRPCRes(ctx, hisCursor.List, "", s.c.Custom.HisHasShare)
	return res, nil
}

func fetchLiveNetType(net network.Network) string {
	switch net.Type {
	case network.TypeWIFI:
		return _liveWifi
	case network.TypeCellular:
		return _liveMobile
	default:
		return _liveOther
	}
}

// CursorV2 for history
func (s *Service) CursorV2(c context.Context, param *history.HisParam, plat int8, liveParam *history.LiveParam) (*history.ListCursor, bool, error) {
	businesses, ok := busTabMap[param.Business]
	if !ok {
		log.Error("historyCursor invalid business(%s)", param.Business)
		return nil, false, ecode.RequestErr
	}
	if s.cheeseDao.HasCheese(plat, int(param.Build), false) && (param.Business == "all" || param.Business == "archive") {
		businesses = append(businesses, _cheeseStr)
	}
	if ((plat == model.PlatIPad && param.Build > int64(s.c.BuildLimit.IPadCheese)) || plat == model.PlatAndroidHD || (plat == model.PlatIpadHD && param.Build > int64(s.c.BuildLimit.IPadHDCheese))) &&
		(param.Business == "archive" || param.Business == "all") {
		businesses = append(businesses, _cheeseIPadStr)
	}
	var paramMaxBus string
	if _, ok := businessMap[param.MaxTP]; ok {
		paramMaxBus = businessMap[param.MaxTP]
	}
	var (
		res []*v1.ModelResource
		err error
	)
	if param.Islocal && param.Mid > 0 { //登录态和选择本机历史记录走新接口
		deviceType := buildDeviceType(param.MobiApp, param.Device)
		res, err = s.historyDao.NativeHistory(c, param.Mid, businesses, param.Buvid, deviceType, param.Max, param.Ps, paramMaxBus)
		if err != nil {
			log.Error("s.historyDao.NativeHistory error(%+v) ", err)
			return nil, false, err
		}
	} else {
		res, err = s.historyDao.Cursor(c, param.Mid, param.Max, param.Ps, paramMaxBus, businesses, param.Buvid)
		if err != nil {
			log.Error("s.historyDao.Cursor error(%+v) ", err)
			return nil, false, err
		}
	}
	if len(res) == 0 {
		return nil, false, err
	}
	var hasMore bool
	if len(res) >= int(param.Ps) {
		hasMore = true
	}

	data := &history.ListCursor{
		List: []*history.ListRes{},
	}
	data.List = s.TogetherHistory(c, res, param.Mid, param.Build, plat, param.MobiApp, true, liveParam)
	if len(data.List) >= int(param.Ps) {
		data.List = data.List[:param.Ps]
	}
	if len(data.List) > 0 {
		data.Cursor = &history.Cursor{
			Max:   data.List[len(data.List)-1].ViewAt,
			MaxTP: data.List[len(data.List)-1].History.Tp,
			Ps:    param.Ps,
		}
	}
	return data, hasMore, nil
}

func buildDeviceType(mobiApp string, device string) int8 {
	switch mobiApp {
	case "android", "android_G", "android_i", "android_b":
		return model.DeviceAndroid
	case "iphone", "iphone_b", "iphone_i":
		if device == "pad" {
			return model.DeviceIpad
		}
		return model.DeviceIphone
	case "ipad", "ipad_i":
		return model.DeviceIpad
	default:
		return model.DeviceUnknown
	}
}

// buildHistoryParam 构建查询参数
func buildHistoryParam(mid int64, arg *api.CursorV2Req, dev device.Device) *historyQuery.HisParam {
	param := &history.HisParam{
		Ps:  int32(20),
		Mid: mid,
	}
	if arg.Cursor != nil {
		param.Max = arg.Cursor.Max
		param.MaxTP = arg.Cursor.MaxTp
	}
	//无参数 默认优先级最高的 视频
	if arg.Business == "" {
		arg.Business = "archive"
	}
	param.Business = arg.Business
	param.Build = dev.Build
	param.Platform = dev.RawPlatform
	param.Device = dev.Device
	param.Buvid = dev.Buvid
	param.MobiApp = dev.RawMobiApp
	param.Islocal = arg.IsLocal
	return param
}

// fetchFocusBusiness 确定定位到的focus
func fetchFocusBusiness(resultBusTab map[string]bool, targaetBusTab []*historyQuery.BusTab, searchBusiness string) string {
	_, focus := resultBusTab[searchBusiness]
	if focus {
		return searchBusiness
	}
	var focusBusiness string
	for _, sortTab := range targaetBusTab {
		_, ok := resultBusTab[sortTab.Business]
		if ok {
			focusBusiness = sortTab.Business
			break
		}
	}
	return focusBusiness
}

// buildSearchBusinesses 构建查询businesses
func buildSearchBusinesses(source api.HistorySource, keyword string) ([]string, []*history.BusTab) {
	var targetBusTab []*history.BusTab
	if source == api.HistorySource_history {
		if keyword == "" {
			targetBusTab = historyBusTab
		} else {
			targetBusTab = searchBusTab
		}
	}
	if source == api.HistorySource_shopping {
		if keyword == "" {
			targetBusTab = shoppingBusTab
		} else {
			targetBusTab = shoppingBusSearchTab
		}
	}
	var businesses []string
	//组装实际请求businesses
	for _, tab := range targetBusTab {
		businesses = append(businesses, busTabMap[tab.Business]...)
	}
	return businesses, targetBusTab
}
