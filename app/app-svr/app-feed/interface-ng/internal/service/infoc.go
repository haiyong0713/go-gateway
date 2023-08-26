package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	binfoc "go-common/library/log/infoc"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	ngmdl "go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"

	"github.com/pkg/errors"
)

//go:generate easyjson -all infoc.go

//easyjson:json
type ShowList struct {
	Section ShowListItem `json:"section"`
}

//easyjson:json
type ShowListItem struct {
	ID    string `json:"id"`
	Pos   int8   `json:"pos"`
	Style string `json:"style"`
	Items []Item `json:"items"`
}

//easyjson:json
type Item struct {
	ID              int64  `json:"id"`
	Pos             int8   `json:"pos"`
	Type            int8   `json:"type"`
	Goto            string `json:"goto"`
	Source          string `json:"source"`
	Tid             int64  `json:"tid"`
	CustomizedTitle string `json:"customized_title,omitempty"`
	CustomizedCover string `json:"customized_cover,omitempty"`
	IsGifCover      int8   `json:"is_gifcover,omitempty"`
	DynamicCover    int32  `json:"dynamic_cover,omitempty"`
	AvFeature       string `json:"av_feature"`
	URL             string `json:"url"`
	RcmdReason      string `json:"rcmd_reason"`
	LivePendent     string `json:"live_pendent,omitempty"`
	Hash            string `json:"hash,omitempty"`
	IsAdLoc         bool   `json:"is_ad_loc,omitempty"`
	ResourceID      int64  `json:"resource_id,omitempty"`
	SourceID        int64  `json:"source_id,omitempty"`
	CreativeID      int64  `json:"creative_id,omitempty"`
}

