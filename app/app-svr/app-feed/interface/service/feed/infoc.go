package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-common/library/utils/collection"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

type infoc struct {
	typ              string
	mid              string
	client           string
	build            string
	buvid            string
	ip               string
	style            string
	api              string
	now              string
	isRcmd           string
	pull             string
	userFeature      json.RawMessage
	code             string
	items            []*ai.Item
	zoneID           string
	adResponse       string
	deviceID         string
	network          string
	newUser          string
	flush            string
	autoPlay         string
	deviceType       string
	isGifCover       map[int64]int
	bannerHash       string
	clientBannerHash string
	loginEvent       string
	adError          string
	adCount          string
	adPkCode         string
	openEvent        string
	pendentMap       map[int64]string
	subGotoMap       map[int][]string
	cardTypes        map[int]string
	oidMap           map[int][]string
	ogvInfoMap       map[int]*feed.BangumiRcmdInfoc
	mobiApp          string
	discardReason    map[int64]*feed.Discard
	cardGotos        map[int]string
	gameBadge        map[int64]string
	badgeMap         map[int64]string
}

func (s *Service) IndexInfoc(c context.Context, mid int64, plat int8, build int, buvid, api string, userFeature json.RawMessage, style, code int, items []*ai.Item, isRcmd, pull, newUser bool, now time.Time, adResponse, deviceID, network string, flush int, autoPlay string, deviceType int, info *locgrpc.InfoReply, isGifCover map[int64]int, bannerHash, clientBannerHash string, loginEvent, adCode int, adError error, addPos, adPkCode []string, openEvent, isMelloi string, pendentMap map[int64]string, subGotoMap map[int][]string, cardTypes map[int]string,
	oidMap map[int][]string, ogvInfoMap map[int]*feed.BangumiRcmdInfoc, mobiApp string, discardReason map[int64]*feed.Discard, cardGotos map[int]string, gameBadge map[int64]string, badgeMap map[int64]string) {
	if isMelloi != "" {
		return
	}
	if items == nil {
		return
	}
	var (
		isRc      = "0"
		isPull    = "0"
		isNewUser = "0"
		zoneID    int64
	)
	if isRcmd {
		isRc = "1"
	}
	if pull {
		isPull = "1"
	}
	if newUser {
		isNewUser = "1"
	}
	ip := metadata.String(c, metadata.RemoteIP)
	if info != nil {
		zoneID = info.ZoneId
	}
	addPosStr := "None"
	if len(addPos) > 0 {
		addPosStr = strings.Join(addPos, ",")
	}
	s.infoc(infoc{"综合推荐", strconv.FormatInt(mid, 10), strconv.Itoa(int(plat)), strconv.Itoa(build), buvid, ip, strconv.Itoa(style), api, strconv.FormatInt(now.Unix(), 10),
		isRc, isPull, userFeature, strconv.Itoa(code), items, strconv.FormatInt(zoneID, 10), adResponse, deviceID, network, isNewUser, strconv.Itoa(flush), autoPlay,
		strconv.Itoa(deviceType), isGifCover, bannerHash, clientBannerHash, strconv.Itoa(loginEvent), fmt.Sprintf("code(%d),message(%v)", adCode, adError), addPosStr,
		strings.Join(adPkCode, ","), openEvent, pendentMap, subGotoMap, cardTypes, oidMap, ogvInfoMap, mobiApp, discardReason, cardGotos, gameBadge, badgeMap})
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

// writeInfoc
// nolint: gocognit
func (s *Service) infocproc() {
	const (
		// infoc format {"section":{"id":"%s推荐","pos":1,"style":%d,"items":[{"id":%s,"pos":%d,"type":1,"url":""}]}}
		noItem1 = `{"section":{"id":"`
		noItem2 = `{","pos":1,"style":`
		noItem3 = `,"items":[]}}`
	)
	// is_ad_loc, resource_id，source_id, creative_id,
	var (
		msg1              = []byte(`{"section":{"id":"`)
		msg2              = []byte(`","pos":1,"style":`)
		msg3              = []byte(`,"items":[`)
		msg4              = []byte(`{"id":`)
		msg5              = []byte(`,"pos":`)
		msg6              = []byte(`,"type":`)
		msg21             = []byte(`,"goto":"`)
		msgPosRecUniqueID = []byte(`,"pos_rec_unique_id":`)
		msg7              = []byte(`","source":"`)
		msg8              = []byte(`","tid":`)
		msg24             = []byte(`,"sub_goto":`)
		msg20             = []byte(`,"is_gifcover":`)
		msg23             = []byte(`,"dynamic_cover":`)
		msg9              = []byte(`,"av_feature":`)
		msg10             = []byte(`,"url":"`)
		msg11             = []byte(`","rcmd_reason":"`)
		msg19             = []byte(`","live_pendent":"`)
		msg22             = []byte(`","hash":"`)
		msg25             = []byte(`","banner_info":"`)
		msg29             = []byte(`","banner_creative_id":"`)
		msg12             = []byte(`","is_ad_loc":`)
		msg13             = []byte(`,"resource_id":`)
		msg14             = []byte(`,"source_id":`)
		msg15             = []byte(`,"creative_id":`)
		msg26             = []byte(`,"material_id":`)
		msg27             = []byte(`,"is_starlight_live":`)
		msg28             = []byte(`,"starlight_order_id":`)
		msg30             = []byte(`","game_pendent":"`)
		msg31             = []byte(`","badge":"`)
		msg16             = []byte(`},`)
		msg17             = []byte(`"},`)
		msg18             = []byte(`]}}`)
		buf               bytes.Buffer
		list              string
	)
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case infoc:
			var showBanner bool
			trackID := ""
			if len(l.items) > 0 {
				buf.Write(msg1)
				buf.WriteString(l.typ)
				buf.Write(msg2)
				buf.WriteString(l.style)
				buf.Write(msg3)
				for i, v := range l.items {
					if v == nil {
						continue
					}
					if v.TrackID != "" {
						trackID = v.TrackID
					}
					buf.Write(msg4)
					itemID := v.ID
					switch v.Goto {
					case model.GotoConvergeAi:
						itemID = v.ID + _convergeAi
					case model.GotoBanner:
						showBanner = true
					}
					buf.WriteString(strconv.FormatInt(itemID, 10))
					buf.Write(msg5)
					buf.WriteString(strconv.Itoa(i + 1))
					buf.Write(msg6)
					buf.WriteString(gotoMapID(v.Goto))
					buf.Write(msgPosRecUniqueID)
					buf.WriteString(fmt.Sprintf("\"%s\"", v.PosRecUniqueID))
					buf.Write(msg21)
					buf.WriteString(v.Goto)
					buf.Write(msg7)
					buf.WriteString(v.Source)
					buf.Write(msg8)
					buf.WriteString(strconv.FormatInt(v.Tid, 10))
					if v.CustomizedTitle != "" {
						buf.WriteString(fmt.Sprintf(`,"customized_title":"%s"`, v.CustomizedTitle))
					}
					if v.CustomizedCover != "" {
						buf.WriteString(fmt.Sprintf(`,"customized_cover":"%s"`, v.CustomizedCover))
					}
					if v.JumpGoto != "" {
						buf.WriteString(fmt.Sprintf(`,"jump_goto":"%s"`, v.JumpGoto))
					}
					if v.StNewCover != 0 {
						buf.WriteString(fmt.Sprintf(`,"st_new_cover":"%s"`, strconv.FormatInt(int64(v.StNewCover), 10)))
					}
					if subGoto, ok := l.subGotoMap[i]; ok {
						bt, err := json.Marshal(subGoto)
						if err != nil {
							log.Error("subGoto marshal failed, subGoto(%+v)", subGoto)
						} else {
							buf.Write(msg24)
							buf.WriteString(string(bt))
						}
					}
					if oids, ok := l.oidMap[i]; ok {
						bt, err := json.Marshal(oids)
						if err != nil {
							log.Error("oids marshal failed, oids(%+v)", oids)
						} else {
							buf.WriteString(fmt.Sprintf(`,"oid":%s`, string(bt)))
						}
					}
					if ogvInfo, ok := l.ogvInfoMap[i]; ok {
						bt, err := json.Marshal(ogvInfo.SeasonId)
						if err != nil {
							log.Error("season_id marshal failed, season_id(%+v)", ogvInfo.SeasonId)
						} else {
							buf.WriteString(fmt.Sprintf(`,"season_id":%s`, string(bt)))
						}
						bt, err = json.Marshal(ogvInfo.Epid)
						if err != nil {
							log.Error("epid marshal failed, epid(%+v)", ogvInfo.Epid)
						} else {
							buf.WriteString(fmt.Sprintf(`,"epid":%s`, string(bt)))
						}
					}
					if cardType, ok := l.cardTypes[i]; ok {
						buf.WriteString(fmt.Sprintf(`,"card_type":"%s"`, cardType))
					}
					if cardGoto, ok := l.cardGotos[i]; ok {
						buf.WriteString(fmt.Sprintf(`,"card_goto":"%s"`, cardGoto))
					}
					if isGif, ok := l.isGifCover[v.ID]; ok {
						buf.Write(msg20)
						buf.WriteString(strconv.Itoa(isGif))
					}
					if v.DynamicCover > 0 {
						buf.Write(msg23)
						buf.WriteString(strconv.FormatInt(int64(v.DynamicCover), 10))
					}
					if v.CreativeId > 0 {
						buf.Write(msg26)
						buf.WriteString(strconv.FormatInt(v.CreativeId, 10))
					}
					if v.IsStarlightLive > 0 {
						buf.Write(msg27)
						buf.WriteString(strconv.FormatInt(v.IsStarlightLive, 10))
					}
					if v.StarlightOrderID > 0 {
						buf.Write(msg28)
						buf.WriteString(strconv.FormatInt(v.StarlightOrderID, 10))
					}
					if v.LiveInlineDanmu > 0 {
						buf.WriteString(fmt.Sprintf(`,"live_inline_danmu":"%d"`, v.LiveInlineDanmu))
					}
					if v.LiveInlineLightDanmu > 0 {
						buf.WriteString(fmt.Sprintf(`,"live_inline_light_danmu":"%d"`, v.LiveInlineLightDanmu))
					}
					if v.LiveInlineLight > 0 {
						buf.WriteString(fmt.Sprintf(`,"live_inline_light":"%d"`, v.LiveInlineLight))
					}
					buf.Write(msg9)
					if v.AvFeature != nil {
						buf.Write(v.AvFeature)
					} else {
						buf.Write([]byte(`""`))
					}
					buf.Write(msg10)
					buf.WriteString("")
					buf.Write(msg11)
					if v.RcmdReason != nil {
						buf.WriteString(v.RcmdReason.Content)
					}
					if pendent, ok := l.pendentMap[v.ID]; ok {
						buf.Write(msg19)
						buf.WriteString(pendent)
					}
					if gamebadge, hasGBadge := l.gameBadge[v.ID]; hasGBadge {
						buf.Write(msg30)
						buf.WriteString(gamebadge)
					}
					if badge, hasBadge := l.badgeMap[v.ID]; hasBadge {
						buf.Write(msg31)
						buf.WriteString(badge)
					}
					switch v.Goto {
					case model.GotoBanner:
						buf.Write(msg22)
						buf.WriteString(l.bannerHash)
						if v.BannerInfo != nil && len(v.BannerInfo.Items) > 0 {
							buf.Write(msg25)
							buf.WriteString(joinBannerInfo(v.BannerInfo.Items))
						}
						buf.Write(msg29)
						buf.WriteString(joinBannerCreativeID(v.Banners))
					}
					if v.Ad != nil {
						buf.Write(msg12)
						buf.WriteString(strconv.FormatBool(v.Ad.IsAdLoc))
						buf.Write(msg13)
						buf.WriteString(strconv.FormatInt(v.Ad.Resource, 10))
						buf.Write(msg14)
						buf.WriteString(strconv.Itoa(int(v.Ad.Source)))
						buf.Write(msg15)
						buf.WriteString(strconv.FormatInt(v.Ad.CreativeID, 10))
						buf.Write(msg16)
					} else {
						buf.Write(msg17)
					}
				}
				buf.Truncate(buf.Len() - 1)
				buf.Write(msg18)
				list = buf.String()
				buf.Reset()
			} else {
				list = noItem1 + l.typ + noItem2 + l.style + noItem3
			}
			bannerCase := strconv.Itoa(bannerShowCase(l, showBanner))
			discardList := make([]*feed.Discard, 0, len(l.discardReason))
			for _, v := range l.discardReason {
				discardList = append(discardList, v)
			}
			dl, err := json.Marshal(discardList)
			if err != nil {
				log.Error("Failed to Marshal discardList: %+v", err)
			}
			event := infocV2.NewLogStreamV(s.c.Feed.FeedInfocID,
				log.String(l.ip),
				log.String(l.now),
				log.String(l.api),
				log.String(l.buvid),
				log.String(l.mid),
				log.String(l.client),
				log.String(l.pull),
				log.String(list),
				log.String(""),
				log.String(l.isRcmd),
				log.String(l.build),
				log.String(l.code),
				log.String(string(l.userFeature)),
				log.String(l.zoneID),
				log.String(l.adResponse),
				log.String(l.deviceID),
				log.String(l.network),
				log.String(l.newUser),
				log.String(l.flush),
				log.String(l.autoPlay),
				log.String(trackID),
				log.String(l.deviceType),
				log.String(bannerCase),
				log.String(l.clientBannerHash),
				log.String(l.loginEvent),
				log.String(l.adError),
				log.String(l.adCount),
				log.String(l.adPkCode),
				log.String(l.style),
				log.String(l.mobiApp),
				log.String(string(dl)),
			)
			if err := s.infocV2Log.Info(context.Background(), event); err != nil {
				log.Error("Failed to infoc feed index: %s, %s, %s, %s, %+v", l.mid, l.buvid, l.mobiApp, l.build, err)
			}
		}
	}
}

