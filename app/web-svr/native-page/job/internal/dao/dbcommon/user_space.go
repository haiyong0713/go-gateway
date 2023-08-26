package dbcommon

import (
	"context"

	"go-common/library/log"
)

const (
	_resetUSpaceSQL = "UPDATE `native_user_space` SET `page_id`=0,`display_space`=0,`state`=? WHERE `mid`=? and `page_id`=?"
)

func (d *Dao) ResetUserSpace(c context.Context, mid, pageID int64, newState string) error {
	_, err := d.db.Exec(c, _resetUSpaceSQL, newState, mid, pageID)
	if err != nil {
		log.Error("Fail to reset native_user_space, mid=%+v pageID=%+v newState=%+v error=%+v", mid, pageID, newState, err)
		return err
	}
	return nil
}
