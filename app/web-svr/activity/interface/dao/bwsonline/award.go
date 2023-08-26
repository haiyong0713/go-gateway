package bwsonline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"

	"github.com/pkg/errors"
)

const _awardPackageListSQL = "SELECT id FROM bws_online_award_package WHERE state=1 AND type_id=0 AND bid=? ORDER BY id ASC"

func (d *Dao) RawAwardPackageList(ctx context.Context, bid int64) ([]int64, error) {
	rows, err := d.db.Query(ctx, _awardPackageListSQL, bid)
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageList Query")
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "RawAwardPackageList Scan")
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageList rows")
	}
	return ids, nil
}

func awardPackageListKey(bid int64) string {
	return fmt.Sprintf("award_pack_list_%d", bid)
}

func (d *Dao) CacheAwardPackageList(ctx context.Context, bid int64) ([]int64, error) {
	data, err := redis.String(d.redis.Do(ctx, "GET", awardPackageListKey(bid)))
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

func (d *Dao) AddCacheAwardPackageList(ctx context.Context, bid int64, ids []int64) error {
	if _, err := d.redis.Do(ctx, "SETEX", awardPackageListKey(bid), d.dataExpire, xstr.JoinInts(ids)); err != nil {
		return err
	}
	return nil
}

const _awardPackageSQL = "SELECT id,title,intro,price,type_id,award_ids,ctime,mtime,bid FROM bws_online_award_package WHERE id=? AND state=1"

func (d *Dao) RawAwardPackage(ctx context.Context, id int64) (*bwsonline.AwardPackage, error) {
	row := d.db.QueryRow(ctx, _awardPackageSQL, id)
	data := new(bwsonline.AwardPackage)
	var awardIDStr string
	if err := row.Scan(&data.ID, &data.Title, &data.Intro, &data.Price, &data.TypeId, &awardIDStr, &data.Ctime, &data.Mtime, &data.Bid); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawAwardPackage Scan")
	}
	data.AwardIds, _ = xstr.SplitInts(awardIDStr)
	return data, nil
}

func awardPackageKey(id int64) string {
	return fmt.Sprintf("award_pack_%d", id)
}

func (d *Dao) CacheAwardPackage(ctx context.Context, id int64) (*bwsonline.AwardPackage, error) {
	key := awardPackageKey(id)
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	res := new(bwsonline.AwardPackage)
	if err = res.Unmarshal(data); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DelCacheAwardPackage(ctx context.Context, id int64) error {
	key := awardPackageKey(id)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheAwardPackage d.redis.Do(ctx, DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddCacheAwardPackage(ctx context.Context, id int64, data *bwsonline.AwardPackage) error {
	key := awardPackageKey(id)
	bytes, err := data.Marshal()
	if err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.dataExpire, bytes); err != nil {
		return err
	}
	return nil
}

const _awardPackageByIDsSQL = "SELECT id,title,intro,price,award_ids,ctime,mtime,bid FROM bws_online_award_package WHERE id IN(%s) AND state=1"

func (d *Dao) CacheAwardPackageByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.AwardPackage, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(awardPackageKey(v))
	}
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	res := make(map[int64]*bwsonline.AwardPackage, len(ids))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		item := &bwsonline.AwardPackage{}
		if err = item.Unmarshal(bs); err != nil {
			log.Errorc(ctx, "CacheAwardPackageByIDs Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		res[item.ID] = item
	}
	return res, nil
}

func (d *Dao) AddCacheAwardPackageByIDs(ctx context.Context, data map[int64]*bwsonline.AwardPackage) error {
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
			log.Errorc(ctx, "AddCacheAwardPackageByIDs Marshal v:%+v error:%v", v, err)
			continue
		}
		key := awardPackageKey(id)
		argsMDs = argsMDs.Add(key).Add(string(bs))
		keys = append(keys, key)
	}
	if err := conn.Send("MSET", argsMDs...); err != nil {
		log.Errorc(ctx, "AddCacheAwardPackageByIDs MSET error(%v)", err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.dataExpire); err != nil {
			log.Errorc(ctx, "AddCacheAwardPackageByIDs conn.Send(Expire, %s, %d) error(%v)", v, d.dataExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheAwardPackageByIDs Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheSubjectRulesBySids conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) RawAwardPackageByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.AwardPackage, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_awardPackageByIDsSQL, xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageByIDs Query")
	}
	defer rows.Close()
	res := make(map[int64]*bwsonline.AwardPackage)
	for rows.Next() {
		r := new(bwsonline.AwardPackage)
		var awardIDStr string
		if err = rows.Scan(&r.ID, &r.Title, &r.Intro, &r.Price, &awardIDStr, &r.Ctime, &r.Mtime, &r.Bid); err != nil {
			return nil, errors.Wrap(err, "RawAwardPackageByIDs Scan")
		}
		r.AwardIds, _ = xstr.SplitInts(awardIDStr)
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawAwardPackageByIDs rows")
	}
	return res, nil
}

