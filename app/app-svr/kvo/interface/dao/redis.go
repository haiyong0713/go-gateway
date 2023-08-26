package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-gateway/app/app-svr/kvo/interface/model"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_rdsuserConfKeyPrefix = "rd_u"
	_rdsdocKeyPrefix      = "rd_d_v1"
	_rdsUcKeyPrefix       = "rd_uc_"
)

func rdsUserConfKey(mid int64, moduleKey int) string {
	return fmt.Sprintf("%v_%v_%v", _rdsuserConfKeyPrefix, mid, moduleKey)
}

func rdsDocKey(checkSum int64) string {
	return fmt.Sprintf("%v_%v", _rdsdocKeyPrefix, checkSum)
}

func rdsUcKey(mid int64, buvid string, moduleKey int) string {
	if mid > 0 {
		buvid = "$"
	}
	return fmt.Sprintf("%s_%d_%s_%d", _rdsUcKeyPrefix, mid, buvid, moduleKey)
}

// UserConfCache user config cache
func (d *Dao) UserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int) (bs []byte, err error) {
	bs, err = redis.Bytes(d.rds.Do(ctx, "GET", rdsUcKey(mid, buvid, moduleKey)))
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
func (d *Dao) SetUserDocRds(ctx context.Context, mid int64, buvid string, moduleKey int, bs []byte) (err error) {
	if _, err = d.rds.Do(ctx, "SET", rdsUcKey(mid, buvid, moduleKey), bs, "EX", d.rdsUcDocExpire); err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,buvid:%s) error(%v)", mid, moduleKey, buvid, err)
	}
	return
}

// UserConfCache user config cache
func (d *Dao) UserConfRds(ctx context.Context, mid int64, moduleKey int) (uc *model.UserConf, err error) {
	var bs []byte
	bs, err = redis.Bytes(d.rds.Do(ctx, "GET", rdsUserConfKey(mid, moduleKey)))
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
func (d *Dao) SetUserConfRds(ctx context.Context, uc *model.UserConf) (err error) {
	bs, err := json.Marshal(uc)
	if err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,uc:%+v) error(%v)", uc.Mid, uc.ModuleKey, uc, err)
		return
	}
	if _, err = d.rds.Do(ctx, "SET", rdsUserConfKey(uc.Mid, uc.ModuleKey), bs, "EX", d.rdsExpire); err != nil {
		log.Error("d.SetUserConfRds(mid:%d,modulekey:%d,uc:%+v) error(%v)", uc.Mid, uc.ModuleKey, uc, err)
	}
	return

}

// DelUserConfCache del user config cache
func (d *Dao) DelUserConfRds(ctx context.Context, mid int64, moduleKey int) (err error) {
	var (
		key = rdsUserConfKey(mid, moduleKey)
	)
	if _, err = d.rds.Do(ctx, "DEL", key); err != nil {
		log.Error("d.DelUserConfRds(mid:%d,modulekey:%d) err(%v)", mid, moduleKey, err)
	}
	return
}

func (d *Dao) DocumentRds(ctx context.Context, checkSum int64) (data json.RawMessage, err error) {
	var bs []byte
	bs, err = redis.Bytes(d.rds.Do(ctx, "GET", rdsDocKey(checkSum)))
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

func (d *Dao) DelDocumentRds(ctx context.Context, checkSum int64) (err error) {
	var (
		key = rdsDocKey(checkSum)
	)
	if _, err = d.rds.Do(ctx, "DEL", key); err != nil {
		log.Error("d.DelDocumentRds(checksum:%d) err(%v)", checkSum, err)
	}
	return
}

func (d *Dao) SetDocumentRds(ctx context.Context, checkSum int64, data json.RawMessage) (err error) {
	if _, err = d.rds.Do(ctx, "SET", rdsDocKey(checkSum), []byte(data), "EX", d.rdsExpire); err != nil {
		log.Error("d.SetDocumentRds(checksum:%d) error(%v)", checkSum, err)
	}
	return
}

func (d *Dao) pingRedis(c context.Context) (err error) {
	if _, err = d.rds.Do(c, "SET", "ping", "pong"); err != nil {
		log.Error("d.PingRedis error(%v)", err)
	}
	return
}
