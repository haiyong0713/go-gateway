package view

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/component/metadata/device"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/pagination"
	egv2 "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/report"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	gameDao "go-gateway/app/app-svr/app-view/interface/dao/game"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/ad"
	"go-gateway/app/app-svr/app-view/interface/model/bangumi"
	"go-gateway/app/app-svr/app-view/interface/model/elec"
	"go-gateway/app/app-svr/app-view/interface/model/game"
	musicmdl "go-gateway/app/app-svr/app-view/interface/model/music"
	"go-gateway/app/app-svr/app-view/interface/model/special"
	"go-gateway/app/app-svr/app-view/interface/model/tag"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	ahApi "go-gateway/app/app-svr/archive-honor/service/api"
	"go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	notes "go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	channelApi "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	api2 "git.bilibili.co/bapis/bapis-go/copyright-manage/interface"
	v14 "git.bilibili.co/bapis/bapis-go/material/creative/interface/v1"
	v13 "git.bilibili.co/bapis/bapis-go/material/interface"
	natgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	esports "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	v12 "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	ogvgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	v1 "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	bgroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	resApiV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	topic "git.bilibili.co/bapis/bapis-go/topic/service"
	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	_dmformat                  = "http://comment.bilibili.com/%d.xml"
	_qnAndroidBuildGt          = 5325000
	_qnIosBuildGt              = 8170
	_previewAndroidBuild       = 5385000
	_previewIosBuild           = 8400
	_bangumiIPadBuild          = 8270
	_bangumiIPadHDBuild        = 12070
	_videoChannel              = 3
	_iphoneRelateRsc           = int64(2029)
	_androidRelateRsc          = int64(2028)
	_iPadRelateRsc             = int32(4489)
	_iPadSourceId              = int32(4490)
	_iphoneCMRsc               = int64(2335) // 框下广告位
	_androidCMRsc              = int64(2337) // 框下广告位
	_iphonePlayerCM            = int32(2642) // 框内广告位
	_androidPlayerCM           = int32(2643) // 框内广告位
	_playlistSpmid             = "playlist.playlist-video-detail.0.0"
	_missABValue               = "miss"
	_popupName                 = "is_play_push"
	_iconS11                   = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/1jx6bIGiWe.png"
	_iconNightS11              = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/5GDE58Yshv.png"
	_bGroupESportsNameS11      = "电竞赛事白名单" //人群包名称
	_bGroupESportsBusinessS11  = "topic"   //人群包business_id
	_textColorNewTag           = "#505050"
	_textColorNightNewTag      = "#828282"
	_cellColorNewTag           = "#F6F7F8"
	_cellColorNightNewTag      = "#1E2022"
	_iconNewTag                = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/GrtXAQ1SBX.png"
	_iconNightNewTag           = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/Ot6GVyHOs9.png"
	_smallWindowKeep           = "keep"    //保持
	_smallWindowOpen           = "open"    //打开
	_notesText                 = "UP主笔记"   //笔记标签-文案
	_notesTextColor            = "#505050" //笔记标签-文本颜色
	_notesTextColorNight       = "#828282" //笔记标签-文本颜色夜间
	_notesCellColor            = "#F6F7F8" //笔记标签-cell颜色
	_notesCellColorNight       = "#1E2022" //笔记标签-cell颜色夜间
	_iconNotes                 = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/xbmrfp92HQ.png"
	_iconNotesNight            = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/5ELu0j6hbf.png"
	_premiereTextTitle         = "首映说明" //首映标题文案
	_premiereTextSubTitle      = "1、预约首映，首映即将开始时将收到系统通知。\n2、首映开始后，视频开始播放，观众同步观看并在首映室实时互动。UP主可能会上线与观众互动噢~"
	_premiereOnlineText        = "进入首映室"
	_premiereOnlineIcon        = "https://i0.hdslb.com/bfs/activity-plat/static/20220331/070b04320d15630cee162e90b8fee339/NYlDEtodiH.png"
	_premiereOnlineIconDark    = "https://i0.hdslb.com/bfs/activity-plat/static/20220331/070b04320d15630cee162e90b8fee339/U3GgQyfMQv.png"
	_premiereOnlineIconNew     = "https://i0.hdslb.com/bfs/activity-plat/static/20220425/070b04320d15630cee162e90b8fee339/PRWszqP3iH.png"
	_premiereOnlineIconDarkNew = "https://i0.hdslb.com/bfs/activity-plat/static/20220425/070b04320d15630cee162e90b8fee339/veAOcRiRTP.png"
	_premiereGuidancePulldown  = "展开了解更多稿件信息~"
	_premiereGuidanceEntry     = "视频正在首映，去看看吧~"
	_premiereIntroTitle        = "首映介绍"
	_premiereIntroIcon         = "https://i0.hdslb.com/bfs/activity-plat/static/20220526/070b04320d15630cee162e90b8fee339/YGFFEQcM5v.png"
	_premiereIntroIconNight    = "https://i0.hdslb.com/bfs/activity-plat/static/20220526/070b04320d15630cee162e90b8fee339/rpj0XxRBzB.png"
	_fromSpmidFullscreen       = "player.ugc-video-detail.full-screen-relatedvideo.0"
	_spmidFullscreen           = "main.ugc-video-detail.0.0"
	_bijianTagIcon             = "http://i0.hdslb.com/bfs/app/8e12b0da3608fb72a5af266a34e40e8aeba286f6.png"
	_bijianTagIconNight        = "http://i0.hdslb.com/bfs/app/f686b0fe17759b0a8b7a33e0c9e52a6e7941dcff.png"
	_toolTagHePaiIcon          = "http://i0.hdslb.com/bfs/app/0195398051f0cb74a4158f3ce56fed0db3d67ad2.png"
	_toolTagHePaiIconNight     = "http://i0.hdslb.com/bfs/app/3a86125f5b98666a67aaed1a3cec10778b6d1787.png"
	_toolTagMuBanIcon          = "http://i0.hdslb.com/bfs/app/1da0b665641278a6d479e8346738e92a6706531c.png"
	_toolTagMuBanIconNight     = "http://i0.hdslb.com/bfs/app/b9cf5a0d5a1307608557675443819b6bd09dc74e.png"
	_toolTagTeXiaoIcon         = "http://i0.hdslb.com/bfs/app/d95ad4716d69a18f2f4ea435a4df7a6676315756.png"
	_toolTagTexiaoIconNight    = "http://i0.hdslb.com/bfs/app/c898285b9b6ce26bdf08ef8a29d58f09475a6b01.png"
	_toolTagMovieIcon          = "http://i0.hdslb.com/bfs/app/9d5c4b18251f618c5924bbc1e38f0e425b4d52c6.png"
	_toolTagMovieIconNight     = "http://i0.hdslb.com/bfs/app/f851eadf140be4515e05ab571a55e2d6cbaa44b9.png"
	_toolTagHePai              = 9  //合拍
	_toolTagTeXiao             = 64 //特效
	_toolTagPaiShe             = 5  //拍摄特效
	_toolTagText               = 44 //文字模板
	_toolTagVideo              = 46 //视频模板
	_toolTagMusic              = 14 //音乐模板
	_toolTagMovie              = 72 //一键成片
	_inspirationKey            = "aid_inspiration_topic"
	_inspirationTagIcon        = "http://i0.hdslb.com/bfs/app/c412bc2e47790279951498156d8c9be36956e10c.png"
	_inspirationTagIconNight   = "http://i0.hdslb.com/bfs/app/2e8361ece8faf3ac4a50eaf1b2020647f9242805.png"
)

type ToolIcon struct {
	Icon      string
	IconNight string
}

var (
	popupFlag = ab.String(_popupName, "popupConfig", _missABValue)
	toolTag   = map[int32]ToolIcon{
		_toolTagHePai:  {Icon: _toolTagHePaiIcon, IconNight: _toolTagHePaiIconNight},   //合拍icon
		_toolTagTeXiao: {Icon: _toolTagTeXiaoIcon, IconNight: _toolTagTexiaoIconNight}, //特效icon
		_toolTagPaiShe: {Icon: _toolTagTeXiaoIcon, IconNight: _toolTagTexiaoIconNight}, //拍摄icon
		_toolTagText:   {Icon: _toolTagMuBanIcon, IconNight: _toolTagMuBanIconNight},   //文字icon
		_toolTagVideo:  {Icon: _toolTagMuBanIcon, IconNight: _toolTagMuBanIconNight},   //视频icon
		_toolTagMusic:  {Icon: _toolTagMuBanIcon, IconNight: _toolTagMuBanIconNight},   //音乐icon
		_toolTagMovie:  {Icon: _toolTagMovieIcon, IconNight: _toolTagMovieIconNight},   //一键成片icon
	}
)

func (s *Service) newLiveBuild(plat int8, build int) bool {
	if (model.IsIPhone(plat) && build > s.c.BuildLimit.LiveIOSBuildLimit) ||
		(model.IsAndroid(plat) && build > s.c.BuildLimit.LiveAndBuildLimit) {
		return true
	}
	return false
}

func (s *Service) newSeasonTypeBuild(c context.Context) bool {
	buildBool := pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">", int64(s.c.BuildLimit.SeasonTypeIOSBuildLimit))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">", int64(s.c.BuildLimit.SeasonTypeAndBuildLimit))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build(">=", int64(s.c.BuildLimit.SeasonBaseAndroidHdBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">=", int64(33300000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build(">=", int64(65600000))
	}).FinishOr(true)
	return buildBool
}

func (s *Service) playStoryABTest(c context.Context, mid int64, buvid string) (bool, string, bool, string) {
	buildBool := pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.StoryPlayIOS)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.StoryPlayAndroid)
	}).FinishOr(true)
	if !buildBool {
		return false, "", false, ""
	}
	//是否是7天新用户
	newDevice := s.accDao.IsNewDevice(c, buvid, "0-168")
	//竖屏实验
	isPlayStory, storyIcon := func() (bool, string) {
		//最高级白名单，防止内部人员进入黑名单之后体验不了功能
		if mid > 0 {
			for _, v := range s.c.Custom.PlayStoryMids {
				if mid == v {
					return true, s.c.Custom.StoryIcon
				}
			}
		}
		if !newDevice { //老用户过滤近7天没有点击过竖屏稿件全屏按钮
			if s.abPlay.HitPlayBlackList(c, buvid) {
				return false, ""
			}
		}
		//实验全量
		return true, s.c.Custom.StoryIcon
	}()
	//横屏实验
	landscapeStory, landscapeIcon := s.abPlay.LandscapeStoryExp(c, buvid, newDevice)
	return isPlayStory, storyIcon, landscapeStory, landscapeIcon
}

