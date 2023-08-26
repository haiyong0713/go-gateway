package archive

import (
	"context"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

func (d *Dao) ContentFlowControlInfosV2(ctx context.Context, req *cfcgrpc.FlowCtlInfosReq) (*cfcgrpc.FlowCtlInfosV2Reply, error) {
	return d.cfcGRPC.InfosV2(ctx, req)
}
