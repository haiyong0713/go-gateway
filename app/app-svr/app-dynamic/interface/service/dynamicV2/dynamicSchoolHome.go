package dynamicV2

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"git.bilibili.co/bapis/bapis-go/bilibili/pagination"
	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

func (s *Service) CampusEntryTab(c context.Context, general *mdlv2.GeneralParam, req *api.CampusEntryTabReq) (*api.CampusEntryTabResp, error) {
	if req.CampusId <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "unexpected empty campus id")
	}
	resp, err := s.dynDao.CampusEntryTab(c, general.Mid, req.CampusId)
	if err != nil {
		return nil, err
	}
	return &api.CampusEntryTabResp{
		EntryType: api.CampusEntryType(resp.GetType()),
	}, nil
}

func (s *Service) FetchTabSetting(c context.Context, general *mdlv2.GeneralParam) (*api.FetchTabSettingReply, error) {
	reply, err := s.dynDao.FetchTabSetting(c, general)
	if err != nil {
		return nil, err
	}
	res := &api.FetchTabSettingReply{
		Status: api.HomePageTabSttingStatus(reply),
	}
	return res, nil
}

func (s *Service) UpdateTabSetting(c context.Context, general *mdlv2.GeneralParam, req *api.UpdateTabSettingReq) (*api.NoReply, error) {
	if err := s.dynDao.UpdateTabSetting(c, general, req); err != nil {
		return nil, err
	}
	return &api.NoReply{}, nil
}

func (s *Service) CampusSquare(c context.Context, general *mdlv2.GeneralParam, req *api.CampusSquareReq) (*api.CampusSquareReply, error) {
	reply, err := s.dynDao.CampusSquare(c, general, req)
	if err != nil {
		return nil, err
	}
	res := &api.CampusSquareReply{
		Title: reply.Title,
		Button: &api.CampusLabel{
			Text: "查看更多",
			Url:  "bilibili://campus/search?action=turn",
		},
	}
	for _, v := range reply.List {
		item := &api.RcmdCampusBrief{
			CampusId:    v.CampusId,
			CampusName:  v.CampusName,
			CampusBadge: v.CampusBadge,
			Url:         fmt.Sprintf("bilibili://campus/detail/%d", v.CampusId),
		}
		res.List = append(res.List, item)
	}
	return res, nil
}

func campusRcmdFrom2SubPageType(from api.CampusRcmdReqFrom) string {
	switch from {
	case api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_HOME_UN_OPEN, api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_UNKNOWN:
		return "homepage"
	case api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_VISIT_OTHER:
		return "other"
	case api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_HOME_MOMENT, api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_PAGE_SUBORDINATE_MOMENT:
		return "homepage_moment"
	case api.CampusRcmdReqFrom_CAMPUS_RCMD_FROM_DYN_MOMENT:
		return "dt_moment"
	default:
		return "homepage"
	}
}

