package rewards

import (
	"context"
	xsql "database/sql"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"strings"
)

/*
CREATE TABLE `rewards_cdkey_v2` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
`mid` bigint(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
`award_id` bigint(11) NOT NULL DEFAULT '0' COMMENT 'cdkey名称',
`cdkey_content` varchar(5000) NOT NULL DEFAULT '0' COMMENT 'cdkey内容',
`unique_id` varchar(50) NOT NULL DEFAULT '0' COMMENT '幂等ID',
`is_used` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未使用 1 已使用',
`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
`activity_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '活动id',
PRIMARY KEY (`id`),
UNIQUE KEY `ix_mid_name` (`mid`,`unique_id`),
KEY `ix_aid_mid` (`award_id`,`mid`),
KEY `ix_mtime` (`mtime`)
) ENGINE=InnoDB AUTO_INCREMENT=122 DEFAULT CHARSET=utf8 COMMENT='cdkey发放表'
*/

const sql4GetCdKeyList = `
SELECT award_id, 
       cdkey_content, 
       mtime 
FROM   rewards_cdkey_v2
WHERE  id = ? 
AND mid = ? 
LIMIT 1`

const sql4GetCdKeyListByActivityId = `
SELECT award_id, 
       cdkey_content, 
       mtime 
FROM   rewards_cdkey_v2
WHERE  activity_id = ? 
AND mid = ? 
LIMIT 1`

const (
	sql4UserUpdateCdKey = `
UPDATE rewards_cdkey_v2 FORCE INDEX(ix_aid_mid)
SET mid = ?, activity_id = ?, unique_id = ?, is_used = 1
WHERE award_id = ?
	AND mid = 0
	AND activity_id = 0
	AND is_used = 0
LIMIT 1;
`

	sql4GetCdKeyIdByMidAndUniqID = `
SELECT /*master*/ id
FROM rewards_cdkey_v2
WHERE mid = ?
	AND unique_id = ?
LIMIT 1;
`
)

func (d *Dao) SendCdKey(ctx context.Context, mid, activityId, awardId int64, uniqueId string) (cdKeyId int64, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "SendCdKey error(%v)", err)
		}
	}()

	var res xsql.Result
	var rf int64
	res, err = d.db.Exec(ctx, sql4UserUpdateCdKey, mid, activityId, uniqueId, awardId)
	if err != nil {
		//主键冲突证明曾经发放过, 直接获取ID并返回
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = nil
			goto GETCODE
		}

		return
	}
	rf, err = res.RowsAffected()
	if err != nil {
		return
	}
	if rf == 0 {
		err = ecode.StockServerNoStockError
		return
	}
GETCODE:
	row := d.db.QueryRow(ctx, sql4GetCdKeyIdByMidAndUniqID, mid, uniqueId)
	err = row.Scan(&cdKeyId)

	return
}

func (d *Dao) GetCdKeyById(ctx context.Context, mid, id int64) (res []*model.CdKeyInfo, err error) {
	res = make([]*model.CdKeyInfo, 0)
	row := d.db.QueryRow(ctx, sql4GetCdKeyList, id, mid)
	t := &model.CdKeyInfo{}
	err = row.Scan(&t.CdKeyName, &t.Cdkey, &t.Mtime)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		return
	}
	t.Mid = mid
	res = append(res, t)
	return
}

func (d *Dao) GetCdKeyByActivityId(ctx context.Context, mid, activityId int64) (res []*model.CdKeyInfo, err error) {
	res = make([]*model.CdKeyInfo, 0)
	row := d.db.QueryRow(ctx, sql4GetCdKeyListByActivityId, activityId, mid)
	t := &model.CdKeyInfo{}
	err = row.Scan(&t.CdKeyName, &t.Cdkey, &t.Mtime)
	if err == sql.ErrNoRows {
		err = nil
		return
	}
	if err != nil {
		return
	}
	t.Mid = mid
	res = append(res, t)
	return
}
