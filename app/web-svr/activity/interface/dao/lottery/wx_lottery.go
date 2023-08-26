package lottery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/crc32"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/lottery"

	"github.com/pkg/errors"
)

const (
	_wxLotteryLogSQL    = "SELECT `id`,`mid`,`buvid`,`lottery_id`,`gift_type`,`gift_id`,`gift_name`,`gift_money`,`ctime`,`mtime` FROM wx_lottery_log_%02d WHERE mid=?"
	_wxLotteryLogAddSQL = "INSERT INTO wx_lottery_log_%02d (`mid`,`platform`,`lottery_from`,`buvid`) VALUES (?,?,?,?)"
	_wxLotteryLogUpSQL  = "UPDATE wx_lottery_log_%02d SET `gift_type`=?,`user_type`=?,`gift_id`=?,`gift_name`=?,`gift_money`=?,`lottery_id`=? WHERE id=? AND mid=?"
	_wxLotteryHisSQL    = "SELECT `id`,`mid`,`buvid`,`ctime`,`mtime` FROM wx_lottery_his_%02d WHERE buvid=?"
	_wxLotteryHisAddSQL = "INSERT INTO wx_lottery_his_%02d (`mid`,`buvid`) VALUES (?,?)"
)

func wxLotteryLogKey(mid int64) string {
	return fmt.Sprintf("wx_lott_log_%d", mid)
}

func wxLotteryHisKey(buvid string) string {
	return fmt.Sprintf("wx_lott_his_%s", buvid)
}

func wxRedDotKey(mid int64) string {
	return fmt.Sprintf("wx_lott_red_%d", mid)
}

func wxLotteryHisHit(buvid string) uint32 {
	return crc32.ChecksumIEEE([]byte(buvid)) % 100
}

func (d *Dao) RawWxLotteryLog(ctx context.Context, mid int64) (*lottery.WxLotteryLog, error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_wxLotteryLogSQL, mid%100), mid)
	res := new(lottery.WxLotteryLog)
	if err := row.Scan(&res.ID, &res.Mid, &res.Buvid, &res.LotteryID, &res.GiftType, &res.GiftID, &res.GiftName, &res.GiftMoney, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			err = errors.Wrapf(err, "RawWxLotteryLog mid:%d", mid)
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) RawWxLotteryHisByBuvid(ctx context.Context, buvid string) (*lottery.WxLotteryHis, error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_wxLotteryHisSQL, wxLotteryHisHit(buvid)), buvid)
	res := new(lottery.WxLotteryHis)
	if err := row.Scan(&res.ID, &res.Mid, &res.Buvid, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			err = errors.Wrapf(err, "RawWxLotteryHisByBuvid buvid:%s", buvid)
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) AddWxLotteryLog(ctx context.Context, mid, platform, lotteryFrom int64, buvid string) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_wxLotteryLogAddSQL, mid%100), mid, platform, lotteryFrom, buvid)
	if err != nil {
		return 0, errors.Wrapf(err, "AddWxLotteryLog:dao.db.Exec mid:%d platform:%d buvid:%s", mid, platform, buvid)
	}
	return res.LastInsertId()
}

func (d *Dao) AddWxLotteryHis(ctx context.Context, mid int64, buvid string) (int64, error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_wxLotteryHisAddSQL, wxLotteryHisHit(buvid)), mid, buvid)
	if err != nil {
		return 0, errors.Wrapf(err, "AddWxLotteryHis:dao.db.Exec mid:%d buvid %s", mid, buvid)
	}
	return res.LastInsertId()
}

func (d *Dao) UpWxLotteryLog(ctx context.Context, mid, id, giftType, userType, giftID, money int64, lotteryID, giftName string) error {
	_, err := d.db.Exec(ctx, fmt.Sprintf(_wxLotteryLogUpSQL, mid%100), giftType, userType, giftID, giftName, money, lotteryID, id, mid)
	if err != nil {
		return errors.Wrapf(err, "UpWxLotteryLog:dao.db.Exec id:%d mid:%d", id, mid)
	}
	return nil
}

func (d *Dao) CacheWxLotteryLog(ctx context.Context, mid int64) (*lottery.WxLotteryLog, error) {
	key := wxLotteryLogKey(mid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheWxLotteryLog GET key:%s", key)
	}
	res := new(lottery.WxLotteryLog)
	if err = json.Unmarshal(bs, &res); err != nil {
		return nil, errors.Wrapf(err, "CacheWxLotteryLog json.Unmarshal %s", string(bs))
	}
	return res, nil
}

func (d *Dao) AddCacheWxLotteryLog(ctx context.Context, mid int64, data *lottery.WxLotteryLog) error {
	key := wxLotteryLogKey(mid)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "AddCacheWxLotteryLog json.Marshal %+v", data)
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.wxLotteryLogExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheWxLotteryLog conn.Send(SETEX, %s, %d, %s)", key, d.wxLotteryLogExpire, string(bs))
	}
	return nil
}

func (d *Dao) CacheWxLotteryHisByBuvid(ctx context.Context, buvid string) (*lottery.WxLotteryHis, error) {
	key := wxLotteryHisKey(buvid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheWxLotteryHisByBuvid GET key:%s", key)
	}
	res := new(lottery.WxLotteryHis)
	if err = json.Unmarshal(bs, &res); err != nil {
		return nil, errors.Wrapf(err, "CacheWxLotteryHisByBuvid json.Unmarshal %s", string(bs))
	}
	return res, nil
}

func (d *Dao) AddCacheWxLotteryHisByBuvid(ctx context.Context, buvid string, data *lottery.WxLotteryHis) error {
	key := wxLotteryHisKey(buvid)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "AddCacheWxLotteryHisByBuvid json.Marshal %+v", data)
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.wxLotteryLogExpire, bs); err != nil {
		return errors.Wrapf(err, "AddCacheWxLotteryHisByBuvid conn.Send(SETEX, %s, %d, %s)", key, d.wxLotteryLogExpire, string(bs))
	}
	return nil
}

func (d *Dao) DelCacheWxLotteryLog(ctx context.Context, mid int64, buvid string) error {
	conn := component.GlobalRedis.Conn(ctx)
	defer conn.Close()
	key := wxLotteryLogKey(mid)
	hisKey := wxLotteryHisKey(buvid)
	if err := conn.Send("DEL", key); err != nil {
		return errors.Wrapf(err, "DelCacheWxLotteryLog conn.Send(DEL, %s)", key)
	}
	if err := conn.Send("DEL", hisKey); err != nil {
		return errors.Wrapf(err, "DelCacheWxLotteryLog conn.Send(DEL, %s)", hisKey)
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "DelCacheWxLotteryLog conn.Flush()")
	}
	return nil
}

func (d *Dao) CacheWxRedDot(ctx context.Context, mid int64) (bool, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := wxRedDotKey(mid)
	locked, err := redis.String(conn.Do("SET", key, "1", "NX", "EX", d.wxRedDotExpire))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Error("CacheWxRedDot SETNX key:%s error:%v", key, err)
		return false, err
	}
	if locked == "OK" {
		return true, nil
	}
	return false, nil
}

func (d *Dao) ExpireCacheWxRedDot(ctx context.Context, mid int64) error {
	key := wxLotteryLogKey(mid)
	if _, err := component.GlobalRedis.Do(ctx, "EXPIRE", key, d.wxRedDotExpire); err != nil {
		log.Error("ExpireCacheWxRedDot key:%s expire:%d error:%v", key, d.wxRedDotExpire, err)
		return err
	}
	return nil
}
