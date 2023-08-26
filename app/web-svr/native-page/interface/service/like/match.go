package like

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	"go-common/library/log"

	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/interface/conf"
	mdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
)

const (
	_nativePageTable        = "native_page"
	_nativePageDynTable     = "native_page_dyn"
	_nativePageExtTable     = "native_page_ext"
	_nativeModuleTable      = "native_module"
	_nativeClickTable       = "native_click"
	_nativeActTable         = "native_act"
	_nativeDnamicTable      = "native_dynamic_ext"
	_nativeVideoTable       = "native_video_ext"
	_nativeMixtureExt       = "native_mixture_ext"
	_nativeParticipationExt = "native_participation_ext"
	_nativeTsPageTable      = "native_ts_page"
	_nativeTsModuleTable    = "native_ts_module"
	_nativeActTab           = "act_tab"
	_nativeTabModule        = "act_tab_module"
)

// ClearCache del cache
func (s *Service) ClearCache(c context.Context, msg string) (err error) {
	var m struct {
		Table string `json:"table"`
	}
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("ClearCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("ClearCache json.Unmarshal msg(%s)", msg)
	switch m.Table {
	case _nativePageTable, _nativeModuleTable, _nativeClickTable, _nativeActTable, _nativeDnamicTable, _nativeVideoTable, _nativeMixtureExt, _nativeParticipationExt, _nativeTabModule, _nativeActTab, _nativeTsPageTable, _nativeTsModuleTable:
		err = s.AutoDispense(c, msg)
	case _nativePageDynTable:
		err = s.clearNativePageDyn(c, msg)
	case _nativePageExtTable:
		err = s.clearNativePageExt(c, msg)
	}
	// 缓存处理出错，需要日志告警
	if err != nil {
		log.Error("日志告警:ClearCache(%s)缓存处理出错数据(%s) error(%v)", m.Table, msg, err)
	}
	return
}

func (s *Service) clearNativePageDyn(c context.Context, msg string) error {
	var m struct {
		New struct {
			ID  int64 `json:"id"`
			Pid int64 `json:"pid"`
		} `json:"new,omitempty"`
	}
	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("Fail to unmarshal native_page_dyn, msg=%s err=%+v", msg, err)
		return err
	}
	log.Info("clearNativePageDyn json.Unmarshal msg(%s)", msg)
	if m.New.ID <= 0 || m.New.Pid <= 0 {
		return nil
	}
	return s.natDao.DelCacheNativePagesExt(c, m.New.Pid)
}

func (s *Service) clearNativePageExt(c context.Context, msg string) error {
	var m struct {
		New struct {
			ID  int64 `json:"id"`
			Pid int64 `json:"pid"`
		} `json:"new,omitempty"`
	}
	if err := json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("Fail to unmarshal native_page_ext, msg=%s err=%+v", msg, err)
		return err
	}
	log.Info("clearNativePageExt json.Unmarshal msg(%s)", msg)
	if m.New.ID <= 0 || m.New.Pid <= 0 {
		return nil
	}
	if err := s.natDao.DelCacheNativeExtend(c, m.New.Pid); err != nil {
		return err
	}
	//回源
	_, err := s.natDao.NativeExtend(c, m.New.Pid)
	return err
}

const (
	_matchShowMaxRank = 3
	_chinaFlag        = "https://i0.hdslb.com/bfs/activity-plat/static/20220105/fd43ade10c04329bcc177dcb1cdefce0/jaLwmxbRzz.png"
	_goldMedal        = "https://i0.hdslb.com/bfs/activity-plat/static/20220105/fd43ade10c04329bcc177dcb1cdefce0/jvhNOAWn8H.png"
	_silverMedal      = "https://i0.hdslb.com/bfs/activity-plat/static/20220105/fd43ade10c04329bcc177dcb1cdefce0/yxq5NJ0SrT.png"
	_bronzeMedal      = "https://i0.hdslb.com/bfs/activity-plat/static/20220105/fd43ade10c04329bcc177dcb1cdefce0/NZcpeuMT3Q.png"
)

