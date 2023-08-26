package black

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
)

const (
	_prefixBlack = "b_"
)

func keyBlack(mid int64) string {
	return _prefixBlack + strconv.FormatInt(mid, 10)
}

func (d *Dao) addBlackCache(c context.Context, mid int64, aids ...int64) (err error) {
	if len(aids) == 0 {
		return
	}
	key := keyBlack(mid)
	conn := d.redis.Conn(c)
	defer conn.Close()
	for _, aid := range aids {
		if err = conn.Send("ZADD", key, aid, aid); err != nil {
			err = errors.Wrapf(err, "conn.Send(ZADD,%s,%d,%d)", key, aid, aid)
			return
		}
	}
	if err = conn.Send("EXPIRE", key, d.expireRds); err != nil {
		err = errors.Wrapf(err, "conn.Send(EXPIRE,%s,%d)", key, d.expireRds)
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for i := 0; i < len(aids)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			return
		}
	}
	return
}

func (d *Dao) delBlackCache(c context.Context, mid, aid int64) (err error) {
	key := keyBlack(mid)
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("ZREM", key, aid); err != nil {
		err = errors.Wrapf(err, "conn.Do(ZREM,%s,%d)", key, aid)
	}
	return
}
