package feed

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-intl/interface/model"
)

// infoc struct
type infoc struct {
	typ         string
	mid         string
	client      string
	build       string
	buvid       string
	ip          string
	style       string
	api         string
	now         string
	isRcmd      string
	pull        string
	userFeature json.RawMessage
	code        string
	items       []*ai.Item
	lp          bool
	zoneID      string
	adResponse  string
	deviceID    string
	network     string
	newUser     string
	flush       string
	autoPlay    string
	deviceType  string
	Slocale     string
	Clocale     string
	Timezone    string
	Lang        string
	SimCode     string
}

// FeedShowInfo new intl struct
type FeedShowInfo struct {
	IP             string `json:"ip"`
	Slocale        string `json:"s_locale"`
	Clocale        string `json:"c_locale"`
	Timezone       string `json:"timezone"`
	Lang           string `json:"lang"`
	Time           string `json:"time"`
	API            string `json:"api"`
	Buvid          string `json:"buvid"`
	Mid            string `json:"mid"`
	Client         string `json:"client"`
	PageType       int    `json:"pagetype"`
	ShowList       string `json:"showlist"`
	DisplayID      string `json:"display_id"`
	IsRec          int    `json:"is_rec"`
	Build          string `json:"build"`
	ReturnCode     string `json:"return_code"`
	UserFeature    string `json:"user_feature"`
	ZoneID         string `json:"zone_id"`
	ADResponse     string `json:"ad_response"`
	DeviceID       string `json:"device_id"`
	Network        string `json:"network"`
	NewUser        string `json:"new_user"`
	Flush          string `json:"flush"`
	AutoplayCard   string `json:"autoplay_card"`
	TrackID        string `json:"track_id"`
	DeviceType     int    `json:"device_type"`
	BannerShowCase int    `json:"banner_show_case"`
	BannerHash     string `json:"banner_hash"`
	LoginEvent     int    `json:"login_event"`
	BizErrCode     string `json:"biz_err_code"`
	BizAdvNum      string `json:"biz_adv_num"`
	BizPkCode      string `json:"biz_pk_code"`
	DisplayType    int    `json:"display_type"`
	SimCode        string `json:"sim_code"`
}

// IndexInfoc is.
func (s *Service) IndexInfoc(c context.Context, mid int64, plat int8, build int, buvid, api string, userFeature json.RawMessage, style, code int, items []*ai.Item, isRcmd, pull, newUser bool, now time.Time, adResponse, deviceID, network string, flush int, autoPlay string, deviceType int, sLocal, cLocal, timeZone, lang, simCode string) {
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
	info, err := s.loc.Info(c, ip)
	if err != nil {
		log.Warn(" s.loc.Info(%v) error(%v)", ip, err)
	}
	if info != nil {
		zoneID = info.ZoneId
	}
	s.infoc(infoc{"综合推荐", strconv.FormatInt(mid, 10), strconv.Itoa(int(plat)), strconv.Itoa(build), buvid, ip, strconv.Itoa(style), api, strconv.FormatInt(now.Unix(), 10), isRc, isPull, userFeature, strconv.Itoa(code), items, false, strconv.FormatInt(zoneID, 10), adResponse, deviceID, network, isNewUser, strconv.Itoa(flush), autoPlay, strconv.Itoa(deviceType), sLocal, cLocal, timeZone, lang, simCode})
}

// infoc is.
func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

