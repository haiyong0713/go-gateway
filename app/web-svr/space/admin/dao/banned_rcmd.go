package dao

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
	"strings"
)

const (
	_bannedRcmdTable   = "up_rcmd_black_list"
	_bannerRcmdReplace = "insert into up_rcmd_black_list (mid, is_deleted) values %s ON DUPLICATE KEY UPDATE is_deleted = 0"
)

// GetBannedRcmdMids .
func (d *Dao) GetBannedRcmdMids(_ context.Context, mid int64, ps int64, pn int64) (blackListItems []*model.UpRcmdBlackListItem, total int64, err error) {
	blackListItems = make([]*model.UpRcmdBlackListItem, 0)

	query := d.DB.Table(_bannedRcmdTable).Where("is_deleted=?", 0)
	if mid != 0 {
		query = query.Where("mid=?", mid)
	}
	err = query.Order("mid desc").
		Offset(ps * (pn - 1)).Limit(ps).
		Find(&blackListItems).Error

	if err != nil {
		log.Error("[up_rcmd_banned]GetBannedRcmdMids find err: %s", err.Error())
		return blackListItems, 0, err
	}

	err = query.Count(&total).Error
	if err != nil {
		log.Error("[up_rcmd_banned]GetBannedRcmdMids total err: %s", err.Error())
	}
	return blackListItems, total, err
}

// GetBannedRcmdMids .
func (d *Dao) CreateBannedRcmdMids(_ context.Context, bannedMids []int64) (err error) {
	recordStr := "(%d, 0)"
	values := make([]string, 0)
	for _, mid := range bannedMids {
		values = append(values, fmt.Sprintf(recordStr, mid))
	}
	err = d.DB.Exec(fmt.Sprintf(_bannerRcmdReplace, strings.Join(values, ","))).Error
	if err != nil {
		log.Error("[up_rcmd_banned]CreateBannedRcmdMids err: %s", err.Error())
	}
	return err
}

// GetBannedRcmdMids .
func (d *Dao) DeleteBannedRcmdMid(_ context.Context, mid int64) (err error) {
	attrs := map[string]interface{}{
		"is_deleted": 1,
	}
	err = d.DB.Table(_bannedRcmdTable).
		Where("mid = ?", mid).
		Update(attrs).Error
	if err != nil {
		log.Error("[up_rcmd_banned]DeleteBannedRcmdMid err: %s", err.Error())
	}
	return err
}
