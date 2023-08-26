package v2

import (
	"context"

	apiV2 "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

func (s *CampusServer) WaterFlowRcmd(ctx context.Context, req *apiV2.WaterFlowRcmdReq) (*apiV2.WaterFlowRcmdResp, error) {
	return s.dynSvr.WaterFlowRcmd(s.buildPlayerArgs(ctx, req.PlayerArgs), mdlv2.NewGeneralParamFromCtx(ctx), req)
}
