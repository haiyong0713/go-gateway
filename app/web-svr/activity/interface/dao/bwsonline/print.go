package bwsonline

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"

	"github.com/pkg/errors"
)

const _printListSQL = "SELECT id FROM bws_online_print WHERE state = 1 and bid = ? ORDER BY id ASC"

func (d *Dao) RawPrintList(ctx context.Context, bid int64) ([]int64, error) {
	rows, err := d.db.Query(ctx, _printListSQL, bid)
	if err != nil {
		return nil, errors.Wrap(err, "RawPrintList Query")
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "RawPrintList Scan")
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawPrintList rows")
	}
	return ids, nil
}

func printListKey(bid int64) string {
	return fmt.Sprintf("bws_print_list_%d", bid)
}

func (d *Dao) CachePrintList(ctx context.Context, bid int64) ([]int64, error) {
	data, err := redis.String(d.redis.Do(ctx, "GET", printListKey(bid)))
	if err != nil {
		log.Errorc(ctx, "CacheAwardPackageList GET error:%v", err)
		return nil, err
	}
	ids, err := xstr.SplitInts(data)
	if err != nil {
		log.Errorc(ctx, "CacheAwardPackageList SplitInts data:%s error:%v", data, err)
		return nil, err
	}
	return ids, nil
}

func (d *Dao) AddCachePrintList(ctx context.Context, bid int64, ids []int64) error {
	if _, err := d.redis.Do(ctx, "SETEX", printListKey(bid), d.dataExpire, xstr.JoinInts(ids)); err != nil {
		return err
	}
	return nil
}

const _printSQL = "SELECT id,title,image,piece_id,jump_url,rarity,package_id,intro,ctime,mtime,bid FROM bws_online_print WHERE id=? AND state=1"

func (d *Dao) RawPrint(ctx context.Context, id int64) (*bwsonline.Print, error) {
	data := new(bwsonline.Print)
	row := d.db.QueryRow(ctx, _printSQL, id)
	if err := row.Scan(&data.ID, &data.Title, &data.Image, &data.PieceId, &data.JumpUrl, &data.Level, &data.PackageId, &data.Intro, &data.Ctime, &data.Mtime, &data.Bid); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawPrint:QueryRow")
	}
	return data, nil
}

func printKey(id int64) string {
	return fmt.Sprintf("bws_print_%d", id)
}

func (d *Dao) CachePrint(ctx context.Context, id int64) (*bwsonline.Print, error) {
	key := printKey(id)
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	res := new(bwsonline.Print)
	if err = res.Unmarshal(data); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DelCachePrint(ctx context.Context, id int64) error {
	key := printKey(id)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCachePrint conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddCachePrint(ctx context.Context, id int64, data *bwsonline.Print) error {
	key := printKey(id)
	bytes, err := data.Marshal()
	if err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.dataExpire, bytes); err != nil {
		return err
	}
	return nil
}

const _printByIDsSQL = "SELECT id,title,image,piece_id,jump_url,rarity,package_id,intro,ctime,mtime,bid FROM bws_online_print WHERE id IN(%s) AND state=1"

func (d *Dao) RawPrintByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Print, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_printByIDsSQL, xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawPrintByIDs Query")
	}
	defer rows.Close()
	data := make(map[int64]*bwsonline.Print)
	for rows.Next() {
		r := new(bwsonline.Print)
		if err = rows.Scan(&r.ID, &r.Title, &r.Image, &r.PieceId, &r.JumpUrl, &r.Level, &r.PackageId, &r.Intro, &r.Ctime, &r.Mtime, &r.Bid); err != nil {
			return nil, errors.Wrap(err, "RawPrintByIDs Scan")
		}
		data[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawPrintByIDs rows")
	}
	return data, nil
}

func (d *Dao) CachePrintByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Print, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(printKey(v))
	}
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	res := make(map[int64]*bwsonline.Print, len(ids))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		item := &bwsonline.Print{}
		if err = item.Unmarshal(bs); err != nil {
			log.Errorc(ctx, "CachePrintByIDs Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		res[item.ID] = item
	}
	return res, nil
}

func (d *Dao) AddCachePrintByIDs(ctx context.Context, data map[int64]*bwsonline.Print) error {
	if len(data) == 0 {
		return nil
	}
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string
	for id, v := range data {
		bs, err := v.Marshal()
		if err != nil {
			log.Errorc(ctx, "AddCachePrintByIDs Marshal v:%+v error:%v", v, err)
			continue
		}
		key := printKey(id)
		argsMDs = argsMDs.Add(key).Add(string(bs))
		keys = append(keys, key)
	}
	if err := conn.Send("MSET", argsMDs...); err != nil {
		log.Errorc(ctx, "AddCacheAwardByIDs MSET error(%v)", err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.dataExpire); err != nil {
			log.Errorc(ctx, "AddCacheAwardByIDs conn.Send(Expire, %s, %d) error(%v)", v, d.dataExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCachePrintByIDs Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheAwardByIDs conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

const _addUserPrintSQL = "INSERT INTO bws_online_user_print(mid,print_id,batch_id,state) VALUES(?,?,?,?)"

func (d *Dao) AddUserPrint(ctx context.Context, mid, id, state int64, batchID string) (int64, error) {
	row, err := d.db.Exec(ctx, _addUserPrintSQL, mid, id, batchID, state)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserPrint")
	}
	return row.RowsAffected()
}

const _userPrintSQL = "SELECT print_id,batch_id FROM bws_online_user_print WHERE mid=?"

func (d *Dao) RawUserPrint(ctx context.Context, mid int64) (map[int64]string, error) {
	rows, err := d.db.Query(ctx, _userPrintSQL, mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserPrint Query")
	}
	defer rows.Close()
	res := make(map[int64]string)
	for rows.Next() {
		var (
			printID int64
			batchID string
		)
		if err = rows.Scan(&printID, &batchID); err != nil {
			return nil, errors.Wrap(err, "RawUserPrint Scan")
		}
		res[printID] = batchID
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserPrint rows")
	}
	return res, nil
}

func userPrintKey(mid int64) string {
	return fmt.Sprintf("bws_user_print_%d", mid)
}

func (d *Dao) CacheUserPrint(ctx context.Context, mid int64) (map[int64]string, error) {
	key := userPrintKey(mid)
	values, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Errorc(ctx, "CacheUserPrint conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	res := make(map[int64]string)
	for len(values) > 0 {
		var (
			batchID string
			printID int64
		)
		if values, err = redis.Scan(values, &batchID, &printID); err != nil {
			log.Errorc(ctx, "CacheUserPrint redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		res[printID] = batchID
	}
	return res, nil
}

func (d *Dao) AddCacheUserPrint(ctx context.Context, mid int64, data map[int64]string) error {
	key := userPrintKey(mid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	err := conn.Send("DEL", key)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserPrint conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for i, v := range data {
		args = args.Add(i).Add(v)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(ctx, "AddCacheUserPrint conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Errorc(ctx, "AddCacheUserPrint conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheUserPrint conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheUserPrint conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) DelCacheUserPrint(ctx context.Context, mid int64) error {
	key := userPrintKey(mid)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserAward conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}
