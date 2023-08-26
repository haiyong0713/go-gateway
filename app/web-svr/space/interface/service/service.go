package service

import (
	"context"
	gaiagrpc "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"regexp"
	"strconv"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/space/ecode"
	xcode "go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/dao"
	"go-gateway/app/web-svr/space/interface/model"
	mainEcode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"

	accwar "git.bilibili.co/bapis/bapis-go/account/service"
	accmgrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	memberclient "git.bilibili.co/bapis/bapis-go/account/service/member"
	relaclient "git.bilibili.co/bapis/bapis-go/account/service/relation"
	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	noteapi "git.bilibili.co/bapis/bapis-go/app/note/service"
	arcclient "git.bilibili.co/bapis/bapis-go/archive/service"
	artclient "git.bilibili.co/bapis/bapis-go/article/service"
	assclient "git.bilibili.co/bapis/bapis-go/assist/service"
	authgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/auth"
	pugvgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	coinclient "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	filterclient "git.bilibili.co/bapis/bapis-go/filter/service"
	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcclient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	progclient "git.bilibili.co/bapis/bapis-go/pgc/service/progress"
	seasonclient "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"
	resgrpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	ugcclient "git.bilibili.co/bapis/bapis-go/ugc-season/service"
	uparcclient "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/robfig/cron"
)

// Service service struct.
type Service struct {
	c   *conf.Config
	dao *dao.Dao
	// rpc
	ass assclient.AssistClient
	tag tagrpc.TagRPCClient
	// grpc
	accClient         accwar.AccountClient
	arcClient         arcclient.ArchiveClient
	accmGRPC          accmgrpc.MemberClient
	pangugsGRPC       pangugsgrpc.GalleryServiceClient
	coinClient        coinclient.CoinClient
	thumbupClient     thumbupgrpc.ThumbupClient
	pgcFollowClient   pgcclient.FollowClient
	pgcSeasonClient   seasonclient.SeasonClient
	pgcProgressClient progclient.ProgressClient
	artClient         artclient.ArticleGRPCClient
	filterClient      filterclient.FilterClient
	liveClient        livexfans.AnchorClient
	liveXRoom         livexgrpc.XroomgateClient
	favClient         favclient.FavoriteClient
	pugvClient        pugvgrpc.SeasonClient
	pugvAuthClient    authgrpc.AuthClient
	memberClient      memberclient.MemberClient
	noteClient        noteapi.HktNoteClient
	relationClient    relaclient.RelationClient
	upArcClient       uparcclient.UpArchiveClient
	payRankGRPC       payrank.UGCPayRankClient
	actGRPC           actgrpc.ActivityClient
	seriesGRPC        seriesgrpc.SeriesClient
	resGRPC           resgrpc.ResourceClient
	liveUserGRPC      livexfans.UserClient
	pgcCardClient     pgccardgrpc.CardClient
	ugcSeasonClient   ugcclient.UGCSeasonClient
	cfcGRPC           cfcgrpc.FlowControlClient
	gaiaGRPC          gaiagrpc.GaiaClient
	// cache proc
	cache *fanout.Fanout
	// cron
	cron *cron.Cron
	// noNoticeMids
	noNoticeMids   map[int64]struct{}
	BlacklistValue map[int64]struct{}
	SysNotice      map[int64]*model.SysNotice
	// photo mall list
	photoMallList []*model.PhotoMall
	// load running
	loadBlacklistRunning bool
	loadSysNoticeRunning bool
	upRcmdBlackList      []int64
}

var (
	regPhotoMallIphone  = regexp.MustCompile(`,?iphone,?`)
	regPhotoMallAndroid = regexp.MustCompile(`,?android,?`)
)

