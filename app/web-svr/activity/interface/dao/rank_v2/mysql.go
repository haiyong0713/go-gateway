package rank

import (
	"context"
	"database/sql"

	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank"

	"go-common/library/log"
)

const (
	rankDBName = "act_rank_config"
)

const (
	getRankConfigBySIDSQL = "SELECT id,sid,sid_source,ratio,rank_type,rank_attribute,rank_top,is_auto,is_show_score,state,stime,etime,statistics_time,ctime,mtime from act_rank_config where sid=? and sid_source=? and state = ?"
)

// GetRankConfigBySid
func (d *dao) GetRankConfigBySid(c context.Context, sid int64, sidSource int) (rank *rankmdl.Rank, err error) {
	var (
		arg []interface{}
	)
	rank = &rankmdl.Rank{}
	arg = append(arg, sid)
	arg = append(arg, sidSource)
	arg = append(arg, rankmdl.RankStateOnline)
	row := d.db.QueryRow(c, getRankConfigBySIDSQL, arg...)
	if err = row.Scan(&rank.ID, &rank.SID, &rank.SIDSource, &rank.Ratio, &rank.RankType, &rank.RankAttribute, &rank.Top, &rank.IsAuto, &rank.IsShowScore, &rank.State, &rank.Stime, &rank.Etime, &rank.StatisticsTime, &rank.Ctime, &rank.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(c, "Rank@GetRankConfigBySid d.db.QueryRow() SELECT failed. error(%v)", err)
	}
	return
}
