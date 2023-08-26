package dao

import (
	"context"
	"strconv"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/job/internal/model"
)

// ArcView int
func (d *Dao) ArcView(ctx context.Context, aid int64) (res *arcgrpc.SteinsGateViewReply, err error) {
	if res, err = d.arcClient.SteinsGateView(ctx, &arcgrpc.SteinsGateViewRequest{
		Aid: aid,
	}); err != nil {
		log.Error("SteinsGateView Aid %d, Err %v", aid, err)
		return
	}
	return
}

// UpArcFirstCid sends
//
//nolint:bilirailguncheck
func (d *Dao) UpArcFirstCid(ctx context.Context, aid, cid int64) {
	if err := d.steinsGatePub.Send(ctx, strconv.FormatInt(aid, 10), &model.SteinsCid{Aid: aid, Cid: cid, Route: model.SteinsRouteForStickVideo}); err != nil {
		log.Error("SendCid Aid %d, Cid %d Err %v", aid, cid, err)
		d.retryCh <- &model.RetryOp{
			Action:   _retrySendDatabus,
			Value:    aid,
			SubValue: cid,
		}
	} else {
		log.Info("SendCid Aid %d, Cid %d Succ", aid, cid)
	}

}
