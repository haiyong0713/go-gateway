package dao

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
	"go-gateway/app/web-svr/esports/service/tool"

	"github.com/jinzhu/gorm"
)

func (d *dao) RawReplyWall() (list []*model.ReplyWallModel, err error) {
	if err = d.orm.Model(&model.ReplyWallModel{}).Where(_isDeletedFilter, model.IsDeletedFalse).Find(&list).Error; err != nil {
		log.Error("GetReplyWall Error (%v)", err)
		return
	}
	return
}

func (d *dao) ReplyWallUpdateTransaction(ctx context.Context, req *pb.SaveReplyWallModel) (err error) {
	return d.orm.Transaction(func(tx *gorm.DB) (err error) {
		if err = d.BatchDeleteReplyWall(ctx, tx); err != nil {
			log.Errorc(ctx, "[DB][ReplyWallUpdateTransaction][d.BatchDeleteReplyWall][Error], error:(%+v)", err)
			return
		}
		if err = d.BatchAddReplyWall(ctx, tx, req); err != nil {
			log.Errorc(ctx, "[DB][ReplyWallUpdateTransaction][d.BatchAddReplyWall][Error], error:(%+v)", err)
			return
		}
		return
	})
}

func (d *dao) BatchDeleteReplyWall(ctx context.Context, tx *gorm.DB) (err error) {
	err = tx.Model(&model.ReplyWallModel{}).
		Where(_isDeletedFilter, model.IsDeletedFalse).
		Updates(map[string]interface{}{"is_deleted": model.IsDeletedTrue}).Error
	if err != nil {
		log.Errorc(ctx, "[DB][BatchDeleteReplyWall][Error] err:(%+v)", err)
		return
	}
	return
}

const _batchAddReplyWallSql = "insert into es_reply_wall (`contest_id`, `mid`, `reply_details`) values %s"

func (d *dao) BatchAddReplyWall(ctx context.Context, tx *gorm.DB, req *pb.SaveReplyWallModel) (err error) {
	if req.ContestID == 0 {
		return
	}
	var rowStrings []string
	param := make([]interface{}, 0)
	for _, replyInfo := range req.GetReplyList() {
		rowStrings = append(rowStrings, "(?,?,?)")
		// 过滤掉表情符号
		details := tool.RemoveEmojis(replyInfo.ReplyDetails)
		param = append(param, req.ContestID, replyInfo.Mid, details)
	}
	sql := fmt.Sprintf(_batchAddReplyWallSql, strings.Join(rowStrings, ","))
	if err = tx.Model(&model.ReplyWallModel{}).Exec(sql, param...).Error; err != nil {
		log.Errorc(ctx, "[DB][BatchAddReplyWall][Error] err:(%+v)", err)
	}
	return
}
