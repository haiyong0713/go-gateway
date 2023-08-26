package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/utils/collection"
	appchanmdl "go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	honorgrpc "go-gateway/app/app-svr/archive-honor/service/api"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	resmdl "go-gateway/app/app-svr/resource/service/api/v1"
	steinsapi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"
	chmdl "go-gateway/app/web-svr/web/interface/model/channel"
	gateecode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relamdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	ugcmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	upmdl "git.bilibili.co/bapis/bapis-go/archive/service/up"
	playeronlinegrpc "git.bilibili.co/bapis/bapis-go/bilibili/app/playeronline/v1"
	v1 "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
	dmmdl "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	sharemdl "git.bilibili.co/bapis/bapis-go/community/interface/share"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	cremdl "git.bilibili.co/bapis/bapis-go/creative/open/service"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	activegrpc "git.bilibili.co/bapis/bapis-go/manager/service/active"
	shareadmin "git.bilibili.co/bapis/bapis-go/platform/admin/share"
	resv2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	videogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
)

var (
	_emptyReplyHot       = new(model.ReplyHot)
	_emptyBvArc          = make([]*model.BvArc, 0)
	_ccTPArchive   int32 = 1
)

const (
	_japan          = "日本"
	_businessAppeal = 1
	_notForward     = 0
	_member         = 10000
	_shareArcType   = 3
	_shareLiveType  = 6
	_hasUGCPay      = 1
	_ugcOtypeArc    = "archive"
	_ugcCurrencyBp  = "bp"
	_ugcAssetPaid   = "paid"
	_ugcPaidState   = 1
	_labelTypeHot   = 1
	_testerGroup    = 20
	_forbidReco     = 1
	// content-flow-control.service gRPC infos limit
	_cfcAttributeLimit = 30
	_banPlatform       = "wechat"

	_noRecommendLive     int64 = 63 // 屏蔽直播推荐
	_noRecommendActivity int64 = 64 // 屏蔽活动推荐
)

