package pgc

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"go-common/library/net/rpc/warden"

	freyaComp "git.bilibili.co/bapis/bapis-go/pgc/service/freya/component"
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
	if d.freyaCompClient, err = freyaCompNewClient(c.FreyaClient); err != nil {
		panic(fmt.Sprintf("freyaCompoent NewClient error (%+v)", err))
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

func (d *Dao) GetUGCPremiereRoomStatus(c context.Context, roomId int64) (*freyaComp.RoomStatusResp, error) {
	req := &freyaComp.GetUGCPremiereStatusReq{
		RoomId: roomId,
	}
	res, err := d.freyaCompClient.GetUGCPremiereRoomStatus(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.GetUGCPremiereRoomStatus err req(%+v)", req)
		return nil, err
	}
	return res, nil
}
