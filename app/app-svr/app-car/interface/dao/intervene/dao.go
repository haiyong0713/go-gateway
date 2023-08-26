package intervene

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/intervene"
)

const (
	_loadInterveneDataSql = "SELECT `key_type`, `keyword`, `card_type`,`aid`,`rank` FROM `tv_car_xiaopeng_intervene` WHERE `is_deleted` = 0 ORDER BY key_type,rank ASC"
)

type Dao struct {
	db *sql.DB
}

// New init 小鹏插卡干预数据
func New(c *conf.Config, db *sql.DB) (d *Dao) {
	d = &Dao{
		db: db,
	}
	return d
}

// mysql获取小鹏查看干预的数据集合
func (d *Dao) LoadInterveneData(ctx context.Context) (items []*model.TvXiaoPengInterveneModel, err error) {
	rows, err := d.db.Query(ctx, _loadInterveneDataSql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item model.TvXiaoPengInterveneModel
		if err = rows.Scan(&item.Type, &item.KeyWord, &item.CardType, &item.Aid, &item.Rank); err != nil {
			return
		}
		items = append(items, &item)
	}
	if err = rows.Err(); err != nil {
		return items, err
	}
	return items, err
}
