package dm

import (
	"context"
	"fmt"

	"go-common/component/metadata/device"

	"go-gateway/app/app-svr/app-view/interface/conf"

	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	buzzword "git.bilibili.co/bapis/bapis-go/community/interface/dm-buzzword"

	"github.com/pkg/errors"
)

type SubjectInfosReq struct {
	Typ  int32
	Plat int8
	Cids []int64
}

type Dao struct {
	dmGRPC   dmApi.DMClient
	buzzword buzzword.BuzzwordInterfaceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.dmGRPC, err = dmApi.NewClient(c.DMClient); err != nil {
		panic(fmt.Sprintf("DMClient not found err(%v)", err))
	}
	if d.buzzword, err = buzzword.NewClient(c.Buzzword); err != nil {
		panic(errors.WithStack(err))
	}
	return
}

func (d *Dao) SubjectInfos(c context.Context, typ int32, plat int8, oids ...int64) (map[int64]*dmApi.SubjectInfo, error) {
	arg := &dmApi.SubjectInfosReq{Type: typ, Plat: int32(plat), Oids: oids}
	reply, err := d.dmGRPC.SubjectInfos(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return reply.GetInfos(), nil
}

func (d *Dao) Commands(c context.Context, aid, cid, mid int64, dev device.Device) (commands []*dmApi.CommandDm, err error) {
	arg := &dmApi.CommandDmsReq{
		Aid: aid,
		Cid: cid,
		Mid: mid,
		Common: &dmApi.CommonParam{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
		},
	}
	reply, err := d.dmGRPC.CommandDms(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%+v", arg)
		return
	}
	commands = reply.GetCommandDms()
	return
}

func (d *Dao) BuzzwordShowConfigPeriod(ctx context.Context, req *buzzword.BuzzwordShowConfigPeriodReq) (*buzzword.BuzzwordShowConfigPeriodReply, error) {
	return d.buzzword.BuzzwordShowConfigPeriod(ctx, req)
}