// nolint: gomnd,gocognit
func (s *Service) View(c context.Context, aid, cid, mid int64, cdnIP, outRefer, buvid string) (*model.View, *arcmdl.Arc, error) {
	var (
		video                                *arcmdl.VideoReply
		pDesc, forward, report, needCheckFav bool
		remoteIP                             = metadata.String(c, metadata.RemoteIP)
	)
	const _platform = "pc" // 临时于fix线上问题
	viewReply, err := s.arcGRPC.SteinsGateView(c, &arcmdl.SteinsGateViewRequest{Aid: aid, Mid: mid, Platform: _platform})
	if err != nil {
		log.Error("s.steinsGRPC.SteinsGateView(%d) error %v", aid, err)
		return nil, nil, s.slbRetryCode(err)
	}
	if viewReply == nil || viewReply.Arc == nil {
		return nil, nil, errors.WithMessage(xecode.NothingFound, "稿件不存在")
	}
	if _, ok := s.specialMids[viewReply.Arc.Author.Mid]; ok && env.DeployEnv == env.DeployEnvProd {
		return nil, nil, errors.WithMessage(xecode.NothingFound, "特殊up主稿件不展示")
	}
	if viewReply.Arc.FirstCid == 0 {
		if len(viewReply.Pages) > 0 {
			viewReply.Arc.FirstCid = viewReply.Pages[0].Cid
		}
	}
	if viewReply.Arc.State == arcmdl.StateForbidLock && viewReply.Arc.Forward > 0 {
		forward = true
	} else if viewReply.Arc.State == arcmdl.StateForbidRecicle || viewReply.Arc.State == arcmdl.StateForbidLock {
		if viewReply.Arc.ReportResult == "" {
			return nil, nil, errors.WithMessage(xecode.NothingFound, "稿件打回或锁定")
		}
		report = true
	}
	check := []int32{arcmdl.StateForbidWait, arcmdl.StateForbidFixed, arcmdl.StateForbidLater, arcmdl.StateForbidAdminDelay}
	for _, v := range check {
		if viewReply.Arc.State == v {
			return nil, nil, ecode.ArchiveChecking
		}
	}
	if err := func() error {
		if viewReply.Arc.IsNormalPremiere() {
			// 首映稿件
			return nil
		}
		if viewReply.Arc.State == arcmdl.StateForbidSteins || viewReply.Arc.State == arcmdl.StateForbidUserDelay {
			return ecode.ArchivePass
		}
		if !viewReply.Arc.IsNormal() && !forward && !report {
			return ecode.ArchiveDenied
		}
		return nil
	}(); err != nil {
		return nil, nil, err
	}
	// pugv pay arc not allow
	if viewReply.Arc.AttrVal(arcmdl.AttrBitIsPUGVPay) == arcmdl.AttrYes {
		return nil, nil, errors.WithMessage(xecode.NothingFound, "pugv付费稿件不允许访问")
	}
	if viewReply.Arc.AttrValV2(arcmdl.AttrBitV2OnlyFavView) == arcmdl.AttrYes {
		if mid == 0 {
			return nil, nil, errors.WithMessage(xecode.NothingFound, "收藏可见稿件未登录不允许访问")
		}
		needCheckFav = func() bool {
			if mid == viewReply.Arc.Author.Mid {
				return false
			}
			for _, sf := range viewReply.Arc.StaffInfo {
				if sf.Mid == mid {
					return false
				}
			}
			return true
		}()
	}
	cfcInfos, err := s.batchCfcInfos(c, []int64{viewReply.Arc.Aid})
	cfcItem, ok := cfcInfos[viewReply.Arc.Aid]
	if !ok {
		log.Warn("s.View forbidden is empty aid:%d", viewReply.Arc.Aid)
	}
	arcForbidden := model.ItemToArcForbidden(cfcItem)
	if arcForbidden.NoSearch {
		if trimOutRefer := strings.TrimSpace(outRefer); trimOutRefer != "" {
			var has bool
			for _, val := range s.c.ArcNoSearch.Referers {
				if strings.Contains(trimOutRefer, val) {
					has = true
					break
				}
			}
			if !has {
				return nil, nil, errors.WithMessage(xecode.NothingFound, "搜索禁止稿件非法referer不允许访问")
			}
		}
	}
	// visible only author themself（to do）
	if viewReply.Arc.AttrValV2(arcmdl.AttrBitV2OnlySely) == arcmdl.AttrYes {
		if mid == 0 {
			return nil, nil, errors.WithMessage(xecode.NothingFound, "非登陆用户自见稿件不可见")
		}
		if mid != viewReply.Arc.Author.Mid {
			return nil, nil, errors.WithMessage(xecode.NothingFound, "自见稿件仅允许稿件作者自见")
		}
	}
	viewArc := new(model.ViewArc)
	viewArc.FmtWebArc(viewReply.Arc)
	if arcForbidden.NoShare {
		// forbid to share
		viewArc.Rights.NoShare = 1
	}
	rs := &model.View{ViewArc: viewArc, Bvid: s.avToBv(viewArc.Aid), Pages: viewReply.Pages, UserGarb: new(model.UserGarb)}
	group := errgroup.WithContext(c)
	if !forward && !report {
		if viewReply.Arc.AttrVal(arcmdl.AttrBitLimitArea) == arcmdl.AttrYes {
			group.Go(func(ctx context.Context) error {
				return s.zlimit(ctx, rs, mid, remoteIP, cdnIP)
			})
		}
		if _, ok := model.LimitTypeIDMap[viewReply.Arc.TypeID]; ok {
			group.Go(func(ctx context.Context) error {
				return s.specialLimit(ctx, rs, remoteIP)
			})
		}
		group.Go(func(ctx context.Context) error {
			return s.checkAccess(ctx, mid, viewReply.Arc.Access, rs)
		})
	}
	if viewReply.Arc.SeasonID > 0 {
		// ugc season info
		group.Go(func(ctx context.Context) error {
			if seasonReply, e := s.ugcSeasonGRPC.View(ctx, &api.ViewRequest{SeasonID: viewReply.Arc.SeasonID}); e != nil {
				log.Error("s.ugcSeasonGRPC.Season seasonID(%d) error(%v)", viewReply.Arc.SeasonID, e)
			} else {
				rs.UGCSeason = model.CopyFromUGCSeason(seasonReply.View)
				if rs.UGCSeason.SeasonType == 1 {
					rs.IsSeasonDisplay = s.basisSeasonABTest(buvid)
				}
			}
			return nil
		})
		// activity season
		if viewReply.Arc.AttrValV2(arcmdl.AttrBitV2ActSeason) == arcmdl.AttrYes && !s.c.Rule.ForbidBnjJump {
			group.Go(func(ctx context.Context) error {
				actData, ok := s.activitySeasonIDMem[viewReply.Arc.SeasonID]
				if ok && actData != nil {
					if actData.ActivityURL == "" {
						log.Error("View s.activitySeasonIDMem seasonID:%d ActivityUrl nil", viewReply.Arc.SeasonID)
						return nil
					}
					if whiteListErr := checkActivitySeasonWhiteList(mid, actData.Whitelist); whiteListErr != nil {
						log.Warn("View checkActivitySeasonWhiteList mid:%d aid:%d seasonID:%d forbid", mid, aid, viewReply.Arc.SeasonID)
						return nil
					}
					rs.ViewArc.FestivalJumpUrl = fmt.Sprintf("%s?bvid=%s", actData.ActivityURL, rs.Bvid)
					return nil
				}
				commonAct, actErr := s.comActiveGRPC.CommonActivity(ctx, &activegrpc.CommonActivityReq{
					SeasonId: viewReply.Arc.SeasonID,
					Plat:     _commonActPlayPc,
					Mid:      mid,
				})
				if actErr != nil {
					log.Error("View s.comActiveGRPC.CommonActivity seasonID:%d mid:%d error:%v", viewReply.Arc.SeasonID, mid, actErr)
					return nil
				}
				if !verifyAct(commonAct) {
					log.Error("View s.comActiveGRPC.CommonActivity seasonID:%d mid:%d aid:%d verifyAct fail", viewReply.Arc.SeasonID, mid, aid)
					return nil
				}
				if whiteListErr := checkActivitySeasonWhiteList(mid, commonAct.ActivePlay.Whitelist); whiteListErr != nil {
					log.Warn("View checkActivitySeasonWhiteList mid:%d aid:%d seasonID:%d forbid", mid, aid, viewReply.Arc.SeasonID)
					return nil
				}
				rs.ViewArc.FestivalJumpUrl = fmt.Sprintf("%s?bvid=%s", commonAct.ActivePlay.ActivityUrl, rs.Bvid)
				return nil
			})
		}
	}
	group.Go(func(ctx context.Context) error {
		dmCid := cid
		if dmCid == 0 {
			dmCid = viewReply.Arc.FirstCid
		}
		rs.Subtitle = s.dmSubtitle(ctx, aid, dmCid)
		return nil
	})
	if rs.ViewArc.Rights.UGCPay == _hasUGCPay {
		group.Go(func(ctx context.Context) error {
			rs.Asset = s.ugcPayAsset(ctx, aid)
			return nil
		})
	}
	// fill staff info
	if viewReply.Arc.AttrVal(arcmdl.AttrBitIsCooperation) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			rs.Staff = s.staffInfo(ctx, viewReply.Arc.Author.Mid, viewReply.Arc.StaffInfo)
			return nil
		})
	}
	if cid > 0 {
		group.Go(func(ctx context.Context) error {
			if video, err = s.arcGRPC.Video(ctx, &arcmdl.VideoRequest{Aid: aid, Cid: cid}); err != nil {
				log.Error("s.arcGRPC.Video(%d,%d) error %v", aid, cid, err)
				err = nil
			} else if video.Page != nil && video.Page.Desc != "" {
				rs.Desc = video.Page.Desc
				pDesc = true
			}
			return nil
		})
	}
	if viewReply.Arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			steinsView, e := s.steinsGRPC.GraphView(ctx, &steinsapi.GraphViewReq{Aid: aid})
			if e != nil {
				return e
			}
			// replace page and first cid
			if steinsView.Page != nil {
				rs.Pages = []*arcmdl.Page{model.ArchivePage(steinsView.Page)}
				rs.FirstCid = steinsView.Page.Cid
			}
			rs.Stat.Evaluation = steinsView.Evaluation
			rs.SteinGuideCid = s.steinGuideCid
			return nil
		})
	}
	if needCheckFav {
		group.Go(func(ctx context.Context) error {
			resp, e := s.favGRPC.IsFavored(ctx, &favgrpc.IsFavoredReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oid: aid})
			if e != nil {
				log.Error("View s.favGRPC.IsFavored(%d,%d) error(%v)", mid, aid, e)
				return xecode.NothingFound
			}
			if resp == nil || !resp.Faved {
				log.Warn("View aid:%d AttrBitV2OnlyFavView mid:%d not faved", aid, mid)
				return xecode.NothingFound
			}
			return nil
		})
	}
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			userEquip, e := s.garbGRPC.ThumbupUserEquip(ctx, &garbgrpc.ThumbupUserEquipReq{
				Mid: mid,
			})
			if e != nil {
				log.Error("View s.garbGRPC.ThumbupUserEquip mid:%d error:%+v", mid, e)
				return nil
			}
			if userEquip == nil {
				log.Warn("View s.garbGRPC.ThumbupUserEquip mid:%d reply nil", mid)
				return nil
			}
			rs.UserGarb = &model.UserGarb{URLImageAniCut: userEquip.URLImageAniCut}
			return nil
		})
	}
	// 点赞动画和icon
	group.Go(func(ctx context.Context) error {
		res, err := s.GetMultiLikeAnimation(ctx, aid)
		if err != nil {
			log.Error("s.GetMultiLikeAnimation aid:%d,err%+v", aid, err)
			return nil
		}
		if like, ok := res[aid]; ok {
			rs.LikeIcon = like.WebLikeIcon
			if like.LikeCartoon != "" {
				// avid纬度 >> 装扮 >> 默认
				rs.UserGarb = &model.UserGarb{URLImageAniCut: like.LikeCartoon}
			}
		}
		return nil
	})
	if viewReply.Arc.AttrVal(arcmdl.AttrBitHasArgument) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			vuRes, e := s.videoUpGRPC.MultiArchiveArgument(ctx, &videogrpc.MultiArchiveArgumentReq{Aids: []int64{aid}})
			if e != nil {
				log.Error("View s.videoUpGRPC.MultiArchiveArgument aid:%d error:%+v", aid, e)
				return nil
			}
			if vuRes == nil || vuRes.Arguments == nil || vuRes.Arguments[aid] == nil {
				log.Error("View s.videoUpGRPC.MultiArchiveArgument res error aid:%d", aid)
				return nil
			}
			argue := vuRes.Arguments[aid]
			rs.Stat.ArgueMsg = argue.GetArgueMsg()
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		var err error
		in := &honorgrpc.HonorRequest{Aid: aid}
		reply, err := s.honorGRPC.Honor(ctx, in)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		rs.HonorReply = &model.HonorReply{}
		for _, honor := range reply.GetHonor() {
			if honor == nil {
				continue
			}
			var weeklyRecommendNum int64
			if honor.Type == 2 { // 荣誉类型 1-入站必刷 2-每周必看 3-日排行榜 4-热门 5-精选频道
				func() {
					u, err := url.Parse(honor.Url)
					if err != nil {
						log.Error("%+v", err)
						return
					}
					if weeklyRecommendNum, err = strconv.ParseInt(u.Query().Get("num"), 10, 64); err != nil {
						log.Error("%+v", err)
						return
					}
				}()
			}
			honor.Url, honor.NaUrl = "", ""
			rs.HonorReply.Honor = append(rs.HonorReply.Honor, &model.Honor{Honor: honor, WeeklyRecommendNum: weeklyRecommendNum})
		}
		return nil
	})
	if rs.Premiere != nil {
		group.Go(func(ctx context.Context) error {
			reply, err := s.dao.GetPremiereSidByAid(ctx, viewReply.GetAid())
			if err != nil {
				log.Error("error:%+v, aid:%d", err, aid)
				return nil
			}
			if reply != nil {
				rs.Premiere.SID = reply.Sid
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		log.Error("View aid:%d group.Wait error:%+v", aid, err)
		return nil, nil, err
	}
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) error {
		if pDesc {
			return nil
		}
		desc, reply, mids, err := s.description(ctx, aid, rs.Desc)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		rs.Desc = desc
		var accReply *accmdl.InfosReply
		if len(mids) != 0 {
			accReply, err = s.accGRPC.Infos3(ctx, &accmdl.MidsReq{Mids: mids})
			if err != nil {
				log.Error("%+v", err)
			}
		}
		rs.DescV2 = s.DescV2ParamsMerge(ctx, reply, accReply)
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if viewReply.Arc.AttrVal(arcmdl.AttrBitIsPGC) == arcmdl.AttrYes {
			return nil
		}
		if err := s.initDownload(c, viewReply.Arc, rs, mid, cdnIP); err != nil {
			log.Error("s.initDownload aid:%d mid:%d ip:%s error:%+v", rs.Aid, mid, cdnIP, err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if !forward {
		rs.Forward = _notForward
	}
	if !report {
		rs.ReportResult = ""
	}
	if viewReply.Arc.AttrVal(arcmdl.AttrBitJumpUrl) == arcmdl.AttrNo {
		rs.RedirectURL = ""
	}
	if viewReply.Arc.AttrVal(arcmdl.AttrBitLimitArea) == arcmdl.AttrYes || viewReply.Arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		rs.NoCache = true
	}
	s.initLabel(rs)
	return rs, viewReply.Arc, nil
}

func (s *Service) zlimit(c context.Context, view *model.View, mid int64, remoteIP, cdnIP string) error {
	ipInfo, err := s.locGRPC.Info(c, &locgrpc.InfoReq{Addr: remoteIP})
	if err != nil {
		log.Error("zlimit s.locGRPC.Info aid:%d,ip:%s error:%v", view.Aid, remoteIP, err)
		return err
	}
	if ipInfo.GetZoneId() == 0 {
		log.Warn("zlimit s.locGRPC.Info aid:%d,ip:%s zone id not found", view.Aid, remoteIP)
		return xecode.NothingFound
	}
	data, err := s.locGRPC.Archive(c, &locgrpc.ArchiveReq{Aid: view.Aid, Mid: mid, IpAddr: remoteIP, CdnAddr: cdnIP})
	if err != nil {
		log.Error("zlimit s.locGRPC.Archive aid:%d mid:%d ip:%s cdnIP:%s error:%v", view.Aid, mid, remoteIP, cdnIP, err)
		return err
	}
	if data.Auth.Play == int64(locgrpc.Status_Forbidden) {
		log.Warn("s.locGRPC.Archive aid:%d ip:%s: zlimit.Forbidden", view.Aid, remoteIP)
		return xecode.NothingFound
	}
	return nil
}

// specialLimit spacialLimit special type id limit in japan
func (s *Service) specialLimit(c context.Context, view *model.View, remoteIP string) (err error) {
	var zone *locgrpc.InfoReply
	view.NoCache = true
	if zone, err = s.locGRPC.Info(c, &locgrpc.InfoReq{Addr: remoteIP}); err != nil || zone == nil {
		log.Error("s.locGRPC.Info(%s) error(%v) or zone is nil", remoteIP, err)
		err = nil
	} else if zone.Country == _japan {
		err = xecode.NothingFound
	}
	return
}

// avid >> up粉丝/up自己 >> 默认
func (s *Service) UpLikeImg(ctx context.Context, mid, vmid, Aid int64, following bool) (*model.UpLikeImg, error) {
	// 未登录的展示默认的
	if mid < 1 {
		return &model.UpLikeImg{}, nil
	}
	// avid 优先
	reply, err := s.creativeGRPC.UpLikeImg(ctx, &cremdl.UpLikeImgReq{Mid: vmid, Avid: Aid})
	if err != nil {
		return nil, err
	}
	avidType := int64(2)
	if reply.GetType() == avidType {
		return &model.UpLikeImg{
			UpLikeImg: reply,
		}, nil
	}
	// UP主有配置自定义一键三连动画
	// 触发用户是UP的粉丝，非粉丝还是展示系统样式
	// UP主自己可以触发蓄力动画
	if ok := func() bool {
		if following || mid == vmid {
			return true
		}
		reply, err := s.accGRPC.Relation3(ctx, &accmdl.RelationReq{Mid: mid, Owner: vmid})
		if err != nil {
			log.Error("%+v", err)
			return false
		}
		return reply.GetFollowing()
	}(); !ok {
		return &model.UpLikeImg{}, nil
	}
	return &model.UpLikeImg{
		UpLikeImg: reply,
	}, nil
}

// checkAccess check mid aid access
func (s *Service) checkAccess(c context.Context, mid int64, access int32, view *model.View) error {
	if access == 0 {
		return nil
	}
	view.NoCache = true
	if mid <= 0 {
		log.Warn("user not login  aid(%d)", view.Aid)
		return xecode.AccessDenied
	}
	p, err := s.accGRPC.Card3(c, &accmdl.MidReq{Mid: mid})
	if err != nil {
		log.Error("s.accGRPC.Card3(%d) error(%v)", mid, err)
		return err
	}
	if p == nil {
		log.Warn("Info2 result is null aid(%d) state(%d) access(%d)", view.Aid, view.State, access)
		return xecode.AccessDenied
	}
	card := p.Card
	isVip := (card.Vip.Type > 0) && (card.Vip.Status == 1)
	if access > 0 && card.Rank < access && (!isVip) {
		log.Warn("mid(%d) rank(%d) vip(tp:%d,status:%d) have not access(%d) view archive(%d) ", mid, card.Rank, card.Vip.Type, card.Vip.Status, access, view.Aid)
		if mid > 0 {
			return xecode.NothingFound
		}
		return ecode.ArchiveNotLogin
	}
	return nil
}

func (s *Service) initDownload(c context.Context, arc *arcmdl.Arc, v *model.View, mid int64, cdnIP string) (err error) {
	var download int64
	if arc.AttrVal(arcmdl.AttrBitLimitArea) == arcmdl.AttrYes {
		if download, err = s.downLimit(c, mid, v.Aid, cdnIP); err != nil {
			return
		}
	} else {
		download = int64(locgrpc.StatusDown_AllowDown)
	}
	if download == int64(locgrpc.StatusDown_ForbiddenDown) {
		v.Rights.Download = int32(download)
		return
	}
	for _, p := range v.Pages {
		if p.From == "qq" {
			download = int64(locgrpc.StatusDown_ForbiddenDown)
			break
		}
	}
	v.Rights.Download = int32(download)
	return
}

// downLimit ip limit
func (s *Service) downLimit(c context.Context, mid, aid int64, cdnIP string) (down int64, err error) {
	var (
		reply *locgrpc.ArchiveReply
		ip    = metadata.String(c, metadata.RemoteIP)
	)
	if reply, err = s.locGRPC.Archive(c, &locgrpc.ArchiveReq{Aid: aid, Mid: mid, IpAddr: ip, CdnAddr: cdnIP}); err != nil {
		log.Error("s.locGRPC.Archive(%d) error(%v)", mid, err)
		return
	}
	if reply.Auth != nil {
		if reply.Auth.Play == int64(locgrpc.Status_Forbidden) {
			err = xecode.AccessDenied
		} else {
			down = reply.Auth.Down
		}
	}
	return
}

func (s *Service) dmSubtitle(c context.Context, aid, cid int64) (subtitle *model.Subtitle) {
	var (
		dmSubReply *dmmdl.SubtitleGetReply
		err        error
		mids       []int64
		infosReply *accmdl.InfosReply
		subs       []*model.SubtitleItem
	)
	subtitle = new(model.Subtitle)
	if dmSubReply, err = s.dmGRPC.SubtitleGet(c, &dmmdl.SubtitleGetReq{Aid: aid, Oid: cid, Type: 1}); err != nil {
		log.Warn("dmSubtitle s.dmGRPC.SubtitleGet aid(%d) cid(%d) warn(%v)", aid, cid, err)
	} else if dmSubReply != nil && dmSubReply.Subtitle != nil {
		subtitle.AllowSubmit = dmSubReply.Subtitle.AllowSubmit
		if len(dmSubReply.Subtitle.Subtitles) > 0 {
			for _, v := range dmSubReply.Subtitle.Subtitles {
				if v.AuthorMid > 0 {
					mids = append(mids, v.AuthorMid)
				}
			}
			infoData := make(map[int64]*accmdl.Info)
			if len(mids) > 0 {
				if infosReply, err = s.accGRPC.Infos3(c, &accmdl.MidsReq{Mids: mids, RealIp: metadata.String(c, metadata.RemoteIP)}); err != nil {
					log.Error("dmSubtitle aid(%d) cid(%d) s.acc.Infos3 mids(%v) error(%v)", aid, cid, mids, err)
				} else {
					infoData = infosReply.Infos
				}
			}
			for _, v := range dmSubReply.Subtitle.Subtitles {
				sub := &model.SubtitleItem{VideoSubtitle: v, Author: &accmdl.Info{Mid: v.AuthorMid}}
				if info, ok := infoData[v.AuthorMid]; ok && info != nil {
					info.Birthday = 0
					sub.Author = info
				}
				subs = append(subs, sub)
			}
			subtitle.List = subs
		}
	}
	if len(subtitle.List) == 0 {
		subtitle.List = make([]*model.SubtitleItem, 0)
	}
	return
}

// ArchiveStat get archive stat data by aid.
func (s *Service) ArchiveStat(c context.Context, aid int64) (*model.Stat, error) {
	arcReply := func() *arcmdl.ArcReply {
		if aid == s.c.Bnj2020.LiveAid && s.bnj20Cache.LiveArc != nil {
			return s.bnj20Cache.LiveArc
		}
		arcReply, err := s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if err != nil {
			log.Error("s.arcGRPC.Arc(%d) error(%v)", aid, err)
			return nil
		}
		return arcReply
	}()
	if arcReply == nil || arcReply.Arc == nil {
		log.Warn("ArchiveStat aid:%d arcReply(%+v) fail", aid, arcReply)
		return nil, xecode.NothingFound
	}
	arc := arcReply.Arc
	if !model.CheckAllowState(arc) {
		return nil, xecode.AccessDenied
	}
	var statView interface{}
	statView = arc.Stat.View
	if arc.Access > 0 {
		statView = "--"
	}
	stat := &model.Stat{
		Aid:       arc.Stat.Aid,
		View:      statView,
		Danmaku:   arc.Stat.Danmaku,
		Reply:     arc.Stat.Reply,
		Fav:       arc.Stat.Fav,
		Coin:      arc.Stat.Coin,
		Share:     arc.Stat.Share,
		Like:      arc.Stat.Like,
		NowRank:   arc.Stat.NowRank,
		HisRank:   arc.Stat.HisRank,
		NoReprint: arc.Rights.NoReprint,
		Copyright: arc.Copyright,
	}
	group := errgroup.WithContext(c)
	if arc.AttrVal(arcmdl.AttrBitHasArgument) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			vuRes, e := s.videoUpGRPC.MultiArchiveArgument(ctx, &videogrpc.MultiArchiveArgumentReq{Aids: []int64{aid}})
			if e != nil {
				log.Error("ArchiveStat s.videoUpGRPC.MultiArchiveArgument aid:%d error:%+v", aid, e)
				return nil
			}
			if vuRes == nil || vuRes.Arguments == nil || vuRes.Arguments[aid] == nil {
				log.Error("ArchiveStat s.videoUpGRPC.MultiArchiveArgument res error aid:%d", aid)
				return nil
			}
			argue := vuRes.Arguments[aid]
			stat.ArgueMsg = argue.ArgueMsg
			return nil
		})
	}
	if arc.AttrVal(arcmdl.AttrBitSteinsGate) == arcmdl.AttrYes {
		group.Go(func(ctx context.Context) error {
			if data, e := s.steinsGRPC.Evaluation(ctx, &steinsapi.EvaluationReq{Aid: aid}); e != nil {
				log.Error("ArchiveStat s.steinsGRPC.Evaluation(%d) error(%v)", aid, e)
			} else {
				stat.Evaluation = data.Eval
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Warn("ArchiveStat aid:%d group.Wait:%+v", aid, err)
	}
	stat.Bvid = s.avToBv(stat.Aid)
	return stat, nil
}

// AddShare share add count
func (s *Service) AddShare(c context.Context, aid, mid, roomID, upID, parentAreaID, areaID int64, riskParams *model.RiskManagement) (res *model.ShareRes, err error) {
	res = &model.ShareRes{
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	remoteIP := metadata.String(c, metadata.RemoteIP)
	arg := &sharemdl.ServiceClickReq{
		Oid:     aid,
		Mid:     mid,
		Type:    _shareArcType,
		Channel: "default",
		Ip:      remoteIP,
		ApiType: "old",
	}
	if roomID > 0 {
		arg.Oid = roomID
		arg.Type = _shareLiveType
		arg.UpId = upID
		arg.ParentAreaId = parentAreaID
		arg.AreaId = areaID
	}
	if arg.Type == _shareArcType {
		var arcReply *arcmdl.ArcReply
		arcReply, err = s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if err != nil {
			res.Shares = 0
			return res, err
		}
		if arcReply == nil || arcReply.Arc == nil || !arcReply.Arc.IsNormal() {
			res.Shares = 0
			return res, xecode.NothingFound
		}
		shareSource := "1" // 未登录
		if mid > 0 {
			shareSource = "2"
		}
		riskParams.ShareSource = shareSource
		riskParams.Action = _shareAddAction
		riskParams.Scene = _shareAddScene
		riskParams.Pubtime = arcReply.Arc.PubDate.Time().Format("2006-01-02 15:04:05")
		riskParams.Title = arcReply.Arc.Title
		riskParams.PlayNum = arcReply.Arc.Stat.View
		riskResult := s.RiskVerifyAndManager(c, riskParams)
		if riskResult != nil {
			res.GaiaResType = riskResult.GaiaResType
			res.IsRisk = riskResult.IsRisk
			res.GaiaData = riskResult.GaiaData
			return res, nil
		}
	}
	shareReply, err := s.shareGRPC.ServiceClick(c, arg)
	if err != nil {
		log.Error("ServiceClick oid=%d,mid=%d,error=%v", aid, mid, err)
		return
	}
	res.Shares = shareReply.GetCount()
	return
}

// Description get archive description by aid.
func (s *Service) Description(c context.Context, aid, page int64) (res string, err error) {
	var (
		viewReply *arcmdl.ViewReply
		video     *arcmdl.VideoReply
		cid       int64
		longDesc  *arcmdl.DescriptionReply
	)
	if viewReply, err = s.arcGRPC.View(c, &arcmdl.ViewRequest{Aid: aid}); err != nil {
		log.Error("s.arcGRPC.View(%d) error %v", aid, err)
		return
	}
	if viewReply != nil && viewReply.Arc != nil {
		if !viewReply.Arc.IsNormal() {
			err = ecode.ArchiveDenied
			return
		}
	}
	if page > 0 {
		if int(page-1) >= len(viewReply.Pages) || viewReply.Pages[page-1] == nil {
			err = xecode.NothingFound
			return
		}
		cid = viewReply.Pages[page-1].Cid
		if cid > 0 {
			if video, err = s.arcGRPC.Video(c, &arcmdl.VideoRequest{Aid: aid, Cid: cid}); err != nil {
				log.Error("s.arcGRPC.Video(%d,%d) error %v", aid, cid, err)
			} else if video.Page != nil && video.Page.Desc != "" {
				res = video.Page.Desc
				return
			}
		}
	} else {
		res = viewReply.Arc.Desc
	}
	if longDesc, err = s.arcGRPC.Description(c, &arcmdl.DescriptionRequest{Aid: aid}); err != nil {
		log.Error("s.arcGRPC.Description(%d) error(%v)", aid, err)
	} else if longDesc.Desc != "" {
		res = longDesc.Desc
	}
	return
}

// Desc2 get archive description by aid.
func (s *Service) Desc2(ctx context.Context, aid, page int64) (*model.DescReply, error) {
	viewReply, err := s.arcGRPC.View(ctx, &arcmdl.ViewRequest{Aid: aid})
	if err != nil {
		log.Error("s.arcGRPC.View(%d) error %v", aid, err)
		return nil, err
	}
	if viewReply != nil && viewReply.Arc != nil {
		if !viewReply.Arc.IsNormal() {
			return nil, ecode.ArchiveDenied
		}
	}
	desc := viewReply.Arc.Desc
	if page > 0 {
		if int(page-1) >= len(viewReply.Pages) || viewReply.Pages[page-1] == nil {
			return nil, xecode.NothingFound
		}
		cid := viewReply.Pages[page-1].Cid
		if cid > 0 {
			video, err := s.arcGRPC.Video(ctx, &arcmdl.VideoRequest{Aid: aid, Cid: cid})
			if err != nil {
				log.Error("s.arcGRPC.Video(%d,%d) error %v", aid, cid, err)
			}
			if video.GetPage().GetDesc() != "" {
				return &model.DescReply{
					Desc: video.GetPage().GetDesc(),
				}, nil
			}
		}
		desc = ""
	}
	desc, reply, mids, err := s.description(ctx, aid, desc)
	if err != nil {
		return nil, err
	}
	var accReply *accmdl.InfosReply
	if len(mids) != 0 {
		accReply, err = s.accGRPC.Infos3(ctx, &accmdl.MidsReq{Mids: mids})
		if err != nil {
			log.Error("%+v", err)
		}
	}
	descV2 := s.DescV2ParamsMerge(ctx, reply, accReply)
	return &model.DescReply{
		Desc:   desc,
		DescV2: descV2,
	}, nil
}

// nolint:gomnd
func (s *Service) description(ctx context.Context, aid int64, arcDesc string) (string, []*arcmdl.DescV2, []int64, error) {
	reply, err := s.arcGRPC.Description(ctx, &arcmdl.DescriptionRequest{Aid: aid})
	if err != nil {
		return "", nil, nil, err
	}
	desc := arcDesc
	if reply.GetDesc() != "" {
		desc = reply.GetDesc()
	}
	descV2 := reply.DescV2Parse
	//简介不为空但是desc_v2为空，则需要拼装desc_v2的type=1类型
	if len(descV2) == 0 {
		if desc != "" {
			descV2 = append(descV2, &arcmdl.DescV2{
				RawText: desc,
				Type:    1,
			})
		}
	}
	var mids []int64
	for _, v := range descV2 {
		if v == nil {
			continue
		}
		if v.Type != 2 {
			continue
		}
		mids = append(mids, v.BizId)
	}
	return desc, descV2, mids, nil
}

func (s *Service) DescV2ParamsMerge(c context.Context, arcDescV2 []*arcmdl.DescV2, accountInfos *accmdl.InfosReply) []*model.DescV2 {
	var descV2 []*model.DescV2
	for _, val := range arcDescV2 {
		if val == nil {
			continue
		}
		rawText, ok := accountInfos.GetInfos()[val.BizId]
		if ok {
			val.RawText = rawText.Name
		}
		descV2 = append(descV2, &model.DescV2{
			RawText: val.RawText,
			Type:    int64(val.Type),
			BizId:   val.BizId,
		})
	}
	return descV2
}

// ArcReport add archive report
func (s *Service) ArcReport(c context.Context, mid, aid, tp int64, reason, pics string) (err error) {
	if err = s.dao.ArcReport(c, mid, aid, tp, reason, pics); err != nil {
		log.Error("s.dao.ArcReport(%d,%d,%d,%s,%s) err (%v)", mid, aid, tp, reason, pics, err)
	}
	return
}

// AppealTags get appeal tags
func (s *Service) AppealTags(c context.Context) (rs json.RawMessage, err error) {
	if rs, err = s.dao.AppealTags(c, _businessAppeal); err != nil {
		log.Error("s.dao.AppealTags(1) error(%v)", err)
	}
	return
}

// ArcAppeal add archive appeal.
func (s *Service) ArcAppeal(c context.Context, mid int64, data map[string]string) (err error) {
	aid, _ := strconv.ParseInt(data["oid"], 10, 64)
	if err = s.dao.ArcAppealCache(c, mid, aid); err != nil {
		if err == ecode.ArcAppealLimit {
			log.Warn("s.arcAppealLimit mid(%d) aid(%d)", mid, aid)
			return
		}
		err = nil
	}
	if err = s.dao.ArcAppeal(c, mid, data, _businessAppeal); err != nil {
		log.Error("s.dao.ArcAppeal(%d,%v,1) error(%v)", mid, data, err)
		return
	}
	if err = s.dao.SetArcAppealCache(c, mid, aid); err != nil {
		log.Error("s.dao.SetArcAppealCache(%d,%d)", mid, aid)
		err = nil
	}
	return
}

// AuthorRecommend get author recommend data
// nolint: gocognit
func (s *Service) AuthorRecommend(c context.Context, aid int64) (res []*model.BvArc, err error) {
	var (
		arcReply   *arcmdl.ArcReply
		arc        *arcmdl.Arc
		arcErr     error
		aids       []int64
		upArcReply *uparcgrpc.ArcPassedReply
		arcs       *arcmdl.ArcsReply
		forbid     int64
	)
	defer func() {
		if len(res) == 0 {
			res = _emptyBvArc
		}
	}()
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if arcReply, arcErr = s.arcGRPC.Arc(ctx, &arcmdl.ArcRequest{Aid: aid}); arcErr != nil {
			log.Error("s.arcGRPC.Arc(%d) error(%v)", aid, arcErr)
		} else if arcReply != nil {
			arc = arcReply.Arc
		}
		return arcErr
	})
	group.Go(func(ctx context.Context) error {
		if viewAddit, e := s.videoUpGRPC.ArcViewAddit(ctx, &vuapi.ArcViewAdditReq{Aid: aid}); e != nil {
			log.Error("s.videoUpGRPC.ArcViewAddit aid(%d) error(%v)", aid, e)
		} else if viewAddit != nil && viewAddit.ForbidReco != nil {
			forbid = viewAddit.ForbidReco.State
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	if arc == nil || !arc.IsNormal() || forbid == _forbidReco {
		return
	}
	resAids := make(map[int64]int64)
	resAids[aid] = aid
	if upArcReply, err = s.upArcGRPC.ArcPassed(c, &uparcgrpc.ArcPassedReq{Mid: arcReply.Arc.Author.Mid, Pn: _firstPn, Ps: int64(s.c.Rule.AuthorRecCnt)}); err != nil {
		log.Error("s.upGRPC.UpArcs(%d) error(%v)", arcReply.Arc.Author.Mid, err)
		err = nil
	} else if upArcReply != nil {
		for _, v := range upArcReply.Archives {
			if _, ok := resAids[v.Aid]; !ok {
				res = append(res, model.CopyFromUpArcToBvArc(v, s.avToBv(v.Aid)))
				resAids[v.Aid] = v.Aid
				if len(res) >= s.c.Rule.AuthorRecCnt {
					return
				}
			}
		}
	}
	if len(res) < s.c.Rule.AuthorRecCnt {
		if aids, err = s.dao.RelatedAids(c, aid); err != nil {
			log.Error("s.dao.RelatedArchives(%d) error(%v)", aid, err)
			err = nil
		} else if len(aids) > 0 {
			ps := s.c.Rule.AuthorRecCnt - len(res)
			if len(aids) > ps {
				aids = aids[0:ps]
			}
			archivesArgLog("AuthorRecommend", aids)
			if arcs, err = s.arcGRPC.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
				log.Error("s.arcGRPC.Arcs(%v) error(%v)", aids, err)
				err = nil
			} else {
				for _, aid := range aids {
					if _, ok := resAids[aid]; !ok {
						if arc, ok := arcs.Arcs[aid]; ok && arc != nil && arc.IsNormal() {
							res = append(res, model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)))
							if len(res) >= s.c.Rule.AuthorRecCnt {
								return
							}
						}
					}
				}
			}
		}
	}
	return
}

// nolint:gocognit
func (s *Service) RelatedArcs(ctx context.Context, aid, mid int64, buvid string, needRcmdReason, needOperation, webRmRepeat, inActivity bool, arc *arcmdl.Arc) ([]*model.BvArc, *model.SpecRecm, *videogrpc.ForbidReco, error) {
	group := errgroup.WithContext(ctx)
	if arc == nil {
		group.Go(func(ctx context.Context) error {
			arcReply, arcErr := s.arcGRPC.Arc(ctx, &arcmdl.ArcRequest{Aid: aid})
			if arcErr != nil {
				log.Error("RelatedArcs s.arcGRPC.Arc(%d) error(%v)", aid, arcErr)
				return arcErr
			}
			arc = arcReply.GetArc()
			return nil
		})
	}
	var forbid int64
	var viewForbidReco *videogrpc.ForbidReco
	group.Go(func(ctx context.Context) error {
		viewAddit, e := s.videoUpGRPC.ArcViewAddit(ctx, &vuapi.ArcViewAdditReq{Aid: aid})
		if e != nil {
			log.Error("RelatedArcs s.videoUpGRPC.ArcViewAddit aid(%d) error(%v)", aid, e)
			return nil
		}
		if viewAddit != nil && viewAddit.ForbidReco != nil {
			forbid = viewAddit.ForbidReco.State
			viewForbidReco = viewAddit.ForbidReco
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, nil, viewForbidReco, err
	}
	if arc == nil || forbid == _forbidReco || arc.AttrVal(arcmdl.AttrBitIsPGC) == arcmdl.AttrYes {
		return _emptyBvArc, nil, viewForbidReco, nil
	}
	if err := func() error {
		if arc.IsNormalPremiere() {
			// 首映稿件
			return nil
		}
		if !arc.IsNormal() {
			return ecode.ArchiveDenied
		}
		return nil
	}(); err != nil {
		return _emptyBvArc, nil, viewForbidReco, nil
	}
	recAids, operAid, rcmd, err := func() ([]int64, int64, *model.SpecRecm, error) {
		rcmdNew, rcmdErr := s.dao.ArcRecommends(ctx, aid, mid, needOperation, webRmRepeat, inActivity, buvid)
		if rcmdErr != nil {
			return nil, 0, nil, rcmdErr
		}
		var (
			aids      []int64
			operAid   int64
			rcmd      *model.SpecRecm
			gameCheck bool
		)
		for _, v := range rcmdNew {
			if v == nil || v.ID <= 0 {
				continue
			}
			switch v.Goto {
			case model.GotoAv:
				aids = append(aids, v.ID)
				// 运营稿件卡
				if v.IsDalao == 1 {
					operAid = v.ID
				}
			case model.GotoGame:
				// 特殊运营卡游戏类型只有一张
				if gameCheck {
					continue
				}
				gameCheck = true
				gameInfo, e := s.dao.GameInfo(ctx, v.ID)
				if e != nil {
					log.Error("RelatedArcs game data (%d) error(%v)", v.ID, e)
					continue
				}
				if gameInfo != nil {
					rcmd = &model.SpecRecm{
						Type: model.SpecRecmTypeGame,
						Game: gameInfo,
					}
				}
			case model.GotoSpecial:
				if card, ok := s.specRcmdCard[v.ID]; ok && card != nil {
					rcmd = &model.SpecRecm{
						Type: model.SpecRecmTypeCard,
						Card: card,
					}
				} else {
					log.Warn("RelatedArcs id:%d card not found", v.ID)
				}
			default:
				log.Warn("RelatedArcs ArcRecommends goto:%s not support", v.Goto)
			}
		}
		return aids, operAid, rcmd, nil
	}()
	if err != nil {
		log.Error("RelatedArcs s.dao.ArcRecommends(%d) error(%v)", aid, err)
		// degrade
		recAids, err = s.dao.RelatedAids(ctx, aid)
		if err != nil {
			log.Error("RelatedArcs s.dao.RelatedAids(%d) error(%v)", aid, err)
			return nil, nil, viewForbidReco, err
		}
	}
	if arc.AttrValV2(arcmdl.AttrBitV2CleanMode) == arcmdl.AttrYes {
		rcmd = nil
		operAid = 0
	}
	if len(recAids) == 0 {
		return _emptyBvArc, rcmd, viewForbidReco, nil
	}
	var afAids []int64
	aidMap := make(map[int64]int64, len(recAids))
	for _, recAid := range recAids {
		if aid == recAid {
			continue
		}
		if _, ok := aidMap[recAid]; ok {
			continue
		}
		aidMap[recAid] = recAid
		afAids = append(afAids, recAid)
	}
	if len(afAids) > s.c.Rule.RelatedArcCnt {
		afAids = afAids[:s.c.Rule.RelatedArcCnt]
	}
	arcsReply, arcErr := s.arcGRPC.Arcs(ctx, &arcmdl.ArcsRequest{Aids: afAids})
	if arcErr != nil {
		log.Error("RelatedArcs s.arcGRPC.Arcs(%v) error(%v)", afAids, arcErr)
		return _emptyBvArc, rcmd, viewForbidReco, nil
	}
	var list []*model.BvArc
	for _, afAid := range afAids {
		if afArc, ok := arcsReply.GetArcs()[afAid]; ok && afArc.IsNormal() {
			bvArc := model.CopyFromArcToBvArc(afArc, s.avToBv(afArc.Aid))
			if afAid == operAid {
				rcmd = &model.SpecRecm{
					Type:    model.SpecRecmTypeArc,
					Archive: bvArc,
				}
				continue
			}
			list = append(list, bvArc)
		}
	}
	if len(list) == 0 {
		return _emptyBvArc, rcmd, viewForbidReco, nil
	}
	s.rcmdExtra(ctx, list, needRcmdReason)
	return list, rcmd, viewForbidReco, nil
}

// nolint:gomnd,gocognit
func (s *Service) rcmdExtra(ctx context.Context, list []*model.BvArc, needRcmdReason bool) {
	var aids, seasonIDs []int64
	seasonIDm := map[int64]struct{}{}
	for _, value := range list {
		seasonID := value.GetSeasonID()
		if _, ok := seasonIDm[seasonID]; !ok && seasonID > 0 && len(seasonIDs) < 50 {
			seasonIDs = append(seasonIDs, seasonID)
			seasonIDm[seasonID] = struct{}{}
		}
		if len(aids) < 40 {
			aids = append(aids, value.GetAid())
		}
	}
	var (
		seasonsMap map[int64]*api.Season
		honorsMap  map[int64]*honorgrpc.HonorReply
	)
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		if len(seasonIDs) == 0 {
			return nil
		}
		reply, err := s.ugcSeasonGRPC.Seasons(ctx, &api.SeasonsRequest{SeasonIds: seasonIDs})
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		seasonsMap = reply.GetSeasons()
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if !needRcmdReason {
			return nil
		}
		honorsReq := &honorgrpc.HonorsRequest{Aids: aids}
		honors, err := s.honorGRPC.Honors(ctx, honorsReq)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		honorsMap = honors.GetHonors()
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	var count int
	for _, val := range list {
		if season, ok := seasonsMap[val.SeasonID]; ok {
			// 是否基础合集
			// 0:精选合集 1:基础合集
			val.SeasonType = season.AttrVal(api.AttrSnType)
		}
		if needRcmdReason {
			if count < 4 {
				val.RcmdReason = func() string {
					if _, ok := s.hotLabelAids[val.GetAid()]; ok {
						return "热门"
					}
					if val.Stat.View > 100000000 {
						return "播放破亿"
					}
					var honorType int32
					if honor, ok := honorsMap[val.Aid]; ok {
						if len(honor.GetHonor()) != 0 {
							honorType = honor.GetHonor()[0].Type
						}
					}
					if honorType == 1 {
						return "入站必刷"
					}
					if val.Stat.View > 10000000 {
						return "千万播放"
					}
					if honorType == 2 {
						return "每周必看"
					}
					if val.Stat.View > 1000000 {
						return "百万播放"
					}
					return ""
				}()
				if val.RcmdReason != "" {
					count++
				}
			}
		}
	}
}

// nolint:gocognit
func (s *Service) Detail(c context.Context, aid, mid int64, cdnIP, outRefer, buvid, recommendType, platform string, needRcmdReason, needOperation, webRmRepeat, needHotShare, needElec bool, pageNo int) (*model.Detail, error) {
	if isInList := s.checkCommonBWList(c, aid, platform); isInList {
		return nil, errors.WithMessage(xecode.NothingFound, "微信小程序涉政稿件下架")
	}
	view, originalArc, err := s.View(c, aid, 0, mid, cdnIP, outRefer, buvid)
	if err != nil {
		log.Error("Detail s.View(%d) error %+v", aid, err)
		return nil, s.slbRetryCode(err)
	}
	res := &model.Detail{View: view, HotShare: &model.HotShare{Show: false, List: _emptyBvArc}}
	group := errgroup.WithContext(c)
	if view.Author.Mid > 0 {
		group.Go(func(ctx context.Context) error {
			var cardErr error
			if res.Card, cardErr = s.Card(c, view.Author.Mid, mid, true, false); cardErr != nil {
				log.Error("Detail s.Card(%d) error %+v", aid, cardErr)
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		cid := view.FirstCid
		index := pageNo - 1
		if view.Pages != nil && index >= 0 && len(view.Pages) > index && view.Pages[index] != nil && view.Pages[index].Page == int32(pageNo) {
			cid = view.Pages[index].Cid
		}
		res.Tags, _ = s.DetailTag(ctx, aid, mid, cid, originalArc, false)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		res.Reply, _ = s.replyHot(ctx, aid)
		return nil
	})
	var viewForbidReco *videogrpc.ForbidReco
	group.Go(func(ctx context.Context) error {
		var relatedErr error
		if res.Related, res.Spec, viewForbidReco, relatedErr = s.RelatedArcs(ctx, aid, mid, buvid, needRcmdReason, needOperation, webRmRepeat, false, originalArc); relatedErr != nil {
			log.Error("s.RelatedArcs(%d) error %+v", aid, relatedErr)
		}
		if viewForbidReco != nil {
			res.ViewAddit = &model.ViewAddit{
				NoRecommendLive: func() bool {
					if state, ok := viewForbidReco.GetGroupState()[_noRecommendLive]; ok && state == 1 {
						return true
					}
					return false
				}(),
				NoRecommendActivity: func() bool {
					if state, ok := viewForbidReco.GetGroupState()[_noRecommendActivity]; ok && state == 1 {
						return true
					}
					return false
				}(),
			}
		}
		return nil
	})
	var shareArcs []*model.BvArc
	if needHotShare && s.c.Rule.HotShareOpen {
		group.Go(func(ctx context.Context) error {
			shareReply, shareErr := s.shareAdminGRPC.TopShareList(ctx, &shareadmin.TopShareListReq{Aid: aid})
			if shareErr != nil {
				log.Error("Detail s.shareAdminGRPC.TopShareList aid:%d error:%v", aid, shareErr)
				return nil
			}
			if shareReply == nil || len(shareReply.Aids) == 0 {
				log.Error("Detail s.shareAdminGRPC.TopShareList aid:%d reply empty", aid)
				return nil
			}
			var aids []int64
			for _, v := range shareReply.Aids {
				if len(aids) >= s.c.Rule.HotShare {
					break
				}
				if v > 0 && v != aid {
					aids = append(aids, v)
				}
			}
			if len(aids) == 0 {
				return nil
			}
			archives, cfcInfos, arcErr := s.batchArchivesAndCfcInfos(ctx, aids)
			if arcErr != nil {
				log.Error("Detail s.batchArchivesAndCfcInfos aids:%+v error:%v", aids, arcErr)
				return nil
			}
			for _, v := range aids {
				if len(shareArcs) >= s.c.Rule.HotShareOut {
					break
				}
				arcItem, ok := archives[v]
				if !ok || arcItem == nil || !arcItem.IsNormal() {
					continue
				}
				var cfcItem []*cfcgrpc.ForbiddenItem
				if cfcItem, ok = cfcInfos[v]; !ok {
					log.Warn("s.Detail forbidden is empty aid:%d", v)
				}
				arcForbidden := model.ItemToArcForbidden(cfcItem)
				if arcForbidden.NoDynamic ||
					arcForbidden.NoRecommend ||
					arcForbidden.NoRank {
					continue
				}
				shareArcs = append(shareArcs, model.CopyFromArcToBvArc(arcItem, s.avToBv(arcItem.Aid)))
			}
			return nil
		})
	}
	if needElec {
		group.Go(func(ctx context.Context) error {
			var err error
			if res.Elec, err = s.ElecShow(ctx, view.Author.Mid, aid, mid, originalArc); err != nil {
				log.Error("Detail s.ElecShow upMid:%d aid:%d error:%v", view.Author.Mid, aid, err)
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		var err error
		if res.Recommend, err = s.recommend(ctx, aid, view.Author.Mid, recommendType); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		redirectsRes, err := s.dao.ArcRedirectUrls(ctx, []int64{aid})
		if err != nil {
			log.Error("ArcRedirectUrls is err (%+v)", err)
			return nil
		}
		redirect, ok := redirectsRes[aid]
		if !ok {
			return nil
		}
		if redirect.RedirectTarget == "" || redirect.PolicyId == 0 {
			return nil
		}
		//location策略获取返回数据
		if redirect.GetPolicyType() == arcmdl.RedirectPolicyType_PolicyTypeLocation {
			locs, err := s.GetGroups(ctx, []int64{redirect.PolicyId}, cdnIP)
			if err != nil {
				log.Error("GetGroups is err (%+v)", err)
				return nil
			}
			loc, ok := locs[redirect.PolicyId]
			if !ok {
				return nil
			}
			//是否需要跳转
			if loc.Play != int64(locgrpc.Status_Forbidden) {
				res.View.RedirectURL = redirect.RedirectTarget
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("Detail aid:%d,mid:%d,error:%+v", aid, mid, err)
	}
	if len(res.Related) > 0 && len(shareArcs) > 0 {
		res.HotShare = &model.HotShare{Show: true, List: shareArcs}
	}
	if len(res.Related) == 0 {
		res.Recommend = nil
	}

	if platform == "wechat" {
		s.detailFilterBindOid(c, &res.Related)
		if res.Recommend != nil && res.Recommend.List != nil {
			s.detailFilterBindOid(c, &res.Recommend.List)
		}
	}
	// 上报
	s.reportWatch(c, originalArc, buvid)
	return res, nil
}

func (s *Service) detailFilterBindOid(c context.Context, containOidSlice *[]*model.BvArc) {
	if len(*containOidSlice) == 0 {
		return
	}
	var oidList []int64
	for _, v := range *containOidSlice {
		if v != nil {
			oidList = append(oidList, v.Aid)
		}
	}

	bindOidList, err := s.dao.TagBind(c, oidList)
	k := 0
	for _, v := range *containOidSlice {
		if err != nil || bindOidList == nil || v == nil || !inIntSlice(bindOidList, v.Aid) {
			(*containOidSlice)[k] = v
			k++
		}
	}
	*containOidSlice = (*containOidSlice)[:k]
}

func (s *Service) GetGroups(c context.Context, groupId []int64, cdnIp string) (map[int64]*locgrpc.Auth, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	req := &locgrpc.GroupsReq{
		Gids:    groupId,
		IpAddr:  ip,
		CdnAddr: cdnIp,
	}
	reply, err := s.locGRPC.Groups(c, req)
	if err != nil {
		return nil, err
	}
	return reply.GetAuths(), nil
}

// nolint:gocognit
func (s *Service) recommend(ctx context.Context, aid, mid int64, recommendType string) (*model.Recommend, error) {
	res := &model.Recommend{}
	var (
		aids      []int64
		arcBySort bool
	)
	switch recommendType {
	case "hot_share":
		res.Title = "热门分享"
		reply, err := s.shareAdminGRPC.TopShareList(ctx, &shareadmin.TopShareListReq{Aid: aid})
		if err != nil {
			return nil, err
		}
		if reply == nil || len(reply.Aids) == 0 {
			return nil, nil
		}
		for _, v := range reply.Aids {
			if len(aids) >= s.c.Rule.Recommend.AidCount {
				break
			}
			if v > 0 && v != aid {
				aids = append(aids, v)
			}
		}
	case "popular":
		res.Title = "B站热门"
		reply := s.wxHotAids
		for _, v := range reply {
			if len(aids) >= s.c.Rule.Recommend.AidCount {
				break
			}
			if v.ID > 0 && v.ID != aid {
				aids = append(aids, v.ID)
			}
		}
	case "arc_by_pubtime", "arc_by_view":
		arcBySort = true
		res.Title = "UP主投稿"
		var order uparcgrpc.SearchOrder
		if recommendType == "arc_by_view" {
			order = uparcgrpc.SearchOrder_click
		}
		reply, err := s.upArcGRPC.ArcPassed(ctx, &uparcgrpc.ArcPassedReq{Mid: mid, Pn: 1, Ps: 50, Order: order})
		if err != nil {
			return nil, err
		}
		for _, v := range reply.Archives {
			if v.Aid > 0 && v.Aid != aid {
				aids = append(aids, v.Aid)
			}
		}
	default:
		return nil, nil
	}
	if len(aids) == 0 {
		return nil, nil
	}
	archives, cfcInfos, err := s.batchArchivesAndCfcInfos(ctx, aids)
	if err != nil {
		return nil, err
	}
	var arcs []*model.BvArc
	for _, v := range aids {
		if len(arcs) == s.c.Rule.Recommend.ItemCount {
			break
		}
		arcItem, ok := archives[v]
		if !ok || arcItem == nil || !arcItem.IsNormal() {
			continue
		}
		var cfcItem []*cfcgrpc.ForbiddenItem
		if cfcItem, ok = cfcInfos[v]; !ok {
			log.Warn("s.recommend forbidden is empty aid:%d", v)
		}
		arcForbidden := model.ItemToArcForbidden(cfcItem)
		if !arcBySort && (arcForbidden.NoDynamic ||
			arcForbidden.NoRecommend ||
			arcForbidden.NoRank) {
			continue
		}
		arcs = append(arcs, model.CopyFromArcToBvArc(arcItem, s.avToBv(arcItem.Aid)))
	}
	if len(aids) < s.c.Rule.Recommend.ItemCount {
		return nil, nil
	}
	res.Show = true
	res.List = arcs
	return res, nil
}

// nolint:gocognit
func (s *Service) DetailTag(c context.Context, aid, mid, cid int64, arc *arcmdl.Arc, IsH5Subdivision bool) ([]*chmdl.VideoTag, error) {
	var (
		tags     []*chmdl.VideoTag
		tagsTmp1 []*chmdl.VideoTag
		tagsTmp2 []*chmdl.VideoTag
		tagsBgm  []*chmdl.VideoTag
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		bgmRep, err := s.dao.BgmEntrance(ctx, aid, cid)
		_bgmExist := 1
		if err != nil {
			log.Error("s.dao.BgmEntrance aid (%d) error (%+v)", aid, err)
			return err
		}
		if bgmRep.State == _bgmExist && bgmRep.Info != nil {
			tagsBgm = append(tagsBgm, &chmdl.VideoTag{
				TagTopTag: chmdl.TagTopTag{
					ID:      0,
					Name:    bgmRep.Info.MusicTitle,
					MusicId: bgmRep.Info.MusicId,
				},
				TagType: "bgm",
				JumpUrl: bgmRep.Info.JumpUrl,
			})
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		args := &topicsvc.BatchResTopicByTypeReq{
			Type:   1, //资源类型 1:视频(ugc)
			ResIds: []int64{aid},
		}
		reply, err := s.dao.BatchResTopicByType(ctx, args)
		if err != nil {
			log.Error("s.dao.BatchResTopicByType args=%+v err=%+v", args, err)
			return nil
		}
		if topicTag, ok := reply.ResTopics[aid]; ok {
			tagsTmp1 = append(tagsTmp1, &chmdl.VideoTag{
				TagTopTag: chmdl.TagTopTag{
					ID:   topicTag.Id,
					Name: topicTag.Name,
				},
				TagType: "topic",
				JumpUrl: topicTag.JumpUrl,
			})
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		resourceChannels, err := s.dao.ResourceChannels(c, mid, aid, chmdl.VideoChannel)
		if err != nil {
			log.Error("s.dao.ResourceChannels(%+v,%+v) (%+v)", mid, aid, err)
			return err
		}
		for _, channel := range resourceChannels.GetChannels() {
			tag := &chmdl.VideoTag{
				TagTopTag: chmdl.TagTopTag{
					ID:           channel.GetID(),
					Name:         channel.GetName(),
					Cover:        channel.GetCover(),
					HeadCover:    channel.GetHeadCover(),
					Content:      channel.GetContent(),
					ShortContent: channel.GetShortContent(),
					Type:         int8(channel.GetType()),
					State:        int8(channel.GetState()),
					CTime:        channel.GetCTime(),
					MTime:        channel.GetMTime(),
					Count: struct {
						View  int `json:"view"`
						Use   int `json:"use"`
						Atten int `json:"atten"`
					}{
						View:  0,
						Use:   int(channel.GetUse()),
						Atten: int(channel.GetAtten()), //关注数
					},
					IsAtten:   0,
					Role:      int8(channel.GetRole()),
					Likes:     channel.GetLikes(),
					Hates:     channel.GetHates(),
					Attribute: int8(channel.GetAttribute()),
					Liked:     int8(channel.GetLiked()),
					Hated:     int8(channel.GetHated()),
				},
				IsActivity:      false,
				Color:           channel.GetColor(),
				Alpha:           channel.GetAlpha(),
				IsSeason:        channel.GetPGC(),
				FeaturedCount:   channel.GetFeaturedCnt(),
				SubscribedCount: channel.GetAtten(),
				ArchiveCount:    appchanmdl.Stat64String(channel.GetResourceCnt(), ""),
			}
			channelType := int32(2) // 频道类型
			if channel.GetCType() == channelType {
				tag.TagType = "new_channel"
			} else {
				tag.TagType = "old_channel"
			}
			if channel.GetSubscribe() > 0 {
				tag.IsAtten = 1
			}
			tagsTmp2 = append(tagsTmp2, tag)
		}
		return nil
	})
	if arc == nil {
		group.Go(func(ctx context.Context) error {
			arcReq := &arcmdl.ArcRequest{Aid: aid}
			arcReply, err := s.arcGRPC.Arc(ctx, arcReq)
			if err != nil {
				log.Error("s.arcGRPC.Arc(%+v) (%+v)", arcReq, err)
				return err
			}
			arc = arcReply.GetArc()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("group.Wait() (%+v)", err)
		return nil, err
	}
	tags = append(tagsBgm, tagsTmp1...) // bgm tag first
	tags = append(tags, tagsTmp2...)
	if arc != nil && arc.AttrValV2(arcmdl.AttrBitV2CleanMode) == arcmdl.AttrYes {
		return []*chmdl.VideoTag{}, nil
	}
	if arc != nil && arc.MissionID > 0 {
		protocol, err := s.dao.ActProtocol(c, arc.MissionID)
		if err != nil {
			log.Error("s.dao.ActProtocol(%+v) (%+v)", arc.MissionID, err)
		} else if protocol.Protocol != nil {
			var (
				actTagName = protocol.Protocol.Tags
				actTags    []*chmdl.VideoTag
				newTags    []*chmdl.VideoTag
				oldTags    []*chmdl.VideoTag
				topicTags  []*chmdl.VideoTag
				bgmTags    []*chmdl.VideoTag
			)
			for _, tag := range tags {
				if tag.TagType == "topic" {
					topicTags = append(topicTags, tag)
					continue
				}
				if tag.TagType == "bgm" {
					bgmTags = append(bgmTags, tag)
					continue
				}
				if tag.Name == actTagName {
					tag.IsActivity = true
					tag.TagType = "activity"
					actTags = append(actTags, tag)
				} else if tag.TagType == "new_channel" {
					newTags = append(newTags, tag)
				} else {
					oldTags = append(oldTags, tag)
				}
			}
			// tag顺序：bgm-新话题-频道-活动-话题
			tags = append(bgmTags, topicTags...)
			tags = append(tags, newTags...)
			if len(topicTags) == 0 {
				// 如果没有新话题，再添加活动话题
				tags = append(tags, actTags...)
			}
			tags = append(tags, oldTags...)
		}
	}
	tagNames := s.c.Rule.H5SubdivisionTags
	if IsH5Subdivision && len(tagNames) > 0 {
		// H5唤端实验,仅用于在H5细分tag补充上无效tag标签，用于展示及后续配置改动方便，无实际意义，pc上的tag不用关注此项
		var customsTags []*chmdl.VideoTag
		for index, tagName := range tagNames {
			tag := &chmdl.VideoTag{
				TagTopTag: chmdl.TagTopTag{
					ID:   int64(index),
					Name: tagName,
				},
				TagType: "custom",
			}
			customsTags = append(customsTags, tag)
		}
		tags = append(customsTags, tags...)
	}
	return removeDuplicateTags(tags), nil
}

func removeDuplicateTags(in []*chmdl.VideoTag) []*chmdl.VideoTag {
	videoTagsSet := sets.NewString()
	res := make([]*chmdl.VideoTag, 0, len(in))
	for _, tag := range in {
		if videoTagsSet.Has(tag.Name) {
			continue
		}
		videoTagsSet.Insert(tag.Name)
		res = append(res, tag)
	}
	return res
}

func (s *Service) DetailGRPC(c context.Context, req *v1.ViewDetailReq, mid int64, buvid string) (data *v1.ViewDetailReply, err error) {
	if req.Bvid != "" {
		if req.Aid, err = bvid.BvToAv(req.Bvid); err != nil {
			log.Error("DetailGRPC bvid.BvToAv(%s) error(%v)", req.Bvid, err)
			err = xecode.RequestErr
			return
		}
	}
	var rs *model.Detail
	if rs, err = s.Detail(c, req.Aid, mid, "", "", buvid, "", "", false, false, false, false, false, 0); err != nil {
		return
	}
	data = model.CopyFromDetail(rs)
	return
}

// ArcUGCPay get arc ugc pay relation.
func (s *Service) ArcUGCPay(c context.Context, mid, aid int64) (data *model.AssetRelation, err error) {
	var relation *ugcmdl.AssetRelationResp
	data = new(model.AssetRelation)
	if arcReply, e := s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid}); e != nil {
		log.Error("s.arcGRPC.Arc(%d) error(%v)", aid, e)
	} else if arcReply.Arc.Author.Mid == mid {
		data.State = _ugcPaidState
		return
	}
	if relation, err = s.ugcPayGRPC.AssetRelation(c, &ugcmdl.AssetRelationReq{Mid: mid, Oid: aid, Otype: _ugcOtypeArc}); err != nil {
		log.Error("ArcUGCPay s.ugcPayGRPC.AssetRelation mid:%d aid:%d error(%v)", mid, aid, err)
		err = nil
		return
	}
	if relation.State == _ugcAssetPaid {
		data.State = _ugcPaidState
	}
	return
}

// ArcRelation .
func (s *Service) ArcRelation(c context.Context, mid, aid int64) (*model.ReqUser, error) {
	data := new(model.ReqUser)
	arcReply, err := s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid})
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	if arcReply == nil || arcReply.Arc == nil || !arcReply.Arc.IsNormal() {
		log.Error("ArcRelation s.arcGRPC.Arc(%d) error(%v)", aid, err)
		return data, nil
	}
	return s.arcRelation(c, mid, aid, arcReply.GetArc().GetSeasonID(), arcReply.Arc.Author.Mid), nil
}

func (s *Service) replyHot(c context.Context, aid int64) (res *model.ReplyHot, err error) {
	if res, err = s.dao.Hot(c, aid); err != nil {
		log.Error("s.dao.Hot(%d) error %+v", aid, err)
	}
	if res == nil {
		res = _emptyReplyHot
	}
	return
}

func (s *Service) ugcPayAsset(c context.Context, aid int64) (data *ugcmdl.AssetQueryResp) {
	asset, err := s.ugcPayGRPC.AssetQuery(c, &ugcmdl.AssetQueryReq{Oid: aid, Otype: _ugcOtypeArc, Currency: _ugcCurrencyBp})
	if err != nil {
		log.Error("ugcPayAsset oid(%d) error(%v)", aid, err)
		data = new(ugcmdl.AssetQueryResp)
		return
	}
	data = asset
	return
}

func (s *Service) staffInfo(c context.Context, authorMid int64, staffs []*arcmdl.StaffInfo) (data []*model.Staff) {
	var (
		mids   []int64
		cards  map[int64]*accmdl.Card
		stats  map[int64]*relamdl.StatReply
		owners []*arcmdl.StaffInfo
		err    error
	)
	owners = append(owners, &arcmdl.StaffInfo{Mid: authorMid, Title: "UP主"})
	owners = append(owners, staffs...)
	for _, v := range owners {
		if v.Mid > 0 {
			mids = append(mids, v.Mid)
		}
	}
	if len(mids) == 0 {
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		cardReply, e := s.accGRPC.Cards3(c, &accmdl.MidsReq{Mids: mids})
		if e != nil {
			log.Error("staffInfo s.accGRPC.Cards3(%v) error(%d)", mids, e)
			return e
		}
		cards = cardReply.Cards
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if statsReply, e := s.relationGRPC.Stats(c, &relamdl.MidsReq{Mids: mids}); e != nil {
			log.Error("staffInfo s.relationGRPC.Stats(%v) error(%v)", mids, e)
		} else {
			stats = statsReply.StatReplyMap
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("staffInfo group.Wait error(%v)", err)
		return
	}
	for _, owner := range owners {
		if card, ok := cards[owner.Mid]; ok && card != nil {
			item := &model.Staff{
				Mid:      owner.Mid,
				Title:    owner.Title,
				Name:     card.Name,
				Face:     card.Face,
				Official: card.Official,
			}
			item.Vip = card.Vip
			if stat, ok := stats[owner.Mid]; ok && stat != nil {
				item.Follower = stat.Follower
			}
			if owner.StaffAttrVal(arcmdl.StaffAttrBitAdOrder) == arcmdl.AttrYes {
				item.LabelStyle = model.StaffLabelAd
			}
			data = append(data, item)
		}
	}
	return
}

func (s *Service) loadManager() {
	if s.managerRunning {
		return
	}
	s.managerRunning = true
	defer func() {
		s.managerRunning = false
	}()
	groupReply, err := s.upGRPC.UpGroupMids(context.Background(), &upmdl.UpGroupMidsReq{GroupID: _testerGroup, Pn: _firstPn, Ps: s.c.Rule.UpGroupCnt})
	if err != nil || groupReply == nil {
		log.Error("loadManager s.upGRPC.UpGroupMids error(%+v)", err)
		return
	}
	midsM := make(map[int64]struct{}, len(groupReply.Mids))
	for _, mid := range groupReply.Mids {
		midsM[mid] = struct{}{}
	}
	log.Info("load special mids(%+v)  group reply count(%d)", midsM, groupReply.Total)
	s.specialMids = midsM
}

func (s *Service) initLabel(view *model.View) {
	if _, ok := s.hotLabelAids[view.Aid]; ok {
		view.Label = &model.Label{Type: _labelTypeHot}
	}
}

func (s *Service) loadHotLabel() {
	if s.hotLabelRunning {
		return
	}
	s.hotLabelRunning = true
	defer func() {
		s.hotLabelRunning = false
	}()
	hot, err := s.dao.HotLabel(context.Background())
	if err != nil {
		log.Warn("loadHotLabel s.dao.HotLabel warn(%v)", err)
		return
	}
	tmpLabel := make(map[int64]struct{}, len(hot))
	for _, aid := range hot {
		tmpLabel[aid] = struct{}{}
	}
	s.hotLabelAids = tmpLabel
}

func (s *Service) loadGuideCid() {
	if s.guideCidRunning {
		return
	}
	s.guideCidRunning = true
	defer func() {
		s.guideCidRunning = false
	}()
	if s.c.Rule.SteinsGuideAid == 0 {
		return
	}
	view, err := s.arcGRPC.View(context.Background(), &arcmdl.ViewRequest{Aid: s.c.Rule.SteinsGuideAid})
	if err != nil {
		log.Error("loadGuideCid s.arcGRPC.View(%d) error(%v)", s.c.Rule.SteinsGuideAid, err)
		return
	}
	if view != nil && view.Arc != nil && view.Arc.IsNormal() {
		if len(view.Pages) > 0 {
			atomic.StoreInt64(&s.steinGuideCid, view.Pages[0].Cid)
			log.Warn("loadGuideCid success cids(%d)", view.Pages[0].Cid)
		}
	}
}

// ArcCustomConfig is
func (s *Service) ArcCustomConfig(ctx context.Context, aid int64) (*model.ArcCustomConfig, error) {
	cc, err := s.resgrpc.CustomConfig(ctx, &resmdl.CustomConfigRequest{
		TP:  _ccTPArchive,
		Oid: aid,
	})
	if err != nil || cc == nil {
		log.Error("Failed to get archvie custom config: %d,err: %+v or cc=nil", aid, err)
		return nil, xecode.NothingFound
	}
	//仅链接，无高亮文字，默认展示"点击前往"；仅高亮文字，无链接，则不展示高亮文字
	highlight := cc.HighlightContent
	if cc.URL != "" && highlight == "" {
		highlight = "点击前往"
	} else if cc.URL == "" {
		highlight = ""
	}
	reply := &model.ArcCustomConfig{
		Aid:       aid,
		Content:   cc.Content,
		URL:       cc.URL,
		Highlight: highlight,
		Image:     cc.Image,
		ImageBig:  cc.ImageBig,
	}
	return reply, nil
}

func (s *Service) arcRelation(ctx context.Context, mid, aid, sid, authorMid int64) *model.ReqUser {
	ip := metadata.String(ctx, metadata.RemoteIP)
	group := errgroup.WithContext(ctx)
	data := new(model.ReqUser)
	// attention
	if authorMid > 0 {
		group.Go(func(ctx context.Context) error {
			if resp, e := s.accGRPC.Relation3(ctx, &accmdl.RelationReq{Mid: mid, Owner: authorMid, RealIp: ip}); e != nil {
				log.Error("arcRelation s.accGRPC.Relation3(%d,%d,%s) error(%v)", mid, authorMid, ip, e)
			} else if resp != nil {
				data.Attention = resp.Following
			}
			return nil
		})
	}
	// favorite
	group.Go(func(ctx context.Context) error {
		resourceMap := make(map[int32]*favgrpc.Oids)
		resourceMap[_favTypeArc] = &favgrpc.Oids{Oid: []int64{aid}}
		if sid > 0 {
			resourceMap[_favOTypeSeason] = &favgrpc.Oids{Oid: []int64{sid}}
		}
		req := &favgrpc.IsFavoredsResourcesReq{Mid: mid, ResourcesMap: resourceMap}
		reply, err := s.favGRPC.IsFavoredsResources(ctx, req)
		if err != nil {
			log.Error("arcRelation s.favGRPC.IsFavoredsResources err(%+v) req(%+v)", err, req)
			return nil
		}
		if favArcMap, ok := reply.GetFavored()[_favTypeArc]; ok {
			if fv, exist := favArcMap.GetOidFavored()[aid]; exist {
				data.Favorite = fv
			}
		}
		if favSeasonMap, ok := reply.GetFavored()[_favOTypeSeason]; ok {
			if fs, exist := favSeasonMap.GetOidFavored()[sid]; exist {
				data.SeasonFav = fs
			}
		}
		return nil
	})
	// like
	group.Go(func(ctx context.Context) error {
		if resp, e := s.thumbupGRPC.HasLike(ctx, &thumbmdl.HasLikeReq{Business: _businessLike, MessageIds: []int64{aid}, Mid: mid, IP: ip}); e != nil {
			log.Error("arcRelation s.thumbup.HasLike(%d,%d,%s) error %v", aid, mid, ip, e)
		} else if resp != nil && resp.States != nil {
			if v, ok := resp.States[aid]; ok {
				switch v.State {
				case thumbmdl.State_STATE_LIKE:
					data.Like = true
				case thumbmdl.State_STATE_DISLIKE:
					data.Dislike = true
				default:
				}
			}
		}
		return nil
	})
	// coin
	group.Go(func(ctx context.Context) error {
		if resp, e := s.coinGRPC.ItemUserCoins(ctx, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness}); e != nil {
			log.Error("arcRelation s.coinGRPC.ItemUserCoins(%d,%d,%s) error %v", mid, aid, ip, e)
		} else if resp != nil {
			data.Coin = resp.Number
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("arcRelation aid:%d group.Wait:%+v", aid, err)
	}
	return data
}

func (s *Service) slbRetryCode(originErr error) error {
	retryCode := []int{-500, -502, -504}
	for _, val := range retryCode {
		if xecode.EqualError(xecode.Int(val), originErr) {
			return errors.Wrapf(gateecode.WebSLBRetry, "%v", originErr)
		}
	}
	return originErr
}

func (s *Service) SLBRetry(err error) bool {
	return xecode.EqualError(gateecode.WebSLBRetry, err)
}

func (s *Service) basisSeasonABTest(buvid string) bool {
	group := crc32.ChecksumIEEE([]byte(buvid)) % s.c.BasisSeasonABTest.Group
	return group < s.c.BasisSeasonABTest.Gray
}

func (s *Service) batchArchivesAndCfcInfos(ctx context.Context, aids []int64) (map[int64]*arcmdl.Arc, map[int64][]*cfcgrpc.ForbiddenItem, error) {
	var (
		archives map[int64]*arcmdl.Arc
		cfcInfos map[int64][]*cfcgrpc.ForbiddenItem
	)
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		var err error
		archives, err = s.batchArchives(ctx, aids)
		if err != nil {
			log.Error("s.batchArchivesAndCfcInfos archives aids:%+v error:%v", aids, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var err error
		cfcInfos, err = s.batchCfcInfos(ctx, aids)
		if err != nil {
			// 打印日志,不用往上层抛出错误
			log.Error("s.batchArchivesAndCfcInfos cfcInfos aids:%+v error:%v", aids, err)
			return nil
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("s.batchArchivesAndCfcInfos aids:%+v error:%v", aids, err)
		return nil, nil, err
	}
	return archives, cfcInfos, nil
}

func (s *Service) batchCfcInfos(ctx context.Context, aids []int64) (map[int64][]*cfcgrpc.ForbiddenItem, error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	cfcInfos := make(map[int64][]*cfcgrpc.ForbiddenItem, aidsLen)
	group := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _cfcAttributeLimit {
		var partAids []int64
		l := i + _cfcAttributeLimit
		if l > aidsLen {
			l = aidsLen
		}
		partAids = aids[i:l]
		group.Go(func(ctx context.Context) error {
			var reply map[int64]*cfcgrpc.FlowCtlInfoReply
			if err := retry(func() (err error) {
				reply, err = s.contentFlowControlInfos(ctx, partAids)
				return err
			}); err != nil {
				log.Error("日志告警 contentFlowControlInfos error:%v", err)
				return errors.Wrapf(err, "ContentFlowControlInfos partAids:%v", partAids)
			}
			if reply == nil {
				return nil
			}
			mutex.Lock()
			for k, v := range reply {
				cfcInfos[k] = v.ForbiddenItems
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("s.batchArcAttribute error:%+v, arg:%+v", err, aids)
		return nil, err
	}
	return cfcInfos, nil
}

func (s *Service) contentFlowControlInfos(ctx context.Context, aids []int64) (map[int64]*cfcgrpc.FlowCtlInfoReply, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", s.c.CfcSvrConfig.Source)
	params.Set("business_id", strconv.FormatInt(s.c.CfcSvrConfig.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	req := &cfcgrpc.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(s.c.CfcSvrConfig.BusinessID),
		Source:     s.c.CfcSvrConfig.Source,
		Sign:       getSign(params, s.c.CfcSvrConfig.Secret),
		Ts:         ts,
	}
	reply, err := s.cfcGRPC.Infos(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, nil
	}
	return reply.ForbiddenItemMap, nil
}

func getSign(params url.Values, secret string) string {
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}

func (s *Service) checkCommonBWList(ctx context.Context, aid int64, platform string) bool {
	if aid == 0 || platform != _banPlatform {
		return false
	}
	rep, err := s.res2grpc.CheckCommonBWList(ctx, &resv2grpc.CheckCommonBWListReq{
		Oid:   strconv.FormatInt(aid, 10),
		Token: s.c.BanArcGRPCToken,
	})
	if err != nil {
		log.Error("s.checkCommonBWList aid:%d, platform:%s, token:%s, err:%v", aid, platform, s.c.BanArcGRPCToken, err)
		return false
	}
	return rep.IsInList
}

func (s *Service) Premiere(c context.Context, aid int64) (*model.Premiere, error) {
	arc, err := s.arcGRPC.SteinsGateView(c, &arcmdl.SteinsGateViewRequest{Aid: aid})
	if err != nil {
		log.Error("s.Premiere error:%+v, aid:%d", err, aid)
		return nil, err
	}
	if !arc.IsNormalPremiere() {
		// 非首映稿件
		return nil, nil
	}
	res := &model.Premiere{
		State:     int32(arc.GetPremiere().GetState()),
		StartTime: arc.GetPremiere().GetStartTime(),
		RoomID:    arc.GetPremiere().GetRoomId(),
		NowTime:   time.Now().Unix(),
	}
	reply, err := s.dao.GetPremiereSidByAid(c, aid)
	if err != nil {
		log.Error("s.Premiere error:%+v, aid:%d", err, aid)
		return nil, nil
	}
	if reply != nil {
		res.SID = reply.Sid
	}
	return res, nil
}

func (s *Service) PremiereInfo(c context.Context, aid int64) (*model.PremiereInfo, error) {
	req := &playeronlinegrpc.PremiereInfoReq{
		Aid: aid,
	}
	reply, err := s.onlineGRPC.PremiereInfo(c, req)
	if err != nil {
		log.Error("s.PremiereInfo err:%v, aid:%d", err, aid)
		return nil, err
	}
	res := &model.PremiereInfo{
		Participant: reply.GetParticipant(),
		Interaction: reply.GetInteraction(),
	}
	return res, nil
}

func (s *Service) reportWatch(ctx context.Context, a *arcmdl.Arc, buvid string) {
	if a == nil || a.Premiere == nil || a.Premiere.State == arcmdl.PremiereState_premiere_none ||
		a.Premiere.State == arcmdl.PremiereState_premiere_after {
		return
	}
	req := &playeronlinegrpc.ReportWatchReq{
		Aid:   a.GetAid(),
		Buvid: buvid,
		Biz:   "web",
	}
	if _, err := s.onlineGRPC.ReportWatch(ctx, req); err != nil {
		log.Error("s.reportWatch aid:%d, err:%v", a.GetAid(), err)
		return
	}
}
