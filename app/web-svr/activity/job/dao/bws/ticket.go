package bws

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

func bwsTicketBindKey(bid int32) string {
	return fmt.Sprintf("bws_ticket_bind_cache:%d", bid)
}

func (d *Dao) SetMaxSyncBindRecordId(ctx context.Context, recordId int64, bid int32) (err error) {
	var (
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	key := bwsTicketBindKey(bid)
	if _, err = conn.Do("SET", key, recordId); err != nil {
		log.Errorc(ctx, "SetMaxSyncBindRecordId conn.Send(SET, %s, %s) error(%v)", key, recordId, err)
	}
	return
}

func (d *Dao) GetMaxSyncBindRecordId(ctx context.Context, bid int32) (recordId int64, err error) {
	var (
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	key := bwsTicketBindKey(bid)
	if recordId, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		log.Errorc(ctx, "GetMaxSyncBindRecordId conn.Send(GET, %s, %s) error(%v)", key, recordId, err)
	}
	return
}
