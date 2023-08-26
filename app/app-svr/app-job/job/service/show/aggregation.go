package show

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/errgroup"
	"go-common/library/xstr"

	showmdl "go-gateway/app/app-svr/app-job/job/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

func (s *Service) aggregationConsumeproc() {
	defer s.waiter.Done()
	var (
		msg *databus.Message
		ok  bool
		err error
	)
	msgs := s.aggregationSub.Messages()
	for {
		if msg, ok = <-msgs; !ok {
			close(s.aggregationChan)
			log.Info("aggregation databus consumer exit")
			return
		}
		_ = msg.Commit()
		log.Info("[databus: app-job aggregationSub] new message: %s", msg.Value)
		var aggmsg = &showmdl.AggregationMsg{}
		if err = json.Unmarshal(msg.Value, aggmsg); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		s.aggregationChan <- aggmsg
	}
}

func (s *Service) aggregationproc() {
	defer s.waiter.Done()
	var (
		ms *showmdl.AggregationMsg
		ok bool
	)
	for {
		time.Sleep(10 * time.Millisecond)
		if ms, ok = <-s.aggregationChan; !ok {
			log.Error("s.aggregationChan id closed")
			return
		}
		s.aggregationCache(ms)
	}
}

// 获取全网热点数据
// nolint: gocognit
func (s *Service) aggregationCache(ms *showmdl.AggregationMsg) {
	var (
		olds, news map[string][]*showmdl.AggregationItem
		aggPass    []*showmdl.Aggregation // 通过审核的热词(db)
		offLineIDs []int64                // 需要置为下线的热词ID
		err        error
	)
	if ms == nil {
		log.Warn("aggregationCache get ms nil")
		return
	}
	log.Info("aggregationCache get spider type: %s", ms.SpiderType)
	// 热点类型 SpiderType: rank_data、account_data、archive_data
	if ms.SpiderType != "" {
		// 缓存中对应SpiderType的热点数据
		if olds, err = s.dao.Aggregations(context.Background(), ms.SpiderType); err != nil {
			log.Error("%v", err)
			return
		}
		// db中的过审热词
		if aggPass, err = s.dao.AggregationPass(context.Background()); err != nil {
			log.Error("%v", err)
			return
		}
		news = make(map[string][]*showmdl.AggregationItem)
		// 对新热点数据做整理
		for nidx, aggTmp := range ms.RankData {
			// 热点数据和对应platform: weibo、douyu、acfun、zhihu
			if aggTmp == nil || aggTmp.Platform == "" {
				continue
			}
			// 不处理 acfunc非热搜榜、weibo非热搜榜-实时热点
			// acfun -（榜单类型: 1 香蕉榜 2 热搜榜）
			// weibo - (榜单类型: 1 热搜榜-实时热点 2 热搜榜-实时上升热点 3 话题榜
			if (aggTmp.Platform == "acfun" && aggTmp.DataType != 2) || (aggTmp.Platform == "weibo" && aggTmp.DataType != 1) {
				continue
			}
			agg := &showmdl.AggregationItem{}
			*agg = *aggTmp
			agg.DatabusTime = ms.Timestamp
			// 计算排名变化
			if oaggs, ok := olds[aggTmp.Platform]; !ok {
				// 新平台,全部置为新
				agg.RankType = showmdl.AggregationNew
			} else {
				// 已有平台，计算 新、上升、下降及具体值
				var isExit bool
				for oidx, oagg := range oaggs {
					if aggTmp.Title != oagg.Title {
						continue
					}
					isExit = true
					if nidx < oidx {
						// 排名上升
						agg.RankType = showmdl.AggregationUp
						agg.RankValue = oidx - nidx
					} else if nidx > oidx {
						// 排名下降
						agg.RankType = showmdl.AggregationDown
						agg.RankValue = nidx - oidx
					}
				}
				if !isExit {
					agg.RankType = showmdl.AggregationNew
				}
			}
			news[agg.Platform] = append(news[agg.Platform], agg)
		}
		if olds == nil {
			olds = make(map[string][]*showmdl.AggregationItem)
		}
		for plat, new := range news {
			olds[plat] = new
		}
		// 覆盖更新后的热点缓存
		if err = s.dao.SetAggregations(context.Background(), ms.SpiderType, olds); err != nil {
			log.Error("%v", err)
			return
		}
		// 聚合需要下线的热点(非B站热门、非人工)
		for _, ap := range aggPass {
			if ap == nil || ap.Plat == "bili_popular" || ap.Plat == "artificial" {
				continue
			}
			old, ok := olds[ap.Plat]
			if !ok {
				offLineIDs = append(offLineIDs, ap.ID)
				continue
			}
			var isExist bool
			for _, o := range old {
				if o.Title == ap.HotTitle {
					isExist = true
				}
			}
			if !isExist {
				offLineIDs = append(offLineIDs, ap.ID)
			}
		}
		// 下线热词
		if len(offLineIDs) > 0 {
			if _, err = s.dao.OffLine(context.Background(), offLineIDs); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
}

// 刷新热点物料
// nolint: gocognit,gomnd
func (s *Service) aggregationMaterial() {
	var (
		aggPass []*showmdl.Aggregation
		err     error
	)
	// 获取db所有通过审核的热词
	if aggPass, err = s.dao.AggregationPass(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	// 缓存热点对应的ai数据和热点下所有稿件的信息
	for _, _ap := range aggPass {
		ap := _ap
		if ap == nil || ap.ID == 0 {
			continue
		}
		// 获取新鲜的ai物料
		var ais []*showmdl.CardList
		g, ctx := errgroup.WithContext(context.Background())
		g.Go(func() (err error) {
			if ais, err = s.dao.AggFromAI(ctx, ap.ID); err != nil {
				log.Error("%v", err)
				err = nil
				return
			}
			if len(ais) > 100 {
				ais = ais[:100]
			}
			return
		})
		// 获取人工操作过的物料(添加和禁止)
		var material []int64
		g.Go(func() (err error) {
			if material, err = s.dao.Materials(ctx, ap.ID); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
		// 缓存中的ai物料
		var oldAI *showmdl.AggAI
		g.Go(func() (err error) {
			if oldAI, err = s.dao.AggAI(ctx, ap.ID); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			continue
		}
		var (
			aidsTmp = make(map[int64]int64) // 聚合aid并做排重所用的临时aid map
			tidsTmp = make(map[int64]int64) // 聚合tid并做排重所用的临时tid map
			aim     = make(map[int64]*showmdl.CardList)
		)
		// 从新鲜ai物料中聚合aid和物料map
		for _, v := range ais {
			if v == nil {
				continue
			}
			if v.Tag != "" {
				tagIDs, _ := xstr.SplitInts(v.Tag)
				for _, v := range tagIDs {
					tidsTmp[v] = v
				}
			}
			if v.ID != 0 {
				aidsTmp[v.ID] = v.ID
				aim[v.ID] = v
			}
		}
		// 从人工物料中聚合aid
		for _, m := range material {
			if m == 0 {
				continue
			}
			aidsTmp[m] = m
		}
		// 排重聚合aid和tagid
		var aids, tids []int64
		for aid := range aidsTmp {
			aids = append(aids, aid)
		}
		for tid := range tidsTmp {
			tids = append(tids, tid)
		}
		var (
			arcm   map[int64]*arcgrpc.Arc
			oldArc map[int64]*showmdl.ArcInfo
			tagm   map[int64]*taggrpc.Tag
		)
		g2, ctx2 := errgroup.WithContext(context.Background())
		if len(tids) > 0 {
			g2.Go(func() (err error) {
				if tagm, err = s.dao.Tags(ctx2, tids); err != nil {
					log.Error("%v", err)
				}
				return
			})
		}
		if len(aids) > 0 {
			g2.Go(func() (err error) {
				if arcm, err = s.arcDao.Arcs(ctx2, aids); err != nil {
					log.Error("%v", err)
				}
				return
			})
			// 获取缓存中热词物料列表
			g2.Go(func() (err error) {
				if oldArc, err = s.dao.AggArc(ctx2, ap.ID); err != nil {
					log.Error("%v", err)
				}
				return
			})
		}
		if err = g2.Wait(); err != nil {
			log.Error("%+v", err)
			continue
		}
		// 缓存ai物料
		var newAI = &showmdl.AggAI{}
		// 计算ai物料更新数量
		newAI.UpCnt = len(aim)
		if oldAI != nil {
			newAI.UpCnt = len(aim) - len(oldAI.CardList)
		}
		newAI.CardList = make(map[int64]*showmdl.CardList)
		for _, v := range aim {
			var (
				tagNames []string
				ai       = &showmdl.CardList{}
			)
			if v.Tag != "" {
				tagIDs, _ := xstr.SplitInts(v.Tag)
				for _, v := range tagIDs {
					if tag, ok := tagm[v]; ok && tag != nil {
						tagNames = append(tagNames, tag.Name)
					}
				}
			}
			*ai = *v
			ai.TagNames = strings.Join(tagNames, ",")
			newAI.CardList[v.ID] = ai
		}
		// 刷新AI物料缓存
		if err = s.dao.SetAggAI(context.Background(), ap.ID, newAI); err != nil {
			log.Error("%v", err)
			return
		}
		// 稿件播放量信息
		var newArc = make(map[int64]*showmdl.ArcInfo)
		// 聚合新稿件缓存
		for _, arc := range arcm {
			if arc == nil {
				continue
			}
			na := &showmdl.ArcInfo{
				ID:     arc.Aid,
				Title:  arc.Title,
				View:   arc.Stat.View,
				Author: arc.Author.Name,
			}
			if oa, ok := oldArc[arc.Aid]; ok && oa != nil {
				na.ViewSpeed = na.View - oa.View
			}
			newArc[arc.Aid] = na
		}
		// 缓存物料稿件信息
		if err = s.dao.SetAggArc(context.Background(), ap.ID, newArc); err != nil {
			log.Error("%v", err)
		}
		return
	}
}
