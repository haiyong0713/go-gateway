package show

import (
	"context"
	"strconv"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/log"
	"go-common/library/xstr"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-show/ecode"
	swEcode "go-gateway/app/app-svr/app-show/ecode"
	svApi "go-gateway/app/app-svr/app-show/interface/api"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/aggregation"
	rcd "go-gateway/app/app-svr/app-show/interface/model/recommend"
	rcmdm "go-gateway/app/app-svr/app-show/interface/model/recommend"

	"go-common/library/sync/errgroup.v2"
)

const _hasMore = 1

// Aggregation .
// nolint:gomnd
func (s *Service) Aggregation(ctx context.Context, hotWordID int64, mid int64, mobiApp, device string) (res *aggregation.AggRes, aggSc []*rcmdm.CardList, err error) {
	var (
		aggRes    *aggregation.Aggregation
		resAICard []*rcd.CardList
	)
	res = &aggregation.AggRes{}
	gp := errgroup.WithCancel(ctx)
	gp.Go(func(ctx context.Context) (err error) {
		if aggRes, err = s.dao.Aggregation(ctx, hotWordID); err != nil || aggRes == nil {
			log.Error("[Aggregation] s.agg.Aggregation() hotWordID(%d) error(%v)", hotWordID, err)
			err = ecode.HotWordNoAuditingErr
			return
		}
		if aggRes.State != _auditingPass {
			log.Warn("[Aggregation] HotWordID(%d) State(%d) 审核状态不是通过", aggRes.ID, aggRes.State)
			err = ecode.HotWordNoAuditingErr
			return
		}
		return
	})
	gp.Go(func(ctx context.Context) (err error) {
		if resAICard, err = s.dao.AIAggregation(ctx, hotWordID); err != nil {
			log.Error("[Aggregation] s.agg.AIAggregation(%d) error(%v)", hotWordID, err)
			return
		}
		if len(resAICard) > 100 { // 最多100
			resAICard = resAICard[:100]
		}
		return
	})
	if err = gp.Wait(); err != nil {
		log.Error("gp.Wait() error(%v)", err)
		return
	}
	if len(resAICard) == 0 {
		log.Warn("[Aggregation] AI card return is 0")
		err = ecode.HotWordAIErr
		return
	}
	if res.Card, err = s.dealHotWordCard(ctx, resAICard, mid, mobiApp, device); err != nil {
		log.Error("[Aggregation]s.dealAggCard() error(%v)", err)
		return
	}
	res.Desc = aggRes.Subtitle
	res.H5Title = aggRes.HotTitle
	res.Image = aggRes.Image
	aggSc = resAICard
	return
}

func (s *Service) AggrSvideo(ctx context.Context, hotWordID int64, idx int64) (res *svApi.AggrSVideoReply, err error) {
	var (
		aggRes  *aggregation.Aggregation
		aggrTop = &svApi.SVideoTop{}
		gp      = errgroup.WithCancel(ctx)
	)
	gp.Go(func(ctx context.Context) (err error) {
		if aggRes, err = s.dao.Aggregation(ctx, hotWordID); err != nil {
			return
		}
		if aggRes == nil {
			err = ecode.HotWordNoAuditingErr
			return
		}
		if aggRes.State != _auditingPass {
			err = ecode.HotWordNoAuditingErr
			return
		}
		aggrTop.Title = aggRes.Title
		aggrTop.Desc = aggRes.Subtitle
		return
	})
	var (
		nextIdx   int
		ps        = 20
		hasMore   int32
		list      = make([]*svApi.SVideoItem, 0)
		resAICard []*rcd.CardList
	)
	gp.Go(func(ctx context.Context) (err error) {
		if resAICard, err = s.dao.AIAggregation(ctx, hotWordID); err != nil {
			return
		}
		if len(resAICard) <= int(idx) {
			err = swEcode.ActivityNothingMore
			return
		}
		cards := resAICard[idx:]
		for k, card := range cards {
			if card.ID == 0 {
				continue
			}
			tmp := &svApi.SVideoItem{
				Rid:   card.ID,
				Uid:   0,
				Index: int64(k) + idx,
			}
			list = append(list, tmp)
			if len(list) >= ps {
				nextIdx = int(idx) + k + 1
				break
			}
		}
		if nextIdx > 0 && nextIdx < len(resAICard) {
			hasMore = _hasMore
		}
		return
	})
	if err = gp.Wait(); err != nil {
		log.Error("AggrSvideo hotwordID(%d) idx(%d) error(%v)", hotWordID, idx, err)
		return
	}
	if len(resAICard) == 0 {
		log.Warn("AggrSvideo [Aggregation] AI card return is 0")
		err = ecode.HotWordAIErr
		return
	}
	res = &svApi.AggrSVideoReply{
		List:    list,
		Offset:  strconv.FormatInt(int64(nextIdx), 10),
		HasMore: hasMore,
		Top:     aggrTop,
	}
	return
}

