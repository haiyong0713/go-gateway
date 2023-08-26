package v1

import (
	"context"

	siriext "go-gateway/app/app-svr/siri-ext/service/api"
)

func (d *dao) ResolveCommand(ctx context.Context, req *siriext.ResolveCommandReq) (*siriext.ResolveCommandReply, error) {
	return d.siriExtClient.ResolveCommand(ctx, req)
}
