package native

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
	"go-gateway/app/web-svr/native-page/interface/model/white_list"
)

func keyNewModuleCache(nid int64, pType int32) string {
	if pType > 0 {
		return fmt.Sprintf("nat_nt_me_spt_%d_%d", nid, pType)
	}
	return fmt.Sprintf("nat_nt_me_s_%d", nid)
}

// keyNewMixtureCache .
func keyNewMixtureCache(moduleID int64, mType int32) string {
	return fmt.Sprintf("nat_ntm_ms_%d_%d", moduleID, mType)
}

// keyNewAllMixtureCache .
func keyNewAllMixtureCache(moduleID int64) string {
	return fmt.Sprintf("n_a_nt_mx_%d", moduleID)
}

// keyNtIDsCache .
func keyNtUIDsCache(uid int64) string {
	return fmt.Sprintf("nat_tsu_%d", uid)
}

func keyNtTsOnlineIDsCache(uid int64) string {
	return fmt.Sprintf("nat_tsoli_%d", uid)
}

func keyNtModuleIDsCache(tsID int64) string {
	return fmt.Sprintf("nat_tsids_%d", tsID)
}

func keyNewParticipationCache(moduleID int64) string {
	return fmt.Sprintf("nat_m_ntp_s_%d", moduleID)
}

func keyWhiteListByMid(mid int64) string {
	return fmt.Sprintf("white_list_mid_%d", mid)
}

func keyPageProgressParams(pageID int64) string {
	return fmt.Sprintf("page_prog_params_%d", pageID)
}

func (d *Dao) cacheSFWhiteListByMid(mid int64) string {
	return fmt.Sprintf("white_list_sf_m_%d", mid)
}

func (d *Dao) cacheSFModuleIDs(nid int64, pType int32, _, _ int64) string {
	return fmt.Sprintf("nat_sf_m_%d_%d", nid, pType)
}

func (d *Dao) cacheSFNatMixIDs(moduleID int64, mType int32, _, _ int64) string {
	return fmt.Sprintf("nat_sf_ix_%d_%d", moduleID, mType)
}

func (d *Dao) cacheSFNatAllMixIDs(moduleID int64, _, _ int64) string {
	return fmt.Sprintf("nat_sf_aix_%d", moduleID)
}

func (d *Dao) cacheSFNtTsUIDs(uid int64, _, _ int64) string {
	return fmt.Sprintf("nat_sf_tsp_%d", uid)
}

func (d *Dao) cacheSFNtTsOnlineIDs(uid int64, _, _ int64) string {
	return fmt.Sprintf("nat_sf_tsonl_%d", uid)
}

func (d *Dao) cacheSFNtTsModuleIDs(tsID int64, _, _ int64) string {
	return fmt.Sprintf("nat_sf_tids_%d", tsID)
}

func (d *Dao) cacheSFPartPids(pid int64, _, _ int64) string {
	return fmt.Sprintf("nat_sf_pid_%d", pid)
}

func (d *Dao) cacheSFNtPidToTsID(pid int64) string {
	return fmt.Sprintf("nat_sf_ptos_%d", pid)
}

func ntPidToTsIDKey(pid int64) string {
	return fmt.Sprintf("nat_ts_pts_%d", pid)
}

func ntPageExtKey(pid int64) string {
	return fmt.Sprintf("nat_pg_dy_%d", pid)
}

func ntTitleUniqueKey(title string) string {
	return fmt.Sprintf("nat_ts_tinu_%s", title)
}

func ntMidUniqueKey(mid int64) string {
	return fmt.Sprintf("nat_ts_mid_%d", mid)
}

// CacheModuleIDs .
func (d *Dao) CacheModuleIDs(c context.Context, nid int64, pType int32, offset, end int64) (idReply *dynmdl.ModuleIDsReply, err error) {
	var (
		key = keyNewModuleCache(nid, pType)
	)
	return d.zrangeCommon(c, offset, end, key)
}

