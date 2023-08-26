package tag

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	tagmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/tag"
)

const (
	_mInfo = "/x/internal/tag/minfo"
)

// Dao is tag dao
type Dao struct {
	client *httpx.Client
	mInfo  string
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient),
		mInfo:  c.Host.APICo + _mInfo,
	}
	return
}

// TagInfos get tag infos by tagIds
func (d *Dao) TagInfos(c context.Context, tags []int64, mid int64) (tagMyInfo []*tagmdl.Tag, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("tag_id", xstr.JoinInts(tags))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int           `json:"code"`
		Data []*tagmdl.Tag `json:"data"`
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
