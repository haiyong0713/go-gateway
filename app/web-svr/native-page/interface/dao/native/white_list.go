package native

import (
	"context"
	"database/sql"

	"go-common/library/log"
	"go-gateway/app/web-svr/native-page/interface/model/white_list"
)

const (
	_stateValid      = 1
	_whiteListMidSQL = "select `id`,`mid`,`creator`,`creator_uid`,`modifier`,`modifier_uid`,`from_type`,`state`,`ctime`,`mtime` from `white_list` where `mid`=? and `state`=? limit 1"
	_sddWhiteSQL     = "INSERT INTO `white_list` (`mid`,`creator`,`creator_uid`,`modifier`,`modifier_uid`,`state`,`from_type`) VALUES (?,?,?,?,?,?,?)"
)

// RawTagIDSearch .
func (d *Dao) RawWhiteListByMid(c context.Context, mid int64) (*white_list.WhiteList, error) {
	row := d.db.QueryRow(c, _whiteListMidSQL, mid, _stateValid)
	t := &white_list.WhiteList{}
	err := row.Scan(&t.ID, &t.Mid, &t.Creator, &t.CreatorUID, &t.Modifier, &t.ModifierUID, &t.FromType, &t.State, &t.Ctime, &t.Mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return nil, err
	}
	return t, nil
}

func (d *Dao) WhiteSave(c context.Context, p *white_list.WhiteList) (int64, error) {
	res, err := d.db.Exec(c, _sddWhiteSQL, p.Mid, p.Creator, p.CreatorUID, p.Modifier, p.ModifierUID, p.State, p.FromType)
	if err != nil {
		log.Error("WhiteSave arg:%v error(%v)", p, err)
		return 0, err
	}
	return res.LastInsertId()
}
