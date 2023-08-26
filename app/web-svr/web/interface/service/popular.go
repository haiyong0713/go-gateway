package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	errGroup "go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	arc "go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	accGrpc "git.bilibili.co/bapis/bapis-go/account/service"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	emotegrpc "git.bilibili.co/bapis/bapis-go/community/service/emote"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	actplatgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

func (s *Service) Popular(ctx context.Context, mid int64, buvid, path string, pn, ps int) (list []*model.PopularArc, noMore bool, err error) {
	aiRcmd, userFeature, resCode, degrade := func() ([]*model.HotItem, string, int, bool) {
		// 如果mid和buvid都是空，直接访问降级数据
		if mid == 0 && buvid == "" {
			log.Warn("Popular mid && buvid empty")
			return nil, "", 0, true
		}
		pn = pn - 1 // ai pn 从0开始
		aiRcmd, userFeature, resCode, aiErr := s.dao.HotAiRcmd(ctx, mid, buvid, pn, ps)
		if aiErr != nil {
			log.Error("日志报警 Popular HotAiRcmd mid:%d buvid:%s pn:%d ps:%d error(%v)", mid, buvid, pn, ps, aiErr)
			return nil, "", 0, true
		}
		return aiRcmd, userFeature, resCode, false
	}()
	if resCode == -3 {
		return []*model.PopularArc{}, true, nil
	}
	if degrade {
		aiRcmd = s.popularDegrade(ctx, 0, (pn-1)*ps, ps)
		if len(aiRcmd) == 0 {
			return []*model.PopularArc{}, true, nil
		}
	}
	var (
		aids         []int64
		firstTrackID string
	)
	for _, v := range aiRcmd {
		if v == nil || v.Goto != "av" || v.ID <= 0 {
			continue
		}
		if firstTrackID == "" && v.TrackID != "" {
			firstTrackID = v.TrackID
		}
		aids = append(aids, v.ID)
	}
	if len(aids) == 0 {
		log.Warn("Popular HotAiRcmd mid:%d buvid:%s pn:%d ps:%d no aids", mid, buvid, pn, ps)
		return []*model.PopularArc{}, false, nil
	}
	arcs, err := s.batchArchives(ctx, aids)
	if err != nil {
		log.Error("Popular HotAiRcmd mid:%d buvid:%s pn:%d ps:%d aids:%v error:%v", mid, buvid, pn, ps, aids, err)
		return []*model.PopularArc{}, false, nil
	}
	var feedInfoc []*model.PopularFeedInfoc
	var i int
	for _, v := range aiRcmd {
		if v == nil || v.Goto != "av" || v.ID <= 0 {
			continue
		}
		if arc, ok := arcs[v.ID]; ok && arc != nil && arc.IsNormal() {
			bvidStr := s.avToBv(arc.Aid)
			list = append(list, &model.PopularArc{
				BvArc:      model.CopyFromArcToBvArc(arc, bvidStr),
				RcmdReason: v.RcmdReason,
			})
			i++
			feedInfoc = append(feedInfoc, &model.PopularFeedInfoc{
				Goto:         "av",
				Param:        strconv.FormatInt(v.ID, 10),
				URI:          fmt.Sprintf("https://www.bilibili.com/video/%s", bvidStr),
				AvFeature:    v.AvFeature,
				Source:       v.Source,
				RPos:         i,
				FromType:     v.FromType,
				CornerMark:   v.RcmdReason.CornerMark,
				RcmdContent:  v.RcmdReason.Content,
				CoverType:    "pic",
				CardStyle:    2,
				HotAggreID:   v.HotwordID,
				ChannelOrder: 0,
				ChannelName:  "全部热门",
				ChannelID:    0,
			})
		}
	}
	s.InfocV2(model.PopularInfoc{
		MobiApp:    "pc",
		Time:       time.Now().Format("2006-01-02 15:04:05"),
		LoginEvent: 2, // web无该字段，按ai要求写死2
		Mid:        mid,
		Buvid:      buvid,
		Feed: func() string {
			res, _ := json.Marshal(feedInfoc)
			return string(res)
		}(),
		Page:    int64(pn),
		URL:     path,
		Env:     env.DeployEnv,
		Trackid: firstTrackID,
		IsRec: func() int64 {
			if !degrade {
				return 1
			}
			return 0
		}(),
		ReturnCode:  strconv.Itoa(resCode),
		UserFeature: userFeature,
		Flush:       "0", // 按ai要求写死0
	})
	return list, false, nil
}