// initReqUser init Req User
// nolint:gocognit,gomnd
func (s *Service) initReqUser(c context.Context, v *view.View, mid int64, plat int8, build int, buvid, platform, brand, net, mobiApp string) {
	// owner ext
	var (
		owners   []int64
		cards    map[int64]*accApi.Card
		fls      map[int64]int8
		staffMap = make(map[int64]*api.StaffInfo)
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	g := egv2.WithContext(c)
	if v.Author.Mid > 0 {
		owners = append(owners, v.Author.Mid)
		for _, staffInfo := range v.StaffInfo {
			if staffInfo == nil {
				continue
			}
			owners = append(owners, staffInfo.Mid)
			staffMap[staffInfo.Mid] = staffInfo
		}
		g.Go(func(ctx context.Context) (err error) {
			v.OwnerExt.OfficialVerify.Type = -1
			cards, err = cfg.dep.Account.Cards3(ctx, owners)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if card, ok := cards[v.Author.Mid]; ok && card != nil {
				if card.Official.Role != 0 {
					v.OwnerExt.OfficialVerify.Desc = card.Official.Title
				}
				v.OwnerExt.OfficialVerify.Type = int(card.Official.Type)
				v.OwnerExt.Vip.Type = int(card.Vip.Type)
				v.OwnerExt.Vip.VipStatus = int(card.Vip.Status)
				v.OwnerExt.Vip.DueDate = card.Vip.DueDate
				v.OwnerExt.Vip.ThemeType = int(card.Vip.ThemeType)
				v.OwnerExt.Vip.Label = card.Vip.Label
				v.Author.Name = card.Name
				v.Author.Face = card.Face
			}
			return
		})
		// 6.16 班车需求
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.UpArcCount, &feature.OriginResutl{
			MobiApp:    mobiApp,
			Build:      int64(build),
			BuildLimit: (mobiApp == "iphone" && build >= 61600000) || (mobiApp == "android" && build >= 6160000),
		}) {
			g.Go(func(ctx context.Context) error {
				count, err := cfg.dep.UpArc.UpArcCount(ctx, v.Author.Mid)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				var text string
				if count >= 10000 {
					text = strconv.FormatFloat(float64(count)/10000, 'f', 1, 64) + "万"
				} else {
					text = fmt.Sprintf("%d", count)
				}
				v.OwnerExt.ArcCount = text + s.c.Custom.UpArcText
				return nil
			})
		}
		if !model.IsIPhoneB(plat) && !model.IsAndroidB(plat) {
			g.Go(func(ctx context.Context) error {
				// 6.16版本之后支持开关控制直播状态
				if s.newLiveBuild(plat, build) && !s.c.Custom.LiveSwitchOn {
					return nil
				}
				l, _ := cfg.dep.Live.LivingRoom(ctx, v.Author.Mid, platform, brand, net, build, mid)
				if l == nil {
					return nil
				}
				v.OwnerExt.Live = l
				return nil
			})
		}
		g.Go(func(ctx context.Context) (err error) {
			stat, err := cfg.dep.Relation.Stat(ctx, v.Author.Mid)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if stat != nil {
				v.OwnerExt.Fans = int(stat.Follower)
			}
			return
		})
		g.Go(func(ctx context.Context) error {
			if ass, err := cfg.dep.Assist.Assist(ctx, v.Author.Mid); err != nil {
				log.Error("%+v", err)
			} else {
				v.OwnerExt.Assists = ass
			}
			return nil
		})
	}
	// req user
	v.ReqUser = &viewApi.ReqUser{Favorite: 0, Attention: -999, Like: 0, Dislike: 0, GuestAttention: -999}
	if mid > 0 || buvid != "" {
		g.Go(func(ctx context.Context) error {
			likeState, err := cfg.dep.ThumbUP.HasLike(ctx, mid, _businessLike, buvid, v.Aid)
			if err != nil {
				log.Error("s.thumbupDao.HasLike mid(%d) buvid(%s) bus(%s) aid(%d) err(%+v)", mid, buvid, _businessLike, v.Aid, err)
				return nil
			}
			if likeState == thumbup.State_STATE_LIKE {
				v.ReqUser.Like = 1
				if mid == 0 { // 未登录&&已点赞的稿件点赞数兼容+1
					v.Stat.Like += 1
				}
			} else if likeState == thumbup.State_STATE_DISLIKE {
				v.ReqUser.Dislike = 1
			}
			return nil
		})
	}
	// check req user
	if mid > 0 {
		g.Go(func(ctx context.Context) error {
			res := cfg.dep.Fav.IsFavoredsResources(ctx, mid, v.Aid, v.SeasonID)
			if res == nil {
				return nil
			}
			if fv, ok := res[model.FavTypeVideo]; ok && fv {
				v.ReqUser.Favorite = 1
			}
			if fs, ok := res[model.FavTypeSeason]; ok && fs {
				v.ReqUser.FavSeason = 1
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			res, err := cfg.dep.Coin.ArchiveUserCoins(ctx, v.Aid, mid, _avTypeAv)
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			if res != nil && res.Multiply > 0 {
				v.ReqUser.Coin = 1
			}
			return
		})
		if v.Author.Mid > 0 {
			g.Go(func(ctx context.Context) error {
				fls = cfg.dep.Account.IsAttention(ctx, owners, mid)
				if _, ok := fls[v.Author.Mid]; ok {
					v.ReqUser.Attention = 1
				}
				return nil
			})
			if mid > 0 && v.Author.Mid != mid {
				g.Go(func(ctx context.Context) error {
					gusfls, _ := cfg.dep.Account.ContractRelation3(ctx, v.Author.Mid, mid)
					if gusfls != nil && (gusfls.Attribute == 2 || gusfls.Attribute == 6) {
						v.ReqUser.GuestAttention = 1
					}
					if gusfls != nil && gusfls.ContractInfo != nil && gusfls.ContractInfo.IsContractor {
						v.IsContractor = true
						if gusfls.ContractInfo.UserAttr == 1 {
							v.IsOldFans = true
						}
					}
					return nil
				})
			}
		}
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// fill staff
	if v.AttrVal(api.AttrBitIsCooperation) == api.AttrYes {
		for _, owner := range owners {
			if card, ok := cards[owner]; ok && card != nil {
				staff := &view.Staff{Mid: owner}
				if owner == v.Author.Mid { // staff中不带up信息，up加在第一位
					staff.Title = "UP主"
				} else if s, ok := staffMap[owner]; ok {
					staff.Title = s.Title
					if s.StaffAttrVal(api.StaffAttrBitAdOrder) == api.AttrYes {
						staff.LabelStyle = model.StaffLabelAd
					}
				}
				staff.Name = card.Name
				staff.Face = card.Face
				staff.OfficialVerify.Type = -1
				otp := -1
				odesc := ""
				if card.Official.Role != 0 {
					odesc = card.Official.Title
					otp = 1
					if card.Official.Role <= 2 || card.Official.Role == 7 {
						otp = 0
					}
				}
				staff.OfficialVerify.Type = otp
				staff.OfficialVerify.Desc = odesc
				staff.Vip.Type = int(card.Vip.Type)
				staff.Vip.VipStatus = int(card.Vip.Status)
				staff.Vip.DueDate = card.Vip.DueDate
				staff.Vip.ThemeType = int(card.Vip.ThemeType)
				staff.Vip.Label = card.Vip.Label
				if _, ok := fls[owner]; ok {
					staff.Attention = 1
				}
				v.Staff = append(v.Staff, staff)
			}
		}
	}
}

func (s *Service) infocAd(c context.Context, mobiApp string, build int, network string, mid int64, buvid string, aid int64, rscIDs []int32, platform, isMelloi, event string) {
	if isMelloi != "" {
		return
	}
	for _, v := range rscIDs {
		event := infoc.NewLogStreamV("002770",
			log.String(event),
			log.String(mobiApp),
			log.String(platform),
			log.String(fmt.Sprintf("%d", build)),
			log.String(network),
			log.String(fmt.Sprintf("%d", mid)),
			log.String(buvid),
			log.Int64(time.Now().Unix()),
			log.String(fmt.Sprintf("%d", v)),
			log.String(fmt.Sprintf("%d", aid)),
		)
		if err := s.infocV2Log.Info(c, event); err != nil {
			log.Error("infocViewAd params(%s,%s,%d,%s,%d,%s,%d,%d) err(%v)",
				mobiApp, platform, build, network, mid, buvid, v, aid, err)
		}
	}
}

// nolint:gocognit
func (s *Service) initRelateCMTag(c context.Context, v *view.View, plat int8, build, parentMode, autoplay int, mid int64,
	buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, trackid, platform, filtered string, tids []int64, isMelloi, slocale, clocale, pageVersion string) {
	const (
		_iPhoneRelateGame  = 6500
		_androidRelateGame = 5210000
	)
	var (
		rls                                []*view.Relate
		aidm                               map[int64]struct{}
		rGameID, cGameID, relateRsc, cmRsc int64
		advert                             *ad.Ad
		adm                                map[int]*ad.AdInfo
		err                                error
		relateConf                         *view.RelateConf
		hasDalao, gamecardStyleExp         int
	)
	// 审核版本，和有屏蔽推荐池属性的稿件下 不出相关推荐任何信息
	if filtered == "1" || v.ForbidRec == 1 {
		log.Warn("no relates aid(%d) filtered(%s) ForbidRec(%d)", v.Aid, filtered, v.ForbidRec)
		return
	}
	g := egv2.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		if v.AttrVal(api.AttrBitIsPGC) == api.AttrYes && v.RedirectURL != "" {
			return nil
		}
		if mid > 0 || buvid != "" {
			isNewColor := false
			if (model.IsAndroid(plat) && build > s.c.BuildLimit.CardOptimizeAndroid) ||
				(model.IsIPhone(plat) && build > s.c.BuildLimit.CardOptimizeIPhone) ||
				(model.PlatIpadHD == plat && build > s.c.BuildLimit.CardOptimizeIPadHD) {
				isNewColor = true
			}
			if rls, v.TabInfo, v.PlayParam, v.UserFeature, v.ReturnCode, v.PvFeature, relateConf, err = s.newRcmdRelate(ctx, plat, v.Aid, mid, v.ZoneID, buvid, mobiApp, from, trackid, model.RelateCmd, "", build, parentMode, autoplay, 0, isNewColor, pageVersion, fromSpmid); err != nil {
				log.Error("s.newRcmdRelate(%d) error(%+v)", v.Aid, err)
			}
			if relateConf != nil {
				hasDalao = relateConf.HasDalao
				gamecardStyleExp = relateConf.GamecardStyleExp
				if v.Config != nil {
					v.Config.AutoplayCountdown = s.c.ViewConfig.AutoplayCountdown
					if relateConf.AutoplayCountdown > 0 {
						v.Config.AutoplayCountdown = relateConf.AutoplayCountdown
					}
					v.Config.PageRefresh = relateConf.ReturnPage
					v.Config.AutoplayDesc = relateConf.AutoplayToast
					v.Config.RelatesStyle = relateConf.RelatesStyle
					v.Config.RelateGifExp = relateConf.GifExp
				}
			}
			if len(v.TabInfo) > 0 {
				var tmpTab []*viewApi.RelateTab
				for _, t := range v.TabInfo {
					if t == nil {
						continue
					}
					tmpTab = append(tmpTab, &viewApi.RelateTab{Id: t.ID, Title: t.Desc})
				}
				v.RelateTab = tmpTab
			}
		}
		// ai：code=-3表示无有效结果稿;code=5表示屏蔽用户黑名单;code=-2表示内部拉用户信息缺失
		if len(rls) == 0 && v.ReturnCode != "-3" && v.ReturnCode != "-5" { // -3和-5不要取灾备数据
			rls, err = s.dealRcmdRelate(ctx, plat, v.Aid, mid, build, mobiApp, device)
			log.Warn("s.dealRcmdRelate aid(%d) mid(%d) buvid(%s) build(%d) mobiApp(%s) device(%s) err(%+v)", v.Aid, mid, buvid, build, mobiApp, device, err)
		} else {
			v.IsRec = 1
			log.Info("s.newRcmdRelate returncode(%s) aid(%d) mid(%d) buvid(%s) hasDalao(%d)", v.ReturnCode, v.Aid, mid, buvid, hasDalao)
		}
		return nil
	})
	if v.AttrValV2(model.AttrBitV2CleanMode) == api.AttrNo {
		if !model.IsIPad(plat) && !model.IsOverseas(plat) {
			g.Go(func(ctx context.Context) (err error) {
				if model.IsIPhone(plat) {
					relateRsc = _iphoneRelateRsc
					cmRsc = _iphoneCMRsc
				} else {
					relateRsc = _androidRelateRsc
					cmRsc = _androidCMRsc
				}
				s.infocAd(ctx, mobiApp, build, network, mid, buvid, v.Aid, []int32{int32(relateRsc), int32(cmRsc)}, platform, isMelloi, "view_request_before")
				if advert, err = s.adDao.Ad(ctx, mobiApp, device, buvid, build, mid, v.Author.Mid, v.Aid, v.TypeID, tids, []int64{relateRsc, cmRsc}, network, adExtra, spmid, fromSpmid, from); err != nil {
					log.Error("s.adDao.Ad err(%+v)", err)
				}
				s.infocAd(ctx, mobiApp, build, network, mid, buvid, v.Aid, []int32{int32(relateRsc), int32(cmRsc)}, platform, isMelloi, "view_request")
				return nil
			})
		}
		if (plat == model.PlatAndroid && build >= _androidRelateGame) || (plat == model.PlatIPhone && build >= _iPhoneRelateGame) {
			if buvid != "" && crc32.ChecksumIEEE([]byte(buvid))%10 == 1 {
				g.Go(func(ctx context.Context) (err error) {
					rGameID = s.relateGame(v.Aid)
					return nil
				})
			}
			if v.AttrVal(api.AttrBitIsPorder) == api.AttrYes || v.OrderID > 0 {
				g.Go(func(ctx context.Context) (err error) {
					if cGameID, err = s.vuDao.ArcCommercial(ctx, v.Aid); err != nil {
						log.Error("s.vuDao.ArcCommercial aid(%d),err(%+v)", v.Aid, err)
					}
					return nil
				})
			}
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("initRelateCMTag g.wait err(%+v)", err)
		return
	}
	if advert != nil {
		if advert.AdsControl != nil {
			v.CMConfig = &view.CMConfig{
				AdsControl: advert.AdsControl,
			}
		}
		if adm, err = s.dealCM(c, advert, relateRsc); err != nil {
			log.Error("%+v", err)
		}
		initCM(v, advert, cmRsc)
	}
	// ad
	if len(rls) == 0 {
		s.prom.Incr("zero_relates")
		return
	}
	var (
		rm map[int]*view.Relate
	)
	// ai已经有dalao卡则直接返回，没有则看要不要第一位拼接其他卡
	if hasDalao > 0 {
		v.Relates = rls
		return
	}
	r := s.dealGame(c, plat, build, gamecardStyleExp, cGameID, model.FromOrder)
	if r == nil {
		r = s.dealGame(c, plat, build, gamecardStyleExp, rGameID, model.FromRcmd)
	}
	if r != nil {
		rm = map[int]*view.Relate{0: r}
		aidm = map[int64]struct{}{r.Aid: {}}
	} else if len(adm) != 0 {
		rm = make(map[int]*view.Relate, len(adm))
		for idx, ad := range adm {
			r = &view.Relate{}
			r.FromCM(ad)
			rm[idx] = r
		}
	}
	if len(rm) != 0 {
		var tmp []*view.Relate
		for _, rl := range rls {
			if _, ok := aidm[rl.Aid]; ok {
				continue
			}
			tmp = append(tmp, rl)
		}
		v.Relates = make([]*view.Relate, 0, len(tmp)+len(rm))
		for _, rl := range tmp {
		LABEL:
			if r, ok := rm[len(v.Relates)]; ok {
				if r.IsAdLoc && r.AdCb == "" {
					rel := &view.Relate{}
					*rel = *rl
					rel.IsAdLoc = r.IsAdLoc
					rel.RequestID = r.RequestID
					rel.SrcID = r.SrcID
					rel.ClientIP = r.ClientIP
					rel.AdIndex = r.AdIndex
					rel.Extra = r.Extra
					rel.CardIndex = r.CardIndex
					v.Relates = append(v.Relates, rel)
				} else if r.Aid != v.Aid {
					v.Relates = append(v.Relates, r)
					goto LABEL
				} else {
					v.Relates = append(v.Relates, rl)
				}
			} else {
				v.Relates = append(v.Relates, rl)
			}
		}
	} else {
		v.Relates = rls
	}
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		for _, rl := range v.Relates {
			i18n.TranslateAsTCV2(&rl.Title)
		}
	}
}

func initCM(v *view.View, advert *ad.Ad, resource int64) {
	ads, _ := advert.Convert(resource)
	sort.Sort(ad.AdInfos(ads))
	if len(ads) == 0 {
		return
	}
	v.CMs = make([]*view.CM, 0, len(ads))
	for _, ad := range ads {
		cm := &view.CM{}
		cm.FromCM(ad)
		v.CMs = append(v.CMs, cm)
	}
}

func (s *Service) initMovie(c context.Context, v *view.View, mid int64, build int, mobiApp, device string, nMovie bool) (err error) {
	s.pHit.Incr("is_movie")
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	var m *bangumi.Movie
	if m, err = cfg.dep.PGC.Movie(c, v.Aid, mid, build, mobiApp, device); err != nil || m == nil {
		log.Error("%+v", err)
		err = ecode.NothingFound
		s.pMiss.Incr("err_is_PGC")
		return
	}
	if v.Rights.HD5 == 1 && m.PayUser.Status == 0 && !s.checkVIP(c, mid) {
		v.Rights.HD5 = 0
	}
	if len(m.List) == 0 {
		err = ecode.NothingFound
		return
	}
	vps := make([]*view.Page, 0, len(m.List))
	for _, l := range m.List {
		vp := &view.Page{
			Page: &api.Page{Cid: l.Cid, Page: int32(l.Page), From: l.Type, Part: l.Part, Vid: l.Vid},
		}
		vps = append(vps, vp)
	}
	m.List = nil
	// view
	v.Pages = vps
	v.Rights.Download = int32(m.AllowDownload)
	m.AllowDownload = 0
	v.Rights.Bp = 0
	if nMovie {
		v.Movie = m
		v.Desc = m.Season.Evaluate
	}
	return
}

func (s *Service) initPGC(c context.Context, v *view.View, mid int64, build int, mobiApp, device string) (err error) {
	s.pHit.Incr("is_PGC")
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	var season *bangumi.Season
	if season, err = cfg.dep.PGC.PGC(c, v.Aid, mid, build, mobiApp, device); err != nil {
		log.Error("%+v", err)
		err = ecode.NothingFound
		s.pMiss.Incr("err_is_PGC")
		return
	}
	if season != nil {
		if season.Player != nil {
			if len(v.Pages) != 0 {
				if season.Player.Cid != 0 {
					v.Pages[0].Cid = season.Player.Cid
				}
				if season.Player.From != "" {
					v.Pages[0].From = season.Player.From
				}
				if season.Player.Vid != "" {
					v.Pages[0].Vid = season.Player.Vid
				}
			}
			season.Player = nil
		}
		if season.AllowDownload == "1" {
			v.Rights.Download = 1
		} else {
			v.Rights.Download = 0
		}
		if season.SeasonID != "" {
			season.AllowDownload = ""
			v.Season = season
		}
	}
	if v.Rights.HD5 == 1 && !s.checkVIP(c, mid) {
		v.Rights.HD5 = 0
	}
	v.Rights.Bp = 0
	return
}

func (s *Service) initPages(c context.Context, vs *view.ViewStatic, ap []*api.Page, mobiApp string, build int) {
	pages := make([]*view.Page, 0, len(ap))
	for _, v := range ap {
		page := &view.Page{}
		metas := view.BuildMetas(v.Duration)
		if vs.AttrVal(api.AttrBitIsBangumi) == api.AttrYes {
			v.From = "bangumi"
		}
		// iphone 于6.0.0版本修复分p标题含有4字节emoji崩溃问题
		if v.Part != "" && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.PageEmoji, &feature.OriginResutl{
			MobiApp:    mobiApp,
			Build:      int64(build),
			BuildLimit: (mobiApp == "iphone" && build < 10000) || (mobiApp == "iphone_b" && build < 9360),
		}) {
			v.Part = s.FilterEmoji(v.Part)
		}
		page.Page = v
		page.Metas = metas
		page.DMLink = fmt.Sprintf(_dmformat, v.Cid)
		page.DlTitle = "视频已缓存完成"
		page.DlSubtitle = fmt.Sprintf("%s %s", vs.Title, v.Part)
		pages = append(pages, page)
	}
	vs.Pages = pages
}

func (s *Service) initUGCPay(c context.Context, v *view.View, plat int8, mid int64, build int) (err error) {
	var (
		asset      *view.Asset
		platform   = model.Platform(plat)
		canPreview bool
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if (v.AttrVal(api.AttrBitUGCPayPreview) == api.AttrYes) && ((plat == model.PlatIPhone && build > _previewIosBuild) || (plat == model.PlatAndroid && build > _previewAndroidBuild)) {
		canPreview = true
	}
	if asset, err = cfg.dep.UGCPay.AssetRelationDetail(c, mid, v.Aid, platform, canPreview); err != nil {
		log.Error("%+v", err)
		return
	}
	if asset != nil {
		v.Asset = asset
	}
	return
}

//nolint:gomnd
func (s *Service) initAudios(c context.Context, v *view.View) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	pLen := len(v.Pages)
	if pLen == 0 || pLen > 100 {
		return
	}
	if pLen > 50 {
		pLen = 50
	}
	cids := make([]int64, 0, len(v.Pages[:pLen]))
	for _, p := range v.Pages[:pLen] {
		cids = append(cids, p.Cid)
	}
	vam, err := cfg.dep.Audio.AudioByCids(c, cids)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(vam) != 0 {
		for _, p := range v.Pages[:pLen] {
			if va, ok := vam[p.Cid]; ok {
				p.Audio = va
			}
		}
		if len(v.Pages) == 1 {
			if va, ok := vam[v.Pages[0].Cid]; ok {
				v.Audio = va
			}
		}
	}
}

