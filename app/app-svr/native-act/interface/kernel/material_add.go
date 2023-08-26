package kernel

import (
	"fmt"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

// weekIDs:[]int64
func (ml *MaterialLoader) addWeekIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	weekIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("weekIDs is not []int64, material_type=%+v", model.MaterialWeeks))
	}
	if len(weekIDs) == 0 {
		return nil
	}
	ml.weekIDs = append(ml.weekIDs, weekIDs...)
	return nil
}

// gameIDs:[]int64
func (ml *MaterialLoader) addGameIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	gameIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("gameIDs is not []int64, material_type=%+v", model.MaterialGame))
	}
	if len(gameIDs) == 0 {
		return nil
	}
	ml.gameIDs = append(ml.gameIDs, gameIDs...)
	return nil
}

// liveIDs:[]int64
func (ml *MaterialLoader) addLiveIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	liveIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("liveIDs is not []int64, material_type=%+v", model.MaterialLive))
	}
	if len(liveIDs) == 0 {
		return nil
	}
	ml.liveIDs = append(ml.liveIDs, liveIDs...)
	return nil
}

// aids:[]int64
func (ml *MaterialLoader) addAids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	aids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("aids is not []int64, material_type=%+v", model.MaterialArchive))
	}
	if len(aids) == 0 {
		return nil
	}
	ml.aids = append(ml.aids, aids...)
	return nil
}

// roomIDs:[]int64, isLive:int64
func (ml *MaterialLoader) addRoomIDs(data ...interface{}) error {
	if len(data) < _roomParams {
		return nil
	}
	roomIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("roomIDs is not []int64, material_type=%+v", model.MaterialLiveRoom))
	}
	if len(roomIDs) == 0 {
		return nil
	}
	isLive, ok := data[1].(int64)
	if !ok {
		return errors.New(fmt.Sprintf("isLive is not int64, material_type=%+v", model.MaterialLiveRoom))
	}
	if ml.roomIDs == nil {
		ml.roomIDs = make(map[int64][]int64)
	}
	ml.roomIDs[isLive] = append(ml.roomIDs[isLive], roomIDs...)
	return nil
}

// cvids:[]int64
func (ml *MaterialLoader) addCvids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	cvids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("cvids is not []int64, material_type=%+v", model.MaterialArticle))
	}
	if len(cvids) == 0 {
		return nil
	}
	ml.cvids = append(ml.cvids, cvids...)
	return nil
}

// epids:[]int64
func (ml *MaterialLoader) addEpids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	epids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("epids is not []int64, material_type=%+v", model.MaterialEpisode))
	}
	if len(epids) == 0 {
		return nil
	}
	ml.epids = append(ml.epids, epids...)
	return nil
}

// folderIDs:[]int64, type:int32
func (ml *MaterialLoader) addFolderIDs(data ...interface{}) error {
	if len(data) < _folderParams {
		return nil
	}
	folderIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("folderIDs is not []int64, material_type=%+v", model.MaterialFolder))
	}
	if len(folderIDs) == 0 {
		return nil
	}
	folderType, ok := data[1].(int32)
	if !ok {
		return errors.New(fmt.Sprintf("folderType is not int32, material_type=%+v", model.MaterialFolder))
	}
	if ml.folderIDs == nil {
		ml.folderIDs = make(map[int32][]int64)
	}
	ml.folderIDs[folderType] = append(ml.folderIDs[folderType], folderIDs...)
	return nil
}

// actSubProtoIDs:[]int64
func (ml *MaterialLoader) addActSubProtoIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	actSubProtoIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("actSubProtoIDs is not []int64, material_type=%+v", model.MaterialActSubProto))
	}
	if len(actSubProtoIDs) == 0 {
		return nil
	}
	ml.actSubProtoIDs = append(ml.actSubProtoIDs, actSubProtoIDs...)
	return nil
}

// mids:[]int64
func (ml *MaterialLoader) addMids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	mids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("mids is not []int64, material_type=%+v", model.MaterialAccount))
	}
	if len(mids) == 0 {
		return nil
	}
	ml.mids = append(ml.mids, mids...)
	return nil
}

