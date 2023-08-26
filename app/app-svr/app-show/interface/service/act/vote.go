package act

import (
	"context"
	"sort"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"

	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	"go-gateway/app/web-svr/native-page/interface/api"
)

func (s *Service) formatVote(c context.Context, mou *api.NativeModule, acts *api.Click, mid int64) *actmdl.Item {
	confSort := mou.ConfUnmarshal()
	switch confSort.SourceType {
	case api.SourceTypeVoteUp:
		return s.formatVoteUp(c, mou, acts, mid)
	case api.SourceTypeVoteAct, "": //兼容老数据
		return s.formatVoteAct(c, mou, acts, mid)
	default:
	}
	return nil
}

func (s *Service) formatVoteAct(c context.Context, mou *api.NativeModule, acts *api.Click, mid int64) *actmdl.Item {
	//目前只支持二选一模式
	confSort := mou.ConfUnmarshal()
	var (
		ps = 2
	)
	voteRly, err := s.actDao.GetVoteActivityRank(c, mou.Fid, confSort.Sid, 1, int64(ps), 3, mid)
	if err != nil {
		log.Error("s.actDao.GetVoteActivityRank(%d,%d,%d) error(%v)", mou.Fid, confSort.Sid, mid, err)
		return nil
	}
	if voteRly == nil || len(voteRly.Rank) < ps { //返回数据缺失，不下发改组件
		return nil
	}
	var clickItem []*actmdl.Item
	if len(acts.Areas) > 0 {
		var (
			firExRankInfo = make(map[int64]*actgrpc.ExternalRankInfo)
			k             = int64(0)
		)
		for _, v := range voteRly.Rank {
			if v == nil {
				continue
			}
			firExRankInfo[k] = v
			k++
		}
		//按照x轴排序
		sort.Slice(acts.Areas, func(i, j int) bool {
			return acts.Areas[i].Leftx < acts.Areas[j].Leftx
		})
		buttonNum := int64(0)
		tipObj := &actmdl.TipCancel{}
		tipObj.FromVoteTip("投票")
		for _, v := range acts.Areas {
			var (
				ext  *actmdl.ClickExt
				dTmp = &actmdl.Item{}
			)
			switch {
			case v.IsVoteButton():
				//判断是否已投
				var (
					isFollow bool
					itemID   int64
				)
				if val, ok := firExRankInfo[buttonNum]; ok && val != nil {
					if mid > 0 {
						isFollow = val.UserCanVoteCount == 0
					}
					itemID = val.SourceItemId
				}
				buttonNum++
				ext = &actmdl.ClickExt{Tip: tipObj, Goto: actmdl.GotoClickVote, IsFollow: isFollow, FID: mou.Fid, GroupID: confSort.Sid, ItemID: itemID, FollowIcon: confSort.Image, UnFollowIcon: v.UnfinishedImage}
				dTmp.FromMVote(v, ext)
			case v.IsVoteUser():
				if mid == 0 { //未登录不展示
					continue
				}
				ext = &actmdl.ClickExt{Num: voteRly.UserAvailVoteCount}
				dTmp.FromMVote(v, ext)
			case v.IsVoteProcess(): //进度数值展示
				dTmp.FromMVoteProcess(v, firExRankInfo, mou)
			default:
				continue
			}
			clickItem = append(clickItem, dTmp)
		}
	}
	res := &actmdl.Item{}
	res.FromVote(mou, clickItem)
	return res
}

func (s *Service) formatVoteUp(c context.Context, mou *api.NativeModule, acts *api.Click, mid int64) *actmdl.Item {
	if mou.Fid <= 0 || acts == nil {
		return nil
	}
	voteInfo, err := s.dynvoteDao.ListFeedVote(c, mou.Fid, mid)
	if err != nil || voteInfo == nil || voteInfo.Status != 1 || len(voteInfo.Options) < actmdl.VoteOptionNum {
		return nil
	}
	myVotes := upMyVotes(voteInfo.MyVotes)
	//按照x轴排序
	sort.Slice(acts.Areas, func(i, j int) bool {
		return acts.Areas[i].Leftx < acts.Areas[j].Leftx
	})
	confSort := mou.ConfUnmarshal()
	tipObj := &actmdl.TipCancel{}
	tipObj.FromVoteTip("投票")
	var buttonNo int64
	areaItems := make([]*actmdl.Item, 0, len(acts.Areas))
	for _, v := range acts.Areas {
		item := &actmdl.Item{}
		switch {
		case v.IsVoteButton():
			if buttonNo >= actmdl.VoteOptionNum {
				continue
			}
			optIdx := voteInfo.Options[buttonNo].OptIdx
			var isFollow bool
			if _, ok := myVotes[optIdx]; ok {
				isFollow = true
			}
			ext := &actmdl.ClickExt{
				Tip:           tipObj,
				Goto:          actmdl.GotoClickVoteUp,
				IsFollow:      isFollow,
				FID:           mou.Fid,
				ItemID:        int64(optIdx),
				FollowIcon:    confSort.Image,
				UnFollowIcon:  v.UnfinishedImage,
				CancelDisable: true,
				Toast:         "您已经投过票了",
			}
			item.FromMVote(v, ext)
			buttonNo++
		case v.IsVoteProcess():
			item.FromUpVoteProgress(v, voteInfo.Options, mou)
		default:
			continue
		}
		areaItems = append(areaItems, item)
	}
	voteItem := &actmdl.Item{}
	voteItem.FromVote(mou, areaItems)
	return voteItem
}

func upMyVotes(votes []int32) map[int32]struct{} {
	res := make(map[int32]struct{}, len(votes))
	for _, vote := range votes {
		res[vote] = struct{}{}
	}
	return res
}
