package bws

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

// bwsAchieveRankReload 广州，上海，成都 总排行reload.
func (s *Service) bwsAchieveRankReload(c context.Context) {
	var (
		id      = int64(0)
		bids    = s.c.Bws.Bws2019
		onlyKey = time.Now().Unix()
		err     error
	)
	if len(bids) == 0 {
		return
	}
	for {
		// 反正都是异步执行 sleep 30ms
		var (
			list   []*bwsmdl.Users
			points map[int64]int64
		)
		time.Sleep(time.Millisecond * 30)
		if list, err = s.dao.RawUsersByBid(c, bids, id); err != nil {
			log.Error("BwsAchieveRankReload s.dao.RawUsersByBid(%v,%d) error(%v)", bids, id, err)
			break
		}
		if len(list) == 0 {
			log.Info("BwsAchieveRankReload success(%v,%d)", bids, id)
			break
		}
		mids := make([]int64, 0)
		midMap := make(map[int64][]string)
		for _, v := range list {
			if v.ID > id {
				id = v.ID
			}
			if v.Mid == 0 || v.Key == "" {
				continue
			}
			mids = append(mids, v.Mid)
			midMap[v.Mid] = append(midMap[v.Mid], v.Key)
		}
		if points, err = s.dao.RawCompositeAchievesPoint(c, mids); err != nil {
			log.Error("BwsAchieveRankReload s.dao.RawCompositeAchievesPoint(%v) error(%v)", mids, err)
			break
		}
		for _, val := range mids {
			var (
				isOk bool
				achi *bwsmdl.Achievement
			)
			if _, ok := points[val]; !ok || points[val] == 0 {
				continue
			}
			if _, k := midMap[val]; !k {
				continue
			}
			// 判断mid是否已经回源 10分钟过期
			if isOk, err = s.dao.AchieveReloadSet(c, val, onlyKey, 600); err == nil && !isOk {
				// isok == false 不是首次写入缓存 直接跳过，mid已经回源
				continue
			}
			if achi, err = s.lastCompositeAchievements(c, val); err != nil {
				log.Error("BwsAchieveRankReload s.dao.LastAchievements(%d) error(%v)", val, err)
				continue
			}
			if achi == nil || achi.ID == 0 {
				continue
			}
			// 回源排行榜
			if err = s.dao.IncrAchievesPoint(c, bids[0], val, points[val], int64(achi.Ctime), true); err != nil {
				log.Error("BwsAchieveRankReload s.dao.IncrAchievesPoint(%d,%s,%d) error(%v)", bids[0], midMap[val], val, err)
			}
		}
	}

}

// AchieveRankReload 单场排行榜回源 .
func (s *Service) achieveRankReload(c context.Context, bid int64) {
	var (
		id = int64(0)
	)
	for {
		var (
			err    error
			list   []*bwsmdl.Users
			points map[string]int64
		)
		// 反正都是异步执行 sleep 30ms
		time.Sleep(time.Millisecond * 30)
		if list, err = s.dao.RawUsersByBid(c, []int64{bid}, id); err != nil {
			log.Error("AchieveRankReload s.dao.RawUsersByBid(%d,%d) error(%v)", bid, id, err)
			break
		}
		if len(list) == 0 {
			log.Info("AchieveRankReload success(%d,%d)", bid, id)
			break
		}
		ukeys := make([]string, 0)
		midMap := make(map[string]int64)
		for _, v := range list {
			if v.ID > id {
				id = v.ID
			}
			if v.Mid > 0 && v.Key != "" {
				ukeys = append(ukeys, v.Key)
				midMap[v.Key] = v.Mid
			}
		}
		if points, err = s.dao.RawAchievesPoint(c, bid, ukeys); err != nil {
			log.Error("AchieveRankReload s.dao.RawAchievesPoint(%d,%v) error(%v)", bid, ukeys, err)
			break
		}
		for _, val := range ukeys {
			if _, ok := points[val]; !ok || points[val] == 0 {
				continue
			}
			if _, k := midMap[val]; !k {
				continue
			}
			var achi *bwsmdl.Achievement
			if achi, err = s.dao.LastAchievements(c, map[int64]string{bid: val}); err != nil {
				log.Error("AchieveRankReload s.dao.LastAchievements(%d,%s) error(%v)", bid, val, err)
				continue
			}
			if achi == nil || achi.ID == 0 {
				continue
			}
			// 回源排行榜
			if err = s.dao.IncrSingleAchievesPoint(c, bid, midMap[val], points[val], int64(achi.Ctime), val, true); err != nil {
				log.Error("AchieveRankReload s.dao.IncrSingleAchievesPoint(%d,%s,%d) error(%v)", bid, val, midMap[val], err)
			}
		}
		log.Info("AchieveRankReload AchieveRankReload(%d,id:%d)", bid, id)
	}
}

