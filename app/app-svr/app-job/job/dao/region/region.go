package region

import (
	"context"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"

	v1 "go-gateway/app/app-svr/app-job/job/api"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/model/region"

	"github.com/pkg/errors"
)

const (
	_regionDefiniteSQL    = "SELECT id,definite_state,definite_time FROM region_copy WHERE definite_state<>0"
	_updateRegionStateSQL = "UPDATE region_copy SET state=?,definite_state=0 WHERE id=?"
	// region
	_allSQL    = "SELECT r.rid,r.reid,r.name,r.logo,r.rank,r.goto,r.param,r.plat,r.area,r.build,r.conditions,r.uri,r.is_logo,r.type,r.is_rank,l.name FROM region AS r, language AS l WHERE r.state=1 AND l.id=r.lang_id ORDER BY r.rank DESC"
	_allSQL2   = "SELECT r.id,r.rid,r.reid,r.name,r.logo,r.rank,r.goto,r.param,r.plat,r.area,r.uri,r.is_logo,r.type,l.name FROM region_copy AS r, language AS l WHERE r.state=1 AND l.id=r.lang_id ORDER BY r.rank DESC"
	_limitSQL  = "SELECT l.id,l.rid,l.build,l.conditions FROM region_limit AS l,region_copy AS r WHERE l.rid=r.id"
	_configSQL = "SELECT c.id,c.rid,c.is_rank FROM region_rank_config AS c,region_copy AS r WHERE c.rid=r.id"
	// region android
	_regionPlatSQL = "SELECT r.rid,r.reid,r.name,r.logo,r.rank,r.goto,r.param,r.plat,r.area,l.name FROM region_copy AS r, language AS l WHERE r.plat=0 AND r.state=1 AND l.id=r.lang_id ORDER BY r.rank DESC"
	// region redis key
	_regionRedisKey     = "region"
	_loadRegionKey      = "loadRegion"
	_loadRegionListKey  = "loadRegionlist"
	_loadRegionCacheKey = "loadRegionListCache"
	_splitToken         = ":"
	_regionExpire       = 604800
)

// Dao is region dao.
type Dao struct {
	c          *conf.Config
	db         *sql.DB
	get        *sql.Stmt
	list       *sql.Stmt
	limit      *sql.Stmt
	config     *sql.Stmt
	regionPlat *sql.Stmt
	redis      *redis.Pool
}

// New new a region dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:     c,
		db:    sql.NewMySQL(c.MySQL.Show),
		redis: redis.NewPool(c.Redis.Entrance.Config),
	}
	// prepare
	d.get = d.db.Prepared(_allSQL)
	d.list = d.db.Prepared(_allSQL2)
	d.limit = d.db.Prepared(_limitSQL)
	d.config = d.db.Prepared(_configSQL)
	d.regionPlat = d.db.Prepared(_regionPlatSQL)
	return
}

