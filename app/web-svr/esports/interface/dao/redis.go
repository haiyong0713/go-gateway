package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	mdlEp "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const (
	_keyCont         = "c_%d"
	_keyS9Cont       = "c_%d_%s_%s_%d_%d"
	_keyTimeCont     = "c_%s_%s_%d_%d"
	_keySeasonCont   = "cpb_%d_%d_%d_%d_%d"
	_keyNoSeasonCont = "cpb_%d_%d_%d_%d"
	_keyVideo        = "v_%d"
	_keyCale         = "c_%s_%s"
	_keyFMat         = "fm"
	_keyFVideo       = "fv"
	_keySeason       = "se"
	_keySeasonM      = "sp"
	_keyC            = "co_%d"
	_keyCF           = "cf_%d"
	_keyCAF          = "caf_%d"
	_keyS            = "s_%d"
	_keyG            = "g_%d"
	_keyGMap         = "g_m_%d_%d"
	_keyAct          = "act_%d"
	_keyLive         = "li_%d"
	_keyModule       = "module_%d"
	_keyTop          = "top_%d_%d"
	_keyPoint        = "point_%d_%d_%d"
	_keyKnock        = "knock_%d"
	_keyMAct         = "ma_%d"
	_keyTeam         = "team_%d"
	_keyCSData       = "c_s_data_v2_%d"
	_keyCSDataNew    = "c:data:1119:%d"
	_keyCRecent      = "c_recent_%d_%d_%d_%d"
	_keyGameS        = "gs_%d"
	_keySpecTeam     = "st_%d_%d_%d_%d"
	_keyGuessGS      = "guess_s_game_%d"
	_keyGuessD       = "guess_d_%d"
	_keyGuessRec     = "guess_r_%d_%d"
	_keySearchMain   = "s_card_m"
	_keySearchMD     = "s_c_%d"
	_gameRank        = "h5_game_new"
	_seasonGame      = "season_game"
	_seasonRank      = "s_rank_new_%d"

	cacheKey4DeployIDPod = "esport:pod:%v"

	cacheKey4ContestOfNew = "contest:1012:%v"
	cacheKey4SeasonOfNew  = "season:1012:%v"
	cacheKey4TeamOfNew    = "team:1012:%v"
	_keyTeamInSeason      = "team:in:season:1021:%d"
	_keyMatchSeason       = "match:seasons:0316:%d"
	_keySeasonInfo        = "season:info:0316:%d"
	_keyVideoList         = "video:list:component:%d"
)

func cacheKey4Contest(contestID int64) string {
	return fmt.Sprintf(cacheKey4ContestOfNew, contestID)
}

func cacheKey4Team(teamID int64) string {
	return fmt.Sprintf(cacheKey4TeamOfNew, teamID)
}

func cacheKey4Season(seasonID int64) string {
	return fmt.Sprintf(cacheKey4SeasonOfNew, seasonID)
}

func keyCale(stime, etime string) string {
	return fmt.Sprintf(_keyCale, stime, etime)
}

func keyCont(ps int) string {
	return fmt.Sprintf(_keyCont, ps)
}

func keyS9Cont(sid int64, stime, etime string, ps, sort int) string {
	return fmt.Sprintf(_keyS9Cont, sid, stime, etime, ps, sort)
}

func keyTimeCont(stime, etime string, ps, sort int) string {
	return fmt.Sprintf(_keyTimeCont, stime, etime, ps, sort)
}

func keyNoSeasonCont(stime, etime, ps, sort int64) string {
	return fmt.Sprintf(_keyNoSeasonCont, stime, etime, ps, sort)
}

func keySeasonCont(sid, stime, etime, ps, sort int64) string {
	return fmt.Sprintf(_keySeasonCont, sid, stime, etime, ps, sort)
}

func keyVideo(ps int) string {
	return fmt.Sprintf(_keyVideo, ps)
}
func keyContID(cid int64) string {
	return fmt.Sprintf(_keyC, cid)
}

func keyCoFav(mid int64) string {
	return fmt.Sprintf(_keyCF, mid)
}
func keyCoAppFav(mid int64) string {
	return fmt.Sprintf(_keyCAF, mid)
}

func keySID(sid int64) string {
	return fmt.Sprintf(_keyS, sid)
}

func keyGID(gid int64) string {
	return fmt.Sprintf(_keyG, gid)
}

func keyGMap(oid, tp int64) string {
	return fmt.Sprintf(_keyGMap, oid, tp)
}

func keyTeamID(tid int64) string {
	return fmt.Sprintf(_keyTeam, tid)
}

func keyMatchAct(aid int64) string {
	return fmt.Sprintf(_keyAct, aid)
}
func keyLive(aid int64) string {
	return fmt.Sprintf(_keyLive, aid)
}

func keyCSData(cid int64) string {
	return fmt.Sprintf(_keyCSData, cid)
}

func keyCSDataV2(cid int64) string {
	return fmt.Sprintf(_keyCSDataNew, cid)
}

func keyGuessD(id int64) string {
	return fmt.Sprintf(_keyGuessD, id)
}

func keyCRecent(param *model.ParamCDRecent) string {
	key := fmt.Sprintf(_keyCRecent, param.CID, param.HomeID, param.AwayID, param.Ps)
	return key
}

func keyGuessRec(param *model.ParamEsGuess) string {
	key := fmt.Sprintf(_keyGuessRec, param.HomeID, param.AwayID)
	return key
}

func keyMatchModule(mmid int64) string {
	return fmt.Sprintf(_keyModule, mmid)
}