func (s *Service) popularDegrade(ctx context.Context, i, index, ps int) []*model.HotItem {
	data, err := s.dao.PopularCardTenCache(ctx, i, index, ps)
	if err != nil {
		log.Error("popularDegrade error:%+v", err)
		return nil
	}
	var res []*model.HotItem
	for _, v := range data {
		if v == nil {
			continue
		}
		res = append(res, &model.HotItem{
			ID:         v.Value,
			Goto:       v.Type,
			RcmdReason: &model.HotRcmdReason{Content: v.Reason, CornerMark: v.CornerMark},
		})
	}
	return res
}

// nolint:gomnd
func (s *Service) PopularSeriesOne(ctx context.Context, typ string, number int64) (*model.SeriesOne, error) {
	data, err := s.dao.CacheWeeklySeries(ctx, typ)
	if err != nil {
		log.Error("%+v", err)
		data = s.seriesConfigBak
	}
	var config *model.SeriesConfig
	for _, val := range data {
		if val == nil {
			continue
		}
		if val.Number == number {
			config = val
			break
		}
	}
	if config == nil {
		return nil, errors.WithStack(ecode.NothingFound)
	}
	if config.MediaID > 0 {
		config.MediaID = config.MediaID*100 + s.c.PopularSeries.MediaMid%100
	}
	detail, err := s.dao.CacheSeriesDetail(ctx, config.ID)
	if err != nil {
		log.Error("%+v", err)
		detail = s.seriesListBak
	}
	list, ok := detail[config.ID]
	if !ok {
		return nil, errors.WithStack(ecode.NothingFound)
	}
	var aids []int64
	for _, val := range list {
		if val == nil || val.Goto != "av" || val.Param <= 0 {
			continue
		}
		aids = append(aids, val.Param)
	}
	if len(aids) == 0 {
		return nil, errors.WithStack(ecode.NothingFound)
	}
	arcs, err := s.batchArchives(ctx, aids)
	if err != nil {
		return nil, err
	}
	var oneList []*model.SeriesArc
	for _, val := range list {
		if val == nil || val.Goto != "av" || val.Param <= 0 {
			continue
		}
		if arc, ok := arcs[val.Param]; ok && arc.IsNormal() {
			oneList = append(oneList, &model.SeriesArc{
				BvArc:      model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)),
				RcmdReason: val.RcmdReason,
			})
		}
	}
	oneConfig := &model.SeriesOneConfig{
		ID:            config.ID,
		Type:          config.Type,
		Number:        config.Number,
		Subject:       config.Subject,
		Stime:         config.Stime,
		Etime:         config.Etime,
		Status:        config.Status,
		Name:          config.Name,
		Hint:          config.Hint,
		Color:         config.Color,
		Cover:         config.Cover,
		ShareTitle:    config.ShareTitle,
		ShareSubtitle: config.ShareSubtitle,
		MediaID:       config.MediaID,
	}
	// 第8期(0102更新)，用于左上角展示
	oneConfig.Label = fmt.Sprintf("第%d期(%s更新)", oneConfig.Number, oneConfig.Etime.Time().AddDate(0, 0, 1).Format("0102"))
	if oneConfig.Status == 4 { // 灾备数据
		oneConfig.Subject = "哔哩哔哩每周必看"
		// 白色
		oneConfig.Color = 2
		// 2019年第8期
		oneConfig.Hint = oneConfig.Stime.Time().Format("2006") + fmt.Sprintf("年第%d期:", oneConfig.Number)
		// 默认头图
		oneConfig.Cover = "http://i0.hdslb.com/bfs/archive/1eed634b8071d0d37dfdb6d68513242ae9e0897b.jpg"
		// [哔哩哔哩每周必看]2019年第3期
		oneConfig.ShareTitle = "「哔哩哔哩每周必看」" + oneConfig.Stime.Time().Format("2006") + fmt.Sprintf("年第%d期", oneConfig.Number)
	}
	return &model.SeriesOne{
		Config:   oneConfig,
		Reminder: s.c.PopularSeries.Reminder,
		List:     oneList,
	}, nil
}

