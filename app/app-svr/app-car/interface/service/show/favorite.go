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
	toviewgrpc "git.bilibili.co/bapis/bapis-go/community/service/toview"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

const (
	_myfavorite = 1
	_upfavorite = 2
	_topview    = 3
)

var (
	favToBusinessMap = map[int64]string{
		_myfavorite: model.EntranceMyFavorite,
		_upfavorite: model.EntranceUpFavorite,
		_topview:    model.EntranceToView,
	}
)

func (s *Service) Media(c context.Context, plat int8, mid int64, buvid, cookie, referer string, param *favorite.MediaParam) ([]cardm.Handler, error) {
	const (
		_max       = 20
		_defaultPn = 1
		_toViewMax = 1
	)
	items := []cardm.Handler{}
	favs, err := s.fav.FolderSpace(c, param.MobiApp, param.Build, param.AccessKey, cookie, buvid, referer, mid)
	if err != nil {
		log.Error("%+v", err)
		return items, nil
	}
	for _, fav := range favs {
		// 如果收藏夹内一个内容都没有直接不下发
		if fav.Media == nil || fav.Media.Count == 0 {
			continue
		}
		// 1-创建的收藏夹，2-收藏的收藏夹，3-稍后再看
		switch fav.ID {
		case _myfavorite, _upfavorite:
			cards := []cardm.Handler{}
			for _, v := range fav.Media.MediaList {
				r := &ai.Item{Goto: model.GotoFavorite, ID: v.ID / 100}
				op := &operate.Card{}
				op.From(model.CardGt(r.Goto), favToBusinessMap[fav.ID], r.ID, plat, param.Build, param.MobiApp)
				cardGoto := model.CardGotoUserFavorite
				// 0 - 是否公开（1为私密） 1-是否默认(1为非默认)
				if model.AttrVal(int32(v.Attr), 1) == model.AttrNo {
					cardGoto = model.CardGotoDefalutFavorite
				}
				h := cardm.Handle(plat, cardGoto, model.SmallCoverV1, r, nil)
				if h == nil || !h.From(v, op) {
					continue
				}
				cards = append(cards, h)
			}
			if len(cards) == 0 {
				continue
			}
			items = append(items, cards...)
		}
	}
	// nolint:gomnd
	if len(items) > 20 {
		items = items[:_max]
	}
	return items, nil
}

func (s *Service) Favorite(c context.Context, plat int8, mid int64, param *favorite.FavoriteParam) ([]*show.Item, error) {
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
		return []*show.Item{}, nil
	}
	if len(favs) == 0 {
		// 四舍五入向上取整
		pn := param.Pn - int(math.Ceil(float64(count)/_favMax))
		favs, err = s.mediaFavorite(c, mid, pn, _favMax)
		if err != nil {
			log.Error("%+v", err)
			return []*show.Item{}, nil
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
		return []*show.Item{}, nil
	}
	materials := &card.Materials{
		Arcs: arcs,
		Epms: seams,
	}
	items := []*show.Item{}
	for _, v := range favs {
		ms, ok := favoritem[v.ID]
		if !ok {
			continue
		}
		var rs []*ai.Item
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
			rs = append(rs, &ai.Item{Goto: gt, ID: v.Oid})
		}
		is := s.dealItemMedia(plat, param.MobiApp, param.Build, rs, model.FromList, model.EntranceMediaList, v.ID, v.Mid, materials)
		if len(is) == 0 {
			continue
		}
		favType := model.CardGotoUserFavorite
		// 0 - 是否公开（1为私密） 1-是否默认(1为非默认)
		if model.AttrVal(int32(v.Attr), 1) == model.AttrNo {
			favType = model.CardGotoDefalutFavorite
		}
		item := &show.Item{Type: string(favType), Items: is, Title: v.Name}
		item.FromFavItem(v.ID, v.Mid)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) myFavorite(c context.Context, mid int64, pn, ps int) ([]*favoritemodel.Folder, int, error) {
	const (
		// 我收藏的视频
		_favVedio = 2
	)
	favs, err := s.fav.UserFolders(c, mid, 0, _favVedio)
	if err != nil {
		log.Error("%+v", err)
		return nil, 0, err
	}
	count := len(favs)
	start := (pn - 1) * ps
	end := start + ps
	if end < len(favs) {
		favs = favs[start:end]
	} else if start < len(favs) {
		favs = favs[start:]
	} else {
		favs = []*favoritemodel.Folder{}
	}
	return favs, count, nil
}