func keyKnock(mdID int64) string {
	return fmt.Sprintf(_keyKnock, mdID)
}

func keyTop(aid, ps int64) string {
	return fmt.Sprintf(_keyTop, aid, ps)
}

func keyPoint(aid, mdID, ps int64) string {
	return fmt.Sprintf(_keyPoint, aid, mdID, ps)
}

func keyMAct(aid int64) string {
	return fmt.Sprintf(_keyMAct, aid)
}

func keyGameS(tp int64) string {
	return fmt.Sprintf(_keyGameS, tp)
}

func keySpecTeam(tid, sid, tp, r int64) string {
	return fmt.Sprintf(_keySpecTeam, tid, sid, tp, r)
}

func keyMatchSeason(matchID int64) string {
	return fmt.Sprintf(_keyMatchSeason, matchID)
}

func keySeasonInfo(sid int64) string {
	return fmt.Sprintf(_keySeasonInfo, sid)
}

// FMatCache get filter match from cache.
func (d *Dao) FMatCache(c context.Context) (res map[string][]*model.Filter, err error) {
	res, err = d.filterCache(c, _keyFMat)
	return
}

// FVideoCache get filter video from cache.
func (d *Dao) FVideoCache(c context.Context) (res map[string][]*model.Filter, err error) {
	res, err = d.filterCache(c, _keyFVideo)
	return
}

func keyGuessGameSeason(id int64) string {
	return fmt.Sprintf(_keyGuessGS, id)
}

func keySearchMD(id int64) string {
	return fmt.Sprintf(_keySearchMD, id)
}

func keySRank(tp int64) string {
	return fmt.Sprintf(_seasonRank, tp)
}

func (d *Dao) filterCache(c context.Context, key string) (rs map[string][]*model.Filter, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var values []byte
	if values, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Error("filterCache (%s) return nil ", key)
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	rs = make(map[string][]*model.Filter)
	if err = json.Unmarshal(values, &rs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", values, err)
	}
	return
}

// SetFMatCache set  filter match to cache.
func (d *Dao) SetFMatCache(c context.Context, fs map[string][]*model.Filter) (err error) {
	err = d.setFilterCache(c, _keyFMat, fs)
	return
}

// SetFVideoCache set  filter match to cache.
func (d *Dao) SetFVideoCache(c context.Context, fs map[string][]*model.Filter) (err error) {
	err = d.setFilterCache(c, _keyFVideo, fs)
	return
}

func (d *Dao) setFilterCache(c context.Context, key string, fs map[string][]*model.Filter) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(fs); err != nil {
		log.Error("json.Marshal(%v) error(%v)", fs, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s,%d) error(%v)", key, d.filterExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Recevie(%d) error(%v0", i, err)
		}
	}
	return
}

// ContestCache get all contest from cache.
func (d *Dao) ContestCache(c context.Context, ps int) (res []*model.Contest, total int, err error) {
	key := keyCont(ps)
	res, total, err = d.cosCache(c, key)
	return
}

// S9ContestCache get  s9 contest from cache.
func (d *Dao) S9ContestCache(c context.Context, sid int64, stime, etime string, ps, sort int) (res []*model.Contest, total int, err error) {
	key := keyS9Cont(sid, stime, etime, ps, sort)
	res, total, err = d.cosCache(c, key)
	return
}

// ImproveContestCache time contest  from cache.
func (d *Dao) ImproveContestCache(c context.Context, stime, etime string, ps, sort int) (res []*model.Contest, total int, err error) {
	key := keyTimeCont(stime, etime, ps, sort)
	res, total, err = d.cosCache(c, key)
	return
}

// FavCoCache get fav contest from cache.
func (d *Dao) FavCoCache(c context.Context, mid int64) (res []*model.Contest, total int, err error) {
	key := keyCoFav(mid)
	res, total, err = d.cosCache(c, key)
	return
}

// FavCoAppCache get fav contest from cache.
func (d *Dao) FavCoAppCache(c context.Context, mid int64) (res []*model.Contest, total int, err error) {
	key := keyCoAppFav(mid)
	res, total, err = d.cosCache(c, key)
	return
}

func (d *Dao) cosCache(c context.Context, key string) (res []*model.Contest, total int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cont := &model.Contest{}
		if err = json.Unmarshal(bs, cont); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cont)
	}
	total = from(num)
	return
}

// SetContestCache set  all contest to cache.
func (d *Dao) SetContestCache(c context.Context, ps int, contests []*model.Contest, total int) (err error) {
	key := keyCont(ps)
	err = d.setCosCache(c, key, contests, total)
	return
}

// SetS9ContestCache set  s9 contest to cache.
func (d *Dao) SetS9ContestCache(c context.Context, sid int64, stime, etime string, ps, sort int, contests []*model.Contest, total int) (err error) {
	key := keyS9Cont(sid, stime, etime, ps, sort)
	err = d.setCosCache(c, key, contests, total)
	return
}

// SetImproveContestCache set  time contest to cache.
func (d *Dao) SetImproveContestCache(c context.Context, stime, etime string, ps, sort int, contests []*model.Contest, total int) (err error) {
	key := keyTimeCont(stime, etime, ps, sort)
	err = d.setCosCache(c, key, contests, total)
	return
}

// SetFavCoCache set  fav contest to cache.
func (d *Dao) SetFavCoCache(c context.Context, mid int64, contests []*model.Contest, total int) (err error) {
	key := keyCoFav(mid)
	err = d.setCosCache(c, key, contests, total)
	return
}

