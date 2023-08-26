package dao

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/model/common"
	"time"
)

const (
	_refreshBWListItem = `SELECT 
		l.oid, l.area_id, l.is_deleted, s.token, s.status, s.is_online, s.default_value
	FROM
		black_white_list AS l
			LEFT JOIN
		black_white_scene AS s ON l.scene_id = s.id
	WHERE
		s.mtime >= ? or l.mtime >= ? Limit ? offset ?`
)

const PageSize = 200

// 从数据库刷新上1分钟的变更内容
func (d *Dao) GetModifiedBWListItemFromDB(ctx context.Context, pageNum int) (list []*common.BWListItem, err error) {
	m, _ := time.ParseDuration("-31s")
	startTime := time.Now().Add(m).Format("2006-01-02 15:04:05")

	var rows *sql.Rows
	rows, err = d.managerDB.Query(
		ctx,
		_refreshBWListItem,
		startTime,
		startTime,
		PageSize,
		pageNum*PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list = make([]*common.BWListItem, 0)

	for rows.Next() {
		row := new(common.BWListItem)
		if err = rows.Scan(&row.Oid, &row.AreaID, &row.IsDeleted, &row.Token, &row.Status, &row.IsOnline, &row.DefaultValue); err != nil {
			return nil, err
		}
		list = append(list, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	log.Error("GetModifiedBWListItemFromDB success len(%d)", len(list))

	return list, nil
}

// redis Dao
func (d *Dao) SetModifiedBWListItemIntoRedis(ctx context.Context, list []*common.BWListItem) (err error) {
	conn := d.redisShow.Get(ctx)
	defer conn.Close()

	var counter int

	//defaultValueKeys := make(map[string]bool)

	for _, ele := range list {
		item := ele

		// 目前仅设置地域配制
		areaKey := fmt.Sprintf("bw_%s_area_%s", item.Token, item.Oid)

		if item.IsOnline == 1 && item.IsDeleted == 0 && item.Status == 0 {
			// 设置key
			if err = conn.Send("SET", areaKey, item.AreaID); err != nil {
				log.Error("SetModifiedBWListItemIntoRedis add fail: %+v, key: %+v, val: %d", err, areaKey, item.AreaID)
				return err
			}
			counter += 1
		} else {
			// 删除key
			if err = conn.Send("DEL", areaKey); err != nil {
				log.Error("SetModifiedBWListItemIntoRedis del fail: %+v, key: %+v", err, areaKey)
				return err
			}
			counter += 1
		}
	}
	if err = conn.Flush(); err != nil {
		return err
	}

	for counter > 0 {
		if _, err = conn.Receive(); err != nil {
			return err
		}
		counter--
	}

	return nil
}
