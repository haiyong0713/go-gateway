package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"
)

const (
	_serialAnimation = 33 // 连载动画rid
	// all region archive
	_allRegion = "%d_a_%s"
	// original archive
	_originReg = "%d_o_%s"
	// every all region
	_everyReg = "%d_a"
	// every original region
	_everyOrReg = "%d_o"
	// today region key
	_todayRegCount  = "%d_reg_today"
	_todayRegExpire = 3 * 24 * time.Hour
	// all region key
	_allTypeKey = "all_region"
	// all type expire
	_allTypeExpire = 7 * 24 * time.Hour
)

func fmtAllKey(rid int32, pubDate string) string {
	return fmt.Sprintf(_allRegion, rid, pubDate)
}

func fmtOriginKey(rid int32, pubDate string) string {
	return fmt.Sprintf(_originReg, rid, pubDate)
}

func fmtEveryRegKey(rid int32) string {
	return fmt.Sprintf(_everyReg, rid)
}

func fmtEveryOrRegKey(rid int32) string {
	return fmt.Sprintf(_everyOrReg, rid)
}

func fmtTodayCntKey(rid int32) string {
	return fmt.Sprintf(_todayRegCount, rid)
}

// AddRegionArcCache add all region archive .
func (d *Dao) AddRegionArcCache(ctx context.Context, rid, reid int32, arc ...*arcmdl.RegionArc) (err error) {
	var (
		count int
		conn  = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	if rid <= 0 {
		log.Info("[AddRegionArcCache] arc rid is 0")
		return
	}
	for _, v := range arc {
		var (
			ridStr            = time.Unix(int64(v.PubDate), 0).Format("200601")
			allKey            = fmtAllKey(rid, ridStr)
			originKey         = fmtOriginKey(rid, ridStr)
			todayCntKey       = fmtTodayCntKey(rid)
			fatherTodayCntKey = fmtTodayCntKey(reid)
		)
		info, err := d.ContentFlowControlInfoV2(ctx, v.Aid)
		if err != nil {
			log.Error("日志告警 ContentFlowControlInfoV2 aid:%d, error:%v", v.Aid, err)
		}
		forbidden := model.ItemToArcForbidden(info)
		if !forbidden.AllowShow() {
			log.Info("archive is not allow show aid(%d)", v.Aid)
			continue
		}
		if err = conn.Send("ZADD", allKey, v.PubDate, v.Aid); err != nil {
			log.Error("conn.Send(ZADD, %s, %d) error(%v)", allKey, v.Aid, err)
			continue
		}
		count++
		if err = conn.Send("ZADD", todayCntKey, v.PubDate, v.Aid); err != nil {
			log.Error("conn.Send(ZADD, %s, %d) error(%v)", todayCntKey, v.Aid, err)
			continue
		}
		count++
		if err = conn.Send("ZREMRANGEBYSCORE", todayCntKey, "-inf", time.Now().Add(-time.Duration(_todayRegExpire)).Unix()); err != nil {
			log.Error("conn.Send(ZREMRANGEBYSCORE, %s) error(%v)", todayCntKey, err)
			continue
		}
		count++
		if reid != 0 { // 一级分区当天投稿总数
			if err = conn.Send("ZADD", fatherTodayCntKey, v.PubDate, v.Aid); err != nil {
				log.Error("conn.Send(ZADD, %s, %d) error(%v)", fatherTodayCntKey, v.Aid, err)
				continue
			}
			count++
			if err = conn.Send("ZREMRANGEBYSCORE", fatherTodayCntKey, "-inf", time.Now().Add(-time.Duration(_todayRegExpire)).Unix()); err != nil {
				log.Error("conn.Send(ZREMRANGEBYSCORE, %s) error(%v)", fatherTodayCntKey, err)
				continue
			}
			count++
		}
		if v.Copyright == model.CopyrightOriginal {
			if err = conn.Send("ZADD", originKey, v.PubDate, v.Aid); err != nil {
				log.Error("conn.Send(ZADD, %s, %d) error(%v)", originKey, v.Aid, err)
				continue
			}
			count++
		}
		// 连载动画不记录
		if rid != _serialAnimation {
			if err = conn.Send("ZADD", _allTypeKey, v.PubDate, v.Aid); err != nil {
				log.Error("conn.Send(ZADD, %s, %d, %d) error(%v)", _allTypeKey, v.PubDate, v.Aid, err)
				continue
			}
			count++
			if err = conn.Send("ZREMRANGEBYSCORE", _allTypeKey, "-inf", time.Now().Add(-time.Duration(_allTypeExpire)).Unix()); err != nil {
				log.Error("conn.Send(ZREMRANGEBYSCORE, %s) error(%v)", _allTypeKey, err)
				continue
			}
			count++
		}
	}
	if count == 0 {
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush rid(%d) error(%v)", rid, err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive rid(%d) error(%v)", rid, err)
		}
	}
	// save region archive count
	if err = d.saveRegCount(ctx, rid, arc...); err != nil {
		log.Error("saveRegCount rid(%d) error(%v)", rid, err)
	}
	return
}

