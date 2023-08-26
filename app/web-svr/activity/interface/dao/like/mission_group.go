package like

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/xstr"
	l "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_likeMissionBuffSQL  = "SELECT id FROM like_mission_group WHERE sid = ? AND mid = ?"
	_likeMissionAddSQL   = "INSERT INTO like_mission_group (`sid`,`mid`,`state`) VALUES (?,?,?)"
	_likeMissionGroupSQL = "SELECT id,sid,mid,state,ctime,mtime FROM like_mission_group WHERE id IN (%s)"
	_likeMissionAwardSQL = "SELECT award FROM act_like_user_achievements WHERE id = ?"
	// MissionStateInit the init state
	MissionStateInit = 0
)

// RawLikeMissionBuff get mid has .
func (d *Dao) RawLikeMissionBuff(c context.Context, sid, mid int64) (ID int64, err error) {
	res := &l.MissionGroup{}
	row := d.db.QueryRow(c, _likeMissionBuffSQL, sid, mid)
	if err = row.Scan(&res.ID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawLikeMissionBuff:QueryRow")
			return
		}
	}
	ID = res.ID
	return
}

// RawActUserAward .
func (d *Dao) RawActUserAward(c context.Context, id int64) (award int64, err error) {
	row := d.db.QueryRow(c, _likeMissionAwardSQL, id)
	if err = row.Scan(&award); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawActUserAward:QueryRow")
			return
		}
	}
	return
}

// MissionGroupAdd add like_mission_group data .
func (d *Dao) MissionGroupAdd(c context.Context, group *l.MissionGroup) (misID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _likeMissionAddSQL, group.Sid, group.Mid, group.State); err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s)", _likeMissionAddSQL)
		return
	}
	return res.LastInsertId()
}

// RawMissionGroupItems get mission_group item by ids.
func (d *Dao) RawMissionGroupItems(c context.Context, lids []int64) (res map[int64]*l.MissionGroup, err error) {
	res = make(map[int64]*l.MissionGroup, len(lids))
	rows, err := d.db.Query(c, fmt.Sprintf(_likeMissionGroupSQL, xstr.JoinInts(lids)))
	if err != nil {
		err = errors.Wrapf(err, "d.db.Query(%s)", _likeMissionGroupSQL)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := &l.MissionGroup{}
		if err = rows.Scan(&n.ID, &n.Sid, &n.Mid, &n.State, &n.Ctime, &n.Mtime); err != nil {
			err = errors.Wrapf(err, "d.db.Scan(%s)", _likeMissionGroupSQL)
			return
		}
		res[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawMissionGroupItem:rows.Err()")
	}
	return
}
