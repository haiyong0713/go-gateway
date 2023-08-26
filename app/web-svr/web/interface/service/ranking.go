package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	_rankIndexLen   = 8
	_rankLen        = 100
	_rankRegionLen  = 10
	_rankOtherLimit = 10
	_aidBulkSize    = 50
)

var (
	_emptyRankArchive = make([]*model.RankArchive, 0)
)

// Ranking get ranking data.
func (s *Service) Ranking(c context.Context, rid int16, rankType, day, arcType int) (res *model.RankData, err error) {
	var (
		rankArc  *model.RankNew
		addCache = true
	)
	if res, err = s.dao.RankingCache(c, rid, rankType, day, arcType); err != nil {
		err = nil
		addCache = false
	} else if res != nil && len(res.List) > 0 {
		return
	}
	if rankArc, err = s.dao.Ranking(c, rid, rankType, day, arcType); err != nil {
		err = nil
	} else if rankArc != nil && len(rankArc.List) > s.c.Rule.MinRankCount {
		res = &model.RankData{Note: rankArc.Note}
		if res.List, err = s.fmtRankArcs(c, rankArc.List, _rankLen); err != nil {
			err = nil
		} else if len(res.List) > 0 {
			if addCache {
				if err := s.cache.Do(c, func(c context.Context) {
					if err := s.dao.SetRankingCache(c, rid, rankType, day, arcType, res); err != nil {
						log.Error("%+v", err)
					}
				}); err != nil {
					log.Error("%+v", err)
				}
			}
			return
		}
	} else {
		log.Error("s.dao.RankingNew(%d,%d,%d) len(aids) (%d)", rid, day, arcType, len(rankArc.List))
	}
	res, err = s.dao.RankingBakCache(c, rid, rankType, day, arcType)
	if res == nil || len(res.List) == 0 {
		res = &model.RankData{List: _emptyRankArchive}
	}
	return
}

