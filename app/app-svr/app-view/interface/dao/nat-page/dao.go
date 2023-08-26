package nat_page

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-view/interface/conf"

	natgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

type Dao struct {
	natGRPC natgrpc.NaPageClient
}

// New elec dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.natGRPC, err = natgrpc.NewClient(c.NatClient); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) NatInfoFromForeign(c context.Context, tids []int64, pageType int64, content map[string]string) (res map[int64]*natgrpc.NativePage, err error) {
	var (
		args = &natgrpc.NatInfoFromForeignReq{
			Fids:     tids,
			PageType: pageType,
			Content:  content,
		}
		resTmp *natgrpc.NatInfoFromForeignReply
	)
	if resTmp, err = d.natGRPC.NatInfoFromForeign(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	// 木有getList方法
	if resTmp != nil {
		res = resTmp.List
	}
	return
}
