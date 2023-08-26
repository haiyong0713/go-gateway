package show

import (
	"context"
	"strings"
	"time"

	"go-common/library/database/sql"

	api "go-gateway/app/app-svr/resource/service/api/v1"

	"github.com/pkg/errors"
)

const (
	_hiddensSQL = "SELECT h.id,h.sid,h.rid,h.cid,h.module_id,h.channel,h.pid,h.stime,h.etime,h.hidden_condition,h.hide_dynamic,hl.plat,hl.build,hl.conditions FROM entrance_hidden as h JOIN entrance_hidden_limit as hl ON h.id=hl.oid WHERE h.stime<=? AND h.etime>=? AND h.state=1 AND hl.state=1"
)

// Hiddens is
func (d *Dao) Hiddens(c context.Context, now time.Time) (hiddens []*api.Hidden, limits map[int64][]*api.HiddenLimit, err error) {
	var rows *sql.Rows
	limits = make(map[int64][]*api.HiddenLimit)
	if rows, err = d.db.Query(c, _hiddensSQL, now, now); err != nil {
		err = errors.Wrapf(err, "d.db.Query err")
		return
	}
	defer rows.Close()
	for rows.Next() {
		h := &api.Hidden{}
		hl := &api.HiddenLimit{}
		if err = rows.Scan(&h.Id, &h.Sid, &h.Rid, &h.Cid, &h.ModuleId, &h.Channel, &h.Pid, &h.Stime, &h.Etime, &h.HiddenCondition, &h.HideDynamic, &hl.Plat, &hl.Build, &hl.Conditions); err != nil {
			err = errors.Wrapf(err, "rows.Scan err")
			return
		}
		channelArr := strings.Split(h.Channel, ",")
		channelMap := make(map[string]string, len(channelArr))
		var channelFuzzy []string
		for _, v := range channelArr {
			if strings.Contains(v, "%") { //如果有%则要单独处理包含逻辑
				channelFuzzy = append(channelFuzzy, v)
				continue
			}
			channelMap[v] = v
		}
		h.ChannelMap = channelMap
		h.ChannelFuzzy = channelFuzzy
		if _, ok := limits[h.Id]; !ok {
			hiddens = append(hiddens, h)
		}
		limit := &api.HiddenLimit{
			Oid:        h.Id,
			Plat:       hl.Plat,
			Build:      hl.Build,
			Conditions: hl.Conditions,
		}
		limits[h.Id] = append(limits[h.Id], limit)
	}
	err = rows.Err()
	return
}
