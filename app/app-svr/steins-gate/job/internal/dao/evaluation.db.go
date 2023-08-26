package dao

import (
	"context"

	"github.com/pkg/errors"
)

const (
	_updateEvaluationSQL = "INSERT INTO arc_evaluation(aid,evaluation) VALUE(?,?) ON DUPLICATE KEY UPDATE evaluation=VALUES(evaluation)"
)

// addEvalDB adds a new evaluation
func (d *Dao) addEvalDB(c context.Context, aid, evaluation int64) (err error) {
	if _, err = d.db.Exec(c, _updateEvaluationSQL, aid, evaluation); err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s) aid = %d", _updateEvaluationSQL, aid)
	}
	return

}
