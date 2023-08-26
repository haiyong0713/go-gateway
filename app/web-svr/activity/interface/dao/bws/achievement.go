package bws

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"go-common/library/sync/errgroup.v2"
)

const (
	_userAchieveKeyFmt     = "bws_u_a_%d_%s"
	_userRankKeyFmt        = "bws_u_r_%d_%s"
	_userAchieveCntFmt     = "bws_a_c_%d_%s"
	_bwsLotteryKeyFmt      = "bws_l_%s_%d"
	_bwsAchievePointFmt    = "bws_a_pit_%d_%s"
	_bwsAchieveRankFmt     = "bws_a_r_%d"
	_bwsAllAchieveRankFmt  = "bws_al_a_r_%d"
	_bwsAllAchievePointFmt = "bws_cnt_pt_%d_%d"
	_rankMaxTime           = 2145888000
	_awardSQL              = "UPDATE act_bws_user_achieves SET award = 2 where `key`= ? AND aid = ?"
	_achievementsSQL       = "SELECT id,`name`,icon,dic,link_type,`unlock`,bid,icon_big,icon_active,icon_active_big,award,ctime,mtime,image,suit_id,`achieve_point`,`level`,`extra_type` FROM act_bws_achievements WHERE del = 0  AND bid = ? ORDER BY ID"
	_userAchieveSQL        = "SELECT id,aid,award FROM act_bws_user_achieves WHERE bid = ? AND `key` = ? AND del = 0"
	_usersAchieveSQL       = "SELECT id,aid,award,`key` FROM act_bws_user_achieves WHERE bid = ? AND `key` IN (%s) AND del = 0"
	_userLastAchieveSQL    = "SELECT `id`,`ctime` FROM `act_bws_user_achieves` WHERE (%s) AND `del` = 0 ORDER BY `id` DESC LIMIT 1"
	_userAchieveAddSQL     = "INSERT INTO act_bws_user_achieves(bid,aid,award,`key`) VALUES(?,?,?,?)"
	_countAchievesSQL      = "SELECT aid,COUNT(1) AS c FROM act_bws_user_achieves WHERE del = 0 AND bid = ? AND ctime BETWEEN ? AND ? GROUP BY aid HAVING c > 0"
	_nextDayHour           = 16
	_maxAchieveTop         = 1000
	_bwsData               = 2019
	_userUserDetailData    = "bws_u_d_%d_%d_%s"
	_midCoupon             = "bws_u_c_%d_%d"
	_userRankDetail        = "bws_rank_detail_%d_%s"
	_userUserRankData      = "bws_rank_limit_%d_%s"
)

func keyUserAchieve(bid int64, key string) string {
	return fmt.Sprintf(_userAchieveKeyFmt, bid, key)
}

func keyUserDetail(bid, mid int64, date string) string {
	return fmt.Sprintf(_userUserDetailData, bid, mid, date)
}

func keyMidCoupon(bid, mid int64) string {
	return fmt.Sprintf(_midCoupon, bid, mid)
}

func keyBwsRank(bid int64, date string) string {
	return fmt.Sprintf(_userRankDetail, bid, date)
}

func keyUserRankLimit(bid int64, date string) string {
	return fmt.Sprintf(_userUserRankData, bid, date)
}
func keyAchieveCnt(bid int64, day string) string {
	return fmt.Sprintf(_userAchieveCntFmt, bid, day)
}

func keyLottery(aid int64, day string) string {
	if day == "" {
		day = time.Now().Format("20060102")
	}
	return fmt.Sprintf(_bwsLotteryKeyFmt, day, aid)
}

func achievePoint(bid int64, ukey string) string {
	return fmt.Sprintf(_bwsAchievePointFmt, bid, ukey)
}

func achieveRank(bid int64) string {
	return fmt.Sprintf(_bwsAchieveRankFmt, bid)
}

func allAchieveRank() string {
	return fmt.Sprintf(_bwsAllAchieveRankFmt, _bwsData)
}

func allAchievePoint(mid int64) string {
	return fmt.Sprintf(_bwsAllAchievePointFmt, _bwsData, mid)
}

func bwsLotteryKey(bid, awardID int64) string {
	return fmt.Sprintf("bws_lott_%d_%d", bid, awardID)
}

