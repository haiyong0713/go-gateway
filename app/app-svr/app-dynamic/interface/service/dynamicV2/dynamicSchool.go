package dynamicV2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	relationmdl "go-gateway/app/app-svr/app-dynamic/interface/model/relation"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	dyncomn "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

func (s *Service) DynAlumniDynamics(c context.Context, general *mdlv2.GeneralParam, req *api.AlumniDynamicsReq) (*api.AlumniDynamicsReply, error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	var (
		attentions *dyncomn.AttentionInfo
		info       *locgrpc.InfoReply
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(ctx, general.Mid, true, true, general)
		if err != nil {
			return err
		}
		attentions = mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if info, err = s.loc.InfoGRPC(ctx, general.IP); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	dynList, err := s.dynDao.AlumniDynamics(c, general, &api.CampusRcmdFeedReq{
		CampusId: req.CampusId, FirstTime: req.FirstTime, Page: req.Page, FromType: req.GetFromType(),
	}, attentions, info.GetZoneId())
	if err != nil {
		return nil, err
	}
	reply := &api.AlumniDynamicsReply{
		Toast: dynList.Toast,
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeSchool)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	reply.List = retDynList
	return reply, nil
}

const (
	// 校友圈
	_campusSchool = 1
	// 入校必看
	_campusOfficialDynamic = 2
	// 官方账号
	_campusOfficialAccount = 3
	// 校园十大榜单
	_campusBillboard = 4
	// 校园话题讨论
	_campusTopic = 5
	// 其他校园
	_campusOther = 6
)

func (s *Service) campusTop(c context.Context, general *mdlv2.GeneralParam, schoolTop *dyncampusgrpc.MajorPageInfo) *api.CampusTop {
	if schoolTop == nil {
		return nil
	}

	res := &api.CampusTop{
		CampusId:   int64(schoolTop.CampusId),
		CampusName: schoolTop.CampusName,
		SwitchLabel: &api.CampusLabel{
			Text: "切换",
			Url:  "bilibili://campus/search",
		},
		Title: schoolTop.CampusName,
		InviteLabel: &api.CampusLabel{
			Text: schoolTop.InviteDesc,
			Url:  s.c.Resource.Others.SchoolInviteURI,
		},
		CampusBadge:      schoolTop.CampusBadge,
		CampusBackground: schoolTop.CampusBackground,
		CampusMotto:      schoolTop.CampusMotto,
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

	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynCampusBanner, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCampusBannerIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynCampusBannerAndroid)}) {
		res.SwitchLabel.Text = "切换学校"
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
		// 版本控制，一期不下发入校必看、官方号
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynSchool, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynSchoolIOS) ||
				(general.IsAndroidPick() && general.GetBuild() <= s.c.BuildLimit.DynSchoolAndroid)}) {
			// 老版本不展示校友圈以外的tab
			if v.TabType != _campusSchool {
				continue
			}
		}
		// 六期以前不下发 校园榜单
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynSchoolBillboard, &feature.OriginResutl{
			BuildLimit: general.IsMobileBuildLimitMet(mdlv2.Less, s.c.BuildLimit.DynSchoolBillboardAndroid, s.c.BuildLimit.DynSchoolBillboardIOS),
		}) {
			if v.TabType == _campusBillboard {
				continue
			}
		}
		// 下发校园榜单但是未开放的情况下 低版本直接屏蔽下发
		// 新版本客户端才兼容开放状态展示
		if s.isDisableDynCampusBillboardAutoOpen(c, general, int64(v.TabType), v.TabStatus) {
			continue
		}
		// 六期以前不下发校园话题
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynSchoolTopicDiscuss, &feature.OriginResutl{
			BuildLimit: general.IsMobileBuildLimitMet(mdlv2.Less, s.c.BuildLimit.DynSchoolTopicDiscussAndroid, s.c.BuildLimit.DynSchoolTopicDiscussIOS),
		}) {
			if v.TabType == _campusTopic {
				continue
			}
		}

		var (
			uri        string
			campusType api.CampusTabType
		)
		switch v.TabType {
		case _campusSchool:
			uri = model.FillURI(model.GotoFeedSchool, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_school
		case _campusOfficialAccount:
			uri = model.FillURI(model.GotoOfficialAccount, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_account
		case _campusOfficialDynamic:
			uri = model.FillURI(model.GotoOfficialDynamic, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_dynamic
		case _campusBillboard:
			uri = model.FillURI(model.GotoSchoolBillboard, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_billboard
		case _campusTopic:
			uri = model.FillURI(model.GotoSchoolTopicHome, strconv.FormatInt(int64(schoolTop.CampusId), 10), nil)
			campusType = api.CampusTabType_campus_topic
		default:
			log.Warn("campusTop miss mid(%d) type(%d)", general.Mid, v.TabType)
			continue
		}
		item := &api.CampusShowTabInfo{
			Name:   v.TabName,
			Url:    uri,
			Type:   campusType,
			RedDot: v.RedDot,
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
		dynCtx, err := s.topicSquareMaterial(c, general, schoolTop.TopicSquare.RcmdCard)
		if err != nil {
			log.Error("%+v", err)
			return res
		}
		res.TopicSquare = s.topicSquareInfo(c, general, schoolTop.TopicSquare, dynCtx)
	}

	return res
}

func (s *Service) SchoolRecommend(c context.Context, general *mdlv2.GeneralParam, param *api.SchoolRecommendReq) (*api.SchoolRecommendReply, error) {
	reply, err := s.dynDao.SchoolRecommend(c, general, param)
	if err != nil {
		return nil, err
	}
	res := &api.SchoolRecommendReply{}
	for _, v := range reply {
		item := &api.CampusInfo{
			CampusId:   int64(v.CampusId),
			CampusName: v.CampusName,
			Online:     int64(v.Online),
			Url:        fmt.Sprintf("bilibili://campus/detail/%d", v.CampusId),
		}
		res.Items = append(res.Items, item)
	}
	return res, nil
}

func (s *Service) SchoolSearch(c context.Context, _ *mdlv2.GeneralParam, param *api.SchoolSearchReq) (*api.SchoolSearchReply, error) {
	reply, err := s.dynDao.SchoolSearch(c, param)
	if err != nil {
		return nil, err
	}
	res := &api.SchoolSearchReply{
		Toast: &api.SearchToast{
			DescText_1: s.c.Resource.Text.SearchToast,
			DescText_2: s.c.Resource.Text.SearchToastDesc,
		},
	}
	for _, v := range reply {
		item := &api.CampusInfo{
			CampusId:   int64(v.CampusId),
			CampusName: v.CampusName,
			// Desc:       fmt.Sprintf("%s/%s", v.CityName, v.Level), 一期不显示
			Online: int64(v.Online),
			Url:    fmt.Sprintf("bilibili://campus/detail/%d", v.CampusId),
		}
		res.Items = append(res.Items, item)
	}
	return res, nil
}

func (s *Service) OfficialAccounts(c context.Context, general *mdlv2.GeneralParam, param *api.OfficialAccountsReq) (*api.OfficialAccountsReply, error) {
	reply, err := s.dynDao.OfficialAccounts(c, param, general)
	if err != nil {
		return nil, err
	}
	var (
		mids      []int64
		passedm   map[int64]int64
		accm      map[int64]*accountgrpc.Card
		relationm map[int64]*relationgrpc.InterrelationReply
		statm     map[int64]*relationgrpc.StatReply
	)
	for _, uid := range reply.GetUids() {
		mids = append(mids, int64(uid))
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		res, err := s.accountDao.Cards3New(ctx, mids)
		if err != nil {
			log.Warn("getMaterial mid(%v) Cards3New(%v) error(%v)", general.Mid, mids, err)
			return nil
		}
		accm = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		res, err := s.accountDao.Interrelations(ctx, general.Mid, mids)
		if err != nil {
			log.Error("getMaterial mid(%v) Interrelations(%v), error %v", general.Mid, mids, err)
			return nil
		}
		relationm = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		res, err := s.upDao.ArcsPassedTotal(ctx, mids)
		if err != nil {
			log.Error("getMaterial mid(%v) ArcsPassedTotal(%v), error %v", general.Mid, mids, err)
			return nil
		}
		passedm = res
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		res, err := s.relationDao.Stats(ctx, mids)
		if err != nil {
			log.Warn("getMaterial mid(%v) Stats(%v), error(%v)", general.Mid, mids, err)
			return nil
		}
		statm = res
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &api.OfficialAccountsReply{
		Offset: int64(reply.Offset),
	}
	if reply.HasMore == 1 {
		res.HasMore = true
	}
	for _, uid := range reply.GetUids() {
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, int64(uid), general) {
			continue
		}
		userInfo, ok := accm[int64(uid)]
		if !ok || userInfo.IsDeleted != 0 {
			continue
		}
		accInfo := &api.OfficialAccountInfo{
			Author: &api.UserInfo{
				Mid:  userInfo.Mid,
				Name: userInfo.Name,
				Face: userInfo.Face,
				Official: &api.OfficialVerify{ // 认证
					Type: int32(userInfo.Official.Type),
					Desc: userInfo.Official.Desc,
				},
				Vip: &api.VipInfo{ // 会员
					Type:    userInfo.Vip.Type,
					Status:  userInfo.Vip.Status,
					DueDate: userInfo.Vip.DueDate,
					Label: &api.VipLabel{
						Path: userInfo.Vip.Label.Path,
					},
					ThemeType: userInfo.Vip.ThemeType,
				},
				Pendant: &api.UserPendant{ // 头像挂件
					Pid:    int64(userInfo.Pendant.Pid),
					Name:   userInfo.Pendant.Name,
					Image:  userInfo.Pendant.Image,
					Expire: int64(userInfo.Pendant.Expire),
				},
				Nameplate: &api.Nameplate{ // 勋章
					Nid:        int64(userInfo.Nameplate.Nid),
					Name:       userInfo.Nameplate.Name,
					Image:      userInfo.Nameplate.Image,
					ImageSmall: userInfo.Nameplate.ImageSmall,
					Level:      userInfo.Nameplate.Level,
					Condition:  userInfo.Nameplate.Condition,
				},
				Uri:        model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
				Level:      userInfo.Level,
				Sign:       userInfo.Sign,
				FaceNft:    userInfo.FaceNft,
				FaceNftNew: userInfo.FaceNftNew,
			},
			Mid:        userInfo.Mid,
			Uri:        model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
			Relation:   relationmdl.RelationChange(int64(uid), relationm),
			DescText_2: userInfo.Official.Desc,
		}
		if relation, ok := statm[int64(uid)]; ok {
			accInfo.DescText_1 = fmt.Sprintf("粉丝：%s", model.StatString(relation.Follower, ""))
		}
		if passed, ok := passedm[int64(uid)]; ok {
			desc := fmt.Sprintf("%s个视频", model.StatString(passed, ""))
			accInfo.DescText_1 = accInfo.DescText_1 + "  " + desc
		}
		res.Items = append(res.Items, accInfo)
	}
	return res, nil
}

func (s *Service) OfficialDynamics(c context.Context, general *mdlv2.GeneralParam, param *api.OfficialDynamicsReq) (*api.OfficialDynamicsReply, error) {
	dynList, err := s.dynDao.OfficialDynamics(c, param, general)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics})
	if err != nil {
		return nil, err
	}
	var rcmdTmps []*api.OfficialItem
	for _, v := range dynList.Dynamics {
		if v.Type != mdlv2.DynTypeVideo {
			continue
		}
		ap, ok := dynCtx.GetArchive(int64(v.Rid))
		if !ok || !ap.Arc.IsNormal() {
			continue
		}
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, ap.Arc.Author.GetMid(), general) {
			continue
		}
		rcmdArcTmp := &api.OfficialRcmdArchive{
			Title:          ap.Arc.GetTitle(),
			Cover:          ap.Arc.GetPic(),
			CoverRightText: s.videoDuration(ap.Arc.Duration),
			DescIcon_1:     api.CoverIcon_cover_icon_up,
			DescText_1:     ap.Arc.Author.GetName(),
			DescIcon_2:     api.CoverIcon_cover_icon_play,
			DescText_2:     fmt.Sprintf("%s观看", s.numTransfer(int(ap.Arc.Stat.View))),
			Reason:         v.Desc,
			ShowThreePoint: true,
			Uri:            model.FillURI(model.GotoAv, strconv.FormatInt(ap.Arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, ap.Arc.GetFirstCid(), true)),
			Aid:            ap.Arc.GetAid(),
			Mid:            ap.Arc.Author.GetMid(),
			Name:           ap.Arc.Author.GetName(),
		}
		rcmdTmp := &api.OfficialItem{
			Type: api.RcmdType_rcmd_archive,
			RcmdItem: &api.OfficialItem_RcmdArchive{
				RcmdArchive: rcmdArcTmp,
			},
		}
		rcmdTmps = append(rcmdTmps, rcmdTmp)
	}
	res := &api.OfficialDynamicsReply{
		Offset:  dynList.OffsetInt,
		HasMore: dynList.HasMore,
		Items:   rcmdTmps,
	}
	return res, nil
}

