package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/interface/api"
	likeconst "go-gateway/app/web-svr/activity/interface/model/like"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
	favoriteapi "go-main/app/community/favorite/service/api"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	apiactplat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

type rule struct {
	*likemdl.SubjectRule
	Coefficient []float64
}

type subRule struct {
	*likemdl.ActSubject
	rs []*rule
}

type reserveList struct {
	*subRule
	users []*likemdl.Reserve
}

type ruleScore struct {
	*rule
	Val int64
}

type userScore struct {
	*likemdl.ActSubject
	*likemdl.Reserve
	Score []*ruleScore
}

const (
	sidBatchNum             = 100
	likeBatchNum            = 1000
	reserveUserBatchNum     = 1000
	getCounterResBatchNum   = 10
	userStateUpdateBatchNum = 10
	filterExpireTime        = 86400 * 180
)

const (
	favFilterName = "filter_aid_favorites"
)

var (
	activityRule4Mid  map[int64][]*likemdl.SubjectRule
	activityRule4Sids map[int64][]*likemdl.SubjectRule
	activityRule4Fav  map[int64][]*likemdl.SubjectRule
	lockSyncData      = sync.Mutex{}
	chanFullSync      = make(chan *likemdl.SubjectRule, 16)
)

var (
	runningSubjectNum, receiveSubRule, receiveReserve, receiveScore,
	databusMidStart, databusAvidStart, databusMidDo, databusAvidDo uint64
)

func newRule(r *likemdl.SubjectRule) *rule {
	c := make([]float64, 0, strings.Count(r.Coefficient, ",")+1)
	for _, p := range strings.Split(r.Coefficient, ",") {
		q, err := strconv.ParseFloat(p, 64)
		if err != nil {
			panic(err)
		}
		c = append(c, q)
	}
	return &rule{
		SubjectRule: r,
		Coefficient: c,
	}
}

