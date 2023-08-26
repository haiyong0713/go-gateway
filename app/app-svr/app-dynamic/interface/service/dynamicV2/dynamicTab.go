package dynamicV2

import (
	"context"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

var (
	tabVideo = &api.DynTab{Title: "视频", Uri: "bilibili://following/index/8", Anchor: "video"}
	tabAll   = &api.DynTab{Title: "综合", Uri: "bilibili://following/index/268435455", DefaultTab: true, Anchor: "all"}
	//tabAllV2 = &api.DynTab{Title: "关注", Uri: "bilibili://following/index/268435455", DefaultTab: true, Anchor: "all"}
)

// DynTab Tab页展示（支持同城和校园的tab）
func (s *Service) DynTab(c context.Context, general *mdlv2.GeneralParam, req *api.DynTabReq) (*api.DynTabReply, error) {
	var (
		tab *dyncampusgrpc.TabShowReply
	)
	if !general.GetDisableRcmd() {
		var sideErr error
		tab, sideErr = s.dynDao.DynTabShow(c, &dyncampusgrpc.TabShowReq{
			FromType: mdlv2.ToCampusFromType(req.GetFromType()),
			IpAddr:   general.IP, Uid: uint64(general.Mid),
		})
		if sideErr != nil {
			log.Errorc(c, "s.dynDao.DynTabShow error(%+v)", sideErr)
		}
	}
	res := &api.DynTabReply{}
	dynTabAll := tabAll
	// 青少年只返回综合tab
	if req.TeenagersMode == 1 {
		res.DynTab = append(res.DynTab, dynTabAll)
		return res, nil
	}

	res.DynTab = append(res.DynTab, tabVideo, dynTabAll)
	// 国际版只返回视频和综合tab
	if general.Device.MobiApp() == "iphone_i" || general.Device.MobiApp() == "android_i" {
		return res, nil
	}

	// 处理动态tab展示内容（校园/同城）
	res.DynTab = append(res.DynTab, dynShowTabContent(tab)...)
	// 筛选器tab + AB实验逻辑
	s.dynAllFilterAbtest(c, general, req, res)

	return res, nil
}

func dynShowTabContent(tab *dyncampusgrpc.TabShowReply) []*api.DynTab {
	// 2021/12/23: 同城全部下线
	// 硬屏蔽掉同城和二选一tab下发
	const (
		_dynShowNoTabType     = 0 // 0:不展示
		_dynShowCityTabType   = 1 // 1:展示同城
		_dynShowCampusTabType = 2 // 2:展示校园
		_dynShowChoiceTabType = 3 // 3:二选一模式(tab全部下发）
	)
	var res []*api.DynTab
	if tab == nil {
		return nil
	}
	switch tab.ShowType {
	case _dynShowNoTabType:
		return nil
	//case _dynShowCityTabType:
	//	if cityTab, openCityTab := dynShowCityTab(tab.CityTab); openCityTab {
	//		res = append(res, cityTab)
	//	}
	case _dynShowCampusTabType:
		if campusTab, openCampusTab := dynShowCampusTab(tab.CampusTab); openCampusTab {
			res = append(res, campusTab)
		}
	//case _dynShowChoiceTabType:
	//	if choiceTab, openChoiceTab := dynShowChoiceTab(tab); openChoiceTab {
	//		res = append(res, choiceTab)
	//	}
	default:
		log.Error("Unrecognized dynShowTabContent tab.ShowType: %d", tab.ShowType)
	}
	return res
}

// dynShowCampusTab 处理校园tab内容
func dynShowCampusTab(campusTab *dyncampusgrpc.CampusTabInfo) (*api.DynTab, bool) {
	const (
		_campusTabSwitchClose = 0
	)
	// 功能关闭的时候，不显示校园tab
	if campusTab == nil || campusTab.GetSwitch() == _campusTabSwitchClose {
		return nil, false
	}
	return &api.DynTab{
		Title:    campusTab.GetTabName(),
		Uri:      "bilibili://campus/home",
		Bubble:   campusTab.GetBubbleDesc(),
		RedPoint: int32(campusTab.GetRedPoint()),
		IsPopup:  int32(campusTab.GetNeedAsk()),
		Popup: &api.Popup{
			Title: "bilibili校园开启邀请函",
			Desc:  "发现校园新鲜事，一键三连校友的视频动态，快来加入吧~",
		},
		Anchor:       "campus",
		InternalTest: "内测",
	}, true
}

// SetDecision 校园-设置同城校园二选一结果
func (s *Service) SetDecision(c context.Context, general *mdlv2.GeneralParam, req *api.SetDecisionReq) (*api.NoReply, error) {
	args := &dyncampusgrpc.SetDecisionReq{
		Uid:      uint64(general.Mid),
		Result:   uint32(req.Result),
		FromType: mdlv2.ToCampusFromType(req.GetFromType()),
	}
	if err := s.dynDao.SetDecision(c, args); err != nil {
		log.Error("SetDecision mid(%v) SetDecision(), error %v", general.Mid, err)
		return nil, err
	}
	return &api.NoReply{}, nil
}

// SubscribeCampus 校园-设置预约校园开放通知
func (s *Service) SubscribeCampus(c context.Context, general *mdlv2.GeneralParam, req *api.SubscribeCampusReq) (*api.NoReply, error) {
	args := &dyncampusgrpc.SubscribeReq{
		FromType:   mdlv2.ToCampusFromType(req.GetFromType()),
		Uid:        uint64(general.Mid),
		CampusId:   uint64(req.CampusId),
		CampusName: req.CampusName,
	}
	if err := s.dynDao.SubscribeCampus(c, args); err != nil {
		log.Error("SubscribeCampus mid(%v) SubscribeCampus(), error %v", general.Mid, err)
		return nil, err
	}
	return &api.NoReply{}, nil
}

// SetRcntCampus 校园-设置访问的校园
func (s *Service) SetRecentCampus(c context.Context, general *mdlv2.GeneralParam, req *api.SetRecentCampusReq) (*api.NoReply, error) {
	args := &dyncampusgrpc.SetRecentReq{
		FromType:   mdlv2.ToCampusFromType(req.GetFromType()),
		Uid:        uint64(general.Mid),
		CampusId:   uint64(req.CampusId),
		CampusName: req.CampusName,
	}
	if err := s.dynDao.SetRcntCampus(c, args); err != nil {
		log.Error("SetRcntCampus mid(%v) SetRcntCampus(), error %v", general.Mid, err)
		return nil, err
	}
	return &api.NoReply{}, nil
}
