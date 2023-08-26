package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_addUserCostSQL    = "INSERT INTO act_vogue_user_cost (mid,cost,goods) values(?,?,?)"
	_updateUserCostSQL = "UPDATE act_vogue_user_cost SET cost=? WHERE id=?"
	_countUserCostSQL  = "SELECT IFNULL(SUM(cost),0) FROM act_vogue_user_cost WHERE mid=?"
	_userInviteListSQL = "SELECT score, ctime FROM act_vogue_user_invite WHERE uid=?"
)

func (d *Dao) InsertUserCost(c context.Context, mid, cost, goods int64) (id int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _addUserCostSQL, mid, cost, goods); err != nil {
		log.Error("d.db.Exec(%s,%d,%d,%d)", _addUserCostSQL, mid, cost, goods)
		return
	}
	return res.LastInsertId()
}

func (d *Dao) UpdateUserCost(c context.Context, id, cost int64) (affect int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _updateUserCostSQL, cost, id); err != nil {
		log.Error("d.db.Exec(%s,%d,%d)", _updateUserCostSQL, cost, id)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) CountUserCost(c context.Context, mid int64) (res int64, err error) {
	row := d.db.QueryRow(c, _countUserCostSQL, mid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Error("CountUserCost(%d) error(%v)", mid, err)
	}
	return
}

func (d *Dao) UserInviteList(c context.Context, mid int64) (res []*model.InviteListItem, err error) {
	res = make([]*model.InviteListItem, 0, 0)
	rows, err := d.db.Query(c, _userInviteListSQL, mid)
	if err != nil {
		log.Error("dmReader.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.InviteListItem{}
		if err = rows.Scan(&r.Score, &r.Ctime); err != nil {
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
