package steins

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_recentArcs = "SELECT DISTINCT (aid) FROM graph WHERE is_preview=0 "
)

// ManagerList .
func (d *Dao) ManagerList(c context.Context, aid int64) (data *model.ManagerGraph, err error) {
	var (
		graphDB *model.GraphDB
		nodes   []*api.GraphNode
		edges   []*api.GraphEdge
	)
	if graphDB, err = d.graph(c, aid, false); err != nil {
		log.Error("ManagerList d.graph(%d) error(%v)", aid, err)
		err = nil
	}
	if graphDB == nil {
		err = ecode.NothingFound
		return
	}
	data = &model.ManagerGraph{Aid: aid}
	if nodes, err = d.GraphNodeList(c, graphDB.Id); err != nil {
		log.Error("ManagerList d.GraphNodeList graphID(%d) error(%v)", graphDB.Id, err)
		err = nil
	} else {
		for _, v := range nodes {
			data.NodeNameArr = append(data.NodeNameArr, v.Name)
		}
		data.NodeNames = strings.Join(data.NodeNameArr, "||")
	}
	if edges, err = d.GraphEdgeList(c, graphDB.Id); err != nil {
		log.Error("ManagerList d.GraphEdgeList graphID(%d) error(%v)", graphDB.Id, err)
		err = nil
	} else {
		for _, v := range edges {
			data.EdgeNameArr = append(data.EdgeNameArr, v.Title)
		}
		data.EdgeNames = strings.Join(data.EdgeNameArr, "||")
	}
	return
}

// RecentArcs .
func (d *Dao) RecentArcs(c context.Context, param *model.RecentArcReq) (res []int64, err error) {
	var query string
	if param.Stime > 0 {
		query += fmt.Sprintf(" AND mtime > \"%s\"", time.Unix(param.Stime, 0).Format("2006-01-02 15:04:05"))
	}
	if param.Etime > 0 {
		query += fmt.Sprintf(" AND mtime < \"%s\"", time.Unix(param.Etime, 0).Format("2006-01-02 15:04:05"))
	}
	if query == "" {
		query += fmt.Sprintf(" AND mtime > \"%s\"", time.Now().Format("2006-01-02")) // 默认出当天结果
	}
	query = _recentArcs + query + " LIMIT 200"
	rows, err := d.db.Query(c, query)
	if err != nil {
		err = errors.Wrapf(err, "db.Query(%s) error(%v)", query, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var aid int64
		if err = rows.Scan(&aid); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		res = append(res, aid)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "db.Rows(%s) error(%v)", query, err)
	}
	return

}
