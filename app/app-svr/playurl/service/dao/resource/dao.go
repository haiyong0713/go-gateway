package resource

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	arcmdl "go-gateway/app/app-svr/playurl/service/model/archive"

	"go-gateway/app/app-svr/playurl/service/conf"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	"github.com/pkg/errors"
)

// Dao is
type Dao struct {
	c *conf.Config
	// rpc
	resRPC     *resrpc.Service
	manager    string
	vipFreeURL string
	client     *httpx.Client
}

const (
	_manager    = "/x/admin/manager/interface/blackwhite/list/scene/oid_list"
	_managerVip = "/x/internal/resource/resolution/limit/free"
)

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// rpc
		resRPC:     resrpc.New(c.ResourceRPC),
		manager:    c.Host.ManagerHost + _manager,
		vipFreeURL: c.Host.APICo + _managerVip,
		client:     httpx.NewClient(c.HTTPClient),
	}
	return
}

// PasterCID get all paster cid.
func (d *Dao) PasterCID(c context.Context) (cids map[int64]int64, err error) {
	if cids, err = d.resRPC.PasterCID(c); err != nil {
		log.Error("d.resRPC.PasterCID() error(%v)", err)
	}
	return
}

func (d *Dao) FetchAllOnlineBlackList(c context.Context) (map[int64]struct{}, error) {
	if d.c.LegoToken == nil {
		return nil, errors.New("d.c.LegoToken.PlayOnlineToken is nil")
	}
	params := url.Values{}
	params.Set("token", d.c.LegoToken.PlayOnlineToken)
	var res struct {
		Code int `json:"code"`
		Data struct {
			Oids []string `json:"oids"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.manager, "", params, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.New(fmt.Sprintf("res.Code(%d)", res.Code))
	}
	blackList := make(map[int64]struct{})
	for _, aidStr := range res.Data.Oids {
		aid, err := strconv.ParseInt(aidStr, 10, 64)
		if err != nil {
			log.Error("strconv.ParseInt error(%+v)", err)
			continue
		}
		blackList[aid] = struct{}{}
	}
	return blackList, nil
}

func (d *Dao) FetchVipList(c context.Context) (map[int64]*arcmdl.VipFree, error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			LimitFreeWithAid map[int64]struct {
				Aid       int64  `json:"aid"`
				LimitFree int32  `json:"limit_free"`
				Subtitle  string `json:"subtitle"`
			} `json:"limit_free_with_aid"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.vipFreeURL, "", nil, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.New(fmt.Sprintf("res.Code(%d)", res.Code))
	}
	rly := make(map[int64]*arcmdl.VipFree)
	if len(res.Data.LimitFreeWithAid) > 0 {
		for _, v := range res.Data.LimitFreeWithAid {
			if v.Aid <= 0 {
				continue
			}
			rly[v.Aid] = &arcmdl.VipFree{
				LimitFree: v.LimitFree,
				Subtitle:  v.Subtitle,
			}
		}
	}
	return rly, nil
}
