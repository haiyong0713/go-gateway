package steins

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_graphSQL                  = "SELECT id,aid,regional_vars,global_vars,state,script,ctime,version,skin_id,no_tutorial,no_backtracking,no_evaluation,guest_overwrite_regional_vars FROM graph WHERE aid=? AND is_preview=0 ORDER BY id DESC LIMIT 1" // pick the latest version of arc's graph
	_graphPreviewSQL           = "SELECT id,aid,regional_vars,global_vars,state,script,ctime,version,skin_id,no_tutorial,no_backtracking,no_evaluation,guest_overwrite_regional_vars FROM graph WHERE aid=? AND is_preview=1 ORDER BY id DESC LIMIT 1"
	_graphIdByAidSQL           = "SELECT id FROM graph WHERE aid=? LIMIT 1"
	_graphListSQL              = "SELECT id,aid,script,ctime FROM graph WHERE aid=? AND state=1 AND is_preview=0 ORDER BY id DESC LIMIT ?"
	_graphByIDSQL              = "SELECT id,aid,script,ctime FROM graph WHERE id=? AND is_preview=0"
	_graphAddSQL               = "INSERT INTO graph(aid,regional_vars,global_vars,script,state,skin_id,is_preview,version) VALUE(?,?,?,?,?,?,?,?)"
	_graphAuditAddSQL          = "INSERT INTO graph_audit(aid,regional_vars,global_vars,script,state,skin_id) VALUE(?,?,?,?,?,?)"
	_graphNodeAddSQL           = "INSERT INTO graph_node(name,graph_id,cid,is_start,otype,show_time,skin_id) VALUE(?,?,?,?,?,?,?)"
	_graphNodeAuditAddSQL      = "INSERT INTO graph_node_audit(name,graph_id,cid,is_start,otype,show_time,skin_id) VALUE(?,?,?,?,?,?,?)"
	_graphEdgeBatchAddSQL      = "INSERT INTO graph_edge(graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`) VALUES%s"
	_graphEdgeAuditBatchAddSQL = "INSERT INTO graph_edge_audit(graph_id,from_node,title,to_node,to_node_cid,weight,text_align,pos_x,pos_y,is_default,script,attribute,`condition`) VALUES%s"
	_steinsRouteForStickVideo  = "stick_video"
	_graphVersionEdge          = 1 // edge剧情图
)

// LatestGraphList get latest graph list
func (d *Dao) LatestGraphList(c context.Context, aid int64, limit int) (list []*model.Graph, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _graphListSQL, aid, limit); err != nil {
		log.Error("LatestGraphList d.db.Query aid(%d) error(%v)", aid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(model.Graph)
		if err = rows.Scan(&r.ID, &r.Aid, &r.Script, &r.Ctime); err != nil {
			log.Error("LatestGraphList rows.Scan aid(%d) error(%v)", aid, err)
			return
		}
		list = append(list, r)
	}
	err = rows.Err()
	return
}

func (d *Dao) graphByID(c context.Context, graphID int64) (data *model.Graph, err error) {
	data = &model.Graph{}
	if err = d.db.QueryRow(c, _graphByIDSQL, graphID).Scan(&data.ID, &data.Aid, &data.Script, &data.Ctime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("graphID %d, empty graph", graphID)
		} else {
			err = errors.Wrapf(err, "graphID %d", graphID)
		}
		data = nil
		return
	}
	return
}