func gotoMapID(gt string) (id string) {
	switch gt {
	case model.GotoAv:
		id = "1"
	case model.GotoBangumi:
		id = "2"
	case model.GotoLive:
		id = "3"
	case model.GotoRank:
		id = "6"
	case model.GotoAdAv:
		id = "8"
	case model.GotoAdWeb:
		id = "9"
	case model.GotoBangumiRcmd:
		id = "10"
	case model.GotoLogin:
		id = "11"
	case model.GotoUpBangumi:
		id = "12"
	case model.GotoBanner:
		id = "13"
	case model.GotoAdWebS:
		id = "14"
	case model.GotoUpArticle:
		id = "15"
	case model.GotoConverge, model.GotoConvergeAi:
		id = "17"
	case model.GotoSpecial:
		id = "18"
	case model.GotoArticleS:
		id = "20"
	case model.GotoGameDownloadS:
		id = "21"
	case model.GotoShoppingS:
		id = "22"
	case model.GotoAudio:
		id = "23"
	case model.GotoPlayer:
		id = "24"
	case model.GotoSpecialS:
		id = "25"
	case model.GotoAdLarge:
		id = "26"
	case model.GotoPlayerLive:
		id = "27"
	case model.GotoSong:
		id = "28"
	case model.GotoLiveUpRcmd:
		id = "29"
	case model.GotoUpRcmdAv:
		id = "30"
	case model.GotoSubscribe:
		id = "31"
	case model.GotoChannelRcmd:
		id = "32"
	case model.GotoMoe:
		id = "33"
	case model.GotoPGC:
		id = "34"
	case model.GotoSearchSubscribe:
		id = "35"
	case model.GotoPicture:
		id = "36"
	case model.GotoInterest:
		id = "37"
	case model.GotoFollowMode:
		id = "38"
	case model.GotoPlayerBangumi:
		id = "39"
	case model.GotoSpecialB:
		id = "40"
	case model.GotoAdPlayer:
		id = "41"
	case model.GotoVipRenew:
		id = "42"
	case model.GotoAvConverge:
		id = "43"
	case model.GotoMultilayerConverge:
		id = "44"
	case model.GotoSpecialChannel:
		id = "45"
	case model.GotoTunnel:
		id = "46"
	case model.GotoInlineAv, model.GotoInlineAvV2:
		id = "47"
	case model.GotoAiStory:
		id = "48"
	case model.GotoInlinePGC:
		id = "49"
	case model.GotoInlineLive:
		id = "50"
	case model.GotoAdInlineGesture:
		id = "51"
	case model.GotoAdInline360:
		id = "52"
	case model.GotoAdInlineLive:
		id = "53"
	case model.GotoAdWebGif:
		id = "54"
	case model.GotoNewTunnel:
		id = "55"
	case model.GotoBigTunnel:
		id = "56"
	case model.GotoAdLive:
		id = "57"
	case model.GotoAdDynamic:
		id = "58"
	case model.GotoAdInlineChoose:
		id = "59"
	case model.GotoAdInlineChooseTeam:
		id = "60"
	case model.GotoAdInlineAv:
		id = "61"
	case model.GotoGame:
		id = "62"
	case model.GotoAdWebGifReservation:
		id = "63"
	case model.GotoAdPlayerReservation:
		id = "64"
	case model.GotoAdInline3D:
		id = "65"
	case model.GotoAdPgc:
		id = "66"
	case model.GotoAdInlinePgc:
		id = "67"
	case model.GotoAdInlineEggs:
		id = "68"
	case model.GotoAdInline3DV2:
		id = "69"
	default:
		id = "-1"
	}
	return
}