var _eventGuideImages = map[esportsgrpc.SportsMatchGuideTypeEnum]string{
	esportsgrpc.SportsMatchGuideTypeEnum_GuideCollection: "https://i0.hdslb.com/bfs/activity-plat/static/20220105/fd43ade10c04329bcc177dcb1cdefce0/9neciluGVX.png",
	esportsgrpc.SportsMatchGuideTypeEnum_GuidePlayback:   "https://i0.hdslb.com/bfs/activity-plat/static/20220121/256a1a14b990ce65d4a3168e1090a5f7/JHBEouwfbh.png",
	esportsgrpc.SportsMatchGuideTypeEnum_GuideLive:       "https://i0.hdslb.com/bfs/activity-plat/static/20220121/256a1a14b990ce65d4a3168e1090a5f7/Qphs7zDgNs.png",
}

var _eventStatus = map[esportsgrpc.SportsMatchStatusEnum]string{
	esportsgrpc.SportsMatchStatusEnum_MatchStatusScheduled:   "未开始",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusRunning:     "进行中",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusFinished:    "已结束",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusCancelled:   "比赛取消",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusDelayed:     "比赛延迟",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusPostponed:   "比赛推迟",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusRescheduled: "比赛改期",
}

func (s *Service) FormatMatchMedal(c context.Context, mou *natpagegrpc.NativeModule) *mdl.Item {
	if mou.Fid <= 0 || s.c.WinterOlyMedal == nil {
		return nil
	}
	rly, err := s.esportsDao.GetSportsSeasonMedalTable(c, mou.Fid, esportsgrpc.MedalTableTypeEnum_Gold)
	if err != nil || rly == nil {
		return nil
	}
	cfg := s.c.WinterOlyMedal
	tableItems := make([]*mdl.Item, 0, len(rly.Table))
	// 中国特殊展示
	var handledSpecial bool
	for k, record := range rly.Table {
		if record == nil || record.GoldRank > _matchShowMaxRank {
			continue
		}
		nameColor, bgColor := func() (nameColor, bgColor string) {
			if record.ParticipantId == rly.SpecialRecord.GetParticipantId() {
				handledSpecial = true
				return cfg.SpecialFontColor, cfg.SpecialBgColor
			}
			if record.ParticipantName == "中国" {
				return cfg.SpecialFontColor, cfg.SpecialBgColor
			}
			if k%2 == 1 {
				return "", cfg.IntervalBgColor
			}
			return "", cfg.DefaultBgColor
		}()
		tableItems = append(tableItems, &mdl.Item{
			Item:  formatMedalItem(record, cfg.RankColor[strconv.FormatInt(int64(record.GoldRank), 10)], nameColor),
			Color: &mdl.Color{BgColor: bgColor},
		})
	}
	// special特殊展示
	if rly.SpecialRecord != nil && !handledSpecial {
		rank := strconv.FormatInt(int64(rly.SpecialRecord.GoldRank), 10)
		tableItems = append(tableItems, &mdl.Item{
			Item:  formatMedalItem(rly.SpecialRecord, cfg.RankColor[rank], cfg.SpecialFontColor),
			Color: &mdl.Color{BgColor: cfg.SpecialBgColor},
		})
	}
	if len(tableItems) == 0 {
		return nil
	}
	matchMedal := &mdl.Item{
		Goto:       mdl.GotoMatchMedal,
		ItemID:     mou.Fid,
		TableAttrs: matchMedalTableAttrs(),
		Header: &mdl.Item{
			Item:  matchHeaderItems(),
			Color: &mdl.Color{BgColor: cfg.HeaderBgColor},
		},
		Item: tableItems,
	}
	if rly.Tips != nil {
		matchMedal.Title = fmt.Sprintf("截止%s", time.Unix(rly.Tips.UpdateTime, 0).Format("01月02日15:04"))
		matchMedal.Color = &mdl.Color{TitleColor: cfg.TitleColor}
	}
	matchMedalModule := &mdl.Item{}
	matchMedalModule.FromMatchMedalModule(mou, []*mdl.Item{matchMedal})
	return matchMedalModule
}

func matchMedalTableAttrs() []*mdl.TableAttr {
	return []*mdl.TableAttr{
		{Ratio: 13, TextAlign: mdl.TextAlignCenter},
		{Ratio: 35, TextAlign: mdl.TextAlignLeft},
		{Ratio: 13, TextAlign: mdl.TextAlignCenter},
		{Ratio: 13, TextAlign: mdl.TextAlignCenter},
		{Ratio: 13, TextAlign: mdl.TextAlignCenter},
		{Ratio: 13, TextAlign: mdl.TextAlignCenter},
	}
}

