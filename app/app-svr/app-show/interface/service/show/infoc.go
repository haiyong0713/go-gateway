package show

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/feed"
	rcmdm "go-gateway/app/app-svr/app-show/interface/model/recommend"
	"go-gateway/app/app-svr/app-show/interface/model/show"
)

type infoc struct {
	mid      string
	client   string
	buvid    string
	ip       string
	api      string
	now      string
	isRcmmnd string
	items    []*show.Item
}

type feedInfoc struct {
	mobiApp     string
	device      string
	build       string
	now         string
	loginEvent  string
	mid         string
	buvid       string
	page        string
	spmid       string
	feed        []*feed.Item
	url         string
	env         string
	returnCode  string
	trackid     string
	userfeature string
	flush       string
	isrec       string
	adBizData   string
}

type aggregationInfoc struct {
	mobiApp    string
	device     string
	build      string
	now        string
	loginEvent string
	mid        string
	buvid      string
	page       string
	spmid      string
	feed       interface{}
	url        string
}

// Infoc write data for Hadoop do analytics
func (s *Service) Infoc(mid int64, plat int8, buvid, ip, api string, items []*show.Item, now time.Time) {
	select {
	case s.logCh <- infoc{strconv.FormatInt(mid, 10), strconv.Itoa(int(plat)), buvid, ip, api, strconv.FormatInt(now.Unix(), 10), "1", items}:
	default:
		log.Warn("infoc log buffer is full")
	}
}

func (s *Service) InfocAggregation(hotWordID, mid int64, buvid, url string, aggSc []*rcmdm.CardList, now time.Time) {
	infoc := &aggregationInfoc{
		mobiApp:    "h5",
		build:      "0",
		now:        now.Format("2006-01-02 15:04:05"),
		loginEvent: "0",
		mid:        strconv.FormatInt(mid, 10),
		buvid:      buvid,
		page:       strconv.FormatInt(hotWordID, 10),
		url:        url,
	}
	var (
		itemsInfoc = []interface{}{}
	)
	for i, r := range aggSc {
		item := map[string]interface{}{
			"goto":         r.Goto,
			"param":        strconv.FormatInt(r.ID, 10),
			"uri":          model.FillURI(r.Goto, strconv.FormatInt(r.ID, 10), nil),
			"r_pos":        i + 1,
			"corner_mark":  r.CornerMark,
			"rcmd_content": r.Desc,
			"card_style":   2,
			"hot_aggre_id": r.HotwordID,
		}
		if r.CoverGif != "" && r.Goto == model.GotoAv {
			item["cover_type"] = "gif"
		} else {
			item["cover_type"] = "pic"
		}
		itemsInfoc = append(itemsInfoc, item)
	}
	infoc.feed = itemsInfoc
	s.infocfeed(infoc)
}

func (s *Service) infocfeed(i interface{}) {
	select {
	case s.logFeedCh <- i:
	default:
		log.Warn("infocfeed chan full")
	}
}

// writeInfoc
func (s *Service) infocproc() {
	const (
		// infoc format {"section":{"id":"热门推荐","pos":1,"items":[{"id":%s,"pos":%d,"type":1,"url":""}]}}
		noItem = `{"section":{"id":"热门推荐","pos":1,"items":[""]}}`
	)
	var (
		msg1 = []byte(`{"section":{"id":"热门推荐","pos":1,"items":[`)
		msg2 = []byte(`{"id":`)
		msg3 = []byte(`,"pos":`)
		msg4 = []byte(`,"type":1,"url":""},`)

		buf  bytes.Buffer
		list string
	)
	for {
		i := <-s.logCh
		if len(i.items) > 0 {
			buf.Write(msg1)
			for i, v := range i.items {
				if v.Goto != model.GotoAv {
					continue
				}
				buf.Write(msg2)
				buf.WriteString(v.Param)
				buf.Write(msg3)
				buf.WriteString(strconv.Itoa(i + 1))
				buf.Write(msg4)
			}
			buf.Truncate(buf.Len() - 1)
			buf.WriteString(`]}}`)
			list = buf.String()
			buf.Reset()
		} else {
			list = noItem
		}
		_ = s.reportV2(context.Background(), &i, list)
	}
}

