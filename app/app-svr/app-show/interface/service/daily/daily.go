package daily

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/app-show/interface/conf"
	arcdao "go-gateway/app/app-svr/app-show/interface/dao/archive"
	carddao "go-gateway/app/app-svr/app-show/interface/dao/card"
	chdao "go-gateway/app/app-svr/app-show/interface/dao/channel"
	tagdao "go-gateway/app/app-svr/app-show/interface/dao/tag"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/app/app-svr/app-show/interface/model/daily"
)

const (
	_initDailyKey  = "daily_key_%d_%d"
	_initColumnKey = "column_key_%d_%d"
)

var (
	_emptyDaily = []*daily.Show{}
)

type Service struct {
	c     *conf.Config
	cdao  *carddao.Dao
	arc   *arcdao.Dao
	tag   *tagdao.Dao
	chDao *chdao.Dao
	// columnsCache
	columnsCache map[string]*card.Column
	// card
	cardCache       map[string][]*daily.Show
	columnCache     map[string]*daily.Show
	columnListCache map[string][]*daily.Item
	dailyJobRailGun *railgun.Railgun
}

// New new a daily service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		cdao:  carddao.New(c),
		arc:   arcdao.New(c),
		tag:   tagdao.New(c),
		chDao: chdao.New(c),
		// columnsCache
		columnsCache: map[string]*card.Column{},
		// card
		cardCache:       map[string][]*daily.Show{},
		columnCache:     map[string]*daily.Show{},
		columnListCache: map[string][]*daily.Item{},
	}
	now := time.Now()
	s.loadColumnsCache()
	s.loadNperCache(now)
	s.initDailyRailGun(now)
	return
}

// Daily
func (s *Service) Daily(c context.Context, plat int8, build, dailyID, pn, ps int) (res []*daily.Show) {
	if pn > 0 {
		pn = pn - 1
	}
	start := pn * ps
	end := start + ps
	key := fmt.Sprintf(_initColumnKey, plat, dailyID)
	if column, ok := s.columnsCache[key]; ok {
		if model.InvalidBuild(build, column.Build, column.Condition) {
			res = _emptyDaily
			return
		}
		cardKey := fmt.Sprintf(_initDailyKey, plat, dailyID)
		if cards, ok := s.cardCache[cardKey]; ok {
			for _, sw := range cards {
				if model.InvalidBuild(build, sw.Build, sw.Condition) {
					continue
				}
				res = append(res, sw)
			}
			resLen := len(res)
			if resLen > end {
				res = res[start:end]
			} else if resLen > start {
				res = res[start:]
			} else {
				res = _emptyDaily
			}
		}
	}
	if len(res) == 0 {
		res = _emptyDaily
	}
	return
}

// initDailyRailGun to init railgun cron job
func (s *Service) initDailyRailGun(now time.Time) {
	r := railgun.NewRailGun("loadDailyCache", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 */3 * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadColumnsCache()
		s.loadNperCache(now)
		return railgun.MsgPolicyNormal
	}))
	s.dailyJobRailGun = r
	r.Start()
}

// loadColumnsCache load all columns cache
func (s *Service) loadColumnsCache() {
	res, err := s.cdao.Columns(context.TODO())
	if err != nil {
		log.Error("s.cdao.Columns error(%v)", err)
		return
	}
	tmp := map[string]*card.Column{}
	for plat, columns := range res {
		for _, column := range columns {
			key := fmt.Sprintf(_initColumnKey, plat, column.ID)
			tmp[key] = column
		}
	}
	s.columnsCache = tmp
	log.Info("column cache size(%d)", len(s.columnsCache))
}

// loadNperCache
func (s *Service) loadNperCache(now time.Time) {
	hdm, err := s.cdao.ColumnNpers(context.TODO())
	if err != nil {
		log.Error("s.cdao.ColumnNpers error(%v)", err)
		return
	}
	itm, aids, err := s.cdao.NperContents(context.TODO())
	if err != nil {
		log.Error("s.cdao.NperContents error(%v)", err)
		return
	}
	tmp, tmpColumns, tmpList := s.mergeCard(context.TODO(), hdm, itm, aids, now, 0, "", "")
	s.cardCache = tmp
	s.columnCache = tmpColumns
	s.columnListCache = tmpList
	log.Info("load cardCache size(%d), columnCache size(%d), columnListCache size(%d)", len(s.cardCache), len(s.columnCache), len(s.columnListCache))
}

