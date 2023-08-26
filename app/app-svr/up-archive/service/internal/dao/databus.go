package dao

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/up-archive/service/internal/model"
)

// nolint:bilirailguncheck
func (d *dao) SendBuildCacheMsg(ctx context.Context, mid, nowTs int64) error {
	data := &model.UpArcPubMsg{Mid: mid, Ctime: nowTs}
	err := d.upArcPub.Send(ctx, strconv.FormatInt(mid, 10), data)
	if err != nil {
		log.Warn("发送稿件列表消息失败,data:%+v,error:%+v", data, err)
		return err
	}
	log.Warn("发送稿件列表消息成功,data:%+v", data)
	return nil
}
