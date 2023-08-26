package dao

import (
	"context"
	"database/sql"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/steins-gate/admin/internal/model"

	"github.com/pkg/errors"
)

const (
	_graphAuditByIDSQL = "SELECT id,aid,regional_vars,global_vars,state FROM graph_audit WHERE id=?"
)

// GraphAuditByID def.
func (d *Dao) GraphAuditByID(c context.Context, graphID int64) (a *model.GraphDB, err error) {
	a = &model.GraphDB{}
	if err = d.db.QueryRow(c, _graphAuditByIDSQL, graphID).Scan(&a.Id, &a.Aid, &a.RegionalVars, &a.GlobalVars, &a.State); err != nil {
		if err == sql.ErrNoRows {
			err = ecode.NothingFound
		}
		err = errors.Wrapf(err, "graphID %d", graphID)
		a = nil
	}
	return

}
