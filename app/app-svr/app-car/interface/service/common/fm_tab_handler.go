package common

import (
	"context"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	archive "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	fmRec "git.bilibili.co/bapis/bapis-go/ott-recommend/automotive-channel"

	"github.com/pkg/errors"
)

var (
	tabItemsHandlerMap map[fm_v2.FmType]TabItemsHandler // 处理接口map
)

func initFmTabHandler(s *Service) {
	tabItemsHandlerMap = map[fm_v2.FmType]TabItemsHandler{
		fm_v2.AudioFeed:     &TabItemsFeed{s: s},
		fm_v2.AudioUp:       &TabItemsUp{s: s},
		fm_v2.AudioRelate:   &TabItemsRelate{s: s},
		fm_v2.AudioHome:     &TabItemsHome{s: s},
		fm_v2.AudioVertical: &TabItemsVertical{s: s},
		fm_v2.AudioSeason:   &TabItemsSeason{s: s},
		fm_v2.AudioSeasonUp: &TabItemsSeasonUp{s: s},
	}
}

func TabItemsStrategy(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	if _, ok := tabItemsHandlerMap[req.FmType]; !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "FmType Illegal: %s", req.FmType)
	}
	return tabItemsHandlerMap[req.FmType].HandleTabItems(c, req)
}

type TabItemsHandler interface {
	// HandleTabItems 根据FM播单类型，生成首页播单卡片
	HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error)
}

type TabItemsFeed struct {
	s *Service
}

type TabItemsUp struct {
	s *Service
}

type TabItemsRelate struct {
	s *Service
}

type TabItemsHome struct {
	s *Service
}

type TabItemsVertical struct {
	s *Service
}

type TabItemsSeason struct {
	s *Service
}

type TabItemsSeasonUp struct {
	s *Service
}

func (t *TabItemsFeed) HandleTabItems(_ context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	if req.FmType != fm_v2.AudioFeed || t.s.c.Custom == nil || len(t.s.c.Custom.FmTabConfigs) == 0 {
		return new(fm_v2.HandleTabItemsResp), nil
	}
	for _, v := range t.s.c.Custom.FmTabConfigs {
		if v.FmType == string(fm_v2.AudioFeed) {
			tabItem := &fm_v2.TabItem{
				FmType: fm_v2.AudioFeed,
				Title:  v.Title,
				Cover:  v.Cover,
				Style:  fm_v2.DefaultStyle,
			}
			return &fm_v2.HandleTabItemsResp{
				TabItems: []*fm_v2.TabItem{tabItem},
			}, nil
		}
	}
	return nil, ecode.NothingFound
}

func (t *TabItemsUp) HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	var (
		upMid   = req.FmId
		profile *accountgrpc.Profile
		tabItem *fm_v2.TabItem
	)
	profile, err = t.s.accountDao.Profile3(c, upMid)
	if err != nil {
		return nil, err
	}
	tabItem = &fm_v2.TabItem{
		FmType: fm_v2.AudioUp,
		FmId:   upMid,
		Title:  profile.Name,
		Cover:  profile.Face,
		Style:  fm_v2.CircleStyle,
	}
	return &fm_v2.HandleTabItemsResp{
		TabItems: []*fm_v2.TabItem{tabItem},
	}, nil
}

func (t *TabItemsRelate) HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	var (
		aid     = req.FmId
		arc     *archive.Arc
		tabItem *fm_v2.TabItem
	)
	arc, err = t.s.archiveDao.Arc(c, aid)
	if err != nil {
		return nil, err
	}
	tabItem = &fm_v2.TabItem{
		FmType: fm_v2.AudioRelate,
		FmId:   aid,
		Title:  arc.Title,
		Cover:  arc.Pic,
		Style:  fm_v2.RectangleStyle,
	}
	return &fm_v2.HandleTabItemsResp{
		TabItems: []*fm_v2.TabItem{tabItem},
	}, nil
}

