package bwsonline

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

const (
	_userBindTicketCachePre = "bwpark:ticket:bind:hash:pre"
	_hashSharding           = 8
)

type bindStruct struct {
	Mid int64
	Id  int64
}

func getHaskKey(mid int64) string {
	return buildKey(_userBindTicketCachePre, mid%_hashSharding)
}
func (d *Dao) BatchCacheBindRecords(ctx context.Context, records []*bwsonline.TicketBindRecord) (err error) {
	if len(records) == 0 {
		return
	}
	var (
		hashMap = make(map[string][]*bindStruct)
	)
	for _, reco := range records {
		if reco.Id <= 0 {
			continue
		}
		tmpList := hashMap[getHaskKey(reco.Mid)]
		hashMap[getHaskKey(reco.Mid)] = append(tmpList, &bindStruct{Mid: reco.Mid, Id: reco.Id})
	}
	for hkey, hashList := range hashMap {
		if err = d.HmsetBindRecords(ctx, hkey, hashList); err != nil {
			log.Infoc(ctx, "BatchCacheBindRecords HmsetBindRecords hkey:%s , hashList:%v", hkey, len(hashList))
			return
		}
	}
	return
}

func (d *Dao) HmsetBindRecords(ctx context.Context, hkey string, hashList []*bindStruct) (err error) {
	var args = redis.Args{}.Add(hkey)
	for _, item := range hashList {
		args = args.Add(item.Mid).Add(item.Id)
	}
	if _, err = d.redis.Do(ctx, "HMSET", args...); err != nil {
		log.Errorc(ctx, "HmsetBindRecords conn.Send(HMSET) error(%v)", err)
		return
	}
	return
}

func (d *Dao) CheckBindRecord(ctx context.Context, mid int64) (id int64, err error) {
	if id, err = redis.Int64(d.redis.Do(ctx, "HGET", getHaskKey(mid), mid)); err != nil {
		log.Errorc(ctx, "CheckBindRecord mid:%v , err:%+v", mid, err)
		if err == redis.ErrNil {
			err = nil
		}
	}
	return
}
