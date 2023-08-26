package dao

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/web-svr/dance-taiko/job/model"
)

const (
	_selectKeyFrames = "SELECT `key_frames` FROM `dance_key_frames` WHERE `aid`=? AND `cid`=? AND `is_deleted`=0"
	_pickExamplesSQL = "SELECT ts, action FROM dance_examples_%02d WHERE aid=? AND cid=? AND is_deleted=0 AND ts>=? AND ts<=? ORDER BY ts"
)

// 单位毫秒的关键帧
func (d *Dao) RawKeyFrames(c context.Context, aid, cid int64) ([]int64, error) {
	row := d.db.QueryRow(c, _selectKeyFrames, aid, cid)
	res := ""
	if err := row.Scan(&res); err != nil {
		return []int64{}, errors.Wrapf(err, "RawKeyFrames cid(%d)", cid)
	}
	if res == "" {
		return []int64{}, nil
	}
	return xstr.SplitInts(res)
}

// 获取一段区间内的example数据
func (d *Dao) PickExamples(c context.Context, aid, cid, start, end int64) ([]model.Example, error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_pickExamplesSQL, aid%100), aid, cid, start, end)
	if err != nil {
		log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.Example, 0)
	for rows.Next() {
		var ts int64
		var action string
		if err = rows.Scan(&ts, &action); err != nil {
			log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
			return nil, err
		}
		stats = append(stats, model.Example{
			Ts:     ts,
			Action: action,
		})
	}
	if err := rows.Err(); err != nil {
		log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
		return nil, err
	}
	return stats, rows.Err()
}
