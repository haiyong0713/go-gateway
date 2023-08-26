package shop

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/native-page/interface/conf"

	"github.com/pkg/errors"
)

const (
	_favStatURI = "/api/ticket/user/favstatusinner"
)

type Dao struct {
	c          *conf.Config
	client     *httpx.Client
	favStatURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		client:     httpx.NewClient(c.HTTPShowCo),
		favStatURL: c.Host.ShowCo + _favStatURI,
	}
	return
}

func (d *Dao) BatchMultiFavStat(c context.Context, itemIDs []int64, mid int64) map[int64]bool {
	// 去重
	idsSet := make(map[int64]struct{})
	for _, v := range itemIDs {
		idsSet[v] = struct{}{}
	}
	var ids []int64
	for id := range idsSet {
		if id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return make(map[int64]bool)
	}
	var (
		idsLen = len(ids)
		mutex  = sync.Mutex{}
		maxIDs = 20
	)
	aidMap := make(map[int64]bool)
	gp := errgroup.WithContext(c)
	for i := 0; i < idsLen; i += maxIDs {
		var partAids []int64
		if i+maxIDs > idsLen {
			partAids = ids[i:]
		} else {
			partAids = ids[i : i+maxIDs]
		}
		gp.Go(func(ctx context.Context) error {
			tmpRes, err := d.MultiFavStat(ctx, partAids, mid)
			if err != nil { //错误忽略，降级处理
				log.Error("d.MultiFavStat(%v,%d) error(%v)", partAids, mid, err)
				return nil
			}
			if len(tmpRes) > 0 {
				mutex.Lock()
				for k, v := range tmpRes {
					aidMap[k] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	_ = gp.Wait()
	return aidMap
}

func (d *Dao) MultiFavStat(c context.Context, itemIDs []int64, mid int64) (map[int64]bool, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("item_id", xstr.JoinInts(itemIDs))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Result map[string]bool `json:"result"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.favStatURL, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.favStatURL+"?"+params.Encode())
		return nil, err
	}
	rly := make(map[int64]bool)
	if res.Data.Result == nil {
		return rly, nil
	}
	for k, v := range res.Data.Result {
		if id, e := strconv.ParseInt(k, 10, 64); e == nil {
			rly[id] = v
		}
	}
	return rly, nil
}
