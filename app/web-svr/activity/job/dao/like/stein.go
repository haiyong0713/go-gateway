package like

import (
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/activity/job/component"

	"go-common/library/database/elastic"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _steinListKey = "stein_l"

func (d *Dao) SteinRuleCount(c context.Context, sid, viewRule, likeRule int64) (total int64, err error) {
	req := d.es.NewRequest(_activity).Index(_activity)
	req.WhereEq("state", 1)
	if sid > 0 {
		req.WhereEq("sid", sid)
	}
	if viewRule > 0 {
		req.WhereRange("click", viewRule, nil, elastic.RangeScopeLoRo)
	}
	if likeRule > 0 {
		req.WhereRange("likes", likeRule, nil, elastic.RangeScopeLoRo)
	}
	req.Pn(1).Ps(1)
	res := new(struct {
		Page struct {
			Total int `json:"total"`
		}
	})
	if err = req.Scan(c, res); err != nil {
		log.Error("SteinRuleCount req.Scan sid(%d) view(%d) like(%d) error(%v)", sid, viewRule, likeRule, err)
		return
	}
	total = int64(res.Page.Total)
	return
}

func (d *Dao) SetSteinCache(c context.Context, data *like.SteinList) (err error) {
	var (
		bs []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		return
	}
	if _, err = component.GlobalRedisStore.Do(c, "SET", _steinListKey, bs); err != nil {
		log.Error("SetSteinCache conn.Send(SET, %s, %s) error(%v)", _steinListKey, string(bs), err)
	}
	return
}
