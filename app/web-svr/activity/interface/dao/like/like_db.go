package like

import (
	"context"
	xsql "database/sql"
	"encoding/json"
	"fmt"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/tool"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_selLikeSQL                              = "SELECT id,wid FROM likes where state = 1 AND sid = ? ORDER BY type"
	_likeSQL                                 = "SELECT id,sid,type,mid,wid,state,stick_top,ctime,mtime FROM likes WHERE id = ? and state = 1"
	_likeMoreLidSQL                          = "SELECT id,sid,type,mid,wid,state,stick_top,ctime,mtime FROM likes WHERE id > ?  order by id asc limit 1000"
	_likesBySidSQL                           = "SELECT id,sid,type,mid,wid,state,stick_top,ctime,mtime FROM likes WHERE id > ? and sid = ? and state = 1 order by id asc limit 1000"
	_likesSQL                                = "SELECT id,sid,type,mid,wid,state,stick_top,ctime,mtime FROM likes WHERE  id IN (%s) and state = 1"
	_likeListSQL                             = "SELECT id,wid,ctime FROM likes WHERE state = 1 AND sid = ?"
	_likeMaxIDSQL                            = "SELECT id FROM likes ORDER BY id DESC limit 1"
	_likeUpStateSQL                          = "UPDATE likes SET `state` = ? where id = ? and mid = ?"
	_likeInsertSQL                           = "INSERT INTO %s (`sid`,`type`,`mid`,`wid`,`state`,`stick_top`,`ctime`,`mtime`,`referer`) VALUES(?,?,?,?,?,?,?,?,?)"
	_likeCountSQL                            = "SELECT count(1) FROM likes WHERE sid = ?"
	_likeCheckSQL                            = "SELECT id,sid,type,mid,wid,state,stick_top,ctime,mtime FROM likes WHERE sid=? AND mid=?"
	_likeUserCheckSQL                        = "SELECT count(1) FROM likes WHERE sid = ? AND mid = ?"
	_activityArcsSQL                         = "SELECT id,wid,ctime,state FROM likes WHERE mid = ? and sid = ? limit 1000"
	_activityGetLidByWidSQL                  = "SELECT id FROM likes WHERE wid = ? and state = ? limit 1"
	_actRelationInfoSQL                      = "SELECT id,name,native_ids,h5_ids,web_ids,lottery_ids,reserve_ids,video_source_ids,follow_ids,season_ids,reserve_config,follow_config,season_config,favorite_info,favorite_config,mall_ids,mall_config,topic_ids,topic_config FROM act_relation_subject WHERE id = ? and state=?"
	_actRelationInfoGetHotIDsSQL             = "SELECT id FROM act_relation_subject WHERE state = %d"
	_actSubjectInfoGetHotIDsSQL              = "SELECT id FROM act_subject WHERE type in (%s) and etime >= '%s' and state = 1"
	_actSubjectReserveInfoGetHotIDsSQL       = "SELECT id FROM act_subject WHERE type=%d and etime >= '%s' and state = 1"
	_upActRelationWithSIDAndStateSQL         = "SELECT id,sid,mid,oid,type,state FROM up_act_reserve_relation WHERE sid IN (%s) and state IN (%s)"
	_upActRelationWithSIDSQL                 = "SELECT id,sid,mid,oid,type,state,live_plan_start_time,audit,audit_channel,dynamic_id,dynamic_audit,lottery_type,lottery_id,lottery_audit FROM up_act_reserve_relation WHERE sid IN (%s)"
	_upActRelationLiveExpireSQL              = "SELECT id,sid,mid,oid,type,state,live_plan_start_time,audit,audit_channel,dynamic_id FROM up_act_reserve_relation WHERE oid = '' and type = %d and state IN (%s) order by `live_plan_start_time` asc limit 100"
	_upActRelationWithStateSQLLimit          = "SELECT id,sid,mid,oid,type,state FROM up_act_reserve_relation WHERE mid = %d and type IN (%s) and state IN (%s) limit %d"
	_upActRelationPublishedSQL               = "SELECT id,sid,mid,oid,type,state FROM up_act_reserve_relation WHERE mid = %d and type IN (%s) limit %d"
	_upActRelationWithMIDAndSIDSQL           = "SELECT id,sid,mid,oid,type,state,live_plan_start_time,audit,audit_channel,lottery_type,lottery_id FROM up_act_reserve_relation WHERE mid = (%d) and sid IN (%s)"
	_upActRelationWithMIDAndSIDFromMasterSQL = "SELECT /*master*/ id,sid,mid,oid,type,state,live_plan_start_time,audit,audit_channel,lottery_type,lottery_id FROM up_act_reserve_relation WHERE sid IN (%s)"
	_upActRelationWithStateSQL               = "SELECT id,sid,mid,oid,type,state,live_plan_start_time FROM up_act_reserve_relation WHERE mid = %d and type IN (%s) and state IN (%s)"
	_upActRelationOthersWithStateSQL         = "SELECT id,sid FROM up_act_reserve_hang WHERE pub_mid = %d"
	_upActRelationWithTypeAndStateSQL        = "SELECT id,sid,mid,oid,type,state,live_plan_start_time FROM up_act_reserve_relation WHERE type IN (%s) and state IN (%s)"
	_upActRelationWithStateLimitSQL          = "SELECT id,sid,mid,oid,type,state,live_plan_start_time FROM up_act_reserve_relation WHERE mid = %d and type IN (%s) and state IN (%s) limit 1"
	_upActRelationWithOIDSQL                 = "SELECT id,sid,mid,oid,type,state FROM up_act_reserve_relation WHERE oid = ? and type = ? and state = ?"
	_upActRelationWithOIDLatestSQL           = "SELECT id,sid,mid,oid,type,state FROM up_act_reserve_relation WHERE oid = ? and type IN ($) and state IN ($) order by sid desc limit 1"
	_upActReserveRelationBindInfo            = "SELECT id,sid,oid,o_type,rid,r_type FROM up_act_reserve_relation_bind WHERE oid = %s and o_type=%d and r_type=%d ORDER BY sid desc limit 1"
	_upActReserveHangRecordInfoSQL           = "SELECT COUNT(*) FROM up_act_reserve_hang WHERE pub_mid = %d and sid = %d limit 1"
	_actSubjectUpdateStatus                  = "UPDATE act_subject set state = ? where id = ?"
	_actSubjectUpdateEtime                   = "UPDATE act_subject set etime = ? where id = ?"
	_upActRelationUpdateStateWithAudit       = "UPDATE up_act_reserve_relation set state = ?, audit = ?, audit_channel = ?,dynamic_id = ? where mid = ? and sid = ?"
	_upActRelationUpdateState                = "UPDATE up_act_reserve_relation set state = ? where mid = ? and sid = ?"
	_upActRelationUpdateStateAndOID          = "UPDATE up_act_reserve_relation set state = ?, oid = ? where mid = ? and sid = ?"
	_upActRelationBindSQL                    = "UPDATE up_act_reserve_relation set oid = ?, state = ? where sid = ?"
	_upActRelationStateSQL                   = "UPDATE up_act_reserve_relation set state = ? where sid = ?"
	_upActRelationDependAuditStateSQL        = "UPDATE up_act_reserve_relation set dynamic_audit = ?, lottery_audit = ? where sid = ?"
	_upActRelationCreateItem                 = "INSERT INTO up_act_reserve_relation (`sid`,`mid`,`oid`,`type`,`state`,`from`,`live_plan_start_time`,`audit`, `audit_channel`, `dynamic_id`, `lottery_type`, `lottery_id`, `lottery_audit`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_upActRelationItemSQL                    = "UPDATE up_act_reserve_relation set live_plan_start_time = ?, lottery_type = ?, lottery_id = ?, lottery_audit = ? where sid = ?"
	_upActSubjectAddSQL                      = "INSERT INTO act_subject (name,type,stime,etime,state) VALUES (?,?,?,?,?)"
	_upActSubjectUpdateSQL                   = "UPDATE `act_subject` set `name` = ? , `stime` = ? , `etime` = ? where `id` = ?"
	_oldTableName                            = "likes"
	_keyLikeTagFmt                           = "l_t_%d_%d"
	_keyLikeTagCntsFmt                       = "l_t_cs_%d"
	_keyLikeRegionFmt                        = "l_r_%d_%d"
	_keyLikeWidMapId                         = "l_t_wid_map_id_%d"
	_keyVoteTotalBySID                       = "act:vote:%d"
	// likeAPI ip frequence key the old is ddos:like:ip:%s
	_keyIPRequestFmt = "go:ddos:l:ip:%s"
	// the cache set of like order by ctime the old is bilibili-activity:ctime:%d
	_keyLikeListCtimeFmt      = "go:bl-a:ctime:%d"
	_keyLikeListRandomFmt     = "go:bl-n:random:%d"
	_keyLikeListRandomTypeFmt = "go:bl-n:random:type:%d:%d"
	_keyLikeStochasticFmt     = "go:bl-n:stochastic:%d"
	_keyLikeStochasticTypeFmt = "go:bl-n:stochastic:type:%d:%d"
	_keyStochasticSingleFmt   = "go:sf:stoch:%d:%d"
	_keyRandomSingleFmt       = "go:sf:random:%d:%d"
	// the cache set of like type order by ctime
	_keyLikeListTypeCtimeFmt = "go:b:a:t:%d:%d"
	// storyKing LikeAct cache
	_keyStoryDilyLikeFmt = "go:s:d:m:%s:%d:%d"
	// storyKing each likeAct cahce
	_keyStoryEachLikeFmt = "go:s:ea:m:%s:%d:%d:%d"
	// es index
	_activity = "activity"
	// storyKing LikeExtendTimes cache
	_keyStoryExtraLikeFmt = "like_extra_times_%d_%d"
	// storyKing up times cache
	_keyUpActionFmt = "like_up_times_%d_%d"
	// storyKing LikeExtendInfo cache
	_keyStoryExtendInfoFmt = "like_extend_info_%d_%s"
	// storyKing LikeExtendToken cache
	_keyStoryExtendTokenFmt = "like_extend_token_%d_%d"
	// EsOrderLikes archive center likes.
	EsOrderLikes = "likes"
	// EsOrderCoin archive center coin .
	EsOrderCoin = "coin"
	// EsOrderReply archive center reply.
	EsOrderReply = "reply"
	// EsOrderShare  archive center share.
	EsOrderShare = "share"
	// EsOrderClick archive center click
	EsOrderClick = "click"
	// EsOrderDm archive center  dm
	EsOrderDm = "dm"
	// EsOrderFav archive center fav
	EsOrderFav = "fav"
	// ActOrderLike activity list like order.
	ActOrderLike = "like"
	// ActOrderCtime activity list ctime order.
	ActOrderCtime = "ctime"
	// ActOrderRandom order random .
	ActOrderRandom = "random"
	// ActOrderStochastic order stochastic .
	ActOrderStochastic = "stochastic"
)

