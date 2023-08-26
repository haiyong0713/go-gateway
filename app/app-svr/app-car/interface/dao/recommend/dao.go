package recommend

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/recommend"

	"github.com/pkg/errors"
)

const (
	_rcmd     = "/recommand"
	_rcmdfeed = "/pegasus/feed/%d"
)

type Dao struct {
	c        *conf.Config
	client   *bm.Client
	rcmd     string
	rcmdfeed string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:        c,
		client:   bm.NewClient(c.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		rcmd:     c.HostDiscovery.Data + _rcmd,
		rcmdfeed: c.HostDiscovery.Data + _rcmdfeed,
	}
	return d
}

func (d *Dao) Relate(c context.Context, mid, fromAv int64, buvid string) ([]*recommend.Item, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("cmd", "web_related")
	timeout := time.Duration(d.c.HTTPData.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("from_av", strconv.FormatInt(fromAv, 10))
	params.Set("buvid", buvid)
	params.Set("webpage", "car_media")
	params.Set("web_rm_repeat", "1")
	var res struct {
		Code int               `json:"code"`
		Data []*recommend.Item `json:"data"`
	}
	if err := d.client.Get(c, d.rcmd, ip, params, &res); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		if res.Code == -3 { // code -3 结果数量不足（暂时无结果可出）
			return []*recommend.Item{}, nil
		}
		err := errors.Wrap(ecode.Int(res.Code), d.rcmd+"?"+params.Encode())
		log.Error("%v", err)
		return nil, err
	}
	return res.Data, nil
}

func (d *Dao) FeedRecommend(c context.Context, plat int8, mobiApp, buvid string, mid int64, build, loginEvent, group, count, mode int) ([]*ai.Item, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if mid == 0 && buvid == "" {
		return nil, nil
	}
	params := url.Values{}
	uri := fmt.Sprintf(d.rcmdfeed, group)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("plat", "2") // 目前直接获取ipad数据
	params.Set("build", strconv.Itoa(build))
	params.Set("login_event", strconv.Itoa(loginEvent))
	params.Set("pagetype", "1")
	params.Set("column", "1")
	params.Set("style", "1")
	params.Set("request_cnt", strconv.Itoa(count))
	timeout := time.Duration(d.c.HTTPData.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("disable_rcmd", strconv.Itoa(mode)) // 1关闭个性化
	var res struct {
		Code int        `json:"code"`
		Data []*ai.Item `json:"data"`
	}
	if err := d.client.Get(c, uri, ip, params, &res); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		if res.Code == -3 { // code -3 结果数量不足（暂时无结果可出）
			return nil, nil
		}
		err := errors.Wrap(ecode.Int(res.Code), uri+"?"+params.Encode())
		log.Error("%v", err)
		return nil, err
	}
	return res.Data, nil
}

// RelatePgc pgc相关推荐.
func (d *Dao) RelatePgc(c context.Context, mid, seasonId, build, loginEvent int64, buvid string) ([]*recommend.Item, error) {
	params := url.Values{}
	//params.Set("cmd", "vehicle_ogv_related")
	params.Set("cmd", "ogv_related") // TODO
	params.Set("timeout", "500")
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.FormatInt(build, 10))
	params.Set("plat", "0")
	params.Set("login_event", strconv.FormatInt(loginEvent, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("request_cnt", "40")
	params.Set("parent_mode", "0")
	params.Set("from_av", strconv.FormatInt(seasonId, 10))
	var res struct {
		Code int               `json:"code"`
		Data []*recommend.Item `json:"data"`
	}
	if err := d.client.Get(c, d.rcmd, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		log.Error("RelatePgc d.client.Get err=%+v, buvid=%s.", err, buvid)
		return nil, err
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		if res.Code == -3 {
			log.Error("RelatePgc ai code=-3. sid=%d, buvid=%s.", seasonId, buvid)
			return []*recommend.Item{}, nil
		}
		err := errors.Wrap(ecode.Int(res.Code), d.rcmd+"?"+params.Encode())
		log.Error("RelatePgc ai err=%+v, code=%d, buvid=%s.", err, res.Code, buvid)
		return nil, err
	}
	return res.Data, nil
}
