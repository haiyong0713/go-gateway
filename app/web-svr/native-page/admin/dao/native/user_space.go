package native

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_natUserSpaceTable    = "native_user_space"
	_updateUSpaceStateSQL = "UPDATE `native_user_space` SET `state`=? WHERE `id`=? and `page_id`=? and `state`=?"
	_resetUSpaceSQL       = "UPDATE `native_user_space` SET `page_id`=0,`display_space`=0,`state`=? WHERE `mid`=? and `page_id`=? and `state`=?"
)

func (d *Dao) UpdateUserSpaceState(c context.Context, id, pid int64, newState, oldState string) error {
	if err := d.DB.Exec(_updateUSpaceStateSQL, newState, id, pid, oldState).Error; err != nil {
		log.Error("Fail to update native_user_space state, newState=%+v id=%+v pid=%+v oldState=%+v error=%+v", newState, id, pid, oldState, err)
		return err
	}
	return nil
}

func (d *Dao) ResetUserSpace(c context.Context, mid, pageID int64, newState, oldState string) error {
	if err := d.DB.Exec(_resetUSpaceSQL, newState, mid, pageID, oldState).Error; err != nil {
		log.Error("Fail to reset native_user_space, mid=%+v pageID=+%v error=%+v", mid, pageID, err)
		return err
	}
	return nil
}

func (d *Dao) UserSpaceByMid(c context.Context, mid int64) (*api.NativeUserSpace, error) {
	res := &api.NativeUserSpace{}
	if err := d.DB.Table(_natUserSpaceTable).Where("mid=?", mid).First(&res).Error; err != nil {
		log.Error("Fail to get userSpace, mid=%+v error=%+v", mid, err)
		return nil, err
	}
	return res, nil
}