// ipRequestKey .
func ipRequestKey(ip string) string {
	return fmt.Sprintf(_keyIPRequestFmt, ip)
}

func likeListCtimeKey(sid int64) string {
	return fmt.Sprintf(_keyLikeListCtimeFmt, sid)
}

func likeListRandomKey(sid int64) string {
	return fmt.Sprintf(_keyLikeListRandomFmt, sid)
}

func likeListRandomTypeKey(ltype, sid int64) string {
	return fmt.Sprintf(_keyLikeListRandomTypeFmt, ltype, sid)
}

func likeStochasticKey(sid int64) string {
	return fmt.Sprintf(_keyLikeStochasticFmt, sid)
}

func likeStochasticTypeKey(ltype int64, sid int64) string {
	return fmt.Sprintf(_keyLikeStochasticTypeFmt, ltype, sid)
}

func likeListTypeCtimeKey(types int64, sid int64) string {
	return fmt.Sprintf(_keyLikeListTypeCtimeFmt, types, sid)
}

func likeTotalKey(sid int64) string {
	return fmt.Sprintf("go:l:t:%d", sid)
}

func keyLikeTag(sid, tagID int64) string {
	return fmt.Sprintf(_keyLikeTagFmt, sid, tagID)
}

func keyLikeTagCounts(sid int64) string {
	return fmt.Sprintf(_keyLikeTagCntsFmt, sid)
}

func keyLikeRegion(sid int64, regionID int32) string {
	return fmt.Sprintf(_keyLikeRegionFmt, sid, regionID)
}

func keyStoryLikeKey(sid, mid int64, daily string) string {
	return fmt.Sprintf(_keyStoryDilyLikeFmt, daily, sid, mid)
}

func keyStoryEachLike(sid, mid, lid int64, daily string) string {
	return fmt.Sprintf(_keyStoryEachLikeFmt, daily, sid, mid, lid)
}

func keyStoryExtraLike(sid, mid int64) string {
	return fmt.Sprintf(_keyStoryExtraLikeFmt, sid, mid)
}

func keyUpAction(sid, mid int64) string {
	return fmt.Sprintf(_keyUpActionFmt, sid, mid)
}

func keyStoryExtendInfo(sid int64, token string) string {
	return fmt.Sprintf(_keyStoryExtendInfoFmt, sid, token)
}

func keyStoryExtendToken(sid, mid int64) string {
	return fmt.Sprintf(_keyStoryExtendTokenFmt, sid, mid)
}

func keyLikeWidMapId(wid int64) string {
	return fmt.Sprintf(_keyLikeWidMapId, wid)
}

func (dao *Dao) cacheSFActStochastic(sid, ltype int64) string {
	return fmt.Sprintf(_keyStochasticSingleFmt, ltype, sid)
}

func (dao *Dao) cacheSFActRandom(sid, ltype, _, _ int64) string {
	return fmt.Sprintf(_keyRandomSingleFmt, ltype, sid)
}

func keyVoteTotalBySID(sid int64) string {
	return fmt.Sprintf(_keyVoteTotalBySID, sid)
}

// LikeTypeList dao sql.
func (dao *Dao) LikeTypeList(c context.Context, sid int64) (ns []*like.Like, err error) {
	rows, err := dao.db.Query(c, _selLikeSQL, sid)
	if err != nil {
		log.Error("LikeTypeList dao.db.Query error(%v)", err)
		return
	}
	ns = make([]*like.Like, 0)
	defer rows.Close()
	for rows.Next() {
		n := &like.Like{
			Item: &like.Item{},
		}
		if err = rows.Scan(&n.ID, &n.Wid); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("row.Scan row error(%v)", err)
	}
	return
}

// LikeList dao sql
func (dao *Dao) LikeList(c context.Context, sid int64) (ns []*like.Item, err error) {
	rows, err := dao.db.Query(c, _likeListSQL, sid)
	if err != nil {
		log.Error("LikeList dao.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.Item)
		if err = rows.Scan(&n.ID, &n.Wid, &n.Ctime); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		ns = append(ns, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("row.Scan row error(%v)", err)
	}
	return
}

// RawLikes get likes by wid.
func (dao *Dao) RawLikes(c context.Context, ids []int64) (data map[int64]*like.Item, err error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_likesSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("Likes dao.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	data = make(map[int64]*like.Item)
	for rows.Next() {
		res := &like.Item{}
		if err = rows.Scan(&res.ID, &res.Sid, &res.Type, &res.Mid, &res.Wid, &res.State, &res.StickTop, &res.Ctime, &res.Mtime); err != nil {
			log.Error("Likes row.Scan error(%v)", err)
			return
		}
		data[res.ID] = res
	}
	if err = rows.Err(); err != nil {
		log.Error("Likes row.Scan row error(%v)", err)
	}
	return
}

// LikeTagCache get like tag cache.
func (dao *Dao) LikeTagCache(c context.Context, sid, tagID int64, start, end int) (likes []*like.Item, err error) {
	var values []interface{}
	key := keyLikeTag(sid, tagID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZREVRANGE", key, start, end)); err != nil {
		log.Error("LikeTagCache conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("LikeRegionCache redis.Scan(%v) error(%v)", values, err)
			return
		}
		like := new(like.Item)
		if err = json.Unmarshal(bs, &like); err != nil {
			log.Error("LikeRegionCache conn.Do(ZRANGE, %s) error(%v)", key, err)
			continue
		}
		if like.ID > 0 {
			likes = append(likes, like)
		}
	}
	return
}

// LikeTagCnt get like tag cnt.
func (dao *Dao) LikeTagCnt(c context.Context, sid, tagID int64) (count int, err error) {
	key := keyLikeTag(sid, tagID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if count, err = redis.Int(conn.Do("ZCARD", key)); err != nil {
		log.Error("LikeRegionCnt conn.Do(ZCARD, %s) error(%v)", key, err)
	}
	return
}

// SetLikeTagCache set like tag cache no expire.
func (dao *Dao) SetLikeTagCache(c context.Context, sid, tagID int64, likes []*like.Item) (err error) {
	var bs []byte
	key := keyLikeTag(sid, tagID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("SetLikeTagCache conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range likes {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("SetLikeTagCache json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// LikeRegionCache get like region cache.
func (dao *Dao) LikeRegionCache(c context.Context, sid int64, regionID int32, start, end int) (likes []*like.Item, err error) {
	var values []interface{}
	key := keyLikeRegion(sid, regionID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZREVRANGE", key, start, end)); err != nil {
		log.Error("LikeRegionCache conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("LikeRegionCache redis.Scan(%v) error(%v)", values, err)
			return
		}
		like := new(like.Item)
		if err = json.Unmarshal(bs, &like); err != nil {
			log.Error("LikeRegionCache conn.Do(ZREVRANGE, %s) error(%v)", key, err)
			continue
		}
		if like.ID > 0 {
			likes = append(likes, like)
		}
	}
	return
}

// LikeRegionCnt get like region cnt.
func (dao *Dao) LikeRegionCnt(c context.Context, sid int64, regionID int32) (count int, err error) {
	key := keyLikeRegion(sid, regionID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if count, err = redis.Int(conn.Do("ZCARD", key)); err != nil {
		log.Error("LikeRegionCnt conn.Do(ZCARD, %s) error(%v)", key, err)
	}
	return
}

// SetLikeRegionCache set like region cache.
func (dao *Dao) SetLikeRegionCache(c context.Context, sid int64, regionID int32, likes []*like.Item) (err error) {
	var bs []byte
	key := keyLikeRegion(sid, regionID)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("SetLikeTagCache conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range likes {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("SetLikeRegionCache json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// SetTagLikeCountsCache .
func (dao *Dao) SetTagLikeCountsCache(c context.Context, sid int64, counts map[int64]int32) (err error) {
	key := keyLikeTagCounts(sid)
	conn := dao.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for tagID, count := range counts {
		args = args.Add(tagID).Add(count)
	}
	if _, err = conn.Do("HMSET", args...); err != nil {
		log.Error("SetLikeCountsCache conn.Do(HMSET) key(%s) error(%v)", key, err)
	}
	return
}

// TagLikeCountsCache get tag like counts cache.
func (dao *Dao) TagLikeCountsCache(c context.Context, sid int64, tagIDs []int64) (counts map[int64]int32, err error) {
	if len(tagIDs) == 0 {
		return
	}
	key := keyLikeTagCounts(sid)
	conn := dao.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(key).AddFlat(tagIDs)
	var tmpCounts []int
	if tmpCounts, err = redis.Ints(conn.Do("HMGET", args...)); err != nil {
		log.Error("redis.Ints(HMGET) key(%s) args(%v) error(%v)", key, args, err)
		return
	}
	if len(tmpCounts) != len(tagIDs) {
		return
	}
	counts = make(map[int64]int32, len(tagIDs))
	for i, tagID := range tagIDs {
		counts[tagID] = int32(tmpCounts[i])
	}
	return
}

// RawLike get like by id .
func (dao *Dao) RawLike(c context.Context, id int64) (res *like.Item, err error) {
	res = new(like.Item)
	row := dao.db.QueryRow(c, _likeSQL, id)
	if err = row.Scan(&res.ID, &res.Sid, &res.Type, &res.Mid, &res.Wid, &res.State, &res.StickTop, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "LikeByID:QueryRow")
		}
	}
	return
}

// LikeListMoreLid get likes data with like.id greater than lid
func (dao *Dao) LikeListMoreLid(c context.Context, lid int64) (res []*like.Item, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _likeMoreLidSQL, lid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "LikeListMoreLid:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make([]*like.Item, 0, 1000)
	for rows.Next() {
		a := &like.Item{}
		if err = rows.Scan(&a.ID, &a.Sid, &a.Type, &a.Mid, &a.Wid, &a.State, &a.StickTop, &a.Ctime, &a.Mtime); err != nil {
			err = errors.Wrap(err, "LikeListMoreLid:rows.Scan()")
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LikeListMoreLid: rows.Err()")
	}
	return
}

// LikesBySid get sid all likes .
func (dao *Dao) LikesBySid(c context.Context, lid, sid int64) (res []*like.Item, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _likesBySidSQL, lid, sid); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "LikesBySid:dao.db.Query()")
		}
		return
	}
	defer rows.Close()
	res = make([]*like.Item, 0, 1000)
	for rows.Next() {
		a := &like.Item{}
		if err = rows.Scan(&a.ID, &a.Sid, &a.Type, &a.Mid, &a.Wid, &a.State, &a.StickTop, &a.Ctime, &a.Mtime); err != nil {
			err = errors.Wrap(err, "LikesBySid:rows.Scan()")
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LikesBySid:rows.Err()")
	}
	return
}

// LikeCtime .
func (dao *Dao) LikeCtime(c context.Context, sid, ltype, start, end int64) ([]int64, error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeListTypeCtimeKey(ltype, sid)
	} else {
		key = likeListCtimeKey(sid)
	}
	return dao.zrevrangeCommon(c, start, end, key)
}

// CacheActRandom .
func (dao *Dao) CacheActRandom(c context.Context, sid, ltype, start, end int64) (res []int64, err error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeListRandomTypeKey(ltype, sid)
	} else {
		key = likeListRandomKey(sid)
	}
	if res, err = redis.Int64s(component.GlobalRedisStore.Do(c, "ZRANGE", key, start, end)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "conn.Do(ZRANGE)")
		}
	}
	return
}

// CacheActStochastic .
func (dao *Dao) CacheActStochastic(c context.Context, sid, ltype int64) (res []int64, err error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeStochasticTypeKey(ltype, sid)
	} else {
		key = likeStochasticKey(sid)
	}
	if res, err = redis.Int64s(component.GlobalRedisStore.Do(c, "SMEMBERS", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "conn.Do(SMEMBERS)")
		}
	}
	return
}

