package like

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/job/component"
	"net/url"
	"strconv"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	l "go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const (
	_lotteryListStateSQL  = "SELECT id,info,sid FROM act_lottery_times WHERE info != ? AND type = ? AND `state`= 0"
	_lotteryAllSQL        = "SELECT id,lottery_id FROM act_lottery WHERE stime<? AND etime>? AND `state`= 0"
	_goAddLotteryTimesURI = "/x/internal/activity/lottery/addtimes"
	_lotteryLikeKey       = "go:lottery_like"
	_lotteryCustomizeKey  = "go:lottery_customize"
	_lotteryVipKey        = "go:lottery_vip"
	_lotteryOGVKey        = "go:lottery_ogv"
	actionKey             = "action"
	timesKey              = "times"
)

// LotteryList get lottery list.
func (dao *Dao) RawLotteryLikeList(c context.Context, ltype int) (list map[string]*l.Lottery, err error) {
	rows, err := dao.db.Query(c, _lotteryListStateSQL, "", ltype)
	if err != nil {
		err = errors.Wrapf(err, "RawLotteryLikeList:d.db.Query()")
		return
	}
	defer rows.Close()
	list = make(map[string]*l.Lottery)
	for rows.Next() {
		ll := &l.Lottery{}
		if err = rows.Scan(&ll.ID, &ll.Info, &ll.Sid); err != nil {
			err = errors.Wrapf(err, "RawLotteryLikeList:row.Scan row (%v)", ll)
			return
		}
		list[ll.Info] = ll
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RawLotteryLikeList:rowsErr()")
	}
	return
}

// RawLotteryAddTimesList get lottery list.
func (dao *Dao) RawLotteryAddTimesList(c context.Context, ltype int) (list []*l.Lottery, err error) {
	rows, err := dao.db.Query(c, _lotteryListStateSQL, "", ltype)
	if err != nil {
		err = errors.Wrapf(err, "RawLotteryAddTimesList:d.db.Query()")
		return
	}
	defer rows.Close()
	list = make([]*l.Lottery, 0)
	for rows.Next() {
		ll := &l.Lottery{}
		if err = rows.Scan(&ll.ID, &ll.Info, &ll.Sid); err != nil {
			err = errors.Wrapf(err, "RawLotteryAddTimesList:row.Scan row (%v)", ll)
			return
		}
		list = append(list, ll)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RawLotteryAddTimesList:rowsErr()")
	}
	return
}

// RawLotteryAllList get lottery list.
func (dao *Dao) RawLotteryAllList(c context.Context) (list []*l.LotteryDetail, err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rows, err := dao.db.Query(c, _lotteryAllSQL, now, now)
	if err != nil {
		err = errors.Wrapf(err, "RawLotteryAllList:d.db.Query()")
		return
	}
	defer rows.Close()
	list = make([]*l.LotteryDetail, 0)
	for rows.Next() {
		ll := &l.LotteryDetail{}
		if err = rows.Scan(&ll.ID, &ll.Sid); err != nil {
			err = errors.Wrapf(err, "RawLotteryAllList:row.Scan row (%v)", ll)
			return
		}
		list = append(list, ll)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "RawLotteryAllList:rowsErr()")
	}
	return
}