// dyntopicgrpc.HasDynsReq
func (ml *MaterialLoader) addHasDynsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*dyntopicgrpc.HasDynsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not dyntopicgrpc.HasDynsReq, material_type=%+v", model.MaterialHasDynsRly))
	}
	if ml.hasDynsReqs == nil {
		ml.hasDynsReqs = make(map[RequestID]*dyntopicgrpc.HasDynsReq)
	}
	reqID := requestID()
	ml.hasDynsReqs[reqID] = req
	return reqID, nil
}

// dyntopicgrpc.ListDynsReq
func (ml *MaterialLoader) addListDynsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*dyntopicgrpc.ListDynsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not dyntopicgrpc.ListDynsReq, material_type=%+v", model.MaterialListDynsRly))
	}
	if ml.listDynsReqs == nil {
		ml.listDynsReqs = make(map[RequestID]*dyntopicgrpc.ListDynsReq)
	}
	reqID := requestID()
	ml.listDynsReqs[reqID] = req
	return reqID, nil
}

// appdyngrpc.DynServerDetailsReq
func (ml *MaterialLoader) addDynDetailReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*appdyngrpc.DynServerDetailsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not appdyngrpc.DynServerDetailsReq, material_type=%+v", model.MaterialDynDetail))
	}
	if ml.dynDetailReqs == nil {
		ml.dynDetailReqs = make(map[RequestID]*appdyngrpc.DynServerDetailsReq)
	}
	reqID := requestID()
	ml.dynDetailReqs[reqID] = req
	return reqID, nil
}

// kernel.ActLikesReq
func (ml *MaterialLoader) addActLikesReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*ActLikesReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.ActLikesReq, material_type=%+v", model.MaterialActLikesRly))
	}
	if ml.actLikesReqs == nil {
		ml.actLikesReqs = make(map[RequestID]*ActLikesReq)
	}
	reqID := requestID()
	ml.actLikesReqs[reqID] = req
	return reqID, nil
}

// dynfeedgrpc.FetchDynIdByRevsReq
func (ml *MaterialLoader) addDynRevsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*dynfeedgrpc.FetchDynIdByRevsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not dynfeedgrpc.FetchDynIdByRevsReq, material_type=%+v", model.MaterialDynRevsRly))
	}
	if ml.dynRevsReqs == nil {
		ml.dynRevsReqs = make(map[RequestID]*dynfeedgrpc.FetchDynIdByRevsReq)
	}
	reqID := requestID()
	ml.dynRevsReqs[reqID] = req
	return reqID, nil
}

// tagIDs:[]int64
func (ml *MaterialLoader) addTagIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	tagIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("tagIDs is not []int64, material_type=%+v", model.MaterialTag))
	}
	if len(tagIDs) == 0 {
		return nil
	}
	ml.tagIDs = append(ml.tagIDs, tagIDs...)
	return nil
}

// kernel.ModuleMixExtsReq
func (ml *MaterialLoader) addMixExtsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*ModuleMixExtsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.ModuleMixExtsReq, material_type=%+v", model.MaterialMixExtsRly))
	}
	if ml.mixExtsReqs == nil {
		ml.mixExtsReqs = make(map[RequestID]*ModuleMixExtsReq)
	}
	reqID := requestID()
	ml.mixExtsReqs[reqID] = req
	return reqID, nil
}

// actplatv2grpc.GetHistoryReq
func (ml *MaterialLoader) addGetHisReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*actplatv2grpc.GetHistoryReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not actplatv2grpc.GetHistoryReq, material_type=%+v", model.MaterialGetHisRly))
	}
	if ml.getHisReqs == nil {
		ml.getHisReqs = make(map[RequestID]*actplatv2grpc.GetHistoryReq)
	}
	reqID := requestID()
	ml.getHisReqs[reqID] = req
	return reqID, nil
}