func (s *Service) CampusRedDot(c context.Context, general *mdlv2.GeneralParam, param *api.CampusRedDotReq) (*api.CampusRedDotReply, error) {
	reply, err := s.dynDao.CampusRedDot(c, param, general)
	if err != nil {
		log.Warn("%+v", err)
		return nil, err
	}
	res := &api.CampusRedDotReply{RedDot: int32(reply.GetRedDot())}
	return res, nil
}

func (s *Service) CampusRcmdFeed(c context.Context, general *mdlv2.GeneralParam, param *api.CampusRcmdFeedReq) (*api.CampusRcmdFeedReply, error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	var (
		attentions *dyncomn.AttentionInfo
		userSchool *dyncampusgrpc.CampusIdentityReply
		info       *locgrpc.InfoReply
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (err error) {
		if userSchool, err = s.dynDao.Identity(ctx, general); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(ctx, general.Mid, true, true, general)
		if err != nil {
			return err
		}
		attentions = mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if info, err = s.loc.InfoGRPC(ctx, general.Network.RemoteIP); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	dynList, err := s.dynDao.AlumniDynamics(c, general, param, attentions, info.GetZoneId())
	if err != nil {
		return nil, err
	}
	reply := &api.CampusRcmdFeedReply{
		Toast:   dynList.Toast,
		HasMore: dynList.HasMore,
		Update:  dynList.CampusFeedUpdate,
	}
	if guideBar := dynList.GuideBar; guideBar != nil {
		reply.GuideBar = &api.GuideBarInfo{
			Show:         int32(guideBar.Show),
			Page:         int32(guideBar.Page),
			Position:     int32(guideBar.Position),
			Desc:         guideBar.Desc,
			JumpPage:     int32(guideBar.JumpPage),
			JumpPosition: int32(guideBar.JumpPosition),
		}
	}
	if len(dynList.Dynamics) == 0 {
		return reply, nil
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dynList.Dynamics})
	if err != nil {
		return nil, err
	}
	// 填入当前上下文的校园id  回填物料时会用到
	dynCtx.CampusID = param.CampusId
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, dynList.Dynamics, dynCtx, general, _handleTypeSchool)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	// Step 6. 折叠判断
	retDynList := s.procFold(foldList, dynCtx, general)
	reply.List = retDynList
	// infoc
	if dynList.RcmdInfo != nil {
		dynList.RcmdInfo.ZoneID = info.GetZoneId()
		dynList.RcmdInfo.SchoolID = userSchool.GetCampusId()
	}
	s.campusRcmdFeedInfoc(c, general, param, dynList.RcmdInfo, foldList.List)

	// 生成热议话题卡
	var hotTopicCards map[int]*api.DynamicItem
	if dynList.CampusHotTopic != nil && s.isCampusHotTopicCapable(c, general) {
		hotTopicCards = dynList.CampusHotTopic.ToV2DynamicItem()
	}
	// 生成校园小黄条
	var yellowBars map[int]*api.DynamicItem
	if dynList.YellowBars != nil && s.isCampusYellowBarCapable(c, general) {
		yellowBars = dynList.GetYellowBarV2DynamicItems()
	}
	// 热议卡和小黄条插入校园feed流
	reply.List = dynList.InsertIntoDynList(hotTopicCards, yellowBars, reply.List)

	return reply, nil
}

