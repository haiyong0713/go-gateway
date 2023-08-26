package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/story"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedcommon "go-gateway/app/app-svr/app-feed/interface/common"
	appResourcegrpc "go-gateway/app/app-svr/app-resource/interface/api/v1"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/story/internal/dao"
	"go-gateway/app/app-svr/story/internal/model"
	gateecode "go-gateway/ecode"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	liverankgrpc "git.bilibili.co/bapis/bapis-go/live/rankdb/v1"
	livegrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	materialgrpc "git.bilibili.co/bapis/bapis-go/material/interface"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcstory "git.bilibili.co/bapis/bapis-go/pgc/service/card/story"
	pgcFollowClient "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicgrpc "git.bilibili.co/bapis/bapis-go/topic/service"
	uparcgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"
	vogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"github.com/pkg/errors"
)

var (
	ErrStoryLiveReserveKeyExist = errors.New("story-reservation: live reserve key already exists")
	ErrValidStoryReserve        = errors.New("story-reservation: invalid reservation data")

	_thumbUpAnimationSlice = []string{"story_like_combo_22", "story_like_combo_33", "story_like_combo_tv"}
)

const (
	_max         = 20
	_pgcNum      = 5
	_degradeCode = -1200
)

// nolint:gomnd
func storyExpGroup(mid int64, buvid string) int64 {
	prefix := buvid
	if mid > 0 {
		prefix = strconv.FormatInt(mid, 10)
	}
	digest := md5.Sum([]byte(fmt.Sprintf("%s%s", prefix, "583eec250c9a3ec7")))
	subStr := []byte(hex.EncodeToString(digest[:]))[18:]
	target, _ := strconv.ParseInt(string(subStr), 16, 64)
	return target % 25
}

// nolint: unparam
func (s *StoryService) storyRcmd(c context.Context, plat int8, build, pull int, buvid string, mid, aid int64, displayID int,
	storyParam, adExtra string, adResource int64, location *locgrpc.InfoReply, mobiApp, network string, feedStatus,
	fromAvID int64, fromTrackId string, disableRcmd, requestFrom int) (*ai.StoryView, int, error) {
	if s.cfg.CustomConfig.DegradeSwitch {
		dg := sets.NewInt64(s.cfg.CustomConfig.DegradeGroup...)
		if dg.Has(storyExpGroup(mid, buvid)) {
			s.infoProm.Incr("story_backup_with_switch")
			return s.storyBackupRcmd(), _degradeCode, nil
		}
	}
	data, respCode, err := s.dao.StoryRcmd(c, plat, build, pull, buvid, mid, aid, adResource, displayID, storyParam,
		adExtra, location, mobiApp, network, feedStatus, fromAvID, fromTrackId, disableRcmd, requestFrom)
	if err != nil || respCode == 500 {
		log.Warn("Failed to request story ai rcmd failback to backup data: %+v: %d: mid: %d: buvid: %q", err, respCode, mid, buvid)
		s.infoProm.Incr("story_backup_with_code")
		return s.storyBackupRcmd(), respCode, nil
	}
	return data, respCode, nil
}

func (s *StoryService) storyBackupRcmd() *ai.StoryView {
	backupItems := s.storyRcmdCacheByRandom(4)
	return &ai.StoryView{Data: backupItems}
}

func (s *StoryService) storyRcmdCacheByRandom(count int) []*ai.SubItems {
	cache := s.storyRcmdCache
	index := len(cache)
	if count > 0 && count < index {
		index = count
	}
	out := make([]*ai.SubItems, 0, index)
	for _, idx := range rand.Perm(len(cache))[:index] {
		out = append(out, cache[idx])
	}
	return out
}

// request_from: 0:天马story入口  1:首页左上角独立入口  2:播放页入口
var NeedAddEntranceRequestFrom = sets.NewInt64(0, 2)

