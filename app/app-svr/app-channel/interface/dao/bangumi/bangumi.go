package bangumi

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-channel/interface/conf"

	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"github.com/pkg/errors"
)

const (
	_seasons = "/api/inner/aid_episodes_v2"
)

// Dao is bangumi dao.
type Dao struct {
	// http client
	client *bm.Client
	// bangumi
	seasons string
	// grpc
	rpcClient      seasongrpc.SeasonClient
	rpcEpidsClient episodegrpc.EpisodeClient
	appCardClient  appCardgrpc.AppCardClient
}

// New new a bangumi dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http clients
		client:  bm.NewClient(c.HTTPClient),
		seasons: c.Host.Bangumi + _seasons,
	}
	var err error
	if d.rpcClient, err = seasongrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("seasongrpc NewClientt error (%+v)", err))
	}
	if d.rpcEpidsClient, err = episodegrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("episodegrpc NewClientt error (%+v)", err))
	}
	if d.appCardClient, err = appCardgrpc.NewClient(c.PGCRPC); err != nil {
		panic(fmt.Sprintf("appCardgrpc NewClientt error (%+v)", err))
	}
	return d
}

// Seasons bangumi Season .
func (d *Dao) Seasons(c context.Context, aids []int64, now time.Time) (sm map[int64]*bangumi.Season, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("aids", xstr.JoinInts(aids))
	params.Set("type", "av")
	params.Set("build", "app-feed")
	params.Set("platform", "Golang")
	var res struct {
		Code   int                       `json:"code"`
		Result map[int64]*bangumi.Season `json:"result"`
	}
	if err = d.client.Get(c, d.seasons, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.seasons+"?"+params.Encode())
		return
	}
	sm = res.Result
	return
}