func (s *Service) TopicSquare(c context.Context, general *mdlv2.GeneralParam, param *api.TopicSquareReq) (*api.TopicSquareReply, error) {
	reply, err := s.dynDao.TopicSquare(c, param, general)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if reply.Info == nil {
		return &api.TopicSquareReply{}, nil
	}
	dynCtx, err := s.topicSquareMaterial(c, general, reply.Info.RcmdCard)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &api.TopicSquareReply{
		Info: s.topicSquareInfo(c, general, reply.Info, dynCtx),
	}
	return res, nil
}

func (s *Service) topicSquareMaterial(c context.Context, general *mdlv2.GeneralParam, rcmd *dyncampusgrpc.TopicRcmdCard) (*mdlv2.DynamicContext, error) {
	var (
		dyns []*mdlv2.Dynamic
	)
	if rcmd != nil {
		dyns = append(dyns, &mdlv2.Dynamic{Type: int64(rcmd.Type), Rid: int64(rcmd.Rid)})
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: dyns})
	return dynCtx, err
}

func (s *Service) topicSquareInfo(c context.Context, general *mdlv2.GeneralParam, topicInfo *dyncampusgrpc.TopicSquareInfo, dynCtx *mdlv2.DynamicContext) *api.TopicSquareInfo {
	if topicInfo == nil {
		return nil
	}
	res := &api.TopicSquareInfo{
		Title: topicInfo.Title,
		Button: &api.CampusLabel{
			Text: topicInfo.ButtonDesc,
			Url:  topicInfo.ButtonUrl,
		},
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynCampusBanner, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynCampusBannerIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynCampusBannerAndroid)}) {
		res.Button.Text = "更多校园话题"
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

func (s *Service) TopicList(c context.Context, general *mdlv2.GeneralParam, param *api.TopicListReq) (*api.TopicListReply, error) {
	reply, err := s.dynDao.TopicList(c, param, general)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &api.TopicListReply{
		Offset:  reply.Offset,
		HasMore: reply.HasMore != 0,
	}
	if len(reply.CreateTopicLink) > 0 {
		res.CreateTopicBtn = &api.IconButton{
			JumpUri: reply.CreateTopicLink,
		}
	}
	var (
		mids []int64
		accm map[int64]*accountgrpc.Card
	)
	for _, v := range reply.List {
		if v.Uid > 0 {
			mids = append(mids, int64(v.Uid))
		}
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		res, err := s.accountDao.Cards3New(ctx, mids)
		if err != nil {
			log.Warn("getMaterial mid(%v) Cards3New(%v) error(%v)", general.Mid, mids, err)
			return nil
		}
		accm = res
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, v := range reply.List {
		item := &api.TopicItem{
			TopicId:   int64(v.TopicId),
			TopicName: v.TopicName,
			Url:       v.TopicLink,
			RcmdDesc:  v.RcmdDesc,
		}
		userInfo, ok := accm[int64(v.Uid)]
		if ok && userInfo.IsDeleted == 0 {
			item.Desc = userInfo.Name
		}
		if v.HeatInfo != nil {
			item.Desc_2 = fmt.Sprintf("%s %s", fmt.Sprintf("%s浏览", s.numTransfer(int(v.HeatInfo.View))), fmt.Sprintf("%s讨论", s.numTransfer(int(v.HeatInfo.Discuss))))
		}
		res.Items = append(res.Items, item)
	}
	return res, nil
}

//nolint:gomnd
func (s *Service) CampusMateLikeList(ctx context.Context, g *mdlv2.GeneralParam, req *api.CampusMateLikeListReq) (resp *api.CampusMateLikeListReply, err error) {
	res, err := s.dynDao.CampusLikeList(ctx, req, g)
	if err != nil {
		log.Errorc(ctx, "dynDao.CampusLikeList error (%v)", err)
		return
	}
	resp = new(api.CampusMateLikeListReply)
	if len(res.GetList()) == 0 {
		return
	}
	list := res.GetList()
	if len(list) > 100 {
		list = list[0:100]
	}
	mids := make([]int64, 0, len(list))
	authors := make([]*api.ModuleAuthor, 0, len(list))
	for _, r := range list {
		mids = append(mids, r.Uid)
		authors = append(authors, &api.ModuleAuthor{Mid: r.Uid, Attend: int32(r.Attend)})
	}
	cards, err := s.accountDao.Cards3New(ctx, mids)
	if err != nil {
		xmetric.DyanmicItemAPI.Inc("/account.service.Account/Cards3", "request_error")
		log.Errorc(ctx, "CampusMateLikeList getMaterial mid(%v) Cards3(%v) error(%v)", g.Mid, mids, err)
		return nil, err
	}
	for _, u := range authors {
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(ctx, u.Mid, g) {
			continue
		}
		userInfo, ok := cards[u.Mid]
		if !ok {
			continue
		}
		u.Author = &api.UserInfo{
			Mid:  userInfo.Mid,
			Name: userInfo.Name,
			Face: userInfo.Face,
			Official: &api.OfficialVerify{ // 认证
				Type: userInfo.Official.Type,
				Desc: userInfo.Official.Desc,
			},
			Vip: &api.VipInfo{ // 会员
				Type:    userInfo.Vip.Type,
				Status:  userInfo.Vip.Status,
				DueDate: userInfo.Vip.DueDate,
				Label: &api.VipLabel{
					Path: userInfo.Vip.Label.Path,
				},
				ThemeType:       userInfo.Vip.ThemeType,
				AvatarSubscript: userInfo.Vip.AvatarSubscript,
				NicknameColor:   userInfo.Vip.NicknameColor,
			},
			Pendant: &api.UserPendant{ // 头像挂件
				Pid:    int64(userInfo.Pendant.Pid),
				Name:   userInfo.Pendant.Name,
				Image:  userInfo.Pendant.Image,
				Expire: userInfo.Pendant.Expire,
			},
			Nameplate: &api.Nameplate{ // 勋章
				Nid:        int64(userInfo.Nameplate.Nid),
				Name:       userInfo.Nameplate.Name,
				Image:      userInfo.Nameplate.Image,
				ImageSmall: userInfo.Nameplate.ImageSmall,
				Level:      userInfo.Nameplate.Level,
				Condition:  userInfo.Nameplate.Condition,
			},
			Uri:        model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(userInfo.Mid, 10), nil),
			Level:      userInfo.Level,
			Sign:       userInfo.Sign,
			FaceNft:    userInfo.FaceNft,
			FaceNftNew: userInfo.FaceNftNew,
		}
		if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
			u.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
		}
	}

	resp.List = authors
	return
}

