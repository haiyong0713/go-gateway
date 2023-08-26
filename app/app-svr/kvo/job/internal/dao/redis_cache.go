package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_rdsuserConfKeyPrefix = "rd_u"
	_rdsdocKeyPrefix      = "rd_d_v1"
	_rdsuserDocKeyPrefix  = "dm_u_doc"
	_rdsUcKeyPrefix       = "rd_uc_"
)

func rdsUserConfKey(mid int64, moduleKey int) string {
	return fmt.Sprintf("%s_%d_%d", _rdsuserConfKeyPrefix, mid, moduleKey)
}

func rdsDocKey(checkSum int64) string {
	return fmt.Sprintf("%s_%d", _rdsdocKeyPrefix, checkSum)
}

func keyUserDoc(mid int64) string {
	return fmt.Sprintf("%s_%d", _rdsuserDocKeyPrefix, mid)
}

func rdsUcKey(mid int64, buvid string, moduleKey int) string {
	if mid > 0 {
		buvid = "$"
	}
	return fmt.Sprintf("%s_%d_%s_%d", _rdsUcKeyPrefix, mid, buvid, moduleKey)
}

// UserConfCache user config cache
func (d *dao) UserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int) (bs []byte, err error) {
	bs, err = redis.Bytes(d.redis.Do(ctx, "GET", rdsUcKey(mid, buvid, moduleKey)))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			bs = nil
			return
		}
		log.Error("d.UserConfRds(mid:%d,modulekey:%d,buvid:%s) error(%v)", mid, moduleKey, buvid, err)
	}
	return
}

// SetUserConfCache set user config cache
func (d *dao) SetUserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int, bs []byte) (err error) {
	if _, err = d.redis.Do(ctx, "SET", rdsUcKey(mid, buvid, moduleKey), bs, "EX", d.rdsUcDocExpire); err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,buvid:%s) error(%v)", mid, moduleKey, buvid, err)
	}
	return
}

// UserConfCache user config cache
func (d *dao) UserConfRds(ctx context.Context, mid int64, moduleKey int) (uc *model.UserConf, err error) {
	var bs []byte
	bs, err = redis.Bytes(d.redis.Do(ctx, "GET", rdsUserConfKey(mid, moduleKey)))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			uc = nil
			return
		}
		log.Error("d.UserConfRds(mid:%d,modulekey:%d) error(%v)", mid, moduleKey, err)
		return
	}
	if err = json.Unmarshal(bs, &uc); err != nil {
		log.Error("d.UserConfRds(mid:%d,modulekey:%d) error(%v)", mid, moduleKey, err)
	}
	return
}

// SetUserConfCache set user config cache
func (d *dao) SetUserConfRds(ctx context.Context, uc *model.UserConf) (err error) {
	bs, err := json.Marshal(uc)
	if err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,uc:%+v) error(%v)", uc.Mid, uc.ModuleKey, uc, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SET", rdsUserConfKey(uc.Mid, uc.ModuleKey), bs, "EX", d.rdsExpire); err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,uc:%+v) error(%v)", uc.Mid, uc.ModuleKey, uc, err)
	}
	return

}

// DelUserConfCache del user config cache
func (d *dao) DelUserConfRds(ctx context.Context, mid int64, moduleKey int) (err error) {
	var (
		key = rdsUserConfKey(mid, moduleKey)
	)
	if _, err = d.redis.Do(ctx, "DEL", key); err != nil {
		log.Error("d.DelUserConfRds(mid:%d,modulekey:%d) err(%v)", mid, moduleKey, err)
	}
	return
}

func (d *dao) DocumentRds(ctx context.Context, checkSum int64) (data json.RawMessage, err error) {
	var bs []byte
	bs, err = redis.Bytes(d.redis.Do(ctx, "GET", rdsDocKey(checkSum)))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			data = nil
			return
		}
		log.Error("d.DocumentRds(checksum:%d) err(%v)", checkSum, err)
		return
	}
	data = json.RawMessage(bs)
	return
}

func (d *dao) DelDocumentRds(ctx context.Context, checkSum int64) (err error) {
	var (
		key = rdsDocKey(checkSum)
	)
	if _, err = d.redis.Do(ctx, "DEL", key); err != nil {
		log.Error("d.DelDocumentRds(checksum:%d) err(%v)", checkSum, err)
	}
	return
}

func (d *dao) SetDocumentRds(ctx context.Context, checkSum int64, data json.RawMessage) (err error) {
	if _, err = d.redis.Do(ctx, "SET", rdsDocKey(checkSum), []byte(data), "EX", d.rdsExpire); err != nil {
		log.Error("d.SetDocumentRds(checksum:%d) error(%v)", checkSum, err)
	}
	return
}

// ExpireUserDoc .
func (d *dao) ExpireUserDoc(c context.Context, mid int64) (ok bool, err error) {
	key := keyUserDoc(mid)
	ok, err = redis.Bool(d.redis.Do(c, "EXPIRE", key, d.rdsExpire))
	return
}

// HMsetUserDoc .
func (d *dao) HMsetUserDoc(c context.Context, mid int64, m map[string]string) (err error) {
	key := keyUserDoc(mid)
	args := redis.Args{}.Add(key)
	for k, v := range m {
		args = args.Add(k).Add(v)
	}
	p := d.redis.Pipeline()
	p.Send("HMSET", args...)
	p.Send("EXPIRE", key, d.rdsExpire)
	replies, nerr := p.Exec(c)
	if nerr != nil {
		err = nerr
		return
	}
	for replies.Next() {
		if _, nerr := replies.Scan(); nerr != nil {
			err = nerr
			return
		}
	}
	return
}

// HgetAllUserDoc .
func (d *dao) HgetAllUserDoc(c context.Context, mid int64) (res map[string]string, err error) {
	key := keyUserDoc(mid)
	if res, err = redis.StringMap(d.redis.Do(c, "HGETALL", key)); err != nil {
		return
	}
	return
}
