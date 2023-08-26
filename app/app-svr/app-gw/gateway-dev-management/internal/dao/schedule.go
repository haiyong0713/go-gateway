package dao

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

const (
	_insertConfig     = "INSERT INTO gwmng_config (config_key, config_value) VALUES (?,?)"
	_selectConfigs    = "SELECT * FROM gwmng_config"
	_selectValueByKey = "SELECT config_value FROM gwmng_config WHERE config_key = ?"
	_updateValueByKey = "UPDATE gwmng_config SET config_value = ? WHERE config_key = ?"
)

func (d *dao) InsertConfig(ctx context.Context, gs *model.GatewaySchedule) error {
	if _, err := d.db.Exec(ctx, _insertConfig, gs.Key, gs.Value); err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return err
	}
	return nil
}

func (d *dao) SelectConfigs(ctx context.Context) ([]*model.GatewaySchedule, error) {
	rows, err := d.db.Query(ctx, _selectConfigs)
	if err != nil {
		return nil, err
	}
	var res []*model.GatewaySchedule
	defer rows.Close()
	for rows.Next() {
		a := &model.GatewaySchedule{}
		if err = rows.Scan(&a.Id, &a.Key, &a.Value, &a.Ctime, &a.Mtime); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) SelectValueByKey(ctx context.Context, key string) (string, error) {
	var value string
	row := d.db.QueryRow(ctx, _selectValueByKey, key)
	err := row.Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *dao) UpdateValueByKey(ctx context.Context, key string, value string) error {
	_, err := d.db.Exec(ctx, _updateValueByKey, value, key)
	if err != nil {
		return err
	}
	return nil
}