var _validFeedbackBizType = map[int32]struct{}{
	1: {}, 2: {}, 3: {}, 4: {},
}

func (s *Service) CampusFeedback(ctx context.Context, general *mdlv2.GeneralParam, req *api.CampusFeedbackReq) (resp *api.CampusFeedbackReply, err error) {
	// 校验
	if len(req.GetInfos()) == 0 {
		return nil, ecode.RequestErr
	}
	for _, info := range req.GetInfos() {
		if _, ok := _validFeedbackBizType[info.BizType]; !ok {
			log.Warnc(ctx, "CampusFeedback unknown bizType(%d): %+v", info.BizType, *info)
			return nil, ecode.RequestErr
		}
		if info.BizId == 0 {
			return nil, ecode.RequestErr
		}
	}
	err = s.dynDao.CampusFeedback(ctx, general.Mid, req)
	if err != nil {
		log.Errorc(ctx, "CampusFeedback dao error: (%v)", err)
		return nil, err
	}
	resp = &api.CampusFeedbackReply{
		Message: "已成功提交",
	}
	return
}

func (s *Service) CampusBillboard(ctx context.Context, general *mdlv2.GeneralParam, req *api.CampusBillBoardReq) (resp *api.CampusBillBoardReply, err error) {
	const _maxBillboardItems = 10 // 排行榜最大稿件数量

	if req.CampusId <= 0 && len(req.VersionCode) == 0 {
		return nil, errors.WithMessage(ecode.RequestErr, "either CampusId or VersionCode must be provided")
	}
	bi, err := s.dynDao.CampusBillboardMeta(ctx, general.Mid, req.CampusId, req.VersionCode, req.GetFromType())
	if err != nil {
		log.Errorc(ctx, "dynDao.CampusBillboardMeta error: (%v)", err)
		return nil, err
	}
	// 优先处理未开放状态
	if opInfo := bi.Meta.GetSubInfo(); opInfo != nil {
		resp = &api.CampusBillBoardReply{
			OpenProgress: &api.CampusFeatureProgress{
				ProgressFull:     opInfo.GetMaxReserve(),
				ProgressAchieved: opInfo.GetUserNum(),
				DescTitle:        "暂未开启",
				Desc_1:           fmt.Sprintf("近7天投稿校友%d人以上时该板块自动开启", opInfo.GetMaxReserve()),
			},
		}
		if opInfo.GetVisitStatus() == dyncomn.UserVisitStatus_VISIT_STATUS_MASTER {
			resp.OpenProgress.Btn = &api.CampusLabel{
				Text: "去投稿",
				Url: "bilibili://uper/center_plus?" + url.Values{
					"relation_from": []string{"campus"},
					"tab_index":     []string{"2"},
					"post_config":   []string{"{\"first_entrance\":\"校园\"}"},
				}.Encode(),
			}
		}
		return
	}

	// 正常渲染内容
	resp = &api.CampusBillBoardReply{
		Title:       bi.Meta.GetTitleName(),
		HelpUri:     bi.Meta.GetJumpUrl(),
		CampusName:  bi.Meta.GetCampusName(),
		BuildTime:   bi.Meta.GetBuildTime(),
		VersionCode: bi.Meta.GetVersionCode(),
		BindNotice:  bi.Meta.GetBindNotice(),
		CampusId:    bi.Meta.GetCampusId(),
		UpdateToast: bi.Meta.GetToast(),
		ShareUri: model.SuffixHandler("version_code=" + model.QueryEscape(bi.Meta.VersionCode))(
			s.c.Resource.Others.SchoolBillboardShareURI),
	}
	if len(bi.Dyns) <= 0 {
		return
	}
	if req.FromType == api.CampusReqFromType_HOME {
		resp.Title = fmt.Sprintf("榜单生成时间：%v", time.Unix(bi.Meta.GetBuildTime(), 0).Format("01-02 15:04"))
	}
	dynCtx, err := s.getMaterial(ctx, getMaterialOption{general: general, dynamics: bi.Dyns})
	if err != nil {
		log.Errorc(ctx, "CampusBillboard getMaterial error: (%v)", err)
		// 只log错误 榜单内容返回空
		return resp, nil
	}
	resp.List = make([]*api.OfficialItem, 0, len(bi.Items))
	// 按照顺序填充排行榜信息
	for _, item := range bi.Items {
		if len(resp.List) >= _maxBillboardItems {
			break
		}
		dyn := item.DynBrief
		if !dyn.Visible {
			continue
		}
		dynCtx.Dyn = dyn
		var filledItem *api.OfficialItem
		switch dyn.Type {
		case mdlv2.DynTypeVideo:
			avInfo, ok := dynCtx.GetArchive(dyn.Rid)
			if !ok || avInfo == nil || avInfo.Arc == nil || !avInfo.Arc.IsNormal() {
				continue
			}
			arc := avInfo.Arc
			// 付费合集
			if mdlv2.PayAttrVal(arc) {
				continue
			}
			filledItem = &api.OfficialItem{
				Type: api.RcmdType_rcmd_archive,
				RcmdItem: &api.OfficialItem_RcmdArchive{
					RcmdArchive: &api.OfficialRcmdArchive{
						Title: arc.Title, Cover: arc.Pic, DynamicId: dyn.DynamicID,
						Aid: arc.Aid, Cid: arc.GetFirstCid(), Mid: arc.Author.GetMid(), Name: arc.Author.GetName(),
						Reason: item.Reason, ShowThreePoint: true,
						CoverRightText: s.videoDuration(arc.Duration),
						Uri: model.FillURI(model.GotoAv, strconv.FormatInt(arc.Aid, 10),
							model.AvPlayHandlerGRPCV2(avInfo, arc.GetFirstCid(), true)),
						DescIcon_1: api.CoverIcon_cover_icon_none,
						DescText_1: arc.PubDate.Time().Format("1-2"),
						DescIcon_2: api.CoverIcon_cover_icon_up,
						DescText_2: arc.Author.GetName(),
					},
				},
			}
		case mdlv2.DynTypeDraw:
			draw, ok := dynCtx.GetResDraw(dyn.Rid)
			if !ok || draw == nil || draw.Item == nil || draw.User == nil {
				log.Warnc(ctx, "CampusBillboard GetResDraw not found: Rid(%d) Res(%+v) Dyn(%+v)", dyn.Rid, draw, dyn)
				continue
			}
			cover := ""
			if len(draw.Item.Pictures) > 0 {
				cover = draw.Item.Pictures[0].ImgSrc
			}
			title := s.descriptionDraw(dynCtx, general)
			if len(title) == 0 {
				title = "图文动态"
			}
			filledItem = &api.OfficialItem{
				Type: api.RcmdType_rcmd_dynamic,
				RcmdItem: &api.OfficialItem_RcmdDynamic{
					RcmdDynamic: &api.OfficialRcmdDynamic{
						Title: title, Cover: cover, Rid: dyn.Rid,
						DynamicId: dyn.DynamicID, Mid: draw.User.UID, UserName: draw.User.Name,
						CoverRightTopText: "图文", Reason: item.Reason,
						Uri: model.FillURI(model.GotoDyn, strconv.FormatInt(dyn.DynamicID, 10),
							model.SuffixHandler(fmt.Sprintf("cardType=%d&rid=%d", dyn.Type, dyn.Rid))),
						DescIcon_1: api.CoverIcon_cover_icon_none,
						DescText_1: time.Unix(dyn.Timestamp, 0).Format("1-2"),
						DescIcon_2: api.CoverIcon_cover_icon_up,
						DescText_2: draw.User.Name,
					},
				},
			}
		default:
			log.Errorc(ctx, "CampusBillboard unhandled billboardItem: (%+v)", dyn)
		}
		if filledItem != nil {
			resp.List = append(resp.List, filledItem)
		}
	}
	return resp, nil
}

