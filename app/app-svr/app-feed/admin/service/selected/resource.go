package selected

import (
	"bytes"
	"context"
	"encoding/csv"
	"math"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
)

// SelList def.
func (s *Service) SelList(c *bm.Context, req *selected.SelResReq) (data *selected.SelResReply, err error) {
	if req.Number != 0 {
		return s.oneSerieList(c, req)
	}
	return s.allSeriesList(c, req)
}

// oneSerieList returns the list with the given serie
func (s *Service) oneSerieList(c *bm.Context, req *selected.SelResReq) (data *selected.SelResReply, err error) {
	var (
		serie *selected.Serie
		reqES = &selected.ReqSelES{}
		reply *selected.SelESReply
	)
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		Type:   req.Type,
		Number: req.Number,
	}); err != nil {
		log.Error("SelList PickSerie Number %d, Type %s, Err %v", req.Number, req.Type, err)
		return
	}
	reqES.FromHTTP(req, serie.ID) // 查ES
	if reply, err = s.esDao.SelResES(c, reqES); err != nil {
		log.Error("SelList ES Err %v", err)
		return
	}
	if data, err = s.filterRes(c, reply, map[int64]*selected.Serie{serie.ID: serie}); err != nil {
		return
	}
	data.GetBvid()
	return
}

// SelList picks the list of selected resources
func (s *Service) allSeriesList(c *bm.Context, req *selected.SelResReq) (data *selected.SelResReply, err error) {
	var (
		reqES  = &selected.ReqSelES{}
		sids   []int64
		series map[int64]*selected.Serie
		reply  *selected.SelESReply
	)
	reqES.FromHTTP(req, 0) // 查ES
	if reply, err = s.esDao.SelResES(c, reqES); err != nil {
		log.Error("SelList ES Err %v", err)
		return
	}
	for _, v := range reply.Result {
		sids = append(sids, v.SerieID)
	}
	if series, err = s.dao.PickSeries(c, sids); err != nil {
		log.Error("SelList SelResES Number %d, Type %s, Sids %v, Err %v", req.Number, req.Type, sids, err)
		return
	}
	if data, err = s.filterRes(c, reply, series); err != nil {
		return
	}
	data.GetBvid()
	return
}

// filterRes 过滤热门禁止和up主删除的稿件，并且拼接上稿件的分区信息
func (s *Service) filterRes(c context.Context, reply *selected.SelESReply, series map[int64]*selected.Serie) (data *selected.SelResReply, err error) {
	var (
		aids        []int64
		cardIDs     []int64
		arcs        map[int64]*api.Arc
		noHotAids   map[int64]struct{}
		hotDownAids map[int64]struct{}
		cards       map[int64]*selected.SelES
		tps         map[int32]*api.Tp
	)
	data = &selected.SelResReply{
		Page:   reply.Page,
		Result: make([]*selected.SelShow, 0),
	}
	if len(reply.Result) == 0 {
		return
	}
	for _, v := range reply.Result {
		aids = append(aids, v.RID)
		cardIDs = append(cardIDs, v.ID)
	}
	if tps, err = s.arcDao.Types(c); err != nil {
		log.Error("filterRes Arcs Aids %v, Err %v", aids, err)
		return
	}
	if arcs, err = s.arcDao.Arcs(c, aids); err != nil {
		log.Error("filterRes Arcs Aids %v, Err %v", aids, err)
		return
	}
	if noHotAids, hotDownAids, err = s.arcDao.FlowJudge(c, aids, s.c.WeeklySelected.FlowCtrl); err != nil {
		log.Error("filterRes FlowJudge Aids %v, Err %v", aids, err)
		return
	}
	if cards, err = s.dao.Resources(c, cardIDs); err != nil {
		log.Error("resources cardIDs %v, err %v", cardIDs, err)
		return
	}
	for _, v := range reply.Result {
		var (
			stitle, zone string
			show         = &selected.SelShow{}
		)
		if card, ok := cards[v.ID]; ok { // 使用db结果替换es结果
			v.Position = card.Position
			v.RcmdReason = card.RcmdReason
			v.RID = card.RID
			v.SerieID = card.SerieID
			v.Source = card.Source
			v.Status = card.Status
		}
		if serie, ok := series[v.SerieID]; ok {
			stitle = serie.SerieName()
		}
		arc, okArc := arcs[v.RID]
		if !okArc { // 找不到稿件信息只能有期名
			log.Warn("filterRes Aid %d, Can't found Arc", v.RID)
			show.FromES(v, stitle, "", "")
			data.Result = append(data.Result, show)
			continue
		}
		if tp, ok := tps[arc.TypeID]; ok { // find the first level zone's name
			if tpP, okP := tps[tp.Pid]; okP {
				zone = tpP.Name
			}
		}
		if !arc.IsNormal() { // 稿件异常但是有稿件信息
			log.Warn("filterRes Aid %d, Arc Abnormal State %d", v.RID, arc.State)
			show.FromES(v, stitle, zone, arc.Pic)
			data.Result = append(data.Result, show)
			continue
		}
		if _, forbid := noHotAids[v.RID]; forbid { // 热门禁止稿件
			log.Warn("filterRes Aid %d, Arc NoHot", v.RID)
			show.FromES(v, stitle, zone, arc.Pic)
			show.IsNoHot = true
			data.Result = append(data.Result, show)
			continue
		}
		if _, forbid := hotDownAids[v.RID]; forbid { // 热门降权稿件
			log.Warn("filterRes Aid %d, Arc HotDown", v.RID)
			show.FromES(v, stitle, zone, arc.Pic)
			show.IsHotDown = true
			data.Result = append(data.Result, show)
			continue
		}
		show.IsNormal = true // 正常稿件 正常输出
		show.FromES(v, stitle, zone, arc.Pic)
		data.Result = append(data.Result, show)
	}
	return
}

// SelExport exports
func (s *Service) SelExport(c *bm.Context, req *selected.SelResReq) (data *bytes.Buffer, err error) {
	var (
		reply     *selected.SelResReply
		resources []*selected.SelShow
		pageCnt   int
	)
	if req.Number != 0 { // 指定单期
		if reply, err = s.oneSerieList(c, req); err != nil {
			log.Error("Export OneSerieList Req Type %s, Number %d, Err %v", req.Type, req.Number, err)
			return
		}
		resources = reply.Result
	} else { // 多期需要翻页拉取
		if reply, err = s.allSeriesList(c, req); err != nil {
			log.Error("Export allSeriesList Req Type %s, Number %d, Err %v", req.Type, req.Number, err)
			return
		}
		resources = reply.Result
		if reply.Page.Total > reply.Page.Size {
			pageCnt = int(math.Ceil(float64(reply.Page.Total) / float64(reply.Page.Size)))
			for i := 2; i <= pageCnt; i++ {
				req.Pn = i
				if reply, err = s.allSeriesList(c, req); err != nil {
					log.Error("Export allSeriesList Req Type %s, Number %d, Err %v", req.Type, req.Number, err)
					return
				}
				resources = append(resources, reply.Result...)
			}
		}
	}
	data = s.treatExport(resources)
	return
}

func (s *Service) treatExport(resources []*selected.SelShow) (data *bytes.Buffer) {
	var cfg = s.c.Cfg.SelCfg
	data = &bytes.Buffer{}
	csvWriter := csv.NewWriter(data)
	//nolint:errcheck
	csvWriter.Write(cfg.ExportTitles)
	for _, v := range resources {
		//nolint:errcheck
		csvWriter.Write(v.Export())
	}
	csvWriter.Flush()
	return
}