// initElecRank .
func (s *Service) initElecRank(c context.Context, v *view.View, mobiApp, platform, device string, build int) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if _, ok := s.allowTypeIds[int16(v.TypeID)]; !ok || int8(v.Copyright) != api.CopyrightOriginal {
		return
	}
	info, err := cfg.dep.UGCPayRank.RankElecMonthUP(c, v.Author.Mid, int64(build), mobiApp, platform, device)
	if err != nil {
		log.Error("initElecRank  upmid:%d %+v", v.Author.Mid, err)
		return
	}
	if info != nil && info.UP != nil {
		var tmp []*viewApi.ElecRankItem
		for _, i := range info.UP.List {
			tmp = append(tmp, &viewApi.ElecRankItem{
				Nickname: i.Nickname,
				Avatar:   i.Avatar,
				Mid:      i.MID,
				Message:  i.Message,
			})
		}
		if tmp != nil {
			v.ElecRank = &viewApi.ElecRank{List: tmp, Count: info.UP.Count}
		}
	}
}

func (s *Service) initElec(c context.Context, v *view.View, mobiApp, platform, device string, build int, mid int64) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if _, ok := s.allowTypeIds[int16(v.TypeID)]; !ok || int8(v.Copyright) != api.CopyrightOriginal {
		return
	}
	info, err := cfg.dep.UGCPayRank.UPRankWithPanelByUPMid(c, mid, v.Author.Mid, int64(build), mobiApp, platform, device)
	if err != nil {
		log.Error("initElec upmid:%d err(%+v) ", v.Author.Mid, err)
		return
	}
	if info != nil && (info.Show || info.State == model.ElecEnableStatus || info.State == model.ElecPlusEnableStatus) {
		v.Rights.Elec = 1
		v.Elec = elec.FormatElec(info)

		var tmp []*viewApi.ElecRankItem
		for _, item := range info.List {
			tmp = append(tmp, &viewApi.ElecRankItem{
				Nickname: item.Nickname,
				Avatar:   item.Avatar,
				Mid:      item.MID,
				Message:  item.Message,
			})
		}
		if len(tmp) > 0 {
			text := "人为TA充电"
			if info.State == model.ElecPlusEnableStatus {
				text = info.RankTitle
			}
			v.ElecRank = &viewApi.ElecRank{List: tmp, Count: info.Count, Text: text}
		}
	}
}

func (s *Service) initResult(v *view.View, in *view.InitTag) {
	if in == nil {
		return
	}
	v.ViewTab = in.ViewTab
	v.Config = in.Config
	v.ActivityURL = in.ActivityURL
	v.Tag = in.Tag
	v.TIcon = in.TIcon
	v.SpecialCell = in.SpecialCell
	v.SpecialCellNew = in.SpecialCellNew
	v.DescTag = in.DescTag
	v.RefreshSpecialCell = in.RefreshSpecialCell
	v.MaterialLeft = in.MaterialLeft
	v.NotesCount = in.NotesCount
	v.ClientAction = in.ClientAction
}

// nolint:gocognit,ineffassign
func (s *Service) initTag(c context.Context, arc *api.Arc, cid, mid int64, plat int8, build int, pageVersion, buvid, mobiApp, spmid, platform string, extra map[string]string) (v *view.InitTag) {
	var (
		actTag         []*tag.Tag //活动tag
		arcTag         []*tag.Tag //非活动tag
		actTagName     string
		esportsCell    *viewApi.SpecialCell
		topicCell      *viewApi.SpecialCell
		note           *notes.ArcTagReply  //笔记
		music          *musicmdl.MusicInfo //bgm
		ogvChan        *channelApi.Channel
		bijianMaterial *v13.StoryPlayerRes      //必剪素材tag
		toolMaterial   *v14.PlayPageMaterialTag //工具类tag
		bgroupReply    map[string]bool
	)
	v = &view.InitTag{}
	//政务白名单
	v.Config = &view.Config{}
	gvn := egv2.WithContext(c)
	//判断up_mid是否在白名单里 人群包参数
	gvn.Go(func(c context.Context) error {
		groups := []*bgroup.MemberInReq_MemberInReqSingle{
			{Name: _bGroupESportsNameS11, Business: _bGroupESportsBusinessS11}, //电竞
			{Name: s.c.Custom.GroupsName, Business: s.c.Custom.GroupsBusiness}, //政务
		}
		var err error
		bgroupReply, err = s.IsMidExists(c, arc.Author.Mid, groups)
		if err != nil {
			log.Error("s.IsMidExists is err %+v %+v %+v", arc.Author.Mid, groups, err)
			return nil
		}
		v.Config.IsAbsoluteTime = bgroupReply[s.c.Custom.GroupsName]
		return nil
	})
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	gvn.Go(func(c context.Context) error {
		if arc.MissionID <= 0 {
			return nil
		}
		protocol, err := cfg.dep.Activity.ActProtocol(c, arc.MissionID)
		if err != nil {
			log.Error("s.actDao.ActProtocol err(%+v)", err)
			return nil
		}
		if protocol != nil && protocol.ShowStatus == int64(3) { //活动进行中才有活动tag
			if protocol.Subject != nil {
				v.ActivityURL = protocol.Subject.AndroidURL
				if model.IsIOS(plat) {
					v.ActivityURL = protocol.Subject.IosURL
				}
			}
			if protocol.Protocol != nil {
				actTagName = protocol.Protocol.Tags
			}
		}
		return nil
	})
	var chans []*channelApi.Channel
	gvn.Go(func(c context.Context) (err error) {
		chans, err = cfg.dep.Channel.ResourceChannels(c, arc.Aid, mid, _videoChannel)
		if err != nil { //err 直接返回
			log.Error("s.channelDao.ResourceChannels err(%+v)", err)
		}
		return
	})
	err := gvn.Wait()
	if err != nil {
		return
	}
	// chans等于0直接返回
	if len(chans) == 0 {
		return
	}
	var actTids []int64
	// 优先级  话题活动的链接模式 > 话题活动 > 新频道 > 旧频道
	arcTag, actTag, ogvChan, actTids = s.ResourceChannelsHandle(chans, plat, build, actTagName, spmid)
	// 获取活动tag的信息
	if (model.IsAndroid(plat) && build > s.c.BuildLimit.ViewChannelActiveAndroid) || (model.IsIOSNormal(plat) && build > s.c.BuildLimit.ViewChannelActiveIOS) ||
		(model.IsIpadHD(plat) && build >= s.c.BuildLimit.NewTopicIPadHDBuild) {
		if s.c.Custom.ViewTag && len(actTids) > 0 {
			var actInfos map[int64]*natgrpc.NativePage
			content := map[string]string{
				"view_aid": strconv.FormatInt(arc.Aid, 10),
			}
			if actInfos, err = cfg.dep.NatPage.NatInfoFromForeign(c, actTids, 1, content); err != nil {
				log.Error("%v", err)
				err = nil // 已经具备新旧频道跳转能力，活动tag跳转属于"锦上添花"逻辑，舍弃报错，确保详情页整体稳定
			}
			for _, actTagTmp := range actTag {
				if actInfo, ok := actInfos[actTagTmp.TagID]; ok {
					if actInfo == nil {
						continue
					}
					if actInfo.SkipURL != "" {
						actTagTmp.URI = actInfo.SkipURL
						continue
					}
					if actInfo.ID > 0 {
						actTagTmp.URI = fmt.Sprintf("bilibili://following/activity_landing/%d", actInfo.ID)
						continue
					}
				}
			}
		}
	}
	// 活动稿件tag放在第一位
	v.Tag = append(actTag, arcTag...)
	v.TIcon = make(map[string]*viewApi.TIcon)
	v.TIcon["act"] = &viewApi.TIcon{Icon: s.c.TagConfig.ActIcon}
	if s.c.TagConfig.OpenIcon {
		v.TIcon["new"] = &viewApi.TIcon{Icon: s.c.TagConfig.NewIcon}
	}
	var tagIDs []int64
	for _, tagData := range v.Tag {
		tagIDs = append(tagIDs, tagData.TagID)
	}
	upIDs := []int64{arc.Author.Mid}
	for _, staff := range arc.StaffInfo {
		upIDs = append(upIDs, staff.Mid)
	}
	if (plat == model.PlatAndroidI && build >= 6790400) || (plat == model.PlatIPhoneI && build >= 67900400) {
		v.DescTag = append(v.DescTag, arcTag...)
	}
	//新详情页跳过
	if !cfg.skipSpecialCell && pageVersion != "v2" {
		//版本判断
		if (plat == model.PlatAndroid && build >= s.c.BuildLimit.SpecialCellAndroidBuild) || (plat == model.PlatIPhone && build >= s.c.BuildLimit.SpecialCellIOSBuild) ||
			(plat == model.PlatIpadHD && build >= s.c.BuildLimit.SpecialCellIPadHDBuild) || (plat == model.PlatIPad && build >= s.c.BuildLimit.SpecialCellIPadBuild) {
			var isNewVersion bool
			if (mobiApp == "android" && build >= s.c.BuildLimit.MusicAndroidBuild) || (mobiApp == "iphone" && build >= s.c.BuildLimit.MusicIOSBuild) {
				isNewVersion = true
			}
			g := egv2.WithContext(c)
			//s11命中白名单 + ESportS12Switch(s12的时候s11需要关闭)
			s11WhiteMidBool := bgroupReply[_bGroupESportsNameS11]
			//是否为s12
			isS12 := false
			if s.c.Custom.ESportS12Switch || funk.Contains(s.c.Custom.ESportS12MidWhite, mid) {
				isS12 = true
			}
			if !isS12 && s11WhiteMidBool {
				//赛事标签
				g.Go(func(ctx context.Context) (err error) {
					esportsCell, err = s.eSportsSpecialCell(ctx, arc.Author.Mid, arc.Aid)
					if err != nil {
						log.Error("s.eSportsSpecialCell is err %+v", err)
					}
					return nil
				})
			}
			if isS12 {
				//s12赛事
				g.Go(func(ctx context.Context) (err error) {
					esportsCell, err = s.eSportsS12(ctx, tagIDs, mid)
					if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
						log.Error("s.eSportsS12 is err %+v", err)
					}
					return nil
				})
			}
			//新话题标签
			if (plat == model.PlatAndroid && build >= s.c.BuildLimit.NewTopicAndroidBuild) || (plat == model.PlatIPhone && build >= s.c.BuildLimit.NewTopicIOSBuild) {
				g.Go(func(ctx context.Context) (err error) {
					topicCell, err = s.newTopicSpecialCell(ctx, arc.Aid)
					if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
						log.Error("s.newTopicSpecialCell is err %+v", err)
					}
					return nil
				})
				//以前服务端下发的icon置为空
				v.TIcon["new"] = nil
			}
			//笔记标签
			g.Go(func(ctx context.Context) (err error) {
				note, err = s.notesDao.ArcNote(ctx, arc.Aid, arc.Author.Mid, mid, arc.TypeID)
				if err != nil {
					log.Error("s.ArcNote is err %+v", err)
				}
				if note != nil && note.NotesCount > 0 {
					v.NotesCount = note.NotesCount
				}
				//笔记编辑旧版本不下发
				if note != nil && note.NoteId == 0 {
					if (mobiApp == "android" && build < s.c.BuildLimit.NoteAndroidBuild) || (mobiApp == "iphone" && build < s.c.BuildLimit.NoteIOSBuild) {
						note = nil
					}
				}
				//是否为白名单拉起笔记浮层
				if note != nil && note.AutoPullCvid > 0 {
					v.ClientAction = &viewApi.PullClientAction{
						Type:       "note",
						PullAction: true,
						Params:     strconv.FormatInt(note.AutoPullCvid, 10),
					}
				}
				return nil
			})
			//bgm  高版本
			if isNewVersion {
				g.Go(func(ctx context.Context) error {
					musicRly, e := cfg.dep.Music.BgmEntrance(ctx, arc.Aid, cid, platform)
					if e != nil {
						log.Error("cfg.dep.Music.BgmEntrance(%d,%d) error(%v)", arc.Aid, cid, e)
						return nil
					}
					if musicRly != nil {
						v.RefreshSpecialCell = musicRly.MusicState == 1
						music = musicRly.MusicInfo
					}
					return nil
				})
			}
			//ugc tab3 后台配置
			if (plat == model.PlatAndroid || plat == model.PlatIPhone) && s.matchNGBuilder(mid, buvid, "view_tab") {
				g.Go(func(ctx context.Context) error {
					viewTab, err := cfg.dep.Resource.ViewTab(ctx, arc.Aid, tagIDs, upIDs, arc.TypeID, int32(plat), int32(build))
					if err != nil {
						if !ecode.EqualError(ecode.NothingFound, err) {
							log.Error("ViewTab err(%+v) aid(%d) tagids(%+v) upids(%+v) typeid(%d) plat(%d) build(%d)", err, arc.Aid, tagIDs, upIDs, arc.TypeID, plat, build)
						}
						return nil
					}
					v.ViewTab = viewTab
					return nil
				})
			}
			//必剪素材tag
			if funk.Contains([]int32{19, 20, 39}, arc.UpFromV2) {
				g.Go(func(ctx context.Context) (err error) {
					bijianMaterial, err = s.archiveMaterialDao.GetPlayerTag(ctx, arc.Aid)
					if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
						log.Error("s.archiveMaterialDao.GetPlayerTag error(%+v %+v)", arc.Aid, err)
						return nil
					}
					return nil
				})
			}
			//工具类标签
			g.Go(func(ctx context.Context) (err error) {
				dev, _ := device.FromContext(ctx)
				req := &v14.ArcMaterialListReq{
					Aid:      arc.GetAid(),
					Cid:      arc.GetFirstCid(),
					Source:   1,
					Platform: dev.RawPlatform,
					MobiApp:  mobiApp,
					Device:   dev.Device,
					Build:    int32(build),
				}
				toolMaterial, err = s.creativeMaterialDao.GetArcMaterialListTag(ctx, req)
				if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
					log.Error("s.creativeMaterialDao.GetArcMaterialListTag error(%+v %+v)", req, err)
					return nil
				}
				if toolMaterial != nil && toolMaterial.Type == _toolTagText && (plat == model.PlatIPhone && build < 68300000 || plat == model.PlatAndroid && build < 6800000) {
					toolMaterial = nil
					return nil
				}
				if toolMaterial != nil && toolMaterial.Type == _toolTagVideo && (plat == model.PlatIPhone && build < 68300000 || plat == model.PlatAndroid && build < 6820000) {
					toolMaterial = nil
					return nil
				}
				return nil
			})
			if err = g.Wait(); err != nil {
				log.Error("special cell is err %+v", err)
				return
			}
			//灵感话题
			inspirationMaterial, isOperation := s.InspirationTopicHandle(topicCell, extra)
			//aid返回灵感话题跳过灰度
			//服务端返回的灵感话题走灰度逻辑
			if s.c.Custom.InspirationTagSwitch && inspirationMaterial != nil && topicCell != nil {
				//未被灰度到的mid都跳转"话题页"
				if isOperation && (mid == 0 || !(mid%100 < s.c.Custom.InspirationTagGrey)) { //话题维度（运营维度）
					inspirationMaterial.Url = topicCell.JumpUrl
				}
			}
			//优先级 赛事 > ugc_tab ＞ 创作类(必剪 > 工具（合拍、特效、模板、一键成片） ＞ 灵感) > bgm > 笔记 > 电影
			// (新话题不参与优先级)
			var cellArr map[string]*viewApi.SpecialCell
			v.SpecialCell, cellArr = s.specialCellPriority(esportsCell, v.ViewTab, bijianMaterial, toolMaterial, inspirationMaterial)
			v.DescTag = FilterTopicNameV2(topicCell, arcTag, ogvChan) //简介tag（不包含活动tag)和ogvtag
			if isNewVersion {
				v.SpecialCellNew = view.SpecialCellPriorityNewVersion(ogvChan, music, note, spmid, topicCell, cellArr)
			} else {
				v.SpecialCellNew = s.specialCellPriorityNew(note, v.SpecialCell)
			}
			if music != nil {
				v.MaterialLeft = &viewApi.MaterialLeft{StaticIcon: view.PlayMusicStaticIcon, Text: music.MusicTitle, Url: music.JumpUrl, LeftType: "bgm", Param: music.MusicId, OperationalType: "7"}
				if !s.c.Custom.CloseMusicIcon {
					v.MaterialLeft.Icon = view.PlayMusicIcon
				}
			}
		}
	}
	return
}

// 灵感话题处理:运营配置灵感话题>服务返回的灵感话题
func (s *Service) InspirationTopicHandle(topicCell *viewApi.SpecialCell, extra map[string]string) (*view.InspirationMaterial, bool) {
	res := &view.InspirationMaterial{}
	isOperation := false
	if topicCell != nil {
		topicId, _ := strconv.Atoi(topicCell.Param)
		v, ok := s.inspirationMaterial[int64(topicId)] // 运营配置的灵感话题
		if ok {
			isOperation = true
			return &view.InspirationMaterial{
				Title:         v.Title,
				Url:           v.Url,
				InspirationId: v.InspirationId,
			}, isOperation
		}
	}
	if extra != nil {
		v, ok := extra[_inspirationKey]
		if !ok {
			return nil, isOperation
		}
		err := json.Unmarshal([]byte(v), res)
		if err != nil {
			log.Error("json.Unmarshal inspiration err %+v %+v", err, extra[_inspirationKey])
			return nil, isOperation
		}
		return res, isOperation
	}

	return nil, isOperation
}

