package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/interface/model"
	mainEcode "go-gateway/ecode"

	accwar "git.bilibili.co/bapis/bapis-go/account/service"
	memmdl "git.bilibili.co/bapis/bapis-go/account/service/member"
	relamdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	api "git.bilibili.co/bapis/bapis-go/article/service"
	pugvapi "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	tagmdl "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favapi "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbupmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	livemdl "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexmdl "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	followmdl "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"
	resmdl "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	ugcmdl "git.bilibili.co/bapis/bapis-go/ugc-season/service"
	uparcmdl "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

const (
	_samplePn        = 1
	_samplePs        = 1
	_silenceForbid   = 1
	_accBlockDefault = 0
	_accBlockDue     = 1
	_officialNoType  = -1
	_audioCardOn     = 1
	_noticeForbid    = 1
	_fake            = 1
	_deleted         = 1
	// 原样式icon
	_sysNoticeOldIcon = "https://i0.hdslb.com/bfs/space/7a89f7ed04b98458b23863846bd2539a90ff1153.png"
	// 缅怀提示icon
	_sysNoticeNewIcon = "https://i0.hdslb.com/bfs/space/ca6d0ed2edae23cf348db19cd2c293f2121c1b59.png"
	// 缅怀背景色
	_sysNoticeNewBgColor = "#e7e7e7"
	// 缅怀文字色
	_sysNoticeNewTextcolor = "#999999"
	// 原样式背景色
	_sysNoticeOldBgColor = "#FFF3DB"
	// 原样式文字色
	_sysNoticeOldTextcolor = "#FFB112"
)

var (
	_emptyThemeList = make([]*model.ThemeDetail, 0)
	_emptyArcItem   = make([]*model.ArcItem, 0)
)

