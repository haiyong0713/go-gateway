package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/story"
	"go-gateway/app/app-svr/app-card/interface/model/stat"
	"go-gateway/app/app-svr/story/internal/model"
)

type storyInfoc struct {
	ip          string
	now         string
	api         string
	buvid       string
	mid         string
	client      string
	pull        string
	list        interface{}
	isRec       string
	build       string
	returnCode  string
	deviceID    string
	network     string
	trackID     string
	userFeature string
	displayID   string
	fromAv      string
	fromTrackid string
	bizInfo     interface{}
}

func (s *StoryService) StoryInfoc(c context.Context, api, buvid string, mid int64, plat int8, pull int,
	list []*story.Item, respCode int, build int, deviceID, network, trackID, userFeature string, displayID int,
	aid int64, now time.Time, param *model.StoryParam) {
	var (
		items     = []interface{}{}
		aiTrackID string
	)
	adIndex := int32(0)
	adActualIndex := 0
	for i, l := range list {
		if l.Rcmd == nil {
			continue
		}
		if l.AdInfo != nil {
			adActualIndex = i + 1
			adIndex = l.AdInfo.CardIndex
		}
		// 去除第一刷第一个强行插入的from_av的数据
		if displayID == 1 && l.Rcmd.ID == aid && i == 0 {
			continue
		}
		stat.MetricStoryCardTotal.Inc(l.Rcmd.Goto, strconv.Itoa(int(plat)))
		items = append(items, map[string]interface{}{
			"id":              l.Rcmd.ID,
			"pos":             i + 1,
			"goto":            l.Rcmd.Goto,
			"source":          l.Rcmd.Source,
			"av_feature":      l.Rcmd.AvFeature,
			"story_up_mid":    l.Rcmd.StoryUpMid,
			"advertise_type":  l.Rcmd.AdvertiseType,
			"epid":            l.Rcmd.EpID,
			"has_icon":        l.Rcmd.HasIcon,
			"icon_type":       l.Rcmd.IconType,
			"icon_id":         l.Rcmd.IconID,
			"icon_title":      l.Rcmd.IconTitle,
			"dalao_uniq_id":   l.Rcmd.PosRecUniqueId,
			"dalao_title":     l.Rcmd.PosRecTitle,
			"has_topic":       l.Rcmd.HasTopic,
			"topic_id":        l.Rcmd.TopicID,
			"topic_title":     l.Rcmd.TopicTitle,
			"ogv_style":       l.Rcmd.OGVStyle,
			"highlight_start": l.Rcmd.HighlightStart,
			"mutual_reason":   l.Rcmd.MutualReason,
			"rcmd_reason":     l.RcmdReason,
			"extra_json":      l.Rcmd.ExtraJson,
		})
		if l.Rcmd.TrackID != "" {
			aiTrackID = l.Rcmd.TrackID
		}
	}
	listJson := map[string]interface{}{
		"section": map[string]interface{}{
			"items":        items,
			"request_from": param.RequestFrom,
		},
	}
	biz := map[string]interface{}{
		"ad_index":        adIndex,
		"ad_actual_index": adActualIndex,
	}
	ip := metadata.String(c, metadata.RemoteIP)
	infoclog := storyInfoc{
		ip:          ip,
		now:         strconv.FormatInt(now.Unix(), 10),
		api:         api,
		buvid:       buvid,
		mid:         strconv.FormatInt(mid, 10),
		client:      strconv.Itoa(int(plat)),
		pull:        strconv.Itoa(pull),
		list:        listJson,
		isRec:       strconv.Itoa(0),
		build:       strconv.Itoa(int(build)),
		returnCode:  strconv.Itoa(respCode),
		deviceID:    deviceID,
		network:     network,
		trackID:     aiTrackID,
		userFeature: userFeature,
		displayID:   strconv.Itoa(int(displayID)),
		fromAv:      strconv.FormatInt(aid, 10),
		fromTrackid: trackID,
		bizInfo:     biz,
	}
	// ai接口正常的时候传1
	if respCode == 0 || respCode == 600 {
		infoclog.isRec = strconv.Itoa(1)
	}
	s.infoc(infoclog)
}

type storyClickInfoc struct {
	ip        string
	now       string
	api       string
	buvid     string
	mid       string
	client    string
	aid       string
	displayID string
	err       string
	from      string
	build     string
	trackid   string
	autoplay  string
	fromSpmid string
	spmid     string
}

func (s *StoryService) StoryClickInfoc(ctx context.Context, api, buvid string, mid, aid int64, plat int8, build, displayID,
	from, autoPlay int, trackID, fromSpmid, spmid string, now time.Time) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	infoclog := storyClickInfoc{
		ip:        ip,
		now:       strconv.FormatInt(now.Unix(), 10),
		api:       api,
		buvid:     buvid,
		mid:       strconv.FormatInt(mid, 10),
		client:    strconv.Itoa(int(plat)),
		aid:       strconv.FormatInt(aid, 10),
		displayID: strconv.Itoa(displayID),
		err:       strconv.Itoa(ecode.Cause(nil).Code()),
		from:      strconv.Itoa(from),
		build:     strconv.Itoa(build),
		trackid:   trackID,
		autoplay:  strconv.Itoa(autoPlay),
		fromSpmid: fromSpmid,
		spmid:     spmid,
	}
	s.infoc(infoclog)
}

// nolint: gocognit,errcheck
func (s *StoryService) infocproc() {
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case storyInfoc:
			showlist, _ := json.Marshal(l.list)
			storyBiz, _ := json.Marshal(l.bizInfo)
			event := infocV2.NewLogStreamV("004071",
				log.String(l.ip),
				log.String(l.now),
				log.String(l.api),
				log.String(l.buvid),
				log.String(l.mid),
				log.String(l.client),
				log.String(l.pull),
				log.String(string(showlist)),
				log.String(l.isRec),
				log.String(l.build),
				log.String(l.returnCode),
				log.String(l.deviceID),
				log.String(l.network),
				log.String(l.trackID),
				log.String(l.userFeature),
				log.String(l.displayID),
				log.String(l.fromAv),
				log.String(l.fromTrackid),
				log.String(string(storyBiz)),
			)
			if err := s.infocV2Log.Info(context.Background(), event); err != nil {
				log.Error("Failed to infoc story: %s, %s, %s, %+v", l.mid, l.buvid, l.build, err)
			}
		case storyClickInfoc:
			event := infocV2.NewLogStreamV("000025",
				log.String(l.ip),
				log.String(l.now),
				log.String(l.api),
				log.String(l.buvid),
				log.String(l.mid),
				log.String(l.client),
				log.String(l.aid),
				log.String(""),
				log.String(l.err),
				log.String(l.from),
				log.String(l.build),
				log.String(l.trackid),
				log.String(l.autoplay),
				log.String(l.fromSpmid),
				log.String(l.spmid),
			)
			if err := s.infocV2Log.Info(context.Background(), event); err != nil {
				log.Error("Failed to infoc story click: %s, %s, %s, %s, %+v", l.mid, l.buvid, l.aid, l.build, err)
			}
		}
	}
}
