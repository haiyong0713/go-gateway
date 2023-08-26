package dynamicV2

import (
	"context"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func (s *Service) statInfo(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) *api.ModuleStat {
	stat := &api.ModuleStat{
		Repost: dynCtx.Dyn.Repost, // 禁止转发会置为0
	}
	var (
		reply, like, share int64
		replyuri           string
		islike             bool
	)
	switch {
	case dynCtx.Dyn.IsForward():
		reply, like, replyuri, islike = s.statForward(dynCtx)
	case dynCtx.Dyn.IsAv():
		reply, like, replyuri, islike = s.statAV(dynCtx)
	case dynCtx.Dyn.IsPGC():
		reply, like, replyuri, islike = s.statPGC(dynCtx)
	case dynCtx.Dyn.IsCheeseBatch():
		reply, like, replyuri, islike = s.statCourBatch(dynCtx)
	case dynCtx.Dyn.IsWord():
		reply, like, replyuri, islike = s.statWord(dynCtx)
	case dynCtx.Dyn.IsDraw():
		reply, like, replyuri, islike = s.statDraw(dynCtx)
	case dynCtx.Dyn.IsArticle():
		reply, like, replyuri, islike = s.statArticle(dynCtx)
	case dynCtx.Dyn.IsMusic():
		reply, like, replyuri, islike = s.statMusic(dynCtx)
	case dynCtx.Dyn.IsCommon():
		reply, like, replyuri, islike = s.statCommon(dynCtx)
	case dynCtx.Dyn.IsAD():
		share, reply, like, islike = s.statAD(c, general, dynCtx)
	case dynCtx.Dyn.IsApplet():
		reply, like, replyuri, islike = s.statApplet(dynCtx)
	case dynCtx.Dyn.IsSubscription(), dynCtx.Dyn.IsSubscriptionNew():
		reply, like, replyuri, islike = s.statSubscription(dynCtx)
	case dynCtx.Dyn.IsLiveRcmd():
		reply, like, replyuri, islike = s.statLiveRcmd(dynCtx)
	case dynCtx.Dyn.IsUGCSeason():
		reply, like, replyuri, islike = s.statUGCSeason(dynCtx)
	case dynCtx.Dyn.IsBatch():
		reply, like, replyuri, islike = s.statBatch(dynCtx)
	case dynCtx.Dyn.IsCourUp():
		reply, like, replyuri, islike = s.statCourUp(dynCtx)
	}
	stat.Reply = reply // 禁止评论会置为0
	stat.Like = like
	stat.ReplyUrl = replyuri
	if share > 0 {
		stat.Repost = share
	}
	if islike {
		if stat.LikeInfo == nil {
			stat.LikeInfo = new(api.LikeInfo)
		}
		stat.LikeInfo.IsLike = true
		// 点赞业务的点赞计数是异步更新的
		// 已点赞都情况下兜底返回个点赞数1
		if stat.Like == 0 {
			stat.Like = 1
		}
	}
	// 点赞动画
	if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.LikeIcon != nil {
		if stat.LikeInfo == nil {
			stat.LikeInfo = new(api.LikeInfo)
		}
		stat.LikeInfo.Animation = &api.LikeAnimation{
			Begin:      dynCtx.Dyn.Extend.LikeIcon.Begin,
			Proc:       dynCtx.Dyn.Extend.LikeIcon.Proc,
			End:        dynCtx.Dyn.Extend.LikeIcon.End,
			LikeIconId: dynCtx.Dyn.Extend.LikeIcon.NewIconID,
		}
	}
	// 优先，评论服务端-禁止评论
	if replyInfo, ok := dynCtx.GetReply(); ok && replyInfo.NoComment {
		stat.NoComment = true
		stat.Reply = 0
		stat.NoCommentText = s.c.Resource.Text.ModuleStatNoComment
	} else if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.CommentBan == 1 { // 否则，动态服务端-禁止评论
		stat.NoComment = true
		stat.Reply = 0
		stat.NoCommentText = s.c.Resource.Text.ModuleStatNoComment
	}
	// 附加大卡稿件首映前禁评文案
	if stat.NoComment && dynCtx.IsAttachCardPremiereBefore() {
		stat.NoCommentText = "首映开始后开放评论区，先点击卡片进入首映室和大家聊聊吧"
	}
	// 禁止转发
	if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.RepostBan == 1 {
		stat.NoForward = true
		stat.Repost = 0
		stat.NoForwardText = s.c.Resource.Text.ModuleStatNoForward
	} else if dynCtx.Dyn.IsForward() && dynCtx.Interim.ForwardOrigFaild {
		stat.NoForward = true
		stat.Repost = 0
		stat.NoForwardText = s.c.Resource.Text.ModuleStatNoForwardForwardFaild
	}
	return stat
}

func (s *Service) stat(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_stat,
		ModuleItem: &api.Module_ModuleStat{
			ModuleStat: s.statInfo(c, dynCtx, general),
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) statShell(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_stat_forward,
		ModuleItem: &api.Module_ModuleStatForward{
			ModuleStatForward: s.statInfo(c, dynCtx, general),
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) statForward(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statAV(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		like = int64(archive.Stat.Like)
		reply = int64(archive.Stat.Reply)
	}
	islike, _ = s.getLike(dynCtx)
	return
}

func (s *Service) statPGC(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid)); ok {
		reply = int64(pgc.Stat.Reply)
	}
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statCourBatch(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if batch, ok := dynCtx.GetResCheeseBatch(dynCtx.Dyn.Rid); ok {
		reply = int64(batch.NewEp.Reply)
	}
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statCourUp(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if batch, ok := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid); ok {
		reply = int64(batch.Stat.Reply)
	}
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statWord(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statDraw(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid); ok {
		reply = int64(draw.Item.Reply)
	}
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statArticle(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	article, _ := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
	reply = article.Stats.Reply
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statMusic(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	music, _ := dynCtx.GetResMusic(dynCtx.Dyn.Rid)
	reply = int64(music.ReplyCnt)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statCommon(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statBatch(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statAD(c context.Context, general *mdlv2.GeneralParam, dynCtx *mdlv2.DynamicContext) (share, reply, like int64, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	var (
		aid       int64
		showReply bool
	)
	// 广告起飞卡
	if ok := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynAdFly, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynAdFlyIOS) ||
			(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynAdFlyAndroid)}); ok &&
		dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		aid = dynCtx.Dyn.PassThrough.AdAvid
		showReply = true
	} else {
		// 创作ID获取视频ID在获取视频的点赞信息
		creativeID := dynCtx.Dyn.Rid
		var ok bool
		aid, ok = dynCtx.ResCreativeIDs[creativeID]
		if !ok {
			return
		}
	}
	ap, ok := dynCtx.GetArchive(aid)
	if !ok {
		return
	}
	var archive = ap.Arc
	like = int64(archive.Stat.Like)
	// 把广告转换成视频卡获取是否点赞
	tmpdynCtx := &mdlv2.DynamicContext{}
	*tmpdynCtx = *dynCtx
	tmpdynCtx.Dyn.Type = mdlv2.DynTypeVideo
	tmpdynCtx.Dyn.Rid = aid
	islike, _ = s.getLike(tmpdynCtx)
	if showReply {
		// 评论数
		reply = int64(archive.Stat.Reply)
		share = int64(archive.Stat.Share)
	}
	return
}

func (s *Service) statApplet(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statSubscription(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statLiveRcmd(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = s.getReply(dynCtx)
	islike, like = s.getLike(dynCtx)
	return
}

func (s *Service) statUGCSeason(dynCtx *mdlv2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		like = int64(archive.Stat.Like)
		reply = int64(archive.Stat.Reply)
	}
	islike, _ = s.getLike(dynCtx)
	return
}

func (s *Service) getLike(dynCtx *mdlv2.DynamicContext) (bool, int64) {
	var (
		busParam *mdlv2.ThumbsRecord
		busType  string
		isThum   bool
	)
	if dynCtx.Dyn.IsPGC() {
		if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid)); ok {
			if busParam, busType, isThum = mdlv2.GetPGCLikeID(pgc); !isThum {
				return false, 0
			}
		}
	} else {
		if busParam, busType, isThum = dynCtx.Dyn.GetLikeID(); !isThum {
			return false, 0
		}
	}
	if busParam == nil {
		return false, 0
	}
	if dynCtx.ResLike != nil && dynCtx.ResLike[busType] != nil {
		if r, ok := dynCtx.ResLike[busType].Records[busParam.MsgID]; ok {
			if r.LikeState == thumgrpc.State_STATE_LIKE {
				return true, r.LikeNumber
			}
			return false, r.LikeNumber
		}
	}
	return false, 0
}

func (s *Service) getReply(dynCtx *mdlv2.DynamicContext) int64 {
	if replyTmp, ok := dynCtx.GetReply(); ok {
		return replyTmp.Count
	}
	return 0
}

func (s *Service) statFake(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	stat := &api.Module_ModuleStat{ModuleStat: &api.ModuleStat{}}
	if dynCtx.Dyn.IsAv() {
		stat.ModuleStat.NoComment = true
		stat.ModuleStat.NoCommentText = "视频转码中，暂不能操作噢~"
		stat.ModuleStat.NoForward = true
		stat.ModuleStat.NoForwardText = "视频转码中，暂不能操作噢~"
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_stat,
		ModuleItem: stat,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) statRepost(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	var (
		like   int64
		islike bool
	)
	stat := &api.Module_ModuleStat{ModuleStat: &api.ModuleStat{
		Repost: dynCtx.Dyn.Repost,
	}}
	islike, like = s.getLike(dynCtx)
	stat.ModuleStat.Like = like
	if islike {
		if stat.ModuleStat.LikeInfo == nil {
			stat.ModuleStat.LikeInfo = new(api.LikeInfo)
		}
		stat.ModuleStat.LikeInfo.IsLike = true
	}
	// 点赞动画
	if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.LikeIcon != nil {
		if stat.ModuleStat.LikeInfo == nil {
			stat.ModuleStat.LikeInfo = new(api.LikeInfo)
		}
		stat.ModuleStat.LikeInfo.Animation = &api.LikeAnimation{
			Begin:      dynCtx.Dyn.Extend.LikeIcon.Begin,
			Proc:       dynCtx.Dyn.Extend.LikeIcon.Proc,
			End:        dynCtx.Dyn.Extend.LikeIcon.End,
			LikeIconId: dynCtx.Dyn.Extend.LikeIcon.NewIconID,
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_stat,
		ModuleItem: stat,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}