// LikeRandomCount .
func (dao *Dao) LikeRandomCount(c context.Context, sid, ltype int64) (res int64, err error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeListRandomTypeKey(ltype, sid)
	} else {
		key = likeListRandomKey(sid)
	}
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "ZCARD", key)); err != nil {
		log.Error("LikeRandomCount conn.Do(ZCARD, %s) error(%v)", key, err)
		return
	}
	if res == 1 {
		if lids, err := dao.CacheActRandom(c, sid, ltype, 0, 1); err == nil && len(lids) == 1 && lids[0] == -1 {
			res = 0
		}
	}
	return
}

// AddCacheActStochastic .
func (dao *Dao) AddCacheActStochastic(c context.Context, sid int64, ids []int64, ltype int64) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  string
	)
	if ltype > 0 {
		key = likeStochasticTypeKey(ltype, sid)
	} else {
		key = likeStochasticKey(sid)
	}
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range ids {
		args = args.Add(v)
	}
	if err = conn.Send("SADD", args...); err != nil {
		err = errors.Wrap(err, "conn.Send(SADD)")
		return
	}
	expire := dao.stochasticExpire
	// 如果是空缓存过期时间设置短一些 30s
	if len(ids) == 1 && ids[0] == -1 {
		expire = 30
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		err = errors.Wrap(err, "conn.Send(EXPIRE)")
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "conn.Flush()")
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrapf(err, "conn.Receive(%d)", i)
			return
		}
	}
	return
}

// AddCacheActRandom .
func (dao *Dao) AddCacheActRandom(c context.Context, sid int64, ids []int64, ltype int64) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  string
	)
	if ltype > 0 {
		key = likeListRandomTypeKey(ltype, sid)
	} else {
		key = likeListRandomKey(sid)
	}
	defer conn.Close()
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}.Add(key)
	for k, v := range ids {
		args = args.Add(k + 1).Add(v)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		err = errors.Wrap(err, "conn.Send(ZADD)")
		return
	}
	expire := dao.randomExpire
	if len(ids) == 1 && ids[0] == -1 {
		expire = 30
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		err = errors.Wrap(err, "conn.Send(EXPIRE)")
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "conn.Flush()")
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrapf(err, "conn.Receive(%d)", i)
			return
		}
	}
	return
}

// LikeCount .
func (dao *Dao) LikeCount(c context.Context, sid, ltype int64) (res int64, err error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeListTypeCtimeKey(ltype, sid)
	} else {
		var isEnt bool
		for _, v := range dao.c.Ent.UpSids {
			if v == sid {
				isEnt = true
				break
			}
		}
		if isEnt {
			key = entUpRankKey(sid)
		} else {
			key = likeListCtimeKey(sid)
		}
	}
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "ZCARD", key)); err != nil {
		log.Error("LikeCount conn.Do(ZCARD, %s) error(%v)", key, err)
	}
	return
}

// LikeListCtime set like list by ctime.
func (dao *Dao) LikeListCtime(c context.Context, sid int64, items []*like.Item) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  = likeListCtimeKey(sid)
		max  = 0
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range items {
		args = args.Add(v.Ctime).Add(v.ID)
		if v.Type != 0 {
			typeKey := likeListTypeCtimeKey(v.Type, sid)
			typeArgs := redis.Args{}.Add(typeKey).Add(v.Ctime).Add(v.ID)
			if err = conn.Send("ZADD", typeArgs...); err != nil {
				log.Error("LikeListCtime:conn.Send(%v) error(%v)", v, err)
				return
			}
			max++
		}
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("LikeListCtime:conn.Send(%v) error(%v)", items, err)
		return
	}
	max++
	if err = conn.Flush(); err != nil {
		log.Error("LikeListCtime:conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("LikeListCtime:conn.Receive(%d) error(%v)", i, err)
			return
		}
	}
	return
}

// DelLikeListCtime delete likeList Ctime cache .
func (dao *Dao) DelLikeListCtime(c context.Context, sid int64, items []*like.Item) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  = likeListCtimeKey(sid)
		max  = 0
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range items {
		args = args.Add(v.ID)
		if v.Type != 0 {
			typeKey := likeListTypeCtimeKey(v.Type, sid)
			if err = conn.Send("ZREM", typeKey, v.ID); err != nil {
				log.Error("DelLikeListCtime:conn.Send(%v) error(%v)", v, err)
				return
			}
			max++
		}
	}
	if err = conn.Send("ZREM", args...); err != nil {
		log.Error("DelLikeListCtime:conn.Send(%v) error(%v)", args, err)
		return
	}
	max++
	if err = conn.Flush(); err != nil {
		log.Error("DelLikeListCtime:conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("DelLikeListCtime:conn.Receive(%d) error(%v)", i, err)
			return
		}
	}
	return
}

// LikeMaxID get likes last id .
func (dao *Dao) LikeMaxID(c context.Context) (res *like.Item, err error) {
	res = new(like.Item)
	rows := dao.db.QueryRow(c, _likeMaxIDSQL)
	if err = rows.Scan(&res.ID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "LikeMaxID:QueryRow")
		}
	}
	return
}

// TxAddLike .
func (dao *Dao) TxAddLike(c context.Context, tx *sql.Tx, item *like.Item) (res int64, err error) {
	var (
		now    = time.Now().Format("2006-01-02 15:04:05")
		sqlRes xsql.Result
	)
	if sqlRes, err = tx.Exec(fmt.Sprintf(_likeInsertSQL, _oldTableName), item.Sid, item.Type, item.Mid, item.Wid, item.State, item.StickTop, now, now, item.Referer); err != nil {
		log.Error("TxAddLike:tx.Exec(%s) error(%v)", _oldTableName, err)
		return
	}
	return sqlRes.LastInsertId()
}

