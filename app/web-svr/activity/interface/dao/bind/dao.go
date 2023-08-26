package bind

import (
	"context"
	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

// Dao bnj dao.
type Dao struct {
	c  *conf.Config
	db *sql.DB
	mc *memcache.Memcache
}

var localD *Dao

// New init bnj dao.
func New(c *conf.Config) (d *Dao) {
	if localD != nil {
		return localD
	}
	d = &Dao{
		c:  c,
		db: component.GlobalDB,
		mc: memcache.New(c.Memcache.Like),
	}
	localD = d
	return d
}

const (
	_mcBindConfigsKeyTtl = 86400

	_getAllBindConfigsSql = "select id, bind_phone, bind_account, bind_type, game_type, act_id, bind_external, status from act_account_bind_config where is_deleted = 0"
	_insertBindConfig     = "insert into act_account_bind_config (bind_phone, bind_account, bind_type, game_type, act_id, bind_external, status) values (?, ?, ?, ?, ?, ?, ?)"
	_updateBindConfig     = "update act_account_bind_config set bind_phone = ?, bind_account = ?, bind_type = ?, game_type = ?, act_id = ?, bind_external = ?, status = ? where id = ?"
	_countBindConfig      = "select count(id) as num from act_account_bind_config where 1 = 1"
	_getBindConfigByPage  = "select id, bind_phone, bind_account, bind_type, game_type, act_id, bind_external, status from act_account_bind_config where is_deleted = 0 order by id desc limit ? offset ?"
)

func (d *Dao) GetAllBindConfigs(ctx context.Context) (configs []*v1.BindConfigInfo, err error) {
	configs = make([]*v1.BindConfigInfo, 0)
	rows, err := d.db.Query(ctx, _getAllBindConfigsSql)
	if err != nil {
		log.Errorc(ctx, "[GetAllBindConfigs][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		record := new(v1.BindConfigInfo)
		if err = rows.Scan(&record.ID, &record.BindPhone, &record.BindAccount, &record.BindType, &record.GameType, &record.ActId, &record.BindExternal, &record.Status); err != nil {
			log.Errorc(ctx, "[GetAllBindConfigs][Query][Error], err:%+v", err)
			return
		}
		configs = append(configs, record)
	}
	return
}

func formatMcBindConfigsKey() string {
	return "activity-interface:bind:configList"
}

func (d *Dao) StoreBindConfigsCache(ctx context.Context, configs []*v1.BindConfigInfo) (err error) {
	configMapping := make(map[int64]*v1.BindConfigInfo)
	for _, v := range configs {
		configMapping[v.ID] = v
	}
	item := &memcache.Item{
		Key:        formatMcBindConfigsKey(),
		Object:     configMapping,
		Expiration: _mcBindConfigsKeyTtl,
		Flags:      memcache.FlagJSON,
	}
	if err = d.mc.Set(ctx, item); err != nil {
		log.Errorc(ctx, "[StoreBindConfigs][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetBindConfigsCache(ctx context.Context) (configMapping map[int64]*v1.BindConfigInfo, err error) {
	configMapping = make(map[int64]*v1.BindConfigInfo)
	if err = d.mc.Get(ctx, formatMcBindConfigsKey()).Scan(&configMapping); err != nil {
		log.Errorc(ctx, "[GetBindConfigs][Get][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) DeleteBindConfigsCache(ctx context.Context) (err error) {
	if err = d.mc.Delete(ctx, formatMcBindConfigsKey()); err != nil {
		log.Errorc(ctx, "[GetBindConfigs][Delete][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) InsertBindConfig(ctx context.Context, bindInfo *v1.BindConfigInfo) (err error) {
	if _, err = d.db.Exec(ctx, _insertBindConfig, bindInfo.BindPhone, bindInfo.BindAccount, bindInfo.BindType, bindInfo.GameType, bindInfo.ActId, bindInfo.BindExternal, bindInfo.Status); err != nil {
		log.Errorc(ctx, "[insertBindConfig][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) UpdateBindConfig(ctx context.Context, bindInfo *v1.BindConfigInfo) (err error) {
	if _, err = d.db.Exec(ctx, _updateBindConfig, bindInfo.BindPhone, bindInfo.BindAccount, bindInfo.BindType, bindInfo.GameType, bindInfo.ActId, bindInfo.BindExternal, bindInfo.Status, bindInfo.ID); err != nil {
		log.Errorc(ctx, "[updateBindConfig][Error], err:%+v", err)
		return
	}
	return
}

func (d *Dao) GetBindConfigByOffset(ctx context.Context, page int, pageSize int) (list []*v1.BindConfigInfo, total int64, err error) {
	type countStruct struct {
		Num int64
	}
	list = make([]*v1.BindConfigInfo, 0)
	count := new(countStruct)
	row := d.db.QueryRow(ctx, _countBindConfig)
	if err = row.Scan(&count.Num); err != nil {
		log.Errorc(ctx, "[GetBinfConfigByOffset][Scan][Error], err:%+v", err)
		return
	}
	total = count.Num
	offset := (page - 1) * pageSize
	rows, err := d.db.Query(ctx, _getBindConfigByPage, pageSize, offset)
	if err != nil {
		log.Errorc(ctx, "[GetBinfConfigByOffset][Query][Error], err:%+v", err)
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		record := new(v1.BindConfigInfo)
		if err = rows.Scan(&record.ID, &record.BindPhone, &record.BindAccount, &record.BindType, &record.GameType, &record.ActId, &record.BindExternal, &record.Status); err != nil {
			log.Errorc(ctx, "[GetAllBindConfigs][Query][Error], err:%+v", err)
			return
		}
		list = append(list, record)
	}
	return
}