func (s *Service) CampusRecommend(c context.Context, general *mdlv2.GeneralParam, req *api.CampusRecommendReq) (*api.CampusRecommendReply, error) {
	var (
		userSchool *dyncampusgrpc.CampusIdentityReply
		info       *locgrpc.InfoReply
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if userSchool, err = s.dynDao.Identity(ctx, general); err != nil {
			log.Errorc(ctx, "%+v", err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if info, err = s.loc.InfoGRPC(ctx, general.IP); err != nil {
			log.Errorc(ctx, "%+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	rcmd, hasMore, err := s.rcmdDao.Recommend(c, general, req.CampusId, userSchool.GetCampusId(), map[string]string{
		"zone_id":        strconv.FormatInt(info.GetZoneId(), 10),
		"page_no":        strconv.FormatInt(req.PageNo, 10),
		"nearby_subpage": campusRcmdFrom2SubPageType(req.From),
	}, true)
	if err != nil {
		log.Error("%+v", err)
		// 走灾备
		const (
			_itemMax   = 80
			_itemCount = 10
		)
		start := rand.Intn(_itemMax)
		end := start + _itemCount
		redisItems, err := s.dynDao.SchoolCache(c, start, end)
		if err != nil {
			return nil, errors.WithMessagef(ecode.ServiceUnavailable, "%v", err)
		}
		if rcmd == nil {
			rcmd = &mdlv2.RcmdReply{Code: -11, Infoc: &mdlv2.RcmdInfo{Code: http.StatusInternalServerError}}
		}
		for _, v := range redisItems {
			rcmdItem := &mdlv2.RcmdItem{}
			rcmdItem.FromItem(v)
			rcmd.Items = append(rcmd.Items, rcmdItem)
		}
	}
	rcmd.Infoc.ZoneID = info.GetZoneId()
	rcmd.Infoc.SchoolID = userSchool.GetCampusId()
	// 解析基础数据&获取物料
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, homeAiRcmd: rcmd.Items})
	if err != nil {
		return nil, err
	}
	res := &api.CampusRecommendReply{
		HasMore: hasMore,
		Items:   s.homeCampusRcmd(dynCtx, rcmd.Items),
	}
	s.campusRecommendInfoc(c, general, req, rcmd.Infoc, res.Items)
	return res, nil
}

func (s *Service) homeCampusRcmd(dynCtx *mdlv2.DynamicContext, rcmds []*mdlv2.RcmdItem) []*api.RcmdItem {
	if len(rcmds) == 0 {
		return nil
	}
	var rcmdTmps []*api.RcmdItem
	for _, v := range rcmds {
		ap, ok := dynCtx.GetArchive(v.ID)
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
			Desc:            v.RcmdReason.Content,
			TrackId:         v.TrackID,
			RcmdReason: &api.RcmdReason{
				CampusName: v.RcmdReason.Content,
				RcmdReason: v.RcmdReason.ReasonDesc,
			},
		}
		if len(v.RcmdReason.ReasonDesc) > 0 {
			rcmdArcTmp.RcmdReason.Style = api.RcmdReasonStyle_rcmd_reason_style_campus_nearby
		}
		// UGC转PGC逻辑
		if ap.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && ap.Arc.RedirectURL != "" {
			rcmdArcTmp.IsPgc = true
			rcmdArcTmp.Uri = ap.Arc.RedirectURL
		}
		rcmdArcTmp.Uri = model.FillReplyURL(rcmdArcTmp.Uri, fmt.Sprintf("trackid=%s", v.TrackID))
		rcmdTmp := &api.RcmdItem{
			Type: api.RcmdType_rcmd_archive,
			RcmdItem: &api.RcmdItem_RcmdArchive{
				RcmdArchive: rcmdArcTmp,
			},
		}
		rcmdTmps = append(rcmdTmps, rcmdTmp)
	}
	if n := len(rcmdTmps); n >= 1 && n%2 != 0 {
		rcmdTmps = rcmdTmps[0 : n-1]
	}
	return rcmdTmps
}

func (s *Service) HomePages(c context.Context, general *mdlv2.GeneralParam, req *api.CampusHomePagesReq) (*api.CampusHomePagesReply, error) {
	// 请求服务端获取基础数据
	reply, err := s.dynDao.HomePages(c, general, req)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	res := &api.CampusHomePagesReply{PageType: int32(reply.PageType)}
	switch reply.PageType {
	case _feedTab:
		top := s.campusHomeTop(c, general, reply.MajorPage)
		if top == nil {
			return res, nil
		}
		res.CampusTop = top
	case _rcmd:
		top, err := s.campusHomeRcmdPage(c, general, reply.NearbyRcmdPage)
		if err != nil {
			return nil, err
		}
		if top == nil {
			return res, nil
		}
		res.Top = top
	default:
		log.Warn("CampusRcmd miss page_type(%d)", reply.PageType)
		return res, nil
	}
	// 非主态首页干掉切换按钮
	if req.PageType != api.CampusHomePageType_PAGE_MAJOR {
		if res.CampusTop != nil {
			res.CampusTop.SwitchLabel = nil
		} else if res.Top != nil {
			res.Top.SwitchLabel = nil
		}
	}

	return res, nil
}