// AllRegion get all region archive .
func (d *Dao) AllRegion(ctx context.Context, param []*model.ResKey) (aids []int64, err error) {
	var (
		conn = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	for _, v := range param {
		if err = conn.Send("ZREVRANGE", v.Reskey, v.Start, v.End); err != nil {
			log.Error("conn.Send error(%v)", err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < len(param); i++ {
		var resAid []int64
		if resAid, err = redis.Int64s(conn.Receive()); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
		aids = append(aids, resAid...)
	}
	return
}

// RegionKeyCount get region key count .
func (d *Dao) RegionKeyCount(ctx context.Context, key string) (res []*model.AllRegKey, err error) {
	var (
		conn   = d.rgRds.Get(ctx)
		values []interface{}
		Regtp  [][]byte
	)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZREVRANGE", key, 0, -1)); err != nil {
		log.Error("GetRegionCount redis.Values(ZREVRANGE) key(%s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	if err = redis.ScanSlice(values, &Regtp); err != nil {
		log.Error("redis.ScanSlice error(%v)", err)
		return
	}
	if len(Regtp) == 0 {
		return
	}
	for _, v := range Regtp {
		r := &model.AllRegKey{}
		if err = json.Unmarshal(v, r); err == nil {
			res = append(res, r)
		}
	}
	return
}

// RegCount get region archive count .
func (d *Dao) RegCount(ctx context.Context, key string) (count int64, err error) {
	var (
		res  [][]byte
		conn = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	if res, err = redis.ByteSlices(conn.Do("ZRANGE", key, 0, -1)); err != nil {
		log.Error("conn.Do(ZRANGE) key(%s) error(%v)", key, err)
		return
	}
	if len(res) == 0 {
		return
	}
	for _, v := range res {
		r := &model.AllRegKey{}
		if err = json.Unmarshal(v, r); err != nil {
			log.Warn("json.Unmarshal() key(%s) error(%v)", key, err)
			err = nil
			continue
		}
		count += r.Count
	}
	return
}

// DelArcCache delete region archive .
func (d *Dao) DelArcCache(ctx context.Context, rid, reid int32, param *arcmdl.RegionArc) (err error) {
	var (
		conn              = d.rgRds.Get(ctx)
		ridStr            = time.Unix(int64(param.PubDate), 0).Format("200601")
		allKey            = fmtAllKey(rid, ridStr)
		originKey         = fmtOriginKey(rid, ridStr)
		todayCntKey       = fmtTodayCntKey(rid)
		fatherTodayCntKey = fmtTodayCntKey(reid)
		count             int
	)
	defer conn.Close()
	if err = conn.Send("ZREM", allKey, param.Aid); err != nil {
		log.Error("conn.Send(ZREM) key(%s) error(%v)", allKey, err)
		return
	}
	count++
	if err = conn.Send("ZREM", originKey, param.Aid); err != nil {
		log.Error("conn.Send(ZREM) key(%s) error(%v)", originKey, err)
		return
	}
	count++
	if err = conn.Send("ZREM", todayCntKey, param.Aid); err != nil {
		log.Error("conn.Send(ZREM) key(%s) error(%v)", todayCntKey, err)
		return
	}
	count++
	if reid != 0 {
		if err = conn.Send("ZREM", fatherTodayCntKey, param.Aid); err != nil {
			log.Error("conn.Send(ZREM) key(%s) error(%v)", fatherTodayCntKey, err)
			return
		}
		count++
	}
	if err = conn.Send("ZREM", _allTypeKey, param.Aid); err != nil {
		log.Error("conn.Send(ZREM)key(%s) aid(%d) error(%v)", _allTypeKey, param.Aid, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() rid(%d) error(%v)", rid, err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive()error(%v)", err)
			return
		}
	}
	// save region archive count
	if err = d.saveRegCount(ctx, rid, param); err != nil {
		log.Error("saveRegCount rid(%d) error(%v)", rid, err)
	}
	return
}

func (d *Dao) saveAllKey(ctx context.Context, key string, score string, v *model.AllRegKey) (err error) {
	var (
		res  []byte
		tmp  int
		conn = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	if tmp, err = strconv.Atoi(score); err != nil {
		log.Error("strconv.Atoi  key(%s) error(%v)", key, err)
		return
	}
	if res, err = json.Marshal(v); err != nil {
		log.Error("json.Marshal key(%s) error(%v)", key, err)
		return
	}
	if _, err = conn.Do("ZREMRANGEBYSCORE", key, score, score); err != nil {
		log.Error("conn.Do(ZREMRANGEBYRANK) key(%s) error(%v)", key, err)
		return
	}
	if _, err = conn.Do("ZADD", key, tmp, res); err != nil {
		log.Error("conn.Do(ZADD) key(%s) error(%v)", key, err)
	}
	return
}

func (d *Dao) saveRegCount(ctx context.Context, rid int32, arc ...*arcmdl.RegionArc) (err error) {
	var (
		conn = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	for _, v := range arc {
		var (
			eAllKey    = fmtEveryRegKey(rid)
			eOrKey     = fmtEveryOrRegKey(rid)
			ridStr     = time.Unix(int64(v.PubDate), 0).Format("200601")
			allKey     = fmtAllKey(rid, ridStr)
			originKey  = fmtOriginKey(rid, ridStr)
			aLen, oLen int
			rg         = &model.AllRegKey{}
		)
		if aLen, err = redis.Int(conn.Do("ZCARD", allKey)); err != nil {
			log.Error("conn.Do(ZCARD) aid(%d) error(%v)", v.Aid, err)
			return
		}
		if aLen == 0 {
			log.Warn("[region] 全部rid(%d) key(%s)已经没有数据", rid, rg.Key)
		}
		rg.Key = allKey
		rg.Count = int64(aLen)
		if err = d.saveAllKey(ctx, eAllKey, ridStr, rg); err != nil {
			log.Error("d.saveAllKey() key(%s) aid(%d) error(%v)", eAllKey, v.Aid, err)
			return
		}
		if v.Copyright == model.CopyrightOriginal {
			if oLen, err = redis.Int(conn.Do("ZCARD", originKey)); err != nil {
				log.Error("conn.Do(ZCARD) aid(%d) error(%v)", v.Aid, err)
				return
			}
			if oLen == 0 {
				log.Warn("[region] 原创rid(%d) key(%s)已经没有数据", rid, rg.Key)
			}
			rg.Key = originKey
			rg.Count = int64(oLen)
			if err = d.saveAllKey(ctx, eOrKey, ridStr, rg); err != nil {
				log.Error("d.saveAllKey() key(%s) aid(%d) error(%v)", eAllKey, v.Aid, err)
				return
			}
		}
	}
	return
}

// PushFail fail aids save redis .
func (d *Dao) PushFail(ctx context.Context, a interface{}) (err error) {
	var (
		conn = d.rgRds.Get(ctx)
		bt   []byte
	)
	defer conn.Close()
	if bt, err = json.Marshal(a); err != nil {
		log.Error("json.Marshal(%v) error(%v)", a, err)
		return
	}
	if _, err := conn.Do("RPUSH", model.FailList, bt); err != nil {
		log.Error("conn.Do(RPUSH key(%s)) error(%v)", model.FailList, err)
	}
	return
}

// PopFail lpop retry .
func (d *Dao) PopFail(ctx context.Context) (bt []byte, err error) {
	var (
		conn = d.rgRds.Get(ctx)
	)
	defer conn.Close()
	if bt, err = redis.Bytes(conn.Do("LPOP", model.FailList)); err != nil && err != redis.ErrNil {
		log.Error("conn.Do(LPOP, key(%s)) error(%v)", model.FailList, err)
	}
	return
}

// RegionCnt get second region count today
func (d *Dao) RegionCnt(ctx context.Context, rids []int32, min, max int64) (res map[int32]int64, err error) {
	var conn = d.rgRds.Get(ctx)
	defer conn.Close()
	res = make(map[int32]int64, len(rids))
	for _, v := range rids {
		var (
			key = fmtTodayCntKey(v)
			cnt int64
		)
		if cnt, err = redis.Int64(conn.Do("ZCOUNT", key, min, max)); err != nil {
			log.Error("[SeRegCount] redis.Int64(ZCOUNT) key(%s) param(%d,%d) error(%v)", key, min, max, err)
			return
		}
		res[v] = cnt
	}
	return
}

// RecentRegArc rid=0最近七天稿件 rid!=0最近三天分区稿件
func (d *Dao) RecentRegArc(ctx context.Context, rid int32, min, max int) (aids []int64, err error) {
	var (
		conn = d.rgRds.Get(ctx)
		key  = fmtTodayCntKey(rid)
	)
	defer conn.Close()
	if rid == 0 {
		key = _allTypeKey
	}
	if aids, err = redis.Int64s(conn.Do("ZREVRANGE", key, min, max)); err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
	}
	return
}

// RecentAllRegArcCnt get left 7 day arcs count
func (d *Dao) RecentAllRegArcCnt(ctx context.Context) (count int64, err error) {
	var conn = d.rgRds.Get(ctx)
	defer conn.Close()
	if count, err = redis.Int64(conn.Do("ZCARD", _allTypeKey)); err != nil {
		log.Error("[RecentAllRegArcCnt] redis.Int(ZCOUNT) key(%s) error(%v)", _allTypeKey, err)
	}
	return
}