func (s *Service) PopularSeries(ctx context.Context, typ string) ([]*model.SeriesRes, error) {
	data, err := s.dao.CacheWeeklySeries(ctx, typ)
	if err != nil {
		log.Error("%+v", err)
		data = s.seriesConfigBak
	}
	var res []*model.SeriesRes
	for _, v := range data {
		if v == nil {
			continue
		}
		res = append(res, &model.SeriesRes{Number: v.Number, Subject: v.Subject, Status: v.Status, Name: v.Name})
	}
	if len(res) == 0 {
		return nil, errors.WithStack(ecode.NothingFound)
	}
	return res, nil
}

func (s *Service) PopularPrecious(ctx context.Context) (*model.PreciousRes, error) {
	arcsReply, err := s.popularGRPC.Arcs(ctx, &populargrpc.Empty{})
	if err != nil {
		log.Error("日志告警 入站必刷获取后台数据错误: %+v", err)
		return nil, err
	}
	list, err := s.dealCard(ctx, arcsReply.GetList())
	if err != nil {
		log.Error("[Precious] s.dealCard() error(%+v)", err)
		return nil, err
	}
	if len(list) == 0 {
		return nil, ecode.NothingFound
	}
	explain := fmt.Sprintf(s.c.PopularPrecious.ExplainFmt, len(list))
	return &model.PreciousRes{
		Title:   s.c.PopularPrecious.Title,
		MediaID: arcsReply.GetMediaId(),
		Explain: explain,
		List:    list,
	}, nil
}

// dealCard .
func (s *Service) dealCard(ctx context.Context, arcList []*populargrpc.ArcList) ([]*model.PreciousArc, error) {
	var (
		aids     []int64
		arcReply map[int64]*arcgrpc.Arc
	)
	for _, v := range arcList {
		aids = append(aids, v.Aid)
	}
	latest, origin := splitLatestPrecious(arcList)
	if len(aids) != 0 {
		var err error
		if arcReply, err = s.batchArchives(ctx, aids); err != nil {
			log.Error("[dealCard] s.arc.ArchivesPB() aids(%s) error(%v)", xstr.JoinInts(aids), err)
			return nil, err
		}
	}
	var list []*model.PreciousArc
	for _, v := range latest {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if arc, ok := arcReply[v.Aid]; ok && arc != nil && arc.IsNormal() {
			list = append(list, &model.PreciousArc{
				BvArc:       model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)),
				Achievement: v.Recommend,
			})
		}
	}
	for _, v := range origin {
		if v == nil || v.Aid <= 0 {
			continue
		}
		if arc, ok := arcReply[v.Aid]; ok && arc != nil && arc.IsNormal() {
			list = append(list, &model.PreciousArc{
				BvArc:       model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)),
				Achievement: v.Recommend,
			})
		}
	}
	return list, nil
}

func splitLatestPrecious(in []*populargrpc.ArcList) ([]*populargrpc.ArcList, []*populargrpc.ArcList) {
	const (
		_TypeOrigin = int64(2)
		_TypeLatest = int64(1)
	)
	latest, origin := []*populargrpc.ArcList{}, []*populargrpc.ArcList{}
	for _, v := range in {
		switch v.Type {
		case _TypeOrigin:
			origin = append(origin, v)
			continue
		case _TypeLatest:
			latest = append(latest, v)
			continue
		default:
			log.Warn("unrecognized arc type: %+v", v)
			continue
		}
	}
	return latest, origin
}

func (s *Service) loadPopularSeries() {
	ctx := context.Background()
	func() {
		seriesConfig, err := s.dao.CacheWeeklySeries(ctx, "weekly_selected")
		if err != nil {
			log.Error("日志告警 构建每周必看灾备错误,err:%+v", err)
			return
		}
		if len(seriesConfig) == 0 {
			log.Error("日志告警 构建每周必看灾备数据为空,data:%+v", seriesConfig)
			return
		}
		var seriesIDs []int64
		for _, val := range seriesConfig {
			seriesIDs = append(seriesIDs, val.ID)
		}
		seriesList, err := s.dao.CacheSeriesDetail(ctx, seriesIDs...)
		if err != nil {
			log.Error("日志告警 构建每周必看灾备错误,err:%+v", err)
			return
		}
		if len(seriesList) == 0 {
			log.Error("日志告警 构建每周必看灾备数据为空,data:%+v", seriesList)
			return
		}
		s.seriesConfigBak = seriesConfig
		s.seriesListBak = seriesList
	}()
}