// 新话题和channel标签名字去重
func FilterTopicNameV2(newTopic *viewApi.SpecialCell, channelTag []*tag.Tag, ogvChan *channelApi.Channel) []*tag.Tag {
	res := []*tag.Tag{}
	//去掉channel标签里的重名
	for _, v := range channelTag {
		if newTopic != nil && v.Name == newTopic.Text {
			continue
		}
		if ogvChan != nil && v.TagID == ogvChan.ID {
			continue
		}
		res = append(res, v)
	}
	return res
}

// 话题标签灰度
func (s *Service) NewTopicActTagGrey(buvid string) bool {
	//灰度总开关
	if !s.c.BuildLimit.NewTopicActTagGreySwitch {
		return false
	}
	if buvid == "" {
		return false
	}
	//获取最后一位
	content := buvid[len(buvid)-1:]
	i, err := strconv.Atoi(content)
	//如果是数字，最后一位是1或者3
	if err == nil {
		if i == 1 || i == 3 {
			return true
		}
	}
	return false
}

// 是否收进简介里灰度
func (s *Service) NewTopicGrey(mid int64) bool {
	//灰度开关
	if !s.c.BuildLimit.NewTopicGreySwitch {
		return false
	}
	//mid是否在白名单里
	if funk.ContainsInt64(s.c.Custom.NewTopicDescTagWhite, mid) {
		return true
	}
	//灰度逻辑
	if mid%20 < s.c.Custom.NewTopicDescGrey {
		return true
	}
	return false
}

// 是否在人群包中
func (s *Service) IsMidExists(c context.Context, mid int64, groups []*bgroup.MemberInReq_MemberInReqSingle) (map[string]bool, error) {
	req := &bgroup.MemberInReq{
		Member: strconv.FormatInt(mid, 10),
		Groups: groups,
	}
	reply, err := s.bGroupDao.GetMidExists(c, req)
	if err != nil {
		log.Error("s.bGroupDao.GetMidExists err:%+v", err)
		return nil, err
	}
	res := make(map[string]bool)
	for _, v := range reply {
		if v == nil {
			continue
		}
		res[v.Name] = v.In
	}
	return res, nil
}

// channel标签数据处理
// nolint:gocognit
func (s *Service) ResourceChannelsHandle(chans []*channelApi.Channel, plat int8, build int, actTagName, spmid string) (arcTag []*tag.Tag, actTag []*tag.Tag, ogvTag *channelApi.Channel, actTids []int64) {
	for _, t := range chans {
		// 电影频道
		var isNewOgv bool
		if t.BizType == channelApi.ChannelBizlType_MOVIE && ((plat == model.PlatIPhone && build >= s.c.BuildLimit.OGVChanIOSBuild) || (plat == model.PlatAndroid && build >= s.c.BuildLimit.OGVChanAndroidBuild)) {
			isNewOgv = true
		}
		if ogvTag == nil && isNewOgv { //第一个电影频道外露
			ogvTag = t
		}
		tempTag := &tag.Tag{TagID: t.ID, Name: t.Name, Cover: t.Cover, Likes: t.Likes, Hates: t.Hates, Liked: t.Liked, Hated: t.Hated, Attribute: t.Attribute}
		// 仅支持 iphone和安卓粉 548版本 版本号修改
		if t.CType == 2 && ((plat == model.PlatIPhone && build > s.c.BuildLimit.ChannelIOS) || (plat == model.PlatAndroid && build >= s.c.BuildLimit.ChannelAndroid)) {
			if isNewOgv {
				tempTag.URI = fmt.Sprintf("bilibili://feed/channel?biz_id=%d&biz_type=0&source=%s", t.ID, spmid)
			} else {
				tempTag.URI = "bilibili://pegasus/channel/v2/" + strconv.FormatInt(tempTag.TagID, 10) + "?tab=select"
			}
			tempTag.TagType = "new"
		} else {
			tempTag.URI = "bilibili://pegasus/channel/" + strconv.FormatInt(tempTag.TagID, 10)
			tempTag.TagType = "common"
		}
		if actTagName == tempTag.Name {
			tempTag.IsActivity = 1
			tempTag.TagType = "act"
			actTag = append(actTag, tempTag)
			// 是否活动tag
			if t.ActAttr == 1 {
				actTids = append(actTids, tempTag.TagID)
			}
		} else {
			//新版本普通标签要跳搜索落地页
			if tempTag.TagType == "common" {
				if (plat == model.PlatAndroid && build >= s.c.BuildLimit.NewTopicAndroidBuild) || (plat == model.PlatIPhone && build >= s.c.BuildLimit.NewTopicIOSBuild) ||
					(plat == model.PlatIpadHD && build >= s.c.BuildLimit.NewTopicIPadHDBuild) || (plat == model.PlatIPad && build >= s.c.BuildLimit.NewTopicIPadBuild) {
					encodeName := url.QueryEscape(t.Name)
					if strings.IndexByte(encodeName, '+') > -1 {
						encodeName = strings.Replace(encodeName, "+", "%20", -1)
					}
					tempTag.URI = "bilibili://search?keyword=" + encodeName
				}
			}
			arcTag = append(arcTag, tempTag)
		}
	}
	return arcTag, actTag, ogvTag, actTids
}

// 赛事标签
func (s *Service) eSportsS12(c context.Context, tagIds []int64, mid int64) (*viewApi.SpecialCell, error) {
	esportsReply, err := s.esportsDao.GetCapsuleCard(c, tagIds, mid)
	if err != nil {
		log.Error("eSportsS12 is err %+v %+v", err, tagIds)
		return nil, err
	}
	if esportsReply == nil {
		return nil, ecode.NothingFound
	}
	if esportsReply.IsMatch == 0 {
		return nil, ecode.NothingFound
	}
	res := &viewApi.SpecialCell{
		Text:             esportsReply.GetTitle(),
		TextColor:        view.CellTextColor,
		TextColorNight:   view.CellTextColorNight,
		CellType:         view.CellS11Type,
		EndIcon:          view.EndIcon,
		EndIconNight:     view.EndIconNight,
		CellBgcolor:      view.CellColor,
		CellBgcolorNight: view.CellColorNight,
		JumpUrl:          esportsReply.GetJumpUrl(),
		JumpType:         "fluid", //半浮层打开
		Icon:             view.IconS11,
		IconNight:        view.IconS11Night,
	}
	if esportsReply.IsLive == 1 { //直播中
		res.JumpType = "new_page"
	}
	return res, nil
}

// 赛事标签
func (s *Service) eSportsSpecialCell(c context.Context, upMid, aid int64) (*viewApi.SpecialCell, error) {
	bvID, _ := bvid.AvToBv(aid)
	//调用电竞的接口
	req := &esports.GetContestInfoByBvIdRequest{
		Mid:  upMid,
		BvId: bvID,
	}
	esportsReply, err := s.esportsDao.GetContestInfo(c, req)
	if err != nil {
		log.Error("GetContestInfo is err %+v %+v", err, req)
		return nil, err
	}
	return &viewApi.SpecialCell{
		Text:             esportsReply.GetContent(),
		TextColor:        view.CellTextColor,
		TextColorNight:   view.CellTextColorNight,
		CellType:         view.CellS11Type,
		EndIcon:          view.EndIcon,
		EndIconNight:     view.EndIconNight,
		CellBgcolor:      view.CellColor,
		CellBgcolorNight: view.CellColorNight,
		JumpUrl:          esportsReply.GetJumpUrl(),
		Param:            strconv.FormatInt(esportsReply.ContestId, 10),
		PageTitle:        esportsReply.GetTitle(),
		JumpType:         "fluid", //半浮层打开
		Icon:             view.IconS11,
		IconNight:        view.IconS11Night,
	}, nil
}

// 新话题标签
func (s *Service) newTopicSpecialCell(c context.Context, aid int64) (*viewApi.SpecialCell, error) {
	req := &topic.BatchResTopicByTypeReq{
		Type:   1,
		ResIds: []int64{aid},
	}
	reply, err := s.topicDao.GetBatchResTopicByType(c, req)
	if err != nil {
		log.Error("GetBatchResTopicByType is err %+v %+v", err, req)
		return nil, err
	}
	topicRes, ok := reply[aid]
	if !ok {
		return nil, ecode.NothingFound
	}
	return &viewApi.SpecialCell{
		Icon:             _iconNewTag,
		IconNight:        _iconNightNewTag,
		Text:             topicRes.GetName(),
		TextColor:        _textColorNewTag,
		TextColorNight:   _textColorNightNewTag,
		JumpUrl:          topicRes.GetJumpUrl(),
		CellType:         view.CellTopicType,
		CellBgcolor:      _cellColorNewTag,
		CellBgcolorNight: _cellColorNightNewTag,
		Param:            strconv.FormatInt(topicRes.Id, 10),
		JumpType:         "new_page", //新页面打开
	}, nil
}

func (s *Service) specialCellPriorityNew(note *notes.ArcTagReply, specialCell *viewApi.SpecialCell) []*viewApi.SpecialCell {
	res := []*viewApi.SpecialCell{}
	if note != nil && note.JumpLink != "" && note.TagShowText != "" {
		res = append(res, &viewApi.SpecialCell{
			Icon:             _iconNotes,
			IconNight:        _iconNotesNight,
			Text:             _notesText,
			TextColor:        _notesTextColor,
			TextColorNight:   _notesTextColorNight,
			CellType:         view.CellNoteType,
			CellBgcolor:      _notesCellColor,
			CellBgcolorNight: _notesCellColorNight,
			JumpUrl:          note.JumpLink + "&detail=ugc_useful_area",
			Param:            strconv.FormatInt(note.NoteId, 10),
			NotesCount:       note.NotesCount,
		},
		)
	}
	if specialCell != nil {
		res = append(res, specialCell)
	}
	return res
}

// 标签优先级
func (s *Service) specialCellPriority(eSports *viewApi.SpecialCell, viewTab *viewApi.Tab, bijianMaterial *v13.StoryPlayerRes, toolMaterial *v14.PlayPageMaterialTag, inspiration *view.InspirationMaterial) (*viewApi.SpecialCell, map[string]*viewApi.SpecialCell) {
	cellArr := make(map[string]*viewApi.SpecialCell)
	highPriorityCell := &viewApi.SpecialCell{}
	//灵感话题
	if inspiration != nil {
		inspirationTmp := &viewApi.SpecialCell{
			Text:             inspiration.Title,
			TextColor:        view.CellTextColor,
			TextColorNight:   view.CellTextColorNight,
			CellType:         view.CellInspirationType,
			EndIcon:          view.EndIcon,
			EndIconNight:     view.EndIconNight,
			CellBgcolor:      view.CellColor,
			CellBgcolorNight: view.CellColorNight,
			JumpUrl:          inspiration.Url,
			Param:            strconv.FormatInt(inspiration.InspirationId, 10),
			JumpType:         "new_page",
			Icon:             _inspirationTagIcon,
			IconNight:        _inspirationTagIconNight,
		}
		highPriorityCell = inspirationTmp
		cellArr[view.CellInspirationType] = inspirationTmp
	}
	//工具
	if toolMaterial != nil {
		tool := &viewApi.SpecialCell{
			Text:             toolMaterial.Tag,
			TextColor:        view.CellTextColor,
			TextColorNight:   view.CellTextColorNight,
			CellType:         view.CellToolType,
			EndIcon:          view.EndIcon,
			EndIconNight:     view.EndIconNight,
			CellBgcolor:      view.CellColor,
			CellBgcolorNight: view.CellColorNight,
			JumpUrl:          toolMaterial.JumpUrl,
			Param:            strconv.FormatInt(toolMaterial.Id, 10),
			JumpType:         "new_page",
		}
		if v, ok := toolTag[toolMaterial.Type]; ok {
			tool.Icon = v.Icon
			tool.IconNight = v.IconNight
		}
		highPriorityCell = tool
		cellArr[view.CellToolType] = tool
	}
	//必剪
	if bijianMaterial != nil {
		bijian := &viewApi.SpecialCell{
			Text:             bijianMaterial.Text,
			TextColor:        view.CellTextColor,
			TextColorNight:   view.CellTextColorNight,
			CellType:         view.CellBiJianType,
			EndIcon:          view.EndIcon,
			EndIconNight:     view.EndIconNight,
			CellBgcolor:      view.CellColor,
			CellBgcolorNight: view.CellColorNight,
			Icon:             _bijianTagIcon,
			IconNight:        _bijianTagIconNight,
			JumpUrl:          bijianMaterial.JumpUrl,
			JumpType:         "new_page",
		}
		highPriorityCell = bijian
		cellArr[view.CellBiJianType] = bijian
	}
	// 复用原播放页第三个tab后台配置的能力, 以配置内容为最高优先级
	if viewTab != nil {
		tab := &viewApi.SpecialCell{
			Text:             viewTab.Text,
			TextColor:        view.CellTextColor,
			TextColorNight:   view.CellTextColorNight,
			CellType:         view.CellUgcTabType,
			EndIcon:          view.EndIcon,
			EndIconNight:     view.EndIconNight,
			CellBgcolor:      view.CellColor,
			CellBgcolorNight: view.CellColorNight,
			JumpUrl:          viewTab.Uri,
			JumpType:         "fluid", //半浮层打开
			PageTitle:        viewTab.Text,
			Icon:             _iconS11,
			IconNight:        _iconNightS11,
		}
		highPriorityCell = tab
		cellArr[view.CellUgcTabType] = tab
	}
	//赛事
	if eSports != nil {
		highPriorityCell = eSports
		cellArr[view.CellS11Type] = eSports
	}
	return highPriorityCell, cellArr
}

//nolint:gomnd
func (s *Service) initDM(c context.Context, v *view.View) {
	const (
		_dmTypeAv    = 1
		_dmPlatMobie = 1
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	pLen := len(v.Pages)
	if pLen == 0 || pLen > 100 {
		return
	}
	if pLen > 50 {
		pLen = 50
	}
	cids := make([]int64, 0, len(v.Pages[:pLen]))
	for _, p := range v.Pages[:pLen] {
		cids = append(cids, p.Cid)
	}
	res, err := cfg.dep.Danmu.SubjectInfos(c, _dmTypeAv, _dmPlatMobie, cids...)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(res) == 0 {
		return
	}
	for _, p := range v.Pages[:pLen] {
		if r, ok := res[p.Cid]; ok {
			p.DM = r
		}
	}
}

func (s *Service) dealCM(c context.Context, advert *ad.Ad, resource int64) (adm map[int]*ad.AdInfo, err error) {
	ads, aids := advert.Convert(resource)
	if len(ads) == 0 {
		return
	}
	adm = make(map[int]*ad.AdInfo, len(ads))
	for _, ad := range ads {
		adm[ad.CardIndex-1] = ad
	}
	if len(aids) == 0 {
		return
	}
	as, err := s.arcDao.Archives(c, aids, 0, "", "")
	if err != nil {
		log.Error("%+v", err)
		err = nil
		return
	}
	for _, ad := range adm {
		if ad.Goto == model.GotoAv && ad.CreativeContent != nil {
			if a, ok := as[ad.CreativeContent.VideoID]; ok {
				ad.View = int(a.Stat.View)
				ad.Danmaku = int(a.Stat.Danmaku)
				ad.Author = a.Author
				ad.Stat = a.Stat
				ad.Duration = a.Duration
				if ad.CreativeContent.Desc == "" {
					ad.CreativeContent.Desc = a.Desc
				}
				ad.URI = model.FillURI(ad.Goto, ad.Param, cdm.ArcPlayHandler(a, nil, "", nil, 0, "", false))
			}
		}
	}
	return
}

func (s *Service) initPremiere(ctx context.Context, v *view.View, mid int64) {
	//判断聊天室是否风控
	roomId := v.Premiere.GetRoomId()
	if roomId > 0 {
		_, err := s.pgcDao.GetUGCPremiereRoomStatus(ctx, v.Premiere.GetRoomId())
		if err != nil {
			if !ecode.EqualError(xecode.PremiereRoomRisk, err) {
				log.Error("s.pgcDao.GetUGCPremiereRoomStatus is err %+v %+v", err, roomId)
			}
		}
		//房间被风控了，则去除首映信息
		if ecode.EqualError(xecode.PremiereRoomRisk, err) {
			v.PremiereRiskStatus = true
			return
		}
	}
	v.PremiereResource = &viewApi.PremiereResource{}
	//首映
	v.PremiereResource.Premiere = &viewApi.Premiere{
		PremiereState: viewApi.PremiereState(v.Premiere.State),
		StartTime:     v.Premiere.StartTime,
		ServiceTime:   time.Now().Unix(),
		RoomId:        v.Premiere.RoomId,
	}
	v.PremiereResource.Text = &viewApi.PremiereText{
		Title:            _premiereTextTitle,
		Subtitle:         _premiereTextSubTitle,
		OnlineText:       _premiereOnlineText,
		OnlineIcon:       _premiereOnlineIcon,
		OnlineIconDark:   _premiereOnlineIconDark,
		GuidancePulldown: _premiereGuidancePulldown,
		GuidanceEntry:    _premiereGuidanceEntry,
		IntroTitle:       _premiereIntroTitle,
		IntroIcon:        _premiereIntroIcon,
		IntroIconNight:   _premiereIntroIconNight,
	}
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", int64(67100000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(6710000))
	}).FinishOr(true) {
		v.PremiereResource.Text.OnlineIcon = _premiereOnlineIconNew
		v.PremiereResource.Text.OnlineIconDark = _premiereOnlineIconDarkNew
	}
	//aid去查活动id
	reply, err := s.actDao.GetPremiereSid(ctx, v.Aid)
	if err != nil {
		log.Error("s.actDao.GetPremiereSid is err %+v %+v", err, v.Aid)
		return
	}
	//活动id + 当前用户mid 查预约状态
	if reply.Sid > 0 {
		r, err := s.actDao.ReserveState(ctx, reply.Sid, mid)
		if err != nil {
			log.Error("s.actDao.ReserveState is err %+v %+v", err, v.Aid)
			return
		}
		//预约
		v.PremiereResource.Reserve = &viewApi.PremiereReserve{
			ReserveId: reply.Sid,
			Count:     r.Total,
			IsFollow:  r.IsFollow,
		}
	}
}

