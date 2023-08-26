package show

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/pkg/errors"
)

const (
	_selectedSerie = "%d_%s_serie"
)

func oneSerieKey(number int64, stype string) string {
	return fmt.Sprintf(_selectedSerie, number, stype)
}

func (d *Dao) PickSerieCache(c context.Context, sType string, number int64) (*show.SerieFull, error) {
	var (
		key  = oneSerieKey(number, sType)
		conn = d.selectedRedis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		log.Error("PickSerieCache redis.Get(%s) error(%v)", key, err)
		return nil, err
	}

	var serie *show.SerieFull
	err = json.Unmarshal(values, &serie)
	if err != nil {
		return nil, errors.Wrapf(err, "PickSerieCache json.Unmarshal error")
	}

	return serie, nil
}
