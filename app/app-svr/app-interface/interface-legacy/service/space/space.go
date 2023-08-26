package space

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup"
	"go-common/library/text/translate/chinese.v2"
	"go-common/library/utils/collection"

	errgroupv2 "go-common/library/sync/errgroup.v2"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/middleware/midInt64"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/audio"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/cm"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/comic"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/community"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/elec"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/favorite"
	gmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/game"
	mallmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/mall"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/platng"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/shop"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	spm "go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"
	spacegrpc "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/pkg/idsafe/bvid"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	cheeseGRPC "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	answergrpc "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	dynamicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	campusapi "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynsharegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/publish"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	actgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	passportuser "git.bilibili.co/bapis/bapis-go/passport/service/user"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcappcard "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"
	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"
	digitalgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/digital/service"
	live2dgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"
	vipinfogrpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

const (
	_shopName             = "的店铺"
	_businessLike         = "archive"
	_articleLike          = "article"
	_dynamicLike          = "dynamic"
	_albumLike            = "album"
	_clipLike             = "clip"
	_cheeseLike           = "cheese"
	_iphoneGarbBuild      = 9120
	_androidGarbBuild     = 5520400
	_iphoneShopRsc        = 3322
	_androidShopRsc       = 3327
	_spaceArchiveUri      = "bilibili://music/playlist/spacepage/%d?%s"
	_spaceArchivePS       = "20"
	_spaceArchiveOffset   = "0"
	_spaceArchiveOrder    = "time"
	_spaceArchiveDesc     = "1"
	_spaceArchivePageType = "1"
	_spaceArchiveOid      = "0"
	_accountBlocked       = 1
	_accountIsDeleted     = 1
	_spaceArchiveBusiness = "media-list"
	// 稿件属性位信息
	_arcAttrNoRecommend = "55"
	// top image
	_defaultTopImage        = "https://i0.hdslb.com/bfs/activity-plat/static/9bdd988aed64a23976d6d5494533a450/tDojHPi2LG.jpg"
	_defaultTopNightImage   = "https://i0.hdslb.com/bfs/activity-plat/static/9bdd988aed64a23976d6d5494533a450/5A10HwGd3.jpg"
	_spaceSeriesListMaxSize = 20
	// 原样式icon
	_prInfoOldIcon = "https://i0.hdslb.com/bfs/space/7a89f7ed04b98458b23863846bd2539a90ff1153.png"
	// 原样式夜间icon
	_prInfoOldIconNight = "https://i0.hdslb.com/bfs/space/cab669b46fc1bce8b8b2fbd0ce19909f9f2299a4.png"
	// 缅怀提示日间icon
	_prInfoNewIcon = "https://i0.hdslb.com/bfs/space/ca6d0ed2edae23cf348db19cd2c293f2121c1b59.png"
	// 缅怀提示夜间icon
	_prInfoNewIconNight = "https://i0.hdslb.com/bfs/space/e2a4c7bb9297e74d1be7467f96086bf33931f9d0.png"
	// 缅怀样式背景色
	_prInfoNewBgColor = "#e7e7e7"
	// 缅怀样式文字色
	_prInfoNewTextcolor = "#999999"
	// 缅怀样式夜间背景色
	_prInfoNewBgColorNight = "#2A2A2A"
	// 缅怀样式夜间文字色
	_prInfoNewTextcolorNight = "#727272"
	// 原样式背景色
	_prInfoOldBgColor = "#FFF3DB"
	// 原样式文字色
	_prInfoOldTextcolor = "#FFB112"
	// 原样式夜间背景色
	_prInfoOLdBgColorNight = "#322D21"
	// 原样式夜间文字色
	_prInfoOldTextcolorNight = "#E6A31D"
	// 花火入口icon
	_PickupEntranceIcon = "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/qGbJfq9VGe.png"
)

func (s *Service) garbInfo(ctx context.Context, garbMob *space.Mob, equip *garbgrpc.SpaceBGUserEquipReply, userAsset *garbgrpc.SpaceBGUserAssetListReply, vmid, mid int64) {
	var (
		fansNbr  int64
		showGarb bool
	)
	garbMob.GoodsAvailable = s.c.Cfg.GarbCfg.GoodsAvailable
	if equip != nil && equip.Item != nil {
		fansNbr, _ = s.garbDao.UserFanInfo(ctx, vmid, equip.Item.SuitItemID)
	}
	if userAsset != nil && len(userAsset.List) > 0 { // 是否购买过和当前出不出粉丝头图无关
		garbMob.HasGarb = true
	}
	if !garbMob.HasGarb { // 未买过
		garbMob.PurchaseButton = &space.GarbButton{
			URI:   s.c.Cfg.GarbCfg.PurchaseButtonURI,
			Title: s.c.Cfg.GarbCfg.PurchaseButtonTitle,
		}
	}
	garb := new(space.GarbInfo)
	if showGarb = garb.FromEquip(equip, fansNbr); showGarb {
		garbMob.GarbInfo = garb                                            // 粉丝标牌等信息
		garbMob.TopPhoto.ImgURL = equip.Item.Images[equip.Index].Landscape // 替换为粉丝头图
		garbMob.TopPhoto.NightImgURL = ""
	}
	if mid != 0 && mid == vmid && (showGarb || garbMob.TopPhoto.ImgURL != "") { // 主态并且不是默认头图才下发恢复默认按钮
		garbMob.ShowReset = true
	}
	//nolint:gosimple
	return
}