func (s *Service) CampusTopicRcmdFeed(ctx context.Context, general *mdlv2.GeneralParam, req *api.CampusTopicRcmdFeedReq) (resp *api.CampusTopicRcmdFeedReply, err error) {
	if req.CampusId <= 0 {
		return nil, errors.WithMessagef(ecode.RequestErr, "invalid req.CampusId")
	}

	// 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(ctx, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)

	eg := errgroup.WithContext(ctx)
	var feedInfo *mdlv2.CampusForumDynamicsInfo
	var topicInfo *mdlv2.CampusForumSquareInfo
	// 获取feed信息
	eg.Go(func(c context.Context) (err error) {
		feedInfo, err = s.dynDao.CampusForumDynamics(c, req.FromType, general.Mid, req.CampusId, req.Offset, attentions, general)
		if err != nil {
			return errors.WithMessagef(err, "failed to get CampusForumDynamics")
		}
		return
	})
	// 首页插入推荐话题卡
	if len(req.Offset) == 0 {
		eg.Go(func(c context.Context) (err error) {
			topicInfo, err = s.dynDao.CampusForumSquare(c, req.FromType, general.Mid, req.CampusId, general)
			if err != nil {
				xmetric.DynamicCardError.Inc(s.fromName(_handleTypeSchoolTopicFeed), mdlv2.DynamicName(int64(api.DynamicType_topic_rcmd)), "handle_error")
				log.Warn("CampusForumSquare mid(%d) campusId(%d) from(%s) failed", general.Mid, req.CampusId, _handleTypeSchoolTopicFeed)
			}
			// 不影响主要流程
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return
	}
	resp = &api.CampusTopicRcmdFeedReply{
		Toast: feedInfo.UpdateToast, HasMore: feedInfo.HasMore, Offset: feedInfo.PageOffset,
	}
	if len(feedInfo.Dyns) > 0 {
		// 处理动态feed
		dynCtx, err := s.getMaterial(ctx, getMaterialOption{general: general, dynamics: feedInfo.Dyns})
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to get dynCtx for feedInfo(%+v)", feedInfo)
		}
		dynCtx.CampusID = req.CampusId
		foldList := s.procListReply(ctx, feedInfo.Dyns, dynCtx, general, _handleTypeSchoolTopicFeed)
		s.procBackfill(ctx, dynCtx, general, foldList)
		resp.List = s.procFold(foldList, dynCtx, general)
	}

	// prepend话题推荐卡
	addPublishBtn := s.isCampusTopicPublishBtn(ctx, general)
	if rcmdTopicCard := topicInfo.ToV2DynamicItem(s.c, addPublishBtn &&
		req.FromType == api.CampusReqFromType_DYNAMIC); rcmdTopicCard != nil {
		finalList := make([]*api.DynamicItem, len(resp.List)+1)
		copy(finalList[1:], resp.List)
		finalList[0] = rcmdTopicCard
		resp.List = finalList
	}
	// 首页情况下，发布链接放页面上
	if req.FromType == api.CampusReqFromType_HOME && topicInfo != nil && len(topicInfo.PublishLink) > 0 {
		resp.JoinDiscuss = &api.IconButton{
			JumpUri: topicInfo.PublishLink,
		}
	}

	return
}

func (s *Service) CampusMngDetail(ctx context.Context, general *mdlv2.GeneralParam, param *api.CampusMngDetailReq) (*api.CampusMngDetailReply, error) {
	if param.CampusId <= 0 {
		return nil, ecode.RequestErr
	}
	req := &dyncampusgrpc.FetchCampusMngDetailReq{
		Uid:      general.Mid,
		CampusId: param.CampusId,
	}
	resp, err := s.dynDao.CampusMngDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	const _auditHintMsg = "内容正在审核中"

	itemList := make([]*api.CampusMngItem, 0, len(resp.Items)+1)
	// BasicInfo模块单独处理
	itemList = append(itemList, &api.CampusMngItem{
		AuditStatus:  api.CampusMngAuditStatus(dyncomn.CampusMngAuditStatus_CAMPUS_MNG_AUDIT_NONE),
		AuditMessage: _auditHintMsg,
		ItemType:     api.CampusMngItemType_campus_mng_basic_info,
		Item: &api.CampusMngItem_BasicInfo{BasicInfo: &api.CampusMngBasicInfo{
			CampusId:   resp.CampusID,
			CampusName: resp.CampusName,
			HintMsg:    "请认真填写如下信息，共建bilibili校园",
		}},
	})

	for _, item := range resp.Items {
		itemTmp := &api.CampusMngItem{
			AuditStatus:  api.CampusMngAuditStatus(item.AuditStatus),
			AuditMessage: _auditHintMsg,
		}
		// 模块类型和内容
		switch item.ItemType {
		case dyncomn.CampusMngItemType_CAMPUS_MNG_ITEM_BADGE:
			itemTmp.ItemType = api.CampusMngItemType_campus_mng_badge
			itemBadge, ok := item.Item.(*dyncampusgrpc.CampusMngItem_Badge)
			if !ok || itemBadge == nil {
				continue
			}
			badge := &api.CampusMngBadge{
				Title:         "校徽",
				BadgeUrl:      itemBadge.Badge.GetBadge(),
				UploadHintMsg: "支持透明背景的PNG圆形图",
			}
			itemTmp.Item = &api.CampusMngItem_Badge{Badge: badge}
		case dyncomn.CampusMngItemType_CAMPUS_MNG_ITEM_MOTTO:
			itemTmp.ItemType = api.CampusMngItemType_campus_mng_slogan
			itemSlogan, ok := item.Item.(*dyncampusgrpc.CampusMngItem_Motto)
			if !ok || itemSlogan == nil {
				continue
			}
			slogan := &api.CampusMngSlogan{
				Title:        "校训/Slogan",
				Slogan:       itemSlogan.Motto.GetMotto(),
				InputHintMsg: "请填写学校校训",
			}
			itemTmp.Item = &api.CampusMngItem_Slogan{Slogan: slogan}
		case dyncomn.CampusMngItemType_CAMPUS_MNG_ITEM_QUIZ:
			itemTmp.ItemType = api.CampusMngItemType_campus_mng_quiz
			itemQuiz, ok := item.Item.(*dyncampusgrpc.CampusMngItem_Quiz)
			if !ok || itemQuiz == nil {
				continue
			}
			quiz := &api.CampusMngQuiz{
				Title:       "入园题目",
				MoreLabel:   &api.CampusLabel{Text: "查看所有", Url: fmt.Sprintf("bilibili://campus/page/manage/quiz/%d", param.CampusId)},
				AddLabel:    "新建题目",
				SubmitLabel: "提交题目",
				QuizCount:   itemQuiz.Quiz.GetTotal(),
			}
			itemTmp.Item = &api.CampusMngItem_Quiz{Quiz: quiz}
		default:
			log.Warnc(ctx, "Unknown CampusMngItemType %+v while Fetching detail, CampusID: %d, CampusName: %s", item, resp.CampusID, resp.CampusName)
			continue
		}

		itemList = append(itemList, itemTmp)
	}

	ret := &api.CampusMngDetailReply{
		Items:               itemList,
		TopHintBarMsg:       "页面出新模块啦，请更新到最新版本使用",
		BottomSubmitHintMsg: "提交内容后会在1周内收到是否上线的通知，请耐心等待，感谢支持与配合~",
		CampusId:            resp.CampusID,
		CampusName:          resp.CampusName,
	}
	return ret, nil
}

func (s *Service) CampusMngSubmit(ctx context.Context, general *mdlv2.GeneralParam, param *api.CampusMngSubmitReq) (*api.CampusMngSubmitReply, error) {
	if param.CampusId <= 0 || len(param.ModifiedItems) <= 0 {
		return nil, ecode.RequestErr
	}

	itemList := make([]*dyncampusgrpc.CampusMngItem, 0, len(param.ModifiedItems))
	for _, item := range param.ModifiedItems {
		itemTmp := &dyncampusgrpc.CampusMngItem{
			IsDel: item.IsDel,
		}
		// 模块类型和内容
		switch item.GetItemType() {
		case api.CampusMngItemType_campus_mng_badge:
			itemTmp.ItemType = dyncomn.CampusMngItemType_CAMPUS_MNG_ITEM_BADGE
			itemBadge, ok := item.Item.(*api.CampusMngItem_Badge)
			if !ok || itemBadge == nil {
				continue
			}
			badge := &dyncampusgrpc.CampusMngBadge{Badge: itemBadge.Badge.GetBadgeUrl()}
			itemTmp.Item = &dyncampusgrpc.CampusMngItem_Badge{Badge: badge}
		case api.CampusMngItemType_campus_mng_slogan:
			itemTmp.ItemType = dyncomn.CampusMngItemType_CAMPUS_MNG_ITEM_MOTTO
			itemSlogan, ok := item.Item.(*api.CampusMngItem_Slogan)
			if !ok || itemSlogan == nil {
				continue
			}
			slogan := &dyncampusgrpc.CampusMngMotto{Motto: itemSlogan.Slogan.GetSlogan()}
			itemTmp.Item = &dyncampusgrpc.CampusMngItem_Motto{Motto: slogan}
		case api.CampusMngItemType_campus_mng_basic_info, api.CampusMngItemType_campus_mng_quiz:
			// do nothing
			continue
		default:
			log.Warnc(ctx, "Unknown CampusMngItemType %+v for submitting, CampusID: %d", item, param.CampusId)
			continue
		}

		itemList = append(itemList, itemTmp)
	}
	if len(itemList) <= 0 {
		return nil, ecode.RequestErr
	}

	req := &dyncampusgrpc.UpdateCampusMngDetailReq{
		Uid:      general.Mid,
		CampusId: param.CampusId,
		Items:    itemList,
	}
	resp, err := s.dynDao.CampusMngSubmit(ctx, req)
	if err != nil {
		return nil, err
	}

	ret := &api.CampusMngSubmitReply{Toast: resp.Toast}
	return ret, nil
}

func (s *Service) CampusMngQuizOperate(ctx context.Context, general *mdlv2.GeneralParam, params *api.CampusMngQuizOperateReq) (*api.CampusMngQuizOperateReply, error) {
	if params.CampusId <= 0 {
		return nil, ecode.RequestErr
	}
	var (
		operateFn func() (*mdlv2.CampusQuizOperateRes, error)
		respFn    func(res *mdlv2.CampusQuizOperateRes) *api.CampusMngQuizOperateReply
	)

	switch params.Action {
	case api.CampusMngQuizAction_campus_mng_quiz_act_add, api.CampusMngQuizAction_campus_mng_quiz_act_del:
		if len(params.Quiz) <= 0 {
			return nil, ecode.RequestErr
		}
		operateFn = func() (*mdlv2.CampusQuizOperateRes, error) {
			quiz := make([]*dyncampusgrpc.QuestionItem, 0, len(params.Quiz))
			for _, q := range params.Quiz {
				quiz = append(quiz, &dyncampusgrpc.QuestionItem{
					Id: q.GetQuizId(), Title: q.GetQuestion(),
					CorrectAnswer: q.GetCorrectAnswer(),
					WrongAnswer:   q.GetWrongAnswerList(),
				})
			}
			return s.dynDao.CampusQuizOperate(ctx, &dyncampusgrpc.OperateQuestionReq{
				CampusId: params.CampusId, Uid: general.Mid,
				OperateType: dyncomn.QuestionOperate(params.Action),
				Question:    quiz,
			})
		}
		respFn = func(_ *mdlv2.CampusQuizOperateRes) *api.CampusMngQuizOperateReply {
			toast := "题目删除成功"
			if params.Action == api.CampusMngQuizAction_campus_mng_quiz_act_add {
				toast = "题目已提交，请等待审核"
			}
			return &api.CampusMngQuizOperateReply{Toast: toast}
		}
	default:
		operateFn = func() (*mdlv2.CampusQuizOperateRes, error) {
			return s.dynDao.CampusQuizList(ctx, &dyncampusgrpc.FetchQuestionListReq{
				CampusId: params.CampusId, Uid: general.Mid,
			})
		}
		respFn = func(res *mdlv2.CampusQuizOperateRes) *api.CampusMngQuizOperateReply {
			return &api.CampusMngQuizOperateReply{
				QuizTotal: res.Total, Quiz: res.ToQuizDetailItems(),
			}
		}
	}
	res, err := operateFn()
	if err != nil {
		return nil, err
	}

	return respFn(res), nil
}
