package show

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	"go-gateway/app/app-svr/app-car/interface/model/history"
	"go-gateway/app/app-svr/app-car/interface/model/mine"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

// nolint: gocognit
func (s *Service) MineWeb(c context.Context, mid int64, plat int8, buvid, cookie, referer string, param *mine.MineParam) (res []*show.ItemWeb, account *mine.MineWeb, err error) {
	var (
		cardsTypes      = []string{_dynamicVideo, _history, _topView, _favorite}
		mutex           sync.Mutex
		aids, hisAids   []int64
		epids           []int32
		arcs            map[int64]*arcgrpc.Arc
		arcViews        map[int64]*arcgrpc.ViewReply
		seams, seamsAid map[int32]*episodegrpc.EpisodeCardsProto
		mineinfo        *mine.MineWeb
		medias          []cardm.Handler
	)
	if tabs, ok := s.c.Custom.MineWebTab[param.Env]; ok {
		cardsTypes = tabs
	}
	listm := map[string][]*ai.Item{}
	deviceInfo := model.DeviceInfo{
		MobiApp:  model.AndroidBilithings,
		Platform: "android",
	}
	group := errgroup.WithContext(c)
	if mid > 0 {
		// 用户信息
		group.Go(func(ctx context.Context) (err error) {
			if mineinfo, err = s.userInfoWeb(ctx, mid); err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
		// 动态
		group.Go(func(ctx context.Context) (err error) {
			dParam := &dynamic.DynamicParam{
				FromType:       model.FromList,
				DeviceInfo:     deviceInfo,
				LocalTime:      8,
				UpdateBaseline: "0",
				Page:           1,
				AssistBaseline: "20",
			}
			item, _, err := s.dynVideoList(ctx, mid, buvid, dParam)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if len(item) > 0 {
				mutex.Lock()
				listm[_dynamicVideo] = item
				mutex.Unlock()
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if medias, err = s.Media(ctx, model.PlatH5, mid, buvid, cookie, referer, &favorite.MediaParam{DeviceInfo: deviceInfo}); err != nil {
				if err != nil {
					log.Error("%+v", err)
					return err
				}
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			toview, err := s.fav.UserToViews(c, mid, _defaultPn, _defaultPs)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			var items []*ai.Item
			for _, v := range toview {
				if v.Aid == 0 {
					continue
				}
				items = append(items, &ai.Item{Goto: model.GotoAv, ID: v.Aid})
			}
			if len(items) > 0 {
				mutex.Lock()
				listm[_topView] = items
				mutex.Unlock()
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) (err error) {
		hParam := &history.HisParam{FromType: model.FromList, DeviceInfo: deviceInfo}
		item, _, err := s.cursorList(ctx, mid, buvid, hParam)
		if err != nil {
			log.Error("%+v", err)
			return err
		}
		if len(item) > 0 {
			mutex.Lock()
			listm[_history] = item
			mutex.Unlock()
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, list := range listm {
		for _, v := range list {
			switch v.Goto {
			case model.GotoAv:
				aids = append(aids, v.ID)
			case model.GotoAvHis:
				hisAids = append(hisAids, v.ID)
			case model.GotoPGCEp:
				epids = append(epids, int32(v.ID))
			case model.GotoPGCEpHis:
				hisAids = append(hisAids, v.ID)
				epids = append(epids, int32(v.ChildID))
			}
		}
	}
	// 第二次批量
	group = errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.ArcsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seamsAid, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(hisAids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcViews, err = s.arc.Views(ctx, hisAids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.EpCards(ctx, epids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	list := cardsTypes
	materials := &cardm.Materials{
		Arcs:               arcs,
		ViewReplym:         arcViews,
		EpisodeCardsProtom: seamsAid,
		Epms:               seams,
	}
	items := []*show.ItemWeb{}
	for _, ct := range list {
		var (
			entrance string
			cardType model.CardType
			list     []cardm.Handler
		)
		op := &operate.Card{}
		if ct == _favorite && len(medias) > 0 {
			list = medias
		} else {
			cards, ok := listm[ct]
			if !ok {
				continue
			}
			cardParam := &cardm.CardParam{
				Plat:         plat,
				Mid:          mid,
				FromType:     model.FromList,
				IsBackUpCard: true,
			}
			switch ct {
			case _dynamicVideo:
				entrance = model.EntranceDynamicVideo
				cardType = model.SmallCoverV1
			case _history:
				entrance = model.EntranceHistoryRecord
				cardType = model.SmallCoverV3
			case _topView:
				entrance = model.EntranceToView
				cardType = model.SmallCoverV1
			case _favorite:
				cardType = model.SmallCoverV1
				entrance = model.EntranceMyFavorite
			}
			list = s.cardDealWebItem(cardParam, cards, entrance, cardType, materials, op)
			if len(list) == 0 {
				continue
			}
		}
		item := &show.ItemWeb{Type: ct, Items: list}
		item.FromItemWeb(entrance, op.Rid)
		items = append(items, item)
	}
	if mineinfo == nil {
		mineinfo = &mine.MineWeb{}
	}
	return items, mineinfo, nil
}

func (s *Service) userInfoWeb(c context.Context, mid int64) (*mine.MineWeb, error) {
	var (
		ps *accountgrpc.Profile
	)
	ps, err := s.acc.Profile3(c, mid)
	if err != nil {
		log.Error("s.acc.Profile3(%d) error(%v)", mid, err)
		return nil, err
	}
	account := &mine.MineWeb{}
	account.FromMineWeb(ps)
	return account, nil
}
