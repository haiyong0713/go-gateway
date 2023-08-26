package copyright

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/playurl/service/conf"
	"go-gateway/app/app-svr/playurl/service/model"

	"github.com/pkg/errors"
)

const (
	_restrictionURI = "/api/x/copyright/gateway/play_restriction"
)

type Dao struct {
	client         *bm.Client
	restrictionURL string
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:         bm.NewClient(c.HTTPCopyRightClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		restrictionURL: c.HostDiscovery.CopyRight + _restrictionURI,
	}
	return
}

// PlayRestriction .
func (d *Dao) PlayRestriction(c context.Context, aid int64) (*model.CopyRightRestriction, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	var res struct {
		Code int                         `json:"code"`
		Data *model.CopyRightRestriction `json:"data"`
	}
	if err := d.client.Get(c, d.restrictionURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.restrictionURL+"?"+params.Encode())
	}
	if res.Data == nil { //兜底不禁止后台播放,小窗,投屏
		return &model.CopyRightRestriction{Aid: aid}, nil
	}
	return res.Data, nil
}
