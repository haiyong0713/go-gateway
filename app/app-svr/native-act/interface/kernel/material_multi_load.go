package kernel

import (
	"strconv"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (ml *MaterialLoader) MultiLoad() *Material {
	material := &Material{}
	multiML, ok := ml.remove2MultiLoader()
	if !ok {
		return ml.Load(material)
	}
	multiML.Load(material)
	addSourceOfMixExts(material, ml, multiML)
	addSourceOfPageArcs(material, ml)
	addSourceOfSelSerie(material, ml)
	addSourceOfActLikes(material, ml, multiML)
	addSourceOfRelInfos(material, ml, multiML)
	addSourceOfBriefDyns(material, ml, multiML)
	addSourceOfChannelFeed(material, ml, multiML)
	addSourceOfUpList(material, ml, multiML)
	addSourceOfRankRst(material, ml, multiML)
	addSourceOfUpRsvInfo(material, ml, multiML)
	addSourceOfActSubject(material, ml, multiML)
	addSourceOfSourceDetail(material, ml, multiML)
	ml.Load(material)
	return material
}

func (ml *MaterialLoader) remove2MultiLoader() (*MaterialLoader, bool) {
	multiML := NewMaterialLoader(ml.c, ml.dep, ml.ss)
	var ok bool
	if len(ml.mixExtsReqs) > 0 {
		multiML.mixExtsReqs = ml.mixExtsReqs
		ml.mixExtsReqs = nil
		ok = true
	}
	if len(ml.pageArcsReqs) > 0 {
		multiML.pageArcsReqs = ml.pageArcsReqs
		ml.mixExtsReqs = nil
		ok = true
	}
	if len(ml.selSerieReqs) > 0 {
		multiML.selSerieReqs = ml.selSerieReqs
		ml.selSerieReqs = nil
		ok = true
	}
	if len(ml.actLikesReqs) > 0 {
		multiML.actLikesReqs = ml.actLikesReqs
		ml.actLikesReqs = nil
		ok = true
	}
	if len(ml.relInfosReqs) > 0 {
		multiML.relInfosReqs = ml.relInfosReqs
		ml.relInfosReqs = nil
		ok = true
	}
	if len(ml.briefDynsReqs) > 0 {
		multiML.briefDynsReqs = ml.briefDynsReqs
		ml.briefDynsReqs = nil
		ok = true
	}
	if len(ml.channelFeedReqs) > 0 {
		multiML.channelFeedReqs = ml.channelFeedReqs
		ml.channelFeedReqs = nil
		ok = true
	}
	if len(ml.upListReqs) > 0 {
		multiML.upListReqs = ml.upListReqs
		ml.upListReqs = nil
		ok = true
	}
	if len(ml.rankRstReqs) > 0 {
		multiML.rankRstReqs = ml.rankRstReqs
		ml.rankRstReqs = nil
		ok = true
	}
	if len(ml.upRsvIDsReqs) > 0 {
		multiML.upRsvIDsReqs = ml.upRsvIDsReqs
		ml.upRsvIDsReqs = nil
		ok = true
	}
	if len(ml.actSidsReqs) > 0 {
		multiML.actSidsReqs = ml.actSidsReqs
		ml.actSidsReqs = nil
		ok = true
	}
	if len(ml.sourceDetailReqs) > 0 {
		multiML.sourceDetailReqs = ml.sourceDetailReqs
		ml.sourceDetailReqs = nil
		ok = true
	}
	return multiML, ok
}

func addSourceOfMixExts(material *Material, ml, multiML *MaterialLoader) {
	var (
		mids    []int64
		aids    []int64
		epids   []int64
		cvids   []int64
		fids    []int64
		ssids   []int32
		roomIDs map[int64][]int64
		playAvs []*arcgrpc.PlayAv
	)
	for reqID, rly := range material.MixExtsRlys {
		req, ok := multiML.mixExtsReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, ext := range rly.List {
			if ext == nil || ext.ForeignID == 0 {
				continue
			}
			switch ext.MType {
			case natpagegrpc.MixTypeRcmd:
				mids = append(mids, ext.ForeignID)
			case natpagegrpc.MixAvidType:
				if req.ArcType == model.MaterialArchive {
					aids = append(aids, ext.ForeignID)
				} else {
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: ext.ForeignID})
				}
			case natpagegrpc.MixEpidType:
				epids = append(epids, ext.ForeignID)
			case natpagegrpc.MixCvidType:
				cvids = append(cvids, ext.ForeignID)
			case natpagegrpc.MixFolder:
				if mixEditor, err := model.UnmarshalMixExtEditor(ext.Reason); err == nil && mixEditor.Fid > 0 {
					aids = append(aids, ext.ForeignID)
					fids = append(fids, mixEditor.Fid)
				}
			case natpagegrpc.MixLive:
				if roomIDs == nil {
					roomIDs = make(map[int64][]int64)
				}
				roomIDs[req.IsLive] = append(roomIDs[req.IsLive], ext.ForeignID)
			case natpagegrpc.MixOgvSsid:
				ssids = append(ssids, int32(ext.ForeignID))
			}
		}
	}
	_ = ml.addMids(mids)
	_ = ml.addAids(aids)
	_ = ml.addEpids(epids)
	_ = ml.addCvids(cvids)
	_ = ml.addFolderIDs(fids, favmdl.TypeVideo)
	for isLive, ids := range roomIDs {
		_ = ml.addRoomIDs(ids, isLive)
	}
	_ = ml.addPlayAvs(playAvs)
	_ = ml.addSsids(ssids)
}

