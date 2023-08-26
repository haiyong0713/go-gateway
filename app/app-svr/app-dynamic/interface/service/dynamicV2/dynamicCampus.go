package dynamicV2

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

const (
	_feedTab = 1
	_rcmd    = 2
)

func (s *Service) CampusRcmd(c context.Context, general *mdlv2.GeneralParam, req *api.CampusRcmdReq) (*api.CampusRcmdReply, error) {
	// 请求服务端获取基础数据
	reply, err := s.dynDao.NearbyRcmd(c, req, general)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	res := &api.CampusRcmdReply{PageType: int32(reply.PageType), JumpHomePop: int32(reply.JumpHomePop)}
	switch reply.PageType {
	case _feedTab:
		top := s.campusTop(c, general, reply.MajorPage)
		if top == nil {
			return res, nil
		}
		res.CampusTop = top
	case _rcmd:
		top, rcmd, hotTopic, err := s.campusRcmdPage(c, general, reply.NearbyRcmdPage)
		if err != nil {
			return nil, err
		}
		if top == nil || rcmd == nil {
			return res, nil
		}

		res.Top = top
		res.Rcmd = rcmd
		if res.Top != nil {
			res.Top.RcmdTopic = hotTopic.ToV2CampusHomeRcmdTopic()
		}
	default:
		log.Warn("CampusRcmd miss page_type(%d)", reply.PageType)
		return res, nil
	}
	return res, nil
}

func (s *Service) campusRcmdPage(c context.Context, general *mdlv2.GeneralParam, rcmdInfo *dyncampusgrpc.NearbyRcmdInfo) (top *api.CampusRcmdTop, rcmd *api.CampusRcmdInfo, hotTopic *mdlv2.CampusHotTopicInfo, err error) {
	if rcmdInfo == nil {
		return nil, nil, nil, nil
	}
	// 解析基础数据&获取物料
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, campusRcmd: rcmdInfo.RcmdList})
	if err != nil {
		return nil, nil, nil, err
	}
	// 聚合&返回
	top, rcmd, hotTopic = s.proCampusRcmd(dynCtx, general, rcmdInfo)
	return
}

const (
	_campusNotChosen     = 0
	_campusSubscribed    = 1
	_campusNonSubscribed = 2
	// 1和2，在预约满了情况下，又命中管控学校，就会变成3
	_campusUnderControlled = 3
)

