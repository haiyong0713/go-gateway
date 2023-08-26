package dao

import (
	"context"

	"go-gateway/app/app-svr/app-gw/baas/api"
)

const (
	_exportAll        = "SELECT id,export_api,model_name,tree_id,ctime,state FROM baas_export ORDER BY id DESC"
	_addBaasExport    = "INSERT INTO baas_export(export_api,model_name,tree_id,state) VALUES (?,?,?,?)"
	_updateBaasExport = "UPDATE baas_export SET export_api=?,model_name=?,tree_id=?,state=? WHERE id=?"
)

func (d *dao) ExportList(ctx context.Context) ([]*api.BaasExport, error) {
	rows, err := d.db.Query(ctx, _exportAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*api.BaasExport
	for rows.Next() {
		item := &api.BaasExport{}
		if err := rows.Scan(&item.Id, &item.ExportApi, &item.ModelName,
			&item.TreeId, &item.Ctime, &item.State); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *dao) AddExport(ctx context.Context, param *api.AddExportRequest) error {
	_, err := d.db.Exec(ctx, _addBaasExport, param.ExportApi, param.ModelName, param.TreeId, param.State)
	return err
}

func (d *dao) UpdateExport(ctx context.Context, export *api.UpdateExportRequest) error {
	_, err := d.db.Exec(ctx, _updateBaasExport, export.ExportApi,
		export.ModelName, export.TreeId, export.State, export.Id)
	return err
}