func bwsSpecLotteryKey(bid int64) string {
	return fmt.Sprintf("bws_spec_lott_%d", bid)
}

func keyUserRank(bid int64, date string) string {
	return fmt.Sprintf(_userRankKeyFmt, bid, date)
}

// rankMaxTime .
func rankMaxTime() int64 {
	return _rankMaxTime
}

// buildAchiveRank only for num lower 1 million.
func buildAchiveRank(num int64, ctime int64) int64 {
	var (
		maxTimeStamp = rankMaxTime()
		timeScore    = maxTimeStamp - ctime
		rankScore    int64
	)
	rankScore = (num & 0xFFFFF) << 32
	rankScore |= timeScore & 0xFFFFFFFF
	return rankScore
}

// Award  achievement award
func (d *Dao) Award(c context.Context, key string, aid int64) (err error) {
	if _, err = d.db.Exec(c, _awardSQL, key, aid); err != nil {
		log.Error("Award: db.Exec(%d,%s) error(%v)", aid, key, err)
	}
	return
}

// LastAchievements .
func (d *Dao) LastAchievements(c context.Context, bidUkey map[int64]string) (res *bwsmdl.Achievement, err error) {
	var (
		bidStr []string
		args   []interface{}
	)
	if len(bidUkey) == 0 {
		return
	}
	for k, v := range bidUkey {
		bidStr = append(bidStr, "(`bid` = ? AND `key` = ?)")
		args = append(args, k, v)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_userLastAchieveSQL, strings.Join(bidStr, " OR ")), args...)
	res = &bwsmdl.Achievement{}
	if err = row.Scan(&res.ID, &res.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// RawAchievements  achievements list
func (d *Dao) RawAchievements(c context.Context, bid int64) (res *bwsmdl.Achievements, err error) {
	var (
		rows *xsql.Rows
		rs   []*bwsmdl.Achievement
	)
	if rows, err = d.db.Query(c, _achievementsSQL, bid); err != nil {
		log.Error("RawAchievements: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.Achievement)
		if err = rows.Scan(&r.ID, &r.Name, &r.Icon, &r.Dic, &r.LockType, &r.Unlock, &r.Bid, &r.IconBig, &r.IconActive, &r.IconActiveBig, &r.Award, &r.Ctime, &r.Mtime, &r.Image, &r.SuitID, &r.AchievePoint, &r.Level, &r.ExtraType); err != nil {
			log.Error("RawAchievements:row.Scan() error(%v)", err)
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		return
	}
	if len(rs) > 0 {
		res = new(bwsmdl.Achievements)
		res.Achievements = rs
	}
	return
}

// RawUserAchieves get user achievements from db.
func (d *Dao) RawUserAchieves(c context.Context, bid int64, key string) (rs []*bwsmdl.UserAchieve, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _userAchieveSQL, bid, key); err != nil {
		log.Error("RawUserAchieves: db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.UserAchieve)
		if err = rows.Scan(&r.ID, &r.Aid, &r.Award); err != nil {
			log.Error("RawUserAchieves:row.Scan() error(%v)", err)
			return
		}
		rs = append(rs, r)
	}
	err = rows.Err()
	return
}

// RawUsersAchieves get user achievements from db.
func (d *Dao) RawUsersAchieves(c context.Context, bid int64, ukeys []string) (rs map[string][]*bwsmdl.UserAchieve, err error) {
	var (
		rows    *xsql.Rows
		ukesLen = len(ukeys)
		args    = make([]interface{}, 0)
		str     []string
	)
	if ukesLen == 0 {
		return
	}
	args = append(args, bid)
	for _, v := range ukeys {
		str = append(str, "?")
		args = append(args, v)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_usersAchieveSQL, strings.Join(str, ",")), args...); err != nil {
		log.Error("RawUserAchieves: db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	rs = make(map[string][]*bwsmdl.UserAchieve)
	for rows.Next() {
		r := new(bwsmdl.UserAchieve)
		if err = rows.Scan(&r.ID, &r.Aid, &r.Award, &r.Key); err != nil {
			log.Error("RawUserAchieves:row.Scan() error(%v)", err)
			return
		}
		rs[r.Key] = append(rs[r.Key], r)
	}
	err = rows.Err()
	return
}

// RawAchieveCounts  achievements user count
func (d *Dao) RawAchieveCounts(c context.Context, bid int64, day string) (res []*bwsmdl.CountAchieves, err error) {
	var (
		rows *xsql.Rows
	)
	start, _ := time.Parse("20060102-15:04:05", day+"-00:00:00")
	end, _ := time.Parse("20060102-15:04:05", day+"-23:59:59")
	if rows, err = d.db.Query(c, _countAchievesSQL, bid, start, end); err != nil {
		log.Error("RawCountAchieves: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.CountAchieves)
		if err = rows.Scan(&r.Aid, &r.Count); err != nil {
			log.Error("RawCountAchieves:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// AddUserAchieve .
func (d *Dao) AddUserAchieve(c context.Context, bid, aid, award int64, key string) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _userAchieveAddSQL, bid, aid, award, key); err != nil {
		log.Error("AddUserAchieve error d.db.Exec(%d,%d,%s) error(%v)", bid, aid, key, err)
		return
	}
	return res.LastInsertId()
}

// CacheUserAchieves .
func (d *Dao) CacheUserAchieves(c context.Context, bid int64, key string) (res []*bwsmdl.UserAchieve, err error) {
	var (
		values   []interface{}
		cacheKey = keyUserAchieve(bid, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, 0, -1)); err != nil {
		log.Error("CacheUserAchieves conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("CacheUserAchieves redis.Scan(%v) error(%v)", values, err)
			return
		}
		item := new(bwsmdl.UserAchieve)
		if err = json.Unmarshal(bs, &item); err != nil {
			log.Error("CacheUserAchieves conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
			continue
		}
		res = append(res, item)
	}
	return
}

// AddCacheUserAchieves .
func (d *Dao) AddCacheUserAchieves(c context.Context, bid int64, data []*bwsmdl.UserAchieve, key string) (err error) {
	var bs []byte
	if len(data) == 0 {
		return
	}
	cacheKey := keyUserAchieve(bid, key)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(cacheKey)
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserAchieves json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUserAchieves conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userAchExpire); err != nil {
		log.Error("AddCacheUserAchieves conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserAchieves conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserAchieves conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelCacheUserAchieves .
func (d *Dao) DelCacheUserAchieves(c context.Context, bid int64, key string) (err error) {
	cacheKey := keyUserAchieve(bid, key)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheUserAchieves conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// AppendUserAchievesCache .
func (d *Dao) AppendUserAchievesCache(c context.Context, bid int64, key string, achieve *bwsmdl.UserAchieve) (err error) {
	var (
		bs       []byte
		ok       bool
		cacheKey = keyUserAchieve(bid, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", cacheKey, d.userAchExpire)); err != nil || !ok {
		log.Error("AppendUserAchievesCache conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(cacheKey)
	if bs, err = json.Marshal(achieve); err != nil {
		log.Error("AppendUserAchievesCache json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(achieve.ID).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUserAchieves conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userAchExpire); err != nil {
		log.Error("AddCacheUserAchieves conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserAchieves conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserAchieves conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheAchieveCounts  get achieve counts from cache
func (d *Dao) CacheAchieveCounts(c context.Context, bid int64, day string) (res []*bwsmdl.CountAchieves, err error) {
	var (
		bss  []int64
		key  = keyAchieveCnt(bid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bss, err = redis.Int64s(conn.Do("HGETALL", key)); err != nil {
		log.Error("CacheAchieveCounts conn.Do(HGETALL,%s) error(%v)", key, err)
		return
	}
	for i := 1; i < len(bss); i += 2 {
		item := &bwsmdl.CountAchieves{Aid: bss[i-1], Count: bss[i]}
		res = append(res, item)
	}
	return
}

// AddCacheAchieveCounts set achieve counts  to cache
func (d *Dao) AddCacheAchieveCounts(c context.Context, bid int64, res []*bwsmdl.CountAchieves, day string) (err error) {
	if len(res) == 0 {
		return
	}
	key := keyAchieveCnt(bid, day)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddCacheAchieveCounts conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range res {
		args = args.Add(v.Aid).Add(v.Count)
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Error("AddCacheAchieveCounts conn.Send(HMSET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.achCntExpire); err != nil {
		log.Error("AddCacheAchieveCounts conn.Send(Expire, %s, %d) error(%v)", key, d.achCntExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheAchieveCounts conn.Flush error(%v)", err)
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

// IncrCacheAchieveCounts incr achieve counts  to cache
func (d *Dao) IncrCacheAchieveCounts(c context.Context, bid, aid int64, day string) (err error) {
	var (
		key  = keyAchieveCnt(bid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("EXPIRE", key, d.achCntExpire); err != nil {
		log.Error("IncrCacheAchieveCounts conn.Send(Expire, %s, %d) error(%v)", key, d.achCntExpire, err)
		return
	}
	if err = conn.Send("HINCRBY", key, aid, 1); err != nil {
		log.Error("IncrCacheAchieveCounts conn.Send(HMSET, %s, %d) error(%v)", key, aid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("IncrCacheAchieveCounts conn.Flush error(%v)", err)
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

// DelCacheAchieveCounts delete achieve cnt cache.
func (d *Dao) DelCacheAchieveCounts(c context.Context, bid int64, day string) (err error) {
	cacheKey := keyAchieveCnt(bid, day)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheAchieveCounts conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// AddLotteryMidCache add lottery mid cache.
func (d *Dao) AddLotteryMidCache(c context.Context, aid, mid int64) (err error) {
	now := time.Now()
	hour := now.Hour()
	dayInt, _ := strconv.ParseInt(now.Format("20060102"), 10, 64)
	if hour >= _nextDayHour {
		dayInt = dayInt + 1
	}
	cacheKey := keyLottery(aid, strconv.FormatInt(dayInt, 10))
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SADD", cacheKey, mid); err != nil {
		log.Error("AddLotteryCache conn.Do(LPUSH, %s, %d) error(%v)", cacheKey, mid, err)
	}
	return
}

// CacheLotteryMid .
func (d *Dao) CacheLotteryMid(c context.Context, aid int64, day string) (mid int64, err error) {
	var (
		cacheKey = keyLottery(aid, day)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if mid, err = redis.Int64(conn.Do("SPOP", cacheKey)); err != nil && err != redis.ErrNil {
		log.Error("LotteryMidCache SPOP key(%s) error(%v)", cacheKey, err)
	}
	return
}

// CacheLotteryMids .
func (d *Dao) CacheLotteryMids(c context.Context, aid int64, day string) (mids []int64, err error) {
	var cacheKey string
	conn := d.redis.Get(c)
	defer conn.Close()
	cacheKey = keyLottery(aid, day)
	if mids, err = redis.Int64s(conn.Do("SMEMBERS", cacheKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("redis.Int64s(conn.Do(SMEMEBERS,%s)) error(%v)", cacheKey, err)
	}
	return
}

// CacheLotteryV1 .
func (d *Dao) CacheLotteryV1(c context.Context, bid, awardID int64) (data []*bwsmdl.LotteryCache, err error) {
	var (
		cacheKey string
		bs       []byte
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	cacheKey = bwsLotteryKey(bid, awardID)
	if bs, err = redis.Bytes(conn.Do("GET", cacheKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("CacheLotteryV1 redis.String(conn.Do(GET,%s)) error(%v)", cacheKey, err)
		return
	}
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("CacheLotteryV1 json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

func (d *Dao) CacheLotterySpec(c context.Context, bid int64) (data []*bwsmdl.LotteryCache, err error) {
	var (
		cacheKey string
		bs       []byte
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	cacheKey = bwsSpecLotteryKey(bid)
	if bs, err = redis.Bytes(conn.Do("GET", cacheKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("CacheLotteryV1 redis.String(conn.Do(GET,%s)) error(%v)", cacheKey, err)
		return
	}
	item := new(bwsmdl.LotteryCache)
	if err = json.Unmarshal(bs, &item); err != nil {
		log.Error("CacheLotterySpec json.Unmarshal(%s) error(%v)", string(bs), err)
		return
	}
	data = []*bwsmdl.LotteryCache{item}
	return
}

// DelCacheAchievesPoint .
func (d *Dao) DelCacheAchievesPoint(c context.Context, bid int64, ukey []string) (err error) {
	if len(ukey) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range ukey {
		args = args.Add(achievePoint(bid, v))
	}
	if _, err = conn.Do("DEL", args...); err != nil {
		log.Error("DelCacheAchievesPoint redis.Ints(conn.Do(DEL,%v) error(%v)", args, err)
	}
	return
}

// AchievesPoint get data from cache if miss will call source method, then add to cache.
func (d *Dao) AchievesPoint(c context.Context, bid int64, ukey []string) (res map[string]int64, err error) {
	if len(ukey) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheAchievesPoint(c, bid, ukey); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []string
	for _, key := range ukey {
		if _, ok := res[key]; !ok {
			miss = append(miss, key)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[string]int64
	missData, err = d.RawAchievesPoint(c, bid, miss)
	if res == nil {
		res = make(map[string]int64, len(ukey))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.AddCacheAchievesPoint(c, bid, missData)
	return
}

// CacheAchievesPoint 获取用户成就点数.
func (d *Dao) CacheAchievesPoint(c context.Context, bid int64, ukey []string) (list map[string]int64, err error) {
	var ss []int64
	if len(ukey) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range ukey {
		args = args.Add(achievePoint(bid, v))
	}
	if ss, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheAchievesPoint redis.Ints(conn.Do(MGET,%v) error(%v)", args, err)
		}
		return
	}
	list = make(map[string]int64, len(ukey))
	for key, val := range ss {
		if val == 0 {
			continue
		}
		list[ukey[key]] = val
	}
	return
}

// AddCacheAchievesPoint 设置用户成就点数(回源时使用).
func (d *Dao) AddCacheAchievesPoint(c context.Context, bid int64, miss map[string]int64) (err error) {
	if len(miss) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for k, v := range miss {
		args = args.Add(achievePoint(bid, k)).Add(v)
	}
	if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
		log.Error("AddCacheAchievesPoint redis.Ints(conn.Do(MSET,%v) error(%v)", args, err)
	}
	return
}

// RawAchievesPoint .
func (d *Dao) RawAchievesPoint(c context.Context, bid int64, ukeys []string) (list map[string]int64, err error) {
	var (
		userAchieve map[string][]*bwsmdl.UserAchieve
		achieve     *bwsmdl.Achievements
		mapAchieve  map[int64]*bwsmdl.Achievement
	)
	if len(ukeys) == 0 {
		return
	}
	// 获取个人成就
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) (e error) {
		if userAchieve, e = d.RawUsersAchieves(ctx, bid, ukeys); e != nil {
			log.Error("d.RawUsersAchieves(%d),%v error(%v)", bid, ukeys, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if achieve, e = d.RawAchievements(ctx, bid); e != nil {
			log.Error("d.RawAchievements(%d),%v error(%v)", bid, ukeys, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	// 啥也没有
	if len(userAchieve) == 0 || achieve == nil || len(achieve.Achievements) == 0 {
		return
	}
	mapAchieve = make(map[int64]*bwsmdl.Achievement)
	for _, v := range achieve.Achievements {
		mapAchieve[v.ID] = v
	}
	list = make(map[string]int64)
	for k, achi := range userAchieve {
		achieveScore := int64(0)
		for _, val := range achi {
			if _, ok := mapAchieve[val.Aid]; ok {
				achieveScore += mapAchieve[val.Aid].AchievePoint
			}
		}
		list[k] = achieveScore
	}
	return
}

// RawCompositeAchievesPoint 计算bws2019总场排名 has check.
func (d *Dao) RawCompositeAchievesPoint(c context.Context, mids []int64) (list map[int64]int64, err error) {
	var (
		scoreSlice []map[int64]int64
	)
	if len(mids) == 0 {
		return
	}
	eg := errgroup.WithContext(c)
	for _, v := range d.bws2019 {
		tempBid := v
		// 分别广州，上海，成都 mids下对应的score
		eg.Go(func(ctx context.Context) error {
			var (
				ukeys   []string
				midsMap map[string]int64
				ss      map[int64]int64
			)
			user, e := d.RawUsersMids(ctx, tempBid, mids)
			if e != nil {
				log.Error("d.UsersMids(%d,%v) error(%v)", tempBid, mids, e)
				return nil
			}
			midsMap = make(map[string]int64)
			for _, val := range user {
				if val.Key == "" {
					continue
				}
				ukeys = append(ukeys, val.Key)
				midsMap[val.Key] = val.Mid

			}
			listPoint, et := d.RawAchievesPoint(ctx, tempBid, ukeys)
			if et != nil {
				log.Error("d.RawAchievesPoint(%d,%v) error(%v)", tempBid, ukeys, e)
				return nil
			}
			// 获取score
			ss = make(map[int64]int64)
			for uk, lt := range listPoint {
				if _, ok := midsMap[uk]; !ok {
					continue
				}
				ss[midsMap[uk]] = lt
			}
			scoreSlice = append(scoreSlice, ss)
			return nil
		})
	}
	eg.Wait()
	// 汇总mid下的score
	list = make(map[int64]int64)
	for _, val := range scoreSlice {
		for m, vt := range val {
			list[m] += vt
		}
	}
	return
}

// CacheAchievesRank .
func (d *Dao) CacheAchievesRank(c context.Context, bid int64, num, ty int) (list []int64, err error) {
	var (
		key  string
		conn = d.redis.Get(c)
	)
	if ty == bwsmdl.CompositeRankType {
		key = allAchieveRank()
	} else {
		key = achieveRank(bid)
	}
	defer conn.Close()
	if list, err = redis.Int64s(conn.Do("ZREVRANGE", key, 0, num)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheAchievesRank conn.Do(ZREVRANGE,%s,%d) error(%v)", key, num, err)
		}
	}
	return
}

// AddCacheAchievesRank 排行榜回源逻辑.
func (d *Dao) AddCacheAchievesRank(c context.Context, bid int64, miss map[int64]*bwsmdl.RankAchieve, ty int) (err error) {
	var (
		ss   bool
		ukey string
	)
	if len(miss) == 0 {
		return
	}
	if ty == bwsmdl.CompositeRankType {
		ukey = allAchieveRank()
	} else {
		ukey = achieveRank(bid)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(ukey)
	for k, v := range miss {
		if v != nil {
			score := buildAchiveRank(v.Num, v.Ctime)
			args = args.Add(score).Add(k)
		}
	}
	if ss, err = redis.Bool(conn.Do("ZADD", args...)); err != nil || !ss {
		log.Error("AddCacheAchievesRank redis.Ints(conn.Do(ZADD,%v) error(%v)", args, err)
	}
	return
}

// isBws2019s bid是否属于广州，上海，成都bws
func (d *Dao) IsBws2019s(bid int64) bool {
	for _, v := range d.bws2019 {
		if bid == v {
			return true
		}
	}
	return false
}

// IncrSingleAchievesPoint 更新成就排名和成就分数.
func (d *Dao) IncrSingleAchievesPoint(c context.Context, bid, mid, score, ctime int64, ukey string, initScore bool) (err error) {
	var (
		// 用户成就点数缓存
		cacheKey = achievePoint(bid, ukey)
		// 当场次成就排行
		rankKey = achieveRank(bid)
		conn    = d.redis.Get(c)
		count   int64
	)

	defer conn.Close()
	if initScore {
		if _, err = conn.Do("SET", cacheKey, score); err != nil {
			log.Error("incrSingleAchivePoint conn.Do(SET) key(%s) error(%v)", cacheKey, err)
			return
		}
		count = score
	} else {
		if count, err = redis.Int64(conn.Do("INCRBY", cacheKey, score)); err != nil {
			log.Error("incrSingleAchivePoint conn.Do(INCRBY) key(%s) error(%v)", cacheKey, err)
			return
		}
	}
	if mid <= 0 {
		return
	}
	rankScore := buildAchiveRank(count, ctime)
	if err = conn.Send("ZADD", rankKey, rankScore, mid); err != nil {
		log.Error("incrSingleAchivePoint conn.Do(ZADD) key(%s) error(%v)", rankKey, err)
		return
	}
	if err = conn.Send("ZREMRANGEBYRANK", rankKey, 0, -(_maxAchieveTop + 1)); err != nil {
		log.Error("incrSingleAchivePoint conn.Do(ZREMRANGEBYRANK) key(%s) error(%v)", rankKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("incrSingleAchivePoint conn.Receive() (%d) error(%v)", i, err)
			return
		}
	}
	return
}

// IncrAchievesPoint bws2019 成就总排行 .
func (d *Dao) IncrAchievesPoint(c context.Context, bid, mid, score, ctime int64, initScore bool) (err error) {
	var (
		// 用户总场次成就点数
		allCacheKey = allAchievePoint(mid)
		// 用户总场次排行
		allRankKey = allAchieveRank()
		conn       = d.redis.Get(c)
		allCount   int64
	)
	defer conn.Close()
	if isBws := d.IsBws2019s(bid); !isBws || mid <= 0 {
		return
	}
	if initScore {
		if _, err = conn.Do("SET", allCacheKey, score); err != nil {
			log.Error("IncrAchievesPoint conn.Do（SET) all key(%s) error(%v)", allCacheKey, err)
			return
		}
		allCount = score
	} else {
		if allCount, err = redis.Int64(conn.Do("INCRBY", allCacheKey, score)); err != nil {
			log.Error("IncrAchievesPoint conn.Do(INCRBY) all key(%s) error(%v)", allCacheKey, err)
			return
		}
	}
	allRankScore := buildAchiveRank(allCount, ctime)
	if err = conn.Send("ZADD", allRankKey, allRankScore, mid); err != nil {
		log.Error("IncrAchievesPoint allRankKey conn.Do(ZADD) key(%s) error(%v)", allRankKey, err)
		return
	}
	if err = conn.Send("ZREMRANGEBYRANK", allRankKey, 0, -(_maxAchieveTop + 1)); err != nil {
		log.Error("IncrAchievesPoint conn.Do(ZREMRANGEBYRANK) allRankKey(%s) error(%v)", allRankKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("IncrAchievesPoint conn.Receive() (%d) error(%v)", i, err)
			return
		}
	}
	return
}

// GetAchieveRank 获取用户当前bws成就排行.
func (d *Dao) GetAchieveRank(c context.Context, bid int64, mid int64, ty int) (rank int, err error) {
	var cacheKey string
	if ty == bwsmdl.CompositeRankType {
		cacheKey = allAchieveRank()
	} else {
		cacheKey = achieveRank(bid)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if rank, err = redis.Int(conn.Do("ZREVRANK", cacheKey, mid)); err != nil {
		// 默认值
		rank = bwsmdl.DefaultRank
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("GetAchiveRank conn.Do(ZREVRANK) error(%v)", err)
		}
	}
	return
}

// DelAllAchieveRank 总场排行版删除.
func (d *Dao) DelAllAchieveRank(c context.Context) (err error) {
	cacheKey := allAchieveRank()
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelAllAchieveRank conn.Do(DEL) error(%v)", err)
	}
	return
}

// DelAchieveRank 排行版数据删除.
func (d *Dao) DelAchieveRank(c context.Context, bid int64) (err error) {
	cacheKey := achieveRank(bid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelAchieveRank conn.Do(DEL) error(%v)", err)
	}
	return
}

// DelCacheCompositeAchievesPoint .
func (d *Dao) DelCacheCompositeAchievesPoint(c context.Context, mids []int64) (err error) {
	if len(mids) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range mids {
		args = args.Add(allAchievePoint(v))
	}
	if _, err = conn.Do("DEL", args...); err != nil {
		log.Error("DelCacheCompositeAchievesPoint redis.Ints(conn.Do(MGET,%v) error(%v)", args, err)
	}
	return
}

// CacheCompositeAchievesPoint 获取用户成就点数.
func (d *Dao) CacheCompositeAchievesPoint(c context.Context, mids []int64) (list map[int64]int64, err error) {
	var ss []int64
	if len(mids) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range mids {
		args = args.Add(allAchievePoint(v))
	}
	if ss, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheCompositeAchievesPoint redis.Ints(conn.Do(MGET,%v) error(%v)", args, err)
		}
		return
	}
	list = make(map[int64]int64, len(mids))
	for key, val := range ss {
		if val == 0 {
			continue
		}
		list[mids[key]] = val
	}
	return
}

// AddCacheAchievesPoint 获取用户成就点数.
func (d *Dao) AddCacheCompositeAchievesPoint(c context.Context, miss map[int64]int64) (err error) {
	if len(miss) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for k, v := range miss {
		args = args.Add(allAchievePoint(k)).Add(v)
	}
	if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("AddCacheAchievesPoint redis.Ints(conn.Do(MSET,%v) error(%v)", args, err)
		}
	}
	return
}
