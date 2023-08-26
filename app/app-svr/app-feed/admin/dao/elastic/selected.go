package elastic

import (
	"context"
	"fmt"

	"go-common/library/database/elastic"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
)

// SelResES picks result from ES
func (d *Dao) SelResES(c context.Context, req *selected.ReqSelES) (data *selected.SelESReply, err error) {
	var (
		cfg = d.c.Cfg.SelCfg
		r   = d.esClient.NewRequest(cfg.Business).
			Index(cfg.Index).WhereEq("deleted", 0).
			Order("serie_id", elastic.OrderDesc).Order("status", elastic.OrderAsc).Order("position", elastic.OrderAsc) // 先按期倒序排序，再把被拒绝的放后面，最后按位置排序
	)
	req.OneSerie()
	if req.SerieID != 0 {
		r = r.WhereEq("serie_id", req.SerieID)
	}
	if req.Status != 0 {
		r = r.WhereEq("status", req.Status)
	}
	if req.AID != 0 {
		r = r.WhereEq("aid", req.AID)
	}
	if req.Mid != 0 {
		r = r.WhereEq("mid", req.Mid)
	}
	if req.Title != "" {
		r = r.WhereLike([]string{"title"}, []string{req.Title}, true, elastic.LikeLevelMiddle)
	}
	if req.Author != "" {
		r = r.WhereLike([]string{"author"}, []string{req.Author}, true, elastic.LikeLevelMiddle)
	}
	if req.Creator != "" {
		r = r.WhereLike([]string{"creator"}, []string{req.Creator}, true, elastic.LikeLevelMiddle)
	}
	r.Ps(req.Ps).Pn(int(req.Pn))
	//log.Info("SelResES Params %s", r.Params())
	if err = r.Scan(c, &data); err != nil {
		log.Error("SelResES :Scan params(%s) error(%v)", r.Params(), err)
		return
	}
	if data == nil || data.Page == nil {
		err = fmt.Errorf("data or data.Page nil")
		log.Error("ArcES params(%s) error(%v)", r.Params(), err)
		return
	}
	return
}