func (s *Service) dealRcmdRelate(c context.Context, plat int8, aid, mid int64, build int, mobiApp, device string) (rls []*view.Relate, err error) {
	var aids []int64
	if aids, err = s.arcDao.RelateAids(c, aid); err != nil {
		return
	}
	if len(aids) == 0 {
		return
	}
	var as map[int64]*api.Arc
	if as, err = s.arcDao.Archives(c, aids, mid, mobiApp, device); err != nil || len(as) == 0 {
		return
	}
	for _, aid := range aids {
		if a, ok := as[aid]; ok {
			if s.overseaCheck(a, plat) || !a.IsNormal() {
				continue
			}
			r := &view.Relate{}
			var (
				cooperation bool
				ogvURL      bool
			)
			if (model.IsAndroid(plat) && build > s.c.BuildLimit.CooperationAndroid) || (model.IsIPhone(plat) && build > s.c.BuildLimit.CooperationIOS) ||
				(model.IsIpadHD(plat) && build > s.c.BuildLimit.CooperationIPadHD) || (model.IsIPad(plat) && build > s.c.BuildLimit.CooperationIPad) {
				cooperation = true
			}
			if (plat == model.PlatAndroid && build > s.c.BuildLimit.OGVURLAndroid) || (plat == model.PlatIPhone && build > s.c.BuildLimit.OGVURLIOS) {
				ogvURL = true
			}
			r.FromAv(a, "", "", "", nil, cooperation, ogvURL, 0, "")
			rls = append(rls, r)
		}
	}
	return
}

func (s *Service) dealPremiereRelate(c context.Context, mid int64, plat int8, build int) ([]*view.Relate, error) {
	req := &upgrpc.ArcPassedReq{
		Mid:  mid,
		Pn:   1,
		Ps:   20,
		Sort: "desc",
	}
	res, err := s.arcDao.UpArchiveList(c, req)
	if err != nil {
		return nil, err
	}
	if len(res.GetArchives()) == 0 {
		return nil, ecode.NothingFound
	}
	upList := res.GetArchives()
	arc := &api.Arc{}
	rls := []*view.Relate{}
	for _, upArc := range upList {
		err := copier.Copy(arc, upArc)
		if err != nil {
			continue
		}
		if !arc.IsNormal() {
			continue
		}
		var (
			cooperation bool
			ogvURL      bool
		)
		r := &view.Relate{}
		if (model.IsAndroid(plat) && build > s.c.BuildLimit.CooperationAndroid) || (model.IsIPhone(plat) && build > s.c.BuildLimit.CooperationIOS) ||
			(model.IsIpadHD(plat) && build > s.c.BuildLimit.CooperationIPadHD) || (model.IsIPad(plat) && build > s.c.BuildLimit.CooperationIPad) {
			cooperation = true
		}
		if (plat == model.PlatAndroid && build > s.c.BuildLimit.OGVURLAndroid) || (plat == model.PlatIPhone && build > s.c.BuildLimit.OGVURLIOS) {
			ogvURL = true
		}
		r.FromAv(arc, "", "", "", nil, cooperation, ogvURL, 0, "")
		rls = append(rls, r)
	}
	return rls, nil
}

func (s *Service) dealGame(c context.Context, plat int8, build, gamecardStyleExp int, id int64, from string) (r *view.Relate) {
	if id < 1 {
		return nil
	}
	info, err := s.gameDao.Info(c, id, plat)
	if err != nil {
		log.Error("s.gameDao.Info err(%+v) id(%d) plat(%d)", err, id, plat)
		return nil
	}
	if info != nil && info.IsOnline {
		r = &view.Relate{}
		r.FromGame(c, s.c.Feature, info, from, plat, build, gamecardStyleExp)
	}
	return r
}

func recommendResHandle(res *view.RelateResV2) *advo.SunspotAdReplyForView {
	if res.BizData == nil || res.BizData.Data == nil {
		return nil
	}
	bizData := res.BizData.Data
	mixContents := make(map[int32]*advo.MixResourceContentDto)
	for key, info := range bizData.AdsInfo {
		if len(info) == 0 {
			continue
		}
		mixSourceContents := make(map[int32]*advo.MixSourceContentDto)
		for k, v := range info {
			if v == nil {
				continue
			}
			tmpMixSourceContent := &advo.MixSourceContentDto{
				AvId:      v.AvId,
				IsAd:      v.IsAd,
				CardIndex: v.CardIndex,
				CardType:  v.CardType,
			}
			if v.SourceContents != "" {
				decodeSourceContent, err := base64.StdEncoding.DecodeString(v.SourceContents)
				if err != nil {
					log.Error("SourceContent base64_decode is error(%+v) aid(%d)", err, v.AvId)
					continue
				}
				sourceContent := &types.Any{}
				if err := sourceContent.Unmarshal(decodeSourceContent); err != nil {
					log.Error("unmarshal is error(%+v) aid(%d) content(%s)", err, v.AvId, decodeSourceContent)
					continue
				}
				tmpMixSourceContent.SourceContent = sourceContent

			}
			mixSourceContents[k] = tmpMixSourceContent
		}
		mixContents[key] = &advo.MixResourceContentDto{
			ResourceId:     key,
			SourceContents: mixSourceContents,
		}
	}
	advertNew := &advo.SunspotAdReplyForView{
		ResourceContents: mixContents,
	}
	if bizData.TabInfo != nil {
		tabInfo := &advo.TabInfoDto{}
		tabInfo.TabName = bizData.TabInfo.TabName
		tabInfo.TabVersion = bizData.TabInfo.TabVersion
		extra, err := mappingToAny(bizData.TabInfo.Extra)
		if err != nil {
			log.Error("Failed to mapping tab extra to Any: %s, %+v", bizData.TabInfo.Extra, err)
		}
		tabInfo.Extra = extra
		advertNew.TabInfo = tabInfo
	}
	if bizData.AdsControl != "" {
		adsControl, err := mappingToAny(bizData.AdsControl)
		if err != nil {
			log.Error("Failed to mapping AdsControl to Any: %s, %+v", bizData.AdsControl, err)
			return advertNew
		}
		advertNew.AdsControl = adsControl
	}
	return advertNew
}

func mappingToAny(in string) (*types.Any, error) {
	decodeString, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return nil, err
	}
	out := &types.Any{}
	if err = out.Unmarshal(decodeString); err != nil {
		return nil, err
	}
	return out, nil
}

