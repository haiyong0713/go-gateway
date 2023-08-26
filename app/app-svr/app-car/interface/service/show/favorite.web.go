package show

import (
	"context"
	"math"
	"strconv"
	"strings"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	favoritemodel "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favoritegrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) FavAddOrDelFoldersWeb(c context.Context, mid int64, param *favorite.FavAddOrDelFolders) error {
	addFidStrs := strings.Split(param.AddFids, ",")
	delFidStrs := strings.Split(param.DelFids, ",")
	var (
		addFids []int64
		delFids []int64
	)
	for _, v := range addFidStrs {
		fid, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		addFids = append(addFids, fid)
	}
	for _, v := range delFidStrs {
		fid, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		delFids = append(delFids, fid)
	}
	group := errgroup.WithContext(c)
	if len(addFids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if err := s.fav.FavAddFolders(c, model.AndroidBilithings, param.Device, param.Platform, addFids, param.Oid, mid); err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
	}
	if len(delFids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if err := s.fav.FavDelFolders(c, model.AndroidBilithings, param.Device, param.Platform, delFids, param.Oid, mid); err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%v", err)
		return err
	}
	return nil
}

func (s *Service) AddFolderWeb(c context.Context, mid int64, param *favorite.AddFolder) (int64, error) {
	fid, err := s.fav.AddFolder(c, mid, param.Name, param.Desc, param.Public)
	if err != nil {
		log.Error("%+v", err)
		return 0, nil
	}
	return fid, nil
}

