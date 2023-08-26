package v1

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	"github.com/pkg/errors"
)

func (d *dao) TagInfos(c context.Context, tags []int64, mid int64) (tagMyInfo []*search.Tag, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("tag_id", xstr.JoinInts(tags))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int           `json:"code"`
		Data []*search.Tag `json:"data"`
	}
	if err = d.client.Get(c, d.mInfo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.mInfo+"?"+params.Encode())
		return
	}
	tagMyInfo = res.Data
	return
}
