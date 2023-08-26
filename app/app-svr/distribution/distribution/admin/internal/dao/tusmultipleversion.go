package dao

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
	tmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"
	vcm "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"github.com/pkg/errors"
)

func (tmv *tusMultipleVersionDao) FetchVersionManager(ctx context.Context, fieldName string) (*vcm.ConfigVersionManager, error) {
	req := tmv.kv.NewGetReq([]byte(vcm.NewTaishanKey(fieldName)))
	record, err := tmv.kv.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	cvm := &vcm.ConfigVersionManager{}
	if err := json.Unmarshal(record.Columns[0].Value, cvm); err != nil {
		return nil, err
	}
	return cvm, nil
}

func (tmv *tusMultipleVersionDao) EditVersions(ctx context.Context, in *vcm.ConfigVersionManager) error {
	cvmbs, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req := tmv.kv.NewPutReq([]byte(vcm.NewTaishanKey(in.Field)), cvmbs)
	if err := tmv.kv.Put(ctx, req); err != nil {
		return err
	}
	return nil
}

func (tmv *tusMultipleVersionDao) BatchFetchVersionManager(ctx context.Context, fieldNames []string) ([]*vcm.ConfigVersionManager, error) {
	var keys []string
	for _, v := range fieldNames {
		keys = append(keys, vcm.NewTaishanKey(v))
	}
	req := tmv.kv.NewBatchGetReq(ctx, keys)
	resp, err := tmv.kv.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.AllSucceed {
		return nil, errors.Errorf("Failed to Fetch all configs")
	}
	var cvms []*vcm.ConfigVersionManager
	for _, v := range resp.Records {
		if len(v.Columns) == 0 || v.Columns[0].Value == nil {
			continue
		}
		cvm := &vcm.ConfigVersionManager{}
		if err := json.Unmarshal(v.Columns[0].Value, cvm); err != nil {
			return nil, err
		}
		cvms = append(cvms, cvm)
	}
	return cvms, nil
}

func (tmv *tusMultipleVersionDao) DeleteVersionConfig(ctx context.Context, fieldName string, versionInfo *vcm.VersionInfo) error {
	configKeys := []string{tmm.KeyFormat(fieldName, versionInfo.ConfigVersion, model.DefaultTusValue)}
	for _, v := range versionInfo.TusValues {
		configKeys = append(configKeys, tmm.KeyFormat(fieldName, versionInfo.ConfigVersion, v))
	}
	batchReply, err := tmv.kv.BatchDel(ctx, tmv.kv.NewBatchDelReq(ctx, configKeys))
	if err != nil {
		return err
	}
	if !batchReply.AllSucceed {
		return errors.New("Failed to delete all config keys")
	}
	return nil
}
