package native

import (
	"context"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	chagrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/admin/model"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
)

func (s *Service) FindCounters(c context.Context, act string) (*natmdl.FindCountersReply, error) {
	rly, err := s.actplatClient.CountersInActivity(c, &actplatapi.CountersInActivityReq{Activity: act})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return &natmdl.FindCountersReply{List: rly.Counters}, nil
}

func (s *Service) ChannelDetail(c context.Context, id int64, ps int32, buvid string) (*natmdl.HmtChannelRly, error) {
	chaRly, err := s.dao.ChannelFeed(c, id, 0, buvid, 0, ps)
	if err != nil {
		return nil, err
	}
	if chaRly == nil || len(chaRly.List) == 0 {
		return nil, ecode.NothingFound
	}
	//拼接aids和epids
	rly := &natmdl.HmtChannelRly{}
	for _, v := range chaRly.List {
		if v == nil || v.Id <= 0 {
			continue
		}
		tmp := &natmdl.HmtChannel{ID: v.Id}
		switch v.Type {
		case chagrpc.ResourceType_UGC_RESOURCE:
			tmp.Type = "UGC"
		case chagrpc.ResourceType_OGV_RESOURCE:
			tmp.Type = "OGV"
		default:
			continue
		}
		rly.IDs = append(rly.IDs, tmp)
	}
	return rly, nil
}

func (s *Service) CartoonDetail(c context.Context, id int64) (*model.ComicItem, error) {
	//获取展示平台
	list, err := s.addao.GetComicInfos(c, []int64{id})
	if err != nil {
		log.Error("s.addao.GetComicInfos(%d) error(%v)", id, err)
		return nil, err
	}
	if lv, ok := list[id]; !ok || lv == nil {
		return nil, ecode.NothingFound
	}
	return list[id], nil
}

func (s *Service) GameDetail(c context.Context, id int64) (*model.FormatItem, error) {
	//获取展示平台
	list, err := s.addao.GameList(c, []int64{id})
	if err != nil {
		log.Error("s.addao.GameList(%d) error(%v)", id, err)
		return nil, err
	}
	if lv, ok := list[id]; !ok || lv == nil {
		return nil, ecode.NothingFound
	}
	//平台类型：0=PC，1=安卓，2=IOS
	var realPlatform string
	switch {
	case list[id].IsShowAndroid == 1:
		realPlatform = "1"
	case list[id].IsShowIos == 1:
		realPlatform = "2"
	case list[id].IsShowPc == 1:
		realPlatform = "0"
	default:
		return nil, ecode.NothingFound
	}
	rly, err := s.addao.MultiGameInfo(c, []int64{id}, 0, realPlatform)
	if err != nil {
		log.Error(" s.addao.MultiGameInfo(%d,%s) error(%v)", id, realPlatform, err)
		return nil, err
	}
	if kv, ok := rly[id]; !ok || kv == nil {
		return nil, ecode.NothingFound
	}
	tmpItem := &model.FormatItem{}
	tmpItem.FromGameExt(rly[id])
	return tmpItem, nil
}

func (s *Service) ReserveDetail(c context.Context, id int64) (*natmdl.ReserveRly, error) {
	//获取展示平台
	list, err := s.dao.UpActReserveRelationInfo(c, 0, []int64{id})
	if err != nil {
		log.Error("s.dao.UpActReserveRelationInfo(%d) error(%v)", id, err)
		return nil, err
	}
	if lv, ok := list[id]; !ok || lv == nil {
		return nil, ecode.NothingFound
	}
	res := &natmdl.ReserveRly{
		SID:   list[id].Sid,
		Title: list[id].Title,
		Type:  int32(list[id].Type),
		Mid:   list[id].Upmid,
	}
	if list[id].Upmid > 0 { //账号信息获取失败，降级处理
		if accRly, err := s.dao.Info3(c, list[id].Upmid); err == nil && accRly != nil {
			res.Name = accRly.Name
		}
	}
	return res, nil
}

func (s *Service) TsPage(c context.Context, req *natmdl.TsPageReq) (*natmdl.TsPageRly, error) {
	pageDyn, err := s.dao.DynExtByPid(c, req.PageID)
	if err != nil {
		return nil, err
	}
	return &natmdl.TsPageRly{Dynamic: pageDyn.Dynamic}, nil
}

func (s *Service) UpVote(c context.Context, voteID int64) (*dyncommongrpc.VoteInfo, error) {
	infos, err := s.dao.ListFeedVotes(c, &dynvotegrpc.ListFeedVotesReq{VoteIds: []int64{voteID}})
	if err != nil {
		return nil, err
	}
	vote, ok := infos.GetVoteInfos()[voteID]
	if !ok || vote == nil {
		return nil, ecode.NothingFound
	}
	return vote, nil
}
