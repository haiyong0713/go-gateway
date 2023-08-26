package ugcpayrank

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/playurl/service/conf"

	ugcpayr "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
)

// Dao is ugcpay dao.
type Dao struct {
	// rpc
	ugcpayrRPC ugcpayr.UGCPayRankClient
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.ugcpayrRPC, err = ugcpayr.NewClient(c.UGCpayRankClient)
	if err != nil {
		panic(fmt.Sprintf("ugcpayrRPC NewClient error(%v)", err))
	}
	return
}

// ArchiveElecStatus
func (d *Dao) ArchiveElecStatus(c context.Context, upmid, aid int64) (bool, error) {
	rly, e := d.ugcpayrRPC.ArchiveElecStatus(c, &ugcpayr.ArchiveElecStatusReq{UPMID: upmid, AVID: aid})
	if e != nil || rly == nil {
		log.Error("d.ugcpayrRPC.ArchiveElecStatus upmid:%d,aid:%d error(%v)", upmid, aid, e)
		return false, e
	}
	return rly.Show, nil
}
