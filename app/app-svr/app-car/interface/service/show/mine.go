package show

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	"go-gateway/app/app-svr/app-car/interface/model/history"
	"go-gateway/app/app-svr/app-car/interface/model/mine"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"go-common/library/sync/errgroup.v2"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_dynamicVideo = "dynamic_video"
	_history      = "history"
	_favorite     = "my_favorite"
	_topView      = "top_view"
)

func (s *Service) Mine(c context.Context, mid int64, plat int8, buvid string, param *mine.MineParam) (res []*show.Item, account *mine.Mine, err error) {
	var (
		cardsTypes = []string{_dynamicVideo, _history, _topView, _favorite}
		cards      = map[string][]cardm.Handler{}
		mine       *mine.Mine
		mutex      sync.Mutex
	)
	group := errgroup.WithContext(c)
	if mid > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if mine, err = s.userInfo(ctx, mid); err != nil {
				log.Error("%+v", err)
				return err
			}
			if mine != nil {
				mine.Mid = 0
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			dParam := &dynamic.DynamicParam{
				FromType:       model.FromList,
				DeviceInfo:     param.DeviceInfo,
				LocalTime:      8,
				UpdateBaseline: "0",
				Page:           1,
				AssistBaseline: "20",
			}
			item, _, err := s.DynVideo(ctx, plat, mid, buvid, dParam)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if len(item) > 0 {
				mutex.Lock()
				cards[_dynamicVideo] = item
				mutex.Unlock()
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) (err error) {
		hParam := &history.HisParam{FromType: model.FromList, DeviceInfo: param.DeviceInfo}
		item, _, err := s.Cursor(ctx, plat, mid, buvid, hParam)
		if err != nil {
			log.Error("%+v", err)
			return err
		}
		if len(item) > 0 {
			mutex.Lock()
			cards[_history] = item
			mutex.Unlock()
		}
		return nil
	})
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.Media, &feature.OriginResutl{
		MobiApp:    param.MobiApp,
		Device:     param.Device,
		Build:      int64(param.Build),
		BuildLimit: param.Build >= 1010000,
	}) && mid > 0 {
		group.Go(func(ctx context.Context) (err error) {
			item, err := s.Media(ctx, plat, mid, buvid, "", "", &favorite.MediaParam{DeviceInfo: param.DeviceInfo})
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if len(item) > 0 {
				mutex.Lock()
				cards[_favorite] = item
				mutex.Unlock()
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			item, err := s.ToView(ctx, plat, mid, &favorite.ToViewParam{Pn: _defaultPn, Ps: _defaultPs, FromType: model.FromList})
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if len(item) > 0 {
				mutex.Lock()
				cards[_topView] = item
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	showItems := []*show.Item{}
	for _, ct := range cardsTypes {
		card, ok := cards[ct]
		if !ok {
			continue
		}
		item := &show.Item{Type: ct, Items: card}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.MineTag, &feature.OriginResutl{
			MobiApp:    param.MobiApp,
			Device:     param.Device,
			Build:      int64(param.Build),
			BuildLimit: param.Build >= 1010000,
		}) {
			item.FromItem2("")
		} else {
			item.FromItem()
		}
		showItems = append(showItems, item)
	}
	return showItems, mine, nil
}

func (s *Service) userInfo(c context.Context, mid int64) (*mine.Mine, error) {
	var (
		ps *accountgrpc.Profile
	)
	ps, err := s.acc.Profile3(c, mid)
	if err != nil {
		log.Error("s.acc.Profile3(%d) error(%v)", mid, err)
		return nil, err
	}
	account := &mine.Mine{}
	account.FromMine(ps)
	return account, nil
}
