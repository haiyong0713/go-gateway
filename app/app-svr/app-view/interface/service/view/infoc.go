package view

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"
)

type viewInfoc struct {
	mid       string
	client    string
	build     string
	buvid     string
	ip        string
	api       string
	now       string
	aid       string
	err       string
	from      string
	trackid   string
	autoplay  string
	fromSpmid string
	spmid     string
	network   string
}

type relateInfoc struct {
	mid          string
	aid          string
	client       string
	buvid        string
	ip           string
	api          string
	now          string
	isRcmmnd     int8
	rls          []*view.Relate
	build        string
	returnCode   string
	userFeature  string
	from         string
	autoplay     int
	playParam    int
	fromTrackID  string
	pvFeature    json.RawMessage
	tabInfo      []*view.TabInfo
	pageType     string
	tabid        string
	fromSpmid    string
	spmid        string
	relatesInfoc *view.RelatesInfoc
	pageIndex    int64
}

// ViewInfoc view infoc
func (s *Service) ViewInfoc(mid int64, plat int, trackid, aid, ip, api, build, buvid, from string, now time.Time, err error, autoplay int, spmid, fromSpmid, network, isMelloi string) {
	if isMelloi != "" {
		return
	}
	s.infoc(viewInfoc{strconv.FormatInt(mid, 10), strconv.Itoa(plat), build, buvid, ip, api, strconv.FormatInt(now.Unix(), 10), aid, strconv.Itoa(ecode.Cause(err).Code()), from, trackid, strconv.Itoa(autoplay), fromSpmid, spmid, network})
}

// RelateInfoc Relate Infoc
func (s *Service) RelateInfoc(mid, aid int64, plat int, build, buvid, ip, api, returnCode, userFeature, from, tabid string,
	rls []*view.Relate, now time.Time, isRec int8, autoplay, playParam int, fromTrackID, pageType, fromSpmid, spmid string,
	pvFeature json.RawMessage, tabInfo []*view.TabInfo, isMelloi string, relatesInfoc *view.RelatesInfoc, pageIndex int64) {
	if isMelloi != "" {
		return
	}
	if relatesInfoc == nil {
		relatesInfoc = &view.RelatesInfoc{}
	}
	s.infoc(relateInfoc{strconv.FormatInt(mid, 10), strconv.FormatInt(aid, 10), strconv.Itoa(plat),
		buvid, ip, api, strconv.FormatInt(now.Unix(), 10), isRec, rls, build, returnCode,
		userFeature, from, autoplay, playParam, fromTrackID, pvFeature,
		tabInfo, pageType, tabid, fromSpmid, spmid, relatesInfoc, pageIndex})
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.inCh <- i:
	default:
		log.Warn("cacheproc chan full")
	}
}

