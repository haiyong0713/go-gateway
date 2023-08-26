package rank

import (
	"context"
	"go-common/library/database/elastic"
	"go-common/library/log"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank_v3"
)

const (
	activityRank = "activity_rank"
)

func (d *dao) RankResult(c context.Context, baseID, rankID, batch, aid, mid, tagID, pn, ps int64) (list []*rankmdl.ResultFromES, total int64, err error) {
	var res struct {
		Page struct {
			Num   int   `json:"num"`
			Size  int   `json:"size"`
			Total int64 `json:"total"`
		} `json:"page"`
		Result []*rankmdl.ResultFromES `json:"result"`
	}
	req := d.es.NewRequest(activityRank).Index("log_user_action_250_all")
	req = req.WhereEq("base_id", baseID)
	req = req.WhereEq("rank_id", rankID)
	req = req.WhereEq("batch", batch)
	req = req.WhereEq("aid", aid)
	if mid > 0 {
		req = req.WhereEq("mid", mid)
	}
	if tagID > 0 {
		req = req.WhereEq("tag_id", tagID)
	}
	if aid > 0 {
		req = req.WhereEq("aid", aid)
	}
	if ps > 0 {
		req = req.Ps(int(ps))
	}
	if err = req.Pn(int(pn)).Order("rank", elastic.OrderDesc).Scan(c, &res); err != nil {
		log.Error("RankResult aid(%d) mid(%d), err: %v", aid, mid, err)
		return
	}
	list = res.Result
	total = res.Page.Total
	return
}