// NavNum get space nav num by mid.
func (s *Service) NavNum(c context.Context, mid, vmid int64) (res *model.NavNum) {
	ip := metadata.String(c, metadata.RemoteIP)
	res = &model.NavNum{
		Favourite: &model.Num{},
	}
	var guestFavCount int
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		setting := s.privacy(ctx, vmid)
		isLivePlayBack := setting[model.LivePlayback]
		without := []uparcmdl.Without{uparcmdl.Without_no_space}
		if isLivePlayBack == 0 {
			without = append(without, uparcmdl.Without_live_playback)
		}
		if reply, err := s.upArcClient.ArcPassedTotal(ctx, &uparcmdl.ArcPassedTotalReq{Mid: vmid, Without: without}); err != nil {
			log.Error("s.upArcClient.ArcPassedTotal(%d) error(%v)", vmid, err)
		} else if reply != nil {
			res.Video = reply.Total
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		res.Channel = s.videoListCount(ctx, vmid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.favClient.CntUserFolders(ctx, &favapi.CntUserFoldersReq{Typ: _typeFavArchive, Mid: mid, Vmid: vmid})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		res.Favourite.Master = int(reply.GetCount())
		res.Favourite.Guest = int(reply.GetCount())
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if mid != vmid {
			return nil
		}
		reply, err := s.favClient.CntUserFolders(ctx, &favapi.CntUserFoldersReq{Typ: _typeFavArchive, Mid: 0, Vmid: vmid})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		guestFavCount = int(reply.GetCount())
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if reply, err := s.pgcFollowClient.MyFollows(ctx, &followmdl.MyFollowsReq{Mid: vmid, FollowType: model.FollowTypeAnime, Pn: _samplePn, Ps: _samplePs}); err != nil {
			log.Error("s.pgcFollowClient.MyFollows Anime(%d) error(%v)", vmid, err)
		} else if reply != nil && reply.Total > 0 {
			res.Bangumi = reply.Total
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if reply, err := s.pgcFollowClient.MyFollows(ctx, &followmdl.MyFollowsReq{Mid: vmid, FollowType: model.FollowTypeCinema, Pn: _samplePn, Ps: _samplePs}); err != nil {
			log.Error("s.pgcFollowClient.MyFollows Cinema(%d) error(%v)", vmid, err)
		} else if reply != nil && reply.Total > 0 {
			res.Cinema = reply.Total
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if tag, err := s.tag.SubTags(ctx, &tagmdl.SubTagsReq{Mid: vmid, Pn: _samplePn, Ps: _samplePs}); err != nil {
			log.Error("s.tag.SubTags(%d) error(%v)", vmid, err)
		} else if tag != nil {
			res.Tag = int(tag.Total)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if art, err := s.artClient.UpArtMetas(ctx, &api.UpArtMetasReq{Mid: vmid, Pn: 1, Ps: 10, Ip: ip}); err != nil {
			log.Error("s.artClient.UpArtMetas(%d) error(%v)", vmid, err)
		} else if art != nil {
			res.Article = art.Count
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if cnt, err := s.dao.AlbumCount(ctx, vmid); err == nil && cnt > 0 {
			res.Album = cnt
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if cnt, err := s.dao.AudioCnt(ctx, vmid); err != nil {
			log.Error("s.dao.AudioCnt(%d) error(%v)", vmid, err)
		} else if cnt > 0 {
			res.Audio = cnt
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if cnt, err := s.pugvClient.UserSeason(ctx, &pugvapi.UserSeasonReq{Mid: vmid, Pn: _samplePn, Ps: _samplePs, NeedAll: 2}); err != nil {
			log.Error("s.pugvClient.UserSeason(%d) error(%v)", vmid, err)
		} else if cnt != nil {
			res.Pugv = cnt.Total
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.ugcSeasonClient.UpperList(ctx, &ugcmdl.UpperListRequest{Mid: vmid, PageNum: _samplePn, PageSize: _samplePs})
		if err != nil {
			log.Error("s.NavNum UpperList vmid:%d, err:%+v", vmid, err)
			return nil
		}
		res.SeasonNum = reply.TotalCount
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if guestFavCount > 0 {
		res.Favourite.Guest = guestFavCount
	}
	return
}

func (s *Service) videoListCount(ctx context.Context, vmid int64) *model.Num {
	seriesNum := &model.Num{}
	reply, err := s.seriesGRPC.ListSeries(ctx, &seriesgrpc.ListSeriesReq{Mid: vmid, State: seriesgrpc.SeriesOnlineAll})
	if err != nil {
		log.Error("%+v", err)
		return seriesNum
	}
	count := len(reply.GetSeriesList())
	seriesNum.Master = count
	seriesNum.Guest = count
	return seriesNum
}

// UpStat get up stat.
func (s *Service) UpStat(c context.Context, mid int64) (res *model.UpStat, err error) {
	var (
		info     *accwar.InfoReply
		likeErr  error
		likeRep  *thumbupmdl.UserLikedCountsReply
		viewData *model.CreativeView
	)
	if info, err = s.accClient.Info3(c, &accwar.MidReq{Mid: mid}); err != nil || info == nil {
		log.Error("s.accClient.Info3(%d) error(%v)", mid, err)
		return
	}
	res = new(model.UpStat)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		var err error
		viewData, err = s.creativeViewData(ctx, mid)
		if err != nil {
			log.Error("s.creativeViewData mid:%d, error:%v", mid, err)
		}
		if viewData != nil {
			res.Archive.View = viewData.ArchivePlay
			res.Article.View = viewData.ArticleView
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		arg := &thumbupmdl.UserLikedCountsReq{Mid: mid, Businesses: []string{
			_businessArchiveLike,
			_businessDynamicLike,
			_businessArticleLike,
			_businessDyAlbumLike,
			_businessDyclipLike,
			_businessDyCheeseLike,
		}}
		if likeRep, likeErr = s.thumbupClient.UserLikedCounts(ctx, arg); likeErr != nil {
			log.Error("s.UserLikedCounts(%d) error(%v)", mid, likeErr)
		} else if likeRep != nil {
			for _, v := range likeRep.LikeCounts {
				res.Likes += v
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// MyInfo get my info.
func (s *Service) MyInfo(c context.Context, mid int64) (res *model.ProfileStat, err error) {
	var reply *accwar.ProfileStatReply
	if reply, err = s.accClient.ProfileWithStat3(c, &accwar.MidReq{Mid: mid}); err != nil {
		log.Error("s.accClient.ProfileWithStat3(%d) error(%v)", mid, err)
		return
	}
	res = &model.ProfileStat{
		Profile:   reply.Profile,
		LevelExp:  reply.LevelInfo,
		Coins:     reply.Coins,
		Following: reply.Follower,
		Follower:  reply.Follower,
	}
	// reset join time
	res.Profile.JoinTime = 0
	return
}

// AccTags get account tags.
func (s *Service) AccTags(c context.Context, mid int64) ([]*memmdl.UserTagReply, error) {
	rly, err := s.memberClient.UserTag(c, &memmdl.MidReq{Mid: mid})
	if err != nil {
		log.Error("AccTags s.memberClient.UserTag mid:%d error:%v", mid, err)
		return nil, err
	}
	return []*memmdl.UserTagReply{rly}, nil
}

// SetAccTags set account tags.
func (s *Service) SetAccTags(c context.Context, mid int64, tags []string) error {
	if _, err := s.memberClient.SetUserTag(c, &memmdl.SetUserTagReq{Mid: mid, Tags: tags}); err != nil {
		log.Error("SetAccTags s.memberClient.SetUserTag mid:%d tags:%v error:%v", mid, tags, err)
		return err
	}
	return nil
}

// nolint:gocognit,gomnd
func (s *Service) AccInfo(ctx context.Context, mid, vmid int64, riskParams *model.RiskManagement) (*model.AccInfo, error) {
	res := &model.AccInfo{
		Theme:       struct{}{},
		Series:      &model.Series{UserUpgradeStatus: seriesgrpc.Upgraded},
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	if env.DeployEnv == env.DeployEnvProd {
		if _, ok := s.BlacklistValue[vmid]; ok {
			return nil, errors.WithMessage(xecode.NothingFound, "命中blacklist")
		}
	}
	var (
		disableUserInfo, disableShowSchool bool
		isFakeAccount                      int32
		hasOldChannel                      bool
	)
	riskResult := s.RiskVerifyAndManager(ctx, riskParams)
	if riskResult != nil {
		res.GaiaResType = riskResult.GaiaResType
		res.IsRisk = riskResult.IsRisk
		res.GaiaData = riskResult.GaiaData
		return res, nil
	}
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		if ok := func() bool {
			if mid == vmid {
				return true
			}
			req := &resmdl.CheckCommonBWListReq{
				Oid:    strconv.FormatInt(vmid, 10),
				Token:  s.c.LegoToken.SpaceIPLimit,
				UserIp: metadata.String(ctx, metadata.RemoteIP),
			}
			reply, err := s.resGRPC.CheckCommonBWList(ctx, req)
			if err != nil {
				log.Error("%+v", err)
				return true
			}
			return reply.GetIsInList()
		}(); !ok {
			return errors.WithMessage(xecode.NothingFound, "命中lego地区黑名单")
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.accClient.ProfileWithStat3(ctx, &accwar.MidReq{Mid: vmid})
		if err != nil {
			if xecode.EqualError(xecode.UserNotExist, err) || xecode.EqualError(mainEcode.MemberNotExist, err) {
				return errors.WithMessage(xecode.NothingFound, "账号告知用户不存在")
			}
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
			reply = model.DefaultProfileStat
		}
		if reply.GetProfile().GetIsDeleted() == _deleted {
			log.Error("AccInfo mid=%d vmid=%d profile=%+v is deleted", mid, vmid, reply.GetProfile())
			return errors.WithMessage(xecode.NothingFound, "账号告知用户被删除")
		}
		isFakeAccount = reply.GetProfile().GetIsFakeAccount()
		res.FromCard(reply)
		midNFTRegionMap := s.BatchNFTRegion(ctx, []int64{vmid})
		res.FaceNftType = midNFTRegionMap[vmid]
		if res.Mid == 0 {
			res.Mid = vmid
		}
		if mid != vmid {
			res.Coins = 0
		}
		if !s.c.SeniorMemberSwitch.ShowSeniorMember {
			res.IsSeniorMember = 0
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.memberClient.UserTag(ctx, &memmdl.MidReq{Mid: vmid})
		if err != nil {
			log.Error("AccTags s.memberClient.UserTag mid:%d error:%v", mid, err)
			return nil
		}
		res.Tags = reply.GetTags()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if mid == vmid {
			return nil
		}
		if err := s.privacyCheck(ctx, vmid, model.PcyUserInfo); err != nil {
			disableUserInfo = true
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if mid == vmid {
			return nil
		}
		reply := s.disableShowSchoolPrivacy(ctx, vmid)
		disableShowSchool = reply == 1
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if mid <= 0 || mid == vmid {
			return nil
		}
		reply, err := s.accClient.Relation3(ctx, &accwar.RelationReq{Mid: mid, Owner: vmid})
		if err != nil {
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
			return nil
		}
		res.IsFollowed = reply.GetFollowing()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.dao.WebTopPhoto(ctx, vmid, mid, "", "")
		if err != nil {
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
		}
		res.TopPhoto = s.c.Rule.TopPhoto
		if reply != nil && reply.LImg != "" {
			res.TopPhoto = reply.LImg
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.liveClient.QueryMedal(ctx, &livemdl.QueryMedalReq{UpUid: vmid})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		res.FansBadge = reply.GetUpMedal().GetId() > 0
		return nil
	})
	g.Go(func(ctx context.Context) error {
		res.FansMedal = &model.FansMedal{}
		// 1.判断主客态
		owner := mid == vmid
		hideFansMedal, _ := s.spaceLiveMedalPrivacy(ctx, vmid)
		// 2.公开显示佩戴的粉丝勋章开关关闭
		if hideFansMedal == 1 {
			// 客态不展示
			if !owner {
				return nil
			}
			// 主态灰度占位
			res.FansMedal.Show = true
			return nil
		}
		reply, err := s.liveUserGRPC.Wearing(ctx, &livemdl.WearingReq{UserId: vmid, ForceLight: true})
		if err != nil {
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
			return nil
		}
		// 3.没有佩戴勋章
		if reply.GetWearing() == livemdl.WearingStatus_NotWearing {
			// 客态不展示
			if !owner {
				return nil
			}
			// 主态灰度占位
			res.FansMedal.Show = true
			return nil
		}
		// 正常展示
		res.FansMedal.Show = true
		res.FansMedal.Wear = true
		res.FansMedal.Medal = reply.GetMedal()
		return nil
	})
	g.Go(func(ctx context.Context) error {
		entryFrom := "NONE"
		req := &livexmdl.EntryRoomInfoReq{
			Uid:        mid,
			Uids:       []int64{vmid},
			EntryFrom:  []string{entryFrom},
			NotPlayurl: 1,
			ReqBiz:     "/x/space/acc/info",
		}
		reply, err := s.liveXRoom.EntryRoomInfo(ctx, req)
		if err != nil {
			log.Error("AccInfo mid=%d vmid=%d error:%+v", mid, vmid, err)
			return nil
		}
		if reply == nil || reply.List[vmid] == nil {
			return nil
		}
		info := reply.List[vmid]
		res.LiveRoom = &model.Live{
			LiveStatus:    info.LiveStatus,
			URL:           info.JumpUrl[entryFrom],
			Title:         info.Title,
			Cover:         info.Cover,
			RoomID:        info.RoomId,
			BroadcastType: info.LiveScreenType,
			WatchedShow:   info.WatchedShow,
			RoomStatus:    1,
		}
		switch info.LiveStatus {
		case 0:
			res.LiveRoom.LiveStatus = 0
			res.LiveRoom.RoundStatus = 0
		case 1:
			res.LiveRoom.LiveStatus = 1
			res.LiveRoom.RoundStatus = 0
			if info.IsEncryptRoom || info.IsHiddenRoom {
				res.LiveRoom.LiveStatus = 0
			}
		case 2:
			res.LiveRoom.LiveStatus = 0
			res.LiveRoom.RoundStatus = 1
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.channelList(ctx, vmid)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		hasOldChannel = len(reply) != 0
		return nil
	})
	if s.c.Rule.McnOn {
		// mcn info
		g.Go(func(ctx context.Context) error {
			reply, err := s.memberClient.GetUserExtraValue(ctx, &memmdl.MemberMidReq{
				Mid:      vmid,
				RemoteIP: metadata.String(ctx, metadata.RemoteIP),
			})
			if err != nil {
				log.Error("@AccInfo , get MCN Info Failed (%v)", err)
				return nil
			}

			infoMap := reply.GetExtraInfo()
			if infoMap == nil {
				return nil
			}

			mcn := ""
			if mcnInfo, isOk := infoMap["video_mcn_info"]; isOk {
				// 优先
				mcn = mcnInfo
			} else if mcnInfo, isOk := infoMap["live_mcn_info"]; isOk {
				mcn = mcnInfo
			}
			if mcn != "" {
				var info *model.MCNInfo
				if err = json.Unmarshal([]byte(mcn), &info); err != nil {
					log.Error("【@GetUserExtraValue】json parse error: (%v), info: (%s)", err, mcn)
					return nil
				}
				res.LiveMCNInfo = &model.MCNInfo{
					Name: info.Name,
					URL:  info.URL,
				}
			}
			return nil
		})
	}
	g.Go(func(ctx context.Context) error {
		// 充电信息
		reply, err := s.payRankGRPC.UPRankWithPanelByUPMid(ctx, &payrank.RankElecUPReq{UPMID: vmid, Mid: mid})
		if err != nil || reply == nil {
			log.Error("ElecShow s.payRankGRPC.UPRankWithPanelByUPMid upMid:%d error:%v", vmid, err)
			return nil
		}
		fmt.Println("elec info", mid, vmid, reply.GetState(), reply.GetUpowerTitle(), reply.GetUpowerIconUrl(), reply.GetUpowerJumpUrl())
		res.ElecInfo = &model.ElecPlusInfo{
			ShowInfo: &model.ShowInfo{
				Show:    reply.GetShow(),
				State:   int8(reply.State),
				Title:   reply.UpowerTitle,
				Icon:    reply.UpowerIconUrl,
				JumpURL: reply.UpowerJumpUrl,
			},
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if disableUserInfo {
		res.Sex = "保密"
		res.Birthday = ""
		res.Tags = nil
	}
	if disableShowSchool {
		res.School = nil
	}
	if !hasOldChannel {
		//已有频道的用户（不含仅有“直播回放”系列的用户）：全量展现频道升级提示，toast+弹窗+常驻升级入口
		//没有频道的用户&仅有“直播回放”系列的用户：直接升级为“视频列表”
		//升级后，用户频道链接channel/index链接重定向至channel/series
		res.Series.UserUpgradeStatus = seriesgrpc.Upgraded
		res.Series.ShowUpgradeWindow = false
	}
	func() {
		if v, ok := s.SysNotice[vmid]; ok {
			// 公告配置类型，1-其他类型，2-去世公告
			notice := &model.SysNotice{}
			*notice = *v
			notice.Icon = _sysNoticeOldIcon
			notice.BgColor = _sysNoticeOldBgColor
			notice.TextColor = _sysNoticeOldTextcolor
			if notice.NoticeType == 2 {
				notice.Icon = _sysNoticeNewIcon
				notice.BgColor = _sysNoticeNewBgColor
				notice.TextColor = _sysNoticeNewTextcolor
			}
			res.SysNotice = notice
			return
		}
		if isFakeAccount == _fake {
			var url string
			content := s.c.Fake.Guest
			if mid == vmid {
				url = s.c.Fake.Url
				content = s.c.Fake.Home
			}
			res.SysNotice = &model.SysNotice{
				Content:   content,
				Url:       url,
				Icon:      _sysNoticeOldIcon,
				BgColor:   _sysNoticeOldBgColor,
				TextColor: _sysNoticeOldTextcolor,
			}
			return
		}
		res.SysNotice = struct{}{}
	}()
	return res, nil
}

// ThemeList get theme list.
func (s *Service) ThemeList(c context.Context, mid int64) (data []*model.ThemeDetail, err error) {
	var theme *model.ThemeDetails
	if theme, err = s.dao.Theme(c, mid); err != nil {
		return
	}
	if theme == nil || len(theme.List) == 0 {
		data = _emptyThemeList
		return
	}
	data = theme.List
	return
}

// ThemeActive theme active.
func (s *Service) ThemeActive(c context.Context, mid, themeID int64) (err error) {
	var (
		theme *model.ThemeDetails
		check bool
	)
	if theme, err = s.dao.Theme(c, mid); err != nil {
		return
	}
	if theme == nil || len(theme.List) == 0 {
		err = xecode.RequestErr
		return
	}
	for _, v := range theme.List {
		if v.ID == themeID {
			if v.IsActivated == 1 {
				err = xecode.NotModified
				return
			}
			check = true
		}
	}
	if !check {
		err = xecode.RequestErr
		return
	}
	if err = s.dao.ThemeActive(c, mid, themeID); err == nil {
		_ = s.dao.DelCacheTheme(c, mid)
	}
	return
}

// Relation .
func (s *Service) Relation(c context.Context, mid, vmid int64) (data *model.Relation) {
	data = &model.Relation{Relation: struct{}{}, BeRelation: struct{}{}}
	ip := metadata.String(c, metadata.RemoteIP)
	if mid == vmid {
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if relation, err := s.relationClient.Relation(ctx, &relamdl.RelationReq{Mid: mid, Fid: vmid, RealIp: ip}); err != nil {
			log.Error("Relation s.relation.Relation(Mid:%d,Fid:%d,%s) error %v", mid, vmid, ip, err)
		} else if relation != nil {
			data.Relation = relation
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if beRelation, err := s.relationClient.Relation(ctx, &relamdl.RelationReq{Mid: vmid, Fid: mid, RealIp: ip}); err != nil {
			log.Error("Relation s.relation.Relation(Mid:%d,Fid:%d,%s) error %v", vmid, mid, ip, err)
		} else if beRelation != nil {
			data.BeRelation = beRelation
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// WebIndex web index.
func (s *Service) WebIndex(c context.Context, mid, vmid int64, pn, ps int32) (data *model.WebIndex, err error) {
	data = new(model.WebIndex)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		info, infoErr := s.AccInfo(ctx, mid, vmid, &model.RiskManagement{})
		if infoErr != nil {
			return infoErr
		}
		data.Account = info
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if setting, e := s.SettingInfo(ctx, vmid); e == nil {
			data.Setting = setting
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if upArc, e := s.UpArcs(ctx, vmid, pn, ps); e != nil {
			data.Archive = &model.WebArc{Archives: _emptyArcItem}
		} else {
			arc := &model.WebArc{
				Page:     model.WebPage{Pn: pn, Ps: ps, Count: upArc.Count},
				Archives: upArc.List,
			}
			data.Archive = arc
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// ValidateTheme .
func (s *Service) ValidateTheme(c context.Context, mid int64) (*model.ThemeDetail, error) {
	theme, err := s.dao.Theme(c, mid)
	if err != nil {
		return nil, err
	}
	if theme == nil || len(theme.List) == 0 {
		return nil, nil
	}
	for _, v := range theme.List {
		if v != nil && v.IsActivated == 1 {
			return v, nil
		}
	}
	return nil, nil
}

func (s *Service) creativeViewData(ctx context.Context, mid int64) (*model.CreativeView, error) {
	addCache := true
	data, err := s.dao.CreativeViewDataCache(ctx, mid)
	if data != nil {
		return data, nil
	}
	if err != nil {
		addCache = false
	}
	if data, err = s.dao.CreativeViewData(ctx, mid); data != nil && addCache {
		s.cache.Do(ctx, func(c context.Context) {
			_ = s.dao.SetCreativeViewDataCache(c, mid, data)
		})
	}
	return data, err
}
