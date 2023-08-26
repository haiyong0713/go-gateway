package util

import (
	arcmidv1 "go-gateway/app/app-svr/archive/middleware/v1"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

func Trans2PlayerArgs(arg *arcgrpc.BatchPlayArg) *arcmidv1.PlayerArgs {
	if arg == nil {
		return nil
	}
	return &arcmidv1.PlayerArgs{
		Qn:        arg.Qn,
		Fnver:     arg.Fnver,
		Fnval:     arg.Fnval,
		ForceHost: arg.ForceHost,
	}
}