// StateModify .
func (dao *Dao) StateModify(c context.Context, lid, mid int64, state int) (ef int64, err error) {
	var res xsql.Result
	if res, err = dao.db.Exec(c, _likeUpStateSQL, state, lid, mid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	ef, _ = res.RowsAffected()
	return
}

// GroupItemData like data.
func (dao *Dao) GroupItemData(c context.Context, sid int64, ck string) (data []*like.GroupItem, err error) {
	var req *http.Request
	if req, err = dao.client.NewRequest(http.MethodGet, fmt.Sprintf(dao.likeItemURL, sid), metadata.String(c, metadata.RemoteIP), url.Values{}); err != nil {
		return
	}
	req.Header.Set("Cookie", ck)
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*like.GroupItem `json:"list"`
		} `json:"data"`
	}
	if err = dao.client.Do(c, req, &res, dao.likeItemURL); err != nil {
		err = errors.Wrapf(err, "LikeData dao.client.Do sid(%d)", sid)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "LikeData sid(%d)", sid)
		return
	}
	data = res.Data.List
	return
}

// RawSourceItemData get source data.
func (dao *Dao) RawSourceItemData(c context.Context, sid int64) (sids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []*struct {
				Data struct {
					Sid string `json:"sid"`
				} `json:"data"`
			} `json:"list"`
		} `json:"data"`
	}
	if err = dao.client.RESTfulGet(c, dao.sourceItemURL, metadata.String(c, metadata.RemoteIP), url.Values{}, &res, sid); err != nil {
		err = errors.Wrapf(err, "LikeData dao.client.RESTfulGet sid(%d)", sid)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "LikeData sid(%d)", sid)
		return
	}
	for _, v := range res.Data.List {
		if sid, e := strconv.ParseInt(v.Data.Sid, 10, 64); e != nil {
			continue
		} else {
			sids = append(sids, sid)
		}
	}
	return
}

// SourceItem get source data json raw message.
func (dao *Dao) SourceItem(c context.Context, sid int64) (source json.RawMessage, err error) {
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = dao.client.RESTfulGet(c, dao.sourceItemURL, metadata.String(c, metadata.RemoteIP), url.Values{}, &res, sid); err != nil {
		err = errors.Wrapf(err, "LikeData dao.client.RESTfulGet sid(%d)", sid)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "LikeData sid(%d)", sid)
		return
	}
	source = res.Data
	return
}

// StoryLikeSum .
func (dao *Dao) StoryLikeSum(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		now = time.Now().Format("2006-01-02")
		key = keyStoryLikeKey(sid, mid, now)
	)
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = -1
		} else {
			err = errors.Wrap(err, "redis.Do(get)")
		}
	}
	return
}

// IncrStoryLikeSum .
func (dao *Dao) IncrStoryLikeSum(c context.Context, sid, mid int64, score int64) (res int64, err error) {
	var (
		now = time.Now().Format("2006-01-02")
		key = keyStoryLikeKey(sid, mid, now)
	)
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "INCRBY", key, score)); err != nil {
		err = errors.Wrap(err, "redis.Do(get)")
	}
	return
}

// SetLikeSum .
func (dao *Dao) SetLikeSum(c context.Context, sid, mid int64, sum int64) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		now  = time.Now().Format("2006-01-02")
		key  = keyStoryLikeKey(sid, mid, now)
		res  bool
	)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", key, sum)); err != nil {
		err = errors.Wrap(err, "redis.Bool(SETNX)")
		return
	}
	if res {
		conn.Do("EXPIRE", key, 86400)
	} else {
		err = errors.New("redis.Bool(SETNX) res false")
	}
	return
}

// BatchStoryEachLikeSum .
func (dao *Dao) BatchStoryEachLikeSum(c context.Context, sid, mid int64, lids []int64) (liked map[int64]int64, err error) {
	var (
		conn = dao.redis.Get(c)
		args = redis.Args{}
		ss   []int64
		now  = time.Now().Format("2006-01-02")
	)
	defer conn.Close()
	for _, lid := range lids {
		args = args.Add(keyStoryEachLike(sid, mid, lid, now))
	}
	if ss, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrapf(err, "redis.Ints(conn.Do(HMGET,%v)", args)
		}
		return
	}
	liked = make(map[int64]int64, len(lids))
	for key, val := range ss {
		liked[lids[key]] = val
	}
	return
}

// StoryEachLikeSum  .
func (dao *Dao) StoryEachLikeSum(c context.Context, sid, mid, lid int64) (res int64, err error) {
	var (
		conn = dao.redis.Get(c)
		now  = time.Now().Format("2006-01-02")
		key  = keyStoryEachLike(sid, mid, lid, now)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = -1
		} else {
			err = errors.Wrap(err, "redis.Do(get)")
		}
	}
	return
}

// IncrStoryEachLikeAct .
func (dao *Dao) IncrStoryEachLikeAct(c context.Context, sid, mid, lid int64, score int64) (res int64, err error) {
	var (
		conn = dao.redis.Get(c)
		now  = time.Now().Format("2006-01-02")
		key  = keyStoryEachLike(sid, mid, lid, now)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("INCRBY", key, score)); err != nil {
		err = errors.Wrap(err, "redis.Do(get)")
	}
	return
}

// SetEachLikeSum .
func (dao *Dao) SetEachLikeSum(c context.Context, sid, mid, lid int64, sum int64) (err error) {
	var (
		conn = dao.redis.Get(c)
		now  = time.Now().Format("2006-01-02")
		key  = keyStoryEachLike(sid, mid, lid, now)
		res  bool
	)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", key, sum)); err != nil {
		err = errors.Wrap(err, "redis.Bool(SETNX)")
		return
	}
	if res {
		conn.Do("EXPIRE", key, 86400)
	} else {
		err = errors.New("redis.Bool(SETNX) res false")
	}
	return
}

// MyListFromEs .
func (dao *Dao) MyListFromEs(c context.Context, sid, mid int64, order string, ps, pn, isOriginal int) (res *like.ListInfo, err error) {
	actResult := new(struct {
		Result []struct {
			ID  int64 `json:"id"`
			Wid int64 `json:"wid"`
			//Ctime     xtime.Time `json:"ctime"`
			Sid       int64 `json:"sid"`
			Type      int64 `json:"type"`
			LikesType int64 `json:"likes_type"` //likes表type
			Mid       int64 `json:"mid"`
			State     int64 `json:"state"`
			Copyright int   `json:"copyright"`
			//Mtime     xtime.Time `json:"mtime"`
		} `json:"result"`
		Page *like.Page `json:"page"`
	})
	req := dao.es.NewRequest(_activity).Index(_activity).WhereEq("sid", sid).WhereEq("mid", mid).Ps(ps).Pn(pn)
	if order != "" {
		req.Order(order, elastic.OrderDesc)
	}
	if isOriginal == 1 {
		req.WhereEq("copyright", 1)
	}
	req.Fields("id", "sid", "wid", "mid", "type", "state", "copyright", "likes_type")
	if err = req.Scan(c, &actResult); err != nil {
		err = errors.Wrap(err, "req.Scan")
		return
	}
	if len(actResult.Result) == 0 {
		return
	}
	res = &like.ListInfo{Page: actResult.Page, List: make([]*like.List, 0, len(actResult.Result))}
	for _, v := range actResult.Result {
		a := &like.List{
			Item: &like.Item{
				ID:  v.ID,
				Wid: v.Wid,
				//Ctime: v.Ctime,
				Sid:   v.Sid,
				Type:  v.LikesType,
				Mid:   v.Mid,
				State: v.State,
				//Mtime: v.Mtime,
			},
		}
		res.List = append(res.List, a)
	}
	return
}

// AllListFromEs .
func (dao *Dao) AllListFromEs(c context.Context, sids []int64, mid int64, order string, ps, pn, isOriginal int) (res *like.ListInfo, err error) {
	actResult := new(struct {
		Result []struct {
			ID        int64 `json:"id"`
			Wid       int64 `json:"wid"`
			Sid       int64 `json:"sid"`
			Type      int64 `json:"type"`
			LikesType int64 `json:"likes_type"` //likes表type
			Mid       int64 `json:"mid"`
			State     int64 `json:"state"`
			Copyright int   `json:"copyright"`
		} `json:"result"`
		Page *like.Page `json:"page"`
	})
	req := dao.es.NewRequest(_activity).Index(_activity).WhereIn("sid", sids).WhereEq("mid", mid).WhereEq("state", 1).Ps(ps).Pn(pn)
	if order != "" {
		req.Order(order, elastic.OrderDesc)
	}
	if isOriginal == 1 {
		req.WhereEq("copyright", 1)
	}
	req.Fields("id", "sid", "wid", "mid", "type", "state", "copyright", "likes_type")
	if err = req.Scan(c, &actResult); err != nil {
		err = errors.Wrap(err, "req.Scan")
		return
	}
	if len(actResult.Result) == 0 {
		return
	}
	res = &like.ListInfo{Page: actResult.Page, List: make([]*like.List, 0, len(actResult.Result))}
	for _, v := range actResult.Result {
		a := &like.List{
			Item: &like.Item{
				ID:    v.ID,
				Wid:   v.Wid,
				Sid:   v.Sid,
				Type:  v.LikesType,
				Mid:   v.Mid,
				State: v.State,
			},
		}
		res.List = append(res.List, a)
	}
	return
}