const (
	_step1           = 1 // 萌新发现官
	_step2           = 2 // 宝藏探索家
	_step3           = 3 // 识宝达人
	_step4           = 4 // 阅宝无数
	_step1Num        = 9
	_step2Num        = 21
	_step3Num        = 42
	_step4Num        = 85
	_popularBusiness = "popular_activity"
)

func (s *Service) PopularActivity(ctx context.Context, mid int64) (*model.PopularActivityReply, error) {
	role, total, err := s.activityProgress(ctx, mid)
	if err != nil {
		return nil, err
	}
	out := &model.PopularActivityReply{
		Role: role,
	}
	fanoutResult, err := s.doPopularFanout(ctx, mid, role)
	if err != nil {
		return nil, err
	}
	out.ActAwardStatus = fanoutResult.actAwardStatus
	out.Rank = fanoutResult.rank
	out.HonorMeta = &model.HonorMeta{
		ViewCount:           total,
		ThumbupCount:        fanoutResult.like.GetTotal(),
		CoinCount:           fanoutResult.coin.GetTotal(),
		FirstWatchTime:      fanoutResult.firstWatchTime,
		CurrentHonorGetTime: fanoutResult.honorGetTime,
		FirstLoginTime:      fanoutResult.account.GetProfile().GetJoinTime(),
	}
	return out, nil
}

func (s *Service) activityProgress(ctx context.Context, mid int64) (int8, int64, error) {
	reply, err := s.actPlatGRPC.GetTotalRes(ctx, &actplatgrpc.GetTotalResReq{
		Counter:  s.c.PopularActivity.ArchiveCounter,
		Activity: strconv.FormatInt(s.c.PopularActivity.Sid, 10),
		Mid:      mid,
	})
	if err != nil {
		return 0, 0, err
	}
	return handleRole(reply.GetTotal()), reply.GetTotal(), nil
}

func handleRole(total int64) int8 {
	if total < _step1Num {
		return 0
	}
	if total < _step2Num {
		return _step1
	}
	if total < _step3Num {
		return _step2
	}
	if total < _step4Num {
		return _step3
	}
	return _step4
}

type popularFanoutResult struct {
	account        *accGrpc.ProfileReply
	coin           *actplatgrpc.GetTotalResResp
	like           *actplatgrpc.GetTotalResResp
	actAwardStatus []*model.ActAwardStatus
	rank           int64
	firstWatchTime xtime.Time
	honorGetTime   xtime.Time
}

func constructAwardStatus(awardName string, resultId int64, hasStock bool) *model.ActAwardStatus {
	const (
		_noReward      = 1
		_alreadyReward = 2
		_noStock       = 3
	)
	status := &model.ActAwardStatus{
		AwardName: awardName,
	}
	if resultId > 0 {
		status.State = _alreadyReward
		return status
	}
	if hasStock {
		status.State = _noReward
		return status
	}
	status.State = _noStock
	return status
}

func constructInnerAwardStatus(role int8, rawStatus map[string]int64) []*model.ActAwardStatus {
	var res []*model.ActAwardStatus
	for key, val := range rawStatus {
		var status *model.ActAwardStatus
		switch key {
		case model.AwardStep1:
			if role >= _step1 {
				status = constructAwardStatus(model.AwardStep1, val, true)
			}
		case model.AwardStep2:
			if role >= _step2 {
				status = constructAwardStatus(model.AwardStep2, val, true)
			}
		case model.AwardStep3:
			if role >= _step3 {
				status = constructAwardStatus(model.AwardStep3, val, true)
			}
		case model.AwardStep4:
			if role >= _step4 {
				status = constructAwardStatus(model.AwardStep4, val, true)
			}
		default:
		}
		if status != nil {
			res = append(res, status)
		}
	}
	return res
}

