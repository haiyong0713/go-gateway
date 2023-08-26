package dao

import (
	"context"

	"go-common/library/database/elastic"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

// SearchArc search archive.
func (d *Dao) SearchArc(c context.Context, p *model.ArcListParam) (rs []*model.SearchArc, total int, err error) {
	req := d.Elastic.NewRequest(_esports).Index(_esports).Pn(p.Pn).Ps(p.Ps)
	req.Fields("id", "aid", "typeid", "title", "state", "mid", "gid", "tags", "teams", "matchs", "year", "source", "ctime")
	if p.Title != "" {
		req.WhereLike([]string{"title"}, []string{p.Title}, true, elastic.LikeLevelLow)
	}
	if p.Aid != "" {
		req.WhereEq("aid", p.Aid)
	}
	if p.TypeID > 0 {
		req.WhereEq("type_id", p.TypeID)
	}
	if p.Copyright > 0 {
		req.WhereEq("copyright", p.Copyright)
	}
	if p.State != "" {
		req.WhereEq("state", p.State)
	}
	if len(p.CheckState) > 0 {
		req.WhereIn("is_deleted", p.CheckState)
	}
	res := new(struct {
		Page struct {
			Num   int `json:"num"`
			Size  int `json:"size"`
			Total int `json:"total"`
		} `json:"page"`
		Result []*model.SearchArc `json:"result"`
	})
	if err = req.Scan(c, &res); err != nil || res == nil {
		log.Error("SearchArc req.Scan error(%v)", err)
		return
	}
	total = res.Page.Total
	rs = res.Result
	return
}