// MyListTotalStateFromEs .
func (dao *Dao) MyListTotalStateFromEs(c context.Context, sid, mid int64, isOriginal int) (data *like.SubjectStat, err error) {
	res := new(struct {
		Result struct {
			SumCoin []struct {
				Value float64 `json:"value"`
			} `json:"sum_coin"`
			SumLikes []struct {
				Value float64 `json:"value"`
			} `json:"sum_likes"`
		} `json:"result"`
		Page struct {
			Total int64 `json:"total"`
		} `json:"page"`
	})
	req := dao.es.NewRequest(_activity).Index(_activity).WhereEq("sid", sid).WhereEq("mid", mid).WhereEq("state", 1).Sum("likes").Sum("coin")
	if isOriginal == 1 {
		req.WhereEq("copyright", 1)
	}
	if err = req.Scan(c, &res); err != nil {
		return
	}
	data = &like.SubjectStat{
		Coin:  int64(res.Result.SumCoin[0].Value),
		Like:  int64(res.Result.SumLikes[0].Value),
		Count: res.Page.Total,
	}
	return
}

// RawActStochastic .
func (dao *Dao) RawActStochastic(c context.Context, sid, ltype int64) ([]int64, error) {
	orderList, err := dao.ListFromES(c, sid, "", 500, 1, time.Now().Unix(), ltype)
	if err != nil {
		log.Error("d.ListFromES(%d,%d) error(%+v)", sid, ltype, err)
		return nil, err
	}
	if orderList == nil || len(orderList.List) == 0 {
		return nil, nil
	}
	orderIDs := make([]int64, 0)
	for _, v := range orderList.List {
		orderIDs = append(orderIDs, v.ID)
	}
	return orderIDs, nil
}

// RawActRandom .
func (dao *Dao) RawActRandom(c context.Context, sid, ltype, start, end int64) ([]int64, []int64, error) {
	orderList, err := dao.ListFromES(c, sid, "", 500, 1, time.Now().Unix(), ltype)
	if err != nil {
		log.Error("d.ListFromES(%d,%d) error(%+v)", sid, ltype, err)
		return nil, nil, err
	}
	if orderList == nil || len(orderList.List) == 0 {
		return nil, nil, nil
	}

	missIDs := make([]int64, 0)
	for _, v := range orderList.List {
		missIDs = append(missIDs, v.ID)
	}
	lidLent := int64(len(missIDs))
	if end == -1 {
		end = lidLent
	} else {
		end += 1
		if lidLent < end {
			end = lidLent
		}
	}
	var lids []int64
	if start > end {
		lids = []int64{}
	} else {
		lids = missIDs[start:end]
	}
	return lids, missIDs, nil
}

// ListFromES .
func (dao *Dao) ListFromES(c context.Context, sid int64, order string, ps, pn int, seed, ltype int64) (res *like.ListInfo, err error) {
	actResult := new(struct {
		Result []struct {
			ID  int64 `json:"id"`
			Wid int64 `json:"wid"`
			//Ctime xtime.Time `json:"ctime"`
			Sid       int64 `json:"sid"`
			Type      int64 `json:"type"`       // subject表type
			LikesType int64 `json:"likes_type"` //likes表type
			Mid       int64 `json:"mid"`
			State     int64 `json:"state"`
			//Mtime xtime.Time `json:"mtime"`
			Likes int64 `json:"likes"`
			Click int64 `json:"click"`
			Coin  int64 `json:"coin"`
			Share int64 `json:"share"`
			Reply int64 `json:"reply"`
			Dm    int64 `json:"dm"`
			Fav   int64 `json:"fav"`
		} `json:"result"`
		Page *like.Page `json:"page"`
	})
	req := dao.es.NewRequest(_activity).Index(_activity).WhereEq("sid", sid).WhereEq("state", 1).Ps(ps).Pn(pn)
	if ltype > 0 {
		req.WhereEq("likes_type", ltype)
	}
	if order != "" {
		req.Order(order, elastic.OrderDesc)
	}
	if seed > 0 {
		req.OrderRandomSeed(time.Unix(seed, 0).Format("2006-01-02 15:04:05"))
	}
	req.Fields("id", "sid", "wid", "mid", "type", "state", "click", "likes", "coin", "share", "reply", "dm", "fav", "likes_type")
	if err = req.Scan(c, &actResult); err != nil {
		err = errors.Wrap(err, "req.Scan")
		return
	}
	if len(actResult.Result) == 0 {
		return
	}
	res = &like.ListInfo{Page: actResult.Page, List: make([]*like.List, 0, len(actResult.Result))}
	for _, v := range actResult.Result {
		a := &like.List{
			Likes: v.Likes,
			Click: v.Click,
			Coin:  v.Coin,
			Share: v.Share,
			Reply: v.Reply,
			Dm:    v.Dm,
			Fav:   v.Fav,
			Item: &like.Item{
				ID:  v.ID,
				Wid: v.Wid,
				//Ctime: v.Ctime,
				Sid:   v.Sid,
				Type:  v.LikesType,
				Mid:   v.Mid,
				State: v.State,
				//Mtime: v.Mtime,
			},
		}
		res.List = append(res.List, a)
	}
	return
}

// MultiTags .
func (dao *Dao) MultiTags(c context.Context, wids []int64) (tagList map[int64][]string, err error) {
	if len(wids) == 0 {
		return
	}
	var res struct {
		Code int                   `json:"code"`
		Data map[int64][]*like.Tag `json:"data"`
	}
	params := url.Values{}
	params.Set("aids", xstr.JoinInts(wids))
	if err = dao.client.Get(c, dao.tagURL, "", params, &res); err != nil {
		log.Error("MultiTags:dao.client.Get(%s) error(%+v)", dao.tagURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		return
	}
	tagList = make(map[int64][]string, len(res.Data))
	for k, v := range res.Data {
		if len(v) == 0 {
			continue
		}
		tagList[k] = make([]string, 0, len(v))
		for _, val := range v {
			tagList[k] = append(tagList[k], val.Name)
		}
	}
	return
}

// OidInfoFromES .
func (dao *Dao) OidInfoFromES(c context.Context, oids []int64, sType, ps, pn int) (res map[int64]*like.Item, err error) {
	actResult := new(struct {
		Result []struct {
			ID  int64 `json:"id"`
			Wid int64 `json:"wid"`
			//Ctime xtime.Time `json:"ctime"`
			Sid       int64 `json:"sid"`
			Type      int64 `json:"type"`
			LikesType int64 `json:"likes_type"` //likes表type
			Mid       int64 `json:"mid"`
			State     int64 `json:"state"`
			//Mtime xtime.Time `json:"mtime"`
			Likes int64 `json:"likes"`
			Click int64 `json:"click"`
			Coin  int64 `json:"coin"`
			Share int64 `json:"share"`
			Reply int64 `json:"reply"`
			Dm    int64 `json:"dm"`
			Fav   int64 `json:"fav"`
		} `json:"result"`
		Page *like.Page `json:"page"`
	})
	req := dao.es.NewRequest(_activity).Index(_activity).WhereIn("wid", oids).WhereEq("type", sType).Ps(ps).Pn(pn)
	req.Fields("id", "sid", "wid", "mid", "type", "state", "likes_type")
	if err = req.Scan(c, &actResult); err != nil {
		err = errors.Wrap(err, "req.Scan")
		return
	}
	if len(actResult.Result) == 0 {
		return
	}
	res = make(map[int64]*like.Item, len(actResult.Result))
	for _, v := range actResult.Result {
		res[v.Wid] = &like.Item{
			ID:  v.ID,
			Wid: v.Wid,
			//Ctime: v.Ctime,
			Sid:   v.Sid,
			Type:  v.LikesType,
			Mid:   v.Mid,
			State: v.State,
			//Mtime: v.Mtime,
		}
	}
	return
}

// RawLikeTotal .
func (dao *Dao) RawLikeTotal(c context.Context, sid int64) (total int64, err error) {
	row := dao.db.QueryRow(c, _likeCountSQL, sid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawLikeCount row.Scan() error(%v)", err)
		}
	}
	return
}

// CacheLikeTotal .
func (dao *Dao) CacheLikeTotal(c context.Context, sid int64) (total int64, err error) {
	key := likeTotalKey(sid)
	if total, err = redis.Int64(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do (GET key(%s)) error(%v)", key, err)
		}
	}
	return
}

// AddCacheLikeTotal .
func (dao *Dao) AddCacheLikeTotal(c context.Context, sid, total int64) (err error) {
	key := likeTotalKey(sid)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	if err = conn.Send("SET", key, total); err != nil {
		log.Error("conn.Send(SET, %s, %d) error(%v)", key, total, err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.likeTotalExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, dao.likeTotalExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

// IncrCacheLikeTotal .
func (dao *Dao) IncrCacheLikeTotal(c context.Context, sid int64) (err error) {
	var ok bool
	key := likeTotalKey(sid)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, dao.likeTotalExpire)); err != nil {
		log.Error("conn.Do(EXPIRE) key(%s) error(%v)", key, err)
		return
	}
	if !ok {
		return
	}
	if _, err = conn.Do("INCR", key); err != nil {
		log.Error("conn.Do(INCR key(%s)) error(%v)", key, err)
	}
	return
}

// RawLikeCheck .
func (dao *Dao) RawLikeCheck(c context.Context, mid, sid int64) (res *like.Item, err error) {
	res = new(like.Item)
	row := dao.db.QueryRow(c, _likeCheckSQL, sid, mid)
	if err = row.Scan(&res.ID, &res.Sid, &res.Type, &res.Mid, &res.Wid, &res.State, &res.StickTop, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			err = errors.Wrap(err, "RawLikeCheck:QueryRow")
		}
	}
	return
}