// SaveGraph .
//
//nolint:gocognit
func (d *Dao) SaveGraph(c context.Context, isPreview int, isAudit bool, param *model.SaveGraphParam, dimensions map[int64]*model.DimensionInfo, mid int64) (graphID int64, err error) {
	var (
		tx                    *xsql.Tx
		result, nodeResult    sql.Result
		lastNodeID, gFirstCid int64
		edgeSQLs              []string
		edgeArgs              []interface{}
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("SaveGraph d.db.Begin error(%v)", err)
		return
	}
	if isAudit {
		result, err = tx.Exec(_graphAuditAddSQL, param.Graph.Aid, param.Graph.RegionalStr, param.Graph.GlobalVars, param.Graph.Script, param.Graph.State, param.Graph.SkinID)
	} else {
		result, err = tx.Exec(_graphAddSQL, param.Graph.Aid, param.Graph.RegionalStr, param.Graph.GlobalVars, param.Graph.Script, param.Graph.State, param.Graph.SkinID, isPreview, _graphVersionEdge) // 全走edge剧情
	}
	if err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("SaveGraph tx.Exec audit graph(%v) error(%v)", param.Graph, err)
		return
	}
	if graphID, err = result.LastInsertId(); err != nil {
		//nolint:errcheck
		tx.Rollback()
		log.Error("SaveGraph LastInsertId error(%v)", err)
		return
	}
	fakeNodeIDs := make(map[string]string, len(param.Graph.Nodes))
	nodeNames := make(map[string]*api.GraphNode, len(param.Graph.Nodes))
	for _, node := range param.Graph.Nodes {
		if isAudit {
			nodeResult, err = tx.Exec(_graphNodeAuditAddSQL, node.Name, graphID, node.Cid, node.IsStart, node.Otype, node.ShowTime, node.SkinID)
		} else {
			nodeResult, err = tx.Exec(_graphNodeAddSQL, node.Name, graphID, node.Cid, node.IsStart, node.Otype, node.ShowTime, node.SkinID)
		}
		if err != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("SaveGraph tx.Exec node(%v) error(%v)", node, err)
			return
		}
		if lastNodeID, err = nodeResult.LastInsertId(); err != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("SaveGraph node LastInsertId error(%v)", err)
			return
		}
		fakeNodeIDs[node.ID] = node.Name
		nodeNames[node.Name] = &api.GraphNode{
			Id:       lastNodeID,
			Name:     node.Name,
			GraphId:  graphID,
			Cid:      node.Cid,
			IsStart:  node.IsStart,
			Otype:    node.Otype,
			ShowTime: node.ShowTime,
		}
		if node.IsStart == 1 { // pick the starting point's cid
			gFirstCid = node.Cid
		}
	}
	// group node id
	for _, node := range param.Graph.Nodes {
		var nodeID int64
		if n, ok := nodeNames[node.Name]; ok {
			nodeID = n.Id
		} else {
			log.Warn("SaveGraph node warn node(%+v)", node)
			continue
		}
		for _, edge := range node.Edges {
			var toNodeID, toNodeCid int64
			if nodeName, ok := fakeNodeIDs[edge.ToNodeID]; ok {
				if toNode, ok := nodeNames[nodeName]; ok {
					toNodeID = toNode.Id
					toNodeCid = toNode.Cid
				}
			}
			if toNodeID == 0 || toNodeCid == 0 {
				log.Warn("SaveGraph edge to node warn edge(%+v) param(%+v)", edge, param)
				continue
			}
			edgeSQLs = append(edgeSQLs, "(?,?,?,?,?,?,?,?,?,?,?,?,?)")
			edgeArgs = append(edgeArgs, graphID, nodeID, edge.Title, toNodeID, toNodeCid, edge.Weight, edge.TextAlign, edge.PosX, edge.PosY, edge.IsDefault, edge.Script, edge.AttributeStr, edge.ConditionStr)
		}
	}
	if len(edgeSQLs) > 0 {
		if isAudit {
			_, err = tx.Exec(fmt.Sprintf(_graphEdgeAuditBatchAddSQL, strings.Join(edgeSQLs, ",")), edgeArgs...)
		} else {
			_, err = tx.Exec(fmt.Sprintf(_graphEdgeBatchAddSQL, strings.Join(edgeSQLs, ",")), edgeArgs...)
		}
		if err != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("SaveGraph edge add error(%v)", err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
		return
	}
	if isPreview == model.GraphIsPreview {
		log.Warn("aid(%d) is preview", param.Graph.Aid)
		return
	}
	// preview don't need cache
	d.cache.Do(c, func(ctx context.Context) {
		//nolint:errcheck
		d.setGraphAllCache(ctx, param.Graph.Aid, graphID, dimensions)
		msg := &model.GraphPubMsg{Aid: param.Graph.Aid, Action: model.MsgActionPass}
		//nolint:bilirailguncheck
		if e := Retry(func() (err error) { // retry 10 times
			if err = d.graphPassPub.Send(ctx, strconv.FormatInt(param.Graph.Aid, 10), msg); err != nil {
				log.Warn("SaveGraph d.graphPassPub.Send(%+v) error(%v)", msg, err)
			} else {
				log.Info("SaveGraph d.graphPassPub.Send(%+v) ok", msg)
			}
			return
		}, 10, RetrySpan); e != nil {
			log.Error("SaveGraph d.graphPassPub.Send(%+v) error(%v)", msg, err)
		}
	})
	d.cache.Do(c, func(ctx context.Context) {
		msg := fmt.Sprintf("aid:%d  查看链接:http://api.bilibili.com/x/stein/manager?aid=%d", param.Graph.Aid, param.Graph.Aid)
		if e := d.SendWechat(ctx, d.c.Wechat.WxTitle, msg, d.c.Wechat.WxUser); e != nil {
			log.Error("SaveGraph d.SendWechat.Send(%+v) error(%v)", msg, e)
		}
	})
	if gFirstCid > 0 { // send databus msg to archive-job to update the firstCid of the archive and the videos
		d.cache.Do(c, func(ctx context.Context) {
			cidMdl := &model.SteinsCid{
				Aid:   param.Graph.Aid,
				Cid:   gFirstCid,
				Route: _steinsRouteForStickVideo,
			}
			//nolint:bilirailguncheck
			if e := Retry(func() (err error) { // retry 10 times
				if err = d.steinsCidPub.Send(ctx, cidMdl.Key(), cidMdl); err != nil {
					log.Warn("SaveGraph d.steinsCidPub.Send Aid %d Cid %d Err %v", cidMdl.Aid, cidMdl.Cid, err)
				} else {
					log.Info("SaveGraph d.steinsCidPub.Send Aid %d Cid %d Succ!", cidMdl.Aid, cidMdl.Cid)
				}
				return
			}, 10, RetrySpan); e != nil {
				log.Error("SaveGraph d.steinsCidPub.Send Aid %d Cid %d Err %v", cidMdl.Aid, cidMdl.Cid, e)
			}
		})
	}
	return
}