// WriteViewInfoc Write View Infoc
func (s *Service) infocproc() {
	const (
		noItem = `{"section":{"id":"相关视频","pos":1,"from_item":"%s","from_tabid":"%s","items":[]}}`
	)
	var (
		msg1  = []byte(`{"section":{"id":"相关视频","pos":1,"from_item":"`)
		ms10  = []byte(`","from_tabid":"`)
		msg2  = []byte(`","items":[`)
		msg3  = []byte(`{"id":`)
		msg4  = []byte(`,"pos":`)
		msg5  = []byte(`,"goto":"`)
		msg6  = []byte(`","from":"`)
		msg7  = []byte(`","source":"`)
		msg8  = []byte(`","av_feature":`)
		msg9  = []byte(`,"dynamic_cover":`)
		msg10 = []byte(`,"type":1,"url":""},`)
		buf   bytes.Buffer
		list  string
	)
	for {
		i := <-s.inCh
		switch v := i.(type) {
		case viewInfoc:
			s.InfocV2(v)
		case relateInfoc:
			var trackID string
			if len(v.rls) == 0 {
				list = fmt.Sprintf(noItem, v.aid, v.tabid)
			} else {
				buf.Write(msg1)
				buf.WriteString(v.aid)
				buf.Write(ms10)
				buf.WriteString(v.tabid)
				buf.Write(msg2)
				for key, value := range v.rls {
					// trackid
					if value.TrackID != "" {
						trackID = value.TrackID
					}
					//list
					id, _ := strconv.ParseInt(value.Param, 10, 64)
					buf.Write(msg3)
					buf.WriteString(strconv.FormatInt(id, 10))
					buf.Write(msg4)
					buf.WriteString(strconv.Itoa(key + 1))
					buf.Write(msg5)
					buf.WriteString(value.Goto)
					buf.Write(msg6)
					buf.WriteString(value.From)
					buf.Write(msg7)
					buf.WriteString(value.Source)
					buf.Write(msg8)
					if value.AvFeature != nil {
						buf.Write(value.AvFeature)
					} else {
						buf.Write([]byte(`""`))
					}
					buf.Write(msg9)
					dynamicCover := "0"
					if value.CoverGif != "" && (value.Goto == model.GotoAv || value.Goto == model.GotoSpecial) {
						dynamicCover = model.DynamicCoverTp[value.Goto]
					}
					buf.WriteString(dynamicCover)
					buf.Write(msg10)
				}
				buf.Truncate(buf.Len() - 1)
				buf.WriteString(`]`)
				if v.tabInfo != nil && len(v.tabInfo) > 0 {
					tabStr, _ := json.Marshal(v.tabInfo)
					buf.WriteString(fmt.Sprintf(`,"tabs":%s`, tabStr))
				}
				buf.WriteString(`}}`)
				list = buf.String()
				buf.Reset()
			}
			s.relateInfocV2(v, list, trackID)
		case *view.ContinuousInfo:
			s.InfocV2(v)
		default:
			log.Warn("infocproc can't process the type")
		}
	}
}

func (s *Service) relateInfocV2(v relateInfoc, list string, trackID string) {
	payload := infocV2.NewLogStream(s.c.InfocRelateV2.LogID, v.ip, v.now, v.api, v.buvid, v.mid, v.client,
		v.pageType, list, v.pageIndex, v.isRcmmnd, trackID, v.build, v.returnCode, v.userFeature, v.from, v.autoplay,
		v.playParam, v.fromTrackID, string(v.pvFeature), v.fromSpmid, v.spmid, v.relatesInfoc.AdCode,
		v.relatesInfoc.AdNum, v.relatesInfoc.PKCode)
	if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
		log.Error("relateInfocV2 v(%+v) list(%+v) trackID(%+v) err(%+v)", v, list, trackID, err)
	}
}

func (s *Service) InfocV2(i interface{}) {
	switch v := i.(type) {
	case view.UserActInfoc:
		payload := infocV2.NewLogStream(s.c.InfocV2LogID.UserActLogID, v.Buvid, v.Build, v.Client, v.Ip, v.Uid, v.Aid, v.Mid, v.Sid, v.Refer, v.Url, v.From, v.ItemID, v.ItemType, v.Action, v.ActionID, v.Ua, v.Ts, v.Extra, v.IsRisk)
		if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
			log.Error("infocproc s.UserActInfoc err(%+v)", err)
		}
	case viewInfoc:
		payload := infocV2.NewLogStream(s.c.InfocViewV2.LogID, v.ip, v.now, v.api, v.buvid, v.mid, v.client, v.aid, "",
			v.err, v.from, v.build, v.trackid, v.autoplay, v.fromSpmid, v.spmid, v.network)
		if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
			log.Error("infocproc s.viewInfoc v(%+v) err(%+v)", v, err)
		}
	case *view.ContinuousInfo:
		payload := infocV2.NewLogStream(s.c.InfocV2LogID.ContinuousLogID, v.Ip, v.Now, v.Api, v.Buvid, v.Mid, v.Client, v.MobiApp, v.From, v.ShowList, v.IsRec, v.Build, v.ReturnCode, v.DeviceId, v.Network, v.TrackId, v.Spmid, v.FromSpmid, v.UserFeature, v.DisplayId, v.FromAv, v.FromTrackId)
		if err := s.infocV2Log.Info(context.Background(), payload); err != nil {
			log.Error("infocproc s.ContinuousInfo err(%+v)", err)
		}
	default:
		log.Warn("infocproc can't process the type")
	}
}
