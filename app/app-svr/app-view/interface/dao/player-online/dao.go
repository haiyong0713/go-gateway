package player_online

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/conf"

	playeronline "git.bilibili.co/bapis/bapis-go/bilibili/app/playeronline/v1"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

const appID = "main.app-svr.player-online"

type Dao struct {
	c        *conf.Config
	poClient playeronline.PlayerOnlineClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.poClient, err = NewClient(c.PlayerOnlineGRPC); err != nil {
		panic(err)
	}
	return
}

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (playeronline.PlayerOnlineClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return playeronline.NewPlayerOnlineClient(conn), nil
}

func (d *Dao) ReportWatch(ctx context.Context, aid int64, buvid string) error {
	_, err := d.poClient.ReportWatch(ctx, &playeronline.ReportWatchReq{
		Aid:   aid,
		Buvid: buvid,
		Biz:   "app",
	})
	if err != nil {
		log.Error("日志告警 ReportWatch error(%+v) aid(%d), buvid(%s)", err, aid, buvid)
	}
	return err
}
