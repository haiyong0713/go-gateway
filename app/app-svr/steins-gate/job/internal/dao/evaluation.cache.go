package dao

import (
	"context"
	"strconv"

	"go-common/library/log"
)

const (
	_postfixEvaluation = "_ev"
)

func evaluationKey(aid int64) string {
	return strconv.FormatInt(aid, 10) + _postfixEvaluation
}

// addEvalCache .
func (d *Dao) addEvalCache(c context.Context, aid int64, val int64) (err error) {
	key := evaluationKey(aid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, val); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, val, err)
		return
	}
	return

}
