package region

import (
	"context"
	"go-gateway/app/app-svr/app-show/interface/component"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	jobApi "go-gateway/app/app-svr/app-job/job/api"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/region"

	"github.com/pkg/errors"
)

const (
	_regionRedisKeyPrefix = "region"
	_loadRegionKey        = "loadRegion"
	_loadRegionListKey    = "loadRegionlist"
	_loadRegionCacheKey   = "loadRegionListCache"
	_splitToken           = ":"
)

type Dao struct {
	db    *sql.DB
	redis *redis.Pool
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:    component.GlobalShowDB,
		redis: redis.NewPool(c.Redis.Entrance),
	}
	return
}

// GetAll get all region.
func (d *Dao) All(ctx context.Context) ([]*region.Region, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", regionActionKey(_loadRegionKey, "RegionReply")))
	if err != nil {
		return nil, err
	}
	res := jobApi.RegionReply{}
	if err = res.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	var regions []*region.Region
	for _, reg := range res.Regions {
		r := region.Region{}
		r.FromJobPBRegion(reg)
		regions = append(regions, &r)
	}
	return regions, nil
}

// RegionPlat get android
func (d *Dao) RegionPlat(ctx context.Context) ([]*region.Region, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", regionActionKey(_loadRegionCacheKey, "RegionReply")))
	if err != nil {
		return nil, err
	}
	res := jobApi.RegionReply{}
	if err = res.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	var regions []*region.Region
	for _, reg := range res.Regions {
		r := region.Region{}
		r.FromJobPBRegion(reg)
		regions = append(regions, &r)
	}
	return regions, nil
}

// AllList get all region.
func (d *Dao) AllList(ctx context.Context) ([]*region.Region, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", regionActionKey(_loadRegionListKey, "RegionReply")))
	if err != nil {
		return nil, err
	}
	res := jobApi.RegionReply{}
	if err = res.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	var regions []*region.Region
	for _, reg := range res.Regions {
		r := region.Region{}
		r.FromJobPBRegion(reg)
		regions = append(regions, &r)
	}
	return regions, nil
}

// Limit region limits
func (d *Dao) Limit(ctx context.Context) (map[int64][]*region.Limit, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", regionActionKey(_loadRegionListKey, "RegionLtmReply")))
	if err != nil {
		return nil, err
	}
	raw := jobApi.RegionLtmReply{}
	if err = raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	limits := map[int64][]*region.Limit{}
	for _, v := range raw.Ltm {
		for _, lmt := range v.Limits {
			limit := region.Limit{}
			limit.FromJobPBLimit(lmt)
			limits[v.Key] = append(limits[v.Key], &limit)
		}
	}
	return limits, nil
}

// Config region configs
func (d *Dao) Config(ctx context.Context) (map[int64][]*region.Config, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", regionActionKey(_loadRegionListKey, "RegionCfgmReply")))
	if err != nil {
		return nil, err
	}
	raw := jobApi.RegionCfgmReply{}
	if err = raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	configs := map[int64][]*region.Config{}
	for _, v := range raw.Cfgm {
		for _, cfg := range v.Configs {
			config := region.Config{}
			config.FromJobPBConfig(cfg)
			configs[v.Key] = append(configs[v.Key], &config)
		}
	}
	return configs, nil
}

// Close close resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}

func regionActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_regionRedisKeyPrefix)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}