func (s *Service) FavoriteWeb(c context.Context, mid int64, param *favorite.FavoriteParam) ([]*show.ItemWeb, error) {
	const (
		// 一页最多3个收藏夹
		_favMax    = 3
		_defaultPn = 1
		// 一个收藏夹最多20个视频数据
		_max = 10
		// 我收藏的视频
		_favVedio = 2
		_pgc      = 24
	)
	var (
		mutex   sync.Mutex
		favAids []int64
		epids   []int32
		arcs    map[int64]*arcgrpc.Arc
		seams   map[int32]*episodegrpc.EpisodeCardsProto
	)
	favoritem := map[int64]*favoritegrpc.ModelFavorites{}
	favs, count, err := s.myFavorite(c, mid, param.Pn, _favMax)
	if err != nil {
		log.Error("%+v", err)
		return []*show.ItemWeb{}, nil
	}
	if len(favs) == 0 {
		// 四舍五入向上取整
		pn := param.Pn - int(math.Ceil(float64(count)/_favMax))
		favs, err = s.mediaFavorite(c, mid, pn, _favMax)
		if err != nil {
			log.Error("%+v", err)
			return []*show.ItemWeb{}, nil
		}
	}
	group := errgroup.WithContext(c)
	for _, v := range favs {
		tmp := &favoritemodel.Folder{}
		*tmp = *v
		group.Go(func(ctx context.Context) error {
			reply, err := s.fav.FavoritesAll(ctx, mid, tmp.Mid, tmp.ID, _favVedio, _defaultPn, _max)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			for _, v := range reply.GetList() {
				if v.Oid == 0 {
					continue
				}
				switch v.Type {
				case _favVedio:
					favAids = append(favAids, v.Oid)
				case _pgc:
					epids = append(epids, int32(v.Oid))
				default:
					continue
				}
			}
			mutex.Lock()
			favoritem[tmp.ID] = reply
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%v", err)
	}
	if len(favAids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, favAids); err != nil {
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
		return []*show.ItemWeb{}, nil
	}
	materials := &card.Materials{
		Arcs: arcs,
		Epms: seams,
	}
	items := []*show.ItemWeb{}
	// 格式转换
	for _, v := range favs {
		ms, ok := favoritem[v.ID]
		if !ok {
			continue
		}
		var cards []*ai.Item
		for _, v := range ms.GetList() {
			var gt string
			if v.Oid == 0 {
				continue
			}
			switch v.Type {
			case _favVedio:
				gt = model.GotoAv
			case _pgc:
				gt = model.GotoPGC
			default:
				continue
			}
			cards = append(cards, &ai.Item{Goto: gt, ID: v.Oid})
		}
		op := &operate.Card{FavID: v.ID, Vmid: v.Mid}
		cardParam := &cardm.CardParam{
			Plat:     model.PlatH5,
			Mid:      mid,
			FromType: model.FromList,
		}
		list := s.cardDealWebItem(cardParam, cards, model.EntranceMediaList, model.SmallCoverV1, materials, op)
		if len(list) == 0 {
			continue
		}
		favType := model.CardGotoUserFavorite
		// 0 - 是否公开（1为私密） 1-是否默认(1为非默认)
		if model.AttrVal(int32(v.Attr), 1) == model.AttrNo {
			favType = model.CardGotoDefalutFavorite
		}
		item := &show.ItemWeb{Type: string(favType), Items: list, Title: v.Name}
		item.FromFavItemWeb(v.ID, v.Mid, model.EntranceMediaList)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) MediaListWeb(c context.Context, plat int8, mid int64, param *favorite.MediaListParam) ([]cardm.Handler, *cardm.Page, error) {
	const (
		// 我收藏的视频
		_favVedio = 2
		_max      = 20
		_vedio    = 2
		_pgc      = 24
	)
	medias, err := s.fav.FavoritesAll(c, mid, param.Vmid, param.FavID, _favVedio, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil
	}
	var (
		aids  []int64
		epids []int32
		arcs  map[int64]*arcgrpc.Arc
		seams map[int32]*episodegrpc.EpisodeCardsProto
		cards []*ai.Item
	)
	for _, v := range medias.GetList() {
		var gt string
		if v.Oid == 0 {
			continue
		}
		switch v.Type {
		case _vedio:
			gt = model.GotoAv
		case _pgc:
			gt = model.GotoPGC
		default:
			continue
		}
		cards = append(cards, &ai.Item{Goto: gt, ID: v.Oid})
	}
	// 插入逻辑
	var pn int
	if param.Pn == 1 {
		pn = 1
	}
	cards = s.listInsert(cards, pn, param.ParamStr)
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
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
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
		return []cardm.Handler{}, nil, nil
	}
	materials := &card.Materials{
		Arcs: arcs,
		Epms: seams,
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromView,
	}
	op := &operate.Card{}
	list := s.cardDealWebItem(cardParam, cards, model.EntrancePopular, model.SmallCoverV4, materials, op)
	itemPage := &cardm.Page{}
	if len(list) > 0 {
		itemPage.Position = list[len(list)-1].Get().Pos
	}
	return list, itemPage, nil
}

func (s *Service) ToViewWeb(c context.Context, mid int64, param *favorite.ToViewParam) ([]cardm.Handler, error) {
	toViews, err := s.fav.UserToViews(c, mid, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}
	var (
		aids  []int64
		arcs  map[int64]*arcgrpc.Arc
		cards []*ai.Item
	)
	for _, v := range toViews {
		if v.Aid == 0 {
			continue
		}
		cards = append(cards, &ai.Item{Goto: model.GotoAv, ID: v.Aid})
	}
	// 插入逻辑
	var pn int
	if param.Pn == 1 {
		pn = 1
	}
	cards = s.listInsert(cards, pn, param.ParamStr)
	for _, v := range cards {
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		}
	}
	if len(aids) > 0 {
		if arcs, err = s.arc.Archives(c, aids); err != nil {
			log.Error("%+v", err)
			return []cardm.Handler{}, nil
		}
	}
	materials := &card.Materials{
		Arcs: arcs,
	}
	cardParam := &cardm.CardParam{
		Plat:     model.PlatH5,
		Mid:      mid,
		FromType: model.FromView,
	}
	op := &operate.Card{}
	list := s.cardDealWebItem(cardParam, cards, model.EntranceToView, model.SmallCoverV4, materials, op)
	return list, nil
}
