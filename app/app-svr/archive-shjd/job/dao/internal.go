package dao

import (
	"context"
	"database/sql"

	achmdl "go-gateway/app/app-svr/archive/service/api"
)

var (
	_internalSQL = "SELECT `id`,`aid`,`attribute` FROM `archive_internal` WHERE `aid` = ?"
)

func (d *Dao) RawInternal(c context.Context, aid int64) (*achmdl.ArcInternal, error) {
	row := d.db.QueryRow(c, _internalSQL, aid)
	ry := &achmdl.ArcInternal{}
	if err := row.Scan(&ry.ID, &ry.Aid, &ry.Attribute); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ry, nil
}
