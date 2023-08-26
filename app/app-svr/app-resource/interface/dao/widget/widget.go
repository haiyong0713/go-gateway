package widget

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	model "go-gateway/app/app-svr/app-resource/interface/model/widget"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"
)

const _hotWord = "/query/recommend"

type Dao struct {
	accClient accgrpc.AccountClient

	client *httpx.Client
	hot    string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClientAsyn),
		hot:    c.Host.Search + _hotWord,
	}
	var err error
	if d.accClient, err = accgrpc.NewClient(c.AccountClient); err != nil {
		panic(err)
	}
	return
}

// HotSearch is to search hot words in a list.
func (d *Dao) HotSearch(c context.Context, r *model.WidgetsMetaReq) (*model.Hot, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("main_ver", "v3")
	params.Set("platform", r.Platform)
	params.Set("mobi_app", r.MobiApp)
	params.Set("clientip", ip)
	params.Set("device", r.Device)
	params.Set("build", strconv.FormatInt(r.Build, 10))
	params.Set("search_type", "default")
	params.Set("req_source", "0")
	params.Set("is_new", "1")
	req, err := d.client.NewRequest("GET", d.hot, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", r.Buvid)
	res := &model.Hot{}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.hot+"?"+params.Encode())
	}
	return res, nil
}

func (d *Dao) Card3(ctx context.Context, req *accgrpc.MidReq) (*accgrpc.Card, error) {
	reply, err := d.accClient.Card3(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.accClient.Cards3 req=%+v", req)
	}
	return reply.Card, nil
}