// populargrpc.PageArcsReq
func (ml *MaterialLoader) addPageArcsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*populargrpc.PageArcsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not populargrpc.PageArcsReq, material_type=%+v", model.MaterialPageArcsRly))
	}
	if ml.pageArcsReqs == nil {
		ml.pageArcsReqs = make(map[RequestID]*populargrpc.PageArcsReq)
	}
	reqID := requestID()
	ml.pageArcsReqs[reqID] = req
	return reqID, nil
}

// natpagegrpc.ModuleMixExtReq
func (ml *MaterialLoader) addMixExtReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*natpagegrpc.ModuleMixExtReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not natpagegrpc.ModuleMixExtReq, material_type=%+v", model.MaterialMixExtRly))
	}
	if ml.mixExtReqs == nil {
		ml.mixExtReqs = make(map[RequestID]*natpagegrpc.ModuleMixExtReq)
	}
	reqID := requestID()
	ml.mixExtReqs[reqID] = req
	return reqID, nil
}

// kernel.RankResultReq
func (ml *MaterialLoader) addRankRstReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*RankResultReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.RankResultReq, material_type=%+v", model.MaterialRankRstRly))
	}
	if ml.rankRstReqs == nil {
		ml.rankRstReqs = make(map[RequestID]*RankResultReq)
	}
	reqID := requestID()
	ml.rankRstReqs[reqID] = req
	return reqID, nil
}

// appshowgrpc.SelectedSerieReq
func (ml *MaterialLoader) addSelSerieReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*appshowgrpc.SelectedSerieReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not appshowgrpc.SelectedSerieReq, material_type=%+v", model.MaterialSelSerieRly))
	}
	if ml.selSerieReqs == nil {
		ml.selSerieReqs = make(map[RequestID]*appshowgrpc.SelectedSerieReq)
	}
	reqID := requestID()
	ml.selSerieReqs[reqID] = req
	return reqID, nil
}

// kernel.UpListReq
func (ml *MaterialLoader) addUpListReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*UpListReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.UpListReq, material_type=%+v", model.MaterialUpListRly))
	}
	if ml.upListReqs == nil {
		ml.upListReqs = make(map[RequestID]*UpListReq)
	}
	reqID := requestID()
	ml.upListReqs[reqID] = req
	return reqID, nil
}

// kernel.RelInfosReq
func (ml *MaterialLoader) addRelInfosReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*RelInfosReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.RelInfosReq, material_type=%+v", model.MaterialRelInfosRly))
	}
	if ml.relInfosReqs == nil {
		ml.relInfosReqs = make(map[RequestID]*RelInfosReq)
	}
	reqID := requestID()
	ml.relInfosReqs[reqID] = req
	return reqID, nil
}

// kernel.BriefDynsReq
func (ml *MaterialLoader) addBriefDynsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*BriefDynsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.BriefDynsReq, material_type=%+v", model.MaterialBriefDynsRly))
	}
	if ml.briefDynsReqs == nil {
		ml.briefDynsReqs = make(map[RequestID]*BriefDynsReq)
	}
	reqID := requestID()
	ml.briefDynsReqs[reqID] = req
	return reqID, nil
}

// wids:[]int32
func (ml *MaterialLoader) addWids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	wids, ok := data[0].([]int32)
	if !ok {
		return errors.New(fmt.Sprintf("wids is not []int32, material_type=%+v", model.MaterialQueryWidRly))
	}
	if len(wids) == 0 {
		return nil
	}
	ml.wids = append(ml.wids, wids...)
	return nil
}

// liveplaygrpc.GetListByActIdReq
func (ml *MaterialLoader) addRoomsByActIdReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*liveplaygrpc.GetListByActIdReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not liveplaygrpc.GetListByActIdReq, material_type=%+v", model.MaterialRoomsByActIdRly))
	}
	if ml.roomsByActIdReqs == nil {
		ml.roomsByActIdReqs = make(map[RequestID]*liveplaygrpc.GetListByActIdReq)
	}
	reqID := requestID()
	ml.roomsByActIdReqs[reqID] = req
	return reqID, nil
}