// bannerShowCase infoc banner case
func bannerShowCase(param infoc, showBanner bool) (bannerCase int) {
	const (
		_openEventCold = "cold"
		_flshNoAuto    = "0"
		_flushAuto     = "1"
		// banner展现时机
		_bannerDefault    = 0 // 默认
		_bannerNo         = 1 // 无banner
		_bannerCold       = 2 // 冷启带banner（open_event = cold）
		_bannerUpdate     = 3 // banner内容更新带banner（flush = 0&&上刷hash值和下刷hash值不一致）
		_bannerClientShow = 4 // 切换到后台30mins后带banner（flush = 1&&上一刷hash值为空）
		_bannerLoginShow  = 5 // 登录状态切换带banner（open_event != cold，login_event为1或2）
		_bannerAiShow     = 6 // 热启动间隔4小时（AI控制）
	)
	if !showBanner {
		bannerCase = _bannerNo
		return
	}
	if param.loginEvent != "0" && param.flush == _flshNoAuto && param.openEvent == _openEventCold {
		bannerCase = _bannerCold
		return
	}
	if param.clientBannerHash != param.bannerHash && param.flush == _flshNoAuto {
		bannerCase = _bannerUpdate
		return
	}
	if param.clientBannerHash == "" && param.flush == _flushAuto {
		bannerCase = _bannerClientShow
		return
	}
	if param.openEvent != _openEventCold && param.loginEvent != "0" {
		bannerCase = _bannerLoginShow
		return
	}
	bannerCase = _bannerAiShow
	return
}

func joinBannerInfo(in []*ai.BannerInfoItem) string {
	typeList := make([]string, 0, len(in))
	for _, item := range in {
		typeList = append(typeList, fmt.Sprintf("%s_%d_%s", item.Type, item.ID, item.InlineID))
	}
	return strings.Join(typeList, ",")
}

func joinBannerCreativeID(banners []*banner.Banner) string {
	creativeIDs := make([]int64, 0, len(banners))
	for _, v := range banners {
		creativeIDs = append(creativeIDs, int64(v.CreativeID))
	}
	return collection.JoinSliceInt(creativeIDs, ",")
}