func (s *Service) doPopularFanout(ctx context.Context, mid int64, role int8) (*popularFanoutResult, error) {
	const (
		_badgeStockNumber = 500
	)
	var (
		innerAwardStatus []*model.ActAwardStatus
		badgeAwardStatus *model.ActAwardStatus
	)
	out := &popularFanoutResult{}
	eg := errGroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		res, err := s.dao.PopularAward(ctx, mid)
		if err != nil {
			log.Error("doPopularFanout() s.dao.PopularAward err: %+v", err)
			return nil
		}
		innerAwardStatus = constructInnerAwardStatus(role, res)
		return nil
	})
	if role == _step4 {
		eg.Go(func(ctx context.Context) (err error) {
			var hasBadgeStock bool
			res, err := s.dao.PopularBadgeAward(ctx, mid)
			if err != nil {
				log.Error("doPopularFanout() s.dao.PopularBadgeAward err: %+v", err)
				return nil
			}
			if s.c.PopularActivity.HasBadgeStock {
				cnt, err := s.dao.CountPopularBadgeAward(ctx)
				if err != nil {
					log.Error("doPopularFanout() s.dao.CountPopularBadgeAward err:%+v", err)
					return nil
				}
				hasBadgeStock = cnt < _badgeStockNumber
			}
			badgeAwardStatus = constructAwardStatus(model.AwardBadge, res, hasBadgeStock)
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			out.honorGetTime, err = s.dao.PopularWatchTime(ctx, mid, int64(role))
			if err != nil {
				log.Error("Failed to raw PopularWatchTime: %+v", err)
			}
			if out.rank, err = s.dao.PopularRank(ctx, mid); err != nil {
				log.Error("Failed to run raw PopularRank: %+v", err)
				return nil
			}
			if out.rank != 0 {
				return nil
			}
			// fixme 补偿逻辑，正式上线后开启
			//if out.rank, err = s.dao.AddPopularRank(ctx, mid); err != nil {
			//	log.Error("Failed to AddPopularRank: %+v", err)
			//	return nil
			//}
			//s.cache.Do(ctx, func(ctx context.Context) {
			//	if err := s.dao.DelCachePopularRank(ctx, mid); err != nil {
			//		log.Error("Failed to DelCachePopularRank: %+v", err)
			//	}
			//})
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		out.firstWatchTime, err = s.dao.PopularWatchTime(ctx, mid, 0)
		if err != nil {
			log.Error("Failed to raw PopularWatchTime: %+v", err)
		}
		return nil
	})
	if role == _step1 {
		eg.Go(func(ctx context.Context) (err error) {
			out.honorGetTime, err = s.dao.PopularWatchTime(ctx, mid, int64(role))
			if err != nil {
				log.Error("Failed to raw PopularWatchTime: %+v", err)
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		out.account, err = s.accGRPC.Profile3(ctx, &accGrpc.MidReq{Mid: mid, RealIp: metadata.RemoteIP})
		if err != nil {
			log.Error("Failed to request Profile3: %+v", err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		out.coin, err = s.popularGetTotalCoinRes(ctx, mid)
		if err != nil {
			log.Error("Failed to request GetTotalRes coin: %+v", err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		out.like, err = s.popularGetTotalLikeRes(ctx, mid)
		if err != nil {
			log.Error("Failed to request GetTotalRes like: %+v", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	out.actAwardStatus = innerAwardStatus
	if badgeAwardStatus != nil {
		out.actAwardStatus = append(out.actAwardStatus, badgeAwardStatus)
	}
	return out, nil
}

type HistoryContent struct {
	Aid int64 `json:"aid"`
}

func (s *Service) PopularActivityArchiveList(ctx context.Context, mid int64) (*model.PopularActivityArchiveList, error) {
	reply, err := s.actPlatGRPC.GetHistory(ctx, &actplatgrpc.GetHistoryReq{
		Activity: strconv.FormatInt(s.c.PopularActivity.Sid, 10),
		Counter:  s.c.PopularActivity.ArchiveCounter,
		Mid:      mid,
	})
	if err != nil {
		log.Error("Failed to GetHistory: %+v", err)
		return nil, err
	}
	aids := make([]int64, 0, len(reply.GetHistory()))
	for _, history := range reply.GetHistory() {
		hc := &HistoryContent{}
		if err := json.Unmarshal([]byte(history.Source), hc); err != nil {
			log.Error("Failed to unmarshal history source: %+v", errors.WithStack(err))
			continue
		}
		if hc.Aid <= 0 {
			continue
		}
		aids = append(aids, hc.Aid)
	}
	out := &model.PopularActivityArchiveList{}
	if len(aids) == 0 {
		return out, nil
	}
	arcsReply, err := s.multiArcs(ctx, aids, mid)
	if err != nil {
		log.Error("Failed to request multiArcs: %+v", err)
		return nil, err
	}
	list := make([]*model.ArchiveMeta, 0, len(aids))
	for _, aid := range aids {
		item, ok := arcsReply[aid]
		if !ok || item == nil || !item.IsNormal() {
			log.Error("Failed to add aid from arcs: %d", aid)
			continue
		}
		list = append(list, &model.ArchiveMeta{
			Aid:        aid,
			Pic:        item.Pic,
			Title:      item.Title,
			AuthorName: item.Author.Name,
			View:       item.Stat.View,
			Danmaku:    item.Stat.Danmaku,
		})
	}
	out.List = list
	return out, nil
}

func (s *Service) multiArcs(ctx context.Context, aids []int64, mid int64) (map[int64]*arc.Arc, error) {
	const _count = 100
	var shard int
	if len(aids) < _count {
		shard = 1
	} else {
		shard = len(aids) / _count
		if len(aids)%(shard*_count) != 0 {
			shard++
		}
	}
	aidss := make([][]int64, shard)
	for i, aid := range aids {
		aidss[i%shard] = append(aidss[i%shard], aid)
	}
	arcms := make([]map[int64]*arc.Arc, len(aidss))
	g := errGroup.WithContext(ctx)
	for idx, aids := range aidss {
		if len(aids) == 0 {
			continue
		}
		tmpIdx, tmpAids := idx, aids
		g.Go(func(ctx context.Context) error {
			arcs, err := s.arcGRPC.Arcs(ctx, &arc.ArcsRequest{Aids: tmpAids, Mid: mid})
			if err != nil {
				return err
			}
			arcms[tmpIdx] = arcs.GetArcs()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	res := map[int64]*arc.Arc{}
	for _, arcm := range arcms {
		for aid, arc := range arcm {
			res[aid] = arc
		}
	}
	return res, nil
}

func (s *Service) sendAward(ctx context.Context, mid int64, awardName string) error {
	switch awardName {
	case model.AwardStep1:
		args := &emotegrpc.UserEmoteUnlockReq{
			Mid:      mid,
			EmoteIds: s.c.PopularActivity.SingleEmoteId,
			Business: _popularBusiness,
		}
		if _, err := s.dao.UserEmoteUnlock(ctx, args); err != nil {
			return errors.Wrapf(err, "sendAward() s.dao.UserEmoteUnlock args: %+v", args)
		}
	case model.AwardStep2:
		args := &emotegrpc.UserEmoteUnlockReq{
			Mid:      mid,
			EmoteIds: s.c.PopularActivity.AllEmoteId,
			Business: _popularBusiness,
		}
		if _, err := s.dao.UserEmoteUnlock(ctx, args); err != nil {
			return errors.Wrapf(err, "sendAward() s.dao.UserEmoteUnlock args: %+v", args)
		}
	case model.AwardStep3:
		args := constructRewardsSendAwardReq(mid, awardName)
		args.AwardId = s.c.PopularActivity.ShortTermSkinAwardId
		if _, err := s.dao.RewardsSendAwardV2(ctx, args); err != nil {
			return errors.Wrapf(err, "sendAward() s.dao.RewardsSendAwardV2 args: %+v", args)
		}
	case model.AwardStep4:
		args := constructRewardsSendAwardReq(mid, awardName)
		args.AwardId = s.c.PopularActivity.LongTermSkinAwardId
		if _, err := s.dao.RewardsSendAwardV2(ctx, args); err != nil {
			return errors.Wrapf(err, "sendAward() s.dao.RewardsSendAwardV2 args: %+v", args)
		}
	case model.AwardBadge:
		args := constructRewardsSendAwardReq(mid, awardName)
		args.AwardId = s.c.PopularActivity.BadgeAwardId
		if _, err := s.dao.RewardsSendAwardV2(ctx, args); err != nil {
			return errors.Wrapf(err, "sendAward() s.dao.RewardsSendAwardV2 args: %+v", args)
		}
	default:
		return errors.New(fmt.Sprintf("Invalid awardName %s", awardName))
	}
	return nil
}

func constructRewardsSendAwardReq(mid int64, awardName string) *actgrpc.RewardsSendAwardV2Req {
	return &actgrpc.RewardsSendAwardV2Req{
		Mid:      mid,
		UniqueId: fmt.Sprintf("%d--mid,%s--awardName", mid, awardName),
		Business: _popularBusiness,
	}
}

func (s *Service) PopularActivityAward(ctx context.Context, mid int64, awardName string) (*model.PopularAwardReply, error) {
	role, _, err := s.activityProgress(ctx, mid)
	if err != nil {
		return nil, err
	}
	// 检查是否满足领取条件
	if checkCannotAchieveAward(role, awardName) {
		return nil, ecode.Error(ecode.RequestErr, "未满足领取条件")
	}
	// 领实物勋章走实物领奖逻辑
	if awardName == model.AwardBadge {
		return s.PopularKillBadgeAward(ctx, mid)
	}
	// 其余走普通领奖逻辑
	res, err := s.dao.PopularAward(ctx, mid)
	if err != nil {
		return nil, err
	}
	// 判断是否已经领取，重复领取返回为true
	if checkAlreadyGetAward(res, awardName) {
		return nil, ecode.Error(ecode.RequestErr, "已领取奖励，请勿重复领取")
	}
	token := strconv.FormatInt(mid, 10) + time.Now().Format("20060102")
	out := &model.PopularAwardReply{}
	// 具体领奖逻辑
	if err = s.sendAward(ctx, mid, awardName); err != nil {
		log.Errorc(ctx, "PopularActivityAward() Failed to sendAward: err: %+v", err)
		return nil, err
	}
	// 如果领奖了，写进库里一条数据
	if out.ID, err = s.dao.AddPopularAward(ctx, mid, awardName, token); err != nil {
		log.Errorc(ctx, "PopularActivityAward() s.dao.AddPopularAward err: %+v", err)
		return nil, err
	}
	// 删缓存，没成功的话不返回err
	if err = s.dao.DelCachePopularAward(ctx, mid); err != nil {
		log.Errorc(ctx, "PopularActivityAward() Failed to DelCachePopularAward: err %+v", err)
		return out, nil
	}
	return out, nil
}

func checkCannotAchieveAward(role int8, awardName string) bool {
	switch awardName {
	case model.AwardStep1:
		return role < _step1
	case model.AwardStep2:
		return role < _step2
	case model.AwardStep3:
		return role < _step3
	case model.AwardStep4:
		return role < _step4
	case model.AwardBadge:
		return role < _step4
	default:
		log.Error("checkCannotAchieveAward() Invalid awardName %s", awardName)
		return true
	}
}

func checkAlreadyGetAward(awardStatus map[string]int64, awardName string) bool {
	val, ok := awardStatus[awardName]
	if !ok {
		return false
	}
	return val > 0
}

func (s *Service) PopularKillBadgeAward(ctx context.Context, mid int64) (*model.PopularAwardReply, error) {
	if !s.c.PopularActivity.HasBadgeStock {
		return nil, ecode.Error(ecode.RequestErr, "限量实物徽章已抢完，敬请期待下次活动")
	}
	// 取实物徽章领奖记录
	id, err := s.dao.PopularBadgeAward(ctx, mid)
	if err != nil {
		return nil, err
	}
	if id > 0 {
		return nil, ecode.Error(ecode.RequestErr, "已领取奖励，请勿重复领取")
	}
	// 剩下的是可以领奖的mid
	out := &model.PopularAwardReply{}
	if out.ID, err = s.dao.UpdatePopularBadgeAward(ctx, mid); err != nil {
		log.Error("PopularKillBadgeAward() s.dao.UpdatePopularBadgeAward err:%+v", err)
		return nil, err
	}
	// 通过检验且已经存表，调用领奖接口逻辑
	if err = s.sendAward(ctx, mid, model.AwardBadge); err != nil {
		log.Errorc(ctx, "Failed to sendAward: err: %+v", err)
		return nil, err
	}
	// 删除缓存中的脏数据,如果发生err，直接日志处理
	if err = s.dao.DelCachePopularBadgeAward(ctx, mid); err != nil {
		log.Error("Failed to DelCachePopularAward: %+v", err)
		return out, nil
	}
	return out, nil
}
