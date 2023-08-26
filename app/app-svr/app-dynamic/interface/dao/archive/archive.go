package archive

import (
	"context"
	"sync"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-dynamic/interface/model"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"
)

func (d *Dao) Archive(c context.Context, aids []int64, mobiApp, device string, mid int64, platform string) (res map[int64]*archivegrpc.Arc, err error) {
	var (
		args    = &archivegrpc.ArcsRequest{Aids: aids, Device: device, MobiApp: mobiApp, Mid: mid, Platform: platform}
		arcsTmp *archivegrpc.ArcsReply
	)
	if arcsTmp, err = d.archiveGRPC.Arcs(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = arcsTmp.GetArcs()
	return
}

func (d *Dao) Pages(ctx context.Context, cids map[int64]int64) (map[int64]*archivegrpc.Page, error) {
	eg := errgroup.WithContext(ctx)
	eg.GOMAXPROCS(20)
	mu := sync.Mutex{}
	ret := make(map[int64]*archivegrpc.Page)

	for c, a := range cids {
		cid := c
		aid := a
		eg.Go(func(ctx context.Context) error {
			resp, err := d.archiveGRPC.Video(ctx, &archivegrpc.VideoRequest{Aid: aid, Cid: cid})
			if err != nil {
				return err
			}
			mu.Lock()
			ret[cid] = resp.Page
			mu.Unlock()
			return nil
		})
	}
	return ret, eg.Wait()
}

// ArcsWithPlayurl get archive player.
func (d *Dao) ArcsPlayer(c context.Context, aids []*archivegrpc.PlayAv, showPgcPlayurl bool, from string) (map[int64]*archivegrpc.ArcPlayer, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*archivegrpc.ArcPlayer)
	batchArg, _ := arcmid.FromContext(c)
	tmpBatchArg := &archivegrpc.BatchPlayArg{}
	if batchArg != nil {
		*tmpBatchArg = *batchArg
		tmpBatchArg.ShowPgcPlayurl = showPgcPlayurl
		tmpBatchArg.From = from
	}
	for i := 0; i < len(aids); i += max50 {
		var partAids []*archivegrpc.PlayAv
		if i+max50 > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &archivegrpc.ArcsPlayerRequest{
				PlayAvs:      partAids,
				BatchPlayArg: tmpBatchArg,
			}
			archives, err := d.archiveGRPC.ArcsPlayer(ctx, arg)
			if err != nil {
				log.Error("ArcsPlayer partAids(%+v) err(%v)", partAids, err)
				return err
			}
			mu.Lock()
			for aid, arc := range archives.GetArcsPlayer() {
				if arc == nil {
					continue
				}
				if !arc.Arc.IsNormal() && !model.IsPremiereBefore(arc.Arc) {
					continue
				}
				res[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
