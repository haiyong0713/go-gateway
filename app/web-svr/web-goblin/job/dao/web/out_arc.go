package web

import (
	"context"
	"time"

	"go-gateway/app/web-svr/web-goblin/job/model/web"

	"github.com/pkg/errors"
)

const _addArcSQL = "INSERT INTO out_arc(aid,snap_view,type) VALUES (?,?,?) ON DUPLICATE KEY UPDATE snap_view=?"

func (d *Dao) AddArc(ctx context.Context, aid, click, typ int64) error {
	if _, err := d.db.Exec(ctx, _addArcSQL, aid, click, typ, click); err != nil {
		return errors.Wrap(err, "AddArc exec")
	}
	return nil
}

const _delArc = "UPDATE out_arc SET is_deleted=1 WHERE aid=?"

func (d *Dao) DelArc(ctx context.Context, aid int64) error {
	if _, err := d.db.Exec(ctx, _delArc, aid); err != nil {
		return errors.Wrap(err, "DelArc exec")
	}
	return nil
}

const _outArcByMtimeSQL = "SELECT id,aid,snap_view,is_deleted FROM out_arc WHERE mtime>=? AND mtime<?"

func (d *Dao) OutArcByMtime(ctx context.Context, from, to time.Time) ([]*web.OutArc, error) {
	rows, err := d.db.Query(ctx, _outArcByMtimeSQL, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*web.OutArc
	for rows.Next() {
		r := new(web.OutArc)
		if err = rows.Scan(&r.ID, &r.Aid, &r.SnapView, &r.IsDeleted); err != nil {
			return nil, errors.Wrap(err, "OutArcByMtime scan")
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "OutArcByMtime rows Err")
	}
	return list, nil
}
