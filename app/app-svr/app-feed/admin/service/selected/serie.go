package selected

import (
	"context"
	"fmt"
	"git.bilibili.co/bapis/bapis-go/archive/service"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"sync"
	"time"
)

// SelSeries picks all existing series
func (s *Service) SelSeries(c *bm.Context, sType string) (results []*selected.SerieFilter, err error) {
	var series []*selected.Serie
	if series, err = s.dao.Series(c, sType); err != nil {
		log.Error("SelSeries Err %v", err)
		return
	}
	results = make([]*selected.SerieFilter, 0)
	for _, v := range series {
		filter := &selected.SerieFilter{}
		filter.FromSerie(v)
		results = append(results, filter)
	}
	return
}

// SelPreview returns preview
func (s *Service) SelPreview(c *bm.Context, req *selected.PreviewReq) (data *selected.SelPreview, err error) {
	var (
		serie     *selected.Serie
		aids      []int64
		arcs      map[int64]*api.Arc
		reply     *selected.SelESReply
		noHotAids map[int64]struct{}
	)
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		Type:   req.Type,
		Number: req.Number,
	}); err != nil {
		log.Error("PickSerie PickSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	serie.Name = serie.SerieName()
	data = &selected.SelPreview{
		Config: serie,
		List:   make([]*selected.SelRes, 0),
	}
	reqES := &selected.ReqSelES{
		Status:  1, // passed cards
		SerieID: serie.ID,
	}
	if reply, err = s.esDao.SelResES(c, reqES); err != nil {
		log.Error("SelList ES Err %v", err)
		return
	}
	//log.Warn("es reply: %s", reply)
	for _, v := range reply.Result {
		aids = append(aids, v.RID)
	}
	if len(aids) == 0 {
		return
	}
	if arcs, err = s.arcDao.Arcs(c, aids); err != nil {
		log.Error("PickSerie Arcs Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	if noHotAids, _, err = s.arcDao.FlowJudge(c, aids, s.c.WeeklySelected.FlowCtrl); err != nil {
		log.Error("filterRes FlowJudge Aids %v, Err %v", aids, err)
		return
	}
	for _, v := range reply.Result {
		if arc, ok := arcs[v.RID]; ok && arc.IsNormal() {
			if _, okForbid := noHotAids[v.RID]; okForbid && v.RID != 290372201 { // filter popular forbidden arcs
				continue
			}
			res := &selected.SelRes{}
			res.FromArc(arc, v.RcmdReason)
			data.List = append(data.List, res)
		}
	}
	return
}

// SelPreview returns preview
func (s *Service) loadSeriesInUse() (err error) {
	if nums, err := s.dao.SeriesNums(context.Background()); err != nil {
		return ecode.Error(-500, "【每周必看-日志报警】获取每周必看获取最大期数失败")
	} else if len(nums) > 0 {
		// 优先刷新新增的一期
		if len(s.SeriesInUse) == 0 {
			s.SeriesInUse = make([]*selected.SelPreview, int(nums[0]))
		}
		if len(s.SeriesInUse) < int(nums[0]) {
			var req = &selected.PreviewReq{
				Type:   "weekly_selected",
				Number: nums[0],
			}
			if data, e := s.SelPreview(&bm.Context{Context: context.Background()}, req); e != nil {
				log.Error(fmt.Sprintf("【每周必看-日志报警】获取每周必看load新增第%d期数失败", nums[0]))
				return err
			} else {
				s.SeriesInUse = append([]*selected.SelPreview{data}, s.SeriesInUse...)
			}
		}
		lock := sync.Mutex{}
		for i, v := range nums {
			number := v
			index := i
			//nolint:biligowordcheck
			go (func() {
				var req = &selected.PreviewReq{
					Type:   "weekly_selected",
					Number: number,
				}
				if data, e := s.SelPreview(&bm.Context{Context: context.Background()}, req); e != nil {
					log.Error(fmt.Sprintf("【每周必看-日志报警】获取每周必看load第%d期数失败", number))
				} else {
					lock.Lock()
					s.SeriesInUse[index] = data
					lock.Unlock()
				}
			})()
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}

// 刷新每周必看数据缓存
func (s *Service) refCache(ctx context.Context, sType string, sNumber int64) (err error) {
	if err = s.dao.RefreshSeries(ctx, sType); err != nil {
		err = errors.WithMessagef(err, "refCache RefreshSeries sType(%s)", sType)
		return err
	}
	if err = s.dao.RefreshSingleSerie(ctx, sType, sNumber); err != nil {
		err = errors.WithMessagef(err, "refCache RefreshSingleSerie sType(%s) sNumber(%d)", sType, sNumber)
		return err
	}
	return nil
}

// 返回最新一期的每周必看核心数据
func (s *Service) LatestSelPreview(c *bm.Context) (data *selected.LatestSelPreviewReply, err error) {
	data = &selected.LatestSelPreviewReply{
		List: make([]*selected.SelResSimple, 0),
	}
	if s.SeriesInUse == nil {
		err = ecode.Error(-777, "【每周必看-日志报警】无法获取每周必看内容")
		return
	}
	if s.SeriesInUse[0] == nil {
		err = ecode.Error(-776, "【每周必看-日志报警】无法获取完整的最新一期每周必看内容")
		return
	}
	data.ID = s.SeriesInUse[0].Config.ID
	data.Number = s.SeriesInUse[0].Config.Number
	if len(s.SeriesInUse[0].List) == 0 {
		err = ecode.Error(-776, "【每周必看-日志报警】无法获取完整的最新一期每周必看内容")
		return
	}
	for _, v := range s.SeriesInUse[0].List {
		data.List = append(data.List, &selected.SelResSimple{Aid: v.Param, RcmdReason: v.RcmdReason})
	}
	return
}
