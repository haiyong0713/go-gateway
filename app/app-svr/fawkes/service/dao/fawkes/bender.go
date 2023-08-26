package fawkes

import (
	"context"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/fawkes/service/model/bender"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_topResource = "/top_resources"
)

func (d *Dao) BenderTopResource(ctx context.Context) (res *bender.ResourceData, err error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	req, err := d.httpClient.NewRequest(http.MethodGet, d.topResource, ip, nil)
	if err != nil {
		return nil, err
	}
	resp := new(bender.TopResourceResp)
	err = d.httpClient.Do(ctx, req, &resp)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if resp.Code != ecode.OK.Code() {
		err = ecode.Error(ecode.Int(resp.Code), resp.Message)
		return
	}
	res = resp.Data
	return
}
