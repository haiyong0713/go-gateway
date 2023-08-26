package dynamicV2

import (
	"context"
	"strconv"

	"go-common/library/log"
	xecode "go-gateway/app/app-svr/app-dynamic/ecode"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
)

func (s *Service) DynDetail(c context.Context, general *mdlv2.GeneralParam, req *api.DynDetailReq) (*api.DynDetailReply, error) {
	general.Config = req.Config
	var (
		requestID, adExtra string
	)
	if req.GetAdParam() != nil {
		requestID = req.GetAdParam().GetRequestId()
		adExtra = req.GetAdParam().GetAdExtra()
	}
	dynDetail, err := s.dynDao.DynDetail(c, general, req)
	if err != nil {
		xmetric.DynamicCoreAPI.Inc("动态详情页", "request_error")
		log.Errorc(c, "DynDetail s.dynDao.DynDetail mid(%d) error(%v)", general.Mid, err)
		return nil, err
	}

	// 部分物料只有详情页需要
	general.AdFrom = req.From
	dynCtx, err := s.getMaterial(c, getMaterialOption{
		general: general, dynamics: []*mdlv2.Dynamic{dynDetail.Dynamic},
		requestID: requestID, adExtra: adExtra,
	})
	if err != nil {
		return nil, err
	}
	item := s.procDetailReply(c, dynDetail, dynCtx, general, _handleTypeView)
	if item == nil {
		return nil, xecode.DynViewNotFound
	}
	s.procDetailBack(c, dynCtx, general, item)
	res := &api.DynDetailReply{
		Item: item,
	}
	return res, nil
}

func (s *Service) procDetailReply(c context.Context, dyn *mdlv2.DynDetailRes, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, from string) *api.DynamicItem {
	dynCtx.From = from
	dynCtx.Dyn = dyn.Dynamic                                     // 原始数据
	dynCtx.DynamicItem = &api.DynamicItem{Extend: &api.Extend{}} // 聚合结果
	dynCtx.Interim = &mdlv2.Interim{}                            // 临时逻辑
	dynCtx.Recmd = dyn.Recommend                                 // 相关推荐
	// mid > int32老版本抛弃当前卡片
	if s.checkMidMaxInt32(c, dynCtx.Dyn.UID, general) {
		return nil
	}
	handlerList, ok := s.getHandlerView(c, dynCtx, general)
	if !ok {
		log.Warn("dynamic mid(%v) getHandlerList !ok", general.Mid)
		return nil
	}
	// 执行拼接func
	if err := s.conveyer(c, dynCtx, general, handlerList...); err != nil {
		xmetric.DynamicCardError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "conveyer_error")
		log.Warn("dynamic mid(%v) conveyer, err %v", general.Mid, err)
		return nil
	}
	if dynCtx.Interim.IsPassCard {
		xmetric.DynamicCardError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "pass_card")
		log.Warn("dynamic mid(%v) IsPassCard dynid %v", general.Mid, dyn.Dynamic.DynamicID)
		return nil
	}
	return dynCtx.DynamicItem
}

func (s *Service) procDetailBack(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, item *api.DynamicItem) {
	// 聚合回填物料
	s.BackfillGetMaterial(c, dynCtx, general)
	// 遍历回填
	if item == nil {
		return
	}
	s.backfill(c, dynCtx, item, general)
}

