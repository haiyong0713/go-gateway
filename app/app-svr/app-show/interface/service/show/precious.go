package show

import (
	"bytes"
	"context"
	"text/template"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/log"
	"go-common/library/xstr"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/precious"

	"go-common/library/sync/errgroup.v2"

	pgrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"

	tagApi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

var (
	_preciousStyleDefault     = int64(0)
	_preciousStyleSplitLatest = int64(1)

	_TypeOrigin = int64(2)
	_TypeLatest = int64(1)

	_PublishVersionClassicial  = int64(1)
	_PublishVersionSplitLatest = int64(2)
)

func (s *Service) PreciousSubAdd(ctx context.Context, mid int64) error {
	if _, err := s.tagClient.AddSub(ctx, &tagApi.AddSubReq{Mid: mid, Tids: []int64{s.c.PreciousInfo.SubTagID}}); err != nil {
		return err
	}
	return nil
}

func (s *Service) PreciousSubDel(ctx context.Context, mid int64) error {
	if _, err := s.tagClient.CancelSub(ctx, &tagApi.CancelSubReq{Mid: mid, Tid: s.c.PreciousInfo.SubTagID}); err != nil {
		return err
	}
	return nil
}

func (s *Service) preciousSubscribed(ctx context.Context, mid int64) bool {
	tagReply, err := s.tagClient.Tag(ctx, &tagApi.TagReq{Mid: mid, Tid: s.c.PreciousInfo.SubTagID})
	if err != nil {
		log.Error("Failed to get precious subscribe status: %+v", err)
		return false
	}
	if tagReply.Tag == nil {
		log.Error("Invalid tag reply: %+v", tagReply)
		return false
	}
	if tagReply.Tag.Attention == 1 {
		return true
	}
	return false
}

func renderTemplate(input string, data interface{}) string {
	tmpl, err := template.New("").Parse(input)
	if err != nil {
		return input
	}
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, data); err != nil {
		return input
	}
	return out.String()
}

// Precious .
func (s *Service) Precious(ctx context.Context, style int64, mid int64, mobiApp, device string) (res *precious.Precious, err error) {
	var (
		cardRes   []cardm.Handler
		latestRes []cardm.Handler
		originRes []cardm.Handler
	)

	arcsReply, err := s.popular.Arcs(ctx)
	if err != nil {
		log.Error("日志告警 入站必刷获取后台数据错误: %+v", err)
		return nil, err
	}

	if cardRes, latestRes, originRes, err = s.dealCard(ctx, arcsReply.List, mid, mobiApp, device); err != nil {
		log.Error("[Precious] s.dealCard() error(%v)", err)
		return
	}
	if len(cardRes) == 0 {
		err = xecode.AppPreciousNotNorm
		return
	}
	res = &precious.Precious{
		MediaID: arcsReply.MediaId,
	}
	switch style {
	case _preciousStyleDefault: // 老版入站必刷
		res.H5Title = "入站必刷"
		res.Explain = renderTemplate("我不允许还有人没看过这{{.len}}个宝藏视频！", map[string]interface{}{"len": len(originRes)})
		res.Card = originRes
	case _preciousStyleSplitLatest: // 分离最近更新后的入站必刷
		res.H5Title = "bilibili入站必刷"
		switch arcsReply.PublishVersion {
		case _PublishVersionClassicial:
			res.OriginCard = cardRes
		case _PublishVersionSplitLatest:
			res.LatestCard = latestRes
			res.OriginCard = originRes
		default:
			log.Warn("Unrecognized publish version: %+v", arcsReply)
			res.OriginCard = cardRes
		}
		res.LatestCount = int64(len(latestRes))
		res.PageSubTitle = arcsReply.PageSubTitle
		res.ShareMainTitle = arcsReply.ShareMainTitle
		res.ShareSubTitle = renderTemplate(arcsReply.ShareSubTitle, map[string]interface{}{
			"len":          len(res.OriginCard),
			"latest_count": res.LatestCount,
		})
		res.Explain = res.ShareSubTitle
		if mid > 0 {
			res.Subscribed = s.preciousSubscribed(ctx, mid)
		}
	}
	return
}

// latest, origin
func splitLatestPrecious(in []*pgrpc.ArcList) ([]*pgrpc.ArcList, []*pgrpc.ArcList) {
	latest, origin := []*pgrpc.ArcList{}, []*pgrpc.ArcList{}
	for _, v := range in {
		switch v.Type {
		case _TypeOrigin:
			origin = append(origin, v)
			continue
		case _TypeLatest:
			latest = append(latest, v)
			continue
		default:
			log.Warn("unrecognized arc type: %+v", v)
			continue
		}
	}
	return latest, origin
}

func aidSet(in []*pgrpc.ArcList) sets.Int64 {
	out := sets.NewInt64()
	for _, i := range in {
		out.Insert(i.Aid)
	}
	return out
}

// dealCard .
func (s *Service) dealCard(ctx context.Context, arcList []*pgrpc.ArcList, mid int64, mobiApp, device string) (cardRes []cardm.Handler, latestRes []cardm.Handler, originRes []cardm.Handler, err error) {
	var (
		// cType cdm.CardType
		aids     []int64
		arcReply map[int64]*arcgrpc.Arc
		cardType = cdm.SmallCoverH6
		flowResp *creativeAPI.FlowJudgesReply
	)
	for _, v := range arcList {
		aids = append(aids, v.Aid)
	}
	latest, origin := splitLatestPrecious(arcList)
	latestAidSet, originAidSet := aidSet(latest), aidSet(origin)
	if len(aids) != 0 {
		gp := errgroup.WithContext(ctx)
		gp.Go(func(ctx context.Context) error {
			if arcReply, err = s.arc.ArchivesPB(ctx, aids, mid, mobiApp, device); err != nil {
				log.Error("[dealCard] s.arc.ArchivesPB() aids(%s) error(%v)", xstr.JoinInts(aids), err)
			}
			return err
		})
		gp.Go(func(ctx context.Context) error {
			if flowResp, err = s.creativeClient.FlowJudges(context.Background(), &creativeAPI.FlowJudgesReq{
				Oids:     aids,
				Business: 4,
				Gid:      24,
			}); err != nil {
				log.Error("s.creativeClient.FlowJudge Aids %v, error(%v)", aids, err)
				err = nil
			}
			return err
		})
		if err = gp.Wait(); err != nil {
			log.Error("gp.wait error(%v)", err)
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
	}
	if len(arcReply) != 0 {
		for _, v := range arcList {
			op := &operate.Card{
				Desc: v.Recommend,
			}
			op.From(cdm.CardGt(cdm.GotoAv), v.Aid, 0, 0, 0, "")
			handle := cardm.Handle(0, cdm.CardGt(cdm.GotoAv), cardType, cdm.ColumnSvrSingle, &ai.Item{Goto: model.GotoAv}, nil, nil, nil, nil, nil, nil)
			if handle == nil {
				continue
			}
			_ = handle.From(arcReply, op)
			if handle.Get().Right {
				cardRes = append(cardRes, handle)
				if latestAidSet.Has(v.Aid) {
					latestRes = append(latestRes, handle)
				}
				if originAidSet.Has(v.Aid) {
					originRes = append(originRes, handle)
				}
			}
		}
	}
	return
}
