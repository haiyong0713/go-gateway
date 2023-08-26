package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	jsonselect "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/select"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accAPI "git.bilibili.co/bapis/bapis-go/account/service"
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	feedMgr "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	vipgrpc "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/pkg/errors"
)

func cardKey(cardType, cardGoto string) string {
	return fmt.Sprintf("%s:%s", cardType, cardGoto)
}

func (s *Service) matchNGBuilder(mid int64, buvid, cardType, cardGoto string, posRecID int64, isAI bool) bool {
	if mid == 0 || buvid == "" {
		return false
	}
	if !isAI {
		return false
	}
	if posRecID != 0 {
		return false
	}
	if s.c.NgSwitch.DisableAll {
		return false
	}
	policy, ok := s.c.NgSwitch.CardSharding[cardKey(cardType, cardGoto)]
	if !ok {
		return false
	}
	if len(policy) == 0 {
		return true
	}
	for _, v := range policy {
		fn, err := parsePolicy(v)
		if err != nil {
			log.Error("Failed to parse policy: %+v", err)
			continue
		}
		if fn(mid, buvid) {
			return true
		}
	}
	return false
}

func parsePolicy(in string) (func(mid int64, buvid string) bool, error) {
	parts := strings.Split(in, ":")
	//nolint:gomnd
	if len(parts) != 2 {
		return nil, errors.Errorf("Invalid policy: %q", in)
	}
	switch parts[0] {
	case "mid":
		matchMid, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			if mid <= 0 {
				return false
			}
			return matchMid == mid
		}, nil
	case "buvid":
		matchBuvid := parts[1]
		return func(mid int64, buvid string) bool {
			if buvid == "" {
				return false
			}
			return matchBuvid == buvid
		}, nil
	case "mid_mod":
		mmParts := strings.Split(parts[1], ",")
		//nolint:gomnd
		if len(mmParts) != 2 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		mod, err := strconv.ParseInt(mmParts[0], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		pivtoal, err := strconv.ParseInt(mmParts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			if mid <= 0 {
				return false
			}
			return mid%mod <= pivtoal
		}, nil
	case "buvidcrc32_mod":
		bcmParts := strings.Split(parts[1], ",")
		//nolint:gomnd
		if len(bcmParts) != 2 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		mod, err := strconv.ParseInt(bcmParts[0], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		pivtoal, err := strconv.ParseInt(bcmParts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			if buvid == "" {
				return false
			}
			return int64(crc32.ChecksumIEEE([]byte(buvid)))%mod <= pivtoal
		}, nil
	default:
		return nil, errors.Errorf("Invalid policy: %q", in)
	}
}

func (s *Service) buildCardNG(ctx cardschema.FeedContext, cardType, cardGoto string, item *ai.Item, index int64, materials *Materials, infoc *feed.Infoc) (cardschema.FeedCard, bool) {
	if item.Ad != nil {
		item.SetCardStatusAd(item.Ad)
	}
	fn, ok := CardMap[cardKey(cardType, cardGoto)]
	if !ok {
		log.Error("Failed to find card Handler: %s, %s", cardType, cardGoto)
		FillDiscard(item.ID, item.Goto, feed.DiscardReasonCannotFindCardHandler, "", infoc)
		return nil, false
	}
	out, err := fn(ctx, index, item, materials, s.c.Feed)
	if err != nil {
		log.Error("Failed to build card: %+v", errors.WithStack(err))
		FillDiscard(item.ID, item.Goto, feed.DiscardReasonCannotBuildCard, err.Error(), infoc)
		return nil, false
	}
	return out, true
}

func buildFeedCtx(ctx context.Context, param *feed.IndexParam, mid int64, isAttentionStore map[int64]int8) cardschema.FeedContext {
	indexParam := buildFeedIndex(param, "", "")
	dev, _ := device.FromContext(ctx)
	userSession := feedcard.NewUserSession(mid, isAttentionStore, indexParam)
	fCtx := feedcard.NewFeedContext(userSession, feedcard.NewCtxDevice(&dev), time.Now())
	return fCtx
}

func buildFeedIndex(param *feed.IndexParam, applist, deviceInfo string) *feedcard.IndexParam {
	return &feedcard.IndexParam{
		Idx:           param.Idx,
		Pull:          param.Pull,
		Column:        param.Column,
		LoginEvent:    param.LoginEvent,
		OpenEvent:     param.OpenEvent,
		BannerHash:    param.BannerHash,
		AdExtra:       param.AdExtra,
		Interest:      param.Interest,
		Flush:         param.Flush,
		AutoPlayCard:  param.AutoPlayCard,
		DeviceType:    param.DeviceType,
		ParentMode:    param.ParentMode,
		RecsysMode:    param.RecsysMode,
		TeenagersMode: param.TeenagersMode,
		LessonsMode:   param.LessonsMode,
		DeviceName:    param.DeviceName,
		AccessKey:     param.AccessKey,
		ActionKey:     param.ActionKey,
		Statistics:    param.Statistics,
		Appver:        param.Appver,
		Filtered:      param.Filtered,
		AppKey:        param.AppKey,
		HttpsUrlReq:   param.HttpsUrlReq,
		InterestV2:    param.InterestV2,
		SplashID:      param.SplashID,
		Guidance:      param.Guidance,
		AppList:       applist,
		DeviceInfo:    deviceInfo,
		DisableRcmd:   param.DisableRcmd,
	}
}

func isHandlerEqual(left, right appcard.Handler) (bool, string, string, error) {
	leftTmpByte, _ := json.Marshal(left)
	rightTmpByte, _ := json.Marshal(right)
	var leftTmp map[string]interface{}
	if err := json.Unmarshal(leftTmpByte, &leftTmp); err != nil {
		log.Error("Failed to unmarshal: %s, %+v", string(leftTmpByte), err)
		return false, "", "", err
	}
	var rightTmp map[string]interface{}
	if err := json.Unmarshal(rightTmpByte, &rightTmp); err != nil {
		log.Error("Failed to unmarshal: %s, %+v", string(rightTmpByte), err)
		return false, "", "", err
	}
	delete(rightTmp, "three_point_v2")
	delete(rightTmp, "three_point")
	leftByte, _ := json.Marshal(leftTmp)
	rightByte, _ := json.Marshal(rightTmp)
	if bytes.Equal(leftByte, rightByte) {
		return true, "", "", nil
	}
	return false, string(leftByte), string(rightByte), nil
}

var CardMap = map[string]func(cardschema.FeedContext, int64, *ai.Item, *Materials, *conf.Feed) (cardschema.FeedCard, error){
	"one_pic_v3:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildOnePicV3FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"one_pic_v2:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildOnePicV2FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"three_pic_v3:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildThreePicV3FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"three_pic_v2:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildThreePicV2FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:article_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromArticle(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromLiveRoom(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:bangumi": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromBangumi(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v6:inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV6FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v5:inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV5FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v7:inline_pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV7FromPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v8:inline_live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV8FromLiveRoom(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v9:inline_av_v2": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV9FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"storys_v2:ai_story": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildStorysV2(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"select:follow_mode": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSelectFromFollowMode(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_v5:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerV5(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v4:bangumi_rcmd": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV4FromBangumiRcmd(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v7:tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV7FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v7:vip_renew": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV7FromVip(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdAv(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_web_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdWebS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_web": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdWeb(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_player": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"one_pic_v1:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildOnePicV1FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"three_pic_v1:picture": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildThreePicV1FromPicture(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"storys_v1:ai_story": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildStorysV1(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v1:av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV1FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v1:live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV1FromLiveRoom(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v1:pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV1FromPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v1:bangumi": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV1FromBangumi(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v1:bangumi_rcmd": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV1FromBangumiRcmd(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_v4:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerV4(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v6:vip_renew": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV6FromVip(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v6:tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV6FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"three_item_h_v3:article_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildThreeItemHV3FromArticle(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_v6:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerV6(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v1:ad_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1AdAv(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v1:ad_web_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1AdWebS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v1:ad_web": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1AdWeb(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v7:bangumi": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV7(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v8:live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV8(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v9:av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV9(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"ogv_small_cover:bangumi": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildOgvSmallCoverFromPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v7:inline_pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV7WithInlinePGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v8:inline_live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV8WithInlineLive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v9:inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV9(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v7:pgc": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV7WithPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v9:live": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV9FromLiveRoom(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_web_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV1AdWebS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_v8:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerV8(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_single_v8:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerSingleV8(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"banner_ipad_v8:banner": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildBannerIPadV8(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"notify_tunnel_v1:new_tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildNotifyTunnelV1FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"notify_tunnel_single_v1:new_tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildNotifyTunnelSingleV1FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"notify_tunnel_large_v1:big_tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildNotifyTunnelLargeV1FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"notify_tunnel_large_single_v1:big_tunnel": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildNotifyTunnelLargeSingleV1FromTunnel(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v7:special_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV7FromSpecialS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v8:special_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV8FromSpecialS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_single_v9:special_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverSingleV9FromSpecialS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"large_cover_v1:article_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildLargeCoverV1FromArticle(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_double_v9:ad_inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV9FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v9:ad_inline_av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV9FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v10:game": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV10FromGame(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_web_gif_reservation": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2FromReservation(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_web_gif_reservation": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1FromReservation(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_player_reservation": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2FromReservation(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_player_reservation": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1FromReservation(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v1:ad_player": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV1AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_player": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV1AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v2:special_s": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV2FromSpecialS(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"small_cover_v11:av": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildSmallCoverV11FromArchive(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_inline_3d": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_inline_3d": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV1AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_inline_3d_v2": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_inline_3d_v2": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV1AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_ogv": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2FromPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_double_v7:ad_inline_ogv": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV7FromInlinePGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v7:ad_inline_ogv": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV7WithPGC(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_v2:ad_inline_eggs": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmV2AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
	"cm_single_v1:ad_inline_eggs": func(feedContext cardschema.FeedContext, i int64, item *ai.Item, materials *Materials, feed *conf.Feed) (cardschema.FeedCard, error) {
		return feedcard.BuildCmSingleV1AdPlayer(feedContext, i, item, setFanoutResult(materials, feed))
	},
}

func setFanoutResult(materials *Materials, feed *conf.Feed) *feedcard.FanoutResult {
	out := &feedcard.FanoutResult{
		Archive: struct {
			Archive      map[int64]*api.ArcPlayer
			StoryArchive map[int64]*api.ArcPlayer
		}{Archive: materials.Archive, StoryArchive: materials.StoryArchive},
		Tag:     materials.Tag,
		Dynamic: struct{ Picture map[int64]*bplus.Picture }{Picture: materials.Picture},
		Account: struct {
			Card            map[int64]*accAPI.Card
			RelationStatMid map[int64]*relationgrpc.StatReply
			IsAttention     map[int64]int8
		}{Card: materials.AccountCard, RelationStatMid: materials.RelationStatMid, IsAttention: materials.IsAttention},
		Article: materials.Article,
		Live: struct {
			Room       map[int64]*live.Room
			InlineRoom map[int64]*live.Room
		}{Room: materials.Room, InlineRoom: materials.InlineRoom},
		Channel: materials.Channel,
		Banner: struct {
			Banners []*banner.Banner
			Version string
		}{Banners: materials.Banner, Version: materials.BannerVersion},
		Bangumi: struct {
			EP                map[int64]*bangumi.EpPlayer
			Season            map[int32]*episodegrpc.EpisodeCardsProto
			SeasonByAid       map[int32]*episodegrpc.EpisodeCardsProto
			InlinePGC         map[int32]*pgcinline.EpisodeCard
			Remind            *bangumi.Remind
			Update            *bangumi.Update
			PgcEpisodeByAids  map[int64]*pgccard.EpisodeCard
			PgcEpisodeByEpids map[int32]*pgccard.EpisodeCard
			PgcSeason         map[int32]*pgcAppGrpc.SeasonCardInfoProto
			EpMaterial        map[int64]*deliverygrpc.EpMaterial
		}{EP: nil, Season: materials.Season, SeasonByAid: materials.SeasonByAid, InlinePGC: materials.InlinePGC,
			Remind: materials.Remind, Update: materials.Update, PgcEpisodeByAids: materials.PgcEpisodeByAids,
			PgcEpisodeByEpids: materials.PgcEpisodeByEpids, PgcSeason: materials.PgcSeason, EpMaterial: materials.EpMaterial},
		Inline:     setInline(feed.Inline),
		ThumbUp:    struct{ HasLikeArchive map[int64]int8 }{HasLikeArchive: materials.HasLike},
		FollowMode: setFollowMode(feed.Index.FollowMode.Card),
		StoryIcon:  ConstructGotoIcon(feed.StoryIcon),
		Tunnel:     materials.Tunnel,
		Vip:        materials.Vip,
		Favourite:  materials.HasFavourite,
		HotAidSet:  materials.HotAidSet,
		Coin:       materials.HasCoin,
		LiveBadge: struct {
			LeftBottomBadgeStyle *operate.LiveBottomBadge
			LeftCoverBadgeStyle  []*operate.V9LiveLeftCoverBadge
		}{
			LeftBottomBadgeStyle: setLeftBottonBadgeStyle(materials),
			LeftCoverBadgeStyle:  materials.LiveLeftCoverBadgeStyle,
		},
		MultiMaterials:     materials.MultiMaterials,
		Specials:           materials.Specials,
		ThreePointMetaText: materials.ThreePointMetaText,
		Game:               materials.Game,
		Reservation:        materials.Reservation,
		SpecialCard:        materials.SpecialCard,
		OpenCourseMark:     materials.OpenCourseMark,
		LikeStatState:      materials.LikeStatState,
	}
	return out
}

func setLeftBottonBadgeStyle(materials *Materials) *operate.LiveBottomBadge {
	if val, ok := materials.LiveLeftBottomBadgeStyle[materials.LiveLeftBottomBadgeKey]; ok {
		return val
	}
	return nil
}

type Materials struct {
	Archive                  map[int64]*arcgrpc.ArcPlayer
	StoryArchive             map[int64]*arcgrpc.ArcPlayer
	Picture                  map[int64]*bplus.Picture
	Tag                      map[int64]*taggrpc.Tag
	AccountCard              map[int64]*accountgrpc.Card
	Channel                  map[int64]*channelgrpc.ChannelCard
	RelationStatMid          map[int64]*relationgrpc.StatReply
	IsAttention              map[int64]int8
	Article                  map[int64]*article.Meta
	Room                     map[int64]*live.Room
	InlineRoom               map[int64]*live.Room
	Season                   map[int32]*episodegrpc.EpisodeCardsProto
	SeasonByAid              map[int32]*episodegrpc.EpisodeCardsProto
	InlinePGC                map[int32]*pgcinline.EpisodeCard
	HasLike                  map[int64]int8
	Banner                   []*banner.Banner
	BannerVersion            string
	Remind                   *bangumi.Remind
	Update                   *bangumi.Update
	Tunnel                   map[int64]*tunnelgrpc.FeedCard
	Vip                      *vipgrpc.TipsRenewReply
	HasFavourite             map[int64]int8
	HotAidSet                sets.Int64
	HasCoin                  map[int64]int64
	PgcEpisodeByAids         map[int64]*pgccard.EpisodeCard
	PgcEpisodeByEpids        map[int32]*pgccard.EpisodeCard
	LiveLeftBottomBadgeKey   string
	LiveLeftBottomBadgeStyle map[string]*operate.LiveBottomBadge
	LiveLeftCoverBadgeStyle  []*operate.V9LiveLeftCoverBadge
	MultiMaterials           map[int64]*feedMgr.Material
	Specials                 map[int64]*operate.Card
	ThreePointMetaText       *threePointMeta.ThreePointMetaText
	Game                     map[int64]*game.Game
	Reservation              map[int64]*activitygrpc.UpActReserveRelationInfo
	IconList                 []*hmtchannelgrpc.Icon
	SpecialCard              map[int64]*resourceV2grpc.AppSpecialCard
	PgcSeason                map[int32]*pgcAppGrpc.SeasonCardInfoProto
	OpenCourseMark           map[int64]bool
	LikeStatState            map[int64]*thumbupgrpc.StatState
	EpMaterial               map[int64]*deliverygrpc.EpMaterial
}

func setInline(req *conf.Inline) *large_cover.Inline {
	if req == nil {
		return nil
	}
	return &large_cover.Inline{
		LikeButtonShowCount:      req.LikeButtonShowCount,
		LikeResource:             req.LikeResource,
		LikeResourceHash:         req.LikeResourceHash,
		DisLikeResource:          req.DisLikeResource,
		DisLikeResourceHash:      req.DisLikeResourceHash,
		LikeNightResource:        req.LikeNightResource,
		LikeNightResourceHash:    req.LikeNightResourceHash,
		DisLikeNightResource:     req.DisLikeNightResource,
		DisLikeNightResourceHash: req.DisLikeNightResourceHash,
		IconDrag:                 req.IconDrag,
		IconDragHash:             req.IconDragHash,
		IconStop:                 req.IconStop,
		IconStopHash:             req.IconStopHash,
		ThreePointPanelType:      req.ThreePointPanelType,
	}
}

func setFollowMode(req *feed.Card) *jsonselect.FollowMode {
	if req == nil {
		return nil
	}
	return &jsonselect.FollowMode{
		Title:   req.Title,
		Desc:    req.Desc,
		Buttons: req.Button,
	}
}

var NgMergeCardSet = sets.NewString("large_cover_single_v7:bangumi", "large_cover_single_v8:live",
	"large_cover_single_v7:inline_pgc", "large_cover_single_v8:inline_live", "large_cover_single_v9:inline_av",
	"large_cover_single_v7:pgc", "large_cover_v6:inline_av", "large_cover_v5:inline_av",
	"one_pic_v3:picture", "one_pic_v2:picture", "one_pic_v1:picture", "three_pic_v3:picture", "three_pic_v2:picture",
	"three_pic_v1:picture", "small_cover_v2:picture", "large_cover_v1:av", "large_cover_v1:bangumi", "large_cover_v1:live",
	"cm_v1:ad_av", "large_cover_single_v9:av", "cm_single_v1:ad_web_s", "ogv_small_cover:bangumi", "small_cover_v2:live",
	"small_cover_v9:live", "small_cover_v2:article_s", "three_item_h_v3:article_s", "large_cover_v9:inline_av_v2",
	"small_cover_v2:bangumi", "large_cover_v1:pgc", "small_cover_v2:pgc", "large_cover_v7:inline_pgc",
	"large_cover_v8:inline_live", "storys_v2:ai_story", "storys_v1:ai_story", "large_cover_single_v7:special_s",
	"large_cover_single_v8:special_s", "large_cover_single_v9:special_s", "notify_tunnel_large_single_v1:big_tunnel",
	"notify_tunnel_large_v1:big_tunnel", "banner_v8:banner", "banner_single_v8:banner", "banner_ipad_v8:banner",
	"large_cover_v1:article_s", "cm_double_v9:ad_inline_av", "cm_single_v9:ad_inline_av", "small_cover_v10:game",
	"small_cover_v2:av", "cm_v2:ad_web_gif_reservation", "cm_v2:ad_player_reservation", "cm_single_v1:ad_web_gif_reservation",
	"cm_single_v1:ad_player_reservation", "cm_v2:ad_av", "cm_v2:ad_web_s", "cm_v2:ad_web", "cm_v2:ad_player",
	"cm_v1:ad_av", "cm_v1:ad_web_s", "cm_v1:ad_web", "small_cover_v11:av", "cm_v2:ad_inline_3d", "cm_single_v1:ad_inline_3d",
	"cm_single_v1:ad_player", "small_cover_v2:special_s", "cm_v2:ad_ogv", "cm_double_v7:ad_inline_ogv",
	"cm_single_v7:ad_inline_ogv", "notify_tunnel_v1:new_tunnel", "notify_tunnel_single_v1:new_tunnel", "cm_v2:ad_inline_eggs",
	"cm_single_v1:ad_inline_eggs", "cm_v2:ad_inline_3d_v2", "cm_single_v1:ad_inline_3d_v2")
