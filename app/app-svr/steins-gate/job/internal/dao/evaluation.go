package dao

import (
	"context"

	"go-gateway/app/app-svr/steins-gate/job/internal/model"

	"go-common/library/log"
)

// addEval def
func (d *Dao) addEval(c context.Context, aid, evaluation int64) (err error) {
	if err = d.addEvalDB(c, aid, evaluation); err != nil {
		return
	}
	err = d.addEvalCache(c, aid, evaluation)
	return
}

// AddEval contains retry
func (d *Dao) AddEval(c context.Context, aid, evaluation int64) {
	if err := d.addEval(c, aid, evaluation); err != nil {
		log.Error("AddEvaluation Aid %d, evaluation %d Err %v", aid, evaluation, err)
		d.retryCh <- &model.RetryOp{
			Action:   _retryAddEvaluation,
			Value:    aid,
			SubValue: evaluation,
		}
		return
	}
	log.Info("AddEvaluation Aid %d, evaluation %d Succ", aid, evaluation)

}