func (s *Service) proCampusRcmd(dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, campusRcmd *dyncampusgrpc.NearbyRcmdInfo) (top *api.CampusRcmdTop, rcmd *api.CampusRcmdInfo, hotTopic *mdlv2.CampusHotTopicInfo) {
	if campusRcmd == nil {
		return nil, nil, nil
	}
	if campusRcmd.TopShow == nil {
		log.Warn("NearbyCampusRcmd get TopShowInfo nil")
	}
	if campusRcmd.TopShow != nil {
		if general.IsIPhonePick() && general.GetBuild() >= 67100000 || general.IsAndroidPick() && general.GetBuild() >= 6710000 {
			top = s.campusRcmdTopShowNew(dynCtx, campusRcmd)
		} else {
			top = s.campusRcmdTopShowOld(dynCtx, campusRcmd)
		}
	}
	if campusRcmd.RcmdList == nil {
		log.Warn("NearbyCampusRcmd get RcmdCampusInfo nil")
	}
	if campusRcmd.HotForumCard != nil {
		hotTopic = &mdlv2.CampusHotTopicInfo{FeedHot: campusRcmd.HotForumCard}
	}
	var items []*api.CampusRcmdItem
	for _, campus := range campusRcmd.RcmdList.List {
		if campus == nil {
			continue
		}
		var rcmdTmps []*api.RcmdItem
		for _, aid := range campus.Aids {
			ap, ok := dynCtx.GetArchive(int64(aid))
			if !ok || !ap.Arc.IsNormal() {
				continue
			}
			// 付费合集
			if mdlv2.PayAttrVal(ap.Arc) {
				continue
			}
			rcmdArcTmp := &api.RcmdArchive{
				Title:           ap.Arc.GetTitle(),
				Cover:           ap.Arc.GetPic(),
				CoverLeftIcon_1: api.CoverIcon_cover_icon_play,
				CoverLeftText_1: s.numTransfer(int(ap.Arc.Stat.View)),
				CoverLeftIcon_2: api.CoverIcon_cover_icon_danmaku,
				CoverLeftText_2: s.numTransfer(int(ap.Arc.Stat.Danmaku)),
				CoverLeftIcon_3: api.CoverIcon_cover_icon_none,
				CoverLeftText_3: s.videoDuration(ap.Arc.Duration),
				Uri:             model.FillURI(model.GotoAv, strconv.FormatInt(ap.Arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, ap.Arc.FirstCid, true)),
				Aid:             ap.Arc.Aid,
			}
			// UGC转PGC逻辑
			if ap.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && ap.Arc.RedirectURL != "" {
				rcmdArcTmp.IsPgc = true
				rcmdArcTmp.Uri = ap.Arc.RedirectURL
			}
			rcmdTmp := &api.RcmdItem{
				Type: api.RcmdType_rcmd_archive,
				RcmdItem: &api.RcmdItem_RcmdArchive{
					RcmdArchive: rcmdArcTmp,
				},
			}
			rcmdTmps = append(rcmdTmps, rcmdTmp)
			// 每个学校最多展示2个
			// nolint:gomnd
			if len(rcmdTmps) == 2 {
				break
			}
		}
		// 过滤小于2个内容的学校
		// nolint:gomnd
		if len(rcmdTmps) < 2 {
			continue
		}
		items = append(items, &api.CampusRcmdItem{
			CampusId: int64(campus.CampusId),
			Title:    campus.CampusName,
			EntryLabel: &api.CampusLabel{
				Text: "进入",
				Url:  fmt.Sprintf("bilibili://campus/detail/%d", campus.CampusId),
			},
			Items: rcmdTmps,
		})
	}
	rcmd = &api.CampusRcmdInfo{
		Title: campusRcmd.RcmdList.Title,
		Items: items,
	}
	return top, rcmd, hotTopic
}

func (s *Service) campusRcmdTopShowOld(_ *mdlv2.DynamicContext, campusRcmd *dyncampusgrpc.NearbyRcmdInfo) *api.CampusRcmdTop {
	top := &api.CampusRcmdTop{
		CampusId:   int64(campusRcmd.TopShow.CampusId),
		CampusName: campusRcmd.TopShow.CampusName,
		Title:      "你还未添加校园",
		Desc:       "立即添加，抢先了解校园新动态哦",
		Type:       int32(campusRcmd.TopShow.ShowType),
		InviteLabel: &api.CampusLabel{
			Text: campusRcmd.TopShow.InviteDesc,
			Url:  s.c.Resource.Others.SchoolInviteURI,
		},
	}
	// 预约信息
	if subInfo := campusRcmd.TopShow.SubInfo; subInfo != nil {
		reserveNumber := subInfo.UserNum
		// nolint:gomnd
		if reserveNumber < 2 {
			reserveNumber = 2
		}
		top.ReserveLabel = &api.CampusLabel{
			Text: fmt.Sprintf("已有%d个校友预约", reserveNumber),
			Url:  subInfo.JumpUrl,
		}
		top.ReserveNumber = int64(subInfo.UserNum)
		top.MaxReserve = subInfo.MaxReserve
	}
	if mngInfo := campusRcmd.TopShow.ManagementInfo; mngInfo != nil {
		top.MngLabel = &api.CampusLabel{
			Text: mngInfo.Desc,
			Url:  mngInfo.JumpUrl,
		}
	}
	switch campusRcmd.TopShow.ShowType {
	case _campusNotChosen: // 未选择学校
		top.Button = &api.RcmdTopButton{
			Text: "去添加",
			Url:  "bilibili://campus/search",
		}
	case _campusNonSubscribed, _campusUnderControlled: // 选了学校未预约
		if campusRcmd.TopShow.ShowType == _campusNonSubscribed {
			top.NoticeLabel = &api.CampusLabel{
				Text: "一键预约",
			}
			top.Desc_3 = "点击下方按钮，第一时间获取上线提醒"
		}
		fallthrough
	default:
		top.Title = campusRcmd.TopShow.CampusName
		// top.Desc = fmt.Sprintf("%sbilibili校园即将上线", campusRcmd.TopShow.CampusName)
		top.Desc = "校园上线后将私信通知你"
		top.Desc_2 = "召唤校友来预约，可以加速校园上线哦~"
		top.SwitchLabel = &api.CampusLabel{
			Text: "切换",
			Url:  "bilibili://campus/search",
		}
		// 管控
		if campusRcmd.TopShow.ShowType == _campusUnderControlled {
			top.Type = _campusSubscribed
			top.AuditBeforeOpen = true
			top.AuditMessage = "受相关部门要求，板块需与学校上级主管部门报备后才能开放，敬请期待哦"
		}
	}
	return top
}

