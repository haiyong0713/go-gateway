package dao

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/player/interface/model"
	"net/url"
)

const _limitFreeURL = "/x/internal/resource/resolution/limit/free"

func (d *Dao) LimitFree(c context.Context) (res map[int64]*model.LimitFreeInfo, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	var rs struct {
		Code int `json:"code"`
		Data struct {
			LimitFreeWithAid map[int64]*model.LimitFreeInfo `json:"limit_free_with_aid"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.c.Host.LimitFreeUrl+_limitFreeURL, ip, params, &rs); err != nil {
		log.Error("limitFreeList d.client.Get(%s,%s,%v) error(%v)", d.c.Host.LimitFreeUrl+_limitFreeURL, ip, params, err)
		return
	}
	if rs.Code != 0 {
		return nil, errors.Wrap(ecode.Int(rs.Code), d.c.Host.LimitFreeUrl+_limitFreeURL+"?"+params.Encode())
	}
	return rs.Data.LimitFreeWithAid, nil
}
