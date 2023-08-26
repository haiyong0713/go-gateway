package steins

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

const (
	_edgeByNodeSQL         = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id,is_hidden,width,height,to_time,to_type FROM graph_edge WHERE from_node=? ORDER BY id ASC"
	_edgeSQL               = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id,is_hidden,width,height,to_time,to_type FROM graph_edge WHERE id=?"
	_edgesSQL              = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id,is_hidden,width,height,to_time,to_type FROM graph_edge WHERE id in (%s)"
	_edgeListSQL           = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id,is_hidden,width,height,to_time,to_type FROM graph_edge WHERE graph_id=?"
	_edgeAuditListSQL      = "SELECT id,graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`,group_id,is_hidden,width,height,to_time,to_type FROM graph_edge_audit WHERE graph_id=?"
	_edgeFrameAnimationSQL = "SELECT edge_id,event,position,source_pic,item_height,item_width,item_count,fps,`columns`,rows,`loop` FROM graph_edge_frame_animation WHERE edge_id IN (%s)"
)

// RawEdge get graphEdge by node_id.
func (d *Dao) RawEdge(c context.Context, id int64) (edge *api.GraphEdge, err error) {
	row := d.db.QueryRow(c, _edgeSQL, id)
	edge = &api.GraphEdge{}
	var tmpScript sql.NullString
	if err = row.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign,
		&edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition, &edge.GroupId, &edge.IsHidden, &edge.Width, &edge.Height, &edge.ToTime, &edge.ToType); err != nil {
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

// RawEdges get graphEdge by from_node.
func (d *Dao) RawEdges(c context.Context, ids []int64) (res map[int64]*api.GraphEdge, err error) {
	query := fmt.Sprintf(_edgesSQL, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, query)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", query, err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*api.GraphEdge)
	for rows.Next() {
		edge := &api.GraphEdge{}
		var tmpScript sql.NullString
		if err = rows.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign,
			&edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition, &edge.GroupId, &edge.IsHidden, &edge.Width, &edge.Height, &edge.ToTime, &edge.ToType); err != nil {
			err = errors.Wrapf(err, "edgeIDs %v", ids)
			return
		}
		edge.Script = tmpScript.String
		res[edge.Id] = edge
	}
	err = rows.Err()
	return
}

// edgeByNode get graphEdge by from_node.
func (d *Dao) edgeByNode(c context.Context, fromNode int64) (res []*api.GraphEdge, err error) {
	rows, err := d.db.Query(c, _edgeByNodeSQL, fromNode)
	if err != nil {
		log.Error("db.Query from_node(%d) error(%v)", fromNode, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		edge := &api.GraphEdge{}
		var tmpScript sql.NullString
		if err = rows.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign,
			&edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition, &edge.GroupId, &edge.IsHidden, &edge.Width, &edge.Height, &edge.ToTime, &edge.ToType); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		edge.Script = tmpScript.String
		res = append(res, edge)
	}
	err = rows.Err()
	return
}

func (d *Dao) GraphEdgeList(c context.Context, graphID int64, opt ...interface{}) (edges []*api.GraphEdge, err error) {
	var rows *xsql.Rows
	if len(opt) > 0 && opt[0].(bool) {
		rows, err = d.db.Query(c, _edgeAuditListSQL, graphID)
	} else {
		rows, err = d.db.Query(c, _edgeListSQL, graphID)
	}
	if err != nil {
		log.Error("db.Query GraphEdgeList(%d) error(%v)", graphID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		edge := new(api.GraphEdge)
		var tmpScript sql.NullString
		if err = rows.Scan(&edge.Id, &edge.GraphId, &edge.FromNode, &edge.Title, &edge.ToNode, &edge.ToNodeCid, &edge.Weight, &edge.TextAlign,
			&edge.PosX, &edge.PosY, &edge.IsDefault, &tmpScript, &edge.Attribute, &edge.Condition, &edge.GroupId, &edge.IsHidden, &edge.Width, &edge.Height, &edge.ToTime, &edge.ToType); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		edge.Script = tmpScript.String
		edges = append(edges, edge)
	}
	err = rows.Err()
	return
}

func ractArray(in []byte) [16]byte {
	out := [16]byte{}
	//nolint:gomnd
	if len(in) <= 16 {
		copy(out[:], in)
		return out
	}
	copy(out[:], in[0:16])
	return out
}

func decodeRact(in [16]byte, r *api.Ract) {
	buf := [4]int32{}
	for i := 0; i < len(in) && i+4 <= len(in); i += 4 {
		ui := binary.BigEndian.Uint32(in[i : i+4])
		buf[i/4] = int32(ui)
	}
	r.X = buf[0]
	r.Y = buf[1]
	r.Width = buf[2]
	r.Height = buf[3]
}

// func encodeRact(r *api.Ract) [16]byte {
// 	out := [16]byte{}
// 	binary.BigEndian.PutUint32(out[0:4], uint32(r.X))
// 	binary.BigEndian.PutUint32(out[4:8], uint32(r.Y))
// 	binary.BigEndian.PutUint32(out[8:12], uint32(r.Width))
// 	binary.BigEndian.PutUint32(out[12:16], uint32(r.Height))
// 	return out
// }

// RawEdgeFrameAnimations is
func (d *Dao) RawEdgeFrameAnimations(ctx context.Context, edgeIDs []int64) (map[int64]*api.EdgeFrameAnimations, error) {
	if len(edgeIDs) <= 0 {
		return map[int64]*api.EdgeFrameAnimations{}, nil
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_edgeFrameAnimationSQL, xstr.JoinInts(edgeIDs)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]*api.EdgeFrameAnimations{}
	for rows.Next() {
		efa := &api.EdgeFrameAnimation{}
		posBin := []byte{}
		if err := rows.Scan(&efa.EdgeId, &efa.Event, &posBin, &efa.SourcePic, &efa.ItemHeight, &efa.ItemWidth,
			&efa.ItemCount, &efa.Fps, &efa.Colums, &efa.Rows, &efa.Loop); err != nil {
			log.Error("Failed to scan edge frame animation: %+v", err)
			continue
		}
		decodeRact(ractArray(posBin), &efa.Position)
		if _, ok := out[efa.EdgeId]; !ok {
			out[efa.EdgeId] = &api.EdgeFrameAnimations{
				Animations: map[string]*api.EdgeFrameAnimation{},
			}
		}
		out[efa.EdgeId].Animations[efa.Event] = efa
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return out, nil

}
