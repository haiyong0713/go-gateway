package dao

import (
	"context"
	"fmt"

	"go-common/library/conf/paladin.v2"
	"go-common/library/database/taishan"

	"github.com/pkg/errors"
)

type TaishanError struct {
	*taishan.Status
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type tableConfig struct {
	Table string
	Token string
}

type statusGetter interface {
	GetStatus() *taishan.Status
}

func wrapError(reply statusGetter) error {
	return NewTaishanError(reply.GetStatus())
}

func NewKV() (*Taishan, func(), error) {
	var (
		cfg tableConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("kv.toml").Unmarshal(&ct); err != nil {
		return nil, nil, err
	}
	if err := ct.Get("Table").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, err
	}
	client, err := taishan.NewClient(nil)
	if err != nil {
		return nil, nil, err
	}
	taishan := &Taishan{
		client:   client,
		tableCfg: cfg,
	}
	cf := func() {}
	return taishan, cf, nil
}

func (te *TaishanError) Error() string {
	return fmt.Sprintf("%+v", te.Status)
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

func (ts *Taishan) NewGetReq(key []byte) *taishan.GetReq {
	req := &taishan.GetReq{
		Table: ts.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: ts.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
		},
	}
	return req
}

func (ts *Taishan) Get(ctx context.Context, req *taishan.GetReq) (*taishan.Record, error) {
	reply, err := ts.client.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	record := reply.Record
	if err := wrapError(record); err != nil {
		return nil, err
	}
	return reply.Record, nil
}

func (ts *Taishan) NewPutReq(key, value []byte, ttl uint32) *taishan.PutReq {
	req := &taishan.PutReq{
		Table: ts.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: ts.tableCfg.Token,
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
	return req
}

func (ts *Taishan) Put(ctx context.Context, req *taishan.PutReq) error {
	reply, err := ts.client.Put(ctx, req)
	if err != nil {
		return err
	}
	return wrapError(reply)
}

func (ts *Taishan) NewScanReq(start, end []byte, limit uint32) *taishan.ScanReq {
	req := &taishan.ScanReq{
		Table: ts.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: ts.tableCfg.Token,
		},
		StartRec: &taishan.Record{
			Key: start,
		},
		EndRec: &taishan.Record{
			Key: end,
		},
		Limit: limit,
	}
	return req
}

func (ts *Taishan) Scan(ctx context.Context, req *taishan.ScanReq) (*taishan.ScanResp, error) {
	reply, err := ts.client.Scan(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := wrapError(reply); err != nil {
		return nil, err
	}
	return reply, nil
}

func (ts *Taishan) NewCASReq(key, oldV, newV []byte) *taishan.CasReq {
	cond := &taishan.CasCond{
		Method: taishan.CasCond_EQUALS,
		Records: []*taishan.Record{
			{
				Columns: []*taishan.Column{
					{
						Value: oldV,
					},
				},
				Key: key,
			},
		},
	}
	req := &taishan.CasReq{
		Table: ts.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: ts.tableCfg.Token,
		},
		Cond: cond,
		Records: []*taishan.Record{
			{
				Key: key,
				Columns: []*taishan.Column{
					{
						Value: newV,
					},
				},
			},
		},
	}
	return req
}

func (ts *Taishan) CAS(ctx context.Context, req *taishan.CasReq) error {
	reply, err := ts.client.Cas(ctx, req)
	if err != nil {
		return err
	}
	return wrapError(reply)
}

func (ts *Taishan) NewDelReq(key []byte) *taishan.DelReq {
	req := &taishan.DelReq{
		Table: ts.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: ts.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
		},
	}
	return req
}

func (ts *Taishan) Del(ctx context.Context, req *taishan.DelReq) error {
	reply, err := ts.client.Del(ctx, req)
	if err != nil {
		return err
	}
	return wrapError(reply)
}

func (ts *Taishan) TryLock(ctx context.Context, key string, oldVal, newVal []byte, ttl uint32) error {
	req := ts.NewCASReq([]byte(key), oldVal, newVal)
	req.Records[0].Ttl = ttl
	return ts.CAS(ctx, req)
}

func (ts *Taishan) GetKey(ctx context.Context, key string) ([]byte, error) {
	req := ts.NewGetReq([]byte(key))
	record, err := ts.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

func (ts *Taishan) DelKey(ctx context.Context, key string) error {
	req := ts.NewDelReq([]byte(key))
	return ts.Del(ctx, req)
}

func (ts *Taishan) PutKey(ctx context.Context, key, value []byte, ttl uint32) error {
	req := ts.NewPutReq(key, value, ttl)
	return ts.Put(ctx, req)
}
