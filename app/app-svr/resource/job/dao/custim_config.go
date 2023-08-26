package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/model/custom"
	"time"
)

// CC: custom config

const (
	_refreshModifiedCCSQL = `select tp, oid, content, url, highlight_content, image, image_big, stime, etime, state from custom_config where mtime between ? AND ? `
)

// 从数据库刷新上1分钟的变更内容
func (d *Dao) GetModifiedCCFromDB(ctx context.Context) (list []*custom.Config, err error) {
	endTime := time.Now()
	m, _ := time.ParseDuration("-1m")
	startTime := endTime.Add(m)

	var rows *sql.Rows
	rows, err = d.resourceDB.Query(ctx, _refreshModifiedCCSQL, startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Error("GetModifiedCCFromDB fail 0: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	list = make([]*custom.Config, 0)

	for rows.Next() {
		row := new(custom.Config)
		if err = rows.Scan(&row.Tp, &row.Oid, &row.Content, &row.Url, &row.HighlightContent, &row.Image, &row.ImageBig, &row.STime, &row.ETime, &row.State); err != nil {
			log.Error("GetModifiedCCFromDB fail 1: %s", err.Error())
			return nil, err
		}
		list = append(list, row)
	}

	if err = rows.Err(); err != nil {
		log.Error("GetModifiedCCFromDB fail 2: %s", err.Error())
		return nil, err
	}

	return list, nil
}

// redis Dao
func (d *Dao) SetModifiedCCIntoRedis(ctx context.Context, list []*custom.Config) (err error) {
	conn := d.redisShow.Get(ctx)
	defer conn.Close()

	var counter int

	var ccJson []byte
	for _, item := range list {
		ccJson, err = json.Marshal(item)
		if err != nil {
			log.Error("SetModifiedCCIntoRedis fail 0 : %+v, data: %+v", err, ccJson)
			return err
		}
		key := fmt.Sprintf("cc_%d", item.Oid)

		if item.State == 1 {
			if err = conn.Send("SET", key, ccJson); err != nil {
				log.Error("SetModifiedCCIntoRedis fail 1: %+v, data: %+v", err, ccJson)
				return err
			}
			if err = conn.Send("EXPIREAT", key, item.ETime.Unix()); err != nil {
				log.Error("SetModifiedCCIntoRedis fail 2: %+v, data: %+v", err, ccJson)
				return err
			}
			counter += 2
		} else if item.State == 0 {
			if err = conn.Send("DEL", key); err != nil {
				log.Error("SetModifiedCCIntoRedis fail 3: %+v, data: %+v", err, ccJson)
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
