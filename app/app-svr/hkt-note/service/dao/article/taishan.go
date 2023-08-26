package article

import (
	"context"
	"go-common/library/database/taishan"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/common"
	"go-gateway/app/app-svr/hkt-note/ecode"
)

const (
	TAISHAN_RECORD_NOT_EXISTED = 404
)

// err不为nil说明Get失败,err为空时，record为nil说明没有对应kv
func (d *Dao) GetTaishan(ctx context.Context, key string, tableConfig *common.TaishanTableConfig) (record *taishan.Record, err error) {
	if len(key) == 0 || tableConfig == nil {
		log.Errorc(ctx, "GetTaishan req invalid record %v ,table %v", key, tableConfig.Name)
		return nil, ecode.TaishanOperationReqInvalid
	}

	//构造请求
	req := &taishan.GetReq{
		Table: tableConfig.Name,
		Auth:  &taishan.Auth{Token: tableConfig.Auth.Token},
		//当前统一读主
		AccessPolicy: &taishan.AccessPolicy{
			Policy: taishan.AccessPolicy_PRIMARY_ONLY,
		},
		Record: &taishan.Record{
			Key: []byte(key),
		},
	}
	rsp, err := d.TaishanCli.Get(ctx, req)
	if err != nil || rsp == nil || rsp.Record == nil || rsp.Record.Status == nil {
		log.Errorc(ctx, "GetTaishan Failed err %v key %v  table %v rsp %v", err, key, tableConfig.Name, rsp)
		return nil, ecode.TaishanOperationFail
	}
	if rsp.Record.Status.ErrNo != 0 && rsp.Record.Status.ErrNo != TAISHAN_RECORD_NOT_EXISTED {
		log.Errorc(ctx, "GetTaishan Failed status err  key %v table %v rsp %v", key, tableConfig.Name, rsp)
		return nil, ecode.TaishanOperationFail
	}
	if rsp.Record.Status.ErrNo == TAISHAN_RECORD_NOT_EXISTED || len(rsp.Record.Columns) <= 0 || rsp.Record.Columns[0] == nil {
		return nil, nil
	}
	return rsp.Record, nil
}
