package playurl

import (
	"context"
	"sync"

	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	playurlgrpc "git.bilibili.co/bapis/bapis-go/playurl/service"
)

type Dao struct {
	c           *conf.Config
	playurlGRPC playurlgrpc.PlayURLClient
}

func New(c *conf.Config) (s *Dao) {
	s = &Dao{
		c: c,
	}
	var err error
	if s.playurlGRPC, err = playurlgrpc.NewClient(c.PlayurlGRPC); err != nil {
		panic(err)
	}
	return s
}

func (d *Dao) PlayOnline(c context.Context, aidm map[int64]int64) (map[int64]*playurlgrpc.PlayOnlineReply, error) {
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := map[int64]*playurlgrpc.PlayOnlineReply{}
	for k, v := range aidm {
		aid := k
		cid := v
		g.Go(func(ctx context.Context) error {
			req := &playurlgrpc.PlayOnlineReq{
				Aid:      aid,
				Cid:      cid,
				Business: playurlgrpc.OnlineBusiness_OnlineUGC,
			}
			reply, err := d.playurlGRPC.PlayOnline(ctx, req)
			if err != nil {
				return err
			}
			mu.Lock()
			res[aid] = reply
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
