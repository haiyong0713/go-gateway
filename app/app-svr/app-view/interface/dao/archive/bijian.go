package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-gateway/app/app-svr/app-view/interface/model/view"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const (
	_biJianURL = "/x/internal/material/views"
)

func (d *Dao) GetBiJianMaterial(c context.Context, req *view.BiJianMaterialReq) (*view.BiJianMaterialReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("type", strconv.FormatInt(req.Type, 10))
	params.Set("ids", strconv.FormatInt(req.Ids, 10))
	params.Set("biz", strconv.FormatInt(req.Biz, 10))
	res := &view.BiJianMaterialReply{}
	err := d.httpClient.Get(c, d.biJianURL, ip, params, &res)
	if err != nil {
		log.Error("GetBiJianMaterial is err(%+v) url(%s)", err, d.biJianURL+"?"+params.Encode())
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.biJianURL+"?"+params.Encode())
		return nil, err
	}
	return res, nil
}
