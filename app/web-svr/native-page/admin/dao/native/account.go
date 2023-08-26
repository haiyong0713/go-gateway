package native

import (
	"context"
	"go-common/library/ecode"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
)

func (d *Dao) Info3(c context.Context, mid int64) (*acccli.Info, error) {
	rly, err := d.accGRPC.Info3(c, &acccli.MidReq{Mid: mid})
	if err != nil {
		log.Error("d.accGRPC.Info3(%d) error(%v)", mid, err)
		return nil, err
	}
	if rly != nil {
		return rly.Info, nil
	}
	return nil, ecode.RequestErr
}

func (d *Dao) Infos3(c context.Context, mids []int64) (map[int64]*acccli.Info, error) {
	rly, err := d.accGRPC.Infos3(c, &acccli.MidsReq{Mids: mids})
	if err != nil {
		log.Error("d.accGRPC.Infos3(%v) error(%v)", mids, err)
		return nil, err
	}
	if rly != nil {
		return rly.Infos, nil
	}
	return nil, ecode.RequestErr
}
