package thumbup

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"github.com/pkg/errors"
)

type Dao struct {
	c        *conf.Config
	thumGRPC thumgrpc.ThumbupClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.thumGRPC, err = thumgrpc.NewClient(c.ThumGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) MultiStats(c context.Context, mid int64, business map[string][]*mdlv2.ThumbsRecord) (*thumgrpc.MultiStatsReply, error) {
	var busiParams = map[string]*thumgrpc.MultiStatsReq_Business{}
	for busType, item := range business {
		busiTmp, ok := busiParams[busType]
		if !ok {
			busiTmp = &thumgrpc.MultiStatsReq_Business{}
			busiParams[busType] = busiTmp
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
		return nil, errors.WithStack(err)
	}
	return likeStats, nil
}
