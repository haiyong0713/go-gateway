package like

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	likemdl "go-gateway/app/web-svr/native-page/interface/model/like"
)

const (
	_arcTypeListURI = "/x/internal/creative/archive/typelist"
)

func (d *Dao) ArcTypeList(c context.Context, mid int64) ([]*likemdl.ArcType, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ip", ip)
	var res struct {
		Code int                `json:"code"`
		Data []*likemdl.ArcType `json:"data"`
	}
	if err := d.client.Get(c, d.arcTypeListURL, ip, params, &res); err != nil {
		log.Error("Fail to request ArcTypeList, mid=%d error=%+v", mid, err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.arcTypeListURL+"?"+params.Encode())
		log.Error("Fail to request ArcTypeList, error=%+v", err)
		return nil, err
	}
	return res.Data, nil
}
