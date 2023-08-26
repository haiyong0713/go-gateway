package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"

	egv2 "go-common/library/sync/errgroup.v2"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/channel"
	"go-gateway/app/app-svr/app-channel/interface/model/feed"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"

	"go-farm"
)

const (
	_initRegionKey = "region_key_%d_%v"
	_initlanguage  = "hans"
	_initVersion   = "region_version"
	_regionRepeat  = "r_%d_%d"
	_maxAtten      = 10 //展示最多10个我的订阅
)

var (
	_tabList = []*channel.TabList{
		{
			Name:  "推荐",
			URI:   "bilibili://pegasus/channel/feed/%d",
			TabID: "multiple",
		},
		{
			Name:  "话题",
			URI:   "bilibili://following/topic_detail?id=%d&name=%s",
			TabID: "topic",
		},
	}
	_reidWhiteList = map[int]struct{}{
		1:     {},
		3:     {},
		129:   {},
		4:     {},
		36:    {},
		188:   {},
		160:   {},
		119:   {},
		155:   {},
		165:   {},
		5:     {},
		181:   {},
		65554: {},
		65553: {},
		65551: {},
		65550: {},
	}
)

// Tab channel tab
func (s *Service) Tab(c context.Context, tid, mid int64, tname string, plat int8, build int) (res *channel.Tab, err error) {
	var (
		t          *tag.ChannelReply
		channelIDs []int64
		channels   map[int64]*channelgrpc.Channel
	)
	if t, err = s.tg.ChannelDetail(c, mid, tid, tname, s.isOverseas(plat)); err != nil || t == nil {
		log.Error("s.tag.ChannelDetail(%d, %d, %v) error(%v)", mid, tid, tname, err)
		return
	}
	if (model.IsIPhone(plat) && build > s.c.BuildLimit.TabSimilarIOS) || (model.IsAndroid(plat) && build > s.c.BuildLimit.TabSimilarAndroid) {
		if t != nil {
			for _, s := range t.Synonyms {
				if s != nil && s.Id != 0 {
					channelIDs = append(channelIDs, s.Id)
				}
			}
		}
		if channels, err = s.chDao.Infos(c, channelIDs, mid); err != nil {
			log.Error("%v", err)
			err = nil
		}
	}
	res = &channel.Tab{}
	res.SimilarTagChange(t, channels)
	res.TabList = s.tablist(t)
	return
}

// SubscribeAdd subscribe add
func (s *Service) SubscribeAdd(c context.Context, mid, id int64, now time.Time) (err error) {
	if err = s.tg.SubscribeAdd(c, mid, id, now); err != nil {
		log.Error("s.tg.SubscribeAdd(%d,%d) error(%v)", mid, id, err)
		return
	}
	return
}

// SubscribeCancel subscribe channel
func (s *Service) SubscribeCancel(c context.Context, mid, id int64, now time.Time) (err error) {
	if err = s.tg.SubscribeCancel(c, mid, id, now); err != nil {
		log.Error("s.tg.SubscribeCancel(%d,%d) error(%v)", mid, id, err)
		return
	}
	return
}

// SubscribeUpdate subscribe update
func (s *Service) SubscribeUpdate(c context.Context, mid int64, ids string) (err error) {
	if err = s.tg.SubscribeUpdate(c, mid, ids); err != nil {
		log.Error("s.tg.SubscribeUpdate(%d,%s) error(%v)", mid, ids, err)
		return
	}
	return
}

