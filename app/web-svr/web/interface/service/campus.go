package service

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"
	cpmodel "go-gateway/app/web-svr/web/interface/model/campus"
	"go-gateway/app/web-svr/web/interface/model/rcmd"

	"github.com/pkg/errors"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	campusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

const (
	_feed = 1
	_rcmd = 2
)

var _emptyOfficialList = make([]*cpmodel.OfficialAccountInfo, 0)

// 校内推荐页顶部展示信息（包括tab/banner/学校）
func (s *Service) Pages(ctx context.Context, req *cpmodel.CampusRcmdReq) (*cpmodel.PagesReply, error) {
	data, err := s.campusDao.Pages(ctx, req)
	if err != nil {
		log.Error("【@Pages】PC校园获取数据失败：%+v", err)
		return nil, err
	}
	res := &cpmodel.PagesReply{
		PageType: data.PageType,
	}

	switch data.PageType {
	case _feed:
		res.MajorPageInfo = cpmodel.FromMajorPageInfo(data.MajorPage)
	case _rcmd:
		res.RcmdPageInfo = &campusgrpc.NearbyRcmdInfo{}
		if data.NearbyRcmdPage != nil && data.NearbyRcmdPage.TopShow != nil {
			res.RcmdPageInfo.TopShow = data.NearbyRcmdPage.TopShow
		} else {
			log.Warn("【@Pages】 data.NearbyRcmdPage.TopShow is Null")
		}
	default:
		log.Warn("【@Pages】 miss page_type(%d)", data.PageType)
		return res, nil
	}
	return res, nil
}

// 学校搜索
func (s *Service) SchoolSearch(ctx context.Context, keywords string, ps, offset uint64, fromType string) (*cpmodel.SchoolSearchRep, error) {
	data, err := s.campusDao.SchoolSearch(ctx, keywords, ps, offset)
	if err != nil {
		log.Error("【@SchoolSearch】学校搜索数据失败：%+v", err)
		return nil, err
	}
	res := &cpmodel.SchoolSearchRep{
		Results: data.Results,
		HasMore: data.HasMore,
		Offset:  data.Offset + uint64(len(data.Results)),
	}
	return res, nil
}

// 学校推荐
func (s *Service) SchoolRecommend(ctx context.Context, mid uint64, lat, lng float32) ([]*campusgrpc.CampusInfo, error) {
	data, err := s.campusDao.SchoolRecommend(ctx, mid, lat, lng)
	if err != nil {
		log.Error("【@SchoolRecommend】获取学校推荐数据失败：%+v", err)
		return make([]*campusgrpc.CampusInfo, 0), nil
	}
	return data, nil
}

