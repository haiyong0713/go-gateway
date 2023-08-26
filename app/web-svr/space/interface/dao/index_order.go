package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_indexOrderKeyFmt = "spc_io_%d"
	_indexOrderSQL    = `SELECT index_order FROM dede_member_up_settings%d WHERE mid = ?`
	_indexOrderAddSQL = `INSERT INTO dede_member_up_settings%d (mid,index_order) VALUES (?,?) ON DUPLICATE KEY UPDATE index_order = ?`
)

// nolint:gomnd
func indexOrderHit(mid int64) int64 {
	return mid % 10
}

func indexOrderKey(mid int64) string {
	return fmt.Sprintf(_indexOrderKeyFmt, mid)
}

// IndexOrder get index order info.
func (d *Dao) IndexOrder(c context.Context, mid int64) (indexOrder string, err error) {
	var row = d.db.QueryRow(c, fmt.Sprintf(_indexOrderSQL, indexOrderHit(mid)), mid)
	if err = row.Scan(&indexOrder); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("IndexOrder row.Scan() error(%v)", err)
		}
	}
	return
}

// IndexOrderModify index order modify.
func (d *Dao) IndexOrderModify(c context.Context, mid int64, orderStr string) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_indexOrderAddSQL, indexOrderHit(mid)), mid, orderStr, orderStr); err != nil {
		log.Error("IndexOrderModify error d.db.Exec(%d,%s) error(%v)", mid, orderStr, err)
	}
	return
}

// IndexOrderCache get index order cache.
func (d *Dao) IndexOrderCache(c context.Context, mid int64) ([]*model.IndexOrder, error) {
	key := indexOrderKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("IndexOrderCache conn.Do(GET,%s) error(%v)", key, err)
		return nil, err
	}
	var data []*model.IndexOrder
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("IndexOrderCache json.Unmarshal(%s) error(%v)", string(bs), err)
		return nil, err
	}
	return data, nil
}

// SetIndexOrderCache set index order cache.
func (d *Dao) SetIndexOrderCache(c context.Context, mid int64, data []*model.IndexOrder) error {
	key := indexOrderKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("SetIndexOrderCache json.Marshal mid:%d req(%+v) error(%v)", mid, data, err)
		return err
	}
	if _, err = conn.Do("SETEX", key, d.settingExpire, bs); err != nil {
		log.Error("SetIndexOrderCache conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}

// DelIndexOrderCache delete index order cache.
func (d *Dao) DelIndexOrderCache(c context.Context, mid int64) (err error) {
	key := indexOrderKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}
