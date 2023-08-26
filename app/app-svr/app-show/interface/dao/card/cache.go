package card

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/model/selected"

	"github.com/pkg/errors"
)

const (
	_allSeries          = "%s_series"
	_selectedSerie      = "%d_%s_serie"
	_weekly             = "weekly_selected"
	_cardRedisKeyPrefix = "card"
	_splitToken         = ":"
)

func oneSerieKey(number int64, stype string) string { // for mc
	return fmt.Sprintf(_selectedSerie, number, stype)
}

func allSeriesKey(stype string) string { // for redis
	return fmt.Sprintf(_allSeries, stype)
}

// SetAllSeries adds all series data into Redis
func (d *Dao) SetAllSeries(c context.Context, sType string, list []*selected.SerieFilter) error {
	if len(list) == 0 {
		return nil
	}
	var (
		conn = d.redis.Get(c)
		key  = allSeriesKey(sType)
	)
	defer conn.Close()

	bs, err := json.Marshal(list)
	if err != nil {
		return errors.Wrapf(err, "SetAllSeries json.Marshal error")
	}

	if _, err := conn.Do("SET", key, bs); err != nil {
		log.Error("SetAllSeries redis error(%v), key(%v)", err, key)
		return err
	}

	return nil
}

// AllSeriesCache gets the cache of all series of one type from redis
func (d *Dao) AllSeriesCache(c context.Context, sType string) ([]*selected.SerieFilter, error) {
	var (
		conn = d.redis.Get(c)
		key  = allSeriesKey(sType)
	)
	defer conn.Close()

	values, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		log.Error("AllSeriesCache redis.Get(%s) error(%v)", key, err)
		return nil, err
	}

	var res []*selected.SerieFilter
	if err := json.Unmarshal(values, &res); err != nil {
		return nil, errors.Wrapf(err, "AllSeriesCache json.Unmarshal error")
	}
	if len(res) <= 0 {
		return nil, errors.Errorf("result is nil. key(%v)", key)
	}

	return res, nil
}

// AddSerieCache adds one serie's data into MC
func (d *Dao) AddSerieCache(c context.Context, serie *selected.SerieFull) error {
	if serie == nil {
		return errors.Errorf("Serie is Nil")
	}
	var (
		key  = oneSerieKey(serie.Config.Number, serie.Config.Type)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	bs, err := json.Marshal(serie)
	if err != nil {
		return errors.Wrapf(err, "AddSerieCache json.Unmarshal error")
	}

	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Error("AddSerieCache redis error(%v), key(%v)", err, key)
		return err
	}

	return nil
}

// PickSerieCache get serie cache
func (d *Dao) PickSerieCache(c context.Context, sType string, number int64) (*selected.SerieFull, error) {
	var (
		key  = oneSerieKey(number, sType)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	values, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		log.Error("PickSerieCache redis.Get(%s) error(%v)", key, err)
		return nil, err
	}

	var serie *selected.SerieFull
	err = json.Unmarshal(values, &serie)
	if err != nil {
		return nil, errors.Wrapf(err, "PickSerieCache json.Unmarshal error")
	}

	return serie, nil
}

func (d *Dao) BatchPickSerieCache(c context.Context, sType string, numbers []int64) (map[int64]*selected.SerieFull, error) {
	if len(numbers) == 0 && sType != _weekly {
		return nil, ecode.RequestErr
	}
	var (
		args  redis.Args
		conn  = d.redis.Get(c)
		items [][]byte
		err   error
		res   = make(map[int64]*selected.SerieFull, len(numbers))
	)
	defer conn.Close()
	for _, number := range numbers {
		args = args.Add(oneSerieKey(number, sType))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return nil, err
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		temp := new(selected.SerieFull)
		if err = json.Unmarshal(item, temp); err != nil {
			log.Error("selected full Unmarshal error(%v)", err)
			continue
		}
		if temp.Config != nil {
			res[temp.Config.SerieCore.Number] = temp
		}
	}
	return res, nil
}

func cardActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_cardRedisKeyPrefix)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}