// kernel.ChannelFeedReq
func (ml *MaterialLoader) addChannelFeedReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*ChannelFeedReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.ChannelFeedReq, material_type=%+v", model.MaterialChannelFeedRly))
	}
	if ml.channelFeedReqs == nil {
		ml.channelFeedReqs = make(map[RequestID]*ChannelFeedReq)
	}
	reqID := requestID()
	ml.channelFeedReqs[reqID] = req
	return reqID, nil
}

// playAvs:[]*arcgrpc.PlayAv
func (ml *MaterialLoader) addPlayAvs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	playAvs, ok := data[0].([]*arcgrpc.PlayAv)
	if !ok {
		return errors.New(fmt.Sprintf("playAvs is not []*arcgrpc.PlayAv, material_type=%+v", model.MaterialArcPlayer))
	}
	if len(playAvs) == 0 {
		return nil
	}
	ml.playAvs = append(ml.playAvs, playAvs...)
	return nil
}

// relFids:[]int64
func (ml *MaterialLoader) addRelFids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	fids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("relFids is not []int64, material_type=%+v", model.MaterialRelation))
	}
	if len(fids) == 0 {
		return nil
	}
	ml.relFids = append(ml.relFids, fids...)
	return nil
}

// cardMids:[]int64
func (ml *MaterialLoader) addCardMids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	cardMids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("cardMids is not []int64, material_type=%+v", model.MaterialAccountCard))
	}
	if len(cardMids) == 0 {
		return nil
	}
	ml.cardMids = append(ml.cardMids, cardMids...)
	return nil
}

// pidsOfNaCard:[]int64
func (ml *MaterialLoader) addPidsOfNaCard(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	pids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("pidsOfNaCard is not []int64, material_type=%+v", model.MaterialNativeCard))
	}
	if len(pids) == 0 {
		return nil
	}
	ml.pidsOfNaCard = append(ml.pidsOfNaCard, pids...)
	return nil
}

// pidsOfNaAll:[]int64
func (ml *MaterialLoader) addPidsOfNaAll(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	pids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("pidsOfNaAll is not []int64, material_type=%+v", model.MaterialNativeAllPage))
	}
	if len(pids) == 0 {
		return nil
	}
	ml.pidsOfNaAll = append(ml.pidsOfNaAll, pids...)
	return nil
}

// pidsOfNaPages:[]int64
func (ml *MaterialLoader) addPidsOfNaPages(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	pids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("pidsOfNaPages is not []int64, material_type=%+v", model.MaterialNativePages))
	}
	if len(pids) == 0 {
		return nil
	}
	ml.pidsOfNaPages = append(ml.pidsOfNaPages, pids...)
	return nil
}

// channelIDs:[]int64
func (ml *MaterialLoader) addChannelIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("channelIDs is not []int64, material_type=%+v", model.MaterialChannel))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.channelIDs = append(ml.channelIDs, ids...)
	return nil
}

// activitygrpc.GetVoteActivityRankReq
func (ml *MaterialLoader) addVoteRankReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*activitygrpc.GetVoteActivityRankReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not activitygrpc.GetVoteActivityRankReq, material_type=%+v", model.MaterialVoteRankRly))
	}
	if ml.VoteRankReqs == nil {
		ml.VoteRankReqs = make(map[RequestID]*activitygrpc.GetVoteActivityRankReq)
	}
	reqID := requestID()
	ml.VoteRankReqs[reqID] = req
	return reqID, nil
}

// kernel.UpRsvIDsReq
func (ml *MaterialLoader) addUpRsvIDsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*UpRsvIDsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.UpRsvIDsReq, material_type=%+v", model.MaterialUpRsvInfo))
	}
	if len(req.IDs) == 0 {
		return "", nil
	}
	if ml.upRsvIDsReqs == nil {
		ml.upRsvIDsReqs = make(map[RequestID]*UpRsvIDsReq)
	}
	reqID := requestID()
	ml.upRsvIDsReqs[reqID] = req
	return reqID, nil
}

