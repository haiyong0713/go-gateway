package web

import (
	"context"

	"go-gateway/app/web-svr/web-goblin/interface/model/web"
)

const _outArcsSQL = `SELECT id,aid,snap_view FROM out_arc WHERE type=? AND is_deleted = 0 AND id>? ORDER BY id ASC LIMIT 1000`

func (d *Dao) OutArcs(c context.Context, typ int, id int64) ([]*web.OutArc, error) {
	rows, err := d.db.Query(c, _outArcsSQL, typ, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*web.OutArc
	for rows.Next() {
		c := &web.OutArc{}
		if err = rows.Scan(&c.ID, &c.Aid, &c.SnapView); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
