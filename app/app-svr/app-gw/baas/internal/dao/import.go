package dao

import (
	"context"
	"fmt"
	"strings"

	"go-gateway/app/app-svr/app-gw/baas/api"
)

const (
	_importByExportIds = "SELECT id,baas_export_id,datasource_api,datasource_type FROM baas_import WHERE baas_export_id IN (%s)"
	_addBaasImport     = "INSERT INTO baas_import(baas_export_id,datasource_api,datasource_type) VALUES (?,?,?)"
	_updateBaasImport  = "UPDATE baas_import SET datasource_api=?,datasource_type=? WHERE id=?"
	_importAll         = "SELECT id,baas_export_id,datasource_api,datasource_type FROM baas_import"
)

func (d *dao) ImportByExportIds(ctx context.Context, ids []int64) (map[int64][]*api.BaasImport, error) {
	out := make(map[int64][]*api.BaasImport)
	if len(ids) == 0 {
		return out, nil
	}
	var (
		args []string
		sqls []interface{}
	)
	for _, sid := range ids {
		args = append(args, "?")
		sqls = append(sqls, sid)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_importByExportIds, strings.Join(args, ",")), sqls...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := &api.BaasImport{}
		if err = rows.Scan(&item.Id, &item.BaasExportId, &item.DatasourceApi, &item.DatasourceType); err != nil {
			return nil, err
		}
		out[item.BaasExportId] = append(out[item.BaasExportId], item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *dao) AddImport(ctx context.Context, param *api.AddImportRequest) error {
	_, err := d.db.Exec(ctx, _addBaasImport, param.BaasExportId, param.DatasourceApi, param.DatasourceType)
	return err
}

func (d *dao) UpdateImport(ctx context.Context, param *api.UpdateImportRequest) error {
	_, err := d.db.Exec(ctx, _updateBaasImport, param.DatasourceApi, param.DatasourceType, param.Id)
	return err
}

func (d *dao) ImportAll(ctx context.Context) (map[int64][]*api.ImportItem, error) {
	rows, err := d.db.Query(ctx, _importAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64][]*api.ImportItem)
	for rows.Next() {
		item := &api.BaasImport{}
		if err = rows.Scan(&item.Id, &item.BaasExportId, &item.DatasourceApi, &item.DatasourceType); err != nil {
			return nil, err
		}
		out[item.BaasExportId] = append(out[item.BaasExportId], api.ConstructImportItem(item))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
