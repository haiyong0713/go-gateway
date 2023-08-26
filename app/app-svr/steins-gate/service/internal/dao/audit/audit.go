package audit

import (
	"context"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/database/sql"

	"github.com/pkg/errors"
)

const (
	_graphAuditUpStateSQL = "UPDATE graph_audit SET state=? WHERE id=?"
	//nolint:gosec
	_graphAuditPassSQL       = "UPDATE graph_audit SET result_gid=?, state=? WHERE id=?"
	_graphResultUpdateSQL    = "UPDATE graph SET state=? WHERE id=?"
	_graphAuditByIDSQL       = "SELECT id,aid,regional_vars,global_vars,state,script,result_gid,skin_id FROM graph_audit WHERE id=?"
	_graphAuditByAidSQL      = "SELECT id,aid,regional_vars,global_vars,state,script,result_gid,skin_id FROM graph_audit WHERE aid=? ORDER BY id DESC LIMIT 1"
	_noDelGraphAuditByAidSQL = "SELECT id,aid,regional_vars,global_vars,state,script,result_gid,skin_id FROM graph_audit WHERE aid=? AND state IN (1,-20) ORDER BY id DESC LIMIT 1"
)

// GraphAuditByID def.
func (d *Dao) GraphAuditByID(c context.Context, graphID int64) (a *model.GraphAuditDB, err error) {
	a = &model.GraphAuditDB{}
	if err = d.db.QueryRow(c, _graphAuditByIDSQL, graphID).Scan(&a.Id, &a.Aid, &a.RegionalVars, &a.GlobalVars, &a.State, &a.Script, &a.ResultGID, &a.SkinId); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "graphID %d", graphID)
		}
		a = nil
		return
	}
	return
}

// GraphAuditByAid .
func (d *Dao) GraphAuditByAid(c context.Context, aid int64) (a *model.GraphAuditDB, err error) {
	a = &model.GraphAuditDB{}
	if err = d.db.QueryRow(c, _graphAuditByAidSQL, aid).Scan(&a.Id, &a.Aid, &a.RegionalVars, &a.GlobalVars, &a.State, &a.Script, &a.ResultGID, &a.SkinId); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "aid %d", aid)
		}
		a = nil
		return
	}
	return
}

// GraphAudit def.
func (d *Dao) GraphAudit(c context.Context, auditGID, resultGID int64, state int) (err error) {
	var tx *sql.Tx
	if tx, err = d.db.Begin(c); err != nil {
		err = errors.Wrapf(err, "auditGID %d", auditGID)
		return
	}
	if _, err = tx.Exec(_graphResultUpdateSQL, state, resultGID); err != nil {
		err = errors.Wrapf(err, "resultGID %d auditGID %d state %d", resultGID, auditGID, state)
		//nolint:errcheck
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(_graphAuditUpStateSQL, state, auditGID); err != nil {
		err = errors.Wrapf(err, "resultGID %d auditGID %d state %d", resultGID, auditGID, state)
		//nolint:errcheck
		tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		err = errors.Wrapf(err, "resultGID %d auditGID %d state %d", resultGID, auditGID, state)
		return
	}
	return
}

// GraphAuditRepulse 将审核表状态更新为拒绝
func (d *Dao) GraphAuditRepulse(c context.Context, auditGID int64) (err error) {
	if _, err = d.db.Exec(c, _graphAuditUpStateSQL, model.GraphStateRepulse, auditGID); err != nil {
		err = errors.Wrapf(err, "auditGID %d ", auditGID)
	}
	return
}

// GraphAuditPass 将审核表状态更新为通过，并且更新resultGID字段
func (d *Dao) GraphAuditPass(c context.Context, resultGID, auditGID int64) (err error) {
	if _, err = d.db.Exec(c, _graphAuditPassSQL, resultGID, model.GraphStatePass, auditGID); err != nil {
		err = errors.Wrapf(err, "auditGID %d resultGID %d", auditGID, resultGID)
	}
	return
}

// NoDelGraphAuditByAid .
func (d *Dao) NoDelGraphAuditByAid(c context.Context, aid int64) (a *model.GraphAuditDB, err error) {
	a = &model.GraphAuditDB{}
	if err = d.db.QueryRow(c, _noDelGraphAuditByAidSQL, aid).Scan(&a.Id, &a.Aid, &a.RegionalVars, &a.GlobalVars, &a.State, &a.Script, &a.ResultGID, &a.SkinId); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "aid %d", aid)
		}
		a = nil
		return
	}
	return

}
