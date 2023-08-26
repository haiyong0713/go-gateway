package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-channel/interface/conf"
	dynmdl "go-gateway/app/app-svr/app-channel/interface/model/dynamic"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// dynGrpc 动态feed流client
	dynGrpc dyngrpc.FeedClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.dynGrpc, err = dyngrpc.NewClient(c.DynGRPC); err != nil {
		panic(fmt.Sprintf("dynGrpc NewClient error(%v)", err))
	}
	return d
}

func (d *Dao) DrawDetails(c context.Context, mid int64, drawIds []int64) (map[int64]*dynmdl.DrawDetailRes, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	for _, id := range drawIds {
		params.Add("ids[]", strconv.FormatInt(id, 10))
	}
	drawDetails := d.c.Host.VcCo + "/link_draw/v0/Doc/dynamicDetailsV2"
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		// Data []*dymdl.DrawDetailRes `json:"data"`
		Data struct {
			DocItems []struct {
				DocID string `json:"doc_id"`
				Item  string `json:"item"`
			} `json:"doc_items"`
		} `json:"data"`
	}
	if err := d.client.Get(c, drawDetails, "", params, &ret); err != nil {
		log.Error("DrawDetails http GET(%s) failed, params:(%s), error(%+v)", drawDetails, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Error("DrawDetails http GET(%s) failed, params:(%s), code: %v, msg: %v", drawDetails, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DrawDetails url(%v) code(%v) msg(%v)", drawDetails, ret.Code, ret.Msg)
		return nil, err
	}
	res := make(map[int64]*dynmdl.DrawDetailRes)
	for _, d := range ret.Data.DocItems {
		rid, err := strconv.ParseInt(d.DocID, 10, 64)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		var item *dynmdl.DrawDetailRes
		err = json.Unmarshal([]byte(d.Item), &item)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		res[rid] = item
	}
	return res, nil
}

func (d *Dao) DynSimpleInfos(ctx context.Context, args *dyngrpc.DynSimpleInfosReq) (*dyngrpc.DynSimpleInfosRsp, error) {
	return d.dynGrpc.DynSimpleInfos(ctx, args)
}