const _userPackage = "SELECT package_id, award_ids FROM bws_online_user_award_package WHERE mid=?"

func (d *Dao) RawUserPackage(ctx context.Context, mid int64) ([]*bwsonline.AwardPackage, error) {
	rows, err := d.db.Query(ctx, _userPackage, mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserPackage Query")
	}
	defer rows.Close()
	var res []*bwsonline.AwardPackage
	for rows.Next() {
		var id int64
		var awardStr string
		if err = rows.Scan(&id, &awardStr); err != nil {
			return nil, errors.Wrap(err, "RawUserPackage Scan")
		}
		awardIDs, e := xstr.SplitInts(awardStr)
		if e != nil {
			log.Errorc(ctx, "RawUserPackage xstr.SplitInts(%s) err[%v]", awardStr, e)
		}
		res = append(res, &bwsonline.AwardPackage{
			ID:       id,
			AwardIds: awardIDs,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserPackage rows")
	}
	return res, nil
}

func userPackageKey(mid int64) string {
	return fmt.Sprintf("bws_user_pack_%d", mid)
}

func (d *Dao) CacheUserPackage(ctx context.Context, mid int64) ([]*bwsonline.AwardPackage, error) {
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", userPackageKey(mid)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Errorc(ctx, "CacheUserPackage GET error:%v", err)
		return nil, err
	}
	res := make([]*bwsonline.AwardPackage, 0, 10)
	err = json.Unmarshal(data, res)
	if err != nil {
		log.Errorc(ctx, "CacheUserPackage json.Unmarshal data:%s error:%v", data, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheUserPackage(ctx context.Context, mid int64, awards []*bwsonline.AwardPackage) error {
	data, err := json.Marshal(awards)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserPackage json.Marshal data:%s error:%v", data, err)
		return err
	}
	if _, err := d.redis.Do(ctx, "SETEX", userPackageKey(mid), d.userExpire, data); err != nil {
		return err
	}
	return nil
}

func (d *Dao) DelCacheUserPackage(ctx context.Context, mid int64) error {
	var (
		key = userPackageKey(mid)
	)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserPackage d.redis.Do(ctx, DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _addUserPackageSQL = "INSERT INTO bws_online_user_award_package (mid,package_id,award_ids) VALUES (?,?,?)"

func (d *Dao) AddUserAwardPackage(ctx context.Context, mid, id int64, awards []int64) (int64, error) {
	row, err := d.db.Exec(ctx, _addUserPackageSQL, mid, id, xstr.JoinInts(awards))
	if err != nil {
		return 0, errors.Wrap(err, "AddUserAwardPackage")
	}
	return row.LastInsertId()
}

const _awardSQL = "SELECT id,title,intro,image,type_id,num,token,expire_time,ctime,mtime FROM bws_online_award WHERE id =? AND state = 1"

func (d *Dao) RawAward(ctx context.Context, id int64) (*bwsonline.Award, error) {
	row := d.db.QueryRow(ctx, _awardSQL, id)
	data := new(bwsonline.Award)
	if err := row.Scan(&data.ID, &data.Title, &data.Intro, &data.Image, &data.TypeId, &data.Num, &data.Token, &data.Expire, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawAward Scan")
	}
	return data, nil
}

func awardKey(id int64) string {
	return fmt.Sprintf("bws_award_%d", id)
}

func (d *Dao) CacheAward(ctx context.Context, id int64) (*bwsonline.Award, error) {
	key := awardKey(id)
	data, err := redis.Bytes(d.redis.Do(ctx, "GET", key))
	if err != nil {
		return nil, err
	}
	res := new(bwsonline.Award)
	if err = res.Unmarshal(data); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DelCacheAward(ctx context.Context, id int64) error {
	key := awardKey(id)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheAward d.redis.Do(ctx, DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddCacheAward(ctx context.Context, id int64, data *bwsonline.Award) error {
	key := awardKey(id)
	bytes, err := data.Marshal()
	if err != nil {
		return err
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.dataExpire, bytes); err != nil {
		return err
	}
	return nil
}

const _awardByIDsSQL = "SELECT id,title,intro,image,type_id,num,token,expire_time,ctime,mtime FROM bws_online_award WHERE id IN(%s) AND state = 1"

func (d *Dao) RawAwardByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Award, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(_awardByIDsSQL, xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawAwardByIDs Query")
	}
	defer rows.Close()
	res := make(map[int64]*bwsonline.Award)
	for rows.Next() {
		r := new(bwsonline.Award)
		if err = rows.Scan(&r.ID, &r.Title, &r.Intro, &r.Image, &r.TypeId, &r.Num, &r.Token, &r.Expire, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawAwardByIDs Scan")
		}
		res[r.ID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawAwardByIDs rows")
	}
	return res, nil
}

func (d *Dao) CacheAwardByIDs(ctx context.Context, ids []int64) (map[int64]*bwsonline.Award, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := redis.Args{}
	for _, v := range ids {
		args = args.Add(awardKey(v))
	}
	bss, err := redis.ByteSlices(d.redis.Do(ctx, "MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	res := make(map[int64]*bwsonline.Award, len(ids))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		item := &bwsonline.Award{}
		if err = item.Unmarshal(bs); err != nil {
			log.Errorc(ctx, "CacheAwardByIDs Unmarshal(%s) error(%v)", string(bs), err)
			continue
		}
		res[item.ID] = item
	}
	return res, nil
}

func (d *Dao) AddCacheAwardByIDs(ctx context.Context, data map[int64]*bwsonline.Award) error {
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
		key := awardKey(id)
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

const _userAwardSQL = "SELECT award_id,state FROM bws_online_user_award WHERE mid=? and bid=?"

func (d *Dao) RawUserAward(ctx context.Context, mid, bid int64) ([]*bwsonline.UserAward, error) {
	rows, err := d.db.Query(ctx, _userAwardSQL, mid, bid)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserAward Query")
	}
	defer rows.Close()
	var res []*bwsonline.UserAward
	for rows.Next() {
		var awardID, state int64
		if err = rows.Scan(&awardID, &state); err != nil {
			return nil, errors.Wrap(err, "RawUserAward Scan")
		}
		res = append(res, &bwsonline.UserAward{Award: &bwsonline.Award{ID: awardID}, State: state})
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawUserAward rows")
	}
	return res, nil
}

func userAwardKey(mid, bid int64) string {
	return fmt.Sprintf("bws_user_award_%d_%d", mid, bid)
}

func (d *Dao) CacheUserAward(ctx context.Context, mid, bid int64) ([]*bwsonline.UserAward, error) {
	key := userAwardKey(mid, bid)
	values, err := redis.Values(d.redis.Do(ctx, "ZREVRANGE", key, 0, -1))
	if err != nil {
		log.Errorc(ctx, "CacheUserAward d.redis.Do(ctx, ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	var res []*bwsonline.UserAward
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Errorc(ctx, "CacheUserAward redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		item := &bwsonline.UserAward{}
		if err = json.Unmarshal(bs, item); err != nil {
			log.Errorc(ctx, "CacheUserAward json.Unmarshal(%v) error(%v)", bs, err)
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (d *Dao) AddCacheUserAward(ctx context.Context, mid int64, data []*bwsonline.UserAward, bid int64) error {
	key := userAwardKey(mid, bid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	err := conn.Send("DEL", key)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserAward conn.Send(DEL, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for _, v := range data {
		if v.Award == nil {
			continue
		}
		var bs []byte
		bs, err = json.Marshal(v)
		if err != nil {
			log.Errorc(ctx, "AddCacheUserAward json.Marshal() error(%v)", err)
			return err
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(ctx, "AddCacheUserAward conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Send("EXPIRE", key, d.userExpire); err != nil {
		log.Errorc(ctx, "AddCacheUserAward conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(ctx, "AddCacheUserAward conn.Flush error(%v)", err)
		return err
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(ctx, "AddCacheUserAward conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) DelCacheUserAward(ctx context.Context, mid, bid int64) error {
	var (
		key = userAwardKey(mid, bid)
	)
	if _, err := d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "DelCacheUserAward d.redis.Do(ctx, DEL, %s) error(%v)", key, err)
		return err
	}
	return nil
}

const _addUserAwardSQL = "INSERT IGNORE INTO bws_online_user_award (mid,award_id,bid) VALUES %s"

func (d *Dao) AddUserAward(ctx context.Context, mid int64, ids []int64, bid int64) (int64, error) {
	var (
		rows    []interface{}
		rowsTmp []string
	)
	for _, id := range ids {
		rowsTmp = append(rowsTmp, "(?,?,?)")
		rows = append(rows, mid, id, bid)
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_addUserAwardSQL, strings.Join(rowsTmp, ",")), rows...)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserAward Exec")
	}
	return row.RowsAffected()
}

const _upUserAwardSQL = "UPDATE bws_online_user_award SET state=1 WHERE mid=? AND award_id=?"

func (d *Dao) UpUserAward(ctx context.Context, mid, id int64) (int64, error) {
	row, err := d.db.Exec(ctx, _upUserAwardSQL, mid, id)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserAward")
	}
	return row.RowsAffected()
}