// 官方账号
func (s *Service) OfficialAccounts(ctx context.Context, req *cpmodel.CampusOfficialReq) (res []*cpmodel.OfficialAccountInfo, err error) {
	data, offErr := s.campusDao.OfficialAccounts(ctx, req)
	if offErr != nil {
		log.Error("【@OfficialAccounts】获取学校官方账号数据失败：%+v", err)
		return _emptyOfficialList, offErr
	}
	var (
		mids       []int64
		cardsReply *accmdl.CardsReply
		cardErr    error
	)
	uids := data.GetUids()
	if len(uids) == 0 {
		res = _emptyOfficialList
		return
	}
	for _, uid := range uids {
		mids = append(mids, int64(uid))
	}
	relInfos := make(map[int64]*relmdl.StatReply, len(mids))
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		if cardsReply, cardErr = s.accGRPC.Cards3(ctx, &accmdl.MidsReq{Mids: mids}); cardErr != nil {
			log.Error("【@OfficialAccounts】 s.accGRPC.Cards3(%v) error(%v)", mids, cardErr)
			return cardErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if relReply, relErr := s.relationGRPC.Stats(ctx, &relmdl.MidsReq{Mids: mids}); relErr != nil {
			log.Error("【@OfficialAccounts】 s.relationGRPC.Stats(%d,%v) error(%v)", req.Mid, mids, relErr)
		} else if relReply != nil {
			relInfos = relReply.StatReplyMap
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}

	for _, mid := range mids {
		if card, ok := cardsReply.Cards[mid]; ok {
			info := &cpmodel.OfficialAccountInfo{}
			info.FromCard(card)
			if stat, ok := relInfos[mid]; ok {
				info.Follower = stat.Follower
			}
			res = append(res, info)
		}
	}
	return
}

// 话题列表
func (s *Service) CampusTopicList(ctx context.Context, mid, campusId uint64, offset int64) (*campusgrpc.TopicListReply, error) {
	data, err := s.campusDao.TopicList(ctx, mid, campusId, offset)
	if err != nil {
		log.Error("【@CampusTopicList】获取学校话题列表数据失败：%+v", err)
		data = &campusgrpc.TopicListReply{
			HasMore: 0,
			Offset:  "",
			List:    make([]*campusgrpc.TopicListInfo, 0),
		}
		return data, nil
	}
	return data, nil
}

// 入校必看
func (s *Service) OfficialDynamics(ctx context.Context, req *cpmodel.CampusOfficialReq) (*cpmodel.OfficialDynamicsReply, error) {
	data, err := s.campusDao.OfficialDynamics(ctx, req)
	if err != nil {
		log.Error("【@OfficialDynamics】获取入校必看数据失败：%+v", err)
		return nil, err
	}
	res := &cpmodel.OfficialDynamicsReply{
		HasMore: int(data.HasMore),
		Offset:  int(data.Offset),
	}
	dyConfig := data.GetDynsConfig()
	if len(dyConfig) == 0 {
		// found nothing
		res.RcmdItems = make([]*cpmodel.OfficialDynamicsItem, 0)
		log.Error("【@OfficialDynamics】not found dyconfig")
		return res, ecode.NothingFound
	}
	var (
		aids []int64
	)
	for _, v := range dyConfig {
		aids = append(aids, int64(v.GetRid()))
	}
	arcs, err1 := s.batchArchives(ctx, aids)
	if err1 != nil {
		log.Error("【@OfficialDynamics】get arcs failed：%v", err1)
		return nil, err1
	}
	for _, dyItem := range dyConfig {
		arc := arcs[int64(dyItem.GetRid())]
		if !arc.IsNormal() {
			continue
		}
		rcmdItem := &cpmodel.OfficialDynamicsItem{
			DyId:    int(dyItem.GetDynamicId()),
			Desc:    dyItem.GetReason(),
			ArcInfo: &rcmd.Item{},
		}
		rcmdItem.FromArc(arc)
		res.RcmdItems = append(res.RcmdItems, rcmdItem)
	}
	return res, nil
}

var _validFeedbackBizType = map[byte]struct{}{
	1: {}, 2: {}, 3: {}, 4: {},
}

// 用户反馈
func (s *Service) CampusFeedback(ctx context.Context, req *cpmodel.CampusFeedbackReq) (resp *cpmodel.CampusFeedbackReply, err error) {
	if err = json.Unmarshal([]byte(req.Infos), &req.List); err != nil {
		log.Error("【@CampusFeedback】json parse error: (%v)", err)
		return nil, errors.New("json parse error")
	}
	for _, info := range req.List {
		if _, ok := _validFeedbackBizType[byte(info.BizType)]; !ok {
			log.Warnc(ctx, "【@CampusFeedback】 unknown bizType(%d): %+v", info.BizType, *info)
			return nil, errors.WithMessage(ecode.RequestErr, "unknown bizType")
		}
		if info.BizId == "" {
			return nil, errors.WithMessage(ecode.RequestErr, "invalid bizId")
		}
	}
	err = s.campusDao.CampusFeedback(ctx, req)
	if err != nil {
		log.Errorc(ctx, "【@CampusFeedback】 dao error: (%v)", err)
		return nil, err
	}
	resp = &cpmodel.CampusFeedbackReply{
		Message: "已成功提交",
	}
	return
}

// 校园榜单（十大热点）
func (s *Service) CampusBillboard(ctx context.Context, mid, campus_id int64, version_code string) (resp *cpmodel.CampusBillBoardReply, err error) {
	bi, err := s.campusDao.CampusBillboard(ctx, mid, campus_id, version_code)
	if err != nil {
		log.Error("【@CampusBillboard】dynDao.CampusBillboardMeta error: (%v)", err)
		return nil, err
	}
	// 正常渲染内容
	resp = &cpmodel.CampusBillBoardReply{
		Title:       bi.GetTitleName(),
		HelpUri:     bi.GetJumpUrl(),
		CampusName:  bi.GetCampusName(),
		BuildTime:   bi.GetBuildTime(),
		VersionCode: bi.GetVersionCode(),
		BindNotice:  bi.GetBindNotice(),
		CampusId:    bi.GetCampusId(),
		UpdateToast: bi.GetToast(),
	}
	if len(bi.List) <= 0 {
		return
	}
	if resp.List, err = s.campusBillboardCards(ctx, bi.GetList(), mid); err != nil {
		log.Error("【@CampusBillboard】Get BillboardCards Failed: (%v)", err)
	}
	return resp, nil
}

func (s *Service) campusBillboardCards(ctx context.Context, bi []*campusgrpc.BoardItem, mid int64) ([]*cpmodel.CampusBillBoardRcmdItem, error) {
	var (
		aids       []int64
		dynamicIDs []int64
		arcm       map[int64]*arcmdl.Arc
		dynamicm   map[int64]*model.DynamicCard
	)
	for _, item := range bi {
		rid := item.Dyns.GetRid()
		dyId := item.GetDyns().GetDynId()
		if rid == 0 {
			continue
		}
		switch item.Dyns.GetType() {
		case cpmodel.Campus_Dy_Arc_Type: // 稿件
			aids = append(aids, rid)
		case cpmodel.Campus_Dy_Draw_Type: // 动态
			dynamicIDs = append(dynamicIDs, dyId)
		}
	}
	group := errgroup.WithContext(ctx)
	// 获取稿件详情
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcm, err = s.dao.Arcs(ctx, aids); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if len(dynamicIDs) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if dynamicm, err = s.dao.DrawInfos(ctx, mid, dynamicIDs); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("@campusBillboardCards 获取稿件出错%v", err)
		return nil, err
	}
	var res []*cpmodel.CampusBillBoardRcmdItem
	const _maxBillboardItems = 10 // 排行榜最大稿件数量
	for _, rc := range bi {
		if len(res) >= _maxBillboardItems {
			break
		}
		dyn := rc.Dyns
		if !dyn.Visible {
			continue
		}
		rid := rc.GetDyns().GetRid()
		dyId := rc.GetDyns().DynId
		var item *cpmodel.CampusBillBoardRcmdItem
		switch rc.GetDyns().GetType() {
		case cpmodel.Campus_Dy_Arc_Type: // 稿件
			if arc, ok := arcm[rid]; ok && arc.IsNormal() {
				item = &cpmodel.CampusBillBoardRcmdItem{}
				item.FromArc(arc)
			}
		case cpmodel.Campus_Dy_Draw_Type: // 动态
			if dynamic, ok := dynamicm[dyId]; ok {
				var drawDetail *model.DrawDetail
				if err := json.Unmarshal([]byte(dynamic.Card), &drawDetail); err != nil {
					log.Error("【@CampusBillboardCards】图文动态解包出错%v", err)
					continue
				}
				item = &cpmodel.CampusBillBoardRcmdItem{}
				item.FromDraw(dynamic, rid, drawDetail)
			}
		default:
			log.Error("CampusBillboard unhandled billboardItem: (%+v)", rc)
		}
		if item != nil {
			item.Reason = rc.GetReason()
			// item.DyId = strconv.FormatInt(dyId, 10)
			res = append(res, item)
		}
	}
	return res, nil
}

