package steins

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

const (
	_edgeGroupsSQL = "SELECT id,graph_id,title,title_color,title_font_size,start_time,duration,type,pause_video,fade_in_time,fade_out_time FROM edge_group WHERE id IN (%s)"
)

// RawEdgeGroups get graphEdgeGroup by ids.
func (d *Dao) RawEdgeGroups(c context.Context, ids []int64) (res map[int64]*api.EdgeGroup, err error) {
	query := fmt.Sprintf(_edgeGroupsSQL, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, query)
	if err != nil {
		log.Error("db.Query(%s) error(%v)", query, err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*api.EdgeGroup)
	for rows.Next() {
		edgeGroup := &api.EdgeGroup{}
		if err = rows.Scan(&edgeGroup.Id, &edgeGroup.GraphId, &edgeGroup.Title, &edgeGroup.TitleColor, &edgeGroup.TitleFontSize, &edgeGroup.StartTime, &edgeGroup.Duration, &edgeGroup.Type, &edgeGroup.PauseVideo, &edgeGroup.FadeInTime, &edgeGroup.FadeOutTime); err != nil {
			err = errors.Wrapf(err, "edgeIDs %v", ids)
			return
		}
		res[edgeGroup.Id] = edgeGroup
	}
	err = rows.Err()
	return

}