func (s *Service) UserActionStatSyncOnce() {
	lockSyncData.Lock()
	defer lockSyncData.Unlock()
	log.Info("UserActionStatSyncOnce start")
	defer log.Info("UserActionStatSyncOnce done")
	var subs []*likemdl.ActSubject
	var err error
	var ctx = context.Background()
	now := time.Now()
	// 拉取进行中的预约行为统计活动
	if subs, err = s.dao.SubjectList(ctx, []int64{likeconst.USERACTIONSTAT}, now); err != nil {
		log.Error("userActionStatSyncOnce s.dao.SubjectList error(%+v)", err)
		return
	}
	subNum := len(subs)
	runningSubjectNum = uint64(subNum)
	if subNum == 0 {
		return
	}

	wg := sync.WaitGroup{}

	chSubRule := make(chan *subRule)

	wg.Add(1)
	go func() {
		defer wg.Done()
		// 活动分批处理，拉取活动规则
		for i := 0; i < subNum; i += sidBatchNum {
			var bSubs []*likemdl.ActSubject
			if i+sidBatchNum <= subNum {
				// 满一批次
				bSubs = subs[i : i+sidBatchNum]
			} else {
				// 不足一批次
				bSubs = subs[i:subNum]
			}
			// 拉取活动对应的统计规则
			sids := make([]int64, 0, len(subs))
			for _, sub := range bSubs {
				sids = append(sids, sub.ID)
			}
			var rules []*likemdl.SubjectRule
			rules, err = s.dao.SubjectRulesBySids(ctx, sids)
			if err != nil {
				log.Error("userActionStatSyncOnce s.dao.SubjectRulesBySids error(%+v) sid(%v)", err, sids)
				continue
			}

			// 按照subject组织分活动并行处理
			for j := 0; j < len(sids); j++ {
				sr := &subRule{
					ActSubject: bSubs[j],
					rs:         []*rule{},
				}
				for _, r := range rules {
					if bSubs[j].ID == r.Sid {
						sr.rs = append(sr.rs, newRule(r))
					}
				}
				if len(sr.rs) > 0 {
					chSubRule <- sr
				}
			}
		}
		close(chSubRule)
	}()

	chReserve := make(chan *reserveList, getCounterResBatchNum)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 拉取分批预约用户列表
		for sr := range chSubRule {
			receiveSubRule++
			var minID int64
			for {
				var list []*likemdl.Reserve
				if list, err = s.dao.RawReserveList(ctx, sr.ID, minID, reserveUserBatchNum); err != nil {
					break
				}
				if len(list) == 0 {
					break
				}
				minID = list[len(list)-1].ID
				for _, r := range list {
					chReserve <- &reserveList{
						subRule: sr,
						users: []*likemdl.Reserve{
							r,
						},
					}
				}
			}
		}
		close(chReserve)
	}()

	chUs := make(chan *userScore, userStateUpdateBatchNum)
	// 拉取counter侧统计结果
	wg.Add(1)
	go func() {
		defer wg.Done()
		wgSub := sync.WaitGroup{}
		wgSub.Add(getCounterResBatchNum)
		for i := 0; i < getCounterResBatchNum; i++ {
			go func() {
				defer wgSub.Done()
				for r := range chReserve {
					atomic.AddUint64(&receiveReserve, 1)
					for _, u := range r.users {
						us := &userScore{
							ActSubject: r.ActSubject,
							Reserve:    u,
							Score:      []*ruleScore{},
						}
						send := true
						for _, ru := range r.rs {
							res, err := s.actplatClient.GetCounterRes(ctx, &apiactplat.GetCounterResReq{
								Counter:  ru.RuleName,
								Activity: fmt.Sprint(ru.Sid),
								Mid:      u.Mid,
								Time:     0,
							})
							if err != nil {
								log.Error("userActionStatSyncOnce s.actplatClient.GetCounterRes error(%+v) Counter(%v) Activity(%v) Mid(%v)", err, ru.RuleName, ru.Sid, u.Mid)
								send = false
								break
							}
							var score int64
							if len(res.CounterList) <= 0 {
								log.Warn("userActionStatSyncOnce s.actplatClient.GetCounterRes empty res Counter(%v) Activity(%v) Mid(%v)", ru.RuleName, ru.Sid, u.Mid)
							} else {
								for _, v := range res.CounterList {
									score += v.Val
								}
							}
							us.Score = append(us.Score, &ruleScore{
								rule: ru,
								Val:  score,
							})
						}
						if send {
							chUs <- us
						}
					}
				}
			}()
		}
		wgSub.Wait()
		close(chUs)
	}()

	// 计算实际积分 & 落地行为维度分数 & 活动维度分数
	wg.Add(1)
	go func() {
		defer wg.Done()
		wgSub := sync.WaitGroup{}
		wgSub.Add(getCounterResBatchNum)
		for i := 0; i < userStateUpdateBatchNum; i++ {
			go func() {
				defer wgSub.Done()
				for us := range chUs {
					var total int64
					for _, r := range us.Score {
						atomic.AddUint64(&receiveScore, 1)
						var score int64
						// 根据公式计算实际分数
						switch r.Category {
						//case likemdl.Publish:
						//	{
						//		score = r.Val
						//	}
						//case likemdl.Watch:
						//	{
						//		score = int64(math.Round(float64(r.Val) / r.Coefficient[0]))
						//	}
						//case likemdl.Coin:
						//case likemdl.Agree:
						//	{
						//		score = int64(math.Round(float64(r.Val) * r.Coefficient[0]))
						//	}
						default:
							score = r.Val
						}
						total += score
						// 落地user_state
						_, err = s.actGRPC.SyncUserState(ctx, &api.SyncUserStateReq{
							TaskID: r.TaskID,
							MID:    us.Mid,
							Count:  score,
							SID:    us.ActSubject.ID,
						})
						if err != nil {
							log.Error("userActionStatSyncOnce s.actGRPC.SyncUserState error(%+v) TaskID(%v) MID(%v) Count(%v) SID(%v)", err, r.TaskID, us.Mid, score, us.ActSubject.ID)
						}
					}
					// 落地 reserve
					_, err = s.actGRPC.SyncUserScore(ctx, &api.SyncUserScoreReq{
						SID:   us.ActSubject.ID,
						MID:   us.Mid,
						Score: total,
					})
					if err != nil {
						log.Error("userActionStatSyncOnce s.actGRPC.SyncUserScore error(%+v) SID(%v) MID(%v) Score(%v)", err, us.ActSubject.ID, us.Mid, total)
					}
				}
			}()
		}
		wgSub.Wait()
	}()

	wg.Wait()
}

