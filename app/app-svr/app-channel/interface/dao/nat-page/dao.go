package nat_page

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-channel/interface/conf"

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

func (d *Dao) NatInfoFromForeigns(c context.Context, tids []int64, pageType int64) (res map[int64]*natgrpc.NativePage, err error) {
	var maxLimit = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res = make(map[int64]*natgrpc.NativePage)
	for i := 0; i < len(tids); i += maxLimit {
		var tidGroups []int64
		if i+maxLimit <= len(tids) {
			tidGroups = tids[i : i+maxLimit]
		} else {
			tidGroups = tids[i:]
		}
		g.Go(func(ctx context.Context) error {
			var tmpRes map[int64]*natgrpc.NativePage
			if tmpRes, err = d.NatInfoFromForeign(ctx, tidGroups, pageType); err != nil {
				log.Error("%+v", err)
				return err
			}
			mu.Lock()
			for key, value := range tmpRes {
				res[key] = value
			}
			mu.Unlock()
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return
}
func (d *Dao) NatInfoFromForeign(c context.Context, tids []int64, pageType int64) (res map[int64]*natgrpc.NativePage, err error) {
	var (
		args   = &natgrpc.NatInfoFromForeignReq{Fids: tids, PageType: pageType}
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