// Space aggregation space data.
//
//nolint:gocognit
func (s *Service) Space(c context.Context, mid, vmid int64, plat int8, build int, pn, ps, teenagersMode, lessonsMode int, fromViewAid int64, platform, device, mobiApp, name string, now time.Time, buvid, network, adExtra, spmid, fromSpmid, filtered string, isHant bool) (sp *space.Space, err error) {
	if _, ok := s.BlackList[vmid]; ok {
		err = ecode.NothingFound
		return
	}
	if ok := func() bool {
		if mid == vmid {
			return true
		}
		ok, err := s.resDao.CheckCommonBWList(c, vmid)
		if err != nil {
			log.Error("%+v", err)
			return true
		}
		return ok
	}(); !ok {
		return nil, ecode.NothingFound
	}
	sp = &space.Space{}
	// 获取空间基本信息
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		// 账号基本信息
		card, err := s.card(ctx, vmid, mid, name, isHant, int64(build), mobiApp, device)
		if err != nil {
			log.Error("s.card vmid(%d) error(%+v)", vmid, err)
			// 账号信息出错时，直接返回
			return ecode.NothingFound
		}
		if vmid < 1 {
			if vmid, _ = strconv.ParseInt(card.Mid, 10, 64); vmid < 1 {
				return ecode.NothingFound
			}
		}
		if model.IsAndroid(plat) && (build >= s.c.SpaceBuildLimit.HideSexAndroidStart && build <= s.c.SpaceBuildLimit.HideSexAndroidEnd) && s.c.SpaceBuildLimit.HideSexSwitch {
			card.Sex = ""
		}
		sp.Card = card
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取空间设置信息
		if sp.Setting, err = s.spcDao.Setting(ctx, vmid); err != nil {
			prom.BusinessInfoCount.Incr("SpSettingNil")
			log.Error("s.spcDao.Setting vmid(%d), err(%+v)", vmid, err)
			// 设置信息出错时，链路继续进行
			return nil
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("Space() eg.Wait() error(%+v)", err)
		return nil, err
	}
	if sp.Card.IsDeleted == _accountIsDeleted {
		log.Error("s.space account is deleted mid=%d vmid=%d card=%+v ", mid, vmid, sp.Card)
		return nil, errors.WithMessage(ecode.NothingFound, "账号已注销")
	}
	if _, ok := s.UpRcmdBlockMap[vmid]; ok {
		sp.DisableUpRcmd = true
	}
	if s.c.SpaceSkipRule.AchieveURL != "" {
		achieveURL := s.c.SpaceSkipRule.AchieveURL + "?navhide=1&mid=" + strconv.FormatInt(vmid, 10)
		if sp.Card.Nameplate.ImageSmall != "" {
			sp.Card.Achieve = &space.Achieve{
				Image:      sp.Card.Nameplate.ImageSmall,
				AchieveURL: achieveURL,
				IsDefault:  false}
		} else if vmid == mid {
			sp.Card.Achieve = &space.Achieve{
				Image:      s.c.SpaceSkipRule.AchieveImage,
				AchieveURL: achieveURL,
				IsDefault:  true,
			}
		}
	}
	// space
	var (
		ownerEquip     *garbgrpc.SpaceBGUserEquipReply
		userAssetReply *garbgrpc.SpaceBGUserAssetListReply
		supportGarb    = false
		hasHomeTab     bool // 是否展示主页tab,针对tab2生效
		characterReply *live2dgrpc.GetUserSpaceCharacterInfoResp
		digitalReply   *digitalgrpc.GetGarbSpaceEntryResp
	)
	g, ctx := errgroup.WithContext(c)
	if vmid == mid {
		sp.Card.PendantURL = s.c.SpaceSkipRule.PendantURL
		sp.Card.PendantTitle = "更换头像挂件"
	} else if sp.Card.Pendant.Pid > 0 { // 客态
		sp.Card.PendantURL = fmt.Sprintf("https://www.bilibili.com/h5/mall/preview/pendant/%d?navhide=1&from=personal_space", sp.Card.Pendant.Pid)
		sp.Card.PendantTitle = "查看TA的头像挂件"
		g.Go(func() error {
			assetRly, e := s.garbDao.UserAsset(ctx, int64(sp.Card.Pendant.Pid), vmid)
			if e != nil {
				log.Error("s.garbDao.UserAsset(%d,%d) error(%v)", sp.Card.Pendant.Pid, vmid, e)
				return nil
			}
			if assetRly != nil && assetRly.Asset != nil {
				// wangyuzhe(20200824)：IsDiy 和 SuitItemID 相比，IsDiy 优先
				if assetRly.Asset.Item != nil && assetRly.Asset.Item.SuitItemID > 0 {
					sp.Card.PendantURL = fmt.Sprintf("https://www.bilibili.com/h5/mall/suit/detail?navhide=1&id=%d&from=personal_space", assetRly.Asset.Item.SuitItemID)
				}
				if assetRly.Asset.IsDiy > 0 {
					sp.Card.PendantURL = fmt.Sprintf("https://www.bilibili.com/h5/mall/preview/pendant/%d?navhide=1&from=personal_space&isdiy=%d", sp.Card.Pendant.Pid, assetRly.Asset.IsDiy)
				}
			}
			return nil
		})
	}
	if sp.Card.Silence == _accountBlocked {
		g.Go(func() error {
			endTime, err := s.accDao.BlockTime(ctx, vmid)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			sp.Card.EndTime = endTime
			sp.Card.SilenceURL = "http://www.bilibili.com/blackroom/releaseexame.html"
			return nil
		})
	}
	// 活动接口获取up主预约卡
	g.Go(func() error {
		data, err := s.actDao.UpActUserSpaceCard(ctx, mid, vmid)
		if err != nil {
			log.Error("s.actDao.UpActUserSpaceCard mid(%d), vmid(%d), err(%+v)", mid, vmid, err)
			return nil
		}
		if len(data) <= 0 {
			return nil
		}
		sortReserveRelationInfo(data)
		simpleInfo := s.getDynSimpleInfo(ctx, makeDynIdsFromReserveInfo(data))
		sp.ReservationCardList = asSpaceReservationCardList(data, mid == vmid, simpleInfo, build, plat)
		if len(sp.ReservationCardList) > 0 {
			sp.ReservationCardInfo = sp.ReservationCardList[0]
		}
		return nil
	})
	// 公告信息会影响高仿号提示的展示
	g.Go(func() error {
		var err error
		if sp.Card.PRInfo, err = s.spcDao.PRInfo(ctx, vmid); err != nil {
			log.Error("%+v", err)
			return nil
		}
		if (model.IsIOSPick(plat) && build > s.c.SpaceBuildLimit.PrInfoCardIOS) ||
			(model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.PrInfoCardAndroid) ||
			(plat == model.PlatIpadHD && build > s.c.SpaceBuildLimit.PrInfoCardIPadHD) ||
			model.IsAndroidHD(plat) {
			if sp.Card.PRInfo == nil {
				log.Warn("sp.Card.PRInfo should not be nil")
				return nil
			}
			// 公告配置类型，1-其他类型，2-去世公告
			//nolint:gomnd
			if sp.Card.PRInfo.NoticeType == 1 {
				sp.Card.PRInfo.Icon = _prInfoOldIcon
				sp.Card.PRInfo.IconNight = _prInfoOldIconNight
				sp.Card.PRInfo.BgColor = _prInfoOldBgColor
				sp.Card.PRInfo.BgColorNight = _prInfoOLdBgColorNight
				sp.Card.PRInfo.TextColor = _prInfoOldTextcolor
				sp.Card.PRInfo.TextColorNight = _prInfoOldTextcolorNight
				//nolint:gomnd
			} else if sp.Card.PRInfo.NoticeType == 2 {
				sp.Card.PRInfo.Icon = _prInfoNewIcon
				sp.Card.PRInfo.IconNight = _prInfoNewIconNight
				sp.Card.PRInfo.BgColor = _prInfoNewBgColor
				sp.Card.PRInfo.BgColorNight = _prInfoNewBgColorNight
				sp.Card.PRInfo.TextColor = _prInfoNewTextcolor
				sp.Card.PRInfo.TextColorNight = _prInfoNewTextcolorNight
			}
		} else {
			//nolint:gomnd
			if sp.Card.PRInfo.NoticeType == 2 {
				sp.Card.PRInfo = nil
			}
		}
		return nil
	})
	g.Go(func() (err error) {
		sp.Space = new(space.Mob)
		data, topPhotoArc, err := s.spcDao.TopPhoto(ctx, mobiApp, device, build, vmid, mid)
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		if data != nil {
			sp.Space.TopPhoto.ImgURL = data.ImgUrl
			sp.Space.TopPhoto.NightImgURL = data.NightImgUrl
		}
		if topPhotoArc != nil {
			// 主人态展示入口
			if topPhotoArc.Show && mid == vmid {
				sp.Space.ShowSetArchive = true
				sp.Space.SetArchiveText = s.c.Custom.SetArchiveText
			}
			if topPhotoArc.Aid > 0 {
				func() {
					apm, err := s.arcDao.ArcsPlayer(ctx, []*api.PlayAv{{Aid: topPhotoArc.Aid}}, false)
					if err != nil {
						log.Error("%+v", err)
						//nolint:ineffassign
						err = nil
						return
					}
					ap, ok := apm[topPhotoArc.Aid]
					if !ok || ap == nil || ap.Arc == nil || !ap.Arc.IsNormal() || ap.Arc.Videos <= 0 {
						log.Warn("topPhoto arc mid:%d aid:%d not allowed", mid, topPhotoArc.Aid)
						return
					}
					arc := ap.Arc
					sp.Space.Archive = &space.TopPhotoArchive{
						Aid:      topPhotoArc.Aid,
						Cid:      arc.FirstCid,
						ImageURL: topPhotoArc.Pic,
						URI: func() string {
							var uri string
							playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
							uri = model.FillURI(model.GotoFullScreen, fmt.Sprintf("%d/%d/", topPhotoArc.Aid, arc.FirstCid), model.AvPlayHandlerGRPC(arc, playInfo))
							bvidStr, _ := bvid.AvToBv(topPhotoArc.Aid)
							//nolint:gosimple
							if strings.Index(uri, "?") == -1 {
								return fmt.Sprintf("%s?bvid=%s", uri, bvidStr)
							}
							return fmt.Sprintf("%s&bvid=%s", uri, bvidStr)
						}(),
					}
				}()
			}
		}
		return nil
	})
	if s.c.Cfg.GarbCfg.GoodsAvailable && (model.IsIPhone(plat) && build >= _iphoneGarbBuild || model.IsAndroid(plat) && build >= _androidGarbBuild) { // 有货并且符合版本过滤要求才进行粉丝相关逻辑
		supportGarb = true
		g.Go(func() error {
			currentEquip, equipErr := s.garbDao.SpaceBGEquip(ctx, vmid)
			if equipErr != nil {
				log.Error("%+v", equipErr)
				return nil
			}
			ownerEquip = currentEquip
			return nil
		})
		g.Go(func() error {
			userAssetReply, _ = s.garbDao.SpaceBGUserAssetList(ctx, mid, 1, 1)
			return nil
		})
	}
	g.Go(func() error {
		sp.Card.FansGroup, _ = s.bplusDao.GroupsCount(ctx, mid, vmid)
		return nil
	})
	if vmid == mid {
		g.Go(func() (err error) {
			if sp.Card.FansUnread, err = s.relDao.FollowersUnread(ctx, vmid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	g.Go(func() (e error) {
		var (
			likeMap map[string]int64
			allNum  int64
		)
		if likeMap, e = s.thumbupDao.UserLikedCounts(ctx, vmid, []string{_businessLike, _articleLike, _dynamicLike, _albumLike, _clipLike, _cheeseLike}); e != nil {
			log.Error("space:%v", e)
			e = nil
			return
		}
		for _, v := range likeMap {
			allNum += v
		}
		sp.Card.Likes = &space.LikesTmp{SkrTip: s.c.SpaceLikeRule.SkrTip, LikeNum: allNum}
		return
	})
	// (iOS蓝版)或者(国际版 iphone_i、android_i)强行去除充电模块
	if model.IsIPhone(plat) && (build >= 7000 && build <= 8000) || teenagersMode != 0 || lessonsMode != 0 || (model.IsOverseas(plat) && !enableOverseaElec(plat, build)) {
		sp.Elec = nil
	} else {
		// elec rank
		g.Go(func() (err error) {
			// http info 迁移到grpc
			info, err := s.elecDao.UPRankWithPanelByUPMid(ctx, vmid, int64(build), mid, mobiApp, platform, device)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if info != nil && info.RankElecUPProto != nil && info.State != -1 {
				sp.Elec = elec.FormatElec(ctx, info)
			}
			return
		})
	}
	g.Go(func() error {
		rel, _ := s.relDao.Relation(ctx, mid, vmid)
		// -999:无关系, -1:我已经拉黑该用户, 0:我悄悄关注了该用户, 1:我公开关注了该用户
		// 1- 悄悄关注 2 关注  6-好友 128-拉黑
		// Special 0-不是特别关注 1-特别关注
		if rel == nil {
			return nil
		}
		//nolint:gomnd
		if rel.Attribute == 1 {
			sp.Relation = 0
		} else if rel.Attribute == 2 || rel.Attribute == 6 {
			sp.Relation = 1
			//nolint:gomnd
		} else if rel.Attribute >= 128 {
			sp.Relation = -1
		} else {
			sp.Relation = -999
		}
		sp.RelSpecial = rel.Special
		return nil
	})
	sp.GuestRelation = -999
	if mid > 0 && mid != vmid {
		g.Go(func() error {
			// -999:无关系, -1:我已经拉黑该用户, 0:我悄悄关注了该用户, 1:我公开关注了该用户
			// 1- 悄悄关注 2 关注  6-好友 128-拉黑
			// Special 0-不是特别关注 1-特别关注
			gusRel, _ := s.relDao.Relation(ctx, vmid, mid)
			if gusRel == nil {
				return nil
			}
			//nolint:gomnd
			if gusRel.Attribute == 1 {
				sp.GuestRelation = 0
			} else if gusRel.Attribute == 2 || gusRel.Attribute == 6 {
				sp.GuestRelation = 1
				//nolint:gomnd
			} else if gusRel.Attribute >= 128 {
				sp.GuestRelation = -1
			}
			sp.GuestSpecial = gusRel.Special
			return nil
		})
		// 客态&已登录
		g.Go(func() error {
			displayNum, e := s.teenDao.CacheAttention(ctx, mid)
			if e != nil {
				log.Error("s.teenDao.CacheAttention(%d) error(%v)", mid, e)
				return nil
			}
			if s.c.Cfg.MaxDisplay > displayNum {
				sp.AttentionTip = &space.AttentionTip{CardNum: s.c.SpaceNewABTest.AtTestNum, Tip: "关注UP主，再也不迷路"}
			}
			return nil
		})
		// 契约者卡片逻辑
		g.Go(func() error {
			contractResource, isFollowDisplay, err := s.getContractResource(ctx, mid, vmid)
			if err != nil {
				log.Error("s.getContractResource mid=%d, vmid=%d, error=%+v", mid, vmid, err)
				return nil
			}
			if isFollowDisplay {
				sp.ContractResource = contractResource
			}
			return nil
		})
	}
	if cdm.ShowLive(mobiApp, device, build) && !model.IsOverseas(plat) {
		g.Go(func() (err error) {
			if sp.Medal, err = s.liveDao.QueryMedalStatus(ctx, vmid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	// g.Go(func() (err error) {
	// 	if sp.Attention, err = s.relDao.Attention(ctx, mid, vmid); err != nil {
	// 		log.Error("%+v", err)
	// 		err = nil
	// 	}
	// 	return
	// })
	if (mid != vmid) && sp.Card.IsDeleted == 1 {
		if err = g.Wait(); err != nil {
			log.Error("%v", err)
			return
		}
		// 默认头图
		sp.Space.TopPhoto.ImgURL = _defaultTopImage
		sp.Space.TopPhoto.NightImgURL = _defaultTopNightImage
		return
	}
	// 下面是各业务方数据
	if teenagersMode == 0 && lessonsMode == 0 && cdm.ShowLive(mobiApp, device, build) && !model.IsOverseas(plat) {
		g.Go(func() (err error) {
			if sp.Live, err = s.liveDao.Live(ctx, vmid, platform); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			// 我的直播间
			var live *struct {
				RoomStatus int `json:"roomStatus"`
			}
			//nolint:errcheck
			json.Unmarshal(sp.Live, &live)
			if live != nil && live.RoomStatus == 1 {
				hasHomeTab = true
			}
			return
		})
	}
	if (mid != vmid) && (s.c.SpaceTabABTest != nil) {
		sp.IsParams = s.c.SpaceTabABTest.IsParams
	}
	sp.Tab = &space.Tab{}
	// 投稿-视频
	g.Go(func() error {
		var err error
		if sp.Archive, err = s.UpArcs(ctx, mobiApp, device, mid, vmid, pn, ps, build, plat, true, space.ArchiveNew, isHant, sp.Setting); err != nil {
			log.Error("%+v", err)
		}
		if sp.Archive != nil && len(sp.Archive.Item) > 0 {
			sp.Tab.Archive = true
			hasHomeTab = true
		}
		return nil
	})
	// 动态
	g.Go(func() (err error) {
		if sp.Tab.Dynamic, err = s.bplusDao.Dynamic(ctx, vmid); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	// 投稿-专栏
	g.Go(func() error {
		sp.Article = s.UpArticles(ctx, vmid, 1, 3)
		if sp.Article != nil && len(sp.Article.Item) != 0 {
			sp.Tab.Article = true
			hasHomeTab = true
		}
		return nil
	})
	// 投稿音频
	g.Go(func() error {
		sp.Audios = s.audios(ctx, vmid, 1, 3)
		if sp.Audios != nil && len(sp.Audios.Item) != 0 {
			sp.Tab.Audios = true
			hasHomeTab = true
		}
		return nil
	})
	var (
		listReply   *garbgrpc.SpaceBGUserAssetListReply
		suitItemIDs []int64
		ownerFanIDs map[int64]*garbgrpc.UserFanInfoReply
	)
	g.Go(func() error {
		var err error
		if listReply, err = s.garbDao.SpaceBGUserAssetListWithFan(ctx, vmid, 1, 4); err != nil {
			log.Error("s.garbDao.SpaceBGUserAssetListWithFan vmid(%d) err(%+v)", vmid, err)
			return nil
		}
		for _, v := range listReply.List {
			if v != nil && v.Item != nil {
				suitItemIDs = append(suitItemIDs, v.Item.SuitItemID)
			}
		}
		if ownerFanIDs, err = s.garbDao.UserFanInfos(ctx, vmid, suitItemIDs); err != nil {
			log.Error("s.garbDao.UserFanInfos vmid(%d) err(%+v)", vmid, err)
		}
		return nil // 粉丝背景图容错
	})
	// 投稿-合集
	if (mobiApp == "android" && build > s.c.SpaceBuildLimit.UGCSeasonAndroid) || (model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.UGCSeasonIOS) || (mobiApp == "android_i" && build > s.c.SpaceBuildLimit.UGCSeasonAndroidI) || (mobiApp == "ipad") || (mobiApp == "android_hd") || (model.IsIPadPink(plat) && build > 65900000) {
		g.Go(func() error {
			sp.UGCSeason = s.UpSeasons(ctx, vmid, 1, 10, build, now, plat, mobiApp)
			if sp.UGCSeason != nil && len(sp.UGCSeason.Item) != 0 {
				sp.Tab.UGCSeason = true
				hasHomeTab = true
			}
			return nil
		})
	}
	// 投稿-系列
	g.Go(func() error {
		if sp.Series, err = s.upListSeries(ctx, vmid, fromViewAid); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		if sp.Series != nil && len(sp.Series.Item) != 0 {
			sp.Tab.Series = true
		}
		return nil
	})
	g.Go(func() error {
		if sp.Setting != nil {
			sp.Card.LiveFansWearing = s.asLiveFansWearing(ctx, sp.Setting.CloseSpaceMedal, vmid)
		}
		// 主态粉丝勋章默认显示客户端坑位图
		if mid == vmid && sp.Card.LiveFansWearing == nil {
			sp.Card.LiveFansWearing = &space.LiveFansWearing{
				ShowDefaultIcon: true,
				MedalJumpUrl:    fmt.Sprintf("https://live.bilibili.com/p/html/live-fansmedal-wall/index.html?is_live_webview=1&tId=%d#/medal", mid),
			}
		}
		return nil
	})
	// 数字艺术品展示
	g.Go(func() error {
		spaceGetNftArgs := &gallerygrpc.SpaceGetNFTReq{Mid: vmid, MobiApp: mobiApp, ViewMid: mid, FaceNftId: sp.Card.NftId}
		hasNftReply, err := s.galleryDao.SpaceHasNFT(ctx, spaceGetNftArgs)
		if err != nil {
			log.Error("s.galleryDao.SpaceHasNFT() spaceGetNftArgs=%+v, error=%+v", spaceGetNftArgs, err)
			return nil
		}
		sp.Card.HasFaceNft = hasNftReply.HasFace
		if sp.Card.FaceNftNew == 1 {
			// 当访问用户佩戴nft头像时，点击头像显示nft头像跳链
			sp.Card.NftFaceJump = replaceFaceJumpNftID(sp.Card.NftId, hasNftReply.FaceJump)
		}
		if hasNftReply.HasCert {
			// 用户是个人认证up主且有nft
			sp.Card.NftCertificate = &space.NftCertificate{
				DetailUrl: hasNftReply.Certificate.DetailUrl,
			}
		}
		if hasNftReply.HasArt {
			// 用户有数字艺术品
			sp.NftShowModule = space.ConvertToNftShowModule(hasNftReply.TotalArts, hasNftReply.ArtsMoreJump, hasNftReply.FloorTitle, hasNftReply.ArtsList)
			hasHomeTab = true
		}
		if hasNftReply.HasFace {
			// 拥有NFT头像
			sp.NftFaceButton = &space.NftFaceButton{
				FaceButtonChs: hasNftReply.FaceButtonChs,
				FaceButtonCht: hasNftReply.FaceButtonCht,
			}
		}
		return nil
	})
	if mid == vmid {
		// 仅主态调用
		g.Go(func() error {
			// 大会员过期动效
			vipSpaceLabelReply, err := s.vipDao.VipSpaceLabel(ctx, &vipinfogrpc.SpaceLabelReq{Mid: mid})
			if err != nil {
				log.Error("s.vipDao.VipSpaceLabel() err=%+v", err)
				return nil
			}
			sp.VipSpaceLabel = &space.VipSpaceLabel{ShowExpire: vipSpaceLabelReply.ShowExpire}
			if vipSpaceLabelReply.ShowExpire {
				sp.VipSpaceLabel.ExpireTextFrom = vipSpaceLabelReply.ExpireText.From
				sp.VipSpaceLabel.ExpireTextTo = vipSpaceLabelReply.ExpireText.To
				sp.VipSpaceLabel.LottieUri = vipSpaceLabelReply.ExpireAnimation.LottieUri
			}
			return nil
		})
	}
	// 青少年模式、课堂模式、国际版，不下发商品
	if teenagersMode == 0 && lessonsMode == 0 && !model.IsOverseas(plat) {
		// 新商品店铺开关开启，同时满足版本需求
		if s.c.Switch.AdOpen && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.SpaceShop, &feature.OriginResutl{
			MobiApp:    mobiApp,
			Device:     device,
			Build:      int64(build),
			BuildLimit: (mobiApp == "iphone" && device == "phone" && build > 8961) || (mobiApp == "android" && build >= 5510000),
		}) {
			g.Go(func() (err error) {
				var (
					adInfo  *cm.Ad
					shopRsc int64
				)
				switch plat {
				case model.PlatIPhone:
					shopRsc = _iphoneShopRsc
				case model.PlatAndroid:
					shopRsc = _androidShopRsc
				}
				if adInfo, err = s.adDao.AdVTwo(ctx, mid, vmid, build, buvid, []int64{shopRsc}, network, mobiApp, device, adExtra, spmid, fromSpmid); err != nil {
					log.Error("s.adDao.AdVTwo(%d,%d) error(%v)", mid, vmid, err)
					err = nil
					return
				}
				if adInfo != nil {
					sp.Tab.Mall = adInfo.ShowShopTab
					sp.AdSourceContentV2 = adInfo.SourceContent
					sp.AdShopType = adInfo.ShopTabType
				}
				return nil
			})
		} else if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.SpaceShopElse, &feature.OriginResutl{
			MobiApp:    mobiApp,
			Device:     device,
			Build:      int64(build),
			BuildLimit: (mobiApp == "iphone" && device == "phone" && build > 8820) || (mobiApp == "android" && build > 5475000),
		}) {
			g.Go(func() (err error) {
				var (
					adInfo  *cm.Ad
					shopRsc int64
				)
				switch plat {
				case model.PlatIPhone:
					shopRsc = _iphoneShopRsc
				case model.PlatAndroid:
					shopRsc = _androidShopRsc
				}
				if adInfo, err = s.adDao.Ad(ctx, mid, vmid, build, buvid, []int64{shopRsc}, network, mobiApp, device, adExtra, spmid, fromSpmid); err != nil {
					err = nil
					return
				}
				if adInfo != nil {
					sp.Tab.Mall = adInfo.ShowShopTab
					sp.AdSourceContent = adInfo.SourceContent
				}
				return nil
			})
		} else if (model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.FavAndroid) || (model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.FavIOS) {
			g.Go(func() (err error) {
				var mall *mallmdl.Mall
				if mall, err = s.mallDao.Mall(ctx, vmid); err != nil || mall == nil {
					log.Error("err(%v) mall(%v)", err, mall)
					err = nil
					return
				}
				if mall.TabState == mallmdl.TabYes {
					sp.Tab.Mall = true
				}
				if mall.Name != "" && mall.URL != "" {
					sm := &space.MallItem{}
					sm.FormMall(mall)
					sp.Mall = sm
					// hasHomeTab = true 产品反馈入口移到顶部 暂时不需要露出 预留以防万一
				}
				return
			})
		} else {
			// 商品
			g.Go(func() (err error) {
				var info *shop.Info
				if info, err = s.shopDao.Info(ctx, vmid, mobiApp, device, build); err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				if info != nil && info.Shop != nil && info.Shop.Status == 2 {
					sp.Tab.Shop = true
					sp.Shop = &space.Shop{ID: info.Shop.ID, Name: info.Shop.Name + _shopName}
					hasHomeTab = true
				}
				return
			})
		}
	}
	// 收藏
	if (model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.FavAndroid) ||
		(model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.FavIOS) ||
		(model.IsAndroidHD(plat)) {
		// build limit 5.38
		g.Go(func() error {
			var mediaCount int
			sp.Favourite2, mediaCount = s.favFolders2(ctx, mid, vmid, sp.Setting, plat, build, mobiApp, device)
			// sp.Favourite2.Item里面只包含我创建的收藏，不包含我的收藏
			if sp.Favourite2 != nil && len(sp.Favourite2.Item) != 0 {
				sp.Tab.Favorite = true
				hasHomeTab = true
				// 客态情况下,主页收藏版块只有一个默认收藏夹并且默认收藏夹内没有收藏任何内容,不下发该模块
				if mid != vmid && sp.Favourite2.Count == 1 && sp.Favourite2.Item[0].Count == 0 && sp.Favourite2.Item[0].IsDefault {
					sp.Favourite2.Item = []*favorite.Folder2{}
					sp.Favourite2.Count = 0
					// 收藏tab在客态情况下不再下发
					sp.Tab.Favorite = false
				}
			}
			// 客人态下，有我的收藏就下发标签
			if mid != vmid && sp.Favourite2 != nil && mediaCount > 0 {
				sp.Tab.Favorite = true
				hasHomeTab = true
			}
			return nil
		})
	} else {
		g.Go(func() error {
			sp.Favourite = s.favFolders(ctx, mid, vmid, sp.Setting, plat, build, mobiApp)
			if sp.Favourite != nil && len(sp.Favourite.Item) != 0 {
				sp.Tab.Favorite = true
				hasHomeTab = true
			}
			return nil
		})
	}
	if teenagersMode == 0 && lessonsMode == 0 {
		// 追番
		g.Go(func() error {
			sp.Season = s.Bangumi(ctx, mid, vmid, sp.Setting, pn, ps)
			if sp.Season != nil && len(sp.Season.Item) != 0 {
				sp.Tab.Bangumi = true
				hasHomeTab = true
			}
			return nil
		})
		if !model.IsOverseas(plat) {
			// 看板娘空间
			g.Go(func() error {
				characterReply, err = s.garbDao.GetUserSpaceCharacterInfo(ctx, &live2dgrpc.GetUserSpaceCharacterInfoReq{Mid: vmid})
				if err != nil {
					log.Error("s.garbDao.GetUserSpaceCharacterInfo vmid=%d, err=%+v", vmid, err)
					err = nil
					return nil
				}
				return nil
			})
			//大航海信息获取
			if s.c.Switch.GuardOpen {
				g.Go(func() error {
					sp.Guard = s.GuardList(ctx, mid, vmid, 1, 4)
					return nil
				})
			}
			// 课堂
			if s.cheeseDao.HasCheese(plat, build, true) {
				g.Go(func() error {
					sp.Cheese = s.UpCheese(ctx, vmid, pn, pageSizeFromCheesePlat(ctx))
					if sp.Cheese != nil && len(sp.Cheese.Item) != 0 {
						sp.Tab.Cheese = true
						hasHomeTab = true
					}
					return nil
				})
			}
			// iPad不用下发，直接屏蔽
			if plat != model.PlatIPad && plat != model.PlatIpadHD && filtered != "1" {
				g.Go(func() error {
					sp.PlayGame = s.PlayGames(ctx, mid, vmid, sp.Setting, pn, ps, platform)
					if sp.PlayGame != nil && len(sp.PlayGame.Item) > 0 {
						hasHomeTab = true
					}
					return nil
				})
			}
		}
	}
	g.Go(func() error {
		sp.CoinArc = s.CoinArcs(ctx, mid, vmid, sp.Setting, pn, ps, isHant, mobiApp, device)
		if sp.CoinArc != nil && len(sp.CoinArc.Item) != 0 {
			sp.Tab.Coin = true
			hasHomeTab = true
		}
		return nil
	})
	g.Go(func() error {
		sp.LikeArc = s.LikeArcs(ctx, mid, vmid, sp.Setting, pn, ps, isHant, mobiApp, device)
		if sp.LikeArc != nil && len(sp.LikeArc.Item) != 0 {
			sp.Tab.Like = true
			hasHomeTab = true
		}
		return nil
	})
	if !model.IsOverseas(plat) {
		// 投稿-漫画
		if (mobiApp == "android" && build > s.c.SpaceBuildLimit.ComicAndroid) || (model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.ComicIOS) || (mobiApp == "android_i" && build > s.c.SpaceBuildLimit.ComicAndroidI) {
			g.Go(func() error {
				sp.Comic = s.UpComics(ctx, vmid, pn, ps, build, now, plat, false)
				if sp.Comic != nil && len(sp.Comic.Item) != 0 {
					sp.Tab.Comic = true
					hasHomeTab = true
				}
				return nil
			})
		}
		// 追漫(暂时无视)
		if (model.IsIPhone(plat) && build > s.c.SpaceBuildLimit.SubComicIOS) || (model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.SubComicAndroid) {
			g.Go(func() (err error) {
				sp.SubComic = s.SubComics(ctx, mid, vmid, sp.Setting, pn, ps, build, plat)
				if sp.SubComic != nil && len(sp.SubComic.Item) != 0 {
					sp.Tab.SubComic = true
					hasHomeTab = true
				}
				return
			})
		}
		// 官号导流下载浮层
		g.Go(func() error {
			sp.LeadDownload, _ = s.spcDao.OfficialDownload(ctx, vmid, plat)
			return nil
		})
	}
	// 粉丝彩蛋
	g.Go(func() error {
		fans, e := s.relDao.SpecialEffect(ctx, mid, vmid, buvid)
		if e != nil {
			log.Error("s.relDao.SpecialEffect(%d,%d) error(%v)", vmid, mid, e)
			return nil
		}
		if fans != nil {
			tmpF := &space.FansEffect{Show: fans.IsShowEffect, ResourceID: fans.ResourceId, AchieveType: fans.AchieveType}
			// 557以下版本，只能下发千万粉丝动效
			if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.SpaceFansEffect, &feature.OriginResutl{
				MobiApp:    mobiApp,
				Device:     device,
				Build:      int64(build),
				BuildLimit: (mobiApp == "android" && build < 5570000) || (mobiApp == "iphone" && build <= 9290),
			}) {
				if fans.AchieveType == 1 {
					sp.FansEffect = tmpF
				}
			} else {
				sp.FansEffect = tmpF
			}
		}
		return nil
	})
	// 活动tab
	var activity *spacegrpc.UserTabReply
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.SpaceTabActivity, &feature.OriginResutl{
		MobiApp:    mobiApp,
		Device:     device,
		Build:      int64(build),
		BuildLimit: (mobiApp == "android" && build >= 6090000) || (mobiApp == "iphone" && build > 10240),
	}) {
		g.Go(func() error {
			act, err := s.spcDao.ActivityTab(ctx, vmid, int32(plat), int32(build))
			if err != nil {
				log.Error("Failed to get activity tab vmid: %d, error: %+v", vmid, err)
				return nil
			}
			if act == nil {
				return nil
			}
			activity = act
			sp.Activity = &space.Activity{PageId: act.TabCont, H5Link: act.H5Link}
			sp.Tab.Activity = true
			return nil
		})
	}
	// 发起活动模块，企业认证账号
	//nolint:gomnd
	if sp.Card.OfficialVerify.Role == 3 {
		g.Go(func() error {
			sp.CreatedActivity = s.createdActList(ctx, vmid, device)
			return nil
		})
	}
	// 数字藏品入口
	g.Go(func() error {
		var err error
		digitalReply, err = s.digitalDao.DigitalEntry(ctx, vmid)
		if err != nil {
			log.Error("s.digitalDao.DigitalEntry vmid=%d, mid=%d, err=%+v", vmid, mid, err)
			return nil
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("g.Wait() %+v", err)
	}
	// 空间公告为空(nil或者结构体数据为空)&是高仿号
	if (sp.Card.PRInfo == nil || (sp.Card.PRInfo.MID == 0 && sp.Card.PRInfo.Content == "" && sp.Card.PRInfo.URL == "")) && sp.Card.IsFakeAccount == 1 {
		sp.Card.PRInfo = &space.PRInfo{Content: "经举报，该账号存在冒充他人账号风险，请勿轻易相信其散布的信息，以防上当受骗。"}
		if vmid == mid {
			sp.Card.PRInfo = &space.PRInfo{Content: "您的账号经举报，存在冒充他人账号的嫌疑，请尽快在PC端进行身份认证。"}
		}
		sp.Card.PRInfo.Icon = _prInfoOldIcon
		sp.Card.PRInfo.IconNight = _prInfoOldIconNight
		sp.Card.PRInfo.BgColor = _prInfoOldBgColor
		sp.Card.PRInfo.BgColorNight = _prInfoOLdBgColorNight
		sp.Card.PRInfo.TextColor = _prInfoOldTextcolor
		sp.Card.PRInfo.TextColorNight = _prInfoOldTextcolorNight
	}
	// 空间拉黑互不可见/隐藏处理
	sp.HiddenAttribute = asHiddenAttribute(c, vmid, sp.Relation, sp.GuestRelation)
	if sp.HiddenAttribute != nil {
		if isHant {
			i18n.TranslateAsTCV2(&sp.HiddenAttribute.Text)
		}
		// 拉黑时服务端暂时单独处理隐藏预约卡，之后统一下发固定结构
		sp.ReservationCardList = nil
		sp.ReservationCardInfo = nil
	}
	// 默认tab逻辑
	sp.DefaultTab = space.HomeTab
	if vmid != mid && sp.Tab != nil {
		switch {
		case sp.Tab.Archive:
			sp.DefaultTab = space.VideoTab
		case sp.Tab.Activity:
			sp.DefaultTab = space.ActivityTab
		case sp.Tab.Article:
			sp.DefaultTab = space.ArticleTab
		case sp.Tab.Audios:
			sp.DefaultTab = space.AudiosTab
		case sp.Tab.Comic:
			sp.DefaultTab = space.ComicTab
		case sp.Tab.UGCSeason:
			sp.DefaultTab = space.SeasonTab
			// case sp.Tab.Album:
			// 	sp.DefaultTab = space.AlbumTab
			// case sp.Tab.Clip:
			// 	sp.DefaultTab = space.ClipTab
		}
	}
	if activity != nil && activity.IsDefault == 1 {
		sp.DefaultTab = space.ActivityTab
		sp.PreferSpaceTab = true
	}
	// 新tab 主页、动态、投稿、商品、收藏、追番、课堂
	var tabTmp []*space.TabItem
	if sp.Tab.Dynamic {
		tabTmp = append(tabTmp, &space.TabItem{Title: "动态", Param: space.DyanmicTab})
	}
	// 投稿(全部、视频、专栏、音频、漫画、合集、系列)
	var contributeTab, contributeTabTmp []*space.TabItem
	if sp.Tab.Archive {
		contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: "视频", Param: space.VideoTab})
	}
	if sp.Tab.Article {
		contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: "专栏", Param: space.ArticleTab})
	}
	if sp.Tab.Audios {
		contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: "音频", Param: space.AudiosTab})
	}
	if sp.Tab.Comic {
		contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: "漫画", Param: space.ComicTab})
	}
	if sp.Tab.UGCSeason {
		contributeTabTmp = resolveUgcSeasonContributeTab(c, contributeTabTmp, sp.UGCSeason)
	}
	// 特殊处理：选项开启 当用户仅有视频，没有专栏|合集|漫画，且仅有直播回放一个系列时，隐藏视频 直播回放一行
	hideLivePlayback := sp.Setting != nil && sp.Setting.LivePlayback == 1 && sp.Tab.Archive && !sp.Tab.Article && !sp.Tab.Audios && !sp.Tab.UGCSeason && !sp.Tab.Comic
	if sp.Tab.Series {
		hideLivePlayback = hideLivePlayback && len(sp.Series.Item) == 1
		for _, series := range sp.Series.Item {
			if hideLivePlayback && series.IsLivePlayBack {
				continue
			}
			contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: series.Name, Param: space.SeriesTab, SeriesId: series.SeriesId, Mtime: series.Mtime})
		}
	}
	if s.c.Switch.SpaceContributeAll && len(contributeTabTmp) > 1 {
		contributeTab = append(contributeTab, &space.TabItem{Title: "全部", Param: space.AllTab})
	}
	contributeTab = append(contributeTab, contributeTabTmp...)
	if len(contributeTab) > 0 {
		tabTmp = append(tabTmp, &space.TabItem{Title: "投稿", Param: space.ContributeTab, Items: contributeTab})
	}
	if sp.Tab.Activity && activity.TabOrder == 0 { // 未配置活动tab位置，默认跟在投稿后面，无投稿时跟在投稿前一个的后面，依次类推
		tabTmp = append(tabTmp, &space.TabItem{Title: activity.TabName, Param: space.ActivityTab})
	}
	if sp.Tab.Mall {
		tabTmp = append(tabTmp, &space.TabItem{Title: "商品", Param: space.ShopTab})
	}
	if sp.Tab.Favorite {
		tabTmp = append(tabTmp, &space.TabItem{Title: "收藏", Param: space.FavoriteTab})
	}
	if sp.Tab.Bangumi {
		tabTmp = append(tabTmp, &space.TabItem{Title: "追番", Param: space.BangumiTab})
	}
	if sp.Tab.Cheese {
		tabTmp = append(tabTmp, &space.TabItem{Title: "课堂", Param: space.CheeseTab})
	}
	if len(tabTmp) > 0 || hasHomeTab {
		sp.Tab2 = append(sp.Tab2, &space.TabItem{Title: "主页", Param: space.HomeTab})
		sp.Tab2 = append(sp.Tab2, tabTmp...)
	}

	if sp.Tab.Activity && activity.TabOrder != 0 { // 配置活动tab位置
		sp.Tab2 = adjustTab2(sp.Tab2, activity)
	}
	if supportGarb {
		s.garbInfo(c, sp.Space, ownerEquip, userAssetReply, vmid, mid)
		if sp.Setting == nil || sp.Setting.DressUp == 1 || vmid == mid {
			if sp.Setting == nil {
				log.Warn("spSetting is nil")
			}
			if listReply != nil {
				sp.FansDress = new(space.GarbDressReply)
				sp.FansDress.Count = listReply.Total
				sp.FansDress.Total = listReply.All
				for _, v := range listReply.List {
					if v == nil || v.Item == nil {
						continue
					}
					item := new(space.GarbDressItem)
					item.FromGarb(v.Item, ownerFanIDs)
					sp.FansDress.Items = append(sp.FansDress.Items, item)
				}
				if len(sp.FansDress.Items) > 0 {
					hasHomeTab = true
				}
			}
		}
	}
	// 空间展示看板娘
	if characterReply != nil {
		sp.Space.ShowCharacter = characterReply.ShowEntry
		if characterReply.Active {
			sp.Space.ImgURL = characterReply.Preview
			sp.Space.NightImgURL = characterReply.PreviewNight
			sp.Space.CharacterInfo = &space.CharacterInfo{IsActive: characterReply.Active}
		}
	}
	// 空间数字藏品
	if digitalReply != nil {
		sp.Space.ShowDigital = digitalReply.ShowEntry
		sp.Space.DigitalInfo = &space.DigitalInfo{
			Active:              digitalReply.Active,
			HeadUrl:             digitalReply.HeadUrl,
			ItemId:              digitalReply.ItemId,
			NftId:               digitalReply.NftId,
			JumpUrl:             digitalReply.JumpUrl,
			RegionType:          int32(digitalReply.NftRegion),
			Icon:                digitalReply.NftRegionIcon,
			AnimationUrlList:    digitalReply.AnimationUrlList,
			NftType:             int32(digitalReply.Type),
			BackgroundHandle:    int32(digitalReply.BackgroundHandle),
			AnimationFirstFrame: digitalReply.AnimationFirstFrame,
			MusicAlbum:          digitalReply.MusicAlbum,
			Animation:           digitalReply.Animation,
			NftRegionTitle:      digitalReply.NftRegionTitle,
			NFTImage:            digitalReply.GetNftImage(),
		}
		if digitalReply.Active {
			sp.Space.ImgURL = digitalReply.PreviewUrl
			sp.Space.NightImgURL = digitalReply.PreviewUrlNight

		}
	}
	// 开关控制信息再处理
	resolveSpaceSetting(sp)
	return
}

