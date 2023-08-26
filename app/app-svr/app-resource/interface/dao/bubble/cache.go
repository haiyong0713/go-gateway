package bubble

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/app-resource/interface/model"

	"go-common/library/log"

	"go-common/library/cache/memcache"
)

const (
	_bubbleKey = "bub_%d_%d"
)

func BubbleKey(buid, mid int64) (key string) {
	return fmt.Sprintf(_bubbleKey, buid, mid)
}

func (d *Dao) SetBubbleConfig(c context.Context, buid, mid int64, state int, expire int32) (err error) {
	var (
		key  = BubbleKey(buid, mid)
		conn = d.bubbleMc.Conn(c)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: state, Flags: memcache.FlagJSON, Expiration: expire}
	if err = conn.Set(item); err != nil {
		log.Error("[SetAggregations] conn.Set() error()")
	}
	return
}

func (d *Dao) BubbleConfig(c context.Context, buid, mid int64) (state int, err error) {
	var (
		key  = BubbleKey(buid, mid)
		conn = d.bubbleMc.Conn(c)
		bc   *memcache.Item
	)
	defer conn.Close()
	if bc, err = conn.Get(key); err != nil {
		if err == memcache.ErrNotFound {
			state = model.BubbleNoExist
			err = nil
		} else {
			log.Error("memcache.Get(%s) error(%v)", key, err)
		}
		return
	}
	if err = conn.Scan(bc, &state); err != nil {
		log.Error("conn.Scan(%s) error(%v)", bc.Value, err)
	}
	return
}
