package dao

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"github.com/pkg/errors"
	"github.com/siddontang/go-mysql/mysql"
)

const (
	removeHvarRec  = "DELETE FROM %s WHERE mtime < \"%s\" LIMIT %d"
	_sharding      = 100
	_rmPiece       = 1000
	_hvarRmLockKey = "RemoveHvarRec"
)

func tableName(mid int64) string {
	return fmt.Sprintf("hidden_vars_rec_%02d", mid%100)
}

func (d *Dao) RemoveHvarRec(periodValid int) error {
	removed := int64(0)
	expired := time.Now().AddDate(0, 0, -periodValid)
	for index := int64(0); index < _sharding; index++ {
		for {
			result, err := d.db.Exec(context.Background(),
				fmt.Sprintf(removeHvarRec, tableName(index), expired.Format(mysql.TimeFormat), _rmPiece))
			if err != nil {
				return errors.Wrapf(err, "RemoveHvarRec Index %d", index)
			}
			var affected int64
			if affected, err = result.RowsAffected(); err != nil {
				return errors.Wrapf(err, "RemoveHvarRec Index %d", index)
			}
			if affected == 0 { // 删除完毕
				break
			}
			removed += affected
			time.Sleep(1 * time.Second) // 防止主从不同步
			log.Info("RemoveHvarRec periodValid %d, index %d removed %d", periodValid, index, removed)
		}
		log.Info("RemoveHvarRec periodValid %d, index %d removed finished %d", periodValid, index, removed)
	}
	log.Info("RemoveHvarRec periodValid %d, removed finished %d", periodValid, removed)
	return nil
}

// delGraphCache def.
func (d *Dao) DelHvarLock(c context.Context) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", _hvarRmLockKey); err != nil {
		log.Error("RemoveHvarRec DelHvarLock Err %v", err)
	}
	return
}

// GetHvarLock def.
func (d *Dao) GetHvarLock(c context.Context) (bool, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", _hvarRmLockKey, time.Now().Unix(), "EX", d.lockExpire, "NX"))
	if err != nil {
		if err == redis.ErrNil { // 锁已存在，未拿到
			return false, nil
		}
		log.Error("GetHvarLock Err %v", err)
		return false, err
	}
	if reply != "OK" {
		return false, nil
	}
	return true, nil

}
