package dao

import (
	"context"
	"database/sql"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/job/internal/model"
)

const (
	_graphSQL       = "SELECT id,aid,state FROM graph WHERE aid=? AND is_preview=0 ORDER BY id DESC LIMIT 1" // pick the latest version of arc's graph
	_nodesSQL       = "SELECT id,cid,is_start FROM graph_node WHERE graph_id=?"
	_returnGraphSQL = "UPDATE graph SET state=-6 WHERE id=?"
	_preFixGraph    = "gh_"
)

func graphKey(aid int64) string {
	return _preFixGraph + strconv.FormatInt(aid, 10)
}

// Graph picks the graph info
func (d *Dao) Graph(c context.Context, aid int64) (graph *model.Graph, err error) {
	graph = new(model.Graph)
	if err = d.db.QueryRow(c, _graphSQL, aid).Scan(&graph.ID, &graph.AID, &graph.State); err != nil {
		if err == sql.ErrNoRows {
			log.Warn("aid %d, empty graph", aid)
		} else {
			log.Error("row.Scan error(%v)", err)
		}
		graph = nil
		return
	}
	return
}

// Nodes picks the nodes of the given graph
func (d *Dao) Nodes(c context.Context, graphID int64) (result []*model.Node, firstCid int64, err error) {
	rows, err := d.db.Query(c, _nodesSQL, graphID)
	if err != nil {
		log.Error("mysql.Query %s error(%v)", _nodesSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.Node{}
		if err = rows.Scan(&r.ID, &r.CID, &r.IsStart); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		result = append(result, r)
		if r.IsStart == 1 { // pick the starting point
			firstCid = r.CID
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.err error(%v)", err)
		return
	}
	return
}

// ReturnGraph def.
func (d *Dao) returnGraph(c context.Context, graphID int64) (err error) {
	if _, err = d.db.Exec(c, _returnGraphSQL, graphID); err != nil {
		log.Error("ReturnGraph GraphID %d, Err %v", graphID, err)
		return
	}
	return
}

// ReturnGraph will retry returnGraph operation in case of error
func (d *Dao) ReturnGraph(c context.Context, graphID int64) {
	var err error
	if err = d.returnGraph(c, graphID); err != nil {
		select {
		case d.retryCh <- &model.RetryOp{
			Action: _retryReturn,
			Value:  graphID,
		}:
		default:
			log.Error("retry channel full value %d, action %s", graphID, _retryReturn)
		}
	} else {
		log.Info("ReturnGraph %d Succ", graphID)
	}
}

// delGraphCache def.
func (d *Dao) delGraphCache(c context.Context, aid int64) (err error) {
	var (
		key  = graphKey(aid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelGraphCache Aid %d, Err %v", aid, err)
	}
	return
}

// DelGraphCache will retry delGraphCache operation in case of error
func (d *Dao) DelGraphCache(c context.Context, aid int64) {
	var err error
	if err = d.delGraphCache(c, aid); err != nil {
		select {
		case d.retryCh <- &model.RetryOp{
			Action: _retryDelCache,
			Value:  aid,
		}:
		default:
			log.Error("retry channel full value %d, action %s", aid, _retryDelCache)
		}
	} else {
		log.Info("DelGraphCache Aid %d Succ", aid)
	}

}
