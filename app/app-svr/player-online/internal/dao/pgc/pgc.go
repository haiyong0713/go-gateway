package pgc

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	freyaComp "git.bilibili.co/bapis/bapis-go/pgc/service/freya/component"

	"go-gateway/app/app-svr/player-online/internal/conf"
)

type Dao struct {
	c               *conf.Config
	freyaCompClient freyaComp.ComponentInnerClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.freyaCompClient, err = freyaCompNewClient(c.PgcRPC); err != nil {
		panic(fmt.Sprintf("strategygrpc NewClient error (%+v)", err))
	}
	return d
}

// NewClient new a grpc client
func freyaCompNewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (freyaComp.ComponentInnerClient, error) {
	const (
		_appID = "pgc.service.freya"
	)
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+_appID)
	if err != nil {
		return nil, err
	}
	return freyaComp.NewComponentInnerClient(conn), nil
}

func (d *Dao) GetUGCPremiereRoomStatistics(c context.Context, roomId int64) (res *freyaComp.RoomStatisticsResp, err error) {
	req := &freyaComp.GetUGCPremiereStatisticsReq{
		RoomId: roomId,
	}
	if res, err = d.freyaCompClient.GetUGCPremiereRoomStatistics(c, req); err != nil {
		err = errors.Wrapf(err, "d.GetUGCPremiereRoomStatistics err req(%+v)", req)
		return
	}
	log.Error("d.GetUGCPremiereRoomStatistics req: %+v, res: %+v", req, res)
	return
}