func (t *TabItemsHome) HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	if t.s.c.Custom == nil || len(t.s.c.Custom.FmTabConfigs) == 0 {
		return new(fm_v2.HandleTabItemsResp), nil
	}
	var (
		defaultVerticals = make([]*fm_v2.TabItem, 0)
		verticalMap      = make(map[int64]*fm_v2.TabItem)
		defaultFeed      *fm_v2.TabItem
		recResp          *fm_v2.RecResp
		resItems         []*fm_v2.TabItem
	)
	// 1. 预构建垂类卡片
	for _, v := range t.s.c.Custom.FmTabConfigs {
		if v.Title == "" || v.Cover == "" {
			continue
		}
		item := &fm_v2.TabItem{
			FmType:   fm_v2.FmType(v.FmType),
			FmId:     v.FmId,
			Title:    v.Title,
			SubTitle: v.SubTitle,
			Cover:    v.Cover,
			Style:    fm_v2.TabStyle(v.Style),
		}
		if v.FmType == string(fm_v2.AudioVertical) {
			defaultVerticals = append(defaultVerticals, item)
			verticalMap[v.FmId] = item
		}
		if v.FmType == string(fm_v2.AudioFeed) {
			defaultFeed = item
		}
	}
	// 2. 调用个性化排序
	recResp, err = t.s.fmDao.FmHome(c, req.Mid, req.Buvid, req.DeviceInfo, req.PageReq)
	if err != nil {
		// 降级为默认排序的垂类列表
		log.Errorc(c, "TabItemsHome HandleTabItems t.s.fmDao.FmHome error:%+v, req:%+v", err, req)
		return &fm_v2.HandleTabItemsResp{TabItems: append([]*fm_v2.TabItem{defaultFeed}, defaultVerticals...)}, nil
	}
	// 3. 构建首页卡片
	resItems, err = t.generateTabItems(c, req, recResp, verticalMap, defaultFeed)
	if err != nil {
		// 降级为默认排序的垂类列表
		log.Errorc(c, "TabItemsHome HandleTabItems t.generateTabItems err:%+v, req:%+v, recResp:%s, verticalMap:%s, defaultFeed:%+v", err, req, toJson(recResp), toJson(verticalMap), defaultFeed)
		return &fm_v2.HandleTabItemsResp{TabItems: append([]*fm_v2.TabItem{defaultFeed}, defaultVerticals...)}, nil
	}
	return &fm_v2.HandleTabItemsResp{TabItems: resItems, PageResp: recResp.PageResp}, nil
}