func pageSizeFromCheesePlat(ctx context.Context) int {
	const (
		_PadCheesePageSize     = 4
		_defaultCheesePageSize = 2
	)
	if pd.WithContext(ctx).IsPlatAndroidHD().Or().IsPlatIPad().Or().IsPlatIPadHD().MustFinish() {
		return _PadCheesePageSize
	}
	return _defaultCheesePageSize
}

func resolveUgcSeasonContributeTab(ctx context.Context, contributeTabTmp []*space.TabItem, ugcSeasonList *space.UGCSeasonList) []*space.TabItem {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidI().Or().IsPlatAndroidB().And().Build("<", int64(6580000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().Or().IsMobiAppIPhoneI().Or().IsPlatIPhoneB().And().Build("<", int64(65500000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33400000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build("<", int64(1110000))
	}).MustFinish() {
		contributeTabTmp = append(contributeTabTmp, &space.TabItem{Title: "合集", Param: space.SeasonTab})
		return contributeTabTmp
	}
	for _, v := range ugcSeasonList.Item {
		if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
			pd.IsPlatAndroid().And().Build(">=", 6780000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPhone().And().Build(">=", 67800100)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPad().And().Build(">=", 68000000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPadHD().And().Build(">=", 34500000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatAndroidHD().And().Build(">=", 1230000)
		}).MustFinish() {
			if v.IsNoSpace {
				// 刷屏合集不展示tab
				continue
			}
		}
		item := &space.TabItem{
			Title: func() string {
				if v.IsPay {
					return fmt.Sprintf("付费·%s", v.Title)
				}
				return v.Title
			}(),
			Param:    space.SeasonVideoTab,
			SeasonId: v.SeasonId,
			Mtime:    v.MTime,
		}
		contributeTabTmp = append(contributeTabTmp, item)
	}
	return contributeTabTmp
}

func (s *Service) getDynSimpleInfo(ctx context.Context, dynIds []int64) map[int64]*dyngrpc.DynSimpleInfo {
	if len(dynIds) == 0 {
		return nil
	}
	res, err := s.dynDao.DynSimpleInfos(ctx, &dyngrpc.DynSimpleInfosReq{
		DynIds: dynIds,
	})
	if err != nil {
		log.Error("getDynSimpleInfo() s.dynDao.DynSimpleInfos, dynIds=%+v, err=%+v", dynIds, err)
		return nil
	}
	return res.DynSimpleInfos
}

func skipReserveVersionControl(cardType activitygrpc.UpActReserveRelationType, build int, plat int8) bool {
	switch cardType {
	case activitygrpc.UpActReserveRelationType_ESports:
		if (model.IsAndroid(plat) && build < 6340000) || (model.IsIOS(plat) && build < 63400000) {
			return true
		}
	default:
	}
	return false
}

func makeDynIdsFromReserveInfo(data []*activitygrpc.UpActReserveRelationInfo) []int64 {
	var dynIds []int64
	for _, v := range data {
		if v.DynamicId == "" {
			continue
		}
		id, err := strconv.ParseInt(v.DynamicId, 10, 64)
		if err != nil {
			log.Error("makeDynIdsFromReserveInfo() strconv.ParseInt DynamicId=%s, err=%+v", v.DynamicId, err)
			continue
		}
		dynIds = append(dynIds, id)
	}
	return dynIds
}

// 设置开关控制空间直播粉丝勋章点亮展示
func (s *Service) asLiveFansWearing(ctx context.Context, closeSpaceMedal int, vmid int64) *space.LiveFansWearing {
	if closeSpaceMedal != 0 {
		return nil
	}
	wearingInfo, err := s.liveDao.WearingInfo(ctx, &livexfans.WearingReq{
		UserId:     vmid,
		ForceLight: true,
	})
	if err != nil {
		log.Error("s.liveDao.WearingInfo %+v", err)
		return nil
	}
	return space.WearingInfoChange(wearingInfo)
}

func asHiddenAttribute(ctx context.Context, vmid int64, relation, guestRelation int) *space.HiddenAttribute {
	// mid int64过滤
	if midInt64.IsDisableInt64MidVersion(ctx) && midInt64.CheckHasInt64InMids(vmid) {
		return &space.HiddenAttribute{
			IsSpaceHidden: true,
			Text:          "APP版本过低，请升级更新后查看",
		}
	}
	if relation != -1 && guestRelation != -1 {
		return nil
	}
	attr := &space.HiddenAttribute{
		IsSpaceHidden: true,
	}
	if guestRelation == -1 {
		attr.Text = "由于对方隐私设置\n无法查看空间内容"
		return attr
	}
	attr.Text = "无法查看空间内容\n请将该用户移除黑名单"
	return attr
}

func asSpaceReservationCardList(data []*activitygrpc.UpActReserveRelationInfo, isSpaceOwner bool, simpleInfo map[int64]*dyngrpc.DynSimpleInfo, build int, plat int8) []*space.UpActReserveRelationInfo {
	var res []*space.UpActReserveRelationInfo
	for _, v := range data {
		// 预约先审后发：主态可见，客态不可见
		if !isSpaceOwner && v.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
			continue
		}
		// 特殊预约类型版本控制
		if skipReserveVersionControl(v.Type, build, plat) {
			continue
		}
		// 预约隐藏特定渠道
		if _, ok := v.Hide[int64(activitygrpc.UpCreateActReserveFrom_FromSpace)]; ok {
			continue
		}
		info := &space.UpActReserveRelationInfo{}
		info.FromUpActReserveRelationInfo(v)
		info.FromUpActReserveLotteryInfo(v)
		if !model.IsOverseas(plat) {
			info.IsDynamicValid = resolveIsDynamicValid(info.DynamicId, simpleInfo)
		}
		if isSpaceOwner {
			// 后置主态判断逻辑
			if v.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
				info.Name = "[审核中]" + info.Name
			}
			info.ShowText2 = true
		}
		res = append(res, info)
	}
	return res
}

func resolveIsDynamicValid(dynamicId string, simpleInfo map[int64]*dyngrpc.DynSimpleInfo) bool {
	if dynamicId == "" {
		return false
	}
	id, err := strconv.ParseInt(dynamicId, 10, 64)
	if err != nil {
		log.Error("resolveIsDynamicValid() strconv.ParseInt dynamicId=%s, err=%+v", dynamicId, err)
		return false
	}
	if v, ok := simpleInfo[id]; ok {
		return v.Visible
	}
	return false
}

func adjustTab2(tab []*space.TabItem, activity *spacegrpc.UserTabReply) []*space.TabItem {
	if int(activity.TabOrder) > len(tab) {
		tab = append(tab, &space.TabItem{Title: activity.TabName, Param: space.ActivityTab})
		return tab
	}
	tab = append(tab[:activity.TabOrder-1],
		append([]*space.TabItem{{Title: activity.TabName, Param: space.ActivityTab}},
			tab[activity.TabOrder-1:]...)...)
	return tab
}

// nolint:gocognit
// UpArcs get upload archive .
func (s *Service) UpArcs(c context.Context, mobiApp, device string, mid, uid int64, pn, ps, build int, plat int8, isPopularBadge bool, order string, isHant bool, setting *space.Setting) (*space.ArcList, error) {
	var (
		arcs     []*api.ArcPlayer
		uname    string
		aids     []int64
		position *spm.HistoryPosition
		playAvs  []*api.PlayAv
	)
	res := &space.ArcList{Item: []*space.ArcItem{}}
	res.Order = append(res.Order, &space.ArcOrder{Title: "最新发布", Value: space.ArchiveNew})
	res.Order = append(res.Order, &space.ArcOrder{Title: "最多播放", Value: space.ArchivePlay})
	if (model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.CooperationAndroid) ||
		(model.IsIOS(plat) && build > s.c.SpaceBuildLimit.CooperationIOS) ||
		(plat == model.PlatIpadHD && build > s.c.SpaceBuildLimit.IPadHDArchiveSort) ||
		(mobiApp == "android_i" && build > 2042030) || (mobiApp == "iphone_i") ||
		(model.IsAndroidHD(plat)) {
		var needPlayurl bool
		if (model.IsAndroidPick(plat) && build > s.c.SpaceBuildLimit.PlayurlAndroid) || (model.IsIOSPick(plat) && build > s.c.SpaceBuildLimit.PlayurlIOS) {
			needPlayurl = true
		}
		var without []uparcapi.Without
		if setting == nil {
			var err error
			setting, err = s.spcDao.Setting(c, uid)
			if err != nil {
				log.Error("%+v", err)
			}
		}
		if setting != nil && setting.LivePlayback != 1 {
			without = append(without, uparcapi.Without_live_playback)
		}
		if !model.IsIPad(plat) {
			without = append(without, uparcapi.Without_no_space)
		}
		if (model.IsIPadPink(plat) && build >= 68000000) || (model.IsIPadHD(plat) && build >= 34500000) || (model.IsAndroidHD(plat) && build >= 1230000) {
			without = append(without, uparcapi.Without_no_space)
		}
		g, ctx := errgroup.WithContext(c)
		switch order {
		case space.ArchiveNew:
			g.Go(func() error {
				arcsTmp, upArcTotal, err := s.upArcDao.ArcPassed(ctx, uid, int64(pn), int64(ps), "", without)
				if err != nil {
					log.Error("UpArcs s.upArcDao.ArcPassed mid:%d pn:%d ps:%d error:%v", mid, pn, ps, err)
					return err
				}
				res.Count = int(upArcTotal)
				for _, item := range arcsTmp {
					if item == nil {
						continue
					}
					playAvs = append(playAvs, &api.PlayAv{Aid: item.Aid})
					aids = append(aids, item.Aid)
				}
				var apm map[int64]*api.ArcPlayer
				if needPlayurl {
					apm, err = s.arcDao.ArcsPlayer(ctx, playAvs, false)
					if err != nil {
						log.Error("%v", err)
						return err
					}
					for _, aid := range aids {
						if arc, ok := apm[aid]; ok {
							arcs = append(arcs, arc)
						}
					}
					return nil
				}
				for _, item := range arcsTmp {
					if item == nil {
						continue
					}
					arcs = append(arcs, &api.ArcPlayer{Arc: space.FromUpArcToArc(item)})
				}
				return nil
			})
		case space.ArchivePlay:
			g.Go(func() error {
				reply, upArcTotal, err := s.upArcDao.ArcPassed(ctx, uid, int64(pn), int64(ps), "click", without)
				if err != nil {
					log.Error("UpArcs s.upArcDao.ArcPassed mid:%d pn:%d ps:%d error:%v", mid, pn, ps, err)
					return err
				}
				res.Count = int(upArcTotal)
				if !needPlayurl {
					for _, val := range reply {
						arc := &api.ArcPlayer{
							Arc: model.CopyFromArc(val),
						}
						arcs = append(arcs, arc)
					}
					return nil
				}
				for _, val := range reply {
					playAvs = append(playAvs, &api.PlayAv{Aid: val.Aid})
					aids = append(aids, val.Aid)
				}
				if len(aids) > 0 {
					var (
						apm map[int64]*api.ArcPlayer
						err error
					)
					if needPlayurl {
						apm, err = s.arcDao.ArcsPlayer(ctx, playAvs, false)
					} else {
						apm, err = s.arcDao.Arcs(ctx, aids, mobiApp, device, mid)
					}
					if err != nil {
						log.Error("%+v", err)
						return err
					}
					for _, aid := range aids {
						if arc, ok := apm[aid]; ok {
							arcs = append(arcs, arc)
						}
					}
				}
				return nil
			})
		default:
			log.Warn("invalid order:%v", order)
		}
		g.Go(func() error {
			account, err := s.accDao.Card(ctx, uid)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			uname = account.Name
			return nil
		})
		// 获取连续播放进度（增加了版本控制，暂时仅IOS 5.59以上版本走该逻辑）
		if mid > 0 && ((model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.ContinuePlayAndroid) || (model.IsIOS(plat) && build > s.c.SpaceBuildLimit.ContinuePlayIOS)) ||
			(plat == model.PlatIpadHD && build > s.c.SpaceBuildLimit.IPadHDArchiveSort) {
			g.Go(func() (err error) {
				if position, err = s.hisDao.Position(ctx, mid, uid, _spaceArchiveBusiness); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
	} else {
		without := []uparcapi.Without{uparcapi.Without_staff}
		arcsTmp, upArcTotal, err := s.upArcDao.ArcPassed(c, uid, int64(pn), int64(ps), "", without)
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res.Count = int(upArcTotal)
		for _, item := range arcsTmp {
			if item == nil {
				continue
			}
			arcs = append(arcs, &api.ArcPlayer{Arc: space.FromUpArcToArc(item)})
		}
	}

	if len(arcs) != 0 {
		var seasons map[int64]*ugcSeasonGrpc.Season // 以aid为key的合集信息
		if pd.WithContext(c).Where(func(pd *pd.PDContext) {
			pd.IsPlatAndroid().And().Build(">=", 6780000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPhone().And().Build(">=", 67800100)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPad().And().Build(">=", 68000000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPadHD().And().Build(">=", 34500000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatAndroidHD().And().Build(">=", 1230000)
		}).MustFinish() {
			if seasonDisplay, err := s.ArcReorderBySeason(c, arcs, nil, order, uid); err == nil {
				arcs = seasonDisplay.ArcPlayer
				seasons = seasonDisplay.Seasons
			}
		}
		res.Item = make([]*space.ArcItem, 0, len(arcs))
		for _, v := range arcs {
			if v.Arc == nil || !v.GetArc().IsNormal() {
				continue
			}
			if isHant {
				out := chinese.Converts(c, v.Arc.Title, v.Arc.TypeName, v.Arc.Desc)
				v.Arc.Title = out[v.Arc.Title]
				v.Arc.TypeName = out[v.Arc.TypeName]
				v.Arc.Desc = out[v.Arc.Desc]
			}
			si := &space.ArcItem{}
			if isPopularBadge {
				si.FromArc(v, s.hotAids, s.c.Custom.UpArcHasShare, true, seasons[v.Arc.Aid])
			} else {
				si.FromArc(v, nil, s.c.Custom.UpArcHasShare, true, seasons[v.Arc.Aid])
			}
			res.Item = append(res.Item, si)
		}
		if res.Count < ps {
			res.Count = len(res.Item)
		}
		if s.c.SpaceArchive.EpisodicOpen &&
			(!model.IsPad(plat) ||
				((plat == model.PlatIpadHD) || (model.IsIPad(plat) && build >= 63100000))) &&
			uname != "" &&
			len(arcs) > 1 { // 只有1个稿件的时候不显示一键连播按钮
			for _, item := range s.c.SpaceArchive.EpisodicMid {
				if item == uid {
					return res, nil
				}
			}
			res.EpisodicButton = new(space.EpisodicButton)
			res.EpisodicButton.Text = s.c.SpaceArchive.EpisodicText
			params := url.Values{}
			params.Set("offset", _spaceArchiveOffset)
			params.Set("desc", _spaceArchiveDesc)
			params.Set("oid", _spaceArchiveOid)
			if order == space.ArchiveNew && position != nil {
				//6.12之前根据offset判断文案，6.12及之后根据oid来判断文案
				if ((model.IsAndroid(plat) && build > s.c.SpaceBuildLimit.ButtonTextAnd) || (model.IsIOS(plat) && build > s.c.SpaceBuildLimit.ButtonTextIOS)) ||
					(plat == model.PlatIpadHD && build > s.c.SpaceBuildLimit.ButtonTextIpad) {
					if position.Oid > 0 {
						res.EpisodicButton.Text = s.c.SpaceArchive.EpisodicText1
					}
				} else {
					if position.Offset > 0 {
						res.EpisodicButton.Text = s.c.SpaceArchive.EpisodicText1
					}
				}
				params.Set("offset", strconv.Itoa(position.Offset))
				params.Set("desc", strconv.Itoa(position.Desc))
				params.Set("oid", strconv.Itoa(position.Oid))
			}
			params.Set("ps", _spaceArchivePS)
			params.Set("order", _spaceArchiveOrder)
			params.Set("page_type", _spaceArchivePageType)
			params.Set("user_name", uname)
			params.Set("playlist_intro", s.c.SpaceArchive.EpisodicDesc)
			params.Set("total_count", strconv.Itoa(res.Count))
			switch order {
			case space.ArchiveNew:
				params.Set("sort_field", "1")
			case space.ArchivePlay:
				params.Set("sort_field", "2")
				params.Set("sort_hidden", "1")
			}
			res.EpisodicButton.Uri = fmt.Sprintf(_spaceArchiveUri, uid, params.Encode())
		}
	}
	return res, nil
}

// // upListSeries get upper listseries
func (s *Service) upListSeries(c context.Context, mid, aid int64) (*space.SeriesList, error) {
	ListSeriesRsp, err := s.seriesDao.ListSeries(c, mid, aid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if len(ListSeriesRsp.SeriesList) > _spaceSeriesListMaxSize {
		ListSeriesRsp.SeriesList = ListSeriesRsp.SeriesList[:_spaceSeriesListMaxSize]
	}
	seriesReply := &space.SeriesList{Item: []*space.SeriesItem{}}
	for _, series := range ListSeriesRsp.SeriesList {
		if series.Meta == nil || series.Meta.Name == "" {
			log.Warn("series meta数据错误,data:%+v", series.Meta)
			continue
		}
		seriesItem := &space.SeriesItem{
			Name:           series.Meta.Name,
			SeriesId:       series.Meta.SeriesId,
			IsLivePlayBack: series.Meta.Category == seriesgrpc.SeriesLiveReplay,
			Mtime:          series.Meta.MTime,
		}
		seriesReply.Item = append(seriesReply.Item, seriesItem)
	}
	return seriesReply, nil
}

// UpSeries get upload series .
func (s *Service) UpSeries(c context.Context, vmid, seriesId, ps, next int64, sort string, isPopularBadge, isHant bool) (*space.SeriesArchiveList, error) {
	var (
		listArchivesCursorResp *seriesgrpc.ListArchivesCursorResp
		series                 *seriesgrpc.SeriesData
	)
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		listArchivesCursorResp, err = s.seriesDao.ListArchivesCursor(ctx, vmid, seriesId, ps, next, sort)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		series, err = s.seriesDao.Series(ctx, seriesId)
		if err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if listArchivesCursorResp.Cursor == nil {
		return nil, errors.New("listArchivesCursorResp.Cursor should not be nil")
	}
	res := &space.SeriesArchiveList{
		Item: []*space.ArcItem{},
	}
	res.Next = listArchivesCursorResp.Cursor.Next
	var (
		apm     map[int64]*api.ArcPlayer
		playAvs []*api.PlayAv
		err     error
	)
	aids := listArchivesCursorResp.Aids
	for _, aid := range aids {
		playAvs = append(playAvs, &api.PlayAv{Aid: aid})
	}
	apm, err = s.arcDao.ArcsPlayer(c, playAvs, false)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	var arcs []*api.ArcPlayer
	for _, aid := range aids {
		arc, ok := apm[aid]
		if !ok {
			log.Warn("arc aid= %d is not found", aid)
			continue
		}
		arcs = append(arcs, arc)
	}
	// 倒序影响播单跳链
	episodicButtonSortDesc := 1
	if sort == "asc" {
		episodicButtonSortDesc = 0
	}
	for _, arc := range arcs {
		v := arc.Arc
		if v == nil || !v.IsNormal() {
			log.Warn("v = %v is not normal", v)
			continue
		}
		if isHant {
			out := chinese.Converts(c, v.Title, v.TypeName, v.Desc)
			v.Title = out[v.Title]
			v.TypeName = out[v.TypeName]
			v.Desc = out[v.Desc]
		}
		si := &space.ArcItem{}
		if isPopularBadge {
			si.FromArc(arc, s.hotAids, s.c.Custom.UpArcHasShare, false, nil)
		} else {
			si.FromArc(arc, nil, s.c.Custom.UpArcHasShare, false, nil)
		}
		if !enableSeriesEpisodicButton(c) && (series.Meta.Category != seriesgrpc.SeriesLiveReplay) { // 并且不能是直播回放
			si.URI = fmt.Sprintf("bilibili://music/playlist/playpage/%d?page_type=5&oid=%d&avid=%d&desc=%d", seriesId, arc.Arc.Aid, arc.Arc.Aid, episodicButtonSortDesc)
		}
		res.Item = append(res.Item, si)
	}
	if len(res.Item) > 0 {
		res.EpisodicButton = &space.EpisodicButton{
			Text: "播放全部",
			Uri:  fmt.Sprintf("bilibili://music/playlist/playpage/%d?page_type=5&desc=%d", seriesId, episodicButtonSortDesc),
		}
		res.Order = constructSeriesSeasonArcListOrder()
	}
	return res, nil
}

func constructSeriesSeasonArcListOrder() []*space.ArcOrder {
	return []*space.ArcOrder{{Title: "默认", Value: "desc"}, {Title: "倒序", Value: "asc"}}
}

func enableSeriesEpisodicButton(ctx context.Context) bool {
	dev, _ := device.FromContext(ctx)
	plat := platng.Plat(dev.RawMobiApp, dev.Device)
	return platng.IsIPad(plat)
}

// UpSeasons get upload ugc_season .
func (s *Service) UpSeasons(c context.Context, uid, pn, ps int64, build int, now time.Time, plat int8, mobiApp string) (res *space.UGCSeasonList) {
	var (
		ugcSeasons *ugcSeasonGrpc.UpperListReply
		err        error
	)
	res = &space.UGCSeasonList{Item: []*space.UGCSeasonItem{}}
	if ugcSeasons, err = s.ugcSeasonDao.UpperList(c, uid, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	if ugcSeasons == nil {
		return
	}
	res.Count = ugcSeasons.GetTotalCount()
	if len(ugcSeasons.GetSeasons()) != 0 {
		res.Item = make([]*space.UGCSeasonItem, 0, len(ugcSeasons.GetSeasons()))
		for _, v := range ugcSeasons.GetSeasons() {
			si := &space.UGCSeasonItem{}
			si.FromUGCSeason(v)
			res.Item = append(res.Item, si)
		}
	}
	sort.SliceStable(res.Item, func(i, j int) bool {
		// 付费合集之间最近更新的在前
		if res.Item[i].IsPay && res.Item[j].IsPay {
			return res.Item[i].MTime > res.Item[j].MTime
		}
		// 付费合集在前
		return res.Item[i].IsPay
	})
	return
}

func (s *Service) SeasonArchiveList(ctx context.Context, isHant bool, params *space.SeasonArchiveParam) (*space.SeasonArchiveResp, error) {
	reply, err := s.ugcSeasonDao.SeasonView(ctx, &ugcSeasonGrpc.ViewRequest{SeasonID: params.SeasonId})
	if err != nil {
		return nil, err
	}
	// 获取指定合集下所有aid
	var playAvs []*api.PlayAv
	for _, v := range reply.View.Sections {
		for _, episode := range v.Episodes {
			playAvs = append(playAvs, &api.PlayAv{Aid: episode.Aid})
		}
	}
	// 获取稿件信息
	apm, err := s.arcDao.ArcsPlayer(ctx, playAvs, false)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var arcs []*api.ArcPlayer
	for _, playAv := range playAvs {
		arc, ok := apm[playAv.Aid]
		if !ok {
			continue
		}
		arcs = append(arcs, arc)
	}
	// 返回结果
	res := &space.SeasonArchiveResp{}
	for _, arc := range arcs {
		v := arc.Arc
		if v == nil || !v.IsNormal() {
			continue
		}
		if isHant {
			out := chinese.Converts(ctx, v.Title, v.TypeName, v.Desc)
			v.Title = out[v.Title]
			v.TypeName = out[v.TypeName]
			v.Desc = out[v.Desc]
		}
		si := &space.ArcItem{}
		si.FromArc(arc, nil, s.c.Custom.UpArcHasShare, false, nil)
		res.Item = append(res.Item, si)
	}
	// 内存排序倒序
	episodicButtonSortDesc := 1
	if params.Sort == "asc" {
		episodicButtonSortDesc = 0
		reverseArcListItem(res.Item)
	}
	// 播放全部按钮
	res.EpisodicButton = &space.EpisodicButton{
		Text: "播放全部",
		Uri:  fmt.Sprintf("bilibili://music/playlist/playpage/%d?page_type=8&desc=%d", reply.View.Season.ID, episodicButtonSortDesc),
	}
	if reply.View.Season.AttrVal(ugcSeasonGrpc.AttrSnType) == 0 && len(res.Item) > 0 {
		// 精品合集跳第一个视频的详情页
		res.EpisodicButton.Uri = res.Item[0].URI
	}
	res.Order = constructSeriesSeasonArcListOrder()
	return res, nil
}

// UpComics get upload comic .
func (s *Service) UpComics(c context.Context, uid int64, pn, ps, build int, now time.Time, plat int8, labelTime bool) (res *space.ComicList) {
	var (
		comics *comic.Comics
		err    error
	)
	res = &space.ComicList{Item: []*space.ComicItem{}}
	if comics, err = s.comicDao.UpComics(c, uid, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	res.Count = comics.Total
	if len(comics.ComicList) != 0 {
		res.Item = make([]*space.ComicItem, 0, len(comics.ComicList))
		for _, v := range comics.ComicList {
			si := &space.ComicItem{}
			si.FromComic(v, labelTime)
			res.Item = append(res.Item, si)
		}
	}
	return
}

// SubComics get sub comic
func (s *Service) SubComics(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps, build int, plat int8) (res *space.SubComicList) {
	var (
		comics []*comic.FavComic
		err    error
	)
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.Comic != 1 {
			return
		}
	}
	res = &space.SubComicList{Item: []*space.SubComicItem{}}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if comics, err = s.comicDao.FavComics(ctx, vmid, pn, ps); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	g.Go(func() (err error) {
		if res.Count, err = s.comicDao.FavComicsCount(ctx, vmid); err != nil {
			log.Error("%v", err)
			err = nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, comic := range comics {
		if comic == nil {
			continue
		}
		ssi := &space.SubComicItem{}
		ssi.FormSubComic(comic)
		res.Item = append(res.Item, ssi)
	}
	return
}

// UpArticles get article.
func (s *Service) UpArticles(c context.Context, uid int64, pn, ps int) (res *space.ArticleList) {
	res = &space.ArticleList{Item: []*space.ArticleItem{}, Lists: []*article.List{}}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		var ams []*article.Meta
		if ams, res.Count, err = s.artDao.UpArticles(ctx, uid, pn, ps); err != nil {
			return err
		}
		if len(ams) != 0 {
			res.Item = make([]*space.ArticleItem, 0, len(ams))
			for _, v := range ams {
				if v.AttrVal(article.AttrBitNoDistribute) {
					continue
				}
				si := &space.ArticleItem{}
				si.FromArticle(ctx, v)
				res.Item = append(res.Item, si)
			}
		}
		return err
	})
	g.Go(func() (err error) {
		var lists []*article.List
		lists, res.ListsCount, err = s.artDao.UpLists(ctx, uid)
		if err != nil {
			return err
		}
		if len(lists) > 0 {
			res.Lists = lists
		}
		return err
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// favFolders get favorite folders
func (s *Service) favFolders(c context.Context, mid, vmid int64, setting *space.Setting, plat int8, build int, mobiApp string) (res *space.FavList) {
	const (
		_oldAndroidBuild = 427100
		_oldIOSBuild     = 3910
	)
	var (
		fs  []*favorite.Folder
		err error
	)
	res = &space.FavList{Item: []*favorite.Folder{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.FavVideo != 1 {
			return
		}
	}
	var mediaList bool
	// 双端版本号限制，符合此条件显示为“默认收藏夹”：
	// iPhone <5.36.1(8300) 或iPhone>5.36.1(8300)
	// Android <5360001或Android>5361000
	// 双端版本号限制，符合此条件显示为“默认播单”：
	// iPhone=5.36.1(8300)
	// 5360001 <=Android <=5361000
	if (plat == model.PlatIPhone && build == 8300) || (plat == model.PlatAndroid && build >= 5360001 && build <= 5361000) {
		mediaList = true
	}
	if fs, err = s.favDao.Folders(c, mid, vmid, mobiApp, build, mediaList); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range fs {
		if ((plat == model.PlatAndroid || plat == model.PlatAndroidG) && build <= _oldAndroidBuild) || ((plat == model.PlatIPhone || plat == model.PlatIPhoneI) && build <= _oldIOSBuild) {
			v.Videos = v.Cover
			v.Cover = nil
		}
	}
	res.Item = fs
	res.Count = len(fs)
	return
}

// GuardList .
func (s *Service) GuardList(c context.Context, mid, vmid int64, pn, ps int) (res *space.Guard) {
	//获取大航海信息
	list, err := s.guardDao.GetTopListGuardAttr(c, vmid, int64(pn), int64(ps), []string{"LEVEL", "FOLLOW_TIME"})
	if err != nil || list == nil || len(list.List) == 0 {
		return
	}
	res = &space.Guard{
		URI:       fmt.Sprintf("https://live.bilibili.com/p/html/live-app-guard-info/index.html?is_live_webview=1&hybrid_set_header=2&data_behavior_id=guard_main_space&uid=%d", vmid),
		Desc:      fmt.Sprintf("%d人加入大航海", list.TargetInfo.GuardCount),
		HighLight: "大航海",
	}
	for _, v := range list.List {
		res.Item = append(res.Item, &space.GuardList{Mid: v.Uid, Face: v.Face})
	}
	if mid == vmid {
		res.ButtonMsg = "我的大航海"
	} else {
		res.ButtonMsg = "Ta的大航海"
	}
	return
}

// favFolders2 get new fav.
func (s *Service) favFolders2(c context.Context, mid, vmid int64, setting *space.Setting, _ int8, _ int, mobiApp, device string) (res *space.FavList2, mediaCount int) {
	var (
		err error
		uf  []*favmdl.Folder
		mf  *favorite.Favorites
	)
	res = &space.FavList2{Item: []*favorite.Folder2{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.FavVideo != 1 {
			return
		}
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() error {
		var err error
		uf, err = s.favDao.UserFolders(ctx, favorite.TypeVideo, mid, vmid)
		if err != nil {
			log.Error("%v", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		mf, err = s.favDao.FavoritesRPC(ctx, favorite.TypeMediaList, mid, vmid, 0)
		if err != nil {
			log.Error("%v", err)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	var userFavCount = len(uf)
	//nolint:gomnd
	if userFavCount > 2 {
		uf = uf[:2]
	}
	faids := make(map[int64][]*favmdl.Resource, len(uf))
	for _, f := range uf {
		faids[f.ID] = f.RecentRes
	}
	var covers map[int64]*favorite.Cover
	if covers, err = s.FavCovers(c, faids, mobiApp, device, mid); err != nil {
		log.Error("s.FavCovers(%v) error(%v)", faids, err)
		//nolint:ineffassign
		err = nil
	}
	for _, f := range uf {
		if f == nil {
			continue
		}
		i := &favorite.Folder2{}
		i.FormFav(f)
		if cover, ok := covers[f.ID]; ok && cover != nil {
			i.Type = cover.Type
			if i.Cover == "" {
				i.Cover = cover.Pic
			}
		}
		res.Item = append(res.Item, i)
	}
	if mf != nil {
		mediaCount = mf.Page.Count
	}
	res.Count = userFavCount + mediaCount
	return
}

// Bangumi get concern season
func (s *Service) Bangumi(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps int) (res *space.BangumiList) {
	var (
		followReply *pgcappcard.FollowReply
		err         error
	)
	res = &space.BangumiList{Item: []*space.BangumiItem{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.Bangumi != 1 {
			return
		}
	}
	if followReply, err = s.bgmDao.Concern(c, mid, vmid, pn, ps); err != nil {
		log.Error("s.bgmDao.Concern err=%+v", err)
		return
	}
	if len(followReply.Seasons) != 0 {
		res.Item = make([]*space.BangumiItem, 0, len(followReply.Seasons))
		for _, v := range followReply.Seasons {
			si := &space.BangumiItem{}
			si.FromSeason(v)
			res.Item = append(res.Item, si)
		}
	}
	res.Count = int(followReply.Total)
	return
}

// Community get community
func (s *Service) Community(c context.Context, uid int64, pn, ps int, ak, platform string) (res *space.CommuList) {
	var (
		comm []*community.Community
		err  error
	)
	res = &space.CommuList{Item: []*space.CommItem{}}
	if comm, res.Count, err = s.commDao.Community(c, uid, ak, platform, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	if len(comm) != 0 {
		res.Item = make([]*space.CommItem, 0, len(comm))
		for _, v := range comm {
			si := &space.CommItem{}
			si.FromCommunity(v)
			res.Item = append(res.Item, si)
		}
	}
	return
}

// CoinCancel .
func (s *Service) CoinCancel(c context.Context, aid, mid int64) (err error) {
	return s.coinDao.UpMemberState(c, aid, mid, _businessLike)
}

// PlayGames .
func (s *Service) PlayGames(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps int, platform string) (res *space.GameList) {
	var (
		err     error
		rlyGame *gmdl.RecentGame
	)
	res = &space.GameList{Item: []*space.GameItem{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("PlayGames %+v", err)
				return
			}
		}
		if setting.PlayedGame != 1 {
			return
		}
	}
	if rlyGame, err = s.gameDao.RecentGame(c, vmid, pn, ps, platform); err != nil {
		log.Error("s.gameDao.RecentGame %+v", err)
		return
	}
	if rlyGame != nil && len(rlyGame.List) > 0 {
		res.Count = rlyGame.TotalCount
		for _, v := range rlyGame.List {
			si := &space.GameItem{}
			si.FromGame(v)
			res.Item = append(res.Item, si)
		}
	}
	return
}

// PlayGamesSub .
func (s *Service) PlayGamesSub(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps int, platform string) (res *space.GameListSub) {
	var (
		err     error
		rlyGame *gmdl.RecentGameSub
	)
	res = &space.GameListSub{Item: []*space.GameItemSub{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("PlayGamesSub %+v", err)
				return
			}
		}
		if setting.PlayedGame != 1 {
			return
		}
	}
	res.Image = s.c.SpaceGame.Image
	res.Uri = s.c.SpaceGame.JumpUri
	if rlyGame, err = s.gameDao.RecentGameSub(c, vmid, pn, ps, platform); err != nil {
		log.Error("s.gameDao.RecentGameSub %+v", err)
		return
	}
	if rlyGame != nil && len(rlyGame.List) > 0 {
		res.Count = rlyGame.TotalCount
		for _, v := range rlyGame.List {
			si := &space.GameItemSub{}
			si.FromGameSub(v)
			res.Item = append(res.Item, si)
		}
	}
	return
}

// CoinArcs get coin archives.
func (s *Service) CoinArcs(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps int, isHant bool, mobiApp, device string) (res *space.ArcList) {
	var (
		coins   []*api.Arc
		coinEps map[int64]*pgccardgrpc.EpisodeCard
		err     error
	)
	res = &space.ArcList{Item: []*space.ArcItem{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.CoinsVideo != 1 {
			return
		}
	}
	if coins, coinEps, res.Count, err = s.coinDao.CoinList(c, vmid, pn, ps, mobiApp, device); err != nil {
		log.Error("CoinArcs s.coinDao.CoinList(%d) error(%+v)", vmid, err)
		return
	}
	if len(coins) != 0 {
		res.Item = make([]*space.ArcItem, 0, len(coins))
		for _, v := range coins {
			si := &space.ArcItem{}
			si.FromCoinArc(v)
			ep, ok := coinEps[v.Aid]
			if ok {
				si.ConvertAsOGVEP(ep)
			}
			if isHant {
				out := chinese.Converts(c, si.Title, si.TypeName)
				si.Title = out[si.Title]
				si.TypeName = out[si.TypeName]
			}
			res.Item = append(res.Item, si)
		}
	}
	return
}

// LikeArcs get like archives.
func (s *Service) LikeArcs(c context.Context, mid, vmid int64, setting *space.Setting, pn, ps int, isHant bool, mobiApp, device string) (res *space.ArcList) {
	var (
		likes []*thumbupgrpc.ItemRecord
		err   error
	)
	res = &space.ArcList{Item: []*space.ArcItem{}}
	if mid != vmid {
		if setting == nil {
			setting, err = s.spcDao.Setting(c, vmid)
			if err != nil {
				log.Error("%+v", err)
				return
			}
		}
		if setting.LikesVideo != 1 {
			return
		}
	}
	if likes, res.Count, err = s.thumbupDao.UserTotalLike(c, vmid, _businessLike, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	if len(likes) != 0 {
		aids := make([]int64, 0, len(likes))
		for _, v := range likes {
			aids = append(aids, v.MessageID)
		}

		eg := errgroupv2.WithContext(c)
		var as map[int64]*api.Arc
		eg.Go(func(ctx context.Context) (err error) {
			as, err = s.arcDao.Archives(ctx, aids, mobiApp, device, mid)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
		var eps map[int64]*pgccardgrpc.EpisodeCard
		eg.Go(func(ctx context.Context) (err error) {
			eps, err = s.bgmDao.EpCardsFromPgcByAids(ctx, aids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})
		if err := eg.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		if len(as) == 0 {
			return
		}
		res.Item = make([]*space.ArcItem, 0, len(as))
		for _, v := range likes {
			if a, ok := as[v.MessageID]; ok {
				si := &space.ArcItem{}
				si.FromLikeArc(a)
				ep, ok := eps[v.MessageID]
				if ok {
					si.ConvertAsOGVEP(ep)
				}
				if isHant {
					out := chinese.Converts(c, si.Title, si.TypeName)
					si.Title = out[si.Title]
					si.TypeName = out[si.TypeName]
				}
				res.Item = append(res.Item, si)
			}
		}
	}
	return
}

// card get card by mid, vmid or name.
func (s *Service) card(c context.Context, vmid, mid int64, name string, isHant bool, build int64, mobiApp, device string) (scard *space.Card, err error) {
	scard = &space.Card{}
	var (
		profile        *account.ProfileStatReply
		relationm      map[int64]*relationgrpc.InterrelationReply
		campusInfo     *campusapi.CampusInfoReply
		seniorGateRsp  *answergrpc.SeniorGateResp
		pickupEntrance *space.PickupEntrance
		nftID          string
		nftRegionInfo  *gallerygrpc.NFTRegion
		userLocation   *passportuser.UserActiveLocationReply
		mcnInfo        *space.McnInfo
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if vmid > 0 {
			profile, err = s.accDao.Profile3(ctx, vmid)
		} else if name != "" {
			profile, err = s.accDao.ProfileByName3(ctx, name)
		}
		if err != nil {
			err = errors.Wrapf(err, "%v,%v or profile.Profile is nil", vmid, name)
			return
		}
		var campusInfoErr error
		if profile.School != nil {
			if profile.School.SchoolId > 0 {
				campusInfo, campusInfoErr = s.schoolDao.CampusInfo(ctx, &campusapi.CampusInfoReq{
					CampusId: profile.School.SchoolId,
				})
				if campusInfoErr != nil {
					log.Error("Failed to get campus info: %+v: %+v", profile.School, campusInfoErr)
				}
			}
		}
		// 当访问用户佩戴nft头像时,获取对应nft_id
		// 获取对应face_icon
		if profile.Profile.FaceNftNew == 1 {
			req := &memberAPI.NFTBatchInfoReq{
				Mids:   []int64{vmid},
				Status: "inUsing",
				Source: "face",
			}
			reply, nftInfoErr := s.accDao.NFTBatchInfo(ctx, req)
			if nftInfoErr != nil {
				log.Error("s.accDao.NFTBatchInfo vmid=%d err=%+v", vmid, nftInfoErr)
				return err
			}
			var (
				info *memberAPI.NFTBatchInfoData
				ok   bool
			)
			if info, ok = reply.GetNftInfos()[strconv.FormatInt(vmid, 10)]; !ok {
				log.Warn("s.accDao.NFTBatchInfo info is empty vmid=%d", vmid)
				return err
			}
			nftID = info.GetNftId()
			var (
				nftRegionErr error
			)
			if nftRegionInfo, nftRegionErr = s.galleryDao.GetNFTRegion(ctx, nftID); nftRegionErr != nil {
				log.Error("s.galleryDao.GetNFTRegion vmid=%d err=%+v", vmid, nftRegionErr)
				return err
			}
			return err
		}
		return err
	})
	if vmid > 0 {
		g.Go(func() error {
			var err error
			if relationm, err = s.relDao.Interrelations(ctx, mid, []int64{vmid}); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
		g.Go(func() error {
			var answerErr error
			seniorGateRsp, answerErr = s.answerDao.SeniorGate(ctx, vmid, build, mobiApp, device)
			if answerErr != nil {
				log.Error("s.answerDao.SeniorGate error(%+v), vmid(%d)", answerErr, vmid)
				return nil
			}
			return nil
		})
		// 登陆的情况下判断是否下发IP属地
		// 增加港澳台版本的屏蔽
		if mid != 0 && !pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
			pd.IsOverseas()
		}).MustFinish() {
			// 主态的情况下开关打开,下发IP属地
			// 客态的情况下开关打开,下发IP属地
			if (vmid == mid && s.c.ActiveLocationSwitch.OwnerSwitch) || (vmid != mid && s.c.ActiveLocationSwitch.GuestSwitch) {
				g.Go(func() error {
					var locationErr error
					userLocation, locationErr = s.accDao.UserActiveLocation(ctx, &passportuser.MidReq{Mid: vmid})
					if locationErr != nil {
						log.Error("s.accDao.UserActiveLocation error(%+v), vmid(%d)", locationErr, vmid)
						return nil
					}
					return nil
				})
			}
		}
	}
	if mid != 0 {
		g.Go(func() error {
			res, err := s.adDao.PickupEntrance(ctx, mid, vmid)
			if err != nil {
				log.Error("s.adDao.PickupEntrance error(%+v), vmid(%d)", err, vmid)
				return nil
			}
			if res != nil {
				pickupEntrance = &space.PickupEntrance{
					JumpUrl:        res.JumpUrl,
					IsShowEntrance: res.IsShowEntrance,
					Icon:           _PickupEntranceIcon,
				}
			}
			return nil
		})
		if s.c.MCNTagSwitch.Switch && pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
			pd.IsPlatAndroid().And().Build(">=", 6790000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPhone().And().Build(">=", 67900000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPad().And().Build(">=", 68200000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPadHD().And().Build(">=", 34600000)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatAndroidHD().And().Build(">=", 1230000)
		}).MustFinish() {
			g.Go(func() error {
				req := &memberAPI.GetUserExtraBasedOnKeyReq{
					Mid:  vmid,
					Keys: []string{"live_mcn_info", "video_mcn_info"},
				}
				res, err := s.accDao.GetUserExtraBasedOnKeys(ctx, req)
				if err != nil {
					log.Error("s.accDao.GetUserExtraBasedOnKeys error(%+v), vmid(%d)", err, vmid)
					return nil
				}
				if res != nil {
					mcnInfo = &space.McnInfo{}
					mcnInfo.FromUserExtraValues(res)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	if profile == nil || profile.Profile == nil {
		err = ecode.NothingFound
		log.Error("profile is null")
		return
	}
	scard = &space.Card{}
	scard.Mid = strconv.FormatInt(profile.Profile.Mid, 10)
	scard.Name = profile.Profile.Name
	scard.Sex = profile.Profile.Sex
	scard.IsFakeAccount = profile.Profile.IsFakeAccount
	scard.Face = profile.Profile.Face
	scard.FaceNftNew = profile.Profile.FaceNftNew // 用户头像是否是nft头像
	scard.Description = profile.Profile.Official.Desc
	scard.Fans = int(profile.Follower)
	scard.Attention = int(profile.Following)
	scard.NftId = nftID
	if nftRegionInfo != nil {
		scard.NftFaceIcon = &space.NftFaceIcon{
			RegionType: int32(nftRegionInfo.Type),
			Icon:       nftRegionInfo.Icon,
			ShowStatus: int32(nftRegionInfo.ShowStatus),
		}
	}
	// 回粉
	scard.Relation = space.RelationChange(vmid, relationm)
	scard.Sign = profile.Profile.Sign
	scard.LevelInfo.Cur = profile.Profile.Level
	scard.LevelInfo.Min = profile.LevelInfo.Min
	scard.LevelInfo.NowExp = profile.LevelInfo.NowExp
	scard.LevelInfo.NextExp = profile.LevelInfo.NextExp
	if seniorGateRsp != nil {
		scard.LevelInfo.Identity = int64(seniorGateRsp.Member)
		if vmid == mid { // 主态下发出题入口
			scard.LevelInfo.SeniorInquiry.InquiryText = seniorGateRsp.InquiryText
			scard.LevelInfo.SeniorInquiry.InquiryUrl = seniorGateRsp.InquiryUrl
		}
	}
	if profile.LevelInfo.NextExp == -1 {
		scard.LevelInfo.NextExp = "--"
	}
	scard.Pendant = profile.Profile.GetPendant()
	scard.Nameplate = profile.Profile.GetNameplate()
	scard.OfficialVerify.Role = profile.Profile.Official.Role
	scard.OfficialVerify.Type = int8(profile.Profile.Official.Type)
	scard.OfficialVerify.Title = profile.Profile.Official.Title
	if isHant {
		scard.OfficialVerify.Title = chinese.Convert(c, scard.OfficialVerify.Title)
	}
	scard.Vip.Type = int(profile.Profile.Vip.Type)
	if profile.Profile.Official.Role != 0 {
		scard.OfficialVerify.Desc = profile.Profile.Official.Title
	}
	scard.Vip.VipStatus = int(profile.Profile.Vip.Status)
	scard.Vip.DueDate = profile.Profile.Vip.DueDate
	scard.Vip.ThemeType = int(profile.Profile.Vip.ThemeType)
	scard.Vip.Label = space.FromVipLabelToVipSpaceLabel(profile.Profile.Vip.Label, isHant)
	if vmid == mid && profile.Profile.Vip.Status == model.VipStatusExpire && profile.Profile.Vip.DueDate > 0 { //空间主人态替换过期铭牌
		scard.Vip.Label.Path = model.VipLabelExpire
	}
	if profile.GetProfile().GetSilence() == _accountBlocked {
		scard.Silence = profile.Profile.Silence
	}
	if isDeleted := profile.GetProfile().GetIsDeleted(); isDeleted == _accountIsDeleted { // 是否是注销账号
		scard.IsDeleted = isDeleted
	}
	if profile.GetUserHonourInfo() != nil {
		if profile.UserHonourInfo.GetColour() != nil {
			scard.Honours.Colour.Dark = profile.UserHonourInfo.Colour.Dark
			scard.Honours.Colour.Normal = profile.UserHonourInfo.Colour.Normal
		}
		scard.Honours.Tags = filterHonorTag(profile.UserHonourInfo.Tags)
	}
	// 支持职业信息展示
	scard.Profession.Id = profile.Profile.Profession.Id
	scard.Profession.Name = profile.Profile.Profession.Name
	scard.Profession.ShowName = profile.Profile.Profession.ShowName
	if professionVerify, ok := makeSpaceProfessionVerify(profile.Profile.Profession); ok {
		scard.ProfessionVerify = professionVerify
	}
	//支持校园信息展示
	scard.School.SchoolId = profile.School.GetSchoolId()
	scard.School.Name = profile.School.GetName()
	scard.SpaceTag = userSpaceTag(c, profile, vmid == mid, campusInfo, userLocation, mcnInfo)
	// 空间花火入口
	scard.PickupEntrance = pickupEntrance
	// 空间底部tag
	scard.SpaceTagBottom = spaceTagBottom(c, userLocation, mcnInfo)
	return scard, nil
}

func makeSpaceProfessionVerify(profession account.Profession) (space.ProfessionVerify, bool) {
	const (
		_professionIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/fwuEQFnjVO.png"
	)
	if profession.IsShow != 1 {
		return space.ProfessionVerify{}, false
	}
	if profession.Title != "" && profession.Department != "" {
		return space.ProfessionVerify{
			Icon:     _professionIcon,
			ShowDesc: fmt.Sprintf("职业资质：%s %s", profession.Title, profession.Department),
		}, true
	}
	if profession.Name != "" {
		return space.ProfessionVerify{
			Icon:     _professionIcon,
			ShowDesc: fmt.Sprintf("职业资质：%s", profession.Name),
		}, true
	}
	return space.ProfessionVerify{}, false
}

func filterHonorTag(in []*account.HonourTag) []*account.HonourTag {
	out := make([]*account.HonourTag, 0, len(in))
	for _, t := range in {
		scene := sets.NewString(t.Scene...)
		if scene.Has("space") {
			out = append(out, t)
		}
	}
	return out
}

func asHighlight(ctx context.Context, in *space.SpaceTag) {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsAndroidAll().And().Build(">=", int64(6520000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsIOSAll().And().Build(">=", int64(65200000))
	}).MustFinish() {
		in.TextColor = "#61666D"
		in.NightTextColor = "#A2A7AE"
		in.BackgroundColor = "#F6F7F8"
		in.NightBackgroundColor = "#0D0D0E"
		return
	}
	in.TextColor = "#FF6699"
	in.NightTextColor = "#D44E7D"
	in.BackgroundColor = "#FFECF1"
	in.NightBackgroundColor = "#2F1A22"
}

func asGray(ctx context.Context, in *space.SpaceTag) {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsAndroidAll().And().Build(">=", int64(6520000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsIOSAll().And().Build(">=", int64(65200000))
	}).MustFinish() {
		in.TextColor = "#9499A0"
		in.NightTextColor = "#757A81"
		in.BackgroundColor = "#FFFFFF"
		in.NightBackgroundColor = "#17181A"
		return
	}
	in.TextColor = "#484C53"
	in.NightTextColor = "#B9BDC2"
	in.BackgroundColor = "#F1F2F3"
	in.NightBackgroundColor = "#000000"
}

func fillSpaceTagColor(ctx context.Context, in *space.SpaceTag) {
	switch in.Type {
	// 粉色标
	case "honour", "school", "bbq", "mcn_info":
		asHighlight(ctx, in)
		if in.Type == "school" && in.URI == "" {
			asGray(ctx, in)
			return
		}
	// 灰色标
	case "profession":
		asGray(ctx, in)
	// 灰色没跳转标
	case "location":
		asNoJumpGray(in)
	// 默认按灰色标来
	default:
		asGray(ctx, in)
	}
}

func userSpaceTag(ctx context.Context, profile *account.ProfileStatReply, isOwner bool, campusInfo *campusapi.CampusInfoReply, userLocation *passportuser.UserActiveLocationReply, mcnInfo *space.McnInfo) []*space.SpaceTag {
	out := []*space.SpaceTag{}
	if profile.UserHonourInfo != nil {
		for _, h := range profile.UserHonourInfo.Tags {
			t := &space.SpaceTag{
				Type:  "honour",
				Title: h.Name,
				URI:   h.Link,
				Icon:  h.Icon,
			}
			fillSpaceTagColor(ctx, t)
			scene := sets.NewString(h.Scene...)
			if scene.Has("space") {
				out = append(out, t)
			}
		}
	}
	// 校园和添加校园标
	func() {
		if profile.School != nil {
			if profile.School.SchoolId > 0 {
				t := &space.SpaceTag{
					Type:  "school",
					Title: profile.School.Name,
					Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/PDqKDoER9Q.png",
				}
				if campusInfo != nil {
					if campusInfo.Online == 1 {
						t.URI = fmt.Sprintf("bilibili://campus/detail/%d", campusInfo.CampusId)
					}
				}
				fillSpaceTagColor(ctx, t)
				out = append(out, t)
				return
			}
		}
		if isOwner {
			t := &space.SpaceTag{
				Type:  "submit_school",
				Title: "添加学校信息",
				URI:   "https://www.bilibili.com/h5/school/edit?navhide=1",
			}
			out = append(out, t)
		}
	}()

	if canShowProfessionSpaceTag(ctx, profile) {
		t := &space.SpaceTag{
			Type:  "profession",
			Title: profile.Profile.Profession.ShowName,
			Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/nRfolXG7RQ.png",
		}
		fillSpaceTagColor(ctx, t)
		out = append(out, t)
	}
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build("<", 6860000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build("<", 68600000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build("<", 68800000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", 34900000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build("<", 1270000)
	}).MustFinish() {
		// ip属地标签
		func() {
			if userLocation != nil {
				if userLocation.Location == "" || userLocation.Location == "hide" {
					return
				}
				t := &space.SpaceTag{
					Type:  "location",
					Title: fmt.Sprintf("IP属地：%s", userLocation.Location),
					Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/fAILMRg9PS.png",
				}
				fillSpaceTagColor(ctx, t)
				out = append(out, t)
				return
			}
		}()
		// mcn机构展示信息
		func() {
			if mcnInfo != nil && mcnInfo.Name != "" {
				t := &space.SpaceTag{
					Type:  "mcn_info",
					Title: mcnInfo.Name,
					URI:   mcnInfo.Url,
					Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/MKBhpAJCsf.png",
				}
				fillSpaceTagColor(ctx, t)
				out = append(out, t)
				return
			}
		}()
	}
	return sortUserSpaceTag(out)
}

func canShowProfessionSpaceTag(ctx context.Context, profile *account.ProfileStatReply) bool {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidB().Or().IsPlatAndroidI().And().Build(">=", int64(6600000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().Or().IsPlatIPhoneB().Or().IsPlatIPhoneI().Or().IsPlatIPad().And().Build(">=", int64(66000000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">=", int64(33600000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build(">=", int64(1140000))
	}).MustFinish() {
		// 新版本不出职业相关的space tag
		return false
	}
	return profile.Profile.Profession.IsShow == 1 && profile.Profile.Profession.Id != 0
}

func sortUserSpaceTag(tags []*space.SpaceTag) []*space.SpaceTag {
	typeOrder := map[string]int64{
		"honour":        1,
		"bbq":           2,
		"school":        3,
		"profession":    4,
		"submit_school": 5,
		"mcn_info":      6,
		"location":      7,
	}
	sort.Slice(tags, func(i, j int) bool {
		pi, ok := typeOrder[tags[i].Type]
		if !ok {
			pi = math.MaxInt64
		}
		pj, ok := typeOrder[tags[j].Type]
		if !ok {
			pj = math.MaxInt64
		}
		return pi < pj
	})
	return tags
}

// audios
func (s *Service) audios(c context.Context, mid int64, pn, ps int) (res *space.AudioList) {
	var (
		audios []*audio.Audio
		err    error
	)
	res = &space.AudioList{Item: []*space.AudioItem{}}
	if audios, res.Count, err = s.audioDao.Audios(c, mid, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	if len(audios) != 0 {
		res.Item = make([]*space.AudioItem, 0, len(audios))
		for _, v := range audios {
			si := &space.AudioItem{}
			si.FromAudio(v)
			res.Item = append(res.Item, si)
		}
	}
	return
}

// Report func
func (s *Service) Report(c context.Context, mid int64, reason, ak string) (err error) {
	return s.spcDao.Report(c, mid, reason, ak)
}

// FavCovers get cover of each fid
func (s *Service) FavCovers(c context.Context, recents map[int64][]*favmdl.Resource, mobiApp, device string, mid int64) (fcvs map[int64]*favorite.Cover, err error) {
	var (
		fids, avids, musicids []int64
		ogvIds                []int32
		avm                   map[int64]*api.Arc
		mm                    map[int64]*audio.Music
		ogvEpcardsm           map[int32]*pgccardgrpc.EpisodeCard
	)
	for fid := range recents {
		fids = append(fids, fid)
	}
	for _, fid := range fids {
		if resources, ok := recents[fid]; ok {
			for _, res := range resources {
				switch int8(res.Typ) {
				case favorite.TypeVideo:
					avids = append(avids, res.Oid)
				case favorite.TypeMusicNew:
					musicids = append(musicids, res.Oid)
				case favorite.TypeOgv:
					ogvIds = append(ogvIds, int32(res.Oid))
				default:
					log.Error("FavCovers: unexpected type from resource %+v", res)
				}
				//nolint:staticcheck
				break // olny need the first oid.
			}
		}
	}
	fcvs = make(map[int64]*favorite.Cover)
	g, ctx := errgroup.WithContext(c)
	if len(avids) != 0 {
		g.Go(func() (err error) {
			if avm, err = s.arcDao.Archives(ctx, avids, mobiApp, device, mid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(musicids) != 0 {
		g.Go(func() (err error) {
			if mm, err = s.audioDao.MusicMap(ctx, musicids); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
	}
	if len(ogvIds) != 0 {
		g.Go(func() (err error) {
			if ogvEpcardsm, err = s.bgmDao.EpCardsFromPgcByEpids(ctx, ogvIds); err != nil {
				log.Error("s.bangumiDao.EpCardsFromPgcByEpids err(%+v)", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	// set miss fids cover
	for _, fid := range fids {
		cv := &favorite.Cover{}
		for _, res := range recents[fid] {
			cv.Aid = int(res.Oid)
			cv.Type = res.Typ
			switch int8(res.Typ) {
			case favorite.TypeVideo:
				if arc, ok := avm[res.Oid]; ok {
					if !arc.IsNormal() {
						continue
					}
					cv.Pic = arc.Pic
				}
			case favorite.TypeMusicNew:
				if music, ok := mm[res.Oid]; ok {
					cv.Pic = music.Cover
				}
			case favorite.TypeOgv:
				if ogv, ok := ogvEpcardsm[int32(res.Oid)]; ok {
					cv.Pic = ogv.Cover
				}
			default:
				log.Error("FavCovers: unexpected type from resource %+v", res)
			}
			break
		}
		fcvs[fid] = cv
	}
	return
}

// UpCheese get up pugv
func (s *Service) UpCheese(c context.Context, vmid int64, pn, ps int) (res *space.CheeseList) {
	var (
		seasons []*cheeseGRPC.SeasonCard
		total   int64
		err     error
	)
	res = &space.CheeseList{Item: []*space.CheeseItem{}}
	if seasons, total, err = s.cheeseDao.UserSeason(c, vmid, pn, ps); err != nil || len(seasons) == 0 {
		log.Error("s.cheeseDao.UserSeason err(%+v) or len(season)=0", err)
		return
	}
	res.Count = total
	res.Item = make([]*space.CheeseItem, 0, len(seasons))
	for _, v := range seasons {
		si := &space.CheeseItem{}
		si.FromCheese(v)
		res.Item = append(res.Item, si)
	}
	return
}

// AttentionMark .
func (s *Service) AttentionMark(c context.Context, mid int64) (err error) {
	if err = s.teenDao.IncrCacheAttention(c, mid); err != nil {
		log.Error("s.teenDao.IncrCacheAttention(%d) error(%v)", mid, err)
		// 错误不抛出
		err = nil
	}
	return
}

func (s *Service) PhotoMallList(c context.Context, params *space.PhotoTopParm, mid int64) (res *space.PhotoMall, err error) {
	var (
		list         []*space.PhotoMallItem
		equip        *garbgrpc.SpaceBGUserEquipReply
		character    *live2dgrpc.GetUserSpaceCharacterInfoResp
		digitalReply *digitalgrpc.GetGarbSpaceEntryResp
	)
	g := errgroupv2.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		// 获取粉丝装扮的头图
		equip, err = s.garbDao.SpaceBGEquip(ctx, mid)
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		// 获取看板娘装扮状态
		character, err = s.garbDao.GetUserSpaceCharacterInfo(ctx, &live2dgrpc.GetUserSpaceCharacterInfoReq{Mid: mid})
		if err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		// 获取默认头图列表
		list, err = s.spcDao.PhotoMallList(ctx, params.MobiApp, params.Device, mid)
		if err != nil {
			log.Error("%+v", err)
		}
		return
	})
	g.Go(func(ctx context.Context) error {
		// 数字藏品入口
		var err error
		digitalReply, err = s.digitalDao.DigitalEntry(ctx, mid)
		if err != nil {
			log.Error("s.digitalDao.DigitalEntry  mid=%d, err=%+v", mid, err)
			return nil
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	res = &space.PhotoMall{
		List:  list,
		Title: s.c.Custom.PhotoMallTitle,
	}
	if blockDefaultDressActivated(equip, character, digitalReply) {
		for _, v := range list {
			v.IsActivated = 0 //强行取消已选中的状态
		}
	}
	return
}

func blockDefaultDressActivated(equip *garbgrpc.SpaceBGUserEquipReply, character *live2dgrpc.GetUserSpaceCharacterInfoResp, digital *digitalgrpc.GetGarbSpaceEntryResp) bool {
	if equip != nil && equip.Item != nil && len(equip.Item.Images) > int(equip.Index) && equip.Item.Images[equip.Index] != nil && equip.Item.Images[equip.Index].Landscape != "" {
		// 判断当前用户是否有粉丝装扮，有粉丝装扮就把 已选中的状态下掉
		return true
	}
	if character != nil && character.Active {
		// 判断当前用户是否有看板娘装扮，有粉丝装扮就把 已选中的状态下掉
		return true
	}
	if digital != nil && digital.Active {
		// 判断当前用户是否装扮数字藏品头图，有就将已选中的状态下掉
		return true
	}
	return false
}

func (s *Service) PhotoTopSet(c context.Context, params *space.PhotoTopParm, mid int64) (err error) {
	// 先卸下粉丝装扮和大会员头图
	// 增加数字藏品卸载
	if err = s.TopphotoReset(c, mid, params.AccessKey, params.Platform, params.Device, params.Type); err != nil {
		log.Error("%+v", err)
		err = nil
	}
	// 在装扮默认头图,视频头图
	if err = s.spcDao.PhotoTopSet(c, params.MobiApp, params.Oid, mid, params.Type); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (s *Service) PhotoArcList(ctx context.Context, mobiApp, platform, device, buvid string, mid int64, plat int8, build, pn, ps int) (*space.ArcList, error) {
	// 限制pgc视频,禁止推荐视频,互动视频,ugc付费
	attrForbid := int64(1<<api.AttrBitIsPGC) | int64(1<<api.AttrBitSteinsGate) | int64(1<<api.AttrBitUGCPay)
	res := new(space.ArcList)
	if (pn-1)*ps > s.c.Custom.PhotoArcCount {
		return res, nil
	}
	searchArchives, err := s.searchDao.Space(ctx, mobiApp, platform, device, "", "", space.ArchiveNew, "", buvid, plat, build, 0, 0, 0, pn, ps, mid, mid, attrForbid, time.Now())
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var aids []int64
	if searchArchives != nil {
		res.Count = searchArchives.Total
		if res.Count > s.c.Custom.PhotoArcCount {
			res.Count = s.c.Custom.PhotoArcCount
		}
		if searchArchives.Result != nil {
			for _, val := range searchArchives.Result.VList {
				aids = append(aids, val.Aid)
			}
		}
	}
	eg := errgroupv2.WithContext(ctx)
	var (
		arcsTmp          map[int64]*arcgrpc.ArcPlayer
		flowInfosV2Reply *cfcgrpc.FlowCtlInfosV2Reply
	)
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcsTmp, err = s.arcDao.Arcs(ctx, aids, mobiApp, device, mid)
			if err != nil {
				return err
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			req := makeContentFlowControlInfosV2Params(s.c.CfcSvrConfig, aids)
			flowInfosV2Reply, err = s.arcDao.ContentFlowControlInfosV2(ctx, req)
			if err != nil {
				log.Error("s.arcDao.ContentFlowControlInfosV2 err=%+v", err)
				return nil
			}
			return nil
		})
		if err := eg.Wait(); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		total := (pn - 1) * ps
		for _, aid := range aids {
			ap, ok := arcsTmp[aid]
			if !ok || ap == nil || ap.Arc == nil || !ap.Arc.IsNormal() {
				continue
			}
			if total >= s.c.Custom.PhotoArcCount {
				break
			}
			arc := ap.Arc
			if arc.AttrVal(api.AttrBitIsPGC) == api.AttrYes ||
				getArcAttrFromInfosV2(flowInfosV2Reply, arc.Aid, _arcAttrNoRecommend) ||
				arc.AttrVal(api.AttrBitSteinsGate) == api.AttrYes ||
				arc.AttrVal(api.AttrBitUGCPay) == api.AttrYes {
				continue
			}
			si := &space.ArcItem{}
			si.FromArc(ap, nil, false, false, nil)
			res.Item = append(res.Item, si)
			total++
		}
		if res.Count < ps {
			res.Count = len(res.Item)
		}
	}
	return res, nil
}

func makeContentFlowControlInfosV2Params(config *conf.CfcSvrConfig, aids []int64) *cfcgrpc.FlowCtlInfosReq {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", config.Source)
	params.Set("business_id", strconv.FormatInt(config.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	return &cfcgrpc.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(config.BusinessID),
		Source:     config.Source,
		Sign:       getFlowCtlInfosReqSign(params, config.Secret),
		Ts:         ts,
	}
}

func getFlowCtlInfosReqSign(params url.Values, secret string) string {
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

func getArcAttrFromInfosV2(reply *cfcgrpc.FlowCtlInfosV2Reply, aid int64, arcsAttrKey string) bool {
	if reply == nil {
		return false
	}
	val, ok := reply.ItemsMap[aid]
	if !ok {
		return false
	}
	for _, v := range val.Items {
		if v.Key == arcsAttrKey {
			return v.Value == 1
		}
	}
	return false
}

//nolint:gocognit
func (s *Service) createdActList(ctx context.Context, vmid int64, device string) []*space.CreatedActivity {
	topicList, err := s.adDao.CreatedTopicList(ctx, vmid)
	if err != nil {
		log.Error("Failed to get CreatedTopicList vmid: %d, error: %+v", vmid, err)
		return nil
	}
	var createdList []*space.CreatedActivity
	var topicIDs []int64
	for _, v := range topicList {
		if v == nil || v.TopicID <= 0 || v.TopicName == "" {
			continue
		}
		if v.TopicType == 1 {
			item := &space.CreatedActivity{
				Name:      fmt.Sprintf("#%s#", v.TopicName),
				TopicID:   v.TopicID,
				Cover:     v.CoverURL,
				CoverMd5:  v.CoverMd5,
				View:      v.View,
				Discuss:   v.Discuss,
				URI:       v.JumpUrl,
				TopicType: v.TopicType,
			}
			createdList = append(createdList, item)
			continue
		}
		topicIDs = append(topicIDs, v.TopicID)
	}
	if len(topicIDs) == 0 {
		return createdList
	}
	eg := errgroupv2.WithContext(ctx)
	var topicChs map[int64]*channelgrpc.Channel
	eg.Go(func(ctx context.Context) error {
		var chErr error
		topicChs, chErr = s.channelDao.ChannelInfos(ctx, topicIDs)
		if chErr != nil {
			return chErr
		}
		return nil
	})
	var topicStats map[int64]*dynamicgrpc.TopicStats
	eg.Go(func(ctx context.Context) error {
		var statErr error
		topicStats, statErr = s.bplusDao.TopicStats(ctx, topicIDs)
		if statErr != nil {
			log.Error("Failed to get topic stat vmid:%d topicIDs: %v, error: %+v", vmid, topicIDs, statErr)
		}
		return nil
	})
	var natPages map[int64]*actgrpc.NativePage
	eg.Go(func(ctx context.Context) error {
		var actErr error
		natPages, actErr = s.actDao.NatActInfo(ctx, topicIDs, 1, nil)
		if actErr != nil {
			log.Error("Failed to get nat act info vmid: %d topicIDs: %v, error: %+v", vmid, topicIDs, actErr)
		}
		return nil
	})
	if egErr := eg.Wait(); egErr != nil {
		log.Error("createdActList vmid:%d eg.Wait() error:%v", vmid, egErr)
		return createdList
	}
	for _, v := range topicList {
		if v == nil || v.TopicID <= 0 || v.TopicName == "" || v.TopicType == 1 {
			continue
		}
		channel, ok := topicChs[v.TopicID]
		if !ok || channel == nil || channel.State != 0 {
			continue
		}
		item := &space.CreatedActivity{
			Name:     fmt.Sprintf("#%s#", v.TopicName),
			TopicID:  v.TopicID,
			Cover:    v.CoverURL,
			CoverMd5: v.CoverMd5,
		}
		//status 0:显示 1:隐藏
		if stat, ok := topicStats[v.TopicID]; ok && stat != nil && stat.Status == 0 {
			item.View = stat.View
			item.Discuss = stat.Discuss
		}
		//nolint:gosimple
		natInfo, _ := natPages[v.TopicID]
		item.URI = func() string {
			if natInfo != nil {
				if natInfo.SkipURL != "" {
					return natInfo.SkipURL
				}
				if natInfo.ID > 0 {
					return fmt.Sprintf("bilibili://following/activity_landing/%d", natInfo.ID)
				}
			}
			// 新频道
			if channel.CType == 2 && device != "pad" {
				return fmt.Sprintf("bilibili://pegasus/channel/v2/%d?tab=topic", v.TopicID)
			}
			return fmt.Sprintf("bilibili://pegasus/channel/%d?type=topic", v.TopicID)
		}()
		createdList = append(createdList, item)
		if len(createdList) >= s.c.Custom.CreatedActCnt {
			break
		}
	}
	return createdList
}

func resolveSpaceSetting(sp *space.Space) {
	// 校园开关控制校园信息
	if sp.Setting == nil || sp.Setting.DisableShowSchool == 1 {
		sp.Card.School = space.School{}
		sp.Card.SpaceTag = removeSchoolTag(sp.Card.SpaceTag)
	}
}

func removeSchoolTag(in []*space.SpaceTag) []*space.SpaceTag {
	out := make([]*space.SpaceTag, 0, len(in))
	for _, t := range in {
		if t.Type == "school" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (s *Service) Reserve(ctx context.Context, req *space.AddReserveReq) (*space.ReserveClickResp, error) {
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		err := s.actDao.AddReserve(ctx, req)
		if err != nil {
			log.Error("s.actDao.AddReserve req(%+v), err(%+v)", req, err)
			return err
		}
		return nil
	})
	var (
		calList []*activitygrpc.ReserveCalendarInfo
	)
	eg.Go(func(ctx context.Context) error {
		calRly, e := s.actDao.GetReserveCalendarInfo(ctx, req.Sid)
		if e != nil { //错误可降级
			log.Error("s.actDao.GetReserveCalendarInfo req(%d) error(%+v)", req.Sid, e)
			return nil
		}
		if calRly != nil {
			calList = calRly.List
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	rly := constructReserveClickResp(req.ReserveTotal+1, nil)
	for _, v := range calList {
		if v == nil {
			continue
		}
		rly.CalendarInfos = append(rly.CalendarInfos, &space.CalendarInfo{
			Title:      v.CalendarTitle,
			STime:      v.LivePlanStartTime.Time().Unix(),
			ETime:      v.LivePlanStartTime.Time().Add(time.Hour).Unix(),
			Comment:    v.Comment,
			BusinessID: v.BusinessId,
		})
	}
	return rly, nil
}

func (s *Service) ReserveCancel(ctx context.Context, mid, sid, total int64) (*space.ReserveClickResp, error) {
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if err = s.actDao.DelReserve(ctx, mid, sid); err != nil {
			log.Error("s.actDao.DelReserve mid(%d), sid(%d), err(%+v)", mid, sid, err)
		}
		return
	})
	var busIDs []string
	eg.Go(func(ctx context.Context) error {
		calRly, e := s.actDao.GetReserveCalendarInfo(ctx, sid)
		if e != nil { //错误可降级
			log.Error("s.actDao.GetReserveCalendarInfo req(%d) error(%+v)", sid, e)
			return nil
		}
		if calRly == nil || len(calRly.List) == 0 {
			return nil
		}
		for _, v := range calRly.List {
			busIDs = append(busIDs, v.BusinessId)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return constructReserveClickResp(total-1, busIDs), nil
}

func constructReserveClickResp(total int64, busIDs []string) *space.ReserveClickResp {
	if total < 0 {
		return nil
	}
	return &space.ReserveClickResp{
		ReserveUpdate: total,
		DescUpdate:    fmt.Sprintf(" %s", model.StatNumberToString(total, "人预约")),
		BusinessIDs:   busIDs,
	}
}

func (s *Service) UpReserveCancel(ctx context.Context, mid int64, sid int64) error {
	err := s.actDao.CancelUpActReserve(ctx, mid, sid)
	if err != nil {
		log.Error("s.actDao.CancelUpActReserve mid(%d), sid(%d), err(%+v)", mid, sid, err)
		return err
	}
	return nil
}

func (s *Service) GetReserveDynShareContent(ctx context.Context, mid int64, req *space.ReserveShareInfoReq, dev device.Device) (*dynsharegrpc.GetReserveDynShareContentRsp, error) {
	args := &dynsharegrpc.GetReserveDynShareContentReq{
		Uid:   mid,
		DynId: req.DynId,
		Meta: &dynamicgrpc.MetaDataCtrl{
			Platform: dev.RawPlatform,
			Build:    strconv.FormatInt(dev.Build, 10),
			MobiApp:  dev.RawMobiApp,
			Buvid:    dev.Buvid,
			Device:   dev.Device,
			Network:  dev.NetworkType,
			Ip:       metadata.RemoteIP,
		},
		ShareId:   req.ShareId,
		ShareMode: req.ShareMode,
	}
	res, err := s.dynDao.GetReserveDynShareContent(ctx, args)
	if err != nil {
		log.Error("GetReserveDynShareContent() s.dynDao.GetReserveDynShareContent args=%+v, err=%+v", args, err)
		return nil, err
	}
	return res, nil
}

// 新版国际版可以用充电了
func enableOverseaElec(plat int8, build int) bool {
	if plat == model.PlatIPhoneI && build >= 63900000 {
		return true
	}
	return false
}

// 用户账号服务的nft_id替换盘古FaceJump的nft_id
func replaceFaceJumpNftID(nftID, faceJump string) string {
	if nftID == "" {
		log.Error("replaceFaceJumpNftID nftID is empty")
		return ""
	}
	u, err := url.Parse(faceJump)
	if err != nil {
		log.Error("replaceFaceJumpNftID err:%v", err)
		return ""
	}
	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Error("replaceFaceJumpNftID err:%v", err)
		return ""
	}
	params.Set("nftid", nftID)
	u.RawQuery = params.Encode()
	return u.String()
}

func asNoJumpGray(in *space.SpaceTag) {
	in.TextColor = "#61666D"
	in.NightTextColor = "#A2A7AE"
}

// 根据合集信息重新排序
func (s *Service) ArcReorderBySeason(ctx context.Context, arcPlayer []*api.ArcPlayer, arcPlayerCursor []*space.ArcPlayerCursor, order string, upMid int64) (*space.SeasonsRankInfo, error) {
	seasonArcMap, seasonIds := getSeasonArcInfo(arcPlayer, arcPlayerCursor, upMid)
	if len(seasonIds) == 0 {
		return nil, errors.New("seasonIds is empty")
	}
	req := &ugcSeasonGrpc.SeasonsRequest{
		SeasonIds: seasonIds,
	}
	reply, err := s.ugcSeasonDao.Seasons(ctx, req)
	if err != nil {
		log.Error("s.UpArcs Seasons error:%+v", err)
		return nil, err
	}
	seasons := make(map[int64]*ugcSeasonGrpc.Season)
	for sId, sn := range reply.GetSeasons() {
		if sn.AttrVal(ugcSeasonGrpc.AttrSnNoSpace) == ugcSeasonGrpc.AttrSnYes {
			// 防刷屏活动合集中,非合集创建人的空间稿件不受防刷逻辑影响
			if arc, ok := seasonArcMap[sId]; ok && sn.Mid == arc.Author.Mid {
				seasons[arc.Aid] = sn
			}
		}
	}
	if len(seasons) == 0 {
		return nil, errors.New("seasons is empty")
	}
	if arcPlayer != nil {
		arcPlayer = sortArcPlayerBySeasons(arcPlayer, seasons, order)
		res := &space.SeasonsRankInfo{
			Seasons:   seasons,
			ArcPlayer: arcPlayer,
		}
		return res, nil
	}
	arcPlayerCursor = sortArcPlayerCursorBySeasons(arcPlayerCursor, seasons, order)
	res := &space.SeasonsRankInfo{
		Seasons:         seasons,
		ArcPlayerCursor: arcPlayerCursor,
	}
	return res, nil
}

// 获取需要防刷屏合集显示优化的合集和稿件信息
func getSeasonArcInfo(arcPlayer []*api.ArcPlayer, arcPlayerCursor []*space.ArcPlayerCursor, upMid int64) (map[int64]*api.Arc, []int64) {
	var (
		seasonArcMap = make(map[int64]*api.Arc)
		seasonIds    []int64
	)
	if arcPlayer != nil {
		for _, v := range arcPlayer {
			// 主投稿人的空间稿件才需要参与显示优化
			if v.Arc.SeasonID != 0 && v.Arc.Author.Mid == upMid {
				seasonIds = append(seasonIds, v.Arc.SeasonID)
				// 防刷屏合集在主投稿人空间只有一个稿件
				seasonArcMap[v.Arc.SeasonID] = v.Arc
			}
		}
		return seasonArcMap, seasonIds
	}
	for _, v := range arcPlayerCursor {
		// 主投稿人的空间稿件才需要参与显示优化
		if v.Arc.SeasonID != 0 && v.Arc.Author.Mid == upMid {
			seasonIds = append(seasonIds, v.Arc.SeasonID)
			// 防刷屏合集在空间只有一个稿件
			seasonArcMap[v.Arc.SeasonID] = v.Arc
		}
	}
	return seasonArcMap, seasonIds
}

// 通过合集信息排序稿件
func sortArcPlayerBySeasons(arcs []*api.ArcPlayer, seasons map[int64]*ugcSeasonGrpc.Season, order string) []*api.ArcPlayer {
	sort.SliceStable(arcs, func(i, j int) bool {
		if order == space.ArchivePlay {
			// 如果是播放数排序,播放多的在前
			// 如果是防刷屏合集稿件,取合集的总播放数
			ai, aj := arcs[i].Arc.Stat.View, arcs[j].Arc.Stat.View
			if season, ok := seasons[arcs[i].Arc.Aid]; ok {
				ai = season.Stat.View
			}
			if season, ok := seasons[arcs[j].Arc.Aid]; ok {
				aj = season.Stat.View
			}
			return ai > aj
		}
		// 如果是默认排序,最近更新的在前
		// 如果是防刷屏合集稿件,取合集的更新时间
		ai, aj := arcs[i].Arc.PubDate, arcs[j].Arc.PubDate
		if season, ok := seasons[arcs[i].Arc.Aid]; ok {
			ai = season.Ptime
		}
		if season, ok := seasons[arcs[j].Arc.Aid]; ok {
			aj = season.Ptime
		}
		return ai > aj
	})
	return arcs
}

// 通过合集信息排序稿件
func sortArcPlayerCursorBySeasons(arcs []*space.ArcPlayerCursor, seasons map[int64]*ugcSeasonGrpc.Season, order string) []*space.ArcPlayerCursor {
	sort.SliceStable(arcs, func(i, j int) bool {
		if order == space.ArchivePlay {
			// 如果是播放数排序,播放多的在前
			// 如果是防刷屏合集稿件,取合集的总播放数
			ai, aj := arcs[i].Arc.Stat.View, arcs[j].Arc.Stat.View
			if season, ok := seasons[arcs[i].Arc.Aid]; ok {
				ai = season.Stat.View
			}
			if season, ok := seasons[arcs[j].Arc.Aid]; ok {
				aj = season.Stat.View
			}
			return ai > aj
		}
		// 如果是默认排序,最近更新的在前
		// 如果是防刷屏合集稿件,取合集的更新时间
		ai, aj := arcs[i].Arc.PubDate, arcs[j].Arc.PubDate
		if season, ok := seasons[arcs[i].Arc.Aid]; ok {
			ai = season.Ptime
		}
		if season, ok := seasons[arcs[j].Arc.Aid]; ok {
			aj = season.Ptime
		}
		return ai > aj
	})
	return arcs
}

// 空间底部tag
func spaceTagBottom(ctx context.Context, userLocation *passportuser.UserActiveLocationReply, mcnInfo *space.McnInfo) []*space.SpaceTagBottom {
	var out []*space.SpaceTagBottom
	// ip属地标签
	func() {
		if userLocation != nil {
			if userLocation.Location == "" || userLocation.Location == "hide" {
				return
			}
			t := &space.SpaceTagBottom{
				Type:  "location",
				Title: fmt.Sprintf("IP属地：%s", userLocation.Location),
				Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/McvOxLw27A.png",
			}
			out = append(out, t)
			return
		}
	}()
	// mcn机构展示信息
	func() {
		if mcnInfo != nil && mcnInfo.Name != "" {
			t := &space.SpaceTagBottom{
				Type:  "mcn_info",
				Title: mcnInfo.Name,
				URI:   mcnInfo.Url,
				Icon:  "https://i0.hdslb.com/bfs/activity-plat/static/2be2c5f696186bad80d4b452e4af2a76/OsYihE3h0w.png",
			}
			out = append(out, t)
			return
		}
	}()
	return sortSpaceTagBottom(out)
}

func sortSpaceTagBottom(tags []*space.SpaceTagBottom) []*space.SpaceTagBottom {
	typeOrder := map[string]int64{
		"mcn_info": 1,
		"location": 2,
	}
	sort.Slice(tags, func(i, j int) bool {
		pi, ok := typeOrder[tags[i].Type]
		if !ok {
			pi = math.MaxInt64
		}
		pj, ok := typeOrder[tags[j].Type]
		if !ok {
			pj = math.MaxInt64
		}
		return pi < pj
	})
	return tags
}
