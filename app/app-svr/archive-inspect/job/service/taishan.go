package service

import (
	"context"

	"go-common/library/database/taishan"
	"go-common/library/ecode"

	"github.com/pkg/errors"
)

func (s *Service) getFromTaishan(c context.Context, key string) ([]byte, error) {
	req := &taishan.GetReq{
		Table: s.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: s.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: []byte(key),
		},
	}
	resp, err := s.Taishan.client.Get(c, req)
	if err != nil {
		return nil, err
	}
	if err = checkRecord(resp.Record); err != nil {
		return nil, err
	}
	return resp.Record.Columns[0].Value, nil
}

// nolint:gomnd
func checkRecord(r *taishan.Record) error {
	if r == nil || r.Status == nil {
		return errors.New("record is nil")
	}
	if r.Status.ErrNo == 404 {
		return ecode.NothingFound
	}
	if r.Columns == nil || len(r.Columns) == 0 || r.Columns[0] == nil {
		return errors.New("Record.Colums is nil")
	}
	return nil
}