// uidLiveIDs:map[int64][]string
func (ml *MaterialLoader) addUidLiveIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].(map[int64][]string)
	if !ok {
		return errors.New(fmt.Sprintf("uidLiveIDs is not map[int64][]string, material_type=%+v", model.MaterialRoomSessionInfo))
	}
	if ml.uidLiveIDs == nil {
		ml.uidLiveIDs = make(map[int64][]string)
	}
	for mid, liveIDs := range ids {
		ml.uidLiveIDs[mid] = append(ml.uidLiveIDs[mid], liveIDs...)
	}
	return nil
}

// populargrpc.TimeLineRequest
func (ml *MaterialLoader) addTimelineReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*populargrpc.TimeLineRequest)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not populargrpc.TimeLineRequest, material_type=%+v", model.MaterialTimelineRly))
	}
	if ml.timelineReqs == nil {
		ml.timelineReqs = make(map[RequestID]*populargrpc.TimeLineRequest)
	}
	reqID := requestID()
	ml.timelineReqs[reqID] = req
	return reqID, nil
}

// ssids:[]int32
func (ml *MaterialLoader) addSsids(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ssids, ok := data[0].([]int32)
	if !ok {
		return errors.New(fmt.Sprintf("ssids is not []int32, material_type=%+v", model.MaterialSeasonCard))
	}
	if len(ssids) == 0 {
		return nil
	}
	ml.ssids = append(ml.ssids, ssids...)
	return nil
}

// pgcappgrpc.SeasonByPlayIdReq
func (ml *MaterialLoader) addSeasonByPlayIdReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*pgcappgrpc.SeasonByPlayIdReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not pgcappgrpc.SeasonByPlayIdReq, material_type=%+v", model.MaterialSeasonByPlayIdRly))
	}
	if ml.seasonByPlayIdReqs == nil {
		ml.seasonByPlayIdReqs = make(map[RequestID]*pgcappgrpc.SeasonByPlayIdReq)
	}
	reqID := requestID()
	ml.seasonByPlayIdReqs[reqID] = req
	return reqID, nil
}

// model.ActiveUsersReq
func (ml *MaterialLoader) addActiveUsersReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*model.ActiveUsersReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not model.ActiveUsersReq, material_type=%+v", model.MaterialActiveUsersRly))
	}
	if ml.activeUsersReqs == nil {
		ml.activeUsersReqs = make(map[RequestID]*model.ActiveUsersReq)
	}
	reqID := requestID()
	ml.activeUsersReqs[reqID] = req
	return reqID, nil
}

// dynVoteIDs:[]int64
func (ml *MaterialLoader) addDynVoteIDs(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	dynVoteIDs, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("dynVoteIDs is not []int64, material_type=%+v", model.MaterialDynVoteInfo))
	}
	if len(dynVoteIDs) == 0 {
		return nil
	}
	ml.dynVoteIDs = append(ml.dynVoteIDs, dynVoteIDs...)
	return nil
}

// kernel.ActSidsReq
func (ml *MaterialLoader) addActSidsReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*ActSidsReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not kernel.ActSidsReq, material_type=%+v", model.MaterialActSubject))
	}
	if len(req.IDs) == 0 {
		return "", nil
	}
	if ml.actSidsReqs == nil {
		ml.actSidsReqs = make(map[RequestID]*ActSidsReq)
	}
	reqID := requestID()
	ml.actSidsReqs[reqID] = req
	return reqID, nil
}

// actSidGroupIDs: sid int64, gids:[]int64
func (ml *MaterialLoader) addActSidGroupIDs(data ...interface{}) error {
	if len(data) < _actProgGroupParams {
		return nil
	}
	sid, ok := data[0].(int64)
	if !ok {
		return errors.New(fmt.Sprintf("sid is not int64, material_type=%+v", model.MaterialActProgressGroup))
	}
	if sid == 0 {
		return nil
	}
	gids, ok := data[1].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("gids is not []int64, material_type=%+v", model.MaterialActProgressGroup))
	}
	if len(gids) == 0 {
		return nil
	}
	if ml.actSidGroupIDs == nil {
		ml.actSidGroupIDs = make(map[int64][]int64)
	}
	ml.actSidGroupIDs[sid] = append(ml.actSidGroupIDs[sid], gids...)
	return nil
}

