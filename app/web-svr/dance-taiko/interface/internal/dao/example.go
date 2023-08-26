package dao

import (
	"context"
	"encoding/json"

	"go-common/library/log"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_pickExamplesSQL = "SELECT ts, action FROM dance_example WHERE aid=? AND deleted=0 ORDER BY ts"
	_pickAidsSQL     = "SELECT DISTINCT(aid) FROM dance_example"
)

func (d *dao) AllAids(c context.Context) ([]int64, error) {
	rows, err := d.db.Query(c, _pickAidsSQL)
	if err != nil {
		log.Error("AllAids row.Scan() error(%v)", err)
		return nil, err
	}
	defer rows.Close()

	aids := make([]int64, 0)
	for rows.Next() {
		aid := int64(0)
		if err = rows.Scan(&aid); err != nil {
			log.Error("AllAids row.Scan() error(%v)", err)
			return nil, err
		}
		aids = append(aids, aid)
	}
	return aids, rows.Err()
}

func (d *dao) PickExamples(c context.Context, aid int64) ([]*model.Stat, error) {
	rows, err := d.db.Query(c, _pickExamplesSQL, aid)
	if err != nil {
		log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
		return nil, err
	}
	defer rows.Close()

	stats := make([]*model.Stat, 0)
	for rows.Next() {
		action := ""
		ts := int64(0)
		if err = rows.Scan(&ts, &action); err != nil {
			log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
			return nil, err
		}
		ac := model.StatCore{}
		if err = json.Unmarshal([]byte(action), &ac); err != nil {
			log.Error("PickExamples aid %d row.Scan() error(%v)", aid, err)
			return nil, err
		}
		stats = append(stats, &model.Stat{
			StatCore: ac,
			TS:       ts,
		})
	}

	return stats, rows.Err()
}
