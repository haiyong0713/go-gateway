package popular

import (
	"context"

	pgrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"
)

type Dao struct {
	c          *conf.Config
	grpcClient pgrpc.PopularClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = pgrpc.NewClient(c.PopularGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) TimeLine(c context.Context, lid int64, offset, ps int32) (*pgrpc.TimeLineReply, error) {
	rly, err := d.grpcClient.TimeLine(c, &pgrpc.TimeLineRequest{LineId: lid, Offset: offset, Ps: ps})
	if err != nil {
		log.Error("d.grpcClient.TimeLine(%d,%d,%d) error(%v)", lid, offset, ps, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) Arcs(c context.Context) (*pgrpc.ArcsResp, error) {
	return d.grpcClient.Arcs(c, &pgrpc.Empty{})
}

// PageArcs .
func (d *Dao) PageArcs(c context.Context, offset, ps int64, arcType int32) (*pgrpc.PageArcsResp, error) {
	rly, err := d.grpcClient.PageArcs(c, &pgrpc.PageArcsReq{Offset: offset, PageSize: ps, ArcType: arcType})
	if err != nil {
		log.Error("d.grpcClient.PageArcse(%d,%d) error(%v)", offset, ps, err)
		return nil, err
	}
	return rly, nil
}
