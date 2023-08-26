package bplus

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	"github.com/pkg/errors"
)

const (
	_dynamicDetail = "/dynamic_detail/v0/dynamic/details"
	_fromFeed      = "tianma"
)

// Dao is a dao.
type Dao struct {
	// http client
	client      *httpx.Client
	dynamicGRPC dyngrpc.FeedClient
	// ad
	dynamicDetail string
}

// New new a dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:        httpx.NewClient(c.HTTPClient),
		dynamicDetail: c.Host.DynamicCo + _dynamicDetail,
	}
	var err error
	if d.dynamicGRPC, err = dyngrpc.NewClient(c.DynamicGRPC); err != nil {
		panic(err)
	}
	return
}

// DynamicDetail is.
func (d *Dao) DynamicDetail(c context.Context, platfrom, mobiApp, device string, build int, ids ...int64) (picm map[int64]*bplus.Picture, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Add("from", _fromFeed)
	for _, id := range ids {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var mate = &struct {
		Platfrom string `json:"platform"`
		MobiApp  string `json:"mobi_app"`
		Device   string `json:"device"`
		Build    string `json:"build"`
	}{
		Platfrom: platfrom,
		MobiApp:  mobiApp,
		Device:   device,
		Build:    strconv.Itoa(build),
	}
	pb, _ := json.Marshal(mate)
	params.Add("meta", string(pb))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			List []*bplus.Picture `json:"list"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.dynamicDetail, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dynamicDetail+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		picm = make(map[int64]*bplus.Picture, len(res.Data.List))
		for _, pic := range res.Data.List {
			picm[pic.DynamicID] = pic
		}
	}
	return
}

func (d *Dao) DynamicGeneralStory(ctx context.Context, param *dyngrpc.GeneralStoryReq) (*dyngrpc.GeneralStoryRsp, error) {
	reply, err := d.dynamicGRPC.GeneralStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) DynamicSpaceStory(ctx context.Context, param *dyngrpc.SpaceStoryReq) (*dyngrpc.SpaceStoryRsp, error) {
	reply, err := d.dynamicGRPC.SpaceStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) DynamicInsert(ctx context.Context, param *dyngrpc.InsertedStoryReq) (*dyngrpc.InsertedStoryRsp, error) {
	reply, err := d.dynamicGRPC.InsertedStory(ctx, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
