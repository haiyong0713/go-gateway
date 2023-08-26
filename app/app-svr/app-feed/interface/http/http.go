package http

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	abtest "go-common/component/tinker/middleware/http"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/feed/thread_sampler"
	"go-gateway/app/app-svr/app-feed/interface/model/sets"
	"go-gateway/app/app-svr/app-feed/interface/service/external"
	"go-gateway/app/app-svr/app-feed/interface/service/feed"
	"go-gateway/app/app-svr/app-feed/interface/service/region"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"github.com/google/uuid"
)

var (
	// depend service
	authSvc *auth.Auth
	// self service
	regionSvc   *region.Service
	feedSvc     *feed.Service
	externalSvc *external.Service
	//nolint:unused
	cfg *conf.Config
	// sampler
	sampler *thread_sampler.ThreadSampler
	// feature service
	featureSvc *feature.Feature
)

// Init is
func Init(c *conf.Config, ic infoc.Infoc) {
	initService(c, ic)
	// init external router
	engineOut := bm.DefaultServer(c.BM.Outer)
	outerRouter(engineOut)
	// init outer server
	if err := engineOut.Start(); err != nil {
		log.Error("engineOut.Start() error(%v)", err)
		panic(err)
	}
	feed.InitDatabus(c)
}

// initService init services.
func initService(c *conf.Config, ic infoc.Infoc) {
	authSvc = auth.New(nil)
	// init self service
	regionSvc = region.New(c)
	feedSvc = feed.New(c, ic)
	externalSvc = external.New(c)
	// conf
	cfg = c
	// sampler
	sampler = initSampler(c)
	featureSvc = feature.New(nil)
}

func initSampler(c *conf.Config) *thread_sampler.ThreadSampler {
	sampler, err := thread_sampler.NewThreadSampler(c)
	if err != nil {
		panic(err)
	}
	return sampler
}

// outerRouter init outer router api path.
func outerRouter(e *bm.Engine) {
	e.Ping(ping)
	// formal api
	e.Use(anticrawler.Report())
	feed := e.Group("/x/feed")
	feed.GET("/region/tags", authSvc.GuestMobile, tags)
	feed.GET("/subscribe/tags", authSvc.UserMobile, subTags)
	feed.POST("/subscribe/tags/add", authSvc.UserMobile, addTag)
	feed.POST("/subscribe/tags/cancel", authSvc.UserMobile, cancelTag)
	feed.GET("/index", authSvc.GuestMobile, feedIndex)
	feed.GET("/index/tab", authSvc.GuestMobile, feedIndexTab)
	feed.GET("/upper", authSvc.UserMobile, feedUpper)
	feed.GET("/upper/archive", authSvc.UserMobile, feedUpperArchive)
	feed.GET("/upper/bangumi", authSvc.UserMobile, feedUpperBangumi)
	feed.GET("/upper/recent", authSvc.UserMobile, feedUpperRecent)
	feed.GET("/upper/article", authSvc.UserMobile, feedUpperArticle)
	feed.GET("/upper/unread/count", authSvc.UserMobile, feedUnreadCount)
	feed.GET("/dislike", authSvc.GuestMobile, feedDislike)
	feed.GET("/dislike/cancel", authSvc.GuestMobile, feedDislikeCancel)

	feedV2 := e.Group("/x/v2/feed")
	feedV2.GET("/index", authSvc.GuestMobile, arcmid.BatchPlayArgs(), sessionRecorder(feedSvc.RecordSession), abtest.Handler(), featureSvc.BuildLimitHttp(), feedSvc.CustomizedMetric(), feedIndex2)
	feedV2.GET("/index/tab", authSvc.GuestMobile, arcmid.BatchPlayArgs(), feedIndexTab2)
	feedV2.GET("/index/converge", authSvc.GuestMobile, arcmid.BatchPlayArgs(), feedIndexConverge)
	feedV2.GET("/index/ai/converge", authSvc.GuestMobile, arcmid.BatchPlayArgs(), feedIndexAvConverge)
	feedV2.GET("/index/vertical/tab", authSvc.GuestMobile, arcmid.BatchPlayArgs(), verticalTab)
	feedV2.GET("/index/vertical/tab/tag", authSvc.GuestMobile, verticalTabTag)
	feedV2.GET("/index/interest", authSvc.GuestMobile, abtest.Handler(), feedIndexInterest)

	// live dynamic
	external := e.Group("/x/feed/external")
	external.GET("/dynamic/count", dynamicCount)
	external.GET("/dynamic/new", dynamicNew)
	external.GET("/dynamic/history", dynamicHistory)
}

func sessionRecorder(recordFn func(*session.IndexSession)) func(*bm.Context) {
	return func(ctx *bm.Context) {
		header := ctx.Request.Header
		query := ctx.Request.URL.Query()
		mid := func() int64 {
			v, ok := ctx.Get("mid")
			if ok {
				return v.(int64)
			}
			return 0
		}()
		si := &session.IndexSession{
			ID:   uuid.New().String(),
			Time: time.Now().UnixNano(),
			Mid:  mid,
		}
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "panic: session recorder context: %+v", si)
				panic(r)
			}
		}()
		si.Request.Header = header
		si.Request.Query = query
		ctx.Context = session.NewContext(ctx.Context, si)
		ctx.Next()
		if ctx.Error != nil {
			// ignore an error response
			return
		}
		if isSampled(ctx) {
			recordFn(si)
		}
	}
}

func setSessionRecordResponse(ctx *bm.Context, data interface{}) {
	si, ok := session.FromContext(ctx)
	if !ok {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	si.Response = string(raw)
	ctx.Context = newContextWithIsSampled(ctx.Context, data)
}

func newContextWithIsSampled(ctx context.Context, data interface{}) context.Context {
	if err := sampler.Reload(); err != nil {
		log.Error("Fail to reload sampler, because=%+v", err)
	}
	index := buildSampleIndex(data)
	isSampled, _ := sampler.IsSampled(index)

	return thread_sampler.NewContext(ctx, isSampled)
}

func isSampled(ctx *bm.Context) bool {
	isSampled, ok := thread_sampler.FromContext(ctx)
	if ok && isSampled {
		return true
	}
	return false
}

func buildSampleIndex(data interface{}) string {
	handlers, ok := data.([]card.Handler)
	if !ok {
		return ""
	}
	stringSet := sets.NewString()
	for _, handler := range handlers {
		base := handler.Get()
		if base == nil {
			continue
		}
		stringSet.Insert(string(base.CardType))
	}

	return strings.Join(stringSet.List(), ",")
}

// Close
func Close() {
	feedSvc.Close()
}
