package dynamicV2

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	infocV2 "go-common/library/log/infoc.v2"
)

type rcmdInfoc struct {
	ip               string
	network          string
	time             string
	api              string
	buvid            string
	mid              string
	plat             string
	build            string
	pageType         string
	rootSource       string
	showlist         []*mdlv2.RcmdInfoItem
	userFeature      string
	returnCode       string
	trackid          string
	zoneid           string
	tabCampusID      string
	userCampusID     string
	previousCampusID string
	freshType        string
}

func (s *Service) campusRcmdFeedInfoc(_ context.Context, general *mdlv2.GeneralParam, param *api.CampusRcmdFeedReq, aiInfo *mdlv2.RcmdInfo, list []*mdlv2.FoldItem) {
	if aiInfo == nil || len(aiInfo.Listm) == 0 {
		return
	}
	var (
		items   []*mdlv2.RcmdInfoItem
		trackId string
	)
	rootSource := "dt"
	if param.FromType == api.CampusReqFromType_HOME {
		rootSource = "homepage"
	}
	for i, v := range list {
		dynId, _ := strconv.ParseInt(v.Item.GetExtend().GetDynIdStr(), 10, 64)
		aiInfo, ok := aiInfo.Listm[dynId]
		if !ok {
			continue
		}
		trackId = aiInfo.TrackID
		tmp := &mdlv2.RcmdInfoItem{}
		*tmp = *aiInfo
		tmp.Pos = i
		items = append(items, tmp)
	}
	infoclog := rcmdInfoc{
		ip:               general.Network.RemoteIP,
		network:          general.GetNetWork(),
		time:             strconv.FormatInt(time.Now().Unix(), 10),
		api:              "/bilibili.app.dynamic.v2.Dynamic/CampusRcmdFeed",
		buvid:            general.GetBuvid(),
		mid:              strconv.FormatInt(general.Mid, 10),
		plat:             strconv.Itoa(int(model.Plat(general.GetMobiApp(), general.GetDevice()))),
		build:            strconv.Itoa(int(general.GetBuild())),
		pageType:         "moment",
		rootSource:       rootSource,
		showlist:         items,
		userFeature:      string(aiInfo.UserFeature),
		returnCode:       strconv.Itoa(aiInfo.Code),
		trackid:          trackId,
		zoneid:           strconv.FormatInt(aiInfo.ZoneID, 10),
		tabCampusID:      strconv.FormatInt(param.CampusId, 10),
		userCampusID:     strconv.FormatInt(aiInfo.SchoolID, 10),
		previousCampusID: "0",
		freshType:        "3",
	}
	if param.Scroll == 0 {
		infoclog.freshType = "2"
	}
	s.infoc(infoclog)
}

func (s *Service) campusRecommendInfoc(_ context.Context, general *mdlv2.GeneralParam, param *api.CampusRecommendReq, aiInfo *mdlv2.RcmdInfo, list []*api.RcmdItem) {
	if aiInfo == nil || len(aiInfo.Listm) == 0 {
		return
	}
	var (
		items   []*mdlv2.RcmdInfoItem
		trackId string
	)
	for i, v := range list {
		if v.GetRcmdArchive() == nil {
			continue
		}
		aiInfo, ok := aiInfo.Listm[v.GetRcmdArchive().Aid]
		if !ok {
			continue
		}
		trackId = aiInfo.TrackID
		tmp := &mdlv2.RcmdInfoItem{}
		*tmp = *aiInfo
		tmp.Pos = i
		items = append(items, tmp)
	}
	infoclog := rcmdInfoc{
		ip:               general.Network.RemoteIP,
		network:          general.GetNetWork(),
		time:             strconv.FormatInt(time.Now().Unix(), 10),
		api:              "/bilibili.app.dynamic.v2.Dynamic/CampusRecommend",
		buvid:            general.GetBuvid(),
		mid:              strconv.FormatInt(general.Mid, 10),
		plat:             strconv.Itoa(int(model.Plat(general.GetMobiApp(), general.GetDevice()))),
		build:            strconv.Itoa(int(general.GetBuild())),
		pageType:         "nearby",
		rootSource:       campusRcmdFrom2SubPageType(param.From),
		showlist:         items,
		userFeature:      string(aiInfo.UserFeature),
		returnCode:       strconv.Itoa(aiInfo.Code),
		trackid:          trackId,
		zoneid:           strconv.FormatInt(aiInfo.ZoneID, 10),
		tabCampusID:      "0",
		userCampusID:     strconv.FormatInt(aiInfo.SchoolID, 10),
		previousCampusID: strconv.FormatInt(param.CampusId, 10),
		freshType:        "null",
	}
	s.infoc(infoclog)
}

