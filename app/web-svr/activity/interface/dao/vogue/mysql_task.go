package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_taskSQL           = "SELECT id,uid,goods,goods_state,goods_address,mtime FROM act_vogue_user_task WHERE uid=?"
	_taskListSQL       = "SELECT id,uid FROM act_vogue_user_task WHERE id > ? ORDER BY id LIMIT 100"
	_insertTask        = "INSERT IGNORE INTO act_vogue_user_task (uid,goods,goods_state) VALUE (?,?,1)"
	_updateTask        = "UPDATE act_vogue_user_task SET goods_state=? WHERE uid=? AND goods_state=1"
	_updateTaskAddress = "UPDATE act_vogue_user_task SET goods_state=?,goods_address=? WHERE uid=?"
	_prizeSQL          = "SELECT id,uid,goods,goods_state,goods_address,mtime FROM act_vogue_user_task WHERE goods_state >= 3 ORDER BY mtime DESC LIMIT 30"
)

func (d *Dao) RawTask(c context.Context, uid int64) (res *model.Task, err error) {
	res = new(model.Task)
	row := d.db.QueryRow(c, _taskSQL, uid)
	if err = row.Scan(&res.Id, &res.Uid, &res.Goods, &res.GoodsState, &res.GoodsAddress, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawTask(%d) error(%v)", uid, err)
	}
	return
}

func (d *Dao) InsertTask(c context.Context, mid, goods int64) (affect int64, err error) {
	res, err := d.db.Exec(c, _insertTask, mid, goods)
	if err != nil {
		log.Error("d.InsertTask(%v,%v) error(%v)", mid, goods, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTask(c context.Context, mid int64, state int) (affect int64, err error) {
	res, err := d.db.Exec(c, _updateTask, state, mid)
	if err != nil {
		log.Error("d.UpdateTask(%v,%v) error(%v)", mid, state, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTaskAddress(c context.Context, mid int64, state int, address int64) (affect int64, err error) {
	res, err := d.db.Exec(c, _updateTaskAddress, state, address, mid)
	if err != nil {
		log.Error("d.UpdateTask(%v,%v) error(%v)", mid, state, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) RawPrizeList(c context.Context) (res []*model.Task, err error) {
	res = make([]*model.Task, 0, 0)
	rows, err := d.db.Query(c, _prizeSQL)
	if err != nil {
		log.Error("dmReader.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.Task{}
		if err = rows.Scan(&r.Id, &r.Uid, &r.Goods, &r.GoodsState, &r.GoodsAddress, &r.Mtime); err != nil {
			log.Error("row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) TaskMid(c context.Context, id int64) (res []int64, max int64, err error) {
	res = make([]int64, 0, 20)
	rows, err := d.db.Query(c, _taskListSQL, id)
	if err != nil {
		log.Error("dmReader.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, uid int64
		if err = rows.Scan(&id, &uid); err != nil {
			log.Error("row.Scan() error(%v)", err)
			return
		}
		res = append(res, uid)
		if id > max {
			max = id
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}