func (d *Dao) graph(c context.Context, aid int64, preview bool) (a *model.GraphDB, err error) {
	var sqlStr string
	if preview {
		sqlStr = _graphPreviewSQL
	} else {
		sqlStr = _graphSQL
	}
	a = &model.GraphDB{}
	if err = d.db.QueryRow(c, sqlStr, aid).Scan(&a.Id, &a.Aid, &a.RegionalVars, &a.GlobalVars, &a.State, &a.Script, &a.Ctime, &a.Version, &a.SkinId, &a.NoTutorial, &a.NoBacktracking, &a.NoEvaluation, &a.GuestOverwriteRegionalVars); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			log.Warn("aid %d, empty graph", aid)
		} else {
			err = errors.Wrapf(err, "aid %d", aid)
		}
		a = nil
		return
	}
	return
}

func (d *Dao) graphWithStarting(c context.Context, aid int64, isPreview bool) (a *api.GraphInfo, err error) {
	var graphDB *model.GraphDB
	if graphDB, err = d.graph(c, aid, isPreview); err != nil {
		log.Error("d.archive(%d) error(%v)", aid, err)
		return
	}
	if graphDB == nil || !graphDB.IsPass() {
		err = ecode.NothingFound
		return
	}
	var fiNid, fiCid int64
	if fiNid, fiCid, err = d.startingPoint(c, graphDB.Id); err != nil {
		return
	}
	a = graphDB.ToGraphInfo(fiNid, fiCid)
	return
}

func (d *Dao) ExistGraph(c context.Context, aid int64) (res bool, err error) {
	var graphId int64
	if err = d.db.QueryRow(c, _graphIdByAidSQL, aid).Scan(&graphId); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "aid %d", aid)
		}
		res = false
		return
	}
	res = graphId > 0
	return

}
