package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"

	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/pkg/errors"
)

const _popularSeriesURL = "/x/admin/feed/popular/selected/series_in_use"

func (d *dao) PopularSeries(ctx context.Context) ([]*model.MgrSeriesData, error) {
	var res struct {
		Code int                    `json:"code"`
		Data []*model.MgrSeriesData `json:"data"`
	}
	if err := d.httpR.Get(ctx, d.popularSeriesURL, "", nil, &res); err != nil {
		return nil, errors.Wrap(err, d.popularSeriesURL)
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.popularSeriesURL)
	}
	return res.Data, nil
}

func (d *dao) AddCacheSeries(ctx context.Context, typ string, data []*model.MgrSeriesConfig) error {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "%+v", data)
	}
	key := fmt.Sprintf("popular_series_%s", typ)
	if _, err := conn.Do("SET", key, bs); err != nil {
		return errors.Wrapf(err, "%v,%s", key, bs)
	}
	return nil
}

func (d *dao) AddCacheSeriesDetail(ctx context.Context, data map[int64][]*model.MgrSeriesList) error {
	if len(data) == 0 {
		return nil
	}
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	var args redis.Args
	for id, val := range data {
		bs, err := json.Marshal(val)
		if err != nil {
			return errors.Wrapf(err, "%+v", val)
		}
		key := fmt.Sprintf("popular_series_detail_%d", id)
		args = args.Add(key).Add(bs)
	}
	if _, err := conn.Do("MSET", args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

const (
	_addPopularRankSQL = "INSERT INTO popular_rank (`mid`,`rank_from`) VALUES (?,?)"
	_rankFrom          = "job"
)

func (d *dao) AddPopularRank(ctx context.Context, mid int64) (int64, error) {
	res, err := d.showDB.Exec(ctx, _addPopularRankSQL, mid, _rankFrom)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

const (
	_addPopularWatchTimeSQL = "INSERT INTO %s (`mid`,`wtime`) VALUES (?,?)"
	_popularWatch           = "popular_watch_%d"
	_popularWatch0          = "popular_watch_0_%d"
)

func watchTable(mid, stage int64) string {
	if stage > 0 {
		return fmt.Sprintf(_popularWatch, stage)
	}
	return fmt.Sprintf(_popularWatch0, mid%10)
}

func (d *dao) AddPopularWatchTime(ctx context.Context, mid int64, stage int8, t time.Time) (int64, error) {
	res, err := d.showDB.Exec(ctx, fmt.Sprintf(_addPopularWatchTimeSQL, watchTable(mid, int64(stage))), mid, t)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func popularWatchKey(mid, stage int64) string {
	return fmt.Sprintf("precious:watch:%d:%d", mid, stage)
}

func (d *dao) DelCachePopularWatchTime(ctx context.Context, mid int64, stage int64) error {
	cacheKey := popularWatchKey(mid, stage)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		return err
	}
	return nil
}

func popularRankKey(mid int64) string {
	return fmt.Sprintf("precious:rank:%d", mid)
}

func (d *dao) DelCachePopularRank(ctx context.Context, mid int64) error {
	cacheKey := popularRankKey(mid)
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		return err
	}
	return nil
}
