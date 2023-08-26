package view

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/app-intl/interface/model/view"
)

// viewInfoc struct
type viewInfoc struct {
	mid      string
	client   string
	build    string
	buvid    string
	ip       string
	api      string
	now      string
	aid      string
	err      string
	from     string
	trackid  string
	autoplay string
}

// relateInfoc struct
type relateInfoc struct {
	mid         string
	aid         string
	client      string
	buvid       string
	ip          string
	api         string
	now         string
	isRcmmnd    int8
	rls         []*view.Relate
	trackid     string
	build       string
	returnCode  string
	userFeature string
	from        string
}

// ViewInfoc view infoc
func (s *Service) ViewInfoc(mid int64, plat int, trackid, aid, ip, api, build, buvid, from string, now time.Time, err error, autoplay int) {
	s.infoc(viewInfoc{strconv.FormatInt(mid, 10), strconv.Itoa(plat), build, buvid, ip, api, strconv.FormatInt(now.Unix(), 10), aid, strconv.Itoa(ecode.Cause(err).Code()), from, trackid, strconv.Itoa(autoplay)})
}

// RelateInfoc Relate Infoc
func (s *Service) RelateInfoc(mid, aid int64, plat int, trackid, build, buvid, ip, api, returnCode, userFeature, from string, rls []*view.Relate, now time.Time, isRec int8) {
	s.infoc(relateInfoc{strconv.FormatInt(mid, 10), strconv.FormatInt(aid, 10), strconv.Itoa(plat), buvid, ip, api, strconv.FormatInt(now.Unix(), 10), isRec, rls, trackid, build, returnCode, userFeature, from})
}

// infoc is.
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
		noItem = `{"section":{"id":"相关视频","pos":1,"from_item":"%s","items":[]}}`
	)
	var (
		msg1 = []byte(`{"section":{"id":"相关视频","pos":1,"from_item":"`)
		msg2 = []byte(`","items":[`)
		msg3 = []byte(`{"id":`)
		msg4 = []byte(`,"pos":`)
		msg5 = []byte(`,"goto":"`)
		msg6 = []byte(`","from":"`)
		msg7 = []byte(`","source":"`)
		msg8 = []byte(`","av_feature":`)
		msg9 = []byte(`,"type":1,"url":""},`)
		buf  bytes.Buffer
		list string
	)
	for {
		i := <-s.inCh
		switch v := i.(type) {
		case viewInfoc:
			_ = s.infocV2SendViewData(context.Background(), &v)
		case relateInfoc:
			if len(v.rls) > 0 {
				buf.Write(msg1)
				buf.WriteString(v.aid)
				buf.Write(msg2)
				for key, value := range v.rls {
					buf.Write(msg3)
					buf.WriteString(strconv.Itoa(int(value.Aid)))
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
				}
				buf.Truncate(buf.Len() - 1)
				buf.WriteString(`]}}`)
				list = buf.String()
				buf.Reset()
			} else {
				list = fmt.Sprintf(noItem, v.aid)
			}
			_ = s.reportV2(context.Background(), &v, list)
		default:
			log.Warn("infocproc can't process the type")
		}
	}
}

// nolint:unparam
func (s *Service) infocV2SendViewData(ctx context.Context, v *viewInfoc) error {
	payload := infocv2.NewLogStreamV(s.c.ViewInfocv2.LogID, log.String(v.ip), log.String(v.now), log.String(v.api),
		log.String(v.buvid), log.String(v.mid), log.String(v.client), log.String(v.aid), log.String(""),
		log.String(v.err), log.String(v.from), log.String(v.build), log.String(v.trackid), log.String(v.autoplay))
	if err := s.viewInfocv2.Info(ctx, payload); err != nil {
		log.Warn("infocV2SendViewData() s.viewInfocv2.Info() viewInfoc(%+v) error(%v)", v, err)
	}
	return nil
}

func (s *Service) reportV2(c context.Context, data *relateInfoc, showlist string) error {
	args := []log.D{
		log.KV("ip", data.ip),
		log.KV("time", data.now),
		log.KV("api", data.api),
		log.KV("buvid", data.buvid),
		log.KV("mid", data.mid),
		log.KV("client", data.client),
		log.KV("pagetype", "2"),
		log.KV("showlist", showlist),
		log.KV("displayid", ""),
		log.KV("is_rec", data.isRcmmnd),
		log.KV("trackid", data.trackid),
		log.KV("build", data.build),
		log.KV("return_code", data.returnCode),
		log.KV("user_feature", data.userFeature),
		log.KV("source_page", data.from),
	}
	payload := infocv2.NewLogStreamV(s.c.RelateInfocv2.LogID, args...)
	if err := s.relateInfocv2.Info(c, payload); err != nil {
		log.Error("Fail to report RelateFeed, data=%+v showlist=%s error=%+v", data, showlist, err)
		return err
	}
	return nil

}