// mediaFavorite 我收藏的收藏夹
func (s *Service) mediaFavorite(c context.Context, mid int64, pn, ps int) ([]*favoritemodel.Folder, error) {
	const (
		// 我收藏的视频
		_favVedio = 2
		// 我收藏的收藏夹
		_mediaList = 11
	)
	var (
		favids []int64
	)
	reply, err := s.fav.FavoritesAll(c, mid, mid, 0, _mediaList, pn, ps)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, v := range reply.GetList() {
		favids = append(favids, v.Oid)
	}
	mediafavs, err := s.fav.Folders(c, favids, _favVedio, mid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return mediafavs, nil
}

func (s *Service) MediaList(c context.Context, plat int8, mid int64, param *favorite.MediaListParam) ([]cardm.Handler, error) {
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
		return []cardm.Handler{}, nil
	}
	var (
		aids  []int64
		epids []int32
		arcs  map[int64]*arcgrpc.Arc
		seams map[int32]*episodegrpc.EpisodeCardsProto
		rs    []*ai.Item
	)
	// 插入逻辑
	if param.Pn == 1 && param.ParamStr != "" {
		gt, id, _, ok := cardm.FromGtPrune(param.ParamStr)
		if ok {
			var isok bool
			for _, v := range medias.GetList() {
				if v.Oid == id {
					isok = true
					break
				}
			}
			if !isok {
				favType := _vedio
				if gt == model.GotoPGC {
					favType = _pgc
				}
				cards := []*favoritegrpc.ModelFavorite{{Oid: id, Type: int32(favType)}}
				cards = append(cards, medias.GetList()...)
				medias.List = cards
			}
		}
	}
	for _, v := range medias.GetList() {
		var gt string
		if v.Oid == 0 {
			continue
		}
		switch v.Type {
		case _vedio:
			aids = append(aids, v.Oid)
			gt = model.GotoAv
		case _pgc:
			epids = append(epids, int32(v.Oid))
			gt = model.GotoPGC
		default:
			continue
		}
		rs = append(rs, &ai.Item{Goto: gt, ID: v.Oid})
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
		return []cardm.Handler{}, nil
	}
	materials := &card.Materials{
		Arcs: arcs,
		Epms: seams,
	}
	return s.dealItemMedia(plat, param.MobiApp, param.Build, rs, param.FromType, model.EntranceMediaList, param.FavID, param.Vmid, materials), nil
}

func (s *Service) ToView(c context.Context, plat int8, mid int64, param *favorite.ToViewParam) ([]cardm.Handler, error) {
	toViews, err := s.fav.UserToViews(c, mid, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}
	var (
		aids []int64
		arcs map[int64]*arcgrpc.Arc
		rs   []*ai.Item
	)
	// 插入逻辑
	if param.Pn == 1 && param.ParamStr != "" {
		_, id, _, ok := cardm.FromGtPrune(param.ParamStr)
		if ok {
			var isok bool
			for _, v := range toViews {
				if v.Aid == id {
					isok = true
					break
				}
			}
			if !isok {
				cards := []*toviewgrpc.ToView{{Aid: id}}
				cards = append(cards, toViews...)
				toViews = cards
			}
		}
	}
	for _, v := range toViews {
		if v.Aid != 0 {
			aids = append(aids, v.Aid)
			rs = append(rs, &ai.Item{Goto: model.GotoAv, ID: v.Aid})
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
	return s.dealItemMedia(plat, param.MobiApp, param.Build, rs, param.FromType, model.EntranceToView, 0, 0, materials), nil
}

func (s *Service) dealItemMedia(plat int8, mobiApp string, build int, rs []*ai.Item, fromType, entrance string, favID, vmid int64, materials *card.Materials) []cardm.Handler {
	is := []cardm.Handler{}
	// 兜底卡片
	backupCard := []cardm.Handler{}
	for _, r := range rs {
		var (
			cardType model.CardType
			op       = &operate.Card{FavID: favID, Vmid: vmid}
			main     interface{}
		)
		op.From(model.CardGt(r.Goto), entrance, r.ID, plat, build, mobiApp)
		switch r.Goto {
		case model.GotoAv:
			main = materials.Arcs
		case model.GotoPGC:
			main = materials.Epms
		}
		switch fromType {
		case model.FromList:
			cardType = model.SmallCoverV1
		default:
			cardType = model.SmallCoverV4
		}
		materials := &cardm.Materials{
			Prune: cardm.GtPrune(r.Goto, r.ID),
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 找出互动视频卡片，并把第一张放入兜底卡片里面，并且长度为0的时候放入
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			if len(backupCard) == 0 {
				backupCard = append(backupCard, h)
			}
			// 互动视频不放入列表里面
			continue
		}
		is = append(is, h)
	}
	// 如果过滤完当前列表一个都没有，放入兜底卡片数据
	if len(is) == 0 {
		is = backupCard
	}
	return is
}

func (s *Service) FavAddOrDelFolders(c context.Context, mid int64, param *favorite.FavAddOrDelFolders) error {
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
			if err := s.fav.FavAddFolders(c, param.MobiApp, param.Device, param.Platform, addFids, param.Aid, mid); err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
	}
	if len(delFids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if err := s.fav.FavDelFolders(c, param.MobiApp, param.Device, param.Platform, delFids, param.Aid, mid); err != nil {
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

func (s *Service) AddFolder(c context.Context, mid int64, param *favorite.AddFolder) (int64, error) {
	fid, err := s.fav.AddFolder(c, mid, param.Name, param.Desc, param.Public)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return fid, nil
}

func (s *Service) UserFolders(c context.Context, mid int64, param *favorite.UserFolderParam) ([]*favorite.UserFolder, error) {
	const (
		// 我收藏的视频
		_favVedio = 2
	)
	if mid == 0 {
		return []*favorite.UserFolder{}, nil
	}
	favs, err := s.fav.UserFolders(c, mid, param.Aid, _favVedio)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var res []*favorite.UserFolder
	for _, v := range favs {
		folder := &favorite.UserFolder{}
		folder.FromUserFolder(v)
		res = append(res, folder)
	}
	return res, nil
}