func matchHeaderItems() []*mdl.Item {
	return []*mdl.Item{
		{Content: "排名"},
		{Content: "国家/地区"},
		{Content: "金", Image: _goldMedal},
		{Content: "银", Image: _silverMedal},
		{Content: "铜", Image: _bronzeMedal},
		{Content: "总数"},
	}
}

func formatMedalItem(record *esportsgrpc.MedalRecordItem, rankFontColor, nameFontColor string) []*mdl.Item {
	var rankColor, nameColor *mdl.Color
	if rankFontColor != "" {
		rankColor = &mdl.Color{FontColor: rankFontColor}
	}
	if nameFontColor != "" {
		nameColor = &mdl.Color{FontColor: nameFontColor}
	}
	item := []*mdl.Item{
		{Content: strconv.FormatInt(int64(record.GoldRank), 10), Color: rankColor},
		{Content: record.ParticipantName, Image: record.ParticipantImg, Color: nameColor},
		{Content: strconv.FormatInt(int64(record.Gold), 10)},
		{Content: strconv.FormatInt(int64(record.Silver), 10)},
		{Content: strconv.FormatInt(int64(record.Bronze), 10)},
		{Content: strconv.FormatInt(int64(record.Total), 10)},
	}
	return item
}

func (s *Service) FormatMatchEvent(c context.Context, mou *natpagegrpc.NativeModule, mixEvent *natpagegrpc.MatchEvent) *mdl.Item {
	if s.c.WinterOlyEvent == nil || mixEvent == nil {
		return nil
	}
	eventIds := extractIdsFromMixture(mixEvent.List)
	if len(eventIds) == 0 {
		return nil
	}
	rly, err := s.esportsDao.GetSportsEventMatches(c, eventIds)
	if err != nil || rly == nil {
		return nil
	}
	cfg := s.c.WinterOlyEvent
	items := make([]*mdl.Item, 0, len(eventIds))
	for _, eventId := range eventIds {
		event, ok := rly.Matches[eventId]
		if !ok || event == nil || !event.FocusOn || event.Content == "" {
			continue
		}
		item := &mdl.Item{
			ItemID:      event.Id,
			Title:       event.Name,
			Images:      eventHeaderImages(event),
			Time:        time.Unix(event.BeginTime, 0).Format("15:04"),
			ImagesUnion: eventImagesUnion(event),
			Content:     event.Content,
			Color:       &mdl.Color{},
		}
		item.Status, item.Color.StatusColor = eventStatus(event, cfg)
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	matchEventModule := &mdl.Item{}
	matchEventModule.FromMatchEventModule(mou, items, cfg)
	return matchEventModule
}

func eventHeaderImages(event *esportsgrpc.SportsEventMatchItem) []*natpagegrpc.Image {
	var images []*natpagegrpc.Image
	if event.MedalUrl != "" {
		images = append(images, &natpagegrpc.Image{Image: event.MedalUrl})
	}
	if event.SpecialParticipant {
		images = append(images, &natpagegrpc.Image{Image: _chinaFlag})
	}
	return images
}

func eventImagesUnion(event *esportsgrpc.SportsEventMatchItem) *mdl.ImagesUnion {
	imageUnion := new(mdl.ImagesUnion)
	if event.Img != "" {
		imageUnion.Event = &mdl.Image{Image: event.Img}
	}
	if event.GuideUrl != "" {
		if image, ok := _eventGuideImages[event.GuideType]; ok {
			imageUnion.Button = &mdl.Image{Image: image, Uri: event.GuideUrl}
		}
	}
	return imageUnion
}

func eventStatus(event *esportsgrpc.SportsEventMatchItem, cfg *conf.WinterOlyEvent) (status, statusColor string) {
	if v, ok := _eventStatus[event.MatchStatus]; ok {
		status = v
	}
	statusColor = func() string {
		if event.MatchStatus == esportsgrpc.SportsMatchStatusEnum_MatchStatusRunning {
			return cfg.RunningStatusColor
		}
		return cfg.DefaultStatusColor
	}()
	return status, statusColor
}

func extractIdsFromMixture(list []*natpagegrpc.NativeMixtureExt) []int64 {
	ids := make([]int64, 0, len(list))
	for _, ext := range list {
		if ext == nil || !ext.IsOnline() || ext.ForeignID <= 0 {
			continue
		}
		ids = append(ids, ext.ForeignID)
	}
	return ids
}