func (s *Service) campusHomeTop(c context.Context, general *mdlv2.GeneralParam, schoolTop *dyncampusgrpc.HomeMajorPageInfo) *api.CampusTop {
	if schoolTop == nil {
		return nil
	}

	res := &api.CampusTop{
		CampusId:   int64(schoolTop.CampusId),
		CampusName: schoolTop.CampusName,
		SwitchLabel: &api.CampusLabel{
			Text: "切换学校",
			Url:  fmt.Sprintf("bilibili://campus/page/recommend/%d", schoolTop.CampusId),
		},
		Title: schoolTop.CampusName,
		InviteLabel: &api.CampusLabel{
			Text: schoolTop.InviteDesc,
			Url:  s.c.Resource.Others.SchoolInviteURI,
		},
		CampusBadge:      schoolTop.CampusBadge,
		CampusBackground: schoolTop.CampusBackground,
		CampusMotto:      schoolTop.CampusMotto,
		CampusIntro:      schoolTop.CampusBrief,
		CampusNameLink:   schoolTop.MinorPageUrl,
	}
	// 官号管理入口
	if mngInfo := schoolTop.ManagementInfo; mngInfo != nil {
		// 检查客户端版本号
		if s.isCampusMngCapable(c, general) {
			res.MngEntry = &api.CampusLabel{
				Text: mngInfo.Desc,
				Url:  mngInfo.JumpUrl,
			}
		}
	}
	// 通知
	if notice := schoolTop.BindNotice; notice != nil && notice.Title != "" && notice.Desc != "" && notice.ButtonDesc != "" {
		res.Notice = &api.CampusNoticeInfo{
			Title: notice.Title,
			Desc:  notice.Desc,
			Button: &api.CampusLabel{
				Text: notice.ButtonDesc,
				Url:  s.c.Resource.Others.SchoolNoticeURI,
			},
		}
	}
	for _, v := range schoolTop.Tabs {
		var (
			uri, icon  string
			campusType api.CampusTabType
		)
		switch v.TabType {
		case _campusSchool:
			uri = model.FillURI(model.GotoFeedSchool, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_school
		case _campusOfficialAccount:
			uri = fmt.Sprintf("bilibili://campus/page/official/%d", schoolTop.CampusId)
			icon = "http://i0.hdslb.com/bfs/feed-admin/9ed4415dfdb4ec95fc388ee9f0a25deb6f260dd8.png"
			campusType = api.CampusTabType_campus_account
		case _campusOfficialDynamic:
			uri = fmt.Sprintf("bilibili://campus/page/read/%d", schoolTop.CampusId)
			icon = "http://i0.hdslb.com/bfs/feed-admin/cab003566ad0b21aefdb2219c963b13836dbdf77.png"
			campusType = api.CampusTabType_campus_dynamic
		case _campusBillboard:
			// 低版本屏蔽热点tab自动开放功能
			if s.isDisableDynCampusBillboardAutoOpen(c, general, v.TabType, v.TabStatus) {
				continue
			}
			uri = fmt.Sprintf("bilibili://campus/page/billboard/%d", schoolTop.CampusId)
			icon = "http://i0.hdslb.com/bfs/feed-admin/c5d99f99c591ae07d88e61c3a420cba70189bc00.png"
			campusType = api.CampusTabType_campus_billboard
		case _campusTopic:
			uri = fmt.Sprintf("bilibili://campus/page/topic_home/%d", schoolTop.CampusId)
			icon = "http://i0.hdslb.com/bfs/feed-admin/836331034d8b6380bf64b89c7d68b8c51c0fc023.png"
			campusType = api.CampusTabType_campus_topic
		case _campusOther:
			uri = fmt.Sprintf("bilibili://campus/page/recommend/%d", schoolTop.CampusId)
			icon = "http://i0.hdslb.com/bfs/feed-admin/e37fc06f236b3c637aa4a7c393e7687a9cd7cdbc.png"
			campusType = api.CampusTabType_campues_other
		default:
			log.Warnc(c, "campusTop miss mid(%d) type(%d)", general.Mid, v.TabType)
			continue
		}
		// 新版本之后，首页的几个tab跳链去掉page部分，否则端上会出现页面头部
		const (
			_androidLimit = 6850000
			_iosLimit     = 68500000
		)
		if general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _androidLimit, _iosLimit) {
			uri = strings.Replace(uri, "/page/", "/", 1)
		}
		item := &api.CampusShowTabInfo{
			Name:    v.TabName,
			Url:     uri,
			Type:    campusType,
			RedDot:  v.RedDot,
			IconUrl: icon,
		}
		res.Tabs = append(res.Tabs, item)
	}
	// banner
	for _, v := range schoolTop.Banner {
		item := &api.CampusBannerInfo{
			Image:   v.PicUrl,
			JumpUrl: v.JumpUrl,
		}
		res.Banner = append(res.Banner, item)
	}
	if s.isCampusNoBanner(c, general) {
		res.Banner = nil
	}
	// 话题
	if schoolTop.TopicSquare != nil {
		dynCtx, err := s.topicHomeSquareMaterial(c, general, schoolTop.TopicSquare.RcmdCard)
		if err != nil {
			log.Error("%+v", err)
			return res
		}
		res.TopicSquare = s.topicHomeSquareInfo(c, general, schoolTop.TopicSquare, dynCtx)
	}
	return res
}

