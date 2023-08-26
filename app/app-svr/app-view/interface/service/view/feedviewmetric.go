package view

import (
	"strconv"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/dao/archive"
	"go-gateway/app/app-svr/app-view/interface/model"

	"go-common/library/log"
	"go-common/library/log/infoc.v2"
)

const SlideViewLogID = "010882"

//easyjson:json
type ShowList struct {
	Section ShowListSection `json:"section"`
}

type ShowListSection struct {
	ID       string                 `json:"id"`
	FromItem string                 `json:"from_item"`
	Items    []*ShowListSectionItem `json:"items"`
}

type ShowListSectionItem struct {
	ID        int64  `json:"id"`
	Pos       int64  `json:"pos"`
	Goto      string `json:"goto"`
	Source    string `json:"source"`
	AVFeature string `json:"av_feature"`
}

func anyTrackID(in *archive.SlidesReply) string {
	for _, i := range in.Data {
		return i.TrackID
	}
	return ""
}

func succeedRecommend(in *archive.SlidesReply) bool {
	return !in.IsBackupReply()
}

func makeShowList(params *FeedItemParams, slidesReply *archive.SlidesReply, reply *viewApi.FeedViewReply) string {
	out := &ShowList{}
	out.Section.ID = "相关视频"
	out.Section.FromItem = strconv.FormatInt(params.FeedViewReq.Aid, 10)

	allItems := map[int64]*ShowListSectionItem{}
	for _, v := range slidesReply.Data {
		allItems[v.ID] = &ShowListSectionItem{
			ID:        v.ID,
			Goto:      v.Goto,
			Source:    v.Source,
			AVFeature: v.AVFeature,
		}
	}

	for i, v := range reply.List {
		if v.View.Arc == nil {
			continue
		}
		item, ok := allItems[v.View.Arc.Aid]
		if !ok {
			continue
		}
		item.Pos = int64(i + 1)
		out.Section.Items = append(out.Section.Items, item)
	}

	jsonBytes, err := out.MarshalJSON()
	if err != nil {
		log.Error("Failed to marshal showlist as json: %+v", err)
		return ""
	}
	return string(jsonBytes)
}

func makeSlideViewEventPayload(params *FeedItemParams, slidesReply *archive.SlidesReply, reply *viewApi.FeedViewReply) infoc.Payload {
	return infoc.NewLogStreamV(SlideViewLogID,
		log.KVString("ip", params.IP),
		log.KVString("time", strconv.FormatInt(params.Now.Unix(), 10)),
		log.KVString("api", model.RPCPathFeedView),
		log.KVString("buvid", params.Device.Buvid),
		log.KVString("mid", strconv.FormatInt(params.AuthN.Mid, 10)),
		log.KVString("device_id", ""),
		log.KVString("client", strconv.FormatInt(int64(params.Plat), 10)),
		log.KVString("mobi_app", params.Device.RawMobiApp),
		log.KVString("build", strconv.FormatInt(params.Device.Build, 10)),
		log.KVString("network", params.Net()),
		log.KVString("trackid", anyTrackID(slidesReply)),
		log.KVString("from_trackid", params.FeedViewReq.FromTrackId),
		log.KVString("from_av", strconv.FormatInt(params.FeedViewReq.Aid, 10)),
		log.KVString("display_id", strconv.FormatInt(params.FeedViewReq.DisplayId, 10)),
		log.KVString("session_id", params.FeedViewReq.SessionId),
		log.KVString("showlist", makeShowList(params, slidesReply, reply)),
		log.KVInt("is_rec", bool2int(succeedRecommend(slidesReply))),
		log.KVString("return_code", strconv.FormatInt(slidesReply.ReturnCode(), 10)),
		log.KVString("spmid", params.FeedViewReq.Spmid),
		log.KVString("from_spmid", params.FeedViewReq.FromSpmid),
		log.KVString("source_page", params.FeedViewReq.From),
		log.KVString("pv_feature", slidesReply.PVFeature),
	)
}

//go:generate easyjson -all feedviewmetric.go