type showInfoc struct {
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
	materials        map[int64]interface{}
	isGifCover       map[int64]int
	bannerHash       string
	clientBannerHash string
	loginEvent       string
	adError          string
	adCount          string
	adPkCode         string
	openEvent        string
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

func (s *Service) IndexInfoc(ctx context.Context, fanoutResult *feedcard.FanoutResult, aiReq *ngmdl.AiReq, cardParam *ngmdl.ConstructCardParam) {
	if len(cardParam.AIResponse.Items) == 0 {
		return
	}
	isRc := "0"
	isPull := "0"
	isNewUser := "0"
	zoneID := int64(0)
	if cardParam.AIResponse.IsRcmd {
		isRc = "1"
	}
	if aiReq.IndexNgReq.FeedParam.Pull {
		isPull = "1"
	}
	if cardParam.AIResponse.NewUser {
		isNewUser = "1"
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	if aiReq.Zone != nil {
		zoneID = aiReq.Zone.ZoneId
	}
	addPosStr := "None"
	if cardParam.AIResponse.BizData != nil && cardParam.AIResponse.BizData.BizResult != "" {
		addPosStr = cardParam.AIResponse.BizData.BizResult
	}
	s.infoc(showInfoc{
		"综合推荐",
		strconv.FormatInt(cardParam.IndexNgReq.Mid, 10),
		strconv.Itoa(int(cardParam.IndexNgReq.Device.Plat())),
		strconv.FormatInt(cardParam.IndexNgReq.Device.Build(), 10),
		cardParam.IndexNgReq.Device.Buvid(),
		ip,
		strconv.Itoa(cardParam.IndexNgReq.Style),
		"/x/feed/index", strconv.FormatInt(aiReq.Now.Unix(), 10),
		isRc,
		isPull,
		cardParam.AIResponse.UserFeature,
		strconv.Itoa(cardParam.AIResponse.RespCode),
		cardParam.AIResponse.Items,
		strconv.FormatInt(zoneID, 10),
		"",
		cardParam.IndexNgReq.Device.Device(),
		cardParam.IndexNgReq.FeedParam.Network,
		isNewUser,
		strconv.Itoa(cardParam.IndexNgReq.FeedParam.Flush),
		aiReq.AutoPlay,
		strconv.Itoa(cardParam.IndexNgReq.FeedParam.DeviceType),
		constructMaterials(fanoutResult),
		constructIsGifCover(cardParam.AIResponse.Items),
		fanoutResult.Banner.Version,
		cardParam.IndexNgReq.FeedParam.BannerHash,
		strconv.Itoa(cardParam.IndexNgReq.FeedParam.LoginEvent),
		fmt.Sprintf("code(%d),message(%v)", constructAdCode(cardParam.IndexNgReq.FeedParam, cardParam.AIResponse), constructAdErr(cardParam.IndexNgReq.FeedParam)),
		addPosStr,
		constructAdPKCode(cardParam.AIResponse),
		cardParam.IndexNgReq.FeedParam.OpenEvent})
}

func constructAdPKCode(param *feed.AIResponse) string {

	out := []string{}
	for _, v := range param.Items {
		if v.AdPKCode() != "" {
			out = append(out, v.AdPKCode())
		}
	}
	if param.BizData == nil {
		return strings.Join(out, ",")
	}
	for _, v := range param.BizData.AdDiscarded {
		if v == nil {
			continue
		}
		missAiAd := cm.ConstructAdInfos(v)
		for _, miss := range missAiAd {
			switch miss.DiscardReason {
			case "gif":
				out = append(out, _adPkGifCard)
			case "banner":
				out = append(out, _adPkBigCard)
			}
		}
	}
	return strings.Join(out, ",")
}

const (
	_recsysMode       = 78050
	_recsysModeMsg    = "关注模式"
	_teenagersMode    = 78051
	_teenagersModeMsg = "青少年模式"
	_lessonsMode      = 78052
	_lessonsModeMsg   = "课堂模式"
)

func constructAdCode(param feed.IndexParam, aiResponse *feed.AIResponse) int {
	if param.RecsysMode == 1 {
		return _recsysMode
	}
	if param.TeenagersMode == 1 {
		return _teenagersMode
	}
	if param.LessonsMode == 1 {
		return _lessonsMode
	}
	return aiResponse.AdCode
}

func constructAdErr(param feed.IndexParam) error {
	if param.RecsysMode == 1 {
		return errors.New(_recsysModeMsg)
	}
	if param.TeenagersMode == 1 {
		return errors.New(_teenagersModeMsg)
	}
	if param.LessonsMode == 1 {
		return errors.New(_lessonsModeMsg)
	}
	return nil
}

func constructMaterials(param *feedcard.FanoutResult) map[int64]interface{} {
	out := make(map[int64]interface{}, len(param.Live.Room)+len(param.Live.InlineRoom))
	for id, room := range param.Live.Room {
		out[id] = room
	}
	for id, inlineRoom := range param.Live.InlineRoom {
		out[id] = inlineRoom
	}
	return out
}

func constructIsGifCover(items []*ai.Item) map[int64]int {
	out := make(map[int64]int)
	for _, item := range items {
		if item.CoverGif == "" {
			continue
		}
		out[item.ID] = 0
		if item.AllowGIF() {
			out[item.ID] = 1
		}
	}
	return out
}

//nolint:gocognit
func (s *Service) infocproc() {
	showInf := binfoc.New(s.customConfig.ShowInfoc)
	list := ""
	trackID := ""
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case showInfoc:
			var showBanner bool
			showList := ShowList{
				Section: ShowListItem{
					ID:    l.typ,
					Pos:   1,
					Style: l.style,
				},
			}
			if len(l.items) > 0 {
				//buf.WriteString(fmt.Sprintf(`{"section":{"id":"%s","pos":1,"style":%s,"items":[`, l.typ, l.style))
				items := make([]Item, 0, len(l.items))
				for i, v := range l.items {
					if v == nil {
						continue
					}
					if v.TrackID != "" {
						trackID = v.TrackID
					}
					itemID := v.ID
					switch v.Goto {
					case model.GotoConvergeAi:
						//nolint:gomnd
						itemID = v.ID + 100000
					case model.GotoBanner:
						showBanner = true
					}
					item := Item{
						ID:     itemID,
						Pos:    int8(i + 1),
						Type:   gotoMapID(v.Goto),
						Source: v.Goto,
						Tid:    v.Tid,
						URL:    "",
					}
					//buf.WriteString(fmt.Sprintf(`{"id":%s,"pos":%s,"type":%s,"goto":"%s","source":"%s","tid":%s`,
					//	strconv.FormatInt(itemID, 10), strconv.Itoa(i+1), gotoMapID(v.Goto), v.Goto, v.Source,
					//	strconv.FormatInt(v.Tid, 10)))
					if v.CustomizedTitle != "" {
						item.CustomizedTitle = v.CustomizedTitle
					}
					if v.CustomizedCover != "" {
						item.CustomizedCover = v.CustomizedCover
					}
					if isGif, ok := l.isGifCover[v.ID]; ok {
						item.IsGifCover = int8(isGif)
					}
					if v.DynamicCoverInfoc() > 0 {
						item.DynamicCover = v.DynamicCoverInfoc()
					}
					//buf.WriteString(`,"av_feature":`)
					if v.AvFeature != nil {
						item.AvFeature = string(v.AvFeature)
					}
					//buf.WriteString(`,"url":"","rcmd_reason":"`)
					if v.RcmdReason != nil {
						item.RcmdReason = v.RcmdReason.Content
					}
					if main, ok := l.materials[v.ID]; ok {
						//nolint:gosimple
						switch main.(type) {
						case *live.Room:
							if r := main.(*live.Room); r != nil {
								item.LivePendent = r.PendentRu
							}
						}
					}
					switch v.Goto {
					case model.GotoBanner:
						item.Hash = l.bannerHash
					}
					if v.Ad != nil {
						item.IsAdLoc = v.Ad.IsAdLoc
						item.ResourceID = v.Ad.Resource
						item.SourceID = int64(v.Ad.Source)
						item.CreativeID = v.Ad.CreativeID
					}
					items = append(items, item)
				}
				showList.Section.Items = items
			}
			bt, err := showList.MarshalJSON()
			if err != nil {
				log.Error("Failed to marshal json: %+v", err)
			}
			list = string(bt)
			bannerCase := strconv.Itoa(bannerShowCase(l, showBanner))
			//nolint:errcheck
			showInf.Info(l.ip, l.now, l.api, l.buvid, l.mid, l.client, l.pull, list, "", l.isRcmd, l.build, l.code,
				string(l.userFeature), l.zoneID, l.adResponse, l.deviceID, l.network, l.newUser, l.flush, l.autoPlay,
				trackID, l.deviceType, bannerCase, l.clientBannerHash, l.loginEvent, l.adError, l.adCount, l.adPkCode,
				l.style)
		}
	}
}