func (s *Service) topicHomeSquareMaterial(c context.Context, general *mdlv2.GeneralParam, rcmd *dyncampusgrpc.HomeTopicRcmdCard) (*mdlv2.DynamicContext, error) {
	var (
		dyns []*mdlv2.Dynamic
	)
	if rcmd != nil {
		dyns = append(dyns, &mdlv2.Dynamic{Type: int64(rcmd.Type), Rid: int64(rcmd.Rid)})
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dyns})
	return dynCtx, err
}

func (s *Service) topicHomeSquareInfo(_ context.Context, _ *mdlv2.GeneralParam, topicInfo *dyncampusgrpc.HomeTopicSquareInfo, dynCtx *mdlv2.DynamicContext) *api.TopicSquareInfo {
	if topicInfo == nil {
		return nil
	}
	res := &api.TopicSquareInfo{
		Title: topicInfo.Title,
		Button: &api.CampusLabel{
			Text: "更多校园话题",
			Url:  topicInfo.ButtonUrl,
		},
	}
	if rcmd := topicInfo.RcmdCard; rcmd != nil {
		res.Rcmd = &api.TopicRcmdCard{
			TopicId:   int64(rcmd.TopicId),
			TopicName: rcmd.TopicName,
			Url:       rcmd.TopicLink,
			Button: &api.CampusLabel{
				Text: "去讨论",
			},
			UpdateDesc: rcmd.UpdateDesc,
		}
		// 跳转动态发布页带新话题
		res.Rcmd.Button.Url = model.FillURI(model.GotoDynPublishWithNewTopic,
			fmt.Sprintf("topicV2ID=%d&topicV2Name=%s", res.Rcmd.TopicId, model.QueryEscape(res.Rcmd.TopicName)), nil)
		switch rcmd.Type {
		case mdlv2.DynTypeVideo:
			// 视频
			if ap, ok := dynCtx.GetArchive(int64(rcmd.Rid)); ok {
				res.Rcmd.Desc_2 = ap.Arc.Title
				if ap.Arc.Dynamic != "" {
					res.Rcmd.Desc_2 = ap.Arc.Dynamic
				}
			}
		case mdlv2.DynTypeWord:
			// 纯文字
			if dynCtx.ResWords != nil && dynCtx.ResWords[int64(rcmd.Rid)] != "" {
				res.Rcmd.Desc_2 = dynCtx.ResWords[int64(rcmd.Rid)]
			}
		case mdlv2.DynTypeDraw:
			// 图文
			if draw, ok := dynCtx.GetResDraw(int64(rcmd.Rid)); ok {
				res.Rcmd.Desc_2 = draw.Item.Description
			}
		}
		if res.Rcmd.Desc_2 != "" {
			res.Rcmd.Desc_1 = "热门动态："
			// 强制去掉换行符
			res.Rcmd.Desc_2 = strings.Replace(res.Rcmd.Desc_2, "\n", "", -1)
		}
	}
	return res
}

