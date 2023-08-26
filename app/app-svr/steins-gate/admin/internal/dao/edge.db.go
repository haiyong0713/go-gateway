package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

const (
	_edgeListSQL   = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition` FROM graph_edge_audit WHERE graph_id=?"
	_edgeByNodeSQL = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition` FROM graph_edge_audit WHERE from_node=? ORDER BY id ASC"
	_edgeSQL       = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id FROM graph_edge_audit WHERE id=?"
)

func (d *Dao) GraphEdgeList(c context.Context, graphID int64) (edges []*api.GraphEdge, err error) {
	rows, err := d.db.Query(c, _edgeListSQL, graphID)
	if err != nil {
		log.Error("db.Query graphEdgeList(%d) error(%v)", graphID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		edge := new(api.GraphEdge)
		var tmpScript sql.NullString
		if err = rows.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign, &edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		edge.Script = tmpScript.String
		edges = append(edges, edge)
	}
	err = rows.Err()
	return
}

// edgeByNode get graphEdge by from_node.
func (d *Dao) EdgeByNode(c context.Context, fromNode int64) (res []*api.GraphEdge, err error) {
	rows, err := d.db.Query(c, _edgeByNodeSQL, fromNode)
	if err != nil {
		log.Error("db.Query from_node(%d) error(%v)", fromNode, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		edge := &api.GraphEdge{}
		var tmpScript sql.NullString
		if err = rows.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign, &edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		edge.Script = tmpScript.String
		res = append(res, edge)
	}
	err = rows.Err()
	return
}

func (d *Dao) RawEdge(c context.Context, id int64) (edge *api.GraphEdge, err error) {
	row := d.db.QueryRow(c, _edgeSQL, id)
	edge = &api.GraphEdge{}
	var tmpScript sql.NullString
	if err = row.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign,
		&edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition, &edge.GroupId); err != nil {
		if err == sql.ErrNoRows {
			edge = nil
			err = nil
		} else {
			err = errors.Wrapf(err, "edgeID %d", id)
		}
		return
	}
	return

}
