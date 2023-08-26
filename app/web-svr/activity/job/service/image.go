package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model/like"
	l "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_imgRkTypeDay = 1
	_imgRkTypeAll = 2
)

func (s *Service) SetImageLikes() {
	go func() {
		s.loadImageLikes()
	}()
}

func (s *Service) loadImageLikes() {
	if time.Now().Unix() > s.c.ImageV2.Etime.Unix() {
		return
	}
	log.Warn("loadImageLikes sid(%d) start", s.c.ImageV2.Sid)
	ctx := context.Background()
	likeArcs, err := s.loadLikeList(ctx, s.c.ImageV2.Sid, _retryTimes)
	if err != nil {
		log.Error("loadImageLikes s.loadLikeList sid(%d) error(%v)", s.c.ImageV2.Sid, err)
		return
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("loadImageLikes len(aids) == 0")
		return
	}
	var archives []*arcmdl.Arc
	for i := 0; i < aidsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("loadImageLikes s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				archives = append(archives, v)
			}
		}
	}
	upScoreMap := make(map[int64]float64)
	for _, v := range archives {
		if v != nil && v.IsNormal() {
			upScoreMap[v.Author.Mid] += float64(v.Stat.Like)*0.3 + float64(v.Stat.Coin) + float64(v.Stat.Fav)*1.5
		}
	}
	log.Warn("loadImageLikes memory len(upScoreMap):%d", len(upScoreMap))
	// get last day data
	now := time.Now()
	today := now.Format("20060102")
	lastDay := now.AddDate(0, 0, -1).Format("20060102")
	lastUpScoreMap, err := s.dao.ImageUpCache(ctx, s.c.ImageV2.Sid, lastDay, _imgRkTypeAll)
	if err != nil {
		log.Error("loadImageLikes lastDayCache sid:%d,day:%s error(%v)", s.c.ImageV2.Sid, lastDay, err)
		return
	}
	var diffUpLikes []*like.ImageUp
	for mid, score := range upScoreMap {
		r := &like.ImageUp{Mid: mid, Score: score}
		if lastScore, ok := lastUpScoreMap[mid]; ok && lastScore > 0 {
			r.Score = score - lastScore
		}
		diffUpLikes = append(diffUpLikes, r)
	}
	upScoreList := make([]*like.ImageUp, 0, len(upScoreMap))
	for k, v := range upScoreMap {
		upScoreList = append(upScoreList, &like.ImageUp{Mid: k, Score: v})
	}
	if err = s.dao.SetImageUpCache(ctx, s.c.ImageV2.Sid, today, _imgRkTypeDay, diffUpLikes); err != nil {
		log.Error("loadImageLikes SetImageUpCache diffUpLikes(%+v) sid:%d,day:%s error(%v)", diffUpLikes, s.c.ImageV2.Sid, today, err)
		return
	}
	if err = s.dao.SetImageUpCache(ctx, s.c.ImageV2.Sid, today, _imgRkTypeAll, upScoreList); err != nil {
		log.Error("loadImageLikes SetImageUpCache all sid:%d,day:%s error(%v)", s.c.ImageV2.Sid, today, err)
		return
	}
	log.Warn("loadImageLikes success day:%s len(upMap):%d", today, len(upScoreMap))
}