// RawTextOnly .
func (dao *Dao) RawTextOnly(c context.Context, sid, mid int64) (cnt int, err error) {
	row := dao.db.QueryRow(c, _likeUserCheckSQL, sid, mid)
	if err = row.Scan(&cnt); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawLikeCount row.Scan(%d,%d) error(%v)", sid, mid, err)
		}
	}
	return
}

// CacheUpActionTimes
func (dao *Dao) CacheUpActionTimes(c context.Context, sid, mid, start, end int64) (res []*like.Action, err error) {
	var (
		key    = keyUpAction(sid, mid)
		values []interface{}
	)
	values, err = redis.Values(component.GlobalRedis.Do(c, "ZREVRANGE", key, start, end))
	if err != nil {
		log.Error("CacheUpActionTimes conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("CacheUpActionTimes redis.Scan(%v) error(%v)", values, err)
			return
		}
		act := &like.Action{}
		if err = json.Unmarshal(bs, act); err != nil {
			log.Error("CacheUpActionTimes json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, act)
	}
	return
}

// CacheLikeExtendTimes
func (dao *Dao) CacheLikeExtraTimes(c context.Context, sid, mid, start, end int64) (res []*like.ExtraTimesDetail, err error) {
	var (
		key    = keyStoryExtraLike(sid, mid)
		conn   = component.GlobalRedis.Conn(c)
		values []interface{}
	)
	defer conn.Close()
	values, err = redis.Values(conn.Do("ZREVRANGE", key, start, end))
	if err != nil {
		log.Error("CacheLikeExtraTimes conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		etd := &like.ExtraTimesDetail{}
		if err = json.Unmarshal(bs, etd); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, etd)
	}
	return
}

func (dao *Dao) AddCacheUpActionTimes(c context.Context, sid, mid int64, list []*like.Action) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = keyUpAction(sid, mid)
		conn = component.GlobalRedis.Conn(c)
		bs   []byte
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddCacheUpActionTimes conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range list {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUpActionTimes json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUpActionTimes conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, 8640000); err != nil {
		log.Error("AddCacheUpActionTimes conn.Send(Expire, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUpActionTimes conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUpActionTimes conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AppendUpActionCache .
func (dao *Dao) AppendUpActionCache(c context.Context, sid, mid int64, act *like.Action) (err error) {
	var (
		bs   []byte
		ok   bool
		key  = keyUpAction(sid, mid)
		conn = component.GlobalRedis.Conn(c)
	)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, 8640000)); err != nil {
		log.Error("AppendUpActionCache conn.Do(EXPIRE %s) sid(%d) mid(%d) error(%v)", key, sid, mid, err)
		return
	}
	if !ok {
		return
	}
	args := redis.Args{}.Add(key)
	if bs, err = json.Marshal(act); err != nil {
		log.Error("AppendUpActionCache json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(act.Ctime).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AppendUpActionCache conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	if err = conn.Send("EXPIRE", key, 8640000); err != nil {
		log.Error("AppendUpActionCache conn.Send(Expire, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AppendUpActionCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AppendUpActionCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (dao *Dao) AddCacheLikeExtraTimes(c context.Context, sid, mid int64, list []*like.ExtraTimesDetail) (err error) {
	if len(list) == 0 {
		return
	}
	var (
		key  = keyStoryExtraLike(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if err = conn.Send("DEL", key); err != nil {
		log.Error("AddCacheLikeExtraTimes conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(key)
	for _, v := range list {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheLikeExtraTimes json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.Ctime).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheLikeExtraTimes conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
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

func (dao *Dao) AddLikeExtraTimes(c context.Context, sid int64, mid int64, etd *like.ExtraTimesDetail) (err error) {
	var (
		key  = keyStoryExtraLike(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	if bs, err = json.Marshal(etd); err != nil {
		log.Error("AddLikeExtraTimes json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(etd.Ctime).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddLikeExtraTimes conn.Send(ZADD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive() error(%v)", err)
		return
	}
	return
}

func (dao *Dao) RawLikeExtraTimes(c context.Context, sid, mid int64) (res []*like.ExtraTimesDetail, err error) {
	var rows *sql.Rows
	if rows, err = dao.db.Query(c, _likeExtraTimesSQL, sid, mid); err != nil {
		err = errors.Wrapf(err, "RawLikeExtraTimes:Query(%s)", _likeExtraTimesSQL)
		return
	}
	defer rows.Close()
	for rows.Next() {
		etd := new(like.ExtraTimesDetail)
		if err = rows.Scan(&etd.ID, &etd.Sid, &etd.Mid, &etd.Num, &etd.Ctime); err != nil {
			err = errors.Wrap(err, "RawLikeExtraTimes:scan()")
			return
		}
		res = append(res, etd)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLikeExtraTimes:rows.Err()")
	}
	return
}

func (dao *Dao) RawLikeExtendInfo(c context.Context, sid int64, token string) (res *like.ExtendTokenDetail, err error) {
	row := dao.db.QueryRow(c, _likeExtendInfoSQL, sid, token)
	res = new(like.ExtendTokenDetail)
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Token, &res.Max, &res.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawLikeExtendToken row.Scan(%d,%s) error(%v)", sid, token, err)
		}
	}
	return
}

func (dao *Dao) CacheLikeExtendInfo(c context.Context, sid int64, token string) (res *like.ExtendTokenDetail, err error) {
	var (
		key  = keyStoryExtendInfo(sid, token)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLikeExtendToken(%s) return nil", key)
		} else {
			log.Error("CacheLikeExtendToken conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (dao *Dao) AddCacheLikeExtendInfo(c context.Context, sid int64, val *like.ExtendTokenDetail, token string) (err error) {
	var (
		key  = keyStoryExtendInfo(sid, token)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		log.Error("json.Marshal(%v) error (%v)", val, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.likeTokenExpire, bs); err != nil {
		log.Error("conn.Send(SET, %s, %d, %s) error(%v)", key, dao.likeTokenExpire, string(bs), err)
	}
	return
}

func (dao *Dao) RawLikeExtendToken(c context.Context, sid, mid int64) (res *like.ExtendTokenDetail, err error) {
	row := dao.db.QueryRow(c, _likeExtendTokenSQL, sid, mid)
	res = &like.ExtendTokenDetail{}
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Token, &res.Max, &res.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawLikeExtendToken row.Scan(%d,%d) error(%v)", sid, mid, err)
		}
	}
	return
}

func (dao *Dao) CacheLikeExtendToken(c context.Context, sid, mid int64) (res *like.ExtendTokenDetail, err error) {
	var (
		key  = keyStoryExtendToken(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLikeExtendToken(%s) return nil", key)
		} else {
			log.Error("CacheLikeExtendToken conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// RsSetNX Dao
func (dao *Dao) AddCacheLikeExtendToken(c context.Context, sid int64, val *like.ExtendTokenDetail, mid int64) (err error) {
	var (
		key  = keyStoryExtendToken(sid, mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		log.Error("json.Marshal(%v) error (%v)", val, err)
		return
	}
	if err = conn.Send("SETEX", key, dao.likeTokenExpire, bs); err != nil {
		log.Error("conn.Send(SET, %s, %d, %s) error(%v)", key, dao.likeTokenExpire, string(bs), err)
	}
	return
}

func (dao *Dao) RawActivityArchives(ctx context.Context, sid, mid int64) ([]*like.Item, error) {
	rows, err := dao.db.Query(ctx, _activityArcsSQL, mid, sid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*like.Item{}
	for rows.Next() {
		arc := &like.Item{}
		if err = rows.Scan(&arc.ID, &arc.Wid, &arc.Ctime, &arc.State); err != nil {
			return nil, err
		}
		if arc.State != 1 {
			continue
		}
		out = append(out, arc)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// 读缓存
func (d *Dao) GetLidByWidFromCache(c context.Context, wid int64) (lid int64, err error) {
	key := keyLikeWidMapId(wid)
	if lid, err = redis.Int64(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err != redis.ErrNil {
			err = errors.Errorf("GetLidByWidFromCache Err conn.Do(GET %s) error(%v)", key, err)
			return
		}
	}
	err = nil
	return
}

// 写缓存
func (d *Dao) SetLidByWidToCache(c context.Context, wid, lid int64) (err error) {
	key := keyLikeWidMapId(wid)
	if _, err = redis.String(component.GlobalRedis.Do(c, "SET", key, lid)); err != nil {
		if err != redis.ErrNil {
			err = errors.Errorf("SetLidByWidToCache Err conn.Do(SET %s %s) error(%v)", key, lid, err)
			return
		}
	}
	err = nil
	return
}

// 设置过期时间
func (d *Dao) SetLidByWidToCacheExpireTime(c context.Context, wid int64) (err error) {
	key := keyLikeWidMapId(wid)
	if _, err = redis.Bool(component.GlobalRedis.Do(c, "EXPIRE", key, 600)); err != nil {
		if err != redis.ErrNil {
			err = errors.Errorf("SetLidByWidToCacheExpireTime Err conn.Do(EXPIRE %s 600) error(%v)", key, err)
			return
		}
	}
	err = nil
	return
}

// db中通过wid和state查询有效数据的id
func (d *Dao) GetLidByWidFromDB(c context.Context, wid int64) (id int64, err error) {
	rows, err := d.db.Query(c, _activityGetLidByWidSQL, wid, 1)
	if err != nil {
		return
	}
	defer rows.Close()
	item := new(like.Item)
	for rows.Next() {
		if err = rows.Scan(&item.ID); err != nil {
			return
		}
	}
	if err = rows.Err(); err != nil {
		return
	}

	return item.ID, nil
}

func (d *Dao) RawGetActRelationInfo(c context.Context, id int64) (*like.ActRelationInfo, error) {
	rows, err := d.db.Query(c, _actRelationInfoSQL, id, like.ActRelationSubjectStatusNormal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	item := new(like.ActRelationInfo)
	for rows.Next() {
		if err = rows.Scan(&item.ID, &item.Name, &item.NativeIDs, &item.H5IDs, &item.WebIDs, &item.LotteryIDs, &item.ReserveIDs, &item.VideoSourceIDs, &item.FollowIDs, &item.SeasonIDs, &item.ReserveConfig, &item.FollowConfig, &item.SeasonConfig, &item.FavoriteInfo, &item.FavoriteConfig, &item.MallIDs, &item.MallConfig, &item.TopicIDs, &item.TopicConfig); err != nil {
			return nil, err
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return item, nil
}

func (d *Dao) HotGetActRelationInfo(c context.Context) ([]int64, error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_actRelationInfoGetHotIDsSQL, like.ActRelationSubjectStatusNormal))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	IDs := make([]int64, 0)
	ID := new(int64)
	for rows.Next() {
		if err = rows.Scan(&ID); err != nil {
			return nil, err
		}
		IDs = append(IDs, *ID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return IDs, nil
}

func (d *Dao) HotGetActSubjectInfo(c context.Context, nowTime string) ([]int64, error) {
	// 目前只查询出来预约活动的subject
	rows, err := d.db.Query(c, fmt.Sprintf(_actSubjectInfoGetHotIDsSQL, xstr.JoinInts([]int64{
		int64(like.RESERVATION),
		int64(like.CLOCKIN),
		int64(like.USERACTIONSTAT),
	}), nowTime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	IDs := make([]int64, 0)
	ID := new(int64)
	for rows.Next() {
		if err = rows.Scan(&ID); err != nil {
			return nil, err
		}
		IDs = append(IDs, *ID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return IDs, nil
}

func (d *Dao) HotGetActSubjectReserveIDsInfo(c context.Context, nowTime string) ([]int64, error) {
	// 目前只查询出来预约活动的subject
	rows, err := d.db.Query(c, fmt.Sprintf(_actSubjectReserveInfoGetHotIDsSQL, like.RESERVATION, nowTime))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	IDs := make([]int64, 0)
	ID := new(int64)
	for rows.Next() {
		if err = rows.Scan(&ID); err != nil {
			return nil, err
		}
		IDs = append(IDs, *ID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return IDs, nil
}

func (d *Dao) RawUpActReserveRelationInfoWithState(ctx context.Context, sids []int64, state []int64) (map[int64]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithSIDAndStateSQL, xstr.JoinInts(sids), xstr.JoinInts(state)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActReserveRelationInfoBySid(ctx context.Context, sids []int64) (map[int64]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithSIDSQL, xstr.JoinInts(sids)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime, &item.Audit, &item.AuditChannel, &item.DynamicID, &item.DynamicAudit, &item.LotteryType, &item.LotteryID, &item.LotteryAudit); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActRelationReserveListWithLimit(ctx context.Context, mid int64, types []int64, state []int64, maxNumLimit int64) (map[int64]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithStateSQLLimit, mid, xstr.JoinInts(types), xstr.JoinInts(state), maxNumLimit))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) CreateUpActReserveItem(ctx context.Context, param *like.CreateUpActReserveItem, mid int64, relationType pb.UpActReserveRelationType, relationState pb.UpActReserveRelationState, extraParams *like.CreateUpActReserveExtra, createDynamic int64) (lastID int64, err error) {
	var res xsql.Result
	var dynamicID int64
	if res, err = d.db.Exec(ctx, _upActSubjectAddSQL, param.Name, param.Type, param.Stime, param.Etime, param.State); err != nil {
		err = errors.Wrap(err, "d.db.Exec err")
		return
	}
	lastID, err = res.LastInsertId()
	if err != nil {
		err = errors.Wrap(err, "res.LastInsertId err")
		return
	}

	if createDynamic == 1 {
		dynamicID, err = d.CreateDynamic(ctx, mid, lastID)
		if err != nil {
			err = errors.Wrap(err, "d.CreateDynamic err")
			log.Errorc(ctx, err.Error())
			err = nil // 忽略错误码
		}
	} else if extraParams.From == int64(pb.UpCreateActReserveFrom_FromDanmaku) {
		// 弹幕来源的预约在不指定createDynamic的情况下，默认创建影子动态
		// 影子动态不在任何渠道下发，仅作为预约分享时的兜底动态
		dynamicID, err = d.CreateShadowDynamic(ctx, mid, lastID)
		if err != nil {
			err = errors.Wrap(err, "d.CreateShadowDynamic err")
		}
	}

	var dynamicIDStr string
	if dynamicID != 0 {
		dynamicIDStr = strconv.FormatInt(dynamicID, 10)
	}

	// 根据lastID去插入关联表数据
	if _, err = d.db.Exec(ctx, _upActRelationCreateItem,
		lastID,
		mid,
		extraParams.Oid,
		relationType,
		relationState,
		extraParams.From,
		extraParams.LivePlanStartTime,
		extraParams.Audit,
		extraParams.AuditChannel,
		dynamicIDStr,
		extraParams.LotteryType,
		extraParams.LotteryID,
		extraParams.LotteryAudit); err != nil {
		err = errors.Wrap(err, "dao.db.Exec err")
		return
	}
	return
}

func (d *Dao) RawGetUpActReserveRelationInfo(ctx context.Context, sids []int64, mid int64) (res map[int64]*like.UpActReserveRelationInfo, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithMIDAndSIDSQL, mid, xstr.JoinInts(sids)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime, &item.Audit, &item.AuditChannel, &item.LotteryType, &item.LotteryID); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActReserveRelationInfoFromMaster(ctx context.Context, sids []int64) (res map[int64]*like.UpActReserveRelationInfo, err error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithMIDAndSIDFromMasterSQL, xstr.JoinInts(sids)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime, &item.Audit, &item.AuditChannel, &item.LotteryType, &item.LotteryID); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) TXUpdateActSubjectFields(tx *sql.Tx, state int64, sid int64) (err error) {
	_, err = tx.Exec(_actSubjectUpdateStatus, state, sid)
	return
}

func (d *Dao) TXUpdateUpActRelationFields(tx *sql.Tx, relationState int64, mid int64, sid int64) (err error) {
	_, err = tx.Exec(_upActRelationUpdateState, relationState, mid, sid)
	return
}

// 事务修改两张表数据
func (d *Dao) TXUpdateSubjectAndRelationData(ctx context.Context, update *like.UpActReserveRelationUpdateFields) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		err = errors.Wrap(err, "d.db.Begin err")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 修改活动基本信息表的状态
	if err = d.TXUpdateActSubjectFields(
		tx,
		update.SubjectState,
		update.Sid); err != nil {
		return
	}
	// 修改关联表信息表的状态
	if err = d.TXUpdateUpActReserveRelationFields(
		tx,
		update.RelationState,
		update.Mid,
		update.Sid,
		update.AuditState,
		update.AuditChannelState,
		update.DynamicID); err != nil {
		return
	}

	return
}

func (d *Dao) TXUpdateUpActReserveRelationFields(tx *sql.Tx, relationState int64, mid int64, sid int64, auditState int64, auditChannelState int64, dynamicID string) (err error) {
	_, err = tx.Exec(_upActRelationUpdateStateWithAudit, relationState, auditState, auditChannelState, dynamicID, mid, sid)
	return
}

// up主关闭自己的预约活动
func (d *Dao) UpActCancelReserve(ctx context.Context, sid int64, mid int64, subjectState int64, relationState pb.UpActReserveRelationState) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		err = errors.Wrap(err, "d.db.Begin err")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 修改活动基本信息表的状态
	if err = d.TXUpdateActSubjectFields(tx, subjectState, sid); err != nil {
		return
	}
	// 修改关联表信息表的状态
	if err = d.TXUpdateUpActRelationFields(tx, int64(relationState), mid, sid); err != nil {
		return
	}

	return
}

func (d *Dao) TXUpdateActReserveItem(tx *sql.Tx, ctx context.Context, sid int64, param *like.CreateUpActReserveItem) (err error) {
	if _, err = tx.Exec(_upActSubjectUpdateSQL, param.Name, param.Stime, param.Etime, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

func (d *Dao) RawGetUpActReserveRelation(ctx context.Context, mid int64, types []int64, state []int64) ([]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithStateSQL, mid, xstr.JoinInts(types), xstr.JoinInts(state)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make([]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActReserveRelationOthers(ctx context.Context, mid int64) ([]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationOthersWithStateSQL, mid))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make([]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActReserveRelationOfAllMid(ctx context.Context, types []int64, state []int64) ([]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithTypeAndStateSQL, xstr.JoinInts(types), xstr.JoinInts(state)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make([]*like.UpActReserveRelationInfo, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActRelationReserveByOID(ctx context.Context, oid string, typ int64, state int64) (res *like.UpActReserveRelationInfo, err error) {
	row := d.db.QueryRow(ctx, _upActRelationWithOIDSQL, oid, typ, state)
	res = &like.UpActReserveRelationInfo{}
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Oid, &res.Type, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		} else {
			err = errors.Wrap(err, "RawGetUpActRelationReserveByOID:QueryRow")
			return
		}
	}

	return
}

func (d *Dao) UpdateUpActReserveBind(ctx context.Context, oid string, state int64, sid int64) (err error) {
	if _, err = d.db.Exec(ctx, _upActRelationBindSQL, oid, state, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

func (d *Dao) TXUpdateUpActReserveBind(tx *sql.Tx, oid string, state int64, sid int64) (err error) {
	if _, err = tx.Exec(_upActRelationBindSQL, oid, state, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

// oid换绑sid 存在旧的绑定数据
func (d *Dao) UpdateUpActReserveBindUnion(ctx context.Context, oid string, oldSid int64, newSid int64) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 绑定新数据
	if err = d.TXUpdateUpActReserveBind(tx, oid, int64(pb.UpActReserveRelationState_UpReserveRelatedOnline), newSid); err != nil {
		return
	}
	// 解绑老数据
	if err = d.TXUpdateUpActReserveBind(tx, "", int64(pb.UpActReserveRelationState_UpReserveRelated), oldSid); err != nil {
		return
	}

	return
}

func (d *Dao) UpdateUpActReserveState(ctx context.Context, state int64, sid int64) (err error) {
	if _, err = d.db.Exec(ctx, _upActRelationStateSQL, state, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

func (d *Dao) RawGetUpActReserveRelationInfo4SpaceCardIDs(ctx context.Context, mid int64) (res []int64, err error) {
	var relationState []int64
	for _, v := range d.UpActUserSpaceCardState() {
		relationState = append(relationState, int64(v))
	}
	var relationType []int64
	for _, v := range d.UpActReserveType() {
		relationType = append(relationType, int64(v))
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationWithStateSQL, mid, xstr.JoinInts(relationType), xstr.JoinInts(relationState)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	res = make([]int64, 0)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime); err != nil {
			return nil, err
		}
		res = append(res, item.Sid)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return
}

// up发起预约编辑状态
func (d *Dao) UpActReserveRelationEdit(ctx context.Context, mid int64, sid int64, subjectState int64, relationState pb.UpActReserveRelationState) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 修改活动基本信息表数据
	if err = d.TXUpdateActSubjectFields(tx, subjectState, sid); err != nil {
		return
	}
	// 修改关联表信息表数据
	if err = d.TXUpdateUpActRelationFields(tx, int64(relationState), mid, sid); err != nil {
		return
	}

	return
}

// 核销 修改预约结束时间和relation中表状态改变
func (d *Dao) UpActReserveRelationCancel4Arc(ctx context.Context, mid int64, sid int64, ts int64, relationState pb.UpActReserveRelationState) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "tx.Rollback() err")
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 修改活动基本信息表的状态
	if err = d.TXUpdateActSubjectEtime(tx, xtime.Time(ts), sid); err != nil {
		return
	}
	// 修改关联表信息表的状态
	if err = d.TXUpdateUpActRelationFields(tx, int64(relationState), mid, sid); err != nil {
		return
	}

	return
}

func (d *Dao) TXUpdateActSubjectEtime(tx *sql.Tx, time xtime.Time, sid int64) (err error) {
	_, err = tx.Exec(_actSubjectUpdateEtime, time, sid)
	return
}

func (d *Dao) UpdateUpActReserveItem(ctx context.Context, sid int64, param *like.CreateUpActReserveItem, extraParam *like.CreateUpActReserveExtra) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "recover() tx.Rollback() err(%+v)", err)
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err)
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
			return
		}
	}()
	// todo 判断预约类型 稿件只需要更新act_subject 直播需要更新act_subject和up_act_reserve_relation
	// 更新预约基本信息数据
	if err = d.TXUpdateActReserveItem(tx, ctx, sid, param); err != nil {
		return
	}
	// 更新关联的预计直播开始时间
	if err = d.TXUpdateUpActReserveItem(tx, ctx, sid, extraParam); err != nil {
		return
	}
	return
}

func (d *Dao) TXUpdateUpActReserveItem(tx *sql.Tx, ctx context.Context, sid int64, extraParam *like.CreateUpActReserveExtra) (err error) {
	if _, err = tx.Exec(_upActRelationItemSQL, extraParam.LivePlanStartTime, extraParam.LotteryType, extraParam.LotteryID, extraParam.LotteryAudit, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

// 核销 修改预约结束时间和relation中表状态改变
func (d *Dao) UpActReserveRelationCancel4Live(ctx context.Context, mid int64, sid int64, ts int64, relationState pb.UpActReserveRelationState, oid string) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "recover() tx.Rollback() err(%+v)", err)
				return
			}
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(ctx, "tx.Rollback() error(%v)", err)
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "tx.Commit() error(%v)", err)
			return
		}
	}()

	// 修改活动基本信息表的状态
	if err = d.TXUpdateActSubjectEtime(tx, xtime.Time(ts), sid); err != nil {
		return
	}
	// 修改关联表信息表的状态
	if err = d.TXUpdateUpActRelationStateAndOID(tx, int64(relationState), mid, sid, oid); err != nil {
		return
	}

	return
}

func (d *Dao) TXUpdateUpActRelationStateAndOID(tx *sql.Tx, relationState int64, mid int64, sid int64, oid string) (err error) {
	_, err = tx.Exec(_upActRelationUpdateStateAndOID, relationState, oid, mid, sid)
	return
}

// 直播过期数据
func (d *Dao) RawGetUpActReserveLiveExpireData(ctx context.Context, typ int64, state []int64) (map[int64]*like.UpActReserveRelationInfo, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_upActRelationLiveExpireSQL, typ, xstr.JoinInts(state)))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*like.UpActReserveRelationInfo)
	for rows.Next() {
		item := &like.UpActReserveRelationInfo{}
		if err = rows.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime, &item.Audit, &item.AuditChannel, &item.DynamicID); err != nil {
			return nil, err
		}
		items[item.Sid] = item
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *Dao) RawGetUpActReserveRelationInfo4Live(ctx context.Context, upmid int64) (res int64, err error) {
	var relationState []int64
	for _, v := range d.UpActUserUGCAndStoryState() {
		relationState = append(relationState, int64(v))
	}
	relationType := []int64{int64(pb.UpActReserveRelationType_Live)}
	row := d.db.QueryRow(ctx, fmt.Sprintf(_upActRelationWithStateLimitSQL, upmid, xstr.JoinInts(relationType), xstr.JoinInts(relationState)))
	item := &like.UpActReserveRelationInfo{}
	if err = row.Scan(&item.ID, &item.Sid, &item.Mid, &item.Oid, &item.Type, &item.State, &item.LivePlanStartTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		} else {
			err = errors.Wrap(err, "RawGetUpActReserveRelationInfo4Live:QueryRow")
			return
		}
	}

	res = item.Sid
	return
}

func (d *Dao) RawGetUpActRelationReserveByOIDLatestData(ctx context.Context, oid string, types []int64, state []int64) (res *like.UpActReserveRelationInfo, err error) {
	query := _upActRelationWithOIDLatestSQL
	query = strings.Replace(query, "$", tool.SeparatorJoin(types, "?", ","), 1)
	query = strings.Replace(query, "$", tool.SeparatorJoin(state, "?", ","), 2)
	args := make([]interface{}, 0)
	args = append(args, oid)
	for _, v := range types {
		args = append(args, v)
	}
	for _, v := range state {
		args = append(args, v)
	}
	row := d.db.QueryRow(ctx, query, args...)
	res = &like.UpActReserveRelationInfo{}
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Oid, &res.Type, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		} else {
			err = errors.Wrap(err, "RawGetUpActRelationReserveByOIDLatestData:QueryRow")
			return
		}
	}

	return
}

func (d *Dao) GetUpActReserveRelationBindInfo(ctx context.Context, oid string, oType int64, rType int64) (res *like.UpActReserveRelationBind, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_upActReserveRelationBindInfo, oid, oType, rType))
	res = new(like.UpActReserveRelationBind)
	if err = row.Scan(&res.ID, &res.Sid, &res.Oid, &res.OType, &res.Rid, &res.RType); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = fmt.Errorf("GetUpActReserveRelationBindInfo d.db.Query _upActReserveRelationByOid rows.Scan error(%+v)", err)
		}
	}

	return
}

func (d *Dao) IsUpActRelationReservePublished(ctx context.Context, mid int64, types []int64, maxNumLimit int64) (res *like.UpActReserveRelationInfo, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_upActRelationPublishedSQL, mid, xstr.JoinInts(types), maxNumLimit))
	res = new(like.UpActReserveRelationInfo)
	if err = row.Scan(&res.ID, &res.Sid, &res.Mid, &res.Oid, &res.Type, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = fmt.Errorf("GetUpActReserveRelationBindInfo d.db.Query _upActReserveRelationByOid rows.Scan error(%+v)", err)
		}
	}
	return
}

func (d *Dao) UpdateUpActReserveRelationDependAuditState(ctx context.Context, sid, dynamicAudit int64, lotteryAudit int64) (err error) {
	if _, err = d.db.Exec(ctx, _upActRelationDependAuditStateSQL, dynamicAudit, lotteryAudit, sid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec")
		return
	}
	return
}

func (d *Dao) CanUpActOthersReserve(ctx context.Context, mid, sid int64) (res int, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_upActReserveHangRecordInfoSQL, mid, sid))
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "d.db.QueryRow")
		}
	}
	return
}
