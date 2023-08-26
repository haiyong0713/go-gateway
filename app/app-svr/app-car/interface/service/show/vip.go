package show

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	cardecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/view"
	"go-gateway/app/app-svr/app-car/interface/model/vip"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	resourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

const (
	_pgcSeason = 1
	_pgcEp     = 2
)

func (s *Service) AddVip(c context.Context, plat int8, mid int64, buvid, path, ua string, param *vip.VipParam) ([]cardm.Handler, error) {
	tec := &view.SilverEventCtx{
		Mid:       mid,
		Buvid:     buvid,
		Ip:        metadata.String(c, metadata.RemoteIP),
		Platform:  param.Platform,
		Ctime:     time.Now().Format("2006-01-02 15:04:05"),
		Api:       path,
		Origin:    param.AppKey,
		UserAgent: ua,
		Build:     strconv.Itoa(param.Build),
		VipEventCtx: &view.VipEventCtx{
			SubScene:    "car_getmemeber",
			ActivityUID: "car_banner_vip",
		},
	}
	if s.silverDao.RuleCheck(c, tec, model.SilverGaiaCommonActivity) {
		return nil, cardecode.AppCarVipRiskUser
	}
	order := s.orderID()
	// 写入db
	if err := s.saveDB(c, mid, buvid, order, param); err != nil {
		return nil, err
	}
	// 领取
	resArg := &resourceapi.ResourceUseAsyncReq{
		BatchToken: s.c.VipConfig.BatchToken,
		Mid:        mid,
		OrderNo:    order,
		Remark:     "车载大会员",
		Appkey:     s.c.VipConfig.AppKey,
		Ts:         time.Now().Unix(),
	}
	if err := s.resDao.ResourceUse(c, resArg); err != nil {
		if xecode.EqualError(xecode.Int(model.VipBatchNotEnoughErr), err) {
			err = cardecode.AppCarVipActivityEnd
		} else {
			// 删除db插入那条
			log.Error("%+v", err)
			if err := s.delDB(c, mid); err != nil {
				log.Error("%+v", err)
				return nil, err
			}
		}
		return nil, err
	}
	return s.pgcCardList(c, plat, param.MobiApp, buvid, param.Build, mid)
}

func (s *Service) CodeOpen(c context.Context, plat int8, mid int64, buvid, path, ua string, param *vip.CodeOpenParam) ([]cardm.Handler, error) {
	tec := &view.SilverEventCtx{
		Mid:       mid,
		Buvid:     buvid,
		Ip:        metadata.String(c, metadata.RemoteIP),
		Platform:  param.Platform,
		Ctime:     time.Now().Format("2006-01-02 15:04:05"),
		Api:       path,
		Origin:    param.AppKey,
		UserAgent: ua,
		Build:     strconv.Itoa(param.Build),
		VipEventCtx: &view.VipEventCtx{
			SubScene:    "car_getmemeber",
			ActivityUID: "car_banner_vip",
		},
	}
	if s.silverDao.RuleCheck(c, tec, model.SilverGaiaCommonActivity) {
		return nil, cardecode.AppCarVipRiskUser
	}
	if err := s.resDao.CodeOpen(c, mid, param.Code); err != nil {
		log.Error("%+v", err)
		return nil, cardecode.AppCarVipError
	}
	return s.pgcCardList(c, plat, param.MobiApp, buvid, param.Build, mid)
}

func (s *Service) pgcCardList(c context.Context, plat int8, mobiApp, buvid string, build int, mid int64) ([]cardm.Handler, error) {
	var (
		seasonIds, epids []int32
		seasonm          map[int32]*seasongrpc.CardInfoProto
		epm              map[int32]*episodegrpc.EpisodeCardsProto
	)
	reply, err := s.bgm.Activity(c, mid, buvid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, v := range reply {
		switch v.Type {
		case _pgcSeason:
			seasonIds = append(seasonIds, int32(v.Oid))
		case _pgcEp:
			epids = append(epids, int32(v.Oid))
		}
	}
	group := errgroup.WithContext(c)
	if len(seasonIds) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seasonm, err = s.bgm.Cards(c, seasonIds); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			epm, err = s.bgm.EpCards(ctx, epids)
			if err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	items := []*ai.Item{}
	for _, v := range seasonIds {
		items = append(items, &ai.Item{Goto: model.GotoPGC, ID: int64(v)})
	}
	for _, v := range epids {
		items = append(items, &ai.Item{Goto: model.GotoPGCEp, ID: int64(v)})
	}
	materials := &card.Materials{
		Seams: seasonm,
		Epms:  epm,
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  mobiApp,
		Build:    build,
	}
	op := &operate.Card{FollowType: _followTypeCinema}
	list := s.cardDealItem(cardParam, items, model.EntrancePgcRcmdList, model.VerticalCoverV1, materials, op)
	if len(list) == 0 {
		return []cardm.Handler{}, nil
	}
	return list, nil
}

func (s *Service) saveDB(c context.Context, mid int64, buvid, order string, param *vip.VipParam) error {
	arg := &vip.VipReceived{
		MID:        mid,
		Buvid:      buvid,
		Channel:    param.Channel,
		BatchToken: s.c.VipConfig.BatchToken,
		OrderNo:    order,
		State:      vip.StateUserIsReceived,
	}
	rows, err := s.resDao.InVipReceived(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	if rows == 0 {
		return cardecode.AppCarVipOnlyOnce
	}
	return nil
}

func (s *Service) delDB(c context.Context, mid int64) error {
	rows, err := s.resDao.DelVipReceived(c, mid)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	if rows == 0 {
		return cardecode.AppCarVipRiskUser
	}
	return nil
}

// orderID get order id
func (s *Service) orderID() string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%05d", s.rnd.Int63n(99999)))
	b.WriteString(fmt.Sprintf("%03d", time.Now().UnixNano()/1e6%1000))
	b.WriteString(time.Now().Format("060102150405"))
	return b.String()
}
