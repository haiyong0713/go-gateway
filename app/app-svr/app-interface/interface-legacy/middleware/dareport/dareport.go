package dareport

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
)

const (
	_degradePrefix = "bm.degrade"
)

type DegradeArgsConf struct {
	Infoc *infocv2.Config
	LogID string
}

type DaReport struct {
	logID string
	ic    infocv2.Infoc
	cache *fanout.Fanout
}

func NewDaReport(cfg *DegradeArgsConf) (*DaReport, error) {
	ic, err := infocv2.New(cfg.Infoc)
	if err != nil {
		return nil, err
	}
	if cfg.LogID == "" {
		return nil, errors.New("logID is empty")
	}
	return &DaReport{
		logID: cfg.LogID,
		ic:    ic,
		cache: fanout.New("da-report", fanout.Worker(8)),
	}, nil
}

func (da *DaReport) Report(args []string) bm.HandlerFunc {
	return func(ctx *bm.Context) {
		logArgs := da.requestLogArgs(ctx, args)
		payload := infocv2.NewLogStreamV(da.logID, logArgs...)
		_ = da.Info(ctx.Context, payload)
	}
}

func (da *DaReport) Info(c context.Context, payload infocv2.Payload) error {
	return da.cache.Do(c, func(ctx context.Context) {
		_ = da.ic.Info(ctx, payload)
	})
}

func (da *DaReport) Close() error {
	if err := da.cache.Close(); err != nil {
		return err
	}
	return da.ic.Close()
}

func (da *DaReport) requestLogArgs(ctx *bm.Context, args []string) []log.D {
	params := ctx.Request.Form
	logArgs := make([]log.D, 0, len(args)+2)
	extArgs := []log.D{
		log.KV("key", da.cacheKey(ctx, args)),
		log.KV("ctime", time.Now().Unix()),
	}
	logArgs = append(logArgs, extArgs...)
	for _, arg := range args {
		logArgs = append(logArgs, log.KV(arg, params.Get(arg)))
	}
	return logArgs
}

func (da *DaReport) cacheKey(ctx *bm.Context, args []string) string {
	req := ctx.Request
	path := ctx.RoutePath
	params := req.Form

	vs := make([]string, 0, len(args))
	for _, arg := range args {
		vs = append(vs, params.Get(arg))
	}
	return fmt.Sprintf("%s:%s_%x", _degradePrefix, strings.Replace(path, "/", "_", -1), md5.Sum([]byte(strings.Join(vs, "-"))))
}
