package bwsonline

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/time"

	"github.com/pkg/errors"
)

const _userCurrencySQL = "SELECT type_id,amount FROM bws_online_user_currency WHERE mid=? and bid=?"

func (d *Dao) RawUserCurrency(ctx context.Context, mid int64, bid int64) (map[int64]int64, error) {
	rows, err := d.db.Query(ctx, _userCurrencySQL, mid, bid)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserCurrency Query")
	}
	defer rows.Close()
	data := make(map[int64]int64)
	for rows.Next() {
		var typeID, amount int64
		if err = rows.Scan(&typeID, &amount); err != nil {
			return nil, errors.Wrap(err, "RawUserCurrency Scan")
		}
		data[typeID] = amount
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserCurrency rows")
	}
	return data, nil
}

func userCurrencyKey(mid, bid int64) string {
	return fmt.Sprintf("bws_user_curr_v1_%d_%d", mid, bid)
}

func (d *Dao) CacheUserCurrency(ctx context.Context, mid int64, bid int64) (map[int64]int64, error) {
	key := userCurrencyKey(mid, bid)
	values, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Errorc(ctx, "CacheUserPrint conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	res := make(map[int64]int64)
	for len(values) > 0 {
		var amount, typeID int64
		if values, err = redis.Scan(values, &typeID, &amount); err != nil {
			log.Errorc(ctx, "CacheUserPrint redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		res[typeID] = amount
	}
	return res, nil
}

func (d *Dao) AddCacheUserCurrency(ctx context.Context, mid int64, data map[int64]int64, bid int64) error {
	key := userCurrencyKey(mid, bid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	err := conn.Send("DEL", key)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserCurrency conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for i, v := range data {
		args = args.Add(v).Add(i)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(ctx, "AddCacheUserCurrency conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Errorc(ctx, "AddCacheUserCurrency conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheUserCurrency conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheUserCurrency conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) DelCacheUserCurrency(ctx context.Context, mid, bid int64) error {
	var (
		key = userCurrencyKey(mid, bid)
	)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserPackage conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _lastAutoEnergySQL = "SELECT ctime FROM bws_online_user_currency_log WHERE mid=? AND bid=? AND type_id=1 AND add_type=1 ORDER BY ID DESC LIMIT 1"

func (d *Dao) RawLastAutoEnergy(ctx context.Context, mid, bid int64) (int64, error) {
	row := d.db.QueryRow(ctx, _lastAutoEnergySQL, mid, bid)
	var lastTime time.Time
	if err := row.Scan(&lastTime); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrap(err, "RawLastAutoEnergy")
	}
	return lastTime.Time().Unix(), nil
}

func lastAutoEnergyKey(mid, bid int64) string {
	return fmt.Sprintf("bws_auto_en_%d_%d", mid, bid)
}

func (d *Dao) CacheLastAutoEnergy(ctx context.Context, mid, bid int64) (int64, error) {
	key := lastAutoEnergyKey(mid, bid)
	res, err := redis.Int64(d.redis.Do(ctx, "GET", lastAutoEnergyKey(mid, bid)))
	if err != nil {
		log.Errorc(ctx, "CacheLastAutoEnergy GET key:%s error:%v", key, err)
		return 0, err
	}
	return res, nil
}

func (d *Dao) AddCacheLastAutoEnergy(ctx context.Context, mid int64, data int64, bid int64) error {
	key := lastAutoEnergyKey(mid, bid)
	if _, err := d.redis.Do(ctx, "SETEX", key, d.dataExpire, data); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DelCacheLastAutoEnergy(ctx context.Context, mid, bid int64) error {
	key := lastAutoEnergyKey(mid, bid)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheLastAutoEnergy conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _addUserCurrLogSQL = "INSERT INTO bws_online_user_currency_log (mid,type_id,add_type,change_amount,bid) VALUES(?,?,?,?,?)"

func (d *Dao) AddUserCurrencyLog(ctx context.Context, mid, typeID, addType, changeAmount, bid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _addUserCurrLogSQL, mid, typeID, addType, changeAmount, bid)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserCurrencyLog")
	}
	return row.LastInsertId()
}

const _upUserCurrencySQL = "INSERT INTO bws_online_user_currency (mid,type_id,amount,bid) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE amount=amount+?"

func (d *Dao) UpUserCurrency(ctx context.Context, mid, typeID, amount, bid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _upUserCurrencySQL, mid, typeID, amount, bid, amount)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserCurrency")
	}
	return row.RowsAffected()
}
