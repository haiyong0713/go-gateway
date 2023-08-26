package result

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
)

const (
	_allSeasonIds = "SELECT season_id FROM season"
	_snAids       = "SELECT season_id,aid FROM season_episode where season_id=?"
	_snArcs       = "SELECT aid FROM season_episode WHERE season_id=?"
)

// AllSeasonIds get all seasonIds
func (d *Dao) AllSeasonIds(c context.Context) (sids []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _allSeasonIds); err != nil {
		log.Error("AllSeasonIds query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sid int64
		if err = rows.Scan(&sid); err != nil {
			log.Error("AllSeasonIds row.Scan() error(%v)", err)
			return
		}
		sids = append(sids, sid)
	}
	if err = rows.Err(); err != nil {
		log.Error("AllSeasonIds rows.Err() error(%v)", err)
	}
	return
}

// SnAids def.
func (d *Dao) SnAids(c context.Context, sid int64) (aids []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _snAids, sid); err != nil {
		log.Error("SnAids error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			seasonID, aid int64
		)
		if err = rows.Scan(&seasonID, &aid); err != nil {
			log.Error("SnAids:row.Scan() error(%v)", err)
			return
		}
		aids = append(aids, aid)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

// SnArcs picks the aids of one season
func (d *Dao) SnArcs(c context.Context, sid int64) (res map[int64]struct{}, aids []int64, err error) {
	var rows *xsql.Rows
	res = make(map[int64]struct{})
	if rows, err = d.db.Query(c, _snArcs, sid); err != nil {
		log.Error("SnArcs error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var aid int64
		if err = rows.Scan(&aid); err != nil {
			log.Error("SnArcs:row.Scan() error(%v)", err)
			return
		}
		res[aid] = struct{}{}
		aids = append(aids, aid)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}
