package prediction

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

const (
	_preSQL     = "select `id`,`sid`,`pid`,`min`,`max`,`name`,`type`,`state`,`ctime`,`mtime` from `prediction` where id in (%s) and state = 1"
	_preListSQL = "select `id`,`state` from `prediction` where id > ? and sid = ? order by id asc limit 1000"
)

// RawPredictions .
func (d *Dao) RawPredictions(c context.Context, ids []int64) (list map[int64]*premdl.Prediction, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_preSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("RawPredictions:d.db.Query(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	list = make(map[int64]*premdl.Prediction, len(ids))
	for rows.Next() {
		n := &premdl.Prediction{}
		if err = rows.Scan(&n.ID, &n.Sid, &n.Pid, &n.Min, &n.Max, &n.Name, &n.Type, &n.State, &n.Ctime, &n.Mtime); err != nil {
			log.Error("RawPredictions:rows.Scan() error(%v)", err)
			return
		}
		list[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("RawPredictions:rows.Err() error(%v)", err)
	}
	return
}

// ListSet .
func (d *Dao) ListSet(c context.Context, id, sid int64) (res []*premdl.Prediction, err error) {
	rows, err := d.db.Query(c, _preListSQL, id, sid)
	if err != nil {
		log.Error("ListSet:d.db.Query(%d,%d) error(%v)", id, sid, err)
		return
	}
	defer rows.Close()
	res = make([]*premdl.Prediction, 0, 1000)
	for rows.Next() {
		n := &premdl.Prediction{}
		if err = rows.Scan(&n.ID, &n.State); err != nil {
			log.Error("ListSet:rows.Scan() error(%v)", err)
			return
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("ListSet:rows.Err() error(%v)", err)
	}
	return
}