// dealHotWordCard .
func (s *Service) dealHotWordCard(ctx context.Context, cards []*rcd.CardList, mid int64, mobiApp, device string) (cardRes []cardm.Handler, err error) {
	var (
		aids     []int64
		arcReply map[int64]*arcgrpc.Arc
		flowResp *creativeAPI.FlowJudgesReply
		rcmdArc  map[int64]struct{}
	)
	for _, v := range cards {
		if v.ID != 0 {
			aids = append(aids, v.ID)
		}
	}
	if len(aids) == 0 {
		log.Warn("[dealHotWordCard] AI return card aid is 0")
		err = ecode.HotWordAIErr
		return
	}
	gp := errgroup.WithContext(ctx)
	gp.Go(func(ctx context.Context) (err error) {
		if arcReply, err = s.arc.ArchivesPB(ctx, aids, mid, mobiApp, device); err != nil {
			log.Error("[dealAggCard] s.arc.ArchivesPB() aids(%s) error(%v)", xstr.JoinInts(aids), err)
		}
		return err
	})
	gp.Go(func(ctx context.Context) (err error) {
		if flowResp, err = s.creativeClient.FlowJudges(ctx, &creativeAPI.FlowJudgesReq{
			Oids:     aids,
			Business: 4,
			Gid:      24,
		}); err != nil {
			log.Error("s.creativeClient.FlowJudge Aids %v, error(%v)", aids, err)
			err = nil
		}
		return err
	})
	gp.Go(func(ctx context.Context) (err error) {
		if rcmdArc, err = s.rcmmnd.Recommend(ctx); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	if err = gp.Wait(); err != nil {
		log.Error("gp.wait error(%v)", err)
		return
	}
	if len(arcReply) == 0 {
		return
	}
	if flowResp != nil && len(flowResp.Oids) > 0 { // filter popular forbidden archive
		for _, v := range flowResp.Oids {
			if _, ok := arcReply[v]; ok {
				delete(arcReply, v)
				log.Info("[dealCard] popular forbidden archive aid(%d)", v)
			}
		}
	}
	for _, v := range cards {
		op := &operate.Card{
			Desc: v.Desc,
		}
		if rcmdArc != nil {
			if _, ok := rcmdArc[v.ID]; ok {
				op.IsPopular = true
			}
		}
		op.From(cdm.CardGt(cdm.GotoAv), v.ID, 0, 0, 0, "")
		handle := cardm.Handle(0, cdm.CardGt(cdm.GotoAv), cdm.SmallCoverH7, cdm.ColumnSvrSingle, &ai.Item{Goto: model.GotoAv}, nil, nil, nil, nil, nil, nil)
		if handle == nil {
			continue
		}
		_ = handle.From(arcReply, op)
		if handle.Get().Right {
			cardRes = append(cardRes, handle)
		}
	}
	return
}
