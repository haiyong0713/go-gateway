package article

//基于taishan的封装,后续统一在这里加泰山的限流

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

//err不为nil说明Get失败,err为空时，record为nil说明没有对应kv
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

func (d *Dao) PutTaishan(ctx context.Context, key string, value []byte, tableConfig *common.TaishanTableConfig) (err error) {
	if len(key) == 0 || len(value) == 0 || tableConfig == nil {
		log.Errorc(ctx, "PutTaishan req invalid record %v ,table %v", key, tableConfig.Name)
		return ecode.TaishanOperationReqInvalid
	}

	req := &taishan.PutReq{
		Table: tableConfig.Name,
		Record: &taishan.Record{
			Key: []byte(key),
		},
		Auth: &taishan.Auth{
			Token: tableConfig.Auth.Token,
		},
	}
	req.Record.Columns = append(req.Record.Columns, &taishan.Column{
		Value: value,
	})
	rspBind, err := d.TaishanCli.Put(ctx, req)
	if err != nil {
		log.Errorc(ctx, "PutTaishan Failed to put{key:%s, value:%s} to taishan for error:%+v", key, string(value), err)
		return ecode.TaishanOperationFail
	}
	if rspBind.Status.ErrNo != 0 {
		log.Errorc(ctx, "PutTaishan Failed to put{key:%s, value:%s} to taishan for error:%d", key, string(value), rspBind.Status.ErrNo)
		return ecode.TaishanOperationFail
	}
	return nil
}

func (d *Dao) DelTaishan(ctx context.Context, key string, tableConfig *common.TaishanTableConfig) (err error) {
	if len(key) == 0 || tableConfig == nil {
		log.Errorc(ctx, "DelTaishan req invalid record %v ,table %v", key, tableConfig.Name)
		return ecode.TaishanOperationReqInvalid
	}
	req := &taishan.DelReq{
		Table: tableConfig.Name,
		Record: &taishan.Record{
			Key: []byte(key),
		},
		Auth: &taishan.Auth{
			Token: tableConfig.Auth.Token,
		},
	}
	rspBind, err := d.TaishanCli.Del(ctx, req)
	if err != nil {
		log.Errorc(ctx, "DelTaishan Failed err %v key %v  table %v", err, key, tableConfig.Name)
		return ecode.TaishanOperationFail
	}
	if rspBind.Status.ErrNo != 0 {
		log.Errorc(ctx, "DelTaishan Failed err %v key %v  table %v", err, key, tableConfig.Name)
		return ecode.TaishanOperationFail
	}
	return nil
}