// infocproc is.
func (s *Service) infocproc() {
	const (
		// infoc format {"section":{"id":"%s推荐","pos":1,"style":%d,"items":[{"id":%s,"pos":%d,"type":1,"url":""}]}}
		noItem1 = `{"section":{"id":"`
		noItem2 = `{","pos":1,"style":`
		noItem3 = `,"items":[]}}`
	)
	var (
		msg1    = []byte(`{"section":{"id":"`)
		msg2    = []byte(`","pos":1,"style":`)
		msg3    = []byte(`,"items":[`)
		msg4    = []byte(`{"id":`)
		msg5    = []byte(`,"pos":`)
		msg6    = []byte(`,"type":`)
		msg7    = []byte(`,"source":"`)
		msg8    = []byte(`","tid":`)
		msg9    = []byte(`,"av_feature":`)
		msg10   = []byte(`,"url":"`)
		msg11   = []byte(`","rcmd_reason":"`)
		msg12   = []byte(`"},`)
		msg13   = []byte(`]}}`)
		buf     bytes.Buffer
		list    string
		trackID string
	)
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case infoc:
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
					buf.WriteString(strconv.FormatInt(v.ID, 10))
					buf.Write(msg5)
					buf.WriteString(strconv.Itoa(i + 1))
					buf.Write(msg6)
					buf.WriteString(gotoMapID(v.Goto))
					buf.Write(msg7)
					buf.WriteString(v.Source)
					buf.Write(msg8)
					buf.WriteString(strconv.FormatInt(v.Tid, 10))
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
					buf.Write(msg12)
				}
				buf.Truncate(buf.Len() - 1)
				buf.Write(msg13)
				list = buf.String()
				buf.Reset()
			} else {
				list = noItem1 + l.typ + noItem2 + l.style + noItem3
			}
			_ = s.infocV2SendFeedData(context.Background(), &l, list, trackID)
		}
	}
}

// nolint:unparam
func (s *Service) infocV2SendFeedData(ctx context.Context, v *infoc, list, trackID string) error {
	payload := infocv2.NewLogStreamV(s.c.FeedInfocv2.LogID, log.String(v.ip), log.String(v.now), log.String(v.api),
		log.String(v.buvid), log.String(v.mid), log.String(v.client), log.String(v.pull), log.String(list),
		log.String(""), log.String(v.isRcmd), log.String(v.build), log.String(v.code), log.String(string(v.userFeature)),
		log.String(v.zoneID), log.String(v.adResponse), log.String(v.deviceID), log.String(v.network), log.String(v.newUser),
		log.String(v.flush), log.String(v.autoPlay), log.String(trackID), log.String(v.deviceType))
	if err := s.infocV2Client.Info(ctx, payload); err != nil {
		log.Warn("infocV2SendFeedData() s.infocV2Client.Info() infoc(%+v) list(%s), trackId(%s), error(%v)", v, list, trackID, err)
	}
	return nil
}

// gotoMapID is.
func gotoMapID(gt string) (id string) {
	if gt == model.GotoAv {
		id = "1"
	} else if gt == model.GotoBangumi {
		id = "2"
	} else if gt == model.GotoLive {
		id = "3"
	} else if gt == model.GotoRank {
		id = "6"
	} else if gt == model.GotoAdAv {
		id = "8"
	} else if gt == model.GotoAdWeb {
		id = "9"
	} else if gt == model.GotoBangumiRcmd {
		id = "10"
	} else if gt == model.GotoLogin {
		id = "11"
	} else if gt == model.GotoBanner {
		id = "13"
	} else if gt == model.GotoAdWebS {
		id = "14"
	} else if gt == model.GotoConverge {
		id = "17"
	} else if gt == model.GotoSpecial {
		id = "18"
	} else if gt == model.GotoArticleS {
		id = "20"
	} else if gt == model.GotoGameDownloadS {
		id = "21"
	} else if gt == model.GotoShoppingS {
		id = "22"
	} else if gt == model.GotoAudio {
		id = "23"
	} else if gt == model.GotoPlayer {
		id = "24"
	} else if gt == model.GotoSpecialS {
		id = "25"
	} else if gt == model.GotoAdLarge {
		id = "26"
	} else if gt == model.GotoPlayerLive {
		id = "27"
	} else if gt == model.GotoSong {
		id = "28"
	} else if gt == model.GotoLiveUpRcmd {
		id = "29"
	} else if gt == model.GotoUpRcmdAv {
		id = "30"
	} else if gt == model.GotoSubscribe {
		id = "31"
	} else if gt == model.GotoChannelRcmd {
		id = "32"
	} else if gt == model.GotoMoe {
		id = "33"
	} else if gt == model.GotoPGC {
		id = "34"
	} else if gt == model.GotoSearchSubscribe {
		id = "35"
	} else {
		id = "-1"
	}
	return
}