// nolint:gocognit
func (s *Service) newRcmdRelateV2(c context.Context, recommendReq *view.RecommendReq) (rls []*view.Relate, advNum string, playParam, pkCode int, userFeature, returnCode string, pvFeature json.RawMessage, relateConf *view.RelateConf, advertNew *advo.SunspotAdReplyForView, isRelateAd bool, adCode string, err error) {
	if recommendReq == nil {
		return
	}
	res, returnCode, err := s.arcDao.NewRelateAidsV2(c, recommendReq)
	//打点记录原地播放实验比例
	if recommendReq.InfeedPlay != 0 {
		s.prom.Incr("相关推荐-InFeedPlay-" + strconv.FormatInt(int64(recommendReq.InfeedPlay), 10))
	}
	if err != nil || res == nil || (recommendReq.InfeedPlay == 0 && len(res.Data) == 0) {
		return
	}
	userFeature = res.UserFeature
	pvFeature = res.PvFeature
	playParam = res.PlayParam
	pkCode = res.BizPkCode //pk_code
	advNum = res.BizAdvNum //库存上报
	relateConf = &view.RelateConf{
		GamecardStyleExp:   1,                // 是否展示游戏新卡, 已推全, 固定为1
		RelatesStyle:       res.RecReasonExp, // 全屏推荐理由样式，0 老样式、1 新样式
		GifExp:             res.GifExp,       // 是否命中gif cover实验
		FeedStyle:          res.FeedStyle,    //相关推荐展现样式
		FeedPopUp:          res.FeedPopUp,    //相关推荐无限下拉引导框
		FeedHasNext:        len(res.Data) != 0 && s.relateFeedHasNext(res.FeedStyle, returnCode),
		LocalPlay:          res.LocalPlay,
		RecThreePointStyle: int32(1),
		Next:               res.Next,
		RefreshIcon:        res.RefreshIcon,
		RefreshShow:        res.RefreshShow,
		RefreshText:        res.RefreshText,
		Refreshable:        res.Refreshable,
	}
	if res.BizData != nil {
		adCode = strconv.Itoa(res.BizData.Code)
	}
	//拼装成广告的返回参数
	advertNew = recommendResHandle(res)
	var (
		arcsPlayAv                      []*api.PlayAv
		ssIDs                           []int32
		epIDs                           []int32
		gameID, specialID, commercialId int64
		arcmV2                          map[int64]*api.ArcPlayer
		banm                            map[int32]*v1.CardInfoProto
		bangumiEp                       map[int32]*ogvgrpc.EpisodeCard
		gameInfos                       map[int64]*game.Game
		commercialInfos                 map[int64]*game.Game
		materialIds                     []int64
		multiMaterials                  map[int64]*resApiV2.Material
		ogvEpMaterialReply              map[int64]*v12.EpMaterial
	)
	ogvMaterialEpId := make(map[int64]int32)
	for _, rec := range res.Data {
		if rec.Goto != model.GotoAv {
			s.prom.Incr("相关推荐-首位卡-" + rec.Goto)
		}
		switch rec.Goto {
		case model.GotoAv:
			tmpAv := &api.PlayAv{
				Aid: rec.Oid,
			}
			if !s.relateNeedPlayer(len(arcsPlayAv), recommendReq.Build, recommendReq.Plat, recommendReq.PageIndex, res.FeedStyle) {
				tmpAv.NoPlayer = true
			}
			arcsPlayAv = append(arcsPlayAv, tmpAv)
			if rec.MaterialId > 0 {
				materialIds = append(materialIds, rec.MaterialId)
			}
		case model.GotoBangumi:
			if (recommendReq.Plat == model.PlatIPad && recommendReq.Build < _bangumiIPadBuild) || (recommendReq.Plat == model.PlatIpadHD && recommendReq.Build < _bangumiIPadHDBuild) {
				continue
			}
			ssIDs = append(ssIDs, int32(rec.Oid))
			if rec.MaterialId > 0 {
				materialIds = append(materialIds, rec.MaterialId)
			}
		case model.GotoBangumiEp:
			if filterOgvIpadHD(recommendReq.Plat, recommendReq.Build) {
				s.prom.Incr("相关推荐-ipadHD-过滤bangumi-ep")
				continue
			}
			epIDs = append(epIDs, int32(rec.Oid))
			if rec.MaterialId > 0 {
				materialIds = append(materialIds, rec.MaterialId)
			}
			if rec.MaterialId > 0 && rec.Source == "pgc" { //算法卡才去获取ogv的物料
				ogvMaterialEpId[rec.MaterialId] = int32(rec.Oid)
			}
			if rec.IsOgvEff > 0 && rec.MaterialId > 0 && rec.Source == "da" { //相关推荐OGV运营卡的效率池部分：goto='bangumi-ep'且source='da'且is_ogv_eff=1
				ogvMaterialEpId[rec.MaterialId] = int32(rec.Oid)
			}
		case model.GotoGame:
			gameID = rec.Oid
			if rec.MaterialId > 0 {
				materialIds = append(materialIds, rec.MaterialId)
			}
		case model.GotoSpecial:
			specialID = rec.Oid
			if rec.MaterialId > 0 {
				materialIds = append(materialIds, rec.MaterialId)
			}
			//针对特殊卡bangumi,ep和av需要校验合法性
			sp, reValue, parseErr := s.parseSpecialCardReValue(specialID)
			if parseErr != nil {
				continue
			}
			if model.OperateType[sp.ReType] == model.GotoEP {
				epIDs = append(epIDs, int32(reValue))
			}
			if model.OperateType[sp.ReType] == model.GotoBangumi {
				ssIDs = append(ssIDs, int32(reValue))
			}
			if model.OperateType[sp.ReType] == model.GotoAv {
				tmpAv := &api.PlayAv{
					Aid:      int64(reValue),
					NoPlayer: true,
				}
				arcsPlayAv = append(arcsPlayAv, tmpAv)
			}
		case model.GotoOrder: //商单
			if res.ArcCommercial == nil {
				continue
			}
			commercialId = res.ArcCommercial.Data.GameId
		case model.GotoCm: //推荐广告
			isRelateAd = true //是否有推荐广告
		}
	}
	eg := egv2.WithContext(c)
	mutex := sync.Mutex{}
	if len(arcsPlayAv) > 0 && s.miaokaiWithRelateCnt(res.FeedStyle) {
		eg.Go(func(ctx context.Context) (err error) {
			if arcmV2, err = s.arcDao.ArcsPlayer(ctx, arcsPlayAv); err != nil {
				log.Error("res.FeedStyle(%s), s.arcDao.ArcsPlayer err(%+v)", res.FeedStyle, err)
			}
			return nil
		})
	} else if len(arcsPlayAv) > 0 {
		s.prom.Incr("相关推荐秒开-" + res.FeedStyle)
		arcmV2 = make(map[int64]*api.ArcPlayer, len(arcsPlayAv))
		for _, part := range s.slicePlayAv(arcsPlayAv) {
			tmpPart := part
			eg.Go(func(ctx context.Context) (err error) {
				partArcsPlayer, err := s.arcDao.ArcsPlayer(ctx, tmpPart)
				if err != nil {
					log.Error("res.FeedStyle(%s), s.arcDao.ArcsPlayer err(%+v)", res.FeedStyle, err)
				} else {
					mutex.Lock()
					defer mutex.Unlock()
					for key, value := range partArcsPlayer {
						arcmV2[key] = value
					}
				}
				return nil
			})
		}
	}
	if len(ssIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if banm, err = s.banDao.CardsInfoReply(ctx, ssIDs); err != nil {
				log.Error("s.banDao.CardsInfoReply err(%+v)", err)
			}
			return nil
		})
	}
	if len(epIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			ogvReq := &ogvgrpc.EpCardsReq{
				EpId:     epIDs,
				BizScene: 2,
			}
			if bangumiEp, err = s.banDao.EpCardsInfo(ctx, ogvReq); err != nil {
				log.Error("s.banDao.EpCardsInfo err(%+v)", err)
			}
			return nil
		})
	}
	if len(ogvMaterialEpId) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			ogvEpMaterialReply, err = s.deliveryDao.GetBatchEpMaterial(ctx, ogvMaterialEpId)
			if err != nil {
				log.Error("s.deliveryDao.GetBatchEpMaterial err(%+v)", err)
			}
			return nil
		})
	}
	if gameID > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if gameInfos, err = s.gameDao.MultiGameInfos(ctx, []int64{gameID}, recommendReq.MobileApp, recommendReq.Build); err != nil {
				log.Error("s.gameDao.MultiGameInfos gameId(%d) recommendReq(%+v) err(%+v)", gameID, recommendReq, err)
			}
			return nil
		})
	}
	//商单物料
	if commercialId > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if commercialInfos, err = s.gameDao.MultiGameInfos(ctx, []int64{commercialId}, recommendReq.MobileApp, recommendReq.Build); err != nil {
				log.Error("s.gameDao.MultiGameInfos commercialId(%d) recommendReq(%+v) err(%+v)", commercialId, recommendReq, err)
			}
			return nil
		})
	}
	//获取物料信息
	if len(materialIds) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			multiMaterials, err = s.rscDao.MultiMaterials(c, materialIds)
			if err != nil {
				log.Error("s.rscDao.MultiMaterials materialIds(%+v), error(%+v)", materialIds, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v)", err)
	}
	var (
		cooperation bool
		ogvURL      bool
	)
	if (model.IsAndroid(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationAndroid) || (model.IsIPhone(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIOS) ||
		(model.IsIpadHD(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIPadHD) || (model.IsIPad(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIPad) {
		cooperation = true
	}
	if (recommendReq.Plat == model.PlatAndroid && recommendReq.Build > s.c.BuildLimit.OGVURLAndroid) || (recommendReq.Plat == model.PlatIPhone && recommendReq.Build > s.c.BuildLimit.OGVURLIOS) {
		ogvURL = true
	}
	hasCoverGif := false
	recAuthorName := ""
	for _, rec := range res.Data {
		r := &view.Relate{AvFeature: rec.AvFeature, Source: rec.Source, TrackID: rec.TrackID}
		switch rec.Goto {
		case model.GotoAv:
			arcwp, ok := arcmV2[rec.Oid]
			if !ok || arcwp == nil || arcwp.Arc == nil {
				continue
			}
			//up主name，在fromAV方法里可能会拼上其他字符，相关推荐负反馈不需要这些字符
			recAuthorName = arcwp.GetArc().GetAuthor().Name
			if s.overseaCheck(arcwp.Arc, recommendReq.Plat) || !arcwp.Arc.IsNormal() {
				continue
			}
			if hasCoverGif { // 控制只返回一张gif卡
				rec.CoverGif = ""
			}
			if rec.IsDalao > 0 {
				r.FromOperate(c, s.c.Feature, rec, arcwp.Arc, nil, nil, model.FromOperation, rec.TrackID,
					cooperation, ogvURL, recommendReq.Plat, recommendReq.Build, relateConf.GamecardStyleExp,
					relateConf.GifExp, nil, 0, multiMaterials)
			} else {
				playInfo := arcwp.PlayerInfo[arcwp.DefaultPlayerCid]
				r.FromAv(arcwp.Arc, "", rec.TrackID, rec.CoverGif, playInfo, cooperation, ogvURL, recommendReq.Build, recommendReq.MobileApp)
			}
			//如果有秒开，就要用秒开的cid
			if arcwp.DefaultPlayerCid > 0 {
				r.Cid = arcwp.DefaultPlayerCid
			}
		case model.GotoBangumi:
			ban, ok := banm[int32(rec.Oid)]
			if !ok {
				continue
			}
			//bangumi卡：运营卡+普通卡，历史原因不能通过is_dalao判断，新加source字段判断
			if rec.IsDalao > 0 && rec.Source == "da" {
				r.FromOperate(c, s.c.Feature, rec, nil, nil, nil, model.FromOperation, rec.TrackID,
					cooperation, false, recommendReq.Plat, recommendReq.Build, relateConf.GamecardStyleExp,
					relateConf.GifExp, ban, recommendReq.Aid, multiMaterials)
			} else {
				r.FromBangumi(ban, recommendReq.Aid, "", rec)
			}
		case model.GotoBangumiEp:
			banEp, ok := bangumiEp[int32(rec.Oid)]
			if !ok {
				continue
			}
			if rec.IsDalao > 0 && rec.Source == "da" {
				//source='da'、goto='bangumi-ep'会有多物料
				r.FromBangumiEpOperate(banEp, recommendReq.Aid, model.FromOperation, rec, multiMaterials, ogvEpMaterialReply)
			} else {
				r.FromBangumiEp(banEp, recommendReq.Aid, "", rec, multiMaterials, ogvEpMaterialReply)
			}
		case model.GotoGame:
			if len(gameInfos) == 0 || gameInfos[gameID] == nil || !gameInfos[gameID].IsOnline || gameInfos[gameID].GameLink == "" {
				errMsg, _ := json.Marshal(gameInfos)
				log.Error("日志告警 游戏卡 出现过滤 gameID(%d) gameInfos(%+v) build(%d）sdkType(%s)", gameID, string(errMsg), recommendReq.Build, gameDao.CastSDKType(recommendReq.MobileApp))
				continue
			}
			r.FromGameCard(rec, gameInfos[gameID], model.FromOperation, recommendReq.Plat, recommendReq.Build, multiMaterials, recommendReq.MobileApp, res.FeedStyle, s.c.Custom.PowerBadgeSwitch)
		case model.GotoSpecial:
			//针对特殊卡bangumi,ep和av需要校验合法性
			sp, reValue, parseErr := s.parseSpecialCardReValue(specialID)
			if parseErr != nil {
				continue
			}
			//校验ep/av合法性
			_, ok := bangumiEp[int32(reValue)]
			if model.OperateType[sp.ReType] == model.GotoEP && !ok {
				continue
			}
			_, ok = banm[int32(reValue)]
			if model.OperateType[sp.ReType] == model.GotoBangumi && !ok {
				continue
			}
			arc, ok := arcmV2[int64(reValue)]
			if model.OperateType[sp.ReType] == model.GotoAv && (!ok || arc == nil || arc.Arc == nil || !arc.Arc.IsNormalV2()) {
				continue
			}
			r.FromOperate(c, s.c.Feature, rec, nil, nil, sp, model.FromOperation, "", cooperation,
				false, recommendReq.Plat, recommendReq.Build, 0, relateConf.GifExp,
				nil, 0, multiMaterials)
		case model.GotoOrder:
			log.Info("稿件商单 aid:%d commercialId:%d", recommendReq.Aid, commercialId)
			if len(commercialInfos) == 0 || commercialInfos[commercialId] == nil || !commercialInfos[commercialId].IsOnline {
				errMsg, _ := json.Marshal(commercialInfos)
				log.Error("日志告警 商单游戏卡 出现过滤 commercialId(%d) gameInfos(%+v) build(%d）sdkType(%s)", commercialId, string(errMsg), recommendReq.Build, gameDao.CastSDKType(recommendReq.MobileApp))
				continue
			}
			r.FromGameCard(rec, commercialInfos[commercialId], model.FromOrder, recommendReq.Plat, recommendReq.Build, nil, recommendReq.MobileApp, res.FeedStyle, s.c.Custom.PowerBadgeSwitch)
		default: // 不识别的goto 不支持
			continue
		}
		r.ReasonStyleFrom(rec.RcmdReason, true)
		if rec.IsDalao > 0 {
			relateConf.HasDalao = int(rec.IsDalao)
		}
		if r.CoverGif != "" {
			hasCoverGif = true
		}
		//负反馈参数
		if recommendReq.RecStyle == 1 { //灰度逻辑
			r.RecThreePointStyle(rec, recAuthorName, relateConf.FeedStyle)
		}
		//dislike参数
		r.DislikeReportData = report.BuildDislikeReportData(r.MaterialId, r.UniqueId)
		// 商单放最前面
		if rec.Goto == model.GotoOrder {
			rls = append([]*view.Relate{r}, rls...)
		} else {
			rls = append(rls, r)
		}
	}
	return
}

// nolint:gocognit
func (s *Service) newRcmdRelate(c context.Context, plat int8, aid, mid, zoneID int64, buvid, mobiApp, sourcePage, trackid, cmd, tabid string,
	build, parentMode, autoplay, isAct int, isNewColor bool, pageVersion, fromSpmid string) (rls []*view.Relate, relateTab []*view.TabInfo, playParam int, userFeature, returnCode string, pvFeature json.RawMessage, relateConf *view.RelateConf, err error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	res, returnCode, err := cfg.dep.Archive.NewRelateAids(c, aid, mid, zoneID, build, parentMode, autoplay, isAct, buvid, sourcePage, trackid, cmd, tabid, plat, pageVersion, fromSpmid)
	if err != nil || res == nil || len(res.Data) == 0 {
		return
	}
	userFeature = res.UserFeature
	pvFeature = res.PvFeature
	relateTab = res.TabInfo
	playParam = res.PlayParam
	relateConf = &view.RelateConf{
		AutoplayCountdown:  res.AutoplayCountdown,
		ReturnPage:         res.ReturnPage,
		GamecardStyleExp:   res.GamecardStyleExp, // 是否展示游戏新卡
		AutoplayToast:      res.AutoplayToast,
		RelatesStyle:       res.RecReasonExp, // 全屏推荐理由样式，0 老样式、1 新样式
		GifExp:             res.GifExp,       // 是否命中gif cover实验
		RecThreePointStyle: int32(1),         // 相关推荐三点新样式
	}
	var (
		arcsPlayAv []*api.PlayAv
		ssIDs      []int32
		epIDs      []int32
		gameID     int64
		specialID  int64
		arcmV2     map[int64]*api.ArcPlayer
		banm       map[int32]*v1.CardInfoProto
		bangumiEp  map[int32]*ogvgrpc.EpisodeCard
		gameInfo   *game.Info
	)
	for _, rec := range res.Data {
		switch rec.Goto {
		case model.GotoAv:
			tmpAv := &api.PlayAv{
				Aid: rec.Oid,
			}
			if !s.relateNeedPlayer(len(arcsPlayAv), build, plat, 0, "") {
				tmpAv.NoPlayer = true
			}
			arcsPlayAv = append(arcsPlayAv, tmpAv)
		case model.GotoBangumi:
			if (plat == model.PlatIPad && build < _bangumiIPadBuild) || (plat == model.PlatIpadHD && build < _bangumiIPadHDBuild) {
				continue
			}
			ssIDs = append(ssIDs, int32(rec.Oid))
		case model.GotoBangumiEp:
			epIDs = append(epIDs, int32(rec.Oid))
		case model.GotoGame:
			gameID = rec.Oid
		case model.GotoSpecial:
			specialID = rec.Oid
		}
	}
	eg := egv2.WithContext(c)
	if len(arcsPlayAv) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if arcmV2, err = cfg.dep.Archive.ArcsPlayer(ctx, arcsPlayAv); err != nil {
				log.Error("s.arcDao.ArcsPlayer err(%+v)", err)
			}
			return nil
		})
	}
	if len(ssIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if banm, err = cfg.dep.PGC.CardsInfoReply(ctx, ssIDs); err != nil {
				log.Error("s.banDao.CardsInfoReply err(%+v)", err)
			}
			return nil
		})
	}
	if len(epIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			ogvReq := &ogvgrpc.EpCardsReq{
				EpId:     epIDs,
				BizScene: 2,
			}
			if bangumiEp, err = s.banDao.EpCardsInfo(ctx, ogvReq); err != nil {
				log.Error("s.banDao.EpCardsInfo err(%+v)", err)
			}
			return nil
		})
	}
	if gameID > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if gameInfo, err = cfg.dep.Game.Info(ctx, gameID, plat); err != nil {
				log.Error("s.gameDao.Info err(%+v)", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v)", err)
	}
	var (
		cooperation bool
		ogvURL      bool
	)
	if (model.IsAndroid(plat) && build > s.c.BuildLimit.CooperationAndroid) || (model.IsIPhone(plat) && build > s.c.BuildLimit.CooperationIOS) ||
		(model.IsIpadHD(plat) && build > s.c.BuildLimit.CooperationIPadHD) || (model.IsIPad(plat) && build > s.c.BuildLimit.CooperationIPad) {
		cooperation = true
	}
	if (plat == model.PlatAndroid && build > s.c.BuildLimit.OGVURLAndroid) || (plat == model.PlatIPhone && build > s.c.BuildLimit.OGVURLIOS) {
		ogvURL = true
	}
	hasCoverGif := false
	recAuthorName := ""
	for _, rec := range res.Data {
		r := &view.Relate{AvFeature: rec.AvFeature, Source: rec.Source, TrackID: rec.TrackID}
		switch rec.Goto {
		case model.GotoAv:
			arcwp, ok := arcmV2[rec.Oid]
			if !ok || arcwp == nil || arcwp.Arc == nil {
				continue
			}
			if s.overseaCheck(arcwp.Arc, plat) || !arcwp.Arc.IsNormal() {
				continue
			}
			//up主name，在fromAV方法里可能会拼上其他字符，相关推荐负反馈不需要这些字符
			recAuthorName = arcwp.GetArc().GetAuthor().Name
			if hasCoverGif { // 控制只返回一张gif卡
				rec.CoverGif = ""
			}
			if rec.IsDalao > 0 {
				r.FromOperate(c, s.c.Feature, rec, arcwp.Arc, nil, nil, model.FromOperation, rec.TrackID, cooperation, ogvURL, plat, build, relateConf.GamecardStyleExp, relateConf.GifExp, nil, 0, nil)
			} else {
				playInfo := arcwp.PlayerInfo[arcwp.DefaultPlayerCid]
				r.FromAv(arcwp.Arc, "", rec.TrackID, rec.CoverGif, playInfo, cooperation, ogvURL, build, mobiApp)
			}
		case model.GotoBangumi:
			ban, ok := banm[int32(rec.Oid)]
			if !ok {
				continue
			}
			//bangumi卡：运营卡+普通卡，历史原因不能通过is_dalao判断，新加source字段判断
			if rec.IsDalao > 0 && rec.Source == "da" {
				r.FromOperate(c, s.c.Feature, rec, nil, nil, nil, model.FromOperation, rec.TrackID, cooperation, false, plat, build, relateConf.GamecardStyleExp, relateConf.GifExp, ban, aid, nil)
			} else {
				r.FromBangumi(ban, aid, "", rec)
			}
		case model.GotoBangumiEp:
			banEp, ok := bangumiEp[int32(rec.Oid)]
			if !ok {
				continue
			}
			r.FromBangumiEp(banEp, aid, "", rec, nil, nil)
		case model.GotoGame:
			if gameInfo == nil || !gameInfo.IsOnline {
				continue
			}
			r.FromOperate(c, s.c.Feature, rec, nil, gameInfo, nil, model.FromOperation, "", cooperation, false, plat, build, relateConf.GamecardStyleExp, relateConf.GifExp, nil, 0, nil)
		case model.GotoSpecial:
			sp, ok := s.specialCache[specialID]
			if !ok {
				continue
			}
			r.FromOperate(c, s.c.Feature, rec, nil, nil, sp, model.FromOperation, "", cooperation, false, plat, build, 0, relateConf.GifExp, nil, 0, nil)
		default: // 不识别的goto 不支持
			log.Error("unknown relate goto:%s rec:%+v", rec.Goto, rec)
			continue
		}
		r.ReasonStyleFrom(rec.RcmdReason, isNewColor)
		if rec.IsDalao > 0 {
			relateConf.HasDalao = int(rec.IsDalao)
		}
		if r.CoverGif != "" {
			hasCoverGif = true
		}
		r.RecThreePointStyle(rec, recAuthorName, "")
		rls = append(rls, r)
	}
	return
}

func (s *Service) relateNeedPlayer(avLen, build int, plat int8, pageIndex int64, feedStyle string) bool {
	if s.miaokaiWithRelateCnt(feedStyle) && avLen < s.c.RelateCnt && ((model.IsAndroid(plat) && build > _qnAndroidBuildGt) || (model.IsIOSNormal(plat) && build > _qnIosBuildGt) || model.IsIPhoneB(plat)) {
		return true
	}
	//v1,v2单双列无限加载, 全部秒开. 当降级生效时, 指定个数秒开
	if (feedStyle == "v1" || feedStyle == "v2") && pageIndex > 0 && !s.c.Custom.MiaoKaiWithRelateCntSwitchOn {
		return true
	}
	return false
}

func (s *Service) miaokaiWithRelateCnt(feedStyle string) bool {
	//旧版本 或者 v3双列非无限加载 或者 降级生效时，指定个数秒开
	return feedStyle == "" || feedStyle == "default" || feedStyle == "v3" || s.c.Custom.MiaoKaiWithRelateCntSwitchOn
}