// 相关推荐
func (s *Service) detailRecommend(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Recmd == nil {
		return nil
	}
	rcmd := &api.Module_ModuleRecommend{
		ModuleRecommend: &api.ModuleRecommend{
			ModuleTitle: dynCtx.Recmd.ModuleTitle,
			Image:       dynCtx.Recmd.Img,
			Tag:         dynCtx.Recmd.Tag,
			Title:       dynCtx.Recmd.Title,
			JumpUrl:     dynCtx.Recmd.JumpUrl,
			Ad:          dynCtx.Recmd.Ad,
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_recommend,
		ModuleItem: rcmd,
	}
	dynCtx.Modules = append(dynCtx.Modules, module)
	return nil
}

func (s *Service) LikeList(c context.Context, general *mdlv2.GeneralParam, req *api.LikeListReq) (*api.LikeListReply, error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, cheese, ugcSeason, batchListFavorite, err := s.followings(c, general.Mid, true, true, general)
	if err != nil {
		return nil, err
	}
	attentions := mdlv2.GetAttentionsParams(general.Mid, following, pgcFollowing, cheese, ugcSeason, batchListFavorite)
	reply, err := s.dynDao.LikeList(c, general, req, attentions)
	if err != nil {
		xmetric.DynamicCoreAPI.Inc("详情页(点赞列表)", "request_error")
		return nil, err
	}
	var (
		mids []int64
	)
	for _, v := range reply.LikeList {
		mids = append(mids, v.Uid)
	}
	cardm, err := s.accountDao.Cards3New(c, mids)
	if err != nil {
		xmetric.DyanmicItemAPI.Inc("/account.service.Account/Cards3", "request_error")
		log.Warn("getMaterial mid(%v) Cards3New(%v) error(%v)", general.Mid, mids, err)
		return nil, err
	}
	var userMdls []*api.ModuleAuthor
	for _, v := range reply.LikeList {
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, v.Uid, general) {
			continue
		}
		userInfo, ok := cardm[v.Uid]
		if !ok {
			continue
		}
		userMdl := &api.ModuleAuthor{
			Mid: userInfo.Mid,
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
			Attend: int32(v.Attend),
		}
		if userInfo.Pendant.ImageEnhance != "" { // 动效图优先
			userMdl.Author.Pendant.Image = userInfo.Pendant.ImageEnhance
		}
		userMdls = append(userMdls, userMdl)
	}
	res := &api.LikeListReply{
		List:       userMdls,
		TotalCount: reply.TotalCount,
	}
	if reply.HasMore == 1 {
		res.HasMore = true
	}
	return res, nil
}

func (s *Service) RepostList(c context.Context, general *mdlv2.GeneralParam, req *api.RepostListReq) (*api.RepostListRsp, error) {
	var (
		_max       = 20 // 转发列表一页最大个数
		reply      *mdlv2.RepostListRes
		err        error
		repostType = api.RepostType_repost_hot // 转发类型，默认热门抓发列表
	)
	switch req.RepostType {
	case api.RepostType_repost_hot:
		// 热门转发
		reply, err = s.dynDao.HotRepostList(c, general, req, _max)
		if err != nil {
			xmetric.DyanmicRelationAPI.Inc("详情页(热门-转发列表)", "request_error")
			log.Error("%+v", err)
			return nil, err
		}
		// 当前是热门，并且没有更多，从普通热门里面获取
		if !reply.HasMore {
			// 第一刷普通offset传空
			req.Offset = ""
			repostType = api.RepostType_repost_general // 热门没有更多则转发类型改为普通类型
			hotPs := len(reply.Dynamics)
			if generalPs := _max - hotPs; generalPs > 0 {
				repost, err := s.dynDao.RepostList(c, general, req, generalPs)
				if err != nil {
					xmetric.DyanmicRelationAPI.Inc("详情页(普通-转发列表)", "request_error")
					log.Error("%+v", err)
					return nil, err
				}
				// 往后追加
				reply.Dynamics = append(reply.Dynamics, repost.Dynamics...)
				reply.Offset = repost.Offset
				reply.HasMore = repost.HasMore
				reply.TotalCount = repost.TotalCount
			}
		}
	default:
		// 普通转发
		reply, err = s.dynDao.RepostList(c, general, req, _max)
		if err != nil {
			xmetric.DyanmicRelationAPI.Inc("详情页(普通-转发列表)", "request_error")
			log.Error("%+v", err)
			return nil, err
		}
		repostType = api.RepostType_repost_general
	}
	dynCtx, err := s.getMaterial(c, getMaterialOption{general: general, dynamics: reply.Dynamics})
	if err != nil {
		return nil, err
	}
	// Step 4. 对物料信息处理，获取详情列表
	foldList := s.procListReply(c, reply.Dynamics, dynCtx, general, _handleTypeRepost)
	// Step 5. 回填
	s.procBackfill(c, dynCtx, general, foldList)
	res := &api.RepostListRsp{
		List:       s.procFold(foldList, dynCtx, general),
		Offset:     reply.Offset,
		HasMore:    reply.HasMore,
		TotalCount: reply.TotalCount,
		RepostType: repostType,
	}
	return res, nil
}