func gotoMapID(gt string) (id int8) {
	switch gt {
	case model.GotoAv:
		id = 1
	case model.GotoBangumi:
		id = 2
	case model.GotoLive:
		id = 3
	case model.GotoRank:
		id = 6
	case model.GotoAdAv:
		id = 8
	case model.GotoAdWeb:
		id = 9
	case model.GotoBangumiRcmd:
		id = 10
	case model.GotoLogin:
		id = 11
	case model.GotoUpBangumi:
		id = 12
	case model.GotoBanner:
		id = 13
	case model.GotoAdWebS:
		id = 14
	case model.GotoUpArticle:
		id = 15
	case model.GotoConverge, model.GotoConvergeAi:
		id = 17
	case model.GotoSpecial:
		id = 18
	case model.GotoArticleS:
		id = 20
	case model.GotoGameDownloadS:
		id = 21
	case model.GotoShoppingS:
		id = 22
	case model.GotoAudio:
		id = 23
	case model.GotoPlayer:
		id = 24
	case model.GotoSpecialS:
		id = 25
	case model.GotoAdLarge:
		id = 26
	case model.GotoPlayerLive:
		id = 27
	case model.GotoSong:
		id = 28
	case model.GotoLiveUpRcmd:
		id = 29
	case model.GotoUpRcmdAv:
		id = 30
	case model.GotoSubscribe:
		id = 31
	case model.GotoChannelRcmd:
		id = 32
	case model.GotoMoe:
		id = 33
	case model.GotoPGC:
		id = 34
	case model.GotoSearchSubscribe:
		id = 35
	case model.GotoPicture:
		id = 36
	case model.GotoInterest:
		id = 37
	case model.GotoFollowMode:
		id = 38
	case model.GotoPlayerBangumi:
		id = 39
	case model.GotoSpecialB:
		id = 40
	case model.GotoAdPlayer:
		id = 41
	case model.GotoVipRenew:
		id = 42
	case model.GotoAvConverge:
		id = 43
	case model.GotoMultilayerConverge:
		id = 44
	case model.GotoSpecialChannel:
		id = 45
	case model.GotoTunnel:
		id = 46
	case model.GotoInlineAv, model.GotoInlineAvV2:
		id = 47
	case model.GotoAiStory:
		id = 48
	case model.GotoInlinePGC:
		id = 49
	case model.GotoInlineLive:
		id = 50
	case model.GotoAdInlineGesture:
		id = 51
	case model.GotoAdInline360:
		id = 52
	case model.GotoAdInlineLive:
		id = 53
	case model.GotoAdWebGif:
		id = 54
	case model.GotoNewTunnel:
		id = 55
	case model.GotoBigTunnel:
		id = 56
	case model.GotoAdPlayerReservation:
		id = 57
	case model.GotoAdWebGifReservation:
		id = 58
	default:
		id = -1
	}
	return
}

// bannerShowCase infoc banner case
func bannerShowCase(param showInfoc, showBanner bool) (bannerCase int) {
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
