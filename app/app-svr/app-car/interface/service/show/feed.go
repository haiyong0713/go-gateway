package show

import (
	"context"
	"fmt"
	"hash/crc32"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/banner"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
)

func (s *Service) group(mid int64, buvid string) (group int) {
	if mid == 0 && buvid == "" {
		group = -1
		return
	}
	if mid != 0 {
		group = int(mid % 20)
		return
	}
	group = int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
	return
}

func (s *Service) Feed(c context.Context, mid int64, plat int8, buvid string, param *banner.ShowBannerParam) []cardm.Handler {
	return s.feed(c, plat, mid, buvid, param, model.SmallCoverV4)
}

func (s *Service) feed(c context.Context, plat int8, mid int64, buvid string, param *banner.ShowBannerParam, cardType model.CardType) []cardm.Handler {
	feedgroup := s.group(mid, buvid)
	feedList, err := s.rcmd.FeedRecommend(c, plat, param.MobiApp, buvid, mid, param.Build, 0, feedgroup, _max, 0)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}
	}
	cards := []*ai.Item{}
	for _, v := range feedList {
		item := &ai.Item{}
		*item = *v
		item.Entrance = model.EntranceCommonSearch
		cards = append(cards, item)
	}
	var (
		aids  []int64
		epids []int32
		arcs  map[int64]*arcgrpc.ArcPlayer
		epm   map[int32]*pgcinline.EpisodeCard
	)
	for _, v := range cards {
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			epids = append(epids, int32(v.ID))
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.ArcsPlayerAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			seams, err := s.bgm.CardsByAidsAll(ctx, aids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, ep := range seams {
				epids = append(epids, ep.EpisodeId)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if len(epids) > 0 {
		if epm, err = s.bgm.InlineCardsAll(c, epids, param.MobiApp, param.Platform, param.Device, param.Build); err != nil {
			log.Error("%+v", err)
		}
	}
	materials := &card.Materials{
		ArcPlayers: arcs,
		EpInlinem:  epm,
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
		IsPlayer: true,
	}
	op := &operate.Card{}
	list := s.cardDealItem(cardParam, cards, model.EntranceCommonSearch, cardType, materials, op)
	if len(list) == 0 {
		return []cardm.Handler{}
	}
	return list
}