// SetAppFavCoCache set  fav contest to cache.
func (d *Dao) SetAppFavCoCache(c context.Context, mid int64, contests []*model.Contest, total int) (err error) {
	key := keyCoAppFav(mid)
	err = d.setCosCache(c, key, contests, total)
	return
}

// DelFavCoCache delete fav contests cache.
func (d *Dao) DelFavCoCache(c context.Context, mid int64) (err error) {
	key := keyCoFav(mid)
	keyApp := keyCoAppFav(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL plaKey(%s) error(%v))", key, err)
		return
	}
	if err = conn.Send("DEL", keyApp); err != nil {
		log.Error("conn.Send(DEL pladKey(%s) error(%v))", keyApp, err)
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

func (d *Dao) setCosCache(c context.Context, key string, contests []*model.Contest, total int) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, contest := range contests {
		bs, _ := json.Marshal(contest)
		args = args.Add(combine(int64(sort), total)).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.filterExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// CalendarCache get all calendar from cache.
func (d *Dao) CalendarCache(c context.Context, p *model.ParamFilter) (res []*model.Calendar, err error) {
	var (
		key  = keyCale(p.Stime, p.Etime)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cale := &model.Calendar{}
		if err = json.Unmarshal(bs, cale); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cale)
	}
	return
}

// SetCalendarCache set  all calendar to cache.
func (d *Dao) SetCalendarCache(c context.Context, p *model.ParamFilter, cales []*model.Calendar) (err error) {
	var (
		key  = keyCale(p.Stime, p.Etime)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, cale := range cales {
		bs, _ := json.Marshal(cale)
		args = args.Add(sort).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.filterExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// VideoCache get all video from cache.
func (d *Dao) VideoCache(c context.Context, ps int) (res []*arcmdl.Arc, total int, err error) {
	var (
		key  = keyVideo(ps)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cont := &arcmdl.Arc{}
		if err = json.Unmarshal(bs, cont); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cont)
	}
	total = from(num)
	return
}

// SetVideoCache set  all contest to cache.
func (d *Dao) SetVideoCache(c context.Context, ps int, videos []*arcmdl.Arc, total int) (err error) {
	var (
		key  = keyVideo(ps)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, video := range videos {
		bs, _ := json.Marshal(video)
		args = args.Add(combine(int64(sort), total)).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.filterExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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
func (d *Dao) seasonsCache(c context.Context, key string, start, end int) (res []*model.Season, total int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		object := &model.Season{}
		if err = json.Unmarshal(bs, object); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, object)
	}
	total = from(num)
	return
}

func (d *Dao) setSeasonsCache(c context.Context, key string, seasons []*model.Season, total int) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	for sort, season := range seasons {
		bs, _ := json.Marshal(season)
		if err = conn.Send("ZADD", key, combine(int64(sort), total), bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.listExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// SeasonCache get season list from cache.
func (d *Dao) SeasonCache(c context.Context, start, end int) (res []*model.Season, total int, err error) {
	res, total, err = d.seasonsCache(c, _keySeason, start, end)
	return
}

// SetSeasonCache set season list cache.
func (d *Dao) SetSeasonCache(c context.Context, seasons []*model.Season, total int) (err error) {
	err = d.setSeasonsCache(c, _keySeason, seasons, total)
	return
}

// SeasonMCache get season list from cache.
func (d *Dao) SeasonMCache(c context.Context, start, end int) (res []*model.Season, total int, err error) {
	res, total, err = d.seasonsCache(c, _keySeasonM, start, end)
	return
}

// SetSeasonMCache set season list cache.
func (d *Dao) SetSeasonMCache(c context.Context, seasons []*model.Season, total int) (err error) {
	err = d.setSeasonsCache(c, _keySeasonM, seasons, total)
	return
}

func from(i int64) int {
	return int(i & 0xffff)
}

func combine(sort int64, count int) int64 {
	return sort<<16 | int64(count)
}

// CacheEpContests .
func (d *Dao) CacheEpContests(c context.Context, ids []int64) (res map[int64]*model.Contest, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, cid := range ids {
		key = keyContID(cid)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheEpContests conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*model.Contest, len(ids))
	for _, bs := range bss {
		con := new(model.Contest)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, con); err != nil {
			log.Error("CacheEpContests json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[con.ID] = con
	}
	return
}

// AddCacheEpContests .
func (d *Dao) AddCacheEpContests(c context.Context, data map[int64]*model.Contest) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsCid = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("json.Marshal err(%v)", err)
			continue
		}
		keyID = keyContID(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsCid = argsCid.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsCid...); err != nil {
		log.Error("AddCacheMatchSubjects conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.filterExpire); err != nil {
			log.Error("AddCacheMatchSubjects conn.Send(Expire, %s, %d) error(%v)", v, d.filterExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// CacheEpSeasons .
func (d *Dao) CacheEpSeasons(c context.Context, ids []int64) (res map[int64]*model.Season, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, sid := range ids {
		key = keySID(sid)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheEpSeasons conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*model.Season, len(ids))
	for _, bs := range bss {
		sea := new(model.Season)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, sea); err != nil {
			log.Error("CacheEpSeasons json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[sea.ID] = sea
	}
	return
}

// CacheEpTeams .
func (d *Dao) CacheEpTeams(c context.Context, ids []int64) (res map[int64]*model.Team, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, tid := range ids {
		key = keyTeamID(tid)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheEpTeams conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*model.Team, len(ids))
	for _, bs := range bss {
		team := new(model.Team)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, team); err != nil {
			log.Error("CacheEpTeams json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[team.ID] = team
	}
	return
}

// AddCacheEpTeams .
func (d *Dao) AddCacheEpTeams(c context.Context, data map[int64]*model.Team) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsCid = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheEpTeams.json.Marshal err(%v)", err)
			continue
		}
		keyID = keyTeamID(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsCid = argsCid.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsCid...); err != nil {
		log.Error("AddCacheEpTeams conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.listExpire); err != nil {
			log.Error("AddCacheEpTeams conn.Send(Expire, %s, %d) error(%v)", v, d.listExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// AddCacheEpSeasons .
func (d *Dao) AddCacheEpSeasons(c context.Context, data map[int64]*model.Season) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsCid = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("json.Marshal err(%v)", err)
			continue
		}
		keyID = keySID(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsCid = argsCid.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsCid...); err != nil {
		log.Error("AddCacheEpSeasons conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.listExpire); err != nil {
			log.Error("AddCacheEpSeasons conn.Send(Expire, %s, %d) error(%v)", v, d.listExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// GetActPageCache get act from cache.
func (d *Dao) GetActPageCache(c context.Context, id int64) (act *model.ActivePage, err error) {
	var (
		bs   []byte
		key  = keyMatchAct(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			act = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	act = new(model.ActivePage)
	if err = json.Unmarshal(bs, act); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetCSingleData get contest single data
func (d *Dao) GetCSingleData(c context.Context, id int64) (data *model.ContestDataPage, err error) {
	return d.innerGetCSingleData(c, id, keyCSData)
}
func (d *Dao) GetCSingleDataV2(c context.Context, id int64) (data *model.ContestDataPage, err error) {
	return d.innerGetCSingleData(c, id, keyCSDataV2)
}
func (d *Dao) innerGetCSingleData(c context.Context, id int64, keyCSDataFunc func(int64) string) (data *model.ContestDataPage, err error) {
	var (
		bs   []byte
		key  = keyCSDataFunc(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			data = nil
		} else {
			log.Error("GetCSingleData conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	data = new(model.ContestDataPage)
	if err = json.Unmarshal(bs, data); err != nil {
		log.Error("GetCSingleData json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCSingleData add act first page value
func (d *Dao) AddCSingleData(c context.Context, id int64, act *model.ContestDataPage) (err error) {
	return d.innerAddCSingleData(c, id, act, keyCSData, d.listExpire)
}
func (d *Dao) AddCSingleDataV2(c context.Context, id int64, act *model.ContestDataPage) (err error) {
	return d.innerAddCSingleData(c, id, act, keyCSDataV2, int32(tool.CalculateExpiredSeconds(1)))
}
func (d *Dao) innerAddCSingleData(c context.Context, id int64, act *model.ContestDataPage, keyCSDataFunc func(int64) string, expire int32) (err error) {
	var (
		bs   []byte
		key  = keyCSDataFunc(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(act); err != nil {
		log.Error("AddCSingleData json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddCSingleData conn.Send(SET,%s,%d) error(%v)", key, id, err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("AddCSingleData conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GetCRecent get contest recent data
func (d *Dao) GetCRecent(c context.Context, param *model.ParamCDRecent) (data []*model.Contest, err error) {
	var (
		bs   []byte
		key  = keyCRecent(param)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			data = nil
		} else {
			log.Error("GetCRecent conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	data = make([]*model.Contest, 0)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("GetCRecent json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCRecent add contest recent data
func (d *Dao) AddCRecent(c context.Context, param *model.ParamCDRecent, data []*model.Contest) (err error) {
	var (
		bs   []byte
		key  = keyCRecent(param)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddCRecent json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddCRecent conn.Send(SET,%s,%v) error(%v)", key, param, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("AddCRecent conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCRecent conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCRecent conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GuessRecent get guess recent match contest data
func (d *Dao) GuessRecent(c context.Context, param *model.ParamEsGuess) (data []*model.Contest, err error) {
	var (
		bs   []byte
		key  = keyGuessRec(param)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			data = nil
		} else {
			log.Error("GuessRecent conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	data = make([]*model.Contest, 0)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("GuessRecent json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddGuessRecent add guess recent match contest data
func (d *Dao) AddGuessRecent(c context.Context, param *model.ParamEsGuess, data []*model.Contest) (err error) {
	var (
		bs   []byte
		key  = keyGuessRec(param)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddGuessRecent json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddGuessRecent conn.Send(SET,%s,%v) error(%v)", key, param, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("AddGuessRecent conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddGuessRecent conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddGuessRecent conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// AddActPageCache add act first page value
func (d *Dao) AddActPageCache(c context.Context, aid int64, act *model.ActivePage) (err error) {
	var (
		bs   []byte
		key  = keyMatchAct(aid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(act); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%d) error(%v)", key, aid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GetActModuleCache get module from cache.
func (d *Dao) GetActModuleCache(c context.Context, mmid int64) (res []*model.Video, err error) {
	var (
		bs   []byte
		key  = keyMatchModule(mmid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("GetModuleCache conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("GetModuleCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddActModuleCache add act first page cache
func (d *Dao) AddActModuleCache(c context.Context, mmid int64, module []*model.Video) (err error) {
	var (
		bs   []byte
		key  = keyMatchModule(mmid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(module); err != nil {
		log.Error("AddActModuleCache json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddActModuleCache conn.Send(SET,%s,%d) error(%v)", key, mmid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("AddActModuleCache conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddActModuleCache add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddActModuleCache add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GetActTopCache get act top value cache
func (d *Dao) GetActTopCache(c context.Context, aid, ps int64) (res []*model.Contest, total int, err error) {
	key := keyTop(aid, ps)
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("GetActTopCache conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cont := &model.Contest{}
		if err = json.Unmarshal(bs, cont); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cont)
	}
	total = from(num)
	return
}

// AddActTopCache add act top cache
func (d *Dao) AddActTopCache(c context.Context, aid, ps int64, tops []*model.Contest, total int) (err error) {
	key := keyTop(aid, ps)
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddActTopCache conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, contest := range tops {
		bs, _ := json.Marshal(contest)
		args = args.Add(combine(int64(sort), total)).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddActTopCache conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("AddActTopCache conn.Send(Expire, %s, %d) error(%v)", key, d.listExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("AddActTopCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddActTopCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// GetActPointsCache get act point value
func (d *Dao) GetActPointsCache(c context.Context, aid, mdID, ps int64) (res []*model.Contest, total int, err error) {
	key := keyPoint(aid, mdID, ps)
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("GetActTopCache conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cont := &model.Contest{}
		if err = json.Unmarshal(bs, cont); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cont)
	}
	total = from(num)
	return
}

// AddActPointsCache add act point data cache
func (d *Dao) AddActPointsCache(c context.Context, aid, mdID, ps int64, points []*model.Contest, total int) (err error) {
	key := keyPoint(aid, mdID, ps)
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddActTopCache conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, contest := range points {
		bs, _ := json.Marshal(contest)
		args = args.Add(combine(int64(sort), total)).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddActTopCache conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("AddActTopCache conn.Send(Expire, %s, %d) error(%v)", key, d.listExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("AddActTopCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddActTopCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// GetActKnockoutCache add act knockout cache value
func (d *Dao) GetActKnockoutCache(c context.Context, mdID int64) (res [][]*model.TreeList, err error) {
	var (
		bs   []byte
		key  = keyKnock(mdID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("GetActKnockoutCache conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("GetActKnockoutCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddActKnockoutCache add act knockout cache value
func (d *Dao) AddActKnockoutCache(c context.Context, mdID int64, knock [][]*model.TreeList) (err error) {
	var (
		bs   []byte
		key  = keyKnock(mdID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(knock); err != nil {
		log.Error("AddActKnockoutCache json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddActKnockoutCache conn.Send(SET,%s,%d) error(%v)", key, mdID, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.treeExpire); err != nil {
		log.Error("AddActKnockoutCache conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddActKnockoutCache add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddActModuleCache add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// AddActKnockCacheTime add act knockout cache value time
func (d *Dao) AddActKnockCacheTime(c context.Context, mdID int64) (err error) {
	var (
		key  = keyKnock(mdID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("EXPIRE", key, d.treeExpire); err != nil {
		log.Error("AddActKnockCacheTime conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddActKnockCacheTime add conn.Flush error(%v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("AddActKnockCacheTime add error(%v)", err)
		return
	}
	return
}

// GetMActCache get act cache value
func (d *Dao) GetMActCache(c context.Context, aid int64) (res *model.Active, err error) {
	var (
		bs   []byte
		key  = keyMAct(aid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("GetMActCache conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("GetMActCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddMActCache add act cache value
func (d *Dao) AddMActCache(c context.Context, aid int64, act *model.Active) (err error) {
	var (
		bs   []byte
		key  = keyMAct(aid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(act); err != nil {
		log.Error("AddMActCache json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddMActCache conn.Send(SET,%s,%d) error(%v)", key, aid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("AddMActCache conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddMActCache add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddMActCache add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GetLiveCache get active live from cache.
func (d *Dao) GetLiveCache(c context.Context, id int64) (live *model.ActiveLive, err error) {
	var (
		bs   []byte
		key  = keyLive(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			live = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	live = new(model.ActiveLive)
	if err = json.Unmarshal(bs, live); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddLiveCache add active live cache.
func (d *Dao) AddLiveCache(c context.Context, aid int64, live *model.ActiveLive) (err error) {
	var (
		bs   []byte
		key  = keyLive(aid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(live); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%d) error(%v)", key, aid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GameSeasonCache leida game season from cache.
func (d *Dao) GameSeasonCache(c context.Context, tp int64) (res []*model.Season, err error) {
	var (
		key  = keyGameS(tp)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	var values []byte
	if values, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Error("GameSeasonCache (%s) return nil ", key)
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = make([]*model.Season, 0)
	if err = json.Unmarshal(values, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", values, err)
	}
	return
}

// SetGameSeasonCache leida game season from cache.
func (d *Dao) SetGameSeasonCache(c context.Context, tp int64, seasons []*model.Season) (err error) {
	var (
		key  = keyGameS(tp)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(seasons); err != nil {
		log.Error("json.Marshal(%v) error(%v)", seasons, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s,%d) error(%v)", key, d.filterExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Recevie(%d) error(%v0", i, err)
		}
	}
	return
}

// SpecTeamCache get team special topic from cache.
func (d *Dao) SpecTeamCache(c context.Context, p *model.ParamSpecial) (res model.SpecialTeam, err error) {
	var (
		bs   []byte
		key  = keySpecTeam(p.ID, p.LeidaSID, p.Tp, p.Recent)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddSpecTeamCache add team special topic to cache.
func (d *Dao) AddSpecTeamCache(c context.Context, p *model.ParamSpecial, res model.SpecialTeam) (err error) {
	var (
		bs   []byte
		key  = keySpecTeam(p.ID, p.LeidaSID, p.Tp, p.Recent)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(res); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// GuessCollecCache get guess collection cache
func (d *Dao) GuessCollecCache(c context.Context, gid int64) (res *model.GuessCollection, err error) {
	var (
		bs   []byte
		key  = keyGuessGameSeason(gid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("dao.GuessCollecCache redis.Bytes(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = new(model.GuessCollection)
	if err = json.Unmarshal(bs, res); err != nil {
		log.Error("dao.GuessCollecCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddGuessCollecCache add guess collection values for game season and calendar
func (d *Dao) AddGuessCollecCache(c context.Context, gid int64, value *model.GuessCollection) (err error) {
	var (
		bs   []byte
		key  = keyGuessGameSeason(gid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(value); err != nil {
		log.Error("dao.AddGuessCollecCache json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("dao.AddGuessCollecCache conn.Send(SET,%s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("dao.AddGuessCollecCache conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("dao.AddGuessCollecCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("dao.AddGuessCollecCache conn.Receive(%d) error(%v)", i+1, err)
			return
		}
	}
	return
}

// GuessDetailCache get guess detail
func (d *Dao) GuessDetailCache(c context.Context, id int64) (data *model.GuessDetail, err error) {
	var (
		bs   []byte
		key  = keyGuessD(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			data = nil
		} else {
			log.Error("GuessDetailCache conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	data = new(model.GuessDetail)
	if err = json.Unmarshal(bs, data); err != nil {
		log.Error("GuessDetailCache json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddGuessDetailCache add guess detail
func (d *Dao) AddGuessDetailCache(c context.Context, id int64, data *model.GuessDetail) (err error) {
	var (
		bs   []byte
		key  = keyGuessD(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddGuessDetailCache json.Marshal(%v) error(%v)", data, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("AddGuessDetailCache conn.Send(SET,%s,%d) error(%v)", key, id, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("AddGuessDetailCache conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddGuessDetailCache add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddGuessDetailCache add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// DelGuessDetailCache delete guess detail cache
func (d *Dao) DelGuessDetailCache(c context.Context, id int64) (err error) {
	var (
		key = keyGuessD(id)
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelGuessDetailCache conn.Do(DEL keyGuessD(%s) error(%v))", key, err)
		return
	}
	return
}

// CacheSearchMainIDs .
func (d *Dao) CacheSearchMainIDs(c context.Context) (res []int64, err error) {
	var (
		bs   []byte
		key  = _keySearchMain
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheSearchMainIDs .
func (d *Dao) AddCacheSearchMainIDs(c context.Context, data []int64) (err error) {
	var (
		bs   []byte
		key  = _keySearchMain
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.filterExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.filterExpire, bs)
		return
	}
	return
}

// CacheSearchMD .
func (d *Dao) CacheSearchMD(c context.Context, mainIDs []int64) (res map[int64]*model.SearchRes, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, mainID := range mainIDs {
		key = keySearchMD(mainID)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheSearchMD conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*model.SearchRes, len(mainIDs))
	for _, bs := range bss {
		md := new(model.SearchRes)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &md); err != nil {
			log.Error("CacheSearchMD json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[md.ID] = md
	}
	return
}

// AddCacheSearchMD .
func (d *Dao) AddCacheSearchMD(c context.Context, data map[int64]*model.SearchRes) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsMDs = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserGuess json.Marshal err(%v)", err)
			continue
		}
		keyID = keySearchMD(k)
		keyIDs = append(keyIDs, keyID)
		argsMDs = argsMDs.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsMDs...); err != nil {
		log.Error("AddCacheUserGuess conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.filterExpire); err != nil {
			log.Error("AddCacheUserGuess conn.Send(Expire, %s, %d) error(%v)", v, d.guessExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// CacheEpGames .
func (d *Dao) CacheEpGames(c context.Context, ids []int64) (res map[int64]*mdlEp.Game, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, gid := range ids {
		key = keyGID(gid)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheEpGames conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*mdlEp.Game, len(ids))
	for _, bs := range bss {
		game := new(mdlEp.Game)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, game); err != nil {
			log.Error("CacheEpGames json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[game.ID] = game
	}
	return
}

// AddCacheEpGames .
func (d *Dao) AddCacheEpGames(c context.Context, data map[int64]*mdlEp.Game) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsCid = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("json.Marshal err(%v)", err)
			continue
		}
		keyID = keyGID(v.ID)
		keyIDs = append(keyIDs, keyID)
		argsCid = argsCid.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsCid...); err != nil {
		log.Error("AddCacheEpGames conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.listExpire); err != nil {
			log.Error("AddCacheEpGames conn.Send(Expire, %s, %d) error(%v)", v, d.listExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// CacheEpGameMap .
func (d *Dao) CacheEpGameMap(c context.Context, oids []int64, tp int64) (res map[int64]int64, err error) {
	if len(oids) == 0 {
		return
	}
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}
		sg   []int64
	)
	defer conn.Close()
	for _, oid := range oids {
		args = args.Add(keyGMap(oid, tp))
	}
	if sg, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		log.Error("CacheEpGameMap redis.Int64s MGET(%v) error(%v)", args, err)
		return
	}
	res = make(map[int64]int64, len(oids))
	for key, val := range sg {
		if val == 0 {
			continue
		}
		res[oids[key]] = val
	}
	return
}

// AddCacheEpGameMap .
func (d *Dao) AddCacheEpGameMap(c context.Context, miss map[int64]int64, tp int64) (err error) {
	if len(miss) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	var gKey []string
	args := redis.Args{}
	for oid, gid := range miss {
		keyStr := keyGMap(oid, tp)
		args = args.Add(keyStr).Add(gid)
		gKey = append(gKey, keyStr)
	}
	var count int
	if err = conn.Send("MSET", args...); err != nil {
		log.Error("AddCacheEpGameMap redis.Int64s(conn.Do(MSET,%v) error(%v)", miss, err)
		return
	}
	count++
	for _, v := range gKey {
		if err = conn.Send("EXPIRE", v, d.listExpire); err != nil {
			log.Error("AddCacheEpGameMap EXPIRE %v error(%v)", miss, err)
			return
		}
		count++
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheEpGameMap Flush %v error(%v)", miss, err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheEpGameMap Receive %v error(%v)", miss, err)
			return
		}
	}
	return
}

// CacheNoSeasonCont no seannn contest from cache.
func (d *Dao) CacheNoSeasonCont(c context.Context, stime, etime, ps, sort int64) (res []*mdlEp.Contest, total int, err error) {
	key := keyNoSeasonCont(stime, etime, ps, sort)
	res, total, err = d.cosPbCache(c, key)
	return
}

// AddCacheNoSeasonCont  set not season contest to cache.
func (d *Dao) AddCacheNoSeasonCont(c context.Context, stime, etime, ps, sort int64, data []*mdlEp.Contest, total int) (err error) {
	key := keyNoSeasonCont(stime, etime, ps, sort)
	err = d.setCosPbCache(c, key, data, total)
	return
}

// CacheSeasonCont season contest from cache.
func (d *Dao) CacheSeasonCont(c context.Context, sid, stime, etime, ps, sort int64) (res []*mdlEp.Contest, total int, err error) {
	key := keySeasonCont(sid, stime, etime, ps, sort)
	res, total, err = d.cosPbCache(c, key)
	return
}

// AddCacheNSeasonCont  set season contest to cache.
func (d *Dao) AddCacheSeasonCont(c context.Context, sid, stime, etime, ps, sort int64, data []*mdlEp.Contest, total int) (err error) {
	key := keySeasonCont(sid, stime, etime, ps, sort)
	err = d.setCosPbCache(c, key, data, total)
	return
}

func (d *Dao) cosPbCache(c context.Context, key string) (res []*mdlEp.Contest, total int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cont := &mdlEp.Contest{}
		if err = json.Unmarshal(bs, cont); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cont)
	}
	total = from(num)
	return
}

func (d *Dao) setCosPbCache(c context.Context, key string, contests []*mdlEp.Contest, total int) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, contest := range contests {
		bs, _ := json.Marshal(contest)
		args = args.Add(combine(int64(sort), total)).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.filterExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.filterExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// AddH5Games.
func (d *Dao) AddH5Games(c context.Context, data []*model.GameRank) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	key := _gameRank
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, game := range data {
		bs, _ := json.Marshal(game)
		args = args.Add(sort).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.listExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// H5Games.
func (d *Dao) H5Games(c context.Context) (res []*model.GameRank, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := _gameRank
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		game := &model.GameRank{}
		if err = json.Unmarshal(bs, game); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, game)
	}
	return
}

// CacheSeasonGames .
func (d *Dao) CacheSeasonGames(c context.Context) (res []int64, err error) {
	var (
		bs   []byte
		key  = _seasonGame
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheSearchMainIDs .
func (d *Dao) AddCacheSeasonGames(c context.Context, data []int64) (err error) {
	var (
		bs   []byte
		key  = _seasonGame
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.listExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.listExpire, bs)
		return
	}
	return
}

// AddCacheSeasonRank.
func (d *Dao) AddCacheSeasonRank(c context.Context, gid int64, data []*model.SeasonRank) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	count := 0
	key := keySRank(gid)
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for sort, season := range data {
		bs, _ := json.Marshal(season)
		args = args.Add(sort).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.listExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.listExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// CacheSeasonRank.
func (d *Dao) CacheSeasonRank(c context.Context, gid int64) (res []*model.SeasonRank, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keySRank(gid)
	values, err := redis.Values(conn.Do("ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		season := &model.SeasonRank{}
		if err = json.Unmarshal(bs, season); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, season)
	}
	return
}

// make sure that only one pod can lock
func (d *Dao) IncrActivityPodIndex(deployID string) (index int64, err error) {
	conn := d.redis.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := fmt.Sprintf(cacheKey4DeployIDPod, deployID)
	reply, incrErr := conn.Do("INCR", cacheKey)
	if incrErr != nil {
		err = incrErr

		return
	}

	if d, ok := reply.(int64); ok {
		index = d
	}

	// set expire as 1 day
	if index == 1 {
		_, _ = conn.Do("EXPIRE", cacheKey, 86400)
	}

	return
}

func keyTeamsInSeason(sid int64) string {
	return fmt.Sprintf(_keyTeamInSeason, sid)
}
func (d *Dao) AddTeamsInSeasonToCache(c context.Context, seasonId int64, teams []*model.TeamInSeason) (err error) {
	var (
		bs   []byte
		key  = keyTeamsInSeason(seasonId)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err = json.Marshal(teams)
	if err != nil {
		return
	}
	expire := int32(tool.CalculateExpiredSeconds(1))
	for i := 0; i <= 3; i++ {
		if _, err = conn.Do("SETEX", key, expire, bs); err == nil {
			break
		}
	}
	if err != nil {
		tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4TeamsInSeason, tool.CacheOfRemote}...).Inc()
		log.Errorc(c, "setex SEND key %s error(%v)", key, err)
	}

	return
}

func (d *Dao) GetTeamsInSeasonFromCache(c context.Context, seasonId int64) (teams []*model.TeamInSeason, err error) {
	var (
		bs   []byte
		key  = keyTeamsInSeason(seasonId)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		log.Errorc(c, "conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	teams = make([]*model.TeamInSeason, 0)
	if err = json.Unmarshal(bs, &teams); err != nil {
		log.Errorc(c, "teams Unmarshal error(%v)", err)
		return
	}
	return
}

// CacheFetchSeasonsByMatchId.
func (d *Dao) CacheFetchSeasonsByMatchId(c context.Context, matchID int64) (seasons []*model.MatchSeason, err error) {
	var (
		bs   []byte
		key  = keyMatchSeason(matchID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		log.Errorc(c, "MatchSeasonsInfo CacheFetchSeasonsByMatchId conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	seasons = make([]*model.MatchSeason, 0)
	if err = json.Unmarshal(bs, &seasons); err != nil {
		log.Errorc(c, "MatchSeasonsInfo  CacheFetchSeasonsByMatchId key(%s) Unmarshal error(%v)", key, err)
		return
	}
	return
}

// AddCacheFetchSeasonsByMatchId.
func (d *Dao) AddCacheFetchSeasonsByMatchId(c context.Context, matchID int64, seasons []*model.MatchSeason) (err error) {
	var (
		bs   []byte
		key  = keyMatchSeason(matchID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err = json.Marshal(seasons)
	if err != nil {
		return
	}
	expire := int32(tool.CalculateExpiredSeconds(10))
	for i := 0; i <= 3; i++ {
		if _, err = conn.Do("SETEX", key, expire, bs); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "MatchSeasonsInfo  AddCacheFetchSeasonsByMatchId SETEX key(%s) error(%v)", key, err)
	}

	return
}

// DelCacheSeasonsByMatchId.
func (d *Dao) DelCacheSeasonsByMatchId(ctx context.Context, matchID int64) (err error) {
	var (
		key = keyMatchSeason(matchID)
	)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(ctx, "MatchSeasonsInfo DelCacheSeasonsByMatchId conn.Do(DEL key(%s) error(%v))", key, err)
		return
	}
	return
}

// CacheFetchSeasonsInfoMap .
func (d *Dao) CacheFetchSeasonsInfoMap(ctx context.Context, sids []int64) (res map[int64]*model.MatchSeason, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, sid := range sids {
		key = keySeasonInfo(sid)
		args = args.Add(key)
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(ctx, "CacheFetchSeasonsInfoMap conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*model.MatchSeason, len(sids))
	for _, bs := range bss {
		season := new(model.MatchSeason)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &season); err != nil {
			log.Errorc(ctx, "CacheFetchSeasonsInfoMap json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[season.SeasonID] = season
	}
	return
}

// AddCacheFetchSeasonsInfoMap .
func (d *Dao) AddCacheFetchSeasonsInfoMap(ctx context.Context, data map[int64]*model.MatchSeason) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsMDs = redis.Args{}
	)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	for k, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheFetchSeasonsInfoMap json.Marshal err(%v)", err)
			continue
		}
		keyID = keySeasonInfo(k)
		keyIDs = append(keyIDs, keyID)
		argsMDs = argsMDs.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsMDs...); err != nil {
		log.Errorc(ctx, "AddCacheFetchSeasonsInfoMap conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	expire := int32(tool.CalculateExpiredSeconds(10))
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, expire); err != nil {
			log.Errorc(ctx, "AddCacheFetchSeasonsInfoMap conn.Send(Expire, %s, %d) error(%v)", v, d.guessExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheFetchSeasonsInfoMap conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheFetchSeasonsInfoMap conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelCacheSeasonInfoByID.
func (d *Dao) DelCacheSeasonInfoByID(ctx context.Context, sid int64) (err error) {
	var (
		key = keySeasonInfo(sid)
	)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheSeasonInfoByID conn.Do(DEL key(%s) error(%v))", key, err)
		return
	}
	return
}

// AddCacheVideoList .
func keyVideoList(id int64) string {
	return fmt.Sprintf(_keyVideoList, id)
}
func (d *Dao) AddCacheVideoList(c context.Context, id int64, videoList *model.VideoListInfo) (err error) {
	var (
		bs   []byte
		key  = keyVideoList(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err = json.Marshal(videoList)
	if err != nil {
		return
	}
	expire := int32(tool.CalculateExpiredSeconds(1))
	for i := 0; i <= 3; i++ {
		if _, err = conn.Do("SETEX", key, expire, bs); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(c, "AddCacheVideoList setex SEND key %s error(%v)", key, err)
	}

	return
}

// CacheVideoList .
func (d *Dao) CacheVideoList(c context.Context, id int64) (videoList *model.VideoListInfo, err error) {
	var (
		bs   []byte
		key  = keyVideoList(id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			videoList = nil
		} else {
			log.Errorc(c, "CacheVideoList conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	videoList = new(model.VideoListInfo)
	if err = json.Unmarshal(bs, &videoList); err != nil {
		log.Errorc(c, "CacheVideoList Unmarshal error(%v)", err)
		return
	}
	return
}

func (d *Dao) delVideoListCache(c context.Context, id int64) (err error) {
	key := keyVideoList(id)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "DelVideoListCache conn.Do(DEL keyVideoList(%s) error(%v))", key, err)
		return
	}
	return
}

func (d *Dao) DelVideoListCacheKey(ctx context.Context, id int64) (err error) {
	if err = retry.WithAttempts(ctx, "video_list_component_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return d.delVideoListCache(ctx, id)
	}); err != nil {
		log.Errorc(ctx, "DelVideoListCacheKey id(%d) error(%+v)", id, err)
		return err
	}
	return
}
