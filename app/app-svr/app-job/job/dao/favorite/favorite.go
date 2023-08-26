package favorite

import (
	"context"
	"time"

	"git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const (
	_popurlarFav        = 14
	_oidWeeklySelected  = 1
	_typeWeeklySelected = "weekly_selected"
	_retryTimes         = 3
	FavedUsersPS        = 1000
)

func favStype(sType string) (oid int64) {
	switch sType {
	case _typeWeeklySelected:
		return _oidWeeklySelected
	}
	return
}

func (d *Dao) Subscribers(c context.Context, sType string, cursor int64) (reply *api.SubscribersReply, err error) {
	oid := favStype(sType)
	if oid == 0 {
		err = ecode.RequestErr
		return
	}
	for i := 0; i < _retryTimes; i++ {
		if reply, err = d.favClient.Subscribers(c, &api.SubscribersReq{Type: _popurlarFav, Oid: oid, Cursor: cursor, Size_: FavedUsersPS}); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if err != nil {
		err = errors.Wrapf(err, "favUsers s.fav.Users sType(%s) cursor(%d)", sType, cursor)
	}
	return
}