// RankingIndex get index ranking data
func (s *Service) RankingIndex(ctx context.Context, day int) ([]*model.IndexArchive, error) {
	arcs, err := func() ([]*arcmdl.Arc, error) {
		aids, memOk := s.rankIndexData[day]
		if !memOk || len(aids) == 0 {
			return nil, errors.New(fmt.Sprintf("RankingIndex day:%d not found", day))
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*arcmdl.Arc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				list = append(list, arc)
			}
		}
		if listLen := len(list); listLen < s.c.Rule.MinRankIndexCount {
			return nil, errors.New(fmt.Sprintf("RankingIndex list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("RankingIndex day:%d error:%v", day, err)
		//get from remote cache
		res, err := s.dao.RankingIndexBakCache(ctx, day)
		if err != nil {
			log.Error("日志告警 RankingIndex RankingIndexBakCache day:%d error:%v", day, err)
			return []*model.IndexArchive{}, nil
		}
		return res, nil
	}
	res := s.fmtIndexArcs(arcs)
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetRankingIndexCache(ctx, day, res); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return res, nil
}

// RankingRegion get region ranking data
// nolint: gocognit
func (s *Service) RankingRegion(ctx context.Context, rid int64, day, original int) ([]*model.RegionArchive, error) {
	arcType, ok := s.typeNames[int32(rid)]
	if !ok {
		log.Warn("RankingRegion rid:%d not found", rid)
		return nil, xecode.RequestErr
	}
	if arcType.Pid == 0 {
		// 一级分区
		var ok bool
		for _, val := range s.c.Rule.RankFirstRegion {
			if rid == val {
				ok = true
				break
			}
		}
		if !ok {
			log.Warn("RankingRegion rid:%d region first not found", rid)
			return nil, xecode.RequestErr
		}
		if _, ok := model.RegionDayType[day]; !ok || original != 0 {
			// 一级分区只有3天总榜,7天榜给h5使用
			log.Warn("RankingRegion rid:%d first region day:%d original:%d", rid, day, original)
			return nil, xecode.RequestErr
		}
	}
	if arcType.Pid != 0 {
		// 二级分区
		for _, val := range s.c.Rule.RankOfflineRegion {
			if rid == val {
				log.Warn("RankingRegion rid:%d offline region", rid)
				return nil, xecode.RequestErr
			}
		}
		if original == 1 {
			for _, val := range s.c.Rule.RankNoOriginalRegion {
				if rid == val {
					log.Warn("RankingRegion rid:%d offline region", rid)
					return nil, xecode.RequestErr
				}
			}
		}
		if !s.regionDayAll(rid) && day == 1 {
			log.Warn("RankingRegion rid:%d not day day:%d", rid, day)
			return nil, xecode.RequestErr
		}
	}
	arcs, err := func() ([]*model.RegionArchive, error) {
		regionAids, memOk := s.rankRegionData[rankRegionMemKey(rid, day, original)]
		if !memOk {
			return nil, errors.New(fmt.Sprintf("RankingRegion rid:%d day:%d original:%d not found", rid, day, original))
		}
		var aids []int64
		for _, v := range regionAids {
			if v != nil && v.Aid > 0 {
				aids = append(aids, v.Aid)
			}
		}
		if len(aids) == 0 {
			return nil, errors.New(fmt.Sprintf("RankingRegion rid:%d day:%d original:%d aids nil", rid, day, original))
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*model.RegionArchive
		for _, v := range regionAids {
			if v == nil {
				continue
			}
			if len(list) > _rankRegionLen {
				break
			}
			if arc, ok := arcs[v.Aid]; ok && arc != nil && arc.IsNormal() {
				var redirectURL string
				if arc.AttrVal(arcmdl.AttrBitJumpUrl) == arcmdl.AttrYes {
					redirectURL = arc.RedirectURL
				}
				list = append(list, &model.RegionArchive{
					Aid:         strconv.FormatInt(arcs[arc.Aid].Aid, 10),
					Bvid:        s.avToBv(arcs[arc.Aid].Aid),
					Typename:    arcs[arc.Aid].TypeName,
					Title:       arcs[arc.Aid].Title,
					Play:        fmtArcView(arcs[arc.Aid]),
					Review:      arcs[arc.Aid].Stat.Reply,
					VideoReview: arcs[arc.Aid].Stat.Danmaku,
					Favorites:   arcs[arc.Aid].Stat.Fav,
					Mid:         arcs[arc.Aid].Author.Mid,
					Author:      arcs[arc.Aid].Author.Name,
					Description: arcs[arc.Aid].Desc,
					Create:      time.Unix(int64(arcs[arc.Aid].PubDate), 0).Format("2006-01-02 15:04"),
					Pic:         arcs[arc.Aid].Pic,
					Coins:       arcs[arc.Aid].Stat.Coin,
					Duration:    fmtDuration(arcs[arc.Aid].Duration),
					Pts:         v.Score,
					Rights:      arcs[arc.Aid].Rights,
					RedirectURL: redirectURL,
				})
			}
		}
		if listLen := len(list); listLen < s.c.Rule.MinRankRegionCount {
			return nil, errors.New(fmt.Sprintf("RankingRegion list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("RankingRegion rid:%d day:%d original:%d error:%v", rid, day, original, err)
		//get from remote cache
		backRes, err := s.dao.RankingRegionBakCache(ctx, rid, day, original)
		if err != nil {
			log.Error("日志告警 RankingTag RankingRegionBakCache rid:%d day:%d original:%d error:%v", rid, day, original, err)
			return []*model.RegionArchive{}, nil
		}
		return backRes, nil
	}
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetRankingRegionCache(ctx, rid, day, original, arcs); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return arcs, nil
}

func rankRegionMemKey(rid int64, day, original int) string {
	return fmt.Sprintf("%d_%d_%d", rid, day, original)
}

// RankingRecommend get rank recommend data.
func (s *Service) RankingRecommend(ctx context.Context, rid int64) ([]*model.IndexArchive, error) {
	var ok bool
	for _, val := range s.c.Rule.RecommendRids {
		if rid == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, xecode.RequestErr
	}
	arcs, err := func() ([]*arcmdl.Arc, error) {
		aids, memOk := s.rankRecommendData[rid]
		if !memOk || len(aids) == 0 {
			return nil, errors.New(fmt.Sprintf("RankingRecommend rid:%d not found", rid))
		}
		if len(aids) == 0 {
			return nil, errors.New("RankingRecommend aids nil")
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*arcmdl.Arc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				list = append(list, arc)
			}
		}
		if listLen := len(list); listLen < s.c.Rule.MinRankRecCount {
			return nil, errors.New(fmt.Sprintf("RankingRecommend list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("RankingRecommend rid:%d error:%v", rid, err)
		//get from remote cache
		backRes, err := s.dao.RankingRecommendBakCache(ctx, rid)
		if err != nil {
			log.Error("日志告警 RankingIndex RankingRecommendBakCache day:%d error:%v", rid, err)
			return []*model.IndexArchive{}, nil
		}
		return backRes, nil
	}
	res := s.fmtIndexArcs(arcs)
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetRankingRecommendCache(ctx, rid, res); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return res, nil
}

func (s *Service) LpRankingRecommend(ctx context.Context, business string) ([]*model.IndexArchive, error) {
	var ok bool
	for key := range s.c.LandingPage {
		if business == key {
			ok = true
			break
		}
	}
	if !ok {
		return nil, xecode.RequestErr
	}
	arcs, err := func() ([]*arcmdl.Arc, error) {
		aids, memOk := s.lpRankRecommendData[business]
		if !memOk || len(aids) == 0 {
			return nil, errors.New(fmt.Sprintf("LpRankingRecommend business:%s not found", business))
		}
		if len(aids) == 0 {
			return nil, errors.New("LpRankingRecommend aids nil")
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*arcmdl.Arc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				list = append(list, arc)
			}
		}
		if listLen := len(list); listLen < s.c.Rule.MinRankRecCount {
			return nil, errors.New(fmt.Sprintf("RankingRecommend list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("LpRankingRecommend business:%s error:%v", business, err)
		//get from remote cache
		backRes, err := s.dao.LpRankingRecommendBakCache(ctx, business)
		if err != nil {
			log.Error("日志告警 RankingIndex RankingRecommendBakCache business:%s error:%v", business, err)
			return []*model.IndexArchive{}, nil
		}
		return backRes, nil
	}
	res := s.fmtIndexArcs(arcs)
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetLpRankingRecommendCache(ctx, business, res); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return res, nil
}

// RankingTag get tag ranking data
func (s *Service) RankingTag(ctx context.Context, rid int16, tagID int64) ([]*model.TagArchive, error) {
	arcs, err := func() ([]*model.TagArchive, error) {
		tagAids, cacheErr := s.dao.RankingTagCache(ctx, rid, tagID)
		if cacheErr != nil {
			return nil, cacheErr
		}
		var aids []int64
		for _, v := range tagAids {
			if v != nil && v.Aid > 0 {
				aids = append(aids, v.Aid)
			}
		}
		if len(aids) == 0 {
			return nil, errors.New("RankingTagCache aids nil")
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*model.TagArchive
		for _, v := range tagAids {
			if v == nil {
				continue
			}
			if arc, ok := arcs[v.Aid]; ok && arc != nil && arc.IsNormal() {
				list = append(list, &model.TagArchive{
					Title:       arc.Title,
					Author:      arc.Author.Name,
					Description: arc.Desc,
					Pic:         arc.Pic,
					Play:        strconv.FormatInt(int64(arc.Stat.View), 10),
					Favorites:   strconv.FormatInt(int64(arc.Stat.Fav), 10),
					Mid:         strconv.FormatInt(arc.Author.Mid, 10),
					Review:      strconv.FormatInt(int64(arc.Stat.Reply), 10),
					CreatedAt:   time.Unix(int64(arcs[arc.Aid].PubDate), 0).Format("2006-01-02 15:04"),
					VideoReview: strconv.FormatInt(int64(arc.Stat.Danmaku), 10),
					Coins:       strconv.FormatInt(int64(arc.Stat.Coin), 10),
					Duration:    strconv.FormatInt(arc.Duration, 10),
					Aid:         arc.Aid,
					Bvid:        s.avToBv(arc.Aid),
					Pts:         v.Score,
					Rights:      arc.Rights,
				})
			}
		}
		if listLen := len(list); listLen < s.c.Rule.MinRankTagCount {
			return nil, errors.New(fmt.Sprintf("RankingTagCache list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("RankingTagCache rid:%d error:%v", rid, err)
		//get from remote cache
		backRes, err := s.dao.RankingTagBakCache(ctx, rid, tagID)
		if err != nil {
			log.Error("日志告警 RankingTag RankingTagBakCache rid:%d tagID:%d error:%v", rid, tagID, err)
			return []*model.TagArchive{}, nil
		}
		return backRes, nil
	}
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetRankingTagCache(ctx, rid, tagID, arcs); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return arcs, nil
}

// WebTop web top data new.
func (s *Service) WebTop(ctx context.Context) ([]*model.BvArc, error) {
	arcs, err := func() ([]*model.BvArc, error) {
		aids := s.webTopData
		if len(aids) == 0 {
			return nil, errors.New("WebTop aids nil")
		}
		arcs, arcErr := s.batchArchives(ctx, aids)
		if arcErr != nil {
			return nil, arcErr
		}
		var list []*model.BvArc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				list = append(list, model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)))
			}
		}
		if listLen := len(list); listLen < s.c.Rule.WebTop {
			return nil, errors.New(fmt.Sprintf("WebTop list count:%d error", listLen))
		}
		return list, nil
	}()
	if err != nil {
		log.Error("WebTop error:%v", err)
		//get from remote cache
		backRes, err := s.dao.WebTopBakCache(ctx)
		if err != nil {
			log.Error("日志告警 WebTop WebTopBakCache error:%v", err)
			return []*model.BvArc{}, nil
		}
		return backRes, nil
	}
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetWebTopCache(ctx, arcs); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return arcs, nil
}

func (s *Service) fmtIndexArcs(arcs []*arcmdl.Arc) (res []*model.IndexArchive) {
	for _, arc := range arcs {
		if len(res) > _rankIndexLen {
			break
		}
		typeName, ok := model.RecSpecTypeName[arc.TypeID]
		if !ok {
			typeName = arc.TypeName
		}
		indexArchive := &model.IndexArchive{
			Aid:         strconv.FormatInt(arc.Aid, 10),
			Bvid:        s.avToBv(arc.Aid),
			Typename:    typeName,
			Title:       arc.Title,
			Play:        fmtArcView(arc),
			Review:      arc.Stat.Reply,
			VideoReview: arc.Stat.Danmaku,
			Favorites:   arc.Stat.Fav,
			Mid:         arc.Author.Mid,
			Author:      arc.Author.Name,
			Description: arc.Desc,
			Create:      time.Unix(int64(arc.PubDate), 0).Format("2006-01-02 15:04"),
			Pic:         arc.Pic,
			Coins:       arc.Stat.Coin,
			Duration:    fmtDuration(arc.Duration),
			Rights:      arc.Rights,
		}
		res = append(res, indexArchive)
	}
	return res
}

func (s *Service) fmtRankArcs(c context.Context, rankArchives []*model.RankNewArchive, arcLen int) (res []*model.RankArchive, err error) {
	var (
		aids []int64
		arcs map[int64]*arcmdl.Arc
	)
	for _, arc := range rankArchives {
		if arc == nil {
			continue
		}
		aids = append(aids, arc.Aid)
		if len(arc.Others) > 0 {
			i := 0
			for _, a := range arc.Others {
				if a == nil {
					continue
				}
				aids = append(aids, a.Aid)
				i++
				if i >= _rankOtherLimit {
					break
				}
			}
		}
	}
	if arcs, err = s.batchArchives(c, aids); err != nil {
		log.Error("s.batchArchives(%v) error(%v)", aids, err)
		return
	}
	for _, arc := range rankArchives {
		if arc == nil {
			continue
		}
		if len(res) > arcLen {
			break
		}
		if _, ok := arcs[arc.Aid]; !ok {
			continue
		}
		var coin, danmu int32
		if arc.RankStat == nil {
			coin = arcs[arc.Aid].Stat.Coin
			danmu = arcs[arc.Aid].Stat.Danmaku
		} else {
			coin = arc.RankStat.Coin
			danmu = arc.RankStat.Danmu
			arcs[arc.Aid].Stat.View = arc.RankStat.Play
		}
		rankArchive := &model.RankArchive{
			Aid:         strconv.FormatInt(arcs[arc.Aid].Aid, 10),
			Bvid:        s.avToBv(arc.Aid),
			Author:      arcs[arc.Aid].Author.Name,
			Coins:       coin,
			Duration:    fmtDuration(arcs[arc.Aid].Duration),
			Mid:         arcs[arc.Aid].Author.Mid,
			Pic:         arcs[arc.Aid].Pic,
			FirstCid:    arcs[arc.Aid].FirstCid,
			Play:        fmtArcView(arcs[arc.Aid]),
			Pts:         arc.Score,
			Title:       arcs[arc.Aid].Title,
			VideoReview: danmu,
			Rights:      arcs[arc.Aid].Rights,
		}
		if len(arc.Others) > 0 {
			for _, a := range arc.Others {
				if a == nil {
					continue
				}
				if _, ok := arcs[a.Aid]; !ok {
					continue
				}
				archive := &model.Other{
					Aid:         a.Aid,
					Bvid:        s.avToBv(a.Aid),
					Play:        fmtArcView(arcs[a.Aid]),
					VideoReview: arcs[a.Aid].Stat.Danmaku,
					Coins:       arcs[a.Aid].Stat.Coin,
					Pts:         a.Score,
					Title:       arcs[a.Aid].Title,
					Pic:         arcs[a.Aid].Pic,
					Duration:    fmtDuration(arcs[a.Aid].Duration),
					Rights:      arcs[a.Aid].Rights,
				}
				rankArchive.Others = append(rankArchive.Others, archive)
			}
		}
		res = append(res, rankArchive)
	}
	return
}

// nolint:gomnd
func fmtDuration(duration int64) (du string) {
	if duration == 0 {
		du = "00:00"
	} else {
		var min, sec string
		min = strconv.Itoa(int(duration / 60))
		if int(duration%60) < 10 {
			sec = "0" + strconv.Itoa(int(duration%60))
		} else {
			sec = strconv.Itoa(int(duration % 60))
		}
		du = min + ":" + sec
	}
	return
}

func fmtArcView(a *arcmdl.Arc) interface{} {
	var view interface{} = a.Stat.View
	if a.Access > 0 {
		view = "--"
	}
	return view
}

func (s *Service) batchArchives(c context.Context, aids []int64) (archives map[int64]*arcmdl.Arc, err error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	group := errgroup.WithContext(c)
	archives = make(map[int64]*arcmdl.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			var arcs *arcmdl.ArcsReply
			arg := &arcmdl.ArcsRequest{Aids: partAids}
			if arcs, err = s.arcGRPC.Arcs(ctx, arg); err != nil {
				log.Error("s.arcGRPC.Arcs(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}

func (s *Service) loadRankIndex() {
	ctx := context.Background()
	tmp := make(map[int][]int64)
	for _, day := range model.IndexDayType {
		tmp[day] = s.rankIndexData[day]
		var aids []int64
		cacheErr := retry(func() (err error) {
			aids, err = s.dao.RankingIndexCache(ctx, day)
			return err
		})
		if cacheErr != nil {
			log.Error("日志告警 loadRankIndex RankingIndexCache day:%d error:%v", day, cacheErr)
			continue
		}
		if cacheCnt := len(aids); cacheCnt < s.c.Rule.MinRankIndexCount {
			log.Error("日志告警 loadRankIndex len RankingIndexCache day:%d count:%d", day, cacheCnt)
			continue
		}
		tmp[day] = aids
	}
	s.rankIndexData = tmp
}

func (s *Service) loadRankRecommend() {
	ctx := context.Background()
	tmp := make(map[int64][]int64)
	for _, rid := range s.c.Rule.RecommendRids {
		tmp[rid] = s.rankRecommendData[rid]
		var aids []int64
		cacheErr := retry(func() (err error) {
			aids, err = s.dao.RankingRecommendCache(ctx, rid)
			return err
		})
		if cacheErr != nil {
			log.Error("日志告警 loadRankRecommend RankingRecommendCache rid:%d error:%v", rid, cacheErr)
			continue
		}
		if cacheCnt := len(aids); cacheCnt < s.c.Rule.MinRankRecCount {
			log.Error("日志告警 loadRankRegion len RankingRecommendCache rid:%d count:%d", rid, cacheCnt)
			continue
		}
		tmp[rid] = aids
	}
	s.rankRecommendData = tmp
}

func (s *Service) loadLpRankRecommend() {
	ctx := context.Background()
	tmp := make(map[string][]int64)
	for business := range s.c.LandingPage {
		tmp[business] = s.lpRankRecommendData[business]
		var aids []int64
		cacheErr := retry(func() (err error) {
			aids, err = s.dao.LpRankingRecommendCache(ctx, business)
			return err
		})
		if cacheErr != nil {
			log.Error("日志告警 loadLpRankRecommend RankingRecommendCache business:%s error:%v", business, cacheErr)
			continue
		}
		if cacheCnt := len(aids); cacheCnt < s.c.Rule.MinRankRecCount {
			log.Error("日志告警 loadLpRankRegion len RankingRecommendCache business:%s count:%d", business, cacheCnt)
			continue
		}
		tmp[business] = aids
	}
	s.lpRankRecommendData = tmp
}

// nolint: gocognit
func (s *Service) loadRankRegion() {
	ctx := context.Background()
	tmp := make(map[string][]*model.NewArchive)
	// 一级分区
	for _, rid := range s.c.Rule.RankFirstRegion {
		for _, day := range model.RegionDayType {
			memKey := rankRegionMemKey(rid, day, 0)
			tmp[memKey] = s.rankRegionData[memKey]
			var cacheData []*model.NewArchive
			cacheErr := retry(func() (err error) {
				cacheData, err = s.dao.RankingRegionCache(ctx, rid, 3, 0)
				return err
			})
			if cacheErr != nil {
				log.Error("日志告警 loadRankRegion RankingRegionCache first rid:%d day:%d error:%v", rid, day, cacheErr)
				continue
			}
			if cacheCnt := len(cacheData); cacheCnt < s.c.Rule.MinRankRegionCount {
				log.Error("日志告警 loadRankRegion len RankingRegionCache first rid:%d day:%d count:%d", rid, day, cacheCnt)
				continue
			}
			tmp[memKey] = cacheData
		}
	}
	// 二级分区
	for _, region := range s.typeNames {
		if region == nil || region.Pid == 0 {
			continue
		}
		for _, val := range s.c.Rule.RankOfflineRegion {
			if int64(region.ID) == val {
				continue
			}
		}
		regionDay := func() map[int]int {
			if s.regionDayAll(int64(region.ID)) {
				return model.RegionDayAll
			}
			return model.RegionDayType
		}()
		for _, day := range regionDay {
			for _, original := range model.RankOriginalType {
				if original == 1 {
					for _, val := range s.c.Rule.RankNoOriginalRegion {
						if int64(region.ID) == val {
							continue
						}
					}
				}
				memKey := rankRegionMemKey(int64(region.ID), day, original)
				tmp[memKey] = s.rankRegionData[memKey]
				var cacheData []*model.NewArchive
				cacheErr := retry(func() (err error) {
					cacheData, err = s.dao.RankingRegionCache(ctx, int64(region.ID), day, original)
					return err
				})
				if cacheErr != nil {
					log.Error("日志告警 loadRankRegion RankingRegionCache second rid:%d day:%d original:%d error:%v", region.ID, day, original, cacheErr)
					continue
				}
				if cacheCnt := len(cacheData); cacheCnt < s.c.Rule.MinRankRegionCount {
					log.Error("日志告警 loadRankRegion len RankingRegionCache second rid:%d day:%d original:%d count:%d", region.ID, day, original, cacheCnt)
					continue
				}
				tmp[memKey] = cacheData
			}
		}
	}
	s.rankRegionData = tmp
}

func (s *Service) loadWebTop() {
	ctx := context.Background()
	var aids []int64
	if err := retry(func() error {
		var err error
		aids, err = s.dao.WebTopCache(ctx)
		return err
	}); err != nil {
		log.Error("日志告警 loadWebTop WebTopCache error:%v", err)
		return
	}
	if cacheCnt := len(aids); cacheCnt < s.c.Rule.WebTop {
		log.Error("日志告警 loadWebTop len WebTopCache count:%d", cacheCnt)
		return
	}
	s.webTopData = aids
	// 多级缓存
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		arcs, err := s.batchArchives(ctx, aids)
		if err != nil {
			log.Error("%+v", err)
			return
		}
		var res []*arcmdl.Arc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				res = append(res, arc)
			}
		}
		if err := s.dao.SetWebTopHotCache(ctx, res); err != nil {
			log.Error("%+v", err)
		}
	})
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func (s *Service) regionDayAll(rid int64) bool {
	for _, allDayRid := range s.c.Rule.DayAllRegion {
		if rid == allDayRid {
			return true
		}
	}
	return false
}