func (t *TabItemsHome) generateTabItems(c context.Context, req *fm_v2.HandleTabItemsReq, recResp *fm_v2.RecResp, verticalMap map[int64]*fm_v2.TabItem, defaultFeed *fm_v2.TabItem) ([]*fm_v2.TabItem, error) {
	var (
		tabMap sync.Map
		err    error
		res    = make([]*fm_v2.TabItem, 0)
		eg     = errgroup.WithContext(c)
	)

	for _i, _v := range recResp.Items {
		i, v := _i, _v
		eg.Go(func(c context.Context) error {
			var item *fm_v2.TabItem

			switch v.FmType {
			case fmRec.FMType_FM_TYPE_AUDIO_VERTICAL:
				item = verticalMap[v.Id]
				if item == nil {
					return nil
				}
			case fmRec.FMType_FM_TYPE_AUDIO_SEASON:
				tab, localErr := t.s.fmSeasonTab(c, &fm_v2.HandleTabItemsReq{
					DeviceInfo: req.DeviceInfo,
					Mid:        req.Mid,
					Buvid:      req.Buvid,
					FmType:     fm_v2.AudioSeason,
					FmId:       v.Id,
				})
				if localErr != nil {
					log.Errorc(c, "TabItemsHome HandleTabItems t.s.fmSeasonTab AudioSeason error:%+v, infoReq:%+v", localErr, req)
					return nil
				}
				if tab == nil || len(tab.TabItems) == 0 {
					log.Warnc(c, "TabItemsHome HandleTabItems AudioSeason tab nil:%+v, req:%+v", tab, req)
					return nil
				}
				item = tab.TabItems[0]
			case fmRec.FMType_FM_TYPE_AUDIO_SEASON_UP:
				tab, localErr := t.s.fmSeasonTab(c, &fm_v2.HandleTabItemsReq{
					DeviceInfo: req.DeviceInfo,
					Mid:        req.Mid,
					Buvid:      req.Buvid,
					FmType:     fm_v2.AudioSeasonUp,
					FmId:       v.Id,
				})
				if localErr != nil {
					log.Errorc(c, "TabItemsHome HandleTabItems t.s.fmSeasonTab AudioSeasonUp error:%+v, infoReq:%+v", localErr, req)
					return nil
				}
				if tab == nil || len(tab.TabItems) == 0 {
					log.Warnc(c, "TabItemsHome HandleTabItems AudioSeason tab nil:%+v, req:%+v", tab, req)
					return nil
				}
				item = tab.TabItems[0]
			case fmRec.FMType_FM_TYPE_AUDIO_FOR_YOU_FEEDS:
				item = defaultFeed
			default:
				log.Error("TabItemsHome HandleTabItems s.fmDao.GetSeasonInfo unknown fmType:%d", v.FmType)
				return nil
			}

			item.ServerInfo = v.ServerInfo
			tabMap.Store(i, item)
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	// tabMap排序，转为slice
	for i := range recResp.Items {
		if item, ok := tabMap.Load(i); ok {
			res = append(res, item.(*fm_v2.TabItem))
		}
	}
	return res, nil
}

func (t *TabItemsVertical) HandleTabItems(_ context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	if req.FmType != fm_v2.AudioVertical || t.s.c.Custom == nil || len(t.s.c.Custom.FmTabConfigs) == 0 {
		return new(fm_v2.HandleTabItemsResp), nil
	}
	for _, v := range t.s.c.Custom.FmTabConfigs {
		if v.FmType == string(fm_v2.AudioVertical) && v.FmId == req.FmId {
			tabItem := &fm_v2.TabItem{
				FmType:   fm_v2.AudioVertical,
				FmId:     v.FmId,
				Title:    v.Title,
				SubTitle: v.SubTitle,
				Cover:    v.Cover,
				Style:    fm_v2.DefaultStyle,
			}
			return &fm_v2.HandleTabItemsResp{
				TabItems: []*fm_v2.TabItem{tabItem},
			}, nil
		}
	}
	return nil, ecode.NothingFound
}

func (t *TabItemsSeason) HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	return t.s.fmSeasonTab(c, req)
}

func (t *TabItemsSeasonUp) HandleTabItems(c context.Context, req *fm_v2.HandleTabItemsReq) (resp *fm_v2.HandleTabItemsResp, err error) {
	return t.s.fmSeasonTab(c, req)
}

func (s *Service) fmSeasonTab(ctx context.Context, req *fm_v2.HandleTabItemsReq) (*fm_v2.HandleTabItemsResp, error) {
	var (
		info *fm_v2.SeasonInfoResp
		arc  *archive.Arc
	)

	if req.FmType != fm_v2.AudioSeason && req.FmType != fm_v2.AudioSeasonUp {
		return nil, ecode.RequestErr
	}

	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var (
			localErr error
			localReq = fm_v2.SeasonInfoReq{Scene: fm_v2.SceneFm, FmType: req.FmType, SeasonId: req.FmId}
		)
		info, localErr = s.fmDao.GetSeasonInfo(ctx, localReq)
		if localErr != nil {
			log.Errorc(ctx, "fmSeasonTab s.fmDao.GetSeasonInfo err:%+v, req:%+v", localErr, localReq)
			return localErr
		}
		if info.Fm == nil || info.Fm.Title == "" || info.Fm.Cover == "" {
			log.Warnc(ctx, "fmSeasonTab s.fmDao.GetSeasonInfo incomplete season info:%+v", info)
			return ecode.NothingFound
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		// 获取合集中首个稿件
		var (
			oids     []int64
			localErr error
			localReq = fm_v2.SeasonOidReq{Scene: fm_v2.SceneFm, FmType: req.FmType, SeasonId: req.FmId, Ps: 1}
		)
		oids, _, localErr = s.fmDao.GetSeasonOid(ctx, localReq)
		if localErr != nil {
			log.Errorc(ctx, "fmSeasonTab s.fmDao.GetSeasonOid err:%+v, req:%+v", localErr, localReq)
			return localErr
		}
		if len(oids) == 0 {
			log.Warnc(ctx, "fmSeasonTab s.fmDao.GetSeasonOid no oid, req:%+v", localReq)
			return ecode.NothingFound
		}

		// 获取首个稿件的标题
		arc, localErr = s.archiveDao.Arc(ctx, oids[0])
		if localErr != nil {
			log.Errorc(ctx, "fmSeasonTab s.archiveDao.SimpleArc err:%+v, oid:%d", localErr, oids[0])
			return localErr
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	item := &fm_v2.TabItem{
		FmType:        req.FmType,
		FmId:          req.FmId,
		Title:         info.Fm.Title,
		FirstArcTitle: arc.Title,
		Cover:         info.Fm.Cover,
	}
	if req.FmType == fm_v2.AudioSeason {
		item.Style = fm_v2.RectangleStyle
	} else if req.FmType == fm_v2.AudioSeasonUp {
		item.Style = fm_v2.CircleStyle
	}
	return &fm_v2.HandleTabItemsResp{TabItems: []*fm_v2.TabItem{item}}, nil
}
