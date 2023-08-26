package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	seasongrpc "go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/app/web-svr/web/interface/model"
	gateecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	webgrpc "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	activegrpc "git.bilibili.co/bapis/bapis-go/manager/service/active"

	"github.com/pkg/errors"
)

const (
	_midBulkSize                   = 50
	_actReserveStateReserved       = 1
	_favTypeArc              int32 = 2
	_favOTypeSeason          int32 = 21
	_platformActivity              = "pc"
	_commonActPlayPc               = 3
)

// nolint: gocognit
func (s *Service) ActivitySeason(ctx context.Context, mid int64, req *webgrpc.ActivitySeasonReq, buvid string) (*webgrpc.ActivitySeasonReply, error) {
	actSeason, err := s.commonActivity(ctx, mid, req.ActivityKey, false, buvid)
	if err != nil {
		return nil, err
	}
	nowTs := time.Now().Unix()
	nowStatus := webgrpc.ActivitySeasonStatus_StatusView
	if actSeason.Live != nil && nowTs >= actSeason.Live.StartTime && nowTs <= actSeason.Live.EndTime {
		nowStatus = webgrpc.ActivitySeasonStatus_StatusLive
		actSeason.Live.NowTime = nowTs
	}
	reply := &webgrpc.ActivitySeasonReply{
		Status: nowStatus,
		Title:  actSeason.Title,
		Theme:  actSeason.Theme,
	}
	// 直播态有游戏
	if nowStatus == webgrpc.ActivitySeasonStatus_StatusLive {
		reply.Live = actSeason.Live
		reply.Game = actSeason.Game
	}
	// 需要稿件数据
	needView := req.Bvid != "" || req.Aid > 0 || nowStatus == webgrpc.ActivitySeasonStatus_StatusView
	// 订阅模块 1-直播前，2-直播中，3-直播后，4-HD通用
	reply.Subscribe = func() *webgrpc.ActivitySubscribe {
		actSub := &webgrpc.ActivitySubscribe{
			Status:            false,
			Title:             actSeason.Title,
			SeasonStatView:    int64(actSeason.SeasonView.Season.Stat.View),
			SeasonStatDanmaku: int64(actSeason.SeasonView.Season.Stat.Danmaku),
		}
		// 直播开始前订阅按钮
		if actSeason.Live != nil && nowTs < actSeason.Live.StartTime {
			menu, ok := actSeason.Subscribe[model.ActSubTypeBeforeLive]
			if !ok || menu == nil || menu.MenuId <= 0 {
				log.Error("大型活动告警 Subscribe name:%s BeforeLive not found", actSeason.Title)
			}
			model.FillSubscribe(actSub, menu, model.ActSubTypeBeforeLive, actSeason.SeasonView.Season.ID, "")
			return actSub
		}
		if nowStatus == webgrpc.ActivitySeasonStatus_StatusLive && needView {
			menu, ok := actSeason.Subscribe[model.ActSubTypeDuringLive]
			if !ok || menu == nil {
				log.Error("大型活动告警 Subscribe name:%s DuringLive not found", actSeason.Title)
			}
			model.FillSubscribe(actSub, menu, model.ActSubTypeDuringLive, 0, actSeason.ActivityURL)
			return actSub
		}
		menu, ok := actSeason.Subscribe[model.ActSubTypeAfterLive]
		if !ok || menu == nil {
			log.Error("大型活动告警 Subscribe name:%s AfterLive not found", actSeason.Title)
		}
		model.FillSubscribe(actSub, menu, model.ActSubTypeAfterLive, actSeason.SeasonView.Season.ID, "")
		return actSub
	}()
	if !needView {
		return reply, nil
	}
	reply.View, err = func() (*webgrpc.ActivityView, error) {
		aid := aidFromAidAndBvid(req.Aid, req.Bvid)
		outView := &webgrpc.ActivityView{Sections: model.CopyFromActivitySection(actSeason.SeasonView.Sections)}
		var firstEpArc, hitEpArc *model.ActivityView
		for _, sec := range actSeason.SeasonView.Sections {
			if sec == nil {
				continue
			}
			for _, ep := range sec.Episodes {
				if ep == nil {
					continue
				}
				if firstEpArc == nil && ep.ActivityView != nil {
					firstEpArc = ep.ActivityView
				}
				if aid > 0 && ep.Aid == aid && ep.ActivityView != nil {
					hitEpArc = ep.ActivityView
				}
			}
		}
		// aid==0或者请求aid未找到，使用剧集第一个有效稿件
		outEpArc := func() *model.ActivityView {
			if hitEpArc != nil {
				return hitEpArc
			}
			return firstEpArc
		}()
		// unexpected error season data 判断了一定有可用稿件
		if outEpArc == nil {
			return nil, ecode.NothingFound
		}
		model.FillFromActivityView(outView, outEpArc)
		outView.Bvid = s.avToBv(outView.Arc.Aid)
		return outView, nil
	}()
	if err != nil {
		log.Error("大型活动告警 ActivitySeason name:%s req:%+v 无稿件数据", actSeason.Title, req)
		return nil, err
	}
	if mid <= 0 {
		return reply, nil
	}
	eg := errgroup.WithContext(ctx)
	// 用户和staff关注关系
	var mids []int64
	for _, v := range reply.View.Staff {
		if v == nil {
			continue
		}
		mids = append(mids, v.Mid)
	}
	if len(mids) > 0 {
		eg.Go(func(ctx context.Context) error {
			relationResp, relaErr := s.relationGRPC.Relations(ctx, &relagrpc.RelationsReq{Mid: mid, Fid: mids, RealIp: metadata.String(ctx, metadata.RemoteIP)})
			if relaErr != nil {
				log.Error("ActivitySeason relationGRPC.Relations mid:%d fid:%v error:%+v", mid, mids, relaErr)
				return nil
			}
			relations := relationResp.GetFollowingMap()
			for _, v := range reply.View.Staff {
				if v == nil {
					continue
				}
				if follow, ok := relations[v.Mid]; ok && follow != nil {
					v.Relation = &webgrpc.Relation{
						Attribute: int64(follow.Attribute),
						Tag:       follow.Tag,
						Special:   int64(follow.Special),
					}
				}
			}
			return nil
		})
	}
	// 判断按钮状态
	var needSeasonFav bool
	switch reply.Subscribe.OrderType {
	case webgrpc.OrderType_TypeOrderActivity:
		eg.Go(func(ctx context.Context) error {
			act, actErr := s.actGRPC.ActRelationInfo(ctx, &actgrpc.ActRelationInfoReq{Id: reply.Subscribe.GetReserveParam().GetReserveId(), Mid: mid, Specific: "reserve"})
			if actErr != nil {
				return actErr
			}
			if act != nil && act.ReserveItems != nil && act.ReserveItems.State == _actReserveStateReserved {
				reply.Subscribe.Status = true
			}
			return nil
		})
	case webgrpc.OrderType_TypeFavSeason:
		needSeasonFav = true
	default:
	}
	// 用户稿件关系
	if aid := reply.View.GetArc().GetAid(); aid > 0 {
		eg.Go(func(ctx context.Context) error {
			var seasonID int64
			if needSeasonFav {
				seasonID = actSeason.SeasonView.Season.ID
			}
			reqUser := s.arcRelation(ctx, mid, aid, seasonID, 0)
			reply.View.ReqUser = &webgrpc.ReqUser{
				Favorite: reqUser.Favorite,
				Like:     reqUser.Like,
				Dislike:  reqUser.Dislike,
				Multiply: reqUser.Coin,
			}
			if needSeasonFav {
				reply.Subscribe.Status = reqUser.SeasonFav
			}
			return nil
		})
		if actSeason.IsContainedRecom { //如果包含AI的相关推荐
			eg.Go(func(ctx context.Context) error {
				arcs, _, _, err := s.RelatedArcs(ctx, aid, mid, buvid, false, false, false, true, nil)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				var rs []*webgrpc.Relate
				for _, val := range arcs {
					r := &webgrpc.Relate{
						Arc:        model.CopyFromBvArc(val),
						Bvid:       val.Bvid,
						SeasonType: val.SeasonType,
					}
					rs = append(rs, r)
				}
				if len(rs) == 0 {
					return nil
				}
				rightRelate := &webgrpc.OperationRelate{}
				if reply.View.RightRelate != nil {
					*rightRelate = *reply.View.RightRelate
				}
				rightRelate.AiRelateItem = rs
				reply.View.RightRelate = rightRelate
				return nil
			})
		}
		eg.Go(func(ctx context.Context) error {
			desc, descV2, mids, err := s.description(ctx, aid, reply.GetView().GetArc().Desc)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			reply.View.Arc.Desc = desc
			var accReply *accmdl.InfosReply
			if len(mids) != 0 {
				accReply, err = s.accGRPC.Infos3(ctx, &accmdl.MidsReq{Mids: mids})
				if err != nil {
					log.Error("%+v", err)
				}
			}
			desc2 := s.DescV2ParamsMerge(ctx, descV2, accReply)
			var dV2 []*webgrpc.DescV2
			for _, val := range desc2 {
				dV2 = append(dV2, &webgrpc.DescV2{
					RawText: val.RawText,
					Type:    int64(val.Type),
					BizId:   val.BizId,
				})
			}
			reply.View.Arc.DescV2 = dV2
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("ActivitySeason req:%+v eg.Wait error:%v", req, err)
	}
	return reply, nil
}

// nolint: gocognit
func (s *Service) ActivityArchive(ctx context.Context, mid int64, req *webgrpc.ActivityArchiveReq, buvid string) (*webgrpc.ActivityArchiveReply, error) {
	aid := aidFromAidAndBvid(req.Aid, req.Bvid)
	if aid <= 0 {
		return nil, ecode.RequestErr
	}
	actSeason, err := s.commonActivity(ctx, mid, req.ActivityKey, false, buvid)
	if err != nil {
		return nil, err
	}
	actView := func() *model.ActivityView {
		for _, sec := range actSeason.SeasonView.Sections {
			if sec == nil {
				continue
			}
			for _, ep := range sec.Episodes {
				if ep == nil {
					continue
				}
				if ep.Aid == aid {
					return ep.ActivityView
				}
			}
		}
		return nil
	}()
	if actView == nil {
		return nil, ecode.NothingFound
	}
	reply := &webgrpc.ActivityArchiveReply{
		Arc:          actView.Arc,
		Bvid:         s.avToBv(aid),
		Pages:        model.CopyFromArcPageGRPC(actView.Pages),
		Staff:        actView.StaffInfo,
		RightRelate:  actView.RightRelate,
		BottomRelate: actView.BottomRelate,
	}
	if mid <= 0 {
		return reply, nil
	}
	eg := errgroup.WithContext(ctx)
	var mids []int64
	for _, v := range reply.Staff {
		if v == nil {
			continue
		}
		mids = append(mids, v.Mid)
	}
	if len(mids) > 0 {
		eg.Go(func(ctx context.Context) error {
			relations := func() map[int64]*relagrpc.FollowingReply {
				relationResp, relaErr := s.relationGRPC.Relations(ctx, &relagrpc.RelationsReq{Mid: mid, Fid: mids, RealIp: metadata.String(ctx, metadata.RemoteIP)})
				if relaErr != nil {
					log.Error("ActivityArchive relationGRPC.Relations mid:%d fid:%v error:%+v", mid, mids, relaErr)
					return nil
				}
				return relationResp.GetFollowingMap()
			}()
			for _, v := range reply.Staff {
				if v == nil {
					continue
				}
				if follow, ok := relations[v.Mid]; ok && follow != nil {
					v.Relation = &webgrpc.Relation{
						Attribute: int64(follow.Attribute),
						Tag:       follow.Tag,
						Special:   int64(follow.Special),
					}
				}
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		reply.ReqUser = func() *webgrpc.ReqUser {
			reqUser := s.arcRelation(ctx, mid, aid, 0, 0)
			return &webgrpc.ReqUser{
				Favorite: reqUser.Favorite,
				Like:     reqUser.Like,
				Dislike:  reqUser.Dislike,
				Multiply: reqUser.Coin,
			}
		}()
		return nil
	})
	if actSeason.IsContainedRecom { //如果包含AI的相关推荐
		eg.Go(func(ctx context.Context) error {
			arcs, _, _, err := s.RelatedArcs(ctx, aid, mid, buvid, false, false, false, true, nil)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			var rs []*webgrpc.Relate
			for _, val := range arcs {
				r := &webgrpc.Relate{
					Arc:        model.CopyFromBvArc(val),
					Bvid:       val.Bvid,
					SeasonType: val.SeasonType,
				}
				rs = append(rs, r)
			}
			if len(rs) == 0 {
				return nil
			}
			rightRelate := &webgrpc.OperationRelate{}
			if reply.RightRelate != nil {
				*rightRelate = *reply.RightRelate
			}
			rightRelate.AiRelateItem = rs
			reply.RightRelate = rightRelate
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		desc, descV2, mids, err := s.description(ctx, aid, reply.GetArc().GetDesc())
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		reply.Arc.Desc = desc
		var accReply *accmdl.InfosReply
		if len(mids) != 0 {
			accReply, err = s.accGRPC.Infos3(ctx, &accmdl.MidsReq{Mids: mids})
			if err != nil {
				log.Error("%+v", err)
			}
		}
		desc2 := s.DescV2ParamsMerge(ctx, descV2, accReply)
		var dV2 []*webgrpc.DescV2
		for _, val := range desc2 {
			dV2 = append(dV2, &webgrpc.DescV2{
				RawText: val.RawText,
				Type:    int64(val.Type),
				BizId:   val.BizId,
			})
		}
		reply.Arc.DescV2 = dV2
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("ActivityArchive mid:%d req:%+v error:%v", mid, req, err)
	}
	return reply, nil
}

func (s *Service) ActivityLiveTimeInfo(ctx context.Context, mid int64, req *webgrpc.ActivityLiveTimeInfoReq, buvid string) (*webgrpc.ActivityLiveTimeInfoReply, error) {
	actSeason, err := s.commonActivity(ctx, mid, req.ActivityKey, true, buvid)
	if err != nil {
		return nil, err
	}
	nowTime := time.Now().Unix()
	liveInfo := &webgrpc.ActivityLiveTimeInfoReply{
		NowTime:   time.Now().Unix(),
		StartTime: actSeason.Live.GetStartTime(),
		EndTime:   actSeason.Live.GetEndTime(),
	}
	if nowTime >= actSeason.Live.GetStartTime() {
		liveInfo.Timeline = actSeason.Live.GetTimeline()
	}
	return liveInfo, nil
}

func (s *Service) ClickActivitySeason(ctx context.Context, mid int64, req *webgrpc.ClickActivitySeasonReq, buvid string) error {
	cancel := req.Action == 1
	err := func() (err error) {
		switch req.OrderType {
		case webgrpc.OrderType_TypeOrderActivity:
			if cancel {
				_, err = s.actGRPC.RelationReserveCancel(ctx, &actgrpc.RelationReserveCancelReq{
					Id:       req.GetReserveParam().GetReserveId(),
					Mid:      mid,
					From:     req.GetReserveParam().GetFrom(),
					Typ:      req.GetReserveParam().GetType(),
					Oid:      strconv.FormatInt(req.GetReserveParam().GetOid(), 10),
					Ip:       metadata.String(ctx, metadata.RemoteIP),
					Platform: _platformActivity,
					Buvid:    buvid,
					Spmid:    req.Spmid,
				})
				return err
			}
			_, err = s.actGRPC.GRPCDoRelation(ctx, &actgrpc.GRPCDoRelationReq{
				Id:       req.GetReserveParam().GetReserveId(),
				Mid:      mid,
				From:     req.GetReserveParam().GetFrom(),
				Typ:      req.GetReserveParam().GetType(),
				Oid:      strconv.FormatInt(req.GetReserveParam().GetOid(), 10),
				Ip:       metadata.String(ctx, metadata.RemoteIP),
				Platform: _platformActivity,
				Buvid:    buvid,
				Spmid:    req.Spmid,
			})
		case webgrpc.OrderType_TypeFavSeason:
			if cancel {
				_, err = s.favGRPC.DelFav(ctx, &favgrpc.DelFavReq{
					Tp:       _favTypeArc,
					Mid:      mid,
					Oid:      req.GetFavParam().GetSeasonId(),
					Otype:    _favOTypeSeason,
					Platform: "pc",
				})
				return err
			}
			_, err = s.favGRPC.AddFav(ctx, &favgrpc.AddFavReq{
				Tp:       _favTypeArc,
				Mid:      mid,
				Oid:      req.GetFavParam().GetSeasonId(),
				Otype:    _favOTypeSeason,
				Platform: "pc",
			})
		default:
			err = ecode.RequestErr
		}
		return err
	}()
	if err != nil {
		log.Error("ClickActivitySeason mid:%d req:%+v error:%v", mid, req, err)
		return err
	}
	return nil
}

func checkActivitySeasonWhiteList(mid int64, whiteList []int64) error {
	if len(whiteList) == 0 {
		return nil
	}
	if mid <= 0 {
		return gateecode.WhiteListErr
	}
	for _, v := range whiteList {
		if mid == v {
			return nil
		}
	}
	return gateecode.WhiteListErr
}

func (s *Service) loadCommonActivities() error {
	ctx := context.Background()
	acts, err := s.comActiveGRPC.CommonActivities(ctx, &activegrpc.CommonActivitiesReq{})
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			s.activitySeasonKeyMem = make(map[string]*model.ActivitySeasonMem)
			s.activitySeasonIDMem = make(map[int64]*model.ActivitySeasonMem)
			return nil
		}
		log.Error("loadActivitySeason CommonActivities error:%v", err)
		return err
	}
	if len(acts.GetActivities()) == 0 {
		log.Error("大型活动告警 loadActivitySeason len(acts.Activities) 0")
		s.activitySeasonKeyMem = make(map[string]*model.ActivitySeasonMem)
		s.activitySeasonIDMem = make(map[int64]*model.ActivitySeasonMem)
		return nil
	}
	memKeyData := make(map[string]*model.ActivitySeasonMem, len(acts.GetActivities()))
	memIDData := make(map[int64]*model.ActivitySeasonMem, len(acts.GetActivities()))
	for id, v := range acts.GetActivities() {
		if !verifyAct(v) {
			log.Warn("loadActivitySeason id:%d nil or activityPlay nil or seasonID <= 0 or PcPlay nil data:%+v", id, v)
			continue
		}
		seasonData, err := s.activitySeasonData(ctx, v, 0, "")
		if err != nil {
			// 非404错误用内存中数据不变
			if !ecode.EqualError(ecode.NothingFound, err) {
				memKeyData[v.PcPlay.ActivityKey] = s.activitySeasonKeyMem[v.PcPlay.ActivityKey]
				memIDData[v.ActivePlay.SeasonId] = s.activitySeasonIDMem[v.ActivePlay.SeasonId]
				log.Error("大型活动告警 loadActivitySeason name:%s id:%d error:%+v", v.ActivePlay.ActivityName, id, err)
			}
			continue
		}
		memKeyData[v.PcPlay.ActivityKey] = seasonData
		memIDData[v.ActivePlay.SeasonId] = seasonData
	}
	s.activitySeasonKeyMem = memKeyData
	s.activitySeasonIDMem = memIDData
	return nil
}

// verifyAct 校验后台接口返回值是否符合预期
func verifyAct(act *activegrpc.CommonActivityResp) bool {
	if act == nil || act.ActivePlay == nil || act.ActivePlay.SeasonId <= 0 || act.ActivePlay.ActivityUrl == "" || act.PcPlay == nil || act.PcPlay.ActivityKey == "" {
		return false
	}
	return true
}

func (s *Service) commonActivity(ctx context.Context, mid int64, actKey string, forbidDegrade bool, buvid string) (*model.ActivitySeasonMem, error) {
	actSeason, ok := s.activitySeasonKeyMem[actKey]
	if !ok || actSeason == nil {
		if forbidDegrade {
			log.Warn("commonActivity memory actKey:%s season nil", actKey)
			return nil, ecode.NothingFound
		}
		// 内存无数据降级实时请求活动数据
		commonAct, err := s.comActiveGRPC.CommonActivity(ctx, &activegrpc.CommonActivityReq{
			Plat:   _commonActPlayPc,
			Schema: actKey,
			Mid:    mid,
		})
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) && !ecode.EqualError(gateecode.WhiteListErr, err) && !ecode.EqualError(gateecode.PlatActivityUnExistErr, err) {
				log.Error("大型活动告警 commonActivity comActiveGRPC.CommonActivity actKey:%s mid:%d error:%v", actKey, mid, err)
			}
			return nil, err
		}
		if !verifyAct(commonAct) {
			log.Error("大型活动告警 后台activegrpc.CommonActivity接口数据不合法 actKey:%s val:%+v", actKey, commonAct)
			return nil, ecode.NothingFound
		}
		actSeason, err = s.activitySeasonData(ctx, commonAct, mid, buvid)
		if err != nil {
			log.Error("大型活动告警 commonActivity activitySeasonData actKey:%s mid:%d error:%+v", actKey, mid, err)
			return nil, err
		}
	}
	if actSeason == nil {
		log.Warn("commonActivity actKey:%s mid:%d data nil", actKey, mid)
		return nil, ecode.NothingFound
	}
	if err := checkActivitySeasonWhiteList(mid, actSeason.Whitelist); err != nil {
		log.Warn("commonActivity checkActivitySeasonWhiteList actKey:%s mid:%d err:%+v", actKey, mid, err)
		return nil, err
	}
	return actSeason, nil
}

// nolint: gocognit
func (s *Service) activitySeasonData(ctx context.Context, act *activegrpc.CommonActivityResp, mid int64, buvid string) (*model.ActivitySeasonMem, error) {
	// 剧集
	season, err := s.ugcSeasonGRPC.View(ctx, &seasongrpc.ViewRequest{SeasonID: act.ActivePlay.SeasonId})
	if err != nil {
		return nil, errors.Wrapf(err, "activitySeasonData ugcSeasonGRPC.View seasonID:%d", act.ActivePlay.SeasonId)
	}
	if season.GetView() == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "activitySeasonData seasonID:%d season.GetView() == nil", act.ActivePlay.SeasonId)
	}
	var aids []int64
	existAid := map[int64]struct{}{}
	for _, sec := range season.View.Sections {
		if sec == nil {
			continue
		}
		for _, ep := range sec.Episodes {
			if ep == nil {
				continue
			}
			if _, ok := existAid[ep.Aid]; ok {
				continue
			}
			aids = append(aids, ep.Aid)
			existAid[ep.Aid] = struct{}{}
		}
	}
	if len(aids) == 0 {
		return nil, errors.Wrapf(ecode.NothingFound, "activitySeasonData seasonID:%d len(aids) == 0", act.ActivePlay.SeasonId)
	}
	// 稿件
	eg := errgroup.WithContext(ctx)
	var (
		views   map[int64]*arcgrpc.ViewReply
		arcDesc map[int64]string
		relatem map[int64][]*webgrpc.Relate
	)
	eg.Go(func(ctx context.Context) error {
		var viewsErr error
		views, viewsErr = s.batchViews(ctx, aids)
		if viewsErr != nil {
			return viewsErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		arcDesc = s.batchArcDesc(ctx, aids)
		return nil
	})
	if act.ActivePlay.IsContainedRecom { //如果包含AI的相关推荐
		eg.Go(func(ctx context.Context) error {
			relatem = s.batchArcRelate(ctx, aids, mid, buvid)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	afViews := make(map[int64]*arcgrpc.ViewReply)
	for _, aid := range aids {
		view, ok := views[aid]
		if !ok || view == nil || view.Arc == nil || !view.Arc.IsNormal() {
			log.Warn("activitySeasonData seasonID:%d aid:%d not allowed", act.ActivePlay.SeasonId, aid)
			continue
		}
		if desc, ok := arcDesc[aid]; ok && desc != "" {
			view.Arc.Desc = desc
		}
		afViews[aid] = view
	}
	if len(afViews) == 0 {
		return nil, errors.Wrapf(ecode.NothingFound, "activitySeasonData seasonID:%d len(afArcs) == 0", act.ActivePlay.SeasonId)
	}
	var mids []int64
	midMap := make(map[int64]struct{})
	for _, arc := range afViews {
		if arc.Author.Mid > 0 {
			if _, ok := midMap[arc.Author.Mid]; !ok {
				mids = append(mids, arc.Author.Mid)
				midMap[arc.Author.Mid] = struct{}{}
			}
		}
		if len(arc.StaffInfo) == 0 {
			continue
		}
		for _, v := range arc.StaffInfo {
			if v != nil && v.Mid > 0 {
				if _, ok := midMap[v.Mid]; !ok {
					mids = append(mids, v.Mid)
					midMap[v.Mid] = struct{}{}
				}
			}
		}
	}
	// 账号 staff
	var cards map[int64]*accgrpc.Card
	if len(mids) > 0 {
		cards = func() map[int64]*accgrpc.Card {
			accs, err := s.batchAccCards(ctx, mids)
			if err != nil {
				log.Error("activitySeasonData batchAccCards(%v) error(%v)", mids, err)
				return nil
			}
			return accs
		}()
	}
	actArcs := make(map[int64]*model.ActivityView, len(afViews))
	for _, afView := range afViews {
		actArc := &model.ActivityView{Arc: model.CopyFromArcToWebGRPC(afView.Arc), Pages: afView.Pages}
		if afView.Author.Mid > 0 {
			authorStaff := func() *webgrpc.Staff {
				if card, ok := cards[afView.Arc.Author.Mid]; ok && card != nil {
					return model.StaffInfoFromCard(card, &arcgrpc.StaffInfo{Mid: afView.Arc.Author.Mid, Title: "UP主"})
				}
				return &webgrpc.Staff{Mid: afView.Arc.Author.Mid, Title: "UP主", Name: afView.Arc.Author.Name, Face: afView.Arc.Author.Face}
			}()
			actArc.StaffInfo = append(actArc.StaffInfo, authorStaff)
		}
		rec := act.PcPlay.CommRecommBr[afView.Arc.Aid]
		relate := relatem[afView.Arc.Aid]
		actArc.RightRelate = model.OperationRelateFromCommRecommends(rec, s.c.ActivitySeason.RightRelatedTitle, relate)
		if beRec, ok := act.PcPlay.CommRecommBe[afView.Arc.Aid]; ok {
			actArc.BottomRelate = model.OperationRelateFromCommRecommends(&activegrpc.CommRecommends{CommRecommends: []*activegrpc.CommRecommend{beRec}}, "", nil)
		}
		// 非联合投稿不用拼其他作者信息
		if afView.Arc.AttrVal(arcgrpc.AttrBitIsCooperation) == arcgrpc.AttrNo {
			actArcs[afView.Arc.Aid] = actArc
			continue
		}
		for _, v := range afView.Arc.StaffInfo {
			if v != nil && v.Mid > 0 {
				if card, ok := cards[v.Mid]; ok && card != nil {
					actArc.StaffInfo = append(actArc.StaffInfo, model.StaffInfoFromCard(card, v))
				}
			}
		}
		actArcs[afView.Arc.Aid] = actArc
	}
	memData := &model.ActivitySeasonMem{
		Whitelist:        act.ActivePlay.Whitelist,
		Title:            act.ActivePlay.ActivityName,
		ActivityURL:      act.ActivePlay.ActivityUrl,
		IsContainedRecom: act.ActivePlay.IsContainedRecom,
	}
	memData.SeasonView = new(model.ActivitySeason)
	memData.SeasonView.CopyFromUgcSeason(season.View, actArcs)
	// 直播
	memData.Live = func() *webgrpc.ActivityLive {
		if !act.ActivePlay.IsContainedLive {
			return nil
		}
		if act.ActivePlay.RoomId <= 0 || act.ActivePlay.Stime <= 0 || act.ActivePlay.Etime <= 0 {
			log.Error("大型活动告警 activitySeasonData 直播数据有误 name:%s playData:%+v", act.ActivePlay.ActivityName, act.ActivePlay)
			return nil
		}
		liveData := &webgrpc.ActivityLive{
			RoomId:          act.ActivePlay.RoomId,
			StartTime:       act.ActivePlay.Stime,
			EndTime:         act.ActivePlay.Etime,
			HoverPic:        act.PcPlay.GuidePic,
			HoverJumpUrl:    act.PcPlay.GuideUrl,
			BreakCycle:      act.PcPlay.BreakCycle,
			OperationRelate: model.OperationRelateFromCommRecommends(&activegrpc.CommRecommends{CommRecommends: act.PcPlay.LiveRecommRight}, "", nil),
			ReplyType:       act.ActivePlay.ReplyType,
			ReplyId:         act.ActivePlay.ReplyId,
			HoverPicClose:   act.PcPlay.HoverPicClose,
			GiftDisclaimer:  act.PcPlay.GiftDisclaimer,
		}
		if s.c.ActivitySeason.BreakCycle > 0 && liveData.BreakCycle < s.c.ActivitySeason.BreakCycle {
			liveData.BreakCycle = s.c.ActivitySeason.BreakCycle
		}
		if act.PcPlay.NeedContent {
			for _, v := range act.PcPlay.PcShows {
				if v == nil {
					continue
				}
				liveData.Timeline = append(liveData.Timeline, &webgrpc.LiveTimeline{
					Name:      v.ShowName,
					StartTime: v.ShowSt,
					EndTime:   v.ShowEt,
					Cover:     v.ShowPic,
					H5Cover:   v.H5Cover,
				})
			}
		}
		return liveData
	}()
	// 直播游戏
	memData.Game = func() *webgrpc.ActivityGame {
		if memData.Live == nil {
			return nil
		}
		var iframes []*webgrpc.ActivityGameIframe
		for _, v := range act.PcPlay.LiveActivity {
			if v == nil {
				continue
			}
			iframes = append(iframes, &webgrpc.ActivityGameIframe{
				Url:    v.LiveActivityUrl,
				Height: v.LiveActivityHeight,
			})
		}
		if len(iframes) == 0 {
			return nil
		}
		return &webgrpc.ActivityGame{
			Iframes:       iframes,
			Disclaimer:    act.PcPlay.Disclaimer,
			DisclaimerUrl: act.PcPlay.DisclaimerUrl,
		}
	}()
	// 主题
	memData.Theme = model.CopyFromPcPlay(act.PcPlay)
	// 预约组件
	memData.Subscribe = act.PcPlay.Menus
	return memData, nil
}

func (s *Service) batchViews(ctx context.Context, aids []int64) (map[int64]*arcgrpc.ViewReply, error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	group := errgroup.WithContext(ctx)
	views := make(map[int64]*arcgrpc.ViewReply, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) error {
			arg := &arcgrpc.ViewsRequest{Aids: partAids}
			reply, arcErr := s.arcGRPC.Views(ctx, arg)
			if arcErr != nil {
				return arcErr
			}
			mutex.Lock()
			for _, v := range reply.GetViews() {
				views[v.Aid] = v
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return views, nil
}

func (s *Service) batchArcRelate(ctx context.Context, aids []int64, mid int64, buvid string) map[int64][]*webgrpc.Relate {
	group := errgroup.WithContext(ctx)
	relatem := make(map[int64][]*webgrpc.Relate, len(aids))
	mutex := sync.Mutex{}
	for _, aid := range aids {
		tmpAid := aid
		group.Go(func(ctx context.Context) error {
			reply, _, _, err := s.RelatedArcs(ctx, tmpAid, mid, buvid, false, false, false, true, nil)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			var rs []*webgrpc.Relate
			for _, val := range reply {
				r := &webgrpc.Relate{
					Arc:        model.CopyFromBvArc(val),
					Bvid:       val.Bvid,
					SeasonType: val.SeasonType,
				}
				rs = append(rs, r)
			}
			mutex.Lock()
			relatem[tmpAid] = rs
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return relatem
}

func (s *Service) batchArcDesc(ctx context.Context, aids []int64) map[int64]string {
	group := errgroup.WithContext(ctx)
	arcDesc := make(map[int64]string, len(aids))
	mutex := sync.Mutex{}
	for _, aid := range aids {
		descAid := aid
		group.Go(func(ctx context.Context) error {
			reply, descErr := s.arcGRPC.Description(ctx, &arcgrpc.DescriptionRequest{Aid: descAid})
			if descErr != nil {
				log.Error("batchArcDesc s.arcGRPC.Description aid:%d error:%+v", descAid, descErr)
				return nil
			}
			mutex.Lock()
			arcDesc[descAid] = reply.GetDesc()
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("batchArcDesc aids:%v error:%+v", aids, err)
	}
	return arcDesc
}

func (s *Service) batchAccCards(ctx context.Context, mids []int64) (map[int64]*accgrpc.Card, error) {
	var (
		mutex   = sync.Mutex{}
		midsLen = len(mids)
	)
	group := errgroup.WithContext(ctx)
	cards := make(map[int64]*accgrpc.Card, midsLen)
	for i := 0; i < midsLen; i += _midBulkSize {
		var partMids []int64
		if i+_aidBulkSize > midsLen {
			partMids = mids[i:]
		} else {
			partMids = mids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) error {
			arg := &accgrpc.MidsReq{Mids: partMids}
			reply, accErr := s.accGRPC.Cards3(ctx, arg)
			if accErr != nil {
				return accErr
			}
			mutex.Lock()
			for _, v := range reply.GetCards() {
				cards[v.Mid] = v
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return cards, nil
}

func aidFromAidAndBvid(aid int64, bvID string) int64 {
	resAid := aid
	if bvID != "" {
		bvAid, err := bvid.BvToAv(bvID)
		if err != nil {
			log.Warn("AidFromAidAndBvid bvID:%s error:%v", bvID, err)
			return resAid
		}
		resAid = bvAid
	}
	return resAid
}
