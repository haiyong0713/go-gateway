package pgc

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"

	freyaComp "git.bilibili.co/bapis/bapis-go/pgc/service/freya/component"

	"go-gateway/app/app-svr/archive/job/conf"
	"go-gateway/app/app-svr/archive/job/model/archive"
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
	if d.freyaCompClient, err = freyaCompNewClient(c.PGCRPC); err != nil {
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

func (d *Dao) Create4UGCPremiere(c context.Context, arc *archive.Archive, premiereTime int64) (res *freyaComp.CreateResp, err error) {
	req := &freyaComp.Create4UGCPremiereReq{
		Aid:            arc.ID,
		Title:          arc.Title,
		Duration:       int64(arc.Duration),
		StartTime:      premiereTime * 1000, //毫秒
		CloseAfter:     d.c.Custom.PremiereCloseAfter,
		Ts:             arc.ID,
		UpMid:          arc.Mid,
		CloseSystemMsg: d.c.Custom.PremiereCloseSystemMsg,
		EndTip:         d.c.Custom.PremiereEndTipSystemMsg,
		RiskCloseTip:   d.c.Custom.PremiereRiskCloseSystemMsg,
	}
	if res, err = d.freyaCompClient.Create4UGCPremiere(c, req); err != nil {
		log.Error("d.freyaCompClient.Create4UGCPremiere error req(%+v) err(%+v)", req, err)
		return
	}
	log.Error("d.freyaCompClient.Create4UGCPremiere req: %+v, res: %+v", req, res)
	return
}
