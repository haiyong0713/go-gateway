package show

import (
	"context"

	"github.com/pkg/errors"
	"go-common/library/ecode"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	v1 "go-gateway/app/app-svr/app-show/interface/api"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/selected"

	"go-common/library/sync/errgroup.v2"

	tagApi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const _weeklySelected = "weekly_selected"

// AddFav def.
func (s *Service) AddFav(c context.Context, mid int64) error {
	if s.c.Custom.TagSwitchOn {
		if _, err := s.tagClient.AddSub(c, &tagApi.AddSubReq{Mid: mid, Tids: []int64{s.c.Custom.SelectedTid}}); err != nil {
			return err
		}
	}
	if s.c.Custom.FavSwitchOn {
		if err := s.favdao.FavAdd(c, mid, _weeklySelected); err != nil {
			return err
		}
	}
	return nil
}

// DelFav def.
func (s *Service) DelFav(c context.Context, mid int64) (err error) {
	if s.c.Custom.TagSwitchOn {
		if _, err := s.tagClient.CancelSub(c, &tagApi.CancelSubReq{Mid: mid, Tid: s.c.Custom.SelectedTid}); err != nil {
			return err
		}
	}
	if s.c.Custom.FavSwitchOn {
		if err := s.favdao.FavDel(c, mid, _weeklySelected); err != nil {
			return err
		}
	}
	return nil
}

// CheckFav def.
func (s *Service) CheckFav(c context.Context, mid int64) (*selected.FavStatus, error) {
	var status bool
	if s.c.Custom.TagSwitchOn {
		tagReply, err := s.tagClient.Tag(c, &tagApi.TagReq{Mid: mid, Tid: s.c.Custom.SelectedTid})
		if err != nil || tagReply == nil || tagReply.Tag == nil {
			return nil, errors.Wrapf(err, "CheckFav reply error")
		}
		if tagReply.Tag.Attention == 1 {
			status = true
		}
	}
	if s.c.Custom.FavSwitchOn {
		favReply, err := s.favdao.FavCheck(c, mid, _weeklySelected)
		if err != nil {
			return nil, err
		}
		status = favReply
	}
	return &selected.FavStatus{Status: status}, nil
}

// AllSeries picks one the available series of one defined type
func (s *Service) AllSeries(c context.Context, sType string) (result []*selected.SerieFilter, err error) {
	if result, err = s.cdao.AllSeriesCache(c, sType); err != nil {
		log.Error("All_Series Stype %s, Redis Err %v", sType, err)
	}
	if len(result) > 0 { // if redis cache already satisfies the need, just return
		s.pHit.Incr("All_Series")
		return
	}
	s.pMiss.Incr("All_Series")
	return s.BackToSrcSeries(c, sType)
}

// getSerie picks one serie with its type and number
func (s *Service) getSerie(c context.Context, sType string, number int64) (result *selected.SerieFull, err error) {
	if result, err = s.cdao.PickSerieCache(c, sType, number); err == nil {
		s.pHit.Incr("SerieFull")
		return
	}
	s.pMiss.Incr("SerieFull")
	return s.BackToSrcSerie(c, sType, number)
}

// SerieShow shows the complete data of one selected serie
func (s *Service) SerieShow(c context.Context, sType string, number int64, mid int64, mobiApp, device string) (result *selected.SerieShow, err error) {
	var (
		full *selected.SerieFull
	)
	if full, err = s.getSerie(c, sType, number); err != nil {
		log.Error("SerieShow getSerie sType %s, Number %d", sType, number)
		return
	}
	result = &selected.SerieShow{
		Config:   full.Config,
		Reminder: s.c.ShowSelectedCfg.Reminder,
	}
	result.List = s.dealRes(c, full.List, mid, mobiApp, device)
	if full.Config.IsDisaster() {
		if len(result.List) > s.c.ShowSelectedCfg.DisasterMax { // 灾备数据最大出卡量限制
			result.List = result.List[0:s.c.ShowSelectedCfg.DisasterMax]
		}
		if len(result.List) > 0 {
			full.Config.ShareSubtitle = result.List[0].Get().Title // 灾备分享副标题为第一个稿件的标题
		}
	}
	return
}

// BatchSerie .
func (s *Service) BatchSerie(c context.Context, typ string, numbers []int64) (*v1.BatchSerieRly, error) {
	serie, err := s.cdao.BatchPickSerieCache(c, typ, numbers)
	if err != nil {
		return nil, err
	}
	if serie == nil {
		return nil, ecode.NothingFound
	}
	rly := &v1.BatchSerieRly{List: make(map[int64]*v1.SerieConfig)}
	for k, v := range serie {
		if v == nil || v.Config == nil {
			continue
		}
		rly.List[k] = &v1.SerieConfig{
			Number:        v.Config.Number,
			Subject:       v.Config.Subject,
			Label:         v.Config.Label,
			Hint:          v.Config.Hint,
			Color:         int64(v.Config.Color),
			Cover:         v.Config.Cover,
			ShareTitle:    v.Config.ShareTitle,
			ShareSubtitle: v.Config.ShareSubtitle,
			MediaId:       v.Config.MediaID,
		}
	}
	return rly, nil
}

func (s *Service) SelectedSerie(c context.Context, typ string, number int64) (*v1.SelectedSerieRly, error) {
	serie, err := s.cdao.PickSerieCache(c, typ, number)
	if err != nil {
		return nil, err
	}
	if serie == nil {
		return nil, ecode.NothingFound
	}
	rly := &v1.SelectedSerieRly{List: make([]*v1.SelectedRes, 0, len(serie.List))}
	if serie.Config != nil {
		rly.Config = &v1.SerieConfig{
			Number:        serie.Config.Number,
			Subject:       serie.Config.Subject,
			Label:         serie.Config.Label,
			Hint:          serie.Config.Hint,
			Color:         int64(serie.Config.Color),
			Cover:         serie.Config.Cover,
			ShareTitle:    serie.Config.ShareTitle,
			ShareSubtitle: serie.Config.ShareSubtitle,
			MediaId:       serie.Config.MediaID,
		}
	}
	for _, res := range serie.List {
		if res == nil {
			continue
		}
		rly.List = append(rly.List, &v1.SelectedRes{
			Rid:        res.RID,
			Rtype:      res.Rtype,
			SerieId:    res.SerieID,
			Position:   int64(res.Position),
			RcmdReason: res.RcmdReason,
		})
	}
	return rly, nil
}

// dealRes treats selected resources
func (s *Service) dealRes(c context.Context, cards []*selected.SelectedRes, mid int64, mobiApp, device string) (is []cardm.Handler) {
	var (
		aids, avUpIDs []int64
		am            map[int64]*arcgrpc.Arc
		err           error
		accountm      map[int64]*accountgrpc.Card
		flowResp      *creativeAPI.FlowJudgesReply
	)
	for _, ca := range cards { // gather resource ids
		switch ca.Rtype {
		case model.GotoAv:
			aids = append(aids, ca.RID)
		}
	}
	if len(aids) != 0 { // pick archive data and author mids
		g := errgroup.WithContext(c)
		g.Go(func(ctx context.Context) (err error) {
			if am, err = s.arc.ArchivesPB(ctx, aids, mid, mobiApp, device); err != nil {
				log.Error("%+v", err)
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if flowResp, err = s.creativeClient.FlowJudges(context.Background(), &creativeAPI.FlowJudgesReq{
				Oids:     aids,
				Business: 4,
				Gid:      24,
			}); err != nil {
				log.Error("s.creativeClient.FlowJudge Aids %v, error(%v)", aids, err)
				err = nil
			}
			return
		})
		if err = g.Wait(); err != nil { // if archive service error, directly return error
			return
		}
		if flowResp != nil && len(flowResp.Oids) > 0 { // filter popular forbidden archive
			for _, v := range flowResp.Oids {
				delete(am, v)
			}
		}
		for _, a := range am {
			avUpIDs = append(avUpIDs, a.Author.Mid)
		}
	}
	if len(avUpIDs) > 0 { // pick upper info
		if accountm, err = s.acc.Cards3GRPC(c, avUpIDs); err != nil {
			log.Error("%+v", err)
		}
	}
	for _, ca := range cards {
		var (
			main     interface{}
			cardType cdm.CardType
			r        = ca.ToAIItem()
		)
		op := &operate.Card{
			Desc: ca.RcmdReason,
		}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, 0, 0, "")
		switch r.Goto {
		case model.GotoAv:
			cardType = cdm.SmallCoverH5
			main = am
		}
		h := cardm.Handle(0, cdm.CardGt(r.Goto), cardType, cdm.ColumnSvrSingle, r, nil, nil, nil, nil, accountm, nil)
		if h == nil {
			continue
		}
		_ = h.From(main, op)
		if h.Get().Right { // filter abnormal cards
			is = append(is, h)
		}
	}
	rl := len(is)
	if rl == 0 {
		is = _emptyList2
		return
	}
	return
}