func (s *Service) userActionStatSyncProc() {
	defer s.waiter.Done()
	for {
		next := time.Now().Add(time.Hour)
		s.UserActionStatSyncOnce()
		now := time.Now()
		if now.Before(next) {
			// 数据规模不够大，一小时执行一次
			time.Sleep(next.Sub(now))
		}
	}
}

func (s *Service) SyncFullData2Counter(ctx context.Context, r *likemdl.SubjectRule) {
	chanFullSync <- r
}

func (s *Service) syncFullData2CounterProc() {
	defer s.waiter.Done()
	for r := range chanFullSync {
		if r.Sids != "" {
			// 存在sid，需要进行一次全量sid同步
			for _, strSid := range strings.Split(r.Sids, ",") {
				sid, _ := strconv.ParseInt(strSid, 10, 64)
				if err := s.syncFullAvid2Counter(sid, r); err != nil {
					log.Error("syncFullData2CounterProc s.syncFullAvid2Counter error(%+v) sid(%v)", err, sid)
				}
			}
		}
		if r.IsStartAtReserve() {
			// 报名之后开始的，需要全量同步一次mid
			if err := s.syncFullMid2Counter(r); err != nil {
				log.Error("syncFullData2CounterProc s.syncFullMid2Counter error(%+v)", err)
			}
		}
	}
}

func (s *Service) syncFullMid2Counter(r *likemdl.SubjectRule) error {
	ctx := context.Background()
	var minID int64
	var err error
	for {
		var list []*likemdl.Reserve
		if list, err = s.dao.RawReserveList(ctx, r.Sid, minID, reserveUserBatchNum); err != nil {
			break
		}
		if len(list) == 0 {
			return nil
		}
		minID = list[len(list)-1].ID
		vals := make([]*apiactplat.FilterMemberInt, 0, len(list))
		ids := make([]int64, 0, len(list))
		for _, p := range list {
			vals = append(vals, &apiactplat.FilterMemberInt{
				Value:      p.Mid,
				ExpireTime: filterExpireTime,
			})
			ids = append(ids, p.Mid)
		}
		log.Info("syncFullMid2Counter s.actplatClient.AddFilterMemberInt Activity[%d] RuleName[%s] mid[%v]",
			r.Sid, r.RuleName, ids)
		_, err = s.actplatClient.AddFilterMemberInt(ctx, &apiactplat.SetFilterMemberIntReq{
			Activity: fmt.Sprint(r.Sid),
			Counter:  r.RuleName,
			Filter:   "filter_mids_apply",
			Values:   vals,
		})
		if err != nil {
			log.Error("syncFullMid2Counter s.actplatClient.AddFilterMemberInt error(%+v)", err)
			return err
		}
	}
	return nil
}