func (s *Service) SetDayImage() {
	ctx := context.Background()
	now := time.Now()
	today := now.Format("20060102")
	lastDay := now.AddDate(0, 0, -1).Format("20060102")
	lastUpScoreMap, err := s.dao.ImageUpCache(ctx, s.c.ImageV2.Sid, lastDay, _imgRkTypeAll)
	if err != nil {
		log.Error("SetDayImage lastDayCache sid:%d,day:%s error(%v)", s.c.ImageV2.Sid, lastDay, err)
		return
	}
	if len(lastUpScoreMap) == 0 {
		log.Warn("SetDayImage len(lastUpScoreMap) == 0")
		return
	}
	todayUpScoreMap, err := s.dao.ImageUpCache(ctx, s.c.ImageV2.Sid, today, _imgRkTypeAll)
	if err != nil {
		log.Error("SetDayImage todayCache sid:%d,day:%s error(%v)", s.c.ImageV2.Sid, today, err)
		return
	}
	if len(todayUpScoreMap) == 0 {
		log.Warn("SetDayImage len(todayUpScoreMap) == 0")
		return
	}
	var diffUpLikes []*like.ImageUp
	for mid, todayScore := range todayUpScoreMap {
		r := &like.ImageUp{Mid: mid, Score: todayScore}
		if lastScore, ok := lastUpScoreMap[mid]; ok && lastScore > 0 {
			r.Score = todayScore - lastScore
		}
		diffUpLikes = append(diffUpLikes, r)
	}
	// set rank
	sort.Slice(diffUpLikes, func(i, j int) bool {
		return diffUpLikes[i].Score > diffUpLikes[j].Score
	})
	if len(diffUpLikes) > s.c.ImageV2.DayLimit {
		diffUpLikes = diffUpLikes[:s.c.ImageV2.DayLimit]
	}
	if err = s.dao.SetImageUpCache(ctx, s.c.ImageV2.Sid, today, _imgRkTypeDay, diffUpLikes); err != nil {
		log.Error("SetDayImage SetImageUpCache diffUpLikes(%+v) sid:%d,day:%s error(%v)", diffUpLikes, s.c.ImageV2.Sid, today, err)
		return
	}
	log.Warn("SetDayImage day:%s success", today)
}

func (s *Service) SetStupidList() {
	go func() {
		s.loadStupidArcs()
	}()
}

func (s *Service) loadStupidArcs() {
	if time.Now().Unix() > s.c.Stupid.Etime {
		return
	}
	ctx := context.Background()
	likeArcs, err := s.loadLikesList(ctx, s.c.Stupid.Sids, _retryTimes)
	if err != nil {
		log.Error("loadStupidArcs s.loadLikeList sids(%v) error(%v)", s.c.Stupid.Sids, err)
		return
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("loadStupidArcs len(aids) == 0")
		return
	}
	var archives []*arcmdl.Arc
	for i := 0; i < aidsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("loadStupidArcs s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				archives = append(archives, v)
			}
		}
	}
	upSlice := make(map[int64][]*like.StupidVv)
	total := int64(0)
	arcs := make(map[int64]int32)
	for _, v := range archives {
		total += int64(v.Stat.View)
		if vv, ok := arcs[v.Author.Mid]; !ok || vv == 0 {
			arcs[v.Author.Mid] = v.Stat.View
			continue
		}
		arcs[v.Author.Mid] += v.Stat.View
	}
	for k, v := range arcs {
		upSlice[k%50] = append(upSlice[k%50], &like.StupidVv{Mid: k, Vv: int64(v)})
	}
	if err := s.dao.AddCacheStupidList(ctx, s.c.Stupid.Sid, upSlice); err != nil {
		log.Error("loadStupidArcs failed to add stupid list: %+v", err)
	}
	if err := s.dao.AddCacheStupidTotal(ctx, s.c.Stupid.Sid, total); err != nil {
		log.Error("loadStupidArcs failed to add stupid total: %+v", err)
	}
	log.Info("loadStupidArcs success")
}

// load stupid act arc
func (s *Service) loadStupidArc() {
	var (
		likes []*l.WebData
		err   error
		tmp   = make(map[int64]struct{})
	)
	if likes, err = s.webDataList(context.Background(), s.c.Stupid.Vid, 0, _objectPieceSize, _retryTimes); err != nil {
		log.Error("loadStupidArc s.webDataList(%d,%d,%d) error(%+v)", s.c.Stupid.Vid, 0, _objectPieceSize, err)
		return
	}
	for _, val := range likes {
		if val == nil || val.Data == "" {
			continue
		}
		aids := &l.AidsData{}
		if err = json.Unmarshal([]byte(val.Data), aids); err != nil {
			log.Error("loadStupidArc json.Unmarshal(%v) error(%+v)", val.Data, err)
			continue
		}
		for _, v := range strings.Split(aids.Aids, ",") {
			if strings.HasPrefix(v, "BV") {
				avid, err := bvid.BvToAv(v)
				if err != nil {
					log.Error("Failed to switch bv to av: %+v %+v", val, err)
					continue
				}
				tmp[avid] = struct{}{}
			} else {
				if avid, _ := strconv.ParseInt(v, 10, 64); avid > 0 {
					tmp[avid] = struct{}{}
				}
			}
		}
	}
	s.stupidArc = tmp
}
