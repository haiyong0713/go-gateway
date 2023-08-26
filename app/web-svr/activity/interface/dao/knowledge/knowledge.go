package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/knowledge"

	"go-common/library/cache/redis"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao  -sync=true  -nullcache=&knowledge.UserInfo{MID:-1} -check_null_code=$.MID==-1
	UsersMid(c context.Context, mid int64, sid int64) (*knowledge.UserInfo, error)
}

// CacheUsersMid ...
func (d *Dao) CacheUsersMid(ctx context.Context, mid, sid int64) (res *knowledge.UserInfo, err error) {

	var (
		key = buildKey(sid, mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return nil, nil
		} else {
			log.Errorc(ctx, "CacheUsersMid conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(ctx, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

const (
	_RawUsersMid = "SELECT mid,archive_count,single_view,all_view from %s WHERE mid = ?"
)

// RawUsersMid 查询用户信息
func (d *Dao) RawUsersMid(ctx context.Context, mid, sid int64) (res *knowledge.UserInfo, err error) {
	res = new(knowledge.UserInfo)
	tableName := fmt.Sprintf("act_knowledge_user_%d", sid)
	row := d.db.QueryRow(ctx, fmt.Sprintf(_RawUsersMid, tableName), mid)
	if err = row.Scan(&res.MID, &res.ArchiveCount, &res.SingleView, &res.AllView); err != nil {
		if err == sql.ErrNoRows {
			return &knowledge.UserInfo{}, nil
		}
		log.Errorc(ctx, "RawUsersMid mid(%d) err(%v)", mid, err)
	}
	return
}

// AddCacheUsersMid 添加缓存
func (d *Dao) AddCacheUsersMid(ctx context.Context, mid int64, user *knowledge.UserInfo, sid int64) (err error) {
	var (
		key = buildKey(sid, mid)
		bs  []byte
	)
	if bs, err = json.Marshal(user); err != nil {
		log.Errorc(ctx, "json.Marshal(%+v) error (%v)", user, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.FiveMinutesExpire, bs); err != nil {
		log.Errorc(ctx, "AddCacheUsersMid conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.FiveMinutesExpire, string(bs), err)
	}
	return
}