func (s *Service) syncFullAvid2Counter(sid int64, r *likemdl.SubjectRule) error {
	ctx := context.Background()
	var offset int
	for {
		l, err := s.dao.LikeList(ctx, sid, offset, likeBatchNum)
		if err != nil {
			return err
		}
		if len(l) == 0 {
			return nil
		}
		offset += len(l)
		vals := make([]*apiactplat.FilterMemberInt, 0, len(l))
		ids := make([]int64, 0, len(l))
		for _, p := range l {
			vals = append(vals, &apiactplat.FilterMemberInt{
				Value:      p.Wid,
				ExpireTime: filterExpireTime,
			})
			ids = append(ids, p.Wid)
		}
		log.Info("syncFullAvid2Counter s.actplatClient.AddFilterMemberInt Activity[%d] RuleName[%s] wid[%v]",
			r.Sid, r.RuleName, ids)
		_, err = s.actplatClient.AddFilterMemberInt(ctx, &apiactplat.SetFilterMemberIntReq{
			Activity: fmt.Sprint(r.Sid),
			Counter:  r.RuleName,
			Filter:   "filter_aid_sources",
			Values:   vals,
		})
		if err != nil {
			log.Error("syncFullAvid2Counter s.actplatClient.AddFilterMemberInt error(%+v)", err)
			return err
		}
	}
	return nil
}

// SyncFavAvid2Counter ...
func (s *Service) SyncFavAvid2Counter() {
	defer s.waiter.Done()
	var c context.Context
	c = trace.SimpleServerTrace(context.Background(), "syncFavAvid")
	for range time.Tick(time.Second * 60) {
		go func() {
			s.syncAllfav2Counter(c)
		}()
	}
}

// syncfav2Counter 同步收藏夹稿件信息
func (s *Service) syncAllfav2Counter(c context.Context) {
	for _, subject := range activityRule4Fav {
		for _, rule := range subject {
			fidMidList, err := s.getFavFidAndMid(c, rule)
			if err != nil {
				log.Errorc(c, "s.getFavFidAndMid err(%v)", err)
				continue
			}
			if fidMidList != nil {
				for _, v := range fidMidList {
					if v.FID > 0 && v.MID > 0 {
						err = s.synvfav2Counter(c, v.FID, v.MID, fmt.Sprintf("%d", rule.Sid), rule.RuleName)
						if err != nil {
							log.Errorc(c, "s.synvfav2Counter err(%v)", err)
						}
					}
				}

			}

		}
	}
}

