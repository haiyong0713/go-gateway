package archive

import (
	"context"

	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

// View3 view archive with pages pb.
func (d *Dao) View(c context.Context, aid int64) (v *api.ViewReply, err error) {
	var arg = &arcgrpc.ViewRequest{Aid: aid}
	if v, err = d.rpcClient.View(c, arg); err != nil {
		return nil, err
	}
	return v, nil
}

// Description get archive description by aid.
func (d *Dao) Description(c context.Context, aid int64) (desc string, err error) {
	var (
		arg     = &arcgrpc.DescriptionRequest{Aid: aid}
		tmpDesc *arcgrpc.DescriptionReply
	)
	if tmpDesc, err = d.rpcClient.Description(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	desc = tmpDesc.Desc
	return
}
