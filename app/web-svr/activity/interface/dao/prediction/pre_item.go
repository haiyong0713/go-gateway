package prediction

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

const (
	_preItemSQL     = "select `id`,`sid`,`pid`,`desc`,`image`,`state`,`ctime`,`mtime` from `prediction_item` where id in (%s) and state = 1"
	_preItemListSQL = "select `id`,`state` from `prediction_item` where id > ? and pid = ? order by id asc limit 1000"
)

// RawPredItems .
func (d *Dao) RawPredItems(c context.Context, ids []int64) (list map[int64]*premdl.PredictionItem, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_preItemSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("RawPredItems:d.db.Query(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	list = make(map[int64]*premdl.PredictionItem, len(ids))
	for rows.Next() {
		n := &premdl.PredictionItem{}
		if err = rows.Scan(&n.ID, &n.Sid, &n.Pid, &n.Desc, &n.Image, &n.State, &n.Ctime, &n.Mtime); err != nil {
			log.Error("RawPredItems:rows.Scan() error(%v)", err)
			return
		}
		list[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("RawPredItems:rows.Err() error(%v)", err)
	}
	return
}

// ItemListSet .
func (d *Dao) ItemListSet(c context.Context, id, pid int64) (res []*premdl.PredictionItem, err error) {
	rows, err := d.db.Query(c, _preItemListSQL, id, pid)
	if err != nil {
		log.Error("ItemListSet:d.db.Query(%d,%d) error(%v)", id, pid, err)
		return
	}
	defer rows.Close()
	res = make([]*premdl.PredictionItem, 0, 1000)
	for rows.Next() {
		n := &premdl.PredictionItem{}
		if err = rows.Scan(&n.ID, &n.State); err != nil {
			log.Error("ItemListSet:rows.Scan() error(%v)", err)
			return
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("ItemListSet:rows.Err() error(%v)", err)
	}
	return
}
