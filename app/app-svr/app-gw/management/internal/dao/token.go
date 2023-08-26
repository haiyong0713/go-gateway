package dao

import (
	"context"

	"github.com/google/uuid"
)

const (
	_tokenKey = "{management}/jwt-token"
)

func (d *dao) TokenSecret(ctx context.Context) ([]byte, error) {
	key := _tokenKey
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

func (d *dao) InitialTokenSecret(ctx context.Context) error {
	req := d.taishan.NewCASReq([]byte(_tokenKey), []byte{}, []byte(uuid.New().String()))
	return d.taishan.CAS(ctx, req)
}
