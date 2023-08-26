package dynamic

import (
	"context"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/log"
	dynmal "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"
)

func (d *Dao) ItemHasLikeRecent(c context.Context, mids []int64, business map[string][]*dynmal.LikeBusiItem) (*thumgrpc.ItemHasLikeRecentReply, error) {
	var busiParams = map[string]*thumgrpc.ItemHasLikeRecentReq_Business{}
	for busi, item := range business {
		busiTmp, ok := busiParams[busi]
		if !ok {
			busiTmp = &thumgrpc.ItemHasLikeRecentReq_Business{}
			busiParams[busi] = busiTmp
		}
		for _, v := range item {
			recTmp := &thumgrpc.ItemHasLikeRecentReq_Record{
				OriginID:  v.OrigID,
				MessageID: v.MsgID,
			}
			busiTmp.Records = append(busiTmp.Records, recTmp)
		}
	}
	in := &thumgrpc.ItemHasLikeRecentReq{
		Mids:       mids,
		Businesses: busiParams,
	}
	like, err := d.thumGRPC.ItemHasLikeRecent(c, in)
	if err != nil {
		log.Errorc(c, "Dao.ItemHasLikeRecent() failed. error(%+v) ", err)
		return nil, err
	}
	return like, nil
}

func (d *Dao) MultiStats(c context.Context, mid int64, business map[string][]*dynmal.LikeBusiItem) (*thumgrpc.MultiStatsReply, error) {
	var busiParams = map[string]*thumgrpc.MultiStatsReq_Business{}
	for busi, item := range business {
		busiTmp, ok := busiParams[busi]
		if !ok {
			busiTmp = &thumgrpc.MultiStatsReq_Business{}
			busiParams[busi] = busiTmp
		}
		for _, v := range item {
			recTmp := &thumgrpc.MultiStatsReq_Record{
				OriginID:  v.OrigID,
				MessageID: v.MsgID,
			}
			busiTmp.Records = append(busiTmp.Records, recTmp)
		}
	}
	in := &thumgrpc.MultiStatsReq{
		Mid:      mid,
		Business: busiParams,
	}
	likeStats, err := d.thumGRPC.MultiStats(c, in)
	if err != nil {
		log.Errorc(c, "Dao.MultiStats() failed. error(%+v)", err)
		return nil, err
	}
	return likeStats, nil
}