func (s *Service) campusHomeRcmdPage(c context.Context, general *mdlv2.GeneralParam, rcmdInfo *dyncampusgrpc.HomeNearbyRcmdInfo) (top *api.CampusRcmdTop, err error) {
	if rcmdInfo == nil {
		return nil, nil
	}
	// 解析基础数据&获取物料
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, homeRecommendItem: rcmdInfo.RcmdList.List})
	if err != nil {
		return nil, err
	}
	// 聚合&返回
	if general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynSchoolShowTabIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynSchoolShowTabAndroid {
		return s.proHomeCampusRcmdV2(dynCtx, rcmdInfo), nil
	}
	return s.proHomeCampusRcmd(dynCtx, rcmdInfo), nil
}

func (s *Service) proHomeCampusRcmd(_ *mdlv2.DynamicContext, campusRcmd *dyncampusgrpc.HomeNearbyRcmdInfo) (top *api.CampusRcmdTop) {
	if campusRcmd == nil {
		return nil
	}
	if campusRcmd.TopShow == nil {
		log.Warn("NearbyCampusRcmd get TopShowInfo nil")
	}
	if campusRcmd.TopShow != nil {
		top = &api.CampusRcmdTop{
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
		if campusRcmd.HotForumCard != nil {
			top.RcmdTopic = (&mdlv2.CampusHotTopicInfo{HomeHead: campusRcmd.HotForumCard}).ToV2CampusHomeRcmdTopic()
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
	}
	return top
}

func (s *Service) proHomeCampusRcmdV2(_ *mdlv2.DynamicContext, campusRcmd *dyncampusgrpc.HomeNearbyRcmdInfo) (top *api.CampusRcmdTop) {
	if campusRcmd == nil {
		return nil
	}
	if campusRcmd.TopShow == nil {
		log.Warn("NearbyCampusRcmd get TopShowInfo nil")
	}
	if campusRcmd.TopShow != nil {
		top = &api.CampusRcmdTop{
			CampusId:   int64(campusRcmd.TopShow.CampusId),
			CampusName: campusRcmd.TopShow.CampusName,
			Title:      "你还未添加校园",
			Desc:       "立即添加，抢先了解校园新动态哦",
			Type:       int32(campusRcmd.TopShow.ShowType),
			InviteLabel: &api.CampusLabel{
				Text: campusRcmd.TopShow.InviteDesc,
				Url:  s.c.Resource.Others.SchoolInviteURI,
			},
			SwitchLabel: &api.CampusLabel{
				Text: "切换学校",
				Url:  fmt.Sprintf("bilibili://campus/page/recommend/%d", campusRcmd.TopShow.CampusId),
			},
		}
		if campusRcmd.HotForumCard != nil {
			top.RcmdTopic = (&mdlv2.CampusHotTopicInfo{HomeHead: campusRcmd.HotForumCard}).ToV2CampusHomeRcmdTopic()
		}
		// 预约信息
		if subInfo := campusRcmd.TopShow.SubInfo; subInfo != nil {
			top.ReserveNumber = subInfo.UserNum
			top.MaxReserve = subInfo.MaxReserve
		}
		// 管理按钮
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
			// 管控
			if campusRcmd.TopShow.ShowType == _campusUnderControlled {
				top.Type = _campusSubscribed
				top.AuditBeforeOpen = true
				top.AuditMessage = "受相关部门要求，板块需与学校上级主管部门报备后才能开放，敬请期待哦"
			}
		}
	}
	return top
}

func (s *Service) HomeSubscribe(c context.Context, general *mdlv2.GeneralParam, req *api.HomeSubscribeReq) (*api.HomeSubscribeReply, error) {
	reply, err := s.dynDao.HomeSubscribe(c, general, req)
	if err != nil {
		return nil, err
	}
	return &api.HomeSubscribeReply{Online: api.CampusOnlineStatus(reply.GetOnline())}, nil
}

func (s *Service) WaterFlowRcmd(ctx context.Context, general *mdlv2.GeneralParam, req *api.WaterFlowRcmdReq) (resp *api.WaterFlowRcmdResp, err error) {
	var (
		userSchool *dyncampusgrpc.CampusIdentityReply
		info       *locgrpc.InfoReply
	)
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if userSchool, err = s.dynDao.Identity(ctx, general); err != nil {
			log.Errorc(ctx, "%+v", err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if info, err = s.loc.InfoGRPC(ctx, general.IP); err != nil {
			log.Errorc(ctx, "%+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	resp = new(api.WaterFlowRcmdResp)

	param := map[string]string{
		"zone_id":        strconv.FormatInt(info.GetZoneId(), 10),
		"nearby_subpage": campusRcmdFrom2SubPageType(req.From),
		"pic_mode":       "1",
		"fresh_type":     "1",
	}
	if req.Page.GetIsRefresh() {
		param["fresh_type"] = "2"
	}
	if len(req.Page.GetOffset()) > 0 {
		pageNo, err := strconv.ParseInt(req.Page.GetOffset(), 10, 64)
		if err != nil {
			return nil, errors.WithMessagef(ecode.RequestErr, "invalid AIrcmd page offset")
		}
		param["page_no"] = req.Page.GetOffset()
		param["fresh_type"] = "3"
		pageNo++ // 加一页
		resp.Offset = &pagination.FeedPaginationReply{
			NextOffset: strconv.FormatInt(pageNo, 10),
		}
	} else {
		param["page_no"] = "0"
		resp.Offset = &pagination.FeedPaginationReply{
			NextOffset: "1", // 下次从第一页开始
		}
	}
	rcmdRes, hasMore, err := s.rcmdDao.Recommend(ctx, general, req.CampusId, userSchool.GetCampusId(), param, false)
	if err != nil {
		log.Errorc(ctx, "school rcmdDao.Recommend failed: %v", err)
		// 走灾备
		const (
			_itemMax   = 80
			_itemCount = 10
		)
		start := rand.Intn(_itemMax)
		end := start + _itemCount
		redisItems, err := s.dynDao.SchoolCache(ctx, start, end)
		if err != nil {
			log.Errorc(ctx, "schoolCache error: %v", err)
			return nil, errors.WithMessagef(ecode.ServiceUnavailable, "%v", err)
		}
		if rcmdRes == nil {
			rcmdRes = &mdlv2.RcmdReply{Code: -11, Infoc: &mdlv2.RcmdInfo{Code: http.StatusInternalServerError}}
		}
		for _, v := range redisItems {
			rcmdItem := &mdlv2.RcmdItem{}
			rcmdItem.FromItem(v)
			rcmdRes.Items = append(rcmdRes.Items, rcmdItem)
		}
	}
	if !hasMore {
		resp.Offset.NextOffset = ""
	}
	// 解析基础数据&获取物料
	dynCtx, err := s.getMaterial(ctx, getMaterialOption{general: general, homeAiRcmd: rcmdRes.Items})
	if err != nil {
		return nil, err
	}
	resp.Items = rcmdRes.ToV2CampusWaterFlowItems(dynCtx, s.c.Ctrl.CampusWaterFlowForceVideoHorizontal)

	// infoc 上报
	// 用于上报推荐的内容下发和实际展现的diff情况
	rcmdRes.Infoc.SchoolID, rcmdRes.Infoc.ZoneID = userSchool.GetCampusId(), info.GetZoneId()
	s.campusWaterFlowInfoc(ctx, general, req, param, rcmdRes.Infoc, resp.Items)

	return resp, nil
}
