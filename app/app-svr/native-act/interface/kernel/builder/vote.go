package builder

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
)

type Vote struct{}

func (bu Vote) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	vtCfg, ok := cfg.(*config.Vote)
	if !ok {
		logCfgAssertionError(config.Vote{})
		return nil
	}
	var items []*api.ModuleItem
	switch vtCfg.SourceType {
	case model.SourceTypeVoteAct, "":
		items = bu.buildModuleItemsOfAct(vtCfg, material, ss)
	case model.SourceTypeVoteUp:
		items = bu.buildModuleItemsOfUp(vtCfg, material)
	}
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeVote.String(),
		ModuleId:      vtCfg.ModuleBase().ModuleID,
		ModuleSetting: &api.Setting{DisplayNum: vtCfg.DisplayNum},
		ModuleItems:   items,
		ModuleUkey:    vtCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Vote) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Vote) buildModuleItemsOfAct(cfg *config.Vote, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	vRankRly, ok := material.VoteRankRlys[cfg.VoteRankReqID]
	if !ok || len(vRankRly.Rank) < model.VoteOptionNum {
		return nil
	}
	cd := &api.VoteCard{
		BgImage:   cfg.BgImage.ToSizeImage(),
		OptionNum: model.VoteOptionNum,
		Buttons:   bu.buildVoteButtonOfAct(cfg, vRankRly.Rank, ss.Mid()),
		LeftNum:   bu.buildVoteLeftNum(cfg, vRankRly.UserAvailVoteCount, ss.Mid()),
		Progress: bu.buildVoteProgress(cfg, func(i int) (num int64, sourceItemID int64) {
			return vRankRly.Rank[i].Vote, vRankRly.Rank[i].SourceItemId
		}),
	}
	item := &api.ModuleItem{
		CardType:   model.CardTypeVote.String(),
		CardDetail: &api.ModuleItem_VoteCard{VoteCard: cd},
	}
	return []*api.ModuleItem{item}
}

func (bu Vote) buildModuleItemsOfUp(cfg *config.Vote, material *kernel.Material) []*api.ModuleItem {
	voteInfo, ok := material.DynVoteInfos[cfg.Sid]
	if !ok || voteInfo == nil || voteInfo.Status != 1 || len(voteInfo.Options) < model.VoteOptionNum {
		return nil
	}
	cd := &api.VoteCard{
		BgImage:   cfg.BgImage.ToSizeImage(),
		OptionNum: model.VoteOptionNum,
		Buttons:   bu.buildVoteButtonOfUp(cfg, voteInfo),
		Progress: bu.buildVoteProgress(cfg, func(i int) (num int64, sourceItemID int64) {
			return int64(voteInfo.Options[i].Cnt), int64(voteInfo.Options[i].OptIdx)
		}),
	}
	item := &api.ModuleItem{
		CardType:   model.CardTypeVote.String(),
		CardDetail: &api.ModuleItem_VoteCard{VoteCard: cd},
	}
	return []*api.ModuleItem{item}
}

func (bu Vote) transArea(cfg *config.Area) *api.Area {
	return &api.Area{Height: cfg.Height, Width: cfg.Width, X: cfg.X, Y: cfg.Y, Ukey: cfg.Ukey}
}

func (bu Vote) buildVoteButtonOfAct(cfg *config.Vote, ranks []*activitygrpc.ExternalRankInfo, mid int64) []*api.VoteButton {
	vts := make([]*api.VoteButton, 0, model.VoteOptionNum)
	for i, v := range cfg.VoteButtons {
		if i >= model.VoteOptionNum {
			continue
		}
		var hasVoted bool
		if mid > 0 {
			hasVoted = ranks[i].UserCanVoteCount == 0
		}
		vt := &api.VoteButton{
			Area:         bu.transArea(&v.Area),
			DoneImage:    cfg.DoneButtonImage,
			UndoneImage:  v.UndoneImage,
			HasVoted:     hasVoted,
			MessageBox:   &api.MessageBox{Type: api.MessageBoxType_Dialog, AlertMsg: "是否取消投票？", ConfirmButtonText: "确定", CancelButtonText: "取消"},
			SourceItemId: ranks[i].SourceItemId,
		}
		vt.VoteParams = bu.voteParams(model.SourceTypeVoteAct, cfg.Sid, cfg.Gid, ranks[i].SourceItemId, hasVoted)
		vts = append(vts, vt)
	}
	return vts
}