// get uri and tostring
func (s *Service) GetArcsPlayerURI(c context.Context, arg *viewApi.GetArcsPlayerReq, arcsPlayerReq []*api.PlayAv, deviceParams device.Device) ([]*viewApi.ArcsPlayer, error) {
	res := []*viewApi.ArcsPlayer{}
	arcmV2, err := s.arcDao.ArcsPlayer(c, arcsPlayerReq)
	if err != nil {
		log.Error("GetArcsPlayerURI s.arcDao.ArcsPlayer err(%+v)", err)
		return nil, err
	}
	for _, playAvsTmp := range arg.PlayAvs {
		aid := playAvsTmp.Aid
		arcPlayTmp, ok := arcmV2[aid]
		if !ok {
			continue
		}
		playerInfo := arcPlayTmp.GetPlayerInfo()
		if playerInfo == nil {
			continue
		}
		playInfoCid, ok := playerInfo[playAvsTmp.Cid]
		if !ok {
			continue
		}
		uri := model.FillURI("", strconv.FormatInt(aid, 10), cdm.ArcPlayHandler(arcmV2[aid].Arc, playInfoCid, "", nil, int(deviceParams.Build), deviceParams.RawMobiApp, true))
		resPlayerInfo := make(map[int64]string)
		resPlayerInfo[playAvsTmp.Cid] = uri
		res = append(res, &viewApi.ArcsPlayer{
			Aid:        aid,
			PlayerInfo: resPlayerInfo,
		})
	}
	return res, nil
}

// nolint:gocognit
func (s *Service) ContinuousPlayRelate(c context.Context, recommendReq *view.RecommendReq, infoParams *view.ContinuousInfo) (rls []*view.Relate, err error) {
	if recommendReq == nil {
		return
	}
	res, returnCode, err := s.arcDao.ContinuousPlayRelate(c, recommendReq)
	infoParams.ReturnCode = returnCode
	if err != nil || res == nil || len(res.Data) == 0 {
		return
	}
	infoParams.IsRec = 1
	infoParams.UserFeature = res.UserFeature
	if len(res.Data) > 0 {
		infoParams.TrackId = res.Data[0].TrackID
	}
	var (
		arcsPlayAv    []*api.PlayAv
		arcmV2        map[int64]*api.ArcPlayer
		cooperation   bool
		ogvURL        bool
		pos           int64 //卡片位置
		copyrightAids []int64
		copyrightRes  map[int64]api2.BanPlayEnum
	)
	for _, rec := range res.Data {
		switch rec.Goto {
		case model.GotoAv:
			//版权
			copyrightAids = append(copyrightAids, rec.Oid)
			//秒开
			tmpAv := &api.PlayAv{
				Aid: rec.Oid,
			}
			if !s.relateNeedPlayer(len(arcsPlayAv), recommendReq.Build, recommendReq.Plat, 0, "") {
				tmpAv.NoPlayer = true
			}
			arcsPlayAv = append(arcsPlayAv, tmpAv)
		}
	}
	eg := egv2.WithContext(c)
	if len(arcsPlayAv) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if arcmV2, err = s.arcDao.ArcsPlayer(ctx, arcsPlayAv); err != nil {
				log.Error("s.arcDao.ArcsPlayer err(%+v)", err)
			}
			return nil
		})
	}
	if len(copyrightAids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if copyrightRes, err = s.copyright.GetArcsBanPlay(ctx, copyrightAids); err != nil {
				log.Error("s.copyright.GetArcsBanPlay err(%+v)", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v)", err)
	}
	if (model.IsAndroid(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationAndroid) || (model.IsIPhone(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIOS) ||
		(model.IsIpadHD(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIPadHD) || (model.IsIPad(recommendReq.Plat) && recommendReq.Build > s.c.BuildLimit.CooperationIPad) {
		cooperation = true
	}
	if (recommendReq.Plat == model.PlatAndroid && recommendReq.Build > s.c.BuildLimit.OGVURLAndroid) || (recommendReq.Plat == model.PlatIPhone && recommendReq.Build > s.c.BuildLimit.OGVURLIOS) {
		ogvURL = true
	}
	for _, rec := range res.Data {
		r := &view.Relate{AvFeature: rec.AvFeature, Source: rec.Source, TrackID: rec.TrackID}
		switch rec.Goto {
		case model.GotoAv:
			arcwp, ok := arcmV2[rec.Oid]
			if !ok || arcwp == nil || arcwp.Arc == nil {
				continue
			}
			copyrightBan, ok := copyrightRes[rec.Oid]
			if ok && copyrightBan == api2.BanPlayEnum_IsBan {
				s.prom.Incr("copyright_ban")
				continue
			}
			if s.overseaCheck(arcwp.Arc, recommendReq.Plat) || !arcwp.Arc.IsNormal() {
				continue
			}
			if arcwp.Arc.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
				continue
			}
			//过滤互动视频
			if arcwp.Arc.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
				continue
			}
			if recommendReq.FromSpmid == _fromSpmidFullscreen && recommendReq.Spmid == _spmidFullscreen {
				//ipadHd全屏连播去掉合集和多p
				if arcwp.Arc.Videos > 1 || arcwp.Arc.SeasonID > 0 {
					continue
				}
			}
			if rec.IsDalao > 0 {
				r.FromOperate(c, s.c.Feature, rec, arcwp.Arc, nil, nil, model.FromOperation, rec.TrackID,
					cooperation, ogvURL, recommendReq.Plat, recommendReq.Build, 0,
					0, nil, 0, nil)
			} else {
				playInfo := arcwp.PlayerInfo[arcwp.DefaultPlayerCid]
				r.FromAv(arcwp.Arc, "", rec.TrackID, rec.CoverGif, playInfo, cooperation, ogvURL, recommendReq.Build, recommendReq.MobileApp)
			}
			rec.Pos = pos //标识卡片位置，透传给ai
			pos++
		default: // 不识别的goto 不支持
			log.Error("unknown continuous relate goto:%s rec:%+v", rec.Goto, rec)
			continue
		}
		rls = append(rls, r)
	}
	relateData, err := json.Marshal(res.Data)
	if err != nil {
		log.Error("marsh error %+v %+v", res.Data, err)
	}
	infoParams.ShowList = string(relateData)
	return
}

// 标签展示优先级：热门 > 活动 > 互动视频 > 百科
func (s *Service) initLabel(c context.Context, v *view.View, displaySteins bool) {
	var (
		_hot        = int32(1)
		_activity   = int32(2)
		_steinsGate = int32(3)
		_premiere   = int32(4)
	)
	if v.Premiere != nil {
		if v.Premiere.State == api.PremiereState_premiere_in {
			v.Label = &viewApi.Label{
				Type:        _premiere,
				Icon:        s.c.LabelIcon.Premiere.Icon,
				IconNight:   s.c.LabelIcon.Premiere.IconNight,
				IconWidth:   s.c.LabelIcon.Premiere.IconWidth,
				IconHeight:  s.c.LabelIcon.Premiere.IconHeight,
				Lottie:      s.c.LabelIcon.Premiere.Lottie,
				LottieNight: s.c.LabelIcon.Premiere.LottieNight,
			}
		}
	}
	if _, ok := s.hotAids[v.Aid]; ok {
		v.Label = &viewApi.Label{
			Type:       _hot,
			Uri:        model.FillURI(model.GotoHotPage, "", nil),
			Icon:       s.c.LabelIcon.Hot.Icon,
			IconNight:  s.c.LabelIcon.Hot.IconNight,
			IconWidth:  s.c.LabelIcon.Hot.IconWidth,
			IconHeight: s.c.LabelIcon.Hot.IconHeight,
		}
		return
	}
	if v.ActivityURL != "" {
		v.Label = &viewApi.Label{
			Type:       _activity,
			Uri:        v.ActivityURL,
			Icon:       s.c.LabelIcon.Act.Icon,
			IconNight:  s.c.LabelIcon.Act.IconNight,
			IconWidth:  s.c.LabelIcon.Act.IconWidth,
			IconHeight: s.c.LabelIcon.Act.IconHeight,
		}
		return
	}
	if displaySteins {
		v.Label = &viewApi.Label{
			Type:       _steinsGate,
			Icon:       s.c.LabelIcon.Steins.Icon,
			IconNight:  s.c.LabelIcon.Steins.IconNight,
			IconWidth:  s.c.LabelIcon.Steins.IconWidth,
			IconHeight: s.c.LabelIcon.Steins.IconHeight,
		}
		return
	}
	isEncyclopedia := s.rscDao.BWList(c, v.Aid)
	if !isEncyclopedia {
		return
	}
	v.Label = &viewApi.Label{
		Icon:       s.c.LabelIcon.Encyclopedia.Icon,
		IconNight:  s.c.LabelIcon.Encyclopedia.IconNight,
		IconWidth:  s.c.LabelIcon.Encyclopedia.IconWidth,
		IconHeight: s.c.LabelIcon.Encyclopedia.IconHeight,
	}
}

func (s *Service) aiRecommendAdParams(c context.Context, plat int8, spmid, buvid, from, trackid, mobiApp, network, adExtra, fromSpmid string,
	mid int64, build, parentMode, autoplay int, tids []int64, device, pageVersion string, viewRes *view.View,
	cfg viewConfig, disableRcmdMode int, deviceType int64, pageIndex int64, sessionId string, inFeedPlay, refreshNum, refreshType int32) (relateRsc int64, cmRsc int64, aiRecommendReq *view.RecommendReq) {
	//广告资源位
	relateRsc = _androidRelateRsc
	cmRsc = _androidCMRsc
	if model.IsIPhone(plat) {
		relateRsc = _iphoneRelateRsc
		cmRsc = _iphoneCMRsc
	}
	adRscs := []int32{int32(relateRsc)}
	if spmid != _playlistSpmid { // 播单页需屏蔽框下广告
		adRscs = append(adRscs, int32(cmRsc))
	}
	if model.IsIpadHD(plat) {
		adRscs = []int32{_iPadRelateRsc}
	}
	//ai推荐请求参数
	aiRecommendReq = &view.RecommendReq{
		Aid:         viewRes.Aid,
		Mid:         mid,
		ZoneId:      viewRes.ZoneID,
		Build:       build,
		ParentMode:  parentMode,
		AutoPlay:    autoplay, //此字段ai已经下线，待确定是否删除
		IsAct:       0,
		Buvid:       buvid,
		SourcePage:  from,
		TrackId:     trackid,
		Cmd:         model.RelateCmd,
		Plat:        plat,
		MobileApp:   mobiApp,
		Network:     network,
		AdExp:       1,
		Device:      device,
		RequestType: "wise",
		PageVersion: pageVersion,
		AdTab:       cfg.adTab,
		DisableRcmd: disableRcmdMode,
		DeviceType:  deviceType,
		PageIndex:   pageIndex,
		SessionId:   sessionId,
		Copyright:   viewRes.Copyright,
		InfeedPlay:  inFeedPlay,
		RefreshNum:  refreshNum,
		RefreshType: refreshType,
	}
	//是否出广告
	if (viewRes.AttrValV2(model.AttrBitV2CleanMode) == api.AttrNo) &&
		(model.IsIpadHD(plat) || model.IsIPhone(plat) || model.IsAndroid(plat)) && !model.IsOverseas(plat) {
		aiRecommendReq.IsAd = 1
		//资源位
		adResource := []string{}
		for _, v := range adRscs {
			tmp := strconv.Itoa(int(v))
			adResource = append(adResource, tmp)
		}
		aiRecommendReq.AdResource = strings.Join(adResource, ",")
		aiRecommendReq.AdExtra = adExtra
		aiRecommendReq.AvRid = viewRes.TypeID
		aiRecommendReq.AvPid = s.ArchiveTypesMap[viewRes.TypeID]
		aiRecommendReq.AvTid = xstr.JoinInts(tids)
		aiRecommendReq.AvUpId = viewRes.Author.Mid
		aiRecommendReq.FromSpmid = fromSpmid
		aiRecommendReq.Spmid = spmid
		aiRecommendReq.AdFrom = from
	}
	//是否出商单
	if (viewRes.AttrValV2(model.AttrBitV2CleanMode) == api.AttrNo) &&
		(viewRes.AttrVal(api.AttrBitIsPorder) == api.AttrYes || viewRes.OrderID > 0) {
		aiRecommendReq.IsCommercial = 1
	}
	//新样式灰度：通知广告
	if s.RecStyleGrey(mid, buvid) || funk.ContainsInt64(s.c.Custom.RecStyleWhite, mid) {
		aiRecommendReq.RecStyle = 1
	}
	//是否是付费合集视频
	if viewRes.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
		aiRecommendReq.IsArcPay = 1
		// 是否免费观看
		if viewRes.Arc.Rights.ArcPayFreeWatch == 1 {
			aiRecommendReq.IsFreeWatch = 1
		}
	}
	// 是否是蓝v
	if s.accDao.IsBlueV(c, viewRes.Author.Mid) {
		aiRecommendReq.IsUpBlue = 1
	}
	return
}

func (s *Service) RecStyleGrey(mid int64, buvid string) bool {
	//总开关
	if !s.c.Custom.RecStyleGreySwitch {
		return true
	}
	salt := "FHSjpZyY6M9LWUeC"
	str := strconv.FormatInt(mid, 10)
	//未登录用户用buvid
	if mid == 0 {
		str = buvid
	}
	str = str + salt
	//md5
	md5Res := md5.Sum([]byte(str))
	//转成16进制
	hexSix := hex.EncodeToString(md5Res[:])
	if len(hexSix) < int(18) {
		return false
	}
	//从第19位截取
	dst := hexSix[18:]
	//再转为10进制
	mm, err := strconv.ParseUint(dst, 16, 64)
	if err != nil {
		return false
	}
	//针对某个组，逐步放量
	return mm%15 == s.c.Custom.RecStyleAiGroup && int(crc32.ChecksumIEEE([]byte(buvid+salt))%100) < s.c.Custom.RecStyleGrey
}

