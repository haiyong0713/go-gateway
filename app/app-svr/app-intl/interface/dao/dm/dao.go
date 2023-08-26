package dm

import (
	"context"
	"fmt"

	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/pkg/errors"
)

// Dao struct
type Dao struct {
	dmGRPC dmApi.DMClient
}

// New a dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.dmGRPC, err = dmApi.NewClient(c.DMClient); err != nil {
		panic(fmt.Sprintf("DMClient not found err(%v)", err))
	}
	return
}

// SubjectInfos
func (d *Dao) SubjectInfos(c context.Context, typ int32, plat int8, oids ...int64) (map[int64]*dmApi.SubjectInfo, error) {
	arg := &dmApi.SubjectInfosReq{Type: typ, Plat: int32(plat), Oids: oids}
	reply, err := d.dmGRPC.SubjectInfos(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return reply.GetInfos(), nil
}
