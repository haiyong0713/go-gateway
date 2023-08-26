package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/esports/job/component"
)

const _stateOk = 1

func (d *Dao) AsyncSendTunnelDatabus(ctx context.Context, platfrom, mid, uniqueID int64) (err error) {
	reqParam := struct {
		Platform int64 `json:"platform"`
		Mid      int64 `json:"mid"`
		State    int64 `json:"state"`
		BizID    int64 `json:"biz_id"`
		UniqueID int64 `json:"unique_id"`
	}{platfrom, mid, _stateOk, d.c.Rule.TunnelBizID, uniqueID}
	if err = d.tunnelPub.Send(ctx, strconv.FormatInt(mid, 10), reqParam); err != nil {
		log.Errorc(ctx, "d.tunnelPub.Send mid(%d) uniqueID(%d) reqParam(%+v) error(%+v)", mid, uniqueID, reqParam, err)
	}
	return
}

func (d *Dao) AsyncSendBGroupDatabus(ctx context.Context, mid, contestID int64) (err error) {
	reqParam := struct {
		Mid       int64  `json:"mid"`
		Source    string `json:"source"`
		Name      string `json:"name"`
		State     int64  `json:"state"`
		Timestamp int64  `json:"timestamp"`
	}{
		mid,
		d.c.TunnelBGroup.Source,
		strconv.FormatInt(contestID, 10),
		_stateOk,
		time.Now().Unix()}
	key := strconv.FormatInt(mid, 10)
	buf, _ := json.Marshal(reqParam)
	if err = retry.WithAttempts(ctx, "job_contest_fav_send_event", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		return component.BGroupMessagePub.Send(ctx, key, buf)
	}); err != nil {
		log.Errorc(ctx, "AsyncSendBGroupDatabus d.BGroupMessagePub.Send mid(%d) contestID(%d) reqParam(%+v) error(%+v)", mid, contestID, reqParam, err)
	}
	return
}
