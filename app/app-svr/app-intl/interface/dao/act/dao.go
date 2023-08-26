package act

import (
	"context"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/pkg/errors"
)

// Dao is act dao.
type Dao struct {
	actGRPC actgrpc.ActivityClient
}

// New elec dao
func New(c *conf.Config) (d *Dao) {
	act, err := actgrpc.NewClient(c.ActivityClient)
	if err != nil {
		panic(err)
	}
	d = &Dao{actGRPC: act}
	return
}

// ActProtocol get act subject & protocol
func (d *Dao) ActProtocol(c context.Context, messionID int64) (protocol *actgrpc.ActSubProtocolReply, err error) {
	arg := &actgrpc.ActSubProtocolReq{Sid: messionID}
	if protocol, err = d.actGRPC.ActSubProtocol(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}
