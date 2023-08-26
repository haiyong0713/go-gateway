package service

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"

	"github.com/pkg/errors"
)

func (s *Service) natInfoFromForeign(c context.Context, tids []int64) (map[int64]*natpagegrpc.NativePage, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*natpagegrpc.NativePage)
	for i := 0; i < len(tids); i += max50 {
		var partTids []int64
		if i+max50 > len(tids) {
			partTids = tids[i:]
		} else {
			partTids = tids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			nfs, err := s.natInfoFromForeignSlice(ctx, partTids, 1)
			if err != nil {
				return err
			}
			mu.Lock()
			for tid, nf := range nfs {
				res[tid] = nf
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("ChannelInfos tids(%+v) eg.wait(%+v)", tids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) natInfoFromForeignSlice(c context.Context, tids []int64, pageType int64) (map[int64]*natpagegrpc.NativePage, error) {
	var (
		args   = &natpagegrpc.NatInfoFromForeignReq{Fids: tids, PageType: pageType}
		resTmp *natpagegrpc.NatInfoFromForeignReply
		err    error
	)
	if resTmp, err = s.natPageGrpcClient.NatInfoFromForeign(c, args); err != nil {
		return nil, err
	}
	// 木有getList方法
	if resTmp == nil {
		return nil, errors.New("s.natPageGrpcClient.NatInfoFromForeign get nil result")
	}
	return resTmp.List, nil
}
