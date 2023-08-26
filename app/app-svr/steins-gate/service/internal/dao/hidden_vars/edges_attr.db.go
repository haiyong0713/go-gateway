package hidden_vars

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

// RawGraphEdgeAttrs .
func (d *Dao) RawGraphEdgeAttrs(c context.Context, graphID int64) (attrsCache *model.EdgeAttrsCache, err error) {
	attrsCache = new(model.EdgeAttrsCache)
	if attrsCache.EdgeAttrs, err = d.graphEdgeList(c, graphID); err != nil {
		err = errors.Wrapf(err, "gid %d", graphID)
		return
	}
	if len(attrsCache.EdgeAttrs) > 0 {
		attrsCache.HasAttrs = true
	}
	return
}

func (d *Dao) graphEdgeList(c context.Context, graphID int64) (edges []*model.EdgeAttr, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _edgesWithAttrSQL, graphID)
	if err != nil {
		log.Error("db.Query GraphEdgeList(%d) error(%v)", graphID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		edge := new(model.EdgeAttr)
		if err = rows.Scan(&edge.ID, &edge.FromNID, &edge.ToNID, &edge.Attribute); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		edges = append(edges, edge)
	}
	err = rows.Err()
	return

}
