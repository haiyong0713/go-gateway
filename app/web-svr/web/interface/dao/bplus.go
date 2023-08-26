package dao

import (
	"context"
	"go-common/library/log"
	"net/url"
	"strconv"

	feedGrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const _dynamicNumURI = "/dynamic_svr/v0/dynamic_svr/space_dy_num"

// DynamicNum get user bplus dynamic num.
func (d *Dao) DynamicNum(c context.Context, mid int64) (num int64, err error) {
	params := url.Values{}
	params.Set("uids", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Items []*struct {
				UID int64 `json:"uid"`
				Num int64 `json:"num"`
			} `json:"items"`
		}
	}
	if err = d.httpR.Get(c, d.dyNumURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dyNumURL+"?"+params.Encode())
		return
	}

	if len(res.Data.Items) > 0 && res.Data.Items[0] != nil && res.Data.Items[0].UID == mid {
		num = res.Data.Items[0].Num
	}
	return
}

func (d *Dao) DynamicNumV2(c context.Context, mid int64) (num int64, err error) {

	res, err := d.dynamicFeedGRPC.SpaceNum(c, &feedGrpc.SpaceNumReq{
		Uid: mid,
	})
	if err != nil {
		log.Error(" dynamicFeedGRPC error(%v)", err)
		return
	}
	num = res.DynNum

	return
}
