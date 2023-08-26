package popups

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/database/taishan"
	"go-common/library/ecode"
)

type TaishanError struct {
	*taishan.Status
}

func (te *TaishanError) Error() string {
	return fmt.Sprintf("%+v", te.Status)
}

func (d *Dao) newGetReq(key string) *taishan.GetReq {
	req := &taishan.GetReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: []byte(key),
		},
	}
	return req
}

func (d *Dao) PutReq(key, value []byte, ttl uint32) error {
	req := &taishan.PutReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
			Columns: []*taishan.Column{
				{
					Value: value,
				},
			},
			Ttl: ttl,
		},
	}
	reply, err := d.Taishan.client.Put(context.Background(), req)
	if err != nil {
		return err
	}
	return wrapError(reply)
}

func checkRecord(r *taishan.Record) error {
	if r == nil || r.Status == nil {
		return errors.New("record is nil")
	}
	//nolint:gomnd
	if r.Status.ErrNo == 404 {
		return ecode.NothingFound
	}
	if r.Columns == nil || len(r.Columns) == 0 || r.Columns[0] == nil {
		return errors.New("Record.Columns is nil")
	}
	return nil
}

func (d *Dao) getFromTaishan(c context.Context, key string) ([]byte, error) {
	req := d.newGetReq(key)
	resp, err := d.Taishan.client.Get(c, req)
	if err != nil {
		return nil, err
	}
	if err = checkRecord(resp.Record); err != nil {
		return nil, err
	}
	return resp.Record.Columns[0].Value, nil
}

func (d *Dao) GetIsPopFromTaishan(c context.Context, key string) (is_pop bool, err error) {
	bs, err := d.getFromTaishan(c, key)
	if err != nil {
		if err == ecode.NothingFound {
			return false, nil
		}
		return true, err
	}
	if err = json.Unmarshal(bs, &is_pop); err != nil {
		return true, err
	}
	return is_pop, nil
}

func (d *Dao) NewDelReq(key []byte) *taishan.DelReq {
	req := &taishan.DelReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
		},
	}
	return req
}

//func (d *Dao) delteTaishanKey(c context.Context, key string) (err error) {
//	req := d.NewDelReq([]byte(key))
//	reply, err := d.Taishan.client.Del(c, req)
//	if err != nil {
//		return err
//	}
//	return wrapError(reply)
//}

type statusGetter interface {
	GetStatus() *taishan.Status
}

func NewTaishanError(status *taishan.Status) error {
	if status == nil {
		return errors.New("input status is invalid")
	}
	if status.ErrNo == 0 {
		return nil
	}
	return errors.WithStack(&TaishanError{Status: status})
}

func wrapError(reply statusGetter) error {
	return NewTaishanError(reply.GetStatus())
}
