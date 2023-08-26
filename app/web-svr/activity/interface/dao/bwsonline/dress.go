package bwsonline

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"

	"github.com/pkg/errors"
)

const _dressSQL = "SELECT id,title,image,pos,pic_pos,ctime,mtime,group_id FROM bws_online_dress WHERE id=? AND state=1"

func (d *Dao) RawDress(ctx context.Context, id int64) (*bwsonline.Dress, error) {
	data := new(bwsonline.Dress)
	row := d.db.QueryRow(ctx, _dressSQL, id)
	if err := row.Scan(&data.ID, &data.Title, &data.Image, &data.Pos, &data.Key, &data.Ctime, &data.Mtime, &data.GroupID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawDress:QueryRow")
	}
	return data, nil
}

func dressKey(id int64) string {
	return fmt.Sprintf("bws_dress_%d", id)
}

func (d *Dao) CacheDress(ctx context.Context, id int64) (*bwsonline.Dress, error) {
	key := dressKey(id)
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		return nil, err
	}
	res := new(bwsonline.Dress)
	if err = res.Unmarshal(data); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DelCacheDress(ctx context.Context, id int64) error {
	key := dressKey(id)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheDress conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddCacheDress(ctx context.Context, id int64, data *bwsonline.Dress) error {
	key := dressKey(id)
	bytes, err := data.Marshal()
	if err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.dataExpire, bytes); err != nil {
		return err
	}
	return nil
}

const _dressByIDsSQL = "SELECT id,title,image,pos,pic_pos,ctime,mtime,group_id FROM bws_online_dress WHERE id IN (%s) AND state=1"

func (d *Dao) RawDressByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Dress, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_dressByIDsSQL, xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawDressByIDs Query")
	}
	defer rows.Close()
	data := make(map[int64]*bwsonline.Dress)
	for rows.Next() {
		r := new(bwsonline.Dress)
		if err = rows.Scan(&r.ID, &r.Title, &r.Image, &r.Pos, &r.Key, &r.Ctime, &r.Mtime, &r.GroupID); err != nil {
			return nil, errors.Wrap(err, "RawDressByIDs Scan")
		}
		data[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawDressByIDs rows")
	}
	return data, nil
}

func (d *Dao) CacheDressByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Dress, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(dressKey(v))
	}
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	res := make(map[int64]*bwsonline.Dress, len(ids))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		item := &bwsonline.Dress{}
		if err = item.Unmarshal(bs); err != nil {
			log.Errorc(ctx, "CacheAwardByIDs Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		res[item.ID] = item
	}
	return res, nil
}

func (d *Dao) AddCacheDressByIDs(ctx context.Context, data map[int64]*bwsonline.Dress) error {
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
			log.Errorc(ctx, "AddCacheAwardByIDs Marshal v:%+v error:%v", v, err)
			continue
		}
		key := dressKey(id)
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
		log.Errorc(ctx, "AddCacheAwardByIDs Flush error(%v)", err)
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

const _userDressSQL = "SELECT id,mid,dress_id,state,ctime,mtime FROM bws_online_user_dress WHERE mid=?"

func (d *Dao) RawUserDress(ctx context.Context, mid int64) ([]*bwsonline.UserDress, error) {
	rows, err := d.db.Query(ctx, _userDressSQL, mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawDressByIDs Query")
	}
	defer rows.Close()
	var list []*bwsonline.UserDress
	for rows.Next() {
		r := new(bwsonline.UserDress)
		if err = rows.Scan(&r.ID, &r.Mid, &r.DressId, &r.State, &r.Ctime, &r.Mtime); err != nil {
			if err != sql.ErrNoRows {
				err = errors.Wrap(err, "RawUserDress:QueryRow")
				return nil, err
			}
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserDress rows")
	}
	return list, nil
}

func userDressKey(mid int64) string {
	return fmt.Sprintf("bws_user_dress_%d", mid)
}

func (d *Dao) CacheUserDress(ctx context.Context, mid int64) ([]*bwsonline.UserDress, error) {
	key := userDressKey(mid)
	values, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", key, 0, -1))
	if err != nil {
		log.Errorc(ctx, "CacheUserDress conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	var res []*bwsonline.UserDress
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Errorc(ctx, "CacheUserAward redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		item := &bwsonline.UserDress{}
		if err = item.Unmarshal(bs); err != nil {
			log.Errorc(ctx, "CacheUserDress Unmarshal(%v) error(%v)", bs, err)
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (d *Dao) AddCacheUserDress(ctx context.Context, mid int64, data []*bwsonline.UserDress) error {
	key := userDressKey(mid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	err := conn.Send("DEL", key)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserAward conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for _, v := range data {
		var bs []byte
		bs, err = v.Marshal()
		if err != nil {
			log.Errorc(ctx, "AddCacheUserDress Marshal() error(%v)", err)
			return err
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(ctx, "AddCacheUserDress conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Errorc(ctx, "AddCacheUserDress conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheUserDress conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheUserDress conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) DelCacheUserDress(ctx context.Context, mid int64) error {
	var (
		key = userDressKey(mid)
	)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserDress conn.Do(DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _dressOffSQL = "UPDATE bws_online_user_dress SET state=0 WHERE mid=?"

func (d *Dao) DressOff(ctx context.Context, mid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _dressOffSQL, mid)
	if err != nil {
		return 0, errors.Wrap(err, "DressOff")
	}
	return row.RowsAffected()
}

const _dressUpSQL = "UPDATE bws_online_user_dress SET state=1 WHERE mid=? AND dress_id IN (%s)"

func (d *Dao) DressUp(ctx context.Context, mid int64, ids []int64) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_dressUpSQL, xstr.JoinInts(ids)), mid)
	if err != nil {
		return 0, errors.Wrap(err, "DressUp")
	}
	return row.RowsAffected()
}

const _dressAddSQL = "INSERT INTO bws_online_user_dress (mid,dress_id) VALUES %s"

func (d *Dao) DressAdd(ctx context.Context, mid int64, ids []int64) (int64, error) {
	args := make([]interface{}, 0, len(ids)*2)
	placeholder := strings.TrimRight(strings.Repeat("(?,?),", len(ids)), ",")
	for _, id := range ids {
		args = append(args, mid, id)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_dressAddSQL, placeholder), args...)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserAward Exec")
	}
	return row.RowsAffected()
}
