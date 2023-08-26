package serial

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/serial/service"
	"github.com/pkg/errors"
)

const _businessSerialHis = "platform-car"

func (d *Dao) SerialHistory(ctx context.Context, mid, serialID, serialIDType int64, ps int32, buvid string) ([]*api.SerialHistory, error) {
	req := &api.HistoryReq{
		Mid:                mid,
		Ps:                 ps,
		SerialId:           serialID,
		Business:           _businessSerialHis,
		BusinessSerialType: serialIDType,
		Buvid:              buvid,
	}
	resp, err := d.serialCli.History(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "SerialHistory mid=%d req=%+v", mid, req)
	}
	return resp.GetList(), nil
}

func (d *Dao) AddHistory(ctx context.Context, mid, aid, viewAt, serialId, serialIDType int64, buvid string) error {
	req := &api.AddHistoryReq{
		Mid:                mid,
		SerialId:           serialId,
		Episode:            aid,
		EpisodeType:        api.EpisodeType_EpisodeTypeUGC,
		ViewAt:             viewAt,
		Business:           _businessSerialHis,
		BusinessSerialType: serialIDType,
		Buvid:              buvid,
	}
	_, err := d.serialCli.AddHistory(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "AddHistory mid=%d req=%+v", mid, req)
	}
	return nil
}

// SerialProgress 查询单个合集历史.
func (d *Dao) SerialProgress(ctx context.Context, mid, serialID, serialIDType int64, buvid string) (*api.SerialHistory, error) {
	req := &api.ProgressReq{
		Mid:                mid,
		SerialId:           serialID,
		Business:           _businessSerialHis,
		BusinessSerialType: serialIDType,
		Buvid:              buvid,
	}
	resp, err := d.serialCli.Progress(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "SerialProgress mid=%d req=%+v", mid, req)
	}
	return resp.GetInfo(), nil
}

// SerialBatchProgress 批量查询合集历史.
func (d *Dao) SerialBatchProgress(ctx context.Context, mid int64, buvid string, s []*api.BatchSerial) ([]*api.SerialHistory, error) {
	req := &api.BatchProgressReq{
		Mid:      mid,
		Business: _businessSerialHis,
		Serials:  s,
		Buvid:    buvid,
	}
	res, err := d.serialCli.BatchProgress(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetInfo(), nil
}