func (s *Service) campusWaterFlowInfoc(_ context.Context, general *mdlv2.GeneralParam, param *api.WaterFlowRcmdReq, aiReqParam map[string]string, aiInfo *mdlv2.RcmdInfo, list []*api.CampusWaterFlowItem) {
	if aiInfo == nil || len(aiInfo.Listm) == 0 {
		return
	}
	var (
		items   []*mdlv2.RcmdInfoItem
		trackId string
	)
	for i, v := range list {
		itemDefault := v.GetItemDefault()
		if itemDefault == nil || itemDefault.Annotations == nil {
			continue
		}
		dynIdStr, ok := itemDefault.Annotations["dynamic_id"]
		if !ok {
			continue
		}
		dynId, _ := strconv.ParseInt(dynIdStr, 10, 64)
		aiInfo, ok := aiInfo.Listm[dynId]
		if !ok {
			continue
		}

		trackId = aiInfo.TrackID
		tmp := &mdlv2.RcmdInfoItem{}
		*tmp = *aiInfo
		tmp.Pos = i
		items = append(items, tmp)
	}
	infoclog := rcmdInfoc{
		ip:               general.Network.RemoteIP,
		network:          general.GetNetWork(),
		time:             strconv.FormatInt(time.Now().Unix(), 10),
		api:              "/bilibili.app.dynamic.v2.Campus/WaterFlowRcmd",
		buvid:            general.GetBuvid(),
		mid:              strconv.FormatInt(general.Mid, 10),
		plat:             strconv.Itoa(int(model.Plat(general.GetMobiApp(), general.GetDevice()))),
		build:            strconv.Itoa(int(general.GetBuild())),
		pageType:         "nearby",
		rootSource:       campusRcmdFrom2SubPageType(param.From),
		showlist:         items,
		userFeature:      string(aiInfo.UserFeature),
		returnCode:       strconv.Itoa(aiInfo.Code),
		trackid:          trackId,
		zoneid:           strconv.FormatInt(aiInfo.ZoneID, 10),
		tabCampusID:      "0",
		userCampusID:     strconv.FormatInt(aiInfo.SchoolID, 10),
		previousCampusID: strconv.FormatInt(param.CampusId, 10),
		freshType:        aiReqParam["fresh_type"],
	}
	s.infoc(infoclog)
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

func (s *Service) infocproc() {
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case rcmdInfoc:
			showlist, _ := json.Marshal(l.showlist)
			event := infocV2.NewLogStreamV(s.c.Infoc.AiSchoolInfocID,
				log.String(l.ip),
				log.String(l.network),
				log.String(l.time),
				log.String(l.api),
				log.String(l.buvid),
				log.String(l.mid),
				log.String(l.plat),
				log.String(l.build),
				log.String(l.pageType),
				log.String(l.rootSource),
				log.String(string(showlist)),
				log.String(l.userFeature),
				log.String(l.returnCode),
				log.String(l.trackid),
				log.String(l.zoneid),
				log.String(l.tabCampusID),
				log.String(l.userCampusID),
				log.String(l.previousCampusID),
				log.String(l.freshType),
			)
			if err := s.infocV2.Info(context.Background(), event); err != nil {
				log.Error("Failed to infoc AiSchool click: %s, %s, %s, %s, %+v", l.mid, l.buvid, l.userFeature, l.build, err)
			}
		}
	}

}
