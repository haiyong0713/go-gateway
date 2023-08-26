package newyear2021

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	model "go-gateway/app/web-svr/activity/interface/model/newyear2021"
)

const (
	sql4GetLatestConf = `
SELECT id,config_content
FROM bnj2021_config where is_deleted = 0
ORDER BY id DESC
LIMIT 1`
	sql4AddConf = `
INSERT INTO bnj2021_config (config_content)
VALUES (?)`
	sql4DelConf = `
UPDATE bnj2021_config SET 
is_deleted = 1
WHERE
id = ?`
)

func (d *Dao) GetLatestConf(ctx context.Context) (version int64, res *model.Config, err error) {
	row := component.GlobalBnjDB.QueryRow(ctx, sql4GetLatestConf)
	configContent := ""
	if err = row.Scan(&version, &configContent); err != nil {
		log.Errorc(ctx, "d.GetLatestConf row.Scan error: %v", err)
		return
	}
	if err = json.Unmarshal([]byte(configContent), &res); err != nil {
		log.Errorc(ctx, "d.GetLatestConf json.Unmarshal error: %v", err)
		return
	}
	return
}

func (d *Dao) UpdateConf(ctx context.Context, config *model.Config) (err error) {
	bs, err := json.Marshal(config)
	if err != nil {
		log.Errorc(ctx, "d.UpdateConf json.Marshal error: %v", err)
		return err
	}
	_, err = component.GlobalBnjDB.Exec(ctx, sql4AddConf, string(bs))
	if err != nil {
		log.Errorc(ctx, "d.UpdateConf Exec error: %v", err)
	}
	return err
}

func (d *Dao) DeleteConf(ctx context.Context, version int64) (err error) {
	_, err = component.GlobalBnjDB.Exec(ctx, sql4DelConf, version)
	if err != nil {
		log.Errorc(ctx, "d.DeleteConf Exec error: %v", err)
	}
	return err
}