// nolint:gocognit
func (s *Service) initRelateCMTagNewV2(c context.Context, v *view.View, plat int8, build, parentMode, autoplay int, mid int64,
	buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, trackid, filtered string, tids []int64, slocale,
	clocale, pageVersion string, cfg viewConfig, disableRcmdMode int, deviceType int64, pageIndex int64, sessionId,
	playMode string, inFeedPlay, refreshNum, refreshType int32) {
	var (
		rls            []*view.Relate
		advertNew      *advo.SunspotAdReplyForView
		relateConf     *view.RelateConf
		pkCode         int
		advNum, adCode string
		isRelateAd     bool
		err            error
	)
	// 审核版本，和有屏蔽推荐池属性的稿件下 后台连播 不出相关推荐任何信息
	if filtered == "1" || v.ForbidRec == 1 || playMode == "background" {
		log.Warn("no relates aid(%d) filtered(%s) ForbidRec(%d) PlayMode(%s)", v.Aid, filtered, v.ForbidRec, playMode)
		return
	}
	//请求参数赋值  推荐广告资源位+框下广告资源位+ai请求参数
	relateRsc, cmRsc, aiRecommendReq := s.aiRecommendAdParams(c, plat, spmid, buvid, from, trackid, mobiApp, network,
		adExtra, fromSpmid, mid, build, parentMode, autoplay, tids, device, pageVersion, v, cfg, disableRcmdMode, deviceType, pageIndex, sessionId, inFeedPlay, refreshNum, refreshType)
	v.RelatesInfoc = &view.RelatesInfoc{}
	v.RelatesInfoc.SetAdCode("NULL")
	if v.AttrVal(api.AttrBitIsPGC) == api.AttrYes && v.RedirectURL != "" {
		return
	}
	if mid > 0 || buvid != "" {
		rls, advNum, v.PlayParam, pkCode, v.UserFeature, v.ReturnCode, v.PvFeature,
			relateConf, advertNew, isRelateAd, adCode, err = s.newRcmdRelateV2(c, aiRecommendReq)
		if err != nil {
			log.Error("s.newRcmdRelateV2(%d) error(%+v)", v.Aid, err)
		}
		//处理ai返回的next标识
		if relateConf != nil && relateConf.Next != "" {
			pi, err := strconv.ParseInt(relateConf.Next, 10, 64)
			if err != nil {
				log.Error("日志告警 pagination aiResponse.Next parse error(%+v)", err)
			} else {
				v.Next = pagination.TokenGeneratorWithSalt(model.PaginationTokenSalt).GetPageToken(pi)
			}
		} else {
			s.prom.Incr("相关推荐-新分页参数-ai空Next-view")
		}

		//设置pk_code
		pkDesc, ok := view.PkCode[pkCode] //pk_code描述，用于prom
		if !ok {
			log.Error("pk_code desc is not exist(%d),(%+v)", pkCode, view.PkCode)
		}
		v.RelatesInfoc.SetPKCode(pkDesc)
		//ad_code
		v.RelatesInfoc.SetAdCode(adCode)
		//设置库存
		v.RelatesInfoc.SetAdNum(advNum)
		if relateConf != nil && v.Config != nil {
			v.Config.AutoplayCountdown = s.c.ViewConfig.AutoplayCountdown
			if relateConf.AutoplayCountdown > 0 {
				v.Config.AutoplayCountdown = relateConf.AutoplayCountdown
			}
			v.Config.RelatesStyle = relateConf.RelatesStyle
			v.Config.RelateGifExp = relateConf.GifExp
			v.Config.RecThreePointStyle = relateConf.RecThreePointStyle
			v.Config.FeedStyle = relateConf.FeedStyle
			v.Config.FeedPopUp = relateConf.FeedPopUp == 1
			v.Config.FeedHasNext = relateConf.FeedHasNext
			v.Config.LocalPlay = int32(relateConf.LocalPlay)
		}
		if relateConf != nil {
			v.RefreshPage = &viewApi.RefreshPage{
				Refreshable: int32(relateConf.Refreshable),
				RefreshText: relateConf.RefreshText,
				RefreshShow: relateConf.RefreshShow,
				RefreshIcon: int32(relateConf.RefreshIcon),
			}
		}
		if advertNew != nil && advertNew.TabInfo != nil && advertNew.TabInfo.TabVersion == 2 {
			//新样式
			v.CmUnderPlayer = advertNew.TabInfo.GetExtra()
		}
	}
	//灾备逻辑 -3和-5,-11不要取灾备数据
	s.prom.Incr("view_relate_return_code_" + v.ReturnCode)
	if len(rls) == 0 && inFeedPlay == 0 && v.ReturnCode != "-3" && v.ReturnCode != "-5" && v.ReturnCode != "-11" {
		rls, err = s.dealRcmdRelate(c, plat, v.Aid, mid, build, mobiApp, device)
		log.Warn("s.dealRcmdRelateV2 aid(%d) mid(%d) buvid(%s) build(%d) mobiApp(%s) device(%s) err(%+v)", v.Aid, mid, buvid, build, mobiApp, device, err)
	} else if v.ReturnCode != "600" { //600：ai推荐返回的数据是走的灾备
		v.IsRec = 1
	}
	//首映稿件相关推荐没返回则拿up主list数据作为推荐数据
	if len(rls) == 0 && v.Arc.AttrValV2(api.AttrBitV2Premiere) == api.AttrYes {
		rls, err = s.dealPremiereRelate(c, v.Author.Mid, plat, build)
		log.Warn("s.dealPremiereRelate aid(%d) mid(%d) build(%d) err(%+v)", v.Aid, mid, build, err)
	}
	var cm *viewApi.CM
	if advertNew != nil {
		if model.IsIpadHD(plat) {
			// iPadHD相关推荐广告是独立模块
			s.initCM4IPad(c, v, advertNew)
		} else {
			v.CMConfigNew = &viewApi.CMConfig{
				AdsControl: advertNew.AdsControl,
			}
			initCMS(v, advertNew, cmRsc)                   //框下广告
			adr := s.initRelateCM(c, advertNew, relateRsc) //推荐广告
			if adr != nil {
				//广告放到第一位
				if isRelateAd && adr.IsAd && adr.Aid != v.Aid {
					rls = append([]*view.Relate{adr}, rls...)
				}
				if advNum == "1S" {
					cm = adr.CM
				}
			}
		}
	}
	if len(rls) == 0 {
		s.prom.Incr("没有任何相关推荐")
		v.RelatesInfoc.SetPKCode(view.AdFirstForRelate0)
		return
	}
	v.Relates = uniqueRelates(v, rls, cm, false, plat, build)
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		for _, rl := range v.Relates {
			i18n.TranslateAsTCV2(&rl.Title)
		}
	}
}

// adFirst=true.  广告 > 运营卡片 > 商单 > AI推荐结果
// adFirst=false. 运营卡片 > 商单 > 广告 > AI推荐结果
func (s *Service) sortRelates(v *view.View, adFirst bool, ai []*view.Relate, dalao *view.Relate, business *view.Relate, adr *view.Relate, plat int8, build int) []*view.Relate {
	var (
		vr          []*view.Relate
		firstRelate *view.Relate
		cm          *viewApi.CM
	)
	if adFirst {
		firstRelate = s.sortRelatesAdFirst(v, adr)
	} else {
		firstRelate, cm = s.sortRelatesAdLast(v, dalao, business, adr)
	}
	if firstRelate != nil {
		vr = []*view.Relate{firstRelate}
	}
	// ai 推荐的优先级最低
	if ai != nil {
		vr = append(vr, ai...)
	}
	return uniqueRelates(v, vr, cm, true, plat, build)
}

func (s *Service) sortRelatesAdFirst(v *view.View, adr *view.Relate) *view.Relate {
	if v.Aid != adr.Aid {
		// 1A 表示第一位出广告
		s.prom.Incr("广告第一优先")
		v.RelatesInfoc.SetPKCode(view.AdFirstForRelate4)
		v.RelatesInfoc.SetAdNum("1A")
		return adr
	}
	v.RelatesInfoc.SetPKCode(view.AdFirstForRelate5)
	log.Error("相关推荐中的广告和视频出现了重复 aid:%d,ad:%+v", v.Aid, adr)
	return nil
}

func (s *Service) sortRelatesAdLast(v *view.View, dalao *view.Relate, business *view.Relate, adr *view.Relate) (*view.Relate, *viewApi.CM) {
	if dalao != nil {
		return dalao, nil
	} else if business != nil {
		v.RelatesInfoc.SetPKCode(view.AdFirstForRelate6)
		return business, nil
	} else if adr != nil {
		if adr.IsAd {
			if v.Aid != adr.Aid {
				v.RelatesInfoc.SetPKCode(view.AdFirstForRelate8)
				v.RelatesInfoc.SetAdNum("1A")
				return adr, nil
			}
			v.RelatesInfoc.SetPKCode(view.AdFirstForRelate7)
			log.Error("相关推荐中的广告和视频出现了重复 aid:%d,ad:%+v", v.Aid, adr)
			return nil, nil
		} else {
			v.RelatesInfoc.SetPKCode(view.AdFirstForRelate9)
			return nil, adr.CM
		}
	} else {
		v.RelatesInfoc.SetPKCode(view.AdFirstForRelate10)
	}
	return nil, nil
}

// 兜底过滤相关推荐内重复aid且不和当前播放页aid重复 注意番剧游戏等无aid
func uniqueRelates(v *view.View, relates []*view.Relate, cm *viewApi.CM, relateOld bool, plat int8, build int) []*view.Relate {
	if len(relates) == 0 {
		return nil
	}
	var (
		vr   []*view.Relate
		aids = make(map[int64]struct{}, len(relates))
	)
	aids[v.Aid] = struct{}{}
	for k, r := range relates {
		if _, ok := aids[r.Aid]; ok && r.Aid > 0 {
			log.Warn("relate duplicated aid:%d rl.aid:%d", v.Aid, r.Aid)
			continue
		}
		if filterOgvIpadHD(plat, build) && v.IPadCM != nil && v.IPadCM.Aid == r.Aid {
			log.Warn("relate duplicated iPadHD aid:%d rl.aid:%d", v.Aid, r.Aid)
			continue
		}
		if k == 0 && cm != nil { // 仅第一位会有保留信息，其他位不关心
			r.CM = cm
			if relateOld {
				v.RelatesInfoc.SetAdNum("1S")
			}
		}
		aids[r.Aid] = struct{}{}
		vr = append(vr, r)
	}
	return vr
}

func initCMS(v *view.View, advertNew *advo.SunspotAdReplyForView, resource int64) {
	rsc, ok := advertNew.ResourceContents[int32(resource)]
	if !ok {
		return
	}
	for _, rv := range rsc.SourceContents {
		v.CMSNew = append(v.CMSNew, &viewApi.CM{
			SourceContent: rv.SourceContent,
		})
	}
}

func (s *Service) initRelateCM(c context.Context, advertNew *advo.SunspotAdReplyForView, resource int64) *view.Relate {
	rsc, ok := advertNew.ResourceContents[int32(resource)]
	if !ok {
		return nil
	}
	var (
		aids []int64
		err  error
	)
	as := make(map[int64]*api.Arc)
	for _, rv := range rsc.SourceContents {
		if rv.AvId > 0 {
			aids = append(aids, rv.AvId)
		}
	}
	if len(aids) > 0 {
		if as, err = s.arcDao.Archives(c, aids, 0, "", ""); err != nil {
			log.Error("initRelateCM %+v", err)
		}
	}
	for _, rv := range rsc.SourceContents {
		if rv.CardIndex != 1 {
			continue
		}
		adr := &view.Relate{
			Aid:  rv.AvId,
			Goto: model.GotoCm,
			CM: &viewApi.CM{
				SourceContent: rv.SourceContent,
			},
			IsAd: rv.IsAd, // 返回是否包含广告内容
		}
		if a, ok := as[rv.AvId]; ok {
			adr.Author = &a.Author
			adr.Stat = a.Stat
			adr.Duration = a.Duration
		}
		return adr
	}
	return nil
}

func (s *Service) initHonor(c context.Context, v *view.View, plat int8, build int, mobiApp string, device string) {
	var (
		honor        *ahApi.Honor
		channelHonor *channelApi.ResourceHonor
		eg           = egv2.WithContext(c)
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	eg.Go(func(ctx context.Context) (err error) {
		honors, err := cfg.dep.ArchiveHornor.Honors(ctx, v.Aid, int64(build), mobiApp, device)
		if err != nil {
			log.Error("s.ahDao.Honors aid:%d err:%+v", v.Aid, err)
			return nil
		}
		if len(honors) == 0 || honors[0] == nil {
			return nil
		}
		if _, ok := model.DisplayHonor[honors[0].Type]; !ok {
			return nil
		}
		// 排行>3的暂不展示 放在配置里
		if honors[0].Type == ahApi.TypeRank && (v.Stat.HisRank > s.c.Custom.HonorRank || v.Stat.HisRank == 0) {
			return nil
		}
		// 兼容548客户端没有做展示互斥逻辑
		if honors[0].Type == ahApi.TypeRank && ((plat == model.PlatIPhone && build <= s.c.BuildLimit.HonorRankIOS) || (plat == model.PlatAndroid && build <= s.c.BuildLimit.HonorRankAndroid)) {
			return nil
		}
		honor = honors[0]
		return nil
	})
	if !model.IsIPad(plat) && v.Honor == nil {
		eg.Go(func(ctx context.Context) (err error) {
			if channelHonor, err = cfg.dep.Channel.ChannelHonor(ctx, v.Aid); err != nil {
				log.Error("s.channelDao.ChannelHonor aid:%d err:%+v", v.Aid, err)
				return nil
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		log.Error("eg.wait err(%+v)", err)
	}
	if honor != nil {
		v.Honor = view.FromHonor(honor)
		return
	}
	if channelHonor != nil {
		v.Honor = view.FromHonor(&ahApi.Honor{
			Aid:  v.Aid,
			Type: ahApi.TypeChannel,
			Url:  channelHonor.Url,
			Desc: channelHonor.Desc,
		})
	}
}

func (s *Service) initCM4IPad(c context.Context, v *view.View, advertNew *advo.SunspotAdReplyForView) {
	rsc, ok := advertNew.ResourceContents[_iPadRelateRsc]
	if !ok {
		return
	}
	mixSourceContent, ok := rsc.SourceContents[_iPadSourceId]
	if !ok {
		return
	}
	if mixSourceContent.AvId == v.Aid {
		return
	}
	v.IPadCM = &viewApi.CmIpad{
		Cm: &viewApi.CM{
			SourceContent: mixSourceContent.SourceContent,
		},
	}
	s.prom.Incr("广告位: ipad_cm")
	if mixSourceContent.AvId == 0 {
		return
	}
	reply, err := s.arcDao.Archive(c, mixSourceContent.AvId)
	if err != nil || reply == nil {
		log.Error("s.arcDao.Archive is err(%+v)", err)
		return
	}
	v.IPadCM.Aid = reply.Aid
	v.IPadCM.Author = &reply.Author
	v.IPadCM.Stat = &reply.Stat
	v.IPadCM.Duration = reply.Duration
}

func (s *Service) popupConfig(ctx context.Context, mid int64, buvid string) bool {
	cfg := FromContextOrCreate(ctx, s.defaultViewConfigCreater())

	_, ok := s.c.PopupWhiteList[strconv.FormatInt(mid, 10)]
	if !ok && !cfg.popupExp { //没有命中白名单也没有命中abtest
		return false
	}
	if feature.GetBuildLimit(ctx, "service.popup", nil) { //628版本之后24h逻辑由客户端来判断
		return true
	}
	conn := s.onlineRedis.Get(ctx)
	defer conn.Close()
	key := fmt.Sprintf("pop_%s", buvid)
	reply, err := redis.String(conn.Do("SET", key, 1, "EX", s.c.Custom.PopupExTime, "NX"))
	if err != nil || reply != "OK" {
		return false
	}
	return true
}

func (s *Service) buvidABTest(ctx context.Context, buvid string, flag *ab.StringFlag) bool {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", buvid))
	exp := flag.Value(t)
	return exp == "1" || exp == "11"
}

func (s *Service) SmallWindowConfig(ctx context.Context, buvid string, flag *ab.StringFlag) bool {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", buvid))
	exp := flag.Value(t)
	//B组 + BB组
	if exp == "2" || exp == "22" {
		return true
	}
	return false
}

//nolint:gomnd
func (s *Service) initOnline(v *view.View, buvid string, mid int64, aid int64) {
	//功能开关
	if s.c.OnlineCtrl == nil || !s.c.OnlineCtrl.SwitchOn {
		return
	}
	v.Online = &viewApi.Online{
		PlayerOnlineLogo: s.c.OnlineCtrl.Logo,
	}
	//白名单+灰度判断
	_, ok := s.c.OnlineCtrl.Mid[strconv.FormatInt(mid, 10)]
	group := crc32.ChecksumIEEE([]byte(buvid+"_online_ctrl")) % 100
	if ok || group < uint32(s.c.OnlineCtrl.Gray) {
		//查看稿件是否屏蔽
		v.Online.OnlineShow = !s.HitBlackList(aid)
	}
}

func (s *Service) HitBlackList(aid int64) bool {
	if _, ok := s.onlineBlackList[aid]; ok {
		return true
	}
	return false
}

func (s *Service) RelatesBiserialConfig(ctx context.Context, build int64, mobiApp, device, buvid string) bool {
	versionMatch := pd.WithDevice(
		pd.NewCommonDevice(mobiApp, device, "", build),
	).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPadHD().And().Build(">=", 33600001)
	}).FinishOr(false)
	if !versionMatch {
		return false
	}
	return s.RelatesBiserialABTest(ctx, buvid, relatesBiserialABtest)
}

func (s *Service) RelatesBiserialABTest(ctx context.Context, buvid string, flag *ab.StringFlag) bool {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", buvid))
	exp := flag.Value(t)
	//1是是实验组，为0时是对照组
	return exp == "1"
}

func (s *Service) slicePlayAv(origin []*api.PlayAv) [][]*api.PlayAv {
	batchCount := 10
	batchNum := len(origin) / batchCount
	if len(origin)%batchCount != 0 {
		batchNum = batchNum + 1
	}
	var res = make([][]*api.PlayAv, batchNum)
	j := 0
	for i := 0; i < len(origin); i += batchCount {
		if i+batchCount > len(origin) {
			// 不足一批次
			res[j] = origin[i:]
		} else {
			// 满一批次
			res[j] = origin[i : i+batchCount]
		}
		j++
	}
	return res
}

// filterOgvIpadHD 老版本pad不下发ogv卡片
func filterOgvIpadHD(plat int8, build int) bool {
	return (plat == model.PlatIPad && build < 66400001) || (plat == model.PlatIpadHD && build < 33700001)
}

func (s *Service) relateFeedHasNext(feedStyle string, returnCode string) bool {
	s.prom.Incr("相关推荐加载-" + feedStyle)
	return (feedStyle == "v1" || feedStyle == "v2") && returnCode != "3" && returnCode != "11"
}

func (s *Service) parseSpecialCardReValue(specialID int64) (*special.Card, int, error) {
	sp, ok := s.specialCache[specialID]
	if !ok || sp == nil {
		log.Error("日志告警 特殊卡 specialCache不存在 specialID(%d)", specialID)
		return nil, 0, errors.New("nothing found")
	}

	if model.OperateType[sp.ReType] == model.GotoEP || model.OperateType[sp.ReType] == model.GotoBangumi ||
		model.OperateType[sp.ReType] == model.GotoAv {
		reValue, err := strconv.Atoi(sp.ReValue)
		if err != nil {
			log.Error("日志告警 特殊卡 sp.ReValue解析失败 reType(%+v) err(%+v)", sp, err)
		}
		return sp, reValue, err
	}
	return sp, 0, nil
}