// nolint: gocognit
func (s *StoryService) Story(c context.Context, plat int8, buvid string, mid int64, param *model.StoryParam, now time.Time) (res []*story.Item, data *ai.StoryView, config *story.StoryConfig, respCode int) {
	var (
		aids, argueAids                         []int64
		storyUpIDs, avUpIDs, liveUpIDs, epUpIDs []int64
		rids                                    []int64
		epids, iconEpids, seasonids             []int32
		iconBcutReq                             []*materialgrpc.StoryReq
		heInlineReq                             []*pgcinline.HeInlineReq
		err                                     error
		userInfo                                *accountgrpc.Card
		haslike, isFav, isFavEp                 map[int64]int8
		likeMap, coinsMap                       map[int64]int64
		likeAnimationIconMap                    map[int64]*thumbupgrpc.LikeAnimation
		followMap                               map[int32]*pgcFollowClient.FollowStatusProto
		amplayer                                map[int64]*arcgrpc.ArcPlayer
		cardm                                   map[int64]*accountgrpc.Card
		statm                                   map[int64]*relationgrpc.StatReply
		authorRelations                         map[int64]*relationgrpc.InterrelationReply
		liveRoomInfos                           map[int64]*livegrpc.EntryRoomInfoResp_EntryList
		upReservationMap                        map[int64]*story.ReservationInfo
		storyTags                               map[int64][]*channelgrpc.Channel
		arguementMap                            map[int64]*vogrpc.Argument
		liveCardInfos                           map[int64]*livegrpc.EntryRoomInfoResp_EntryList
		eppm                                    map[int32]*pgcinline.EpisodeCard
		bcutStoryCart                           map[string]*materialgrpc.StoryRes
		liveRankInfos                           map[int64]*liverankgrpc.IsInHotRankResp_HotRankData
		contractShowConfig                      map[int64]*story.ContractResource
	)
	config = s.constructStoryConfig(mid, buvid, param)
	ip := metadata.String(c, metadata.RemoteIP)
	zone, infoErr := s.dao.InfoGRPC(c, ip)
	if infoErr != nil {
		log.Warn("Failed to get location info: %v, %+v", ip, infoErr)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		data, respCode, err = s.storyRcmd(ctx, plat, param.Build, param.Pull, buvid, mid, param.AID, param.DisplayID, param.StoryParam,
			param.AdExtra, constructStoryAdResource(plat), zone, param.MobiApp, param.Network, param.FeedStatus, param.AID,
			param.TrackID, param.DisableRcmd, param.RequestFrom)
		return nil
	})
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if userInfo, err = s.dao.Card3(ctx, mid); err != nil {
				log.Error("Failed to request card3: %+v", err)
				err = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("Failed to eg.Wait: %+v", err)
	}
	// 确保第一刷的时候，把天马点击进来的story卡片插入到详情页的story页里面的第一个
	if (param.DisplayID == 1 && NeedAddEntranceRequestFrom.Has(int64(param.RequestFrom))) ||
		(param.DisplayID == 1 && param.RequestFrom == 1 && param.AID > 0) { // 特殊处理左上角进小窗返回story定位问题
		if data == nil {
			data = &ai.StoryView{}
		}
		if len(data.Data) > 0 {
			// 第一刷第一张上报天马track_id，后面的全部上报story的track_id
			data.Data = append(data.Data[:0], append([]*ai.SubItems{{Goto: model.GotoVerticalAv, ID: param.AID,
				TrackID: param.TrackID}}, data.Data[0:]...)...)
		} else {
			data.Data = append(data.Data, &ai.SubItems{Goto: model.GotoVerticalAv, ID: param.AID})
		}
	}
	if err != nil {
		err = nil
	}
	if data == nil || len(data.Data) == 0 {
		res = []*story.Item{}
		return
	}
	epidSet := sets.NewInt64()
	for _, v := range data.Data {
		switch v.Goto {
		case model.GotoVerticalAv:
			aids = append(aids, v.ID)
		case model.GotoVerticalAdAv:
			aids = append(aids, v.ID)
			if v.AdvertiseType == model.StoryHasMidAd && v.StoryUpMid > 0 {
				storyUpIDs = append(storyUpIDs, v.StoryUpMid)
			}
		case model.GotoVerticalLive, model.GotoVerticalAdLive:
			if v.ID > 0 {
				rids = append(rids, v.ID)
			}
			v.SetLiveAttentionExp(s.liveAttentionExp(mid))
		case model.GotoVerticalPgc:
			aids = append(aids, v.ID)
			epidSet.Insert(int64(v.EpID))
			heInlineReq = append(heInlineReq, &pgcinline.HeInlineReq{
				EpId:             v.EpID,
				HeBeginPlayPoint: v.HighlightStart,
			})
		default:
			log.Error("Unsupported story goto: %s", v.Goto)
		}
		count := len(aids) + len(rids) + epidSet.Len()
		if count == _max {
			log.Warn("Ai story card is max")
			break
		}
	}
	iconBcutReq, iconEpids = filterIconIds(data.Data, buvid)
	epids, heInlineReq = mergeEpids(epidSet, iconEpids, heInlineReq)
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		if amplayer, err = s.dao.ArcsPlayer(ctx, aids, "story", need1080plus(userInfo)); err != nil {
			log.Error("Failed to request ArcsPlayer: %+v", err)
			return
		}
		for _, a := range amplayer {
			if a.Arc.AttrVal(arcgrpc.AttrBitHasArgument) == arcgrpc.AttrYes {
				argueAids = append(argueAids, a.Arc.Aid)
			}
			if a.Arc.Author.Mid <= 0 {
				log.Warn("Unexpected mid with ugc: %d", a.Arc.Aid)
				continue
			}
			avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if storyTags, err = s.dao.ResourceChannels(ctx, aids, mid); err != nil {
			log.Error("Failed to request ResourceChannels: %+v", err)
			err = nil
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if haslike, err = s.dao.HasLike(ctx, buvid, mid, aids); err != nil {
			log.Error("Failed to request HasLike: %+v", err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		if likeAnimationIconMap, err = s.dao.MultiLikeAnimation(ctx, aids); err != nil {
			log.Error("Failed to request MultiLikeAnimation: %+v", err)
			err = nil
		}
		return
	})
	if len(iconBcutReq) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if bcutStoryCart, err = s.dao.StoryTagList(ctx, iconBcutReq); err != nil {
				log.Error("Failed to request StoryTagList: %+v", err)
			}
			return nil
		})
	}
	if mid > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if isFav, err = s.dao.IsFavVideos(ctx, mid, aids); err != nil {
				log.Error("Failed to request IsFavVideos: %+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if coinsMap, err = s.dao.ArchiveUserCoins(ctx, aids, mid); err != nil {
				log.Error("Failed to request ArchiveUserCoins: %+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) error {
			var err error
			if contractShowConfig, err = s.dao.ContractShowConfig(ctx, aids, mid); err != nil {
				log.Error("s.dao.ContractShowConfig aids:%v, mid:%d, err:%v", aids, mid, err)
				return nil
			}
			return nil
		})
	}
	if len(rids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if liveCardInfos, err = s.liveRoomInfos(ctx, []string{story.EntryFromStoryLive,
				story.EntryFromStoryLiveCloseEntryButton, story.EntryFromStoryAdLive,
				story.EntryFromStoryAdLiveCloseEntryButton}, rids, []int64{}, mid, param); err != nil {
				return nil
			}
			for _, lv := range liveCardInfos {
				if lv.Uid <= 0 {
					log.Warn("Unexpected mid with live: %d", lv.LiveId)
					continue
				}
				liveUpIDs = append(liveUpIDs, lv.Uid)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if liveRankInfos, err = s.dao.LiveHotRank(ctx, rids); err != nil {
				log.Error("Failed to request LiveHotRank: %v, %+v", rids, err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if eppm, err = s.dao.InlineCards(ctx, epids, param.MobiApp, param.Platform, param.Device, param.Build, mid,
				true, buvid, heInlineReq); err != nil {
				log.Error("Failed to request InlineCards: %+v", err)
				return nil
			}
			for _, ep := range eppm {
				seasonids = append(seasonids, ep.GetSeason().GetSeasonId())
				if ep.GetContributeUpInfo().GetMid() <= 0 {
					log.Warn("Unexpected mid with pgc: %d", ep.GetEpisodeId())
					continue
				}
				epUpIDs = append(epUpIDs, ep.GetContributeUpInfo().GetMid())
			}
			return nil
		})
		if mid > 0 {
			g.Go(func(ctx context.Context) (err error) {
				if isFavEp, err = s.dao.IsFavEp(ctx, mid, feedcommon.Int32SliceToInt64Slice(epids)); err != nil {
					log.Error("Failed to request IsFavEp: %+v", err)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("eg.Wait: %+v", err)
		return
	}
	storyUpIDs = append(storyUpIDs, avUpIDs...)
	storyUpIDs = append(storyUpIDs, liveUpIDs...)
	storyUpIDs = append(storyUpIDs, epUpIDs...)
	g = errgroup.WithContext(c)
	if len(storyUpIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if cardm, err = s.dao.Cards3GRPC(ctx, storyUpIDs); err != nil {
				log.Error("Failed to request Cards3GRPC: %+v", err)
				err = nil
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if statm, err = s.dao.StatsGRPC(ctx, storyUpIDs); err != nil {
				log.Error("Failed to request StatsGRPC: %+v", err)
				err = nil
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if authorRelations, err = s.dao.RelationsInterrelations(ctx, mid, storyUpIDs); err != nil {
				log.Error("Failed to request RelationsInterrelations: %+v", err)
				err = nil
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if likeMap, err = s.dao.UserLikedCounts(ctx, storyUpIDs); err != nil {
				log.Error("Failed to request UserLikedCounts: %+v", err)
				err = nil
			}
			return
		})
		if len(seasonids) > 0 && mid > 0 {
			g.Go(func(ctx context.Context) (err error) {
				if followMap, err = s.dao.StatusByMid(ctx, mid, seasonids); err != nil {
					log.Error("Failed to StatusByMid: %d, %+v, %+v", mid, seasonids, err)
					err = nil
				}
				return
			})
		}
		g.Go(func(ctx context.Context) (err error) {
			if liveRoomInfos, err = s.liveRoomInfos(ctx,
				[]string{story.EntryFromStoryFeedUpIcon, story.EntryFromStoryFeedUpPanel},
				[]int64{}, storyUpIDs, mid, param); err != nil {
				err = nil
			}
			return
		})
		if s.matchLiveReservationGroup() {
			g.Go(func(ctx context.Context) (err error) {
				upReservationMap = s.doReservationInfo(ctx, storyUpIDs, mid)
				return
			})
		}
	}
	if len(argueAids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if arguementMap, err = s.dao.Arguments(ctx, argueAids); err != nil {
				log.Error("Failed to request Arguments, param: %+v, %+v", argueAids, err)
				err = nil
			}
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	dev, _ := device.FromContext(c)
	iconm := &story.IconMaterial{
		BcutStoryCart: bcutStoryCart,
		Eppm:          eppm,
	}
	buttonFilter := s.buttonFilter(param.MobiApp, param.Build)
	enableScreencast := s.enableScreencast(buttonFilter)
	needCoin := needCoinDev(param.MobiApp, param.Build)
	jumpToSeason := s.jumpToSeason(mid, buvid)
	for _, v := range data.Data {
		v.SetDisableRcmd(param.DisableRcmd)
		v.SetVideoMode(int(param.VideoMode))
		v.SetMobiApp(param.MobiApp)
		v.SetBuild(param.Build)
		switch v.Goto {
		case model.GotoVerticalAv:
			a, ok := amplayer[v.ID]
			if !ok || !cdm.AvIsNormalGRPC(a) {
				continue
			}
			if cdm.AvIsCharging(a) {
				log.Warn("Filtered by charging aid: %d", a.Arc.Aid)
				continue
			}
			if card.CheckMidMaxInt32(a.Arc.Author.Mid) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", a.Arc.Author.Mid)
				continue
			}
			i := &story.Item{}
			fns := []story.StoryFn{
				story.OptCreativeEntrance(v),
				story.OptAiStoryCommonJumpIcon(iconm, v, jumpToSeason),
				story.OptPosRecTitle(v),
				story.OptThreePointButton(story.NeedDislike(true),
					story.NeedReport(true),
					story.NeedCoin(needCoin),
					story.NeedScreencast(enableScreencast),
					story.NoPlayBackground(a.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
				),
				story.OptShareBottomButton(story.NeedDislike(true),
					story.NeedReport(true),
					story.NeedCoin(needCoin),
					story.NeedScreencast(enableScreencast),
					story.CoinNum(cdm.StatString(a.Arc.Stat.Coin, "")),
					story.NoPlayBackground(a.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
				),
			}
			i.StoryFrom(v, a, cardm, statm, authorRelations, likeMap, coinsMap, haslike, isFav, storyTags[v.ID], s.hotAids,
				getLiveRoomBuilder(liveRoomInfos[a.Arc.Author.Mid], story.EntryFromStoryFeedUpIcon, story.EntryFromStoryFeedUpPanel),
				plat, param.Build, param.MobiApp, getRandomThumbUpAnimation(), upReservationMap[a.Arc.Author.Mid],
				arguementMap, cdm.FfCoverFromStory, mid, contractShowConfig, likeAnimationIconMap, fns...)
			res = append(res, i)
		case model.GotoVerticalAdAv:
			a, ok := amplayer[v.ID]
			if !ok || !cdm.AvIsNormalGRPC(a) {
				log.Error("Failed to get archive: %d, %+v, trackID: %s", v.ID, a, v.TrackID)
				continue
			}
			if cdm.AvIsCharging(a) {
				log.Warn("Filtered by charging aid: %d", a.Arc.Aid)
				continue
			}
			if card.CheckMidMaxInt32(a.Arc.Author.Mid) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", a.Arc.Author.Mid)
				continue
			}
			if !hasStoryAdResource(data.StoryBiz) {
				log.Error("Failed to get story ad resource: %d, %+v, trackID: %s", v.ID, data.StoryBiz, v.TrackID)
				continue
			}
			if data.StoryBiz.Data.StoryVideoID != a.Arc.Aid {
				log.Error("Failed to match AI id and story video id: %d, %d, trackId: %s, requestId: %s", a.Arc.Aid, data.StoryBiz.Data.StoryVideoID, v.TrackID, data.StoryBiz.Data.RequestID)
			}
			if data.StoryBiz.Data.AdvertiseType != v.AdvertiseType {
				log.Error("Failed to match AI advertiseType and ad advertiseType: %d, %d, trackId: %s, requestId: %s", v.AdvertiseType, data.StoryBiz.Data.AdvertiseType, v.TrackID, data.StoryBiz.Data.RequestID)
			}
			i := &story.Item{}
			fns := []story.StoryFn{
				story.OptThreePointButton(story.NeedDislike(true),
					story.NeedReport(true),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.NoPlayBackground(a.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
					story.SmallWindowFilter(true),
				),
				story.OptShareBottomButton(story.NeedDislike(true),
					story.NeedReport(true),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.CoinNum(cdm.StatString(a.Arc.Stat.Coin, "")),
					story.NoPlayBackground(a.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
					story.SmallWindowFilter(true),
				),
			}
			i.StoryFrom(v, a, cardm, statm, authorRelations, likeMap, coinsMap, haslike, isFav, storyTags[v.ID], s.hotAids,
				getLiveRoomBuilder(liveRoomInfos[a.Arc.Author.Mid], story.EntryFromStoryFeedUpIcon, story.EntryFromStoryFeedUpPanel),
				plat, param.Build, param.MobiApp, getRandomThumbUpAnimation(), upReservationMap[a.Arc.Author.Mid],
				arguementMap, cdm.FfCoverFromStory, mid, contractShowConfig, likeAnimationIconMap, fns...)
			i.Goto = cdm.Gt(v.Goto)
			i.AdType = v.AdvertiseType
			switch v.AdvertiseType {
			case model.StoryHasMidAd:
				if card.CheckMidMaxInt32(v.StoryUpMid) && card.CheckMidMaxInt32Version(dev) {
					log.Warn("Filtered by mid int64: %d", v.StoryUpMid)
					continue
				}
				fns := []story.StoryFn{
					story.OptThreePointButton(story.NeedDislike(true),
						story.NeedReport(true),
						story.NeedCoin(needCoin),
						story.NeedScreencast(false),
						story.NoPlayBackground(a.Arc.Rights.NoBackground),
						story.ButtonFilter(buttonFilter),
						story.SmallWindowFilter(true),
					),
					story.OptShareBottomButton(story.NeedDislike(true),
						story.NeedReport(true),
						story.NeedCoin(needCoin),
						story.NeedScreencast(false),
						story.CoinNum(cdm.StatString(a.Arc.Stat.Coin, "")),
						story.NoPlayBackground(a.Arc.Rights.NoBackground),
						story.ButtonFilter(buttonFilter),
						story.SmallWindowFilter(true),
					),
				}
				i.AdStoryWithMid(v, cardm, statm, authorRelations, likeMap, upReservationMap[v.StoryUpMid],
					getLiveRoomBuilder(liveRoomInfos[v.StoryUpMid], story.EntryFromStoryFeedUpIcon, story.EntryFromStoryFeedUpPanel), fns...)
			case model.StoryNoMidAd:
				fns := []story.StoryFn{
					story.OptThreePointButton(story.NeedDislike(true),
						story.NeedReport(true),
						story.NeedCoin(false),
						story.NeedScreencast(false),
						story.ButtonFilter(buttonFilter),
						story.NoPlayBackground(a.Arc.Rights.NoBackground),
						story.SmallWindowFilter(true),
					),
					story.OptShareBottomButton(story.NeedDislike(true),
						story.NeedReport(true),
						story.NeedCoin(false),
						story.NeedScreencast(false),
						story.CoinNum(cdm.StatString(a.Arc.Stat.Coin, "")),
						story.ButtonFilter(buttonFilter),
						story.NoPlayBackground(a.Arc.Rights.NoBackground),
						story.SmallWindowFilter(true),
					),
				}
				i.AdStoryWithoutMid(fns...)
			default:
			}
			hasAdContent := false
			i.AdInfo, hasAdContent = cm.AsStoryAdInfo(data.StoryBiz.Data.StoryAdResource)
			if !hasAdContent {
				log.Error("Failed to get story adContent: %+v, %s", data.StoryBiz.Data.StoryAdResource, v.TrackID)
				continue
			}
			i.FillStoryCartIcon(data.StoryBiz.Data.StoryAdResource)
			res = append(res, i)
		case model.GotoVerticalLive:
			room, ok := liveCardInfos[v.ID]
			if !ok || room.LiveStatus != 1 {
				log.Error("Failed to get live: %d, %+v, trackID: %s", v.ID, room, v.TrackID)
				continue
			}
			if card.CheckMidMaxInt32(room.Uid) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", room.Uid)
				continue
			}
			i := &story.Item{}
			fns := []story.StoryFn{
				story.OptPosRecTitle(v),
			}
			i.StoryFromLiveRoom(v, room, cardm, statm, authorRelations, story.EntryFromStoryLive,
				buildLiveCardLiveRoom(room, story.EntryFromStoryLiveCloseEntryButton), liveRankInfos, fns...)
			res = append(res, i)
		case model.GotoVerticalAdLive:
			room, ok := liveCardInfos[v.ID]
			if !ok || room.LiveStatus != 1 {
				log.Error("Failed to get live: %d, %+v, trackID: %s", v.ID, room, v.TrackID)
				continue
			}
			if card.CheckMidMaxInt32(room.Uid) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", room.Uid)
				continue
			}
			if !hasStoryAdResource(data.StoryBiz) {
				log.Error("Failed to get story ad resource: %d, %+v, trackID: %s", v.ID, data.StoryBiz, v.TrackID)
				continue
			}
			if data.StoryBiz.Data.StoryLiveRoomId != room.RoomId {
				log.Error("Failed to match AI id and story live id: %d, %d, trackId: %s, requestId: %s", room.RoomId, data.StoryBiz.Data.StoryLiveRoomId, v.TrackID, data.StoryBiz.Data.RequestID)
			}
			i := &story.Item{}
			i.StoryFromLiveRoom(v, room, cardm, statm, authorRelations, story.EntryFromStoryAdLive,
				buildLiveCardLiveRoom(room, story.EntryFromStoryAdLiveCloseEntryButton), liveRankInfos)
			hasAdContent := false
			i.AdInfo, hasAdContent = cm.AsStoryAdInfo(data.StoryBiz.Data.StoryAdResource)
			if !hasAdContent {
				log.Error("Failed to get story adContent: %+v, %s", data.StoryBiz.Data.StoryAdResource, v.TrackID)
				continue
			}
			res = append(res, i)
		case model.GotoVerticalPgc:
			ep, ok := eppm[v.EpID]
			if !ok {
				log.Warn("Invalid ep: %d", v.EpID)
				continue
			}
			if card.CheckMidMaxInt32(ep.Season.GetUpInfo().GetMid()) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", ep.Season.GetUpInfo().GetMid())
				continue
			}
			if forbiddenPgcType(ep) {
				log.Warn("Filter by pgc type: %+v, scene: %+v, epid: %d", ep.GetInlineType(), ep.GetInlineScene(), ep.GetEpisodeId())
				continue
			}
			i := &story.Item{}
			fns := []story.StoryFn{
				story.OptAiStoryCommonJumpIcon(iconm, v, jumpToSeason),
				story.OptPosRecTitle(v),
				story.OptThreePointButton(story.NeedDislike(true),
					story.NeedReport(false),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.NoPlayBackground(0),
					story.ButtonFilter(buttonFilter),
					story.NoWatchLater(true),
					story.SmallWindowFilter(true),
				),
				story.OptShareBottomButton(story.NeedDislike(true),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.CoinNum(cdm.Stat64String(ep.Stat.Coin, "")),
					story.NoPlayBackground(0),
					story.ButtonFilter(buttonFilter),
					story.NoWatchLater(true),
					story.SmallWindowFilter(true),
				),
				story.OptPGCStyle(v, ep),
			}
			i.StoryFromPGC(v, ep, cardm, statm, authorRelations, likeMap, coinsMap, haslike, isFavEp, getRandomThumbUpAnimation(), mid, followMap, fns...)
			res = append(res, i)
		default:
			log.Warn("Unsupported goto: %s, track_id: %s", v.Goto, v.TrackID)
		}
	}
	if hasStoryAdResource(data.StoryBiz) {
		markAsAdStock(res, data.StoryBiz, param)
	}
	return
}

func (s *StoryService) appendToViewCfg(cfg *story.StoryConfig, mobiApp string, build int) {
	if jumpToViewFilter(mobiApp, build) {
		return
	}
	cfg.EnableJumpToView, cfg.JumpToViewIcon = true, s.cfg.CustomConfig.JumpToViewIcon
}

func jumpToViewFilter(mobiApp string, build int) bool {
	return mobiApp == "iphone" && build < 68000000
}

func forbiddenPgcType(ep *pgcinline.EpisodeCard) bool {
	return !(ep.GetInlineType() == pgcinline.InlineType_TYPE_WHOLE &&
		(ep.GetInlineScene() == pgcinline.InlineScene_SCENE_HE ||
			ep.GetInlineScene() == pgcinline.InlineScene_SCENE_SKIP ||
			ep.GetInlineScene() == pgcinline.InlineScene_SCENE_RE_START))
}

func (s *StoryService) liveRoomInfos(ctx context.Context, entryFrom []string, rids []int64, uids []int64, mid int64, param *model.StoryParam) (map[int64]*livegrpc.EntryRoomInfoResp_EntryList, error) {
	result, err := s.dao.LiveRoomInfos(ctx, &livegrpc.EntryRoomInfoReq{
		EntryFrom:     entryFrom,
		RoomIds:       rids,
		Uids:          uids,
		Uid:           mid,
		Uipstr:        metadata.String(ctx, metadata.RemoteIP),
		Platform:      param.Platform,
		Build:         int64(param.Build),
		DeviceName:    param.DeviceName,
		Network:       param.Network,
		FilterOffline: 1,
		ReqBiz:        "/x/v2/feed/index/story",
		MobiApp:       param.MobiApp,
	})
	if err != nil {
		log.Error("Failed to request LiveRoomInfos: %+v", err)
		return nil, err
	}
	return result, nil
}

func buildLiveCardLiveRoom(room *livegrpc.EntryRoomInfoResp_EntryList, entryFrom string) *story.LiveRoom {
	return &story.LiveRoom{
		LiveStatus:     room.LiveStatus,
		CloseButtonURI: room.JumpUrl[entryFrom],
		AreaID:         room.AreaId,
		ParentAreaID:   room.ParentAreaId,
		LiveType:       parseLiveType(room),
	}
}

// nolint:gomnd
func parseLiveType(room *livegrpc.EntryRoomInfoResp_EntryList) string {
	if room.LiveInfo == nil {
		return ""
	}
	switch room.LiveInfo.LiveModel {
	case 1:
		return "video"
	case 2:
		return "screen_record"
	case 3:
		return "voice"
	default:
		return ""
	}
}

func mergeEpids(epids sets.Int64, iconEpids []int32, heInlineReq []*pgcinline.HeInlineReq) ([]int32, []*pgcinline.HeInlineReq) {
	for _, id := range iconEpids {
		if epids.Has(int64(id)) {
			continue
		}
		epids.Insert(int64(id))
		heInlineReq = append(heInlineReq, &pgcinline.HeInlineReq{
			EpId: id,
		})
	}
	out := make([]int32, 0, epids.Len())
	for _, v := range epids.List() {
		out = append(out, int32(v))
	}
	return out, heInlineReq
}

func filterIconIds(data []*ai.SubItems, buvid string) (bcupIds []*materialgrpc.StoryReq, epIds []int32) {
	for _, item := range data {
		if item.HasIcon == 0 {
			continue
		}
		switch item.IconType {
		case "ogv":
			epIds = append(epIds, int32(item.IconID))
		case "bcut":
			// 应ai要求更换使用item.ID
			bcupIds = append(bcupIds, &materialgrpc.StoryReq{
				Avid:  item.ID,
				Type:  int32(item.IconID),
				Buvid: buvid,
			})
		case "cart":
		default:
			log.Warn("Failed to match iconType: %s", item.IconType)
		}
	}
	return bcupIds, epIds
}

func (s *StoryService) doReservationInfo(c context.Context, storyUpIDs []int64, mid int64) map[int64]*story.ReservationInfo {
	lock := sync.Mutex{}
	eg := errgroup.WithContext(c)
	upReservationMap := make(map[int64]*story.ReservationInfo, len(storyUpIDs))
	for _, upId := range storyUpIDs {
		upId := upId
		if mid == upId {
			continue
		}
		eg.Go(func(ctx context.Context) error {
			reserve, err := s.reserveRelationInfoFrom(ctx, mid, upId)
			if err != nil {
				if errors.Cause(err) == dao.ErrNoLiveReservation || errors.Cause(err) == ErrStoryLiveReserveKeyExist || errors.Cause(err) == ErrValidStoryReserve {
					return nil
				}
				log.Error("s.reserveRelationInfoFrom err(%+v)", err)
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			upReservationMap[upId] = reserve
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("error group err(%+v) while constructing reservation info", err)
	}
	return upReservationMap
}

func (s *StoryService) reserveRelationInfoFrom(c context.Context, mid, upmid int64) (*story.ReservationInfo, error) {
	resReq := &appResourcegrpc.CheckEntranceInfocRequest{
		Mid:      mid,
		UpMid:    upmid,
		Business: "live_reserve",
	}
	isExist, err := s.dao.StoryLiveReserveKeyExists(c, resReq)
	if err != nil {
		return nil, errors.Wrapf(err, "s.actDao.StoryLiveReserveKeyExists resReq(%+v)", resReq)
	}
	if isExist {
		return nil, ErrStoryLiveReserveKeyExist
	}
	actReq := &activitygrpc.UpActReserveRelationInfo4LiveReq{
		Mid:   mid,
		Upmid: upmid,
		From:  activitygrpc.UpCreateActReserveFrom_FromStory,
	}
	data, err := s.dao.StoryLiveReserveCard(c, actReq)
	if err != nil {
		return nil, errors.Wrapf(err, "s.actDao.StoryLiveReserveCard actReq(%+v)", resReq)
	}
	if !isValidReservation(data) {
		return nil, ErrValidStoryReserve
	}
	info := &story.ReservationInfo{}
	info.Sid = data.Sid
	info.Name = "开始直播"
	info.IsFollow = data.IsFollow
	info.LivePlanStartTime = data.LivePlanStartTime
	return info, nil
}

func isValidReservation(data *activitygrpc.UpActReserveRelationInfo) bool {
	// story预约仅客态可见，在未预约且非审核态展示
	return data.IsFollow == 0 && data.UpActVisible != activitygrpc.UpActVisible_OnlyUpVisible
}

func (s *StoryService) matchLiveReservationGroup() bool {
	return !(s.cfg.CustomConfig.DisableStoryLiveReserveMid)
}

func (s *StoryService) constructStoryConfig(mid int64, buvid string, param *model.StoryParam) *story.StoryConfig {
	config := &story.StoryConfig{
		ProgressBar:     s.setupProgressBar(),
		EnableRcmdGuide: true,
		SlideGuidanceAb: 1,
		ShowButton:      _storyButton,
		ReplyZoomExp:    s.replyZoomExp(mid, buvid, param.MobiApp, param.Build),
		ReplyNoDanmu:    s.replyNoDanmu(mid, buvid),
		ReplyHighRaised: s.replyHighRaised(mid, buvid),
		SpeedPlayExp:    s.speedPlayExp(mid, buvid),
	}
	s.appendToViewCfg(config, param.MobiApp, param.Build)
	return config
}

var (
	_storyButton = []string{"like", "reply", "coin", "fav", "share"}
)

func (s *StoryService) replyZoomExp(mid int64, buvid, mobiApp string, build int) int8 {
	const (
		_reply_horzion = 1
		_reply_all     = 2
	)
	if (mobiApp == "iphone" && build < 68300000) || (mobiApp == "android" && build < 6830000) {
		return _reply_horzion
	}
	if s.matchFeatureControl(mid, buvid, "reply_all") {
		return _reply_all
	}
	bucket := int(crc32.ChecksumIEEE([]byte(buvid+"_reply_all_")) % 20)
	verticalSet := sets.NewInt(s.cfg.CustomConfig.ReplyVerticalGroup...)
	if verticalSet.Has(bucket) {
		return _reply_all
	}
	return _reply_horzion
}

func (s *StoryService) replyNoDanmu(mid int64, buvid string) bool {
	if s.matchFeatureControl(mid, buvid, "reply_no_danmu") {
		return true
	}
	bucket := int(crc32.ChecksumIEEE([]byte(buvid+"_reply_all_")) % 20)
	set := sets.NewInt(s.cfg.CustomConfig.ReplyNoDanmuGroup...)
	return set.Has(bucket)
}

func (s *StoryService) replyHighRaised(mid int64, buvid string) bool {
	if s.matchFeatureControl(mid, buvid, "reply_high_raised") {
		return true
	}
	bucket := int(crc32.ChecksumIEEE([]byte(buvid+"_reply_all_")) % 20)
	set := sets.NewInt(s.cfg.CustomConfig.ReplyHighRaisedGroup...)
	return set.Has(bucket)
}

func (s *StoryService) speedPlayExp(mid int64, buvid string) bool {
	if s.matchFeatureControl(mid, buvid, "speed_play") {
		return true
	}
	bucket := int(crc32.ChecksumIEEE([]byte(buvid+"_speed_play")) % 20)
	return bucket < s.cfg.CustomConfig.SpeedPlay
}

func (s *StoryService) setupProgressBar() *story.ProgressBar {
	if s.cfg.CustomConfig == nil {
		return nil
	}
	return &story.ProgressBar{
		IconDrag:     s.cfg.CustomConfig.IconDrag,
		IconDragHash: s.cfg.CustomConfig.IconDragHash,
		IconStop:     s.cfg.CustomConfig.IconStop,
		IconStopHash: s.cfg.CustomConfig.IconStopHash,
		IconZoom:     s.cfg.CustomConfig.IconZoom,
		IconZoomHash: s.cfg.CustomConfig.IconZoomHash,
	}
}

func hasStoryAdResource(biz *ai.StoryBiz) bool {
	if biz == nil || biz.Code != 0 || biz.Data == nil {
		return false
	}
	return true
}

func constructStoryAdResource(plat int8) int64 {
	switch plat {
	case model.PlatAndroid, model.PlatAndroidB:
		//nolint:gomnd
		return 4355
	case model.PlatIPhone, model.PlatIPhoneB:
		//nolint:gomnd
		return 4352
	}
	return 0
}

func markAsAdStock(items []*story.Item, biz *ai.StoryBiz, param *model.StoryParam) {
	index := int(biz.Data.CardIndex)
	if index <= 0 {
		log.Error("Ad card_index is illegal: %d", index)
		return
	}
	if param.DisplayID == 1 {
		index++
	}
	if index > len(items) { // 若库存位置大于当前卡片长度，则将库存挂在最后一张卡片上
		index = len(items)
	}
	if index <= 0 {
		return
	}
	if items[index-1].Goto == model.GotoVerticalAdAv {
		return
	}
	adInfo, hasAdContent := cm.AsStoryAdInfo(biz.Data.StoryAdResource)
	if hasAdContent {
		log.Error("Failed to mark as ad stock, story has adContent: %+v, trackID: %s", biz.Data.StoryAdResource,
			items[index-1].Rcmd.TrackID)
		return
	}
	items[index-1].AdInfo = adInfo
}

// SpaceStory is
// nolint: gocognit
func (s *StoryService) SpaceStory(ctx context.Context, param *model.SpaceStoryParam) (*model.SpaceStoryReply, error) {
	if param.PN <= 0 {
		param.PN = 1
	}
	if param.PS > 50 || param.PS <= 0 {
		param.PS = 20
	}
	attrNot := uint64(
		(1 << arcgrpc.AttrBitIsPGC) |
			(1 << arcgrpc.AttrBitIsPUGVPay) |
			(1 << arcgrpc.AttrBitSteinsGate),
	)
	spaceArc, arcTotal, err := s.dao.ArcSpaceSearch(ctx, &model.ArcSearchParam{
		Mid:     param.VMid,
		Order:   "pubdate",
		Pn:      param.PN,
		Ps:      param.PS,
		AttrNot: attrNot,
	})
	if err != nil {
		return nil, err
	}

	reply := &model.SpaceStoryReply{}
	reply.Meta.TitleTail = "的更多投稿"
	reply.Page.PN = param.PN
	reply.Page.PS = param.PS
	reply.Page.Total = arcTotal
	reply.Page.HasNext = true
	if reply.Page.PN*reply.Page.PS >= arcTotal {
		reply.Page.HasNext = false
	}
	reply.Items = make([]*story.SpaceItem, 0, len(spaceArc.VList))
	if len(spaceArc.VList) <= 0 {
		return reply, nil
	}

	aids := make([]int64, 0, len(spaceArc.VList))
	for _, v := range spaceArc.VList {
		aids = append(aids, v.Aid)
	}

	var (
		storyUpIDs        []int64
		amplayer          map[int64]*arcgrpc.ArcPlayer
		cardm             map[int64]*accountgrpc.Card
		statm             map[int64]*relationgrpc.StatReply
		authorRelations   map[int64]*relationgrpc.InterrelationReply
		likeMap, coinsMap map[int64]int64
		storyTags         map[int64][]*channelgrpc.Channel
		haslike, isFav    map[int64]int8
		liveRoomInfos     map[int64]*livegrpc.EntryRoomInfoResp_EntryList
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) (err error) {
		if amplayer, err = s.dao.ArcsPlayer(ctx, aids, "story", false); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, a := range amplayer {
			storyUpIDs = append(storyUpIDs, a.Arc.Author.Mid)
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if storyTags, err = s.dao.ResourceChannels(ctx, aids, param.Mid); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	g.Go(func(ctx context.Context) (err error) {
		if haslike, err = s.dao.HasLike(ctx, param.Buvid, param.Mid, aids); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if param.Mid > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if isFav, err = s.dao.IsFavVideos(ctx, param.Mid, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if coinsMap, err = s.dao.ArchiveUserCoins(ctx, aids, param.Mid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}
	g = errgroup.WithContext(ctx)
	if len(storyUpIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if cardm, err = s.dao.Cards3GRPC(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if statm, err = s.dao.StatsGRPC(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if authorRelations, err = s.dao.RelationsInterrelations(ctx, param.Mid, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if likeMap, err = s.dao.UserLikedCounts(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			if liveRoomInfos, err = s.dao.LiveRoomInfos(ctx, &livegrpc.EntryRoomInfoReq{
				EntryFrom:     []string{story.EntryFromStoryVideoUpIcon, story.EntryFromStoryVideoUpPanel},
				Uids:          storyUpIDs,
				Uid:           param.Mid,
				Uipstr:        metadata.String(ctx, metadata.RemoteIP),
				HttpsUrlReq:   true,
				Platform:      param.Platform,
				Build:         int64(param.Build),
				DeviceName:    param.DeviceName,
				Network:       param.Network,
				FilterOffline: 1,
				ReqBiz:        "/x/v2/feed/index/space/story",
			}); err != nil {
				log.Error("Failed to request LiveRoomInfos: %+v", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Failed to execute errgroup: %+v", err)
	}
	dev, _ := device.FromContext(ctx)
	for _, aid := range aids {
		arc, ok := amplayer[aid]
		if !ok || !cdm.AvIsNormalGRPC(arc) {
			continue
		}
		if uint64(arc.Arc.Attribute)&attrNot > 0 {
			log.Warn("Manually filtered archive: %+v: %d", arc, arc.Arc.Attribute)
			continue
		}
		if arc.Arc.RedirectURL != "" {
			log.Warn("Manually filtered archive by `RedirectURL`: %+v: %s", arc, arc.Arc.RedirectURL)
			continue
		}
		if card.CheckMidMaxInt32(arc.Arc.Author.Mid) && card.CheckMidMaxInt32Version(dev) {
			log.Warn("Filtered by mid int64: %d", arc.Arc.Author.Mid)
			continue
		}
		si := &story.SpaceItem{}
		si.StoryFrom(arc, cardm, statm, authorRelations, likeMap, coinsMap, haslike, isFav, storyTags[aid], s.hotAids,
			getLiveRoomBuilder(liveRoomInfos[arc.Arc.Author.Mid], story.EntryFromStoryVideoUpIcon, story.EntryFromStoryVideoUpPanel),
			param.Plat, param.Build, param.MobiApp, getRandomThumbUpAnimation(), param.Mid)
		reply.Items = append(reply.Items, si)
	}
	if itemLen := len(reply.Items); reply.Page.HasNext && itemLen%2 == 1 {
		reply.Items = reply.Items[0 : itemLen-1]
	}
	return reply, nil
}

func getRandomThumbUpAnimation() string {
	return _thumbUpAnimationSlice[rand.Intn(len(_thumbUpAnimationSlice))]
}

func mergeAids(in *uparcgrpc.ArcPassedStoryReply) []int64 {
	out := []int64{}
	for _, i := range in.NextArcs {
		out = append(out, i.Aid)
	}
	for _, i := range in.PrevArcs {
		out = append(out, i.Aid)
	}
	return out
}

func mergeArcs(in *uparcgrpc.ArcPassedStoryReply) []*uparcgrpc.StoryArcs {
	out := []*uparcgrpc.StoryArcs{}
	out = append(out, in.PrevArcs...)
	out = append(out, in.NextArcs...)
	return out
}

func need1080plus(userInfo *accountgrpc.Card) bool {
	return userInfo != nil && userInfo.Vip.Type > 0 && userInfo.Vip.Status == 1
}

// SpaceStoryCursor is
func (s *StoryService) SpaceStoryCursor(ctx context.Context, param *model.SpaceStoryCursorParam) (*model.SpaceStoryCursorReply, error) {
	//nolint:gomnd
	if param.AfterSize > 20 {
		param.AfterSize = 20
	}
	//nolint:gomnd
	if param.BeforeSize > 20 {
		param.BeforeSize = 20
	}
	reply := &model.SpaceStoryCursorReply{}
	reply.Meta.TitleTail = "的更多投稿"
	reply.Config.ShowButton = _storyButton
	reply.Config.ReplyZoomExp = s.replyZoomExp(param.Mid, param.Buvid, param.MobiApp, param.Build)
	reply.Config.ReplyNoDanmu = s.replyNoDanmu(param.Mid, param.Buvid)
	reply.Config.ReplyHighRaised = s.replyHighRaised(param.Mid, param.Buvid)
	reply.Config.SpeedPlayExp = s.speedPlayExp(param.Mid, param.Buvid)

	eg := errgroup.WithContext(ctx)
	var storyArc *uparcgrpc.ArcPassedStoryReply
	eg.Go(func(ctx context.Context) (err error) {
		storyArc, err = s.dao.ArcPassedStory(ctx, &uparcgrpc.ArcPassedStoryReq{
			Mid:       param.VMid,
			Aid:       param.Aid,
			Sort:      "desc",
			PrevCount: param.BeforeSize,
			NextCount: param.AfterSize,
			Rank:      param.Index,
		})
		return err
	})
	var userInfo *accountgrpc.Card
	if param.Mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if userInfo, err = s.dao.Card3(ctx, param.Mid); err != nil {
				log.Error("Failed to request card3: %+v", err)
				err = nil
			}
			return
		})
	}

	if err := eg.Wait(); err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			return reply, nil
		}
		return nil, s.slbRetryCode(err)
	}
	allArcs := mergeArcs(storyArc)
	allAids := mergeAids(storyArc)
	if param.Contain {
		allAids = append(allAids, param.Aid)
	}

	fanoutResult, err := s.doStoryFanout(ctx, &storyFanoutArgs{
		aids:              allAids,
		mid:               param.Mid,
		build:             param.Build,
		buvid:             param.Buvid,
		platform:          param.Platform,
		deviceName:        param.DeviceName,
		network:           param.Network,
		reqBiz:            "/x/v2/feed/index/space/story/cursor",
		liveEntryFrom:     []string{story.EntryFromStoryVideoUpIcon, story.EntryFromStoryVideoUpPanel},
		needUpReservation: false,
		need1080plus:      need1080plus(userInfo),
	})
	if err != nil {
		return nil, err
	}

	reply.Page.Total = storyArc.Total
	reply.Page.HasPrev, reply.Page.HasNext = cursorPageInitialState(storyArc.Total, allArcs)

	beforeArchives := s.makeSpaceCursorItems(ctx, storyArc.PrevArcs, storyArc.Total, fanoutResult, param)
	cursorArchives := []*story.SpaceCursorItem{}
	if param.Contain && storyArc.Rank > 0 {
		cursorMeta := []*uparcgrpc.StoryArcs{{Aid: param.Aid, Rank: storyArc.Rank}}
		cursorArchives = s.makeSpaceCursorItems(ctx, cursorMeta, storyArc.Total, fanoutResult, param)
	}
	afterArchives := s.makeSpaceCursorItems(ctx, storyArc.NextArcs, storyArc.Total, fanoutResult, param)
	func() {
		switch param.Position {
		case "left":
			beforeArchives = trimHead(beforeArchives, true)
			afterArchives = trimTail(afterArchives, !reply.Page.HasNext, false)
			return
		case "right":
			beforeArchives = trimHead(beforeArchives, false)
			afterArchives = trimTail(afterArchives, !reply.Page.HasNext, true)
			return
		default:
			if storyArc.Rank > 0 {
				if isOdd(int(storyArc.Rank)) {
					beforeArchives = trimHead(beforeArchives, true)
					afterArchives = trimTail(afterArchives, !reply.Page.HasNext, false)
					return
				}
				beforeArchives = trimHead(beforeArchives, false)
				afterArchives = trimTail(afterArchives, !reply.Page.HasNext, true)
				return
			}
			return
		}
	}()
	if param.BeforeSize > 0 && len(beforeArchives) <= 0 {
		reply.Page.HasPrev = false
	}
	if param.AfterSize > 0 && len(afterArchives) <= 0 {
		reply.Page.HasNext = false
	}
	reply.Items = append(reply.Items, beforeArchives...)
	reply.Items = append(reply.Items, cursorArchives...)
	reply.Items = append(reply.Items, afterArchives...)
	// 强制修正第一个稿件和最后一个稿件的 index
	if !reply.Page.HasNext {
		forceCastLastIndex(reply, param)
	}
	if !reply.Page.HasPrev {
		forceCastFirstIndex(reply, param)
	}
	return reply, nil
}

func (s *StoryService) slbRetryCode(originErr error) error {
	retryCode := []int{-500, -502, -504}
	for _, val := range retryCode {
		if xecode.EqualError(xecode.Int(val), originErr) {
			return errors.Wrapf(gateecode.AppSLBRetry, "%v", originErr)
		}
	}
	return originErr
}

func forceCastFirstIndex(in *model.SpaceStoryCursorReply, param *model.SpaceStoryCursorParam) {
	if len(in.Items) <= 0 {
		return
	}
	first := in.Items[0]
	if first.Index == 1 {
		return
	}
	log.Warn("Force cast first index to total: %d: %d: %+v", param.VMid, param.Aid, first)
	first.Index = 1
}

func forceCastLastIndex(in *model.SpaceStoryCursorReply, param *model.SpaceStoryCursorParam) {
	if len(in.Items) <= 0 {
		return
	}
	last := in.Items[len(in.Items)-1]
	if last.Index == in.Page.Total {
		return
	}
	log.Warn("Force cast last index to total: %d: %d: %+v", param.VMid, param.Aid, last)
	last.Index = in.Page.Total
}

func isOdd(in int) bool {
	return in%2 == 1
}

func trimHead(in []*story.SpaceCursorItem, toEven bool) []*story.SpaceCursorItem {
	if len(in) <= 0 {
		return in
	}
	if toEven {
		if isOdd(len(in)) {
			return in[1:]
		}
		return in
	}
	if !isOdd(len(in)) {
		return in[1:]
	}
	return in
}

func trimTail(in []*story.SpaceCursorItem, atLastPage bool, toEven bool) []*story.SpaceCursorItem {
	if len(in) <= 0 {
		return in
	}
	// 只在中间页去空窗
	if atLastPage {
		return in
	}
	if toEven {
		if isOdd(len(in)) {
			return in[:len(in)-1]
		}
		return in
	}
	if !isOdd(len(in)) {
		return in[:len(in)-1]
	}
	return in
}

type storyFanoutArgs struct {
	aids              []int64
	mid               int64
	build             int
	buvid             string
	platform          string
	deviceName        string
	network           string
	reqBiz            string
	liveEntryFrom     []string
	needUpReservation bool
	epids             []int32
	mobiApp           string
	device            string
	uids              []int64
	need1080plus      bool
}

type storyFanoutResult struct {
	amplayer          map[int64]*arcgrpc.ArcPlayer
	cardm             map[int64]*accountgrpc.Card
	statm             map[int64]*relationgrpc.StatReply
	authorRelations   map[int64]*relationgrpc.InterrelationReply
	likeMap, coinsMap map[int64]int64
	storyTags         map[int64][]*channelgrpc.Channel
	haslike           map[int64]int8
	isFavVideo        map[int64]int8
	isFavEp           map[int64]int8
	likeAnimationIcon map[int64]*thumbupgrpc.LikeAnimation
	liveRoomInfos     map[int64]*livegrpc.EntryRoomInfoResp_EntryList
	upReservationMap  map[int64]*story.ReservationInfo
	arguementMap      map[int64]*vogrpc.Argument
	eppm              map[int32]*pgcinline.EpisodeCard
}

// nolint:gocognit
func (s *StoryService) doStoryFanout(ctx context.Context, args *storyFanoutArgs) (*storyFanoutResult, error) {
	result := &storyFanoutResult{}
	g := errgroup.WithContext(ctx)
	avUpIDs := []int64{}
	argueAids := []int64{}
	g.Go(func(ctx context.Context) (err error) {
		if result.amplayer, err = s.dao.ArcsPlayer(ctx, args.aids, "story", args.need1080plus); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, a := range result.amplayer {
			avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
			if a.Arc.AttrVal(arcgrpc.AttrBitHasArgument) == arcgrpc.AttrYes {
				argueAids = append(argueAids, a.Arc.Aid)
			}
		}
		return
	})
	epUpIDs := []int64{}
	if len(args.epids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if result.eppm, err = s.dao.InlineCards(ctx, args.epids, args.mobiApp, args.platform, args.device,
				args.build, args.mid, false, args.buvid, nil); err != nil {
				log.Error("%+v", err)
				return
			}
			for _, ep := range result.eppm {
				epUpIDs = append(epUpIDs, ep.GetContributeUpInfo().GetMid())
			}
			return
		})
		if args.mid > 0 {
			g.Go(func(ctx context.Context) (err error) {
				if result.isFavEp, err = s.dao.IsFavEp(ctx, args.mid, feedcommon.Int32SliceToInt64Slice(args.epids)); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		}
	}
	g.Go(func(ctx context.Context) (err error) {
		if result.storyTags, err = s.dao.ResourceChannels(ctx, args.aids, args.mid); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		if result.haslike, err = s.dao.HasLike(ctx, args.buvid, args.mid, args.aids); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		if result.likeAnimationIcon, err = s.dao.MultiLikeAnimation(ctx, args.aids); err != nil {
			log.Error("Failed to MultiLikeAnimation: %+v", err)
			err = nil
		}
		return
	})
	if args.mid > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if result.isFavVideo, err = s.dao.IsFavVideos(ctx, args.mid, args.aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if result.coinsMap, err = s.dao.ArchiveUserCoins(ctx, args.aids, args.mid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	g = errgroup.WithContext(ctx)
	uidSet := sets.NewInt64(avUpIDs...)
	uidSet.Insert(epUpIDs...)
	uidSet.Insert(args.uids...)
	storyUpIDs := uidSet.List()
	if len(storyUpIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if result.cardm, err = s.dao.Cards3GRPC(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if result.statm, err = s.dao.StatsGRPC(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if result.authorRelations, err = s.dao.RelationsInterrelations(ctx, args.mid, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if result.likeMap, err = s.dao.UserLikedCounts(ctx, storyUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if result.liveRoomInfos, err = s.dao.LiveRoomInfos(ctx, &livegrpc.EntryRoomInfoReq{
				EntryFrom:     args.liveEntryFrom,
				Uids:          storyUpIDs,
				Uid:           args.mid,
				Uipstr:        metadata.String(ctx, metadata.RemoteIP),
				HttpsUrlReq:   true,
				Platform:      args.platform,
				Build:         int64(args.build),
				DeviceName:    args.deviceName,
				Network:       args.network,
				FilterOffline: 1,
				ReqBiz:        args.reqBiz,
			}); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		if args.needUpReservation {
			g.Go(func(ctx context.Context) (err error) {
				result.upReservationMap = s.doReservationInfo(ctx, storyUpIDs, args.mid)
				return nil
			})
		}
	}
	if len(argueAids) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if result.arguementMap, err = s.dao.Arguments(ctx, argueAids); err != nil {
				log.Error("Failed to request Arguments, param: %+v, %+v", argueAids, err)
				err = nil
			}
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Failed to execute errgroup: %+v", err)
	}
	return result, nil
}

func (s *StoryService) makeSpaceCursorItems(ctx context.Context, metas []*uparcgrpc.StoryArcs, total int64, fanoutResult *storyFanoutResult, param *model.SpaceStoryCursorParam) []*story.SpaceCursorItem {
	dev, _ := device.FromContext(ctx)
	attrNot := uint64(
		(1 << arcgrpc.AttrBitIsPGC) |
			(1 << arcgrpc.AttrBitIsPUGVPay) |
			(1 << arcgrpc.AttrBitSteinsGate),
	)
	buttonFilter := s.buttonFilter(param.MobiApp, param.Build)
	enableScreencast := s.enableScreencast(buttonFilter)
	needCoin := needCoinDev(param.MobiApp, param.Build)
	out := make([]*story.SpaceCursorItem, 0, len(metas))
	for _, m := range metas {
		arc, ok := fanoutResult.amplayer[m.Aid]
		if !ok || !cdm.AvIsNormalGRPC(arc) {
			log.Warn("Invalid archive: %d: %+v", m.Aid, arc)
			continue
		}
		if uint64(arc.Arc.Attribute)&attrNot > 0 {
			log.Warn("Manually filtered archive: %+v: %d", arc, arc.Arc.Attribute)
			continue
		}
		if cdm.AvIsCharging(arc) {
			log.Warn("Filtered by charging aid: %d", arc.Arc.Aid)
			continue
		}
		if arc.Arc.RedirectURL != "" {
			log.Warn("Manually filtered archive by `RedirectURL`: %+v: %s", arc, arc.Arc.RedirectURL)
			continue
		}
		if card.CheckMidMaxInt32(arc.Arc.Author.Mid) && card.CheckMidMaxInt32Version(dev) {
			log.Warn("Filtered by mid int64: %d", arc.Arc.Author.Mid)
			continue
		}
		si := &story.SpaceCursorItem{}
		si.StoryFrom(arc, fanoutResult.cardm, fanoutResult.statm, fanoutResult.authorRelations, fanoutResult.likeMap,
			fanoutResult.coinsMap, fanoutResult.haslike, fanoutResult.isFavVideo, fanoutResult.storyTags[m.Aid], s.hotAids,
			getLiveRoomBuilder(fanoutResult.liveRoomInfos[arc.Arc.Author.Mid], story.EntryFromStoryVideoUpIcon, story.EntryFromStoryVideoUpPanel),
			m, total, param.Plat, param.Build, param.MobiApp, getRandomThumbUpAnimation(), fanoutResult.arguementMap,
			param.Mid, enableScreencast, buttonFilter, needCoin, fanoutResult.likeAnimationIcon)
		if si.PlayerArgs == nil {
			log.Warn("Manually filtered archive by PlayerArgs nil: %+v: %d", arc, arc.Arc.Attribute)
			continue
		}
		out = append(out, si)
	}
	return out
}

func cursorPageInitialState(total int64, all []*uparcgrpc.StoryArcs) (bool, bool) {
	if len(all) <= 0 {
		return false, false
	}
	hasPrev := false
	if all[0].Rank > 1 {
		hasPrev = true
	}
	hasNext := false
	if all[len(all)-1].Rank < total {
		hasNext = true
	}
	return hasPrev, hasNext
}

func getLiveRoomBuilder(liveRoomInfo *livegrpc.EntryRoomInfoResp_EntryList, entryFromIcon, entryFromPanel string) func() *story.LiveRoom {
	return func() *story.LiveRoom {
		if liveRoomInfo == nil || liveRoomInfo.LiveStatus != 1 {
			return nil
		}
		return &story.LiveRoom{
			LiveStatus:      liveRoomInfo.LiveStatus,
			UpJumpURI:       liveRoomInfo.JumpUrl[entryFromIcon],
			UpPannelJumpURI: liveRoomInfo.JumpUrl[entryFromPanel],
		}
	}
}

type dynamicStoryArg struct {
	versionctrl *dyncommongrpc.VersionCtrlMeta
	uid         int64
	vmid        int64
	offset      string
	pageNumber  int64
	pageSize    int64
	typeList    []string
	scene       string
	aid         int64
	nextUid     int64
	uidPos      int64
	topicId     int64
	topicRid    int64
	topicType   int64
	offsetType  string
	seasonId    int32
	buvid       string
	pull        int
	storyParam  string
}

func fakeAvItem(in []int64) []*ai.SubItems {
	out := make([]*ai.SubItems, 0, len(in))
	for _, v := range in {
		out = append(out, &ai.SubItems{
			ID:   v,
			Goto: model.GotoVerticalAv,
		})
	}
	return out
}

func (s *StoryService) DynamicStory(ctx context.Context, param *story.DynamicStoryParam) (*story.DynamicStoryReply, error) {
	versionalCtrl := &dyncommongrpc.VersionCtrlMeta{
		Build:     strconv.Itoa(param.Build),
		Platform:  param.Platform,
		MobiApp:   param.MobiApp,
		Buvid:     param.Buvid,
		Device:    param.Device,
		Ip:        metadata.String(ctx, metadata.RemoteIP),
		From:      strconv.Itoa(param.From),
		FromSpmid: param.FromSpmid,
	}
	dynArg := &dynamicStoryArg{
		versionctrl: versionalCtrl,
		uid:         param.Mid,
		vmid:        param.Vmid,
		offset:      param.Offset,
		pageNumber:  param.DisplayID,
		pageSize:    _max,
		typeList:    []string{"8_1", "8_3"},
		scene:       param.Scene,
		aid:         param.Aid,
		nextUid:     param.NextUid,
		uidPos:      param.UidPos,
		topicId:     param.TopicId,
		topicRid:    param.TopicRid,
		topicType:   param.TopicType,
		offsetType:  param.OffsetType,
		seasonId:    param.SeasonID,
		buvid:       param.Buvid,
		pull:        param.Pull,
		storyParam:  param.StoryParam,
	}
	eg := errgroup.WithContext(ctx)
	var resp *dynamicVideoStoryReply
	eg.Go(func(ctx context.Context) (err error) {
		resp, err = s.dynamicVideoStory(ctx, dynArg)
		if err != nil {
			log.Error("Failed to request dynamicVideoStory, args: %+v, version: %+v, err: %+v", dynArg,
				versionalCtrl, err)
		}
		return err
	})
	var userInfo *accountgrpc.Card
	if param.Mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if userInfo, err = s.dao.Card3(ctx, param.Mid); err != nil {
				log.Error("Failed to request card3: %+v", err)
				err = nil
			}
			return
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	out := &story.DynamicStoryReply{
		Items: []*story.Item{},
		Page: &story.DynamicStoryPage{
			HasMore:        resp.HasMore,
			Offset:         resp.Offset,
			NextUid:        resp.NextUid,
			OffsetType:     resp.OffsetType,
			SeasonID:       resp.SeasonId,
			HasPrev:        resp.HasPrev,
			PrevOffset:     resp.PrevOffset,
			PrevOffsetType: resp.PrevOffsetType,
		},
		Config: &story.StoryConfig{
			ProgressBar:     s.setupProgressBar(),
			ShowButton:      _storyButton,
			ReplyZoomExp:    s.replyZoomExp(param.Mid, param.Buvid, param.MobiApp, param.Build),
			ReplyNoDanmu:    s.replyNoDanmu(param.Mid, param.Buvid),
			ReplyHighRaised: s.replyHighRaised(param.Mid, param.Buvid),
			SpeedPlayExp:    s.speedPlayExp(param.Mid, param.Buvid),
		},
	}
	s.appendToViewCfg(out.Config, param.MobiApp, param.Build)
	allAids, allEpids, items := mergeDynIds(resp.Items, param)
	if len(allAids)+len(allEpids) == 0 {
		// 客户端要求，没有数据时返回空而非-404
		return out, nil
	}
	fanoutResult, err := s.doStoryFanout(ctx, &storyFanoutArgs{
		aids:              allAids,
		mid:               param.Mid,
		build:             param.Build,
		buvid:             param.Buvid,
		platform:          param.Platform,
		deviceName:        param.DeviceName,
		network:           param.Network,
		reqBiz:            "/x/v2/feed/index/dynamic/story",
		liveEntryFrom:     []string{story.EntryFromStoryDtUpicon, story.EntryFromStoryDtUpPannel},
		needUpReservation: true,
		epids:             allEpids,
		mobiApp:           param.MobiApp,
		device:            param.Device,
		uids:              filterUids(items),
		need1080plus:      need1080plus(userInfo),
	})
	if err != nil {
		return nil, err
	}
	dev, _ := device.FromContext(ctx)
	buttonFilter := s.buttonFilter(param.MobiApp, param.Build)
	enableScreencast := s.enableScreencast(buttonFilter)
	needCoin := needCoinDev(param.MobiApp, param.Build)
	for _, item := range items {
		item.SetDisableRcmd(param.DisableRcmd)
		switch item.Goto {
		case model.GotoVerticalAv:
			arc, ok := fanoutResult.amplayer[item.ID]
			if !ok || !cdm.AvIsNormalGRPC(arc) {
				log.Warn("Invalid archive: %d: %+v", item.ID, arc)
				continue
			}
			if cdm.AvIsCharging(arc) {
				log.Warn("Filtered by charging aid: %d", arc.Arc.Aid)
				continue
			}
			if card.CheckMidMaxInt32(arc.Arc.Author.Mid) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", arc.Arc.Author.Mid)
				continue
			}
			i := &story.Item{}
			fns := []story.StoryFn{
				story.OptThreePointButton(story.NeedDislike(false),
					story.NeedReport(false),
					story.NeedCoin(needCoin),
					story.NeedScreencast(enableScreencast),
					story.NoPlayBackground(arc.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
				),
				story.OptShareBottomButton(story.NeedDislike(false),
					story.NeedCoin(needCoin),
					story.NeedScreencast(enableScreencast),
					story.CoinNum(cdm.StatString(arc.Arc.Stat.Coin, "")),
					story.NoPlayBackground(arc.Arc.Rights.NoBackground),
					story.ButtonFilter(buttonFilter),
				),
			}
			i.StoryFrom(item, arc, fanoutResult.cardm, fanoutResult.statm, fanoutResult.authorRelations,
				fanoutResult.likeMap, fanoutResult.coinsMap, fanoutResult.haslike, fanoutResult.isFavVideo,
				fanoutResult.storyTags[item.ID], s.hotAids,
				getLiveRoomBuilder(fanoutResult.liveRoomInfos[story.TargetMid(item, arc.Arc.Author.Mid)],
					story.EntryFromStoryDtUpicon, story.EntryFromStoryDtUpPannel),
				param.Plat, param.Build, param.MobiApp, getRandomThumbUpAnimation(),
				fanoutResult.upReservationMap[story.TargetMid(item, arc.Arc.Author.Mid)],
				fanoutResult.arguementMap, cdm.FfCoverFromDynamicStory, param.Mid, nil,
				fanoutResult.likeAnimationIcon, fns...)
			out.Items = append(out.Items, i)
		case model.GotoVerticalPgc:
			ep, ok := fanoutResult.eppm[int32(item.ID)]
			if !ok {
				log.Warn("Invalid ep: %d", item.ID)
				continue
			}
			if card.CheckMidMaxInt32(ep.Season.GetUpInfo().GetMid()) && card.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", ep.Season.GetUpInfo().GetMid())
				continue
			}
			fns := []story.StoryFn{
				story.OptThreePointButton(story.NeedDislike(false),
					story.NeedReport(false),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.NoPlayBackground(0),
					story.ButtonFilter(buttonFilter),
					story.NoWatchLater(true),
				),
				story.OptShareBottomButton(story.NeedDislike(false),
					story.NeedCoin(needCoin),
					story.NeedScreencast(false),
					story.CoinNum(cdm.Stat64String(ep.Stat.Coin, "")),
					story.ButtonFilter(buttonFilter),
					story.NoPlayBackground(0),
					story.NoWatchLater(true),
				),
				story.OptPGCStyle(item, ep),
			}
			i := &story.Item{}
			i.StoryFromPGC(item, ep, fanoutResult.cardm, fanoutResult.statm, fanoutResult.authorRelations,
				fanoutResult.likeMap, fanoutResult.coinsMap, fanoutResult.haslike, fanoutResult.isFavEp,
				getRandomThumbUpAnimation(), param.Mid, nil, fns...)
			out.Items = append(out.Items, i)
		default:
			log.Warn("Unsupported goto: %+v", item.Goto)
		}
	}
	return out, nil
}

func filterUids(items []*ai.SubItems) []int64 {
	var out []int64
	for _, item := range items {
		if item.Uid() > 0 {
			out = append(out, item.Uid())
		}
	}
	return out
}

func (s *StoryService) dynamicVideoStory(ctx context.Context, arg *dynamicStoryArg) (*dynamicVideoStoryReply, error) {
	switch arg.scene {
	case "dynamic":
		resp, err := s.dao.DynamicGeneralStory(ctx, &dyngrpc.GeneralStoryReq{
			Uid:         arg.uid,
			Offset:      arg.offset,
			PageNumber:  arg.pageNumber,
			PageSize:    30,
			TypeList:    arg.typeList,
			VersionCtrl: arg.versionctrl,
			Rid:         arg.aid,
			ExtraParam:  arg.storyParam,
		})
		if err != nil {
			return nil, err
		}
		return &dynamicVideoStoryReply{
			Items:      fakeAvItem(resp.GetRid()),
			Offset:     resp.GetOffset(),
			OffsetType: "dynamic",
			HasMore:    resp.GetHasMore(),
		}, nil
	case "dynamic_space":
		resp, err := s.dao.DynamicSpaceStory(ctx, &dyngrpc.SpaceStoryReq{
			HostUid:     arg.vmid,
			Uid:         arg.uid,
			Offset:      arg.offset,
			PageNumber:  arg.pageNumber,
			PageSize:    arg.pageSize,
			TypeList:    arg.typeList,
			VersionCtrl: arg.versionctrl,
			Rid:         arg.aid,
		})
		if err != nil {
			return nil, err
		}
		return &dynamicVideoStoryReply{
			Items:      fakeAvItem(resp.GetRid()),
			Offset:     resp.GetOffset(),
			OffsetType: "dynamic",
			HasMore:    resp.GetHasMore(),
		}, nil
	case "dynamic_insert":
		resp, err := s.dynInsert(ctx, arg)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case "topic_rcmd", "topic_hot", "topic_new":
		vsr := &topicgrpc.VideoStoryReq{
			TopicId:    arg.topicId,
			FromSortBy: fakeTopicFrom(arg.scene),
			Offset:     arg.offset,
			PageSize:   _max,
			Rid:        arg.topicRid,
			Type:       int32(arg.topicType),
			Uid:        arg.uid,
		}
		if arg.topicType == 1 {
			vsr.MetaData = &topiccommon.MetaDataCtrl{
				From: "topic_from_story_mode",
			}
		}
		resp, err := s.dao.TopicStory(ctx, vsr)
		if err != nil {
			return nil, err
		}
		return &dynamicVideoStoryReply{
			Items:      fakeAvItem(resp.GetRid()),
			Offset:     resp.GetOffset(),
			OffsetType: "topic",
			HasMore:    resp.GetHasMore(),
		}, nil
	case "ogv_playlist":
		resp, err := s.ogvPlaylist(ctx, arg)
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		return nil, ecode.RequestErr
	}
}

type dynamicVideoStoryReply struct {
	Items          []*ai.SubItems
	Offset         string
	OffsetType     string
	HasMore        bool
	SeasonId       int32
	NextUid        int64
	HasPrev        bool
	PrevOffset     string
	PrevOffsetType string
}

func (s *StoryService) ogvPlaylist(ctx context.Context, arg *dynamicStoryArg) (*dynamicVideoStoryReply, error) {
	id, err := strconv.ParseInt(arg.offset, 10, 64)
	if err != nil {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("%+v", errors.WithStack(err)))
	}
	index := &pgcstory.PlayIndex{
		IsPgc:    arg.offsetType == "pgc",
		SeasonId: arg.seasonId,
	}
	if index.IsPgc {
		index.EpisodeId = int32(id)
	} else {
		index.Avid = id
	}
	reply, err := s.dao.OgvPlaylist(ctx, &pgcstory.StoryPlayListReq{
		Cursor:   index,
		PageSize: _pgcNum,
		User: &pgcstory.User{
			Mid:   arg.uid,
			Ip:    metadata.String(ctx, metadata.RemoteIP),
			Buvid: arg.buvid,
		},
		ToPrev: parseToPrev(arg.pull),
	})
	if err != nil {
		return nil, err
	}
	fakeItems := make([]*ai.SubItems, 0, len(reply.PlayIndex))
	for _, v := range reply.PlayIndex {
		if v.IsPgc {
			item := &ai.SubItems{ID: int64(v.EpisodeId), Goto: model.GotoVerticalPgc, OGVStyle: 1}
			item.SetPgcAid(v.Avid)
			fakeItems = append(fakeItems, item)
			continue
		}
		fakeItems = append(fakeItems, &ai.SubItems{ID: v.Avid, Goto: model.GotoVerticalAv})
	}
	offset, offsetType := parseOffset(reply.GetCursor())
	prevOffset, prevOffsetType := parseOffset(reply.GetPrevCursor())
	return &dynamicVideoStoryReply{
		Items:          fakeItems,
		Offset:         offset,
		OffsetType:     offsetType,
		HasMore:        reply.GetHasMore(),
		SeasonId:       reply.GetCursor().GetSeasonId(),
		NextUid:        0,
		HasPrev:        reply.GetHasPrev(),
		PrevOffset:     prevOffset,
		PrevOffsetType: prevOffsetType,
	}, nil
}

func parseToPrev(pull int) bool {
	return pull == 1
}

func parseOffset(arg *pgcstory.PlayIndex) (string, string) {
	if arg == nil {
		return "", ""
	}
	if arg.IsPgc {
		return strconv.FormatInt(int64(arg.EpisodeId), 10), "pgc"
	}
	return strconv.FormatInt(arg.Avid, 10), "ugc"
}

func (s *StoryService) dynInsert(ctx context.Context, arg *dynamicStoryArg) (*dynamicVideoStoryReply, error) {
	nextOffset := &dyngrpc.NextOffset{
		Uid:    arg.nextUid,
		Offset: arg.offset,
	}
	req := &dyngrpc.InsertedStoryReq{
		Uid:        arg.uid,
		NextOffset: nextOffset,
		UidPos:     arg.uidPos,
		PageNumber: arg.pageNumber,
		PageSize:   arg.pageSize,
		TypeList:   arg.typeList,
		Meta: &dyncommongrpc.MetaDataCtrl{
			Platform:  arg.versionctrl.Platform,
			Build:     arg.versionctrl.Build,
			MobiApp:   arg.versionctrl.MobiApp,
			Buvid:     arg.versionctrl.Buvid,
			Device:    arg.versionctrl.Device,
			FromSpmid: arg.versionctrl.FromSpmid,
			From:      arg.versionctrl.From,
			Ip:        arg.versionctrl.Ip,
		},
	}
	reply, err := s.dao.DynamicInsert(ctx, req)
	if err != nil {
		return nil, err
	}
	fakeItems := make([]*ai.SubItems, 0, len(reply.GetRidInfo()))
	for _, info := range reply.GetRidInfo() {
		item := &ai.SubItems{ID: info.GetRid(), Goto: model.GotoVerticalAv}
		item.SetUid(info.GetUid())
		fakeItems = append(fakeItems, item)
	}
	return &dynamicVideoStoryReply{
		Items:      fakeItems,
		Offset:     reply.GetNextOffset().GetOffset(),
		OffsetType: "dynamic_insert",
		HasMore:    reply.GetHasMore(),
		NextUid:    reply.GetNextOffset().GetUid(),
	}, nil
}

func fakeTopicFrom(scene string) int64 {
	const (
		_rcmd   = 1
		_hot    = 2
		_newest = 3
	)
	switch scene {
	case "topic_rcmd":
		return _rcmd
	case "topic_hot":
		return _hot
	case "topic_new":
		return _newest
	default:
		return -1
	}
}

var NoAppendScene = sets.NewString("dynamic_insert", "ogv_playlist")

func mergeDynIds(items []*ai.SubItems, param *story.DynamicStoryParam) ([]int64, []int32, []*ai.SubItems) {
	var aids []int64
	var result []*ai.SubItems
	if param.DisplayID == 1 && param.Aid > 0 && !NoAppendScene.Has(param.Scene) {
		result = append(result, &ai.SubItems{ID: param.Aid, Goto: model.GotoVerticalAv})
	}
	result = append(result, items...)
	var epids []int32
	for _, item := range result {
		switch item.Goto {
		case model.GotoVerticalAv:
			aids = append(aids, item.ID)
		case model.GotoVerticalPgc:
			epids = append(epids, int32(item.ID))
			aids = append(aids, item.PgcAid())
		default:
			log.Warn("Unsupported goto: %+s", item.Goto)
		}
	}
	return aids, epids, result
}

func (s *StoryService) StoryCart(ctx context.Context, param *model.StoryCartParam) (*model.StoryCartReply, error) {
	if param.Aid == 0 {
		return nil, ecode.NothingFound
	}
	param.IP = metadata.String(ctx, metadata.RemoteIP)
	g := errgroup.WithContext(ctx)
	var archiveMap map[int64]*arcgrpc.Arc
	g.Go(func(ctx context.Context) (err error) {
		archiveMap, err = s.dao.Archives(ctx, []int64{param.Aid}, param.Mid, param.MobiApp, param.Device)
		if err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	var location *locgrpc.InfoReply
	g.Go(func(ctx context.Context) (err error) {
		if location, err = s.dao.InfoGRPC(ctx, param.IP); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	FillStoryCartParam(param, archiveMap, location)
	out, err := s.dao.StoryCart(ctx, param)
	if err != nil {
		log.Error("Failed to request StoryCart: %+v", err)
		return nil, err
	}
	return out, nil
}

func FillStoryCartParam(param *model.StoryCartParam, arcs map[int64]*arcgrpc.Arc, location *locgrpc.InfoReply) {
	archive, ok := arcs[param.Aid]
	if ok {
		param.AvRid = int64(archive.TypeID)
		param.AvUpId = archive.Author.Mid
	}
	if location != nil {
		param.Country = location.Country
		param.Province = location.Province
		param.City = location.City
	}
}

// nolint:gomnd
func liveAttentionGroup(mid int64) int {
	digest := md5.Sum([]byte(fmt.Sprintf("%d_rubick_live", mid)))
	md5_sum := 0
	for i := 9; i < 16; i++ {
		md5_sum = (md5_sum << 8) + int(digest[i])
	}
	return md5_sum % 14
}

func (s *StoryService) liveAttentionExp(mid int64) int {
	if val, ok := s.cfg.CustomConfig.StoryLiveAttentionMidGroup[strconv.FormatInt(mid, 10)]; ok {
		return val
	}
	group := liveAttentionGroup(mid)
	return s.cfg.CustomConfig.StoryLiveAttentionGroup[strconv.Itoa(group)]
}

func (s *StoryService) matchFeatureControl(mid int64, buvid, feature string) bool {
	if s.cfg.FeatureControl.DisableAll {
		return false
	}
	if feature == "" {
		return false
	}
	policy, ok := s.cfg.FeatureControl.Feature[feature]
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

func (s *StoryService) enableScreencast(buttonFilter bool) bool {
	return !buttonFilter
}

func (s *StoryService) buttonFilter(mobiApp string, build int) bool {
	return mobiApp == "iphone" && build < 67900000
}

func (s *StoryService) jumpToSeason(mid int64, buvid string) bool {
	if s.matchFeatureControl(mid, buvid, "jump_to_season") {
		return true
	}
	return crc32.ChecksumIEEE([]byte(buvid+"_jump_to_season"))%10 < uint32(s.cfg.CustomConfig.JumpToSeason)
}

func (s *StoryService) StoryGameStatus(ctx context.Context, param *model.StoryGameParam) (*model.StoryGameReply, error) {
	out, err := s.dao.GameGifts(ctx, param)
	if err != nil {
		log.Error("Failed to request GameGifts: %+v, %+v", param, err)
	}
	return out, nil
}

func needCoinDev(mobiApp string, build int) bool {
	return (mobiApp == "android" && build < 6810000) || (mobiApp == "iphone" && build < 68100000)
}
