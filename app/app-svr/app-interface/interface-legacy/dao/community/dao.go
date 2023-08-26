package community

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/community"

	contractgrpc "git.bilibili.co/bapis/bapis-go/community/service/contract"

	"github.com/pkg/errors"
)

const (
	_comm = "/api/query.my.community.list.do"
)

// Dao is community dao
type Dao struct {
	client         *httpx.Client
	community      string
	contractClient contractgrpc.ContractClient
}

// New initial community dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:    httpx.NewClient(c.HTTPIm9),
		community: c.Host.Im9 + _comm,
	}
	var err error
	if d.contractClient, err = contractgrpc.NewClient(c.ContractGRPC); err != nil {
		panic(err)
	}
	return
}

// 契约者卡片配置
func (d *Dao) ContractShowConfig(ctx context.Context, req *contractgrpc.ShowConfigReq) (*contractgrpc.ShowConfigReply, error) {
	return d.contractClient.ShowConfig(ctx, req)
}

// Community get community data from api.
func (d *Dao) Community(c context.Context, mid int64, ak, platform string, pn, ps int) (co []*community.Community, count int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("actionKey", "appkey")
	params.Set("data_type", "2")
	params.Set("access_key", ak)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("page_no", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("platform", platform)
	var res struct {
		Code int `json:"code"`
		Data *struct {
			Count  int                    `json:"total_count"`
			Result []*community.Community `json:"result"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.community, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.community+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		co = res.Data.Result
		count = res.Data.Count
	}
	return
}