func (s *Service) campusRcmdTopShowNew(_ *mdlv2.DynamicContext, campusRcmd *dyncampusgrpc.NearbyRcmdInfo) *api.CampusRcmdTop {
	top := &api.CampusRcmdTop{
		CampusId:   int64(campusRcmd.TopShow.CampusId),
		CampusName: campusRcmd.TopShow.CampusName,
		Title:      "你还未添加校园",
		Desc:       "立即添加，抢先了解校园新动态哦",
		Type:       int32(campusRcmd.TopShow.ShowType),
		InviteLabel: &api.CampusLabel{
			Text: campusRcmd.TopShow.InviteDesc,
			Url:  s.c.Resource.Others.SchoolInviteURI,
		},
	}
	// 预约信息
	if subInfo := campusRcmd.TopShow.SubInfo; subInfo != nil {
		top.ReserveNumber = int64(subInfo.UserNum)
		top.MaxReserve = subInfo.MaxReserve
	}
	if mngInfo := campusRcmd.TopShow.ManagementInfo; mngInfo != nil {
		top.MngLabel = &api.CampusLabel{
			Text: mngInfo.Desc,
			Url:  mngInfo.JumpUrl,
		}
	}
	switch campusRcmd.TopShow.ShowType {
	case _campusNotChosen: // 未选择学校
		top.Button = &api.RcmdTopButton{
			Text: "去添加",
			Url:  "bilibili://campus/search",
		}
	case _campusNonSubscribed, _campusUnderControlled: // 选了学校未预约
		if campusRcmd.TopShow.ShowType == _campusNonSubscribed {
			top.NoticeLabel = &api.CampusLabel{
				Text: "一键预约",
			}
			top.Desc_2 = "预约可以加速学校开放哦！"
			top.SchoolLabel = &api.CampusLabel{
				Text: "如何开校",
			}
			if subInfo := campusRcmd.TopShow.SubInfo; subInfo != nil {
				top.SchoolLabel.Url = subInfo.JumpUrl
			}
		}
		fallthrough
	default:
		top.Title = campusRcmd.TopShow.CampusName
		top.Desc = "板块正在筹备中..."
		top.Desc_3 = "校园开放后将私信通知你，先看看其它内容吧"
		top.SwitchLabel = &api.CampusLabel{
			Text: "切换",
			Url:  "bilibili://campus/search",
		}
		// 管控
		if campusRcmd.TopShow.ShowType == _campusUnderControlled {
			top.Type = _campusSubscribed
			top.AuditBeforeOpen = true
			top.AuditMessage = "受相关部门要求，板块需与学校上级主管部门报备后才能开放，敬请期待哦"
		}
	}
	return top
}
