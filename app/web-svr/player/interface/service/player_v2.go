package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/utils/collection"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/player/interface/model"

	"github.com/pkg/errors"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	memberapi "git.bilibili.co/bapis/bapis-go/account/service/member"
	ugcpayapi "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	assitapi "git.bilibili.co/bapis/bapis-go/assist/service"
	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/auth"
	ansapi "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	hisapi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	locapi "git.bilibili.co/bapis/bapis-go/community/service/location"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	_defaultDmMaxLimit       = 1500
	_firstPage         int32 = 1
	_defaultPermission       = "0"
	_blockForever            = -1
	// content-flow-control.service gRPC infos limit
	_cfcAttributeLimit = 30
)

// nolint:gocognit,gomnd
func (s *Service) PlayerV2(ctx context.Context, arg *model.PlayerV2Arg, mid int64) (*model.PlayerV2, error) {
	viewData, err := s.view(ctx, arg.Aid)
	if err != nil {
		return nil, err
	}
	if viewData == nil || viewData.Arc == nil || !viewData.Arc.IsNormal() {
		log.Warn("PlayerV2 aid(%d) vi nil or state not allow", arg.Aid)
		return nil, ecode.NothingFound
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	res := &model.PlayerV2{
		Aid:     arg.Aid,
		Bvid:    s.avToBv(arg.Aid),
		AllowBp: viewData.Arc.AttrVal(arcapi.AttrBitAllowBp) == arcapi.AttrYes,
		// NoShare:         viewData.AttrValV2(arcapi.AttrBitV2OnlyFavView) == arcapi.AttrYes,
		IsUgcPayPreview: viewData.Arc.AttrVal(arcapi.AttrBitUGCPay) == arcapi.AttrYes && viewData.Arc.AttrVal(arcapi.AttrBitUGCPayPreview) == arcapi.AttrYes,
		Cid:             arg.Cid,
		IPInfo:          &model.PlayerIPInfo{IP: ip, ZoneIP: arg.CdnIP},
		MaxLimit:        _defaultDmMaxLimit,
		PageNo:          _firstPage,
		Permission:      _defaultPermission,
		NowTime:         arg.Now.Unix(),
		ViewPoints:      []*model.Point{},
		OnlineSwitch:    s.paramsMap,
		PreviewToast:    _previewToast,
		ShowSwitch:      &model.ShowSwitch{},
	}
	// if viewData.GetDuration
	res.Options, _ = s.playerOptions(viewData.Arc)
	// find page
	if cuPage := func() *arcapi.Page {
		for _, page := range viewData.Pages {
			if page.GetCid() == arg.Cid {
				return page
			}
		}
		log.Warn("PlayerV2 aid(%d) cid(%d) refer(%s) page not found", arg.Aid, arg.Cid, arg.Refer)
		return nil
	}(); cuPage != nil {
		res.MaxLimit = dmLimit(cuPage.Duration)
		res.PageNo = cuPage.Page
		longProcessUGC := int64(time.Duration(s.c.LongProgress.UGC).Seconds())
		isUGC := viewData.Arc.AttrVal(arcapi.AttrBitIsPGC) == arcapi.AttrNo && viewData.Arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrNo
		if longProcessUGC > 0 && cuPage.GetDuration() >= longProcessUGC && isUGC {
			res.ShowSwitch.LongProgress = true
		}
	}
	for index, pa := range viewData.Pages {
		if pa != nil && arg.Cid == pa.Cid && index+1 < len(viewData.Pages) {
			res.HasNext = true
			break
		}
	}
	eg := errgroup.WithCancel(ctx)
	// ip
	eg.Go(func(ctx context.Context) error {
		ipInfo, ipErr := s.locGRPC.Info(ctx, &locapi.InfoReq{Addr: ip})
		if ipErr != nil {
			log.Error("PlayerV2 s.locGRPC.Info ip:%s error:%+v", ip, ipErr)
			return nil
		}
		if ipInfo != nil {
			res.IPInfo.ZoneID = ipInfo.ZoneId
			res.IPInfo.Country = ipInfo.Country
			res.IPInfo.Province = ipInfo.Province
			res.IPInfo.City = ipInfo.City
		}
		return nil
	})
	// 进度条icon
	eg.Go(func(ctx context.Context) error {
		res.PlayerIcon, _ = s.tagPlayerIcon(ctx, arg.Aid, arg.SeasonID, viewData.Arc.TypeID, mid)
		return nil
	})
	// 在线人数
	eg.Go(func(ctx context.Context) error {
		// 默认1人，未连接上广播服务器使用
		res.OnlineCount = 1
		onlineCount, onlineErr := s.dao.OnlineCount(ctx, arg.Aid, arg.Cid)
		if onlineErr != nil {
			log.Error("PlayerV2 s.dao.OnlineCount aid:%d cid:%d error:%v", arg.Aid, arg.Cid, onlineErr)
			return nil
		}
		if onlineCount > 1 {
			res.OnlineCount = onlineCount
		}
		return nil
	})
	// 蒙板
	eg.Go(func(ctx context.Context) error {
		res.DmMask, _ = s.dmMask(ctx, arg.Cid)
		return nil
	})
	// 字幕
	eg.Go(func(ctx context.Context) error {
		res.Subtitle, _ = s.dmSubtitle(ctx, arg.Aid, arg.Cid)
		return nil
	})
	// 高能看点和章节
	eg.Go(func(ctx context.Context) error {
		res.ViewPoints, _ = s.viewPoints(ctx, arg.Aid, arg.Cid)
		return nil
	})
	// 互动视频
	if viewData.Arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes {
		if _, ok := s.steinGuideCids[arg.Cid]; !ok {
			eg.Go(func(ctx context.Context) error {
				interaction, _, steinErr := s.interaction(ctx, arg.Aid, arg.Cid, mid, arg.GraphVersion, arg.Buvid)
				if steinErr != nil {
					log.Error("PlayerV2 s.interaction aid:%d cid:%d mid:%d) error(%v)", arg.Aid, arg.Cid, mid, steinErr)
					return steinErr
				}
				res.Interaction = interaction
				return nil
			})
		}
	}
	// pcdn
	eg.Go(func(ctx context.Context) error {
		res.PcdnLoader, _ = s.pcdnLoader(ctx, arg.Cid, mid, arg.Refer, arg.InnerSign)
		return nil
	})
	// 关注引导和跳转卡
	eg.Go(func(ctx context.Context) error {
		res.GuideAttention, res.JumpCard, res.OperationCard, _, _, _ = s.playerCards(ctx, arg.Aid, arg.Cid, arg.EpID, mid)
		return nil
	})
	// pugv
	if viewData.Arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrYes {
		eg.Go(func(ctx context.Context) error {
			pugvStatus, pugvErr := s.pugvGRPC.SeasonPlayStatus(ctx, &cheeseapi.SeasonPlayStatusReq{Aid: arg.Aid, Mid: mid})
			if pugvErr != nil {
				log.Error("PlayerV2 s.pugvGRPC.SeasonPlayStatus aid:%d,mid:%d error(%v)", arg.Aid, mid, pugvErr)
				return pugvErr
			}
			if pugvStatus != nil {
				res.Pugv = &model.PlayerPugv{
					WatchStatus:  pugvStatus.WatchStatus,
					PayStatus:    pugvStatus.PayStatus,
					SeasonStatus: pugvStatus.SeasonStatus,
				}
			}
			return nil
		})
	}
	// bgm
	eg.Go(func(ctx context.Context) error {
		bgmReply, bgmErr := s.dao.BgmEntrance(ctx, arg.Aid, arg.Cid)
		if bgmErr != nil {
			log.Error("PlayerV2 BgmEntrance aid:%d,cid:%d error(%v)", arg.Aid, arg.Cid, bgmErr)
			return bgmErr
		}
		res.BgmInfo = bgmReply.Info
		return nil
	})

	if mid > 0 {
		res.LoginMid = mid
		res.LoginMidHash = midCrc(mid)
		res.IsOwner = mid == viewData.Author.Mid
		func() {
			// 获取用户信息
			profileReply, profileErr := s.accGRPC.ProfileWithStat3(ctx, &accapi.MidReq{Mid: mid})
			if profileErr != nil {
				log.Error("PlayerV2 s.accGRPC.ProfileWithStat3 mid:%d error:%v", mid, profileErr)
				return
			}
			if profileReply == nil || profileReply.Profile == nil {
				log.Error("PlayerV2 s.accGRPC.ProfileWithStat3 mid:%d profile:%+v nil", mid, profileReply)
				return
			}
			res.Name = profileReply.Profile.Name
			res.LevelInfo = profileReply.LevelInfo
			res.Vip = profileReply.Profile.Vip
			if profileReply.Profile.Silence == _accBanNor || isAdmin(profileReply.Profile.Rank) {
				res.Permission = strings.Join([]string{strconv.FormatInt(int64(profileReply.Profile.Rank), 10), "1001"}, ",")
			}
			// 答题状态
			if profileReply.Profile.Rank < _member {
				res.AnswerStatus = _notAnswer
				eg.Go(func(ctx context.Context) error {
					answerReply, answerErr := s.ansGRPC.Status(ctx, &ansapi.StatusReq{Mid: mid})
					if answerErr != nil {
						log.Error("PlayerV2 s.ansRPC.Status(%d) error(%v)", mid, answerErr)
						return nil
					}
					if answerReply == nil || answerReply.Status == nil {
						log.Error("PlayerV2 s.ansRPC.Status mid:%d reply:%+v nil", mid, answerReply)
						return nil
					}
					res.AnswerStatus = answerReply.Status.Status
					return nil
				})
			}
			// 封禁时间
			if profileReply.Profile.Silence == _accBanSta {
				eg.Go(func(ctx context.Context) error {
					blockTime, blockErr := s.memberGRPC.BlockInfo(ctx, &memberapi.MemberMidReq{Mid: mid})
					if blockErr != nil {
						log.Error("PlayerV2 s.memberGRPC.BlockInfo mid:%d error(%v)", mid, blockErr)
						return nil
					}
					switch blockTime.GetBlockStatus() {
					case _accBlockLimit: //限时封禁
						res.BlockTime = blockTime.EndTime - res.NowTime
						if res.BlockTime < 0 {
							res.BlockTime = 0
						}
					case _accBlockAll: //永久封禁
						res.BlockTime = _blockForever
					default:
						res.BlockTime = 0
					}
					return nil
				})
			}
		}()
		// 协管
		if s.c.Rule.NoAssistMid != viewData.Arc.Author.Mid {
			eg.Go(func(ctx context.Context) error {
				assistReply, assistErr := s.assistGRPC.Assist(ctx, &assitapi.AssistReq{Mid: viewData.Arc.Author.Mid, AssistMid: mid, Tp: 2})
				if assistErr != nil {
					log.Error("PlayerV2 s.assistGRPC.Assist authorMid:%d mid:%d error(%v)", viewData.Arc.Author.Mid, mid, assistErr)
					return nil
				}
				if assistReply.Ar != nil {
					res.Role = strconv.FormatInt(assistReply.Ar.Assist, 10)
				}
				return nil
			})
		}
		// ugc付费
		if res.IsUgcPayPreview && mid == viewData.Arc.Author.Mid {
			res.IsUgcPayPreview = false
		}
		if res.IsUgcPayPreview {
			eg.Go(func(ctx context.Context) error {
				assetReply, assetErr := s.ugcPayGRPC.AssetRelation(ctx, &ugcpayapi.AssetRelationReq{Mid: mid, Oid: viewData.Arc.Aid, Otype: _ugcPayOtypeArc})
				if assetErr != nil {
					log.Error("PlayerV2 s.ugcPayGRPC.AssetRelation mid:%d aid:%d error(%+v)", mid, viewData.Arc.Aid, assetErr)
					return assetErr
				}
				if assetReply != nil && assetReply.State == _relationPaid {
					res.IsUgcPayPreview = false
				}
				return nil
			})
		}
		// 播放进度
		eg.Go(func(ctx context.Context) error {
			proReply, proErr := s.hisGRPC.Progress(ctx, &hisapi.ProgressReq{Mid: mid, Aids: []int64{viewData.Arc.Aid}})
			if proErr != nil || proReply == nil {
				log.Error("PlayerV2 s.hisGRPC.Progress mid:%d aid:%d error(%v)", mid, viewData.Arc.Aid, proErr)
				return nil
			}
			progress, ok := proReply.Res[viewData.Arc.Aid]
			if !ok || progress == nil || progress.Cid <= 0 {
				return nil
			}
			if progress.Pro >= 0 && progress.Cid == arg.Cid {
				res.LastPlayTime = 1000 * progress.Pro
				res.LastPlayCid = progress.Cid
				return nil
			}
			for _, page := range viewData.Pages {
				if page.GetCid() == progress.Cid {
					res.LastPlayTime = 1000 * progress.Pro
					res.LastPlayCid = progress.Cid
					break
				}
			}
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		log.Error("PlayerV2 eg.Wait() aid:%d cid:%d error:%v", arg.Aid, arg.Cid, err)
		return nil, err
	}

	// 禁止项接入
	cfcInfos, err := s.batchCfcInfos(ctx, []int64{viewData.Arc.Aid})
	if cfcInfos != nil && err == nil {
		noShare, ok := cfcInfos[viewData.Arc.Aid]
		if !ok {
			log.Warn("s.View forbidden is empty aid:%d", viewData.Arc.Aid)
		}
		res.NoShare = noShare
	}

	// 植入fawkes的config和ff数据
	if fawekesVersion, ok := s.fawkesVersionCache[arg.FawkesEnv]; ok {
		if fv, ok := fawekesVersion[arg.FawkesAppKey]; ok {
			res.Fawkes = &model.Fawkes{
				ConfigVersion: fv.Config,
				FFVersion:     fv.FF,
			}
		}
	}
	return res, nil
}

func (s *Service) batchCfcInfos(ctx context.Context, aids []int64) (map[int64]bool, error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	cfcInfos := make(map[int64]bool, aidsLen)
	group := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _cfcAttributeLimit {
		var partAids []int64
		l := i + _cfcAttributeLimit
		if l > aidsLen {
			l = aidsLen
		}
		partAids = aids[i:l]
		group.Go(func(ctx context.Context) error {
			var reply map[int64]bool
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
				cfcInfos[k] = v
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

// 目前仅会去匹配是否禁止分享【使用的match方法, 如果需要再接入其他禁止项，需要咨询91路人】
func (s *Service) contentFlowControlInfos(ctx context.Context, aids []int64) (map[int64]bool, error) {
	params := url.Values{}
	params.Set("source", s.c.CfcSvrConfig.Source)
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	req := &cfcgrpc.FlowCtlMatchReq{
		Oids:   aids,
		Source: s.c.CfcSvrConfig.Source,
	}
	reply, err := s.cfcGRPC.Match(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, nil
	}
	return reply.MatchInfo, nil
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}