func (dao *Dao) CacheLotteryList(c context.Context, ltype int) (res map[string]*l.Lottery, err error) {
	var (
		bs []byte
	)
	key := lotteryTypeKey(ltype)
	if key == "" {
		return
	}
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLotteryList(%s) return nil", key)
		} else {
			log.Error("CacheLotteryList conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	res = make(map[string]*l.Lottery)
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (dao *Dao) AddCacheLotteryList(c context.Context, list map[string]*l.Lottery, ltype int) (err error) {
	var (
		bs []byte
	)
	key := lotteryTypeKey(ltype)
	if key == "" || len(list) == 0 {
		return
	}
	if bs, err = json.Marshal(list); err != nil {
		log.Error("json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, dao.lotteryExpire, bs); err != nil {
		log.Error("AddCacheLotteryList conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.lotteryExpire, string(bs), err)
	}
	return
}

func (dao *Dao) LotteryLikeList(c context.Context, ltype int) (res map[string]*l.Lottery, err error) {
	addCache := true
	res, err = dao.CacheLotteryList(c, ltype)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("LotteryLikeList")
		return
	}
	cache.MetricMisses.Inc("LotteryLikeList")
	res, err = dao.RawLotteryLikeList(c, ltype)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(ctx context.Context) {
		dao.AddCacheLotteryList(c, miss, ltype)
	})
	return
}

// Golang AddLotteryTimes .
func (dao *Dao) GoAddLotteryTimes(c context.Context, sid string, cid, mid int64, actionType int, orderNo string) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("sid", sid)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("action_type", strconv.Itoa(actionType))
	params.Set("order_no", orderNo)
	params.Set("cid", strconv.FormatInt(cid, 10))
	if err = dao.httpClient.Post(c, dao.lotteryAddTimesURL, "", params, &res); err != nil {
		log.Error("GoAddLotteryTimes:d.httpClient.Post sid(%s) mid(%d) type(%d) error(%v)", sid, mid, actionType, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), dao.lotteryAddTimesURL+"?"+params.Encode())
	}
	return
}

func lotteryTypeKey(ltype int) (key string) {
	switch ltype {
	case l.LotteryArcType:
		key = _lotteryLikeKey
	case l.LotteryCustomizeType:
		key = _lotteryCustomizeKey
	case l.LotteryVip:
		key = _lotteryVipKey
	case l.LotteryOgvType:
		key = _lotteryOGVKey
	default:
		return
	}
	return
}

// DeleteLotteryActionLog ...
func (dao *Dao) DeleteLotteryActionLog(c context.Context, sid int64, mid int64) (err error) {
	conn := dao.redis.Get(c)
	defer conn.Close()
	var (
		key = buildKeyLottery(actionKey, sid, mid)
	)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "DeleteLotteryActionLog conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// NewDeleteLotteryActionLog ...
func (dao *Dao) NewDeleteLotteryActionLog(c context.Context, sid int64, mid int64) (err error) {
	conn := dao.redisNew.Get(c)
	defer conn.Close()
	var (
		key = buildKeyLottery(actionKey, sid, mid)
	)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "NewDeleteLotteryActionLog conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// DeleteCacheLotteryTimes ...
func (dao *Dao) DeleteCacheLotteryTimes(c context.Context, sid int64, mid int64, remark string) (err error) {
	conn := dao.redis.Get(c)
	defer conn.Close()
	var (
		key = buildKeyLottery(timesKey, sid, remark, mid)
	)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "DeleteCacheLotteryTimes conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// NewDeleteCacheLotteryTimes ...
func (dao *Dao) NewDeleteCacheLotteryTimes(c context.Context, sid int64, mid int64, remark string) (err error) {
	conn := dao.redisNew.Get(c)
	defer conn.Close()
	var (
		key = buildKeyLottery(timesKey, sid, remark, mid)
	)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "NewDeleteCacheLotteryTimes conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

func lotteryTimesKey(sid, mid int64, remark string) string {
	return fmt.Sprintf("lottery_times_%s_%d_%d", remark, sid, mid)
}

// DeleteOldCacheLotteryTimes ...
func (dao *Dao) DeleteOldCacheLotteryTimes(c context.Context, sid int64, mid int64, remark string) (err error) {
	conn := dao.redis.Get(c)
	defer conn.Close()
	var (
		key = lotteryTimesKey(sid, mid, remark)
	)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorc(c, "DeleteOldCacheLotteryTimes conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// LotteryAddTimesPub 自定义流
func (dao *Dao) LotteryAddTimesPub(ctx context.Context, mid int64, data *l.LotteryAddTimesMsg) (err error) {
	midStr := fmt.Sprintf("%d", mid)

	buf, _ := json.Marshal(*data)
	if err = component.LotteryAddTimesPub.Send(ctx, midStr, buf); err != nil {
		log.Errorc(ctx, "LotteryAddTimesPub error : dao.lotteryAddtimesPub.Send(%d,%v) err(%v)", mid, *data, err)
		return err
	}
	log.Infoc(ctx, "LotteryAddTimesPub success : dao.lotteryAddtimesPub.Send(%d,%v)", mid, *data)
	return
}