func (bu Vote) buildVoteLeftNum(cfg *config.Vote, availNum, mid int64) *api.VoteNum {
	if mid <= 0 || cfg.VoteLeftNum == nil {
		return nil
	}
	return &api.VoteNum{
		Area: bu.transArea(&cfg.VoteLeftNum.Area),
		Num:  availNum,
	}
}

func (bu Vote) buildVoteProgress(cfg *config.Vote, progressItem func(i int) (num int64, sourceItemID int64)) *api.VoteProgress {
	if cfg.VoteProgress == nil {
		return nil
	}
	progress := &api.VoteProgress{
		Area:  bu.transArea(&cfg.VoteProgress.Area),
		Style: bu.progressStyle(cfg.VoteProgress.Style),
		Items: make([]*api.VoteProgress_VoteProgressItem, 0, model.VoteOptionNum),
	}
	for i, color := range cfg.VoteProgress.OptionColors {
		if i >= model.VoteOptionNum {
			continue
		}
		num, sourceItemID := progressItem(i)
		progress.Items = append(progress.Items, &api.VoteProgress_VoteProgressItem{
			Color:        color,
			Num:          num,
			SourceItemId: sourceItemID,
		})
	}
	return progress
}

func (bu Vote) buildVoteButtonOfUp(cfg *config.Vote, voteInfo *dyncommongrpc.VoteInfo) []*api.VoteButton {
	myVotes := bu.myVotesOfUp(voteInfo.MyVotes)
	vts := make([]*api.VoteButton, 0, model.VoteOptionNum)
	for i, v := range cfg.VoteButtons {
		if i >= model.VoteOptionNum {
			continue
		}
		optIdx := voteInfo.Options[i].OptIdx
		var hasVoted bool
		if _, ok := myVotes[optIdx]; ok {
			hasVoted = true
		}
		vt := &api.VoteButton{
			Area:         bu.transArea(&v.Area),
			DoneImage:    cfg.DoneButtonImage,
			UndoneImage:  v.UndoneImage,
			HasVoted:     hasVoted,
			MessageBox:   &api.MessageBox{Type: api.MessageBoxType_Toast, AlertMsg: "您已经投过票了"},
			SourceItemId: int64(optIdx),
		}
		vt.VoteParams = bu.voteParams(model.SourceTypeVoteUp, cfg.Sid, 0, int64(optIdx), hasVoted)
		vts = append(vts, vt)
	}
	return vts
}

func (bu Vote) voteParams(typ string, sid, gid, sourceItemID int64, hasVoted bool) string {
	action := api.ActionType_Do
	if hasVoted {
		action = api.ActionType_Undo
	}
	return passthrough.Marshal(&api.VoteParams{Action: action, Type: typ, Sid: sid, Gid: gid, SourceItemId: sourceItemID})
}

func (bu Vote) myVotesOfUp(votes []int32) map[int32]struct{} {
	res := make(map[int32]struct{}, len(votes))
	for _, vote := range votes {
		res[vote] = struct{}{}
	}
	return res
}

var _voteProgressStyle = map[string]api.VoteProgressStyle{
	model.VPStyleCircle: api.VoteProgressStyle_VPStyleCircle,
	model.VPStyleSquare: api.VoteProgressStyle_VPStyleSquare,
}

func (bu Vote) progressStyle(in string) api.VoteProgressStyle {
	if out, ok := _voteProgressStyle[in]; ok {
		return out
	}
	return api.VoteProgressStyle_VPStyleDefault
}
