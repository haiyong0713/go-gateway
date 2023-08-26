package anticrawler

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler/model"
)

type Config struct {
	LogID  string
	Worker int
	Buffer int
	Infoc  *infoc.Config
}

// Send is send func.
type Send func(ctx context.Context, data interface{}) error

// Filter is filter func.
type Filter func(context.Context) bool

// antiCrawler is the common anti-crawler instance.
type antiCrawler struct {
	send Send
}

var (
	_antiCrawler *antiCrawler
	once         sync.Once
)

func Init(c *Config) {
	once.Do(func() {
		_antiCrawler = Default(c)
	})
}

// Default return default anti-crawler.
func Default(c *Config) *antiCrawler {
	if c == nil {
		c = &Config{}
	}
	if c.LogID == "" {
		c.LogID = "009236"
	}
	if c.Worker == 0 {
		c.Worker = 10
	}
	if c.Buffer == 0 {
		c.Buffer = 10240
	}
	if c.Infoc == nil {
		c.Infoc = &infoc.Config{
			Name:    "anticrawler.log",
			Rotated: true,
		}
	}
	return &antiCrawler{
		send: AsyncInfocSend(c),
	}
}

// AsyncInfocSend return async infoc send.
func AsyncInfocSend(c *Config) Send {
	acInfoc, err := infoc.New(c.Infoc)
	if err != nil {
		panic(err)
	}
	cache := fanout.New("anticrawler_cache", fanout.Worker(c.Worker), fanout.Buffer(c.Buffer))
	return func(ctx context.Context, data interface{}) error {
		v, ok := data.(*model.InfocMsg)
		if !ok {
			return nil
		}
		if !isSample(v.Path, v.Sample) {
			return nil
		}
		return cache.Do(ctx, func(ctx context.Context) {
			payload := infoc.NewLogStream(c.LogID, v.Mid, v.Buvid, v.Host, v.Path, v.Method, v.Header, v.Query, v.Body, v.Referer, v.IP, v.Ctime, v.ResponseHeader, v.ResponseBody, v.Sample)
			if err := acInfoc.Info(ctx, payload); err != nil {
				log.Error("failed to send infoc error:%+v", err)
				return
			}
			log.Info("success to send infoc %+v,%+v,%+v,%+v,%+v", v.Mid, v.Buvid, v.Host, v.Path, v.Sample)
		})
	}
}