func addSourceOfPageArcs(material *Material, ml *MaterialLoader) {
	var (
		aids []int64
		fids []int64
	)
	for _, pageArcs := range material.PageArcsRlys {
		for _, item := range pageArcs.GetList() {
			aids = append(aids, item.GetAid())
		}
		fids = append(fids, pageArcs.GetMediaId())
	}
	_ = ml.addAids(aids)
	_ = ml.addFolderIDs(fids, favmdl.TypeVideo)
}

func addSourceOfSelSerie(material *Material, ml *MaterialLoader) {
	var (
		aids []int64
		fids []int64
	)
	for _, rly := range material.SelSerieRlys {
		if len(rly.List) == 0 {
			continue
		}
		if rly.Config != nil && rly.Config.MediaId > 0 {
			fids = append(fids, rly.Config.MediaId)
		}
		for _, res := range rly.List {
			if res == nil || res.Rid == 0 {
				continue
			}
			if res.Rtype == model.SelRtypeArchive {
				aids = append(aids, res.Rid)
			}
		}
	}
	_ = ml.addAids(aids)
	_ = ml.addFolderIDs(fids, favmdl.TypeVideo)
}

func addSourceOfActLikes(material *Material, ml, multiML *MaterialLoader) {
	var (
		aids, cvids []int64
		playAvs     []*arcgrpc.PlayAv
	)
	for reqID, rly := range material.ActLikesRlys {
		req, ok := multiML.actLikesReqs[reqID]
		if !ok || !req.NeedMultiML || rly.Subject == nil {
			continue
		}
		for _, v := range rly.List {
			if v == nil || v.Item == nil || v.Item.Wid == 0 {
				continue
			}
			switch rly.Subject.Type {
			case model.ActSubTypeArticle:
				cvids = append(cvids, v.Item.Wid)
			case model.ActSubTypeVideoLike, model.ActSubTypeVideo2, model.ActSubTypePhoneVideo:
				if req.ArcType == model.MaterialArchive {
					aids = append(aids, v.Item.Wid)
				} else {
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.Item.Wid})
				}
			}
		}
	}
	_ = ml.addAids(aids)
	_ = ml.addCvids(cvids)
	_ = ml.addPlayAvs(playAvs)
}

func addSourceOfRelInfos(material *Material, ml, multiML *MaterialLoader) {
	var epids []int64
	for reqID, rly := range material.RelInfosRlys {
		req, ok := multiML.relInfosReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		var ids []int64
		for _, info := range rly.GetInfos() {
			for _, epList := range info.GetCharacterEp() {
				for _, ep := range epList.GetCharacterEp() {
					if ep == nil || ep.GetEpId() == 0 {
						continue
					}
					ids = append(ids, int64(ep.GetEpId()))
				}
			}
		}
		if req.ShowNum > 0 && int64(len(ids)) > req.ShowNum {
			ids = ids[:req.ShowNum]
		}
		epids = append(epids, ids...)
	}
	_ = ml.addEpids(epids)
}

func addSourceOfBriefDyns(material *Material, ml, multiML *MaterialLoader) {
	var (
		aids, cvids []int64
		playAvs     []*arcgrpc.PlayAv
	)
	for reqID, rly := range material.BriefDynsRlys {
		req, ok := multiML.briefDynsReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, dyn := range rly.Dynamics {
			if dyn == nil || dyn.Rid == 0 {
				continue
			}
			switch dyn.Type {
			case model.DynTypeArticle:
				cvids = append(cvids, dyn.Rid)
			case model.DynTypeVideo:
				if req.ArcType == model.MaterialArchive {
					aids = append(aids, dyn.Rid)
				} else {
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: dyn.Rid})
				}
			}
		}
	}
	_ = ml.addAids(aids)
	_ = ml.addCvids(cvids)
	_ = ml.addPlayAvs(playAvs)
}

