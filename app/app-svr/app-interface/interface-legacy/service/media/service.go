package media

import (
	"context"

	"strconv"

	pgcreview "git.bilibili.co/bapis/bapis-go/pgc/service/review"

	baikegrpc "git.bilibili.co/bapis/bapis-go/community/interface/baike"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	channelmdl "git.bilibili.co/bapis/bapis-go/community/model/channel"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	pgcmedia "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/media"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	arcdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/archive"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/bangumi"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/channel"
	favdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/favorite"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/thumbup"
	archiveapi "go-gateway/app/app-svr/archive/service/api"
)

// Service service struct
type Service struct {
	c          *conf.Config
	channelDao *channel.Dao
	bangumiDao *bangumi.Dao
	arcDao     *arcdao.Dao
	thumbupDao *thumbup.Dao
	accDao     *accdao.Dao
	favDao     *favdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		channelDao: channel.New(c),
		bangumiDao: bangumi.New(c),
		arcDao:     arcdao.New(c),
		thumbupDao: thumbup.New(c),
		accDao:     accdao.New(c),
		favDao:     favdao.New(c),
	}
	return
}

func (s *Service) MediaTab(c context.Context, arg *api.MediaTabReq, mid int64, dev device.Device) (*api.MediaTabReply, error) {
	//参数校验
	chanArg := &channelgrpc.ChannelDetailReq{Mid: mid, Meta: &channelmdl.MetaDataCtrl{Page: "movie", Platform: dev.RawPlatform, From: arg.Spmid, Args: arg.Args, MobiApp: dev.RawMobiApp, Build: dev.Build}}
	if arg.BizType == api.BizMediaType {
		chanArg.BizId = arg.BizId
		chanArg.BizType = channelgrpc.ChannelBizlType_MOVIE
	} else {
		chanArg.ChannelId = arg.BizId
	}
	//获取频道信息
	chanRly, err := s.channelDao.ChannelDetail(c, chanArg)
	if err != nil { //错误，展示兜底态
		log.Error("s.channelDao.ChannelDetail(%v) error(%v)", chanArg, err)
		return &api.MediaTabReply{}, nil
	}
	// 频道信息为空 ｜｜ 不是电影频道
	if chanRly.GetChannel() == nil || chanRly.GetChannel().BizType != channelgrpc.ChannelBizlType_MOVIE || chanRly.GetChannel().BizId == 0 {
		return &api.MediaTabReply{}, nil
	}
	var (
		mediaRly   *pgcmedia.MediaBizInfoGetReply
		isLike     bool
		bizId      = int32(chanRly.GetChannel().BizId)
		seasonInfo *seasongrpc.CardInfoProto
	)
	eg := errgroup.WithContext(c)
	//获取媒资信息
	eg.Go(func(ctx context.Context) (e error) {
		mediaRly, e = s.bangumiDao.GetMediaBizInfoByMediaBizId(ctx, chanRly.GetChannel().BizId)
		if e != nil { //强依赖，返回err
			log.Error("s.bangumiDao.GetMediaBizInfoByMediaBizId(%d) error(%v)", chanRly.GetChannel().BizId, e)
			return
		}
		if mediaRly == nil {
			return ecode.NothingFound
		}
		return
	})
	//获取想看
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			staRly, err := s.bangumiDao.StatusByMid(ctx, mid, chanRly.GetChannel().BizId)
			if err != nil {
				log.Error("s.bangumiDao.StatusByMid(%d,%d) error(%v)", mid, chanRly.GetChannel().BizId, err)
				return nil
			}
			staMap := staRly.GetResult()
			if staMap == nil {
				return nil
			}
			if val, ok := staMap[bizId]; ok && val != nil {
				if val.FollowStatus == 1 {
					isLike = true
				}
			}
			return nil
		})
	}
	//获取是否支持立即观看
	eg.Go(func(ctx context.Context) error {
		seaRly, err := s.bangumiDao.CardsByMediaBizIds(ctx, []int32{bizId})
		if err != nil {
			log.Error("s.bangumiDao.CardsByMediaBizIds(%d) error(%v)", bizId, err)
			return nil
		}
		if _, ok := seaRly[bizId]; ok && seaRly[bizId].GetIsDelete() == 0 {
			seasonInfo = seaRly[bizId]
		}
		return nil
	})
	//获取是否允许展示写长评&写短评
	var isAllowReply int32
	eg.Go(func(ctx context.Context) error {
		if review, e := s.bangumiDao.AllowReview(ctx, bizId); e != nil {
			log.Error("s.bangumiDao.AllowReview(%d) error(%v)", bizId, e)
		} else {
			isAllowReply = review.GetAllowReview()
		}
		return nil
	})
	var articleId int32
	eg.Go(func(ctx context.Context) error {
		if rev, e := s.bangumiDao.MediaStatus(ctx, bizId, mid); e != nil {
			log.Error("s.bangumiDao.MediaStatus(%d) error(%v)", bizId, e)
		} else {
			articleId = rev.GetLongReview().GetArticleId()
		}
		return nil
	})
	// 获取评分信息
	var reviewInfo *pgcreview.ReviewInfoReply
	eg.Go(func(ctx context.Context) (e error) {
		if reviewInfo, e = s.bangumiDao.ReviewInfo(ctx, chanRly.GetChannel().BizId); e != nil {
			log.Error("s.bangumiDao.ReviewInfo(%d) error(%v)", bizId, e)
			e = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return &api.MediaTabReply{}, nil
	}
	//拼接数据
	return &api.MediaTabReply{
		MediaCard:       api.CardItem(mediaRly, seasonInfo, isLike, chanRly.GetChannel().GetChannelName(), reviewInfo, isAllowReply, articleId),
		Tab:             api.ShowTabs(chanRly.GetTabs(), isAllowReply),
		DefaultTabIndex: chanRly.GetDefaultTabIdx(),
		ChannelInfo:     &api.ChannelInfo{ChannelId: chanRly.GetChannel().ChannelId, Subscribed: chanRly.GetChannel().Subscribed},
	}, nil
}

func (s *Service) MediaDetail(c context.Context, arg *api.MediaDetailReq) (*api.MediaDetailReply, error) {
	//目前只支持媒资id
	if arg.BizType != api.BizMediaType {
		return nil, ecode.NothingFound
	}
	mediaRly, e := s.bangumiDao.GetMediaBizInfoByMediaBizId(c, arg.BizId)
	if e != nil { //强依赖，返回err
		log.Error("s.bangumiDao.GetMediaBizInfoByMediaBizId(%d) error(%v)", arg.BizId, e)
		return nil, e
	}
	if mediaRly == nil {
		return nil, ecode.NothingFound
	}
	//拼接数据
	return api.DetailItem(mediaRly), nil
}

func (s *Service) MediaVideo(c context.Context, arg *api.MediaVideoReq, mid int64, dev device.Device) (*api.MediaVideoReply, error) {
	//目前只支持频道
	if arg.BizType != api.BizChannelType {
		return nil, ecode.NothingFound
	}
	req := &baikegrpc.ChannelFeedReq{Cid: arg.BizId, FeedID: arg.FeedId, Mid: mid, Offset: arg.Offset, Ps: arg.Ps, Buvid: dev.Buvid}
	feedRly, err := s.channelDao.ChannelFeed(c, req)
	if err != nil {
		log.Error("s.channelDao.ChannelFeed(%v) error(%v)", req, err)
		return nil, err
	}
	out := &api.MediaVideoReply{HasMore: feedRly.GetHasMore(), Offset: feedRly.GetOffset()}
	var (
		aidMaps = make(map[int64]struct{})
		aids    []int64
		playAvs []*archiveapi.PlayAv
	)
	for _, v := range feedRly.GetList() {
		if v.GetType() != 0 || v.GetRid() == 0 {
			continue
		}
		if _, ok := aidMaps[v.GetRid()]; !ok {
			aids = append(aids, v.GetRid())
			playAvs = append(playAvs, &archiveapi.PlayAv{Aid: v.GetRid()})
			aidMaps[v.GetRid()] = struct{}{}
		}
	}
	//获取稿件信息
	if len(aids) == 0 {
		return out, nil
	}
	var (
		arcsPlay   map[int64]*archiveapi.ArcPlayer
		likeStates map[int64]thumbupgrpc.State
	)
	eg := errgroup.WithContext(c)
	if arg.PlayerArgs != nil { //有秒开参数进秒开
		eg.Go(func(c context.Context) (e error) {
			arcsPlay, e = s.arcDao.ArcsPlayer(c, playAvs, false)
			if e != nil {
				log.Error("s.arcDao.ArcsPlayer(%+v) error(%v)", aids, e)
			}
			return
		})
	} else {
		eg.Go(func(c context.Context) error {
			arcs, e := s.arcDao.Archives(c, aids, dev.MobiApp(), dev.Device, mid)
			if e != nil {
				log.Error("s.arcDao.Archives(%+v) error(%v)", aids, e)
				return nil
			}
			arcsPlay = make(map[int64]*archiveapi.ArcPlayer)
			for k, v := range arcs {
				arcsPlay[k] = &archiveapi.ArcPlayer{Arc: v}
			}
			return nil
		})
	}
	//是否点赞
	eg.Go(func(c context.Context) (e error) {
		if likeStates, e = s.thumbupDao.HasLike(c, dev.Buvid, mid, aids); e != nil {
			log.Error("s.thumbupDao.HasLike(%+v) error(%v)", aids, e)
			e = nil
		}
		return
	})
	var hasFav map[int64]int8
	if mid > 0 {
		eg.Go(func(c context.Context) (e error) {
			if hasFav, e = s.favDao.IsFavVideos(c, mid, aids); e != nil {
				log.Error("s.favDao.IsFavVideos: %+v", e)
				e = nil
			}
			return
		})
	}
	if err := eg.Wait(); err != nil {
		return out, nil
	}
	var bigCard []*api.BigItem
	for _, v := range aids {
		if _, ok := arcsPlay[v]; !ok || arcsPlay[v].GetArc() == nil || !arcsPlay[v].GetArc().IsNormalV2() {
			continue
		}
		var isFav bool
		if fa, ok := hasFav[v]; ok && fa == 1 {
			isFav = true
		}
		playerInfo := arcsPlay[v].GetPlayerInfo()[arcsPlay[v].GetDefaultPlayerCid()]
		bigCard = append(bigCard, api.BigCardItem(arcsPlay[v].GetArc(), playerInfo, int32(likeStates[v]), dev, isFav))
	}
	out.List = bigCard
	return out, nil
}

func (s *Service) MediaRelation(c context.Context, arg *api.MediaRelationReq, mid int64, dev device.Device) (*api.MediaRelationReply, error) {
	//目前只支持频道
	if arg.BizType != api.BizChannelType {
		return nil, ecode.NothingFound
	}
	req := &baikegrpc.ChannelFeedReq{Cid: arg.BizId, FeedID: arg.FeedId, Mid: mid, Offset: arg.Offset, Ps: arg.Ps, Buvid: dev.Buvid}
	feedRly, err := s.channelDao.ChannelFeed(c, req)
	if err != nil {
		log.Error("s.channelDao.ChannelFeed(%v) error(%v)", req, err)
		return nil, err
	}
	out := &api.MediaRelationReply{HasMore: feedRly.GetHasMore(), Offset: feedRly.GetOffset()}
	var (
		aidMaps = make(map[int64]struct{})
		aids    []int64
	)
	for _, v := range feedRly.GetList() {
		if v.GetType() != 0 || v.GetRid() == 0 {
			continue
		}
		if _, ok := aidMaps[v.GetRid()]; !ok {
			aids = append(aids, v.GetRid())
			aidMaps[v.GetRid()] = struct{}{}
		}
	}
	//获取稿件信息
	if len(aids) == 0 {
		return out, nil
	}
	arcs, err := s.arcDao.Archives(c, aids, dev.MobiApp(), dev.Device, mid)
	if err != nil {
		log.Error("s.arcDao.Archives(%+v) error(%v)", aids, err)
		return out, nil
	}
	var smallCard []*api.SmallItem
	for _, v := range aids {
		if _, ok := arcs[v]; !ok || !arcs[v].IsNormalV2() {
			continue
		}
		smallCard = append(smallCard, api.SmallCardItem(arcs[v], dev))
	}
	out.List = smallCard
	return out, nil
}

func (s *Service) MediaFollow(c context.Context, arg *api.MediaFollowReq, mid int64) (*api.MediaFollowReply, error) {
	//目前只支持媒资
	if arg.Type != api.ButType_BUT_LIKE {
		return nil, ecode.RequestErr
	}
	meid, err := strconv.ParseInt(arg.Id, 10, 64)
	if err != nil || meid == 0 {
		return nil, ecode.RequestErr
	}
	if err = s.bangumiDao.AddMediaFollow(c, mid, meid); err != nil {
		return nil, err
	}
	return &api.MediaFollowReply{}, nil
}

func (s *Service) MediaComment(c context.Context, arg *api.MediaCommentReq, mid int64) (*api.MediaCommentReply, error) {
	//获取账号信息
	acProfile, e := s.accDao.Profile3(c, mid)
	if e != nil { //强依赖，返回err
		log.Error("s.accDao.Profile3(%d) error(%v)", mid, e)
		return &api.MediaCommentReply{ErrMsg: "网络错误，请稍后重试～"}, nil
	}
	//判断账号等级
	if acProfile.GetProfile().GetLevel() < 4 {
		return &api.MediaCommentReply{ErrMsg: "账号等级到达LV4才可以评分呐～"}, nil
	}
	//判断手机号码
	if acProfile.GetProfile().GetTelStatus() != 1 {
		return &api.MediaCommentReply{ErrMsg: "请先前往【我的-设置-安全隐私-账号安全中心】绑定手机喔～"}, nil
	}
	return &api.MediaCommentReply{}, nil
}
