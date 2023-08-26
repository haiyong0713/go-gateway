package native

import (
	"context"
	"database/sql"

	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_userSpaceByMidSQL    = "SELECT `id`,`mid`,`title`,`page_id`,`display_space`,`state` FROM `native_user_space` WHERE `mid`=?"
	_addUserSpaceSQL      = "INSERT INTO `native_user_space` (`mid`,`title`,`page_id`,`display_space`,`state`) VALUES (?,?,?,?,?)"
	_updateUSpaceSQL      = "UPDATE `native_user_space` SET `title`=?,`page_id`=?,`display_space`=?,`state`=? WHERE `id`=? and `page_id`=? and `state`=?"
	_updateUSpaceStateSQL = "UPDATE `native_user_space` SET `state`=? WHERE `id`=? and `page_id`=? and `state`=?"
)

func (d *Dao) RawUserSpaceByMid(c context.Context, mid int64) (*api.NativeUserSpace, error) {
	row := d.db.QueryRow(c, _userSpaceByMidSQL, mid)
	t := &api.NativeUserSpace{}
	err := row.Scan(&t.Id, &t.Mid, &t.Title, &t.PageId, &t.DisplaySpace, &t.State)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return nil, err
	}
	return t, nil
}

func (d *Dao) AddUserSpace(c context.Context, sp *api.NativeUserSpace) (int64, error) {
	res, err := d.db.Exec(c, _addUserSpaceSQL, sp.Mid, sp.Title, sp.PageId, sp.DisplaySpace, sp.State)
	if err != nil {
		log.Errorc(c, "Fail to create native_user_space, space=%+v error=%+v", sp, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) UpdateUserSpace(c context.Context, sp *api.NativeUserSpace, oldPid int64, oldState string) error {
	if _, err := d.db.Exec(c, _updateUSpaceSQL, sp.Title, sp.PageId, sp.DisplaySpace, sp.State, sp.Id, oldPid, oldState); err != nil {
		log.Errorc(c, "Fail to update native_user_space, space=%+v oldPid=%+v oldState=%+v error=%+v", sp, oldPid, oldState, err)
		return err
	}
	return nil
}

func (d *Dao) UpdateUserSpaceState(c context.Context, id, pid int64, newState, oldState string) error {
	if _, err := d.db.Exec(c, _updateUSpaceStateSQL, newState, id, pid, oldState); err != nil {
		log.Errorc(c, "Fail to update native_user_space state, newState=%+v id=%+v pid=%+v oldState=%+v error=%+v", newState, id, pid, oldState, err)
		return err
	}
	return nil
}
