package dao

import (
	"context"
	"net/http"
)

func (d *dao) RawHttpImpl(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	_, body, err := d.http.RawResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
