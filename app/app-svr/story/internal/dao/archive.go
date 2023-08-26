package dao

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	"github.com/pkg/errors"
)

const _maxAids = 100

func (d *dao) ArcsPlayer(ctx context.Context, aids []int64, from string, need1080plus bool) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	batchArg, _ := arcmid.FromContext(ctx)
	duplicateBatchArg := *batchArg
	duplicateBatchArg.From = from
	playAvs := make([]*arcgrpc.PlayAv, 0, len(aids))
	for _, aid := range aids {
		item := &arcgrpc.PlayAv{
			Aid:         aid,
			HighQnExtra: need1080plus,
		}
		playAvs = append(playAvs, item)
	}
	arg := &arcgrpc.ArcsPlayerRequest{
		PlayAvs:      playAvs,
		BatchPlayArg: &duplicateBatchArg,
	}
	info, err := d.archiveClient.ArcsPlayer(ctx, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.ArcsPlayer
	return
}

func (d *dao) ArcPassedStory(ctx context.Context, in *uparcgrpc.ArcPassedStoryReq) (*uparcgrpc.ArcPassedStoryReply, error) {
	return d.upArcClient.ArcPassedStory(ctx, in)
}

// Archives multi get archives.
func (d *dao) Archives(c context.Context, aids []int64, mid int64, mobiApp, device string) (am map[int64]*arcgrpc.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	var (
		mutexArcs = sync.Mutex{}
	)
	am = map[int64]*arcgrpc.Arc{}
	g := errgroup.WithContext(c)
	if arcsLen := len(aids); arcsLen > 0 {
		for i := 0; i < arcsLen; i += _maxAids {
			var partAids []int64
			if i+_maxAids > arcsLen {
				partAids = aids[i:]
			} else {
				partAids = aids[i : i+_maxAids]
			}
			g.Go(func(ctx context.Context) (err error) {
				var (
					tmpRes *arcgrpc.ArcsReply
					arg    = &arcgrpc.ArcsRequest{
						Aids:    partAids,
						Mid:     mid,
						MobiApp: mobiApp,
						Device:  device,
					}
				)
				if tmpRes, err = d.archiveClient.Arcs(ctx, arg); err != nil {
					log.Error("Archives aids(%v) d.rpcClient.Arcs error(%v)", aids, err)
					return
				}
				if len(tmpRes.GetArcs()) > 0 {
					mutexArcs.Lock()
					for aid, stat := range tmpRes.GetArcs() {
						am[aid] = stat
					}
					mutexArcs.Unlock()
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}
