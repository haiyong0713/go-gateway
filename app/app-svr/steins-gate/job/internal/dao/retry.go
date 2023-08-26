package dao

import (
	"context"
	"time"

	"go-common/library/log"
)

const (
	_retryReturn        = "return"
	_retryDelCache      = "del_cache"
	_retrySendDatabus   = "cid_dbus"
	_retryAddEvaluation = "add_eval"
)

func (d *Dao) retryproc() {
	defer d.waiter.Done()
	for {
		retryMdl, ok := <-d.retryCh
		if !ok || d.daoClosed {
			log.Warn("Retryproc exit")
			return
		}
		time.Sleep(100 * time.Millisecond)
		switch retryMdl.Action {
		case _retryReturn:
			d.ReturnGraph(context.Background(), retryMdl.Value)
		case _retryDelCache:
			d.DelGraphCache(context.Background(), retryMdl.Value)
		case _retrySendDatabus:
			d.UpArcFirstCid(context.Background(), retryMdl.Value, retryMdl.SubValue)
		case _retryAddEvaluation:
			d.AddEval(context.Background(), retryMdl.Value, retryMdl.SubValue)
		default:
			log.Warn("Retry Illegal Type %s ", retryMdl.Action)
		}
	}

}
