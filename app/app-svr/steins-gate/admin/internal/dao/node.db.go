package dao

import (
	"context"
	"database/sql"
	"go-common/library/ecode"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	_graphNodeListSQL = "SELECT id,name,graph_id,cid,is_start,otype,show_time FROM graph_node_audit WHERE graph_id=?"
	_graphNodeSQL     = "SELECT id,name,graph_id,cid,is_start,otype,show_time FROM graph_node_audit WHERE id=?"
	_startPointSQL    = "SELECT id,cid FROM graph_node_audit WHERE graph_id=? AND is_start=1 LIMIT 1"
)

// GraphNodeList graph node list.
func (d *Dao) GraphNodeList(c context.Context, graphID int64) (res []*api.GraphNode, err error) {
	rows, err := d.db.Query(c, _graphNodeListSQL, graphID)
	if err != nil {
		log.Error("d.db.Query(%d) error(%v)", graphID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		node := &api.GraphNode{}
		if err = rows.Scan(&node.Id, &node.Name, &node.GraphId, &node.Cid, &node.IsStart, &node.Otype, &node.ShowTime); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		res = append(res, node)
	}
	err = rows.Err()
	return
}

// RawNode get graphNode by node_id.
func (d *Dao) RawNode(c context.Context, id int64) (node *api.GraphNode, err error) {
	row := d.db.QueryRow(c, _graphNodeSQL, id)
	node = &api.GraphNode{}
	if err = row.Scan(&node.Id, &node.Name, &node.GraphId, &node.Cid, &node.IsStart, &node.Otype, &node.ShowTime); err != nil {
		if err == sql.ErrNoRows {
			node = nil
			err = nil
		} else {
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	return
}

func (d *Dao) StartingPoint(c context.Context, graphID int64) (firstNid, firstCid int64, err error) {
	if err = d.db.QueryRow(c, _startPointSQL, graphID).Scan(&firstNid, &firstCid); err != nil {
		if err == sql.ErrNoRows { // if the starting point is missing, the tree is not complete !!!!! so we return the error!
			err = ecode.NothingFound
			log.Error("graph %d, starting point not found", graphID)
		} else {
			log.Error("row.Scan error(%v)", err)
		}
		return
	}
	if firstNid == 0 || firstCid == 0 {
		err = ecode.NothingFound
		log.Error("graph %d firstNid %d firstCid %d is Empty", graphID, firstNid, firstCid)
	}
	return

}