// New new service.
func New(c *conf.Config) *Service {
	s := &Service{
		c:     c,
		dao:   dao.New(c),
		cache: fanout.New("cache"),
	}
	var err error
	if s.thumbupClient, err = thumbupgrpc.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	if s.accClient, err = accwar.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.coinClient, err = coinclient.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if s.pgcFollowClient, err = pgcclient.NewClient(c.PGCFollowClient); err != nil {
		panic(err)
	}
	if s.pgcSeasonClient, err = seasonclient.NewClient(c.PGCSeasonClient); err != nil {
		panic(err)
	}
	if s.pgcProgressClient, err = progclient.NewClient(c.PGCProgressClient); err != nil {
		panic(err)
	}
	if s.arcClient, err = arcclient.NewClient(c.ArticleClient); err != nil {
		panic(err)
	}
	if s.accmGRPC, err = accmgrpc.NewClient(c.AccmClient); err != nil {
		panic(err)
	}
	if s.pangugsGRPC, err = pangugsgrpc.NewClient(c.PanguGSClient); err != nil {
		panic(err)
	}
	if s.filterClient, err = filterclient.NewClient(c.FilterClient); err != nil {
		panic(err)
	}
	if s.artClient, err = artclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.ass, err = assclient.NewClient(c.AssistClient); err != nil {
		panic(err)
	}
	if s.liveClient, err = NewLiveClient(c.LiveGRPC); err != nil {
		panic(err)
	}
	if s.liveXRoom, err = livexgrpc.NewClientXroomgate(c.LiveXRoomGRPC); err != nil {
		panic(err)
	}
	if s.favClient, err = favclient.NewClient(c.FavGRPC); err != nil {
		panic(err)
	}
	if s.pugvClient, err = pugvgrpc.NewClient(c.PugvGRPC); err != nil {
		panic(err)
	}
	if s.pugvAuthClient, err = authgrpc.NewClient(c.PugvGRPC); err != nil {
		panic(err)
	}
	if s.memberClient, err = memberclient.NewClient(c.MemberClient); err != nil {
		panic(err)
	}
	if s.relationClient, err = relaclient.NewClient(c.RelationGRPC); err != nil {
		panic(err)
	}
	if s.noteClient, err = noteapi.NewClient(c.NoteClient); err != nil {
		panic(err)
	}
	if s.upArcClient, err = uparcclient.NewClient(c.UpArcClient); err != nil {
		panic(err)
	}
	if s.payRankGRPC, err = payrank.NewClient(c.PayRankClient); err != nil {
		panic(err)
	}
	if s.actGRPC, err = actgrpc.NewClient(c.ActivityClient); err != nil {
		panic(err)
	}
	if s.seriesGRPC, err = seriesgrpc.NewClient(c.SeriesGRPC); err != nil {
		panic(err)
	}
	if s.resGRPC, err = resgrpc.NewClient(c.ResourceGRPC); err != nil {
		panic(err)
	}
	if s.liveUserGRPC, err = livexfans.NewClient(c.LiveUserGRPC); err != nil {
		panic(err)
	}
	if s.tag, err = tagrpc.NewClient(c.TagGRPC); err != nil {
		panic(err)
	}
	if s.pgcCardClient, err = pgccardgrpc.NewClient(c.PGCCardClient); err != nil {
		panic(err)
	}
	if s.ugcSeasonClient, err = ugcclient.NewClient(c.UGCSeasonClient); err != nil {
		panic(err)
	}
	if s.cfcGRPC, err = cfcgrpc.NewClient(c.CfcGRPC); err != nil {
		panic(err)
	}
	if s.gaiaGRPC, err = gaiagrpc.NewClient(c.GaiaGRPC); err != nil {
		panic(err)
	}
	s.initMids()
	if err = s.initCron(); err != nil {
		panic(err)
	}
	return s
}

func (s *Service) initMids() {
	tmp := make(map[int64]struct{}, len(s.c.Rule.NoNoticeMids))
	for _, id := range s.c.Rule.NoNoticeMids {
		tmp[id] = struct{}{}
	}
	s.noNoticeMids = tmp
}

func (s *Service) realName(c context.Context, mid int64) (profile *accwar.Profile, err error) {
	var reply *accwar.ProfileReply
	if reply, err = s.accClient.Profile3(c, &accwar.MidReq{Mid: mid}); err != nil || reply == nil {
		log.Error("s.accClient.Profile3(%d) error(%v)", mid, err)
		return
	}
	profile = reply.Profile
	if !s.c.Rule.RealNameOn {
		return
	}
	if profile.Identification == 0 && profile.TelStatus == 0 {
		err = mainEcode.UserCheckNoPhone
		return
	}
	if profile.Identification == 0 && profile.TelStatus == 2 {
		err = mainEcode.UserCheckInvalidPhone
		return
	}
	return
}

