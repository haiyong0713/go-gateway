package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_prefixArchiveExtra = "aid_extra_%d"
)

func archiveExtraKey(aid int64) string {
	return fmt.Sprintf(_prefixArchiveExtra, aid)
}

func (d *Dao) ArchiveExtraCacheByAid(c context.Context, aid int64) (values map[string]string, err error) {
	if values, err = redis.StringMap(d.redis.Do(c, "HGETALL", archiveExtraKey(aid))); err != nil {
		log.Error("ArchiveExtraCache HGETALL aid(%d) error(%v)", aid, err)
		return values, err
	}
	return values, nil
}

func (d *Dao) ArchiveExtraCacheByAids(c context.Context, aids []int64) (map[int64]map[string]string, error) {
	valuesMap := make(map[int64]map[string]string)

	conn := d.redis.Get(c)
	defer conn.Close()

	for _, aid := range aids {
		if err := conn.Send("HGETALL", archiveExtraKey(aid)); err != nil {
			log.Error("ArchiveExtraCache HGETALL aid(%d) error(%v)", aid, err)
		}
	}
	if err := conn.Flush(); err != nil {
		log.Error("ArchiveExtraCache conn.Flush aids(%d) error(%v)", aids, err)
		return valuesMap, err
	}
	for _, aid := range aids {
		value, err := redis.StringMap(conn.Receive())
		if err != nil {
			log.Error("ArchiveExtraCache conn.Receive aid(%d) error(%v)", aid, err)
			continue
		}
		valuesMap[aid] = value
	}
	return valuesMap, nil
}
