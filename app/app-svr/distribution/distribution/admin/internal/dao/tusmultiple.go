package dao

import (
	"context"

	tmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"

	"github.com/pkg/errors"
)

func (d *dao) SaveMultipleTusConfigs(ctx context.Context, details []*tmm.Detail, fieldName, configVersion string) error {
	taishanBatchPutKey := make(map[string][]byte, len(details))
	for _, v := range details {
		key := tmm.KeyFormat(fieldName, configVersion, v.TusValue)
		taishanBatchPutKey[key] = v.Config
	}
	req := d.kv.NewBatchPutReq(ctx, taishanBatchPutKey)
	resp, err := d.kv.BatchPut(ctx, req)
	if err != nil {
		return err
	}
	if !resp.AllSucceed {
		return errors.Errorf("Failed to put all config to taishan")
	}
	return nil
}