func (s *Service) privacyCheck(c context.Context, vmid int64, field string) (err error) {
	privacy := s.privacy(c, vmid)
	if value, ok := privacy[field]; !ok || value != _defaultPrivacy {
		err = ecode.SpaceNoPrivacy
		return
	}
	return
}

// loadBlacklist load space blacklist
func (s *Service) loadBlacklist() {
	if s.loadBlacklistRunning {
		return
	}
	defer func() {
		s.loadBlacklistRunning = false
	}()
	s.loadBlacklistRunning = true
	s.blacklist(context.Background())
}

// loadSysNotice load system notice
func (s *Service) loadSysNotice() {
	if s.loadSysNoticeRunning {
		return
	}
	defer func() {
		s.loadSysNoticeRunning = false
	}()
	s.loadSysNoticeRunning = true
	s.sysNoticeMap(context.Background())
}

func (s *Service) avToBv(aid int64) (bvID string) {
	var err error
	if bvID, err = bvid.AvToBv(aid); err != nil {
		log.Warn("avToBv(%d) error(%v)", aid, err)
	}
	return
}

// Filter .
func (s *Service) Filter(c context.Context, msgs []string) (err error) {
	msgMap := make(map[string]string, len(msgs))
	for _, v := range msgs {
		msgMap[v] = v
	}
	res, err := s.filterClient.MFilter(c, &filterclient.MFilterReq{Area: model.FilterArea, MsgMap: msgMap})
	if err != nil {
		log.Error("Filter msg(%v) error(%v)", msgs, err)
		return
	}
	for _, msg := range msgMap {
		if v, ok := res.RMap[msg]; ok {
			if v.Level >= model.FilterLevel {
				return xcode.SpaceBanFilter
			}
		}
	}
	return nil
}

func (s *Service) initCron() error {
	s.cron = cron.New()
	var err error
	s.loadPhotoMallList()
	if err = s.cron.AddFunc(s.c.Spec.PhotoMall, s.loadPhotoMallList); err != nil {
		return err
	}
	s.loadBlacklist()
	if err = s.cron.AddFunc(s.c.Spec.BlackList, s.loadBlacklist); err != nil {
		return err
	}
	s.loadSysNotice()
	if err = s.cron.AddFunc(s.c.Spec.SysNotice, s.loadSysNotice); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s *Service) Close() (err error) {
	s.cron.Stop()
	return nil
}

func (s *Service) BatchNFTRegion(ctx context.Context, mids []int64) map[int64]pangugsgrpc.NFTRegionType {
	midNFTRegionMap := make(map[int64]pangugsgrpc.NFTRegionType)
	if len(mids) == 0 {
		return midNFTRegionMap
	}
	nftMidMap := make(map[string]int64)
	if res, e := s.accmGRPC.NFTBatchInfo(ctx, &accmgrpc.NFTBatchInfoReq{Mids: mids, Status: "inUsing", Source: "face"}); e != nil {
		log.Error("s.accmGRPC.NFTBatchInfo(%v) error(%v)", mids, e)
	} else {
		if res == nil || len(res.NftInfos) == 0 {
			return midNFTRegionMap
		}
		var nftIds []string
		for midS, nftInfo := range res.NftInfos {
			if mid, err := strconv.ParseInt(midS, 10, 64); err == nil && mid > 0 && nftInfo != nil && nftInfo.NftId != "" {
				nftMidMap[nftInfo.NftId] = mid
				nftIds = append(nftIds, nftInfo.NftId)
			}
		}
		if res, e := s.pangugsGRPC.GetNFTRegion(ctx, &pangugsgrpc.GetNFTRegionReq{NftId: nftIds}); e != nil {
			log.Error("s.pangugsGRPC.GetNFTRegion(%v) error(%v)", nftIds, e)
		} else {
			if res == nil || len(res.Region) == 0 {
				return midNFTRegionMap
			}
			for nftId, nftRegion := range res.Region {
				mid := nftMidMap[nftId]
				if nftRegion != nil && mid > 0 {
					if nftRegion.Type == pangugsgrpc.NFTRegionType_DEFAULT {
						midNFTRegionMap[mid] = pangugsgrpc.NFTRegionType_MAINLAND
					} else {
						midNFTRegionMap[mid] = nftRegion.Type
					}
				}
			}
		}
	}
	return midNFTRegionMap
}