// RedisInfo .
func (s *Service) RedisInfo(c context.Context, bid, loginMid, mid int64, key, day, typ string, del int, lockType int64, pids []int64) (data json.RawMessage, err error) {
	if !s.isAdmin(loginMid) {
		err = ecode.ActivityNotAdmin
		return
	}
	if key == "" {
		if key, err = s.midToKey(c, bid, mid); err != nil {
			return
		}
	} else {
		if mid, _, err = s.keyToMid(c, bid, key); err != nil {
			return
		}
	}
	var bs []byte
	switch typ {
	case "user_hp": // 用户point点数
		if del == 1 {
			// 删除点数
			if err = s.dao.DelCacheUserHp(c, bid, key); err != nil {
				return
			}
			if _, err = s.dao.UserHp(c, bid, key); err != nil {
				log.Error("s.dao.UserHp(%d,%s) error(%v)", bid, key, err)
			}
			return
		}
		var (
			mapPt int64
		)
		if mapPt, err = s.dao.CacheUserHp(c, bid, key); err != nil {
			log.Error("user_hp s.dao.CacheUserHp error (%v)", err)
			return
		}
		if bs, err = json.Marshal(mapPt); err != nil {
			log.Error("user_hp bws_point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "achieve_rank": //成就排行回源，成就排行获取没有回源逻辑，支持手动回源
		if del == 1 {
			s.cache.Do(c, func(c context.Context) {
				s.achieveRankReload(c, bid)
			})
		} else if del == 2 { //测试数据删除时使用
			s.dao.DelAchieveRank(c, bid)
		}
	case "achieve_rank_2019": // 广州,上海,成都成就排行手动回源
		if del == 1 {
			s.cache.Do(c, func(c context.Context) {
				s.bwsAchieveRankReload(c)
			})
		} else if del == 2 { //测试数据删除时使用
			s.dao.DelAllAchieveRank(c)
		}
	case "achieve_user_rank": // 个人成就排行榜更新
		if del == 1 {
			s.userAchieveRankLoad(c, bid, mid, key)
		}
	case "achieve_user_rank_2019": // 总场成就排行更新
		if del == 1 {
			s.userAllAchieveRankLoad(c, bid, mid, key)
		}
	case "achieve_point": //成就点数
		if del == 1 {
			// 删除缓存
			if err = s.dao.DelCacheAchievesPoint(c, bid, []string{key}); err != nil {
				return
			}
			// 回源数据库
			if _, err = s.dao.AchievesPoint(c, bid, []string{key}); err != nil {
				log.Error("s.dao.AchievesPoint(%d) %s error(%v)", bid, key, err)
			}
			return
		}
		var (
			mapPt map[string]int64
		)
		if mapPt, err = s.dao.CacheAchievesPoint(c, bid, []string{key}); err != nil {
			log.Error("RedisInfo s.dao.CacheAchievesPoint error(%v)", err)
			return
		}
		if bs, err = json.Marshal(mapPt); err != nil {
			log.Error("RedisInfo achieve_point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "achieve_point_2019": //广州上海成都 总成就点数
		if del == 1 {
			if err = s.dao.DelCacheCompositeAchievesPoint(c, []int64{mid}); err != nil {
				return
			}
			if _, err = s.dao.CompositeAchievesPoint(c, []int64{mid}); err != nil {
				log.Error("s.dao.CompositeAchievesPoint(%d) error(%v)", mid, err)
			}
			return
		}
		var (
			mapPt map[int64]int64
		)
		if mapPt, err = s.dao.CacheCompositeAchievesPoint(c, []int64{mid}); err != nil {
			log.Error("RedisInfo s.dao.CacheCompositeAchievesPoint error(%v)", err)
			return
		}
		if bs, err = json.Marshal(mapPt); err != nil {
			log.Error("RedisInfo achieve_point_2019 json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "bws_point":
		if del == 1 {
			err = s.dao.DelCacheBwsPoints(c, pids)
			return
		}
		var (
			mapPt map[int64]*bwsmdl.Point
		)
		if mapPt, err = s.dao.CacheBwsPoints(c, pids); err != nil {
			log.Error("RedisInfo s.dao.CacheBwsPoints error (%v)", err)
			return
		}
		if bs, err = json.Marshal(mapPt); err != nil {
			log.Error("RedisInfo bws_point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "point": // point点 缓存
		if del == 1 {
			err = s.dao.DelCachePoints(c, bid)
			return
		}
		var (
			pids   []int64
			points []*bwsmdl.Point
			mapPt  map[int64]*bwsmdl.Point
		)
		if pids, err = s.dao.CachePoints(c, bid); err != nil || len(pids) == 0 {
			log.Error("RedisInfo point error (%v)", err)
			return
		}
		if mapPt, err = s.dao.CacheBwsPoints(c, pids); err != nil {
			log.Error("RedisInfo s.dao.CacheBwsPoints error (%v)", err)
			return
		}
		for _, v := range pids {
			if _, ok := mapPt[v]; ok {
				points = append(points, mapPt[v])
			}
		}
		if bs, err = json.Marshal(points); err != nil {
			log.Error("RedisInfo point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "achieve": // bis 下achieve列表
		if del == 1 {
			err = s.dao.DelCacheAchievements(c, bid)
			return
		}
		var achieves *bwsmdl.Achievements
		if achieves, err = s.dao.CacheAchievements(c, bid); err != nil || achieves == nil || len(achieves.Achievements) == 0 {
			log.Error("RedisInfo achieve error (%v)", err)
			return
		}
		if bs, err = json.Marshal(achieves.Achievements); err != nil {
			log.Error("RedisInfo achieve json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "user_point": // 失效
		if del == 1 {
			err = s.dao.DelCacheUserPoints(c, bid, key)
			return
		}
		var res []*bwsmdl.UserPoint
		if res, err = s.dao.CacheUserPoints(c, bid, key); err != nil {
			log.Error("RedisInfo user point key(%s) error (%v)", key, err)
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			log.Error("RedisInfo user point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "user_lock_point": // 用户特定类型下完成任务列表
		if del == 1 {
			err = s.dao.DelCacheUserLockPoints(c, bid, lockType, key)
			return
		}
		var res []*bwsmdl.UserPoint
		if res, err = s.dao.CacheUserLockPoints(c, bid, lockType, key); err != nil {
			log.Error("RedisInfo user point key(%s) error (%v)", key, err)
			return
		}
		var pids []int64
		for _, val := range res {
			pids = append(pids, val.Pid)
		}
		var mapPt map[int64]*bwsmdl.Point
		if mapPt, err = s.dao.CacheBwsPoints(c, pids); err != nil {
			log.Error("RedisInfo s.dao.CacheBwsPoints error (%v)", err)
			return
		}
		reply := make([]*bwsmdl.UserLockPointReply, 0)
		for _, val := range res {
			tp := &bwsmdl.UserLockPointReply{UserPoint: val}
			if _, ok := mapPt[val.Pid]; ok {
				tp.Info = mapPt[val.Pid]
			}
			reply = append(reply, tp)
		}
		sort.Slice(reply, func(i, j int) bool {
			return reply[i].ID > reply[j].ID
		})
		if bs, err = json.Marshal(reply); err != nil {
			log.Error("RedisInfo user point json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "user_achieve": //用户成就列表
		if del == 1 {
			err = s.dao.DelCacheUserAchieves(c, bid, key)
			return
		}
		var res []*bwsmdl.UserAchieve
		if res, err = s.dao.CacheUserAchieves(c, bid, key); err != nil {
			log.Error("RedisInfo user achieve key(%s) error (%v)", key, err)
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			log.Error("RedisInfo user achieve json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "achieve_cnt":
		if day == "" {
			day = today()
		}
		if del == 1 {
			err = s.dao.DelCacheAchieveCounts(c, bid, day)
			return
		}
		var res []*bwsmdl.CountAchieves
		if res, err = s.dao.CacheAchieveCounts(c, bid, day); err != nil {
			log.Error("RedisInfo achieve_cnt day(%s) error (%v)", day, err)
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			log.Error("RedisInfo achieve_cnt json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	case "achieve_cnt_db":
		if day == "" {
			day = today()
		}
		var res []*bwsmdl.CountAchieves
		if res, err = s.dao.RawAchieveCounts(c, bid, day); err != nil {
			log.Error("RedisInfo achieve_cnt_db day(%s) error (%v)", day, err)
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			log.Error("RedisInfo achieve_cnt_db json error (%v)", err)
			return
		}
		data = json.RawMessage(bs)
	default:
		err = xecode.RequestErr
	}
	return
}

// KeyInfo .
func (s *Service) KeyInfo(c context.Context, bid, loginMid, keyID, mid int64, key, typ string, del int) (data json.RawMessage, err error) {
	if !s.isAdmin(loginMid) {
		err = ecode.ActivityNotAdmin
		return
	}
	var (
		bs []byte
	)
	switch typ {
	case "id":
		if keyID == 0 {
			err = xecode.RequestErr
			return
		}
		var user *bwsmdl.Users
		if user, err = s.dao.UserByID(c, keyID); err != nil {
			return
		}
		if bs, err = json.Marshal(user); err != nil {
			return
		}
		data = json.RawMessage(bs)
	case "mid":
		if mid == 0 {
			err = xecode.RequestErr
			return
		}
		if del == 1 {
			err = s.dao.DelCacheUsersMid(c, bid, mid)
			return
		}
		var res *bwsmdl.Users
		if res, err = s.dao.CacheUsersMid(c, bid, mid); err != nil {
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			return
		}
		data = json.RawMessage(bs)
	case "key":
		if key == "" {
			err = xecode.RequestErr
			return
		}
		if del == 1 {
			err = s.dao.DelCacheUsersKey(c, bid, key)
			return
		}
		var res *bwsmdl.Users
		if res, err = s.dao.CacheUsersKey(c, bid, key); err != nil {
			return
		}
		if bs, err = json.Marshal(res); err != nil {
			return
		}
		data = json.RawMessage(bs)
	default:
		err = xecode.RequestErr
	}
	return
}

// AdminInfo get admin info.
func (s *Service) AdminInfo(c context.Context, bid, mid int64) (data *bwsmdl.AdminInfo, err error) {
	data = new(bwsmdl.AdminInfo)
	if s.isAdmin(mid) {
		data.IsAdmin = true
	}
	var points *bwsmdl.Points
	if points, err = s.dao.PointsByBid(c, bid); err != nil || points == nil || len(points.Points) == 0 {
		log.Error("s.dao.Points error(%v)", err)
		err = ecode.ActivityPointFail
		return
	}
	for _, v := range points.Points {
		if v.Ower == mid {
			data.Point = v
			break
		}
	}
	for _, v := range s.bwsAllAwards {
		if v != nil && v.Owner == mid {
			data.Award = v
		}
	}
	if data.Point == nil {
		data.Point = struct{}{}
	}
	return
}