// SourceDetailReq
func (ml *MaterialLoader) addSourceDetailReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*SourceDetailReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not SourceDetailReq, material_type=%+v", model.MaterialSourceDetail))
	}
	if ml.sourceDetailReqs == nil {
		ml.sourceDetailReqs = make(map[RequestID]*SourceDetailReq)
	}
	reqID := requestID()
	ml.sourceDetailReqs[reqID] = req
	return reqID, nil
}

// model.ProductDetailReq
func (ml *MaterialLoader) addProductDetailReq(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*model.ProductDetailReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not model.ProductDetailReq, material_type=%+v", model.MaterialProductDetail))
	}
	if ml.productDetailReqs == nil {
		ml.productDetailReqs = make(map[RequestID]*model.ProductDetailReq)
	}
	reqID := requestID()
	ml.productDetailReqs[reqID] = req
	return reqID, nil
}

// pgcFollowSeasonIds:[]int32
func (ml *MaterialLoader) addPgcFollowSeasonIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int32)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int32, material_type=%+v", model.MaterialPgcFollowStatus))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.pgcFollowSeasonIds = append(ml.pgcFollowSeasonIds, ids...)
	return nil
}

// comicIds:[]int64
func (ml *MaterialLoader) addComicIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialComicInfo))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.comicIds = append(ml.comicIds, ids...)
	return nil
}

// actRsvIds:[]int64
func (ml *MaterialLoader) addActRsvIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialActReserveFollow))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.actRsvIds = append(ml.actRsvIds, ids...)
	return nil
}

// awardIds:[]int64
func (ml *MaterialLoader) addAwardIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialActAwardState))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.awardIds = append(ml.awardIds, ids...)
	return nil
}

// ticketFavIds:[]int64
func (ml *MaterialLoader) addTicketFavIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialTicketFavState))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.ticketFavIds = append(ml.ticketFavIds, ids...)
	return nil
}

// actRelationIds:[]int64
func (ml *MaterialLoader) addActRelationIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialActRelationInfo))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.actRelationIds = append(ml.actRelationIds, ids...)
	return nil
}

// actplatv2grpc.GetCounterResReq
func (ml *MaterialLoader) addPlatCounterResReqs(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*actplatv2grpc.GetCounterResReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not actplatv2grpc.GetCounterResReq, material_type=%+v", model.MaterialPlatCounterRes))
	}
	if ml.platCounterReqs == nil {
		ml.platCounterReqs = make(map[RequestID]*actplatv2grpc.GetCounterResReq)
	}
	reqID := requestID()
	ml.platCounterReqs[reqID] = req
	return reqID, nil
}

// actplatv2grpc.GetTotalResReq
func (ml *MaterialLoader) addPlatTotalResReqs(data ...interface{}) (RequestID, error) {
	if len(data) == 0 {
		return "", nil
	}
	req, ok := data[0].(*actplatv2grpc.GetTotalResReq)
	if !ok {
		return "", errors.New(fmt.Sprintf("req is not actplatv2grpc.GetTotalResReq, material_type=%+v", model.MaterialPlatTotalRes))
	}
	if ml.platTotalReqs == nil {
		ml.platTotalReqs = make(map[RequestID]*actplatv2grpc.GetTotalResReq)
	}
	reqID := requestID()
	ml.platTotalReqs[reqID] = req
	return reqID, nil
}

// lotteryIds:[]string
func (ml *MaterialLoader) addLotteryIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]string)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []string, material_type=%+v", model.MaterialLotteryUnused))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.lotteryIds = append(ml.lotteryIds, ids...)
	return nil
}

// scoreIds:[]int64
func (ml *MaterialLoader) addScoreIds(data ...interface{}) error {
	if len(data) == 0 {
		return nil
	}
	ids, ok := data[0].([]int64)
	if !ok {
		return errors.New(fmt.Sprintf("ids is not []int64, material_type=%+v", model.MaterialScoreTarget))
	}
	if len(ids) == 0 {
		return nil
	}
	ml.scoreIds = append(ml.scoreIds, ids...)
	return nil
}