func addSourceOfChannelFeed(material *Material, ml, multiML *MaterialLoader) {
	var aids, epids []int64
	for reqID, rly := range material.ChannelFeedRlys {
		req, ok := multiML.channelFeedReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, resource := range rly.GetList() {
			if resource == nil || resource.GetId() <= 0 {
				continue
			}
			switch resource.GetType() {
			case hmtgrpc.ResourceType_UGC_RESOURCE:
				aids = append(aids, resource.GetId())
			case hmtgrpc.ResourceType_OGV_RESOURCE:
				epids = append(epids, resource.GetId())
			default:
				continue
			}
		}
	}
	_ = ml.addAids(aids)
	_ = ml.addEpids(epids)
}

func addSourceOfUpList(material *Material, ml, multiML *MaterialLoader) {
	var mids []int64
	for reqID, rly := range material.UpListRlys {
		req, ok := multiML.upListReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, item := range rly.List {
			if item == nil {
				continue
			}
			if item.Account != nil {
				mids = append(mids, item.Account.Mid)
			}
		}
	}
	_ = ml.addRelFids(mids)
	_ = ml.addCardMids(mids)
}

func addSourceOfRankRst(material *Material, ml, multiML *MaterialLoader) {
	var mids []int64
	for reqID, rly := range material.RankRstRlys {
		req, ok := multiML.rankRstReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, item := range rly.List {
			if item == nil || item.Account == nil || item.ObjectType != 1 {
				continue
			}
			mids = append(mids, item.Account.MID)
		}
	}
	_ = ml.addRelFids(mids)
}

func addSourceOfUpRsvInfo(material *Material, ml, multiML *MaterialLoader) {
	var (
		mids, aids []int64
		uidLiveIDs = make(map[int64][]string)
	)
	for reqID, rly := range material.UpRsvInfos {
		req, ok := multiML.upRsvIDsReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, info := range rly {
			if info == nil {
				continue
			}
			if info.UpActVisible != activitygrpc.UpActVisible_DefaultVisible {
				continue
			}
			if info.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedWaitCallBack ||
				info.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackCancel ||
				info.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				switch info.Type {
				case activitygrpc.UpActReserveRelationType_Archive:
					if aid, err := strconv.ParseInt(info.Oid, 10, 64); err == nil && aid > 0 {
						aids = append(aids, aid)
					}
				case activitygrpc.UpActReserveRelationType_Live:
					uidLiveIDs[info.Upmid] = append(uidLiveIDs[info.Upmid], info.Oid)
				default:
					continue
				}
			}
			if info.Upmid > 0 {
				mids = append(mids, info.Upmid)
			}
		}
	}
	_ = ml.addCardMids(mids)
	_ = ml.addAids(aids)
	_ = ml.addUidLiveIDs(uidLiveIDs)
}

func addSourceOfActSubject(material *Material, ml, multiML *MaterialLoader) {
	var mids []int64
	for reqID, rly := range material.ActSubjects {
		req, ok := multiML.actSidsReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, info := range rly {
			if info == nil {
				continue
			}
			if req.NeedAccount && info.ActivityInitiator > 0 {
				mids = append(mids, info.ActivityInitiator)
			}
		}
	}
	_ = ml.addCardMids(mids)
}

func addSourceOfSourceDetail(material *Material, ml, multiML *MaterialLoader) {
	var (
		aids  []int64
		epids []int64
		cvids []int64
		fids  []int64
	)
	for reqID, rly := range material.SourceDetailRlys {
		req, ok := multiML.sourceDetailReqs[reqID]
		if !ok || !req.NeedMultiML {
			continue
		}
		for _, item := range rly.ItemList {
			if item == nil || item.ItemId == 0 {
				continue
			}
			switch item.Type {
			case natpagegrpc.MixAvidType:
				aids = append(aids, item.ItemId)
			case natpagegrpc.MixEpidType:
				epids = append(epids, item.ItemId)
			case natpagegrpc.MixCvidType:
				cvids = append(cvids, item.ItemId)
			case natpagegrpc.MixFolder:
				aids = append(aids, item.ItemId)
				fids = append(fids, item.Fid)
			}
		}
	}
	_ = ml.addAids(aids)
	_ = ml.addEpids(epids)
	_ = ml.addCvids(cvids)
	_ = ml.addFolderIDs(fids, favmdl.TypeVideo)
}
