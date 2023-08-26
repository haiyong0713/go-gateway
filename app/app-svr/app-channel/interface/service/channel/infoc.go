package channel

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/log/infoc"
	"go-gateway/app/app-svr/app-channel/interface/model/feed"
)

type feedInfoc struct {
	mobiApp     string
	device      string
	build       string
	now         string
	pull        string
	loginEvent  string
	channelID   string
	channelName string
	mid         string
	buvid       string
	displayID   string
	feed        *feed.Show
	Items       []*feed.Item
	isRec       string
	topChannel  string
	ServerCode  string
	FromSpmid   string
	FromPage    string
}

type channelOperation struct {
	mobiApp   string
	device    string
	build     string
	now       string
	channelID string
	operation string
	mid       string
	from      string
	fromSpmid string
	modelID   string
}

type channelSquare struct {
	mobiApp string
	device  string
	build   string
	now     string
	mid     string
	buvid   string
	feed    []*feed.ChannelInfo
}

// OperationInfoc channel operation infoc
func (s *Service) OperationInfoc(mobiApp, device, operation, fromSpmid, modelID string, build, from int, channelID, mid int64, now time.Time) {
	infoc := &channelOperation{
		mobiApp:   mobiApp,
		device:    device,
		build:     strconv.Itoa(build),
		now:       now.Format("2006-01-02 15:04:05"),
		channelID: strconv.FormatInt(channelID, 10),
		operation: operation,
		mid:       strconv.FormatInt(mid, 10),
		from:      strconv.Itoa(from),
		fromSpmid: fromSpmid,
		modelID:   modelID,
	}
	s.infoc(infoc)
}

// SquareInfoc channel square infoc
func (s *Service) SquareInfoc(mobiApp, device, buvid string, build int, mid int64, now time.Time, items []*feed.ChannelInfo) {
	infoc := &channelSquare{
		mobiApp: mobiApp,
		device:  device,
		build:   strconv.Itoa(build),
		buvid:   buvid,
		now:     now.Format("2006-01-02 15:04:05"),
		mid:     strconv.FormatInt(mid, 10),
		feed:    items,
	}
	s.infoc(infoc)
}

func (s *Service) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}

func (s *Service) infocproc() {
	const (
		noItem = `[]`
	)
	var (
		msg1       = []byte(`[`)
		msg2       = []byte(`{"goto":"`)
		msg3       = []byte(`","param":"`)
		msg4       = []byte(`","uri":"`)
		msg5       = []byte(`","r_pos":`)
		msg6       = []byte(`,"from_type":"`)
		msg7       = []byte(`"},`)
		msg8       = []byte(`]`)
		buf        bytes.Buffer
		list       string
		feedInf2   = infoc.New(s.c.FeedInfoc2)
		channelnf2 = infoc.New(s.c.ChannelInfoc2)
		squarelnf2 = infoc.New(s.c.SquareInfoc2)
	)
	for {
		i, ok := <-s.logCh
		if !ok {
			log.Warn("infoc proc exit")
			return
		}
		var pos int
		switch l := i.(type) {
		case *feedInfoc:
			if l.feed != nil {
				if f := l.feed; len(f.Feed) == 0 && f.Topic == nil {
					list = noItem
				} else {
					buf.Write(msg1)
					if t := f.Topic; t != nil {
						buf.Write(msg2)
						buf.WriteString(t.Goto)
						buf.Write(msg3)
						buf.WriteString(t.Param)
						buf.Write(msg4)
						buf.WriteString(t.URI)
						buf.Write(msg5)
						pos = t.Pos
						buf.WriteString(strconv.Itoa(pos))
						buf.Write(msg6)
						buf.WriteString(t.FromType)
						buf.Write(msg7)
					}
					if items := f.Feed; len(items) == 0 {
						buf.Truncate(buf.Len() - 1)
					} else {
						for _, item := range items {
							buf.Write(msg2)
							buf.WriteString(item.Goto)
							buf.Write(msg3)
							buf.WriteString(item.Param)
							buf.Write(msg4)
							buf.WriteString(item.URI)
							buf.Write(msg5)
							if item.Pos == 0 {
								pos++
							} else {
								pos = item.Pos
							}
							buf.WriteString(strconv.Itoa(pos))
							buf.Write(msg6)
							buf.WriteString(item.FromType)
							buf.Write(msg7)
						}
						buf.Truncate(buf.Len() - 1)
					}
					buf.Write(msg8)
					list = buf.String()
					buf.Reset()
				}
			} else if items := l.Items; len(items) > 0 {
				buf.Write(msg1)
				for _, item := range items {
					buf.Write(msg2)
					buf.WriteString(item.Goto)
					buf.Write(msg3)
					buf.WriteString(item.Param)
					buf.Write(msg4)
					buf.WriteString(item.URI)
					buf.Write(msg5)
					if item.Pos == 0 {
						pos++
					} else {
						pos = item.Pos
					}
					buf.WriteString(strconv.Itoa(pos))
					buf.Write(msg6)
					buf.WriteString(item.FromType)
					buf.Write(msg7)
				}
				buf.Truncate(buf.Len() - 1)
				buf.Write(msg8)
				list = buf.String()
				buf.Reset()
			} else if len(l.Items) == 0 {
				list = noItem
			}
			log.Info("channel_infoc_index(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)_list(%s)", l.mobiApp, l.device, l.build, l.now, l.pull, l.loginEvent,
				l.channelID, l.channelName, l.mid, l.buvid, l.displayID, l.ServerCode, l.FromPage, l.FromSpmid, list)
			_ = feedInf2.Info(l.mobiApp, l.device, l.build, l.now, l.pull, l.loginEvent, l.channelID, l.channelName, l.mid, l.buvid, l.displayID,
				list, "9", l.isRec, l.topChannel, l.ServerCode, l.FromPage, l.FromSpmid)
		case *channelOperation:
			log.Info("channel_infoc_operation_(%s,%s,%s,%s,%s,%s,%s,%s)", l.mobiApp, l.device, l.build, l.now, l.channelID, l.operation, l.mid, l.from)
			_ = channelnf2.Info(l.mobiApp, l.device, l.build, l.now, l.channelID, l.operation, l.mid, l.from)
		case *channelSquare:
			b, _ := json.Marshal(&l.feed)
			log.Info("channel_infoc_square_(%s,%s,%s,%s,%s,%s,%s)", l.mobiApp, l.device, l.build, l.now, l.mid, l.buvid, string(b))
			_ = squarelnf2.Info(l.mobiApp, l.device, l.build, l.now, l.mid, l.buvid, string(b))
		}
	}
}