// ZREVRANGE
func (d *Dao) zrevangeCommon(c context.Context, start, end int64, key string) (res *dynmdl.ModuleRankReply, err error) {
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if err = conn.Send("ZREVRANGE", key, start, end, "WITHSCORES"); err != nil {
		log.Error("zrevangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZCARD", key); err != nil {
		log.Error("zrevangeCommon conn.Do(ZCARD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZREVRANGE", key, 0, 1); err != nil {
		log.Error("zrevangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("zrevangeCommon conn.Flush() error(%v)", err)
		return
	}
	var (
		items     []interface{}
		lids      []*dynmdl.RankInfo
		checkLids []int64
		count     int64
	)
	if items, err = redis.Values(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	for len(items) > 0 {
		var id, t int64
		if items, err = redis.Scan(items, &id, &t); err != nil {
			log.Error("d.CacheUserLikeList error(%+v)", err)
			return
		}
		lids = append(lids, &dynmdl.RankInfo{ID: id, Score: t})
	}
	if count, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if checkLids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	// 缓存内有数据
	if count > 0 {
		res = &dynmdl.ModuleRankReply{IDs: lids, Count: count}
		if count == 1 && len(checkLids) == 1 && checkLids[0] == -1 {
			res.Count = 0
		}
	}
	return
}

// zrevrangeCommon .
func (d *Dao) zrangeCommon(c context.Context, start, end int64, key string) (res *dynmdl.ModuleIDsReply, err error) {
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if err = conn.Send("ZRANGE", key, start, end); err != nil {
		log.Error("zrangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZCARD", key); err != nil {
		log.Error("zrangeCommon conn.Do(ZCARD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZRANGE", key, 0, 1); err != nil {
		log.Error("zrangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("zrangeCommon conn.Flush() error(%v)", err)
		return
	}
	var (
		lids, checkLids []int64
		count           int64
	)
	if lids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if count, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if checkLids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	// 缓存内有数据
	if count > 0 {
		res = &dynmdl.ModuleIDsReply{IDs: lids, Count: count}
		if count == 1 && len(checkLids) == 1 && checkLids[0] == -1 {
			res.Count = 0
		}
	}
	return
}

// AddCacheModuleIDs .
func (d *Dao) AddCacheModuleIDs(c context.Context, nid int64, miss *dynmdl.ModuleIDsReply, pType int32) (err error) {
	if miss == nil || len(miss.IDs) == 0 {
		return
	}
	var (
		key = keyNewModuleCache(nid, pType)
	)
	return d.AddMixCache(c, key, miss.IDs)
}

// RawModuleIDs .
func (d *Dao) RawModuleIDs(c context.Context, nid int64, pType int32, start, end int64) (*dynmdl.ModuleIDsReply, *dynmdl.ModuleIDsReply, error) {
	var (
		missData    map[int64]int64
		sortLy      []*dynmdl.SortRly
		sortLySlice []int64
		err         error
	)
	// 获取数据库失败
	if missData, err = d.RawSortModules(c, nid, pType); err != nil {
		log.Error(" d.RawSortModules(%d) error(%v)", nid, err)
		return &dynmdl.ModuleIDsReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleIDsReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for k, v := range missData {
		sortLy = append(sortLy, &dynmdl.SortRly{ID: k, Rank: v})
	}
	sort.Slice(sortLy, func(i, j int) bool {
		if sortLy[i].Rank == sortLy[j].Rank {
			return sortLy[i].ID < sortLy[j].ID
		}
		return sortLy[i].Rank < sortLy[j].Rank
	})
	for _, val := range sortLy {
		sortLySlice = append(sortLySlice, val.ID)
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleIDsReply{IDs: sortLySlice, Count: lenCount}, nil
}

// DeleteModuleCache
func (d *Dao) DeleteModuleCache(c context.Context, nid int64, pType int32) (err error) {
	var (
		key = keyNewModuleCache(nid, pType)
	)
	return d.DelMixCache(c, key)
}

// CachePartPids .
func (d *Dao) CachePartPids(c context.Context, moduleID int64, start, end int64) (*dynmdl.ModuleIDsReply, error) {
	var (
		key = keyNewParticipationCache(moduleID)
	)
	return d.zrangeCommon(c, start, end, key)
}

// AddCachePartPids .
func (d *Dao) AddCachePartPids(c context.Context, moduleID int64, miss *dynmdl.ModuleIDsReply) (err error) {
	var (
		key = keyNewParticipationCache(moduleID)
	)
	if miss == nil {
		return
	}
	return d.AddMixCache(c, key, miss.IDs)
}

// DelParticipationCache .
func (d *Dao) DelParticipationCache(c context.Context, moduleID int64) (err error) {
	var (
		key = keyNewParticipationCache(moduleID)
	)
	return d.DelMixCache(c, key)
}

// CacheNatAllMixIDs.
func (d *Dao) CacheNatAllMixIDs(c context.Context, moduleID, start, end int64) (*dynmdl.ModuleIDsReply, error) {
	var (
		key = keyNewAllMixtureCache(moduleID)
	)
	return d.zrangeCommon(c, start, end, key)
}

// CacheNatAllMixIDs.
func (d *Dao) CacheNtTsUIDs(c context.Context, uid, start, end int64) (*dynmdl.ModuleRankReply, error) {
	var (
		key = keyNtUIDsCache(uid)
	)
	return d.zrevangeCommon(c, start, end, key)
}

// CacheNtTsOnlineIDs .
func (d *Dao) CacheNtTsOnlineIDs(c context.Context, uid, start, end int64) (*dynmdl.ModuleRankReply, error) {
	var (
		key = keyNtTsOnlineIDsCache(uid)
	)
	return d.zrevangeCommon(c, start, end, key)
}

// AddCacheNtTsOnlineIDs .
func (d *Dao) AddCacheNtTsOnlineIDs(c context.Context, uid int64, miss *dynmdl.ModuleRankReply) (err error) {
	var (
		key = keyNtTsOnlineIDsCache(uid)
	)
	if miss == nil {
		return
	}
	return d.AddMixScoreCache(c, key, miss.IDs)
}

// AddCacheNtTsIDs .
func (d *Dao) AddCacheNtTsUIDs(c context.Context, uid int64, miss *dynmdl.ModuleRankReply) (err error) {
	var (
		key = keyNtUIDsCache(uid)
	)
	if miss == nil {
		return
	}
	return d.AddMixScoreCache(c, key, miss.IDs)
}

// AddSingleCacheNtTsUIDs .
func (d *Dao) AddSingleCacheNtTsUIDs(c context.Context, uid, id, score int64) (err error) {
	var (
		key = keyNtUIDsCache(uid)
	)
	return d.addSingleCache(c, key, id, score)
}

// AddSingleCacheNtTsOnlineIDs .
func (d *Dao) AddSingleCacheNtTsOnlineIDs(c context.Context, uid, id, score int64) (err error) {
	var (
		key = keyNtTsOnlineIDsCache(uid)
	)
	return d.addSingleCache(c, key, id, score)
}

// ZremSingleCacheNtTsOnlineIDs .
func (d *Dao) ZremSingleCacheNtTsOnlineIDs(c context.Context, uid, id int64) (err error) {
	var (
		key = keyNtTsOnlineIDsCache(uid)
	)
	return d.zremSingleCache(c, key, id)
}

// ZremSingleCacheNtTsUIDs .
func (d *Dao) ZremSingleCacheNtTsUIDs(c context.Context, uid, id int64) (err error) {
	var (
		key = keyNtUIDsCache(uid)
	)
	return d.zremSingleCache(c, key, id)
}

// zremSingleCache .
func (d *Dao) zremSingleCache(c context.Context, key string, id int64) (err error) {
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if _, err = conn.Do("ZREM", key, id); err != nil {
		log.Error("zremSingleCache %s %d error(%v)", key, id, err)
	}
	return
}

func (d *Dao) addSingleCache(c context.Context, key string, id, score int64) (err error) {
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key).Add(score).Add(id)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.send(ZADD %v) error(%v)", args, err)
		return
	}
	// 删除空缓存时写入的标兵
	if err = conn.Send("ZREM", key, -1); err != nil {
		log.Error("conn.send(ZREM) error(%v)", err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheNtTsModuleIDs.
func (d *Dao) CacheNtTsModuleIDs(c context.Context, tsID, start, end int64) (*dynmdl.ModuleIDsReply, error) {
	var (
		key = keyNtModuleIDsCache(tsID)
	)
	return d.zrangeCommon(c, start, end, key)
}

// AddCacheNtTsModuleIDs .
func (d *Dao) AddCacheNtTsModuleIDs(c context.Context, tsID int64, miss *dynmdl.ModuleIDsReply) (err error) {
	var (
		key = keyNtModuleIDsCache(tsID)
	)
	if miss == nil {
		return
	}
	return d.AddMixCache(c, key, miss.IDs)
}

// DelCacheNtTsModuleIDs.
func (d *Dao) DelCacheNtTsModuleIDs(c context.Context, tsID int64) (err error) {
	var (
		key = keyNtModuleIDsCache(tsID)
	)
	return d.DelMixCache(c, key)
}

// CacheNatMixIDs.
func (d *Dao) CacheNatMixIDs(c context.Context, moduleID int64, MType int32, start, end int64) (*dynmdl.ModuleIDsReply, error) {
	var (
		key = keyNewMixtureCache(moduleID, MType)
	)
	return d.zrangeCommon(c, start, end, key)
}

func (d *Dao) RawNatMixIDs(c context.Context, moduleID int64, MType int32, start, end int64) (*dynmdl.ModuleIDsReply, *dynmdl.ModuleIDsReply, error) {
	var (
		missData    []*v1.NativeMixtureExt
		sortLy      []*dynmdl.SortRly
		sortLySlice []int64
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NatMixIDsSearch(c, moduleID, MType); err != nil {
		log.Error(" d.NatMixIDsSearch(%d) error(%v)", moduleID, err)
		return &dynmdl.ModuleIDsReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleIDsReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, v := range missData {
		sortLy = append(sortLy, &dynmdl.SortRly{ID: v.ID, Rank: v.Rank})
	}
	sort.Slice(sortLy, func(i, j int) bool {
		if sortLy[i].Rank == sortLy[j].Rank {
			return sortLy[i].ID < sortLy[j].ID
		}
		return sortLy[i].Rank < sortLy[j].Rank
	})
	for _, val := range sortLy {
		sortLySlice = append(sortLySlice, val.ID)
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleIDsReply{IDs: sortLySlice, Count: lenCount}, nil
}

func (d *Dao) RawNatAllMixIDs(c context.Context, moduleID int64, start, end int64) (*dynmdl.ModuleIDsReply, *dynmdl.ModuleIDsReply, error) {
	var (
		missData    []*v1.NativeMixtureExt
		sortLy      []*dynmdl.SortRly
		sortLySlice []int64
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NatAllMixIDsSearch(c, moduleID); err != nil {
		log.Error(" d.NatMixIDsSearch(%d) error(%v)", moduleID, err)
		return &dynmdl.ModuleIDsReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleIDsReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, v := range missData {
		sortLy = append(sortLy, &dynmdl.SortRly{ID: v.ID, Rank: v.Rank})
	}
	sort.Slice(sortLy, func(i, j int) bool {
		if sortLy[i].Rank == sortLy[j].Rank {
			return sortLy[i].ID < sortLy[j].ID
		}
		return sortLy[i].Rank < sortLy[j].Rank
	})
	for _, val := range sortLy {
		sortLySlice = append(sortLySlice, val.ID)
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleIDsReply{IDs: sortLySlice, Count: lenCount}, nil
}

func (d *Dao) RawPartPids(c context.Context, moduleID int64, start, end int64) (*dynmdl.ModuleIDsReply, *dynmdl.ModuleIDsReply, error) {
	var (
		missData    []*v1.NativeParticipationExt
		sortLy      []*dynmdl.SortRly
		sortLySlice []int64
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NatPartIDsSearch(c, moduleID); err != nil {
		log.Error(" d.NatMixIDsSearch(%d) error(%v)", moduleID, err)
		return &dynmdl.ModuleIDsReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleIDsReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, v := range missData {
		sortLy = append(sortLy, &dynmdl.SortRly{ID: v.ID, Rank: v.Rank})
	}
	sort.Slice(sortLy, func(i, j int) bool {
		if sortLy[i].Rank == sortLy[j].Rank {
			return sortLy[i].ID < sortLy[j].ID
		}
		return sortLy[i].Rank < sortLy[j].Rank
	})
	for _, val := range sortLy {
		sortLySlice = append(sortLySlice, val.ID)
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleIDsReply{IDs: sortLySlice, Count: lenCount}, nil
}

// DelCacheNatAllMixIDs.
func (d *Dao) DelCacheNatAllMixIDs(c context.Context, moduleID int64) (err error) {
	var (
		key = keyNewAllMixtureCache(moduleID)
	)
	return d.DelMixCache(c, key)
}

// DelCacheNatMixIDs .
func (d *Dao) DelCacheNatMixIDs(c context.Context, moduleID int64, MType int32) (err error) {
	var (
		key = keyNewMixtureCache(moduleID, MType)
	)
	return d.DelMixCache(c, key)
}

// AddCacheNatAllMixIDs .
func (d *Dao) AddCacheNatAllMixIDs(c context.Context, moduleID int64, miss *dynmdl.ModuleIDsReply) (err error) {
	var (
		key = keyNewAllMixtureCache(moduleID)
	)
	if miss == nil {
		return
	}
	return d.AddMixCache(c, key, miss.IDs)
}

// AddCacheNatMixIDs .
func (d *Dao) AddCacheNatMixIDs(c context.Context, moduleID int64, miss *dynmdl.ModuleIDsReply, MType int32) (err error) {
	var (
		key = keyNewMixtureCache(moduleID, MType)
	)
	if miss == nil {
		return
	}
	return d.AddMixCache(c, key, miss.IDs)
}

// DelMixCache .
func (d *Dao) DelMixCache(c context.Context, key string) (err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("conn.Send(ZREM %s) error(%v)", key, err)
	}
	return
}

// AddMixScoreCache .
func (d *Dao) AddMixScoreCache(c context.Context, key string, ids []*dynmdl.RankInfo) (err error) {
	if len(ids) == 0 {
		return
	}
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range ids {
		args = args.Add(v.Score).Add(v.ID)
	}
	if _, err = conn.Do("ZADD", args...); err != nil {
		log.Error("conn.send(ZADD %v) error(%v)", args, err)
	}
	return
}

// AddMixCache .
func (d *Dao) AddMixCache(c context.Context, key string, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for k, v := range ids {
		args = args.Add(k + 1).Add(v)
	}
	// 当缓存数据为空缓存时设置过期时间
	expire := 0
	if len(ids) == 1 && ids[0] == -1 {
		expire = 86400 // 1天过期
	}
	count := 1
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.send(ZADD %v) error(%v)", args, err)
		return
	}
	if expire > 0 {
		if err = conn.Send("EXPIRE", key, expire); err != nil {
			log.Error("conn.send(EXPIRE %v) error(%v)", expire, err)
			return
		}
		count++
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) RawNtTsOnlineIDs(c context.Context, uid int64, start, end int64) (*dynmdl.ModuleRankReply, *dynmdl.ModuleRankReply, error) {
	var (
		missData    []*v1.NativePage
		sortLySlice []*dynmdl.RankInfo
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NtTsOnlineIDsSearch(c, uid); err != nil {
		log.Error(" d.NtTsUIDsSearch(%d) error(%v)", uid, err)
		return &dynmdl.ModuleRankReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleRankReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, val := range missData {
		sortLySlice = append(sortLySlice, &dynmdl.RankInfo{ID: val.ID, Score: int64(val.Mtime)})
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleRankReply{IDs: sortLySlice, Count: lenCount}, nil
}

// RawNtTsUIDs .
func (d *Dao) RawNtTsUIDs(c context.Context, uid int64, start, end int64) (*dynmdl.ModuleRankReply, *dynmdl.ModuleRankReply, error) {
	var (
		missData    []*v1.NativePage
		sortLySlice []*dynmdl.RankInfo
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NtTsUIDsSearch(c, uid); err != nil {
		log.Error(" d.NtTsUIDsSearch(%d) error(%v)", uid, err)
		return &dynmdl.ModuleRankReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleRankReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, val := range missData {
		sortLySlice = append(sortLySlice, &dynmdl.RankInfo{ID: val.ID, Score: int64(val.Mtime)})
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleRankReply{IDs: sortLySlice, Count: lenCount}, nil
}

func (d *Dao) RawNtTsModuleIDs(c context.Context, tsID int64, start, end int64) (*dynmdl.ModuleIDsReply, *dynmdl.ModuleIDsReply, error) {
	var (
		missData    []*v1.NativeTsModule
		sortLy      []*dynmdl.SortRly
		sortLySlice []int64
		err         error
	)
	// 获取数据库失败
	if missData, err = d.NtTsModuleIDSearch(c, tsID); err != nil {
		log.Error(" d.NtTsModuleIDSearch(%d) error(%v)", tsID, err)
		return &dynmdl.ModuleIDsReply{}, nil, err
	}
	lenCount := int64(len(missData))
	rly := &dynmdl.ModuleIDsReply{Count: lenCount}
	if lenCount == 0 { // 没有数据
		return rly, nil, nil
	}
	for _, v := range missData {
		sortLy = append(sortLy, &dynmdl.SortRly{ID: v.Id, Rank: v.Rank})
	}
	sort.Slice(sortLy, func(i, j int) bool {
		if sortLy[i].Rank == sortLy[j].Rank {
			return sortLy[i].ID < sortLy[j].ID
		}
		return sortLy[i].Rank < sortLy[j].Rank
	})
	for _, val := range sortLy {
		sortLySlice = append(sortLySlice, val.ID)
	}
	if end == -1 {
		end = lenCount - 1
	}
	end += 1
	if lenCount < end {
		end = lenCount
	}
	if start < lenCount {
		rly.IDs = sortLySlice[start:end]
	}
	return rly, &dynmdl.ModuleIDsReply{IDs: sortLySlice, Count: lenCount}, nil
}

// CacheNtPidToTsIDs
func (d *Dao) CacheNtPidToTsIDs(c context.Context, ids []int64) (res map[int64]int64, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		conn = d.redis.Conn(c)
		list []int64
	)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(ntPidToTsIDKey(v))
	}
	if list, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		log.Error("CacheNtPidToTsIDs %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]int64, len(ids))
	for key, val := range list {
		if val == 0 {
			continue
		}
		res[ids[key]] = val
	}
	return
}

// AddCacheNtPidToTsIDs Set data to mc
func (d *Dao) AddCacheNtPidToTsIDs(c context.Context, val map[int64]int64) (err error) {
	if len(val) == 0 {
		return
	}
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	args := redis.Args{}
	for k, v := range val {
		args = args.Add(ntPidToTsIDKey(k)).Add(v)
	}
	if _, err = conn.Do("MSET", args...); err != nil {
		log.Error("CacheNtPidToTsIDs %v error(%v)", val, err)
	}
	return
}

// CacheNtPidToTsID .
func (d *Dao) CacheNtPidToTsID(c context.Context, id int64) (res int64, err error) {
	key := ntPidToTsIDKey(id)
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Warn("CacheNtPidToTsID %d error(%v)", id, err)
		}
	}
	return
}

// CacheNativePagesExt .
func (d *Dao) CacheNativePagesExt(c context.Context, ids []int64) (res map[int64]*v1.NativePageDyn, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		conn  = d.redis.Conn(c)
		items [][]byte
	)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(ntPageExtKey(v))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("CacheNativePagesExt %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativePageDyn, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativePageDyn)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativePagesExt Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativePagesExt .
func (d *Dao) AddCacheNativePagesExt(c context.Context, missData map[int64]*v1.NativePageDyn) error {
	if len(missData) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range missData {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativePagesExt json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := ntPageExtKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	if err := conn.Send("MSET", args...); err != nil {
		log.Error("AddCacheNativePagesExt MSET error(%v)", err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.mcRegularExpire); err != nil {
			log.Error("AddCacheNativePagesExt conn.Send(Expire, %s, %d) error(%v)", v, d.mcRegularExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Error("AddCacheNativePagesExt Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("AddCacheNativePagesExt conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

// DelCacheNativePagesExt delete data from mc
func (d *Dao) DelCacheNativePagesExt(c context.Context, id int64) (err error) {
	key := ntPageExtKey(id)
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheNtPidToTsID %d error(%v)", id, err)
	}
	return
}

// AddCacheNtPidToTsID Set data to mc
func (d *Dao) AddCacheNtPidToTsID(c context.Context, id int64, val int64) (err error) {
	key := ntPidToTsIDKey(id)
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SET", key, val); err != nil {
		log.Error("AddCacheNtPidToTsID %d error(%v)", id, err)
	}
	return
}

// DelCacheNtPidToTsID delete data from mc
func (d *Dao) DelCacheNtPidToTsID(c context.Context, id int64) (err error) {
	key := ntPidToTsIDKey(id)
	var (
		conn = d.redis.Conn(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheNtPidToTsID %d error(%v)", id, err)
	}
	return
}

// NtTsTitleUnique 与admin后台运营发起活动共用一个key，程序防并发.
func (d *Dao) NtTsTitleUnique(c context.Context, title string) (bool, error) {
	var (
		key = ntTitleUniqueKey(title)
	)
	// 2 过期
	return d.setNXLockCache(c, key, 2)
}

func (d *Dao) NtTsMidUnique(c context.Context, mid int64) (bool, error) {
	var (
		key = ntMidUniqueKey(mid)
	)
	// 5 过期
	return d.setNXLockCache(c, key, 5)
}

func (d *Dao) setNXLockCache(c context.Context, key string, times int32) (bool, error) {
	var ok bool
	conn := d.redis.Conn(c)
	defer conn.Close()
	res, err := conn.Do("SET", key, 1, "EX", times, "NX")
	if err != nil {
		log.Error("conn.Do(SETNX(%s)) error(%v)", key, err)
		return ok, err
	}
	if res == "OK" {
		ok = true
	}
	return ok, nil
}

func (d *Dao) CachePageProgressParams(c context.Context, pageID int64) ([]*v1.ProgressParam, error) {
	key := keyPageProgressParams(pageID)
	conn := d.redis.Conn(c)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return []*v1.ProgressParam{}, nil
		}
		log.Errorc(c, "d.CachePageProgressParams(get key: %v) err: %+v", key, err)
		return nil, err
	}
	res := []*v1.ProgressParam{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CachePageProgressParams(get key: %v) err: %+v", key, err)
		return nil, err
	}
	return res, nil
}

// AddCacheWhiteListByMid Set data to redis
func (d *Dao) AddCacheWhiteListByMid(c context.Context, id int64, val *white_list.WhiteList) (err error) {
	if val == nil {
		return
	}
	key := keyWhiteListByMid(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheWhiteListByMid(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.wlByMidExpire
	if val != nil && val.ID == -1 {
		expire = d.wlByMidNullExpire
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheWhiteListByMid(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheWhiteListByMid get data from redis
func (d *Dao) CacheWhiteListByMid(c context.Context, id int64) (res *white_list.WhiteList, err error) {
	key := keyWhiteListByMid(id)
	conn := d.redis.Conn(c)
	defer conn.Close()
	reply, err1 := redis.Bytes(conn.Do("GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheWhiteListByMid(get key: %v) err: %+v", key, err)
		return
	}
	res = &white_list.WhiteList{}
	err = json.Unmarshal(reply, res)
	if err != nil {
		//log.Errorc(c, "d.CacheWhiteListByMid(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeTabBind get data from redis
func (d *Dao) CacheNativeTabBind(c context.Context, ids []int64, category int32) (res map[int64]int64, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		list []int64
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeTabBindKey(v, category))
	}
	if list, err = redis.Int64s(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNtPidToTsIDs %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]int64, len(ids))
	for key, val := range list {
		if val == 0 {
			continue
		}
		res[ids[key]] = val
	}
	return
}

// AddCacheNativeTabBind Set data to redis
func (d *Dao) AddCacheNativeTabBind(c context.Context, values map[int64]int64, category int32) error {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for id, val := range values {
		// key
		key := nativeTabBindKey(id, category)
		// val
		args = args.Add(key).Add(val)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeTabBind")
}

// CacheNativePart get data from redis
func (d *Dao) CacheNativePart(c context.Context, ids []int64) (res map[int64]*v1.NativeParticipationExt, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeParticipationKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativePart %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeParticipationExt, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeParticipationExt)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativePart Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativePart Set data to redis
func (d *Dao) AddCacheNativePart(c context.Context, values map[int64]*v1.NativeParticipationExt) error {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativePart json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeParticipationKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativePart")
}

// CacheNativeTabs get data from redis
func (d *Dao) CacheNativeTabs(c context.Context, ids []int64) (res map[int64]*v1.NativeActTab, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeTabKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeTabs %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeActTab, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeActTab)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeTabs Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeTabs Set data to redis
func (d *Dao) AddCacheNativeTabs(c context.Context, values map[int64]*v1.NativeActTab) error {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeTabs json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeTabKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeTabs")
}

// CacheNativeTabModules get data from redis
func (d *Dao) CacheNativeTabModules(c context.Context, ids []int64) (res map[int64]*v1.NativeTabModule, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeTabModuleKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeTabModules %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeTabModule, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeTabModule)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeTabModules Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeTabModules Set data to redis
func (d *Dao) AddCacheNativeTabModules(c context.Context, values map[int64]*v1.NativeTabModule) error {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeTabs json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeTabModuleKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeTabs")
}

// CacheNativePages get data from redis
func (d *Dao) CacheNativePages(c context.Context, ids []int64) (res map[int64]*v1.NativePage, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativePageKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativePages %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativePage, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativePage)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativePages Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativePages Set data to redis 永不过期
func (d *Dao) AddCacheNativePages(c context.Context, values map[int64]*v1.NativePage) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativePages json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativePageKey(pid)
		args = args.Add(key).Add(bs)
	}
	if _, err := d.redis.Do(c, "MSET", args...); err != nil {
		log.Error("AddCacheNativePages MSET error(%v)", err)
		return err
	}
	return nil
}

// CacheNtTsPages get data from redis
func (d *Dao) CacheNtTsPages(c context.Context, ids []int64) (res map[int64]*v1.NativeTsPage, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(ntTsPageKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNtTsPages %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeTsPage, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeTsPage)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNtTsPages Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNtTsPages Set data to redis
func (d *Dao) AddCacheNtTsPages(c context.Context, values map[int64]*v1.NativeTsPage) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNtTsPages json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := ntTsPageKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNtTsPages")
}

// CacheNtTsModulesExt get data from redis
func (d *Dao) CacheNtTsModulesExt(c context.Context, ids []int64) (res map[int64]*dynmdl.NativeTsModuleExt, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(ntTsModuleExtKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNtTsModulesExt %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*dynmdl.NativeTsModuleExt, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(dynmdl.NativeTsModuleExt)
		if e := json.Unmarshal(item, a); e != nil {
			log.Error("CacheNtTsModulesExt json Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNtTsModulesExt Set data to redis
func (d *Dao) AddCacheNtTsModulesExt(c context.Context, values map[int64]*dynmdl.NativeTsModuleExt) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := json.Marshal(val)
		if err != nil {
			log.Error("AddCacheNtTsModulesExt json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := ntTsModuleExtKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNtTsModulesExt")
}

// CacheNativeModules get data from redis
func (d *Dao) CacheNativeModules(c context.Context, ids []int64) (res map[int64]*v1.NativeModule, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeModuleKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeModules %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeModule, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeModule)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeModules Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeModules Set data to redis 永远不过期
func (d *Dao) AddCacheNativeModules(c context.Context, values map[int64]*v1.NativeModule) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeModules json.Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeModuleKey(pid)
		args = args.Add(key).Add(bs)
	}
	if _, err := d.redis.Do(c, "MSET", args...); err != nil {
		log.Error("AddCacheNativeModules MSET error(%v)", err)
		return err
	}
	return nil
}

// CacheNativeForeigns get data from redis
func (d *Dao) CacheNativeForeigns(c context.Context, ids []int64, pageType int64) (res map[int64]int64, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		list []int64
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeForeignKey(v, pageType))
	}
	if list, err = redis.Int64s(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeForeigns %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]int64, len(ids))
	for key, val := range list {
		if val == 0 {
			continue
		}
		res[ids[key]] = val
	}
	return
}

// AddCacheNativeForeigns Set data to redis
func (d *Dao) AddCacheNativeForeigns(c context.Context, values map[int64]int64, pageType int64) error {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	for id, val := range values {
		key := nativeForeignKey(id, pageType)
		args = args.Add(key).Add(val)
	}
	if _, err := d.redis.Do(c, "MSET", args...); err != nil {
		log.Error("AddCacheNativeForeigns MSET error(%v)", err)
		return err
	}
	return nil
}

// CacheNativeClicks get data from redis
func (d *Dao) CacheNativeClicks(c context.Context, ids []int64) (res map[int64]*v1.NativeClick, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeClickKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeClicks %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeClick, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeClick)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeClicks Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeClicks Set data to redis
func (d *Dao) AddCacheNativeClicks(c context.Context, values map[int64]*v1.NativeClick) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeClicks Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeClickKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeClicks")
}

// CacheNativeDynamics get data from redis
func (d *Dao) CacheNativeDynamics(c context.Context, ids []int64) (res map[int64]*v1.NativeDynamicExt, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeDynamicKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeDynamics %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeDynamicExt, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeDynamicExt)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeDynamics Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeDynamics Set data to redis
func (d *Dao) AddCacheNativeDynamics(c context.Context, values map[int64]*v1.NativeDynamicExt) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeDynamics Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeDynamicKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeDynamics")
}

// CacheNativeVideos get data from redis
func (d *Dao) CacheNativeVideos(c context.Context, ids []int64) (res map[int64]*v1.NativeVideoExt, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeVideoKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeVideos %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeVideoExt, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeVideoExt)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeVideos Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeVideos Set data to redis
func (d *Dao) AddCacheNativeVideos(c context.Context, values map[int64]*v1.NativeVideoExt) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeVideos Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeVideoKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeVideos")
}

// CacheNativeMixtures get data from redis
func (d *Dao) CacheNativeMixtures(c context.Context, ids []int64) (res map[int64]*v1.NativeMixtureExt, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		items [][]byte
	)
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(nativeMixtureKey(v))
	}
	if items, err = redis.ByteSlices(d.redis.Do(c, "MGET", args...)); err != nil {
		log.Error("CacheNativeMixtures %v error(%v)", ids, err)
		return
	}
	res = make(map[int64]*v1.NativeMixtureExt, len(ids))
	for k, item := range items {
		if item == nil {
			continue
		}
		a := new(v1.NativeMixtureExt)
		if e := a.Unmarshal(item); e != nil {
			log.Error("CacheNativeMixtures Unmarshal error(%v)", e)
			continue
		}
		res[ids[k]] = a
	}
	return
}

// AddCacheNativeMixtures Set data to redis
func (d *Dao) AddCacheNativeMixtures(c context.Context, values map[int64]*v1.NativeMixtureExt) (err error) {
	if len(values) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for pid, val := range values {
		if val == nil {
			continue
		}
		bs, err := val.Marshal()
		if err != nil {
			log.Error("AddCacheNativeMixtures Marshal(%v) error(%v)", val, err)
			continue
		}
		key := nativeMixtureKey(pid)
		args = args.Add(key).Add(bs)
		keys = append(keys, key)
	}
	return d.commonAddCache(c, args, keys, "AddCacheNativeMixtures")
}

func (d *Dao) commonAddCache(c context.Context, args redis.Args, keys []string, funcName string) error {
	conn := d.redis.Conn(c)
	defer conn.Close()
	if err := conn.Send("MSET", args...); err != nil {
		log.Error("commonAddCache-%s MSET error(%v)", funcName, err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.mcRegularExpire); err != nil {
			log.Error("commonAddCache-%s conn.Send(Expire, %s, %d) error(%v)", funcName, v, d.mcRegularExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Error("commonAddCache-%s Flush error(%v)", funcName, err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("commonAddCache-%s conn.Receive() error(%v)", funcName, err)
			return err
		}
	}
	return nil
}

// OnlineNativeModules .
func (d *Dao) OnlineNativeModules(c context.Context, ids []int64) (list map[int64]*v1.NativeModule, err error) {
	res, err := d.NativeModules(c, ids)
	if err != nil {
		return
	}
	list = make(map[int64]*v1.NativeModule)
	for k, v := range res {
		if v.IsOnline() {
			list[k] = v
		}
	}
	return
}

// AddCacheNativeForeign Set data to redis
func (d *Dao) AddCacheNativeForeign(c context.Context, id int64, val int64, pageType int64) (err error) {
	key := nativeForeignKey(id, pageType)
	bs := []byte(strconv.FormatInt(int64(val), 10))
	if _, err = d.redis.Do(c, "set", key, bs); err != nil {
		log.Errorc(c, "d.AddCacheNativeForeign(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeUkey Set data to redis
func (d *Dao) AddCacheNativeUkey(c context.Context, id int64, val int64, ukey string) (err error) {
	key := nativeUkeyKey(id, ukey)
	bs := []byte(strconv.FormatInt(int64(val), 10))
	if _, err = d.redis.Do(c, "set", key, bs); err != nil {
		log.Errorc(c, "d.AddCacheNativeUkey(get key: %v) err: %+v", key, err)
		return
	}
	return
}

func (d *Dao) AddCacheSponsoredUp(c context.Context, mid int64) error {
	key := sponsoredUpKey(mid)
	if _, err := d.redis.Do(c, "set", key, true); err != nil {
		log.Errorc(c, "Fail to AddCacheSponsoredUp, key=%s error=%+v", key, err)
		return err
	}
	return nil
}