// mergeCard
func (s *Service) mergeCard(c context.Context, hdm map[int8][]*card.ColumnNper, itm map[int][]*card.Content, itmaids map[int][]int64, now time.Time, mid int64, mobiApp, device string) (res map[string][]*daily.Show, columns map[string]*daily.Show, columnList map[string][]*daily.Item) {
	var (
		dailyMAX = 31
	)
	res = map[string][]*daily.Show{}
	columnList = map[string][]*daily.Item{}
	columns = map[string]*daily.Show{}
	for plat, hds := range hdm {
		for _, hd := range hds {
			var (
				ok     bool
				column *card.Column
			)
			columnskey := fmt.Sprintf(_initColumnKey, plat, hd.ColumnID)
			if column, ok = s.columnsCache[columnskey]; !ok {
				continue
			}
			switch column.Type {
			case model.GotoDaily:
				if dailykey := fmt.Sprintf(_initDailyKey, plat, hd.ColumnID); len(res[dailykey]) > dailyMAX {
					continue
				}
			}
			var (
				sis []*daily.Item
			)
			its, ok := itm[hd.ID]
			if !ok {
				its = []*card.Content{}
			}
			// nolint:gomnd
			switch column.Tpl {
			case 1, 2:
				var tmpItem = map[int64]*daily.Item{}
				if aids, ok := itmaids[hd.ID]; ok {
					tmpItem = s.fromCardAids(context.TODO(), aids, mid, mobiApp, device)
				}
				for _, ci := range its {
					si := s.fillCardItem(ci, tmpItem)
					if si.Title == "" {
						continue
					}
					if ci.TagID > 0 {
						si.TagName, si.TagID, si.TagURI = s.fromTagIDByName(c, ci.TagID, now)
					}
					sis = append(sis, si)
				}
			}
			if len(sis) == 0 {
				continue
			}
			sw := &daily.Show{}
			sw.Head = &daily.Head{
				ColumnID:  hd.ID,
				Build:     hd.Build,
				Condition: hd.Condition,
				Plat:      hd.Plat,
				Desc:      hd.Desc,
				Type:      column.Type,
			}
			if hd.Cover != "" {
				sw.Cover = hd.Cover
			}
			var key string
			switch sw.Head.Type {
			case model.GotoDaily:
				key = fmt.Sprintf(_initDailyKey, plat, hd.ColumnID)
				sw.Head.Title = hd.Name
				sw.Head.Date = int64(hd.NperTime)
				sw.Body = sis
				res[key] = append(res[key], sw)
			case model.GotoColumn:
				key = fmt.Sprintf(_initDailyKey, plat, hd.ID)
				sw.Head.Title = hd.Name
				sw.Head.Goto = hd.Goto
				sw.Head.Param = hd.Param
				sw.Head.URI = hd.URI
				columnList[key] = sis
				columns[key] = sw
			}
		}
	}
	return
}

// fillCardItem
func (s *Service) fillCardItem(csi *card.Content, tsi map[int64]*daily.Item) (si *daily.Item) {
	si = &daily.Item{}
	switch csi.Type {
	case model.CardGotoAv:
		si.Goto = model.GotoAv
		si.Param = csi.Value
	}
	si.URI = model.FillURI(si.Goto, si.Param, nil)
	if si.Goto == model.GotoAv {
		aid, err := strconv.ParseInt(si.Param, 10, 64)
		if err != nil {
			return
		}
		if it, ok := tsi[aid]; ok {
			si = it
			if csi.Title != "" {
				si.Title = csi.Title
			}
		} else {
			si = &daily.Item{}
		}
	}
	return
}

// fromCardAids get Aids.
func (s *Service) fromCardAids(ctx context.Context, aids []int64, mid int64, mobiApp, device string) (data map[int64]*daily.Item) {
	var (
		arc *api.Arc
		ok  bool
	)
	as, err := s.arc.ArchivesPB(ctx, aids, mid, mobiApp, device)
	if err != nil {
		log.Error("s.arc.ArchivesPB(%v) error(%v)", aids, err)
		return
	}
	if len(as) == 0 {
		log.Warn("s.arc.ArchivesPB(%v) length is 0", aids)
		return
	}
	data = map[int64]*daily.Item{}
	for _, aid := range aids {
		if arc, ok = as[aid]; ok {
			if !arc.IsNormal() {
				continue
			}
			i := &daily.Item{}
			i.FromArchivePB(arc)
			data[aid] = i
		}
	}
	return
}

// fromTagIDByName from tag_id by tag_name
func (s *Service) fromTagIDByName(ctx context.Context, tagID int, now time.Time) (tagName string, tagIDInt int64, tagURI string) {
	tag, err := s.tag.TagInfo(ctx, 0, tagID, now)
	if err != nil {
		log.Error("s.tag.TagInfo(%d) error(%v)", tagID, err)
		return
	}
	tagName = tag.Name
	tagIDInt = tag.Tid
	channels, err := s.chDao.Infos(ctx, []int64{tagIDInt}, 0)
	if err != nil {
		log.Error("%v", err)
		return
	}
	if channel, ok := channels[tagIDInt]; ok && channel != nil {
		if channel.CType == model.NewChannel {
			tagURI = model.FillURI(model.GotoChannelNewAll, strconv.FormatInt(tagIDInt, 10), nil)
		}
	}
	return
}

func (s *Service) Close() {
	s.dailyJobRailGun.Close()
	s.cdao.Close()
}