func (s *Service) synvfav2Counter(c context.Context, fid, mid int64, activityID, counter string) error {
	// 增加锁
	err := s.dao.FavSyncLock(c, fid, activityID, counter)
	if err != nil {
		return err
	}
	defer s.dao.DelFavSyncLock(c, fid, activityID, counter)
	ch := make(chan []*favoriteapi.ModelFavorite, favChannelLength)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.favoritesIntoChannel(c, mid, fid, ch)
	})
	eg.Go(func(ctx context.Context) (err error) {
		return s.favoritesOutChannel(c, ch, activityID, counter)
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	return err
}

// favoritesAll 获取收藏夹中的数据
func (s *Service) favoritesIntoChannel(c context.Context, mid int64, fid int64, ch chan []*favoriteapi.ModelFavorite) error {
	batch := favPnStart
	var (
		err error
		fav *favoriteapi.FavoritesReply
	)
	defer close(ch)
	for {
		fav, err = s.favoriteClient.FavoritesAll(c, &favoriteapi.FavoritesReq{
			Tp:  favVideoType,
			Mid: mid,
			Uid: mid,
			Fid: fid,
			Pn:  int32(batch),
			Ps:  favPnSize,
			// Tv:  favNeedFilter,
		})
		if err != nil {
			log.Errorc(c, "s.favoriteClient.FavoritesAll: error(%v)", err)
			break
		}
		if fav.Res == nil {
			break
		}
		if len(fav.Res.List) > 0 {
			ch <- fav.Res.List
		}
		if len(fav.Res.List) < favPnSize {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return err

}

// 收藏夹从channel中取出
func (s *Service) favoritesOutChannel(c context.Context, ch chan []*favoriteapi.ModelFavorite, activityID, counter string) (err error) {
	for v := range ch {
		aids := []int64{}
		for _, item := range v {
			if item.State == 0 {
				aids = append(aids, item.Oid)
			}
		}
		err = s.syncAllAidsToActPlat(c, aids, activityID, counter)
		if err != nil {
			log.Error("s.syncAidsToActPlat: error(%v)", err)
		}
	}
	return err
}

func (s *Service) syncAllAidsToActPlat(c context.Context, aids []int64, activityID, counter string) error {
	values := []*actplatapi.FilterMemberInt{}
	expireTime := int64(1200)
	for _, i := range aids {
		values = append(values, &actplatapi.FilterMemberInt{Value: i, ExpireTime: expireTime})
	}
	_, err := s.actplatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
		Activity: activityID,
		Counter:  counter,
		Filter:   favFilterName,
		Values:   values,
	})
	return err
}

func (s *Service) SyncAvid2Counter(ctx context.Context, i *likemdl.Item) {
	atomic.AddUint64(&databusAvidStart, 1)
	if rs, ok := activityRule4Sids[i.Sid]; ok {
		atomic.AddUint64(&databusAvidDo, 1)
		for _, rule := range rs {
			if i.State == 1 {
				log.Info("SyncAvid2Counter s.actplatClient.AddFilterMemberInt Activity[%d] RuleName[%s] wid[%d]",
					rule.Sid, rule.RuleName, i.Wid)
				_, err := s.actplatClient.AddFilterMemberInt(ctx, &apiactplat.SetFilterMemberIntReq{
					Activity: fmt.Sprint(rule.Sid),
					Counter:  rule.RuleName,
					Filter:   "filter_aid_sources",
					Values: []*apiactplat.FilterMemberInt{
						{
							Value:      i.Wid,
							ExpireTime: filterExpireTime,
						},
					},
				})
				if err != nil {
					log.Error("syncAvid2Counter s.actplatClient.AddFilterMemberInt error(%+v)", err)
				}
			} else {
				log.Info("SyncAvid2Counter s.actplatClient.DelFilterMemberInt Activity[%d] RuleName[%s] wid[%d]",
					rule.Sid, rule.RuleName, i.Wid)
				_, err := s.actplatClient.DelFilterMemberInt(ctx, &apiactplat.SetFilterMemberIntReq{
					Activity: fmt.Sprint(rule.Sid),
					Counter:  rule.RuleName,
					Filter:   "filter_aid_sources",
					Values: []*apiactplat.FilterMemberInt{
						{
							Value:      i.Wid,
							ExpireTime: filterExpireTime,
						},
					},
				})
				if err != nil {
					log.Error("syncAvid2Counter s.actplatClient.DelFilterMemberInt error(%+v)", err)
				}
			}
		}
	}
}

func (s *Service) SyncMid2Counter(ctx context.Context, r *likemdl.Reserve) {
	atomic.AddUint64(&databusMidStart, 1)
	if rs, ok := activityRule4Mid[r.Sid]; ok {
		atomic.AddUint64(&databusMidDo, 1)
		for _, rule := range rs {
			log.Info("SyncMid2Counter s.actplatClient.AddFilterMemberInt Activity[%d] RuleName[%s] mid[%d]",
				rule.Sid, rule.RuleName, r.Mid)
			_, err := s.actplatClient.AddFilterMemberInt(ctx, &apiactplat.SetFilterMemberIntReq{
				Activity: fmt.Sprint(r.Sid),
				Counter:  rule.RuleName,
				Filter:   "filter_mids_apply",
				Values: []*apiactplat.FilterMemberInt{
					{
						Value:      r.Mid,
						ExpireTime: filterExpireTime,
					},
				},
			})
			if err != nil {
				log.Error("syncMid2Counter s.actplatClient.AddFilterMemberInt error(%+v)", err)
			}
		}
	}
}

func (s *Service) syncUserActionActivityProc() {
	defer s.waiter.Done()
	var ctx = context.Background()
	s.LoadActivityRule(ctx)
	for range time.Tick(time.Minute) {
		s.LoadActivityRule(ctx)
	}
}

func (s *Service) LoadActivityRule(ctx context.Context) {
	var subs []*likemdl.ActSubject
	var err error
	var arm = make(map[int64][]*likemdl.SubjectRule)
	var ars = make(map[int64][]*likemdl.SubjectRule)
	var arf = make(map[int64][]*likemdl.SubjectRule)
	if subs, err = s.dao.SubjectList(ctx, []int64{likeconst.USERACTIONSTAT}, time.Now()); err != nil {
		log.Error("syncUserActionActivityProc s.dao.SubjectList error(%+v)", err)
		return
	}
	if len(subs) == 0 {
		activityRule4Mid = arm
		activityRule4Sids = ars
		activityRule4Fav = arf
		return
	}
	subNum := len(subs)
	for i := 0; i < subNum; i += sidBatchNum {
		var bSubs []*likemdl.ActSubject
		if i+sidBatchNum <= subNum {
			// 满一批次
			bSubs = subs[i : i+sidBatchNum]
		} else {
			// 不足一批次
			bSubs = subs[i:subNum]
		}
		// 拉取活动对应的统计规则
		sids := make([]int64, 0, len(subs))
		for _, sub := range bSubs {
			sids = append(sids, sub.ID)
		}
		var rules []*likemdl.SubjectRule
		rules, err = s.dao.SubjectRulesBySids(ctx, sids)
		if err != nil {
			log.Error("syncUserActionActivityProc s.dao.SubjectRulesBySids error(%+v)", err)
			return
		}
		for _, r := range rules {
			if r.State != likemdl.RuleStateOnline {
				continue
			}
			// 报名后开始统计
			if r.IsStartAtReserve() {
				var p []*likemdl.SubjectRule
				var ok bool
				if p, ok = arm[r.Sid]; !ok {
					p = make([]*likemdl.SubjectRule, 0, 10)
				}
				p = append(p, r)
				arm[r.Sid] = p
			}
			if r.Sids != "" {
				for _, sid := range strings.Split(r.Sids, ",") {
					iSid, _ := strconv.ParseInt(sid, 10, 64)
					var p []*likemdl.SubjectRule
					var ok bool
					if p, ok = ars[iSid]; !ok {
						p = make([]*likemdl.SubjectRule, 0, 10)
					}
					p = append(p, r)
					ars[iSid] = p
				}
			}
			// 收藏夹
			if r.AidSource != "" && r.AidSourceType == likemdl.SubjectRuleAidSourceTypeFav {
				fidList, err := s.getFavFidAndMid(ctx, r)
				if err != nil {
					continue
				}
				if fidList == nil {
					continue
				}
				p := make([]*likemdl.SubjectRule, 0, 10)
				p = append(p, r)
				arf[r.ID] = p
			}
		}
	}
	activityRule4Mid = arm
	activityRule4Sids = ars
	activityRule4Fav = arf
}

func (s *Service) getFavFidAndMid(ctx context.Context, r *likemdl.SubjectRule) (favList []*likemdl.Fav, err error) {
	err = json.Unmarshal([]byte(r.AidSource), &r.AidSourceMap)
	if err != nil {
		log.Errorc(ctx, "aid source  json error(%+v)", err)
		return
	}
	favList = make([]*likemdl.Fav, 0)
	if r.AidSourceMap != nil && len(r.AidSourceMap) > 0 {
		for _, v := range r.AidSourceMap {
			var fid, mid int64
			aidSourceMap := *v
			if _, ok := aidSourceMap["fid"]; !ok {
				log.Errorc(ctx, "aid source fid err  json")
				return
			}
			if _, ok := aidSourceMap["mid"]; !ok {
				log.Errorc(ctx, "aid source mid err  json")
				return
			}
			switch aidSourceMap["fid"].(type) {
			case int:
				fid = int64(aidSourceMap["fid"].(int))
			case int64:
				fid = aidSourceMap["fid"].(int64)
			case float64:
				fid = int64(aidSourceMap["fid"].(float64))
			}
			switch aidSourceMap["mid"].(type) {
			case int:
				mid = int64(aidSourceMap["mid"].(int))
			case int64:
				mid = aidSourceMap["mid"].(int64)
			case float64:
				mid = int64(aidSourceMap["mid"].(float64))
			}
			favList = append(favList, &likemdl.Fav{FID: fid, MID: mid})
		}

	}

	return

}
func (s *Service) UserActionSyncInfo() interface{} {
	return map[string]interface{}{
		"activityRule4Mid":  activityRule4Mid,
		"activityRule4Sids": activityRule4Sids,
		"activityRule4SFav": activityRule4Fav,
		"counter": map[string]interface{}{
			"runningSubjectNum": runningSubjectNum,
			"receiveSubRule":    receiveSubRule,
			"receiveReserve":    receiveReserve,
			"receiveScore":      receiveScore,
			"databusMidStart":   databusMidStart,
			"databusAvidStart":  databusAvidStart,
			"databusMidDo":      databusMidDo,
			"databusAvidDo":     databusAvidDo,
		},
	}
}

func (s *Service) TunnelGroupAllUser(sid int64) error {
	var (
		ctx       = context.Background()
		err       error
		minID     int64
		pushCount int
	)
	subject, err := s.dao.ActSubject(ctx, sid)
	if err != nil {
		log.Errorc(ctx, "TunnelGroupAllUser s.dao.ActSubject sid(%d) error(%+v)", sid, err)
		return err
	}
	if subject == nil || subject.ID == 0 {
		log.Errorc(ctx, "TunnelGroupAllUser s.dao.ActSubject sid(%d) subject(%+v)", sid, subject)
		return err
	}
	yyType := reserveType()
	if _, ok := yyType[subject.Type]; !ok {
		log.Errorc(ctx, "TunnelGroupAllUser s.dao.ActSubject sid(%d) type(%d) error", sid, subject.Type)
		return xecode.RequestErr
	}
	for {
		var list []*likemdl.ReserveTunnel
		if list, err = s.dao.TunnelReserveList(ctx, sid, minID, reserveUserBatchNum); err != nil {
			log.Errorc(ctx, "TunnelGroupAllUser s.dao.AsyncSendGroupDatabus sid(%d) error(%+v)", sid, err)
			break
		}
		if len(list) == 0 {
			log.Infoc(ctx, "TunnelGroupAllUser s.dao.AsyncSendGroupDatabus sid(%d) count(%+v)", sid, pushCount)
			return nil
		}
		minID = list[len(list)-1].ID
		for _, p := range list {
			pushCount++
			if err = s.dao.AsyncSendGroupDatabus(ctx, &likemdl.Reserve{
				ID:       p.ID,
				Mid:      p.Mid,
				State:    p.State,
				Num:      p.Num,
				Sid:      sid,
				Platform: p.Platform,
				From:     p.From,
			}, p.Ctime.Unix()); err != nil {
				log.Errorc(ctx, "TunnelGroupAllUser s.dao.AsyncSendGroupDatabus sid(%d) reserveID(%d) reserveMid(%d) error(%+v)", sid, p.ID, p.Mid, err)
			} // 预约活动 人群包
			log.Infoc(ctx, "TunnelGroupAllUser s.dao.AsyncSendGroupDatabus sid(%d) mid(%d)", sid, p.Mid)
		}
	}
	log.Infoc(ctx, "TunnelGroupAllUser s.dao.AsyncSendGroupDatabus sid(%d) count(%+v)", sid, pushCount)
	return nil
}

func reserveType() (res map[int]string) {
	res = make(map[int]string, 3)
	res[18] = "预约活动"
	res[22] = "预约打卡活动"
	res[23] = "预约积分活动"
	return res
}
