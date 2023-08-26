package steins

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/steins-gate/service/api"
)

const (
	_graphNodeSQL          = "SELECT id,name,graph_id,cid,is_start,otype,show_time,skin_id FROM graph_node WHERE id=?"
	_graphNodesSQL         = "SELECT id,name,graph_id,cid,is_start,otype,show_time,skin_id FROM graph_node WHERE id in (%s)"
	_graphNodeListSQL      = "SELECT id,name,graph_id,cid,is_start,otype,show_time,skin_id FROM graph_node WHERE graph_id=?"
	_graphNodeAuditListSQL = "SELECT id,name,graph_id,cid,is_start,otype,show_time,skin_id FROM graph_node_audit WHERE graph_id=?"
	_startPointSQL         = "SELECT id,cid FROM graph_node WHERE graph_id=? AND is_start=1 LIMIT 1"
)

// RawNode get graphNode by node_id.
func (d *Dao) RawNode(c context.Context, id int64) (node *api.GraphNode, err error) {
	row := d.db.QueryRow(c, _graphNodeSQL, id)
	node = &api.GraphNode{}
	if err = row.Scan(&node.Id, &node.Name, &node.GraphId, &node.Cid, &node.IsStart, &node.Otype, &node.ShowTime, &node.SkinId); err != nil {
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

// RawNodes get graphEdge by from_node.
func (d *Dao) RawNodes(c context.Context, ids []int64) (res map[int64]*api.GraphNode, err error) {
	query := fmt.Sprintf(_graphNodesSQL, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, query)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", query, err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*api.GraphNode)
	for rows.Next() {
		node := &api.GraphNode{}
		if err = rows.Scan(&node.Id, &node.Name, &node.GraphId, &node.Cid, &node.IsStart, &node.Otype, &node.ShowTime, &node.SkinId); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		res[node.Id] = node
	}
	err = rows.Err()
	return
}

// GraphNodeList graph node list.
func (d *Dao) GraphNodeList(c context.Context, graphID int64, opt ...interface{}) (res []*api.GraphNode, err error) {
	var rows *xsql.Rows
	if len(opt) > 0 && opt[0].(bool) { // 是否是audit表
		rows, err = d.db.Query(c, _graphNodeAuditListSQL, graphID)
	} else {
		rows, err = d.db.Query(c, _graphNodeListSQL, graphID)
	}
	if err != nil {
		log.Error("d.db.Query(%d) error(%v)", graphID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		node := &api.GraphNode{}
		if err = rows.Scan(&node.Id, &node.Name, &node.GraphId, &node.Cid, &node.IsStart, &node.Otype, &node.ShowTime, &node.SkinId); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		res = append(res, node)
	}
	err = rows.Err()
	return
}

func (d *Dao) startingPoint(c context.Context, graphID int64) (firstNid, firstCid int64, err error) {
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