func (d *Dao) RegionDefinite(ctx context.Context) (res []*region.Region, err error) {
	rows, err := d.db.Query(ctx, _regionDefiniteSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &region.Region{}
		if err = rows.Scan(&r.ID, &r.DefiniteState, &r.DefiniteTime); err != nil {
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// BeginTran begin a transaction
func (d *Dao) BeginTran(ctx context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(ctx)
}

func (d *Dao) UpdateRegionStateSQL(tx *sql.Tx, state int, id int64) (err error) {
	_, err = tx.Exec(_updateRegionStateSQL, state, id)
	return
}

// GetAll get all region.
func (d *Dao) All(ctx context.Context) (*v1.RegionReply, error) {
	rows, err := d.get.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reply := &v1.RegionReply{}
	for rows.Next() {
		a := &v1.Region{}
		if err = rows.Scan(&a.Rid, &a.Reid, &a.Name, &a.Logo, &a.Rank, &a.Goto, &a.Param, &a.Plat, &a.Area, &a.Build, &a.Condition, &a.Uri, &a.IsLogo, &a.Rtype, &a.Entrance, &a.Language); err != nil {
			return nil, err
		}
		reply.Regions = append(reply.Regions, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reply, err
}

func (d *Dao) AddCacheRegion(ctx context.Context, regions *v1.RegionReply) error {
	if regions.Size() <= 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := regions.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := regionActionKey(_loadRegionKey, "RegionReply")
	if _, err := conn.Do("SETEX", key, _regionExpire, bs); err != nil {
		return err
	}
	return nil
}

// RegionPlat get android
func (d *Dao) RegionPlat(ctx context.Context) (*v1.RegionReply, error) {
	rows, err := d.regionPlat.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reply := &v1.RegionReply{}
	for rows.Next() {
		a := &v1.Region{}
		if err = rows.Scan(&a.Rid, &a.Reid, &a.Name, &a.Logo, &a.Rank, &a.Goto, &a.Param, &a.Plat, &a.Area, &a.Language); err != nil {
			return nil, err
		}
		reply.Regions = append(reply.Regions, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reply, err
}

func (d *Dao) AddCacheRegionList(ctx context.Context, regions *v1.RegionReply) error {
	if regions.Size() <= 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := regions.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := regionActionKey(_loadRegionCacheKey, "RegionReply")
	if _, err := conn.Do("SETEX", key, _regionExpire, bs); err != nil {
		return err
	}
	return nil
}

// AllList get all region.
func (d *Dao) AllList(ctx context.Context) (*v1.RegionReply, error) {
	rows, err := d.list.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	apps := &v1.RegionReply{}
	for rows.Next() {
		a := &v1.Region{}
		if err = rows.Scan(&a.Id, &a.Rid, &a.Reid, &a.Name, &a.Logo, &a.Rank, &a.Goto, &a.Param, &a.Plat, &a.Area, &a.Uri, &a.IsLogo, &a.Rtype, &a.Language); err != nil {
			return nil, err
		}
		apps.Regions = append(apps.Regions, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return apps, err
}

// Limit region limits
func (d *Dao) Limit(ctx context.Context) (*v1.RegionLtmReply, error) {
	rows, err := d.limit.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	limits := map[int64][]*v1.Limit{}
	for rows.Next() {
		a := &v1.Limit{}
		if err = rows.Scan(&a.Id, &a.Rid, &a.Build, &a.Condition); err != nil {
			return nil, err
		}
		limits[a.Rid] = append(limits[a.Rid], a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.RegionLtmReply{}
	for k, v := range limits {
		ltm := &v1.RegionLtMap{}
		ltm.Key = k
		ltm.Limits = v
		res.Ltm = append(res.Ltm, ltm)
	}
	return res, err
}

// Config region configs
func (d *Dao) Config(ctx context.Context) (*v1.RegionCfgmReply, error) {
	rows, err := d.config.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	configs := map[int64][]*v1.Config{}
	for rows.Next() {
		a := &v1.Config{}
		if err = rows.Scan(&a.Id, &a.Rid, &a.ScenesId); err != nil {
			return nil, err
		}
		configs[a.Rid] = append(configs[a.Rid], a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.RegionCfgmReply{}
	for k, v := range configs {
		config := &v1.RegionCfgMap{}
		config.Key = k
		config.Configs = v
		res.Cfgm = append(res.Cfgm, config)
	}
	return res, err
}

func (d *Dao) AddRegionList(ctx context.Context, res *v1.RegionReply, limit *v1.RegionLtmReply, config *v1.RegionCfgmReply) error {
	if res.Size() <= 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string

	bs, err := res.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := regionActionKey(_loadRegionListKey, "RegionReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = limit.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key = regionActionKey(_loadRegionListKey, "RegionLtmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = config.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key = regionActionKey(_loadRegionListKey, "RegionCfgmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)
	if err = conn.Send("MSET", argsMDs...); err != nil {
		return err
	}
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, _regionExpire); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
}

func regionActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_regionRedisKey)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}