// List 频道tab页
func (s *Service) List(c context.Context, mid int64, plat int8, build, limit, teenagersMode int, ver, mobiApp, device, lang, paramChannel string) (res *channel.List, err error) {
	var (
		rec, atten  []*channel.Channel
		top, bottom []*channel.Region
		max         = 3
	)
	g, _ := errgroup.WithContext(c)
	//获取推荐的三个频道
	g.Go(func() (err error) {
		rec, err = s.Recommend(c, mid, plat)
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	//获取我的订阅
	if mid > 0 {
		g.Go(func() (err error) {
			atten, err = s.Subscribe(c, mid, limit)
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	//获取分区
	g.Go(func() (err error) {
		top, bottom, _, err = s.RegionList(c, build, teenagersMode, mobiApp, device, lang, paramChannel)
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	_ = g.Wait()
	if tl := len(rec); tl < max {
		if last := max - tl; len(atten) > last {
			rec = append(rec, atten[:last]...)
		} else {
			rec = append(rec, atten...)
		}
	} else {
		rec = rec[:max]
	}
	res = &channel.List{
		RegionTop:    top,
		RegionBottom: bottom,
	}
	if isAudit := s.auditList(mobiApp, plat, build); !isAudit {
		res.RecChannel = rec
		res.AttenChannel = atten
	}
	res.Ver = s.hash(res)
	return
}

// Recommend 推荐
func (s *Service) Recommend(c context.Context, mid int64, plat int8) (res []*channel.Channel, err error) {
	list, err := s.tg.Discover(c, mid, s.isOverseas(plat))
	if err != nil {
		log.Error("%+v", err)
		return
	}
	for _, chann := range list {
		item := &channel.Channel{
			ID:      chann.Id,
			Name:    chann.Name,
			Cover:   chann.Cover,
			IsAtten: chann.Attention,
			Atten:   chann.Sub,
		}
		res = append(res, item)
	}
	return
}

// Subscribe 我订阅的tag（老） standard放前面用户自定义custom放后面
func (s *Service) Subscribe(c context.Context, mid int64, limit int) (res []*channel.Channel, err error) {
	var (
		tinfo []*tag.Tag
	)
	list, err := s.tg.Subscribe(c, mid)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	tinfo = list.Standard
	tinfo = append(tinfo, list.Custom...)
	for _, chann := range tinfo {
		item := &channel.Channel{
			ID:      chann.Id,
			Name:    chann.Name,
			Cover:   chann.Cover,
			Atten:   chann.Sub,
			IsAtten: chann.Attention,
			Content: chann.Content,
		}
		res = append(res, item)
	}
	if len(res) > limit && limit > 0 {
		res = res[:limit]
	} else if len(res) == 0 {
		res = []*channel.Channel{}
	}
	return
}

// Discover 发现频道页（推荐走recommend接口，有分类的揍list接口）
func (s *Service) Discover(c context.Context, id, mid int64, plat int8) (res []*channel.Channel, err error) {
	var (
		list []*tag.Channel
	)
	if id > 0 {
		list, err = s.tg.ListByCategory(c, id, mid, s.isOverseas(plat))
		if err != nil {
			log.Error("%+v", err)
			return
		}
	} else {
		list, err = s.tg.Recommend(c, mid, s.isOverseas(plat))
		if err != nil {
			log.Error("%+v", err)
			return
		}
	}
	if len(list) == 0 {
		res = []*channel.Channel{}
		return
	}
	for _, chann := range list {
		item := &channel.Channel{
			ID:      chann.Id,
			Name:    chann.Name,
			Cover:   chann.Cover,
			Atten:   chann.Sub,
			IsAtten: chann.Attention,
			Content: chann.Content,
		}
		res = append(res, item)
	}
	return
}

// Category 频道分类
func (s *Service) Category(c context.Context, plat int8) (res []*channel.Category, err error) {
	category, err := s.tg.Category(c, s.isOverseas(plat))
	if err != nil {
		log.Error("%+v", err)
		return
	}
	res = append(res, &channel.Category{
		ID:   0,
		Name: "推荐",
	})
	for _, cat := range category {
		item := &channel.Category{
			ID:   cat.Id,
			Name: cat.Name,
		}
		res = append(res, item)
	}
	return
}

// RegionList 分区信息
// nolint:gocognit
func (s *Service) RegionList(c context.Context, build, teenagersMode int, mobiApp, device, lang, paramChannel string) (regionTop, regionBottom, regions []*channel.Region, err error) {
	var (
		hantlanguage = "hant"
		plat         = model.Plat2(mobiApp, device)
	)
	if ok := model.IsOverseas(plat); ok && lang != _initlanguage && lang != hantlanguage {
		lang = hantlanguage
	} else if lang == "" {
		lang = _initlanguage
	}
	var (
		rs = s.cachelist[fmt.Sprintf(_initRegionKey, plat, lang)]
		// maxTop = 8
		ridtmp    = map[string]struct{}{}
		pids      []string
		auths     = make(map[string]*locgrpc.Auth)
		hiddenMap = make(map[int64]bool)
		ip        = metadata.String(c, metadata.RemoteIP)
		rids      []int64
	)
	regionTop = []*channel.Region{}
	regionBottom = []*channel.Region{}
	regions = []*channel.Region{}
	for _, rtmp := range rs {
		if rtmp.ReID != 0 { //过滤二级分区
			continue
		}
		if rtmp.Area != "" {
			pids = append(pids, rtmp.Area)
		}
		rids = append(rids, int64(rtmp.RID))
	}
	eg := egv2.WithContext(c)
	if len(pids) > 0 {
		eg.Go(func(ctx context.Context) error {
			auths, _ = s.loc.AuthPIDs(ctx, strings.Join(pids, ","), ip)
			return nil
		})
	}
	if len(rids) > 0 && model.IsAndroidAll(plat) {
		eg.Go(func(ctx context.Context) error {
			reply, err := s.resDao.EntrancesIsHidden(ctx, rids, build, plat, paramChannel)
			if err != nil {
				log.Error("s.resDao.EntrancesIsHidden err(%+v)", err)
				return nil
			}
			if reply != nil {
				hiddenMap = reply.Infos
			}
			return nil
		})
	}
	_ = eg.Wait()
LOOP:
	for _, rtmp := range rs {
		r := &channel.Region{}
		*r = *rtmp
		if r.ReID != 0 { //过滤二级分区
			continue
		}
		var tmpl, limitshow bool
		if limit, ok := s.limitCache[r.ID]; ok {
			for i, l := range s.limitCache[r.ID] {
				if i+1 <= len(limit)-1 {
					if ((l.Condition == "gt" && limit[i+1].Condition == "lt") && (l.Build < limit[i+1].Build)) ||
						((l.Condition == "lt" && limit[i+1].Condition == "gt") && (l.Build > limit[i+1].Build)) {
						if (l.Condition == "gt" && limit[i+1].Condition == "lt") &&
							(build > l.Build && build < limit[i+1].Build) {
							break
						} else if (l.Condition == "lt" && limit[i+1].Condition == "gt") &&
							(build < l.Build && build > limit[i+1].Build) {
							break
						} else {
							tmpl = true
							continue
						}
					}
				}
				if tmpl {
					if i == len(limit)-1 {
						limitshow = true
						break
						// continue LOOP
					}
					tmpl = false
					continue
				}
				if model.InvalidBuild(build, l.Build, l.Condition) {
					limitshow = true
					continue
					// continue LOOP
				} else {
					limitshow = false
					break
				}
			}
		}
		if limitshow {
			continue LOOP
		}
		// nolint:gomnd
		if r.RID == 65539 {
			if model.IsIOS(plat) {
				r.URI = fmt.Sprintf("%s?from=category", r.URI)
			} else {
				r.URI = fmt.Sprintf("%s?sourceFrom=541", r.URI)
			}
		}
		if auth, ok := auths[r.Area]; ok && auth.Play == int64(locgrpc.Status_Forbidden) {
			log.Warn("s.invalid area(%v) ip(%v) error(%v)", r.Area, ip, err)
			continue
		}
		if isAudit := s.auditRegion(mobiApp, plat, build, r.RID); isAudit {
			continue
		}
		if _, ok := _reidWhiteList[r.RID]; teenagersMode != 0 && !ok {
			continue
		}
		config, ok := s.configCache[r.ID]
		if !ok {
			continue
		}
		// 判断是否在入口屏蔽配置里
		if isHidden, ok := hiddenMap[int64(r.RID)]; ok && isHidden {
			continue
		}
		for _, conf := range config {
			key := fmt.Sprintf(_regionRepeat, r.RID, r.ReID)
			switch conf.ScenesID {
			case 0, 1: // 分区入口
				if _, ok := ridtmp[key]; !ok {
					ridtmp[key] = struct{}{}
				} else {
					continue LOOP
				}
			default:
				continue
			}
			switch conf.ScenesID {
			case 1: //顶部分区
				regionTop = append(regionTop, r)
				regions = append(regions, r)
			case 0: //底部分区
				regionBottom = append(regionBottom, r)
				regions = append(regions, r)
			}
		}
	}
	return
}

func (s *Service) hash(v *channel.List) string {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return _initVersion
	}
	return strconv.FormatUint(farm.Hash64(bs), 10)
}

func (s *Service) loadRegionlist() {
	log.Info("cronLog start loadRegionlist")
	res, err := s.rg.AllList(context.TODO())
	if err != nil {
		log.Error("s.dao.All error(%v)", err)
		return
	}
	tmp := map[string][]*channel.Region{}
	for _, v := range res {
		key := fmt.Sprintf(_initRegionKey, v.Plat, v.Language)
		tmp[key] = append(tmp[key], v)
	}
	if len(tmp) > 0 {
		s.cachelist = tmp
	}
	log.Info("region list cacheproc success")
	limit, err := s.rg.Limit(context.TODO())
	if err != nil {
		log.Error("s.dao.limit error(%v)", err)
		return
	}
	s.limitCache = limit
	log.Info("region limit cacheproc success")
	config, err := s.rg.Config(context.TODO())
	if err != nil {
		log.Error("s.dao.Config error(%v)", err)
		return
	}
	s.configCache = config
	log.Info("region config cacheproc success")
}

// Square 频道广场页
// nolint:gocognit
func (s *Service) Square(c context.Context, mid int64, plat int8, build, teenagersMode int, loginEvent int32, mobiApp, device, lang, buvid, paramChannel string, now time.Time) (res *channel.Square, err error) {
	res = new(channel.Square)
	var (
		squ     *taggrpc.ChannelSquareReply
		regions []*channel.Region
		oidNum  = 2
	)
	isAudit := s.auditList(mobiApp, plat, build)
	eg := errgroup.Group{}
	//获取分区
	eg.Go(func() (err error) {
		_, _, regions, err = s.RegionList(c, build, teenagersMode, mobiApp, device, lang, paramChannel)
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		res.Region = regions
		return
	})
	if !isAudit {
		//获取推荐频道
		eg.Go(func() (err error) {
			var (
				oids            []int64
				tagm            = map[int64]*taggrpc.Tag{}
				chanOids        = map[int64][]*channel.ChanOids{}
				channelCards    = map[int64][]*card.Card{}
				initCardPlatKey = "card_platkey_%d_%d"
				infocFeed       []*feed.ChannelInfo
			)
			squ, err = s.tg.Square(c, mid, s.c.SquareCount, oidNum, build, loginEvent, plat, buvid, s.isOverseas(plat))
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, rec := range squ.GetSquares() {
				cards, ok := s.cardCache[rec.GetChannel().Id]
				if !ok {
					continue
				}
			LOOP:
				for _, c := range cards {
					key := fmt.Sprintf(initCardPlatKey, plat, c.ID)
					cardPlat, ok := s.cardPlatCache[key]
					if !ok {
						continue
					}
					if c.Type != model.GotoAv {
						continue
					}
					for _, l := range cardPlat {
						if model.InvalidBuild(build, l.Build, l.Condition) {
							continue LOOP
						}
					}
					channelCards[c.ChannelID] = append(channelCards[c.ChannelID], c)
				}
			}
			for _, v := range squ.GetSquares() {
				if v == nil {
					continue
				}
				oids = append(oids, v.Oids...)
				if cards, ok := channelCards[v.Channel.Id]; ok {
					for _, c := range cards {
						if c.Type == model.GotoAv {
							chanOids[v.Channel.Id] = append(chanOids[v.Channel.Id], &channel.ChanOids{Oid: c.Value, FromType: _fTypeOperation})
							oids = append(oids, c.Value)
						}
					}
				}
				for _, tmpOid := range v.Oids {
					chanOids[v.Channel.Id] = append(chanOids[v.Channel.Id], &channel.ChanOids{Oid: tmpOid, FromType: _fTypeRecommend})
				}
			}
			var aids []*arcgrpc.PlayAv
			for _, oid := range oids {
				if oid != 0 {
					aids = append(aids, &arcgrpc.PlayAv{Aid: oid})
				}
			}
			am, err := s.Archives(c, aids, false)
			if err != nil {
				return
			}
			for _, v := range squ.GetSquares() {
				var cardItem []*operate.Card
				rec := v.GetChannel()
				if rec == nil {
					continue
				}
				tagm[rec.Id] = &taggrpc.Tag{
					Id:        rec.Id,
					Name:      rec.Name,
					Cover:     rec.Cover,
					Content:   rec.ShortContent,
					Type:      int32(rec.Type),
					State:     rec.State,
					Attention: rec.Attention,
				}
				tagm[rec.Id].Sub = rec.Sub
				for _, oidItem := range chanOids[rec.Id] {
					// nolint:gomnd
					if len(cardItem) >= 2 {
						break
					}
					if _, ok := am[oidItem.Oid]; !ok {
						continue
					}
					cardItem = append(cardItem, &operate.Card{ID: oidItem.Oid, FromType: oidItem.FromType})
				}
				// nolint:gomnd
				if len(cardItem) < 2 {
					continue
				}
				var (
					h = cardm.Handle(plat, cdm.CardGt("channel_square"), "channel_square", cdm.ColumnSvrSingle, nil, tagm, nil, nil, nil, nil, nil)
				)
				if h == nil {
					continue
				}
				op := &operate.Card{
					ID:      rec.Id,
					Items:   cardItem,
					Plat:    plat,
					Param:   strconv.FormatInt(rec.Id, 10),
					Build:   build,
					MobiApp: mobiApp,
				}
				_ = h.From(am, op)
				if h.Get() != nil && h.Get().Right {
					res.Square = append(res.Square, h)
				}
				//infoc
				var (
					infoItems []*feed.ChannelInfoItem
					pos       int
				)
				for _, tmp := range cardItem {
					item := &feed.ChannelInfoItem{}
					item.FromChannelInfoItem(tmp)
					pos++
					item.Pos = pos
					infoItems = append(infoItems, item)
				}
				infocFeed = append(infocFeed, &feed.ChannelInfo{
					ChannelID:   rec.Id,
					ChannelName: rec.Name,
					Item:        infoItems,
				})
			}
			//infoc
			s.SquareInfoc(mobiApp, device, buvid, build, mid, now, infocFeed)
			return
		})
	}
	_ = eg.Wait()
	return
}

// Mysub 我订阅的tag（新） standard放前面用户自定义custom放后面
func (s *Service) Mysub(c context.Context, mid int64, limit int) (res *channel.Mysub, err error) {
	var (
		tinfo      []*tag.Tag
		subChannel []*channel.Channel
	)
	res = new(channel.Mysub)
	list, err := s.tg.Subscribe(c, mid)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	tinfo = list.Standard
	tinfo = append(tinfo, list.Custom...)
	if len(tinfo) > 0 {
		for _, chann := range tinfo {
			subChannel = append(subChannel, &channel.Channel{
				ID:      chann.Id,
				Name:    chann.Name,
				Cover:   chann.Cover,
				Atten:   chann.Sub,
				IsAtten: chann.Attention,
				Content: chann.Content,
			})
		}
		if len(subChannel) > limit && limit > 0 {
			subChannel = subChannel[:limit]
		}
	}
	res.List = subChannel
	res.DisplayCount = _maxAtten
	return
}

func (s *Service) isOverseas(plat int8) (res int32) {
	if ok := model.IsOverseas(plat); ok {
		res = 1
	} else {
		res = 0
	}
	return
}

func (s *Service) tablist(t *tag.ChannelReply) (res []*channel.TabList) {
	res = s.defaultTab(t)
	var (
		mpos        []int
		tmpmenus    = map[int]*tab.Menu{}
		menus       = s.menuCache[t.Channel.Id]
		menusTabIDs = map[int64]struct{}{}
	)
	if len(menus) == 0 {
		return
	}
	for _, m := range menus {
		tmpmenus[m.Priority] = m
		mpos = append(mpos, m.Priority)
	}
	for _, pos := range mpos {
		var (
			tmpm *tab.Menu
			ok   bool
		)
		if tmpm, ok = tmpmenus[pos]; !ok || pos == 0 {
			continue
		}
		if _, ok := menusTabIDs[tmpm.TabID]; !ok {
			menusTabIDs[tmpm.TabID] = struct{}{}
		} else {
			continue
		}
		tl := &channel.TabList{}
		tl.TabListChange(tmpm)
		if len(res) < pos {
			res = append(res, tl)
			continue
		}
		res = append(res[:pos-1], append([]*channel.TabList{tl}, res[pos-1:]...)...)
	}
	return
}

func (s *Service) defaultTab(t *tag.ChannelReply) (res []*channel.TabList) {
	for _, tmp := range _tabList {
		r := &channel.TabList{}
		*r = *tmp
		switch tmp.TabID {
		case "multiple":
			r.URI = fmt.Sprintf(r.URI, t.Channel.Id)
		case "topic":
			r.URI = fmt.Sprintf(r.URI, t.Channel.Id, url.QueryEscape(t.Channel.Name))
		}
		res = append(res, r)
	}
	return
}