func (s *Service) infocfeedproc() {
	const (
		noItem = `[]`
	)
	var (
		msg1  = []byte(`[`)
		msg2  = []byte(`{"goto":"`)
		msg3  = []byte(`","param":"`)
		msg4  = []byte(`","uri":"`)
		msg23 = []byte(`","av_feature":`)
		msg24 = []byte(`,"source":"`)
		msg5  = []byte(`","r_pos":`)
		msg6  = []byte(`,"from_type":"`)
		msg9  = []byte(`","corner_mark":`)
		msg10 = []byte(`,"rcmd_content":"`)
		msg18 = []byte(`","cover_type":"`)
		msg11 = []byte(`","card_style":`)
		msg12 = []byte(`,"items":[`)
		msg13 = []byte(`{"goto":"`)
		msg14 = []byte(`","param":"`)
		msg19 = []byte(`,"hot_aggre_id":`)
		msg20 = []byte(`,"channel_order":`)
		msg21 = []byte(`,"channel_name":"`)
		msg22 = []byte(`","channel_id":`)
		msg17 = []byte(`","pos":`)
		msg15 = []byte(`},`)
		msg16 = []byte(`]`)
		msg7  = []byte(`},`)
		msg8  = []byte(`]`)
		buf   bytes.Buffer
		list  string
	)
	for {
		i, ok := <-s.logFeedCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		switch l := i.(type) {
		case *feedInfoc:
			if f := l.feed; len(f) == 0 {
				list = noItem
			} else {
				buf.Write(msg1)
				for _, item := range f {
					buf.Write(msg2)
					buf.WriteString(item.Goto)
					buf.Write(msg3)
					buf.WriteString(item.Param)
					buf.Write(msg4)
					buf.WriteString(infocURIChange(item))
					buf.Write(msg23)
					buf.WriteString(string(item.AvFeature))
					buf.Write(msg24)
					buf.WriteString(item.Source)
					buf.Write(msg5)
					buf.WriteString(strconv.FormatInt(item.Idx, 10))
					buf.Write(msg6)
					buf.WriteString(item.FromType)
					buf.Write(msg9)
					buf.WriteString(strconv.Itoa(int(item.CornerMark)))
					buf.Write(msg10)
					buf.WriteString(item.RcmdContent)
					buf.Write(msg18)
					buf.WriteString(item.CoverType)
					buf.Write(msg11)
					buf.WriteString(strconv.Itoa(int(item.CardStyle)))
					buf.Write(msg19)
					buf.WriteString(strconv.FormatInt(item.HotAggreID, 10))
					buf.Write(msg20)
					buf.WriteString(item.ChannelOrder)
					buf.Write(msg21)
					buf.WriteString(item.ChannelName)
					buf.Write(msg22)
					buf.WriteString(item.ChannelID)
					if len(item.Item) > 0 {
						buf.Write(msg12)
						for pos, it := range item.Item {
							buf.Write(msg13)
							buf.WriteString(it.Goto)
							buf.Write(msg14)
							buf.WriteString(it.Param)
							buf.Write(msg17)
							buf.WriteString(strconv.Itoa(pos + 1))
							buf.Write(msg15)
						}
						buf.Truncate(buf.Len() - 1)
						buf.Write(msg16)
					}
					buf.Write(msg7)
				}
				buf.Truncate(buf.Len() - 1)
				buf.Write(msg8)
				list = buf.String()
				buf.Reset()
			}
			log.Info("showtab_infoc_index (mobiApp(%s), device(%s), build(%s), now(%s), loginEvent(%s), mid(%s), buvid(%s), list(%s), page(%s), spmid(%s), url(%s), env(%s), trackid(%s), isrec(%s), returnCode(%s), userfeature(%s), flush(%s), bizData(%s))",
				l.mobiApp, l.device, l.build, l.now, l.loginEvent, l.mid, l.buvid, list, l.page, l.spmid, l.url, l.env, l.trackid, l.isrec, l.returnCode, l.userfeature, l.flush, l.adBizData)
			payload := infocv2.NewLogStream(s.c.FeedTabInfocv2.LogID, l.mobiApp, l.device, l.build, l.now, l.loginEvent, l.mid, l.buvid, list, l.page, l.spmid, l.url, l.env, l.trackid, l.isrec, l.returnCode, l.userfeature, l.flush, l.adBizData)
			if err := s.feedTabInfocv2.Info(context.Background(), payload); err != nil {
				log.Error("Fail to report feedTabInfocv2, error=%+v", err)
			}
		case *aggregationInfoc:
			b, _ := json.Marshal(l.feed)
			log.Info("aggregation_infoc_index(%s,%s,%s,%s,%s,%s,%s,%s,%s)_list(%s)", l.mobiApp, l.device, l.build, l.now, l.loginEvent, l.mid, l.buvid, l.page, l.spmid, string(b))
			payload := infocv2.NewLogStream(s.c.FeedTabInfocv2.LogID, l.mobiApp, l.device, l.build, l.now, l.loginEvent, l.mid, l.buvid, string(b), l.page, l.spmid, l.url)
			if err := s.feedTabInfocv2.Info(context.Background(), payload); err != nil {
				log.Error("Fail to report feedTabInfocv2, error=%+v", err)
			}
		}
	}
}

func (s *Service) reportV2(c context.Context, data *infoc, showlist string) error {
	args := []log.D{
		log.KV("ip", data.ip),
		log.KV("time", data.now),
		log.KV("api", data.api),
		log.KV("buvid", data.buvid),
		log.KV("mid", data.mid),
		log.KV("client", data.client),
		log.KV("pagetype", "1"),
		log.KV("showlist", showlist),
		log.KV("displayid", ""),
		log.KV("is_rec", data.isRcmmnd),
	}
	payload := infocv2.NewLogStreamV(s.c.Infocv2.LogID, args...)
	if err := s.infocv2.Info(c, payload); err != nil {
		log.Error("Fail to report ShowInfoc, data=%+v error=%+v", data, err)
		return err
	}
	return nil
}

func infocURIChange(item *feed.Item) (uri string) {
	switch item.Goto {
	case model.GotoAv:
		uri = model.FillURI(item.Goto, item.Param, nil)
	default:
		uri = item.URI
	}
	return
}
