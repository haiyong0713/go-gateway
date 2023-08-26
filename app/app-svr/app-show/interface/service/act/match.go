package act

import (
	"context"
	"fmt"
	"strconv"
	"time"

	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"

	"go-gateway/app/app-svr/app-show/interface/conf"
	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

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

var _eventStatusImg = map[esportsgrpc.SportsMatchStatusEnum]string{
	esportsgrpc.SportsMatchStatusEnum_MatchStatusScheduled:   "https://i0.hdslb.com/bfs/kfptfe/floor/0c3d65430f815f0a437f7ca8d425040e91bb33d2.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusRunning:     "https://i0.hdslb.com/bfs/kfptfe/floor/0196023567d9c3635ef7bed364544310fe153f60.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusFinished:    "https://i0.hdslb.com/bfs/kfptfe/floor/b527e62d280e226466fe88d3932f37f8447f7b35.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusCancelled:   "https://i0.hdslb.com/bfs/kfptfe/floor/13ba8ce1aa823f3a53cdd5cb9c10e6ed53c8843e.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusDelayed:     "https://i0.hdslb.com/bfs/kfptfe/floor/455a4222611cbabcdcda566d053d7090f60a4335.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusPostponed:   "https://i0.hdslb.com/bfs/kfptfe/floor/c8e704bad343e7ffb619fda63d32c6a45c4636a4.png",
	esportsgrpc.SportsMatchStatusEnum_MatchStatusRescheduled: "https://i0.hdslb.com/bfs/kfptfe/floor/56dbafc6f19cea1c4a6ae1926b2317f33a198160.png",
}

func (s *Service) FormatMatchMedal(c context.Context, mou *natpagegrpc.NativeModule) *actmdl.Item {
	if mou.Fid <= 0 || s.c.WinterOlyMedal == nil {
		return nil
	}
	rly, err := s.esportsDao.GetSportsSeasonMedalTable(c, mou.Fid, esportsgrpc.MedalTableTypeEnum_Gold)
	if err != nil || rly == nil {
		return nil
	}
	cfg := s.c.WinterOlyMedal
	tableItems := make([]*actmdl.Item, 0, len(rly.Table))
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
		tableItems = append(tableItems, &actmdl.Item{
			Item:  formatMedalItem(record, cfg.RankColor[strconv.FormatInt(int64(record.GoldRank), 10)], nameColor),
			Color: &actmdl.Color{BgColor: bgColor},
		})
	}
	// special特殊展示
	if rly.SpecialRecord != nil && !handledSpecial {
		rank := strconv.FormatInt(int64(rly.SpecialRecord.GoldRank), 10)
		tableItems = append(tableItems, &actmdl.Item{
			Item:  formatMedalItem(rly.SpecialRecord, cfg.RankColor[rank], cfg.SpecialFontColor),
			Color: &actmdl.Color{BgColor: cfg.SpecialBgColor},
		})
	}
	if len(tableItems) == 0 {
		return nil
	}
	matchMedal := &actmdl.Item{
		Goto:       actmdl.GotoMatchMedal,
		ItemID:     mou.Fid,
		TableAttrs: matchMedalTableAttrs(),
		Header: &actmdl.Item{
			Item:  matchHeaderItems(),
			Color: &actmdl.Color{BgColor: cfg.HeaderBgColor},
		},
		Item: tableItems,
	}
	if rly.Tips != nil {
		matchMedal.Title = fmt.Sprintf("截止%s", time.Unix(rly.Tips.UpdateTime, 0).Format("01月02日15:04"))
		matchMedal.Color = &actmdl.Color{TitleColor: cfg.TitleColor}
	}
	matchMedalModule := &actmdl.Item{}
	matchMedalModule.FromMatchMedalModule(mou, []*actmdl.Item{matchMedal})
	return matchMedalModule
}

func matchMedalTableAttrs() []*actmdl.TableAttr {
	return []*actmdl.TableAttr{
		{Ratio: 13, TextAlign: actmdl.TextAlignCenter},
		{Ratio: 35, TextAlign: actmdl.TextAlignLeft},
		{Ratio: 13, TextAlign: actmdl.TextAlignCenter},
		{Ratio: 13, TextAlign: actmdl.TextAlignCenter},
		{Ratio: 13, TextAlign: actmdl.TextAlignCenter},
		{Ratio: 13, TextAlign: actmdl.TextAlignCenter},
	}
}

func matchHeaderItems() []*actmdl.Item {
	return []*actmdl.Item{
		{Content: "排名"},
		{Content: "国家/地区"},
		{Content: "金", Image: _goldMedal},
		{Content: "银", Image: _silverMedal},
		{Content: "铜", Image: _bronzeMedal},
		{Content: "总数"},
	}
}

func formatMedalItem(record *esportsgrpc.MedalRecordItem, rankFontColor, nameFontColor string) []*actmdl.Item {
	var rankColor, nameColor *actmdl.Color
	if rankFontColor != "" {
		rankColor = &actmdl.Color{FontColor: rankFontColor}
	}
	if nameFontColor != "" {
		nameColor = &actmdl.Color{FontColor: nameFontColor}
	}
	item := []*actmdl.Item{
		{Content: strconv.FormatInt(int64(record.GoldRank), 10), Color: rankColor},
		{Content: record.ParticipantName, Image: record.ParticipantImg, Color: nameColor},
		{Content: strconv.FormatInt(int64(record.Gold), 10)},
		{Content: strconv.FormatInt(int64(record.Silver), 10)},
		{Content: strconv.FormatInt(int64(record.Bronze), 10)},
		{Content: strconv.FormatInt(int64(record.Total), 10)},
	}
	return item
}

func (s *Service) FormatMatchEvent(c context.Context, mou *natpagegrpc.NativeModule, mixEvent *natpagegrpc.MatchEvent, params *actmdl.ParamFormatModule) *actmdl.Item {
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
	items := make([]*actmdl.Item, 0, len(eventIds))
	for _, eventId := range eventIds {
		event, ok := rly.Matches[eventId]
		if !ok || event == nil || !event.FocusOn || event.Content == "" {
			continue
		}
		item := &actmdl.Item{
			ItemID:      event.Id,
			Title:       event.Name,
			Images:      eventHeaderImages(event),
			Time:        time.Unix(event.BeginTime, 0).Format("15:04"),
			ImagesUnion: eventImagesUnion(event),
			Content:     event.Content,
			Color:       &actmdl.Color{},
		}
		// 兼容ios
		if params != nil && params.Platform == "ios" {
			if item.ImagesUnion == nil {
				item.ImagesUnion = new(actmdl.ImagesUnion)
			}
			if item.ImagesUnion.Button == nil {
				if img, ok := _eventStatusImg[event.MatchStatus]; ok {
					item.ImagesUnion.Button = &actmdl.Image{Image: img}
				}
			}
		}
		item.Status, item.Color.StatusColor = eventStatus(event, cfg)
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil
	}
	matchEventModule := &actmdl.Item{}
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

func eventImagesUnion(event *esportsgrpc.SportsEventMatchItem) *actmdl.ImagesUnion {
	imageUnion := new(actmdl.ImagesUnion)
	if event.Img != "" {
		imageUnion.Event = &actmdl.Image{Image: event.Img}
	}
	if event.GuideUrl != "" {
		if image, ok := _eventGuideImages[event.GuideType]; ok {
			imageUnion.Button = &actmdl.Image{Image: image, Uri: event.GuideUrl}
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