// 校园推荐（缺少上报）
func (s *Service) CampusNearbyRcmd(ctx context.Context, req *cpmodel.CampusNearbyRcmdReq) (resp *cpmodel.CampusNearbyRcmdReply, err error) {
	const (
		_av = "av"
	)
	var list []*rcmd.AITopRcmd

	// 获取数据
	data, _, userFeature, err := s.campusDao.CampusNearbyRcmd(ctx, req)
	resp = &cpmodel.CampusNearbyRcmdReply{
		UserFeature: userFeature,
	}
	list = data
	if err != nil {
		log.Error("【@CampusNearbyRcmd】获取校园附近推荐数据失败：%v", err)
		dataHot, errHot := s.campusDao.CampusHotRcmd(ctx)
		if errHot != nil {
			log.Error("【@CampusNearbyRcmd】获取校园附近推荐灾备数据失败：%v", errHot)
			err = errHot
			return
		}
		list = dataHot
	}
	if len(list) <= 0 {
		return
	}

	// 获取稿件
	var aids []int64
	for _, v := range list {
		switch v.Goto {
		case _av:
			aids = append(aids, v.ID)
		default:
			log.Error("【@CampusNearbyRcmd】无效的goto:%v", v.Goto)
		}
	}
	if len(aids) == 0 {
		return nil, errors.New("【@CampusNearbyRcmd】 rcmd aids len is 0")
	}
	arcs, err1 := s.batchArchives(ctx, aids)
	if err1 != nil {
		return nil, err1
	}
	var (
		rcmdItem []*rcmd.Item
	)
	for _, v := range data {
		switch v.Goto {
		case _av:
			if arc, ok := arcs[v.ID]; ok && arc != nil && arc.IsNormal() {
				i := &rcmd.Item{}
				i.FromArc(arc, v)
				rcmdItem = append(rcmdItem, i)
			}
		}
	}
	resp.Items = rcmdItem
	return
}

// 红点
func (s *Service) CampusRedDot(ctx context.Context, req *cpmodel.CampusRedDotReq) (*cpmodel.CampusRedDotReply, error) {
	return s.campusDao.RedDot(ctx, req)
}
